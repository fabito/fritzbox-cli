package main

import (
	"encoding/xml"
	"fmt"
	"os"
	"strings"

	"github.com/fabito/fritzboxctl/internal/soap"
	"github.com/fabito/fritzboxctl/internal/soap/services"

	"github.com/spf13/cobra"
)

// HostEntry represents a device on the network (for display only)
type HostEntry struct {
	HostName      string
	IPAddress     string
	MACAddress    string
	InterfaceType string
	Active        bool
}

// formatHostList formats a list of hosts for display
func formatHostList(hosts []HostEntry) string {
	if len(hosts) == 0 {
		return "No devices found on the network.\n"
	}

	var result string
	result += fmt.Sprintf("%-20s %-15s %-17s %-6s\n", "Hostname", "IP", "MAC", "Type")
	result += fmt.Sprintf("%s\n", strings.Repeat("-", 60))

	for _, host := range hosts {
		hostname := host.HostName
		if hostname == "" {
			hostname = "(unknown)"
		}
		active := ""
		if !host.Active {
			active = "(inactive)"
		}
		result += fmt.Sprintf("%-20s %-15s %-17s %-6s %s\n", hostname, host.IPAddress, host.MACAddress, host.InterfaceType, active)
	}

	return result
}

// newDeviceCommand creates the device command tree
func newDeviceCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "device",
		Short: "Device-related commands",
		Long:  `Commands for interacting with the Fritz!Box device itself.`,
	}

	cmd.AddCommand(newDeviceInfoCommand())
	cmd.AddCommand(newDeviceListCommand())
	cmd.AddCommand(newDeviceBlockCommand())
	cmd.AddCommand(newDeviceUnblockCommand())
	cmd.AddCommand(newDeviceRebootCommand())
	cmd.AddCommand(newDeviceBackupCommand())

	return cmd
}

// newDeviceInfoCommand creates the device info command
func newDeviceInfoCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "info",
		Short: "Get device information",
		Long:  `Retrieves information about the Fritz!Box such as model name, firmware version, serial number, etc.`,
		RunE:  runDeviceInfo,
	}
}

// runDeviceInfo executes the device info command
func runDeviceInfo(cmd *cobra.Command, args []string) error {
	// Call the service layer to get device info
	info, err := services.GetDeviceInfo(soapClient)
	if err != nil {
		return fmt.Errorf("failed to get device info: %w", err)
	}

	// Display based on output format
	switch cfg.OutputFormat {
	case "json":
		// JSON output
		fmt.Printf(`{"model": "%s", "serial": "%s", "firmware": "%s"}\n`,
			info.NewModelName,
			info.NewSerialNumber,
			info.NewSoftwareVersion)
	default:
		// Text output
		fmt.Println("Fritz!Box Device Information")
		fmt.Println("============================")
		fmt.Printf("Model:          %s\n", info.NewModelName)
		fmt.Printf("Model Number:    %s\n", info.NewModelNumber)
		fmt.Printf("Serial Number:   %s\n", info.NewSerialNumber)
		fmt.Printf("Firmware:        %s\n", info.NewSoftwareVersion)
		fmt.Printf("Hardware:        %s\n", info.NewHardwareVersion)
		fmt.Printf("Manufacturer:    %s\n", info.NewManufacturerName)
		fmt.Printf("Description:     %s\n", info.NewDescription)
	}

	return nil
}

// newDeviceListCommand creates the device list command
func newDeviceListCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List devices on the network",
		Long:  `Lists all devices connected to the Fritz!Box (LAN and WLAN).`,
		RunE:  runDeviceList,
	}
}

// runDeviceList executes the device list command
func runDeviceList(cmd *cobra.Command, args []string) error {
	// Step 1: Get number of hosts
	soapBody := `<?xml version="1.0" encoding="utf-8"?>
<s:Envelope s:encodingStyle="http://schemas.xmlsoap.org/soap/encoding/" xmlns:s="http://schemas.xmlsoap.org/soap/envelope/">
    <s:Body>
        <u:GetHostNumberOfEntries xmlns:u="urn:dslforum-org:service:Hosts:1">
        </u:GetHostNumberOfEntries>
    </s:Body>
</s:Envelope>`

	resp, err := soapClient.Call(
		"/upnp/control/hosts",
		"urn:dslforum-org:service:Hosts:1#GetHostNumberOfEntries",
		soapBody,
	)
	if err != nil {
		return fmt.Errorf("failed to get host count: %w", err)
	}

	// Parse number of hosts using shared CleanSoapResponse
	resp = soap.CleanSoapResponse(resp)
	var numResponse services.HostNumberResponse
	if err := xml.Unmarshal([]byte(resp), &numResponse); err != nil {
		return fmt.Errorf("failed to parse host count: %w", err)
	}

	// Convert to int
	numHosts := 0
	fmt.Sscanf(numResponse.NewHostNumberOfEntries, "%d", &numHosts)

	if numHosts == 0 {
		fmt.Println("No devices found on the network.")
		return nil
	}

	// Step 2: Get each host entry
	hosts := make([]HostEntry, 0, numHosts)
	for i := 1; i <= numHosts; i++ {
		soapBody = fmt.Sprintf(`<?xml version="1.0" encoding="utf-8"?>
<s:Envelope s:encodingStyle="http://schemas.xmlsoap.org/soap/encoding/" xmlns:s="http://schemas.xmlsoap.org/soap/envelope/">
    <s:Body>
        <u:GetGenericHostEntry xmlns:u="urn:dslforum-org:service:Hosts:1">
            <NewIndex>%d</NewIndex>
        </u:GetGenericHostEntry>
    </s:Body>
</s:Envelope>`, i)

		resp, err = soapClient.Call(
			"/upnp/control/hosts",
			"urn:dslforum-org:service:Hosts:1#GetGenericHostEntry",
			soapBody,
		)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Failed to get host %d: %v\n", i, err)
			continue
		}

		// Parse host entry using shared CleanSoapResponse
		resp = soap.CleanSoapResponse(resp)
		var hostResp services.HostListResponse
		if err := xml.Unmarshal([]byte(resp), &hostResp); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Failed to parse host %d: %v\n", i, err)
			continue
		}

		// Convert to HostEntry for display
		active := hostResp.NewActive == "1"
		host := HostEntry{
			HostName:      hostResp.NewHostName,
			IPAddress:     hostResp.NewIPAddress,
			MACAddress:    hostResp.NewMACAddress,
			InterfaceType: hostResp.NewInterfaceType,
			Active:        active,
		}
		hosts = append(hosts, host)
	}

	// Display the list
	fmt.Print(formatHostList(hosts))
	return nil
}

// newDeviceBlockCommand creates the device block command
func newDeviceBlockCommand() *cobra.Command {
	var ipAddress string
	cmd := &cobra.Command{
		Use:   "block",
		Short: "Block device internet access",
		Long:  "Blocks internet access for a device by IP address.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDeviceBlock(ipAddress)
		},
	}
	cmd.Flags().StringVarP(&ipAddress, "ip", "i", "", "IP address of the device to block")
	cmd.MarkFlagRequired("ip")
	return cmd
}

// runDeviceBlock executes the device block command
func runDeviceBlock(ipAddress string) error {
	err := services.BlockDevice(ipAddress, soapClient)
	if err != nil {
		return fmt.Errorf("failed to block device: %w", err)
	}
	fmt.Printf("Device %s blocked successfully.\n", ipAddress)
	return nil
}

// newDeviceUnblockCommand creates the device unblock command
func newDeviceUnblockCommand() *cobra.Command {
	var ipAddress string
	cmd := &cobra.Command{
		Use:   "unblock",
		Short: "Unblock device internet access",
		Long:  "Unblocks internet access for a device by IP address.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDeviceUnblock(ipAddress)
		},
	}
	cmd.Flags().StringVarP(&ipAddress, "ip", "i", "", "IP address of the device to unblock")
	cmd.MarkFlagRequired("ip")
	return cmd
}

// runDeviceUnblock executes the device unblock command
func runDeviceUnblock(ipAddress string) error {
	err := services.UnblockDevice(ipAddress, soapClient)
	if err != nil {
		return fmt.Errorf("failed to unblock device: %w", err)
	}
	fmt.Printf("Device %s unblocked successfully.\n", ipAddress)
	return nil
}

// newDeviceRebootCommand creates the reboot command
func newDeviceRebootCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "reboot",
		Short: "Reboot the Fritz!Box",
		RunE:  runDeviceReboot,
	}
}

// runDeviceReboot executes the reboot command
func runDeviceReboot(cmd *cobra.Command, args []string) error {
	soapBody := `<?xml version="1.0" encoding="utf-8"?>
<s:Envelope s:encodingStyle="http://schemas.xmlsoap.org/soap/encoding/" xmlns:s="http://schemas.xmlsoap.org/soap/envelope/">
    <s:Body>
        <u:Reboot xmlns:u="urn:dslforum-org:service:DeviceConfig:1">
        </u:Reboot>
    </s:Body>
</s:Envelope>`

	_, err := soapClient.Call(
		"/upnp/control/deviceconfig",
		"urn:dslforum-org:service:DeviceConfig:1#Reboot",
		soapBody,
	)
	if err != nil {
		return fmt.Errorf("failed to reboot: %w", err)
	}

	fmt.Println("Reboot command sent successfully.")
	fmt.Println("The Fritz!Box should reboot within a few seconds.")
	return nil
}

// newDeviceBackupCommand creates the backup command
func newDeviceBackupCommand() *cobra.Command {
	var backupPassword string
	cmd := &cobra.Command{
		Use:   "backup",
		Short: "Backup Fritz!Box configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDeviceBackup(backupPassword)
		},
	}
	cmd.Flags().StringVarP(&backupPassword, "password", "p", "", "Password for the backup file")
	return cmd
}

// runDeviceBackup executes the backup command
func runDeviceBackup(backupPassword string) error {
	if backupPassword == "" {
		return fmt.Errorf("backup password is required (use --password flag)")
	}

	// First get the security port
	soapBody := `<?xml version="1.0" encoding="utf-8"?>
<s:Envelope s:encodingStyle="http://schemas.xmlsoap.org/soap/encoding/" xmlns:s="http://schemas.xmlsoap.org/soap/envelope/">
    <s:Body>
        <u:GetSecurityPort xmlns:u="urn:dslforum-org:service:DeviceInfo:1">
        </u:GetSecurityPort>
    </s:Body>
</s:Envelope>`

	resp, err := soapClient.Call(
		"/upnp/control/deviceinfo",
		"urn:dslforum-org:service:DeviceInfo:1#GetSecurityPort",
		soapBody,
	)
	if err != nil {
		return fmt.Errorf("failed to get security port: %w", err)
	}

	// Parse security port
	securityPort := parseSecurityPort(resp)
	if securityPort == "" {
		return fmt.Errorf("could not determine security port")
	}

	// Get the config file URL
	soapBody = fmt.Sprintf(`<?xml version="1.0" encoding="utf-8"?>
<s:Envelope s:encodingStyle="http://schemas.xmlsoap.org/soap/encoding/" xmlns:s="http://schemas.xmlsoap.org/soap/envelope/">
    <s:Body>
        <u:X_AVM-DE_GetConfigFile xmlns:u="urn:dslforum-org:service:DeviceConfig:1">
            <NewX_AVM-DE_Password>%s</NewX_AVM-DE_Password>
        </u:X_AVM-DE_GetConfigFile>
    </s:Body>
</s:Envelope>`, backupPassword)

	resp, err = soapClient.Call(
		"/upnp/control/deviceconfig",
		"urn:dslforum-org:service:DeviceConfig:1#X_AVM-DE_GetConfigFile",
		soapBody,
	)
	if err != nil {
		return fmt.Errorf("failed to get config file URL: %w", err)
	}

	// Parse download URL and download
	downloadURL := parseConfigFileURL(resp)
	if downloadURL == "" {
		return fmt.Errorf("could not get config file download URL")
	}

	fmt.Printf("Backup URL obtained. Downloading from: %s\n", downloadURL)
	fmt.Println("Backup functionality is partially implemented. Full download requires additional work.")
	return nil
}

// parseSecurityPort extracts the security port from SOAP response
func parseSecurityPort(resp string) string {
	// Simple parsing
	idx := strings.Index(resp, "NewSecurityPort")
	if idx == -1 {
		return ""
	}
	// Extract the port number
	start := strings.Index(resp[idx:], ">")
	if start == -1 {
		return ""
	}
	start += idx + 1
	end := strings.Index(resp[start:], "<")
	if end == -1 {
		return ""
	}
	return resp[start : start+end]
}

// parseConfigFileURL extracts the config file URL from SOAP response
func parseConfigFileURL(resp string) string {
	idx := strings.Index(resp, "https://")
	if idx == -1 {
		return ""
	}
	end := strings.Index(resp[idx:], "<")
	if end == -1 {
		return resp[idx:]
	}
	return resp[idx : idx+end]
}
