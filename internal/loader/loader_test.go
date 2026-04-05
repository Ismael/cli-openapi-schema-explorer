package loader

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func testdataPath(name string) string {
	return filepath.Join("..", "..", "testdata", name)
}

func TestLoadFromFile_OpenAPI3JSON(t *testing.T) {
	doc, err := Load(testdataPath("sample-api.json"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if doc.Info.Title != "Sample API" {
		t.Errorf("expected title 'Sample API', got %q", doc.Info.Title)
	}
	if doc.OpenAPI != "3.0.3" {
		t.Errorf("expected openapi '3.0.3', got %q", doc.OpenAPI)
	}
}

func TestLoadFromFile_Swagger2(t *testing.T) {
	doc, err := Load(testdataPath("sample-v2-api.json"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if doc.Info.Title != "Simple Swagger 2.0 API" {
		t.Errorf("expected title 'Simple Swagger 2.0 API', got %q", doc.Info.Title)
	}
	// After conversion, should have openapi 3.x field
	if doc.OpenAPI == "" {
		t.Error("expected openapi version to be set after conversion")
	}
	// definitions/Pong should be converted to components/schemas/Pong
	if doc.Components == nil || doc.Components.Schemas == nil {
		t.Fatal("expected components.schemas after conversion")
	}
	pong := doc.Components.Schemas["Pong"]
	if pong == nil {
		t.Error("expected Pong schema in components.schemas")
	}
}

func TestLoadFromURL(t *testing.T) {
	data, err := os.ReadFile(testdataPath("sample-api.json"))
	if err != nil {
		t.Fatalf("failed to read fixture: %v", err)
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write(data)
	}))
	defer srv.Close()

	doc, err := Load(srv.URL + "/spec.json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if doc.Info.Title != "Sample API" {
		t.Errorf("expected title 'Sample API', got %q", doc.Info.Title)
	}
}

func TestLoadFromFile_NotFound(t *testing.T) {
	_, err := Load("/nonexistent/file.json")
	if err == nil {
		t.Fatal("expected error for nonexistent file")
	}
}

func TestLoadFromFile_Invalid(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "bad.json")
	os.WriteFile(path, []byte("not valid json or yaml"), 0644)
	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for invalid content")
	}
}

func TestLoadFromFile_NumericExclusiveMinMax(t *testing.T) {
	// Spec using JSON Schema 2020-12 style numeric exclusiveMinimum/exclusiveMaximum
	spec := `{
		"openapi": "3.0.3",
		"info": {"title": "Numeric ExclusiveMin API", "version": "1.0.0"},
		"paths": {},
		"components": {
			"schemas": {
				"Rating": {
					"type": "integer",
					"exclusiveMinimum": 0,
					"exclusiveMaximum": 100
				}
			}
		}
	}`
	tmp := t.TempDir()
	path := filepath.Join(tmp, "numeric-exclusive.json")
	os.WriteFile(path, []byte(spec), 0644)

	doc, err := Load(path)
	if err != nil {
		t.Fatalf("expected spec with numeric exclusiveMinimum/Maximum to load, got: %v", err)
	}
	if doc.Info.Title != "Numeric ExclusiveMin API" {
		t.Errorf("expected title 'Numeric ExclusiveMin API', got %q", doc.Info.Title)
	}
}

func TestLoadFromFile_OpenAPI31Spec(t *testing.T) {
	doc, err := Load(testdataPath("openapi31-spec.json"))
	if err != nil {
		t.Fatalf("expected OpenAPI 3.1 spec with numeric exclusiveMin/Max to load, got: %v", err)
	}
	if doc.Info.Title != "OpenAPI 3.1 Test API" {
		t.Errorf("expected title 'OpenAPI 3.1 Test API', got %q", doc.Info.Title)
	}
	// Verify paths loaded
	if doc.Paths == nil || doc.Paths.Len() != 2 {
		t.Errorf("expected 2 paths, got %d", doc.Paths.Len())
	}
	// Verify schemas loaded
	if doc.Components == nil || len(doc.Components.Schemas) != 3 {
		t.Errorf("expected 3 schemas, got %d", len(doc.Components.Schemas))
	}
}

func TestLoadFromURL_NumericExclusiveMinMax(t *testing.T) {
	// Serve an OpenAPI 3.1 spec with numeric exclusiveMinimum from a test server
	data, err := os.ReadFile(testdataPath("openapi31-spec.json"))
	if err != nil {
		t.Fatalf("failed to read fixture: %v", err)
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write(data)
	}))
	defer srv.Close()

	doc, err := Load(srv.URL + "/openapi.json")
	if err != nil {
		t.Fatalf("expected URL-loaded OpenAPI 3.1 spec to parse, got: %v", err)
	}
	if doc.Info.Title != "OpenAPI 3.1 Test API" {
		t.Errorf("expected title 'OpenAPI 3.1 Test API', got %q", doc.Info.Title)
	}
}

func TestLoadFromFile_EmptySpec(t *testing.T) {
	doc, err := Load(testdataPath("empty-api.json"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if doc.Info.Title != "Empty API" {
		t.Errorf("expected title 'Empty API', got %q", doc.Info.Title)
	}
}
