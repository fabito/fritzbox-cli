package services

import (
	"encoding/xml"
	"testing"
)

// TestGetWLANStats tests the GetWLANStats function
func TestGetWLANStats(t *testing.T) {
	// Test with nil client (returns mock data)
	resp, err := GetWLANStats(nil, 1)
	if err != nil {
		t.Fatalf("GetWLANStats failed: %v", err)
	}
	if resp.NewTotalPacketsSent != "12345" {
		t.Errorf("Expected '12345', got '%s'", resp.NewTotalPacketsSent)
	}
}

// TestParseWLANStatsResponse tests XML parsing of WLAN stats response
func TestParseWLANStatsResponse(t *testing.T) {
	// Real XML response from Fritz!Box (after cleanSoapResponse)
	xmlData := `<?xml version="1.0"?>
<Envelope>
<Body>
<GetStatisticsResponse>
<NewTotalPacketsSent>1234567</NewTotalPacketsSent>
<NewTotalPacketsReceived>7654321</NewTotalPacketsReceived>
<NewTotalBytesSent>1234567890</NewTotalBytesSent>
<NewTotalBytesReceived>9876543210</NewTotalBytesReceived>
</GetStatisticsResponse>
</Body>
</Envelope>`

	var resp WLANStatsResponse
	if err := xml.Unmarshal([]byte(xmlData), &resp); err != nil {
		t.Fatalf("Failed to parse XML: %v", err)
	}

	// Verify parsed values
	if resp.NewTotalPacketsSent != "1234567" {
		t.Errorf("Expected '1234567', got '%s'", resp.NewTotalPacketsSent)
	}
	if resp.NewTotalPacketsReceived != "7654321" {
		t.Errorf("Expected '7654321', got '%s'", resp.NewTotalPacketsReceived)
	}
	if resp.NewTotalBytesSent != "1234567890" {
		t.Errorf("Expected '1234567890', got '%s'", resp.NewTotalBytesSent)
	}
	if resp.NewTotalBytesReceived != "9876543210" {
		t.Errorf("Expected '9876543210', got '%s'", resp.NewTotalBytesReceived)
	}

	t.Logf("Successfully parsed: %+v", resp)
}
