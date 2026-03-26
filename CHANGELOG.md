# v0.2.0 (2026-03-24)

## Features

### Query & Output
- **`--jq` global filter** — Apply jq expressions to filter JSON/YAML output from any command (powered by gojq). Single-key `{"result": [...]}` responses are auto-unwrapped. String results output as plain text (no JSON quotes); other types output as compact JSON. Usage: `incloud device list --jq '.[].name'`
- **Smart output format** — Auto-detect TTY to choose output format: table by default in terminal, JSON when piped or redirected. `--output` flag overrides. Time-series commands also default to table output

### Device Management
- **Config schema commands** — New `device config schema list/get/overview/validate` subcommands. Query by device ID (`--device`) or product model + firmware version (`--product` + `--version`). `list` supports `--name` regex filter; `validate` accepts payload via `--payload` or `--file` for JSON Schema pre-validation; auto-suggests available firmware versions when schema is not found
- **Enhanced device creation** — SN pre-validation (auto-checks serial number validity and required verification fields before creation), conditional MAC/IMEI prompts (interactive in TTY, error with flag hints in pipe mode), rich error messages (duplicate name, duplicate SN, MAC/IMEI mismatch all have clear diagnostics and fix suggestions)
- **Synchronous packet capture** — `device exec capture` changed from async to sync mode, blocking until completion. `--download <file>` auto-downloads pcap file on completion, Ctrl+C cancels device-side task, failed downloads auto-clean incomplete files
- **Interactive speed test** — `device exec speedtest` redesigned with interactive guidance: auto-fetches available interface list for selection, retrieves matching speed test nodes per interface, streams progress (TTY real-time refresh, non-TTY final result only). New `speedtest-config` subcommand to view available options
- **Syslog live upload** — `device log syslog --fetch` triggers device to upload current log buffer (waits up to 40s). Without `--fetch`, queries existing platform logs (`--after`/`--before` required); with `--fetch`, time range is optional, defaults to today

### Alerts
- **Rule type parameters** — `alert rule create/update` supports type-specific parameters (offline retention duration, CPU threshold, signal strength limit, etc.). `--type` accepts plain type name, comma-separated parameters, or JSON format, and can be repeated for multiple types
- **Type discovery command** — New `alert rule types` lists all 26 alert types with parameter descriptions, `alert rule types <type>` shows single type details

### Development & Debugging
- **Debug mode** — Enable with `--debug` flag or `INCLOUD_DEBUG=1` env var. Outputs HTTP request/response headers and status codes, request body (truncated at 4KB), response timing, token refresh events, config context source. Authorization is always redacted to `****`
- **User impersonation (Sudo)** — Super admins can impersonate any user. Hidden `--sudo <user>` flag or `INCLOUD_SUDO=<user>` env var. Only available to super admins; non-admin calls are silently ignored by the backend. Sudo header is only injected for same-origin requests to prevent credential leakage

## Improvements
- Localized timestamps: ISO 8601 times in table output are auto-converted to local time
- Renamed flags to improve AI discoverability (`--to` → `--target`, `--out` → `--output-file`), old names preserved as hidden aliases
- Required parameters now use `MarkFlagRequired` instead of manual validation, help auto-annotates required flags
- Connector deletion: batch name lookup for confirmation, new typed HTTPError

## Bug Fixes
- Fix `device perf` missing disk and microSD formatters, causing related metrics to not display
- Fix `overview` average/max offline duration showing raw seconds instead of human-readable time
- Fix `device config history list` returning oversized mergedConfig field
- Fix device parsing using wrong `partNumber` field (changed to `product`)
- Fix UTF-8 character truncation in config schema queries
- Fix streaming commands incorrectly warning when `--output` is not explicitly set
- Fix `alert rule --type` help text and examples referencing undiscoverable type names
- Fix Authorization header not being stripped on cross-origin redirects, preventing credential leakage

---

# v0.1.0

First release of InCloud CLI, providing full command-line management for the InHand Cloud platform.

## Features

### Core
- OAuth browser login (PKCE) with automatic token refresh and login status view
- Multi-format output: JSON, YAML, table with built-in human-friendly formatting
- Generic `api` command with query parameters, request body, and custom headers
- Config context management (use/list/set/delete/current)
- Real-time SSE streaming for ping and traceroute diagnostics

### Device Management
- Full device lifecycle: list, get, create, update, delete
- Device groups, configurations, shadows, connected clients
- Signal, interfaces, online/offline events, syslog viewing
- Location management, traffic statistics, performance monitoring, online status
- Antenna info, uplink, remote execution
- Batch import (CSV/XLSX) and export (CSV)
- Assign, unassign, transfer devices

### Firmware
- Firmware listing and details
- Upgrade tasks: create, list, cancel, execution details, retry

### Alerts
- Alert list, details, acknowledge, acknowledge statistics
- Alert rule CRUD, alert export

### Overview Dashboard
- Dashboard, device, alert, traffic, offline summaries
- Top devices and top alert type analysis

### Network & Connectivity
- SD-WAN module: networks, devices, tunnels, connections
- OOBM out-of-band management commands
- Connector usage: statistics, trends, top-K

### Organization
- Organization, user, role management
- Operation audit log queries

### Products
- Product CRUD: list, get, create, update, delete

## Improvements
- Human-friendly formatting for bytes, bitrates, percentages, latency, jitter, duration
- Table exclude-column mode (! prefix), dot-path parsing for nested fields
- TTY table pagination summary header, auto-flattening of nested objects
- Interactive confirmation prompts using charmbracelet/huh

## Bug Fixes
- Fix default output format when `--output` is not specified
- Fix ANSI style nesting in pagination header causing color corruption
- Fix exported file permissions (0o600)
- Fix fallback prompt when browser fails to close after OAuth login
- Fix device import polling and validation status handling
