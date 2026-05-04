package auth

import (
	"errors"
	"testing"
	"time"
)

func TestNewSIDManager(t *testing.T) {
	mockSoapCall := func(service, action, body string) (string, error) {
		return "", nil
	}

	manager := NewSIDManager(mockSoapCall)
	if manager == nil {
		t.Fatal("NewSIDManager returned nil")
	}
	if manager.soapCall == nil {
		t.Error("Expected soapCall to be set")
	}
	if manager.sid != "" {
		t.Error("Expected sid to be empty initially")
	}
}

func TestSIDManagerLogin(t *testing.T) {
	// Create a mock SOAP call function that returns a valid SID response
	mockSoapCall := func(service, action, body string) (string, error) {
		// Verify correct service path and action
		if service != "/upnp/control/deviceconfig" {
			t.Errorf("Expected service '/upnp/control/deviceconfig', got '%s'", service)
		}
		if action != "urn:dslforum-org:service:DeviceConfig:1#X_AVM-DE_CreateUrlSID" {
			t.Errorf("Expected CreateUrlSID action, got '%s'", action)
		}

		// Return a mock response with SID
		return `<?xml version="1.0"?>
<s:Envelope xmlns:s="http://schemas.xmlsoap.org/soap/envelope/">
<s:Body>
<u:X_AVM-DE_CreateUrlSIDResponse xmlns:u="urn:dslforum-org:service:DeviceConfig:1">
<NewX_AVM-DE_UrlSID>sid=abc123def456</NewX_AVM-DE_UrlSID>
</u:X_AVM-DE_CreateUrlSIDResponse>
</s:Body>
</s:Envelope>`, nil
	}

	manager := NewSIDManager(mockSoapCall)
	sid, err := manager.Login()

	if err != nil {
		t.Fatalf("Login failed: %v", err)
	}
	if sid != "abc123def456" {
		t.Errorf("Expected SID 'abc123def456', got '%s'", sid)
	}
	if manager.sid != "abc123def456" {
		t.Errorf("Expected manager.sid to be 'abc123def456', got '%s'", manager.sid)
	}
	// Verify expiration is set to roughly 20 minutes from now
	expectedExpiry := time.Now().Add(20 * time.Minute)
	if manager.expiresAt.Before(expectedExpiry.Add(-1*time.Minute)) || manager.expiresAt.After(expectedExpiry.Add(1*time.Minute)) {
		t.Errorf("Expected expiresAt to be around 20 minutes from now, got %v", manager.expiresAt)
	}
}

func TestSIDManagerLoginError(t *testing.T) {
	mockSoapCall := func(service, action, body string) (string, error) {
		return "", errors.New("connection failed")
	}

	manager := NewSIDManager(mockSoapCall)
	_, err := manager.Login()

	if err == nil {
		t.Error("Expected error from Login when SOAP call fails")
	}
}

func TestSIDManagerLoginInvalidResponse(t *testing.T) {
	mockSoapCall := func(service, action, body string) (string, error) {
		// Return response without SID
		return `<?xml version="1.0"?><Envelope><Body></Body></Envelope>`, nil
	}

	manager := NewSIDManager(mockSoapCall)
	_, err := manager.Login()

	if err == nil {
		t.Error("Expected error when SID cannot be extracted from response")
	}
}

func TestSIDManagerGetSID(t *testing.T) {
	callCount := 0
	mockSoapCall := func(service, action, body string) (string, error) {
		callCount++
		return `<NewX_AVM-DE_UrlSID>sid=test-sid-123</NewX_AVM-DE_UrlSID>`, nil
	}

	manager := NewSIDManager(mockSoapCall)

	// First call should trigger Login
	sid1, err := manager.GetSID()
	if err != nil {
		t.Fatalf("GetSID failed: %v", err)
	}
	if sid1 != "test-sid-123" {
		t.Errorf("Expected 'test-sid-123', got '%s'", sid1)
	}
	if callCount != 1 {
		t.Errorf("Expected 1 SOAP call, got %d", callCount)
	}

	// Second call should return cached SID (no additional SOAP call)
	sid2, err := manager.GetSID()
	if err != nil {
		t.Fatalf("GetSID failed: %v", err)
	}
	if sid2 != "test-sid-123" {
		t.Errorf("Expected 'test-sid-123', got '%s'", sid2)
	}
	if callCount != 1 {
		t.Errorf("Expected still 1 SOAP call (cached), got %d", callCount)
	}
}

func TestSIDManagerGetSIDExpired(t *testing.T) {
	callCount := 0
	mockSoapCall := func(service, action, body string) (string, error) {
		callCount++
		return `<NewX_AVM-DE_UrlSID>sid=new-sid-456</NewX_AVM-DE_UrlSID>`, nil
	}

	manager := NewSIDManager(mockSoapCall)

	// Set an expired SID
	manager.sid = "old-sid"
	manager.expiresAt = time.Now().Add(-1 * time.Minute) // Expired 1 minute ago

	// GetSID should trigger a new Login
	sid, err := manager.GetSID()
	if err != nil {
		t.Fatalf("GetSID failed: %v", err)
	}
	if sid != "new-sid-456" {
		t.Errorf("Expected 'new-sid-456', got '%s'", sid)
	}
	if callCount != 1 {
		t.Errorf("Expected 1 SOAP call (refresh), got %d", callCount)
	}
}

func TestSIDManagerLogout(t *testing.T) {
	mockSoapCall := func(service, action, body string) (string, error) {
		return `<NewX_AVM-DE_UrlSID>sid=test-sid</NewX_AVM-DE_UrlSID>`, nil
	}

	manager := NewSIDManager(mockSoapCall)
	manager.Login()

	if manager.sid == "" {
		t.Fatal("Expected SID to be set after Login")
	}

	err := manager.Logout()
	if err != nil {
		t.Fatalf("Logout failed: %v", err)
	}
	if manager.sid != "" {
		t.Errorf("Expected SID to be cleared after Logout, got '%s'", manager.sid)
	}
	if !manager.expiresAt.IsZero() {
		t.Error("Expected expiresAt to be zero after Logout")
	}
}

func TestSIDManagerRefresh(t *testing.T) {
	callCount := 0
	mockSoapCall := func(service, action, body string) (string, error) {
		callCount++
		return `<NewX_AVM-DE_UrlSID>sid=refreshed-sid</NewX_AVM-DE_UrlSID>`, nil
	}

	manager := NewSIDManager(mockSoapCall)
	manager.sid = "old-sid"
	manager.expiresAt = time.Now().Add(10 * time.Minute) // Not expired

	// Refresh should clear and re-login even if not expired
	err := manager.Refresh()
	if err != nil {
		t.Fatalf("Refresh failed: %v", err)
	}
	if manager.sid != "refreshed-sid" {
		t.Errorf("Expected 'refreshed-sid', got '%s'", manager.sid)
	}
	if callCount != 1 {
		t.Errorf("Expected 1 SOAP call, got %d", callCount)
	}
}

func TestExtractSIDFromString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Valid SID",
			input:    `<NewX_AVM-DE_UrlSID>sid=abc123def456</NewX_AVM-DE_UrlSID>`,
			expected: "abc123def456",
		},
		{
			name:     "SID with XML closing tag",
			input:    `sid=test123</NewX_AVM-DE_UrlSID>`,
			expected: "test123",
		},
		{
			name:     "SID with space after",
			input:    `sid=mysid `,
			expected: "mysid",
		},
		{
			name:     "SID with ampersand after",
			input:    `sid=abc&other=value`,
			expected: "abc",
		},
		{
			name:     "SID with quote after",
			input:    `sid=xyz"`,
			expected: "xyz",
		},
		{
			name:     "No SID in response",
			input:    `<Response></Response>`,
			expected: "",
		},
		{
			name:     "Empty SID",
			input:    `sid=<`,
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractSIDFromString(tt.input)
			if result != tt.expected {
				t.Errorf("extractSIDFromString(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}
