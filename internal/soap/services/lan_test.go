package services

import (
	"encoding/xml"
	"testing"
)

// TestGetLANStats tests the GetLANStats function (RED phase - should fail to compile)
func TestGetLANStats(t *testing.T) {
	// Test with nil client (should return mock data)
	resp, err := GetLANStats(nil)
	if err != nil {
		t.Fatalf("GetLANStats failed: %v", err)
	}
	if resp.NewBytesSent != "123456789" {
		t.Errorf("Expected '123456789', got '%s'", resp.NewBytesSent)
	}
	if resp.NewBytesReceived != "987654321" {
		t.Errorf("Expected '987654321', got '%s'", resp.NewBytesReceived)
	}
	if resp.NewPacketsSent != "12345" {
		t.Errorf("Expected '12345', got '%s'", resp.NewPacketsSent)
	}
	if resp.NewPacketsReceived != "54321" {
		t.Errorf("Expected '54321', got '%s'", resp.NewPacketsReceived)
	}
}

// TestParseLANStatsResponse tests XML parsing of LAN stats response
func TestParseLANStatsResponse(t *testing.T) {
	// Real XML response from Fritz!Box (after cleanSoapResponse)
	xmlData := `<?xml version="1.0"?>
<Envelope>
<Body>
<GetStatisticsResponse>
<NewBytesSent>123456789</NewBytesSent>
<NewBytesReceived>987654321</NewBytesReceived>
<NewPacketsSent>12345</NewPacketsSent>
<NewPacketsReceived>54321</NewPacketsReceived>
<NewErrorsSent>10</NewErrorsSent>
<NewErrorsReceived>5</NewErrorsReceived>
</GetStatisticsResponse>
</Body>
</Envelope>`

	var resp LANStatsResponse
	if err := xml.Unmarshal([]byte(xmlData), &resp); err != nil {
		t.Fatalf("Failed to parse XML: %v", err)
	}

	// Verify parsed values
	if resp.NewBytesSent != "123456789" {
		t.Errorf("Expected '123456789', got '%s'", resp.NewBytesSent)
	}
	if resp.NewBytesReceived != "987654321" {
		t.Errorf("Expected '987654321', got '%s'", resp.NewBytesReceived)
	}
	if resp.NewPacketsSent != "12345" {
		t.Errorf("Expected '12345', got '%s'", resp.NewPacketsSent)
	}
	if resp.NewPacketsReceived != "54321" {
		t.Errorf("Expected '54321', got '%s'", resp.NewPacketsReceived)
	}
	if resp.NewErrorsSent != "10" {
		t.Errorf("Expected '10', got '%s'", resp.NewErrorsSent)
	}
	if resp.NewErrorsReceived != "5" {
		t.Errorf("Expected '5', got '%s'", resp.NewErrorsReceived)
	}

	t.Logf("Successfully parsed LAN stats: %+v", resp)
}
