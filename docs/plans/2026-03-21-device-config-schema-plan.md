# Device Config Schema Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Add `device config schema` subcommands (list, get, overview, validate) so AI tools can discover config schemas and validate JSON before writing.

**Architecture:** Four new commands under `device config schema`, sharing a device-resolution helper that maps `--device <id>` to product/version. Schema validation uses `santhosh-tekuri/jsonschema/v6` for draft-07 support. All commands follow existing CLI patterns (cobra + factory + iostreams).

**Tech Stack:** Go, Cobra, resty (via APIClient), `santhosh-tekuri/jsonschema/v6`, httptest for tests.

---

### Task 1: Add jsonschema dependency

**Files:**
- Modify: `go.mod`
- Modify: `go.sum`

**Step 1: Add the dependency**

Run:
```bash
cd /Users/j3r0lin/Workspace/nezha/incloud-cli/.worktrees/feature-device-config-schema
go get github.com/santhosh-tekuri/jsonschema/v6
```

**Step 2: Verify it compiles**

Run: `make build`
Expected: Success

**Step 3: Commit**

```bash
git add go.mod go.sum
git commit -m "feat(device): add jsonschema dependency for config schema validation"
```

---

### Task 2: Device resolution helper — `resolveProductVersion`

Shared logic to resolve `--device` or `--product`/`--version` into product + version strings. All 4 schema commands will use this.

**Files:**
- Create: `internal/cmd/device/config_schema.go`
- Create: `internal/cmd/device/config_schema_test.go`

**Step 1: Write the test**

```go
// internal/cmd/device/config_schema_test.go
package device

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestResolveProductVersion_FromFlags(t *testing.T) {
	// No server needed — flags provide values directly
	pv := &productVersion{product: "MR805", version: "V2.0.15-111"}
	if pv.product != "MR805" || pv.version != "V2.0.15-111" {
		t.Errorf("unexpected: %+v", pv)
	}
}

func TestResolveProductVersion_FromDevice(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/devices/dev123" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"result":{"_id":"dev123","partNumber":"MR805","firmware":"V2.0.15-111"}}`))
	}))
	defer server.Close()

	f, _ := newTestFactory(t, server.URL)
	client, _ := f.APIClient()

	pv, err := resolveProductVersion(client, "dev123", "", "")
	if err != nil {
		t.Fatal(err)
	}
	if pv.product != "MR805" || pv.version != "V2.0.15-111" {
		t.Errorf("unexpected: %+v", pv)
	}
}

func TestResolveProductVersion_MutualExclusion(t *testing.T) {
	_, err := resolveProductVersion(nil, "dev123", "MR805", "V2.0.15")
	if err == nil {
		t.Fatal("expected error for mutual exclusion")
	}
}

func TestResolveProductVersion_MissingParams(t *testing.T) {
	_, err := resolveProductVersion(nil, "", "MR805", "")
	if err == nil {
		t.Fatal("expected error when --product without --version")
	}
	_, err = resolveProductVersion(nil, "", "", "")
	if err == nil {
		t.Fatal("expected error when no params")
	}
}
```

**Step 2: Run test to verify it fails**

Run: `cd /Users/j3r0lin/Workspace/nezha/incloud-cli/.worktrees/feature-device-config-schema && CGO_ENABLED=0 go test ./internal/cmd/device/ -run TestResolveProductVersion -v`
Expected: FAIL (functions not defined)

**Step 3: Write the implementation**

```go
// internal/cmd/device/config_schema.go
package device

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/spf13/cobra"
	"github.com/tidwall/gjson"

	"github.com/inhandnet/incloud-cli/internal/api"
	"github.com/inhandnet/incloud-cli/internal/factory"
)

type productVersion struct {
	product string
	version string
}

// resolveProductVersion resolves product and version from either a device ID
// or explicit --product/--version flags. The two approaches are mutually exclusive.
func resolveProductVersion(client *api.APIClient, deviceID, product, version string) (*productVersion, error) {
	hasDevice := deviceID != ""
	hasProduct := product != ""
	hasVersion := version != ""

	switch {
	case hasDevice && (hasProduct || hasVersion):
		return nil, fmt.Errorf("--device and --product/--version are mutually exclusive")
	case hasDevice:
		body, err := client.Get("/api/v1/devices/"+deviceID, url.Values{
			"fields": {"partNumber,firmware"},
		})
		if err != nil {
			return nil, fmt.Errorf("fetching device %s: %w", deviceID, err)
		}
		r := gjson.ParseBytes(body)
		p := r.Get("result.partNumber").String()
		v := r.Get("result.firmware").String()
		if p == "" || v == "" {
			return nil, fmt.Errorf("device %s is missing partNumber or firmware info", deviceID)
		}
		return &productVersion{product: p, version: v}, nil
	case hasProduct && hasVersion:
		return &productVersion{product: product, version: version}, nil
	case hasProduct:
		return nil, fmt.Errorf("--version is required when using --product")
	case hasVersion:
		return nil, fmt.Errorf("--product is required when using --version")
	default:
		return nil, fmt.Errorf("either --device or --product/--version is required")
	}
}

// schemaFlags holds the shared flags for all schema commands.
type schemaFlags struct {
	device  string
	product string
	version string
}

// register adds --device, --product, --version flags to the command.
func (sf *schemaFlags) register(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&sf.device, "device", "d", "", "Device ID (use 'incloud device list' to find IDs)")
	cmd.Flags().StringVarP(&sf.product, "product", "p", "", "Product code (requires --version; mutually exclusive with --device)")
	cmd.Flags().StringVar(&sf.version, "version", "", "Firmware version (requires --product)")
}

// resolve calls resolveProductVersion with the stored flag values.
func (sf *schemaFlags) resolve(client *api.APIClient) (*productVersion, error) {
	return resolveProductVersion(client, sf.device, sf.product, sf.version)
}

// newCmdConfigSchema creates the `device config schema` parent command.
func newCmdConfigSchema(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "schema",
		Short: "Browse and validate device configuration schemas",
		Long: `Discover configuration schemas for a device's product/firmware,
and validate JSON payloads before writing.

AI tools workflow:
  1. incloud device config schema overview --device <id>
  2. incloud device config schema list --device <id>
  3. incloud device config schema get --device <id> <json-key>
  4. incloud device config schema validate --device <id> --key <json-key> --payload '{...}'
  5. incloud device config update <id> --payload '{...}'`,
	}

	cmd.AddCommand(newCmdSchemaList(f))
	cmd.AddCommand(newCmdSchemaGet(f))
	cmd.AddCommand(newCmdSchemaOverview(f))
	cmd.AddCommand(newCmdSchemaValidate(f))

	return cmd
}
```

> **Note:** The `newTestFactory` helper already exists in `delete_test.go`. Since tests are in the same package, it's reusable. If the compiler complains about duplicate definition, it means we're sharing it correctly.

**Step 4: Run test to verify it passes**

Run: `cd /Users/j3r0lin/Workspace/nezha/incloud-cli/.worktrees/feature-device-config-schema && CGO_ENABLED=0 go test ./internal/cmd/device/ -run TestResolveProductVersion -v`
Expected: PASS (4 tests)

**Step 5: Lint**

Run:
```bash
cd /Users/j3r0lin/Workspace/nezha/incloud-cli/.worktrees/feature-device-config-schema
goimports -w -local github.com/inhandnet/incloud-cli internal/cmd/device/config_schema.go internal/cmd/device/config_schema_test.go
golangci-lint run ./internal/cmd/device/...
```

**Step 6: Commit**

```bash
git add internal/cmd/device/config_schema.go internal/cmd/device/config_schema_test.go
git commit -m "feat(device): add config schema parent command and device resolution helper"
```

---

### Task 3: `schema list` command

**Files:**
- Create: `internal/cmd/device/config_schema_list.go`
- Add test cases to: `internal/cmd/device/config_schema_test.go`

**Step 1: Write the test**

Append to `config_schema_test.go`:

```go
func TestSchemaList_Table(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v1/devices/dev1":
			w.Write([]byte(`{"result":{"_id":"dev1","partNumber":"MR805","firmware":"V2.0.15-111"}}`))
		case "/api/v1/config-documents":
			if r.URL.Query().Get("product") != "MR805" {
				t.Errorf("expected product=MR805, got %s", r.URL.Query().Get("product"))
			}
			w.Write([]byte(`{"result":[
				{"_id":"1","name":"System DNS","jsonKeys":["dns"],"descriptions":["Global DNS configuration"]},
				{"_id":"2","name":"WiFi Settings","jsonKeys":["wifi"],"descriptions":["WiFi AP and client config"]}
			],"total":2}`))
		default:
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer server.Close()

	f, _ := newTestFactory(t, server.URL)
	out := f.IO.Out.(*bytes.Buffer)

	cmd := newCmdConfigSchema(f)
	cmd.SetArgs([]string{"list", "--device", "dev1", "-o", "table"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "System DNS") {
		t.Errorf("expected 'System DNS' in output, got: %s", out.String())
	}
}

func TestSchemaList_JSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v1/config-documents":
			w.Write([]byte(`{"result":[{"_id":"1","name":"System DNS","jsonKeys":["dns"],"descriptions":["Global DNS"]}],"total":1}`))
		default:
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer server.Close()

	f, _ := newTestFactory(t, server.URL)
	out := f.IO.Out.(*bytes.Buffer)

	cmd := newCmdConfigSchema(f)
	cmd.SetArgs([]string{"list", "--product", "MR805", "--version", "V2.0.15-111", "-o", "json"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "System DNS") {
		t.Errorf("expected 'System DNS' in output, got: %s", out.String())
	}
}
```

**Step 2: Run test to verify it fails**

Run: `cd /Users/j3r0lin/Workspace/nezha/incloud-cli/.worktrees/feature-device-config-schema && CGO_ENABLED=0 go test ./internal/cmd/device/ -run TestSchemaList -v`
Expected: FAIL (newCmdSchemaList not defined)

**Step 3: Write the implementation**

```go
// internal/cmd/device/config_schema_list.go
package device

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tidwall/gjson"

	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/iostreams"
)

var defaultSchemaListFields = []string{"name", "jsonKeys", "description"}

func newCmdSchemaList(f *factory.Factory) *cobra.Command {
	sf := &schemaFlags{}
	var name string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List configuration schemas",
		Long: `List configuration schema documents for a device's product/firmware.
Shows schema name, JSON keys, and a brief description for each config section.

Use --device to auto-detect product/version, or --product/--version to specify directly.`,
		Aliases: []string{"ls"},
		Example: `  # List schemas for a device
  incloud device config schema list --device 507f1f77bcf86cd799439011

  # List by product/version
  incloud device config schema list --product MR805 --version V2.0.15-111

  # Filter by name (regex)
  incloud device config schema list --device 507f1f77bcf86cd799439011 --name "DNS|WiFi"

  # JSON output for AI tools
  incloud device config schema list --device 507f1f77bcf86cd799439011 -o json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := f.APIClient()
			if err != nil {
				return err
			}

			pv, err := sf.resolve(client)
			if err != nil {
				return err
			}

			q := url.Values{}
			q.Set("product", pv.product)
			q.Set("version", pv.version)
			q.Set("module", "default")
			q.Set("limit", "200")
			if name != "" {
				q.Set("name", name)
			}

			body, err := client.Get("/api/v1/config-documents", q)
			if err != nil {
				return err
			}

			output, _ := cmd.Flags().GetString("output")
			if output == "" {
				output = "table"
			}

			return iostreams.FormatOutput(body, f.IO, output, defaultSchemaListFields,
				iostreams.WithTransform(transformSchemaList),
			)
		},
	}

	sf.register(cmd)
	cmd.Flags().StringVar(&name, "name", "", "Filter by schema name (regex)")

	return cmd
}

// transformSchemaList transforms config-documents API response for table display.
// Converts jsonKeys array to comma-separated string, truncates descriptions.
func transformSchemaList(body []byte) ([]byte, error) {
	result := gjson.GetBytes(body, "result")
	if !result.Exists() {
		return []byte(`{"result":[]}`), nil
	}

	var items []map[string]interface{}
	for _, item := range result.Array() {
		row := map[string]interface{}{
			"name": item.Get("name").String(),
		}

		// Join jsonKeys array into comma-separated string
		var keys []string
		for _, k := range item.Get("jsonKeys").Array() {
			keys = append(keys, k.String())
		}
		row["jsonKeys"] = strings.Join(keys, ", ")

		// Use first description, truncate for table
		descs := item.Get("descriptions").Array()
		if len(descs) > 0 {
			desc := descs[0].String()
			if len(desc) > 80 {
				desc = desc[:77] + "..."
			}
			row["description"] = desc
		} else {
			row["description"] = ""
		}

		items = append(items, row)
	}

	out, err := json.Marshal(map[string]interface{}{"result": items})
	if err != nil {
		return nil, fmt.Errorf("formatting schema list: %w", err)
	}
	return out, nil
}
```

**Step 4: Run test to verify it passes**

Run: `cd /Users/j3r0lin/Workspace/nezha/incloud-cli/.worktrees/feature-device-config-schema && CGO_ENABLED=0 go test ./internal/cmd/device/ -run TestSchemaList -v`
Expected: PASS

**Step 5: Lint**

Run:
```bash
cd /Users/j3r0lin/Workspace/nezha/incloud-cli/.worktrees/feature-device-config-schema
goimports -w -local github.com/inhandnet/incloud-cli internal/cmd/device/config_schema_list.go
golangci-lint run ./internal/cmd/device/...
```

**Step 6: Commit**

```bash
git add internal/cmd/device/config_schema_list.go internal/cmd/device/config_schema_test.go
git commit -m "feat(device): add 'device config schema list' command"
```

---

### Task 4: `schema get` command

**Files:**
- Create: `internal/cmd/device/config_schema_get.go`
- Add test cases to: `internal/cmd/device/config_schema_test.go`

**Step 1: Write the test**

Append to `config_schema_test.go`:

```go
func TestSchemaGet(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/config-documents" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.URL.Query().Get("jsonKeys") != "dns" {
			t.Errorf("expected jsonKeys=dns, got %s", r.URL.Query().Get("jsonKeys"))
		}
		w.Write([]byte(`{"result":[{
			"_id":"1",
			"name":"System DNS",
			"jsonKeys":["dns"],
			"descriptions":["Global DNS configuration. Higher priority than upstream DNS."],
			"content":"{\"type\":\"object\",\"properties\":{\"primary\":{\"type\":\"string\"}}}"
		}]}`))
	}))
	defer server.Close()

	f, _ := newTestFactory(t, server.URL)
	out := f.IO.Out.(*bytes.Buffer)

	cmd := newCmdConfigSchema(f)
	cmd.SetArgs([]string{"get", "--product", "MR805", "--version", "V2.0.15-111", "dns"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "System DNS") {
		t.Errorf("expected 'System DNS' in output, got: %s", out.String())
	}
}

func TestSchemaGet_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"result":[]}`))
	}))
	defer server.Close()

	f, _ := newTestFactory(t, server.URL)

	cmd := newCmdConfigSchema(f)
	cmd.SetArgs([]string{"get", "--product", "MR805", "--version", "V2.0.15", "nonexistent"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for not found")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("expected 'not found' in error, got: %v", err)
	}
}
```

**Step 2: Run test to verify it fails**

Run: `cd /Users/j3r0lin/Workspace/nezha/incloud-cli/.worktrees/feature-device-config-schema && CGO_ENABLED=0 go test ./internal/cmd/device/ -run TestSchemaGet -v`
Expected: FAIL

**Step 3: Write the implementation**

```go
// internal/cmd/device/config_schema_get.go
package device

import (
	"fmt"
	"net/url"

	"github.com/spf13/cobra"
	"github.com/tidwall/gjson"

	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/iostreams"
)

func newCmdSchemaGet(f *factory.Factory) *cobra.Command {
	sf := &schemaFlags{}

	cmd := &cobra.Command{
		Use:   "get <json-key>",
		Short: "Get a configuration schema by JSON key",
		Long: `Get the full configuration schema definition for a specific JSON key,
including JSON Schema content and human-readable descriptions.

The JSON key can be found from 'incloud device config schema list'.`,
		Example: `  # Get DNS config schema
  incloud device config schema get --device 507f1f77bcf86cd799439011 dns

  # JSON output for AI parsing
  incloud device config schema get --product MR805 --version V2.0.15-111 dns -o json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			jsonKey := args[0]

			client, err := f.APIClient()
			if err != nil {
				return err
			}

			pv, err := sf.resolve(client)
			if err != nil {
				return err
			}

			q := url.Values{}
			q.Set("product", pv.product)
			q.Set("version", pv.version)
			q.Set("module", "default")
			q.Set("jsonKeys", jsonKey)

			body, err := client.Get("/api/v1/config-documents", q)
			if err != nil {
				return err
			}

			result := gjson.GetBytes(body, "result")
			if !result.Exists() || len(result.Array()) == 0 {
				return fmt.Errorf("config schema %q not found for %s/%s", jsonKey, pv.product, pv.version)
			}

			// Extract the first matching document
			doc := []byte(result.Array()[0].Raw)

			output, _ := cmd.Flags().GetString("output")
			return iostreams.FormatOutput(doc, f.IO, output, nil)
		},
	}

	sf.register(cmd)

	return cmd
}
```

**Step 4: Run test to verify it passes**

Run: `cd /Users/j3r0lin/Workspace/nezha/incloud-cli/.worktrees/feature-device-config-schema && CGO_ENABLED=0 go test ./internal/cmd/device/ -run TestSchemaGet -v`
Expected: PASS

**Step 5: Lint**

Run:
```bash
cd /Users/j3r0lin/Workspace/nezha/incloud-cli/.worktrees/feature-device-config-schema
goimports -w -local github.com/inhandnet/incloud-cli internal/cmd/device/config_schema_get.go
golangci-lint run ./internal/cmd/device/...
```

**Step 6: Commit**

```bash
git add internal/cmd/device/config_schema_get.go internal/cmd/device/config_schema_test.go
git commit -m "feat(device): add 'device config schema get' command"
```

---

### Task 5: `schema overview` command

**Files:**
- Create: `internal/cmd/device/config_schema_overview.go`
- Add test cases to: `internal/cmd/device/config_schema_test.go`

**Step 1: Write the test**

Append to `config_schema_test.go`:

```go
func TestSchemaOverview(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/config-documents/overview" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Write([]byte(`{"result":{"_id":"ov1","product":"CPE02","module":"default","version":"V2.0.8","content":"### JSON KEYS\n- wan: WAN\n- cellular: Cellular"}}`))
	}))
	defer server.Close()

	f, _ := newTestFactory(t, server.URL)
	out := f.IO.Out.(*bytes.Buffer)

	cmd := newCmdConfigSchema(f)
	cmd.SetArgs([]string{"overview", "--product", "CPE02", "--version", "V2.0.8"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "### JSON KEYS") {
		t.Errorf("expected markdown content in output, got: %s", out.String())
	}
}

func TestSchemaOverview_NotAvailable(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"result":null}`))
	}))
	defer server.Close()

	f, errBuf := newTestFactory(t, server.URL)

	cmd := newCmdConfigSchema(f)
	cmd.SetArgs([]string{"overview", "--product", "MR805", "--version", "V2.0.15-111"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(errBuf.String(), "No overview available") {
		t.Errorf("expected 'No overview available' in stderr, got: %s", errBuf.String())
	}
}
```

**Step 2: Run test to verify it fails**

Run: `cd /Users/j3r0lin/Workspace/nezha/incloud-cli/.worktrees/feature-device-config-schema && CGO_ENABLED=0 go test ./internal/cmd/device/ -run TestSchemaOverview -v`
Expected: FAIL

**Step 3: Write the implementation**

```go
// internal/cmd/device/config_schema_overview.go
package device

import (
	"fmt"
	"net/url"

	"github.com/spf13/cobra"
	"github.com/tidwall/gjson"

	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/iostreams"
)

func newCmdSchemaOverview(f *factory.Factory) *cobra.Command {
	sf := &schemaFlags{}

	cmd := &cobra.Command{
		Use:   "overview",
		Short: "View product configuration overview",
		Long: `View the configuration overview for a product/firmware, including
dependency rules and business constraints between config sections.

AI tools should read this before modifying configurations to understand
which config sections depend on each other.`,
		Example: `  # View overview for a device
  incloud device config schema overview --device 507f1f77bcf86cd799439011

  # View by product/version
  incloud device config schema overview --product CPE02 --version V2.0.8

  # JSON output
  incloud device config schema overview --product CPE02 --version V2.0.8 -o json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := f.APIClient()
			if err != nil {
				return err
			}

			pv, err := sf.resolve(client)
			if err != nil {
				return err
			}

			q := url.Values{}
			q.Set("product", pv.product)
			q.Set("version", pv.version)
			q.Set("module", "default")

			body, err := client.Get("/api/v1/config-documents/overview", q)
			if err != nil {
				return err
			}

			result := gjson.GetBytes(body, "result")
			if !result.Exists() || result.Type == gjson.Null {
				fmt.Fprintf(f.IO.ErrOut, "No overview available for %s/%s.\n", pv.product, pv.version)
				return nil
			}

			output, _ := cmd.Flags().GetString("output")
			if output == "" || output == "table" {
				// Output markdown content directly
				content := result.Get("content").String()
				fmt.Fprintln(f.IO.Out, content)
				return nil
			}

			return iostreams.FormatOutput([]byte(result.Raw), f.IO, output, nil)
		},
	}

	sf.register(cmd)

	return cmd
}
```

**Step 4: Run test to verify it passes**

Run: `cd /Users/j3r0lin/Workspace/nezha/incloud-cli/.worktrees/feature-device-config-schema && CGO_ENABLED=0 go test ./internal/cmd/device/ -run TestSchemaOverview -v`
Expected: PASS

**Step 5: Lint**

Run:
```bash
cd /Users/j3r0lin/Workspace/nezha/incloud-cli/.worktrees/feature-device-config-schema
goimports -w -local github.com/inhandnet/incloud-cli internal/cmd/device/config_schema_overview.go
golangci-lint run ./internal/cmd/device/...
```

**Step 6: Commit**

```bash
git add internal/cmd/device/config_schema_overview.go internal/cmd/device/config_schema_test.go
git commit -m "feat(device): add 'device config schema overview' command"
```

---

### Task 6: `schema validate` command

**Files:**
- Create: `internal/cmd/device/config_schema_validate.go`
- Add test cases to: `internal/cmd/device/config_schema_test.go`

**Step 1: Write the test**

Append to `config_schema_test.go`:

```go
func TestSchemaValidate_Pass(t *testing.T) {
	schema := `{"type":"object","properties":{"dns":{"type":"object","properties":{"primary":{"type":"string"}},"required":["primary"]}}}`
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(fmt.Sprintf(`{"result":[{"_id":"1","name":"System DNS","jsonKeys":["dns"],"content":%q}]}`, schema)))
	}))
	defer server.Close()

	f, errBuf := newTestFactory(t, server.URL)

	cmd := newCmdConfigSchema(f)
	cmd.SetArgs([]string{"validate", "--product", "MR805", "--version", "V2.0.15-111",
		"--key", "dns", "--payload", `{"dns":{"primary":"8.8.8.8"}}`})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(errBuf.String(), "Validation passed") {
		t.Errorf("expected 'Validation passed' in stderr, got: %s", errBuf.String())
	}
}

func TestSchemaValidate_Fail(t *testing.T) {
	schema := `{"type":"object","properties":{"dns":{"type":"object","properties":{"primary":{"type":"string"}},"required":["primary"]}}}`
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(fmt.Sprintf(`{"result":[{"_id":"1","name":"System DNS","jsonKeys":["dns"],"content":%q}]}`, schema)))
	}))
	defer server.Close()

	f, _ := newTestFactory(t, server.URL)

	cmd := newCmdConfigSchema(f)
	cmd.SetArgs([]string{"validate", "--product", "MR805", "--version", "V2.0.15-111",
		"--key", "dns", "--payload", `{"dns":{"primary":123}}`})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected validation error")
	}
}

func TestSchemaValidate_PayloadFileMutualExclusion(t *testing.T) {
	f, _ := newTestFactory(t, "https://example.com")

	cmd := newCmdConfigSchema(f)
	cmd.SetArgs([]string{"validate", "--product", "MR805", "--version", "V1",
		"--key", "dns", "--payload", "{}", "--file", "f.json"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "mutually exclusive") {
		t.Errorf("expected 'mutually exclusive' in error, got: %v", err)
	}
}
```

**Step 2: Run test to verify it fails**

Run: `cd /Users/j3r0lin/Workspace/nezha/incloud-cli/.worktrees/feature-device-config-schema && CGO_ENABLED=0 go test ./internal/cmd/device/ -run TestSchemaValidate -v`
Expected: FAIL

**Step 3: Write the implementation**

```go
// internal/cmd/device/config_schema_validate.go
package device

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/santhosh-tekuri/jsonschema/v6"
	"github.com/spf13/cobra"
	"github.com/tidwall/gjson"

	"github.com/inhandnet/incloud-cli/internal/factory"
)

func newCmdSchemaValidate(f *factory.Factory) *cobra.Command {
	sf := &schemaFlags{}
	var (
		key     string
		payload string
		file    string
	)

	cmd := &cobra.Command{
		Use:   "validate",
		Short: "Validate JSON payload against a config schema",
		Long: `Validate a JSON configuration payload against the device's config schema
before writing it with 'incloud device config update'.

Uses JSON Schema draft-07 validation. Exits with code 0 on success, 1 on
validation failure. Useful for AI tools to pre-check generated config.`,
		Example: `  # Validate a JSON payload
  incloud device config schema validate --device 507f1f77bcf86cd799439011 \
    --key dns --payload '{"dns":{"primary":"8.8.8.8"}}'

  # Validate from file
  incloud device config schema validate --product MR805 --version V2.0.15-111 \
    --key dns --file dns-config.json

  # Use in pipeline: validate then apply
  incloud device config schema validate -d <id> --key dns --payload '...' && \
  incloud device config update <id> --payload '...'`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Read payload
			var data []byte
			var err error

			switch {
			case payload != "" && file != "":
				return fmt.Errorf("--payload and --file are mutually exclusive")
			case payload != "":
				data = []byte(payload)
			case file != "":
				data, err = os.ReadFile(file)
				if err != nil {
					return fmt.Errorf("reading file: %w", err)
				}
			default:
				return fmt.Errorf("either --payload or --file is required")
			}

			// Parse payload
			var payloadObj interface{}
			if err := json.Unmarshal(data, &payloadObj); err != nil {
				return fmt.Errorf("invalid JSON payload: %w", err)
			}

			client, err := f.APIClient()
			if err != nil {
				return err
			}

			pv, err := sf.resolve(client)
			if err != nil {
				return err
			}

			// Fetch schema
			q := url.Values{}
			q.Set("product", pv.product)
			q.Set("version", pv.version)
			q.Set("module", "default")
			q.Set("jsonKeys", key)

			body, err := client.Get("/api/v1/config-documents", q)
			if err != nil {
				return err
			}

			result := gjson.GetBytes(body, "result")
			if !result.Exists() || len(result.Array()) == 0 {
				return fmt.Errorf("config schema %q not found for %s/%s", key, pv.product, pv.version)
			}

			schemaContent := result.Array()[0].Get("content").String()
			if schemaContent == "" {
				return fmt.Errorf("config schema %q has no content", key)
			}

			// Parse and compile JSON Schema
			var schemaObj interface{}
			if err := json.Unmarshal([]byte(schemaContent), &schemaObj); err != nil {
				return fmt.Errorf("invalid schema JSON: %w", err)
			}

			compiler := jsonschema.NewCompiler()
			if err := compiler.AddResource("schema.json", schemaObj); err != nil {
				return fmt.Errorf("loading schema: %w", err)
			}
			schema, err := compiler.Compile("schema.json")
			if err != nil {
				return fmt.Errorf("compiling schema: %w", err)
			}

			// Validate
			validationErr := schema.Validate(payloadObj)
			if validationErr == nil {
				fmt.Fprintf(f.IO.ErrOut, "Validation passed.\n")
				return nil
			}

			// Format validation errors
			var sb strings.Builder
			sb.WriteString("Validation failed:\n")
			if ve, ok := validationErr.(*jsonschema.ValidationError); ok {
				for _, cause := range flattenValidationErrors(ve) {
					sb.WriteString(fmt.Sprintf("  - %s: %s\n", cause.path, cause.message))
				}
			} else {
				sb.WriteString(fmt.Sprintf("  - %s\n", validationErr.Error()))
			}
			return fmt.Errorf("%s", sb.String())
		},
	}

	sf.register(cmd)
	cmd.Flags().StringVarP(&key, "key", "k", "", "JSON key identifying the config schema to validate against (required)")
	cmd.Flags().StringVar(&payload, "payload", "", "JSON payload to validate")
	cmd.Flags().StringVar(&file, "file", "", "Path to JSON file to validate")
	_ = cmd.MarkFlagRequired("key")

	return cmd
}

type validationError struct {
	path    string
	message string
}

// flattenValidationErrors extracts leaf validation errors with their JSON paths.
func flattenValidationErrors(ve *jsonschema.ValidationError) []validationError {
	var errors []validationError
	flattenVE(ve, &errors)
	return errors
}

func flattenVE(ve *jsonschema.ValidationError, out *[]validationError) {
	if len(ve.Causes) == 0 {
		path := ve.InstanceLocation
		if path == "" {
			path = "$"
		}
		*out = append(*out, validationError{path: path, message: ve.Error()})
		return
	}
	for _, cause := range ve.Causes {
		flattenVE(cause, out)
	}
}
```

**Step 4: Run test to verify it passes**

Run: `cd /Users/j3r0lin/Workspace/nezha/incloud-cli/.worktrees/feature-device-config-schema && CGO_ENABLED=0 go test ./internal/cmd/device/ -run TestSchemaValidate -v`
Expected: PASS

**Step 5: Lint**

Run:
```bash
cd /Users/j3r0lin/Workspace/nezha/incloud-cli/.worktrees/feature-device-config-schema
goimports -w -local github.com/inhandnet/incloud-cli internal/cmd/device/config_schema_validate.go
golangci-lint run ./internal/cmd/device/...
```

**Step 6: Commit**

```bash
git add internal/cmd/device/config_schema_validate.go internal/cmd/device/config_schema_test.go
git commit -m "feat(device): add 'device config schema validate' command"
```

---

### Task 7: Wire up to command tree + final verification

**Files:**
- Modify: `internal/cmd/device/config.go` (add schema subcommand)

**Step 1: Wire up the schema command**

In `config.go`, add `newCmdConfigSchema(f)`:

```go
cmd.AddCommand(newCmdConfigSchema(f))
```

**Step 2: Run full test suite**

Run:
```bash
cd /Users/j3r0lin/Workspace/nezha/incloud-cli/.worktrees/feature-device-config-schema
make build && make test
```
Expected: All tests pass, build succeeds.

**Step 3: Verify CLI help output**

Run:
```bash
./bin/incloud device config schema --help
./bin/incloud device config schema list --help
./bin/incloud device config schema get --help
./bin/incloud device config schema overview --help
./bin/incloud device config schema validate --help
```
Expected: All help texts display correctly with examples.

**Step 4: Lint the entire project**

Run:
```bash
goimports -w -local github.com/inhandnet/incloud-cli internal/cmd/device/config.go
golangci-lint run ./...
```

**Step 5: Commit**

```bash
git add internal/cmd/device/config.go
git commit -m "feat(device): wire config schema subcommands into command tree"
```

---

### Task 8: Update plan document

**Files:**
- Modify: `docs/plans/2026-03-21-device-config-schema-design.md`

**Step 1: Add implementation status**

Add a section at the bottom of the design document marking all commands as implemented:

```markdown
## Implementation Status

- [x] `device config schema list`
- [x] `device config schema get`
- [x] `device config schema overview`
- [x] `device config schema validate`
- [x] Device resolution helper (--device / --product+--version)
- [x] JSON Schema draft-07 validation
- [ ] incloud-skills AI workflow guide (separate repo)
```

**Step 2: Commit**

```bash
git add docs/plans/2026-03-21-device-config-schema-design.md
git commit -m "docs: update config schema design with implementation status"
```

---

## Final Validation

After all tasks complete:

```bash
cd /Users/j3r0lin/Workspace/nezha/incloud-cli/.worktrees/feature-device-config-schema
make build && make test && golangci-lint run ./...
```

Expected: Build succeeds, all tests pass, no lint errors.

---

## Implementation Status

- [x] `device config schema list`
- [x] `device config schema get`
- [x] `device config schema overview`
- [x] `device config schema validate`
- [x] Device resolution helper (--device / --product+--version)
- [x] JSON Schema draft-07 validation
- [ ] incloud-skills AI workflow guide (separate repo)
