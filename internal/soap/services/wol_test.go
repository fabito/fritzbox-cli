package services

import (
	"fmt"
	"testing"
)

// TestWakeOnLAN tests the WakeOnLAN function with mock client
func TestWakeOnLAN(t *testing.T) {
	// Create a mock client that implements the SOAP caller interface
	mockClient := &mockSOAPCaller{
		response: `<?xml version="1.0"?>
<Envelope>
  <Body>
    <X_AVM-DE_WakeOnLANByMACAddressResponse xmlns="urn:dslforum-org:service:Hosts:1">
    </X_AVM-DE_WakeOnLANByMACAddressResponse>
  </Body>
</Envelope>`,
		err: nil,
	}

	// Call WakeOnLAN - this should compile after implementation
	err := WakeOnLAN("AA:BB:CC:DD:EE:FF", mockClient)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestWakeOnLAN_InvalidMAC(t *testing.T) {
	// Test with invalid MAC (should still send, Fritz!Box validates)
	mockClient := &mockSOAPCaller{
		response: `<?xml version="1.0"?><Envelope><Body><X_AVM-DE_WakeOnLANByMACAddressResponse></X_AVM-DE_WakeOnLANByMACAddressResponse></Body></Envelope>`,
		err:       nil,
	}

	err := WakeOnLAN("INVALID:MAC", mockClient)
	// Should not error (Fritz!Box handles validation)
	if err != nil {
		t.Errorf("Expected no error for invalid MAC, got: %v", err)
	}
}

func TestWakeOnLAN_Error(t *testing.T) {
	// Test with error from SOAP client
	mockClient := &mockSOAPCaller{
		response: "",
		err:       fmt.Errorf("connection failed"),
	}

	err := WakeOnLAN("AA:BB:CC:DD:EE:FF", mockClient)
	if err == nil {
		t.Error("Expected error, got nil")
	}
}
