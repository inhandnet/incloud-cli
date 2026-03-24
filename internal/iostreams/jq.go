package iostreams

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/itchyny/gojq"
)

// ApplyJQ executes a jq expression against JSON data and returns the results
// as newline-separated output. Each result is printed as raw text for strings,
// or compact JSON for other types.
func ApplyJQ(data []byte, expr string) (string, error) {
	query, err := gojq.Parse(expr)
	if err != nil {
		return "", fmt.Errorf("invalid jq expression: %w", err)
	}

	var input interface{}
	if err := json.Unmarshal(data, &input); err != nil {
		return "", fmt.Errorf("parsing JSON for jq: %w", err)
	}

	var sb strings.Builder
	iter := query.Run(input)
	for {
		v, ok := iter.Next()
		if !ok {
			break
		}
		if err, isErr := v.(error); isErr {
			return "", fmt.Errorf("jq: %w", err)
		}
		if s, isStr := v.(string); isStr {
			sb.WriteString(s)
		} else {
			b, err := json.Marshal(v)
			if err != nil {
				return "", fmt.Errorf("jq: marshaling result: %w", err)
			}
			sb.Write(b)
		}
		sb.WriteByte('\n')
	}
	return strings.TrimRight(sb.String(), "\n"), nil
}
