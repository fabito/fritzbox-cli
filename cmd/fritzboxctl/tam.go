package main

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/fabito/fritzboxctl/internal/soap/services"

	"github.com/spf13/cobra"
)

// newTAMCommand creates the TAM (Answering Machine) command tree
func newTAMCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tam",
		Short: "TAM (Answering Machine) commands",
		Long:  `Commands for managing TAM (Answering Machine) on the Fritz!Box.`,
	}

	cmd.AddCommand(newTAMListCommand())
	cmd.AddCommand(newTAMSetCommand())

	return cmd
}

// newTAMListCommand creates the TAM list command
func newTAMListCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List TAM devices and their status",
		Long:  `Lists all TAM (Answering Machine) devices connected to the Fritz!Box, showing their status and message counts.`,
		RunE:  runTAMList,
	}
}

// runTAMList executes the TAM list command
func runTAMList(cmd *cobra.Command, args []string) error {
	// Call the services layer to get TAM list
	tams, err := services.ListTAMs(soapClient)
	if err != nil {
		return fmt.Errorf("failed to list TAMs: %w", err)
	}

	if len(tams) == 0 {
		fmt.Println("No TAM devices found.")
		return nil
	}

	// Display based on output format
	switch cfg.OutputFormat {
	case "json":
		// JSON output (simplified)
		fmt.Println("[")
		for i, tam := range tams {
			enabled := false
			if tam.NewTAMEnable == "1" {
				enabled = true
			}
			fmt.Printf(`  {"index": %s, "name": "%s", "enabled": %v, "new_messages": %s, "total_messages": %s}`,
				tam.NewTAMIndex, tam.NewTAMName, enabled, tam.NewTAMNewMessageCount, tam.NewTAMTotalMessageCount)
			if i < len(tams)-1 {
				fmt.Println(",")
			} else {
				fmt.Println()
			}
		}
		fmt.Println("]")
	default:
		// Text output
		fmt.Println("TAM Devices")
		fmt.Println("===========")
		fmt.Printf("%-5s %-25s %-8s %-15s %-15s\n", "Idx", "Name", "Status", "New Msgs", "Total Msgs")
		fmt.Println(strings.Repeat("-", 70))

		for _, tam := range tams {
			status := "Disabled"
			if tam.NewTAMEnable == "1" {
				status = "Enabled"
			}

			newMsgs := tam.NewTAMNewMessageCount
			if newMsgs == "0" || newMsgs == "" {
				newMsgs = "-"
			}

			totalMsgs := tam.NewTAMTotalMessageCount
			if totalMsgs == "0" || totalMsgs == "" {
				totalMsgs = "-"
			}

			idx, _ := strconv.Atoi(tam.NewTAMIndex)
			fmt.Printf("%-5d %-25s %-8s %-15s %-15s\n", idx, tam.NewTAMName, status, newMsgs, totalMsgs)
		}
	}

	return nil
}

// newTAMSetCommand creates the TAM set command (enable/disable)
func newTAMSetCommand() *cobra.Command {
	var index int
	var on bool
	var off bool

	cmd := &cobra.Command{
		Use:   "set",
		Short: "Enable or disable a TAM (Answering Machine)",
		Long:  `Enables or disables a specific TAM (Answering Machine) on the Fritz!Box.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Determine enabled state from --on/--off flags
			if !on && !off {
				return fmt.Errorf("must specify either --on or --off")
			}
			if on && off {
				return fmt.Errorf("--on and --off are mutually exclusive")
			}
			enabled := on
			return runTAMSet(index, enabled)
		},
	}

	cmd.Flags().IntVarP(&index, "index", "i", 0, "TAM index (default: 0)")
	cmd.Flags().BoolVar(&on, "on", false, "Enable the TAM")
	cmd.Flags().BoolVar(&off, "off", false, "Disable the TAM")

	return cmd
}

// runTAMSet executes the TAM set command
func runTAMSet(index int, enabled bool) error {
	err := services.SetTAMEnabled(index, enabled, soapClient)
	if err != nil {
		return fmt.Errorf("failed to set TAM: %w", err)
	}

	status := "disabled"
	if enabled {
		status = "enabled"
	}
	fmt.Printf("TAM %d %s successfully.\n", index, status)
	return nil
}
