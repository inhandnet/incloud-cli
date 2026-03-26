# InCloud CLI

Command-line tool for [InCloud Manager](https://www.inhandnetworks.com/products/incloud-manager.html) IoT device management platform. Supports authentication, multi-environment context switching, API calls, and multiple output formats.

## Installation

### Install via Claude Code

Paste the following into [Claude Code](https://claude.ai/code):

```
Read https://raw.githubusercontent.com/inhandnet/incloud-cli/main/INSTALL.md and follow the instructions to install incloud CLI.
```

### Download Binary

Download pre-built binaries from the [Releases](https://github.com/inhandnet/incloud-cli/releases) page.

### Build from Source

```bash
# Requires Go 1.25+
make build    # Output to bin/incloud
make install  # Install to $GOPATH/bin
```

> On macOS, `CGO_ENABLED=0` is required (already set in Makefile) to avoid dyld LC_UUID errors.

## Quick Start

### 1. Configure Context

```bash
incloud config set-context dev --host nezha.inhand.dev --org myorg
incloud config use-context dev
```

### 2. Login

```bash
incloud auth login              # Global region (default)
incloud auth login --host cn    # China region
```

Login uses OAuth 2.0 Authorization Code + PKCE flow and automatically opens a browser for authorization.

The CLI reuses the platform frontend's SPA OAuth client, automatically fetching `client_id` and `client_secret` from the platform API at login time, sending credentials via `client_secret_post` (consistent with frontend behavior). You can also specify `--client-id` manually.

### 3. Verify

```bash
incloud auth status
incloud api /api/v1/users/me
```

## Command Reference

### Authentication

```bash
incloud auth login                    # Browser-based OAuth login
incloud auth status                   # View current auth status
incloud auth logout                   # Log out
```

### Context Management

```bash
incloud config set-context <name> --host <url> --org <org>
incloud config use-context <name>
incloud config current-context
incloud config list-contexts
incloud config delete-context <name>
```

### Device Management

```bash
incloud device list                                  # List devices
incloud device get <id>                              # Get device details
incloud device create --name test --sn SN123         # Create device
incloud device update <id> --name new-name           # Update device
incloud device delete <id>                           # Delete device
incloud device group list                            # List device groups
incloud device signal <id>                           # View signal quality
incloud device perf <id>                             # View performance metrics
incloud device exec ping <id> --target 8.8.8.8       # Remote ping
incloud device exec traceroute <id> --target 8.8.8.8 # Remote traceroute
incloud device exec capture <id> --download out.pcap # Packet capture with download
incloud device exec speedtest <id>                   # Interactive speed test
incloud device log syslog <id> --fetch               # Fetch live syslog
incloud device config schema list --device <id>      # List config schemas
incloud device config schema validate --device <id> --file config.json  # Validate config
```

### Alerts

```bash
incloud alert list                                   # List alerts
incloud alert get <id>                               # Get alert details
incloud alert ack <id>                               # Acknowledge alert
incloud alert rule list                              # List alert rules
incloud alert rule create --name test --type disconnected,retention=600
incloud alert rule types                             # List all alert types
```

### Firmware

```bash
incloud firmware list                                # List firmware
incloud firmware get <id>                            # Get firmware details
incloud firmware upgrade create --firmware <id> --device <id>  # Create upgrade task
incloud firmware upgrade list                        # List upgrade tasks
```

### Overview Dashboard

```bash
incloud overview dashboard                           # Dashboard summary
incloud overview device                              # Device statistics
incloud overview alert                               # Alert statistics
incloud overview traffic                             # Traffic summary
```

### Network & Connectivity

```bash
incloud sdwan network list                           # SD-WAN networks
incloud oobm list                                    # OOBM sessions
incloud connector list                               # List connectors
```

### Organization

```bash
incloud org get                                      # View organization info
incloud user list                                    # List users
incloud role list                                    # List roles
incloud activity list                                # Audit log
```

### Generic API Call

```bash
incloud api /api/v1/users/me                            # GET request
incloud api /api/v1/devices -q page=0 -q limit=10       # With query params
incloud api /api/v1/devices -X POST -f name=test         # POST with body fields
echo '{}' | incloud api /api/v1/devices -X POST --input - # Read JSON body from stdin
incloud api /api/v1/users/me -H "Sudo: user@example.com" # Custom header
```

### Self-Update

```bash
incloud update                                       # Update to latest version
incloud update --version v0.2.0                      # Update to specific version
```

### Debugging

```bash
incloud device list --debug                              # Output debug info to stderr
INCLOUD_DEBUG=1 incloud device list                      # Enable via env var
incloud device list --debug -o json 2>/tmp/debug.log     # Debug to file, keep stdout clean
```

### Global Flags

```bash
incloud --context prod api /api/v1/users/me              # Temporarily switch context
incloud --debug device list                              # Enable debug output
incloud version                                          # Show version
```

## Output Formats

Specify output format with `-o`:

| Format | TTY Behavior | Pipe Behavior |
|--------|-------------|---------------|
| `table` (default in TTY) | Aligned table + pagination summary | TSV |
| `json` (default in pipe) | Colorized pretty JSON | Compact JSON |
| `yaml` | YAML | YAML |

```bash
incloud device list -o table                         # Table output
incloud device list -o json                          # JSON output
incloud device list -o yaml                          # YAML output
```

### JQ Filter

Apply jq expressions to filter output from any command:

```bash
incloud device list --jq '.[].name'                  # Extract device names
incloud device list --jq '.[] | select(.online)'     # Filter online devices
```

### Field Selection (`--fields`)

Domain commands (`device list`, `alert list`, etc.) support `--fields`/`-f` to control returned and displayed fields:

```bash
incloud device list -o table -f name -f serialNumber -f online   # Show specific fields
incloud device list -o json -f name -f status                    # JSON with specific fields
```

`--fields` is passed to the API's `fields` parameter, reducing data transfer.

### Column Selection (`--column`)

The generic `api` command uses `--column`/`-c` for client-side column filtering (not sent to API):

```bash
incloud api /api/v1/devices -o table -c name -c status
```

### Pagination

```bash
incloud device list --page 1 --limit 20              # First page (default), 20 per page
incloud device list --page 2 --limit 50              # Second page, 50 per page
```

`--page` starts from 1. Table mode shows pagination summary: `Showing 20 of 96 results (Page 1 of 5)`.

## Environment Variables

| Variable | Description |
|----------|-------------|
| `INCLOUD_CONTEXT` | Override current context |
| `INCLOUD_HOST` | Override host in context |
| `INCLOUD_TOKEN` | Override token in context |
| `INCLOUD_DEBUG` | Set to any non-empty value to enable debug output |
| `INCLOUD_SUDO` | Impersonate a user (super admin only) |

## Configuration

Path: `~/.config/incloud/config.yaml` (permissions `0600`)

The config file stores all context information (host, org, token, etc.), managed via `incloud config` subcommands.

## Development

### Prerequisites

- Go 1.25+
- [golangci-lint](https://golangci-lint.run/)
- [goimports](https://pkg.go.dev/golang.org/x/tools/cmd/goimports)

### Build & Test

```bash
make build    # Build to bin/incloud
make test     # Run tests
make lint     # Run golangci-lint
make clean    # Clean build artifacts
```

### Project Structure

```
cmd/incloud/        # CLI entrypoint
internal/
  api/              # OAuth authentication, token transport
  build/            # Build-time version info
  cmd/              # Subcommand implementations
    device/         #   Device management
    alert/          #   Alert management
    firmware/       #   Firmware management
    config/         #   Config context management
    auth/           #   Authentication
    ...             #   (and more)
  config/           # Config file I/O, context model
  debug/            # Debug output (--debug / INCLOUD_DEBUG)
  factory/          # Dependency injection factory
  iostreams/        # Terminal output, formatting (JSON/Table/YAML)
  ui/               # Interactive UI components
```

## License

Proprietary — InHand Networks
