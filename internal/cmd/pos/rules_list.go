package pos

import (
	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/cmdutil"
	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/iostreams"
)

type rulesListOptions struct {
	cmdutil.ListFlags
	DeviceID string
}

func newCmdRulesList(f *factory.Factory) *cobra.Command {
	opts := &rulesListOptions{}

	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List devices with POS custom rules",
		Long:    "List devices that have POS custom rules configured, across all your devices.",
		Example: `  # All devices with custom rules
  incloud pos rules list

  # Filter by device
  incloud pos rules list --device 507f1f77bcf86cd799439011`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := f.APIClient()
			if err != nil {
				return err
			}

			q := cmdutil.NewQuery(cmd, nil)
			if opts.DeviceID != "" {
				q.Set("deviceId", opts.DeviceID)
			}

			body, err := client.Get("/api/v1/network/pos/custom-rules", q)
			if err != nil {
				return err
			}

			output, _ := cmd.Flags().GetString("output")
			return iostreams.FormatOutput(body, f.IO, output)
		},
	}

	cmd.Flags().StringVar(&opts.DeviceID, "device", "", "Filter by device ID")
	opts.Register(cmd)

	return cmd
}
