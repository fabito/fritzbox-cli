package aha

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
)

// Profile represents a device profile
type Profile struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// ListProfiles retrieves all device profiles from Fritz!Box
func (c *Client) ListProfiles() ([]Profile, error) {
	// Get SID
	sid, err := c.getSID()
	if err != nil {
		return nil, fmt.Errorf("failed to get SID: %w", err)
	}

	// Build URL and POST data
	profileURL := fmt.Sprintf("%s/data.lua", c.baseURL)
	formData := url.Values{
		"xhr":  {"1"},
		"sid":  {sid},
		"page": {"kisi_profilelist"},
	}

	// Make POST request
	resp, err := c.httpClient.PostForm(profileURL, formData)
	if err != nil {
		return nil, fmt.Errorf("failed to get profile list: %w", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get profile list: HTTP %d", resp.StatusCode)
	}

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	slog.Debug("ListProfiles: response received", "body", string(body))

	// Parse JSON response
	var result map[string]json.RawMessage
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse JSON response: %w", err)
	}

	// Extract profiles from nested structure: data.vars.kisi.profiles
	data, ok := result["data"]
	if !ok {
		return nil, fmt.Errorf("no data in response")
	}

	var dataObj struct {
		Vars struct {
			Kisi struct {
				Profiles []Profile `json:"profiles"`
			} `json:"kisi"`
		} `json:"vars"`
	}

	if err := json.Unmarshal(data, &dataObj); err != nil {
		return nil, fmt.Errorf("failed to parse data: %w", err)
	}

	return dataObj.Vars.Kisi.Profiles, nil
}

// GetDeviceProfile retrieves the current profile ID for a device
// Returns the profile ID (e.g., "1", "2") or empty string if no profile set
func (c *Client) GetDeviceProfile(deviceID string) (string, error) {
	// Get SID
	sid, err := c.getSID()
	if err != nil {
		return "", fmt.Errorf("failed to get SID: %w", err)
	}

	// Build URL and POST data
	profileURL := fmt.Sprintf("%s/data.lua", c.baseURL)
	formData := url.Values{
		"xhr":  {"1"},
		"sid":  {sid},
		"page": {"edit_device"},
		"dev":  {deviceID},
	}

	// Make POST request
	resp, err := c.httpClient.PostForm(profileURL, formData)
	if err != nil {
		return "", fmt.Errorf("failed to get device profile: %w", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to get device profile: HTTP %d", resp.StatusCode)
	}

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	slog.Debug("GetDeviceProfile: response received", "body", string(body))

	// Parse JSON response
	var result map[string]json.RawMessage
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("failed to parse JSON response: %w", err)
	}

	// Extract selected profile from: data.vars.dev.netAccess.kisi.profiles.selected
	data, ok := result["data"]
	if !ok {
		return "", fmt.Errorf("no data in response")
	}

	var dataObj struct {
		Vars struct {
			Dev struct {
				NetAccess struct {
					Kisi struct {
						Profiles struct {
							Selected string `json:"selected"`
						} `json:"profiles"`
					} `json:"kisi"`
				} `json:"netAccess"`
			} `json:"dev"`
		} `json:"vars"`
	}

	if err := json.Unmarshal(data, &dataObj); err != nil {
		return "", fmt.Errorf("failed to parse data: %w", err)
	}

	selected := dataObj.Vars.Dev.NetAccess.Kisi.Profiles.Selected

	// Convert "filtprofXXXX" to "XXXX"
	if strings.HasPrefix(selected, "filtprof") {
		return strings.TrimPrefix(selected, "filtprof"), nil
	}

	// Return empty string if no profile selected
	return "", nil
}

// SetDeviceProfile sets the profile for a device
func (c *Client) SetDeviceProfile(deviceID string, profileID string) error {
	// Get SID
	sid, err := c.getSID()
	if err != nil {
		return fmt.Errorf("failed to get SID: %w", err)
	}

	// Build profile ID in format "filtprofXXXX"
	filtprofID := fmt.Sprintf("filtprof%s", profileID)

	// Get device name (required by FritzBoxShell script)
	// For now, we'll use empty string - the script uses dev_name but it might not be strictly required
	deviceName := ""

	// Build URL and POST data
	profileURL := fmt.Sprintf("%s/data.lua", c.baseURL)
	formData := url.Values{
		"sid":          {sid},
		"dev_name":      {deviceName},
		"dev":           {deviceID},
		"kisi_profile":  {filtprofID},
		"page":          {"edit_device"},
		"apply":         {"true"},
	}

	// Make POST request
	resp, err := c.httpClient.PostForm(profileURL, formData)
	if err != nil {
		return fmt.Errorf("failed to set device profile: %w", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to set device profile: HTTP %d", resp.StatusCode)
	}

	slog.Debug("SetDeviceProfile: success", "deviceID", deviceID, "profileID", profileID)
	return nil
}
