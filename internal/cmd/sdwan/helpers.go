package sdwan

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/api"
	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/iostreams"
	"github.com/inhandnet/incloud-cli/internal/ui"
)

const apiBase = "/api/v1/autovpn"

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
func formatOutput(cmd *cobra.Command, io *iostreams.IOStreams, body []byte) error {
	output, _ := cmd.Flags().GetString("output")
	return iostreams.FormatOutput(body, io, output)
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

// toMembers converts a slice of device IDs into NetworkMember payloads.
func toMembers(ids []string) []map[string]interface{} {
	members := make([]map[string]interface{}, len(ids))
	for i, id := range ids {
		members[i] = map[string]interface{}{"deviceId": id}
	}
	return members
}

// deleteResource handles single resource delete with confirmation.
func deleteResource(f *factory.Factory, client *api.APIClient, id string, yes bool, resource, basePath string) error {
	name := id
	if !yes {
		body, err := client.Get(basePath+"/"+id, nil)
		if err != nil {
			return fmt.Errorf("%s %s not found", resource, id)
		}
		if _, n := resultIDName(body); n != "" {
			name = n
		}

		prompt := fmt.Sprintf("Delete %s %q (%s)?", resource, name, id)
		confirmed, err := ui.Confirm(f, prompt)
		if err != nil {
			return err
		}
		if !confirmed {
			return nil
		}
	}

	_, err := client.Delete(basePath + "/" + id)
	if err != nil {
		return err
	}

	writeDeleted(f, resource, name, id)
	return nil
}
