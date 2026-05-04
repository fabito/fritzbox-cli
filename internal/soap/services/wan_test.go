package services

import (
	"encoding/xml"
	"errors"
	"testing"
)

// TestGetExternalIPAddress tests the GetExternalIPAddress function
func TestGetExternalIPAddress(t *testing.T) {
	// Test with nil client (returns mock data)
	result, err := GetExternalIPAddress(nil)
	if err != nil {
		t.Fatalf("GetExternalIPAddress failed: %v", err)
	}
	if result.NewExternalIPAddress != "100.77.123.45" {
		t.Errorf("Expected '100.77.123.45', got '%s'", result.NewExternalIPAddress)
	}
}

// TestGetExternalIPAddressXMLParsing tests XML parsing of external IP response
func TestGetExternalIPAddressXMLParsing(t *testing.T) {
	// Test XML parsing directly
	mockResponse := `<?xml version="1.0"?>
<Envelope>
<Body>
<GetExternalIPAddressResponse>
<NewExternalIPAddress>100.77.123.45</NewExternalIPAddress>
</GetExternalIPAddressResponse>
</Body>
</Envelope>`

	var result GetExternalIPAddressResponse
	err := xml.Unmarshal([]byte(mockResponse), &result)
	if err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if result.NewExternalIPAddress != "100.77.123.45" {
		t.Errorf("Expected '100.77.123.45', got '%s'", result.NewExternalIPAddress)
	}

	t.Logf("Successfully got External IP: %s", result.NewExternalIPAddress)
}

// TestGetWANStatus tests the GetWANStatus function
func TestGetWANStatus(t *testing.T) {
	// Test with nil client (returns mock data)
	result, err := GetWANStatus(nil)
	if err != nil {
		t.Fatalf("GetWANStatus failed: %v", err)
	}
	if result.NewConnectionStatus != "Connected" {
		t.Errorf("Expected 'Connected', got '%s'", result.NewConnectionStatus)
	}
	if result.NewLastConnectionError != "ERROR_NONE" {
		t.Errorf("Expected 'ERROR_NONE', got '%s'", result.NewLastConnectionError)
	}
	if result.NewUptime != "12345" {
		t.Errorf("Expected '12345', got '%s'", result.NewUptime)
	}
}

// TestWANStatusResponseXMLParsing tests XML parsing of WAN status response
func TestWANStatusResponseXMLParsing(t *testing.T) {
	tests := []struct {
		name           string
		xml            string
		expectedStatus string
		expectedError  string
		expectedUptime string
	}{
		{
			name: "Connected status",
			xml: `<?xml version="1.0"?>
<Envelope>
<Body>
<GetStatusInfoResponse>
<NewConnectionStatus>Connected</NewConnectionStatus>
<NewLastConnectionError>ERROR_NONE</NewLastConnectionError>
<NewUptime>86400</NewUptime>
</GetStatusInfoResponse>
</Body>
</Envelope>`,
			expectedStatus: "Connected",
			expectedError:  "ERROR_NONE",
			expectedUptime: "86400",
		},
		{
			name: "Disconnected status",
			xml: `<?xml version="1.0"?>
<Envelope>
<Body>
<GetStatusInfoResponse>
<NewConnectionStatus>Disconnected</NewConnectionStatus>
<NewLastConnectionError>ERROR_AUTH_FAILURE</NewLastConnectionError>
<NewUptime>0</NewUptime>
</GetStatusInfoResponse>
</Body>
</Envelope>`,
			expectedStatus: "Disconnected",
			expectedError:  "ERROR_AUTH_FAILURE",
			expectedUptime: "0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result WANStatusResponse
			err := xml.Unmarshal([]byte(tt.xml), &result)
			if err != nil {
				t.Fatalf("Failed to parse response: %v", err)
			}

			if result.NewConnectionStatus != tt.expectedStatus {
				t.Errorf("Expected status '%s', got '%s'", tt.expectedStatus, result.NewConnectionStatus)
			}
			if result.NewLastConnectionError != tt.expectedError {
				t.Errorf("Expected error '%s', got '%s'", tt.expectedError, result.NewLastConnectionError)
			}
			if result.NewUptime != tt.expectedUptime {
				t.Errorf("Expected uptime '%s', got '%s'", tt.expectedUptime, result.NewUptime)
			}
		})
	}
}

// mockWANSOAPCaller is a mock SOAP caller for WAN tests
type mockWANSOAPCaller struct {
	response string
	err      error
}

func (m *mockWANSOAPCaller) Call(servicePath, action, body string) (string, error) {
	return m.response, m.err
}

// TestWANIPConnectionResponseParsing tests various IP address formats
func TestWANIPConnectionResponseParsing(t *testing.T) {
	tests := []struct {
		name       string
		ip         string
		shouldPass bool
	}{
		{"IPv4 public", "203.0.113.45", true},
		{"IPv4 private", "192.168.1.1", true},
		{"IPv4 localhost", "127.0.0.1", true},
		{"Empty IP", "", true}, // Parser should handle this
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			xmlData := `<?xml version="1.0"?>
<Envelope>
<Body>
<GetExternalIPAddressResponse>
<NewExternalIPAddress>` + tt.ip + `</NewExternalIPAddress>
</GetExternalIPAddressResponse>
</Body>
</Envelope>`

			var result GetExternalIPAddressResponse
			err := xml.Unmarshal([]byte(xmlData), &result)

			if tt.shouldPass {
				if err != nil {
					t.Fatalf("Failed to parse response: %v", err)
				}
				if result.NewExternalIPAddress != tt.ip {
					t.Errorf("Expected '%s', got '%s'", tt.ip, result.NewExternalIPAddress)
				}
			}
		})
	}
}

// TestWANStatusDNSServers tests parsing of DNS servers field
func TestWANStatusDNSServers(t *testing.T) {
	xmlData := `<?xml version="1.0"?>
<Envelope>
<Body>
<GetStatusInfoResponse>
<NewConnectionStatus>Connected</NewConnectionStatus>
<NewLastConnectionError>ERROR_NONE</NewLastConnectionError>
<NewUptime>3600</NewUptime>
<NewDNSServers>8.8.8.8,8.8.4.4</NewDNSServers>
</GetStatusInfoResponse>
</Body>
</Envelope>`

	var result WANStatusResponse
	err := xml.Unmarshal([]byte(xmlData), &result)
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	if result.NewDNSServers != "8.8.8.8,8.8.4.4" {
		t.Errorf("Expected '8.8.8.8,8.8.4.4', got '%s'", result.NewDNSServers)
	}
}

// TestWANStatusEmptyResponse tests handling of empty response fields
func TestWANStatusEmptyResponse(t *testing.T) {
	xmlData := `<?xml version="1.0"?>
<Envelope>
<Body>
<GetStatusInfoResponse>
<NewConnectionStatus></NewConnectionStatus>
<NewLastConnectionError></NewLastConnectionError>
<NewUptime></NewUptime>
</GetStatusInfoResponse>
</Body>
</Envelope>`

	var result WANStatusResponse
	err := xml.Unmarshal([]byte(xmlData), &result)
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	if result.NewConnectionStatus != "" {
		t.Errorf("Expected empty status, got '%s'", result.NewConnectionStatus)
	}
}

// TestWANServiceMockClient demonstrates testing with mock client
func TestWANServiceMockClient(t *testing.T) {
	// Note: GetWANStatus and GetExternalIPAddress use *soap.Client directly
	// To fully test with mocks, the functions would need to accept an interface
	// For now, we test the nil case and XML parsing separately

	// Nil client returns mock data
	status, err := GetWANStatus(nil)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if status.NewConnectionStatus != "Connected" {
		t.Errorf("Expected mock status 'Connected', got '%s'", status.NewConnectionStatus)
	}

	ip, err := GetExternalIPAddress(nil)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if ip.NewExternalIPAddress != "100.77.123.45" {
		t.Errorf("Expected mock IP '100.77.123.45', got '%s'", ip.NewExternalIPAddress)
	}
}

// Verify interface compatibility
var _ = errors.New // suppress unused import
