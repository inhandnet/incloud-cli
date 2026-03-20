package api

import (
	"fmt"
	"io"
	"net/url"
	"os"
	"strings"

	"github.com/spf13/cobra"

	inapi "github.com/inhandnet/incloud-cli/internal/api"
	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/iostreams"
)

type ApiOptions struct {
	Path        string
	Method      string
	QueryParams []string
	BodyFields  []string
	Headers     []string
	InputFile   string
	OutputFile  string
	Columns     []string
}

func NewCmdApi(f *factory.Factory) *cobra.Command {
	opts := &ApiOptions{}

	cmd := &cobra.Command{
		Use:   "api <path>",
		Short: "Make an authenticated API request",
		Long: `Make an authenticated API request to the InCloud platform.

The path is appended to the current context's host URL.
Authorization header is automatically injected.`,
		Example: `  # GET current user
  incloud api /api/v1/users/me

  # List devices with query params
  incloud api /api/v1/devices -q page=0 -q limit=10

  # Output formats
  incloud api /api/v1/devices -o json
  incloud api /api/v1/devices -o table -c name -c status -c product
  incloud api /api/v1/devices -o yaml

  # Create device
  incloud api /api/v1/devices -X POST -f name=test -f product=router

  # POST with JSON from stdin
  echo '{"name":"test"}' | incloud api /api/v1/devices -X POST --input -

  # Custom header
  incloud api /api/v1/users/me -H "Sudo: user@example.com"

  # Download a file
  incloud api /api/v1/devices/DEVICE_ID/files/capture_result.pcap --output-file capture.pcap`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Path = args[0]
			if opts.Method == "" {
				opts.Method = "GET"
			}
			opts.Method = strings.ToUpper(opts.Method)

			client, err := f.APIClient()
			if err != nil {
				return err
			}

			// File download mode: stream response directly to file
			if opts.OutputFile != "" {
				if err := client.Download(opts.Path, opts.OutputFile); err != nil {
					return err
				}
				fmt.Fprintf(f.IO.ErrOut, "Saved to %s\n", opts.OutputFile)
				return nil
			}

			reqOpts, err := buildRequestOptions(opts)
			if err != nil {
				return err
			}

			body, err := client.Do(opts.Method, opts.Path, reqOpts)

			// Format output based on TTY and -o flag
			output, _ := cmd.Flags().GetString("output")
			if body != nil {
				if fmtErr := iostreams.FormatOutput(body, f.IO, output, opts.Columns); fmtErr != nil {
					return fmtErr
				}
			}

			if err != nil {
				return err
			}
			return nil
		},
	}

	cmd.Flags().StringVarP(&opts.Method, "method", "X", "", "HTTP method (default: GET)")
	cmd.Flags().StringArrayVarP(&opts.QueryParams, "query", "q", nil, "Query parameter (key=value)")
	cmd.Flags().StringArrayVarP(&opts.BodyFields, "field", "f", nil, "Body field (key=value), sent as JSON")
	cmd.Flags().StringArrayVarP(&opts.Headers, "header", "H", nil, "Additional header (Key: Value)")
	cmd.Flags().StringVar(&opts.InputFile, "input", "", "Read body from file (use - for stdin)")
	cmd.Flags().StringVar(&opts.OutputFile, "output-file", "", "Save response body to file (for binary downloads)")
	cmd.Flags().StringArrayVarP(&opts.Columns, "column", "c", nil, "Columns to show in table output")

	return cmd
}

func buildRequestOptions(opts *ApiOptions) (*inapi.RequestOptions, error) {
	reqOpts := &inapi.RequestOptions{}

	// query params
	if len(opts.QueryParams) > 0 {
		q := make(url.Values)
		for _, param := range opts.QueryParams {
			parts := strings.SplitN(param, "=", 2)
			if len(parts) != 2 {
				return nil, fmt.Errorf("invalid query param: %s (expected key=value)", param)
			}
			q.Set(parts[0], parts[1])
		}
		reqOpts.Query = q
	}

	// body
	if len(opts.BodyFields) > 0 {
		data := make(map[string]interface{})
		for _, field := range opts.BodyFields {
			parts := strings.SplitN(field, "=", 2)
			if len(parts) != 2 {
				return nil, fmt.Errorf("invalid field: %s (expected key=value)", field)
			}
			data[parts[0]] = parts[1]
		}
		reqOpts.Body = data
	} else if opts.InputFile != "" {
		var reader io.Reader
		if opts.InputFile == "-" {
			reader = os.Stdin
		} else {
			f, err := os.Open(opts.InputFile)
			if err != nil {
				return nil, err
			}
			reader = f
		}
		reqOpts.RawBody = reader
		reqOpts.ContentType = "application/json"
	}

	// custom headers
	if len(opts.Headers) > 0 {
		headers := make(map[string]string)
		for _, h := range opts.Headers {
			parts := strings.SplitN(h, ":", 2)
			if len(parts) != 2 {
				return nil, fmt.Errorf("invalid header: %s (expected Key: Value)", h)
			}
			headers[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
		}
		reqOpts.Headers = headers
	}

	return reqOpts, nil
}
