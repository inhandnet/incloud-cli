package device

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tidwall/gjson"

	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/iostreams"
)

func newCmdSchemaGet(f *factory.Factory) *cobra.Command {
	sf := &schemaFlags{}

	cmd := &cobra.Command{
		Use:   "get <json-key>",
		Short: "Get a configuration schema by JSON key",
		Long: `Get the full configuration schema definition for a specific JSON key,
including JSON Schema content and human-readable descriptions.

The JSON key can be found from 'incloud device config schema list'.`,
		Example: `  # Get DNS config schema
  incloud device config schema get --device 507f1f77bcf86cd799439011 dns

  # JSON output for AI parsing
  incloud device config schema get --product MR805 --version V2.0.15-111 dns -o json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			jsonKey := args[0]

			client, err := f.APIClient()
			if err != nil {
				return err
			}

			pv, err := sf.resolve(client)
			if err != nil {
				return err
			}

			q := pv.configDocumentQuery()
			q.Set("jsonKeys", jsonKey)

			body, err := client.Get("/api/v1/config-documents", q)
			if err != nil {
				return err
			}

			result := gjson.GetBytes(body, "result")
			if !result.Exists() || len(result.Array()) == 0 {
				return fmt.Errorf("config schema %q not found for %s/%s", jsonKey, pv.product, pv.version)
			}

			// Extract the first matching document
			doc := []byte(result.Array()[0].Raw)

			output, _ := cmd.Flags().GetString("output")
			return iostreams.FormatOutput(doc, f.IO, output, nil)
		},
	}

	sf.register(cmd)

	return cmd
}
