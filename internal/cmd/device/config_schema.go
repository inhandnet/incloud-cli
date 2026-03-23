package device

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tidwall/gjson"

	"github.com/inhandnet/incloud-cli/internal/api"
	"github.com/inhandnet/incloud-cli/internal/factory"
)

const defaultConfigModule = "default"

type productVersion struct {
	product string
	version string
}

// configDocumentQuery builds common query params for the config-documents API.
func (pv *productVersion) configDocumentQuery() url.Values {
	q := url.Values{}
	q.Set("product", pv.product)
	q.Set("version", pv.version)
	q.Set("module", defaultConfigModule)
	return q
}

// resolveProductVersion resolves product and version from either a device ID
// or explicit --product/--version flags. The two approaches are mutually exclusive.
func resolveProductVersion(client *api.APIClient, deviceID, product, version string) (*productVersion, error) {
	hasDevice := deviceID != ""
	hasProduct := product != ""
	hasVersion := version != ""

	switch {
	case hasDevice && (hasProduct || hasVersion):
		return nil, fmt.Errorf("--device and --product/--version are mutually exclusive")
	case hasDevice:
		body, err := client.Get("/api/v1/devices/"+deviceID, url.Values{
			"fields": {"product,firmware"},
		})
		if err != nil {
			return nil, fmt.Errorf("fetching device %s: %w", deviceID, err)
		}
		r := gjson.ParseBytes(body)
		p := r.Get("result.product").String()
		v := r.Get("result.firmware").String()
		if p == "" || v == "" {
			return nil, fmt.Errorf("device %s is missing product or firmware info", deviceID)
		}
		return &productVersion{product: p, version: v}, nil
	case hasProduct && hasVersion:
		return &productVersion{product: product, version: version}, nil
	case hasProduct:
		return nil, fmt.Errorf("--version is required when using --product")
	case hasVersion:
		return nil, fmt.Errorf("--product is required when using --version")
	default:
		return nil, fmt.Errorf("either --device or --product/--version is required")
	}
}

// suggestAvailableVersions queries the API for other versions that have schema
// data for the given product. Returns a hint string like "(schemas available for
// versions: V1.0, V2.0)" or empty string if nothing found.
func suggestAvailableVersions(client *api.APIClient, product string) string {
	q := url.Values{}
	q.Set("product", product)
	q.Set("limit", "5")
	q.Set("fields", "version")

	body, err := client.Get("/api/v1/config-documents", q)
	if err != nil {
		return ""
	}

	result := gjson.GetBytes(body, "result")
	if !result.Exists() || len(result.Array()) == 0 {
		return ""
	}

	seen := make(map[string]bool)
	var versions []string
	for _, item := range result.Array() {
		v := item.Get("version").String()
		if v != "" && !seen[v] {
			seen[v] = true
			versions = append(versions, v)
		}
	}

	if len(versions) == 0 {
		return ""
	}
	return fmt.Sprintf("(schemas available for versions: %s)", strings.Join(versions, ", "))
}

// newCmdConfigSchema creates the `device config schema` parent command.
func newCmdConfigSchema(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "schema",
		Short: "Browse and validate device configuration schemas",
		Long: `Discover configuration schemas for a device's product/firmware,
and validate JSON payloads before writing.

AI tools workflow:
  1. incloud device config schema overview --device <id>
  2. incloud device config schema list --device <id>
  3. incloud device config schema get --device <id> <json-key>
  4. incloud device config schema validate --device <id> --key <json-key> --payload '{...}'
  5. incloud device config update <id> --payload '{...}'`,
	}

	cmd.AddCommand(newCmdSchemaList(f))
	cmd.AddCommand(newCmdSchemaGet(f))
	cmd.AddCommand(newCmdSchemaOverview(f))
	cmd.AddCommand(newCmdSchemaValidate(f))

	return cmd
}
