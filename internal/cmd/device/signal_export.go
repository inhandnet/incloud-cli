package device

import (
	"fmt"
	"net/url"
	"os"

	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
)

type signalExportOptions struct {
	After  string
	Before string
	File   string
}

func newCmdSignalExport(f *factory.Factory) *cobra.Command {
	opts := &signalExportOptions{}

	cmd := &cobra.Command{
		Use:   "export <device-id>",
		Short: "Export signal data to Excel",
		Long:  "Export signal quality data as a server-generated Excel file.",
		Example: `  # Export signal data to file
  incloud device signal export 507f1f77bcf86cd799439011 --file signal.xlsx

  # Export with time range
  incloud device signal export 507f1f77bcf86cd799439011 --after 2024-01-01T00:00:00Z --before 2024-01-07T00:00:00Z --file signal.xlsx

  # Export to stdout (pipe to file)
  incloud device signal export 507f1f77bcf86cd799439011 > signal.xlsx`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			deviceID := args[0]

			client, err := f.APIClient()
			if err != nil {
				return err
			}

			q := url.Values{}
			if opts.After != "" {
				q.Set("after", opts.After)
			}
			if opts.Before != "" {
				q.Set("before", opts.Before)
			}

			body, err := client.Get("/api/v1/devices/"+deviceID+"/signal/export", q)
			if err != nil {
				return err
			}

			if opts.File != "" {
				if err := os.WriteFile(opts.File, body, 0o600); err != nil {
					return fmt.Errorf("writing file: %w", err)
				}
				fmt.Fprintf(f.IO.Out, "Exported to %s (%d bytes)\n", opts.File, len(body))
				return nil
			}

			_, err = f.IO.Out.Write(body)
			return err
		},
	}

	cmd.Flags().StringVar(&opts.After, "after", "", "Start time (ISO 8601, e.g. 2024-01-01T00:00:00Z)")
	cmd.Flags().StringVar(&opts.Before, "before", "", "End time (ISO 8601, e.g. 2024-01-02T00:00:00Z)")
	cmd.Flags().StringVar(&opts.File, "file", "", "Write output to file instead of stdout")

	return cmd
}
