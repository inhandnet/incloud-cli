package pos

import (
	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/cmdutil"
	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/iostreams"
)

type forwardedOptions struct {
	cmdutil.ListFlags
	ActiveWithin string
	ClientType   string
	Vendor       string
	DeviceID     string
	OID          string
}

func newCmdForwarded(f *factory.Factory) *cobra.Command {
	opts := &forwardedOptions{}

	cmd := &cobra.Command{
		Use:   "forwarded",
		Short: "List clients with recently forwarded POS traffic",
		Long:  "List clients whose POS traffic was matched/forwarded within a recent time window, with the matched vendor tags.",
		Example: `  # Clients with POS hits in the last 24h
  incloud pos forwarded

  # Last 7 days, filter by vendor
  incloud pos forwarded --active-within 7d --vendor Verifone`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := f.APIClient()
			if err != nil {
				return err
			}

			q := cmdutil.NewQuery(cmd, nil)
			if opts.ActiveWithin != "" {
				q.Set("activeWithin", opts.ActiveWithin)
			}
			if opts.ClientType != "" {
				q.Set("clientType", opts.ClientType)
			}
			if opts.Vendor != "" {
				q.Set("vendor", opts.Vendor)
			}
			if opts.DeviceID != "" {
				q.Set("deviceId", opts.DeviceID)
			}
			if opts.OID != "" {
				q.Set("oid", opts.OID)
			}

			body, err := client.Get("/api/v1/pos-ready/forwarded", q)
			if err != nil {
				return err
			}

			output, _ := cmd.Flags().GetString("output")
			return iostreams.FormatOutput(body, f.IO, output)
		},
	}

	cmd.Flags().StringVar(&opts.ActiveWithin, "active-within", "", "Time window: 1h, 6h, 24h (default), 7d")
	cmd.Flags().StringVar(&opts.ClientType, "client-type", "", "Filter by client type (e.g. POS_TERMINAL)")
	cmd.Flags().StringVar(&opts.Vendor, "vendor", "", "Filter by matched vendor")
	cmd.Flags().StringVar(&opts.DeviceID, "device", "", "Filter by device ID")
	cmd.Flags().StringVar(&opts.OID, "oid", "", "Filter by organization ID")
	opts.Register(cmd)

	return cmd
}
