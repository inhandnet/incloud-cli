package connector

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/iostreams"
)

type accountUpdateOptions struct {
	Name     string
	StaticIp bool
	Vip      string
}

func newCmdAccountUpdate(f *factory.Factory) *cobra.Command {
	opts := &accountUpdateOptions{}

	cmd := &cobra.Command{
		Use:   "update <network-id> <account-id>",
		Short: "Update an account in a connector network",
		Example: `  # Update name
  incloud connector account update <network-id> <account-id> --name new-name

  # Enable static IP
  incloud connector account update <network-id> <account-id> --static-ip --vip 10.32.1.100`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			networkID, accountID := args[0], args[1]

			client, err := f.APIClient()
			if err != nil {
				return err
			}

			body := make(map[string]interface{})
			if cmd.Flags().Changed("name") {
				body["name"] = opts.Name
			}
			if cmd.Flags().Changed("static-ip") {
				body["staticIp"] = opts.StaticIp
			}
			if cmd.Flags().Changed("vip") {
				body["vip"] = opts.Vip
			}

			if len(body) == 0 {
				return fmt.Errorf("no fields to update; specify at least one of --name, --static-ip, --vip")
			}

			respBody, err := client.Put("/api/v1/connectors/"+networkID+"/accounts/"+accountID, body)
			if err != nil {
				output, _ := cmd.Flags().GetString("output")
				if respBody != nil {
					_ = iostreams.FormatOutput(respBody, f.IO, output, nil)
				}
				return err
			}

			writeUpdated(f, "Connector account", respBody)
			return formatOutput(cmd, f.IO, respBody, nil)
		},
	}

	cmd.Flags().StringVar(&opts.Name, "name", "", "Account name (max 64 chars)")
	cmd.Flags().BoolVar(&opts.StaticIp, "static-ip", false, "Use static IP address")
	cmd.Flags().StringVar(&opts.Vip, "vip", "", "Virtual IP (required when --static-ip is set)")

	return cmd
}
