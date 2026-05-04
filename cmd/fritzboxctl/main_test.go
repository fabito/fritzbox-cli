package main

import (
	"testing"
)

func TestHasPort(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		expected bool
	}{
		{"URL without protocol, no port", "192.168.178.1", false},
		{"URL without protocol, with port", "192.168.178.1:49000", true},
		{"HTTP URL without port", "http://192.168.178.1", false},
		{"HTTP URL with port", "http://192.168.178.1:49000", true},
		{"HTTPS URL without port", "https://192.168.178.1", false},
		{"HTTPS URL with port", "https://192.168.178.1:443", true},
		{"URL with path, no port", "http://192.168.178.1/path", false},
		{"URL with path and port", "http://192.168.178.1:8080/path", true},
		{"URL with query, no port", "http://192.168.178.1/path?query=1", false},
		{"URL with query and port", "http://192.168.178.1:9000/path?query=1", true},
		{"URL with fragment, no port", "http://192.168.178.1/path#section", false},
		{"URL with fragment and port", "http://192.168.178.1:9999/path#section", true},
		{"Hostname without port", "http://fritz.box", false},
		{"Hostname with port", "http://fritz.box:49000", true},
		{"IPv6 without port", "http://[::1]", false},
		{"Empty URL", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := hasPort(tt.url)
			if result != tt.expected {
				t.Errorf("hasPort(%q) = %v, want %v", tt.url, result, tt.expected)
			}
		})
	}
}

func TestNewRootCommand(t *testing.T) {
	cmd := NewRootCommand()

	if cmd == nil {
		t.Fatal("NewRootCommand returned nil")
	}
	if cmd.Use != "fritzboxctl" {
		t.Errorf("Expected Use 'fritzboxctl', got '%s'", cmd.Use)
	}
	if cmd.Short == "" {
		t.Error("Expected Short description to be set")
	}
	if cmd.Long == "" {
		t.Error("Expected Long description to be set")
	}

	// Check that subcommands are registered
	subcommands := cmd.Commands()
	expectedSubcommands := []string{"device", "wlan", "network", "tam", "vpn", "led", "log"}

	subcommandNames := make(map[string]bool)
	for _, sub := range subcommands {
		subcommandNames[sub.Name()] = true
	}

	for _, expected := range expectedSubcommands {
		if !subcommandNames[expected] {
			t.Errorf("Expected subcommand '%s' to be registered", expected)
		}
	}
}

func TestRootCommandFlags(t *testing.T) {
	cmd := NewRootCommand()

	// Check persistent flags
	flags := cmd.PersistentFlags()

	testCases := []struct {
		name      string
		shorthand string
	}{
		{"router-uri", "H"},
		{"username", "u"},
		{"password", "p"},
		{"timeout", "t"},
		{"output", "o"},
		{"config", "c"},
		{"verbose", "v"},
	}

	for _, tc := range testCases {
		flag := flags.Lookup(tc.name)
		if flag == nil {
			t.Errorf("Expected flag '%s' to be registered", tc.name)
			continue
		}
		if flag.Shorthand != tc.shorthand {
			t.Errorf("Expected flag '%s' to have shorthand '%s', got '%s'", tc.name, tc.shorthand, flag.Shorthand)
		}
	}
}

func TestGetBandName(t *testing.T) {
	tests := []struct {
		band     int
		expected string
	}{
		{1, "2.4 GHz"},
		{2, "5 GHz"},
		{3, "5 GHz (2nd channel)"},
		{4, "Guest Network"},
		{5, "Unknown (5)"},
		{0, "Unknown (0)"},
		{-1, "Unknown (-1)"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := getBandName(tt.band)
			if result != tt.expected {
				t.Errorf("getBandName(%d) = %q, want %q", tt.band, result, tt.expected)
			}
		})
	}
}

func TestFormatHostListOutput(t *testing.T) {
	// Test empty list
	hosts := []HostEntry{}
	output := formatHostList(hosts)
	if output != "No devices found on the network.\n" {
		t.Errorf("Expected 'No devices found' message, got: %s", output)
	}

	// Test with hosts
	hosts = []HostEntry{
		{
			HostName:      "TestHost",
			IPAddress:     "192.168.178.20",
			MACAddress:    "AA:BB:CC:DD:EE:FF",
			InterfaceType: "WLAN",
			Active:        true,
		},
	}
	output = formatHostList(hosts)

	// Check output contains expected data
	expectedContents := []string{"TestHost", "192.168.178.20", "AA:BB:CC:DD:EE:FF", "WLAN"}
	for _, expected := range expectedContents {
		if !containsStr(output, expected) {
			t.Errorf("Expected output to contain '%s', got: %s", expected, output)
		}
	}
}

func TestFormatHostListInactiveHost(t *testing.T) {
	hosts := []HostEntry{
		{
			HostName:      "InactiveHost",
			IPAddress:     "192.168.178.30",
			MACAddress:    "11:22:33:44:55:66",
			InterfaceType: "LAN",
			Active:        false,
		},
	}
	output := formatHostList(hosts)

	// Should contain "(inactive)" marker
	if !containsStr(output, "inactive") {
		t.Errorf("Expected output to contain 'inactive' marker, got: %s", output)
	}
}

func TestFormatHostListUnknownHostname(t *testing.T) {
	hosts := []HostEntry{
		{
			HostName:      "", // Empty hostname
			IPAddress:     "192.168.178.40",
			MACAddress:    "AA:BB:CC:DD:EE:00",
			InterfaceType: "WLAN",
			Active:        true,
		},
	}
	output := formatHostList(hosts)

	// Should contain "(unknown)" for empty hostname
	if !containsStr(output, "(unknown)") {
		t.Errorf("Expected output to contain '(unknown)' for empty hostname, got: %s", output)
	}
}

// Helper function
func containsStr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
