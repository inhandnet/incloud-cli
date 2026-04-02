package license

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/iostreams"
	"github.com/inhandnet/incloud-cli/internal/ui"
)

func NewCmdTransfer(f *factory.Factory) *cobra.Command {
	var (
		org string
		yes bool
	)

	cmd := &cobra.Command{
		Use:   "transfer <license-id> [<license-id>...]",
		Short: "Transfer licenses to another organization",
		Long: `Transfer one or more licenses to another organization.

Only inactivated and unattached licenses can be transferred.
Maximum 1000 licenses per operation.`,
		Example: `  # Transfer licenses to another organization
  incloud license transfer LIC_ID1 LIC_ID2 LIC_ID3 --org ORG_ID

  # Skip confirmation
  incloud license transfer LIC_ID1 --org ORG_ID --yes`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 1000 {
				return fmt.Errorf("maximum 1000 licenses per transfer operation, got %d", len(args))
			}

			client, err := f.APIClient()
			if err != nil {
				return err
			}

			// Fetch each license for preview
			var previews []any
			for _, id := range args {
				resp, err := client.Get("/api/v1/billing/licenses/"+id, nil)
				if err != nil {
					return fmt.Errorf("failed to fetch license %s: %w", id, err)
				}
				var parsed struct {
					Result json.RawMessage `json:"result"`
				}
				if err := json.Unmarshal(resp, &parsed); err == nil && parsed.Result != nil {
					var item any
					_ = json.Unmarshal(parsed.Result, &item)
					previews = append(previews, item)
				}
			}

			previewJSON, _ := json.Marshal(previews)
			output, _ := cmd.Flags().GetString("output")
			if err := iostreams.FormatOutput(previewJSON, f.IO, output); err != nil {
				return err
			}

			if !yes {
				confirmed, err := ui.Confirm(f, fmt.Sprintf("Transfer %d license(s) to organization %s?", len(args), org))
				if err != nil {
					return err
				}
				if !confirmed {
					return nil
				}
			}

			body := map[string]any{
				"ids": args,
				"to":  org,
			}

			_, err = client.Post("/api/v1/billing/licenses/move", body)
			if err != nil {
				return err
			}

			fmt.Fprintf(f.IO.ErrOut, "Transferred %d license(s) to organization %s.\n", len(args), org)
			return nil
		},
	}

	cmd.Flags().StringVar(&org, "org", "", "Target organization ID (required)")
	_ = cmd.MarkFlagRequired("org")
	cmd.Flags().BoolVarP(&yes, "yes", "y", false, "Skip confirmation prompt")

	return cmd
}
