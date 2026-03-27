package device

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/iostreams"
)

type antennaOptions struct {
	After  string
	Before string
	Fields []string
}

func NewCmdAntenna(f *factory.Factory) *cobra.Command {
	opts := &antennaOptions{}

	cmd := &cobra.Command{
		Use:   "antenna <device-id>",
		Short: "Antenna signal data",
		Long:  "Display per-antenna signal metrics (RSRP, RSRQ, SINR, ssRsrp, ssRsrq, ssSinr) with GPS correlation.",
		Example: `  # Show antenna signal data for a device
  incloud device antenna 507f1f77bcf86cd799439011 --after 2024-01-01T00:00:00Z --before 2024-01-02T00:00:00Z

  # Table output with selected fields
  incloud device antenna 507f1f77bcf86cd799439011 --after 2024-01-01T00:00:00Z --before 2024-01-02T00:00:00Z -o table -f time -f antenna -f rsrp`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			deviceID := args[0]

			client, err := f.APIClient()
			if err != nil {
				return err
			}

			q := url.Values{}
			q.Set("after", opts.After)
			q.Set("before", opts.Before)

			body, err := client.Get("/api/v1/devices/"+deviceID+"/antenna-signal", q)
			if err != nil {
				return err
			}

			output, _ := cmd.Flags().GetString("output")
			return iostreams.FormatOutput(body, f.IO, output, iostreams.WithTransform(flattenAntennaSeries))
		},
	}

	cmd.Flags().StringVar(&opts.After, "after", "", "Start time (ISO 8601, e.g. 2024-01-01T00:00:00Z) [required]")
	cmd.Flags().StringVar(&opts.Before, "before", "", "End time (ISO 8601, e.g. 2024-01-02T00:00:00Z) [required]")
	cmd.Flags().StringSliceVarP(&opts.Fields, "fields", "f", nil, "Fields to display in table mode")
	_ = cmd.MarkFlagRequired("after")
	_ = cmd.MarkFlagRequired("before")

	return cmd
}

// flattenAntennaSeries converts antenna-signal series into flat rows,
// injecting "antenna" (antennaIndex) from each series into every row.
func flattenAntennaSeries(body []byte) ([]byte, error) {
	var envelope struct {
		Result struct {
			Series []struct {
				Antenna string          `json:"antenna"`
				Fields  []string        `json:"fields"`
				Data    [][]interface{} `json:"data"`
			} `json:"series"`
		} `json:"result"`
	}
	if err := json.Unmarshal(body, &envelope); err != nil {
		return nil, fmt.Errorf("parsing antenna-signal response: %w", err)
	}

	var rows []map[string]interface{}
	for _, s := range envelope.Result.Series {
		for _, row := range s.Data {
			obj := map[string]interface{}{"antenna": s.Antenna}
			for i, field := range s.Fields {
				if i < len(row) {
					obj[field] = row[i]
				}
			}
			rows = append(rows, obj)
		}
	}

	if len(rows) == 0 {
		return json.Marshal(map[string]interface{}{"result": []interface{}{}})
	}
	return json.Marshal(map[string]interface{}{"result": rows})
}
