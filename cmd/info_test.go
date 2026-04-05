package cmd

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"gopkg.in/yaml.v3"
)

func testdataPath(name string) string {
	return filepath.Join("..", "testdata", name)
}

func executeCommand(args ...string) (string, error) {
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs(args)
	// Reset all command flags to defaults before each execution to avoid state leaking between tests.
	resetCmdFlags(rootCmd)
	err := rootCmd.Execute()
	return buf.String(), err
}

// resetCmdFlags resets all flags on cmd and its subcommands to their default values,
// clearing the "changed" state so that subsequent executions start fresh.
func resetCmdFlags(cmd *cobra.Command) {
	specPath = ""
	cmd.Flags().VisitAll(func(f *pflag.Flag) {
		f.Changed = false
		f.Value.Set(f.DefValue) //nolint
	})
	for _, sub := range cmd.Commands() {
		resetCmdFlags(sub)
	}
}

func TestInfo_SampleAPI(t *testing.T) {
	out, err := executeCommand("--spec", testdataPath("sample-api.json"), "info")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var info map[string]any
	if err := yaml.Unmarshal([]byte(out), &info); err != nil {
		t.Fatalf("output is not valid YAML: %v\noutput: %s", err, out)
	}
	if info["title"] != "Sample API" {
		t.Errorf("expected title 'Sample API', got %v", info["title"])
	}
	if info["version"] != "1.0.0" {
		t.Errorf("expected version '1.0.0', got %v", info["version"])
	}
}

// Converted from: e2e/resources.test.ts — retrieve info for complex endpoint
func TestInfo_ComplexEndpoint(t *testing.T) {
	out, err := executeCommand("--spec", testdataPath("complex-endpoint.json"), "info")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var info map[string]any
	if err := yaml.Unmarshal([]byte(out), &info); err != nil {
		t.Fatalf("output is not valid YAML: %v\noutput: %s", err, out)
	}
	if info["title"] != "Complex Endpoint Test API" {
		t.Errorf("expected title 'Complex Endpoint Test API', got %v", info["title"])
	}
	if info["version"] != "1.0.0" {
		t.Errorf("expected version '1.0.0', got %v", info["version"])
	}
}

// Converted from: top-level-field-handler.test.ts — info field returns title and version
func TestInfo_HasTitleAndVersion(t *testing.T) {
	out, err := executeCommand("--spec", testdataPath("sample-api.json"), "info")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var info map[string]any
	if err := yaml.Unmarshal([]byte(out), &info); err != nil {
		t.Fatalf("output is not valid YAML: %v\noutput: %s", err, out)
	}
	if _, ok := info["title"]; !ok {
		t.Error("expected 'title' key in info output")
	}
	if _, ok := info["version"]; !ok {
		t.Error("expected 'version' key in info output")
	}
}

// Converted from: top-level-field-handler.test.ts — empty spec info
func TestInfo_EmptySpec(t *testing.T) {
	out, err := executeCommand("--spec", testdataPath("empty-api.json"), "info")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var info map[string]any
	if err := yaml.Unmarshal([]byte(out), &info); err != nil {
		t.Fatalf("output is not valid YAML: %v\noutput: %s", err, out)
	}
	if info["title"] != "Empty API" {
		t.Errorf("expected title 'Empty API', got %v", info["title"])
	}
}

// Converted from: spec-loader.test.ts — error for invalid spec path
func TestInfo_InvalidSpecPath(t *testing.T) {
	out, err := executeCommand("--spec", "/nonexistent/spec.json", "info")
	if err == nil {
		t.Fatal("expected error for nonexistent spec file")
	}
	if !strings.Contains(out, "Error:") {
		t.Error("expected error message in output")
	}
	if !strings.Contains(out, "no such file") {
		t.Errorf("expected 'no such file' in error, got: %s", out)
	}
}

// Test OpenAPI 3.1 spec with numeric exclusiveMinimum/exclusiveMaximum
func TestInfo_OpenAPI31(t *testing.T) {
	out, err := executeCommand("--spec", testdataPath("openapi31-spec.json"), "info")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var info map[string]any
	if err := yaml.Unmarshal([]byte(out), &info); err != nil {
		t.Fatalf("output is not valid YAML: %v\noutput: %s", err, out)
	}
	if info["title"] != "OpenAPI 3.1 Test API" {
		t.Errorf("expected title 'OpenAPI 3.1 Test API', got %v", info["title"])
	}
	if info["description"] != "A spec using OpenAPI 3.1 / JSON Schema 2020-12 features" {
		t.Errorf("expected description, got %v", info["description"])
	}
}

func TestInfo_Swagger2(t *testing.T) {
	out, err := executeCommand("--spec", testdataPath("sample-v2-api.json"), "info")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var info map[string]any
	if err := yaml.Unmarshal([]byte(out), &info); err != nil {
		t.Fatalf("output is not valid YAML: %v\noutput: %s", err, out)
	}
	if info["title"] != "Simple Swagger 2.0 API" {
		t.Errorf("expected title 'Simple Swagger 2.0 API', got %v", info["title"])
	}
}
