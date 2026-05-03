package main

import (
	"testing"
)

// TestFormatHostList tests the formatting of host list output
// RED PHASE: This references formatHostList which doesn't exist yet
func TestFormatHostList(t *testing.T) {
	// Mock host data
	hosts := []HostEntry{
		{HostName: "MyLaptop", IPAddress: "192.168.178.20", MACAddress: "AA:BB:CC:DD:EE:FF", InterfaceType: "WLAN", Active: true},
		{HostName: "MyPhone", IPAddress: "192.168.178.21", MACAddress: "11:22:33:44:55:66", InterfaceType: "WLAN", Active: true},
		{HostName: "", IPAddress: "192.168.178.22", MACAddress: "AA:BB:CC:DD:EE:11", InterfaceType: "LAN", Active: false},
	}

	// This should FAIL - formatHostList doesn't exist yet!
	output := formatHostList(hosts)

	// Check output contains expected data
	expectedStrings := []string{
		"MyLaptop",
		"192.168.178.20",
		"AA:BB:CC:DD:EE:FF",
		"WLAN",
		"MyPhone",
		"192.168.178.21",
	}

	for _, expected := range expectedStrings {
		if !contains(output, expected) {
			t.Errorf("Expected output to contain '%s', but got:\n%s", expected, output)
		}
	}

	t.Logf("Formatted output:\n%s", output)
}

// TestFormatHostListEmpty tests formatting with no hosts
func TestFormatHostListEmpty(t *testing.T) {
	hosts := []HostEntry{}

	output := formatHostList(hosts)

	if !contains(output, "No devices found") {
		t.Errorf("Expected 'No devices found' message, got:\n%s", output)
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsString(s, substr))
}

func containsString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
