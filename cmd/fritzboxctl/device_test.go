package main

import (
	"encoding/xml"
	"strings"
	"testing"
)

// TestParseRealDeviceInfo tests with REAL Fritz!Box 7530 XML data
func TestParseRealDeviceInfo(t *testing.T) {
	// Exact XML after cleaning (from debug logs - this WORKS in simple test)
	resp := `<?xml version="1.0"?>
<Envelope>
<Body>
<GetInfoResponse >
<NewManufacturerName>AVM</NewManufacturerName>
<NewModelName>FRITZ!Box 7530</NewModelName>
<NewSerialNumber>DC396F6BC160</NewSerialNumber>
<NewSoftwareVersion>164.08.02</NewSoftwareVersion>
</GetInfoResponse>
</Body>
</Envelope>`
	
	// Clean up to match what device.go expects
	resp = strings.ReplaceAll(resp, ` encodingStyle="http://schemas.xmlsoap.org/soap/encoding/"`, ">")
	resp = strings.ReplaceAll(resp, " >", ">")

	t.Logf("Cleaned XML: %s", resp)

	// Parse using the EXACT same Envelope struct as device.go
	var info Envelope
	if err := xml.Unmarshal([]byte(resp), &info); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	t.Logf("Parsed: %+v", info.Body.GetInfoResponse)

	// Verify
	if info.Body.GetInfoResponse.NewManufacturerName != "AVM" {
		t.Errorf("Expected ManufacturerName 'AVM', got '%s'", info.Body.GetInfoResponse.NewManufacturerName)
	}
	if info.Body.GetInfoResponse.NewModelName != "FRITZ!Box 7530" {
		t.Errorf("Expected NewModelName 'FRITZ!Box 7530', got '%s'", info.Body.GetInfoResponse.NewModelName)
	}
	if info.Body.GetInfoResponse.NewSerialNumber != "DC396F6BC160" {
		t.Errorf("Expected NewSerialNumber 'DC396F6BC160', got '%s'", info.Body.GetInfoResponse.NewSerialNumber)
	}
	if info.Body.GetInfoResponse.NewSoftwareVersion != "164.08.02" {
		t.Errorf("Expected NewSoftwareVersion '164.08.02', got '%s'", info.Body.GetInfoResponse.NewSoftwareVersion)
	}
}
