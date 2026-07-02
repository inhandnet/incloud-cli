package apidoc

import (
	"fmt"
	"net/url"
	"os"
	"slices"
	"strings"

	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/iostreams"
)

// validApps and validLangs mirror the nezha-support ApiDocController params:
// GET /api/v1/api-docs?lang={en|zh}&app={star|devicelive}.
var (
	validApps  = []string{"star", "devicelive"}
	validLangs = []string{"en", "zh"}
)

type apiDocOptions struct {
	Lang       string
	App        string
	OutputFile string
}

func NewCmdApiDoc(f *factory.Factory) *cobra.Command {
	opts := &apiDocOptions{}

	cmd := &cobra.Command{
		Use:   "apidoc",
		Short: "Download the InCloud public API spec for integration",
		Long: `Download the OpenAPI spec for InCloud's public (integration) APIs.

This is the same spec that backs the online API reference. Use it to build
against the platform: generate client SDKs, import into Postman/Insomnia, or
browse the available endpoints, parameters and schemas offline.

Use --lang to choose the language and --app to choose which spec to fetch.`,
		Example: `  # Print the English InCloud API spec
  incloud apidoc

  # Chinese spec
  incloud apidoc --lang zh

  # Device Live spec, saved to a file for code generation
  incloud apidoc --app devicelive --output-file device-live-api.json

  # List all available endpoints
  incloud apidoc --jq '.paths | keys'`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Lang = strings.ToLower(opts.Lang)
			opts.App = strings.ToLower(opts.App)
			if !slices.Contains(validLangs, opts.Lang) {
				return fmt.Errorf("invalid --lang %q (expected one of: %s)", opts.Lang, strings.Join(validLangs, ", "))
			}
			if !slices.Contains(validApps, opts.App) {
				return fmt.Errorf("invalid --app %q (expected one of: %s)", opts.App, strings.Join(validApps, ", "))
			}

			client, err := f.APIClient()
			if err != nil {
				return err
			}

			q := url.Values{}
			q.Set("lang", opts.Lang)
			q.Set("app", opts.App)

			body, err := client.Get("/api/v1/api-docs", q)
			if err != nil {
				return err
			}

			if opts.OutputFile != "" {
				if err := os.WriteFile(opts.OutputFile, body, 0o644); err != nil {
					return fmt.Errorf("writing file: %w", err)
				}
				fmt.Fprintf(f.IO.ErrOut, "Saved to %s\n", opts.OutputFile)
				return nil
			}

			output, _ := cmd.Flags().GetString("output")
			return iostreams.FormatOutput(body, f.IO, output)
		},
	}

	cmd.Flags().StringVar(&opts.Lang, "lang", "en", "Spec language: en, zh")
	cmd.Flags().StringVar(&opts.App, "app", "star", "Which spec to fetch: star (InCloud API), devicelive (Device Live API)")
	cmd.Flags().StringVar(&opts.OutputFile, "output-file", "", "Save the spec to a file instead of printing")

	return cmd
}
