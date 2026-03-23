package connector

import (
	"encoding/json"
	"fmt"
	"net/url"

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

// lookupNamesByList fetches the list endpoint with ids filter and builds an id→name map.
func lookupNamesByList(client *api.APIClient, listPath string, ids []string) (map[string]string, error) {
	body, err := client.Get(listPath, url.Values{
		"fields": []string{"_id,name"},
		"ids":    ids,
	})
	if err != nil {
		return nil, err
	}
	var resp struct {
		Result []struct {
			ID   string `json:"_id"`
			Name string `json:"name"`
		} `json:"result"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, err
	}
	wanted := make(map[string]struct{}, len(ids))
	for _, id := range ids {
		wanted[id] = struct{}{}
	}
	nameMap := make(map[string]string, len(ids))
	for _, r := range resp.Result {
		if _, ok := wanted[r.ID]; ok {
			nameMap[r.ID] = r.Name
		}
	}
	return nameMap, nil
}

// deleteConnectorResources handles single and bulk delete for connector resources.
// basePath is the base for list/GET/DELETE (e.g. "/api/v1/connectors").
// bulkPath is the POST endpoint for bulk delete (e.g. "/api/v1/connectors/bulk/delete").
// When useListLookup is true, names are resolved via the list endpoint instead of
// GET single — use this for sub-resources that lack a GET-by-ID endpoint.
func deleteConnectorResources(f *factory.Factory, client *api.APIClient, ids []string, yes bool, resource, basePath, bulkPath string, useListLookup bool) error {
	nameMap := make(map[string]string, len(ids))

	if useListLookup {
		listNames, err := lookupNamesByList(client, basePath, ids)
		if err != nil {
			return fmt.Errorf("failed to look up %s: %w", resource, err)
		}
		for _, id := range ids {
			if name, ok := listNames[id]; ok {
				nameMap[id] = name
			} else {
				return fmt.Errorf("%s %s not found", resource, id)
			}
		}
	} else {
		for _, id := range ids {
			body, err := client.Get(basePath+"/"+id, nil)
			if err != nil {
				return fmt.Errorf("%s %s not found", resource, id)
			}
			if _, n := resultIDName(body); n != "" {
				nameMap[id] = n
			}
		}
	}

	if !yes {
		var prompt string
		if len(ids) == 1 {
			name := nameMap[ids[0]]
			if name == "" {
				name = ids[0]
			}
			prompt = fmt.Sprintf("Delete %s %q (%s)?", resource, name, ids[0])
		} else {
			prompt = fmt.Sprintf("Delete %d resources (%s)?", len(ids), resource)
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
		_, err := client.Delete(basePath + "/" + ids[0])
		if err != nil {
			return err
		}
	} else {
		_, err := client.Post(bulkPath, map[string]interface{}{"ids": ids})
		if err != nil {
			return err
		}
	}

	for _, id := range ids {
		name := nameMap[id]
		if name == "" {
			name = id
		}
		writeDeleted(f, resource, name, id)
	}
	return nil
}
