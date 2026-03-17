package device

import (
	"encoding/json"
	"fmt"
	"net/url"
	"sort"

	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/iostreams"
)

type datausageListOptions struct {
	After  string
	Before string
	Groups []string
	Fields []string
}

var defaultDatausageListFields = []string{"deviceId", "sim.tx", "sim.rx", "sim.total", "esim.tx", "esim.rx", "esim.total"}

func newCmdDatausageList(f *factory.Factory) *cobra.Command {
	opts := &datausageListOptions{}

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List data usage per device",
		Long:  "List aggregated data usage for all devices, broken down by interface type (sim, esim, wan, etc.).",
		Example: `  # List all devices' data usage
  incloud device datausage list

  # Filter by time range
  incloud device datausage list --after 2024-03-01 --before 2024-03-31

  # Filter by device groups
  incloud device datausage list --groups 507f1f77bcf86cd799439011

  # Table with custom fields
  incloud device datausage list -o table -f deviceId -f sim.total -f esim.total`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := f.APIClient()
			if err != nil {
				return err
			}

			q := url.Values{
				"after":  {opts.After},
				"before": {opts.Before},
			}
			for _, g := range opts.Groups {
				q.Add("groups", g)
			}

			body, err := client.Get("/api/v1/devices/datausage/details", q)
			if err != nil {
				return err
			}

			output, _ := cmd.Flags().GetString("output")
			fields := opts.Fields
			if len(fields) == 0 {
				fields = defaultDatausageListFields
			}
			return iostreams.FormatOutput(body, f.IO, output, fields, iostreams.WithTransform(flattenDatausageDetails))
		},
	}

	cmd.Flags().StringVar(&opts.After, "after", "", "Start date (e.g. 2024-03-01)")
	cmd.Flags().StringVar(&opts.Before, "before", "", "End date (e.g. 2024-03-31)")
	cmd.Flags().StringSliceVar(&opts.Groups, "groups", nil, "Filter by device group IDs")
	cmd.Flags().StringSliceVarP(&opts.Fields, "fields", "f", nil, "Fields to display in table mode")

	return cmd
}

// flattenDatausageDetails strips the redundant "time" field from each nested
// interface type and sorts results by deviceId for stable output.
// The nested structure is preserved so that FormatTable can resolve
// dot-paths like "sim.tx" or "esim.total" via gjson.
func flattenDatausageDetails(body []byte) ([]byte, error) {
	var envelope struct {
		Result []map[string]interface{} `json:"result"`
	}
	if err := json.Unmarshal(body, &envelope); err != nil {
		return nil, fmt.Errorf("parsing details response: %w", err)
	}

	for _, device := range envelope.Result {
		for key, val := range device {
			if key == "deviceId" {
				continue
			}
			if nested, ok := val.(map[string]interface{}); ok {
				delete(nested, "time")
			}
		}
	}

	// Sort by deviceId for stable output
	sort.Slice(envelope.Result, func(i, j int) bool {
		a, _ := envelope.Result[i]["deviceId"].(string)
		b, _ := envelope.Result[j]["deviceId"].(string)
		return a < b
	})

	return json.Marshal(map[string]interface{}{"result": envelope.Result})
}
