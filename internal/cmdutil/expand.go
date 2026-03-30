package cmdutil

import "strings"

// ExpandDesc generates a description for the --expand flag.
// If supported is non-empty, it includes the list of supported values.
func ExpandDesc(supported []string) string {
	desc := "Expand related resources"
	if len(supported) > 0 {
		desc += ". Supported: " + strings.Join(supported, ", ")
	}
	return desc
}
