package auth

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/http"
	"os"
	"time"
)

// TLSConfig holds TLS configuration options
type TLSConfig struct {
	InsecureSkipVerify bool
	CertFile         string
	KeyFile          string
	CACertFile       string
}

// NewTLSConfig creates a new TLS configuration
func NewTLSConfig(insecureSkipVerify bool) *tls.Config {
	return &tls.Config{
		InsecureSkipVerify: insecureSkipVerify,
		MinVersion:       tls.VersionTLS12,
	}
}

// NewHTTPClient creates an HTTP client with TLS configuration
func NewHTTPClient(tlsConfig *TLSConfig, timeout time.Duration) (*http.Client, error) {
	tlsCfg := NewTLSConfig(tlsConfig.InsecureSkipVerify)

	// Load custom CA cert if provided
	if tlsConfig.CACertFile != "" {
		caCert, err := os.ReadFile(tlsConfig.CACertFile)
		if err != nil {
			return nil, fmt.Errorf("error reading CA cert: %w", err)
		}
		caCertPool := x509.NewCertPool()
		if !caCertPool.AppendCertsFromPEM(caCert) {
			return nil, fmt.Errorf("failed to parse CA cert")
		}
		tlsCfg.RootCAs = caCertPool
	}

	// Load client cert if provided
	if tlsConfig.CertFile != "" && tlsConfig.KeyFile != "" {
		cert, err := tls.LoadX509KeyPair(tlsConfig.CertFile, tlsConfig.KeyFile)
		if err != nil {
			return nil, fmt.Errorf("error loading client cert: %w", err)
		}
		tlsCfg.Certificates = []tls.Certificate{cert}
	}

	transport := &http.Transport{
		TLSClientConfig: tlsCfg,
	}

	return &http.Client{
		Timeout:   timeout,
		Transport: transport,
	}, nil
}

// DefaultTLSConfig returns a default TLS config for Fritz!Box
// Fritz!Box uses self-signed certificates by default
func DefaultTLSConfig() *TLSConfig {
	return &TLSConfig{
		InsecureSkipVerify: true, // Self-signed certs are normal for local Fritz!Box
	}
}
