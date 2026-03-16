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

func NewCmdRuleDelete(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <rule-id> [<rule-id>...]",
		Short: "Delete alert rules",
		Long:  "Delete one or more alert rules by ID. Multiple IDs triggers bulk delete.",
		Example: `  # Delete a single rule
  incloud alert rule delete 507f1f77bcf86cd799439011

  # Delete multiple rules
  incloud alert rule delete 507f1f77bcf86cd799439011 507f1f77bcf86cd799439012`,
		Aliases: []string{"rm"},
		Args:    cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
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

			if len(args) == 1 {
				reqURL := ctx.Host + "/api/v1/alerts/rules/" + args[0]
				req, err := http.NewRequestWithContext(context.Background(), http.MethodDelete, reqURL, http.NoBody)
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

				fmt.Fprintf(f.IO.ErrOut, "Deleted alert rule %s.\n", args[0])
			} else {
				bodyBytes, err := json.Marshal(map[string]any{
					"ids": args,
				})
				if err != nil {
					return fmt.Errorf("marshaling request body: %w", err)
				}

				reqURL := ctx.Host + "/api/v1/alerts/rules/bulk-delete"
				req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, reqURL, bytes.NewReader(bodyBytes))
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

				fmt.Fprintf(f.IO.ErrOut, "Deleted %d alert rule(s).\n", len(args))
			}

			return nil
		},
	}

	return cmd
}
