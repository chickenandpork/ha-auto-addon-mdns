package ha

import (
	"net/url"
	"testing"
)

func TestNewAddonService(t *testing.T) {
	addon := Addon{
		Slug: "test-addon",
		Name: "Test Addon",
		Port: 8080,
		SSL:  true,
	}
	svc := NewAddonService("my-addon.local", "test-addon", map[string]int{"http": 8080}, 80, "https", addon)

	if svc.MdnsName != "my-addon.local" {
		t.Errorf("MdnsName = %q, want %q", svc.MdnsName, "my-addon.local")
	}
	if svc.Hostname != "test-addon" {
		t.Errorf("Hostname = %q, want %q", svc.Hostname, "test-addon")
	}
	if svc.ProxyPort != 80 {
		t.Errorf("ProxyPort = %d, want %d", svc.ProxyPort, 80)
	}
	if svc.Protocol != "https" {
		t.Errorf("Protocol = %q, want %q", svc.Protocol, "https")
	}
	if svc.Original.Name != "Test Addon" {
		t.Errorf("Original.Name = %q, want %q", svc.Original.Name, "Test Addon")
	}
	if svc.ServicePorts["http"] != 8080 {
		t.Errorf("ServicePorts[http] = %d, want 8080", svc.ServicePorts["http"])
	}
}

func TestAddonServiceGetters(t *testing.T) {
	addon := Addon{Slug: "test"}
	svc := NewAddonService("test.local", "test", map[string]int{"http": 80}, 80, "http", addon)

	if got := svc.GetMDNSName(); got != "test.local" {
		t.Errorf("GetMDNSName() = %q, want %q", got, "test.local")
	}
	if got := svc.GetHostname(); got != "test" {
		t.Errorf("GetHostname() = %q, want %q", got, "test")
	}
	if got := svc.GetProtocol(); got != "http" {
		t.Errorf("GetProtocol() = %q, want %q", got, "http")
	}
	if got := svc.GetProxyPort(); got != 80 {
		t.Errorf("GetProxyPort() = %d, want %d", got, 80)
	}
	if got := svc.GetServicePorts(); got == nil {
		t.Error("GetServicePorts() returned nil, want non-nil")
	}
}

func TestAddonDefaultValues(t *testing.T) {
	a := Addon{Slug: "test"}
	if a.Name != "" {
		t.Errorf("Name = %q, want empty", a.Name)
	}
	if a.IP != "" {
		t.Errorf("IP = %q, want empty", a.IP)
	}
	if a.Port != 0 {
		t.Errorf("Port = %d, want 0", a.Port)
	}
	if a.SSL {
		t.Error("SSL should be false by default")
	}
	if a.StartOnBoot {
		t.Error("StartOnBoot should be false by default")
	}
}

func TestNewClient(t *testing.T) {
	c, err := NewClient("http://localhost:8123", "test-token")
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	if c.token != "test-token" {
		t.Errorf("token = %q, want %q", c.token, "test-token")
	}
	if c.baseURL.Host != "localhost:8123" {
		t.Errorf("baseURL.Host = %q, want %q", c.baseURL.Host, "localhost:8123")
	}
}

func TestNewClientInvalidURL(t *testing.T) {
	_, err := NewClient("://invalid", "token")
	if err == nil {
		t.Error("NewClient() expected error for invalid URL, got nil")
	}
}

func TestNewClientBaseURL(t *testing.T) {
	c, err := NewClient("https://hass.example.com:8123", "token")
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	if c.baseURL.Scheme != "https" {
		t.Errorf("baseURL.Scheme = %q, want %q", c.baseURL.Scheme, "https")
	}
	if c.baseURL.Host != "hass.example.com:8123" {
		t.Errorf("baseURL.Host = %q, want %q", c.baseURL.Host, "hass.example.com:8123")
	}
}

func TestNewClientPathResolution(t *testing.T) {
	c, err := NewClient("http://hass.example.com", "token")
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	path := c.baseURL.ResolveReference(&url.URL{Path: "/api/addons"})
	if path.Path != "/api/addons" {
		t.Errorf("resolved path = %q, want %q", path.Path, "/api/addons")
	}
}
