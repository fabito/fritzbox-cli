package auth

import (
	"fmt"
	"log/slog"
	"strings"
	"time"
)

// SIDManager manages AHA session IDs (SID)
type SIDManager struct {
	soapCall  func(service, action, body string) (string, error)
	sid       string
	expiresAt time.Time
}

// NewSIDManager creates a new SIDManager
func NewSIDManager(soapCall func(service, action, body string) (string, error)) *SIDManager {
	return &SIDManager{
		soapCall: soapCall,
	}
}

// GetSID returns a valid SID, obtaining a new one if necessary
func (s *SIDManager) GetSID() (string, error) {
	slog.Debug("GetSID: called", "current_sid", s.sid, "expires_at", s.expiresAt)
	if s.sid != "" && time.Now().Before(s.expiresAt) {
		slog.Debug("GetSID: returning cached SID", "sid", s.sid)
		return s.sid, nil
	}
	return s.Login()
}

// Login obtains a new SID from the Fritz!Box
func (s *SIDManager) Login() (string, error) {
	slog.Debug("Login: calling soapCall")
	soapBody := `<?xml version="1.0" encoding="utf-8"?>
<s:Envelope s:encodingStyle="http://schemas.xmlsoap.org/soap/encoding/" xmlns:s="http://schemas.xmlsoap.org/soap/envelope/">
    <s:Body>
        <u:X_AVM-DE_CreateUrlSID xmlns:u="urn:dslforum-org:service:DeviceConfig:1">
        </u:X_AVM-DE_CreateUrlSID>
    </s:Body>
</s:Envelope>`

	resp, err := s.soapCall(
		"/upnp/control/deviceconfig",
		"urn:dslforum-org:service:DeviceConfig:1#X_AVM-DE_CreateUrlSID",
		soapBody,
	)
	if err != nil {
		return "", fmt.Errorf("failed to call CreateUrlSID: %w", err)
	}

	slog.Debug("Login: SOAP response received", "response", resp[:min(len(resp), 500)])

	// Extract SID using string parsing (more reliable than XML for this case)
	sid := extractSIDFromString(resp)
	if sid == "" {
		return "", fmt.Errorf("failed to extract SID from response")
	}

	s.sid = sid
	s.expiresAt = time.Now().Add(20 * time.Minute)
	slog.Debug("Login: SID obtained", "sid", sid)
	return sid, nil
}

// Logout clears the current SID
func (s *SIDManager) Logout() error {
	s.sid = ""
	s.expiresAt = time.Time{}
	return nil
}

// Refresh forces a new SID to be obtained
func (s *SIDManager) Refresh() error {
	s.sid = ""
	_, err := s.Login()
	return err
}

// extractSIDFromString extracts SID from SOAP response string
// Format: <NewX_AVM-DE_UrlSID>sid=XXX</NewX_AVM-DE_UrlSID>
func extractSIDFromString(s string) string {
	// Find "sid=" in the response
	idx := strings.Index(s, "sid=")
	if idx == -1 {
		return ""
	}

	// Skip "sid="
	start := idx + 4

	// Find end of SID (until '<', '&', '"', or space)
	end := start
	for end < len(s) && s[end] != '<' && s[end] != '&' && s[end] != '"' && s[end] != ' ' {
		end++
	}

	if start >= end {
		return ""
	}

	return s[start:end]
}
