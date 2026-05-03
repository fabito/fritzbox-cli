package main

import (
	"fmt"
	"strings"

	"github.com/fabito/fritzboxctl/internal/soap/services"

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
