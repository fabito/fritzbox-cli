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

// DeviceProfile represents a device with its assigned profile
type DeviceProfile struct {
	DeviceID   string `json:"device_id"`
	DeviceName string `json:"name"`
	MACAddress string `json:"mac"`
	IPAddress  string `json:"ip"`
	ProfileID  string `json:"profile_id"`
	ProfileName string `json:"profile_name"`
}

// ListDevicesWithProfiles lists all devices with their assigned profiles
// Uses page=netDev&xhrId=all to get device list with profile info
func (c *Client) ListDevicesWithProfiles() ([]DeviceProfile, error) {
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
		"page": {"netDev"},
		"xhrId": {"all"},
	}

	slog.Debug("ListDevicesWithProfiles: requesting", "url", profileURL, "page", "netDev")

	// Make POST request
	resp, err := c.httpClient.PostForm(profileURL, formData)
	if err != nil {
		return nil, fmt.Errorf("failed to get device list: %w", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get device list: HTTP %d", resp.StatusCode)
	}

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	slog.Debug("ListDevicesWithProfiles: response received", "body", string(body))

	// Parse JSON response
	var result map[string]json.RawMessage
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse JSON response: %w", err)
	}

	// Extract devices from data.active array
	data, ok := result["data"]
	if !ok {
		return nil, fmt.Errorf("no data in response")
	}

	var dataObj struct {
		Active []struct {
			Name  string `json:"name"`
			MAC   string `json:"mac"`
			UID   string `json:"UID"`
			IPv4  struct {
				IP string `json:"ip"`
			} `json:"ipv4"`
			// Profile info may be in different fields
			ProfileID   string `json:"profile_id"`
			ProfileName string `json:"profile_name"`
		} `json:"active"`
	}

	if err := json.Unmarshal(data, &dataObj); err != nil {
		slog.Debug("ListDevicesWithProfiles: failed to parse", "error", err)
		return nil, fmt.Errorf("failed to parse data: %w", err)
	}

	slog.Debug("ListDevicesWithProfiles: parsed devices", "count", len(dataObj.Active))

	// Convert to DeviceProfile slice
	devices := make([]DeviceProfile, 0, len(dataObj.Active))
	for _, d := range dataObj.Active {
		dp := DeviceProfile{
			DeviceID:   d.UID,
			DeviceName: d.Name,
			MACAddress: d.MAC,
			IPAddress:  d.IPv4.IP,
			ProfileID:  d.ProfileID,
			ProfileName: d.ProfileName,
		}
		slog.Debug("ListDevicesWithProfiles: device", "name", d.Name, "mac", d.MAC, "ip", d.IPv4.IP)
		devices = append(devices, dp)
	}

	return devices, nil
}

// ListProfiles retrieves all available device profiles from Fritz!Box
// Returns profile definitions like "Standard", "Restricted", etc.
func (c *Client) ListProfiles() ([]Profile, error) {
	// Get SID
	sid, err := c.getSID()
	if err != nil {
		return nil, fmt.Errorf("failed to get SID: %w", err)
	}

	// Build URL and POST data
	// From FritzBoxShell: page=kisi_profilelist
	profileURL := fmt.Sprintf("%s/data.lua", c.baseURL)
	formData := url.Values{
		"xhr":  {"1"},
		"sid":  {sid},
		"page": {"kisi_profilelist"},
	}

	slog.Debug("ListProfiles: requesting", "url", profileURL, "page", "kisi_profilelist")

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

	// Parse JSON response - format from FritzBoxShell getProfileName function
	// The response is an array of [id, name] pairs or an object with profiles
	var raw json.RawMessage
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("failed to parse JSON response: %w", err)
	}

	// Try to parse as array of [id, name] pairs first
	var profileArray [][]interface{}
	if err := json.Unmarshal(raw, &profileArray); err == nil {
		profiles := make([]Profile, 0, len(profileArray))
		for _, p := range profileArray {
			if len(p) >= 2 {
				id, _ := p[0].(string)
				name, _ := p[1].(string)
				profiles = append(profiles, Profile{ID: id, Name: name})
			}
		}
		return profiles, nil
	}

	// Try to parse as object with profiles field
	var result map[string]json.RawMessage
	if err := json.Unmarshal(raw, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response as array or object: %w", err)
	}

	// Try data.profiles structure
	if data, ok := result["data"]; ok {
		var dataObj struct {
			Profiles []Profile `json:"profiles"`
		}
		if err := json.Unmarshal(data, &dataObj); err == nil {
			return dataObj.Profiles, nil
		}
	}

	// Try direct profiles field
	if profilesRaw, ok := result["profiles"]; ok {
		var profiles []Profile
		if err := json.Unmarshal(profilesRaw, &profiles); err == nil {
			return profiles, nil
		}
	}

	return nil, fmt.Errorf("could not parse profiles from response")
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
