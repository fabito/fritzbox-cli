package services

import (
	"encoding/xml"
	"fmt"
	"strings"
	"testing"

	"github.com/fabito/fritzboxctl/internal/soap"
)

// TestGetDeviceInfo tests the GetDeviceInfo function
func TestGetDeviceInfo(t *testing.T) {
	// Test with nil client (returns mock data for testing)
	resp, err := GetDeviceInfo(nil)
	if err != nil {
		t.Fatalf("GetDeviceInfo failed: %v", err)
	}
	if resp.NewModelName != "FRITZ!Box 7530" {
		t.Errorf("Expected 'FRITZ!Box 7530', got '%s'", resp.NewModelName)
	}
}

// TestParseGetInfoResponse tests XML parsing of device info response
func TestParseGetInfoResponse(t *testing.T) {
	// Real XML response from Fritz!Box 7530 (after cleanSoapResponse)
	xmlData := `<?xml version="1.0"?>
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

	var resp GetInfoResponse
	if err := xml.Unmarshal([]byte(xmlData), &resp); err != nil {
		t.Fatalf("Failed to parse XML: %v", err)
	}

	// Verify parsed values
	if resp.NewManufacturerName != "AVM" {
		t.Errorf("Expected 'AVM', got '%s'", resp.NewManufacturerName)
	}
	if resp.NewModelName != "FRITZ!Box 7530" {
		t.Errorf("Expected 'FRITZ!Box 7530', got '%s'", resp.NewModelName)
	}
	if resp.NewSerialNumber != "DC396F6BC160" {
		t.Errorf("Expected 'DC396F6BC160', got '%s'", resp.NewSerialNumber)
	}
	if resp.NewSoftwareVersion != "164.08.02" {
		t.Errorf("Expected '164.08.02', got '%s'", resp.NewSoftwareVersion)
	}
	if resp.NewHardwareVersion != "FRITZ!Box 7530" {
		t.Errorf("Expected 'FRITZ!Box 7530', got '%s'", resp.NewHardwareVersion)
	}

	t.Logf("Successfully parsed: %+v", resp)
}

// TestReboot tests the Reboot function
func TestReboot(t *testing.T) {
	// Test with nil client (mock - should succeed)
	err := Reboot(nil)
	if err != nil {
		t.Errorf("Reboot with nil client should not error: %v", err)
	}

	// Test with mock client that succeeds
	mockClient := &mockSoapClient{err: nil}
	err = Reboot(mockClient)
	if err != nil {
		t.Errorf("Reboot with mock success client should not error: %v", err)
	}

	// Test with mock client that fails
	mockClientErr := &mockSoapClient{err: fmt.Errorf("mock error")}
	err = Reboot(mockClientErr)
	if err == nil {
		t.Error("Reboot with mock error client should return error")
	}

	// Test that correct SOAP parameters are passed
	mockRecorder := &mockSoapRecorder{}
	err = Reboot(mockRecorder)
	if err != nil {
		t.Errorf("Reboot should not error with recorder: %v", err)
	}
	if mockRecorder.servicePath != "/upnp/control/deviceconfig" {
		t.Errorf("Expected servicePath '/upnp/control/deviceconfig', got '%s'", mockRecorder.servicePath)
	}
	if !strings.Contains(mockRecorder.soapAction, "Reboot") {
		t.Errorf("Expected soapAction to contain 'Reboot', got '%s'", mockRecorder.soapAction)
	}
}

// mockSoapClient is a simple mock SOAP client for testing
type mockSoapClient struct {
	err error
}

func (m *mockSoapClient) Call(servicePath, soapAction, soapBody string) (string, error) {
	return "", m.err
}

// mockSoapRecorder records call arguments for verification
type mockSoapRecorder struct {
	servicePath string
	soapAction  string
	soapBody    string
}

func (m *mockSoapRecorder) Call(servicePath, soapAction, soapBody string) (string, error) {
	m.servicePath = servicePath
	m.soapAction = soapAction
	m.soapBody = soapBody
	return "", nil
}

// TestCleanSoapResponse tests the CleanSoapResponse function
func TestCleanSoapResponse(t *testing.T) {
	input := `<Envelope encodingStyle="http://schemas.xmlsoap.org/soap/encoding/"><Body><GetInfoResponse ></Body></Envelope>`
	output := soap.CleanSoapResponse(input)

	// Check that encodingStyle was removed
	if strings.Contains(output, "encodingStyle") {
		t.Error("CleanSoapResponse should remove encodingStyle attribute")
	}

	// Check that spaces before > are fixed
	if strings.Contains(output, " >") {
		t.Error("CleanSoapResponse should fix spaces before >")
	}

	t.Logf("Cleaned: %s", output)
}
