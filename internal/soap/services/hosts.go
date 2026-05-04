package services

import (
	"encoding/xml"
	"fmt"

	"github.com/fabito/fritzboxctl/internal/soap"
)

// HostListResponse represents the SOAP response for GetGenericHostEntry
// The XML tags flatten the nested SOAP structure
// Body > GetGenericHostEntryResponse > FieldName
type HostListResponse struct {
	XMLName          xml.Name `xml:"Envelope"`
	NewHostName      string   `xml:"Body>GetGenericHostEntryResponse>NewHostName"`
	NewIPAddress     string   `xml:"Body>GetGenericHostEntryResponse>NewIPAddress"`
	NewMACAddress    string   `xml:"Body>GetGenericHostEntryResponse>NewMACAddress"`
	NewInterfaceType string   `xml:"Body>GetGenericHostEntryResponse>NewInterfaceType"`
	NewActive        string   `xml:"Body>GetGenericHostEntryResponse>NewActive"`
	NewPort          string   `xml:"Body>GetGenericHostEntryResponse>NewX_AVM-DE_Port"`
}

// HostNumberResponse represents the SOAP response for GetHostNumberOfEntries
// The XML tags flatten the nested SOAP structure
// Body > GetHostNumberOfEntriesResponse > FieldName
type HostNumberResponse struct {
	XMLName                xml.Name `xml:"Envelope"`
	NewHostNumberOfEntries string   `xml:"Body>GetHostNumberOfEntriesResponse>NewHostNumberOfEntries"`
}

// BlockDevice blocks internet access for a device by IP address
// Uses X_AVM-DE_HostFilter:1 service with DisallowWANAccessByIP action
func BlockDevice(ipAddress string, soapClient *soap.Client) error {
	if soapClient == nil {
		// Mock mode for testing - return nil (success)
		return nil
	}

	// Build SOAP request for DisallowWANAccessByIP with NewDisallow=1 (block)
	soapBody := fmt.Sprintf(`<?xml version="1.0" encoding="utf-8"?>
<s:Envelope s:encodingStyle="http://schemas.xmlsoap.org/soap/encoding/" xmlns:s="http://schemas.xmlsoap.org/soap/envelope/">
    <s:Body>
        <u:DisallowWANAccessByIP xmlns:u="urn:dslforum-org:service:X_AVM-DE_HostFilter:1">
            <NewIPv4Address>%s</NewIPv4Address>
            <NewDisallow>1</NewDisallow>
        </u:DisallowWANAccessByIP>
    </s:Body>
</s:Envelope>`, ipAddress)

	_, err := soapClient.Call(
		"/upnp/control/hostfilter",
		"urn:dslforum-org:service:X_AVM-DE_HostFilter:1#DisallowWANAccessByIP",
		soapBody,
	)
	if err != nil {
		return fmt.Errorf("failed to block device %s: %w", ipAddress, err)
	}

	return nil
}

// UnblockDevice unblocks internet access for a device by IP address
// Uses X_AVM-DE_HostFilter:1 service with DisallowWANAccessByIP action
func UnblockDevice(ipAddress string, soapClient *soap.Client) error {
	if soapClient == nil {
		// Mock mode for testing - return nil (success)
		return nil
	}

	// Build SOAP request for DisallowWANAccessByIP with NewDisallow=0 (unblock)
	soapBody := fmt.Sprintf(`<?xml version="1.0" encoding="utf-8"?>
<s:Envelope s:encodingStyle="http://schemas.xmlsoap.org/soap/encoding/" xmlns:s="http://schemas.xmlsoap.org/soap/envelope/">
    <s:Body>
        <u:DisallowWANAccessByIP xmlns:u="urn:dslforum-org:service:X_AVM-DE_HostFilter:1">
            <NewIPv4Address>%s</NewIPv4Address>
            <NewDisallow>0</NewDisallow>
        </u:DisallowWANAccessByIP>
    </s:Body>
</s:Envelope>`, ipAddress)

	_, err := soapClient.Call(
		"/upnp/control/hostfilter",
		"urn:dslforum-org:service:X_AVM-DE_HostFilter:1#DisallowWANAccessByIP",
		soapBody,
	)
	if err != nil {
		return fmt.Errorf("failed to unblock device %s: %w", ipAddress, err)
	}

	return nil
}
