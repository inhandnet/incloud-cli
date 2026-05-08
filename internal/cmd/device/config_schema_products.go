package device

import (
	"encoding/json"
	"fmt"
	"sort"

	"github.com/spf13/cobra"
	"github.com/tidwall/gjson"

	"github.com/inhandnet/incloud-cli/internal/cmdutil"
	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/iostreams"
)

func newCmdSchemaProducts(f *factory.Factory) *cobra.Command {
	var (
		product string
		version string
		lf      cmdutil.ListFlags
	)

	cmd := &cobra.Command{
		Use:   "products",
		Short: "List products and versions with configuration schemas",
		Long: `List all product/version combinations that have configuration schemas.

Use this to discover which devices support AI-assisted configuration
before using 'incloud device config schema overview/get/validate'.`,
		Example: `  # List all products with config schemas
  incloud device config schema products

  # Filter by product
  incloud device config schema products --product MR805

  # JSON output for AI parsing
  incloud device config schema products -o json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := f.APIClient()
			if err != nil {
				return err
			}

			q := cmdutil.NewQuery(cmd, []string{"product", "version"})
			q.Set("module", defaultConfigModule)
			if product != "" {
				q.Set("product", product)
			}
			if version != "" {
				q.Set("version", version)
			}

			body, err := client.Get("/api/v1/config-documents/overviews", q)
			if err != nil {
				return err
			}

			output, _ := cmd.Flags().GetString("output")
			if output == "" {
				output = "table"
			}

			return iostreams.FormatOutput(body, f.IO, output,
				iostreams.WithTransform(transformSchemaProducts),
				iostreams.WithColumns("product", "version"),
			)
		},
	}

	lf.Register(cmd)
	cmd.Flags().StringVarP(&product, "product", "p", "", "Filter by product code")
	cmd.Flags().StringVar(&version, "version", "", "Filter by firmware version")

	return cmd
}

// schemaProduct holds a single product/version pair for display.
type schemaProduct struct {
	Product string `json:"product"`
	Version string `json:"version"`
}

// transformSchemaProducts extracts unique product/version pairs from the
// config-documents/overviews API response, deduplicates, and sorts them.
func transformSchemaProducts(body []byte) ([]byte, error) {
	result := gjson.GetBytes(body, "result")
	if !result.Exists() {
		return []byte(`{"result":[]}`), nil
	}

	seen := make(map[string]bool)
	products := make([]schemaProduct, 0)

	for _, item := range result.Array() {
		p := item.Get("product").String()
		v := item.Get("version").String()
		if p == "" || v == "" {
			continue
		}
		key := p + "\x00" + v
		if seen[key] {
			continue
		}
		seen[key] = true
		products = append(products, schemaProduct{Product: p, Version: v})
	}

	sort.Slice(products, func(i, j int) bool {
		if products[i].Product != products[j].Product {
			return products[i].Product < products[j].Product
		}
		return products[i].Version < products[j].Version
	})

	out, err := json.Marshal(map[string]interface{}{"result": products})
	if err != nil {
		return nil, fmt.Errorf("formatting schema products: %w", err)
	}
	return out, nil
}
