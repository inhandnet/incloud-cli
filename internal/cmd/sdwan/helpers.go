package sdwan

import (
	"fmt"

	"github.com/inhandnet/incloud-cli/internal/api"
	"github.com/inhandnet/incloud-cli/internal/cmdutil"
	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/ui"
)

const apiBase = "/api/v1/autovpn"

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
		if _, n := api.ResultIDName(body); n != "" {
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

	cmdutil.WriteDeleted(f, resource, name, id)
	return nil
}
