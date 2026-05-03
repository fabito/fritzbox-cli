package services

import (
	"encoding/xml"
	"strings"
	"testing"
)

// TestGetWLANStatus tests the GetWLANStatus function (RED phase - test first)
func TestGetWLANStatus(t *testing.T) {
	// Test with nil client (returns mock data)
	resp, err := GetWLANStatus(nil, 1)
	if err != nil {
		t.Fatalf("GetWLANStatus failed: %v", err)
	}
	if resp.NewEnable != "1" {
		t.Errorf("Expected '1', got '%s'", resp.NewEnable)
	}
	if resp.NewSSID != "TestSSID" {
		t.Errorf("Expected 'TestSSID', got '%s'", resp.NewSSID)
	}

	// Test with invalid band
	resp, err = GetWLANStatus(nil, 99)
	if err == nil {
		t.Error("Expected error for invalid band")
	}
}

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

// TestParseWLANStatusResponse tests XML parsing of WLAN status response
func TestParseWLANStatusResponse(t *testing.T) {
	// Real XML response from Fritz!Box WLANConfiguration:GetInfo (after cleanSoapResponse)
	xmlData := `<?xml version="1.0"?>
<Envelope>
<Body>
<GetInfoResponse>
<NewEnable>1</NewEnable>
<NewSSID>MyWLAN</NewSSID>
<NewBeaconType>WPA2</NewBeaconType>
<NewChannel>6</NewChannel>
<NewMaxBitRate>866</NewMaxBitRate>
</GetInfoResponse>
</Body>
</Envelope>`

	var resp WLANStatusResponse
	if err := xml.Unmarshal([]byte(xmlData), &resp); err != nil {
		t.Fatalf("Failed to parse XML: %v", err)
	}

	// Verify parsed values
	if resp.NewEnable != "1" {
		t.Errorf("Expected '1', got '%s'", resp.NewEnable)
	}
	if resp.NewSSID != "MyWLAN" {
		t.Errorf("Expected 'MyWLAN', got '%s'", resp.NewSSID)
	}
	if resp.NewBeaconType != "WPA2" {
		t.Errorf("Expected 'WPA2', got '%s'", resp.NewBeaconType)
	}
	if resp.NewChannel != "6" {
		t.Errorf("Expected '6', got '%s'", resp.NewChannel)
	}

	t.Logf("Successfully parsed: %+v", resp)
}

// TestSetWLANEnabled tests the SetWLANEnabled function (RED phase - test first)
func TestSetWLANEnabled(t *testing.T) {
	// Test with nil client (returns mock success for enable)
	err := SetWLANEnabled(1, true, nil)
	if err != nil {
		t.Errorf("Expected no error with nil client (enable), got: %v", err)
	}

	// Test with nil client (disable)
	err = SetWLANEnabled(1, false, nil)
	if err != nil {
		t.Errorf("Expected no error with nil client (disable), got: %v", err)
	}

	// Test with all valid bands
	for _, band := range []int{1, 2, 3, 4} {
		err = SetWLANEnabled(band, true, nil)
		if err != nil {
			t.Errorf("Expected no error for band %d, got: %v", band, err)
		}
	}

	// Test with invalid band
	err = SetWLANEnabled(99, true, nil)
	if err == nil {
		t.Error("Expected error for invalid band")
	}

	// Verify error message contains band number
	if err != nil && !strings.Contains(err.Error(), "99") {
		t.Errorf("Expected error message to contain '99', got: %v", err)
	}
}

// TestSetWLANEnabledEnabledValue tests that enabled value is correctly converted

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
