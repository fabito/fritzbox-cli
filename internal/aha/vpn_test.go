package aha

import (
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

	// Create client with mock SID
	client := &Client{
		baseURL: server.URL,
		sid:     "test-sid",
		httpClient: http.DefaultClient,
	}

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

	client := &Client{
		baseURL: server.URL,
		sid:     "test-sid",
		httpClient: http.DefaultClient,
	}

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

	client := &Client{
		baseURL: server.URL,
		sid:     "test-sid",
		httpClient: http.DefaultClient,
	}

	_, err := client.ListVPNConnections()
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
}
