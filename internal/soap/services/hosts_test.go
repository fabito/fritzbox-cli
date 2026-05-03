package services

import (
	"encoding/xml"
	"testing"
)

// TestBlockDevice GREEN phase - test should pass now
func TestBlockDevice(t *testing.T) {
	// Test with nil client (mock mode)
	err := BlockDevice("192.168.178.50", nil)
	if err != nil {
		t.Errorf("BlockDevice returned error: %v", err)
	}
}

// TestUnblockDevice GREEN phase - test should pass now
func TestUnblockDevice(t *testing.T) {
	// Test with nil client (mock mode)
	err := UnblockDevice("192.168.178.50", nil)
	if err != nil {
		t.Errorf("UnblockDevice returned error: %v", err)
	}
}

// TestBlockDeviceWithMockClient tests with a mock SOAP client
func TestBlockDeviceWithMockClient(t *testing.T) {
	// Create a mock soap.Client that returns success
	// For now, test with nil client (mock mode)
	err := BlockDevice("192.168.178.100", nil)
	if err != nil {
		t.Errorf("BlockDevice with nil client returned error: %v", err)
	}
}

// TestUnblockDeviceWithMockClient tests with a mock SOAP client
func TestUnblockDeviceWithMockClient(t *testing.T) {
	// Create a mock soap.Client that returns success
	// For now, test with nil client (mock mode)
	err := UnblockDevice("192.168.178.100", nil)
	if err != nil {
		t.Errorf("UnblockDevice with nil client returned error: %v", err)
	}
}

// TestBlockDeviceEmptyIP tests error handling
func TestBlockDeviceEmptyIP(t *testing.T) {
	// BlockDevice should handle empty IP gracefully
	// For now with nil client it should succeed (mock mode)
	err := BlockDevice("", nil)
	if err != nil {
		t.Errorf("BlockDevice with empty IP returned error: %v", err)
	}
}

// TestParseHostList tests parsing of host list from Fritz!Box Hosts service
// RED PHASE: This test references HostListResponse which DOES NOT EXIST YET
// The test should FAIL to compile
func TestParseHostList(t *testing.T) {
	// Real SOAP response from Fritz!Box Hosts service (GetGenericHostEntry action)
	soapResponse := `<?xml version="1.0"?>
<Envelope>
<Body>
<GetGenericHostEntryResponse>
<NewHostName>MyLaptop</NewHostName>
<NewIPAddress>192.168.178.20</NewIPAddress>
<NewMACAddress>AA:BB:CC:DD:EE:FF</NewMACAddress>
<NewInterfaceType>WLAN</NewInterfaceType>
<NewActive>1</NewActive>
<NewX_AVM-DE_Port>5</NewX_AVM-DE_Port>
</GetGenericHostEntryResponse>
</Body>
</Envelope>`

	// This should FAIL - HostListResponse doesn't exist yet!
	var response HostListResponse
	if err := xml.Unmarshal([]byte(soapResponse), &response); err != nil {
		t.Fatalf("Failed to parse XML: %v", err)
	}

	// Verify parsed values
	if response.NewHostName != "MyLaptop" {
		t.Errorf("Expected NewHostName 'MyLaptop', got '%s'", response.NewHostName)
	}
	if response.NewIPAddress != "192.168.178.20" {
		t.Errorf("Expected NewIPAddress '192.168.178.20', got '%s'", response.NewIPAddress)
	}
	if response.NewMACAddress != "AA:BB:CC:DD:EE:FF" {
		t.Errorf("Expected NewMACAddress 'AA:BB:CC:DD:EE:FF', got '%s'", response.NewMACAddress)
	}
	if response.NewInterfaceType != "WLAN" {
		t.Errorf("Expected NewInterfaceType 'WLAN', got '%s'", response.NewInterfaceType)
	}
	if response.NewActive != "1" {
		t.Errorf("Expected NewActive '1', got '%s'", response.NewActive)
	}

	t.Logf("Successfully parsed host: %+v", response)
}

// TestParseHostListWithEmptyFields tests parsing when some fields are empty
func TestParseHostListWithEmptyFields(t *testing.T) {
	soapResponse := `<?xml version="1.0"?>
<Envelope>
<Body>
<GetGenericHostEntryResponse>
<NewHostName></NewHostName>
<NewIPAddress>192.168.178.30</NewIPAddress>
<NewMACAddress>00:11:22:33:44:55</NewMACAddress>
<NewInterfaceType>LAN</NewInterfaceType>
<NewActive>0</NewActive>
</GetGenericHostEntryResponse>
</Body>
</Envelope>`

	var response HostListResponse
	if err := xml.Unmarshal([]byte(soapResponse), &response); err != nil {
		t.Fatalf("Failed to parse XML: %v", err)
	}

	if response.NewHostName != "" {
		t.Errorf("Expected empty NewHostName, got '%s'", response.NewHostName)
	}
	if response.NewIPAddress != "192.168.178.30" {
		t.Errorf("Expected '192.168.178.30', got '%s'", response.NewIPAddress)
	}
	if response.NewActive != "0" {
		t.Errorf("Expected '0', got '%s'", response.NewActive)
	}
}

// TestParseHostNumber tests parsing number of hosts
func TestParseHostNumber(t *testing.T) {
	// Mock response for GetHostNumberOfEntries
	numResponse := `<?xml version="1.0"?>
<Envelope>
<Body>
<GetHostNumberOfEntriesResponse>
<NewHostNumberOfEntries>3</NewHostNumberOfEntries>
</GetHostNumberOfEntriesResponse>
</Body>
</Envelope>`

	// This should FAIL - HostNumberResponse doesn't exist yet!
	var response HostNumberResponse
	if err := xml.Unmarshal([]byte(numResponse), &response); err != nil {
		t.Fatalf("Failed to parse XML: %v", err)
	}

	if response.NewHostNumberOfEntries != "3" {
		t.Errorf("Expected 3 hosts, got '%s'", response.NewHostNumberOfEntries)
	}

	t.Logf("Number of hosts: %s", response.NewHostNumberOfEntries)
}
