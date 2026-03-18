package connector

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/iostreams"
)

type accountCreateOptions struct {
	Name     string
	StaticIp bool
	Vip      string
}

func newCmdAccountCreate(f *factory.Factory) *cobra.Command {
	opts := &accountCreateOptions{}

	cmd := &cobra.Command{
		Use:   "create <network-id>",
		Short: "Create an account in a connector network",
		Example: `  # Create account
  incloud connector account create 66827b3ccfb1842140f4222f --name user1

  # With static IP
  incloud connector account create 66827b3ccfb1842140f4222f --name user1 --static-ip --vip 10.32.1.100`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			networkID := args[0]

			client, err := f.APIClient()
			if err != nil {
				return err
			}

			if opts.StaticIp && opts.Vip == "" {
				return fmt.Errorf("--vip is required when --static-ip is set")
			}

			body := map[string]interface{}{
				"name":     opts.Name,
				"staticIp": opts.StaticIp,
			}
			if opts.Vip != "" {
				body["vip"] = opts.Vip
			}

			respBody, err := client.Post("/api/v1/connectors/"+networkID+"/accounts", body)
			if err != nil {
				output, _ := cmd.Flags().GetString("output")
				if respBody != nil {
					_ = iostreams.FormatOutput(respBody, f.IO, output, nil)
				}
				return err
			}

			writeCreated(f, "Connector account", respBody)
			return formatOutput(cmd, f.IO, respBody, nil)
		},
	}

	cmd.Flags().StringVar(&opts.Name, "name", "", "Account name (required, max 64 chars)")
	cmd.Flags().BoolVar(&opts.StaticIp, "static-ip", false, "Use static IP address")
	cmd.Flags().StringVar(&opts.Vip, "vip", "", "Virtual IP (required when --static-ip is set)")

	_ = cmd.MarkFlagRequired("name")

	return cmd
}
