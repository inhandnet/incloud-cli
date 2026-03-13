package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/spf13/cobra"
)

type ApiOptions struct {
	Path        string
	Method      string
	QueryParams []string
	BodyFields  []string
	Headers     []string
	InputFile   string
	Host        string
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

  # Create device
  incloud api /api/v1/devices -X POST -f name=test -f product=router

  # POST with JSON from stdin
  echo '{"name":"test"}' | incloud api /api/v1/devices -X POST --input -

  # Custom header
  incloud api /api/v1/users/me -H "Sudo: user@example.com"`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Path = args[0]
			if opts.Method == "" {
				opts.Method = "GET"
			}
			opts.Method = strings.ToUpper(opts.Method)

			cfg, err := f.Config()
			if err != nil {
				return err
			}
			ctx, err := cfg.ActiveContext()
			if err != nil {
				return err
			}
			opts.Host = ctx.Host

			client, err := f.HttpClient()
			if err != nil {
				return err
			}

			req, err := buildRequest(opts)
			if err != nil {
				return err
			}

			resp, err := client.Do(req)
			if err != nil {
				return fmt.Errorf("request failed: %w", err)
			}
			defer resp.Body.Close()

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				return fmt.Errorf("reading response: %w", err)
			}

			// pretty print JSON if possible
			output, _ := cmd.Flags().GetString("output")
			if output == "json" {
				// raw JSON for piping
				fmt.Fprintln(f.IO.Out, string(body))
			} else {
				// pretty print
				var prettyJSON bytes.Buffer
				if err := json.Indent(&prettyJSON, body, "", "  "); err != nil {
					// not JSON, print raw
					fmt.Fprintln(f.IO.Out, string(body))
				} else {
					fmt.Fprintln(f.IO.Out, prettyJSON.String())
				}
			}

			if resp.StatusCode >= 400 {
				return fmt.Errorf("HTTP %d", resp.StatusCode)
			}
			return nil
		},
	}

	cmd.Flags().StringVarP(&opts.Method, "method", "X", "", "HTTP method (default: GET)")
	cmd.Flags().StringArrayVarP(&opts.QueryParams, "query", "q", nil, "Query parameter (key=value)")
	cmd.Flags().StringArrayVarP(&opts.BodyFields, "field", "f", nil, "Body field (key=value), sent as JSON")
	cmd.Flags().StringArrayVarP(&opts.Headers, "header", "H", nil, "Additional header (Key: Value)")
	cmd.Flags().StringVar(&opts.InputFile, "input", "", "Read body from file (use - for stdin)")

	return cmd
}

func buildRequest(opts *ApiOptions) (*http.Request, error) {
	method := opts.Method
	if method == "" {
		method = "GET"
	}

	u, err := url.Parse(opts.Host + opts.Path)
	if err != nil {
		return nil, fmt.Errorf("invalid URL: %w", err)
	}

	// query params
	if len(opts.QueryParams) > 0 {
		q := u.Query()
		for _, param := range opts.QueryParams {
			parts := strings.SplitN(param, "=", 2)
			if len(parts) != 2 {
				return nil, fmt.Errorf("invalid query param: %s (expected key=value)", param)
			}
			q.Set(parts[0], parts[1])
		}
		u.RawQuery = q.Encode()
	}

	// body
	var body io.Reader
	if len(opts.BodyFields) > 0 {
		data := make(map[string]interface{})
		for _, field := range opts.BodyFields {
			parts := strings.SplitN(field, "=", 2)
			if len(parts) != 2 {
				return nil, fmt.Errorf("invalid field: %s (expected key=value)", field)
			}
			data[parts[0]] = parts[1]
		}
		jsonBytes, err := json.Marshal(data)
		if err != nil {
			return nil, err
		}
		body = bytes.NewReader(jsonBytes)
	} else if opts.InputFile != "" {
		if opts.InputFile == "-" {
			body = os.Stdin
		} else {
			f, err := os.Open(opts.InputFile)
			if err != nil {
				return nil, err
			}
			body = f
		}
	}

	req, err := http.NewRequest(method, u.String(), body)
	if err != nil {
		return nil, err
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	// custom headers
	for _, h := range opts.Headers {
		parts := strings.SplitN(h, ":", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid header: %s (expected Key: Value)", h)
		}
		req.Header.Set(strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1]))
	}

	return req, nil
}
