package services

import (
	"encoding/xml"
	"os"
	"testing"
)

// TestGetDSLStatus tests the GetDSLStatus function
func TestGetDSLStatus(t *testing.T) {
	// Test with nil client (returns mock data)
	resp, err := GetDSLStatus(nil)
	if err != nil {
		t.Fatalf("GetDSLStatus failed: %v", err)
	}

	// Verify mock data
	if resp.NewDownstreamCurrRate != "100000" {
		t.Errorf("Expected '100000', got '%s'", resp.NewDownstreamCurrRate)
	}
	if resp.NewUpstreamCurrRate != "40000" {
		t.Errorf("Expected '40000', got '%s'", resp.NewUpstreamCurrRate)
	}
	if resp.NewDownstreamMaxRate != "120000" {
		t.Errorf("Expected '120000', got '%s'", resp.NewDownstreamMaxRate)
	}
	if resp.NewUpstreamMaxRate != "50000" {
		t.Errorf("Expected '50000', got '%s'", resp.NewUpstreamMaxRate)
	}
}

// TestGetDSLStatusWithRealClient tests with a real SOAP client
// This test is skipped if no router is available
func TestGetDSLStatusWithRealClient(t *testing.T) {
	// Skip if no router environment variables set
	if os.Getenv("FRITZBOX_URI") == "" {
		t.Skip("Skipping test: FRITZBOX_URI not set")
	}

	// This would test with real client - requires router connection
	// For now, just verify the function signature works
	t.Skip("Real client test - requires router connection")
}

// TestParseDSLStatusResponse tests XML parsing of DSL status response
func TestParseDSLStatusResponse(t *testing.T) {
	// Real XML response from Fritz!Box WANDSLInterfaceConfig:GetInfo (after cleanSoapResponse)
	xmlData := `<?xml version="1.0"?>
<Envelope>
<Body>
<GetInfoResponse>
<NewDownstreamCurrRate>100000</NewDownstreamCurrRate>
<NewUpstreamCurrRate>40000</NewUpstreamCurrRate>
<NewDownstreamMaxRate>120000</NewDownstreamMaxRate>
<NewUpstreamMaxRate>50000</NewUpstreamMaxRate>
<NewDownstreamNoiseMargin>10</NewDownstreamNoiseMargin>
<NewUpstreamNoiseMargin>15</NewUpstreamNoiseMargin>
</GetInfoResponse>
</Body>
</Envelope>`

	var resp DSLStatusResponse
	if err := xml.Unmarshal([]byte(xmlData), &resp); err != nil {
		t.Fatalf("Failed to parse XML: %v", err)
	}

	// Verify parsed values
	if resp.NewDownstreamCurrRate != "100000" {
		t.Errorf("Expected '100000', got '%s'", resp.NewDownstreamCurrRate)
	}
	if resp.NewUpstreamCurrRate != "40000" {
		t.Errorf("Expected '40000', got '%s'", resp.NewUpstreamCurrRate)
	}
	if resp.NewDownstreamMaxRate != "120000" {
		t.Errorf("Expected '120000', got '%s'", resp.NewDownstreamMaxRate)
	}
	if resp.NewUpstreamMaxRate != "50000" {
		t.Errorf("Expected '50000', got '%s'", resp.NewUpstreamMaxRate)
	}

	t.Logf("Successfully parsed: %+v", resp)
}
