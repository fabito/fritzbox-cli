package auth

import (
	"crypto/tls"
	"fmt"
	"log/slog"
	"net/http"
	"time"
)

// Auth provides authentication services for Fritz!Box communication
type Auth struct {
	DigestClient *DigestClient
	SIDManager   *SIDManager
	HTTPClient   *http.Client
}

// NewAuth creates a new Auth instance with default settings
func NewAuth(username, password string, timeout time.Duration, insecureSkipVerify bool) (*Auth, error) {
	slog.Debug("NewAuth called", "username", username)

	// Create TLS config
	tlsCfg := &tls.Config{
		InsecureSkipVerify: insecureSkipVerify,
		MinVersion:         tls.VersionTLS12,
	}

	// Create HTTP client
	httpClient := &http.Client{
		Timeout: timeout,
		Transport: &http.Transport{
			TLSClientConfig: tlsCfg,
		},
	}

	// Create digest client
	digestClient := NewDigestClient(httpClient, username, password)

	// Create auth instance (SIDManager will be set up later when SOAP client is available)
	return &Auth{
		DigestClient: digestClient,
		HTTPClient:   httpClient,
	}, nil
}

// SetSIDManager sets the SID manager (called after SOAP client is created)
func (a *Auth) SetSIDManager(sidManager *SIDManager) {
	a.SIDManager = sidManager
}

// GetSID returns a valid SID (convenience method)
func (a *Auth) GetSID() (string, error) {
	if a.SIDManager == nil {
		return "", fmt.Errorf("SID manager not initialized")
	}
	return a.SIDManager.GetSID()
}

// Login obtains a new SID
func (a *Auth) Login() (string, error) {
	if a.SIDManager == nil {
		return "", fmt.Errorf("SID manager not initialized")
	}
	return a.SIDManager.Login()
}

// Logout clears the current SID
func (a *Auth) Logout() error {
	if a.SIDManager == nil {
		return nil
	}
	return a.SIDManager.Logout()
}

// Do executes an HTTP request with digest authentication
func (a *Auth) Do(req *http.Request) (*http.Response, error) {
	slog.Debug("Auth.Do called", "url", req.URL, "method", req.Method)
	return a.DigestClient.Do(req, "")
}
