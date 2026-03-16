package iostreams

import (
	"strings"
	"testing"
)

func TestFormatYAML_Object(t *testing.T) {
	data := []byte(`{"name":"alice","age":30,"active":true}`)
	result, err := FormatYAML(data)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(result, "name: alice") {
		t.Errorf("expected 'name: alice', got:\n%s", result)
	}
	if !strings.Contains(result, "age: 30") {
		t.Errorf("expected 'age: 30', got:\n%s", result)
	}
	if !strings.Contains(result, "active: true") {
		t.Errorf("expected 'active: true', got:\n%s", result)
	}
}

func TestFormatYAML_Nested(t *testing.T) {
	data := []byte(`{"result":{"user":"admin","roles":["root","viewer"]}}`)
	result, err := FormatYAML(data)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(result, "user: admin") {
		t.Errorf("expected 'user: admin', got:\n%s", result)
	}
	if !strings.Contains(result, "- root") {
		t.Errorf("expected '- root' in roles list, got:\n%s", result)
	}
}

func TestFormatYAML_InvalidJSON(t *testing.T) {
	_, err := FormatYAML([]byte(`not json`))
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}
