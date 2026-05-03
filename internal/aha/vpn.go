package aha

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
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
