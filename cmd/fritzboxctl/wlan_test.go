package main

import (
	"encoding/xml"
	"testing"

	"github.com/fabito/fritzboxctl/internal/soap"
	"github.com/fabito/fritzboxctl/internal/soap/services"
)

// Test WLAN statistics parsing with real-like XML response
// Now uses services.WLANStatsResponse from the services package
func TestParseWLANStats(t *testing.T) {
	// Real-like SOAP response from Fritz!Box for GetStatistics
	resp := `<?xml version="1.0"?>
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

	// Parse using services.WLANStatsResponse
	var stats services.WLANStatsResponse
	if err := xml.Unmarshal([]byte(resp), &stats); err != nil {
		t.Fatalf("Failed to parse XML: %v", err)
	}

	// Verify parsed values
	if stats.NewTotalPacketsSent != "1234567" {
		t.Errorf("Expected NewTotalPacketsSent '1234567', got '%s'", stats.NewTotalPacketsSent)
	}
	if stats.NewTotalPacketsReceived != "7654321" {
		t.Errorf("Expected NewTotalPacketsReceived '7654321', got '%s'", stats.NewTotalPacketsReceived)
	}
	if stats.NewTotalBytesSent != "1234567890" {
		t.Errorf("Expected NewTotalBytesSent '1234567890', got '%s'", stats.NewTotalBytesSent)
	}
	if stats.NewTotalBytesReceived != "9876543210" {
		t.Errorf("Expected NewTotalBytesReceived '9876543210', got '%s'", stats.NewTotalBytesReceived)
	}
}

// Test with real router-like response that has namespace prefixes
func TestParseWLANStatsWithNamespaces(t *testing.T) {
	// Simulate response with namespace prefixes
	resp := `<?xml version="1.0"?>
<s:Envelope s:encodingStyle="http://schemas.xmlsoap.org/soap/encoding/" xmlns:s="http://schemas.xmlsoap.org/soap/envelope/">
  <s:Body>
    <u:GetStatisticsResponse xmlns:u="urn:dslforum-org:service:WLANConfiguration:1">
      <NewTotalPacketsSent>100</NewTotalPacketsSent>
      <NewTotalPacketsReceived>200</NewTotalPacketsReceived>
    </u:GetStatisticsResponse>
  </s:Body>
</s:Envelope>`

	// Clean the response using shared function from soap package
	resp = soap.CleanSoapResponse(resp)

	// Parse
	var stats services.WLANStatsResponse
	if err := xml.Unmarshal([]byte(resp), &stats); err != nil {
		t.Fatalf("Failed to parse cleaned XML: %v", err)
	}

	if stats.NewTotalPacketsSent != "100" {
		t.Errorf("Expected '100', got '%s'", stats.NewTotalPacketsSent)
	}
	if stats.NewTotalPacketsReceived != "200" {
		t.Errorf("Expected '200', got '%s'", stats.NewTotalPacketsReceived)
	}
}

// Test getBandName function (still in wlan.go)
func TestGetBandName(t *testing.T) {
	tests := []struct {
		band     int
		expected string
	}{
		{1, "2.4 GHz"},
		{2, "5 GHz"},
		{3, "5 GHz (2nd channel)"},
		{4, "Guest Network"},
		{99, "Unknown (99)"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := getBandName(tt.band)
			if result != tt.expected {
				t.Errorf("getBandName(%d) = '%s', want '%s'", tt.band, result, tt.expected)
			}
		})
	}
}
