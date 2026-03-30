package device

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/cmdutil"
	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/iostreams"
)

var defaultGroupListFields = []string{"_id", "name", "product", "firmware", "createdAt"}
var defaultGroupListSummaryFields = []string{"_id", "name", "product", "firmware", "online", "offline", "total", "createdAt"}

type GroupListOptions struct {
	cmdutil.ListFlags
	Name     string
	Product  []string
	Firmware string
	Org      string
	Summary  bool
}

func newCmdGroupList(f *factory.Factory) *cobra.Command {
	opts := &GroupListOptions{}

	cmd := &cobra.Command{
		Use:     "list",
		Short:   "List device groups",
		Aliases: []string{"ls"},
		Example: `  # List device groups
  incloud device group list

  # Filter by product
  incloud device group list --product ER805

  # Search by name
  incloud device group list --name "Edge"

  # Show device counts per group
  incloud device group list --summary

  # Filter by organization
  incloud device group list --org 60a1b2c3d4e5f6a7b8c9d0e1

  # Expand device count and org info
  incloud device group list --expand org,nezha-iot-device-summary -o json

  # Paginate
  incloud device group list --page 2 --limit 10`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := f.APIClient()
			if err != nil {
				return err
			}

			q := cmdutil.NewQuery(cmd, nil)
			if opts.Name != "" {
				q.Set("name", opts.Name)
			}
			if opts.Firmware != "" {
				q.Set("firmware", opts.Firmware)
			}
			for _, p := range opts.Product {
				q.Add("product", p)
			}
			if opts.Org != "" {
				q.Set("oid", opts.Org)
			}
			output, _ := cmd.Flags().GetString("output")
			if !cmd.Flags().Changed("fields") && (output == "" || output == "table") {
				if opts.Summary {
					q.Set("fields", strings.Join(defaultGroupListSummaryFields, ","))
				} else {
					q.Set("fields", strings.Join(defaultGroupListFields, ","))
				}
			}

			body, err := client.Get("/api/v1/devicegroups", q)
			if err != nil {
				return err
			}

			if opts.Summary {
				merged, err := mergeGroupSummary(client, body)
				if err != nil {
					fmt.Fprintf(f.IO.ErrOut, "Warning: %v\n", err)
				} else {
					body = merged
				}
			}

			return iostreams.FormatOutput(body, f.IO, output)
		},
	}

	opts.ListFlags.Register(cmd)
	cmd.Flags().StringVar(&opts.Name, "name", "", "Filter by group name (fuzzy match)")
	cmd.Flags().StringArrayVar(&opts.Product, "product", nil, "Filter by product (can be repeated)")
	cmd.Flags().StringVar(&opts.Firmware, "firmware", "", "Filter by firmware version (fuzzy match)")
	cmd.Flags().StringVar(&opts.Org, "org", "", "Filter by organization ID")
	cmd.Flags().BoolVar(&opts.Summary, "summary", false, "Include device counts (online/offline/total) per group")
	opts.ListFlags.RegisterExpand(cmd, "org", "nezha-iot-device-summary")

	return cmd
}

// mergeGroupSummary fetches device summary for all groups in the list response
// and merges online/offline/total counts into each group object.
func mergeGroupSummary(client interface {
	Post(path string, body interface{}) ([]byte, error)
}, body []byte) ([]byte, error) {
	var listResp struct {
		Result []map[string]any `json:"result"`
		Total  int              `json:"total"`
		Page   int              `json:"page"`
		Limit  int              `json:"limit"`
	}
	if err := json.Unmarshal(body, &listResp); err != nil {
		return nil, fmt.Errorf("failed to parse group list: %w", err)
	}
	if len(listResp.Result) == 0 {
		return body, nil
	}

	ids := make([]string, len(listResp.Result))
	for i, g := range listResp.Result {
		ids[i], _ = g["_id"].(string)
	}

	summaryBody, err := client.Post("/api/v1/devicegroups/devices/summary", map[string]any{"ids": ids})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch device summary: %w", err)
	}

	var summaryResp struct {
		Result []struct {
			GID          string `json:"gid"`
			Online       int64  `json:"online"`
			Offline      int64  `json:"offline"`
			Incompatible int64  `json:"incompatible"`
			Total        int64  `json:"total"`
		} `json:"result"`
	}
	if err := json.Unmarshal(summaryBody, &summaryResp); err != nil {
		return nil, fmt.Errorf("failed to parse device summary: %w", err)
	}

	summaryMap := make(map[string]int, len(summaryResp.Result))
	for i, s := range summaryResp.Result {
		summaryMap[s.GID] = i
	}

	for _, g := range listResp.Result {
		gid, _ := g["_id"].(string)
		if idx, ok := summaryMap[gid]; ok {
			s := summaryResp.Result[idx]
			g["online"] = s.Online
			g["offline"] = s.Offline
			g["total"] = s.Total
		}
	}

	return json.Marshal(listResp)
}
