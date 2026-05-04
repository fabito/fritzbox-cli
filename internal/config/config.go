package config

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/viper"
)

// Config holds all configuration for fritzboxctl
type Config struct {
	RouterURI    string        `yaml:"router_uri" mapstructure:"router_uri"`
	Username     string        `yaml:"username" mapstructure:"username"`
	Password     string        `yaml:"password" mapstructure:"password"`
	RepeaterURI  string        `yaml:"repeater_uri" mapstructure:"repeater_uri"`
	RepeaterUser string        `yaml:"repeater_user" mapstructure:"repeater_user"`
	RepeaterPass string        `yaml:"repeater_password" mapstructure:"repeater_password"`
	Timeout      time.Duration `yaml:"timeout" mapstructure:"timeout"`
	OutputFormat string        `yaml:"output_format" mapstructure:"output_format"`
}

// Default values
const (
	DefaultRouterURI = "192.168.178.1"
	DefaultTimeout   = 10 * time.Second
	DefaultOutput    = "text"
	ConfigDir        = ".config/fritzboxctl"
	ConfigFile       = "config.yaml"
)

// Load loads configuration from all sources
// Precedence: flags > env vars > config file > defaults
func Load() (*Config, error) {
	v := viper.New()

	// Set defaults
	v.SetDefault("router_uri", DefaultRouterURI)
	v.SetDefault("timeout", DefaultTimeout)
	v.SetDefault("output_format", DefaultOutput)

	// Config file settings
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath(".")                                    // Current directory
	v.AddConfigPath(filepath.Join("$HOME", ConfigDir))      // ~/.config/fritzboxctl/
	v.AddConfigPath(filepath.Join("$HOME", ".fritzboxctl")) // Legacy location

	// Environment variable settings (auto-discovered)
	v.SetEnvPrefix("FRITZBOX") // Will match FRITZBOX_URI, FRITZBOX_USERNAME, etc.
	v.AutomaticEnv()

	// Manual env bindings for FritzBoxShell compatibility
	v.BindEnv("router_uri", "FRITZBOX_URI", "BOXIP", "FRITZBOX_IP")
	v.BindEnv("username", "FRITZBOX_USERNAME", "BOXUSER", "FRITZBOX_USER")
	v.BindEnv("password", "FRITZBOX_PASSWORD", "BOXPW", "FRITZBOX_PW")
	v.BindEnv("repeater_uri", "FRITZBOX_REPEATER_URI", "REPEATERIP")
	v.BindEnv("repeater_user", "FRITZBOX_REPEATER_USERNAME", "REPEATERUSER")
	v.BindEnv("repeater_password", "FRITZBOX_REPEATER_PASSWORD", "REPEATERPW")
	v.BindEnv("timeout", "FRITZBOX_TIMEOUT")
	v.BindEnv("output_format", "FRITZBOX_OUTPUT")

	// Read config file (optional)
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
		// Config file not found is OK - we'll use defaults/env/flags
	}

	// Unmarshal into Config struct
	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("error unmarshaling config: %w", err)
	}

	return &cfg, nil
}

// GetConfigFile returns the path to the config file
func GetConfigFile() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ConfigDir, ConfigFile)
}

// Save saves the configuration to the default config file
func (c *Config) Save() error {
	v := viper.New()
	v.Set("router_uri", c.RouterURI)
	v.Set("username", c.Username)
	v.Set("password", c.Password)
	v.Set("repeater_uri", c.RepeaterURI)
	v.Set("repeater_user", c.RepeaterUser)
	v.Set("repeater_password", c.RepeaterPass)
	v.Set("timeout", c.Timeout)
	v.Set("output_format", c.OutputFormat)

	configFile := GetConfigFile()
	dir := filepath.Dir(configFile)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("error creating config directory: %w", err)
	}

	return v.WriteConfigAs(configFile)
}
