package services

import (
	"encoding/xml"
	"fmt"

	"github.com/fabito/fritzboxctl/internal/soap"
)

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
