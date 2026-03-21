package device

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tidwall/gjson"

	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/iostreams"
)

func newCmdSchemaOverview(f *factory.Factory) *cobra.Command {
	sf := &schemaFlags{}

	cmd := &cobra.Command{
		Use:   "overview",
		Short: "View product configuration overview",
		Long: `View the configuration overview for a product/firmware, including
dependency rules and business constraints between config sections.

AI tools should read this before modifying configurations to understand
which config sections depend on each other.`,
		Example: `  # View overview for a device
  incloud device config schema overview --device 507f1f77bcf86cd799439011

  # View by product/version
  incloud device config schema overview --product CPE02 --version V2.0.8

  # JSON output
  incloud device config schema overview --product CPE02 --version V2.0.8 -o json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := f.APIClient()
			if err != nil {
				return err
			}

			pv, err := sf.resolve(client)
			if err != nil {
				return err
			}

			q := pv.configDocumentQuery()

			body, err := client.Get("/api/v1/config-documents/overview", q)
			if err != nil {
				return err
			}

			result := gjson.GetBytes(body, "result")
			if !result.Exists() || result.Type == gjson.Null {
				fmt.Fprintf(f.IO.ErrOut, "No overview available for %s/%s.\n", pv.product, pv.version)
				return nil
			}

			output, _ := cmd.Flags().GetString("output")
			if output == "" || output == "table" {
				// Output markdown content directly
				content := result.Get("content").String()
				fmt.Fprintln(f.IO.Out, content)
				return nil
			}

			return iostreams.FormatOutput([]byte(result.Raw), f.IO, output, nil)
		},
	}

	sf.register(cmd)

	return cmd
}
