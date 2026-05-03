package services

import (
	"encoding/xml"
	"testing"

	"github.com/fabito/fritzboxctl/internal/soap"
)

// TestGetWANStatus tests the GetWANStatus function with mock SOAP response
func TestGetWANStatus(t *testing.T) {
	// RED phase: This test will fail because GetWANStatus doesn't exist yet

	// Mock SOAP response for WAN status (from real Fritz!Box)
	mockResponse := `<?xml version="1.0"?>
<Envelope>
<Body>
<GetStatusInfoResponse>
<NewConnectionStatus>Connected</NewConnectionStatus>
<NewExternalIPAddress>192.168.1.100</NewExternalIPAddress>
<NewLastConnectionError>ERROR_NONE</NewLastConnectionError>
<NewUptime>12345</NewUptime>
<NewDNSServers>100.95.255.100,100.95.255.101</NewDNSServers>
</GetStatusInfoResponse>
</Body>
</Envelope>`

	// Clean and parse the response
	resp := soap.CleanSoapResponse(mockResponse)
	var status WANStatusResponse
	if err := xml.Unmarshal([]byte(resp), &status); err != nil {
		t.Fatalf("Failed to parse WAN status response: %v", err)
	}

	// Verify parsed values
	if status.NewConnectionStatus != "Connected" {
		t.Errorf("Expected NewConnectionStatus 'Connected', got '%s'", status.NewConnectionStatus)
	}
	if status.NewExternalIPAddress != "192.168.1.100" {
		t.Errorf("Expected NewExternalIPAddress '192.168.1.100', got '%s'", status.NewExternalIPAddress)
	}
	if status.NewLastConnectionError != "ERROR_NONE" {
		t.Errorf("Expected NewLastConnectionError 'ERROR_NONE', got '%s'", status.NewLastConnectionError)
	}
	if status.NewUptime != "12345" {
		t.Errorf("Expected NewUptime '12345', got '%s'", status.NewUptime)
	}
	if status.NewDNSServers != "100.95.255.100,100.95.255.101" {
		t.Errorf("Expected NewDNSServers '100.95.255.100,100.95.255.101', got '%s'", status.NewDNSServers)
	}

	t.Logf("Successfully parsed WAN status: %+v", status)
}

// TestGetWANStatusDisconnected tests parsing of disconnected state
func TestGetWANStatusDisconnected(t *testing.T) {
	mockResponse := `<?xml version="1.0"?>
<Envelope>
<Body>
<GetStatusInfoResponse>
<NewConnectionStatus>Disconnected</NewConnectionStatus>
<NewExternalIPAddress>0.0.0.0</NewExternalIPAddress>
<NewLastConnectionError>ERROR_NO_CARRIER</NewLastConnectionError>
<NewUptime>0</NewUptime>
</GetStatusInfoResponse>
</Body>
</Envelope>`

	resp := soap.CleanSoapResponse(mockResponse)
	var status WANStatusResponse
	if err := xml.Unmarshal([]byte(resp), &status); err != nil {
		t.Fatalf("Failed to parse WAN status response: %v", err)
	}

	if status.NewConnectionStatus != "Disconnected" {
		t.Errorf("Expected NewConnectionStatus 'Disconnected', got '%s'", status.NewConnectionStatus)
	}
	if status.NewLastConnectionError != "ERROR_NO_CARRIER" {
		t.Errorf("Expected NewLastConnectionError 'ERROR_NO_CARRIER', got '%s'", status.NewLastConnectionError)
	}
}

// TestGetWANStatusWithNilClient tests the mock data return when client is nil
func TestGetWANStatusWithNilClient(t *testing.T) {
	resp, err := GetWANStatus(nil)
	if err != nil {
		t.Fatalf("GetWANStatus with nil client failed: %v", err)
	}
	if resp.NewConnectionStatus != "Connected" {
		t.Errorf("Expected 'Connected', got '%s'", resp.NewConnectionStatus)
	}
	if resp.NewExternalIPAddress != "192.168.1.100" {
		t.Errorf("Expected '192.168.1.100', got '%s'", resp.NewExternalIPAddress)
	}
}
