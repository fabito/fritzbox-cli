package soap

import "encoding/xml"

// Envelope represents a SOAP envelope
type Envelope struct {
	XMLName xml.Name    `xml:"s:Envelope"`
	XmlnsS  string     `xml:"xmlns:s,attr"`
	Encoding string     `xml:"s:encodingStyle,attr"`
	Body    BodyEnvelope `xml:"s:Body"`
}

// BodyEnvelope represents the SOAP body
type BodyEnvelope struct {
	Action interface{} `xml:",any"`
}

// Response represents a SOAP response
type Response struct {
	XMLName xml.Name `xml:"Envelope"`
	Body    ResponseBody `xml:"Body"`
}

// ResponseBody represents the SOAP response body
type ResponseBody struct {
	Content []byte `xml:",innerxml"`
}

// ServiceDescription describes a TR-064 service
type ServiceDescription struct {
	ServiceType string `xml:"serviceType"`
	ServiceId   string `xml:"serviceId"`
	ControlURL  string `xml:"controlURL"`
	EventSubURL string `xml:"eventSubURL"`
	SCPDURL     string `xml:"SCPDURL"`
}

// DeviceDescription describes the device from tr64desc.xml
type DeviceDescription struct {
	Device Device `xml:"device"`
}

// Device represents a device in the description
type Device struct {
	DeviceType  string        `xml:"deviceType"`
	FriendlyName string       `xml:"friendlyName"`
	Manufacturer string       `xml:"manufacturer"`
	ModelName   string       `xml:"modelName"`
	ModelNumber string       `xml:"modelNumber"`
	SerialNumber string      `xml:"serialNumber"`
	Services    []Service   `xml:"serviceList>service"`
	Devices     []Device    `xml:"deviceList>device"`
}

// Service represents a service in the description XML
type Service struct {
	ServiceType string `xml:"serviceType"`
	ServiceId   string `xml:"serviceId"`
	ControlURL  string `xml:"controlURL"`
	EventSubURL string `xml:"eventSubURL"`
	SCPDURL     string `xml:"SCPDURL"`
}

// NewEnvelope creates a new SOAP envelope
func NewEnvelope(actionNamespace, actionName string, params map[string]string) *Envelope {
	env := &Envelope{
		XmlnsS:  "http://schemas.xmlsoap.org/soap/envelope/",
		Encoding: "http://schemas.xmlsoap.org/soap/encoding/",
		Body: BodyEnvelope{
			Action: NewAction(actionNamespace, actionName, params),
		},
	}
	return env
}

// Action represents a SOAP action
type Action struct {
	XMLName xml.Name `xml:"u:{ActionName}"`
	Params  []ActionParam
}

// NewAction creates a new SOAP action
func NewAction(namespace, name string, params map[string]string) *Action {
	action := &Action{
		XMLName: xml.Name{
			Space: namespace,
			Local: name,
		},
		Params: make([]ActionParam, 0, len(params)),
	}

	for key, value := range params {
		action.Params = append(action.Params, ActionParam{XMLName: xml.Name{Local: key}, Value: value})
	}

	return action
}

// ActionParam represents a parameter in a SOAP action
type ActionParam struct {
	XMLName xml.Name `xml:",any"`
	Value   string   `xml:",chardata"`
}
