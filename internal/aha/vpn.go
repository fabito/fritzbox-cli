package aha

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// VPNConnection represents a VPN connection on the Fritz!Box
type VPNConnection struct {
	UID     string `json:"UID"`
	Name    string `json:"name"`
	Enabled bool   `json:"enabled"`
	Status  string `json:"status"`
	Type    string `json:"type"`
}

// ListVPNConnections lists all VPN connections
func (c *Client) ListVPNConnections() ([]VPNConnection, error) {
	// Get SID
	sid, err := c.getSID()
	if err != nil {
		return nil, fmt.Errorf("failed to get SID: %w", err)
	}

	url := c.baseURL + "/api/v0/generic/vpn"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Authorization", "AVM-SID "+sid)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call VPN API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("VPN API returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Parse response - it's wrapped in an object with "connection" array
	var result struct {
		Connections []VPNConnection `json:"connection"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse VPN response: %w", err)
	}

	return result.Connections, nil
}

// SetVPNEnabled enables or disables a VPN connection by name
func (c *Client) SetVPNEnabled(name string, enabled bool) error {
	// Step 1: Get SID
	sid, err := c.getSID()
	if err != nil {
		return fmt.Errorf("failed to get SID: %w", err)
	}

	// Step 2: List connections to find the UID by name
	connections, err := c.ListVPNConnections()
	if err != nil {
		return fmt.Errorf("failed to list VPN connections: %w", err)
	}

	// Find connection by name
	var uid string
	for _, conn := range connections {
		if conn.Name == name {
			uid = conn.UID
			break
		}
	}

	if uid == "" {
		return fmt.Errorf("VPN connection '%s' not found", name)
	}

	// Step 3: Make PUT request to enable/disable
	var activated string
	if enabled {
		activated = "1"
	} else {
		activated = "0"
	}

	url := c.baseURL + "/api/v0/generic/vpn/connection/" + uid
	body := fmt.Sprintf(`{"activated":"%s"}`, activated)

	req, err := http.NewRequest("PUT", url, strings.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Authorization", "AVM-SID "+sid)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to call VPN API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("VPN set API returned status %d", resp.StatusCode)
	}

	return nil
}
