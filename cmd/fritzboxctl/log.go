package main

import (
	"fmt"
	"strings"

	"github.com/fabito/fritzboxctl/internal/soap/services"

	"github.com/spf13/cobra"
)

// newLogCommand creates the log command tree
func newLogCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "log",
		Short: "Event log commands",
		Long:  `Commands for viewing the Fritz!Box event log.`,
	}

	cmd.AddCommand(newLogReadCommand())
	cmd.AddCommand(newLogClearCommand())

	return cmd
}

// newLogReadCommand creates the log read command
func newLogReadCommand() *cobra.Command {
	var lines int
	cmd := &cobra.Command{
		Use:   "read",
		Short: "Read the event log",
		Long:  `Reads and displays the Fritz!Box event log.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runLogRead(lines)
		},
	}
	cmd.Flags().IntVarP(&lines, "lines", "n", 0, "Number of recent lines to show (0 = all)")
	return cmd
}

// runLogRead executes the log read command
func runLogRead(lines int) error {
	log, err := services.GetEventLog(soapClient)
	if err != nil {
		return fmt.Errorf("failed to get event log: %w", err)
	}

	// Display based on output format
	switch cfg.OutputFormat {
	case "json":
		fmt.Printf(`{"log": "%s"}\n`, strings.ReplaceAll(log, "\n", "\\n"))
	default:
		// Text output
		fmt.Println("Fritz!Box Event Log")
		fmt.Println("======================")
		if lines > 0 {
			// Show only the last N lines
			allLines := strings.Split(log, "\n")
			start := len(allLines) - lines
			if start < 0 {
				start = 0
			}
			for _, line := range allLines[start:] {
				fmt.Println(line)
			}
		} else {
			fmt.Println(log)
		}
	}

	return nil
}

// newLogClearCommand creates the log clear command
func newLogClearCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "clear",
		Short: "Clear the event log",
		Long:  `Clears the Fritz!Box event log. (WRITE OPERATION - use with caution)`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf("log clear is a write operation - not implemented in Phase 2 (read-only)")
		},
	}
}
