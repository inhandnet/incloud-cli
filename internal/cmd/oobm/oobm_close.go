package oobm

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
)

type OobmCloseOptions struct {
	Services []string
}

func NewCmdOobmClose(f *factory.Factory) *cobra.Command {
	opts := &OobmCloseOptions{}

	cmd := &cobra.Command{
		Use:   "close <id>",
		Short: "Close an OOBM connection",
		Long: `Close an Out-of-Band Management connection to tear down remote access tunnels.

Without --service, all services on the resource are closed.
Use --service to close only specific services (protocol:port[:usage] format).`,
		Example: `  # Close all services on the resource
  incloud oobm close 507f1f77bcf86cd799439011

  # Close only SSH service
  incloud oobm close 507f1f77bcf86cd799439011 --service ssh:22:cli`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id := args[0]

			client, err := f.APIClient()
			if err != nil {
				return err
			}

			var services []map[string]any
			if len(opts.Services) > 0 {
				services, err = parseServices(opts.Services)
				if err != nil {
					return err
				}
			} else {
				res, err := getOobmResource(client, id)
				if err != nil {
					return err
				}
				services = res.Services
				if len(services) == 0 {
					return fmt.Errorf("resource %q has no services defined", id)
				}
			}

			endpoint := fmt.Sprintf("/api/v1/oobm/resources/%s/close", id)
			closed := 0

			for _, svc := range services {
				_, err := client.Post(endpoint, svc)
				if err != nil {
					fmt.Fprintf(f.IO.ErrOut, "Failed to close %s: %v\n", serviceLabel(svc), err)
					continue
				}
				closed++
				fmt.Fprintf(f.IO.ErrOut, "Closed %s tunnel for resource %s.\n", svc["protocol"], id)
			}

			if closed == 0 {
				return fmt.Errorf("no services closed successfully")
			}

			return nil
		},
	}

	cmd.Flags().StringArrayVar(&opts.Services, "service", nil, "Service in protocol:port[:usage] format (omit to close all)")

	return cmd
}
