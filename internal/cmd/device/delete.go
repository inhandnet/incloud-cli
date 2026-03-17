package device

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/mattn/go-isatty"
	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
)

func NewCmdDelete(f *factory.Factory) *cobra.Command {
	var yes bool

	cmd := &cobra.Command{
		Use:     "delete <id>",
		Aliases: []string{"rm"},
		Short:   "Delete a device",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id := args[0]

			if !yes {
				// Check if stdin is a TTY
				if file, ok := f.IO.In.(*os.File); !ok || !isatty.IsTerminal(file.Fd()) {
					return fmt.Errorf("terminal is non-interactive; use --yes to confirm")
				}

				fmt.Fprintf(f.IO.ErrOut, "Delete device %s? (y/N) ", id)
				reader := bufio.NewReader(f.IO.In)
				answer, err := reader.ReadString('\n')
				if err != nil && err != io.EOF {
					return err
				}
				answer = strings.TrimSpace(answer)
				if answer != "y" && answer != "Y" {
					fmt.Fprintln(f.IO.ErrOut, "Aborted.")
					return nil
				}
			}

			client, err := f.APIClient()
			if err != nil {
				return err
			}

			_, err = client.Delete("/api/v1/devices/" + id)
			if err != nil {
				return err
			}

			fmt.Fprintf(f.IO.ErrOut, "Device %s deleted.\n", id)
			return nil
		},
	}

	cmd.Flags().BoolVarP(&yes, "yes", "y", false, "Skip confirmation prompt")

	return cmd
}
