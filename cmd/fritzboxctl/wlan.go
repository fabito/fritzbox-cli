package main

import (
	"encoding/xml"
	"fmt"
	"log/slog"
	"strings"

	"github.com/spf13/cobra"
)

// newWLANCommand creates the WLAN command tree
func newWLANCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "wlan",
		Short: "WLAN-related commands",
		Long:  `Commands for managing WLAN (WiFi) settings.`,
	}

	cmd.AddCommand(newWLANStatusCommand())
	cmd.AddCommand(newWLANSetCommand())
	cmd.AddCommand(newWLANStatsCommand())

	return cmd
}

// newWLANStatusCommand creates the WLAN status command
func newWLANStatusCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Get WLAN status",
		RunE:  runWLANStatus,
	}
}

// runWLANStatus executes the WLAN status command
func runWLANStatus(cmd *cobra.Command, args []string) error {
	return fmt.Errorf("WLAN status command not yet implemented")
}

// newWLANSetCommand creates the WLAN set command
func newWLANSetCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "set",
		Short: "Set WLAN state (on/off)",
		RunE:  runWLANSet,
	}
}

// runWLANSet executes the WLAN set command
func runWLANSet(cmd *cobra.Command, args []string) error {
	return fmt.Errorf("WLAN set command not yet implemented")
}

// newWLANStatsCommand creates the WLAN stats command
func newWLANStatsCommand() *cobra.Command {
	var band int
	cmd := &cobra.Command{
		Use:   "stats",
		Short: "Get WLAN statistics",
		Long:  `Retrieves WLAN statistics such as packets sent/received, errors, etc.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runWLANStats(band)
		},
	}
	cmd.Flags().IntVarP(&band, "band", "b", 1, "WLAN band (1=2.4GHz, 2=5GHz, 3=5GHz ch2, 4=Guest)")
	return cmd
}

// WLANBand represents the WLAN band
type WLANBand int

const (
	WLANBand2_4GHz WLANBand = 1
	WLANBand5GHz   WLANBand = 2
	WLANBand5GHz2  WLANBand = 3
	WLANBandGuest  WLANBand = 4
)

// wlanServiceInfo maps band to service path and type
var wlanServiceInfo = map[int]struct {
	Path string
	Type string
}{
	1: {Path: "/upnp/control/wlanconfig1", Type: "urn:dslforum-org:service:WLANConfiguration:1"},
	2: {Path: "/upnp/control/wlanconfig2", Type: "urn:dslforum-org:service:WLANConfiguration:2"},
	3: {Path: "/upnp/control/wlanconfig3", Type: "urn:dslforum-org:service:WLANConfiguration:3"},
	4: {Path: "/upnp/control/wlanconfig4", Type: "urn:dslforum-org:service:WLANConfiguration:4"},
}

// runWLANStats executes the WLAN stats command
func runWLANStats(band int) error {
	info, ok := wlanServiceInfo[band]
	if !ok {
		return fmt.Errorf("invalid band: %d (use 1-4)", band)
	}

	// Build SOAP request for GetStatistics
	soapBody := fmt.Sprintf(`<?xml version="1.0" encoding="utf-8"?>
<s:Envelope s:encodingStyle="http://schemas.xmlsoap.org/soap/encoding/" xmlns:s="http://schemas.xmlsoap.org/soap/envelope/">
    <s:Body>
        <u:GetStatistics xmlns:u="%s">
        </u:GetStatistics>
    </s:Body>
</s:Envelope>`, info.Type)

	resp, err := soapClient.Call(
		info.Path,
		info.Type+"#GetStatistics",
		soapBody,
	)
	if err != nil {
		return fmt.Errorf("failed to get WLAN statistics: %w", err)
	}

	// Parse and display the response
	return displayWLANStats(resp, band)
}

// WLANStatsEnvelope represents the SOAP response for GetStatistics
type WLANStatsEnvelope struct {
	XMLName xml.Name `xml:"Envelope"`
	Body    struct {
		GetStatisticsResponse struct {
			NewTotalPacketsSent        string `xml:"NewTotalPacketsSent"`
			NewTotalPacketsReceived    string `xml:"NewTotalPacketsReceived"`
			NewPacketErrorsReceived    string `xml:"NewPacketErrorsReceived"`
			NewPacketErrorsSent        string `xml:"NewPacketErrorsSent"`
			NewTotalBytesSent          string `xml:"NewTotalBytesSent"`
			NewTotalBytesReceived      string `xml:"NewTotalBytesReceived"`
			NewErrorsReceived          string `xml:"NewErrorsReceived"`
			NewErrorsSent              string `xml:"NewErrorsSent"`
			NewUnicastPacketsSent      string `xml:"NewUnicastPacketsSent"`
			NewUnicastPacketsReceived  string `xml:"NewUnicastPacketsReceived"`
			NewMulticastPacketsSent    string `xml:"NewMulticastPacketsSent"`
			NewMulticastPacketsReceived string `xml:"NewMulticastPacketsReceived"`
			NewBroadcastPacketsSent    string `xml:"NewBroadcastPacketsSent"`
			NewBroadcastPacketsReceived string `xml:"NewBroadcastPacketsReceived"`
		} `xml:"GetStatisticsResponse"`
	} `xml:"Body"`
}

// displayWLANStats parses and displays WLAN statistics
func displayWLANStats(resp string, band int) error {
	// Clean up response for XML parsing
	resp = cleanSoapResponse(resp)

	slog.Debug("displayWLANStats: cleaned response", "response", resp[:min(len(resp), 500)])

	var envelope WLANStatsEnvelope
	decoder := xml.NewDecoder(strings.NewReader(resp))
	if err := decoder.Decode(&envelope); err != nil {
		return fmt.Errorf("failed to parse WLAN statistics response: %w", err)
	}

	stats := envelope.Body.GetStatisticsResponse

	// Display based on output format
	switch cfg.OutputFormat {
	case "json":
		fmt.Printf(`{"band": %d, "packets_sent": "%s", "packets_received": "%s", "bytes_sent": "%s", "bytes_received": "%s"}\n`,
			band, stats.NewTotalPacketsSent, stats.NewTotalPacketsReceived,
			stats.NewTotalBytesSent, stats.NewTotalBytesReceived)
	default:
		bandName := getBandName(band)
		fmt.Printf("WLAN Statistics for %s\n", bandName)
		fmt.Println("=" + strings.Repeat("=", len("WLAN Statistics for ")+len(bandName)))
		if stats.NewTotalPacketsSent != "" {
			fmt.Printf("Total Packets Sent:        %s\n", stats.NewTotalPacketsSent)
		}
		if stats.NewTotalPacketsReceived != "" {
			fmt.Printf("Total Packets Received:    %s\n", stats.NewTotalPacketsReceived)
		}
		if stats.NewTotalBytesSent != "" {
			fmt.Printf("Total Bytes Sent:          %s\n", stats.NewTotalBytesSent)
		}
		if stats.NewTotalBytesReceived != "" {
			fmt.Printf("Total Bytes Received:      %s\n", stats.NewTotalBytesReceived)
		}
		if stats.NewPacketErrorsReceived != "" {
			fmt.Printf("Packet Errors Received:    %s\n", stats.NewPacketErrorsReceived)
		}
		if stats.NewPacketErrorsSent != "" {
			fmt.Printf("Packet Errors Sent:        %s\n", stats.NewPacketErrorsSent)
		}
		if stats.NewErrorsReceived != "" {
			fmt.Printf("Errors Received:            %s\n", stats.NewErrorsReceived)
		}
		if stats.NewErrorsSent != "" {
			fmt.Printf("Errors Sent:                %s\n", stats.NewErrorsSent)
		}
		if stats.NewUnicastPacketsSent != "" {
			fmt.Printf("Unicast Packets Sent:      %s\n", stats.NewUnicastPacketsSent)
		}
		if stats.NewUnicastPacketsReceived != "" {
			fmt.Printf("Unicast Packets Received:  %s\n", stats.NewUnicastPacketsReceived)
		}
		if stats.NewMulticastPacketsSent != "" {
			fmt.Printf("Multicast Packets Sent:    %s\n", stats.NewMulticastPacketsSent)
		}
		if stats.NewMulticastPacketsReceived != "" {
			fmt.Printf("Multicast Packets Received:%s\n", stats.NewMulticastPacketsReceived)
		}
		if stats.NewBroadcastPacketsSent != "" {
			fmt.Printf("Broadcast Packets Sent:    %s\n", stats.NewBroadcastPacketsSent)
		}
		if stats.NewBroadcastPacketsReceived != "" {
			fmt.Printf("Broadcast Packets Received:%s\n", stats.NewBroadcastPacketsReceived)
		}
	}

	return nil
}

// getBandName returns a human-readable band name
func getBandName(band int) string {
	switch band {
	case 1:
		return "2.4 GHz"
	case 2:
		return "5 GHz"
	case 3:
		return "5 GHz (2nd channel)"
	case 4:
		return "Guest Network"
	default:
		return fmt.Sprintf("Unknown (%d)", band)
	}
}

// cleanSoapResponse removes namespace prefixes and fixes XML for parsing
func cleanSoapResponse(resp string) string {
	resp = strings.ReplaceAll(resp, ` encodingStyle="http://schemas.xmlsoap.org/soap/encoding/"`, "")
	resp = strings.ReplaceAll(resp, " >", ">")
	resp = strings.ReplaceAll(resp, ` xmlns:s="http://schemas.xmlsoap.org/soap/envelope/"`, "")
	resp = strings.ReplaceAll(resp, ` xmlns:u="urn:dslforum-org:service:WLANConfiguration:1"`, "")
	resp = strings.ReplaceAll(resp, ` xmlns:u="urn:dslforum-org:service:WLANConfiguration:2"`, "")
	resp = strings.ReplaceAll(resp, ` xmlns:u="urn:dslforum-org:service:WLANConfiguration:3"`, "")
	resp = strings.ReplaceAll(resp, ` xmlns:u="urn:dslforum-org:service:WLANConfiguration:4"`, "")
	resp = strings.ReplaceAll(resp, "s:", "")
	resp = strings.ReplaceAll(resp, "u:", "")
	return resp
}
