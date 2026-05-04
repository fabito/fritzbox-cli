package services

import (
	"encoding/xml"
	"fmt"

	"github.com/fabito/fritzboxctl/internal/soap"
)

// GetLANCount returns the number of LAN devices
func GetLANCount(soapClient *soap.Client) (int, error) {
	if soapClient == nil {
		return 0, fmt.Errorf("SOAP client is nil")
	}

	// Build SOAP request
	soapBody := `<?xml version="1.0" encoding="utf-8"?>
<s:Envelope s:encodingStyle="http://schemas.xmlsoap.org/soap/encoding/" xmlns:s="http://schemas.xmlsoap.org/soap/envelope/">
    <s:Body>
        <u:GetHostNumberOfEntries xmlns:u="urn:dslforum-org:service:Hosts:1">
        </u:GetHostNumberOfEntries>
    </s:Body>
</s:Envelope>`

	resp, err := soapClient.Call(
		"/upnp/control/hosts",
		"urn:dslforum-org:service:Hosts:1#GetHostNumberOfEntries",
		soapBody,
	)
	if err != nil {
		return 0, fmt.Errorf("failed to get LAN count: %w", err)
	}

	// Clean and parse response
	resp = soap.CleanSoapResponse(resp)
	var result struct {
		XMLName                xml.Name `xml:"Envelope"`
		NewHostNumberOfEntries string   `xml:"Body>GetHostNumberOfEntriesResponse>NewHostNumberOfEntries"`
	}
	if err := xml.Unmarshal([]byte(resp), &result); err != nil {
		return 0, fmt.Errorf("failed to parse response: %w", err)
	}

	// Convert to int
	count := 0
	fmt.Sscanf(result.NewHostNumberOfEntries, "%d", &count)
	return count, nil
}
