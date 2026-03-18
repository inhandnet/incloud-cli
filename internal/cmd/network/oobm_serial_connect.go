package network

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/iostreams"
)

func NewCmdOobmSerialConnect(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "connect <id>",
		Short: "Connect an OOBM serial port",
		Long: `Connect an OOBM serial port to establish a remote console tunnel.

On success, prints the connection URL and credentials. For CLI usage,
the URL is an SSH command you can run directly.`,
		Example: `  # Connect a serial port
  incloud network oobm serial connect 507f1f77bcf86cd799439011`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id := args[0]

			client, err := f.APIClient()
			if err != nil {
				return err
			}

			respBody, err := client.Post("/api/v1/oobm/serials/"+id+"/connect", nil)
			if err != nil {
				return err
			}

			var resp struct {
				Result struct {
					URL      string `json:"url"`
					Username string `json:"username"`
					Password string `json:"password"`
					Usage    string `json:"usage"`
				} `json:"result"`
			}
			_ = json.Unmarshal(respBody, &resp)

			fmt.Fprintf(f.IO.ErrOut, "Connected: %s\n", resp.Result.URL)
			if resp.Result.Username != "" {
				fmt.Fprintf(f.IO.ErrOut, "Username: %s\nPassword: %s\n", resp.Result.Username, resp.Result.Password)
			}

			output, _ := cmd.Flags().GetString("output")
			return iostreams.FormatOutput(respBody, f.IO, output, nil)
		},
	}

	return cmd
}
