// Package ha provides a Home Assistant Supervisor API client.
package ha

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

// Client queries the Home Assistant Supervisor API for addon information.
type Client struct {
	baseURL    *url.URL
	httpClient *http.Client
	// token is the supervisor token for authentication
	token string
}

// NewClient creates a new Home Assistant Supervisor client.
func NewClient(baseURL string, token string) (*Client, error) {
	u, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("parse supervisor URL: %w", err)
	}
	return &Client{
		baseURL:    u,
		httpClient: &http.Client{Timeout: 10 * time.Second},
		token:      token,
	}, nil
}

// ListAddons returns all available addons from the Supervisor API.
func (c *Client) ListAddons(ctx context.Context) ([]Addon, error) {
	path := c.baseURL.ResolveReference(&url.URL{Path: "/api/addons"})
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, path.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("list addons: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("list addons: unexpected status %d: %s", resp.StatusCode, string(body))
	}

	var result struct {
		Ok     bool       `json:"ok"`
		Data   []Addon    `json:"data"`
		Error  string     `json:"error,omitempty"`
		Message string   `json:"message,omitempty"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}
	if !result.Ok {
		return nil, fmt.Errorf("list addons: API error: %s", result.Message)
	}

	// Also try the v2 endpoint if v1 returned empty
	if len(result.Data) == 0 {
		return c.listAddonsV2(ctx)
	}

	return result.Data, nil
}

// listAddonsV2 tries the /api/hassio/addons endpoint (v2 format).
func (c *Client) listAddonsV2(ctx context.Context) ([]Addon, error) {
	path := c.baseURL.ResolveReference(&url.URL{Path: "/api/hassio/addons"})
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, path.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("list addons v2: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("list addons v2: unexpected status %d", resp.StatusCode)
	}

	var result map[string]json.RawMessage
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode v2 response: %w", err)
	}

	// Try data key first
	if data, ok := result["data"]; ok {
		var addons []Addon
		if err := json.Unmarshal(data, &addons); err != nil {
			return nil, fmt.Errorf("unmarshal addons: %w", err)
		}
		return addons, nil
	}

	return nil, nil
}

// GetAddonDetails retrieves detailed info for a specific addon by slug.
func (c *Client) GetAddonDetails(ctx context.Context, slug string) (*Addon, error) {
	path := c.baseURL.ResolveReference(&url.URL{Path: fmt.Sprintf("/api/addons/%s/info", slug)})
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, path.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("get addon details: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("get addon details: unexpected status %d", resp.StatusCode)
	}

	var result struct {
		Ok     bool       `json:"ok"`
		Data   *AddonInfo `json:"data"`
		Error  string     `json:"error,omitempty"`
		Message string   `json:"message,omitempty"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}
	if !result.Ok {
		return nil, fmt.Errorf("get addon details: API error: %s", result.Message)
	}

	if result.Data == nil {
		return nil, nil
	}

	return &Addon{
		Slug:         result.Data.Slug,
		Name:         result.Data.Name,
		Hostname:     result.Data.Hostname,
		IP:           result.Data.IP,
		Port:         result.Data.Port,
		SSL:          result.Data.SSL,
		WebUI:        result.Data.WebUI,
		Version:      result.Data.Version,
		StartOnBoot:  result.Data.Startup == "always",
	}, nil
}

// AddonInfo is the detailed info response from the Supervisor.
type AddonInfo struct {
	Slug        string              `json:"slug"`
	Name        string              `json:"name"`
	Hostname    string              `json:"hostname"`
	IP          string              `json:"ip"`
	Port        int                 `json:"port"`
	SSL         bool                `json:"ssl"`
	WebUI       string              `json:"webui"`
	Version     string              `json:"version"`
	Startup     string              `json:"startup"`
	Ports       map[string]int      `json:"ports"`
	DNS         []string            `json:"dns"`
	HostNetwork bool                `json:"host_network"`
}
