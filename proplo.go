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

var errNetClosing = "use of closed network connection"
var encoder = json.NewEncoder(os.Stdout)

// LogConnect represents log at connected
type LogConnect struct {
	ID           string    `json:"id"`
	Type         string    `json:"type"`
	Time         time.Time `json:"time"`
	ClientAddr   string    `json:"client_addr"`
	ProxyAddr    string    `json:"proxy_addr"`
	UpstreamAddr string    `json:"upstream_addr"`
	Status       string    `json:"status"`
	ClientAt     time.Time `json:"client_at"`
	UpstreamAt   time.Time `json:"upstream_at"`
}

func (l LogConnect) Print(status string) error {
	l.Type = "connect"
	l.Status = status
	l.Time = time.Now()
	return encoder.Encode(l)
}

type LogProxy struct {
	ID        string    `json:"id"`
	Type      string    `json:"type"`
	Time      time.Time `json:"time"`
	SrcAddr   string    `json:"src_addr"`
	ProxyAddr string    `json:"proxy_addr"`
	DestAddr  string    `json:"dest_addr"`
	Bytes     int64     `json:"bytes"`
	Duration  float64   `json:"duration"`
	Error     error     `json:"error"`
}

func (l LogProxy) Print() error {
	l.Type = "transfer"
	l.Time = time.Now()
	return encoder.Encode(l)
}

// Run runs proplo
func Run(ctx context.Context, localAddr, upstreamAddr string) error {
	log.Println("[info] Upstream", upstreamAddr)
	log.Println("[info] Listening", localAddr)
	l, err := net.Listen("tcp", localAddr)
	if err != nil {
		log.Fatalf("couldn't listen to %q: %q\n", localAddr, err.Error())
	}

	// Wrap listener in a proxyproto listener
	proxyListener := &proxyproto.Listener{Listener: l}
	defer proxyListener.Close()

	// Wait for a connection and accept it
	for {
		conn, err := proxyListener.Accept()
		if err != nil {
			log.Println("[error]", err)
			continue
		}
		go proxy(ctx, conn, upstreamAddr)
	}
}

func proxy(ctx context.Context, clientConn net.Conn, upstreamAddr string) {
	id := uuid.New()
	start := time.Now()
	defer clientConn.Close()
	logConnect := LogConnect{
		ID:           id.String(),
		ClientAt:     start,
		ClientAddr:   clientConn.RemoteAddr().String(),
		UpstreamAddr: upstreamAddr,
	}
	d := &net.Dialer{
		Timeout: time.Second * 30,
	}
	upstreamConn, err := d.DialContext(ctx, "tcp", upstreamAddr)
	if err != nil {
		log.Println("[error] couldn't dial to upstream", err)
		logConnect.Print("upstream_failed")
		return
	}
	defer upstreamConn.Close()
	logConnect.ProxyAddr = upstreamConn.LocalAddr().String()
	logConnect.UpstreamAddr = upstreamConn.RemoteAddr().String()
	logConnect.UpstreamAt = time.Now()
	logConnect.Print("connected")

	clientCh := make(chan struct{})
	upstreamCh := make(chan struct{})
	go func() {
		n, err := io.Copy(upstreamConn, clientConn)
		if err != nil && strings.Contains(err.Error(), errNetClosing) {
			err = nil
		}
		logProxy := LogProxy{
			ID:        id.String(),
			SrcAddr:   clientConn.RemoteAddr().String(),
			ProxyAddr: upstreamConn.LocalAddr().String(),
			DestAddr:  upstreamConn.RemoteAddr().String(),
			Bytes:     n,
			Duration:  time.Now().Sub(start).Seconds(),
			Error:     err,
		}
		logProxy.Print()
		clientCh <- struct{}{}
	}()
	go func() {
		n, err := io.Copy(clientConn, upstreamConn)
		if err != nil && strings.Contains(err.Error(), errNetClosing) {
			err = nil
		}
		logProxy := LogProxy{
			ID:        id.String(),
			DestAddr:  clientConn.RemoteAddr().String(),
			ProxyAddr: upstreamConn.LocalAddr().String(),
			SrcAddr:   upstreamConn.RemoteAddr().String(),
			Bytes:     n,
			Duration:  time.Now().Sub(start).Seconds(),
			Error:     err,
		}
		logProxy.Print()
		upstreamCh <- struct{}{}
	}()
	select {
	case <-upstreamCh:
	case <-clientCh:
	}
	return
}
