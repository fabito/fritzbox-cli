package aha

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestListDevicesWithProfiles is the RED phase - test should FAIL initially
func TestListDevicesWithProfiles(t *testing.T) {
	// Mock JSON response from page=netDev&xhrId=all
	mockResponse := `{
		"data": {
			"active": [
				{
					"name": "Fabios-MBP",
					"mac": "3C:8D:20:E8:3C:24",
					"ipv4": {"ip": "192.168.178.29"},
					"UID": "landevice3019"
				},
				{
					"name": "xavier1",
					"mac": "48:B0:2D:51:0B:E6",
					"ipv4": {"ip": "192.168.178.133"},
					"UID": "landevice5980"
				}
			]
		}
	}`

	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(mockResponse))
	}))
	defer server.Close()

	// Create client with test server URL
	client := &Client{
		baseURL: server.URL,
		httpClient: server.Client(),
		getSID: func() (string, error) {
			return "test-sid", nil
		},
	}

	// Call function (should fail - function doesn't exist yet)
	devices, err := client.ListDevicesWithProfiles()
	if err != nil {
		t.Fatalf("ListDevicesWithProfiles() returned error: %v", err)
	}

	// Verify results
	if len(devices) != 2 {
		t.Errorf("Expected 2 devices, got %d", len(devices))
	}

	// Check first device
	if devices[0].DeviceName != "Fabios-MBP" {
		t.Errorf("Expected device name 'Fabios-MBP', got '%s'", devices[0].DeviceName)
	}
	if devices[0].MACAddress != "3C:8D:20:E8:3C:24" {
		t.Errorf("Expected MAC '3C:8D:20:E8:3C:24', got '%s'", devices[0].MACAddress)
	}
	if devices[0].IPAddress != "192.168.178.29" {
		t.Errorf("Expected IP '192.168.178.29', got '%s'", devices[0].IPAddress)
	}

	// Check second device
	if devices[1].DeviceName != "xavier1" {
		t.Errorf("Expected device name 'xavier1', got '%s'", devices[1].DeviceName)
	}

	t.Logf("Successfully parsed devices: %+v", devices)
}

// TestListDevicesWithProfilesEmpty tests with empty device list
func TestListDevicesWithProfilesEmpty(t *testing.T) {
	mockResponse := `{"data": {"active": []}}`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(mockResponse))
	}))
	defer server.Close()

	client := &Client{
		baseURL: server.URL,
		httpClient: server.Client(),
		getSID: func() (string, error) {
			return "test-sid", nil
		},
	}

	devices, err := client.ListDevicesWithProfiles()
	if err != nil {
		t.Fatalf("ListDevicesWithProfiles() returned error: %v", err)
	}

	if len(devices) != 0 {
		t.Errorf("Expected 0 devices, got %d", len(devices))
	}
}

// TestListDevicesWithProfilesJSON tests JSON parsing logic
func TestListDevicesWithProfilesJSON(t *testing.T) {
	// Test the JSON structure parsing
	jsonData := `{
		"data": {
			"active": [
				{
					"name": "TestDevice",
					"mac": "AA:BB:CC:DD:EE:FF",
					"ipv4": {"ip": "192.168.178.100"},
					"UID": "landevice1234"
				}
			]
		}
	}`

	var result map[string]json.RawMessage
	if err := json.Unmarshal([]byte(jsonData), &result); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	data, ok := result["data"]
	if !ok {
		t.Fatal("No data field in response")
	}

	var dataObj struct {
		Active []map[string]interface{} `json:"active"`
	}
	if err := json.Unmarshal(data, &dataObj); err != nil {
		t.Fatalf("Failed to parse data: %v", err)
	}

	if len(dataObj.Active) != 1 {
		t.Errorf("Expected 1 device, got %d", len(dataObj.Active))
	}

	// Extract fields
	name, _ := dataObj.Active[0]["name"].(string)
	mac, _ := dataObj.Active[0]["mac"].(string)

	if name != "TestDevice" {
		t.Errorf("Expected 'TestDevice', got '%s'", name)
	}
	if mac != "AA:BB:CC:DD:EE:FF" {
		t.Errorf("Expected 'AA:BB:CC:DD:EE:FF', got '%s'", mac)
	}

	t.Log("JSON parsing logic works!")
}
