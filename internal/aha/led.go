package aha

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
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
