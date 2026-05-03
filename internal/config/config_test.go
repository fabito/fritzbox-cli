package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestDefaultValues(t *testing.T) {
	if DefaultRouterURI != "192.168.178.1" {
		t.Errorf("DefaultRouterURI = %s, want 192.168.178.1", DefaultRouterURI)
	}
	if DefaultTimeout != 10*time.Second {
		t.Errorf("DefaultTimeout = %v, want 10s", DefaultTimeout)
	}
	if DefaultOutput != "text" {
		t.Errorf("DefaultOutput = %s, want 'text'", DefaultOutput)
	}
}

func TestGetConfigFile(t *testing.T) {
	path := GetConfigFile()
	if path == "" {
		t.Fatal("GetConfigFile returned empty string")
	}

	expected := filepath.Join(os.Getenv("HOME"), ".config/fritzboxctl", "config.yaml")
	if path != expected {
		t.Errorf("GetConfigFile() = %s, want %s", path, expected)
	}
}

func TestLoadWithEnvVars(t *testing.T) {
	// Set environment variables
	os.Setenv("FRITZBOX_URI", "192.168.1.1")
	os.Setenv("FRITZBOX_USERNAME", "testuser")
	os.Setenv("FRITZBOX_PASSWORD", "testpass")
	defer func() {
		os.Unsetenv("FRITZBOX_URI")
		os.Unsetenv("FRITZBOX_USERNAME")
		os.Unsetenv("FRITZBOX_PASSWORD")
	}()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	if cfg.RouterURI != "192.168.1.1" {
		t.Errorf("RouterURI = %s, want 192.168.1.1", cfg.RouterURI)
	}
	if cfg.Username != "testuser" {
		t.Errorf("Username = %s, want testuser", cfg.Username)
	}
	if cfg.Password != "testpass" {
		t.Errorf("Password = %s, want testpass", cfg.Password)
	}
}

func TestLoadWithDefaults(t *testing.T) {
	// Clear any existing env vars
	os.Unsetenv("FRITZBOX_URI")
	os.Unsetenv("FRITZBOX_USERNAME")
	os.Unsetenv("FRITZBOX_PASSWORD")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	if cfg.RouterURI != DefaultRouterURI {
		t.Errorf("RouterURI = %s, want default %s", cfg.RouterURI, DefaultRouterURI)
	}
	if cfg.Timeout != DefaultTimeout {
		t.Errorf("Timeout = %v, want default %v", cfg.Timeout, DefaultTimeout)
	}
	if cfg.OutputFormat != DefaultOutput {
		t.Errorf("OutputFormat = %s, want default %s", cfg.OutputFormat, DefaultOutput)
	}
}

func TestSaveAndLoad(t *testing.T) {
	// Create a temp config
	cfg := &Config{
		RouterURI:    "10.0.0.1",
		Username:     "saveuser",
		Password:     "savepass",
		RepeaterURI:  "10.0.0.2",
		RepeaterUser: "repuser",
		RepeaterPass: "reppass",
		Timeout:      30 * time.Second,
		OutputFormat:  "json",
	}

	// Save
	err := cfg.Save()
	if err != nil {
		t.Fatalf("Save() failed: %v", err)
	}
	defer func() {
		// Cleanup
		os.Remove(GetConfigFile())
		os.Remove(filepath.Dir(GetConfigFile()))
	}()

	// Load and verify
	loaded, err := Load()
	if err != nil {
		t.Fatalf("Load() after save failed: %v", err)
	}

	if loaded.RouterURI != cfg.RouterURI {
		t.Errorf("loaded RouterURI = %s, want %s", loaded.RouterURI, cfg.RouterURI)
	}
	if loaded.Username != cfg.Username {
		t.Errorf("loaded Username = %s, want %s", loaded.Username, cfg.Username)
	}
	if loaded.Timeout != cfg.Timeout {
		t.Errorf("loaded Timeout = %v, want %v", loaded.Timeout, cfg.Timeout)
	}
}

func TestConfigPrecedence(t *testing.T) {
	// Test that env vars override config file
	// First save a config file
	cfg := &Config{
		RouterURI: "10.0.0.1",
		Username:  "fileuser",
	}
	cfg.Save()
	defer func() {
		os.Remove(GetConfigFile())
		os.Remove(filepath.Dir(GetConfigFile()))
	}()

	// Now set env var (should override)
	os.Setenv("FRITZBOX_URI", "192.168.1.1")
	defer os.Unsetenv("FRITZBOX_URI")

	loaded, err := Load()
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	// Env var should win
	if loaded.RouterURI != "192.168.1.1" {
		t.Errorf("RouterURI = %s, want 192.168.1.1 (from env)", loaded.RouterURI)
	}
	// File value should still be used for non-env fields
	if loaded.Username != "fileuser" {
		t.Errorf("Username = %s, want fileuser (from file)", loaded.Username)
	}
}

func TestEnvVarAutoDiscovery(t *testing.T) {
	// Test that env vars with FRITZBOX_ prefix are auto-discovered
	os.Setenv("FRITZBOX_URI", "172.16.0.1")
	os.Setenv("FRITZBOX_USERNAME", "envuser")
	os.Setenv("FRITZBOX_OUTPUT", "json")
	defer func() {
		os.Unsetenv("FRITZBOX_URI")
		os.Unsetenv("FRITZBOX_USERNAME")
		os.Unsetenv("FRITZBOX_OUTPUT")
	}()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	if cfg.RouterURI != "172.16.0.1" {
		t.Errorf("RouterURI = %s, want 172.16.0.1", cfg.RouterURI)
	}
	if cfg.Username != "envuser" {
		t.Errorf("Username = %s, want envuser", cfg.Username)
	}
	if cfg.OutputFormat != "json" {
		t.Errorf("OutputFormat = %s, want json", cfg.OutputFormat)
	}
}
