package services

import (
	"encoding/xml"
	"fmt"

	"github.com/fabito/fritzboxctl/internal/soap"
)
// soapCaller interface for mocking in tests
type soapCaller interface {
	Call(servicePath, action, body string) (string, error)
}


// WLANStatusResponse represents the SOAP response for GetInfo on WLANConfiguration
// The XML tags flatten the nested SOAP structure using path syntax
type WLANStatusResponse struct {
	XMLName            xml.Name `xml:"Envelope"`
	NewEnable         string   `xml:"Body>GetInfoResponse>NewEnable"`
	NewSSID           string   `xml:"Body>GetInfoResponse>NewSSID"`
	NewBeaconType     string   `xml:"Body>GetInfoResponse>NewBeaconType"`
	NewChannel        string   `xml:"Body>GetInfoResponse>NewChannel"`
	NewMaxBitRate     string   `xml:"Body>GetInfoResponse>NewMaxBitRate"`
	NewMACAddress     string   `xml:"Body>GetInfoResponse>NewMACAddress"`
	NewBSSID          string   `xml:"Body>GetInfoResponse>NewBSSID"`
}

// WLANStatsResponse represents the SOAP response for GetStatistics
// The XML tags flatten the nested SOAP structure using path syntax
type WLANStatsResponse struct {
	XMLName                  xml.Name `xml:"Envelope"`
	NewTotalPacketsSent      string   `xml:"Body>GetStatisticsResponse>NewTotalPacketsSent"`
	NewTotalPacketsReceived  string   `xml:"Body>GetStatisticsResponse>NewTotalPacketsReceived"`
	NewPacketErrorsReceived  string   `xml:"Body>GetStatisticsResponse>NewPacketErrorsReceived"`
	NewPacketErrorsSent      string   `xml:"Body>GetStatisticsResponse>NewPacketErrorsSent"`
	NewTotalBytesSent        string   `xml:"Body>GetStatisticsResponse>NewTotalBytesSent"`
	NewTotalBytesReceived    string   `xml:"Body>GetStatisticsResponse>NewTotalBytesReceived"`
	NewErrorsReceived        string   `xml:"Body>GetStatisticsResponse>NewErrorsReceived"`
	NewErrorsSent            string   `xml:"Body>GetStatisticsResponse>NewErrorsSent"`
	NewUnicastPacketsSent    string   `xml:"Body>GetStatisticsResponse>NewUnicastPacketsSent"`
	NewUnicastPacketsReceived string  `xml:"Body>GetStatisticsResponse>NewUnicastPacketsReceived"`
	NewMulticastPacketsSent  string   `xml:"Body>GetStatisticsResponse>NewMulticastPacketsSent"`
	NewMulticastPacketsReceived string `xml:"Body>GetStatisticsResponse>NewMulticastPacketsReceived"`
	NewBroadcastPacketsSent  string   `xml:"Body>GetStatisticsResponse>NewBroadcastPacketsSent"`
	NewBroadcastPacketsReceived string `xml:"Body>GetStatisticsResponse>NewBroadcastPacketsReceived"`
}

// wlanServiceInfo maps band to service path and type
var wlanServiceInfo = map[int]struct {
	Path string
	Type string
}{
	1: {Path: "/upnp/control/wlanconfig1", Type: "urn:dslforum-org:service:WLANConfiguration:1"},
	2: {Path: "/upnp/control/wlanconfig2", Type: "urn:dslforum-org:service:WLANConfiguration:2"},
	3: {Path: "/upnp/control/wlanconfig3", Type: "urn:dslforum-org:service:WLANConfiguration:3"},
	4: {Path: "/upnp/control/wlanconfig4", Type: "urn:dslforum-org:service:WLANConfiguration:4"},
}

// GetWLANStats retrieves WLAN statistics from Fritz!Box
// If soapClient is nil, returns mock data for testing
func GetWLANStats(soapClient *soap.Client, band int) (*WLANStatsResponse, error) {
	info, ok := wlanServiceInfo[band]
	if !ok {
		return nil, fmt.Errorf("invalid band: %d (use 1-4)", band)
	}

	if soapClient == nil {
		// Return mock data for testing
		return &WLANStatsResponse{
			NewTotalPacketsSent:     "12345",
			NewTotalPacketsReceived: "54321",
		}, nil
	}

	// Build SOAP request
	soapBody := fmt.Sprintf(`<?xml version="1.0" encoding="utf-8"?>
<s:Envelope s:encodingStyle="http://schemas.xmlsoap.org/soap/encoding/" xmlns:s="http://schemas.xmlsoap.org/soap/envelope/">
    <s:Body>
        <u:GetStatistics xmlns:u="%s">
        </u:GetStatistics>
    </s:Body>
</s:Envelope>`, info.Type)

	resp, err := soapClient.Call(
		info.Path,
		info.Type+"#GetStatistics",
		soapBody,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get WLAN statistics: %w", err)
	}

	// Clean and parse response using shared function
	resp = soap.CleanSoapResponse(resp)
	var stats WLANStatsResponse
	if err := xml.Unmarshal([]byte(resp), &stats); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &stats, nil
}

// GetWLANStatus retrieves WLAN status from Fritz!Box
// If soapClient is nil, returns mock data for testing
func GetWLANStatus(soapClient *soap.Client, band int) (*WLANStatusResponse, error) {
	info, ok := wlanServiceInfo[band]
	if !ok {
		return nil, fmt.Errorf("invalid band: %d (use 1-4)", band)
	}

	if soapClient == nil {
		// Return mock data for testing
		return &WLANStatusResponse{
			NewEnable:     "1",
			NewSSID:       "TestSSID",
			NewBeaconType: "WPA2",
			NewChannel:    "6",
		}, nil
	}

	// Build SOAP request for GetInfo
	soapBody := fmt.Sprintf(`<?xml version="1.0" encoding="utf-8"?>
<s:Envelope s:encodingStyle="http://schemas.xmlsoap.org/soap/encoding/" xmlns:s="http://schemas.xmlsoap.org/soap/envelope/">
    <s:Body>
        <u:GetInfo xmlns:u="%s">
        </u:GetInfo>
    </s:Body>
</s:Envelope>`, info.Type)

	resp, err := soapClient.Call(
		info.Path,
		info.Type+"#GetInfo",
		soapBody,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get WLAN status: %w", err)
	}

	// Clean and parse response using shared function
	resp = soap.CleanSoapResponse(resp)
	var status WLANStatusResponse
	if err := xml.Unmarshal([]byte(resp), &status); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &status, nil
}

// SetWLANEnabled enables or disables WLAN for the specified band
// If soapClient is nil, returns nil (mock success for testing)
func SetWLANEnabled(band int, enabled bool, soapClient *soap.Client) error {
	info, ok := wlanServiceInfo[band]
	if !ok {
		return fmt.Errorf("invalid band: %d (use 1-4)", band)
	}

	// Return mock success if client is nil
	if soapClient == nil {
		return nil
	}

	// Convert enabled to NewEnable parameter
	newEnable := "0"
	if enabled {
		newEnable = "1"
	}

	// Build SOAP request for SetEnable
	soapBody := fmt.Sprintf(`<?xml version="1.0" encoding="utf-8"?>
<s:Envelope s:encodingStyle="http://schemas.xmlsoap.org/soap/encoding/" xmlns:s="http://schemas.xmlsoap.org/soap/envelope/">
    <s:Body>
        <u:SetEnable xmlns:u="%s">
            <NewEnable>%s</NewEnable>
        </u:SetEnable>
    </s:Body>
</s:Envelope>`, info.Type, newEnable)

	_, err := soapClient.Call(
		info.Path,
		info.Type+"#SetEnable",
		soapBody,
	)
	if err != nil {
		return fmt.Errorf("failed to set WLAN enabled: %w", err)
	}

	return nil
}

// GetWLANChannel retrieves the WLAN channel for the specified band
// If soapClient is nil, returns mock data for testing
func GetWLANChannel(soapClient soapCaller, band int) (string, error) {
	info, ok := wlanServiceInfo[band]
	if !ok {
		return "", fmt.Errorf("invalid band: %d (use 1-4)", band)
	}

	if soapClient == nil {
		// Return mock data for testing
		return "6", nil
	}

	// Build SOAP request for GetInfo
	soapBody := fmt.Sprintf(`<?xml version="1.0" encoding="utf-8"?>
<s:Envelope s:encodingStyle="http://schemas.xmlsoap.org/soap/encoding/" xmlns:s="http://schemas.xmlsoap.org/soap/envelope/">
    <s:Body>
        <u:GetInfo xmlns:u="%s">
        </u:GetInfo>
    </s:Body>
</s:Envelope>`, info.Type)

	resp, err := soapClient.Call(
		info.Path,
		info.Type+"#GetInfo",
		soapBody,
	)
	if err != nil {
		return "", fmt.Errorf("failed to get WLAN info: %w", err)
	}

	// Clean and parse response
	resp = soap.CleanSoapResponse(resp)
	var result struct {
		XMLName xml.Name `xml:"Envelope"`
		NewChannel string   `xml:"Body>GetInfoResponse>NewChannel"`
	}
	if err := xml.Unmarshal([]byte(resp), &result); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	return result.NewChannel, nil
}

// GetWLANQRCode returns the QR code string for WLAN connection
// QR format: WIFI:T:WPA;S:<ssid>;P:<password>;;
func GetWLANQRCode(soapClient soapCaller, band int) (string, error) {
	// Get service info for the band
	info, ok := wlanServiceInfo[band]
	if !ok {
		return "", fmt.Errorf("invalid band: %d (use 1-4)", band)
	}

	// If soapClient is nil, return mock data for testing
	if soapClient == nil {
		return "WIFI:T:WPA;S:TestSSID;P:TestPass;;", nil
	}

	// Step 1: Get SSID from GetInfo
	soapBody1 := fmt.Sprintf(`<?xml version="1.0" encoding="utf-8"?>
<s:Envelope s:encodingStyle="http://schemas.xmlsoap.org/soap/encoding/" xmlns:s="http://schemas.xmlsoap.org/soap/envelope/">
    <s:Body>
        <u:GetInfo xmlns:u="%s">
        </u:GetInfo>
    </s:Body>
</s:Envelope>`, info.Type)

	resp1, err := soapClient.Call(
		info.Path,
		info.Type+"#GetInfo",
		soapBody1,
	)
	if err != nil {
		return "", fmt.Errorf("failed to get WLAN info: %w", err)
	}

	// Clean and parse response for SSID
	resp1 = soap.CleanSoapResponse(resp1)
	var infoResult struct {
		XMLName  xml.Name `xml:"Envelope"`
		NewSSID string   `xml:"Body>GetInfoResponse>NewSSID"`
	}
	if err := xml.Unmarshal([]byte(resp1), &infoResult); err != nil {
		return "", fmt.Errorf("failed to parse GetInfo response: %w", err)
	}

	// Step 2: Get KeyPassphrase from GetSecurityKeys
	soapBody2 := fmt.Sprintf(`<?xml version="1.0" encoding="utf-8"?>
<s:Envelope s:encodingStyle="http://schemas.xmlsoap.org/soap/encoding/" xmlns:s="http://schemas.xmlsoap.org/soap/envelope/">
    <s:Body>
        <u:GetSecurityKeys xmlns:u="%s">
        </u:GetSecurityKeys>
    </s:Body>
</s:Envelope>`, info.Type)

	resp2, err := soapClient.Call(
		info.Path,
		info.Type+"#GetSecurityKeys",
		soapBody2,
	)
	if err != nil {
		return "", fmt.Errorf("failed to get WLAN security keys: %w", err)
	}

	// Clean and parse response for KeyPassphrase
	resp2 = soap.CleanSoapResponse(resp2)
	var keyResult struct {
		XMLName        xml.Name `xml:"Envelope"`
		NewKeyPassphrase string   `xml:"Body>GetSecurityKeysResponse>NewKeyPassphrase"`
	}
	if err := xml.Unmarshal([]byte(resp2), &keyResult); err != nil {
		return "", fmt.Errorf("failed to parse GetSecurityKeys response: %w", err)
	}

	// Generate QR code string
	qrData := fmt.Sprintf("WIFI:T:WPA;S:%s;P:%s;;", infoResult.NewSSID, keyResult.NewKeyPassphrase)
	return qrData, nil
}
