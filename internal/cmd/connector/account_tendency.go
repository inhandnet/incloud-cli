package connector

import (
	"net/url"

	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/iostreams"
)

type accountTendencyOptions struct {
	After  string
	Before string
}

func newCmdAccountTendency(f *factory.Factory) *cobra.Command {
	opts := &accountTendencyOptions{}

	cmd := &cobra.Command{
		Use:   "tendency <network-id> <account-id>",
		Short: "Show account connection usage trend",
		Example: `  # Show usage trend
  incloud connector account tendency <network-id> <account-id> --after 2025-01-01T00:00:00 --before 2025-01-31T23:59:59`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			networkID, accountID := args[0], args[1]

			client, err := f.APIClient()
			if err != nil {
				return err
			}

			q := url.Values{}
			q.Set("after", opts.After)
			q.Set("before", opts.Before)

			body, err := client.Get("/api/v1/connectors/"+networkID+"/accounts/"+accountID+"/online-tendency", q)
			if err != nil {
				return err
			}

			output, _ := cmd.Flags().GetString("output")
			return iostreams.FormatOutput(body, f.IO, output, nil,
				iostreams.WithTransform(iostreams.FlattenSeries))
		},
	}

	cmd.Flags().StringVar(&opts.After, "after", "", "Start time (ISO 8601, required)")
	cmd.Flags().StringVar(&opts.Before, "before", "", "End time (ISO 8601, required)")
	_ = cmd.MarkFlagRequired("after")
	_ = cmd.MarkFlagRequired("before")

	return cmd
}
