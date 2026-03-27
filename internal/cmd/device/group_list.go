package device

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/iostreams"
)

var defaultGroupListFields = []string{"_id", "name", "product", "firmware", "createdAt"}
var defaultGroupListSummaryFields = []string{"_id", "name", "product", "firmware", "online", "offline", "total", "createdAt"}

type GroupListOptions struct {
	Page     int
	Limit    int
	Sort     string
	Name     string
	Product  []string
	Firmware string
	Fields   []string
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

  # Paginate
  incloud device group list --page 2 --limit 10`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := f.APIClient()
			if err != nil {
				return err
			}

			q := url.Values{}
			q.Set("page", strconv.Itoa(opts.Page-1))
			q.Set("limit", strconv.Itoa(opts.Limit))
			if opts.Sort != "" {
				q.Set("sort", opts.Sort)
			}
			if opts.Name != "" {
				q.Set("name", opts.Name)
			}
			if opts.Firmware != "" {
				q.Set("firmware", opts.Firmware)
			}
			for _, p := range opts.Product {
				q.Add("product", p)
			}
			output, _ := cmd.Flags().GetString("output")
			fields := opts.Fields
			if len(fields) == 0 && output == "table" {
				if opts.Summary {
					fields = defaultGroupListSummaryFields
				} else {
					fields = defaultGroupListFields
				}
			}
			if len(fields) > 0 {
				q.Set("fields", strings.Join(fields, ","))
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

	cmd.Flags().IntVar(&opts.Page, "page", 1, "Page number (starting from 1)")
	cmd.Flags().IntVar(&opts.Limit, "limit", 20, "Number of items per page")
	cmd.Flags().StringVar(&opts.Sort, "sort", "", `Sort order (e.g. "createdAt,desc")`)
	cmd.Flags().StringVar(&opts.Name, "name", "", "Filter by group name (fuzzy match)")
	cmd.Flags().StringArrayVar(&opts.Product, "product", nil, "Filter by product (can be repeated)")
	cmd.Flags().StringVar(&opts.Firmware, "firmware", "", "Filter by firmware version (fuzzy match)")
	cmd.Flags().StringSliceVarP(&opts.Fields, "fields", "f", nil, "Fields to return and display")
	cmd.Flags().BoolVar(&opts.Summary, "summary", false, "Include device counts (online/offline/total) per group")

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
