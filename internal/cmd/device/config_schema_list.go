package device

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tidwall/gjson"

	"github.com/inhandnet/incloud-cli/internal/api"
	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/iostreams"
)

// schemaFlags holds the shared flags for all schema commands.
type schemaFlags struct {
	device  string
	product string
	version string
}

// register adds --device, --product, --version flags to the command.
func (sf *schemaFlags) register(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&sf.device, "device", "d", "", "Device ID (use 'incloud device list' to find IDs)")
	cmd.Flags().StringVarP(&sf.product, "product", "p", "", "Product code (requires --version; mutually exclusive with --device)")
	cmd.Flags().StringVar(&sf.version, "version", "", "Firmware version (requires --product)")
}

// resolve calls resolveProductVersion with the stored flag values.
func (sf *schemaFlags) resolve(client *api.APIClient) (*productVersion, error) {
	return resolveProductVersion(client, sf.device, sf.product, sf.version)
}

var defaultSchemaListFields = []string{"name", "jsonKeys", "description"}

func newCmdSchemaList(f *factory.Factory) *cobra.Command {
	sf := &schemaFlags{}
	var name string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List configuration schemas",
		Long: `List configuration schema documents for a device's product/firmware.
Shows schema name, JSON keys, and a brief description for each config section.

Use --device to auto-detect product/version, or --product/--version to specify directly.`,
		Aliases: []string{"ls"},
		Example: `  # List schemas for a device
  incloud device config schema list --device 507f1f77bcf86cd799439011

  # List by product/version
  incloud device config schema list --product MR805 --version V2.0.15-111

  # Filter by name (regex)
  incloud device config schema list --device 507f1f77bcf86cd799439011 --name "DNS|WiFi"

  # JSON output for AI tools
  incloud device config schema list --device 507f1f77bcf86cd799439011 -o json`,
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
			q.Set("limit", "200")
			if name != "" {
				q.Set("name", name)
			}

			body, err := client.Get("/api/v1/config-documents", q)
			if err != nil {
				return err
			}

			output, _ := cmd.Flags().GetString("output")
			if output == "" {
				output = "table"
			}

			return iostreams.FormatOutput(body, f.IO, output, defaultSchemaListFields,
				iostreams.WithTransform(transformSchemaList),
			)
		},
	}

	sf.register(cmd)
	cmd.Flags().StringVar(&name, "name", "", "Filter by schema name (regex)")

	return cmd
}

// transformSchemaList transforms config-documents API response for table display.
// Converts jsonKeys array to comma-separated string, truncates descriptions.
func transformSchemaList(body []byte) ([]byte, error) {
	result := gjson.GetBytes(body, "result")
	if !result.Exists() {
		return []byte(`{"result":[]}`), nil
	}

	var items []map[string]interface{}
	for _, item := range result.Array() {
		row := map[string]interface{}{
			"name": item.Get("name").String(),
		}

		// Join jsonKeys array into comma-separated string
		var keys []string
		for _, k := range item.Get("jsonKeys").Array() {
			keys = append(keys, k.String())
		}
		row["jsonKeys"] = strings.Join(keys, ", ")

		// Use first description, truncate for table (rune-safe for CJK)
		descs := item.Get("descriptions").Array()
		if len(descs) > 0 {
			desc := descs[0].String()
			runes := []rune(desc)
			if len(runes) > 80 {
				desc = string(runes[:77]) + "..."
			}
			row["description"] = desc
		} else {
			row["description"] = ""
		}

		items = append(items, row)
	}

	out, err := json.Marshal(map[string]interface{}{"result": items})
	if err != nil {
		return nil, fmt.Errorf("formatting schema list: %w", err)
	}
	return out, nil
}
