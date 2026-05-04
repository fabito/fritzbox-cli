package aha

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
)

// LEDStatus represents the LED status response from Fritz!Box
type LEDStatus struct {
	LEDDisplay int `json:"ledDisplay"`
	CanDim     int `json:"canDim"`
	DimValue   int `json:"dimValue"`
}

// GetLEDStatus retrieves LED status from Fritz!Box using AHA-HTTP-Interface
func (c *Client) GetLEDStatus() (*LEDStatus, error) {
	// Get SID
	slog.Debug("GetLEDStatus: calling getSID()")
	getSID := c.getSID
	sid, err := getSID()
	//sid, err := c.getSID()
	if err != nil {
		return nil, fmt.Errorf("failed to get SID: %w", err)
	}
	slog.Debug("GetLEDStatus: got SID", "sid", sid)

	// Build URL with SID
	ledURL := fmt.Sprintf("%s/data.lua", c.baseURL)
	formData := url.Values{
		"xhr":  {"1"},
		"sid":  {sid},
		"page": {"led"},
	}

	// Make POST request
	resp, err := c.httpClient.PostForm(ledURL, formData)
	if err != nil {
		return nil, fmt.Errorf("failed to get LED status: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	slog.Debug("GetLEDStatus: response received", "body", string(body))

	// Parse JSON response
	var status LEDStatus
	if err := json.Unmarshal(body, &status); err != nil {
		return nil, fmt.Errorf("failed to parse JSON response: %w", err)
	}

	return &status, nil
}

// SetLEDEnabled sets the LED state (on/off)
// enabled: true = ON (led_display=0), false = OFF (led_display=2)
func (c *Client) SetLEDEnabled(enabled bool) error {
	// Get SID
	slog.Debug("SetLEDEnabled: calling getSID()")
	sid, err := c.getSID()
	if err != nil {
		return fmt.Errorf("failed to get SID: %w", err)
	}
	slog.Debug("SetLEDEnabled: got SID", "sid", sid)

	// Determine led_display value
	// From FritzBoxShell: led_display=0 -> ON, led_display=2 -> OFF
	ledDisplay := "0"
	if !enabled {
		ledDisplay = "2"
	}

	// Build URL and POST data
	ledURL := fmt.Sprintf("%s/data.lua", c.baseURL)
	formData := url.Values{
		"sid":         {sid},
		"page":        {"led"},
		"led_display": {ledDisplay},
		"apply":       {""},
	}

	// Make POST request
	resp, err := c.httpClient.PostForm(ledURL, formData)
	if err != nil {
		return fmt.Errorf("failed to set LED state: %w", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to set LED state: HTTP %d", resp.StatusCode)
	}

	slog.Debug("SetLEDEnabled: success", "enabled", enabled)
	return nil
}
