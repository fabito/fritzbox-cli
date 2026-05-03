package main

import (
	"fmt"
	"strings"

	"github.com/fabito/fritzboxctl/internal/aha"
	"github.com/spf13/cobra"
)

// newVPNCommand creates the VPN command tree
func newVPNCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "vpn",
		Short: "VPN-related commands",
		Long:  `Commands for managing VPN connections (WireGuard and IPsec).`,
	}

	cmd.AddCommand(newVPNListCommand())

	return cmd
}

// newVPNListCommand creates the VPN list command
func newVPNListCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List VPN connections",
		Long:  `Lists all VPN connections (WireGuard and IPsec) and their status.`,
		RunE:  runVPNList,
	}
}

// runVPNList executes the VPN list command
func runVPNList(cmd *cobra.Command, args []string) error {
	// Get SID first
	sid, err := authObj.GetSID()
	if err != nil {
		return fmt.Errorf("failed to get SID: %w", err)
	}

	// Create AHA client
	client := aha.NewClient(cfg.RouterURI, sid)

	// List VPN connections
	connections, err := client.ListVPNConnections()
	if err != nil {
		return fmt.Errorf("failed to list VPN connections: %w", err)
	}

	// Display based on output format
	switch cfg.OutputFormat {
	case "json":
		fmt.Println(`{"vpn_connections": "json output not yet implemented"}`)
	default:
		// Text output
		fmt.Println("VPN Connections")
		fmt.Println("================")
		if len(connections) == 0 {
			fmt.Println("No VPN connections found.")
			return nil
		}
		// Table header
		fmt.Printf("%-20s %-10s %-15s %-12s\n", "Name", "Type", "Status", "Enabled")
		fmt.Println(strings.Repeat("-", 60))
		for _, conn := range connections {
			enabled := "No"
			if conn.Enabled {
				enabled = "Yes"
			}
			fmt.Printf("%-20s %-10s %-15s %-12s\n", conn.Name, conn.Type, conn.Status, enabled)
		}
	}

	return nil
}
