package device

import (
	"encoding/json"
	"fmt"
	"math"
	"sort"

	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/iostreams"
)

type apListOptions struct {
	Snapshot bool
}

var apListColumns = []string{"level", "mac", "alias", "productModel", "serialNumber", "status", "ip", "connectionType", "upperMac", "upperRssi", "staCount", "onlineTime"}

func newCmdAPList(f *factory.Factory) *cobra.Command {
	opts := &apListOptions{}

	cmd := &cobra.Command{
		Use:     "list <device-id>",
		Short:   "List APs managed by an AC device",
		Aliases: []string{"ls"},
		Long: `List the APs managed by an AC device, with their uplink topology.

By default the AC is queried in real time (this can take up to ~30 seconds). If the
AC is offline, use --snapshot to read the last networking state successfully fetched.

Topology: level 0 is the AC itself; every other AP hangs off the device whose MAC is
its upperMac. Table output is sorted by level and shows the number of connected
stations (staCount); use -o json to get the full staList of each AP.`,
		Example: `  # List APs of an AC device (real time)
  incloud device ap list 507f1f77bcf86cd799439011

  # Read the last fetched state (works while the AC is offline)
  incloud device ap list 507f1f77bcf86cd799439011 --snapshot

  # Full output including connected stations
  incloud device ap list 507f1f77bcf86cd799439011 -o json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := f.APIClient()
			if err != nil {
				return err
			}

			path := "/api/v1/network/devices/" + args[0] + "/ap-list"
			if opts.Snapshot {
				path += "/snapshot"
			}
			body, err := client.Get(path, nil)
			if err != nil {
				return err
			}

			if err := checkAPListResult(body, opts.Snapshot); err != nil {
				return err
			}

			output, _ := cmd.Flags().GetString("output")
			return iostreams.FormatOutput(body, f.IO, output,
				iostreams.WithTransform(apListTableRows),
				iostreams.WithColumns(apListColumns...),
			)
		},
	}

	cmd.Flags().BoolVar(&opts.Snapshot, "snapshot", false, "Read the last successfully fetched state instead of querying the AC")

	return cmd
}

// checkAPListResult turns the business errors that the API reports with HTTP 200
// (error/message in the body) into command errors, so they are not shown as an
// empty table.
func checkAPListResult(body []byte, snapshot bool) error {
	var resp struct {
		Error   string          `json:"error"`
		Message string          `json:"message"`
		Result  json.RawMessage `json:"result"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return fmt.Errorf("parsing response: %w", err)
	}

	switch resp.Error {
	case "":
	case "unlicensed_device":
		return fmt.Errorf("device is not licensed: AC-AP networking requires a valid license (%s)", resp.Message)
	case "device offline":
		if snapshot {
			return fmt.Errorf("device is offline")
		}
		return fmt.Errorf("device is offline; use --snapshot to read the last fetched state")
	case "timeout":
		return fmt.Errorf("request timed out waiting for the device; retry or use --snapshot")
	default:
		if resp.Message != "" {
			return fmt.Errorf("failed to get AP list: %s: %s", resp.Error, resp.Message)
		}
		return fmt.Errorf("failed to get AP list: %s", resp.Error)
	}

	if snapshot && (len(resp.Result) == 0 || string(resp.Result) == "null") {
		return fmt.Errorf("no networking snapshot for this device yet; run without --snapshot while the AC is online")
	}
	return nil
}

// apListTableRows flattens result.apList into table rows sorted by level,
// replacing staList with its length.
func apListTableRows(body []byte) ([]byte, error) {
	var resp struct {
		Result struct {
			APList []map[string]any `json:"apList"`
		} `json:"result"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("parsing response: %w", err)
	}

	rows := resp.Result.APList
	for _, ap := range rows {
		staList, _ := ap["staList"].([]any)
		ap["staCount"] = len(staList)
		delete(ap, "staList")
	}
	// Offline APs report no level; keep them after the topology.
	level := func(ap map[string]any) float64 {
		if l, ok := ap["level"].(float64); ok {
			return l
		}
		return math.MaxFloat64
	}
	sort.SliceStable(rows, func(i, j int) bool {
		return level(rows[i]) < level(rows[j])
	})
	if rows == nil {
		rows = []map[string]any{}
	}
	return json.Marshal(rows)
}
