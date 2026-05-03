package soap

import (
	"fmt"
	"strings"
	"testing"
)

func TestBuildSoapEnvelope(t *testing.T) {
	tests := []struct {
		name        string
		serviceType string
		action      string
		params      map[string]string
		contains    []string
	}{
		{
			name:        "Simple GetInfo",
			serviceType: "urn:dslforum-org:service:DeviceInfo:1",
			action:      "GetInfo",
			params:      map[string]string{},
			contains:    []string{"GetInfo", "DeviceInfo:1"},
		},
		{
			name:        "GetInfo with params",
			serviceType: "urn:dslforum-org:service:WLANConfiguration:1",
			action:      "SetEnable",
			params:      map[string]string{"NewEnable": "1"},
			contains:    []string{"SetEnable", "WLANConfiguration:1", "NewEnable", "1"},
		},
		{
			name:        "Reboot",
			serviceType: "urn:dslforum-org:service:DeviceConfig:1",
			action:      "Reboot",
			params:      map[string]string{},
			contains:    []string{"Reboot", "DeviceConfig:1"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := BuildSoapEnvelope(tt.serviceType, tt.action, tt.params)
			if err != nil {
				t.Fatalf("BuildSoapEnvelope failed: %v", err)
			}

			// Check that result contains expected strings
			for _, expected := range tt.contains {
				if !strings.Contains(result, expected) {
					t.Errorf("expected result to contain '%s', got:\n%s", expected, result)
				}
			}

			// Check XML structure
			if !strings.HasPrefix(result, `<?xml version="1.0"`) {
				t.Errorf("expected XML header, got: %s", result[:50])
			}
			if !strings.Contains(result, "</s:Envelope>") {
				t.Errorf("expected closing Envelope tag")
			}
		})
	}
}

func TestBuildSoapEnvelopeMultipleParams(t *testing.T) {
	params := map[string]string{
		"NewIndex":      "0",
		"NewEnable":     "1",
		"NewSomeValue": "test",
	}

	result, err := BuildSoapEnvelope(
		"urn:dslforum-org:service:X_AVM-DE_TAM:1",
		"SetEnable",
		params,
	)
	if err != nil {
		t.Fatalf("BuildSoapEnvelope failed: %v", err)
	}

	// Check all params are present
	for key, value := range params {
		expected := fmt.Sprintf("<%s>%s</%s>", key, value, key)
		if !strings.Contains(result, expected) {
			t.Errorf("expected param '%s' in result, got:\n%s", expected, result)
		}
	}
}

func TestNewClient(t *testing.T) {
	// This is a minimal test since we need actual auth/client
	// In a real test, we'd mock these
	t.Skip("NewClient requires auth.DigestClient and http.Client - skipping")
}

func TestTruncateString(t *testing.T) {
	tests := []struct {
		input    string
		maxLen   int
		expected string
	}{
		{"short", 10, "short"},
		{"exactlength", 12, "exactlength"},
		{"longstringthatneedstruncating", 10, "longstring..."},
		{"", 5, ""},
	}

	for _, tt := range tests {
		result := truncateString(tt.input, tt.maxLen)
		if result != tt.expected {
			t.Errorf("truncateString(%q, %d) = %q, want %q", tt.input, tt.maxLen, result, tt.expected)
		}
	}
}

func TestClientExtractServices(t *testing.T) {
	t.Skip("Requires full XML parsing test with mock description")
}
