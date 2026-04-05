package output

import (
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

// FormatYAML marshals v to YAML.
func FormatYAML(v any) (string, error) {
	data, err := yaml.Marshal(v)
	if err != nil {
		return "", fmt.Errorf("failed to marshal YAML: %w", err)
	}
	return strings.TrimRight(string(data), "\n"), nil
}

// FormatList formats a list of items as plain text with a hint comment.
func FormatList(items []string, hint string) string {
	var sb strings.Builder
	sb.WriteString("# ")
	sb.WriteString(hint)
	sb.WriteString("\n")
	for _, item := range items {
		sb.WriteString(item)
		sb.WriteString("\n")
	}
	return strings.TrimRight(sb.String(), "\n")
}

// FormatError formats an error message.
func FormatError(msg string) string {
	return fmt.Sprintf("Error: %s", msg)
}
