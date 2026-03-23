package device

import (
	"encoding/json"
	"fmt"

	"github.com/pmezard/go-difflib/difflib"
	"github.com/spf13/cobra"
	"github.com/tidwall/gjson"

	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/iostreams"
)

func newCmdConfigHistoryDiff(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "diff <device-id> <snapshot-id-1> <snapshot-id-2>",
		Short: "Compare two configuration snapshots",
		Long:  "Compare two configuration snapshots side by side using unified diff format.",
		Example: `  # Compare two snapshots
  incloud device config snapshots diff 507f1f77bcf86cd799439011 SNAP_ID_1 SNAP_ID_2

  # JSON output with structured differences
  incloud device config snapshots diff 507f1f77bcf86cd799439011 SNAP_ID_1 SNAP_ID_2 -o json`,
		Args: cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			deviceID := args[0]
			snapshotID1 := args[1]
			snapshotID2 := args[2]

			client, err := f.APIClient()
			if err != nil {
				return err
			}

			// Fetch both snapshots concurrently
			type fetchResult struct {
				body []byte
				err  error
			}
			ch := make(chan fetchResult, 1)
			go func() {
				p := fmt.Sprintf("/api/v1/devices/%s/config/history/%s", deviceID, snapshotID2)
				b, e := client.Get(p, nil)
				ch <- fetchResult{b, e}
			}()

			path1 := fmt.Sprintf("/api/v1/devices/%s/config/history/%s", deviceID, snapshotID1)
			body1, err := client.Get(path1, nil)
			if err != nil {
				return fmt.Errorf("fetching snapshot %s: %w", snapshotID1, err)
			}

			r2 := <-ch
			if r2.err != nil {
				return fmt.Errorf("fetching snapshot %s: %w", snapshotID2, r2.err)
			}
			body2 := r2.body

			config1 := gjson.GetBytes(body1, "result.mergedConfig")
			config2 := gjson.GetBytes(body2, "result.mergedConfig")

			yaml1, err := configToYAML(&config1)
			if err != nil {
				return fmt.Errorf("serializing snapshot %s config: %w", snapshotID1, err)
			}
			yaml2, err := configToYAML(&config2)
			if err != nil {
				return fmt.Errorf("serializing snapshot %s config: %w", snapshotID2, err)
			}

			output, _ := cmd.Flags().GetString("output")
			if output == "json" {
				return outputJSONDiff(f, snapshotID1, snapshotID2, &config1, &config2)
			}

			diff := difflib.UnifiedDiff{
				A:        difflib.SplitLines(yaml1),
				B:        difflib.SplitLines(yaml2),
				FromFile: snapshotID1,
				ToFile:   snapshotID2,
				Context:  3,
			}

			text, err := difflib.GetUnifiedDiffString(diff)
			if err != nil {
				return fmt.Errorf("computing diff: %w", err)
			}

			if text == "" {
				fmt.Fprintln(f.IO.Out, "No differences found.")
				return nil
			}

			fmt.Fprint(f.IO.Out, text)
			return nil
		},
	}

	return cmd
}

func configToYAML(r *gjson.Result) (string, error) {
	if !r.Exists() || r.Type == gjson.Null {
		return "", nil
	}
	s, err := iostreams.FormatYAML([]byte(r.Raw))
	if err != nil {
		return "", err
	}
	return s + "\n", nil
}

func outputJSONDiff(f *factory.Factory, id1, id2 string, config1, config2 *gjson.Result) error {
	diffs := collectDiffs("", config1, config2)

	result := map[string]any{
		"snapshot1":   id1,
		"snapshot2":   id2,
		"differences": diffs,
	}

	out, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}
	fmt.Fprintln(f.IO.Out, string(out))
	return nil
}

func collectDiffs(prefix string, a, b *gjson.Result) []map[string]any {
	var diffs []map[string]any

	if a.IsObject() && b.IsObject() {
		keys := make(map[string]bool)
		a.ForEach(func(key, _ gjson.Result) bool {
			keys[key.String()] = true
			return true
		})
		b.ForEach(func(key, _ gjson.Result) bool {
			keys[key.String()] = true
			return true
		})

		for key := range keys {
			path := key
			if prefix != "" {
				path = prefix + "." + key
			}
			va := a.Get(key)
			vb := b.Get(key)

			switch {
			case !va.Exists():
				diffs = append(diffs, map[string]any{
					"path": path,
					"old":  nil,
					"new":  jsonValue(&vb),
				})
			case !vb.Exists():
				diffs = append(diffs, map[string]any{
					"path": path,
					"old":  jsonValue(&va),
					"new":  nil,
				})
			case va.IsObject() && vb.IsObject():
				diffs = append(diffs, collectDiffs(path, &va, &vb)...)
			case va.Raw != vb.Raw:
				diffs = append(diffs, map[string]any{
					"path": path,
					"old":  jsonValue(&va),
					"new":  jsonValue(&vb),
				})
			}
		}
		return diffs
	}

	if a.Raw != b.Raw {
		diffs = append(diffs, map[string]any{
			"path": prefix,
			"old":  jsonValue(a),
			"new":  jsonValue(b),
		})
	}
	return diffs
}

func jsonValue(r *gjson.Result) any {
	if !r.Exists() {
		return nil
	}
	var v any
	if err := json.Unmarshal([]byte(r.Raw), &v); err != nil {
		return r.String()
	}
	return v
}
