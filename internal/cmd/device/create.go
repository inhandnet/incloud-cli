package device

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tidwall/gjson"

	"github.com/inhandnet/incloud-cli/internal/api"
	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/iostreams"
	"github.com/inhandnet/incloud-cli/internal/ui"
)

type CreateOptions struct {
	Name        string
	SN          string
	Product     string
	Description string
	Group       string
	Mac         string
	IMEI        string
	Labels      []string
	Metadata    []string
}

func NewCmdCreate(f *factory.Factory) *cobra.Command {
	opts := &CreateOptions{}

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a device",
		Long: `Create a new device on the InCloud platform.

The serial number is validated before creation to detect the product model
and determine whether a MAC address or IMEI is required. If the required
credential is not provided via flags and the terminal is interactive, you
will be prompted to enter it.`,
		Example: `  # Create a device (product auto-detected from serial number)
  incloud device create --name "My Router" --sn "ABC123456789012" --mac "AA:BB:CC:DD:EE:FF"

  # MAC or IMEI will be prompted in TTY if required but not provided
  incloud device create --name "My Router" --sn "ABC123456789012"

  # With device group and labels
  incloud device create --name "My Router" --sn "ABC123456789012" --mac "AA:BB:CC:DD:EE:FF" \
    --group "group-id" --label env=prod --label region=us

  # With metadata
  incloud device create --name "My Router" --sn "ABC123456789012" --mac "AA:BB:CC:DD:EE:FF" \
    --metadata key1=val1`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCreate(cmd, f, opts)
		},
	}

	cmd.Flags().StringVar(&opts.Name, "name", "", "Device name (required)")
	cmd.Flags().StringVar(&opts.SN, "sn", "", "Serial number (required)")
	cmd.Flags().StringVar(&opts.Product, "product", "", "Product model (auto-detected from serial number)")
	cmd.Flags().StringVar(&opts.Description, "description", "", "Device description (max 256 chars)")
	cmd.Flags().StringVar(&opts.Group, "group", "", "Device group ID (use 'incloud device group list' to find IDs)")
	cmd.Flags().StringVar(&opts.Mac, "mac", "", "MAC address (required for some products; prompted if omitted in TTY)")
	cmd.Flags().StringVar(&opts.IMEI, "imei", "", "IMEI (required for some products; prompted if omitted in TTY)")
	cmd.Flags().StringArrayVar(&opts.Labels, "label", nil, "Label in key=value format (repeatable, max 10)")
	cmd.Flags().StringArrayVar(&opts.Metadata, "metadata", nil, "Metadata in key=value format (repeatable)")

	_ = cmd.MarkFlagRequired("name")
	_ = cmd.MarkFlagRequired("sn")

	return cmd
}

func runCreate(cmd *cobra.Command, f *factory.Factory, opts *CreateOptions) error {
	client, err := f.APIClient()
	if err != nil {
		return err
	}

	sn := strings.ToUpper(opts.SN)

	// Step 1: Validate serial number via API
	fmt.Fprintf(f.IO.ErrOut, "Validating serial number %s...\n", sn)
	validation, err := validateSerialNumber(client, sn)
	if err != nil {
		return err
	}
	fmt.Fprintf(f.IO.ErrOut, "Product: %s (requires %s)\n", validation.Product, validation.ValidatedField)

	// Step 2: Prompt for MAC/IMEI if required but not provided
	if validation.ValidatedField == "mac" && opts.Mac == "" {
		if ui.IsTTY(f) {
			mac, promptErr := ui.Input(f, "MAC Address", "AA:BB:CC:DD:EE:FF", nil)
			if promptErr != nil {
				return promptErr
			}
			opts.Mac = mac
		} else {
			return fmt.Errorf("serial number %s requires a MAC address; use --mac XX:XX:XX:XX:XX:XX", sn)
		}
	}
	if validation.ValidatedField == "imei" && opts.IMEI == "" {
		if ui.IsTTY(f) {
			imei, promptErr := ui.Input(f, "IMEI", "15-17 digits", nil)
			if promptErr != nil {
				return promptErr
			}
			opts.IMEI = imei
		} else {
			return fmt.Errorf("serial number %s requires an IMEI; use --imei <15-17 digits>", sn)
		}
	}

	// Step 3: Build request body
	body := map[string]interface{}{
		"name":         opts.Name,
		"serialNumber": sn,
	}

	product := validation.Product
	if opts.Product != "" {
		product = opts.Product
	}
	if product != "" {
		body["product"] = product
	}
	if opts.Description != "" {
		body["description"] = opts.Description
	}
	if opts.Group != "" {
		body["devicegroupId"] = opts.Group
	}
	if opts.Mac != "" {
		body["mac"] = opts.Mac
	}
	if opts.IMEI != "" {
		body["imei"] = opts.IMEI
	}
	if len(opts.Labels) > 0 {
		labels, labelsErr := parseLabels(opts.Labels)
		if labelsErr != nil {
			return labelsErr
		}
		body["labels"] = labels
	}
	if len(opts.Metadata) > 0 {
		meta, metaErr := parseKeyValues(opts.Metadata)
		if metaErr != nil {
			return metaErr
		}
		body["metadata"] = meta
	}

	// Step 4: Create device
	respBody, err := client.Post("/api/v1/devices", body)
	if err != nil {
		return formatCreateError(sn, opts, respBody, err)
	}

	// Step 5: Success message
	name := gjson.GetBytes(respBody, "result.name").String()
	id := gjson.GetBytes(respBody, "result._id").String()
	fmt.Fprintf(f.IO.ErrOut, "Device %q created. (id: %s)\n", name, id)

	return formatOutput(cmd, f.IO, respBody, nil)
}

// snValidation holds the result of serial number validation.
type snValidation struct {
	Product        string // product model, e.g. "IR615"
	ValidatedField string // "mac" or "imei"
}

// validateSerialNumber calls the API to validate a serial number and returns
// the product type and which credential field (MAC or IMEI) is required.
func validateSerialNumber(client *api.APIClient, sn string) (*snValidation, error) {
	body, err := client.Post("/api/v1/serialnumber/"+sn+"/validate", nil)
	if err != nil {
		return nil, formatSNValidationError(sn, body, err)
	}
	product := gjson.GetBytes(body, "result.product").String()
	field := gjson.GetBytes(body, "result.validatedField").String()
	return &snValidation{Product: product, ValidatedField: field}, nil
}

// formatSNValidationError maps SN validation API errors to user-friendly messages.
func formatSNValidationError(sn string, body []byte, originalErr error) error {
	if len(body) == 0 {
		return originalErr
	}

	errCode := gjson.GetBytes(body, "error").String()
	switch errCode {
	case "resource_not_found":
		return fmt.Errorf("serial number %q is not recognized; it may be unsupported or the product is obsolete", sn)
	case "invalid_request":
		return fmt.Errorf("serial number %q has an invalid format", sn)
	case "invalid_state":
		return fmt.Errorf("product for serial number %q is no longer supported", sn)
	default:
		return originalErr
	}
}

// formatCreateError maps device creation API errors to user-friendly messages.
func formatCreateError(sn string, opts *CreateOptions, body []byte, originalErr error) error {
	if len(body) == 0 {
		return originalErr
	}

	errCode := gjson.GetBytes(body, "error").String()
	extType := gjson.GetBytes(body, "ext.type").String()
	message := gjson.GetBytes(body, "message").String()

	switch errCode {
	case "resource_already_exists":
		switch extType {
		case "name":
			return fmt.Errorf("device name %q already exists", opts.Name)
		case "serialNumber":
			return fmt.Errorf("serial number %q already exists", sn)
		}
	case "resource_not_found":
		switch extType {
		case "serialNumber":
			return fmt.Errorf("serial number %q is not valid", sn)
		case "mac":
			return fmt.Errorf("MAC address %q does not match serial number %q", opts.Mac, sn)
		}
	case "request_not_allowed":
		desc := gjson.GetBytes(body, "description").String()
		if strings.Contains(strings.ToLower(message), "device info is incorrect") ||
			strings.Contains(message, "device_info_is_incorrect") {
			switch desc {
			case "MAC_INVALID":
				return fmt.Errorf("MAC address %q does not match serial number %q", opts.Mac, sn)
			case "IMEI_INVALID":
				return fmt.Errorf("IMEI %q does not match serial number %q", opts.IMEI, sn)
			default:
				return fmt.Errorf("device information is incorrect; verify SN and MAC/IMEI combination")
			}
		}
	}

	return originalErr
}

// parseLabels converts ["key=value", ...] into [{"name":"key","value":"value"}, ...]
func parseLabels(pairs []string) ([]map[string]string, error) {
	labels := make([]map[string]string, 0, len(pairs))
	for _, pair := range pairs {
		parts := strings.SplitN(pair, "=", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid label: %s (expected key=value)", pair)
		}
		labels = append(labels, map[string]string{
			"name":  parts[0],
			"value": parts[1],
		})
	}
	return labels, nil
}

// parseKeyValues converts ["key=value", ...] into {"key":"value", ...}
func parseKeyValues(pairs []string) (map[string]string, error) {
	m := make(map[string]string, len(pairs))
	for _, pair := range pairs {
		parts := strings.SplitN(pair, "=", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid key=value: %s", pair)
		}
		m[parts[0]] = parts[1]
	}
	return m, nil
}

// formatOutput handles -o flag output formatting, shared by create/update.
func formatOutput(cmd *cobra.Command, streams *iostreams.IOStreams, body []byte, columns []string) error {
	output, _ := cmd.Flags().GetString("output")
	return iostreams.FormatOutput(body, streams, output, columns)
}
