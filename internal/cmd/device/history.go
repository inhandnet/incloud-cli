package device

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/iostreams"
)

// diagnosisRecordsPath is the copilot REST endpoint backing `device history`.
const diagnosisRecordsPath = "/api/v1/copilot/diagnosis-records"

// shortRecordIDLen is the length of the short id shown as #xxxxxxxx in the
// assistant; record ids are lowercase UUIDs, so it is the first UUID segment.
const shortRecordIDLen = 8

var recordIDPattern = regexp.MustCompile(`^[0-9a-f][0-9a-f-]{7,35}$`)

var historyColumns = []string{"record_id", "created_at", "status", "outcome", "symptom", "root_cause", "fix"}

func NewCmdHistory(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "history",
		Short: "Assistant diagnosis records",
		Long: `View diagnosis records produced by the AI assistant: the symptom, root cause and
remediation of each diagnosis, and whether the user has confirmed the problem is resolved.`,
	}

	cmd.AddCommand(newCmdHistoryList(f))
	cmd.AddCommand(newCmdHistoryGet(f))

	return cmd
}

// normalizeRecordID accepts a full record id or a short id (optionally with a
// leading '#') and returns it lowercased.
func normalizeRecordID(raw string) (string, error) {
	id := strings.ToLower(strings.TrimPrefix(strings.TrimSpace(raw), "#"))
	if !recordIDPattern.MatchString(id) {
		return "", fmt.Errorf("invalid record id %q: use the full id or at least the %d-character short id (e.g. e95bba0b)", raw, shortRecordIDLen)
	}
	return id, nil
}

// rewrapRecords turns the copilot {"records": [...], "count": n} response into
// the {"result": [...]} envelope the table, --jq and structured output expect.
func rewrapRecords(body []byte) ([]byte, []json.RawMessage, error) {
	var resp struct {
		Records []json.RawMessage `json:"records"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, nil, fmt.Errorf("parsing diagnosis records: %w", err)
	}
	if resp.Records == nil {
		resp.Records = []json.RawMessage{}
	}
	out, err := json.Marshal(map[string]interface{}{"result": resp.Records})
	if err != nil {
		return nil, nil, err
	}
	return out, resp.Records, nil
}

func writeHistory(f *factory.Factory, cmd *cobra.Command, body []byte) error {
	output, _ := cmd.Flags().GetString("output")
	return iostreams.FormatOutput(body, f.IO, output,
		iostreams.WithColumns(historyColumns...),
		iostreams.WithFormatters(iostreams.ColumnFormatters{
			"record_id":  shortRecordID,
			"symptom":    truncateCell,
			"root_cause": truncateCell,
			"fix":        truncateCell,
		}),
	)
}

// truncateCell shortens long text for the table; a null field reaches column
// formatters as "<nil>" and renders empty.
func truncateCell(s string) string {
	if s == "<nil>" {
		return ""
	}
	return iostreams.TruncateRunes(s, 40)
}

func shortRecordID(s string) string {
	if len(s) > shortRecordIDLen {
		return s[:shortRecordIDLen]
	}
	return s
}
