# v0.6.0 (2026-04-16)

## New Features

### Knowledge Base
- **`knowledge` command group** — Search device documentation and get AI-generated answers
  - `knowledge search` — Search the knowledge base with optional model filter and query rewriting
  - `knowledge ask` — Get AI-generated answers from device documentation

### Device Logs
- **`device log local`** — Retrieve logs stored locally on the device (complements existing `device log remote`)

## Improvements

- **Time flag flexibility** — `--after` / `--before` flags now accept local time (`2025-01-01T08:00:00`) and date-only (`2025-01-01`) formats in addition to UTC RFC 3339; values are automatically converted to UTC before being sent to the API
- **AI agent compatibility** — `--limit` flag now accepts `--size` as a hidden alias for better compatibility with AI agents that use the API parameter name directly

---

# v0.5.0 (2026-04-01)

## New Features

### License Management
- **`license` command group** — Full license lifecycle management
  - `license list` / `license get` — View license list and details
  - `license attach` / `license detach` — Bind / unbind device licenses
  - `license transfer` — Transfer licenses between devices
  - `license upgrade` — Upgrade license type
  - `license align-expiry` — Align expiry dates across multiple licenses
  - `license history` — View license operation history
- **`license order` subcommands** — Order queries (list / get)
- **`license type` subcommands** — License type queries (list / get)

## Bug Fixes
- Fix `device config schema validate` failing on PCRE syntax (e.g. `(?!)`) by switching from Go's standard regexp engine to a PCRE-compatible engine
- Fix `product create` missing `firmwareVersionRule` field, preventing firmware version rules from being set during product creation
- Fix `license list` overlay operation executing unexpectedly in non-TTY environments without `--yes`
- Fix incorrect `orderId` query parameter name in `license order list`

---

# v0.4.1 (2026-03-30)

## New Features

### Feedback
- **`feedback resolve`** — Update feedback ticket resolution status

### Device Data
- **`device datausage --interval`** — Device traffic statistics now support `--interval` to specify aggregation interval
- **`device list` filter enhancements** — Add missing filter parameters to device list, consistent with other list commands

### Query Enhancements
- **`--expand` flag** — Multiple commands (device list, device group list, etc.) now support `--expand` to include related fields
- **`--sort` flag** — Multiple list commands now support `--sort` for ordering results
- **`device group list --org`** — Device group list supports filtering by organization
- **Offline statistics filtering** — `overview offline` command supports conditional filtering

### Super Admin
- **`--sudo` flag visibility** — Super admins can now see the `--sudo` flag in all subcommand help text

## Bug Fixes
- Fix `feedback list` not displaying reply content in default output
- Fix traffic statistics table output fields inconsistent with JSON output
- Fix `--expand` parameter values not matching backend API

## Changes
- Remove `--count` flag from `activity`, `alert`, and `feedback` list commands

---

# v0.4.0 (2026-03-27)

## New Features

### Remote Tunnels
- **`tunnel` command group** — Manage device remote access tunnels (list/get/create/delete), view tunnel connection status and details
- **`tunnel forward`** — Forward device ports to localhost via tunnel, using smux multiplexing for efficient TCP forwarding
- **`tunnel exec`** — Execute CLI commands on devices via tunnel, without direct SSH access

### Webhook Management
- **`webhook` command group** — Full webhook lifecycle management (list/get/create/update/delete)
- **`webhook test`** — Test webhook connectivity, supports testing by ID or generic provider test

### Device Assets
- **`device asset` command group** — Device asset CRUD management (list/get/create/update/delete)
- **`client mark-asset`** — Mark a client as an asset

### Client Management
- **`client set-pos-ready`** — Set client POS ready status

### Overview
- **`overview trend`** — Query daily device online count / total count trends

### Query Enhancements
- **`--order` flag** — `device log mqtt` and `signal list` support `--order` to specify sort direction
- **`--timeout` flag** — `device log syslog --fetch` supports custom timeout duration

### Other
- **User-Agent header** — All CLI requests now include a User-Agent header for server-side traffic identification
- **Windows installation docs** — INSTALL.md now includes Windows installation steps

## Bug Fixes
- Fix `api` command printing error messages twice
- Fix 0-based page numbers not converted to 1-based in JSON/YAML/JQ output modes
- Fix pagination query parameter using `size` instead of `limit`
- Fix syslog timestamp parameter appending duplicate `Z` suffix
- Fix timestamp parameters not properly normalized

## Changes
- Remove `--forward` flag from `open-web`, port forwarding consolidated into `tunnel forward`

---

# v0.3.0 (2026-03-26)

## New Features

### Self-Update
- **`incloud update` command** — Check and install new versions from GitHub Releases. Supports `--check` (check only), `--version` (specific version), `--yes` (skip confirmation). Auto-fallback to S3 China mirror when GitHub is unreachable

### Feedback
- **`feedback create`** — Submit feedback tickets, supports `--file` for attachments
- **`feedback list`** — View feedback list with attachment info
- **`feedback download`** — Download feedback attachments

### Authentication
- **Zero-config login** — `incloud login` works with no arguments (defaults to global region, default context)
- **Region shorthand** — `--host` accepts `global`, `cn`, `dev`, `beta` etc. instead of full domain names
- **Top-level login alias** — `incloud login` as shortcut for `incloud auth login`
- **401 auto-prompt** — Prompts re-login on 401 errors
- **Remove stored credentials** — OAuth client credentials no longer saved to config file, fetched dynamically to reduce sensitive data on disk

### Device Management
- **Batch import enhancements** — `device import` adds `--group` and `--org` flags to specify group and sub-organization during import
- **`device import-status` command** — Query import task status, displays per-line error details on failure (serial number, failure reason)
- **Diagnostic log auto-decryption** — `device log diagnostic` auto-detects AES encryption on download and decrypts, outputting .tar.gz directly

### Users & Organizations
- **`user identity list`** — View current user's roles across all accessible organizations, supports filtering by org name
- **`--tenant` global flag** — Switch organization context per-request, multi-org users can operate on external organizations without admin privileges

### Architecture
- **API/Auth URL separation** — Host config split into API address (star.*) and auth address (portal.*), supports direct IP connections

## Bug Fixes
- Add `--yes` confirmation prompt to `alert rule delete` and `user unlock` to prevent accidental operations
- `device log syslog` always outputs plain text lines, ignoring `-o json` to stay grep-friendly

---

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
