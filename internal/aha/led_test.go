package aha

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestGetLEDStatus tests the GetLEDStatus function
func TestGetLEDStatus(t *testing.T) {
	// Mock JSON response from Fritz!Box LED page
	mockResponse := `{
		"ledDisplay": 0,
		"canDim": 1,
		"dimValue": 3
	}`

	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request has SID in POST form
		r.ParseForm()
		if r.FormValue("sid") == "" {
			t.Errorf("Expected SID in POST form")
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(mockResponse))
	}))
	defer server.Close()

	// Create client with test server URL and a mock getSID function
	client := NewClient(server.URL, func() (string, error) {
		return "test-sid", nil
	})

	// Call GetLEDStatus
	status, err := client.GetLEDStatus()
	if err != nil {
		t.Fatalf("GetLEDStatus failed: %v", err)
	}

	// Verify parsed values
	if status.LEDDisplay != 0 {
		t.Errorf("Expected LEDDisplay=0, got %d", status.LEDDisplay)
	}
	if status.CanDim != 1 {
		t.Errorf("Expected CanDim=1, got %d", status.CanDim)
	}
	if status.DimValue != 3 {
		t.Errorf("Expected DimValue=3, got %d", status.DimValue)
	}
}

// TestGetLEDStatusNoDim tests when device doesn't support dimming
func TestGetLEDStatusNoDim(t *testing.T) {
	mockResponse := `{
		"ledDisplay": 2,
		"canDim": 0
	}`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(mockResponse))
	}))
	defer server.Close()

	client := NewClient(server.URL, func() (string, error) {
		return "test-sid", nil
	})

	status, err := client.GetLEDStatus()
	if err != nil {
		t.Fatalf("GetLEDStatus failed: %v", err)
	}

	if status.LEDDisplay != 2 {
		t.Errorf("Expected LEDDisplay=2, got %d", status.LEDDisplay)
	}
	if status.CanDim != 0 {
		t.Errorf("Expected CanDim=0, got %d", status.CanDim)
	}
}

// TestGetLEDStatusInvalidSID tests error when SID is invalid
func TestGetLEDStatusInvalidSID(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	}))
	defer server.Close()

	client := NewClient(server.URL, func() (string, error) {
		return "invalid-sid", nil
	})

	_, err := client.GetLEDStatus()
	if err == nil {
		t.Error("Expected error for invalid SID, got nil")
	}
}

// TestGetLEDStatusInvalidJSON tests error when response is not valid JSON
func TestGetLEDStatusInvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("invalid json"))
	}))
	defer server.Close()

	client := NewClient(server.URL, func() (string, error) {
		return "test-sid", nil
	})

	_, err := client.GetLEDStatus()
	if err == nil {
		t.Error("Expected error for invalid JSON, got nil")
	}
}
