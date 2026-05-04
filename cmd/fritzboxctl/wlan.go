package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/fabito/fritzboxctl/internal/soap/services"

	"encoding/json"
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
	cmd.AddCommand(newWLANQRCodeCommand())
	cmd.AddCommand(newWLANChannelCommand())

	return cmd
}

// newWLANStatusCommand creates the WLAN status command
func newWLANStatusCommand() *cobra.Command {
	var band int
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Get WLAN status",
		Long:  `Retrieves WLAN status such as enabled/disabled, SSID, channel, etc.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runWLANStatus(band)
		},
	}
	cmd.Flags().IntVarP(&band, "band", "b", 0, "WLAN band (1=2.4GHz, 2=5GHz, 3=5GHz 2nd, 4=Guest, 0=all)")
	return cmd
}

// runWLANStatus executes the WLAN status command
func runWLANStatus(band int) error {
	// If band is 0, show all bands
	bands := []int{}
	if band == 0 {
		bands = []int{1, 2, 3, 4}
	} else {
		bands = []int{band}
	}

	for _, b := range bands {
		status, err := services.GetWLANStatus(soapClient, b)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to get WLAN status for band %d: %v\n", b, err)
			continue
		}

		// Display status
		displayWLANStatus(status, b)
	}

	return nil
}

// newWLANSetCommand creates the WLAN set command
func newWLANSetCommand() *cobra.Command {
	var band int
	var enableFlag bool
	var disableFlag bool

	cmd := &cobra.Command{
		Use:   "set",
		Short: "Set WLAN state (on/off)",
		Long:  `Enables or disables WLAN (WiFi) for the specified band.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Determine enable/disable state
			if enableFlag && disableFlag {
				return fmt.Errorf("cannot specify both --on and --off")
			}
			if !enableFlag && !disableFlag {
				return fmt.Errorf("must specify --on or --off")
			}
			enabled := enableFlag

			// Determine bands to affect
			bands := []int{}
			if band == 0 {
				// All bands (1-4)
				bands = []int{1, 2, 3, 4}
			} else {
				bands = []int{band}
			}

			// Call service layer for each band
			for _, b := range bands {
				if err := services.SetWLANEnabled(b, enabled, soapClient); err != nil {
					fmt.Fprintf(os.Stderr, "Failed to set WLAN for band %d: %v\n", b, err)
					continue
				}
				bandName := getBandName(b)
				state := "disabled"
				if enabled {
					state = "enabled"
				}
				fmt.Printf("WLAN %s %s\n", bandName, state)
			}
			return nil
		},
	}
	cmd.Flags().IntVarP(&band, "band", "b", 0, "WLAN band (1=2.4GHz, 2=5GHz, 3=5GHz 2nd, 4=Guest, 0=all)")
	cmd.Flags().BoolVar(&enableFlag, "on", false, "Enable WLAN")
	cmd.Flags().BoolVar(&disableFlag, "off", false, "Disable WLAN")
	return cmd
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

// runWLANStats executes the WLAN stats command
func runWLANStats(band int) error {
	// Call the service layer to get WLAN statistics
	stats, err := services.GetWLANStats(soapClient, band)
	if err != nil {
		return fmt.Errorf("failed to get WLAN statistics: %w", err)
	}

	// Display based on output format
	switch cfg.OutputFormat {
	case "json":
		// JSON output
		fmt.Printf(`{"band": %d, "packets_sent": "%s", "packets_received": "%s", "bytes_sent": "%s", "bytes_received": "%s"}\n`,
			band, stats.NewTotalPacketsSent, stats.NewTotalPacketsReceived,
			stats.NewTotalBytesSent, stats.NewTotalBytesReceived)
	default:
		// Text output
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

// displayWLANStatus formats and displays WLAN status
func displayWLANStatus(status *services.WLANStatusResponse, band int) {
	bandName := getBandName(band)

	switch cfg.OutputFormat {
	case "json":
		fmt.Printf(`{"band": %d, "name": "%s", "enabled": "%s", "ssid": "%s", "channel": "%s"}\n`,
			band, bandName, status.NewEnable, status.NewSSID, status.NewChannel)
	default:
		fmt.Printf("\nWLAN Status for %s\n", bandName)
		fmt.Println("=" + strings.Repeat("=", len("WLAN Status for ")+len(bandName)))
		fmt.Printf("Enabled:      %s\n", status.NewEnable)
		fmt.Printf("SSID:         %s\n", status.NewSSID)
		fmt.Printf("Beacon Type:  %s\n", status.NewBeaconType)
		fmt.Printf("Channel:      %s\n", status.NewChannel)
		fmt.Printf("Max Bit Rate: %s\n", status.NewMaxBitRate)
		fmt.Printf("MAC Address:  %s\n", status.NewMACAddress)
		fmt.Printf("BSSID:        %s\n", status.NewBSSID)
	}
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

// newWLANQRCodeCommand creates the WLAN QR code command
// newWLANQRCodeCommand creates the WLAN QR code command
func newWLANQRCodeCommand() *cobra.Command {
	var band int
	cmd := &cobra.Command{
		Use:   "qrcode",
		Short: "Generate QR code for WLAN connection",
		Long:  `Generates a QR code string for connecting to the WLAN. Use with a QR code generator.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runWLANQRCode(band)
		},
	}
	cmd.Flags().IntVarP(&band, "band", "b", 1, "WLAN band (1=2.4GHz, 2=5GHz, 3=5GHz ch2, 4=Guest)")
	return cmd
}

func runWLANQRCode(band int) error {
	if soapClient == nil {
		return fmt.Errorf("SOAP client not initialized (check router URI, username, and password)")
	}

	// Get QR code string from service layer
	qrData, err := services.GetWLANQRCode(soapClient, band)
	if err != nil {
		return fmt.Errorf("failed to get QR code: %w", err)
	}

	// Output based on format
	switch cfg.OutputFormat {
	case "json":
		type QRCodeOutput struct {
			Band   int    `json:"band"`
			QRData string `json:"qr_data"`
		}
		output := QRCodeOutput{
			Band:   band,
			QRData: qrData,
		}
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(output)
	default:
		// Text output
		fmt.Println("WLAN QR Code Data:")
		fmt.Println("===================")
		fmt.Printf("Band: %d\n", band)
		fmt.Printf("QR Data: %s\n", qrData)
		fmt.Println()
		fmt.Println("To generate a QR code image, use a QR code generator with the above data.")
		fmt.Println("Example with qrencode:")
		fmt.Printf("  echo '%s' | qrencode -t PNG -o wlan-qr.png\n", qrData)
	}
	return nil
}

// newWLANChannelCommand creates the WLAN channel command
func newWLANChannelCommand() *cobra.Command {
	var band int
	var setChannel string

	cmd := &cobra.Command{
		Use:   "channel",
		Short: "Get/set WLAN channel",
		Long:  `Gets or sets the WLAN channel for the specified band.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if setChannel != "" {
				// Write operation - do NOT test on real router
				return runSetWLANChannel(band, setChannel)
			}
			// Read operation - OK to test on real router
			return runGetWLANChannel(band)
		},
	}

	cmd.Flags().IntVarP(&band, "band", "b", 1, "WLAN band (1=2.4GHz, 2=5GHz, 3=5GHz ch2, 4=Guest)")
	cmd.Flags().StringVar(&setChannel, "set", "", "Set channel (write operation, use with caution)")
	return cmd
}

// runGetWLANChannel executes the get channel command (read-only)
func runGetWLANChannel(band int) error {
	if soapClient == nil {
		return fmt.Errorf("SOAP client not initialized (check router URI, username, and password)")
	}

	channel, err := services.GetWLANChannel(soapClient, band)
	if err != nil {
		return fmt.Errorf("failed to get WLAN channel: %w", err)
	}

	// Output based on format
	switch cfg.OutputFormat {
	case "json":
		type Output struct {
			Band    int    `json:"band"`
			Channel string `json:"channel"`
		}
		output := Output{Band: band, Channel: channel}
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(output)
	default:
		bandName := getBandName(band)
		fmt.Printf("WLAN Channel for %s\n", bandName)
		fmt.Println("=========================")
		fmt.Printf("Channel: %s\n", channel)
	}
	return nil
}

// runSetWLANChannel executes the set channel command (write operation)
// WARNING: Do NOT test on real router without caution
func runSetWLANChannel(band int, channel string) error {
	// This is a write operation - implement in service layer
	return fmt.Errorf("SetChannel not yet implemented in service layer")
}
