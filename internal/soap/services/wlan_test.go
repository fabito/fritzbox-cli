package services

import (
	"errors"
	"testing"
)

// Mock SOAP caller for testing
type mockSOAPCaller struct {
	response string
	err       error
}

func (m *mockSOAPCaller) Call(servicePath, action, body string) (string, error) {
	return m.response, m.err
}

// TestGetWLANQRCode_Red tests the RED phase (should fail initially)
// This is now the GREEN/REFACTOR phase - tests should pass
func TestGetWLANQRCode_NilClient(t *testing.T) {
	// With nil client, should return mock data
	qrCode, err := GetWLANQRCode(nil, 1)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	expected := "WIFI:T:WPA;S:TestSSID;P:TestPass;;"
	if qrCode != expected {
		t.Errorf("Expected '%s', got '%s'", expected, qrCode)
	}
}

func TestGetWLANQRCode_InvalidBand(t *testing.T) {
	_, err := GetWLANQRCode(nil, 99)
	if err == nil {
		t.Error("Expected error for invalid band 99")
	}
}

func TestGetWLANQRCode_MockClient(t *testing.T) {
	// Mock response with WPS info
	mockResp := `<?xml version="1.0"?>
<Envelope>
<Body>
<GetDefaultWPSInfoResponse>
<NewSSID>MyWifi</NewSSID>
<NewKeyPassphrase>MySecretKey123</NewKeyPassphrase>
</GetDefaultWPSInfoResponse>
</Body>
</Envelope>`

	mockClient := &mockSOAPCaller{
		response: mockResp,
		err:       nil,
	}

	qrCode, err := GetWLANQRCode(mockClient, 1)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Verify QR code format
	expected := "WIFI:T:WPA;S:MyWifi;P:MySecretKey123;;"
	if qrCode != expected {
		t.Errorf("Expected '%s', got '%s'", expected, qrCode)
	}
}

func TestGetWLANQRCode_SoapError(t *testing.T) {
	mockClient := &mockSOAPCaller{
		response: "",
		err:       errors.New("SOAP call failed"),
	}

	_, err := GetWLANQRCode(mockClient, 1)
	if err == nil {
		t.Error("Expected error from SOAP call")
	}
}
