package auth

import (
	"bytes"
	"crypto/md5"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"sync"
)

// DigestClient implements HTTP Digest Access Authentication (RFC 7616)
// Simplified implementation for Fritz!Box TR-064 protocol
type DigestClient struct {
	client   *http.Client
	username string
	password string
	mu       sync.RWMutex

	// Digest session state
	nonce   string
	opaque  string
	realm    string
	qop      string
	algorithm string
	nc        int
	lastNonce string
}

// NewDigestClient creates a new DigestClient
func NewDigestClient(client *http.Client, username, password string) *DigestClient {
	return &DigestClient{
		client:   client,
		username: username,
		password: password,
		nc:       0,
	}
}

// Do executes an HTTP request with digest authentication
// The body string is used to replay the body on retry after 401
func (d *DigestClient) Do(req *http.Request, body string) (*http.Response, error) {
	slog.Debug("DigestClient.Do: starting request", "method", req.Method, "url", req.URL)
	// Buffer the body for potential retry
	var bodyBytes []byte
	if body != "" {
		bodyBytes = []byte(body)
		req.Body = io.NopCloser(bytes.NewReader(bodyBytes))
		req.ContentLength = int64(len(bodyBytes))
	}

	// First attempt
	slog.Debug("DigestClient.Do: sending first attempt")
	resp, err := d.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	// Check if we got a 401 Unauthorized with WWW-Authenticate header
	if resp.StatusCode == http.StatusUnauthorized {
		slog.Debug("DigestClient.Do: got 401, parsing challenge")
		authHeader := resp.Header.Get("WWW-Authenticate")
		if strings.Contains(authHeader, "Digest") {
			// Parse the digest challenge
			if err := d.parseChallenge(authHeader); err != nil {
				resp.Body.Close()
				return nil, fmt.Errorf("failed to parse digest challenge: %w", err)
			}

			// Close the response body
			resp.Body.Close()

			// Create new request with Authorization header
			newReq := req.Clone(req.Context())
			if bodyBytes != nil {
				// Re-create the body for the retry
				newReq.Body = io.NopCloser(bytes.NewReader(bodyBytes))
				newReq.ContentLength = int64(len(bodyBytes))
			}

			auth := d.generateAuthorization(req)
			newReq.Header.Set("Authorization", auth)

			// Retry with authorization
			slog.Debug("DigestClient.Do: retrying with auth header")
			resp, err = d.client.Do(newReq)
			if err != nil {
				return nil, fmt.Errorf("authorized request failed: %w", err)
			}
		}
	}

	return resp, nil
}

// parseChallenge parses the WWW-Authenticate header
func (d *DigestClient) parseChallenge(header string) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	// Remove "Digest " prefix
	header = strings.TrimPrefix(header, "Digest ")
	header = strings.TrimSpace(header)

	// Parse key-value pairs
	pairs := parseAuthHeader(header)
	d.realm = pairs["realm"]
	d.nonce = pairs["nonce"]
	d.opaque = pairs["opaque"]
	d.qop = pairs["qop"]
	d.algorithm = pairs["algorithm"]
	if d.algorithm == "" {
		d.algorithm = "MD5" // Default
	}

	return nil
}

// generateAuthorization creates the Authorization header value
func (d *DigestClient) generateAuthorization(req *http.Request) string {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.nc++
	cnonce := generateCNonce()
	nonceCount := fmt.Sprintf("%08x", d.nc)

	// Calculate hashes
	ha1 := d.h(d.username + ":" + d.realm + ":" + d.password)
	var ha2 string
	if req.Method == http.MethodGet || req.Method == http.MethodPost {
		// For POST with SOAP, body is typically empty or we use empty body hash
		ha2 = d.h(req.Method + ":" + req.URL.RequestURI())
	}

	var response string
	if d.qop != "" {
		response = d.h(ha1 + ":" + d.nonce + ":" + nonceCount + ":" + cnonce + ":" + d.qop + ":" + ha2)
	} else {
		response = d.h(ha1 + ":" + d.nonce + ":" + ha2)
	}

	// Build authorization header
	parts := []string{
		fmt.Sprintf(`username="%s"`, d.username),
		fmt.Sprintf(`realm="%s"`, d.realm),
		fmt.Sprintf(`nonce="%s"`, d.nonce),
		fmt.Sprintf(`uri="%s"`, req.URL.RequestURI()),
		fmt.Sprintf(`response="%s"`, response),
	}
	if d.opaque != "" {
		parts = append(parts, fmt.Sprintf(`opaque="%s"`, d.opaque))
	}
	if d.qop != "" {
		parts = append(parts, fmt.Sprintf(`qop="%s"`, d.qop))
		parts = append(parts, fmt.Sprintf(`nc=%s`, nonceCount))
		parts = append(parts, fmt.Sprintf(`cnonce="%s"`, cnonce))
	}
	if d.algorithm != "" && d.algorithm != "MD5" {
		parts = append(parts, fmt.Sprintf(`algorithm=%s`, d.algorithm))
	}

	return "Digest " + strings.Join(parts, ", ")
}

// h creates an MD5 hash (Fritz!Box typically uses MD5)
func (d *DigestClient) h(data string) string {
	if d.algorithm == "SHA-256" {
		// SHA-256 support if needed
		// For now, Fritz!Box uses MD5
	}
	h := md5.New()
	io.WriteString(h, data)
	return hex.EncodeToString(h.Sum(nil))
}

// generateCNonce generates a random client nonce
func generateCNonce() string {
	b := make([]byte, 8)
	rand.Read(b)
	return hex.EncodeToString(b)
}

// parseAuthHeader parses a key-value auth header
func parseAuthHeader(header string) map[string]string {
	pairs := make(map[string]string)
	parts := strings.Split(header, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		kv := strings.SplitN(part, "=", 2)
		if len(kv) == 2 {
			key := strings.TrimSpace(kv[0])
			value := strings.TrimSpace(kv[1])
			value = strings.Trim(value, "\"")
			pairs[key] = value
		}
	}
	return pairs
}
