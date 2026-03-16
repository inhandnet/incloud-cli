package alert

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
)

func NewCmdAck(f *factory.Factory) *cobra.Command {
	var (
		all      bool
		typeFlag []string
	)

	cmd := &cobra.Command{
		Use:   "ack [<id>...]",
		Short: "Acknowledge alerts",
		Long:  "Acknowledge one or more alerts by ID, or acknowledge all alerts with --all.",
		Example: `  # Acknowledge specific alerts
  incloud alert ack 507f1f77bcf86cd799439011 507f1f77bcf86cd799439012

  # Acknowledge all alerts
  incloud alert ack --all

  # Acknowledge all alerts of specific types
  incloud alert ack --all --type offline --type reboot`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if all && len(args) > 0 {
				return fmt.Errorf("cannot specify both --all and alert IDs")
			}
			if !all && len(args) == 0 {
				return fmt.Errorf("must specify alert IDs or --all")
			}

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

			var reqURL string
			var bodyMap map[string]any

			if all {
				reqURL = ctx.Host + "/api/v1/alerts/acknowledge/all"
				bodyMap = make(map[string]any)
			} else {
				reqURL = ctx.Host + "/api/v1/alerts/acknowledge"
				bodyMap = map[string]any{
					"ids": args,
				}
			}

			if len(typeFlag) > 0 {
				bodyMap["type"] = typeFlag
			}

			bodyBytes, err := json.Marshal(bodyMap)
			if err != nil {
				return fmt.Errorf("marshaling request body: %w", err)
			}

			req, err := http.NewRequestWithContext(context.Background(), http.MethodPut, reqURL, bytes.NewReader(bodyBytes))
			if err != nil {
				return fmt.Errorf("building request: %w", err)
			}
			req.Header.Set("Content-Type", "application/json")

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

			if all {
				fmt.Fprintln(f.IO.ErrOut, "Acknowledged all alerts.")
			} else {
				fmt.Fprintf(f.IO.ErrOut, "Acknowledged %d alert(s).\n", len(args))
			}

			return nil
		},
	}

	cmd.Flags().BoolVar(&all, "all", false, "Acknowledge all alerts")
	cmd.Flags().StringArrayVar(&typeFlag, "type", nil, "Filter by alert type (can be specified multiple times)")

	return cmd
}
