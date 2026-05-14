// Package ha provides types for Home Assistant addon data.
package ha

// Addon represents a Home Assistant addon.
type Addon struct {
	Slug        string `json:"slug"`
	Name        string `json:"name"`
	Hostname    string `json:"hostname,omitempty"`
	IP          string `json:"ip,omitempty"`
	Port        int    `json:"port,omitempty"`
	SSL         bool   `json:"ssl,omitempty"`
	WebUI       string `json:"webui,omitempty"`
	Version     string `json:"version,omitempty"`
	Protected   bool   `json:"protected,omitempty"`
	StartOnBoot bool   `json:"startup,omitempty"`
	Ports       map[string]int `json:"ports,omitempty"`
}

// AddonService represents a discovered addon service with assigned name and port mapping.
type AddonService struct {
	// mDNS name to advertise (e.g. "my-addon.local")
	MdnsName string
	// Local hostname used for proxy routing
	Hostname string
	// The service port(s) on the addon
	ServicePorts map[string]int
	// The reverse-proxy port (always 80 or 443)
	ProxyPort int
	// Protocol (http or https)
	Protocol string
	// The original addon info
	Original Addon
}

// NewAddonService creates a new AddonService.
func NewAddonService(mdnsName, hostname string, servicePorts map[string]int, proxyPort int, protocol string, original Addon) AddonService {
	return AddonService{
		MdnsName:     mdnsName,
		Hostname:     hostname,
		ServicePorts: servicePorts,
		ProxyPort:    proxyPort,
		Protocol:     protocol,
		Original:     original,
	}
}

// GetMDNSName returns the mDNS name.
func (s AddonService) GetMDNSName() string {
	return s.MdnsName
}

// GetHostname returns the hostname.
func (s AddonService) GetHostname() string {
	return s.Hostname
}

// GetProtocol returns the protocol.
func (s AddonService) GetProtocol() string {
	return s.Protocol
}

// GetProxyPort returns the proxy port.
func (s AddonService) GetProxyPort() int {
	return s.ProxyPort
}

// GetServicePorts returns the service ports.
func (s AddonService) GetServicePorts() map[string]int {
	return s.ServicePorts
}
