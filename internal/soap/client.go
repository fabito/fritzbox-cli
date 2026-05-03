package soap

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"

	"github.com/fabito/fritzboxctl/internal/auth"
)

// Client is a SOAP client for TR-064 protocol
type Client struct {
	baseURL    string
	authClient *auth.DigestClient
	services   map[string]ServiceDescription
	httpClient *http.Client
}

// NewClient creates a new SOAP client
func NewClient(baseURL string, authClient *auth.DigestClient, httpClient *http.Client) *Client {
	// Ensure baseURL doesn't end with slash
	baseURL = strings.TrimSuffix(baseURL, "/")

	return &Client{
		baseURL:    baseURL,
		authClient: authClient,
		httpClient: httpClient,
		services:   make(map[string]ServiceDescription),
	}
}

// Call executes a SOAP action
func (c *Client) Call(servicePath, soapAction, soapBody string) (string, error) {
	slog.Debug("SOAP Call", "servicePath", servicePath, "action", soapAction)

	// Build the full URL
	url := c.baseURL + servicePath

	// Create the request
	req, err := http.NewRequest("POST", url, bytes.NewBufferString(soapBody))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", `text/xml; charset="utf-8"`)
	req.Header.Set("SoapAction", soapAction)

	// Execute the request with digest auth
	slog.Debug("SOAP Call: executing request")
	resp, err := c.authClient.Do(req, soapBody)
	if err != nil {
		return "", fmt.Errorf("SOAP request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	slog.Debug("SOAP Call: response received", "body_length", len(body))

	// Check for SOAP fault
	bodyStr := string(body)
	if strings.Contains(bodyStr, "soap:Fault") || strings.Contains(bodyStr, "UPnPError") {
		return bodyStr, fmt.Errorf("SOAP fault in response: %s", truncateString(bodyStr, 200))
	}

	return bodyStr, nil
}

// Discover parses tr64desc.xml and igddesc.xml to discover services
func (c *Client) Discover() error {
	// Try tr64desc.xml first
	if err := c.parseDescription("/tr64desc.xml"); err != nil {
		// Try igddesc.xml as fallback
		if err2 := c.parseDescription("/igddesc.xml"); err2 != nil {
			return fmt.Errorf("failed to discover services from both tr64desc.xml and igddesc.xml: %v, %v", err, err2)
		}
	}
	return nil
}

// parseDescription parses a description XML file
func (c *Client) parseDescription(path string) error {
	url := c.baseURL + path

	resp, err := c.httpClient.Get(url)
	if err != nil {
		return fmt.Errorf("failed to fetch %s: %w", path, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read %s: %w", path, err)
	}

	var deviceDesc DeviceDescription
	decoder := xml.NewDecoder(bytes.NewReader(body))
	if err := decoder.Decode(&deviceDesc); err != nil {
		return fmt.Errorf("failed to parse %s: %w", path, err)
	}

	// Extract services from device and sub-devices
	c.extractServices(deviceDesc.Device)

	return nil
}

// extractServices recursively extracts services from a device
func (c *Client) extractServices(device Device) {
	for _, svc := range device.Services {
		key := svc.ServiceType
		c.services[key] = ServiceDescription{
			ServiceType: svc.ServiceType,
			ServiceId:   svc.ServiceId,
			ControlURL:  svc.ControlURL,
			EventSubURL: svc.EventSubURL,
			SCPDURL:     svc.SCPDURL,
		}
	}

	for _, dev := range device.Devices {
		c.extractServices(dev)
	}
}

// GetService returns a service description by type
func (c *Client) GetService(serviceType string) (ServiceDescription, bool) {
	svc, ok := c.services[serviceType]
	return svc, ok
}

// BuildSoapEnvelope builds a SOAP envelope XML string
func BuildSoapEnvelope(serviceType, action string, params map[string]string) (string, error) {
	// Custom XML marshaling to get the exact format Fritz!Box expects
	var buf bytes.Buffer
	buf.WriteString(`<?xml version="1.0" encoding="utf-8"?>`)
	buf.WriteString(`<s:Envelope s:encodingStyle="http://schemas.xmlsoap.org/soap/encoding/" `)
	buf.WriteString(`xmlns:s="http://schemas.xmlsoap.org/soap/envelope/">`)
	buf.WriteString(`<s:Body>`)
	buf.WriteString(fmt.Sprintf(`<u:%s xmlns:u="%s">`, action, serviceType))

	for key, value := range params {
		buf.WriteString(fmt.Sprintf(`<%s>%s</%s>`, key, value, key))
	}

	buf.WriteString(fmt.Sprintf(`</u:%s>`, action))
	buf.WriteString(`</s:Body>`)
	buf.WriteString(`</s:Envelope>`)

	return buf.String(), nil
}

// GetSoapCallFunc returns a function that can be used for SOAP calls
// This is used by SIDManager to make SOAP calls
func (c *Client) GetSoapCallFunc() func(service, action, body string) (string, error) {
	return func(service, action, body string) (string, error) {
		soapAction := action
		// Add service type to action if not already present
		if !strings.Contains(action, "#") {
			// Look up the service to get the full SOAP action
			// For simplicity, just use the action as-is
		}
		return c.Call(service, soapAction, body)
	}
}

// Helper to truncate strings
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
