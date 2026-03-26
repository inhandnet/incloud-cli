package tunnel

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
)

func NewCmdTunnelClose(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "close <tunnel-id>",
		Short: "Close a tunnel",
		Long:  "Close an active remote access tunnel and release its resources.",
		Example: `  # Close a tunnel
  incloud tunnel close abc123def456`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			tunnelID := args[0]

			client, err := f.APIClient()
			if err != nil {
				return err
			}

			endpoint := fmt.Sprintf("/api/v1/ngrok/tunnels/%s", tunnelID)
			_, err = client.Delete(endpoint)
			if err != nil {
				return err
			}

			fmt.Fprintf(f.IO.ErrOut, "Tunnel %s closed.\n", tunnelID)
			return nil
		},
	}

	return cmd
}
