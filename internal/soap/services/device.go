package services

import (
	"encoding/xml"
	"fmt"

	"github.com/fabito/fritzboxctl/internal/soap"
)

// SOAPCaller is an interface for making SOAP calls
// This allows mocking in tests
type SOAPCaller interface {
	Call(servicePath, soapAction, soapBody string) (string, error)
}

// GetInfoResponse represents the SOAP response for GetInfo
// The XML tags flatten the nested SOAP structure using path syntax
// Body > GetInfoResponse > FieldName
type GetInfoResponse struct {
	XMLName              xml.Name `xml:"Envelope"`
	NewManufacturerName  string   `xml:"Body>GetInfoResponse>NewManufacturerName"`
	NewManufacturerOUI   string   `xml:"Body>GetInfoResponse>NewManufacturerOUI"`
	NewModelName         string   `xml:"Body>GetInfoResponse>NewModelName"`
	NewModelNumber       string   `xml:"Body>GetInfoResponse>NewModelNumber"`
	NewSerialNumber      string   `xml:"Body>GetInfoResponse>NewSerialNumber"`
	NewDescription       string   `xml:"Body>GetInfoResponse>NewDescription"`
	NewProductClass      string   `xml:"Body>GetInfoResponse>NewProductClass"`
	NewSoftwareVersion   string   `xml:"Body>GetInfoResponse>NewSoftwareVersion"`
	NewHardwareVersion   string   `xml:"Body>GetInfoResponse>NewHardwareVersion"`
	NewDeviceLog       string   `xml:"Body>GetInfoResponse>NewDeviceLog"`
}

// GetEventLog retrieves the event log from Fritz!Box
// If soapClient is nil, returns mock data for testing
func GetEventLog(soapClient *soap.Client) (string, error) {
	if soapClient == nil {
		// Return mock data for testing
		return "30.04.26 11:26:46 IPv6 prefix obtained successfully\n19.04.26 18:40:38 IPv6 internet connection established", nil
	}

	info, err := GetDeviceInfo(soapClient)
	if err != nil {
		return "", fmt.Errorf("failed to get device info: %w", err)
	}

	return info.NewDeviceLog, nil
}

// GetDeviceInfo retrieves device information from Fritz!Box
// If soapClient is nil, returns mock data for testing
func GetDeviceInfo(soapClient *soap.Client) (*GetInfoResponse, error) {
	if soapClient == nil {
		// Return mock data for testing
		return &GetInfoResponse{
			NewModelName: "FRITZ!Box 7530",
		}, nil
	}

	// Build SOAP request
	soapBody := `<?xml version="1.0" encoding="utf-8"?>
<s:Envelope s:encodingStyle="http://schemas.xmlsoap.org/soap/encoding/" xmlns:s="http://schemas.xmlsoap.org/soap/envelope/">
    <s:Body>
        <u:GetInfo xmlns:u="urn:dslforum-org:service:DeviceInfo:1">
        </u:GetInfo>
    </s:Body>
</s:Envelope>`

	resp, err := soapClient.Call(
		"/upnp/control/deviceinfo",
		"urn:dslforum-org:service:DeviceInfo:1#GetInfo",
		soapBody,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get device info: %w", err)
	}

	// Clean and parse response using shared function from soap package
	resp = soap.CleanSoapResponse(resp)
	var info GetInfoResponse
	if err := xml.Unmarshal([]byte(resp), &info); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &info, nil
}

// Reboot reboots the Fritz!Box
// If caller is nil, returns nil (mock success)
// SOAP service: urn:dslforum-org:service:DeviceConfig:1
// SOAP action: Reboot
// Service path: /upnp/control/deviceconfig
func Reboot(caller SOAPCaller) error {
	if caller == nil {
		// Mock success for testing
		return nil
	}

	// Build SOAP request for Reboot
	soapBody := `<?xml version="1.0" encoding="utf-8"?>
<s:Envelope s:encodingStyle="http://schemas.xmlsoap.org/soap/encoding/" xmlns:s="http://schemas.xmlsoap.org/soap/envelope/">
    <s:Body>
        <u:Reboot xmlns:u="urn:dslforum-org:service:DeviceConfig:1">
        </u:Reboot>
    </s:Body>
</s:Envelope>`

	_, err := caller.Call(
		"/upnp/control/deviceconfig",
		"urn:dslforum-org:service:DeviceConfig:1#Reboot",
		soapBody,
	)
	if err != nil {
		return fmt.Errorf("failed to reboot: %w", err)
	}

	return nil
}
