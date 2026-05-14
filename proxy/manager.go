// Package proxy manages reverse-proxy configurations for addon services.
package proxy

import (
	"fmt"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"sync"

	"ha-addon-auto-mdns/ha"
)

// Manager handles the lifecycle of reverse-proxy instances.
type Manager struct {
	mu        sync.Mutex
	svcs      []*ServiceConfig
	listeners []*http.Server
}

// ServiceConfig holds the configuration for a single reverse-proxy service.
type ServiceConfig struct {
	// mDNSName is the advertised mDNS hostname (e.g. "my-addon.local")
	MdnsName string
	// HostHeader sets the Host header forwarded to the backend
	HostHeader string
	// BackendURL is the URL of the addon service to proxy to
	BackendURL *url.URL
	// ListenAddr is the address to listen on (e.g. ":80")
	ListenAddr string
	// Protocol is "http" or "https"
	Protocol string
	// Addon is the original addon info
	Addon ha.Addon
}

// NewServiceConfig creates a new ServiceConfig.
func NewServiceConfig(mdnsName, hostHeader string, backendURL *url.URL, listenAddr string, protocol string, addon ha.Addon) ServiceConfig {
	return ServiceConfig{
		MdnsName:   mdnsName,
		HostHeader: hostHeader,
		BackendURL: backendURL,
		ListenAddr: listenAddr,
		Protocol:   protocol,
		Addon:      addon,
	}
}

// NewManager creates a new reverse-proxy manager.
func NewManager() *Manager {
	return &Manager{}
}

// AddService adds a service configuration and starts proxying if not already running.
func (m *Manager) AddService(cfg *ServiceConfig) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check if we already have this service
	for _, existing := range m.svcs {
		if existing.MdnsName == cfg.MdnsName {
			// Update existing
			existing.BackendURL = cfg.BackendURL
			existing.HostHeader = cfg.HostHeader
			return
		}
	}

	m.svcs = append(m.svcs, cfg)
}

// RemoveService removes a service configuration and stops proxying.
func (m *Manager) RemoveService(mDNSName string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for i, svc := range m.svcs {
		if svc.MdnsName == mDNSName {
			m.svcs = append(m.svcs[:i], m.svcs[i+1:]...)
			break
		}
	}
}

// UpdateServices replaces all current service configurations.
func (m *Manager) UpdateServices(services []ha.AddonService) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.svcs = m.svcs[:0]
	for _, svc := range services {
		cfg := m.serviceToConfig(svc)
		if cfg != nil {
			m.svcs = append(m.svcs, cfg)
		}
	}
}

// serviceToConfig converts an AddonService to a ServiceConfig.
func (m *Manager) serviceToConfig(svc ha.AddonService) *ServiceConfig {
	if svc.Original.IP == "" {
		return nil
	}

	// Build backend URL from the addon's service port
	backendProtocol := "http"
	if svc.Protocol == "https" || svc.Original.SSL {
		backendProtocol = "https"
	}

	// Use the primary service port
	backendPort := svc.ProxyPort
	if len(svc.ServicePorts) > 0 {
		// Use the first service port as the default
		for _, port := range svc.ServicePorts {
			backendPort = port
			break
		}
	}

	backendURL := &url.URL{
		Scheme: backendProtocol,
		Host:   net.JoinHostPort(svc.Original.IP, fmt.Sprintf("%d", backendPort)),
	}

	return &ServiceConfig{
		MdnsName:   svc.MdnsName,
		HostHeader: svc.Hostname,
		BackendURL: backendURL,
		ListenAddr: fmt.Sprintf(":%d", svc.ProxyPort),
		Protocol:   svc.Protocol,
		Addon:      svc.Original,
	}
}

// StartProxyServer starts a reverse proxy server that routes based on host header.
func (m *Manager) StartProxyServer(listenAddr string) (*http.Server, error) {
	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		host := r.Host
		// Strip port if present
		host = strings.Split(host, ":")[0]

		m.mu.Lock()
		defer m.mu.Unlock()

		// Find matching service
		for _, svc := range m.svcs {
			svcHost := strings.Split(svc.MdnsName, ".")[0]
			if svcHost == host || svc.MdnsName == host {
				proxy := httputil.NewSingleHostReverseProxy(svc.BackendURL)
				// Set the host header to the backend
				proxy.Transport = &headerRoundTripper{
					RoundTripper: http.DefaultTransport,
					hostHeader:   svc.HostHeader,
				}
				proxy.ServeHTTP(w, r)
				return
			}
		}

		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("Service not found"))
	})

	server := &http.Server{
		Addr:    listenAddr,
		Handler: mux,
	}

	go func() {
		var err error
		if server.Addr == ":443" {
			err = server.ListenAndServeTLS("", "")
		} else {
			err = server.ListenAndServe()
		}
		if err != nil && err != http.ErrServerClosed {
			// Log but don't fail - the server may be killed
		}
	}()

	return server, nil
}

// GetServices returns the current list of service configurations.
func (m *Manager) GetServices() []ServiceConfig {
	m.mu.Lock()
	defer m.mu.Unlock()

	result := make([]ServiceConfig, len(m.svcs))
	for i, svc := range m.svcs {
		result[i] = *svc
	}
	return result
}

// headerRoundTripper wraps an http.RoundTripper to set the Host header.
type headerRoundTripper struct {
	http.RoundTripper
	hostHeader string
}

func (t *headerRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	if t.hostHeader != "" {
		req.Host = t.hostHeader
	}
	return t.RoundTripper.RoundTrip(req)
}
