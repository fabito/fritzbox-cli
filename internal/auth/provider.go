package auth

import (
	"fmt"

	"github.com/fabito/fritzboxctl/internal/config"
)

// CredentialProvider supplies credentials from multiple sources
type CredentialProvider struct {
	config   *config.Config
	username string
	password string
}

// NewCredentialProvider creates a new CredentialProvider
// Order of precedence: flags > env vars > config file
func NewCredentialProvider(cfg *config.Config, usernameFlag, passwordFlag string) *CredentialProvider {
	cp := &CredentialProvider{
		config: cfg,
	}

	// Flags take highest precedence
	if usernameFlag != "" {
		cp.username = usernameFlag
	} else if cfg.Username != "" {
		cp.username = cfg.Username
	}

	if passwordFlag != "" {
		cp.password = passwordFlag
	} else if cfg.Password != "" {
		cp.password = cfg.Password
	}

	return cp
}

// Username returns the username
func (c *CredentialProvider) Username() string {
	return c.username
}

// Password returns the password
func (c *CredentialProvider) Password() string {
	return c.password
}

// HasCredentials checks if credentials are available
func (c *CredentialProvider) HasCredentials() bool {
	return c.username != "" && c.password != ""
}

// Validate checks if credentials are properly configured
func (c *CredentialProvider) Validate() error {
	if c.username == "" {
		return fmt.Errorf("username not configured (set FRITZBOX_USERNAME or use --username flag)")
	}
	if c.password == "" {
		return fmt.Errorf("password not configured (set FRITZBOX_PASSWORD or use --password flag)")
	}
	return nil
}

// RedactPassword returns a redacted version of the password for logging
func RedactPassword(password string) string {
	if len(password) <= 2 {
		return "***"
	}
	return password[:2] + "***"
}
