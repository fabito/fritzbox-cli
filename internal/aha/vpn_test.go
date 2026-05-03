package aha

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestListVPNConnections is the RED phase test
func TestListVPNConnections(t *testing.T) {
	// Mock VPN API response
	mockResponse := `{
		"connection": [
			{
				"UID": "12345",
				"name": "Work VPN",
				"enabled": true,
				"status": "connected",
				"type": "wireguard"
			},
			{
				"UID": "67890",
				"name": "Home VPN",
				"enabled": false,
				"status": "disconnected",
				"type": "ipsec"
			}
		]
	}`

	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request
		if r.URL.Path != "/api/v0/generic/vpn" {
			t.Errorf("Expected path /api/v0/generic/vpn, got %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "AVM-SID test-sid" {
			t.Errorf("Expected AVM-SID header, got %s", r.Header.Get("Authorization"))
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(mockResponse))
	}))
	defer server.Close()

	// Create client with mock getSID function
	client := NewClient(server.URL, func() (string, error) {
		return "test-sid", nil
	})

	// Call ListVPNConnections
	connections, err := client.ListVPNConnections()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Verify results
	if len(connections) != 2 {
		t.Errorf("Expected 2 connections, got %d", len(connections))
	}

	if connections[0].Name != "Work VPN" {
		t.Errorf("Expected 'Work VPN', got '%s'", connections[0].Name)
	}
	if connections[0].Enabled != true {
		t.Errorf("Expected enabled=true")
	}
	if connections[0].Type != "wireguard" {
		t.Errorf("Expected type 'wireguard', got '%s'", connections[0].Type)
	}

	if connections[1].Name != "Home VPN" {
		t.Errorf("Expected 'Home VPN', got '%s'", connections[1].Name)
	}
}

// TestListVPNConnectionsEmpty tests empty response
func TestListVPNConnectionsEmpty(t *testing.T) {
	mockResponse := `{"connection": []}`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(mockResponse))
	}))
	defer server.Close()

	client := NewClient(server.URL, func() (string, error) {
		return "test-sid", nil
	})

	connections, err := client.ListVPNConnections()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(connections) != 0 {
		t.Errorf("Expected 0 connections, got %d", len(connections))
	}
}

// TestListVPNConnectionsError tests API error
func TestListVPNConnectionsError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := NewClient(server.URL, func() (string, error) {
		return "test-sid", nil
	})

	_, err := client.ListVPNConnections()
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
}

// TestSetVPNEnabled is the RED phase test for SetVPNEnabled
func TestSetVPNEnabled(t *testing.T) {
	// Track request to verify correct API call
	var capturedMethod string
	var capturedBody string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedMethod = r.Method
		body, _ := io.ReadAll(r.Body)
		capturedBody = string(body)

		// Verify auth header
		if r.Header.Get("Authorization") != "AVM-SID test-sid" {
			t.Errorf("Expected AVM-SID header, got %s", r.Header.Get("Authorization"))
		}

		// Return connection list for the first call (to find connection by name)
		if r.URL.Path == "/api/v0/generic/vpn" {
			w.Write([]byte(`{"connection": [{"UID": "vpn-123", "name": "Test VPN", "enabled": false}]}`))
			return
		}

		// Return success for PUT
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient(server.URL, func() (string, error) {
		return "test-sid", nil
	})

	// Call SetVPNEnabled (function doesn't exist yet - RED phase)
	err := client.SetVPNEnabled("Test VPN", true)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Verify correct API call was made
	// Expected: PUT /api/v0/generic/vpn/connection/{uid} with {"activated": "1"}
	if capturedMethod != "PUT" {
		t.Errorf("Expected PUT method, got %s", capturedMethod)
	}
	if capturedBody != `{"activated":"1"}` {
		t.Errorf("Expected body {\"activated\":\"1\"}, got %s", capturedBody)
	}
}

// TestSetVPNEnabledDisable tests disabling a VPN connection
func TestSetVPNEnabledDisable(t *testing.T) {
	var capturedBody string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		capturedBody = string(body)

		// Return connection list for the first call
		if r.URL.Path == "/api/v0/generic/vpn" {
			w.Write([]byte(`{"connection": [{"UID": "vpn-123", "name": "Test VPN", "enabled": true}]}`))
			return
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient(server.URL, func() (string, error) {
		return "test-sid", nil
	})

	err := client.SetVPNEnabled("Test VPN", false)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if capturedBody != `{"activated":"0"}` {
		t.Errorf("Expected body {\"activated\":\"0\"}, got %s", capturedBody)
	}
}

// TestSetVPNEnabledNotFound tests VPN connection not found
func TestSetVPNEnabledNotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Return empty connection list when listing
		if r.URL.Path == "/api/v0/generic/vpn" {
			w.Write([]byte(`{"connection": []}`))
		}
	}))
	defer server.Close()

	client := NewClient(server.URL, func() (string, error) {
		return "test-sid", nil
	})

	err := client.SetVPNEnabled("Non-existent VPN", true)
	if err == nil {
		t.Fatal("Expected error for non-existent VPN")
	}
}

// TestSetVPNEnabledSIDError tests error when SID fails
func TestSetVPNEnabledSIDError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient(server.URL, func() (string, error) {
		return "", fmt.Errorf("SID error")
	})

	err := client.SetVPNEnabled("Test VPN", true)
	if err == nil {
		t.Fatal("Expected error when SID fails")
	}
}

// TestSetVPNEnabledListError tests error when listing VPN connections fails
func TestSetVPNEnabledListError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v0/generic/vpn" {
			w.WriteHeader(http.StatusInternalServerError)
		}
	}))
	defer server.Close()

	client := NewClient(server.URL, func() (string, error) {
		return "test-sid", nil
	})

	err := client.SetVPNEnabled("Test VPN", true)
	if err == nil {
		t.Fatal("Expected error when listing VPN connections fails")
	}
}

// TestSetVPNEnabledPutError tests error when PUT request fails
func TestSetVPNEnabledPutError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Return connection list for the first call
		if r.URL.Path == "/api/v0/generic/vpn" {
			w.Write([]byte(`{"connection": [{"UID": "vpn-123", "name": "Test VPN"}]}`))
			return
		}
		// Return error for PUT
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := NewClient(server.URL, func() (string, error) {
		return "test-sid", nil
	})

	err := client.SetVPNEnabled("Test VPN", true)
	if err == nil {
		t.Fatal("Expected error when PUT request fails")
	}
}
