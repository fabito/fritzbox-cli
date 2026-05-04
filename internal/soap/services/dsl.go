package services

import (
	"encoding/xml"
	"fmt"

	"github.com/fabito/fritzboxctl/internal/soap"
)

// DSLStatusResponse represents the SOAP response for GetInfo on WANDSLInterfaceConfig
// The XML tags flatten the nested SOAP structure using path syntax
type DSLStatusResponse struct {
	XMLName                  xml.Name `xml:"Envelope"`
	NewEnable                string   `xml:"Body>GetInfoResponse>NewEnable"`
	NewStatus                string   `xml:"Body>GetInfoResponse>NewStatus"`
	NewDataPath              string   `xml:"Body>GetInfoResponse>NewDataPath"`
	NewDownstreamCurrRate    string   `xml:"Body>GetInfoResponse>NewDownstreamCurrRate"`
	NewUpstreamCurrRate      string   `xml:"Body>GetInfoResponse>NewUpstreamCurrRate"`
	NewDownstreamMaxRate     string   `xml:"Body>GetInfoResponse>NewDownstreamMaxRate"`
	NewUpstreamMaxRate       string   `xml:"Body>GetInfoResponse>NewUpstreamMaxRate"`
	NewDownstreamNoiseMargin string   `xml:"Body>GetInfoResponse>NewDownstreamNoiseMargin"`
	NewUpstreamNoiseMargin   string   `xml:"Body>GetInfoResponse>NewUpstreamNoiseMargin"`
}

// GetDSLStatus retrieves DSL status from Fritz!Box
// If soapClient is nil, returns mock data for testing
func GetDSLStatus(soapClient *soap.Client) (*DSLStatusResponse, error) {
	if soapClient == nil {
		// Return mock data for testing
		return &DSLStatusResponse{
			NewDownstreamCurrRate:    "100000",
			NewUpstreamCurrRate:      "40000",
			NewDownstreamMaxRate:     "120000",
			NewUpstreamMaxRate:       "50000",
			NewDownstreamNoiseMargin: "10",
			NewUpstreamNoiseMargin:   "15",
		}, nil
	}

	// Build SOAP request for WANDSLInterfaceConfig:GetInfo
	soapBody := `<?xml version="1.0" encoding="utf-8"?>
<s:Envelope s:encodingStyle="http://schemas.xmlsoap.org/soap/encoding/" xmlns:s="http://schemas.xmlsoap.org/soap/envelope/">
    <s:Body>
        <u:GetInfo xmlns:u="urn:dslforum-org:service:WANDSLInterfaceConfig:1">
        </u:GetInfo>
    </s:Body>
</s:Envelope>`

	resp, err := soapClient.Call(
		"/upnp/control/wandslifconfig1",
		"urn:dslforum-org:service:WANDSLInterfaceConfig:1#GetInfo",
		soapBody,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get DSL status: %w", err)
	}

	// Clean and parse response using shared function
	resp = soap.CleanSoapResponse(resp)

	var status DSLStatusResponse
	if err := xml.Unmarshal([]byte(resp), &status); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &status, nil
}
