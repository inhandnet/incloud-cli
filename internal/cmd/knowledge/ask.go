package knowledge

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/iostreams"
)

type askRequest struct {
	Question string `json:"question"`
	Model    string `json:"model,omitempty"`
}

type askResponse struct {
	Answer  string `json:"answer"`
	Sources []struct {
		Source  string `json:"source"`
		Heading string `json:"heading"`
	} `json:"sources"`
}

func NewCmdAsk(f *factory.Factory) *cobra.Command {
	var model string

	cmd := &cobra.Command{
		Use:   "ask <question>",
		Short: "Ask a question to the knowledge base",
		Long:  "Ask a question and get an AI-generated answer based on device documentation.",
		Example: `  # Ask a question
  incloud knowledge ask "How do I set up a VPN tunnel on IR915L?"

  # Specify device model for better context
  incloud knowledge ask "How to configure SNMP?" --model IR305

  # Get raw JSON output
  incloud knowledge ask "What is the default IP address?" -o json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := f.APIClient()
			if err != nil {
				return err
			}

			req := askRequest{
				Question: args[0],
				Model:    model,
			}

			body, err := client.Post("/api/v1/knowledge/ask", req)
			if err != nil {
				return err
			}

			output, _ := cmd.Flags().GetString("output")
			if output != "table" {
				return iostreams.FormatOutput(body, f.IO, output)
			}

			var resp askResponse
			if err := json.Unmarshal(body, &resp); err != nil {
				return fmt.Errorf("parsing ask response: %w", err)
			}

			out := f.IO.Out
			c := iostreams.NewColorizer(f.IO.TermOutput())

			fmt.Fprintln(out, resp.Answer)

			if len(resp.Sources) > 0 {
				fmt.Fprintln(out)
				fmt.Fprintln(out, c.Bold(c.Gray("Sources:")))
				for _, s := range resp.Sources {
					line := fmt.Sprintf("  - %s", s.Heading)
					if s.Source != "" {
						line += c.Gray(fmt.Sprintf(" (%s)", s.Source))
					}
					fmt.Fprintln(out, line)
				}
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&model, "model", "", "Filter by device model (e.g. IR915L)")

	return cmd
}
