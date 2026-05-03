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
	cmd.AddCommand(newVPNSetCommand())

	return cmd
}

// newVPNSetCommand creates the VPN set command
func newVPNSetCommand() *cobra.Command {
	var enabled bool
	cmd := &cobra.Command{
		Use:   "set",
		Short: "Enable or disable a VPN connection",
		Long:  `Enables or disables a VPN connection by name.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 {
				return fmt.Errorf("please specify a VPN connection name")
			}
			name := args[0]
			return runVPNSet(name, enabled)
		},
	}
	cmd.Flags().BoolVar(&enabled, "on", false, "Enable the VPN connection")
	cmd.Flags().BoolVar(&enabled, "off", false, "Disable the VPN connection")
	// Custom help to show on/off options
	cmd.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		fmt.Fprintf(cmd.OutOrStdout(), "Usage:\n  %s [flags] <name>\n\nFlags:\n  --on\tEnable the VPN connection\n  --off\tDisable the VPN connection\n", cmd.CommandPath())
	})
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
	// Create AHA client with getSID function
	routerURI := cfg.RouterURI
	if !strings.HasPrefix(routerURI, "http://") && !strings.HasPrefix(routerURI, "https://") {
		routerURI = "http://" + routerURI
	}
	
	client := aha.NewClient(routerURI, func() (string, error) {
		return authObj.SIDManager.GetSID()
	})

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

// runVPNSet executes the VPN set command
func runVPNSet(name string, enabled bool) error {
	// Create AHA client with getSID function
	routerURI := cfg.RouterURI
	if !strings.HasPrefix(routerURI, "http://") && !strings.HasPrefix(routerURI, "https://") {
		routerURI = "http://" + routerURI
	}

	client := aha.NewClient(routerURI, func() (string, error) {
		return authObj.SIDManager.GetSID()
	})

	// Call SetVPNEnabled
	err := client.SetVPNEnabled(name, enabled)
	if err != nil {
		return fmt.Errorf("failed to set VPN: %w", err)
	}

	// Output result
	switch cfg.OutputFormat {
	case "json":
		fmt.Printf(`{"name": "%s", "enabled": %v}\n`, name, enabled)
	default:
		state := "disabled"
		if enabled {
			state = "enabled"
		}
		fmt.Printf("VPN connection '%s' %s successfully.\n", name, state)
	}

	return nil
}
