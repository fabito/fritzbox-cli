package aha

import "net/http"

// Client is the AHA (AVM Home Automation) HTTP client
// It uses SID authentication for Fritz!Box REST APIs
type Client struct {
	baseURL    string
	sid        string
	httpClient *http.Client
	getSID     func() (string, error)
}

// NewClient creates a new AHA client
func NewClient(baseURL string, getSID func() (string, error)) *Client {
	return &Client{
		baseURL:    baseURL,
		httpClient: http.DefaultClient,
		getSID:     getSID,
	}
}
