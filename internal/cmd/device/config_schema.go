package device

import (
	"fmt"
	"net/url"

	"github.com/spf13/cobra"
	"github.com/tidwall/gjson"

	"github.com/inhandnet/incloud-cli/internal/api"
	"github.com/inhandnet/incloud-cli/internal/factory"
)

type productVersion struct {
	product string
	version string
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
			"fields": {"partNumber,firmware"},
		})
		if err != nil {
			return nil, fmt.Errorf("fetching device %s: %w", deviceID, err)
		}
		r := gjson.ParseBytes(body)
		p := r.Get("result.partNumber").String()
		v := r.Get("result.firmware").String()
		if p == "" || v == "" {
			return nil, fmt.Errorf("device %s is missing partNumber or firmware info", deviceID)
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

	return cmd
}
