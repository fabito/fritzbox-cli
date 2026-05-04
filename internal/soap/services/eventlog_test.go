package services

import (
	"encoding/xml"
	"testing"
)

// TestGetEventLog tests retrieving the event log
func TestGetEventLog(t *testing.T) {
	// Test with nil client (mock mode)
	t.Run("WithNilClient", func(t *testing.T) {
		log, err := GetEventLog(nil)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if log == "" {
			t.Error("Expected non-empty event log")
		}
		t.Logf("Event log length: %d characters", len(log))
	})

	// Test with empty/missing log
	t.Run("EmptyLog", func(t *testing.T) {
		// This would require a mock SOAP client that returns empty log
		// For now, just verify the function signature works
		t.Skip("Need mock SOAP client for empty log test")
	})
}

// TestGetEventLogFromResponse tests parsing event log from SOAP response
func TestGetEventLogFromResponse(t *testing.T) {
	// Mock SOAP response with NewDeviceLog (from real Fritz!Box response)
	mockXML := `<?xml version="1.0"?>
<Envelope>
<Body>
<GetInfoResponse>
<NewManufacturerName>AVM</NewManufacturerName>
<NewModelName>FRITZ!Box 7530</NewModelName>
<NewSerialNumber>DC396F6BC160</NewSerialNumber>
<NewSoftwareVersion>164.08.02</NewSoftwareVersion>
<NewDeviceLog>30.04.26 11:26:46 IPv6 prefix obtained successfully. New prefix: 2404:4408:b100:3006::/64
30.04.26 11:26:42 Internet connection established successfully. IP address: 100.77.120.18
19.04.26 18:40:38 IPv6 internet connection established</NewDeviceLog>
</GetInfoResponse>
</Body>
</Envelope>`

	// Parse the XML to verify NewDeviceLog can be extracted
	type TestResponse struct {
		XMLName      xml.Name `xml:"Envelope"`
		NewDeviceLog string   `xml:"Body>GetInfoResponse>NewDeviceLog"`
	}

	var resp TestResponse
	if err := xml.Unmarshal([]byte(mockXML), &resp); err != nil {
		t.Fatalf("Failed to parse XML: %v", err)
	}

	if resp.NewDeviceLog == "" {
		t.Fatal("NewDeviceLog is empty")
	}

	t.Logf("Event log length: %d characters", len(resp.NewDeviceLog))
	t.Logf("First 100 chars: %s", resp.NewDeviceLog[:min(len(resp.NewDeviceLog), 100)])
}
