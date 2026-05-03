package main

import (
	"fmt"
	"log/slog"

	"github.com/fabito/fritzboxctl/internal/soap/services"

	"github.com/spf13/cobra"
)

// newNetworkCommand creates the network command tree
func newNetworkCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "network",
		Short: "Network-related commands",
		Long:  `Commands for managing network settings (WAN, DSL, LAN).`,
	}

	cmd.AddCommand(newNetworkWANCommand())
	cmd.AddCommand(newNetworkDSLCommand())
	cmd.AddCommand(newNetworkLANCommand())

	return cmd
}

// newNetworkWANCommand creates the WAN command
func newNetworkWANCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "wan",
		Short: "WAN-related commands",
	}

	cmd.AddCommand(newNetworkWANStatusCommand())
	cmd.AddCommand(newNetworkWANReconnectCommand())

	return cmd
}

// newNetworkWANStatusCommand creates the WAN status command
func newNetworkWANStatusCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Get WAN connection status",
		Long:  `Retrieves WAN (internet) connection status such as connection state, external IP, uptime, etc.`,
		RunE:  runNetworkWANStatus,
	}
}

// runNetworkWANStatus executes the WAN status command
func runNetworkWANStatus(cmd *cobra.Command, args []string) error {
	// Call the service layer (thin CLI - just delegates)
	status, err := services.GetWANStatus(soapClient)
	if err != nil {
		return fmt.Errorf("failed to get WAN status: %w", err)
	}

	// Also get external IP (separate call)
	extIPResp, err := services.GetExternalIPAddress(soapClient)
	if err != nil {
		slog.Warn("Failed to get external IP", "error", err)
	}

	// Display the result (format output in cmd/)
	slog.Debug("WAN status retrieved", "status", status.NewConnectionStatus)

	switch cfg.OutputFormat {
	case "json":
		extIP := ""
		if extIPResp != nil {
			extIP = extIPResp.NewExternalIPAddress
		}
		fmt.Printf(`{"connection_status": "%s", "external_ip": "%s", "uptime": "%s"}\n`,
			status.NewConnectionStatus, extIP, status.NewUptime)
	default:
		fmt.Println("WAN Connection Status")
		fmt.Println("=============================")
		fmt.Printf("Connection:      %s\n", status.NewConnectionStatus)
		if extIPResp != nil && extIPResp.NewExternalIPAddress != "" {
			fmt.Printf("External IP:     %s\n", extIPResp.NewExternalIPAddress)
		}
		fmt.Printf("Uptime:          %s seconds\n", status.NewUptime)
		if status.NewLastConnectionError != "" {
			fmt.Printf("Last Error:      %s\n", status.NewLastConnectionError)
		}
		if status.NewDNSServers != "" {
			fmt.Printf("DNS Servers:     %s\n", status.NewDNSServers)
		}
	}

	return nil
}

// newNetworkWANReconnectCommand creates the WAN reconnect command
func newNetworkWANReconnectCommand() *cobra.Command {
	return &cobra.Command{
		Use:  "reconnect",
		Short: "Reconnect WAN (redial)",
		RunE: func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf("WAN reconnect command not yet implemented")
		},
	}
}

// newNetworkDSLCommand creates the DSL command
func newNetworkDSLCommand() *cobra.Command {
	return &cobra.Command{
		Use:  "dsl",
		Short: "DSL-related commands",
	}
}

// newNetworkLANCommand creates the LAN command
func newNetworkLANCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:  "lan",
		Short: "LAN-related commands",
	}

	cmd.AddCommand(&cobra.Command{
		Use:  "stats",
		RunE: func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf("LAN stats command not yet implemented")
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:  "count",
		RunE: func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf("LAN count command not yet implemented")
		},
	})

	return cmd
}
