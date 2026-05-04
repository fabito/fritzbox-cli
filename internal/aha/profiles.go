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
	DeviceID    string `json:"device_id"`
	DeviceName  string `json:"name"`
	MACAddress  string `json:"mac"`
	IPAddress   string `json:"ip"`
	ProfileID   string `json:"profile_id"`
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
		"xhr":   {"1"},
		"sid":   {sid},
		"page":  {"netDev"},
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
			Name string `json:"name"`
			MAC  string `json:"mac"`
			UID  string `json:"UID"`
			IPv4 struct {
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
			DeviceID:    d.UID,
			DeviceName:  d.Name,
			MACAddress:  d.MAC,
			IPAddress:   d.IPv4.IP,
			ProfileID:   d.ProfileID,
			ProfileName: d.ProfileName,
		}
		slog.Debug("ListDevicesWithProfiles: device", "name", d.Name, "mac", d.MAC, "ip", d.IPv4.IP)
		devices = append(devices, dp)
	}

	return devices, nil
}

// ListAvailableProfiles lists all available profile definitions (Standard, Restricted, etc.)
// Parses HTML response from page=kidPro&xhrId=all
func (c *Client) ListAvailableProfiles() ([]Profile, error) {
	// Get SID
	sid, err := c.getSID()
	if err != nil {
		return nil, fmt.Errorf("failed to get SID: %w", err)
	}

	// Build URL and POST data
	profileURL := fmt.Sprintf("%s/data.lua", c.baseURL)
	formData := url.Values{
		"xhr":   {"1"},
		"sid":   {sid},
		"page":  {"kidPro"},
		"xhrId": {"all"},
	}

	slog.Debug("ListAvailableProfiles: requesting", "url", profileURL)

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

	// Read response body (HTML)
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	slog.Debug("ListAvailableProfiles: response received", "body_length", len(body))

	// Parse HTML to extract profiles from input tags with id="filtprofXXX"
	// Format: <input type="radio" ... id="filtprof1" title="Standard" ...>
	html := string(body)
	profiles := []Profile{}

	// Find all filtprof patterns
	// Use simple string search to find id="filtprofXXX"
	for i := 0; i < len(html); {
		idx := strings.Index(html[i:], "filtprof")
		if idx == -1 {
			break
		}
		idx += i // Adjust to absolute position

		// Extract the full id="filtprofXXX"
		idStart := idx
		idEnd := idx + len("filtprof")
		for idEnd < len(html) && html[idEnd] >= '0' && html[idEnd] <= '9' {
			idEnd++
		}

		if idEnd > idStart {
			profileID := html[idStart:idEnd] // e.g., "filtprof1"
			numericID := strings.TrimPrefix(profileID, "filtprof")

			// Look for title="..." before this id
			titleStart := strings.LastIndex(html[:idStart], "title=\"")
			if titleStart != -1 {
				titleStart += len("title=\"")
				titleEnd := strings.Index(html[titleStart:], "\"")
				if titleEnd != -1 {
					profileName := html[titleStart : titleStart+titleEnd]

					profiles = append(profiles, Profile{
						ID:   numericID,
						Name: profileName,
					})
				}
			}
		}

		i = idEnd // Move past this match
	}

	// Remove duplicates (same profile ID might appear multiple times in HTML)
	seen := map[string]bool{}
	uniqueProfiles := []Profile{}
	for _, p := range profiles {
		if !seen[p.ID] {
			seen[p.ID] = true
			uniqueProfiles = append(uniqueProfiles, p)
		}
	}

	return uniqueProfiles, nil
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
		"dev_name":     {deviceName},
		"dev":          {deviceID},
		"kisi_profile": {filtprofID},
		"page":         {"edit_device"},
		"apply":        {"true"},
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

// parseProfilesFromHTML extracts profile IDs and names from HTML
// Format from Fritz!Box:
//
//	<td class="name" title="Standard" data-label="Standard"><span>Standard</span></td>
//	<button type="submit" name="edit" value="filtprof1" class="icon edit" title="Edit"></button>
func parseProfilesFromHTML(html string) []Profile {
	lines := strings.Split(html, "\n")

	// Extract profile names from title="NAME"
	names := []string{}
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.Contains(line, `title="`) {
			temp := line
			for {
				idx := strings.Index(temp, `title="`)
				if idx == -1 {
					break
				}
				temp = temp[idx+len(`title="`):]
				end := strings.Index(temp, `"`)
				if end == -1 {
					break
				}
				name := temp[:end]
				// Check if this is a profile name
				if name == "Standard" || name == "Guest" || name == "Unrestricted" ||
					name == "Restricted" || name == "Kids" || name == "Leo" {
					names = append(names, name)
				}
			}
		}
	}

	// Extract profile IDs from value="filtprofXXXX"
	ids := []string{}
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.Contains(line, `value="filtprof`) {
			idx := strings.Index(line, `"filtprof`)
			if idx != -1 {
				start := idx + 1 // skip opening quote
				end := strings.Index(line[start:], `"`)
				if end != -1 {
					fullID := line[start : start+end]
					numericID := strings.TrimPrefix(fullID, "filtprof")
					ids = append(ids, numericID)
				}
			}
		}
	}

	// Pair them
	profiles := []Profile{}
	for i := 0; i < len(names) && i < len(ids); i++ {
		profiles = append(profiles, Profile{ID: ids[i], Name: names[i]})
	}

	return profiles
}

// ListAvailableProfiles lists all available profile definitions
