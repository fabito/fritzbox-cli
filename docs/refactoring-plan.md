# Refactoring Plan: Separate Concerns (cmd/ vs internal/)

## Goal
Make `cmd/fritzboxctl/` a **thin CLI layer** that only handles command-line parsing and delegates to `internal/` packages for business logic.

## Current Problems
1. `cmd/fritzboxctl/device.go` contains XML parsing structs (`Envelope`, `GetInfoResponse`, `Body`)
2. `cmd/fritzboxctl/wlan.go` contains XML parsing structs (`WLANStatsEnvelope`)
3. `cleanSoapResponse()` function is **duplicated** in both `device.go` and `wlan.go`
4. `displayDeviceInfo()` and `displayWLANStats()` contain business logic + display formatting
5. Unit tests for device info are in `cmd/fritzboxctl/device_test.go` instead of proper locations

## Target Architecture (from design.md)
```
cmd/fritzboxctl/           # THIN LAYER - CLI only
    main.go                # Entry point, cobra setup
    device.go              # Commands only, calls internal services
    wlan.go                # Commands only, calls internal services
    network.go             # Commands only, calls internal services

internal/soap/
    client.go              # SOAP client
    types.go               # Common SOAP types
    services/              # Service-specific implementations
        device.go         # DeviceInfo service (GetInfo, Reboot, etc.) + XML types
        wlan.go           # WLANConfiguration service + XML types
        hosts.go           # Hosts service + XML types
        wan.go             # WAN services (to be created)
```

## TDD Refactoring Steps

### Step 1: RED - Write failing tests for internal/services
- Create `internal/soap/services/device_test.go` with tests that will fail (no implementation yet)
- Tests should cover: `GetDeviceInfo()`, XML parsing of `GetInfoResponse`

### Step 2: GREEN - Implement internal/services/device.go
- Move `Envelope`, `GetInfoResponse`, `Body` structs to `internal/soap/services/device.go`
- Move `cleanSoapResponse()` to `internal/soap/client.go` (or a new `internal/soap/utils.go`)
- Create `GetDeviceInfo(soapClient) (*DeviceInfo, error)` function
- Run tests to verify they pass

### Step 3: REFACTOR - Clean up cmd/device.go
- Remove XML structs from `cmd/fritzboxctl/device.go`
- Make `cmd/fritzboxctl/device.go` call `internal/soap/services.GetDeviceInfo()`
- Keep only CLI formatting in `cmd/` (or move formatting to internal too)

### Step 4: Repeat for WLAN
- RED: Write failing tests for WLAN statistics in `internal/soap/services/wlan_test.go`
- GREEN: Implement `internal/soap/services/wlan.go` with `GetWLANStats()`
- REFACTOR: Clean up `cmd/fritzboxctl/wlan.go`

### Step 5: Review
- Use a subagent to review the refactoring
- Ensure `cmd/` is a thin layer
- Ensure no XML parsing logic remains in `cmd/`
- Ensure no code duplication

## Definition of Done
1. `cmd/fritzboxctl/*.go` files contain ONLY:
   - Cobra command definitions
   - CLI flag handling
   - Output formatting (printing to user)
   - NO XML parsing structs
   - NO business logic

2. `internal/soap/services/*.go` files contain:
   - Service-specific SOAP calls
   - XML request/response structs with proper tags
   - Response parsing and error handling
   - Unit tests

3. No duplicated code (`cleanSoapResponse` should exist in ONE place)

4. All tests pass: `go test ./...`

5. Works against real router: `./fritzboxctl device info` and `./fritzboxctl wlan stats`
