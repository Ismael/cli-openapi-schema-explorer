package cmd

import (
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestComponents_ListTypes(t *testing.T) {
	out, err := executeCommand("--spec", testdataPath("multi-component-types.json"), "components")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "schemas") {
		t.Error("expected 'schemas' in component types")
	}
	if !strings.Contains(out, "parameters") {
		t.Error("expected 'parameters' in component types")
	}
}

func TestComponents_ListSchemaNames(t *testing.T) {
	out, err := executeCommand("--spec", testdataPath("sample-api.json"), "components", "schemas")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "User") {
		t.Error("expected 'User' in schema names")
	}
	if !strings.Contains(out, "UserList") {
		t.Error("expected 'UserList' in schema names")
	}
}

func TestComponents_FilterSchemaNames(t *testing.T) {
	out, err := executeCommand("--spec", testdataPath("complex-endpoint.json"), "components", "schemas", "--filter", "Task")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "Task") {
		t.Error("expected 'Task' in filtered results")
	}
	if !strings.Contains(out, "TaskList") {
		t.Error("expected 'TaskList' in filtered results")
	}
	if !strings.Contains(out, "CreateTaskRequest") {
		t.Error("expected 'CreateTaskRequest' in filtered results")
	}
}

func TestComponents_FilterNoMatch(t *testing.T) {
	out, err := executeCommand("--spec", testdataPath("sample-api.json"), "components", "schemas", "--filter", "zzz")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Should show hint but no names
	lines := strings.Split(strings.TrimSpace(out), "\n")
	// Only the hint line
	if len(lines) > 1 {
		t.Errorf("expected no matches, got: %s", out)
	}
}

func TestComponents_SchemaDetail(t *testing.T) {
	out, err := executeCommand("--spec", testdataPath("sample-api.json"), "components", "schemas", "User")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var schema map[string]any
	if err := yaml.Unmarshal([]byte(out), &schema); err != nil {
		t.Fatalf("output is not valid YAML: %v\noutput: %s", err, out)
	}
	if schema["type"] != "object" {
		t.Errorf("expected type 'object', got %v", schema["type"])
	}
}

func TestComponents_SchemaDetailResolvesRefs(t *testing.T) {
	out, err := executeCommand("--spec", testdataPath("sample-api.json"), "components", "schemas", "UserList")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var schema map[string]any
	if err := yaml.Unmarshal([]byte(out), &schema); err != nil {
		t.Fatalf("output is not valid YAML: %v\noutput: %s", err, out)
	}
	// UserList.properties.users.items should be resolved (not $ref)
	props := schema["properties"].(map[string]any)
	users := props["users"].(map[string]any)
	items := users["items"].(map[string]any)
	if _, hasRef := items["$ref"]; hasRef {
		t.Error("$ref in UserList.users.items should be resolved inline")
	}
	if items["type"] != "object" {
		t.Errorf("expected resolved User type 'object', got %v", items["type"])
	}
}

func TestComponents_MultipleSchemas(t *testing.T) {
	out, err := executeCommand("--spec", testdataPath("complex-endpoint.json"), "components", "schemas", "Task", "TaskList")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var schemas []any
	if err := yaml.Unmarshal([]byte(out), &schemas); err != nil {
		t.Fatalf("output is not valid YAML array: %v\noutput: %s", err, out)
	}
	if len(schemas) != 2 {
		t.Fatalf("expected 2 schemas, got %d", len(schemas))
	}
}

func TestComponents_UnknownType(t *testing.T) {
	out, err := executeCommand("--spec", testdataPath("sample-api.json"), "components", "foobar")
	if err == nil {
		t.Fatal("expected error for unknown component type")
	}
	if !strings.Contains(out, "Error:") {
		t.Error("expected error message in output")
	}
}

func TestComponents_UnknownName(t *testing.T) {
	out, err := executeCommand("--spec", testdataPath("sample-api.json"), "components", "schemas", "NonExistent")
	if err == nil {
		t.Fatal("expected error for unknown schema name")
	}
	if !strings.Contains(out, "Error:") {
		t.Error("expected error message in output")
	}
}

func TestComponents_ParameterType(t *testing.T) {
	out, err := executeCommand("--spec", testdataPath("multi-component-types.json"), "components", "parameters", "--filter", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "TraceId") {
		t.Error("expected 'TraceId' in parameter names")
	}
}

// Converted from: e2e/resources.test.ts — list schemas from complex endpoint
func TestComponents_ComplexEndpointSchemaList(t *testing.T) {
	out, err := executeCommand("--spec", testdataPath("complex-endpoint.json"), "components", "schemas", "--filter", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "CreateTaskRequest") {
		t.Error("expected 'CreateTaskRequest' in schema names")
	}
	if !strings.Contains(out, "Task") {
		t.Error("expected 'Task' in schema names")
	}
	if !strings.Contains(out, "TaskList") {
		t.Error("expected 'TaskList' in schema names")
	}
}

// Converted from: e2e/resources.test.ts — Task schema detail
func TestComponents_TaskSchemaDetail(t *testing.T) {
	out, err := executeCommand("--spec", testdataPath("complex-endpoint.json"), "components", "schemas", "Task")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var schema map[string]any
	if err := yaml.Unmarshal([]byte(out), &schema); err != nil {
		t.Fatalf("output is not valid YAML: %v\noutput: %s", err, out)
	}
	if schema["type"] != "object" {
		t.Errorf("expected type 'object', got %v", schema["type"])
	}
	props := schema["properties"].(map[string]any)
	id := props["id"].(map[string]any)
	if id["type"] != "string" {
		t.Errorf("expected id type 'string', got %v", id["type"])
	}
	title := props["title"].(map[string]any)
	if title["type"] != "string" {
		t.Errorf("expected title type 'string', got %v", title["type"])
	}
}

// Converted from: e2e/resources.test.ts — TaskList schema resolves refs
func TestComponents_TaskListResolvesRefs(t *testing.T) {
	out, err := executeCommand("--spec", testdataPath("complex-endpoint.json"), "components", "schemas", "TaskList")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var schema map[string]any
	if err := yaml.Unmarshal([]byte(out), &schema); err != nil {
		t.Fatalf("output is not valid YAML: %v\noutput: %s", err, out)
	}
	props := schema["properties"].(map[string]any)
	items := props["items"].(map[string]any)
	arrayItems := items["items"].(map[string]any)
	// Task ref should be resolved inline
	if _, hasRef := arrayItems["$ref"]; hasRef {
		t.Error("$ref in TaskList.items.items should be resolved inline")
	}
	if arrayItems["type"] != "object" {
		t.Error("expected Task to be resolved to an object")
	}
}

// Converted from: e2e/resources.test.ts — multiple schemas Task TaskList
func TestComponents_MultipleSchemasContent(t *testing.T) {
	out, err := executeCommand("--spec", testdataPath("complex-endpoint.json"), "components", "schemas", "Task", "TaskList")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var schemas []any
	if err := yaml.Unmarshal([]byte(out), &schemas); err != nil {
		t.Fatalf("output is not valid YAML array: %v\noutput: %s", err, out)
	}
	if len(schemas) != 2 {
		t.Fatalf("expected 2 schemas, got %d", len(schemas))
	}
	// Both should be objects
	for i, schemaAny := range schemas {
		schema := schemaAny.(map[string]any)
		if schema["type"] != "object" {
			t.Errorf("expected schema[%d] type 'object', got %v", i, schema["type"])
		}
	}
}

// Converted from: component-detail-handler.test.ts — single valid parameter detail
func TestComponents_ParameterDetail(t *testing.T) {
	out, err := executeCommand("--spec", testdataPath("multi-component-types.json"), "components", "parameters", "TraceId")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var param map[string]any
	if err := yaml.Unmarshal([]byte(out), &param); err != nil {
		t.Fatalf("output is not valid YAML: %v\noutput: %s", err, out)
	}
	if param["name"] != "X-Trace-ID" {
		t.Errorf("expected parameter name 'X-Trace-ID', got %v", param["name"])
	}
	if param["in"] != "header" {
		t.Errorf("expected parameter in 'header', got %v", param["in"])
	}
}

// Converted from: component-map-handler.test.ts — list names are sorted
func TestComponents_SchemaNamesSorted(t *testing.T) {
	out, err := executeCommand("--spec", testdataPath("sample-api.json"), "components", "schemas", "--filter", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	lines := strings.Split(strings.TrimSpace(out), "\n")
	// Find the lines with schema names (skip hint line)
	var names []string
	for _, line := range lines {
		if !strings.HasPrefix(line, "#") && line != "" {
			names = append(names, strings.TrimSpace(line))
		}
	}
	// Names should be sorted: User before UserList
	if len(names) >= 2 {
		for i := 1; i < len(names); i++ {
			if names[i] < names[i-1] {
				t.Errorf("names not sorted: %v comes after %v", names[i], names[i-1])
			}
		}
	}
}

// Converted from: component-map-handler.test.ts — component type listing shows hint
func TestComponents_TypeListShowsHint(t *testing.T) {
	out, err := executeCommand("--spec", testdataPath("multi-component-types.json"), "components")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "#") {
		t.Error("expected hint comment in component types output")
	}
}

// Converted from: component-detail-handler.test.ts — error includes available names
func TestComponents_UnknownNameShowsAvailable(t *testing.T) {
	out, err := executeCommand("--spec", testdataPath("sample-api.json"), "components", "schemas", "NonExistent")
	if err == nil {
		t.Fatal("expected error for unknown schema name")
	}
	if !strings.Contains(out, "Error:") {
		t.Error("expected error message")
	}
	if !strings.Contains(out, "NonExistent") {
		t.Error("expected error to mention the requested name")
	}
}

// Converted from: component-map-handler.test.ts — error for unknown type shows available types
func TestComponents_UnknownTypeShowsAvailable(t *testing.T) {
	out, err := executeCommand("--spec", testdataPath("sample-api.json"), "components", "foobar")
	if err == nil {
		t.Fatal("expected error for unknown component type")
	}
	if !strings.Contains(out, "Error:") {
		t.Error("expected error message")
	}
	if !strings.Contains(out, "foobar") {
		t.Error("expected error to mention the requested type")
	}
}

// Converted from: e2e/resources.test.ts — components type listing for single-type spec
func TestComponents_SingleTypeSpec(t *testing.T) {
	out, err := executeCommand("--spec", testdataPath("complex-endpoint.json"), "components")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "schemas") {
		t.Error("expected 'schemas' in component types")
	}
}

// Converted from: component-detail-handler.test.ts — User schema has expected properties
func TestComponents_UserSchemaProperties(t *testing.T) {
	out, err := executeCommand("--spec", testdataPath("sample-api.json"), "components", "schemas", "User")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var schema map[string]any
	if err := yaml.Unmarshal([]byte(out), &schema); err != nil {
		t.Fatalf("output is not valid YAML: %v\noutput: %s", err, out)
	}
	props := schema["properties"].(map[string]any)
	if _, ok := props["email"]; !ok {
		t.Error("expected 'email' property in User schema")
	}
	if _, ok := props["name"]; !ok {
		t.Error("expected 'name' property in User schema")
	}
	if _, ok := props["id"]; !ok {
		t.Error("expected 'id' property in User schema")
	}
}

// OpenAPI 3.1 spec — list schema names
func TestComponents_OpenAPI31SchemaList(t *testing.T) {
	out, err := executeCommand("--spec", testdataPath("openapi31-spec.json"), "components", "schemas", "--filter", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "Item") {
		t.Error("expected 'Item' in schema names")
	}
	if !strings.Contains(out, "ItemCreate") {
		t.Error("expected 'ItemCreate' in schema names")
	}
	if !strings.Contains(out, "ItemList") {
		t.Error("expected 'ItemList' in schema names")
	}
}

// OpenAPI 3.1 spec — schema detail with resolved refs
func TestComponents_OpenAPI31SchemaDetail(t *testing.T) {
	out, err := executeCommand("--spec", testdataPath("openapi31-spec.json"), "components", "schemas", "ItemList")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var schema map[string]any
	if err := yaml.Unmarshal([]byte(out), &schema); err != nil {
		t.Fatalf("output is not valid YAML: %v\noutput: %s", err, out)
	}
	if schema["type"] != "object" {
		t.Errorf("expected type 'object', got %v", schema["type"])
	}
	props := schema["properties"].(map[string]any)
	items := props["items"].(map[string]any)
	arrayItems := items["items"].(map[string]any)
	// Item ref should be resolved
	if _, hasRef := arrayItems["$ref"]; hasRef {
		t.Error("$ref in ItemList.items.items should be resolved")
	}
	if arrayItems["type"] != "object" {
		t.Error("expected Item to be resolved to an object")
	}
}

// OpenAPI 3.1 spec — Item schema has correct properties
func TestComponents_OpenAPI31ItemProperties(t *testing.T) {
	out, err := executeCommand("--spec", testdataPath("openapi31-spec.json"), "components", "schemas", "Item")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var schema map[string]any
	if err := yaml.Unmarshal([]byte(out), &schema); err != nil {
		t.Fatalf("output is not valid YAML: %v\noutput: %s", err, out)
	}
	props := schema["properties"].(map[string]any)
	if _, ok := props["id"]; !ok {
		t.Error("expected 'id' property")
	}
	if _, ok := props["name"]; !ok {
		t.Error("expected 'name' property")
	}
	if _, ok := props["rating"]; !ok {
		t.Error("expected 'rating' property")
	}
	if _, ok := props["tags"]; !ok {
		t.Error("expected 'tags' property")
	}
}

func TestSwagger2_ComponentDetail(t *testing.T) {
	out, err := executeCommand("--spec", testdataPath("sample-v2-api.json"), "components", "schemas", "Pong")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var schema map[string]any
	if err := yaml.Unmarshal([]byte(out), &schema); err != nil {
		t.Fatalf("output is not valid YAML: %v\noutput: %s", err, out)
	}
	if schema["type"] != "object" {
		t.Errorf("expected type 'object', got %v", schema["type"])
	}
}
