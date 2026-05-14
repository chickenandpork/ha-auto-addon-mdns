// Package mdns provides mDNS advertisement for addon services.
package mdns

import (
	"fmt"
	"log"
	"net"
	"sync"

	"github.com/jezek/gomdns"
)

// Advertiser publishes addon service names via mDNS.
type Advertiser struct {
	mu       sync.Mutex
	server   gomdns.Server
	records  map[string]gomdns.Record
	stopChan chan struct{}
}

// Service represents an addon service for mDNS advertisement.
type Service struct {
	// mDNS name to advertise (e.g. "my-addon.local")
	MdnsName string
	// IP is the service IP address
	IP net.IP
}

// NewService creates a new mDNS service entry.
func NewService(name string, ip net.IP) Service {
	return Service{
		MdnsName: name,
		IP:       ip,
	}
}

// NewAdvertiser creates a new mDNS advertiser.
func NewAdvertiser() *Advertiser {
	return &Advertiser{
		records:  make(map[string]gomdns.Record),
		stopChan: make(chan struct{}),
	}
}

// Start begins publishing mDNS records.
func (a *Advertiser) Start(localHostname string) error {
	server, err := gomdns.NewServer(fmt.Sprintf(":%d", 5353), gomdns.Options{
		LocalHostname: localHostname,
		InterfaceIndex: -1,
	})
	if err != nil {
		return fmt.Errorf("create mDNS server: %w", err)
	}

	a.server = server
	go func() {
		if err := server.Start(); err != nil {
			log.Printf("mDNS server error: %v", err)
		}
	}()
	return nil
}

// AddService advertises a service via mDNS.
func (a *Advertiser) AddService(name string, ip net.IP) {
	if ip == nil {
		return
	}
	a.mu.Lock()
	a.records[name] = gomdns.Record{
		DnsType:  gomdns.TYPE_A,
		DnsClass: gomdns.CLASS_IN,
		Ttl:      120,
		Name:     name,
		Address:  ip.To4().String(),
	}
	a.mu.Unlock()
	if a.server != nil {
		a.server.Announce(gomdns.TXT{
			Domain: name,
			Txt:    []string{"ha-addon"},
		})
	}
}

// RemoveService stops advertising a service.
func (a *Advertiser) RemoveService(name string) {
	a.mu.Lock()
	delete(a.records, name)
	a.mu.Unlock()
	if a.server != nil {
		a.server.Retract(gomdns.TXT{
			Domain: name,
			Txt:    []string{"ha-addon"},
		})
	}
}

// UpdateServices replaces all mDNS records.
func (a *Advertiser) UpdateServices(services []Service) {
	a.mu.Lock()
	a.records = make(map[string]gomdns.Record)
	a.mu.Unlock()

	for _, svc := range services {
		if svc.IP != nil {
			a.AddService(svc.MdnsName, svc.IP)
		}
	}
}

// Stop stops the mDNS advertiser.
func (a *Advertiser) Stop() {
	close(a.stopChan)
	if a.server != nil {
		a.server.Stop()
	}
}
