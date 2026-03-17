package alert

import (
	"fmt"

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

			client, err := f.APIClient()
			if err != nil {
				return err
			}

			var path string
			var bodyMap map[string]any

			if all {
				path = "/api/v1/alerts/acknowledge/all"
				bodyMap = make(map[string]any)
			} else {
				path = "/api/v1/alerts/acknowledge"
				bodyMap = map[string]any{
					"ids": args,
				}
			}

			if len(typeFlag) > 0 {
				bodyMap["type"] = typeFlag
			}

			if _, err := client.Put(path, bodyMap); err != nil {
				return err
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
