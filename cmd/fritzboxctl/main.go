package main

import (
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/fabito/fritzboxctl/internal/auth"
	"github.com/fabito/fritzboxctl/internal/config"
	"github.com/fabito/fritzboxctl/internal/soap"
)

var (
	// Global flags
	routerURI    string
	username     string
	password     string
	timeout      int
	outputFormat string
	configFile   string
	verbose      bool

	// Internal clients
	soapClient *soap.Client
	authObj    *auth.Auth
	cfg        *config.Config
)

func main() {
	if err := NewRootCommand().Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// NewRootCommand creates the root cobra command
func NewRootCommand() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "fritzboxctl",
		Short: "A modern CLI tool for controlling AVM Fritz!Box routers",
		Long: `fritzboxctl is a modern Go CLI tool for controlling AVM Fritz!Box routers.
It provides a type-safe, efficient, and user-friendly interface to manage
Fritz!Box devices via TR-064 (SOAP/UPnP) and AHA-HTTP protocols.`,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return initializeClients(cmd)
		},
	}

	// Global flags
	rootCmd.PersistentFlags().StringVarP(&routerURI, "router-uri", "H", "", "Fritz!Box URI (IP or hostname) (env: FRITZBOX_URI)")
	rootCmd.PersistentFlags().StringVarP(&username, "username", "u", "", "Username (env: FRITZBOX_USERNAME)")
	rootCmd.PersistentFlags().StringVarP(&password, "password", "p", "", "Password (env: FRITZBOX_PASSWORD)")
	rootCmd.PersistentFlags().IntVarP(&timeout, "timeout", "t", 0, "Request timeout in seconds (env: FRITZBOX_TIMEOUT)")
	rootCmd.PersistentFlags().StringVarP(&outputFormat, "output", "o", "", "Output format (text, json, yaml) (env: FRITZBOX_OUTPUT)")
	rootCmd.PersistentFlags().StringVarP(&configFile, "config", "c", "", "Config file path (env: FRITZBOX_CONFIG)")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose/debug output")

	// Add subcommands
	rootCmd.AddCommand(newDeviceCommand())
	rootCmd.AddCommand(newWLANCommand())
	rootCmd.AddCommand(newNetworkCommand())

	return rootCmd
}

// initializeClients initializes the authentication and SOAP clients
func initializeClients(cmd *cobra.Command) error {
	// Initialize slog based on verbose flag
	if verbose {
		slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug})))
	} else {
		slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelWarn})))
	}

	// Load configuration
	var err error
	cfg, err = config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Override with CLI flags if provided
	if routerURI != "" {
		cfg.RouterURI = routerURI
	}
	if username != "" {
		cfg.Username = username
	}
	if password != "" {
		cfg.Password = password
	}
	if timeout > 0 {
		cfg.Timeout = time.Duration(timeout) * time.Second
	}
	if outputFormat != "" {
		cfg.OutputFormat = outputFormat
	}

	// Create auth module
	authObj, err = auth.NewAuth(cfg.Username, cfg.Password, cfg.Timeout, true)
	if err != nil {
		return fmt.Errorf("failed to create auth: %w", err)
	}

	// Create SOAP client
	soapClient = soap.NewClient(
		fmt.Sprintf("http://%s:49000", cfg.RouterURI),
		authObj.DigestClient,
		authObj.HTTPClient,
	)

	// Set up SID manager
	soapCallFunc := soapClient.GetSoapCallFunc()
	authObj.SetSIDManager(auth.NewSIDManager(soapCallFunc))

	// Discover services
	if err := soapClient.Discover(); err != nil {
		// Non-fatal, some services may not be available
		fmt.Fprintf(os.Stderr, "Warning: Failed to discover services: %v\n", err)
	}

	return nil
}
