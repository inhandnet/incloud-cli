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
		key           string
		payload       string
		file          string
		wholeDocument bool
	)

	cmd := &cobra.Command{
		Use:   "validate",
		Short: "Validate JSON payload against a config schema",
		Long: `Validate a JSON configuration payload against the device's config schema
before writing it with 'incloud device config update'.

With --device, the payload is validated the way 'config update' actually applies
it: the payload is treated as a JSON merge patch over the device's current
configuration, and the merged result is validated. This means an incremental
payload that only touches some fields of a config block is accepted, matching
what 'config update' does.

With --product/--version there is no current configuration to merge onto, so the
payload is validated as a whole document. Use --whole-document to force that
behaviour even when --device is given.

Uses JSON Schema draft-07 validation. Exits with code 0 on success, 1 on
validation failure. Useful for AI tools to pre-check generated config.`,
		Example: `  # Validate a JSON payload
  incloud device config schema validate --device 507f1f77bcf86cd799439011 \
    --key dns --payload '{"dns":{"primary":"8.8.8.8"}}'

  # Validate from file
  incloud device config schema validate --product MR805 --version V2.0.15-111 \
    --key dns --file dns-config.json

  # Incremental payload: validated against the merged result, like 'config update'
  incloud device config schema validate --device 507f1f77bcf86cd799439011 \
    --key qos --payload '{"qos":{"user_rules":[]}}'

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

			doc := result.Array()[0]
			schemaContent := doc.Get("content").String()
			if schemaContent == "" {
				return fmt.Errorf("config schema %q has no content", key)
			}

			var schemaKeys []string
			for _, jk := range doc.Get("jsonKeys").Array() {
				if s := jk.String(); s != "" {
					schemaKeys = append(schemaKeys, s)
				}
			}
			if len(schemaKeys) == 0 {
				schemaKeys = []string{key}
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

			// With a device, model what 'config update' actually does: the payload is a
			// JSON merge patch over the device's current configuration. Validate the
			// merged result rather than the bare payload, so that an incremental payload
			// is judged the same way the write path will treat it.
			merged := sf.device != "" && !wholeDocument
			target := payloadObj
			var preExisting []validationError
			if merged {
				current, err := fetchMergedConfig(client, sf.device)
				if err != nil {
					// Never fall back to whole-document validation here: silently
					// downgrading would let the caller believe the merged result was
					// checked when it was not, which is the very bug this guards.
					return err
				}
				base := pruneToKeys(current, schemaKeys)
				target = applyMergePatch(base, payloadObj)

				// Stored device configs are not guaranteed to satisfy their own
				// schema: fields the schema does not declare and values that fail
				// its formats both occur in practice. Those violations are not
				// caused by this payload and must not fail it, or an incremental
				// update gets rejected for something it never touched.
				preExisting = validationErrors(sch.Validate(base))
			}

			// Validate
			validationErr := sch.Validate(target)
			causes := validationErrors(validationErr)
			introduced := subtractErrors(causes, preExisting)

			if len(introduced) == 0 {
				if merged {
					fmt.Fprintf(f.IO.ErrOut, "Validation passed (against merged config for device %s).\n", sf.device)
					writePreExistingNote(f.IO.ErrOut, preExisting)
				} else {
					fmt.Fprintf(f.IO.ErrOut, "Validation passed.\n")
				}
				return nil
			}

			// Format validation errors
			var sb strings.Builder
			sb.WriteString("Validation failed:\n")
			for _, cause := range introduced {
				fmt.Fprintf(&sb, "  - %s: %s\n", cause.path, cause.message)
			}
			blockLevel := hasBlockLevelRequired(introduced)
			if !merged && blockLevel {
				sb.WriteString("\nNote: this key's schema requires the whole block. " +
					"'config update' performs a partial merge, so an incremental payload may " +
					"still be accepted. Re-run with --device <id> to validate against the " +
					"actual merged result.\n")
			}
			return fmt.Errorf("%s", sb.String())
		},
	}

	sf.register(cmd)
	cmd.Flags().StringVarP(&key, "key", "k", "", "JSON key identifying the config schema to validate against (required; use 'incloud device config schema list' to find keys)")
	cmd.Flags().StringVar(&payload, "payload", "", "JSON payload to validate")
	cmd.Flags().StringVar(&file, "file", "", "Path to JSON file to validate")
	cmd.Flags().BoolVar(&wholeDocument, "whole-document", false, "Validate the payload as a complete document instead of merging it over the device's current configuration (no effect without --device)")
	_ = cmd.MarkFlagRequired("key")

	return cmd
}

type validationError struct {
	path    string
	message string
	// depth is the number of segments in the instance location. 0 is the document
	// root, 1 is a top-level config block such as /qos.
	depth int
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
		*out = append(*out, validationError{path: path, message: ve.Error(), depth: len(ve.InstanceLocation)})
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
