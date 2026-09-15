package device

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/url"
	"regexp"
	"sort"
	"strings"

	"github.com/santhosh-tekuri/jsonschema/v6"
	"github.com/tidwall/gjson"

	"github.com/inhandnet/incloud-cli/internal/api"
)

// fetchMergedConfig returns the device's current merged configuration, which is
// the same document 'config update' merges its payload into.
func fetchMergedConfig(client *api.APIClient, deviceID string) (map[string]interface{}, error) {
	q := url.Values{}
	q.Set("module", defaultConfigModule)

	body, err := client.Get("/api/v1/devices/"+deviceID+"/merge-config", q)
	if err != nil {
		return nil, fmt.Errorf("fetching current config for device %s: %w", deviceID, err)
	}

	result := gjson.GetBytes(body, "result")
	if !result.Exists() {
		return nil, fmt.Errorf("fetching current config for device %s: response has no result", deviceID)
	}

	var current map[string]interface{}
	if err := json.Unmarshal([]byte(result.Raw), &current); err != nil {
		return nil, fmt.Errorf("parsing current config for device %s: %w", deviceID, err)
	}
	if current == nil {
		current = map[string]interface{}{}
	}
	return current, nil
}

// pruneToKeys narrows a full device configuration down to the paths a single
// schema document governs, given as dotted JSON keys such as "qos" or
// "l2tp.clients".
//
// This matters because each schema document declares additionalProperties:false
// around its own keys. Validating a whole device configuration against the "qos"
// schema would report every unrelated top-level block as an additional property,
// so only the governed subtree may be used as the merge base.
func pruneToKeys(config map[string]interface{}, keys []string) map[string]interface{} {
	out := map[string]interface{}{}
	for _, key := range keys {
		copyPath(config, out, strings.Split(key, "."))
	}
	return out
}

// copyPath copies the value at segments from src into dst, creating intermediate
// objects as needed. Missing paths are skipped, leaving the key absent in dst so
// the schema sees it as absent rather than null.
func copyPath(src, dst map[string]interface{}, segments []string) {
	head := segments[0]
	value, ok := src[head]
	if !ok {
		return
	}

	if len(segments) == 1 {
		dst[head] = value
		return
	}

	child, ok := value.(map[string]interface{})
	if !ok {
		// An intermediate segment is not an object, so the deeper path does not
		// exist. Keep what is there so the schema can report the type mismatch.
		dst[head] = value
		return
	}

	next, ok := dst[head].(map[string]interface{})
	if !ok {
		next = map[string]interface{}{}
		dst[head] = next
	}
	copyPath(child, next, segments[1:])
}

// applyMergePatch applies an RFC 7386 JSON Merge Patch to target and returns the
// result. Objects are merged recursively, null removes a key, and every other
// value (arrays and scalars included) replaces what was there.
//
// This mirrors the server's write path: nezha-device-config converts the update
// body with NullableUpdates/Updates.convertMapToUpdate, which recurses into maps
// to build dotted $set paths, replaces arrays and scalars wholesale, and turns
// null into $unset.
func applyMergePatch(target, patch interface{}) interface{} {
	patchMap, ok := patch.(map[string]interface{})
	if !ok {
		return patch
	}

	targetMap, ok := target.(map[string]interface{})
	if !ok {
		targetMap = map[string]interface{}{}
	}

	out := make(map[string]interface{}, len(targetMap)+len(patchMap))
	for k, v := range targetMap {
		out[k] = v
	}
	for k, v := range patchMap {
		if v == nil {
			delete(out, k)
			continue
		}
		out[k] = applyMergePatch(out[k], v)
	}
	return out
}

// hasBlockLevelRequired reports whether any error is a missing required property
// at the document root or at a top-level config block. Those are exactly the
// errors an incremental payload trips over while 'config update' would still
// accept it.
func hasBlockLevelRequired(causes []validationError) bool {
	for _, c := range causes {
		if c.depth <= 1 && strings.Contains(c.message, "missing propert") {
			return true
		}
	}
	return false
}

// validationErrors normalises a schema validation result into leaf errors.
// A nil error yields no entries.
func validationErrors(err error) []validationError {
	if err == nil {
		return nil
	}
	var ve *jsonschema.ValidationError
	if errors.As(err, &ve) {
		return flattenValidationErrors(ve)
	}
	return []validationError{{path: "$", message: err.Error()}}
}

// subtractErrors returns the errors in got that are not already in base, i.e.
// the violations this payload introduces rather than the ones the stored
// configuration already had.
func subtractErrors(got, base []validationError) []validationError {
	if len(base) == 0 {
		return got
	}
	seen := make(map[string]bool, len(base))
	for _, e := range base {
		seen[errorKey(e)] = true
	}
	var out []validationError
	for _, e := range got {
		if !seen[errorKey(e)] {
			out = append(out, e)
		}
	}
	return out
}

// quotedToken matches the quoted property names the validator embeds in its
// messages, e.g. additional properties 'a', 'b' not allowed.
var quotedToken = regexp.MustCompile(`'[^']*'`)

// errorKey builds a comparison key that ignores the order of quoted property
// names. The validator derives those from Go map iteration, so the same
// violation is reported with a different ordering from one run to the next and
// a plain string comparison would treat it as a new error.
func errorKey(e validationError) string {
	var tokens []string
	template := quotedToken.ReplaceAllStringFunc(e.message, func(m string) string {
		tokens = append(tokens, m)
		return "\x01"
	})
	sort.Strings(tokens)
	return e.path + "\x00" + template + "\x00" + strings.Join(tokens, ",")
}

// writePreExistingNote reports violations that the stored configuration already
// had. They do not fail the payload, but hiding them entirely would misreport
// the merged document as fully schema-clean.
func writePreExistingNote(w io.Writer, preExisting []validationError) {
	if len(preExisting) == 0 {
		return
	}
	fmt.Fprintf(w, "\nNote: the device's current configuration already violates this schema "+
		"in %d place(s), independent of this payload:\n", len(preExisting))
	for _, e := range preExisting {
		fmt.Fprintf(w, "  - %s: %s\n", e.path, e.message)
	}
}
