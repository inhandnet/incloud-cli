package device

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net/http"
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

			cfg, err := f.Config()
			if err != nil {
				return err
			}
			ctx, err := cfg.ActiveContext()
			if err != nil {
				return err
			}

			client, err := f.HttpClient()
			if err != nil {
				return err
			}

			reqURL := ctx.Host + "/api/v1/devices/" + id
			req, err := http.NewRequestWithContext(context.Background(), http.MethodDelete, reqURL, http.NoBody)
			if err != nil {
				return err
			}

			resp, err := client.Do(req)
			if err != nil {
				return fmt.Errorf("request failed: %w", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode == http.StatusNoContent {
				fmt.Fprintf(f.IO.ErrOut, "Device %s deleted.\n", id)
				return nil
			}

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				return fmt.Errorf("reading response: %w", err)
			}
			fmt.Fprintln(f.IO.ErrOut, string(body))
			return fmt.Errorf("HTTP %d", resp.StatusCode)
		},
	}

	cmd.Flags().BoolVarP(&yes, "yes", "y", false, "Skip confirmation prompt")

	return cmd
}
