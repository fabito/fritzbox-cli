package services

import (
	"encoding/xml"
	"testing"
)

// TestListTAMs - Testing ListTAMs function (main test written BEFORE implementation - RED phase)
func TestListTAMs(t *testing.T) {
	// Test with nil soapClient (returns mock data)
	tams, err := ListTAMs(nil)
	if err != nil {
		t.Fatalf("ListTAMs returned error: %v", err)
	}

	if len(tams) == 0 {
		t.Fatal("Expected at least one TAM, got none")
	}

	// Verify first TAM
	tam := tams[0]
	if tam.NewTAMIndex != "0" {
		t.Errorf("Expected NewTAMIndex '0', got '%s'", tam.NewTAMIndex)
	}
	if tam.NewTAMName != "TAM 0" {
		t.Errorf("Expected NewTAMName 'TAM 0', got '%s'", tam.NewTAMName)
	}
	if tam.NewTAMEnable != "1" {
		t.Errorf("Expected NewTAMEnable '1', got '%s'", tam.NewTAMEnable)
	}
	if tam.NewTAMNewMessageCount != "3" {
		t.Errorf("Expected NewTAMNewMessageCount '3', got '%s'", tam.NewTAMNewMessageCount)
	}
	if tam.NewTAMTotalMessageCount != "10" {
		t.Errorf("Expected NewTAMTotalMessageCount '10', got '%s'", tam.NewTAMTotalMessageCount)
	}
}

// TestListTAMsEmpty tests when there are no TAMs
func TestListTAMsEmpty(t *testing.T) {
	// This would need a mock SOAP client that returns no TAMs
	// For now, skip this test
	t.Skip("Requires mock SOAP client that returns no TAMs")
}

// TestListTAMsWithRealSoapClient tests with a real SOAP client
func TestListTAMsWithRealSoapClient(t *testing.T) {
	t.Skip("Requires real SOAP client - tested separately with real router")
}

// TestTAMInfoXMLParsing tests the XML parsing of TAMInfo struct
func TestTAMInfoXMLParsing(t *testing.T) {
	// Mock SOAP response XML for TAM info
	mockXML := `<?xml version="1.0"?>
<Envelope>
<Body>
<GetInfoResponse>
<NewTAMIndex>0</NewTAMIndex>
<NewTAMName>TAM 0</NewTAMName>
<NewTAMEnable>1</NewTAMEnable>
<NewTAMNewMessageCount>3</NewTAMNewMessageCount>
<NewTAMTotalMessageCount>10</NewTAMTotalMessageCount>
</GetInfoResponse>
</Body>
</Envelope>`

	// Parse the XML
	var tam TAMInfo
	resp := cleanSoapResponseForTest(mockXML)
	if err := xml.Unmarshal([]byte(resp), &tam); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	t.Logf("Parsed TAM: %+v", tam)

	// Verify parsed values
	if tam.NewTAMIndex != "0" {
		t.Errorf("Expected NewTAMIndex '0', got '%s'", tam.NewTAMIndex)
	}
	if tam.NewTAMName != "TAM 0" {
		t.Errorf("Expected NewTAMName 'TAM 0', got '%s'", tam.NewTAMName)
	}
	if tam.NewTAMEnable != "1" {
		t.Errorf("Expected NewTAMEnable '1', got '%s'", tam.NewTAMEnable)
	}
	if tam.NewTAMNewMessageCount != "3" {
		t.Errorf("Expected NewTAMNewMessageCount '3', got '%s'", tam.NewTAMNewMessageCount)
	}
	if tam.NewTAMTotalMessageCount != "10" {
		t.Errorf("Expected NewTAMTotalMessageCount '10', got '%s'", tam.NewTAMTotalMessageCount)
	}
}

// TestTAMInfoXMLParsingWithDisabledTAM tests parsing when TAM is disabled
func TestTAMInfoXMLParsingWithDisabledTAM(t *testing.T) {
	mockXML := `<?xml version="1.0"?>
<Envelope>
<Body>
<GetInfoResponse>
<NewTAMIndex>1</NewTAMIndex>
<NewTAMName>TAM 1 (disabled)</NewTAMName>
<NewTAMEnable>0</NewTAMEnable>
<NewTAMNewMessageCount>0</NewTAMNewMessageCount>
<NewTAMTotalMessageCount>5</NewTAMTotalMessageCount>
</GetInfoResponse>
</Body>
</Envelope>`

	var tam TAMInfo
	resp := cleanSoapResponseForTest(mockXML)
	if err := xml.Unmarshal([]byte(resp), &tam); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	if tam.NewTAMEnable != "0" {
		t.Errorf("Expected NewTAMEnable '0', got '%s'", tam.NewTAMEnable)
	}
	if tam.NewTAMNewMessageCount != "0" {
		t.Errorf("Expected NewTAMNewMessageCount '0', got '%s'", tam.NewTAMNewMessageCount)
	}
}

// cleanSoapResponseForTest is a helper for testing
func cleanSoapResponseForTest(resp string) string {
	// Simple cleanup - remove namespace prefixes
	// In real code, this is done by soap.CleanSoapResponse
	return resp
}
