package main

import (
	"encoding/xml"
	"testing"

	"github.com/fabito/fritzboxctl/internal/soap/services"
)

// TestParseRealDeviceInfo tests with REAL Fritz!Box 7530 XML data
// Now uses services.GetInfoResponse from the services package
func TestParseRealDeviceInfo(t *testing.T) {
	// Exact XML after cleaning (from debug logs)
	resp := `<?xml version="1.0"?>
<Envelope>
<Body>
<GetInfoResponse>
<NewManufacturerName>AVM</NewManufacturerName>
<NewModelName>FRITZ!Box 7530</NewModelName>
<NewSerialNumber>DC396F6BC160</NewSerialNumber>
<NewSoftwareVersion>164.08.02</NewSoftwareVersion>
<NewHardwareVersion>FRITZ!Box 7530</NewHardwareVersion>
</GetInfoResponse>
</Body>
</Envelope>`

	// Parse using services.GetInfoResponse (from services package)
	var info services.GetInfoResponse
	if err := xml.Unmarshal([]byte(resp), &info); err != nil {
		t.Fatalf("Failed to parse XML: %v", err)
	}

	t.Logf("Parsed: %+v", info)

	// Verify
	if info.NewManufacturerName != "AVM" {
		t.Errorf("Expected ManufacturerName 'AVM', got '%s'", info.NewManufacturerName)
	}
	if info.NewModelName != "FRITZ!Box 7530" {
		t.Errorf("Expected NewModelName 'FRITZ!Box 7530', got '%s'", info.NewModelName)
	}
	if info.NewSerialNumber != "DC396F6BC160" {
		t.Errorf("Expected NewSerialNumber 'DC396F6BC160', got '%s'", info.NewSerialNumber)
	}
	if info.NewSoftwareVersion != "164.08.02" {
		t.Errorf("Expected NewSoftwareVersion '164.08.02', got '%s'", info.NewSoftwareVersion)
	}
	if info.NewHardwareVersion != "FRITZ!Box 7530" {
		t.Errorf("Expected NewHardwareVersion 'FRITZ!Box 7530', got '%s'", info.NewHardwareVersion)
	}
}
