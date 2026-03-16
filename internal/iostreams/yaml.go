package iostreams

import (
	"encoding/json"
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

// FormatYAML converts JSON bytes to YAML string.
func FormatYAML(data []byte) (string, error) {
	var obj any
	if err := json.Unmarshal(data, &obj); err != nil {
		return "", fmt.Errorf("parsing JSON for YAML conversion: %w", err)
	}
	out, err := yaml.Marshal(obj)
	if err != nil {
		return "", fmt.Errorf("marshaling YAML: %w", err)
	}
	return strings.TrimRight(string(out), "\n"), nil
}
