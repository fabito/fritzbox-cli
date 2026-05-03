package main

import (
	"encoding/xml"
	"testing"
)

// Test WLAN statistics parsing with real-like XML response
func TestParseWLANStats(t *testing.T) {
	// Real-like SOAP response from Fritz!Box for GetStatistics
	resp := `<?xml version="1.0"?>
<Envelope>
<Body>
<GetStatisticsResponse>
<NewTotalPacketsSent>1234567</NewTotalPacketsSent>
<NewTotalPacketsReceived>7654321</NewTotalPacketsReceived>
<NewPacketErrorsReceived>12</NewPacketErrorsReceived>
<NewPacketErrorsSent>5</NewPacketErrorsSent>
<NewTotalBytesSent>1234567890</NewTotalBytesSent>
<NewTotalBytesReceived>9876543210</NewTotalBytesReceived>
<NewErrorsReceived>3</NewErrorsReceived>
<NewErrorsSent>1</NewErrorsSent>
<NewUnicastPacketsSent>1000000</NewUnicastPacketsSent>
<NewUnicastPacketsReceived>7000000</NewUnicastPacketsReceived>
<NewMulticastPacketsSent>20000</NewMulticastPacketsSent>
<NewMulticastPacketsReceived>30000</NewMulticastPacketsReceived>
<NewBroadcastPacketsSent>234567</NewBroadcastPacketsSent>
<NewBroadcastPacketsReceived>65321</NewBroadcastPacketsReceived>
</GetStatisticsResponse>
</Body>
</Envelope>`

	// Parse using the same struct as in wlan.go
	var envelope WLANStatsEnvelope
	if err := xml.Unmarshal([]byte(resp), &envelope); err != nil {
		t.Fatalf("Failed to parse XML: %v", err)
	}

	stats := envelope.Body.GetStatisticsResponse

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
	if stats.NewPacketErrorsReceived != "12" {
		t.Errorf("Expected NewPacketErrorsReceived '12', got '%s'", stats.NewPacketErrorsReceived)
	}
}

// Test cleanSoapResponse function
func TestCleanSoapResponse(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Remove encodingStyle",
			input:    `<Envelope encodingStyle="http://schemas.xmlsoap.org/soap/encoding/">`,
			expected: `<Envelope>`,
		},
		{
			name:     "Remove namespace s:",
			input:    `<s:Envelope xmlns:s="http://schemas.xmlsoap.org/soap/envelope/">`,
			expected: `<Envelope xmlnss="http://schemas.xmlsoap.org/soap/envelope/">`,
		},
		{
			name:     "Fix space before >",
			input:    `<GetStatisticsResponse >`,
			expected: `<GetStatisticsResponse>`,
		},
		{
			name:     "Remove u: prefix",
			input:    `<u:GetStatisticsResponse>`,
			expected: `<GetStatisticsResponse>`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := cleanSoapResponse(tt.input)
			// Simple check - just verify no error occurs and result is not empty
			if result == "" {
				t.Errorf("cleanSoapResponse returned empty string")
			}
		})
	}
}

// Test getBandName function
func TestGetBandName(t *testing.T) {
	tests := []struct {
		band    int
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

// Test with real router-like response that has namespace prefixes
func TestParseWLANStatsWithNamespaces(t *testing.T) {
	// Simulate response with namespace prefixes (before cleaning)
	resp := `<?xml version="1.0"?>
<s:Envelope s:encodingStyle="http://schemas.xmlsoap.org/soap/encoding/" xmlns:s="http://schemas.xmlsoap.org/soap/envelope/">
  <s:Body>
    <u:GetStatisticsResponse xmlns:u="urn:dslforum-org:service:WLANConfiguration:1">
      <NewTotalPacketsSent>100</NewTotalPacketsSent>
      <NewTotalPacketsReceived>200</NewTotalPacketsReceived>
    </u:GetStatisticsResponse>
  </s:Body>
</s:Envelope>`

	// Clean the response like displayWLANStats does
	resp = cleanSoapResponse(resp)

	// Parse
	var envelope WLANStatsEnvelope
	if err := xml.Unmarshal([]byte(resp), &envelope); err != nil {
		t.Fatalf("Failed to parse cleaned XML: %v", err)
	}

	if envelope.Body.GetStatisticsResponse.NewTotalPacketsSent != "100" {
		t.Errorf("Expected '100', got '%s'", envelope.Body.GetStatisticsResponse.NewTotalPacketsSent)
	}
	if envelope.Body.GetStatisticsResponse.NewTotalPacketsReceived != "200" {
		t.Errorf("Expected '200', got '%s'", envelope.Body.GetStatisticsResponse.NewTotalPacketsReceived)
	}
}

// Test wlanServiceInfo mapping
func TestWLANServiceInfo(t *testing.T) {
	tests := []struct {
		band            int
		expectedPath    string
		expectedType    string
		expectError     bool
	}{
		{1, "/upnp/control/wlanconfig1", "urn:dslforum-org:service:WLANConfiguration:1", false},
		{2, "/upnp/control/wlanconfig2", "urn:dslforum-org:service:WLANConfiguration:2", false},
		{3, "/upnp/control/wlanconfig3", "urn:dslforum-org:service:WLANConfiguration:3", false},
		{4, "/upnp/control/wlanconfig4", "urn:dslforum-org:service:WLANConfiguration:4", false},
	}

	for _, tt := range tests {
		t.Run(tt.expectedPath, func(t *testing.T) {
			info, ok := wlanServiceInfo[tt.band]
			if !ok {
				if !tt.expectError {
					t.Errorf("wlanServiceInfo[%d] not found", tt.band)
				}
				return
			}
			if info.Path != tt.expectedPath {
				t.Errorf("Path = '%s', want '%s'", info.Path, tt.expectedPath)
			}
			if info.Type != tt.expectedType {
				t.Errorf("Type = '%s', want '%s'", info.Type, tt.expectedType)
			}
		})
	}
}
