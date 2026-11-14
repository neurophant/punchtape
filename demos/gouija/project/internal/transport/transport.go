// Package transport provides pluggable transport mechanisms for the
// Gouija two-stage proxy, supporting both raw TCP and HTTP/WebSocket modes.
package transport

import (
	"io"
	"net"
	"time"
)

// Mode represents the transport mode
type Mode string

const (
	// ModeTCP uses raw TCP connections
	ModeTCP Mode = "tcp"

	// ModeWebSocket uses WebSocket over HTTP/HTTPS
	ModeWebSocket Mode = "ws"
)

// Conn is the interface that wraps a transport connection.
// It extends io.ReadWriteCloser with additional connection methods.
type Conn interface {
	io.ReadWriteCloser

	// LocalAddr returns the local network address.
	LocalAddr() net.Addr

	// RemoteAddr returns the remote network address.
	RemoteAddr() net.Addr

	// SetDeadline sets the read and write deadlines.
	SetDeadline(t time.Time) error

	// SetReadDeadline sets the read deadline.
	SetReadDeadline(t time.Time) error

	// SetWriteDeadline sets the write deadline.
	SetWriteDeadline(t time.Time) error
}

// Dialer is the interface for establishing outbound connections.
type Dialer interface {
	// Dial establishes a connection to the address.
	Dial(address string) (Conn, error)

	// DialTimeout establishes a connection with a timeout.
	DialTimeout(address string, timeout time.Duration) (Conn, error)
}

// Listener is the interface for accepting inbound connections.
type Listener interface {
	// Accept waits for and returns the next connection.
	Accept() (Conn, error)

	// Close closes the listener.
	Close() error

	// Addr returns the listener's network address.
	Addr() net.Addr
}

// Config holds transport configuration
type Config struct {
	// Mode specifies the transport mode (tcp or ws)
	Mode Mode

	// Path is the WebSocket endpoint path (for ws mode)
	Path string

	// TLS enables TLS for WebSocket connections (wss://)
	TLS bool

	// TLSCert is the path to TLS certificate (for proxy in ws mode)
	TLSCert string

	// TLSKey is the path to TLS private key (for proxy in ws mode)
	TLSKey string

	// TCP-specific settings (only used for tcp mode)
	// KeepAlive enables TCP keepalive probes
	KeepAlive bool

	// KeepAliveInterval sets the interval between keepalive probes
	KeepAliveInterval time.Duration
}

// DefaultConfig returns the default transport configuration
func DefaultConfig() Config {
	return Config{
		Mode:              ModeTCP,
		Path:              "/ws",
		TLS:               false,
		KeepAlive:         true,
		KeepAliveInterval: 30 * time.Second,
	}
}

// NewDialer creates a new dialer for the specified transport mode
func NewDialer(cfg Config) Dialer {
	switch cfg.Mode {
	case ModeWebSocket:
		return &wsDialer{
			path: cfg.Path,
			tls:  cfg.TLS,
		}
	default:
		tcpCfg := TCPConfig{
			KeepAlive:         cfg.KeepAlive,
			KeepAliveInterval: cfg.KeepAliveInterval,
		}
		// Apply defaults if not set
		if !cfg.KeepAlive && cfg.KeepAliveInterval == 0 {
			tcpCfg = DefaultTCPConfig()
		}
		return &tcpDialer{tcpConfig: tcpCfg}
	}
}

// NewListener creates a new listener for the specified transport mode
func NewListener(cfg Config, address string) (Listener, error) {
	switch cfg.Mode {
	case ModeWebSocket:
		return newWSListener(address, cfg)
	default:
		tcpCfg := TCPConfig{
			KeepAlive:         cfg.KeepAlive,
			KeepAliveInterval: cfg.KeepAliveInterval,
		}
		// Apply defaults if not set
		if !cfg.KeepAlive && cfg.KeepAliveInterval == 0 {
			tcpCfg = DefaultTCPConfig()
		}
		return newTCPListenerWithConfig(address, tcpCfg)
	}
}
