package services

import (
	"encoding/xml"
	"fmt"

	"github.com/fabito/fritzboxctl/internal/soap"
)

// WANStatusResponse represents the SOAP response for GetStatusInfo
type WANStatusResponse struct {
	XMLName              xml.Name `xml:"Envelope"`
	NewConnectionStatus  string   `xml:"Body>GetStatusInfoResponse>NewConnectionStatus"`
	NewExternalIPAddress string   `xml:"Body>GetStatusInfoResponse>NewExternalIPAddress"`
	NewLastConnectionError string `xml:"Body>GetStatusInfoResponse>NewLastConnectionError"`
	NewUptime           string   `xml:"Body>GetStatusInfoResponse>NewUptime"`
	NewDNSServers       string   `xml:"Body>GetStatusInfoResponse>NewDNSServers"`
}

// GetWANStatus retrieves WAN connection status from Fritz!Box
// If soapClient is nil, returns mock data for testing
func GetWANStatus(soapClient *soap.Client) (*WANStatusResponse, error) {
	if soapClient == nil {
		// Return mock data for testing
		return &WANStatusResponse{
			NewConnectionStatus:  "Connected",
			NewExternalIPAddress: "192.168.1.100",
		}, nil
	}

	// Build SOAP request for GetStatusInfo
	soapBody := `<?xml version='1.0' encoding='utf-8'?>
<s:Envelope s:encodingStyle='http://schemas.xmlsoap.org/soap/encoding/' xmlns:s='http://schemas.xmlsoap.org/soap/envelope/'>
  <s:Body>
    <u:GetStatusInfo xmlns:u='urn:schemas-upnp-org:service:WANIPConnection:1' />
  </s:Body>
</s:Envelope>`

	resp, err := soapClient.Call(
		"/igdupnp/control/WANIPConn1",
		"urn:schemas-upnp-org:service:WANIPConnection:1#GetStatusInfo",
		soapBody,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get WAN status: %w", err)
	}

	// Clean and parse response
	resp = soap.CleanSoapResponse(resp)
	var status WANStatusResponse
	if err := xml.Unmarshal([]byte(resp), &status); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &status, nil
}
