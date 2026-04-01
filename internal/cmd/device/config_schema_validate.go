package device

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/dlclark/regexp2"
	"github.com/santhosh-tekuri/jsonschema/v6"
	"github.com/spf13/cobra"
	"github.com/tidwall/gjson"

	"github.com/inhandnet/incloud-cli/internal/factory"
)

func newCmdSchemaValidate(f *factory.Factory) *cobra.Command {
	sf := &schemaFlags{}
	var (
		key     string
		payload string
		file    string
	)

	cmd := &cobra.Command{
		Use:   "validate",
		Short: "Validate JSON payload against a config schema",
		Long: `Validate a JSON configuration payload against the device's config schema
before writing it with 'incloud device config update'.

Uses JSON Schema draft-07 validation. Exits with code 0 on success, 1 on
validation failure. Useful for AI tools to pre-check generated config.`,
		Example: `  # Validate a JSON payload
  incloud device config schema validate --device 507f1f77bcf86cd799439011 \
    --key dns --payload '{"dns":{"primary":"8.8.8.8"}}'

  # Validate from file
  incloud device config schema validate --product MR805 --version V2.0.15-111 \
    --key dns --file dns-config.json

  # Use in pipeline: validate then apply
  incloud device config schema validate -d <id> --key dns --payload '...' && \
  incloud device config update <id> --payload '...'`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Read payload
			var data []byte
			var err error

			switch {
			case payload != "" && file != "":
				return fmt.Errorf("--payload and --file are mutually exclusive")
			case payload != "":
				data = []byte(payload)
			case file != "":
				data, err = os.ReadFile(file)
				if err != nil {
					return fmt.Errorf("reading file: %w", err)
				}
			default:
				return fmt.Errorf("either --payload or --file is required")
			}

			// Parse payload
			var payloadObj interface{}
			if err := json.Unmarshal(data, &payloadObj); err != nil {
				return fmt.Errorf("invalid JSON payload: %w", err)
			}

			client, err := f.APIClient()
			if err != nil {
				return err
			}

			pv, err := sf.resolve(client)
			if err != nil {
				return err
			}

			// Fetch schema
			q := pv.configDocumentQuery()
			q.Set("jsonKeys", key)

			body, err := client.Get("/api/v1/config-documents", q)
			if err != nil {
				return err
			}

			result := gjson.GetBytes(body, "result")
			if !result.Exists() || len(result.Array()) == 0 {
				return fmt.Errorf("config schema %q not found for %s/%s", key, pv.product, pv.version)
			}

			schemaContent := result.Array()[0].Get("content").String()
			if schemaContent == "" {
				return fmt.Errorf("config schema %q has no content", key)
			}

			// Parse and compile JSON Schema
			var schemaObj interface{}
			if err := json.Unmarshal([]byte(schemaContent), &schemaObj); err != nil {
				return fmt.Errorf("invalid schema JSON: %w", err)
			}

			compiler := jsonschema.NewCompiler()
			compiler.UseRegexpEngine(regexp2Engine)
			if err := compiler.AddResource("schema.json", schemaObj); err != nil {
				return fmt.Errorf("loading schema: %w", err)
			}
			sch, err := compiler.Compile("schema.json")
			if err != nil {
				return fmt.Errorf("compiling schema: %w", err)
			}

			// Validate
			validationErr := sch.Validate(payloadObj)
			if validationErr == nil {
				fmt.Fprintf(f.IO.ErrOut, "Validation passed.\n")
				return nil
			}

			// Format validation errors
			var sb strings.Builder
			sb.WriteString("Validation failed:\n")
			if ve, ok := validationErr.(*jsonschema.ValidationError); ok {
				for _, cause := range flattenValidationErrors(ve) {
					fmt.Fprintf(&sb, "  - %s: %s\n", cause.path, cause.message)
				}
			} else {
				fmt.Fprintf(&sb, "  - %s\n", validationErr.Error())
			}
			return fmt.Errorf("%s", sb.String())
		},
	}

	sf.register(cmd)
	cmd.Flags().StringVarP(&key, "key", "k", "", "JSON key identifying the config schema to validate against (required; use 'incloud device config schema list' to find keys)")
	cmd.Flags().StringVar(&payload, "payload", "", "JSON payload to validate")
	cmd.Flags().StringVar(&file, "file", "", "Path to JSON file to validate")
	_ = cmd.MarkFlagRequired("key")

	return cmd
}

type validationError struct {
	path    string
	message string
}

// flattenValidationErrors extracts leaf validation errors with their JSON paths.
func flattenValidationErrors(ve *jsonschema.ValidationError) []validationError {
	var errors []validationError
	flattenVE(ve, &errors)
	return errors
}

func flattenVE(ve *jsonschema.ValidationError, out *[]validationError) {
	if len(ve.Causes) == 0 {
		path := "/" + strings.Join(ve.InstanceLocation, "/")
		if path == "/" {
			path = "$"
		}
		*out = append(*out, validationError{path: path, message: ve.Error()})
		return
	}
	for _, cause := range ve.Causes {
		flattenVE(cause, out)
	}
}

// regexp2Engine is a jsonschema.RegexpEngine that uses regexp2 (PCRE-compatible)
// instead of Go's RE2. This supports Unicode escapes (\u4e00) and lookaheads
// (?!...) found in JSON Schema patterns from the backend.
func regexp2Engine(pattern string) (jsonschema.Regexp, error) {
	re, err := regexp2.Compile(pattern, regexp2.None)
	if err != nil {
		return nil, err
	}
	return &regexp2Regexp{re: re}, nil
}

type regexp2Regexp struct {
	re *regexp2.Regexp
}

func (r *regexp2Regexp) MatchString(s string) bool {
	matched, _ := r.re.MatchString(s)
	return matched
}

func (r *regexp2Regexp) String() string {
	return r.re.String()
}
