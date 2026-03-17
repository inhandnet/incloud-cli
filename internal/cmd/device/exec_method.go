package device

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
)

func NewCmdExecMethod(f *factory.Factory) *cobra.Command {
	var (
		payload string
		timeout int
	)

	cmd := &cobra.Command{
		Use:   "method <id>[,<id2>,...] <method>",
		Short: "Invoke a custom remote method on device(s)",
		Long: `Invoke a custom remote method on one or more devices.

When multiple device IDs are provided (comma-separated), the bulk endpoint
is used and the request is processed asynchronously.`,
		Example: `  # Invoke a custom method
  incloud device exec method 507f1f77bcf86cd799439011 getConfig --payload '{"module":"wan"}'

  # Bulk invoke a method on multiple devices
  incloud device exec method 507f1f77bcf86cd799439011,653b1ff2a84e171614d88695 syncTime`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			idsArg := args[0]
			method := args[1]

			cfg, err := f.Config()
			if err != nil {
				return err
			}
			actx, err := cfg.ActiveContext()
			if err != nil {
				return err
			}

			client, err := f.HttpClient()
			if err != nil {
				return err
			}

			var payloadObj interface{}
			if payload != "" {
				if err := json.Unmarshal([]byte(payload), &payloadObj); err != nil {
					return fmt.Errorf("invalid --payload JSON: %w", err)
				}
			}

			ids := strings.Split(idsArg, ",")
			if len(ids) > 1 {
				return bulkInvokeMethod(cmd, f, client, actx.Host, ids, method, payloadObj)
			}
			return invokeMethod(cmd, f, client, actx.Host, ids[0], method, timeout, payloadObj)
		},
	}

	cmd.Flags().StringVarP(&payload, "payload", "p", "", "JSON payload for the method")
	cmd.Flags().IntVarP(&timeout, "timeout", "t", 30, "Timeout in seconds (5-300, single device only)")

	return cmd
}

func invokeMethod(cmd *cobra.Command, f *factory.Factory, client *http.Client, host, id, method string, timeout int, payload interface{}) error {
	body := map[string]interface{}{
		"method":  method,
		"timeout": timeout,
	}
	if payload != nil {
		body["payload"] = payload
	}

	jsonBytes, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("encoding request body: %w", err)
	}

	reqURL := host + "/api/v1/devices/" + id + "/methods"
	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, reqURL, bytes.NewReader(jsonBytes))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("reading response: %w", err)
	}

	if err := formatOutput(cmd, f.IO, respBody, nil); err != nil {
		return err
	}

	if resp.StatusCode >= 400 {
		return fmt.Errorf("HTTP %d", resp.StatusCode)
	}
	return nil
}

func bulkInvokeMethod(cmd *cobra.Command, f *factory.Factory, client *http.Client, host string, ids []string, method string, payload interface{}) error {
	body := map[string]interface{}{
		"deviceIds": ids,
		"method":    method,
	}
	if payload != nil {
		body["payload"] = payload
	}

	jsonBytes, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("encoding request body: %w", err)
	}

	reqURL := host + "/api/v1/devices/bulk-invoke-methods"
	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, reqURL, bytes.NewReader(jsonBytes))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("reading response: %w", err)
	}

	if resp.StatusCode >= 400 {
		fmt.Fprintln(f.IO.ErrOut, string(respBody))
		return fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	fmt.Fprintf(f.IO.ErrOut, "Method %s submitted for %d device(s).\n", method, len(ids))
	return nil
}
