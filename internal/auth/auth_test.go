package auth

import (
	"testing"
	"time"
)

func TestNewAuth(t *testing.T) {
	auth, err := NewAuth("testuser", "testpass", 30*time.Second, true)

	if err != nil {
		t.Fatalf("NewAuth failed: %v", err)
	}
	if auth == nil {
		t.Fatal("NewAuth returned nil")
	}
	if auth.DigestClient == nil {
		t.Error("Expected DigestClient to be set")
	}
	if auth.HTTPClient == nil {
		t.Error("Expected HTTPClient to be set")
	}
	if auth.SIDManager != nil {
		t.Error("Expected SIDManager to be nil initially (set later)")
	}
}

func TestNewAuthWithTimeout(t *testing.T) {
	timeout := 60 * time.Second
	auth, err := NewAuth("user", "pass", timeout, false)

	if err != nil {
		t.Fatalf("NewAuth failed: %v", err)
	}
	if auth.HTTPClient.Timeout != timeout {
		t.Errorf("Expected timeout %v, got %v", timeout, auth.HTTPClient.Timeout)
	}
}

func TestAuthSetSIDManager(t *testing.T) {
	auth, _ := NewAuth("user", "pass", 10*time.Second, true)

	// Initially nil
	if auth.SIDManager != nil {
		t.Error("Expected SIDManager to be nil initially")
	}

	// Create and set a mock SID manager
	mockSoapCall := func(service, action, body string) (string, error) {
		return "", nil
	}
	sidManager := NewSIDManager(mockSoapCall)
	auth.SetSIDManager(sidManager)

	if auth.SIDManager == nil {
		t.Error("Expected SIDManager to be set after SetSIDManager")
	}
	if auth.SIDManager != sidManager {
		t.Error("Expected SIDManager to match the one that was set")
	}
}

func TestAuthGetSIDWithoutManager(t *testing.T) {
	auth, _ := NewAuth("user", "pass", 10*time.Second, true)

	// Don't set SIDManager
	_, err := auth.GetSID()
	if err == nil {
		t.Error("Expected error when GetSID called without SIDManager")
	}
}

func TestAuthGetSIDWithManager(t *testing.T) {
	auth, _ := NewAuth("user", "pass", 10*time.Second, true)

	mockSoapCall := func(service, action, body string) (string, error) {
		return `<NewX_AVM-DE_UrlSID>sid=test-auth-sid</NewX_AVM-DE_UrlSID>`, nil
	}
	auth.SetSIDManager(NewSIDManager(mockSoapCall))

	sid, err := auth.GetSID()
	if err != nil {
		t.Fatalf("GetSID failed: %v", err)
	}
	if sid != "test-auth-sid" {
		t.Errorf("Expected 'test-auth-sid', got '%s'", sid)
	}
}

func TestAuthLoginWithoutManager(t *testing.T) {
	auth, _ := NewAuth("user", "pass", 10*time.Second, true)

	_, err := auth.Login()
	if err == nil {
		t.Error("Expected error when Login called without SIDManager")
	}
}

func TestAuthLoginWithManager(t *testing.T) {
	auth, _ := NewAuth("user", "pass", 10*time.Second, true)

	mockSoapCall := func(service, action, body string) (string, error) {
		return `<NewX_AVM-DE_UrlSID>sid=login-sid</NewX_AVM-DE_UrlSID>`, nil
	}
	auth.SetSIDManager(NewSIDManager(mockSoapCall))

	sid, err := auth.Login()
	if err != nil {
		t.Fatalf("Login failed: %v", err)
	}
	if sid != "login-sid" {
		t.Errorf("Expected 'login-sid', got '%s'", sid)
	}
}

func TestAuthLogoutWithoutManager(t *testing.T) {
	auth, _ := NewAuth("user", "pass", 10*time.Second, true)

	// Logout without SIDManager should not error (just return nil)
	err := auth.Logout()
	if err != nil {
		t.Errorf("Logout without SIDManager should not error: %v", err)
	}
}

func TestAuthLogoutWithManager(t *testing.T) {
	auth, _ := NewAuth("user", "pass", 10*time.Second, true)

	mockSoapCall := func(service, action, body string) (string, error) {
		return `<NewX_AVM-DE_UrlSID>sid=test-sid</NewX_AVM-DE_UrlSID>`, nil
	}
	manager := NewSIDManager(mockSoapCall)
	auth.SetSIDManager(manager)

	// Login first
	auth.Login()
	if manager.sid == "" {
		t.Fatal("Expected SID to be set after Login")
	}

	// Logout should clear the SID
	err := auth.Logout()
	if err != nil {
		t.Fatalf("Logout failed: %v", err)
	}
	if manager.sid != "" {
		t.Errorf("Expected SID to be cleared after Logout, got '%s'", manager.sid)
	}
}
