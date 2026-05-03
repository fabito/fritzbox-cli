package main

import (
	"fmt"

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

	cmd.AddCommand(&cobra.Command{
		Use:  "status",
		RunE: func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf("WAN status command not yet implemented")
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:  "reconnect",
		RunE: func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf("WAN reconnect command not yet implemented")
		},
	})

	return cmd
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
