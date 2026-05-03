package aha

import "net/http"

// Client is the AHA (AVM Home Automation) HTTP client
// It uses SID authentication for Fritz!Box REST APIs
type Client struct {
	baseURL    string
	sid        string
	httpClient *http.Client
}

// NewClient creates a new AHA client
func NewClient(baseURL string, sid string) *Client {
	return &Client{
		baseURL:    baseURL,
		sid:        sid,
		httpClient: http.DefaultClient,
	}
}
