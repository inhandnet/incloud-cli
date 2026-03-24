# Debug Mode Design

## Overview

Add `--debug` mode to incloud-cli for developer troubleshooting. Outputs diagnostic
information to stderr covering config resolution, auth events, and HTTP request/response details.

## Trigger

- `--debug` persistent flag (on root command)
- `INCLOUD_DEBUG=1` environment variable
- Flag takes precedence over env var; either one enables debug output

## Output Target

All debug output goes to **stderr**, preserving stdout for data (tables, JSON, etc.).
This ensures `incloud device list --debug -o json | jq ...` works correctly.

## Format

All lines prefixed with `[debug]` for easy filtering (`2>&1 | grep '\[debug\]'`).

## Coverage

### 1. Config Resolution

Printed once at command startup, shows which context is active and where values come from.

```
[debug] context: dev (from: config)
[debug] api:  https://star.nezha.inhand.dev
[debug] auth: https://portal.nezha.inhand.dev
[debug] org: myorg
```

Source attribution (`from: config`, `from: env INCLOUD_HOST`, `from: flag --context`)
helps identify when env vars or flags override config file values.

### 2. Auth Events

Printed during token transport (RoundTripper), shows token state and refresh activity.

```
[debug] token expires at 2026-03-20T18:45:00Z (valid)
[debug] sudo: admin@example.com
```

On token refresh:
```
[debug] token expired, refreshing...
[debug] token refreshed, new expiry: 2026-03-20T19:45:00Z
```

On refresh failure:
```
[debug] token refresh failed: 401 Unauthorized
```

### 3. HTTP Request/Response

Printed for every HTTP call. Request body included, response body NOT included
(use `-o json` to inspect response data).

```
[debug] > POST https://star.nezha.inhand.dev/api/v1/devices?page=0&limit=20
[debug] > Content-Type: application/json
[debug] > Authorization: Bearer ****
[debug] > Body: {"name":"test-device","model":"IR915"}
[debug] < 201 Created (235ms)
[debug] < X-Request-Id: abc-123
```

Rules:
- `Authorization` header value redacted to `****`
- Request timing measured from send to first response byte
- Response: status line + selected headers (X-Request-Id, Content-Type, X-Total-Count)
- Request body printed as-is (POST/PUT/PATCH only); no body = no Body line
- Response body NOT printed (available via `-o json`)

## Implementation Notes

### Debug Logger

Create a simple `debugger` (or similar) in `internal/debug/` package:

```go
package debug

var Enabled bool

func Log(format string, args ...any) {
    if !Enabled {
        fmt.Fprintf(os.Stderr, "[debug] "+format+"\n", args...)
    }
}
```

### Integration Points

1. **Root command (`PersistentPreRunE`)** — read `--debug` flag and `INCLOUD_DEBUG` env,
   set `debug.Enabled`
2. **Factory / Config resolution** — after resolving context, host, org, call `debug.Log`
   with source info
3. **TokenTransport (`RoundTrip`)** — log token expiry check, refresh events, sudo header
4. **APIClient (`execute`)** — log request method/URL/headers/body before send,
   log response status/timing/headers after receive

### What's NOT Included

- **Output pipeline** (envelope unwrap, field filtering, timestamp formatting) — too noisy,
  low diagnostic value. Use `-o json` to see raw API response instead.
- **Verbose/debug level split** — single level for now; revisit if user-facing debug is needed later.

## Examples

```bash
# Debug a single command
incloud device list --debug

# Debug multiple commands via env var
export INCLOUD_DEBUG=1
incloud device list
incloud alert list
unset INCLOUD_DEBUG

# Debug without polluting stdout pipe
incloud device list --debug -o json 2>/tmp/debug.log | jq '.[] | .name'
```
