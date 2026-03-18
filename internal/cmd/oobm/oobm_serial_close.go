package oobm

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/iostreams"
)

func NewCmdOobmSerialClose(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "close <id>",
		Short: "Close an OOBM serial port connection",
		Example: `  # Close a serial port connection
  incloud oobm serial close 507f1f77bcf86cd799439011`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id := args[0]

			client, err := f.APIClient()
			if err != nil {
				return err
			}

			respBody, err := client.Post("/api/v1/oobm/serials/"+id+"/close", nil)
			if err != nil {
				return err
			}

			fmt.Fprintf(f.IO.ErrOut, "OOBM serial connection (%s) closed.\n", id)

			output, _ := cmd.Flags().GetString("output")
			return iostreams.FormatOutput(respBody, f.IO, output, nil)
		},
	}

	return cmd
}
