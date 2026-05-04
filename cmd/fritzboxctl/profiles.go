package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/fabito/fritzboxctl/internal/aha"

	"github.com/spf13/cobra"
)

// newDeviceProfilesCommand creates the device profiles command tree
func newDeviceProfilesCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "profiles",
		Short: "Device profile commands",
		Long:  `List and manage device profiles (parental control).`,
	}

	cmd.AddCommand(newDeviceProfilesListCommand())
	cmd.AddCommand(newDeviceProfilesSetCommand())

	return cmd
}

// newDeviceProfilesListCommand creates the profiles list command
func newDeviceProfilesListCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List profiles and devices",
		Long:  `Lists all available profiles and devices with their assigned profiles.`,
		RunE:  runDeviceProfilesList,
	}
}

// runDeviceProfilesList executes the profiles list command
func runDeviceProfilesList(cmd *cobra.Command, args []string) error {
	// Get AHA client
	if ahaClient == nil {
		return fmt.Errorf("AHA client not initialized (check router URI, username, and password)")
	}

	// First, list available profiles
	profiles, err := ahaClient.ListAvailableProfiles()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Failed to list available profiles: %v\n", err)
	}

	// Then, list devices with their profiles
	devices, err := ahaClient.ListDevicesWithProfiles()
	if err != nil {
		return fmt.Errorf("failed to list devices: %w", err)
	}

	// Output based on format
	switch cfg.OutputFormat {
	case "json":
		return outputProfilesJSON(profiles, devices)
	default:
		return outputProfilesText(profiles, devices)
	}
}

// outputProfilesJSON outputs profiles in JSON format
func outputProfilesJSON(profiles []aha.Profile, devices []aha.DeviceProfile) error {
	// Combine into a single struct
	type Output struct {
		Profiles []aha.Profile       `json:"profiles"`
		Devices  []aha.DeviceProfile `json:"devices"`
	}

	output := Output{
		Profiles: profiles,
		Devices:  devices,
	}

	// Marshal to JSON
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(output)
}

// outputProfilesText outputs profiles in text format
func outputProfilesText(profiles []aha.Profile, devices []aha.DeviceProfile) error {
	// First, show available profiles
	fmt.Println("Available Profiles")
	fmt.Println("===================")
	if len(profiles) == 0 {
		fmt.Println("No profiles found.")
	} else {
		for _, p := range profiles {
			fmt.Printf("  ID: %s - %s\n", p.ID, p.Name)
		}
	}

	fmt.Println()

	// Then, show devices with their profiles
	fmt.Println("Devices with Profiles")
	fmt.Println("=====================")
	if len(devices) == 0 {
		fmt.Println("No devices found.")
		return nil
	}

	// Print header
	fmt.Printf("%-20s %-17s %-15s %-20s\n", "Device Name", "MAC", "IP", "Profile")
	fmt.Println(strings.Repeat("-", 75))

	// Print devices
	for _, d := range devices {
		name := d.DeviceName
		if name == "" {
			name = "(unknown)"
		}
		profile := d.ProfileName
		if profile == "" {
			profile = "(none)"
		}
		fmt.Printf("%-20s %-17s %-15s %-20s\n", name, d.MACAddress, d.IPAddress, profile)
	}

	return nil
}

// newDeviceProfilesSetCommand creates the profiles set command
func newDeviceProfilesSetCommand() *cobra.Command {
	var deviceID string
	var profileID string

	cmd := &cobra.Command{
		Use:   "set",
		Short: "Set profile for a device",
		Long:  `Sets the profile (parental control) for a specific device.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDeviceProfilesSet(deviceID, profileID)
		},
	}

	cmd.Flags().StringVar(&deviceID, "device-id", "", "Device ID (LAN device ID from Fritz!Box)")
	cmd.Flags().StringVar(&profileID, "profile-id", "", "Profile ID (numeric, e.g., 1, 2, 3)")
	cmd.MarkFlagRequired("device-id")
	cmd.MarkFlagRequired("profile-id")

	return cmd
}

// runDeviceProfilesSet executes the profiles set command
func runDeviceProfilesSet(deviceID, profileID string) error {
	// Validate profile ID is numeric
	if _, err := strconv.Atoi(profileID); err != nil {
		return fmt.Errorf("invalid profile ID: must be numeric (e.g., 1, 2, 3)")
	}

	// Get AHA client
	if ahaClient == nil {
		return fmt.Errorf("AHA client not initialized (check router URI, username, and password)")
	}

	// Set profile
	err := ahaClient.SetDeviceProfile(deviceID, profileID)
	if err != nil {
		return fmt.Errorf("failed to set device profile: %w", err)
	}

	fmt.Printf("Profile %s set for device %s\n", profileID, deviceID)
	return nil
}
