# Unified Query Params Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Extract common query param handling (page/limit/sort/fields/expand) into a shared `cmdutil.NewQuery()` factory function, eliminating ~200 lines of duplicated boilerplate across ~45 commands.

**Architecture:** Create a new `internal/cmdutil` package with a `NewQuery(cmd)` function that reads well-known flags from the cobra command and builds `url.Values`. Each command still defines its own flags (with command-specific help text and defaults), but the query param construction is centralized. The function uses `cmd.Flags().Lookup()` to detect which flags exist and whether they were changed.

**Tech Stack:** Go, cobra, net/url

---

## Design

### NewQuery function

```go
// internal/cmdutil/query.go
package cmdutil

func NewQuery(cmd *cobra.Command) url.Values {
    q := url.Values{}

    // page: CLI 1-based → API 0-based
    if f := cmd.Flags().Lookup("page"); f != nil {
        page, _ := strconv.Atoi(f.Value.String())
        q.Set("page", strconv.Itoa(page-1))
    }

    // limit
    if f := cmd.Flags().Lookup("limit"); f != nil {
        q.Set("limit", f.Value.String())
    }

    // sort
    if f := cmd.Flags().Lookup("sort"); f != nil && f.Changed {
        q.Set("sort", f.Value.String())
    }

    // fields: StringSlice → comma-joined
    if f := cmd.Flags().Lookup("fields"); f != nil && f.Changed {
        q.Set("fields", strings.ReplaceAll(f.Value.String(), " ", ""))
    }

    // expand: StringSlice → comma-joined; also set if has non-empty default
    if f := cmd.Flags().Lookup("expand"); f != nil {
        val := strings.ReplaceAll(f.Value.String(), " ", "")
        if val != "" && val != "[]" {
            q.Set("expand", val)
        }
    }

    return q
}
```

Key design decisions:
- Uses `Lookup()` to detect flag existence — works for any command, no interface needed
- Uses `f.Changed` for optional flags (sort, fields) — only set when user explicitly provides
- For expand: checks non-empty value (not `Changed`) — this way hardcoded defaults like `--expand "creator,jobProcessDetails"` are included even when user doesn't explicitly pass the flag
- StringSlice `.Value.String()` returns `[a,b,c]` format — need to strip brackets. Actually cobra's `StringSlice.String()` returns the joined form. Need to verify.

### Flag type standardization

All `expand` flags must be `StringSliceVar` (not `StringVar`). Commands currently using `StringVar` for expand:
- `user/get.go` — default "roles"
- `user/list.go` — no default
- `user/me.go` — no default
- `org/get.go` — no default
- `org/list.go` — no default
- `sdwan/network_get.go` — default "tunnels"
- `connector/network_get.go` — no default

### Hardcoded expand → flag with default

- `firmware/job_list.go`: `q.Set("expand", "creator,jobProcessDetails")` → add `--expand` flag with default `[]string{"creator", "jobProcessDetails"}`
- `auth/status.go`: internal API call, NOT a user-facing list command — leave as-is

### Bug fix

- `device/asset_list.go`: `q.Set("size", ...)` → `q.Set("limit", ...)` (matches frontend behavior)

---

## Tasks

### Task 1: Create cmdutil package with NewQuery

**Files:**
- Create: `internal/cmdutil/query.go`
- Create: `internal/cmdutil/query_test.go`

**Step 1: Write test for NewQuery**

Test all 5 flags: page (1→0 conversion), limit, sort, fields (StringSlice), expand (StringSlice). Test that missing flags are skipped, unchanged optional flags are skipped, and `fields=*` passes through (cleanValues handles it).

**Step 2: Run test to verify it fails**

Run: `go test ./internal/cmdutil/ -v`

**Step 3: Implement NewQuery**

Key consideration: cobra `StringSliceVar` `.Value.String()` returns `[a,b,c]` with brackets. Use `cmd.Flags().GetStringSlice()` instead of raw `.Value.String()` for slice types, then `strings.Join()`.

```go
func NewQuery(cmd *cobra.Command) url.Values {
    q := url.Values{}
    setPage(cmd, q)
    setLimit(cmd, q)
    setSort(cmd, q)
    setFields(cmd, q)
    setExpand(cmd, q)
    return q
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./internal/cmdutil/ -v`

**Step 5: Commit**

```
feat: add cmdutil.NewQuery for unified query param handling
```

---

### Task 2: Standardize expand flags to StringSlice

**Files to modify** (change `StringVar` → `StringSliceVar` for expand):
- `internal/cmd/user/get.go` — default `"roles"` → `[]string{"roles"}`
- `internal/cmd/user/list.go` — `StringVar` → `StringSliceVar`
- `internal/cmd/user/me.go` — `StringVar` → `StringSliceVar`
- `internal/cmd/org/get.go` — `StringVar` → `StringSliceVar`
- `internal/cmd/org/list.go` — `StringVar` → `StringSliceVar`
- `internal/cmd/sdwan/network_get.go` — default `"tunnels"` → `[]string{"tunnels"}`
- `internal/cmd/connector/network_get.go` — `StringVar` → `StringSliceVar`

Also convert hardcoded expand to flag with default:
- `internal/cmd/firmware/job_list.go` — add `--expand` flag with default `[]string{"creator", "jobProcessDetails"}`

**Steps:**

1. For each file: change `opts.Expand` type from `string` to `[]string`, update flag registration, update `q.Set("expand", ...)` to use `strings.Join()`
2. For `firmware/job_list.go`: add Expand field to opts, add flag with default, remove hardcoded `q.Set`
3. Run: `go build ./cmd/incloud` — compiler catches any missed changes
4. Run: `go test ./...`
5. Commit: `refactor: standardize all expand flags to StringSlice`

---

### Task 3: Fix asset_list.go limit param name

**Files:**
- Modify: `internal/cmd/device/asset_list.go`

**Steps:**

1. Change `q.Set("size", strconv.Itoa(opts.Limit))` → `q.Set("limit", strconv.Itoa(opts.Limit))`
2. Run: `go build ./cmd/incloud`
3. Commit: `fix: use correct param name "limit" in device asset list`

---

### Task 4: Migrate all commands to use NewQuery

**Strategy:** Replace the common query param boilerplate in each command's RunE with `cmdutil.NewQuery(cmd)`. The compiler won't help here (old code still compiles), so use grep to find all instances.

**Pattern to replace:**

```go
// Before:
q := url.Values{}
q.Set("page", strconv.Itoa(opts.Page-1))
q.Set("limit", strconv.Itoa(opts.Limit))
if opts.Sort != "" {
    q.Set("sort", opts.Sort)
}
// ... fields/expand handling ...

// After:
q := cmdutil.NewQuery(cmd)
```

**Commands to migrate** (grouped by package for efficient editing):

- `internal/cmd/device/`: list, asset_list, client_list, client_online_events, config_history_list, datausage_list, group_list, log_mqtt, online, signal_list, uplink_perf
- `internal/cmd/alert/`: list, rule_list
- `internal/cmd/firmware/`: list, status, job_list, job_executions
- `internal/cmd/user/`: list, me, get, identity_list
- `internal/cmd/org/`: list, get
- `internal/cmd/connector/`: account_list, device_list, device_list_all, endpoint_list, network_list, network_get
- `internal/cmd/sdwan/`: candidates, devices, network_list, network_tunnels, network_connections, network_get
- `internal/cmd/tunnel/`: tunnel_logs
- `internal/cmd/overview/`: devices, offline, traffic, trend
- `internal/cmd/webhook/`: list
- `internal/cmd/role/`: list
- `internal/cmd/product/`: list, get
- `internal/cmd/feedback/`: list
- `internal/cmd/activity/`: list
- `internal/cmd/oobm/`: oobm_list, oobm_logs, oobm_serial_list

**Steps:**

1. Use a subagent per package group (device, alert+firmware, user+org, connector+sdwan, rest)
2. Each migration: replace boilerplate with `cmdutil.NewQuery(cmd)`, add import, remove now-unused `strconv`/`strings` imports if applicable
3. Special cases:
   - `alert/list.go`: use NewQuery, then override page/limit when `opts.Count`
   - `overview/*`: some don't have standard flags — check each
   - Commands with `fields` default logic: keep the default fields assignment for API filtering, just use NewQuery for page/limit/sort/expand
4. Run: `go build ./cmd/incloud` after each package group
5. Run: `go test ./...` after all migrations
6. Commit: `refactor: migrate commands to use cmdutil.NewQuery`

---

### Task 5: Clean up unused code

After migration, some commands may have unused imports (`strconv`, `strings`, `net/url`) or unused variables.

**Steps:**

1. Run: `goimports -w internal/cmd/`
2. Run: `go vet ./...`
3. Run: `go build ./cmd/incloud && go test ./...`
4. Commit: `chore: clean up unused imports after NewQuery migration`

---

### Task 6: End-to-end verification

**Steps:**

1. `make build`
2. Test device list scenarios:
   - `./bin/incloud device list --limit 2` — default table
   - `./bin/incloud device list --limit 1 -f '*'` — all fields
   - `./bin/incloud device list --limit 1 --expand org` — expand
   - `./bin/incloud device list --limit 1 --expand org -f '*'` — expand + all fields
3. Test other commands:
   - `./bin/incloud user list --limit 2 --expand roles` — expand as StringSlice
   - `./bin/incloud firmware job list --limit 2` — default expand preserved
   - `./bin/incloud alert list --count` — count mode
   - `./bin/incloud device asset list --limit 2` — fixed limit param
4. Confirm all pass

---

## Notes

- `auth/status.go` hardcoded expand is an internal API call, not a user-facing list — leave as-is
- `device/client.go` `fetchClientSeries` has its own query building pattern — leave as-is (not a standard list command)
- `overview/` commands have some unique patterns (e.g., `traffic.go` builds query differently) — evaluate each individually during Task 4
- The `fields` default logic (`if len(fields) == 0 && output == "table" { fields = defaultFields }`) stays in each command — NewQuery only reads the raw flag value
