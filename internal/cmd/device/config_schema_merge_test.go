package device

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
)

// qosSchema mirrors the shape that exposed IM-3188: "uplink_rules" and
// "user_rules" are both required at the block level, so a payload that only
// carries "user_rules" is rejected as a whole document even though
// 'config update' merges it happily.
const qosSchema = `{
  "type": "object",
  "properties": {
    "qos": {
      "type": "object",
      "properties": {
        "uplink_rules": {"type": "object"},
        "user_rules": {"type": "array", "items": {
          "type": "object",
          "properties": {"uuid": {"type": "string"}, "name": {"type": "string"}},
          "required": ["uuid", "name"]
        }}
      },
      "required": ["uplink_rules", "user_rules"],
      "additionalProperties": false
    }
  },
  "required": ["qos"],
  "additionalProperties": false
}`

const qosCurrent = `{
  "qos": {"uplink_rules": {"0000f0804da7846f": {"interface": "wan1"}}, "user_rules": []},
  "lan": {"irrelevant": true},
  "dns": {"dns1": ""}
}`

// newConfigServer serves the three endpoints validate needs: device lookup,
// schema lookup, and the device's current merged config.
func newConfigServer(t *testing.T, schema, current string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasSuffix(r.URL.Path, "/merge-config"):
			fmt.Fprintf(w, `{"result":%s}`, current)
		case r.URL.Path == "/api/v1/config-documents":
			fmt.Fprintf(w, `{"result":[{"_id":"1","name":"Network QoS","jsonKeys":["qos"],"content":%q}]}`, schema)
		default:
			fmt.Fprint(w, `{"result":{"_id":"dev123","product":"FWA02-NAVA","firmware":"V2.0.16"}}`)
		}
	}))
}

func runValidate(t *testing.T, srv *httptest.Server, args ...string) (string, error) {
	t.Helper()
	f, errBuf := newTestFactory(t, srv.URL)
	root := newSchemaRoot(f)
	root.SetArgs(append([]string{"schema", "validate"}, args...))
	err := root.Execute()
	return errBuf.String(), err
}

// This is the IM-3188 regression: before the fix this payload failed with
// "missing property 'uplink_rules'" while 'config update' accepted it.
func TestSchemaValidate_DeviceMergesIncrementalPayload(t *testing.T) {
	srv := newConfigServer(t, qosSchema, qosCurrent)
	defer srv.Close()

	out, err := runValidate(t, srv, "--device", "dev123", "--key", "qos",
		"--payload", `{"qos":{"user_rules":[{"uuid":"a","name":"r1"}]}}`)
	if err != nil {
		t.Fatalf("incremental payload should validate against the merged config, got: %v", err)
	}
	if !strings.Contains(out, "against merged config") {
		t.Errorf("expected merged-config wording, got: %s", out)
	}
}

func TestSchemaValidate_WholeDocumentFlagKeepsStrictMode(t *testing.T) {
	srv := newConfigServer(t, qosSchema, qosCurrent)
	defer srv.Close()

	_, err := runValidate(t, srv, "--device", "dev123", "--key", "qos", "--whole-document",
		"--payload", `{"qos":{"user_rules":[{"uuid":"a","name":"r1"}]}}`)
	if err == nil {
		t.Fatal("--whole-document should keep the strict whole-document behaviour")
	}
	if !strings.Contains(err.Error(), "uplink_rules") {
		t.Errorf("expected block-level required error, got: %v", err)
	}
}

func TestSchemaValidate_BlockLevelHintWithoutDevice(t *testing.T) {
	srv := newConfigServer(t, qosSchema, qosCurrent)
	defer srv.Close()

	_, err := runValidate(t, srv, "--product", "FWA02-NAVA", "--version", "V2.0.16",
		"--key", "qos", "--payload", `{"qos":{"user_rules":[{"uuid":"a","name":"r1"}]}}`)
	if err == nil {
		t.Fatal("expected failure without a device to merge onto")
	}
	if !strings.Contains(err.Error(), "--device") {
		t.Errorf("expected a hint pointing at --device, got: %v", err)
	}
}

// Merging must not turn validate into a rubber stamp: errors inside the payload
// itself still have to fail.
func TestSchemaValidate_ItemLevelErrorsStillFail(t *testing.T) {
	srv := newConfigServer(t, qosSchema, qosCurrent)
	defer srv.Close()

	_, err := runValidate(t, srv, "--device", "dev123", "--key", "qos",
		"--payload", `{"qos":{"user_rules":[{"uuid":"a"}]}}`)
	if err == nil {
		t.Fatal("item-level required field is missing, should fail")
	}
	if !strings.Contains(err.Error(), "name") {
		t.Errorf("expected the missing item field to be reported, got: %v", err)
	}
}

// null in a merge patch deletes the key, so it must be able to break a
// block-level required constraint that the current config was satisfying.
func TestSchemaValidate_NullDeletionFails(t *testing.T) {
	srv := newConfigServer(t, qosSchema, qosCurrent)
	defer srv.Close()

	_, err := runValidate(t, srv, "--device", "dev123", "--key", "qos",
		"--payload", `{"qos":{"uplink_rules":null}}`)
	if err == nil {
		t.Fatal("deleting a required property should fail")
	}
	if !strings.Contains(err.Error(), "uplink_rules") {
		t.Errorf("expected uplink_rules to be reported missing, got: %v", err)
	}
}

// Stored configs are not always schema-clean. Violations that already exist must
// not fail a payload that does not touch them.
func TestSchemaValidate_PreExistingViolationsDoNotFail(t *testing.T) {
	// The current config carries an extra property the schema forbids.
	dirty := `{"qos":{"uplink_rules":{"a":{}},"user_rules":[],"stale_field":1}}`
	srv := newConfigServer(t, qosSchema, dirty)
	defer srv.Close()

	out, err := runValidate(t, srv, "--device", "dev123", "--key", "qos",
		"--payload", `{"qos":{"user_rules":[{"uuid":"a","name":"r1"}]}}`)
	if err != nil {
		t.Fatalf("pre-existing violation must not fail this payload, got: %v", err)
	}
	if !strings.Contains(out, "already violates") {
		t.Errorf("pre-existing violations should still be reported, got: %s", out)
	}
}

// Falling back to whole-document validation when the current config cannot be
// read would recreate the bug while looking like a success.
func TestSchemaValidate_MergeConfigFailureDoesNotFallBack(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasSuffix(r.URL.Path, "/merge-config"):
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprint(w, `{"error":"boom"}`)
		case r.URL.Path == "/api/v1/config-documents":
			fmt.Fprintf(w, `{"result":[{"_id":"1","jsonKeys":["qos"],"content":%q}]}`, qosSchema)
		default:
			fmt.Fprint(w, `{"result":{"_id":"dev123","product":"FWA02-NAVA","firmware":"V2.0.16"}}`)
		}
	}))
	defer srv.Close()

	out, err := runValidate(t, srv, "--device", "dev123", "--key", "qos",
		"--payload", `{"qos":{"user_rules":[]}}`)
	if err == nil {
		t.Fatal("expected an error when the current config cannot be fetched")
	}
	if strings.Contains(out, "Validation passed") {
		t.Errorf("must not report success after failing to read the current config: %s", out)
	}
}

func TestApplyMergePatch(t *testing.T) {
	tests := []struct {
		name          string
		target, patch string
		want          string
	}{
		{"object deep merge", `{"a":{"x":1,"y":2}}`, `{"a":{"y":3}}`, `{"a":{"x":1,"y":3}}`},
		{"array replaced wholesale", `{"a":[1,2,3]}`, `{"a":[9]}`, `{"a":[9]}`},
		{"null deletes", `{"a":1,"b":2}`, `{"b":null}`, `{"a":1}`},
		{"new key added", `{"a":1}`, `{"b":2}`, `{"a":1,"b":2}`},
		{"scalar replaces object", `{"a":{"x":1}}`, `{"a":5}`, `{"a":5}`},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := applyMergePatch(mustJSON(t, tc.target), mustJSON(t, tc.patch))
			if !jsonEqual(t, got, mustJSON(t, tc.want)) {
				t.Errorf("got %v, want %s", got, tc.want)
			}
		})
	}
}

func TestPruneToKeys(t *testing.T) {
	cfg := mustJSONMap(t, `{"qos":{"a":1},"lan":{"b":2},"l2tp":{"clients":{"c":3},"server":{"d":4}}}`)

	got := pruneToKeys(cfg, []string{"qos"})
	if !jsonEqual(t, got, mustJSON(t, `{"qos":{"a":1}}`)) {
		t.Errorf("single key: got %v", got)
	}

	// Dotted keys must not drag in sibling branches the schema does not govern.
	got = pruneToKeys(cfg, []string{"l2tp.clients"})
	if !jsonEqual(t, got, mustJSON(t, `{"l2tp":{"clients":{"c":3}}}`)) {
		t.Errorf("dotted key: got %v", got)
	}

	// Absent paths stay absent rather than becoming null.
	got = pruneToKeys(cfg, []string{"missing"})
	if len(got) != 0 {
		t.Errorf("absent key should yield empty object, got %v", got)
	}
}

// The validator builds these messages from Go map iteration, so the same
// violation comes back with the property names in a different order. The diff
// must not treat that as a newly introduced error.
func TestSubtractErrors_IgnoresQuotedOrder(t *testing.T) {
	base := []validationError{{path: "/a", message: "additional properties 'x', 'y', 'z' not allowed"}}
	got := []validationError{{path: "/a", message: "additional properties 'z', 'x', 'y' not allowed"}}

	if diff := subtractErrors(got, base); len(diff) != 0 {
		t.Errorf("reordered property names should compare equal, got %v", diff)
	}

	newErr := []validationError{{path: "/a", message: "additional properties 'q' not allowed"}}
	if diff := subtractErrors(newErr, base); len(diff) != 1 {
		t.Errorf("a genuinely different error must survive, got %v", diff)
	}
}

func mustJSON(t *testing.T, s string) interface{} {
	t.Helper()
	var v interface{}
	if err := jsonUnmarshalString(s, &v); err != nil {
		t.Fatalf("bad test JSON %s: %v", s, err)
	}
	return v
}

func mustJSONMap(t *testing.T, s string) map[string]interface{} {
	t.Helper()
	v, ok := mustJSON(t, s).(map[string]interface{})
	if !ok {
		t.Fatalf("not an object: %s", s)
	}
	return v
}

func jsonUnmarshalString(s string, v interface{}) error {
	return json.Unmarshal([]byte(s), v)
}

func jsonEqual(t *testing.T, a, b interface{}) bool {
	t.Helper()
	return reflect.DeepEqual(a, b)
}
