package services

import (
	"encoding/xml"
	"fmt"

	"github.com/fabito/fritzboxctl/internal/soap"
)

// LANStatsResponse represents the SOAP response for GetStatistics on LANEthernetInterfaceConfig
// The XML tags flatten the nested SOAP structure using path syntax
type LANStatsResponse struct {
	XMLName            xml.Name `xml:"Envelope"`
	NewBytesSent       string   `xml:"Body>GetStatisticsResponse>NewBytesSent"`
	NewBytesReceived   string   `xml:"Body>GetStatisticsResponse>NewBytesReceived"`
	NewPacketsSent     string   `xml:"Body>GetStatisticsResponse>NewPacketsSent"`
	NewPacketsReceived string   `xml:"Body>GetStatisticsResponse>NewPacketsReceived"`
	NewErrorsSent      string   `xml:"Body>GetStatisticsResponse>NewErrorsSent"`
	NewErrorsReceived  string   `xml:"Body>GetStatisticsResponse>NewErrorsReceived"`
}

// GetLANStats retrieves LAN statistics from Fritz!Box
// If soapClient is nil, returns mock data for testing
func GetLANStats(soapClient *soap.Client) (*LANStatsResponse, error) {
	if soapClient == nil {
		// Return mock data for testing
		return &LANStatsResponse{
			NewBytesSent:       "123456789",
			NewBytesReceived:   "987654321",
			NewPacketsSent:     "12345",
			NewPacketsReceived: "54321",
			NewErrorsSent:      "10",
			NewErrorsReceived:  "5",
		}, nil
	}

	// Build SOAP request for GetStatistics
	soapBody := fmt.Sprintf(`<?xml version="1.0" encoding="utf-8"?>
<s:Envelope s:encodingStyle="http://schemas.xmlsoap.org/soap/encoding/" xmlns:s="http://schemas.xmlsoap.org/soap/envelope/">
    <s:Body>
        <u:GetStatistics xmlns:u="urn:dslforum-org:service:LANEthernetInterfaceConfig:1">
        </u:GetStatistics>
    </s:Body>
</s:Envelope>`)

	resp, err := soapClient.Call(
		"/upnp/control/lanethernetifcfg",
		"urn:dslforum-org:service:LANEthernetInterfaceConfig:1#GetStatistics",
		soapBody,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get LAN statistics: %w", err)
	}

	// Clean and parse response using shared function
	resp = soap.CleanSoapResponse(resp)
	var stats LANStatsResponse
	if err := xml.Unmarshal([]byte(resp), &stats); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &stats, nil
}
