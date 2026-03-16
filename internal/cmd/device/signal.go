package device

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/iostreams"
)

type SignalOptions struct {
	After   string
	Before  string
	Columns []string
}

var defaultSignalColumns = []string{"time", "type", "rsrp", "rsrq", "sinr", "networkType", "carrier", "band"}

func NewCmdSignal(f *factory.Factory) *cobra.Command {
	opts := &SignalOptions{}

	cmd := &cobra.Command{
		Use:   "signal <device-id>",
		Short: "Show device signal quality",
		Long:  "Display signal quality metrics (RSRP, RSRQ, SINR, etc.) for a device over time.",
		Example: `  # Show signal data for a device
  incloud device signal 507f1f77bcf86cd799439011

  # Filter by time range
  incloud device signal 507f1f77bcf86cd799439011 --after 2024-01-01T00:00:00 --before 2024-01-02T00:00:00

  # Table output with selected columns
  incloud device signal 507f1f77bcf86cd799439011 -o table -c time -c rsrp -c sinr`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			deviceID := args[0]

			cfg, err := f.Config()
			if err != nil {
				return err
			}
			ctx, err := cfg.ActiveContext()
			if err != nil {
				return err
			}

			client, err := f.HttpClient()
			if err != nil {
				return err
			}

			u, err := url.Parse(ctx.Host + "/api/v1/devices/" + deviceID + "/signal")
			if err != nil {
				return fmt.Errorf("invalid URL: %w", err)
			}

			q := u.Query()
			if opts.After != "" {
				q.Set("after", opts.After)
			}
			if opts.Before != "" {
				q.Set("before", opts.Before)
			}
			u.RawQuery = q.Encode()

			req, err := http.NewRequestWithContext(context.Background(), "GET", u.String(), http.NoBody)
			if err != nil {
				return fmt.Errorf("building request: %w", err)
			}

			resp, err := client.Do(req)
			if err != nil {
				return fmt.Errorf("request failed: %w", err)
			}
			defer resp.Body.Close()

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				return fmt.Errorf("reading response: %w", err)
			}

			if resp.StatusCode >= 400 {
				return fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
			}

			output, _ := cmd.Flags().GetString("output")
			switch output {
			case "table":
				flat, err := flattenSignalSeries(body)
				if err != nil {
					return err
				}
				columns := opts.Columns
				if len(columns) == 0 && f.IO.IsStdoutTTY() {
					columns = defaultSignalColumns
				}
				if err := iostreams.FormatTable(flat, f.IO, columns); err != nil {
					return err
				}
			case "yaml":
				s, err := iostreams.FormatYAML(body)
				if err != nil {
					return err
				}
				fmt.Fprintln(f.IO.Out, s)
			default:
				if json.Valid(body) {
					fmt.Fprintln(f.IO.Out, iostreams.FormatJSON(body, f.IO, output))
				} else {
					fmt.Fprintln(f.IO.Out, string(body))
				}
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&opts.After, "after", "", "Start time (ISO 8601, e.g. 2024-01-01T00:00:00)")
	cmd.Flags().StringVar(&opts.Before, "before", "", "End time (ISO 8601, e.g. 2024-01-02T00:00:00)")
	cmd.Flags().StringArrayVarP(&opts.Columns, "column", "c", nil, "Columns to show in table output")

	return cmd
}

// flattenSignalSeries converts the signal API's series format (fields + data matrix)
// into a flat JSON array of objects suitable for FormatTable.
func flattenSignalSeries(body []byte) ([]byte, error) {
	var envelope struct {
		Result struct {
			Series []struct {
				Type   string          `json:"type"`
				Fields []string        `json:"fields"`
				Data   [][]interface{} `json:"data"`
			} `json:"series"`
		} `json:"result"`
	}
	if err := json.Unmarshal(body, &envelope); err != nil {
		return nil, fmt.Errorf("parsing signal response: %w", err)
	}

	var rows []map[string]interface{}
	for _, s := range envelope.Result.Series {
		for _, row := range s.Data {
			obj := map[string]interface{}{
				"type": s.Type,
			}
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
