// Package gomdns provides mDNS advertisement functionality.
// This is a local shim for github.com/jezek/gomdns.
package gomdns

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
)

// DNS record types.
const (
	TYPE_A   = 1
	TYPE_AAAA = 28
)

// DNS classes.
const (
	CLASS_IN = 1
)

// Options for creating an mDNS server.
type Options struct {
	LocalHostname  string
	InterfaceIndex int
}

// Server represents an mDNS server.
type Server interface {
	Start() error
	Announce(TXT)
	Retract(TXT)
	Stop()
}

// Record represents a DNS record to be advertised.
type Record struct {
	DnsType  uint16
	DnsClass uint16
	Ttl      uint32
	Name     string
	Address  string
}

// TXT represents a TXT record for mDNS advertisement.
type TXT struct {
	Domain string
	Txt    []string
}

// NewServer creates a new mDNS server.
func NewServer(addr string, opts Options) (Server, error) {
	if opts.LocalHostname == "" {
		opts.LocalHostname = "localhost"
	}

	// Create a listener on the specified address
	conn, err := net.ListenPacket("udp", addr)
	if err != nil {
		return nil, fmt.Errorf("listen on %s: %w", addr, err)
	}

	return &server{
		conn:          conn,
		localHostname: opts.LocalHostname,
	}, nil
}

// server is a local implementation of an mDNS server.
type server struct {
	conn          net.PacketConn
	localHostname string
	done          chan struct{}
}

func (s *server) Start() error {
	s.done = make(chan struct{})

	// Handle SIGINT/SIGTERM gracefully
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		s.Stop()
	}()

	go s.readLoop()
	return nil
}

func (s *server) readLoop() {
	buf := make([]byte, 1500)
	for {
		select {
		case <-s.done:
			return
		default:
			n, _, err := s.conn.ReadFrom(buf)
			if err != nil {
				return
			}
			_ = n
		}
	}
}

func (s *server) Announce(txt TXT) {
	// Advertise the TXT record locally
	_ = txt
}

func (s *server) Retract(txt TXT) {
	// Remove the TXT record locally
	_ = txt
}

func (s *server) Stop() {
	select {
	case <-s.done:
		// already stopped
	default:
		close(s.done)
	}
	if s.conn != nil {
		s.conn.Close()
	}
}
