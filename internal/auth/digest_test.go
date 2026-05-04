package auth

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
)

func TestNewDigestClient(t *testing.T) {
	client := NewDigestClient(&http.Client{}, "user", "pass")
	if client == nil {
		t.Fatal("NewDigestClient returned nil")
	}
	if client.username != "user" {
		t.Errorf("expected username 'user', got '%s'", client.username)
	}
	if client.password != "pass" {
		t.Errorf("expected password 'pass', got '%s'", client.password)
	}
	if client.nc != 0 {
		t.Errorf("expected nc to be 0, got %d", client.nc)
	}
}

func TestParseChallenge(t *testing.T) {
	client := &DigestClient{
		username: "testuser",
		password: "testpass",
		nc:       0,
	}

	header := `Digest realm="testrealm", nonce="testnonce", opaque="testopaque", qop="auth", algorithm="MD5"`

	err := client.parseChallenge(header)
	if err != nil {
		t.Fatalf("parseChallenge failed: %v", err)
	}

	if client.realm != "testrealm" {
		t.Errorf("expected realm 'testrealm', got '%s'", client.realm)
	}
	if client.nonce != "testnonce" {
		t.Errorf("expected nonce 'testnonce', got '%s'", client.nonce)
	}
	if client.opaque != "testopaque" {
		t.Errorf("expected opaque 'testopaque', got '%s'", client.opaque)
	}
	if client.qop != "auth" {
		t.Errorf("expected qop 'auth', got '%s'", client.qop)
	}
	if client.algorithm != "MD5" {
		t.Errorf("expected algorithm 'MD5', got '%s'", client.algorithm)
	}
}

func TestParseChallengeNoAlgorithm(t *testing.T) {
	client := &DigestClient{
		username: "testuser",
		password: "testpass",
	}

	header := `Digest realm="testrealm", nonce="testnonce"`

	err := client.parseChallenge(header)
	if err != nil {
		t.Fatalf("parseChallenge failed: %v", err)
	}

	if client.algorithm != "MD5" {
		t.Errorf("expected default algorithm 'MD5', got '%s'", client.algorithm)
	}
}

func TestH(t *testing.T) {
	client := &DigestClient{
		algorithm: "MD5",
	}

	// Test MD5 hash
	input := "teststring"
	h := md5.New()
	h.Write([]byte(input))
	expected := hex.EncodeToString(h.Sum(nil))
	result := client.h(input)

	if result != expected {
		t.Errorf("expected hash '%s', got '%s'", expected, result)
	}
}

func TestGenerateCNonce(t *testing.T) {
	nonce1 := generateCNonce()
	nonce2 := generateCNonce()

	if len(nonce1) == 0 {
		t.Error("generated cnonce is empty")
	}
	if len(nonce2) == 0 {
		t.Error("generated cnonce is empty")
	}

	// Nonces should be different (very high probability)
	if nonce1 == nonce2 {
		t.Error("two generated nonces are identical (very unlikely for random)")
	}
}

func TestParseAuthHeader(t *testing.T) {
	header := `realm="myrealm", nonce="mynonce", opaque="myopaque", qop="auth"`

	pairs := parseAuthHeader(header)

	if pairs["realm"] != "myrealm" {
		t.Errorf("expected realm 'myrealm', got '%s'", pairs["realm"])
	}
	if pairs["nonce"] != "mynonce" {
		t.Errorf("expected nonce 'mynonce', got '%s'", pairs["nonce"])
	}
	if pairs["opaque"] != "myopaque" {
		t.Errorf("expected opaque 'myopaque', got '%s'", pairs["opaque"])
	}
	if pairs["qop"] != "auth" {
		t.Errorf("expected qop 'auth', got '%s'", pairs["qop"])
	}
}

func TestGenerateAuthorization(t *testing.T) {
	client := &DigestClient{
		username:  "testuser",
		password:  "testpass",
		realm:     "testrealm",
		nonce:     "testnonce",
		opaque:    "testopaque",
		qop:       "auth",
		algorithm: "MD5",
		nc:        1,
	}

	req := httptest.NewRequest("POST", "http://192.168.178.1:49000/upnp/control/deviceinfo", nil)

	auth := client.generateAuthorization(req)

	// Check that required fields are present
	if !strings.Contains(auth, `username="testuser"`) {
		t.Error("Authorization missing username")
	}
	if !strings.Contains(auth, `realm="testrealm"`) {
		t.Error("Authorization missing realm")
	}
	if !strings.Contains(auth, `nonce="testnonce"`) {
		t.Error("Authorization missing nonce")
	}
	if !strings.Contains(auth, `uri="/upnp/control/deviceinfo"`) {
		t.Error("Authorization missing uri")
	}
	if !strings.Contains(auth, "response=") {
		t.Error("Authorization missing response")
	}
	if !strings.Contains(auth, `opaque="testopaque"`) {
		t.Error("Authorization missing opaque")
	}
	if !strings.Contains(auth, "qop=") {
		t.Error("Authorization missing qop")
	}
}

func TestDigestClientConcurrentAccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("WWW-Authenticate", `Digest realm="test", nonce="testnonce", qop="auth"`)
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer server.Close()

	client := NewDigestClient(server.Client(), "user", "pass")

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			req, _ := http.NewRequest("GET", server.URL, nil)
			_, err := client.Do(req, "")
			if err != nil {
				t.Errorf("Do failed: %v", err)
			}
		}()
	}
	wg.Wait()
}

func TestGenerateAuthorizationIncrementNc(t *testing.T) {
	client := &DigestClient{
		username:  "testuser",
		password:  "testpass",
		realm:     "testrealm",
		nonce:     "testnonce",
		qop:       "auth",
		algorithm: "MD5",
		nc:        0,
	}

	req := httptest.NewRequest("POST", "http://test/uri", nil)

	// First call
	auth1 := client.generateAuthorization(req)
	// Second call
	auth2 := client.generateAuthorization(req)

	// Extract nc values
	nc1 := extractNc(auth1)
	nc2 := extractNc(auth2)

	if nc1 == "" || nc2 == "" {
		t.Fatal("could not extract nc values")
	}
	if nc1 == nc2 {
		t.Error("nc should increment between calls")
	}
}

func extractNc(authHeader string) string {
	parts := strings.Split(authHeader, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if strings.HasPrefix(part, "nc=") {
			return strings.TrimPrefix(part, "nc=")
		}
	}
	return ""
}

// Test helper to verify HA1 calculation
func TestHA1Calculation(t *testing.T) {
	client := &DigestClient{
		username:  "testuser",
		password:  "testpass",
		realm:     "testrealm",
		algorithm: "MD5",
	}

	// HA1 = MD5(username:realm:password)
	h := md5.New()
	h.Write([]byte("testuser:testrealm:testpass"))
	expectedHA1 := hex.EncodeToString(h.Sum(nil))
	actualHA1 := client.h(fmt.Sprintf("%s:%s:%s", client.username, client.realm, client.password))

	if actualHA1 != expectedHA1 {
		t.Errorf("HA1 mismatch: expected %s, got %s", expectedHA1, actualHA1)
	}
}
