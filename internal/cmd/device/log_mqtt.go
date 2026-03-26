package device

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/iostreams"
)

type LogMqttOptions struct {
	After  string
	Before string
	Type   string
	Topic  string
	Limit  int
	Next   string
	Prev   string
	Fields []string
}

var defaultMqttLogFields = []string{"timestamp", "logType", "topic"}

func NewCmdLogMqtt(f *factory.Factory) *cobra.Command {
	opts := &LogMqttOptions{}

	cmd := &cobra.Command{
		Use:   "mqtt <device-id>",
		Short: "View MQTT communication logs",
		Long:  "View MQTT message logs for a device, including publish, connect, and disconnect events.",
		Example: `  # View recent MQTT logs
  incloud device log mqtt 507f1f77bcf86cd799439011

  # Filter by time range
  incloud device log mqtt 507f1f77bcf86cd799439011 --after 2024-01-01T00:00:00Z --before 2024-01-02T00:00:00Z

  # Filter by message type
  incloud device log mqtt 507f1f77bcf86cd799439011 --type publish

  # Filter by topic
  incloud device log mqtt 507f1f77bcf86cd799439011 --topic shadow

  # Paginate with cursor
  incloud device log mqtt 507f1f77bcf86cd799439011 --next <cursor-token>

  # Show all fields including payload
  incloud device log mqtt 507f1f77bcf86cd799439011 -o table -f timestamp -f logType -f topic -f payload`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			deviceID := args[0]

			client, err := f.APIClient()
			if err != nil {
				return err
			}

			q := url.Values{}
			if opts.After != "" {
				q.Set("after", opts.After)
			}
			if opts.Before != "" {
				q.Set("before", opts.Before)
			}
			if opts.Type != "" {
				q.Set("type", opts.Type)
			}
			if opts.Topic != "" {
				q.Set("topic", opts.Topic)
			}
			if opts.Next != "" {
				q.Set("next", opts.Next)
			}
			if opts.Prev != "" {
				q.Set("prev", opts.Prev)
			}
			q.Set("limit", strconv.Itoa(opts.Limit))

			body, err := client.Get("/api/v1/devices/"+deviceID+"/mqttlogs", q)
			if err != nil {
				return err
			}

			output, _ := cmd.Flags().GetString("output")
			fields := opts.Fields
			if len(fields) == 0 {
				fields = defaultMqttLogFields
			}
			if err := iostreams.FormatOutput(body, f.IO, output, fields, iostreams.WithTransform(extractResultArray)); err != nil {
				return err
			}
			if output == "table" {
				printCursorHint(f, body)
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&opts.After, "after", "", "Start time (ISO 8601, e.g. 2024-01-01T00:00:00Z)")
	cmd.Flags().StringVar(&opts.Before, "before", "", "End time (ISO 8601, e.g. 2024-01-02T00:00:00Z)")
	cmd.Flags().StringVar(&opts.Type, "type", "", "Filter by log type: publish, connected, disconnected")
	cmd.Flags().StringVar(&opts.Topic, "topic", "", "Filter by MQTT topic (regex)")
	cmd.Flags().IntVar(&opts.Limit, "limit", 20, "Number of results per page")
	cmd.Flags().StringVar(&opts.Next, "next", "", "Cursor token for next page")
	cmd.Flags().StringVar(&opts.Prev, "prev", "", "Cursor token for previous page")
	cmd.Flags().StringSliceVarP(&opts.Fields, "fields", "f", nil, "Fields to display in table mode")
	_ = cmd.RegisterFlagCompletionFunc("type", func(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
		return []string{"publish", "connected", "disconnected"}, cobra.ShellCompDirectiveNoFileComp
	})

	return cmd
}

// extractResultArray extracts the "result" array from a cursor-paginated response.
func extractResultArray(body []byte) ([]byte, error) {
	var wrapper struct {
		Result json.RawMessage `json:"result"`
	}
	if err := json.Unmarshal(body, &wrapper); err != nil {
		return nil, fmt.Errorf("parsing response: %w", err)
	}
	if wrapper.Result == nil {
		return []byte("[]"), nil
	}
	return wrapper.Result, nil
}

// printCursorHint prints next/prev cursor hints in table mode.
func printCursorHint(f *factory.Factory, body []byte) {
	var cursor struct {
		Next string `json:"next"`
		Prev string `json:"prev"`
	}
	if err := json.Unmarshal(body, &cursor); err != nil {
		return
	}
	if cursor.Next != "" {
		fmt.Fprintf(f.IO.ErrOut, "\nNext page: --next %s\n", cursor.Next)
	}
	if cursor.Prev != "" {
		fmt.Fprintf(f.IO.ErrOut, "Prev page: --prev %s\n", cursor.Prev)
	}
}
