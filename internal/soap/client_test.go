package soap

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/fabito/fritzboxctl/internal/auth"
)

func TestBuildSoapEnvelope(t *testing.T) {
	tests := []struct {
		name        string
		serviceType string
		action      string
		params      map[string]string
		contains    []string
	}{
		{
			name:        "Simple GetInfo",
			serviceType: "urn:dslforum-org:service:DeviceInfo:1",
			action:      "GetInfo",
			params:      map[string]string{},
			contains:    []string{"GetInfo", "DeviceInfo:1"},
		},
		{
			name:        "GetInfo with params",
			serviceType: "urn:dslforum-org:service:WLANConfiguration:1",
			action:      "SetEnable",
			params:      map[string]string{"NewEnable": "1"},
			contains:    []string{"SetEnable", "WLANConfiguration:1", "NewEnable", "1"},
		},
		{
			name:        "Reboot",
			serviceType: "urn:dslforum-org:service:DeviceConfig:1",
			action:      "Reboot",
			params:      map[string]string{},
			contains:    []string{"Reboot", "DeviceConfig:1"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := BuildSoapEnvelope(tt.serviceType, tt.action, tt.params)
			if err != nil {
				t.Fatalf("BuildSoapEnvelope failed: %v", err)
			}

			// Check that result contains expected strings
			for _, expected := range tt.contains {
				if !strings.Contains(result, expected) {
					t.Errorf("expected result to contain '%s', got:\n%s", expected, result)
				}
			}

			// Check XML structure
			if !strings.HasPrefix(result, `<?xml version="1.0"`) {
				t.Errorf("expected XML header, got: %s", result[:50])
			}
			if !strings.Contains(result, "</s:Envelope>") {
				t.Errorf("expected closing Envelope tag")
			}
		})
	}
}

func TestBuildSoapEnvelopeMultipleParams(t *testing.T) {
	params := map[string]string{
		"NewIndex":     "0",
		"NewEnable":    "1",
		"NewSomeValue": "test",
	}

	result, err := BuildSoapEnvelope(
		"urn:dslforum-org:service:X_AVM-DE_TAM:1",
		"SetEnable",
		params,
	)
	if err != nil {
		t.Fatalf("BuildSoapEnvelope failed: %v", err)
	}

	// Check all params are present
	for key, value := range params {
		expected := fmt.Sprintf("<%s>%s</%s>", key, value, key)
		if !strings.Contains(result, expected) {
			t.Errorf("expected param '%s' in result, got:\n%s", expected, result)
		}
	}
}

func TestNewClient(t *testing.T) {
	httpClient := &http.Client{}
	digestClient := auth.NewDigestClient(httpClient, "testuser", "testpass")

	client := NewClient("http://192.168.178.1:49000", digestClient, httpClient)

	if client == nil {
		t.Fatal("NewClient returned nil")
	}
	if client.baseURL != "http://192.168.178.1:49000" {
		t.Errorf("Expected baseURL 'http://192.168.178.1:49000', got '%s'", client.baseURL)
	}
	if client.authClient == nil {
		t.Error("Expected authClient to be set")
	}
	if client.httpClient == nil {
		t.Error("Expected httpClient to be set")
	}
}

func TestNewClientTrimsTrailingSlash(t *testing.T) {
	httpClient := &http.Client{}
	digestClient := auth.NewDigestClient(httpClient, "testuser", "testpass")

	client := NewClient("http://192.168.178.1:49000/", digestClient, httpClient)

	if client.baseURL != "http://192.168.178.1:49000" {
		t.Errorf("Expected baseURL without trailing slash, got '%s'", client.baseURL)
	}
}

func TestClientCall(t *testing.T) {
	// Create a test server that simulates Fritz!Box responses
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request method
		if r.Method != "POST" {
			t.Errorf("Expected POST method, got %s", r.Method)
		}

		// Verify content type header
		contentType := r.Header.Get("Content-Type")
		if !strings.Contains(contentType, "text/xml") {
			t.Errorf("Expected Content-Type text/xml, got %s", contentType)
		}

		// Verify SOAP action header
		soapAction := r.Header.Get("SoapAction")
		if soapAction == "" {
			t.Error("Expected SoapAction header to be set")
		}

		// Return a mock SOAP response
		w.Header().Set("Content-Type", "text/xml")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`<?xml version="1.0"?>
<s:Envelope xmlns:s="http://schemas.xmlsoap.org/soap/envelope/">
<s:Body>
<u:GetInfoResponse xmlns:u="urn:dslforum-org:service:DeviceInfo:1">
<NewManufacturerName>AVM</NewManufacturerName>
</u:GetInfoResponse>
</s:Body>
</s:Envelope>`))
	}))
	defer server.Close()

	// Create client pointing to test server
	httpClient := server.Client()
	digestClient := auth.NewDigestClient(httpClient, "testuser", "testpass")
	client := NewClient(server.URL, digestClient, httpClient)

	// Make a SOAP call
	soapBody := `<?xml version="1.0" encoding="utf-8"?>
<s:Envelope xmlns:s="http://schemas.xmlsoap.org/soap/envelope/">
<s:Body><u:GetInfo xmlns:u="urn:dslforum-org:service:DeviceInfo:1"></u:GetInfo></s:Body>
</s:Envelope>`

	resp, err := client.Call(
		"/upnp/control/deviceinfo",
		"urn:dslforum-org:service:DeviceInfo:1#GetInfo",
		soapBody,
	)

	if err != nil {
		t.Fatalf("Call failed: %v", err)
	}

	if !strings.Contains(resp, "AVM") {
		t.Errorf("Expected response to contain 'AVM', got: %s", resp)
	}
}

func TestClientCallSOAPFault(t *testing.T) {
	// Create a test server that returns a SOAP fault
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/xml")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`<?xml version="1.0"?>
<s:Envelope xmlns:s="http://schemas.xmlsoap.org/soap/envelope/">
<s:Body>
<s:Fault>
<faultcode>s:Client</faultcode>
<faultstring>UPnPError</faultstring>
</s:Fault>
</s:Body>
</s:Envelope>`))
	}))
	defer server.Close()

	httpClient := server.Client()
	digestClient := auth.NewDigestClient(httpClient, "testuser", "testpass")
	client := NewClient(server.URL, digestClient, httpClient)

	soapBody := `<?xml version="1.0"?><s:Envelope xmlns:s="http://schemas.xmlsoap.org/soap/envelope/"><s:Body></s:Body></s:Envelope>`

	_, err := client.Call("/test", "TestAction", soapBody)

	if err == nil {
		t.Error("Expected error for SOAP fault response")
	}
	if !strings.Contains(err.Error(), "SOAP fault") {
		t.Errorf("Expected SOAP fault error, got: %v", err)
	}
}

func TestClientGetService(t *testing.T) {
	httpClient := &http.Client{}
	digestClient := auth.NewDigestClient(httpClient, "testuser", "testpass")
	client := NewClient("http://192.168.178.1:49000", digestClient, httpClient)

	// Initially no services
	_, found := client.GetService("urn:dslforum-org:service:DeviceInfo:1")
	if found {
		t.Error("Expected service not to be found initially")
	}

	// Manually add a service for testing
	client.services["urn:dslforum-org:service:DeviceInfo:1"] = ServiceDescription{
		ServiceType: "urn:dslforum-org:service:DeviceInfo:1",
		ServiceId:   "urn:DeviceInfo-com:serviceId:DeviceInfo1",
		ControlURL:  "/upnp/control/deviceinfo",
	}

	svc, found := client.GetService("urn:dslforum-org:service:DeviceInfo:1")
	if !found {
		t.Error("Expected service to be found after adding")
	}
	if svc.ControlURL != "/upnp/control/deviceinfo" {
		t.Errorf("Expected ControlURL '/upnp/control/deviceinfo', got '%s'", svc.ControlURL)
	}
}

func TestClientExtractServices(t *testing.T) {
	httpClient := &http.Client{}
	digestClient := auth.NewDigestClient(httpClient, "testuser", "testpass")
	client := NewClient("http://192.168.178.1:49000", digestClient, httpClient)

	// Create a device with services
	device := Device{
		DeviceType:   "urn:dslforum-org:device:InternetGatewayDevice:1",
		FriendlyName: "FRITZ!Box 7530",
		Services: []Service{
			{
				ServiceType: "urn:dslforum-org:service:DeviceInfo:1",
				ServiceId:   "urn:DeviceInfo-com:serviceId:DeviceInfo1",
				ControlURL:  "/upnp/control/deviceinfo",
			},
			{
				ServiceType: "urn:dslforum-org:service:DeviceConfig:1",
				ServiceId:   "urn:DeviceConfig-com:serviceId:DeviceConfig1",
				ControlURL:  "/upnp/control/deviceconfig",
			},
		},
		Devices: []Device{
			{
				DeviceType: "urn:dslforum-org:device:LANDevice:1",
				Services: []Service{
					{
						ServiceType: "urn:dslforum-org:service:Hosts:1",
						ServiceId:   "urn:Hosts-com:serviceId:Hosts1",
						ControlURL:  "/upnp/control/hosts",
					},
				},
			},
		},
	}

	client.extractServices(device)

	// Verify all services were extracted including nested ones
	if len(client.services) != 3 {
		t.Errorf("Expected 3 services, got %d", len(client.services))
	}

	// Check parent device service
	svc, found := client.GetService("urn:dslforum-org:service:DeviceInfo:1")
	if !found {
		t.Error("Expected DeviceInfo service to be extracted")
	}
	if svc.ControlURL != "/upnp/control/deviceinfo" {
		t.Errorf("Expected ControlURL '/upnp/control/deviceinfo', got '%s'", svc.ControlURL)
	}

	// Check nested device service
	svc, found = client.GetService("urn:dslforum-org:service:Hosts:1")
	if !found {
		t.Error("Expected Hosts service to be extracted from nested device")
	}
}

func TestClientDiscover(t *testing.T) {
	// Create a test server that returns tr64desc.xml
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/tr64desc.xml" {
			w.Header().Set("Content-Type", "text/xml")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`<?xml version="1.0"?>
<root xmlns="urn:dslforum-org:device-1-0">
<device>
<deviceType>urn:dslforum-org:device:InternetGatewayDevice:1</deviceType>
<friendlyName>FRITZ!Box 7530</friendlyName>
<serviceList>
<service>
<serviceType>urn:dslforum-org:service:DeviceInfo:1</serviceType>
<serviceId>urn:DeviceInfo-com:serviceId:DeviceInfo1</serviceId>
<controlURL>/upnp/control/deviceinfo</controlURL>
</service>
</serviceList>
</device>
</root>`))
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	httpClient := server.Client()
	digestClient := auth.NewDigestClient(httpClient, "testuser", "testpass")
	client := NewClient(server.URL, digestClient, httpClient)

	err := client.Discover()
	if err != nil {
		t.Fatalf("Discover failed: %v", err)
	}

	// Verify service was discovered
	svc, found := client.GetService("urn:dslforum-org:service:DeviceInfo:1")
	if !found {
		t.Error("Expected DeviceInfo service to be discovered")
	}
	if svc.ControlURL != "/upnp/control/deviceinfo" {
		t.Errorf("Expected ControlURL '/upnp/control/deviceinfo', got '%s'", svc.ControlURL)
	}
}

func TestGetSoapCallFunc(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/xml")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`<?xml version="1.0"?><Envelope><Body>OK</Body></Envelope>`))
	}))
	defer server.Close()

	httpClient := server.Client()
	digestClient := auth.NewDigestClient(httpClient, "testuser", "testpass")
	client := NewClient(server.URL, digestClient, httpClient)

	callFunc := client.GetSoapCallFunc()

	if callFunc == nil {
		t.Fatal("GetSoapCallFunc returned nil")
	}

	// Test that the function works
	resp, err := callFunc("/test", "TestAction", "<body/>")
	if err != nil {
		t.Fatalf("SoapCallFunc failed: %v", err)
	}
	if !strings.Contains(resp, "OK") {
		t.Errorf("Expected response to contain 'OK', got: %s", resp)
	}
}

func TestCleanSoapResponse(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Remove encodingStyle",
			input:    `<Envelope encodingStyle="http://schemas.xmlsoap.org/soap/encoding/">`,
			expected: `<Envelope>`,
		},
		{
			name:     "Fix spaces before >",
			input:    `<tag >`,
			expected: `<tag>`,
		},
		{
			name:     "Remove soap namespace",
			input:    `<s:Envelope><s:Body></s:Body></s:Envelope>`,
			expected: `<Envelope><Body></Body></Envelope>`,
		},
		{
			name:     "Remove u: namespace",
			input:    `<u:GetInfoResponse></u:GetInfoResponse>`,
			expected: `<GetInfoResponse></GetInfoResponse>`,
		},
		{
			name:     "Remove xmlns:s attribute",
			input:    `<Envelope xmlns:s="http://schemas.xmlsoap.org/soap/envelope/">`,
			expected: `<Envelope>`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CleanSoapResponse(tt.input)
			if result != tt.expected {
				t.Errorf("CleanSoapResponse(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestTruncateString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		maxLen   int
		expected string
	}{
		{"short", "short", 10, "short"},
		{"exactlength", "exactlength", 11, "exactlength"},
		{"truncate", "longstringthatneedstruncating", 10, "longstring..."},
		{"empty", "", 5, ""},
		{"boundary", "abc", 3, "abc"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := truncateString(tt.input, tt.maxLen)
			if result != tt.expected {
				t.Errorf("truncateString(%q, %d) = %q, want %q", tt.input, tt.maxLen, result, tt.expected)
			}
		})
	}
}
