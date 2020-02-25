package proplo

import (
	"sync"
	"time"
)

// LogConnect represents log at connected
type LogConnect struct {
	Type         string    `json:"type"`
	Time         time.Time `json:"time"`
	ClientAddr   string    `json:"client_addr"`
	ProxyAddr    string    `json:"proxy_addr"`
	UpstreamAddr string    `json:"upstream_addr"`
	Status       string    `json:"status"`
	Error        error     `json:"error"`
	ClientAt     time.Time `json:"client_at"`
	UpstreamAt   time.Time `json:"upstream_at"`
	ID           string    `json:"id"`
}

// Print prints a log message to STDOUT as JSON.
func (l *LogConnect) Print(status string) error {
	l.Type = "connect"
	l.Status = status
	l.Time = time.Now()
	return encoder.Encode(l)
}

// LogDisconnect represents log at disconnected
type LogDisconnect struct {
	Type         string    `json:"type"`
	Time         time.Time `json:"time"`
	ClientAddr   string    `json:"client_addr"`
	ProxyAddr    string    `json:"proxy_addr"`
	UpstreamAddr string    `json:"upstream_addr"`
	Src          string    `json:"src"`
	Dest         string    `json:"dest"`
	Bytes        int64     `json:"bytes"`
	Duration     float64   `json:"duration"`
	Error        error     `json:"error"`
	ID           string    `json:"id"`
}

// Print prints a log message to STDOUT as JSON.
func (l *LogDisconnect) Print() error {
	l.Type = "disconnect"
	l.Time = time.Now()
	return encoder.Encode(l)
}

// LogStatus represents a log while in connecting.
type LogStatus struct {
	Type         string    `json:"type"`
	Time         time.Time `json:"time"`
	ClientAddr   string    `json:"client_addr"`
	ProxyAddr    string    `json:"proxy_addr"`
	UpstreamAddr string    `json:"upstream_addr"`
	ClientAt     time.Time `json:"client_at"`
	UpstreamAt   time.Time `json:"upstream_at"`
	Duration     float64   `json:"duration"`
	ID           string    `json:"id"`
}

// Print prints a log message to STDOUT as JSON.
func (l *LogStatus) Print() error {
	l.Type = "status"
	l.Time = time.Now()
	l.Duration = l.Time.Sub(l.ClientAt).Seconds()
	return encoder.Encode(l)
}

// LogSummary represents log message for summary of status.
type LogSummary struct {
	Type        string    `json:"type"`
	Time        time.Time `json:"time"`
	Connections int       `json:"connections"`
}

func (l *LogSummary) Print() error {
	l.Type = "summary"
	l.Time = time.Now()
	return encoder.Encode(l)
}

// Dashboard is a log status storage.
type Dashboard struct {
	LogStatuses map[string]*LogStatus
	mu          sync.Mutex
}

func (d *Dashboard) Post(l *LogConnect) {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.LogStatuses[l.ID] = &LogStatus{
		ID:           l.ID,
		ClientAddr:   l.ClientAddr,
		ProxyAddr:    l.ProxyAddr,
		UpstreamAddr: l.UpstreamAddr,
		ClientAt:     l.ClientAt,
		UpstreamAt:   l.UpstreamAt,
	}
}

func (d *Dashboard) Remove(id string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	delete(d.LogStatuses, id)
}

func (d *Dashboard) Print() error {
	d.mu.Lock()
	defer d.mu.Unlock()
	s := &LogSummary{Connections: len(d.LogStatuses)}
	s.Print()
	for _, l := range d.LogStatuses {
		if err := l.Print(); err != nil {
			return err
		}
	}
	return nil
}
