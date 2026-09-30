package device

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/cmdutil"
	"github.com/inhandnet/incloud-cli/internal/factory"
)

type historyListOptions struct {
	cmdutil.ListFlags
	Status string
	After  string
	Before string
}

func newCmdHistoryList(f *factory.Factory) *cobra.Command {
	opts := &historyListOptions{}

	cmd := &cobra.Command{
		Use:   "list <device-id>",
		Short: "List diagnosis records of a device",
		Long: `List the AI assistant's diagnosis records for a device, newest first.

By default only records still awaiting the user's confirmation are listed;
use --status all to include records whose result has been confirmed.
Use 'incloud device history get <record-id>' for the full content of a record.`,
		Example: `  # Records awaiting confirmation
  incloud device history list 507f1f77bcf86cd799439011

  # All records since a date
  incloud device history list 507f1f77bcf86cd799439011 --status all --after 2026-09-01

  # Full fields as JSON
  incloud device history list 507f1f77bcf86cd799439011 -o json`,
		Aliases: []string{"ls"},
		Args:    cmdutil.ObjectIDArgs(cobra.ExactArgs(1), 0, "device id", "incloud device list -q %s"),
		RunE: func(cmd *cobra.Command, args []string) error {
			if opts.Status != "open" && opts.Status != "all" {
				return fmt.Errorf("invalid --status %q: use open or all", opts.Status)
			}

			client, err := f.APIClient()
			if err != nil {
				return err
			}

			q := cmdutil.NewQuery(cmd, nil)
			q.Set("device_id", args[0])
			q.Set("status", opts.Status)
			if opts.After != "" {
				q.Set("after", cmdutil.ParseTimeFlag(opts.After))
			}
			if opts.Before != "" {
				q.Set("before", cmdutil.ParseTimeFlag(opts.Before))
			}

			body, err := client.Get(diagnosisRecordsPath, q)
			if err != nil {
				return err
			}
			body, _, err = rewrapRecords(body)
			if err != nil {
				return err
			}
			return writeHistory(f, cmd, body)
		},
	}

	opts.Register(cmd)
	cmd.Flags().StringVar(&opts.Status, "status", "open", "Record status: open (awaiting confirmation) or all")
	cmd.Flags().StringVar(&opts.After, "after", "", "Start time (e.g. 2026-09-01, 2026-09-01T08:00:00, 2026-09-01T00:00:00Z)")
	cmd.Flags().StringVar(&opts.Before, "before", "", "End time (e.g. 2026-09-30, 2026-09-30T08:00:00, 2026-09-30T23:59:59Z)")

	return cmd
}
