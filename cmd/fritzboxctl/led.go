package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

// newLEDCommand creates the LED command tree
func newLEDCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "led",
		Short: "LED-related commands",
		Long:  `Commands for managing the Fritz!Box LED display.`,
	}

	cmd.AddCommand(newLEDStatusCommand())
	cmd.AddCommand(newLEDSetCommand())

	return cmd
}

// newLEDStatusCommand creates the LED status command
func newLEDStatusCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Get LED status",
		Long:  `Shows the current LED status (on/off) and brightness if supported.`,
		RunE:  runLEDStatus,
	}
}

// runLEDStatus executes the LED status command
func runLEDStatus(cmd *cobra.Command, args []string) error {
	// Get LED status using AHA client
	status, err := ahaClient.GetLEDStatus()
	if err != nil {
		return fmt.Errorf("failed to get LED status: %w", err)
	}

	// Display based on output format
	switch cfg.OutputFormat {
	case "json":
		fmt.Printf(`{"led_display": %d, "can_dim": %d, "dim_value": %d}`+"\n",
			status.LEDDisplay, status.CanDim, status.DimValue)
	default:
		// Text output
		fmt.Println("LED Status")
		fmt.Println("==========")
		
		// LEDDisplay: 0=ON, 2=OFF
		ledState := "ON"
		if status.LEDDisplay == 2 {
			ledState = "OFF"
		}
		fmt.Printf("LED State:      %s\n", ledState)
		
		if status.CanDim > 0 {
			fmt.Printf("Dimming:        Supported (level %d)\n", status.DimValue)
		} else {
			fmt.Println("Dimming:        Not supported")
		}
	}

	return nil
}

// newLEDSetCommand creates the LED set command
func newLEDSetCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "set [on|off]",
		Short: "Set LED state (on/off)",
		Long:  `Enables or disables the Fritz!Box LEDs.`,
		Args:  cobra.ExactArgs(1),
		RunE:  runLEDSet,
	}
}

// runLEDSet executes the LED set command
func runLEDSet(cmd *cobra.Command, args []string) error {
	// Parse the argument
	var enabled bool
	switch args[0] {
	case "on":
		enabled = true
	case "off":
		enabled = false
	default:
		return fmt.Errorf("invalid argument: %s (use 'on' or 'off')", args[0])
	}

	// Call AHA client to set LED state (thin layer - only CLI handling here)
	err := ahaClient.SetLEDEnabled(enabled)
	if err != nil {
		return fmt.Errorf("failed to set LED state: %w", err)
	}

	// Output result (formatting only - no business logic)
	switch cfg.OutputFormat {
	case "json":
		fmt.Printf(`{"enabled": %v}`+"\n", enabled)
	default:
		state := "OFF"
		if enabled {
			state = "ON"
		}
		fmt.Printf("LEDs switched %s\n", state)
	}

	return nil
}
