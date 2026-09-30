package device

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
)

// historyGetLimit caps the records returned for one get: a short id may match
// more than one record.
const historyGetLimit = 50

func newCmdHistoryGet(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <record-id> [<record-id>...]",
		Short: "Get diagnosis records by id",
		Long: `Get the full content and current status of diagnosis records.

Each id may be the full record id or the 8-character short id shown in the
assistant (e.g. #e95bba0b). A short id matches every record whose id starts
with it.`,
		Example: `  # By short id
  incloud device history get e95bba0b

  # Several records, full fields as JSON
  incloud device history get e95bba0b 68343905 -o json`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ids := make([]string, 0, len(args))
			for _, raw := range args {
				id, err := normalizeRecordID(raw)
				if err != nil {
					return err
				}
				ids = append(ids, id)
			}

			client, err := f.APIClient()
			if err != nil {
				return err
			}

			q := url.Values{}
			for _, id := range ids {
				q.Add("record_ids", id)
			}
			q.Set("limit", strconv.Itoa(historyGetLimit))

			body, err := client.Get(diagnosisRecordsPath, q)
			if err != nil {
				return err
			}
			_, records, err := rewrapRecords(body)
			if err != nil {
				return err
			}
			// Keep only records matching a requested id, so a backend that
			// ignores record_ids never passes unrelated records off as results.
			records, missing := matchRecordIDs(ids, records)
			if len(records) == 0 {
				return fmt.Errorf("no diagnosis record found for %s", strings.Join(args, ", "))
			}
			if len(missing) > 0 {
				fmt.Fprintf(f.IO.ErrOut, "No diagnosis record found for %s\n", strings.Join(missing, ", "))
			}
			body, err = json.Marshal(map[string]interface{}{"result": records})
			if err != nil {
				return err
			}
			return writeHistory(f, cmd, body)
		},
	}

	return cmd
}

// matchRecordIDs returns the records whose id starts with one of the requested
// ids, and the requested ids no record matched.
func matchRecordIDs(ids []string, records []json.RawMessage) ([]json.RawMessage, []string) {
	matched := make([]json.RawMessage, 0, len(records))
	hit := make(map[string]bool, len(ids))
	for _, raw := range records {
		var r struct {
			RecordID string `json:"record_id"`
		}
		if json.Unmarshal(raw, &r) != nil {
			continue
		}
		keep := false
		for _, id := range ids {
			if strings.HasPrefix(r.RecordID, id) {
				hit[id] = true
				keep = true
			}
		}
		if keep {
			matched = append(matched, raw)
		}
	}
	var missing []string
	for _, id := range ids {
		if !hit[id] {
			missing = append(missing, id)
		}
	}
	return matched, missing
}
