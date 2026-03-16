package device

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/iostreams"
)

type GetOptions struct {
	Fields []string
}

func NewCmdGet(f *factory.Factory) *cobra.Command {
	opts := &GetOptions{}

	cmd := &cobra.Command{
		Use:   "get <device-id>",
		Short: "Get device details",
		Long:  "Get detailed information about a specific device by its ID.",
		Example: `  # Get device details (colorized JSON in TTY)
  incloud device get 507f1f77bcf86cd799439011

  # Only specific fields
  incloud device get 507f1f77bcf86cd799439011 -f name -f serialNumber -f online

  # Table output (KEY/VALUE pairs)
  incloud device get 507f1f77bcf86cd799439011 -o table

  # YAML output
  incloud device get 507f1f77bcf86cd799439011 -o yaml`,
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

			u, err := url.Parse(ctx.Host + "/api/v1/devices/" + deviceID)
			if err != nil {
				return fmt.Errorf("invalid URL: %w", err)
			}
			if len(opts.Fields) > 0 {
				q := u.Query()
				q.Set("fields", strings.Join(opts.Fields, ","))
				u.RawQuery = q.Encode()
			}

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
				if err := iostreams.FormatTable(body, f.IO, opts.Fields); err != nil {
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

	cmd.Flags().StringSliceVarP(&opts.Fields, "fields", "f", nil, "Fields to return and display")

	return cmd
}
