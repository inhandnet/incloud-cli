package connector

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/api"
	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/iostreams"
	"github.com/inhandnet/incloud-cli/internal/ui"
)

// resultIDName extracts _id and name from {"result": {...}} response.
func resultIDName(body []byte) (id, name string) {
	var resp struct {
		Result struct {
			ID   string `json:"_id"`
			Name string `json:"name"`
		} `json:"result"`
	}
	_ = json.Unmarshal(body, &resp)
	return resp.Result.ID, resp.Result.Name
}

// formatOutput is a shorthand for FormatOutput with the --output flag.
func formatOutput(cmd *cobra.Command, io *iostreams.IOStreams, body []byte, fields []string) error {
	output, _ := cmd.Flags().GetString("output")
	return iostreams.FormatOutput(body, io, output, fields)
}

// writeCreated writes a "<resource> created" confirmation to stderr.
func writeCreated(f *factory.Factory, resource string, body []byte) {
	id, name := resultIDName(body)
	fmt.Fprintf(f.IO.ErrOut, "%s %q created. (id: %s)\n", resource, name, id)
}

// writeUpdated writes a "<resource> updated" confirmation to stderr.
func writeUpdated(f *factory.Factory, resource string, body []byte) {
	id, name := resultIDName(body)
	fmt.Fprintf(f.IO.ErrOut, "%s %q (%s) updated.\n", resource, name, id)
}

// writeDeleted writes a "<resource> deleted" confirmation to stderr.
func writeDeleted(f *factory.Factory, resource, name, id string) {
	fmt.Fprintf(f.IO.ErrOut, "%s %q (%s) deleted.\n", resource, name, id)
}

// deleteConnectorResources handles single and bulk delete for connector resources.
// singleBasePath is the base for GET and DELETE (e.g. "/api/v1/connectors").
// bulkPath is the POST endpoint for bulk delete (e.g. "/api/v1/connectors/bulk/delete").
func deleteConnectorResources(f *factory.Factory, client *api.APIClient, ids []string, yes bool, resource, singleBasePath, bulkPath string) error {
	// Collect names for confirmation
	type entry struct {
		id   string
		name string
	}
	entries := make([]entry, 0, len(ids))
	for _, id := range ids {
		body, err := client.Get(singleBasePath+"/"+id, nil)
		if err != nil {
			return fmt.Errorf("%s %s not found", resource, id)
		}
		_, name := resultIDName(body)
		if name == "" {
			name = id
		}
		entries = append(entries, entry{id: id, name: name})
	}

	if !yes {
		var prompt string
		if len(entries) == 1 {
			prompt = fmt.Sprintf("Delete %s %q (%s)?", resource, entries[0].name, entries[0].id)
		} else {
			prompt = fmt.Sprintf("Delete %d resources (%s)?", len(entries), resource)
		}
		confirmed, err := ui.Confirm(f, prompt)
		if err != nil {
			return err
		}
		if !confirmed {
			return nil
		}
	}

	if len(ids) == 1 {
		_, err := client.Delete(singleBasePath + "/" + ids[0])
		if err != nil {
			return err
		}
	} else {
		_, err := client.Post(bulkPath, map[string]interface{}{"ids": ids})
		if err != nil {
			return err
		}
	}

	for _, e := range entries {
		writeDeleted(f, resource, e.name, e.id)
	}
	return nil
}
