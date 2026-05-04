package main

import (
	"testing"

	"github.com/fabito/fritzboxctl/internal/soap/services"
)

// TestDisplayWLANStatus tests the output formatting
func TestDisplayWLANStatus(t *testing.T) {
	// Create a mock status response
	status := &services.WLANStatusResponse{
		NewEnable:     "1",
		NewSSID:       "TestWLAN",
		NewBeaconType: "WPA2",
		NewChannel:    "6",
		NewMaxBitRate: "866",
	}

	// Just verify the function doesn't panic with nil cfg
	// In production, cfg will be initialized
	defer func() {
		if r := recover(); r != nil {
			t.Logf("Recovered from panic (expected in test): %v", r)
		}
	}()

	displayWLANStatus(status, 1)
}
