package services

import (
	"encoding/xml"
	"fmt"

	"github.com/fabito/fritzboxctl/internal/soap"
)

// WakeOnLAN sends a Wake-on-LAN packet to the specified MAC address
// Uses Hosts:1 service, X_AVM-DE_WakeOnLANByMACAddress action
// If soapClient is nil, returns nil (mock success for testing)
func WakeOnLAN(mac string, soapClient soapCaller) error {
	if soapClient == nil {
		return nil
	}

	// Build SOAP request
	soapBody := fmt.Sprintf(`<?xml version="1.0" encoding="utf-8"?>
<s:Envelope s:encodingStyle="http://schemas.xmlsoap.org/soap/encoding/" xmlns:s="http://schemas.xmlsoap.org/soap/envelope/">
    <s:Body>
        <u:X_AVM-DE_WakeOnLANByMACAddress xmlns:u="urn:dslforum-org:service:Hosts:1">
            <NewMACAddress>%s</NewMACAddress>
        </u:X_AVM-DE_WakeOnLANByMACAddress>
    </s:Body>
</s:Envelope>`, mac)

	resp, err := soapClient.Call(
		"/upnp/control/hosts",
		"urn:dslforum-org:service:Hosts:1#X_AVM-DE_WakeOnLANByMACAddress",
		soapBody,
	)
	if err != nil {
		return fmt.Errorf("failed to send Wake-on-LAN: %w", err)
	}

	// Clean and parse response
	resp = soap.CleanSoapResponse(resp)
	var envelope struct {
		XMLName xml.Name `xml:"Envelope"`
	}
	if err := xml.Unmarshal([]byte(resp), &envelope); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	return nil
}
