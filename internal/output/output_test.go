package output

import (
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestFormatYAML_BasicObject(t *testing.T) {
	v := map[string]any{"type": "object", "title": "User"}
	result, err := FormatYAML(v)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// YAML output should be multi-line
	if !strings.Contains(result, "\n") {
		t.Error("YAML output for a multi-key object should contain newlines")
	}
	// Round-trip validation
	var parsed map[string]any
	if err := yaml.Unmarshal([]byte(result), &parsed); err != nil {
		t.Fatalf("output is not valid YAML: %v", err)
	}
	if parsed["type"] != "object" {
		t.Errorf("expected type 'object', got %v", parsed["type"])
	}
	if parsed["title"] != "User" {
		t.Errorf("expected title 'User', got %v", parsed["title"])
	}
}

func TestFormatYAML_ComplexData(t *testing.T) {
	v := map[string]any{
		"method":  "GET",
		"path":    "/test",
		"summary": "Test endpoint",
		"parameters": []any{
			map[string]any{
				"name":     "id",
				"in":       "path",
				"required": true,
				"schema":   map[string]any{"type": "string"},
			},
		},
	}
	result, err := FormatYAML(v)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Round-trip validation
	var parsed map[string]any
	if err := yaml.Unmarshal([]byte(result), &parsed); err != nil {
		t.Fatalf("output is not valid YAML: %v", err)
	}
	if parsed["method"] != "GET" {
		t.Errorf("expected method 'GET', got %v", parsed["method"])
	}
}

func TestFormatYAML_NilValue(t *testing.T) {
	result, err := FormatYAML(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// nil marshals to "null" in YAML
	if strings.TrimSpace(result) != "null" {
		t.Errorf("expected 'null', got %q", result)
	}
}

func TestFormatYAML_NoTrailingNewline(t *testing.T) {
	v := map[string]any{"key": "value"}
	result, err := FormatYAML(v)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.HasSuffix(result, "\n") {
		t.Error("FormatYAML should trim trailing newline")
	}
}

func TestFormatList(t *testing.T) {
	items := []string{"/users", "/organizations", "/projects"}
	hint := "Use: openapi-explorer --spec <spec> paths <path> to explore"
	result := FormatList(items, hint)

	lines := strings.Split(result, "\n")
	if !strings.HasPrefix(lines[0], "# ") {
		t.Errorf("expected first line to be a comment, got %q", lines[0])
	}
	if lines[1] != "/users" {
		t.Errorf("expected second line '/users', got %q", lines[1])
	}
}

func TestFormatList_Empty(t *testing.T) {
	result := FormatList([]string{}, "hint")
	if !strings.Contains(result, "# hint") {
		t.Error("empty list should still show hint")
	}
}

func TestFormatError(t *testing.T) {
	result := FormatError("path not found")
	if result != "Error: path not found" {
		t.Errorf("expected 'Error: path not found', got %q", result)
	}
}

// Converted from: formatters.test.ts — FormatList sorted items
func TestFormatList_Sorted(t *testing.T) {
	items := []string{"Error", "User"}
	hint := "Available schemas"
	result := FormatList(items, hint)
	lines := strings.Split(result, "\n")
	if len(lines) < 3 {
		t.Fatalf("expected at least 3 lines, got %d", len(lines))
	}
	if lines[1] != "Error" {
		t.Errorf("expected 'Error' on line 2, got %q", lines[1])
	}
	if lines[2] != "User" {
		t.Errorf("expected 'User' on line 3, got %q", lines[2])
	}
}
