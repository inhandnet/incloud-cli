package device

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/iostreams"
)

type uplinkGetOptions struct {
	Fields []string
}

func newCmdUplinkGet(f *factory.Factory) *cobra.Command {
	opts := &uplinkGetOptions{}

	cmd := &cobra.Command{
		Use:   "get <uplink-id>",
		Short: "Get uplink details",
		Long:  "Get detailed information for a specific uplink by its ID.",
		Example: `  # Get uplink details
  incloud device uplink get 69b27e3e6e65fb572c20fab4

  # Table output
  incloud device uplink get 69b27e3e6e65fb572c20fab4 -o table`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			uplinkID := args[0]

			cfg, err := f.Config()
			if err != nil {
				return err
			}
			actx, err := cfg.ActiveContext()
			if err != nil {
				return err
			}

			client, err := f.HttpClient()
			if err != nil {
				return err
			}

			reqURL := actx.Host + "/api/v1/uplinks/" + uplinkID
			req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, reqURL, http.NoBody)
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
			fields := opts.Fields
			if len(fields) == 0 {
				fields = defaultUplinkDetailFields
			}
			return iostreams.FormatOutput(body, f.IO, output, fields)
		},
	}

	cmd.Flags().StringSliceVarP(&opts.Fields, "fields", "f", nil, "Fields to display in table mode")

	return cmd
}
