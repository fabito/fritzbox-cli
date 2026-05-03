package services

import (
	"encoding/xml"
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

	// Test XML parsing directly
	mockResponse := `<?xml version="1.0"?>
<Envelope>
<Body>
<GetExternalIPAddressResponse>
<NewExternalIPAddress>100.77.123.45</NewExternalIPAddress>
</GetExternalIPAddressResponse>
</Body>
</Envelope>`

	resp := xml.Unmarshal([]byte(mockResponse), &result)
	if resp != nil {
		t.Fatalf("Failed to parse response: %v", resp)
	}

	if result.NewExternalIPAddress != "100.77.123.45" {
		t.Errorf("Expected '100.77.123.45', got '%s'", result.NewExternalIPAddress)
	}

	t.Logf("Successfully got External IP: %s", result.NewExternalIPAddress)
}
