# fritzboxctl - Design Document

## Overview

**fritzboxctl** is a modern Go CLI tool for controlling AVM Fritz!Box routers, inspired by the [FritzBoxShell](https://github.com/jhubig/FritzBoxShell) bash script. It provides a type-safe, efficient, and user-friendly interface to manage Fritz!Box devices via TR-064 (SOAP/UPnP) and AHA-HTTP protocols.

## Technical Analysis of FritzBoxShell

### Protocols Used

1. **TR-064 (SOAP/UPnP)** - Primary protocol
   - Port: 49000 (HTTP), 49443 (HTTPS)
   - Services discovered via `tr64desc.xml` and `igddesc.xml`
   - Used for: WLAN control, device info, statistics, TAM, WAN/DSL operations, reboot, backup

2. **AHA-HTTP-Interface** - Secondary protocol
   - Uses session ID (SID) obtained via `X_AVM-DE_CreateUrlSID`
   - Used for: LED control, KeyLock, signal strength, VPN, device profiles, event log

3. **Mesh API**
   - Endpoint: `meshlist.lua` via security port
   - Used for: Device listing across the mesh network

### Key TR-064 Services and Actions

| Service | Service Type | Common Actions |
|---------|--------------|----------------|
| DeviceInfo:1 | urn:dslforum-org:service:DeviceInfo:1 | GetInfo, GetSecurityPort |
| WLANConfiguration:1 | urn:dslforum-org:service:WLANConfiguration:1 | GetInfo, SetEnable, GetStatistics, GetTotalAssociations, SetChannel, GetSSID, GetSecurityKeys |
| WLANConfiguration:2 | urn:dslforum-org:service:WLANConfiguration:2 | Same as above (5GHz) |
| WLANConfiguration:3 | urn:dslforum-org:service:WLANConfiguration:3 | Same as above (5GHz 2nd channel) |
| Hosts:1 | urn:dslforum-org:service:Hosts:1 | GetHostNumberOfEntries, GetGenericHostEntry, X_AVM-DE_GetMeshListPath, X_AVM-DE_WakeOnLANByMACAddress |
| WANIPConnection:1 | urn:dslforum-org:service:WANIPConnection:1 | GetExternalIPAddress, ForceTermination, GetConnectionStatus |
| WANCommonInterfaceConfig:1 | urn:dslforum-org:service:WANCommonInterfaceConfig:1 | GetTotalBytesReceived/Sent, GetCommonLinkProperties |
| DeviceConfig:1 | urn:dslforum-org:service:DeviceConfig:1 | Reboot, X_AVM-DE_CreateUrlSID, X_AVM-DE_GetConfigFile |
| X_AVM-DE_TAM:1 | urn:dslforum-org:service:X_AVM-DE_TAM:1 | GetInfo, SetEnable, GetMessageList |
| X_AVM-DE_OnTel:1 | urn:dslforum-org:service:X_AVM-DE_OnTel:1 | GetCallList |
| X_AVM-DE_Messaging:1 | urn:dslforum-org:service:X_AVM-DE_Messaging:1 | X_AVM-DE_SendSMS |
| LANEthernetInterfaceConfig:1 | urn:dslforum-org:service:LANEthernetInterfaceConfig:1 | GetStatistics |
| WANDSLInterfaceConfig:1 | urn:dslforum-org:service:WANDSLInterfaceConfig:1 | GetInfo |
| X_AVM-DE_HostFilter:1 | urn:dslforum-org:service:X_AVM-DE_HostFilter:1 | DisallowWANAccessByIP |

### Limitations of FritzBoxShell

1. **Bash dependency**: Requires curl, wget, xmlstarlet, jq, qrencode
2. **No concurrency safety**: Parallel processing uses temp files
3. **Error handling**: Inconsistent error handling across functions
4. **Configuration**: Simple bash source, no validation
5. **Protocol fragmentation**: Mixed TR-064 and Lua/AHA interfaces
6. **No tests**: No automated testing framework
7. **Output parsing**: Fragile text parsing with awk/grep

## Architecture Design

### Project Structure

```
fritzboxctl/
├── cmd/
│   └── fritzboxctl/
│       └── main.go              # Entry point
├── internal/
│   ├── auth/
│   │   ├── digest.go          # HTTP Digest Authentication (RFC 7616)
│   │   ├── sid.go             # SID session management for AHA interface
│   │   ├── tls.go             # TLS configuration (self-signed cert handling)
│   │   ├── provider.go        # Credential provider (config/env/flags)
│   │   └── auth.go           # Main auth interface and types
│   ├── config/
│   │   ├── config.go           # Configuration management
│   │   └── config_test.go
│   ├── soap/
│   │   ├── client.go           # SOAP client implementation
│   │   ├── types.go            # SOAP types and envelopes
│   │   └── services/          # Service-specific implementations
│   │       ├── device.go       # DeviceInfo service
│   │       ├── wlan.go        # WLANConfiguration service
│   │       ├── hosts.go       # Hosts service
│   │       ├── wan.go         # WAN services
│   │       └── tam.go         # TAM service
│   ├── aha/
│   │   ├── client.go          # AHA-HTTP-Interface client
│   │   └── types.go          # AHA types
│   └── mesh/
│       └── client.go          # Mesh API client
├── pkg/
│   └── fritzbox/
│       ├── client.go           # Main FritzBox client
│       └── types.go           # Common types
├── docs/
│   └── design.md              # This file
├── go.mod
├── go.sum
├── Makefile
└── README.md
```

### Core Components

#### 1. Configuration Manager (`internal/config`)

```go
type Config struct {
    RouterURI     string        `yaml:"router_uri" env:"FRITZBOX_URI"`
    Username      string        `yaml:"username" env:"FRITZBOX_USERNAME"`
    Password      string        `yaml:"password" env:"FRITZBOX_PASSWORD"`
    RepeaterURI   string        `yaml:"repeater_uri" env:"FRITZBOX_REPEATER_URI"`
    RepeaterUser  string        `yaml:"repeater_user" env:"FRITZBOX_REPEATER_USERNAME"`
    RepeaterPass  string        `yaml:"repeater_password" env:"FRITZBOX_REPEATER_PASSWORD"`
    Timeout       time.Duration `yaml:"timeout" env:"FRITZBOX_TIMEOUT"`
    OutputFormat  string        `yaml:"output_format" env:"FRITZBOX_OUTPUT"`
}

// Config sources (in order of precedence):
// 1. Command-line flags (highest priority)
// 2. Environment variables (auto-discovered)
//    - Primary: FRITZBOX_URI, FRITZBOX_USERNAME, FRITZBOX_PASSWORD
//    - Fallback: BOXIP, BOXUSER, BOXPW (FritzBoxShell compatibility)
// 3. Config file (~/.config/fritzboxctl/config.yaml)
// 4. Defaults (lowest priority)
```

### Quick Setup Examples

```bash
# Option 1: Environment variables (auto-discovered)
export FRITZBOX_URI="192.168.178.1"
export FRITZBOX_USERNAME="admin"
export FRITZBOX_PASSWORD="secret"
fritzboxctl device info  # Just works!

# Option 2: Flags (override env vars)
fritzboxctl --router-uri 192.168.178.1 --username admin --password secret device info

# Option 3: Config file (~/.config/fritzboxctl/config.yaml)
cat > ~/.config/fritzboxctl/config.yaml <<EOF
router_uri: 192.168.178.1
username: admin
password: secret
EOF
fritzboxctl device info
```
```

#### 2. SOAP Client (`internal/soap`)

Uses `auth.DigestClient` from the auth module to handle HTTP Digest Authentication required for TR-064 SOAP requests.

```go
type Client struct {
    baseURL    string
    authClient *auth.DigestClient
    services   map[string]ServiceDescription
}

// Key methods:
func (c *Client) Call(service, action string, body interface{}) (*Response, error)
func (c *Client) Discover() error  // Parse tr64desc.xml and igddesc.xml
```

**SOAP Envelope Structure:**
```xml
<s:Envelope s:encodingStyle="http://schemas.xmlsoap.org/soap/encoding/"
            xmlns:s="http://schemas.xmlsoap.org/soap/envelope/">
    <s:Body>
        <u:{Action} xmlns:u="{ServiceType}">
            <!-- Parameters -->
        </u:{Action}>
    </s:Body>
</s:Envelope>
```

#### 3. AHA Client (`internal/aha`)

Uses `auth.SIDManager` from the auth module to obtain and manage session IDs (SID) for AHA-HTTP-Interface requests.

```go
type Client struct {
    baseURL    string
    sidManager *auth.SIDManager
    httpClient *http.Client
}

// Key methods:
func (c *Client) Login() error
func (c *Client) Logout() error
func (c *Client) Get(url string, params map[string]string) (json.RawMessage, error)
func (c *Client) Post(url string, data url.Values) (json.RawMessage, error)
```

#### 4. FritzBox Client (`pkg/fritzbox`)

High-level client that combines SOAP, AHA, and auth modules:

```go
type Client struct {
    soapClient *soap.Client
    ahaClient  *aha.Client
    config     *config.Config
}

// Device operations
func (c *Client) GetDeviceInfo() (*DeviceInfo, error)
func (c *Client) Reboot() error

// WLAN operations
func (c *Client) GetWLANStatus(band WLANBand) (*WLANStatus, error)
func (c *Client) SetWLANEnabled(band WLANBand, enabled bool) error
func (c *Client) GetWLANStatistics(band WLANBand) (*WLANStats, error)
func (c *Client) SetWLANChannel(band WLANBand, channel int) error
func (c *Client) GetWLANQRCode(band WLANBand) (string, error)

// Device operations
func (c *Client) ListDevices() ([]*Device, error)
func (c *Client) WakeOnLAN(mac string) error
func (c *Client) BlockDevice(ip string) error
func (c *Client) UnblockDevice(ip string) error

// VPN operations
func (c *Client) SetVPNEnabled(name string, enabled bool) error

// LED operations
func (c *Client) SetLEDsEnabled(enabled bool) error
func (c *Client) SetLEDBrightness(level int) error

// TAM operations
func (c *Client) GetTAMInfo(index int) (*TAMInfo, error)
func (c *Client) SetTAMEnabled(index int, enabled bool) error

// WAN operations
func (c *Client) GetWANStatus() (*WANStatus, error)
func (c *Client) ReconnectWAN() error

// Backup
func (c *Client) Backup(password string) ([]byte, error)
```

#### 5. Authentication Module (`internal/auth`)

Centralizes all authentication logic required to interact with Fritz!Box devices. Handles three auth mechanisms:
1. **HTTP Digest Authentication** for TR-064 SOAP requests (port 49000/49443)
2. **SID Session Management** for AHA-HTTP-Interface
3. **TLS Configuration** for self-signed certificates (Fritz!Box default)

##### Key Components:

- **DigestClient**: Implements HTTP Digest Access Authentication (RFC 7616). Handles challenge-response cycles, nonce tracking, and opaque values. Wraps `http.Client` with digest auth capabilities.

- **SIDManager**: Manages session IDs (SID) for AHA interface. Obtains SID via `X_AVM-DE_CreateUrlSID` SOAP action, refreshes expired SIDs, and handles logout.

- **TLSConfig**: Configures TLS settings with `InsecureSkipVerify` option for local network use with self-signed certs.

- **CredentialProvider**: Supplies credentials from multiple sources (config file > env vars > CLI flags) with proper precedence. Ensures credentials are never logged.

##### Key Types:

```go
// AuthConfig holds authentication-related configuration
type AuthConfig struct {
    InsecureSkipVerify bool          `yaml:"insecure_skip_verify" env:"FRITZBOX_INSECURE"`
    Digest             *DigestConfig `yaml:"digest"`
    SID                *SIDConfig    `yaml:"sid"`
}

// DigestClient handles HTTP Digest Authentication
type DigestClient struct {
    client   *http.Client
    username string
    password string
    nonce    string
    opaque   string
    realm    string
    qop      string
}

// SIDManager manages AHA session IDs
type SIDManager struct {
    soapClient *soap.Client
    sid        string
    expiresAt  time.Time
}

// CredentialProvider supplies credentials
type CredentialProvider struct {
    config   *config.Config
    username string
    password string
}
```

##### Key Methods:

```go
// DigestClient methods
func (d *DigestClient) Do(req *http.Request) (*http.Response, error)
func (d *DigestClient) UpdateChallenge(authHeader string) error

// SIDManager methods
func (s *SIDManager) Login() error
func (s *SIDManager) GetSID() (string, error)
func (s *SIDManager) Logout() error
func (s *SIDManager) Refresh() error

// CredentialProvider methods
func (c *CredentialProvider) Username() string
func (c *CredentialProvider) Password() string
```

##### Auth Flow:

1. **TR-064 SOAP requests**: SOAP client uses `DigestClient.Do()` which automatically handles digest auth handshake
2. **AHA-HTTP requests**: AHA client calls `SIDManager.GetSID()` to get valid SID, includes it as `sid` query parameter
3. **TLS**: All clients use `TLSConfig` that skips verification for local Fritz!Box addresses by default

---

## CLI Design (using Cobra)

### Command Structure

```bash
fritzboxctl [global flags] <command> [subcommand] [flags]
```

### Global Flags

| Flag | Short | Description | Default | Auto-Discovered Env Var |
|------|-------|-------------|---------|---------------------------|
| `--router-uri` | `-H` | Fritz!Box URI (IP or hostname) | 192.168.178.1 | `FRITZBOX_URI` |
| `--username` | `-u` | Username | - | `FRITZBOX_USERNAME` |
| `--password` | `-p` | Password | - | `FRITZBOX_PASSWORD` |
| `--timeout` | `-t` | Request timeout | 10s | `FRITZBOX_TIMEOUT` |
| `--output` | `-o` | Output format (text, json, yaml) | text | `FRITZBOX_OUTPUT` |
| `--config` | `-c` | Config file path | ~/.config/fritzboxctl/config.yaml | `FRITZBOX_CONFIG` |

**Note:** Environment variables are automatically discovered - just set them in your shell and the CLI will pick them up without additional configuration.

For backwards compatibility with FritzBoxShell:
- `FRITZBOX_URI` falls back to `BOXIP`/`FRITZBOX_IP`
- `FRITZBOX_USERNAME` falls back to `BOXUSER`/`FRITZBOX_USER`
- `FRITZBOX_PASSWORD` falls back to `BOXPW`/`FRITZBOX_PW`

### Commands

#### 1. Device Commands

```bash
# Device information
fritzboxctl device info

# Reboot
fritzboxctl device reboot [--box/--repeater]

# Backup
fritzboxctl device backup --password <pwd> [--output-file <path>]
```

#### 2. WLAN Commands

```bash
# Status
fritzboxctl wlan status [--band 2.4|5|guest|all]

# Enable/disable
fritzboxctl wlan set [--band 2.4|5|guest|all] [--on/--off]

# Statistics
fritzboxctl wlan stats [--band 2.4|5|guest]

# QR Code
fritzboxctl wlan qrcode [--band 2.4|5|guest]

# Channel
fritzboxctl wlan channel [--band 2.4|5] [--set <ch>|--random]
```

#### 3. Device Management

```bash
# List devices
fritzboxctl device list [--type 2.4|5|ETH|all] [--with-ip]

# Wake on LAN
fritzboxctl device wol <mac-address>

# Block/unblock
fritzboxctl device block <device-name-or-ip>
fritzboxctl device unblock <device-name-or-ip>

# Profiles
fritzboxctl device profiles list
fritzboxctl device profile set <device> <profile-id>
```

#### 4. Network Commands

```bash
# WAN status and reconnect
fritzboxctl network wan [status|reconnect]

# DSL status
fritzboxctl network dsl status

# LAN statistics
fritzboxctl network lan stats
fritzboxctl network lan count
```

#### 5. VPN Commands

```bash
# List VPN connections
fritzboxctl vpn list

# Enable/disable VPN
fritzboxctl vpn set --name <name> [--on/--off]
```

#### 6. TAM (Answering Machine)

```bash
# List TAMs
fritzboxctl tam list

# Enable/disable
fritzboxctl tam set --index <n> [--on/--off]

# Get messages
fritzboxctl tam messages --index <n>
```

#### 7. LED Control

```bash
# LED status
fritzboxctl led status

# Enable/disable LEDs
fritzboxctl led set [--on/--off]

# Brightness
fritzboxctl led brightness [--level 1|2|3]
```

#### 8. Misc Commands

```bash
# Event log
fritzboxctl log [read|reset]

# Key lock
fritzboxctl keylock [--on/--off]

# Signal strength
fritzboxctl signal [--strength 100|50|25|12|6]

# Call list
fritzboxctl tel calllist [--days <n>]

# Send SMS
fritzboxctl tel sms --to <number> --message <text>

# Count connections
fritzboxctl count [--type 2.4|5|ETH|all] [--with-ip]
```

## Implementation Plan

### Phase 1: Core Infrastructure

1. **Project setup**
   - Initialize Go module
   - Set up Cobra CLI framework
   - Create configuration management

2. **SOAP client**
   - Implement SOAP envelope creation
   - HTTP client with digest auth
   - Service discovery (parse tr64desc.xml)

3. **Basic commands**
   - Device info
   - WLAN status/control

### Phase 2: WLAN and Network Features

1. WLAN statistics
2. WLAN QR code generation
3. Channel management
4. WAN/DSL/LAN statistics
5. WAN reconnect

### Phase 3: Device Management

1. Device listing (mesh API)
2. Wake-on-LAN
3. Device blocking/unblocking
4. Profile management

### Phase 4: Advanced Features

1. TAM (answering machine)
2. VPN control (WireGuard, IPsec)
3. LED control
4. Key lock
5. Signal strength
6. Backup/restore

### Phase 5: Polish

1. Error handling improvements
2. Tests (unit + integration)
3. Documentation
4. CI/CD setup
5. Release automation

## Dependencies

```go
module github.com/fabio/fritzboxctl

go 1.22

require (
    github.com/spf13/cobra v1.8.0                  // CLI framework
    github.com/spf13/viper v1.18.0                 // Configuration
    github.com/abbot/go-http-auth v0.0.0-20210526062804-ef4b02c45d22 // HTTP Digest Auth
    github.com/mdp/qrterminal v3.2.0                // QR code generation
    gopkg.in/yaml.v3 v3.0.1                        // YAML config
)
```

## Output Formats

### Text (default)
```
WLAN 2.4 GHz (MyNetwork):
  Status:   Enabled
  Channel:  6
  Clients:  5
```

### JSON
```json
{
  "band": "2.4",
  "ssid": "MyNetwork",
  "enabled": true,
  "channel": 6,
  "clients": 5
}
```

### InfluxDB Line Protocol (for monitoring)
```
wlan,band=2.4,ssid=MyNetwork enabled=1,channel=6,clients=5
```

## Error Handling

1. **Connection errors**: Clear message with troubleshooting hints
2. **Auth errors**: Prompt for credentials or show config instructions
3. **Action not available**: List available actions for the service
4. **Timeout**: Retry with backoff (configurable)

## Testing Strategy

1. **Unit tests**: Mock SOAP responses
2. **Integration tests**: Test against real Fritz!Box (optional, with env var)
3. **CLI tests**: Test command parsing and output formats

## Security Considerations

1. **Password handling**: Support env vars, config file (warn if plaintext). Never log credentials.
2. **TLS**: Support self-signed certs (Fritz!Box default) with `InsecureSkipVerify` for local networks.
3. **SID handling**: Store SIDs in memory only, clear on logout or timeout. Never expose in logs.
4. **No logging**: Don't log passwords, SIDs, or sensitive data. Use redacted output for debug.
5. **Digest auth**: Implement proper RFC 7616 compliance, avoid credential exposure in challenges.

## Future Enhancements

1. **Daemon mode**: Run as daemon for continuous monitoring
2. **Prometheus exporter**: Export metrics in Prometheus format
3. **REST API**: Expose functionality via HTTP API
4. **Web UI**: Simple web interface
5. **MQTT integration**: Publish device status to MQTT broker
6. **Home Assistant integration**: Custom component
