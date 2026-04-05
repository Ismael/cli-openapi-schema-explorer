package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func absTestdataPath(name string) string {
	abs, _ := filepath.Abs(filepath.Join("..", "testdata", name))
	return abs
}

func TestConfig_FileProvideSpec(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, ".openapi-explorer")
	os.WriteFile(configPath, []byte("spec: "+absTestdataPath("sample-api.json")+"\n"), 0644)

	orig, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(orig)

	out, err := executeCommand("info")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "Sample API") {
		t.Errorf("expected config file to provide spec, got: %s", out)
	}
}

func TestConfig_FlagOverridesFile(t *testing.T) {
	complexSpecPath := absTestdataPath("complex-endpoint.json")

	dir := t.TempDir()
	configPath := filepath.Join(dir, ".openapi-explorer")
	os.WriteFile(configPath, []byte("spec: "+absTestdataPath("sample-api.json")+"\n"), 0644)

	orig, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(orig)

	out, err := executeCommand("--spec", complexSpecPath, "info")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "Complex Endpoint Test API") {
		t.Errorf("expected --spec to override config file, got: %s", out)
	}
}

func TestConfig_MissingBothErrors(t *testing.T) {
	dir := t.TempDir()
	orig, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(orig)

	out, err := executeCommand("info")
	if err == nil {
		t.Fatal("expected error when no spec and no config")
	}
	if !strings.Contains(out, "no spec provided") {
		t.Errorf("expected clear error message, got: %s", out)
	}
}
