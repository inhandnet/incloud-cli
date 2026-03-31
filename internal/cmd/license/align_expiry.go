package license

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/iostreams"
	"github.com/inhandnet/incloud-cli/internal/ui"
)

func NewCmdAlignExpiry(f *factory.Factory) *cobra.Command {
	var yes bool

	cmd := &cobra.Command{
		Use:   "align-expiry <license-id> [<license-id>...]",
		Short: "Align expiry dates of multiple licenses",
		Long: `Align (co-terminate) multiple licenses to the same expiry date.

This is a free operation that redistributes remaining license time so all
selected licenses expire on the same date. Only activated or to-be-expired
licenses can be aligned.`,
		Example: `  # Align expiry dates for multiple licenses
  incloud license align-expiry LIC_ID1 LIC_ID2 LIC_ID3

  # Skip confirmation
  incloud license align-expiry LIC_ID1 LIC_ID2 --yes`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 1000 {
				return fmt.Errorf("maximum 1000 licenses per align-expiry operation, got %d", len(args))
			}

			client, err := f.APIClient()
			if err != nil {
				return err
			}

			body := map[string]any{
				"licenses": args,
			}

			previewResp, err := client.Post("/api/v1/billing/coterm", map[string]any{
				"licenses": args,
				"expand":   "type",
			})
			if err != nil {
				return err
			}

			var wrapper struct {
				Result []json.RawMessage `json:"result"`
			}
			if err := json.Unmarshal(previewResp, &wrapper); err == nil && len(wrapper.Result) == 0 {
				return fmt.Errorf("no eligible licenses to align (only activated or to-be-expired licenses can be aligned)")
			}

			output, _ := cmd.Flags().GetString("output")
			if err := iostreams.FormatOutput(previewResp, f.IO, output); err != nil {
				return err
			}

			if !yes {
				confirmed, err := ui.Confirm(f, fmt.Sprintf("Align expiry dates for %d license(s)?", len(args)))
				if err != nil {
					return err
				}
				if !confirmed {
					return nil
				}
			}

			_, err = client.Post("/api/v1/billing/coterm/apply", body)
			if err != nil {
				return err
			}

			fmt.Fprintf(f.IO.ErrOut, "Aligned expiry dates for %d license(s).\n", len(args))
			return nil
		},
	}

	cmd.Flags().BoolVarP(&yes, "yes", "y", false, "Skip confirmation prompt")

	return cmd
}
