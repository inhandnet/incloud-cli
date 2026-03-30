package org

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
)

// parseLabels converts ["key=value", ...] into [{"key":"key","value":"value"}, ...]
func parseLabels(pairs []string) ([]map[string]string, error) {
	labels := make([]map[string]string, 0, len(pairs))
	for _, pair := range pairs {
		parts := strings.SplitN(pair, "=", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid label: %s (expected key=value)", pair)
		}
		labels = append(labels, map[string]string{
			"key":   parts[0],
			"value": parts[1],
		})
	}
	return labels, nil
}

// defaultListFields defines the default table columns for org list/self.
// Org responses contain 20+ fields; only show the most useful ones by default.
var defaultListFields = []string{
	"_id", "name", "email", "countryCode", "bizCategory",
	"active", "userCount", "deviceCount", "createdAt",
}

func NewCmdOrg(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "org",
		Short: "Manage organizations",
		Long:  "List, create, update, delete, and inspect organizations on the InCloud platform.",
	}

	cmd.AddCommand(NewCmdSelf(f))
	cmd.AddCommand(NewCmdUpdateSelf(f))
	cmd.AddCommand(NewCmdList(f))
	cmd.AddCommand(NewCmdGet(f))
	cmd.AddCommand(NewCmdCreate(f))
	cmd.AddCommand(NewCmdUpdate(f))
	cmd.AddCommand(NewCmdDelete(f))

	return cmd
}
