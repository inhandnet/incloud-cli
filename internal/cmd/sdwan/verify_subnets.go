package sdwan

import (
	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
)

func newCmdVerifySubnets(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "verify-subnets <subnet> [subnet...]",
		Short: "Verify subnets for conflicts",
		Example: `  # Check if subnets conflict with each other
  incloud sdwan verify-subnets 10.0.0.0/24 10.0.0.0/16 192.168.1.0/24`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := f.APIClient()
			if err != nil {
				return err
			}

			body := map[string]interface{}{
				"subnets": args,
			}

			respBody, err := client.Post(apiBase+"/devices/subnets/verify", body)
			if err != nil {
				return err
			}

			return formatOutput(cmd, f.IO, respBody)
		},
	}

	return cmd
}
