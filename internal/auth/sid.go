package auth

import (
	"encoding/xml"
	"fmt"
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
	if s.sid != "" && time.Now().Before(s.expiresAt) {
		return s.sid, nil
	}
	return s.Login()
}

// Login obtains a new SID from the Fritz!Box
func (s *SIDManager) Login() (string, error) {
	soapBody := `<?xml version="1.0" encoding="utf-8"?>
<s:Envelope s:encodingStyle="http://schemas.xmlsoap.org/soap/encoding/"
            xmlns:s="http://schemas.xmlsoap.org/soap/envelope/">
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

	sid, err := extractSID(resp)
	if err != nil {
		return "", fmt.Errorf("failed to parse SID: %w", err)
	}

	s.sid = sid
	s.expiresAt = time.Now().Add(20 * time.Minute)
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

// extractSID extracts SID from SOAP response
func extractSID(resp string) (string, error) {
	type Envelope struct {
		Body struct {
			CreateUrlSIDResponse struct {
				URLSID string `xml:"NewX_AVM-DE_UrlSID"`
			} `xml:"Body>X_AVM-DE_CreateUrlSIDResponse"`
		} `xml:"Body"`
	}

	var env Envelope
	decoder := xml.NewDecoder(strings.NewReader(resp))
	if err := decoder.Decode(&env); err == nil {
		return extractSIDFromURL(env.Body.CreateUrlSIDResponse.URLSID), nil
	}

	return extractSIDFromString(resp)
}

// extractSIDFromString does simple string parsing as fallback
func extractSIDFromString(s string) (string, error) {
	idx := strings.Index(s, "sid=")
	if idx == -1 {
		return "", fmt.Errorf("sid not found in response")
	}

	start := idx + 4
	end := start
	for end < len(s) && s[end] != '&' && s[end] != '<' && s[end] != '"' && s[end] != ' ' {
		end++
	}

	if start >= end {
		return "", fmt.Errorf("invalid sid format")
	}

	return s[start:end], nil
}

func extractSIDFromURL(url string) string {
	parts := strings.Split(url, "sid=")
	if len(parts) < 2 {
		return ""
	}
	sid := parts[1]
	if idx := strings.Index(sid, "&"); idx != -1 {
		sid = sid[:idx]
	}
	return sid
}
