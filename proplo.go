package proplo

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/pires/go-proxyproto"
)

var (
	errNetClosing = "use of closed network connection"
	encoder       = json.NewEncoder(os.Stdout)
	clientStr     = "client"
	upstreamStr   = "upstream"
)

var (
	PrintStatusInterval = time.Minute
	UpstreamTimeout     = 30 * time.Second
)

var dashboard = &Dashboard{
	LogStatuses: make(map[string]*LogStatus),
}

// Run runs proplo
func Run(ctx context.Context, opt *Options) error {
	log.Println("[info] Upstream", opt.UpstreamAddr)
	log.Println("[info] Listening", opt.LocalAddr)
	if opt.IgnoreCIDR != "" {
		log.Println("[info] Ingore CIDR", opt.IgnoreCIDR)
	}
	l, err := net.Listen("tcp", opt.LocalAddr)
	if err != nil {
		log.Fatalf("couldn't listen to %q: %q\n", opt.LocalAddr, err.Error())
	}

	// Wrap listener in a proxyproto listener
	proxyListener := &proxyproto.Listener{Listener: l}
	defer proxyListener.Close()
	go printStatus()

	// Wait for a connection and accept it
	for {
		conn, err := proxyListener.Accept()
		if err != nil {
			log.Println("[error]", err)
			continue
		}
		go proxy(ctx, conn, opt)
	}
}

func proxy(ctx context.Context, clientConn net.Conn, opt *Options) {
	id := uuid.New()
	start := time.Now()
	defer clientConn.Close()

	clientAddr := clientConn.RemoteAddr().String()
	clientHost, _, err := net.SplitHostPort(clientAddr)
	log.Println("[debug] clientAddr", clientAddr, "clientHost", clientHost)
	if clientIP := net.ParseIP(clientHost); clientIP != nil {
		log.Println("[debug] clientIP", clientIP)
		if opt.Ignore(clientIP) {
			log.Println("[debug] ignore client addr", clientConn.RemoteAddr().String())
			return
		}
	}

	logConnect := &LogConnect{
		ID:           id.String(),
		ClientAt:     start,
		ClientAddr:   clientConn.RemoteAddr().String(),
		UpstreamAddr: opt.UpstreamAddr,
	}
	d := &net.Dialer{
		Timeout: UpstreamTimeout,
	}
	upstreamConn, err := d.DialContext(ctx, "tcp", opt.UpstreamAddr)
	if err != nil {
		log.Println("[error] couldn't dial to upstream", err)
		logConnect.Error = err
		logConnect.Print("upstream_failed")
		return
	}
	defer upstreamConn.Close()
	logConnect.ProxyAddr = upstreamConn.LocalAddr().String()
	logConnect.UpstreamAddr = upstreamConn.RemoteAddr().String()
	logConnect.UpstreamAt = time.Now()
	logConnect.Print("connected")

	dashboard.Post(logConnect)
	defer dashboard.Remove(id.String())

	clientCh := make(chan struct{})
	upstreamCh := make(chan struct{})
	go func() {
		n, err := io.Copy(upstreamConn, clientConn)
		if err != nil && strings.Contains(err.Error(), errNetClosing) {
			err = nil
		}
		l := &LogDisconnect{
			ID:           id.String(),
			ClientAddr:   clientConn.RemoteAddr().String(),
			ProxyAddr:    upstreamConn.LocalAddr().String(),
			UpstreamAddr: upstreamConn.RemoteAddr().String(),
			Bytes:        n,
			Duration:     time.Now().Sub(start).Seconds(),
			Error:        err,
			Src:          clientStr,
			Dest:         upstreamStr,
		}
		l.Print()
		clientCh <- struct{}{}
	}()
	go func() {
		n, err := io.Copy(clientConn, upstreamConn)
		if err != nil && strings.Contains(err.Error(), errNetClosing) {
			err = nil
		}
		l := &LogDisconnect{
			ID:           id.String(),
			ClientAddr:   clientConn.RemoteAddr().String(),
			ProxyAddr:    upstreamConn.LocalAddr().String(),
			UpstreamAddr: upstreamConn.RemoteAddr().String(),
			Bytes:        n,
			Duration:     time.Now().Sub(start).Seconds(),
			Error:        err,
			Src:          upstreamStr,
			Dest:         clientStr,
		}
		l.Print()
		upstreamCh <- struct{}{}
	}()

	select {
	case <-upstreamCh:
	case <-clientCh:
	}
	return
}

func printStatus() {
	ticker := time.NewTicker(PrintStatusInterval)
	for range ticker.C {
		dashboard.Print()
	}
}
