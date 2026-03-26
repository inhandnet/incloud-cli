package oobm

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/inhandnet/incloud-cli/internal/api"
)

// parseServices converts a slice of "protocol:port[:usage]" strings into
// API-ready maps. It is shared by create, update, connect, and close commands.
func parseServices(strs []string) ([]map[string]any, error) {
	services := make([]map[string]any, 0, len(strs))
	for _, s := range strs {
		svc, err := parseService(s)
		if err != nil {
			return nil, err
		}
		services = append(services, svc)
	}
	return services, nil
}

// serviceLabel returns a human-readable label like "ssh:22" for error messages.
func serviceLabel(svc map[string]any) string {
	return fmt.Sprintf("%s:%v", svc["protocol"], svc["port"])
}

// oobmResource holds the fields needed for connect/close auto-resolution.
type oobmResource struct {
	ID       string           `json:"_id"`
	Name     string           `json:"name"`
	DeviceID string           `json:"deviceId"`
	Services []map[string]any `json:"services"`
}

// getOobmResource fetches a single OOBM resource by ID from the list API.
func getOobmResource(client *api.APIClient, id string) (*oobmResource, error) {
	q := make(url.Values)
	q.Set("page", "0")
	q.Set("limit", "100")

	for {
		body, err := client.Get("/api/v1/oobm/resources", q)
		if err != nil {
			return nil, err
		}

		var resp struct {
			Result []oobmResource `json:"result"`
			Total  int            `json:"total"`
			Page   int            `json:"page"`
			Limit  int            `json:"limit"`
		}
		if err := json.Unmarshal(body, &resp); err != nil {
			return nil, fmt.Errorf("parsing response: %w", err)
		}

		for i := range resp.Result {
			if resp.Result[i].ID == id {
				return &resp.Result[i], nil
			}
		}

		// Check if there are more pages.
		if (resp.Page+1)*resp.Limit >= resp.Total {
			break
		}
		q.Set("page", fmt.Sprintf("%d", resp.Page+1))
	}

	return nil, fmt.Errorf("OOBM resource %s not found", id)
}
