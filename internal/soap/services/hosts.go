package services

import "encoding/xml"

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
	XMLName              xml.Name `xml:"Envelope"`
	NewHostNumberOfEntries string   `xml:"Body>GetHostNumberOfEntriesResponse>NewHostNumberOfEntries"`
}
