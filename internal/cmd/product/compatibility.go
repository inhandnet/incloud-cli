package product

import (
	"encoding/json"
	"net/url"

	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/iostreams"
)

func NewCmdCompatibility(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "compatibility <product-id-or-name>",
		Short: "List product compatibilities",
		Long:  "List all compatibilities defined for a product, including minimum firmware version requirements.",
		Example: `  # List compatibilities for a product
  incloud product compatibility IR915

  # JSON output
  incloud product compatibility IR915 -o json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			idOrName := args[0]

			client, err := f.APIClient()
			if err != nil {
				return err
			}

			q := url.Values{}
			q.Set("fields", "compatibilities")

			body, err := client.Get("/api/v1/products/"+url.PathEscape(idOrName), q)
			if err != nil {
				return err
			}

			output, _ := cmd.Flags().GetString("output")
			return iostreams.FormatOutput(body, f.IO, output,
				iostreams.WithTransform(extractCompatibilities),
			)
		},
	}

	return cmd
}

// extractCompatibilities extracts the compatibilities array from a product response.
func extractCompatibilities(data []byte) ([]byte, error) {
	var resp struct {
		Result struct {
			Compatibilities json.RawMessage `json:"compatibilities"`
		} `json:"result"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return data, nil
	}
	if resp.Result.Compatibilities == nil {
		return []byte("[]"), nil
	}
	return resp.Result.Compatibilities, nil
}
