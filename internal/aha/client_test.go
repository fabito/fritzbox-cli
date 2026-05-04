package aha

import (
	"net/http"
	"testing"
)

func TestNewClient(t *testing.T) {
	getSID := func() (string, error) {
		return "test-sid", nil
	}

	client := NewClient("http://192.168.178.1", getSID)

	if client == nil {
		t.Fatal("NewClient returned nil")
	}
	if client.baseURL != "http://192.168.178.1" {
		t.Errorf("Expected baseURL 'http://192.168.178.1', got '%s'", client.baseURL)
	}
	if client.httpClient == nil {
		t.Error("Expected httpClient to be set (default)")
	}
	if client.getSID == nil {
		t.Error("Expected getSID function to be set")
	}
}

func TestNewClientWithDefaultHTTPClient(t *testing.T) {
	client := NewClient("http://test.local", func() (string, error) {
		return "", nil
	})

	// Verify it uses http.DefaultClient
	if client.httpClient != http.DefaultClient {
		t.Error("Expected httpClient to be http.DefaultClient")
	}
}

func TestClientGetSIDCalled(t *testing.T) {
	getSIDCalled := false
	getSID := func() (string, error) {
		getSIDCalled = true
		return "my-sid", nil
	}

	client := NewClient("http://test.local", getSID)

	// Access the getSID function (simulate internal usage)
	sid, _ := client.getSID()
	
	if !getSIDCalled {
		t.Error("Expected getSID function to be called")
	}
	if sid != "my-sid" {
		t.Errorf("Expected 'my-sid', got '%s'", sid)
	}
}
