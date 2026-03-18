package device

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/iostreams"
)

func newCmdConfigError(f *factory.Factory) *cobra.Command {
	var module string

	cmd := &cobra.Command{
		Use:   "error <device-id>",
		Short: "Get configuration delivery errors",
		Long:  "Get the most recent configuration delivery errors and pending changes for a device.",
		Example: `  # Check for config errors
  incloud device config error 507f1f77bcf86cd799439011

  # JSON output with full pending details
  incloud device config error 507f1f77bcf86cd799439011 -o json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			deviceID := args[0]

			client, err := f.APIClient()
			if err != nil {
				return err
			}

			q := url.Values{}
			if module != "" {
				q.Set("module", module)
			}

			body, err := client.Get("/api/v1/devices/"+deviceID+"/config/error", q)
			if err != nil {
				return err
			}

			output, _ := cmd.Flags().GetString("output")

			// For table mode, extract and display only the error list
			if output == "" || output == "table" {
				return formatConfigErrors(body, f)
			}

			return iostreams.FormatOutput(body, f.IO, output, nil,
				iostreams.WithTransform(extractResultArray),
			)
		},
	}

	cmd.Flags().StringVar(&module, "module", "", "Module name (defaults to 'default' on the server)")

	return cmd
}

func formatConfigErrors(body []byte, f *factory.Factory) error {
	var resp struct {
		Result struct {
			Error []struct {
				Field   string `json:"field"`
				Message string `json:"message"`
			} `json:"error"`
		} `json:"result"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return fmt.Errorf("parsing response: %w", err)
	}

	if len(resp.Result.Error) == 0 {
		fmt.Fprintln(f.IO.ErrOut, "No configuration errors.")
		return nil
	}

	// Convert error list to JSON array for table rendering
	errJSON, err := json.Marshal(resp.Result.Error)
	if err != nil {
		return err
	}
	return iostreams.FormatTable(errJSON, f.IO, nil)
}
