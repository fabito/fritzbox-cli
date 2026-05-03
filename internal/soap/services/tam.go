package services

import (
	"encoding/xml"
	"fmt"

	"github.com/fabito/fritzboxctl/internal/soap"
)

// TAMInfo represents a TAM (answering machine) on the Fritz!Box
type TAMInfo struct {
	XMLName                xml.Name `xml:"Envelope"`
	NewTAMIndex            string   `xml:"Body>GetInfoResponse>NewTAMIndex"`
	NewTAMName             string   `xml:"Body>GetInfoResponse>NewTAMName"`
	NewTAMEnable           string   `xml:"Body>GetInfoResponse>NewTAMEnable"`
	NewTAMNewMessageCount  string   `xml:"Body>GetInfoResponse>NewTAMNewMessageCount"`
	NewTAMTotalMessageCount string  `xml:"Body>GetInfoResponse>NewTAMTotalMessageCount"`
}

// ListTAMs lists all TAMs (answering machines) on the Fritz!Box
// If soapClient is nil, returns mock data for testing
func ListTAMs(soapClient *soap.Client) ([]TAMInfo, error) {
	if soapClient == nil {
		// Return mock data for testing
		return []TAMInfo{
			{
				NewTAMIndex:            "0",
				NewTAMName:             "TAM 0",
				NewTAMEnable:           "1",
				NewTAMNewMessageCount:  "3",
				NewTAMTotalMessageCount: "10",
			},
		}, nil
	}

	var tams []TAMInfo
	// Iterate through TAM indices (0, 1, 2...) until we get an error
	for i := 0; i < 10; i++ { // Max 10 TAMs should be enough
		tam, err := getTAMInfo(soapClient, i)
		if err != nil {
			// No more TAMs at this index
			break
		}
		tams = append(tams, *tam)
	}

	return tams, nil
}

// getTAMInfo gets info for a specific TAM index
func getTAMInfo(soapClient *soap.Client, index int) (*TAMInfo, error) {
	soapBody := fmt.Sprintf(`<?xml version="1.0" encoding="utf-8"?>
<s:Envelope s:encodingStyle="http://schemas.xmlsoap.org/soap/encoding/" xmlns:s="http://schemas.xmlsoap.org/soap/envelope/">
    <s:Body>
        <u:GetInfo xmlns:u="urn:dslforum-org:service:X_AVM-DE_TAM:1">
            <NewTAMIndex>%d</NewTAMIndex>
        </u:GetInfo>
    </s:Body>
</s:Envelope>`, index)

	resp, err := soapClient.Call(
		"/upnp/control/x_tam",
		"urn:dslforum-org:service:X_AVM-DE_TAM:1#GetInfo",
		soapBody,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get TAM info for index %d: %w", index, err)
	}

	// Clean and parse response
	resp = soap.CleanSoapResponse(resp)
	var tam TAMInfo
	if err := xml.Unmarshal([]byte(resp), &tam); err != nil {
		return nil, fmt.Errorf("failed to parse TAM response: %w", err)
	}

	return &tam, nil
}

// SetTAMEnabled enables or disables a TAM (answering machine)
// If soapClient is nil, returns nil (mock mode for testing)
func SetTAMEnabled(index int, enabled bool, soapClient *soap.Client) error {
	if soapClient == nil {
		// Mock mode - return success
		return nil
	}

	// Convert enabled to string "1" or "0"
	enabledStr := "0"
	if enabled {
		enabledStr = "1"
	}

	// Build SOAP request for SetEnable
	soapBody := fmt.Sprintf(`<?xml version="1.0" encoding="utf-8"?>
<s:Envelope s:encodingStyle="http://schemas.xmlsoap.org/soap/encoding/" xmlns:s="http://schemas.xmlsoap.org/soap/envelope/">
    <s:Body>
        <u:SetEnable xmlns:u="urn:dslforum-org:service:X_AVM-DE_TAM:1">
            <NewTAMIndex>%d</NewTAMIndex>
            <NewEnable>%s</NewEnable>
        </u:SetEnable>
    </s:Body>
</s:Envelope>`, index, enabledStr)

	_, err := soapClient.Call(
		"/upnp/control/x_tam",
		"urn:dslforum-org:service:X_AVM-DE_TAM:1#SetEnable",
		soapBody,
	)
	if err != nil {
		return fmt.Errorf("failed to set TAM %d enabled=%t: %w", index, enabled, err)
	}

	return nil
}
