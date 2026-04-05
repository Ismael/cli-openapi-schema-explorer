package cmd

import (
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestPaths_ListDepth1(t *testing.T) {
	out, err := executeCommand("--spec", testdataPath("paths-test.json"), "paths", "--depth", "1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// paths-test.json has: /project/..., /article/..., /sub/...
	// depth 1 should show /project, /article, /sub
	if !strings.Contains(out, "/project") {
		t.Error("expected /project in output")
	}
	if !strings.Contains(out, "/article") {
		t.Error("expected /article in output")
	}
	if !strings.Contains(out, "/sub") {
		t.Error("expected /sub in output")
	}
	// Should NOT show full paths at depth 1
	if strings.Contains(out, "/project/tasks") {
		t.Error("depth 1 should not show /project/tasks")
	}
}

func TestPaths_ListDepth2(t *testing.T) {
	out, err := executeCommand("--spec", testdataPath("paths-test.json"), "paths", "--depth", "2")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "/project/tasks") {
		t.Error("expected /project/tasks in depth 2 output")
	}
	if !strings.Contains(out, "/article/{articleId}") {
		t.Error("expected /article/{articleId} in depth 2 output")
	}
}

func TestPaths_ListMethods(t *testing.T) {
	out, err := executeCommand("--spec", testdataPath("sample-api.json"), "paths", "/users")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(strings.ToUpper(out), "GET") {
		t.Error("expected GET method listed for /users")
	}
	if !strings.Contains(out, "List users") {
		t.Error("expected summary 'List users' in output")
	}
}

func TestPaths_OperationDetail(t *testing.T) {
	out, err := executeCommand("--spec", testdataPath("complex-endpoint.json"), "paths", "/api/v1/organizations/{orgId}/projects/{projectId}/tasks", "get")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var op map[string]any
	if err := yaml.Unmarshal([]byte(out), &op); err != nil {
		t.Fatalf("output is not valid YAML: %v\noutput: %s", err, out)
	}
	if op["summary"] != "Get Tasks" {
		t.Errorf("expected summary 'Get Tasks', got %v", op["summary"])
	}
	// In shallow mode (default), $ref in response schema should be PRESERVED
	responses := op["responses"].(map[string]any)
	resp200 := responses["200"].(map[string]any)
	content := resp200["content"].(map[string]any)
	appJSON := content["application/json"].(map[string]any)
	schema := appJSON["schema"].(map[string]any)
	if _, hasRef := schema["$ref"]; !hasRef {
		t.Error("$ref in response schema should be preserved in shallow mode")
	}
}

func TestPaths_OperationDetailWithResolve(t *testing.T) {
	out, err := executeCommand("--spec", testdataPath("complex-endpoint.json"), "paths", "/api/v1/organizations/{orgId}/projects/{projectId}/tasks", "get", "--resolve")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var op map[string]any
	if err := yaml.Unmarshal([]byte(out), &op); err != nil {
		t.Fatalf("output is not valid YAML: %v\noutput: %s", err, out)
	}
	if op["summary"] != "Get Tasks" {
		t.Errorf("expected summary 'Get Tasks', got %v", op["summary"])
	}
	// With --resolve, $ref in response should be inlined
	responses := op["responses"].(map[string]any)
	resp200 := responses["200"].(map[string]any)
	content := resp200["content"].(map[string]any)
	appJSON := content["application/json"].(map[string]any)
	schema := appJSON["schema"].(map[string]any)
	if _, hasRef := schema["$ref"]; hasRef {
		t.Error("$ref should be resolved inline with --resolve flag")
	}
	if schema["type"] != "object" {
		t.Error("expected TaskList to be resolved to an object with --resolve flag")
	}
	// TaskList.items.$ref -> Task should also be resolved
	props := schema["properties"].(map[string]any)
	items := props["items"].(map[string]any)
	itemsItems := items["items"].(map[string]any)
	if _, hasRef := itemsItems["$ref"]; hasRef {
		t.Error("nested $ref (Task in TaskList.items) should be resolved inline with --resolve flag")
	}
}

func TestPaths_MultipleMethods(t *testing.T) {
	out, err := executeCommand("--spec", testdataPath("complex-endpoint.json"), "paths", "/api/v1/organizations/{orgId}/projects/{projectId}/tasks", "get", "post")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var ops []any
	if err := yaml.Unmarshal([]byte(out), &ops); err != nil {
		t.Fatalf("output is not valid YAML array: %v\noutput: %s", err, out)
	}
	if len(ops) != 2 {
		t.Fatalf("expected 2 operations, got %d", len(ops))
	}
}

func TestPaths_NotFound(t *testing.T) {
	out, err := executeCommand("--spec", testdataPath("sample-api.json"), "paths", "/nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent path")
	}
	if !strings.Contains(out, "Error:") {
		t.Error("expected error message in output")
	}
}

func TestPaths_InvalidMethod(t *testing.T) {
	out, err := executeCommand("--spec", testdataPath("sample-api.json"), "paths", "/users", "delete")
	if err == nil {
		t.Fatal("expected error for invalid method")
	}
	if !strings.Contains(out, "Error:") {
		t.Error("expected error message in output")
	}
}

func TestPaths_EmptySpec(t *testing.T) {
	out, err := executeCommand("--spec", testdataPath("empty-api.json"), "paths")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Should show hint but no paths
	if !strings.Contains(out, "#") {
		t.Error("expected hint comment even for empty paths")
	}
}

func TestPaths_SubPathDiscovery(t *testing.T) {
	// paths-test.json: /project/tasks/{taskId}
	// Querying /project should show sub-paths
	out, err := executeCommand("--spec", testdataPath("paths-test.json"), "paths", "/project")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "/project/tasks") {
		t.Error("expected /project/tasks as sub-path")
	}
}

// Converted from: e2e/resources.test.ts — GET operation detail includes operationId
func TestPaths_OperationDetailOperationId(t *testing.T) {
	out, err := executeCommand("--spec", testdataPath("complex-endpoint.json"), "paths", "/api/v1/organizations/{orgId}/projects/{projectId}/tasks", "get")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var op map[string]any
	if err := yaml.Unmarshal([]byte(out), &op); err != nil {
		t.Fatalf("output is not valid YAML: %v\noutput: %s", err, out)
	}
	if op["operationId"] != "getProjectTasks" {
		t.Errorf("expected operationId 'getProjectTasks', got %v", op["operationId"])
	}
}

// Converted from: e2e/resources.test.ts — POST operation detail includes operationId
func TestPaths_PostOperationDetail(t *testing.T) {
	out, err := executeCommand("--spec", testdataPath("complex-endpoint.json"), "paths", "/api/v1/organizations/{orgId}/projects/{projectId}/tasks", "post")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var op map[string]any
	if err := yaml.Unmarshal([]byte(out), &op); err != nil {
		t.Fatalf("output is not valid YAML: %v\noutput: %s", err, out)
	}
	if op["operationId"] != "createProjectTask" {
		t.Errorf("expected operationId 'createProjectTask', got %v", op["operationId"])
	}
	if op["summary"] != "Create Task" {
		t.Errorf("expected summary 'Create Task', got %v", op["summary"])
	}
}

// Converted from: e2e/resources.test.ts — multiple methods returns distinct operations
func TestPaths_MultipleMethodsDistinct(t *testing.T) {
	out, err := executeCommand("--spec", testdataPath("complex-endpoint.json"), "paths", "/api/v1/organizations/{orgId}/projects/{projectId}/tasks", "get", "post")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var ops []any
	if err := yaml.Unmarshal([]byte(out), &ops); err != nil {
		t.Fatalf("output is not valid YAML array: %v\noutput: %s", err, out)
	}
	if len(ops) != 2 {
		t.Fatalf("expected 2 operations, got %d", len(ops))
	}
	// Check that both operations have distinct operationIds
	ids := map[string]bool{}
	for _, opAny := range ops {
		op := opAny.(map[string]any)
		if id, ok := op["operationId"].(string); ok {
			ids[id] = true
		}
	}
	if !ids["getProjectTasks"] {
		t.Error("expected getProjectTasks in results")
	}
	if !ids["createProjectTask"] {
		t.Error("expected createProjectTask in results")
	}
}

// Converted from: operation-handler.test.ts — case-insensitive method lookup
func TestPaths_CaseInsensitiveMethod(t *testing.T) {
	out, err := executeCommand("--spec", testdataPath("sample-api.json"), "paths", "/users", "GET")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var op map[string]any
	if err := yaml.Unmarshal([]byte(out), &op); err != nil {
		t.Fatalf("output is not valid YAML: %v\noutput: %s", err, out)
	}
	if op["summary"] != "List users" {
		t.Errorf("expected summary 'List users', got %v", op["summary"])
	}
}

// Converted from: path-item-handler.test.ts — list methods includes METHOD: Summary format
func TestPaths_ListMethodsFormat(t *testing.T) {
	out, err := executeCommand("--spec", testdataPath("complex-endpoint.json"), "paths", "/api/v1/organizations/{orgId}/projects/{projectId}/tasks")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "GET") {
		t.Error("expected GET in methods list")
	}
	if !strings.Contains(out, "POST") {
		t.Error("expected POST in methods list")
	}
	if !strings.Contains(out, "Get Tasks") {
		t.Error("expected 'Get Tasks' summary")
	}
	if !strings.Contains(out, "Create Task") {
		t.Error("expected 'Create Task' summary")
	}
}

// Converted from: operation-handler.test.ts — operation detail preserves requestBody refs in shallow mode
func TestPaths_OperationDetailResolvesRequestBody(t *testing.T) {
	out, err := executeCommand("--spec", testdataPath("complex-endpoint.json"), "paths", "/api/v1/organizations/{orgId}/projects/{projectId}/tasks", "post")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var op map[string]any
	if err := yaml.Unmarshal([]byte(out), &op); err != nil {
		t.Fatalf("output is not valid YAML: %v\noutput: %s", err, out)
	}
	// In shallow mode (default), $ref in requestBody schema should be PRESERVED
	rb := op["requestBody"].(map[string]any)
	content := rb["content"].(map[string]any)
	appJSON := content["application/json"].(map[string]any)
	schema := appJSON["schema"].(map[string]any)
	if _, hasRef := schema["$ref"]; !hasRef {
		t.Error("$ref in requestBody schema should be preserved in shallow mode")
	}
}

// Converted from: e2e/resources.test.ts — error message for invalid method includes available methods
func TestPaths_InvalidMethodShowsAvailable(t *testing.T) {
	out, err := executeCommand("--spec", testdataPath("complex-endpoint.json"), "paths", "/api/v1/organizations/{orgId}/projects/{projectId}/tasks", "put")
	if err == nil {
		t.Fatal("expected error for invalid method")
	}
	if !strings.Contains(out, "Error:") {
		t.Error("expected error message")
	}
	// The MCP version shows "None of the requested methods (put) are valid"
	// Our CLI shows a simpler error, but it should indicate the method wasn't found
	if !strings.Contains(strings.ToLower(out), "put") || !strings.Contains(strings.ToLower(out), "not found") {
		t.Errorf("expected error mentioning 'put' and 'not found', got: %s", out)
	}
}

// Converted from: e2e/resources.test.ts — paths listing for complex endpoint
func TestPaths_ComplexEndpointListPaths(t *testing.T) {
	out, err := executeCommand("--spec", testdataPath("complex-endpoint.json"), "paths")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "/api") {
		t.Error("expected /api in paths list")
	}
}

// Converted from: paths.test.ts — deep nesting paths
func TestPaths_DeepNesting(t *testing.T) {
	out, err := executeCommand("--spec", testdataPath("paths-test.json"), "paths", "--depth", "5")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "/sub/sub/sub/sub/folded") {
		t.Error("expected deep path at depth 5")
	}
}

// Converted from: path-item-handler.test.ts — deeply nested path with parameters
func TestPaths_ParameterizedSubpaths(t *testing.T) {
	out, err := executeCommand("--spec", testdataPath("paths-test.json"), "paths", "/article")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "/article/{articleId}") {
		t.Error("expected /article/{articleId} sub-path")
	}
}

// OpenAPI 3.1 spec — list paths
func TestPaths_OpenAPI31ListPaths(t *testing.T) {
	out, err := executeCommand("--spec", testdataPath("openapi31-spec.json"), "paths")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "/items") {
		t.Error("expected /items in paths list")
	}
}

// OpenAPI 3.1 spec — list methods for path
func TestPaths_OpenAPI31ListMethods(t *testing.T) {
	out, err := executeCommand("--spec", testdataPath("openapi31-spec.json"), "paths", "/items")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "GET") {
		t.Error("expected GET method")
	}
	if !strings.Contains(out, "POST") {
		t.Error("expected POST method")
	}
	if !strings.Contains(out, "List Items") {
		t.Error("expected 'List Items' summary")
	}
}

// OpenAPI 3.1 spec — operation detail with shallow mode (default)
func TestPaths_OpenAPI31OperationDetail(t *testing.T) {
	out, err := executeCommand("--spec", testdataPath("openapi31-spec.json"), "paths", "/items", "get")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var op map[string]any
	if err := yaml.Unmarshal([]byte(out), &op); err != nil {
		t.Fatalf("output is not valid YAML: %v\noutput: %s", err, out)
	}
	if op["operationId"] != "listItems" {
		t.Errorf("expected operationId 'listItems', got %v", op["operationId"])
	}
	// In shallow mode (default), $ref in response schema should be PRESERVED
	responses := op["responses"].(map[string]any)
	resp200 := responses["200"].(map[string]any)
	content := resp200["content"].(map[string]any)
	appJSON := content["application/json"].(map[string]any)
	schema := appJSON["schema"].(map[string]any)
	if _, hasRef := schema["$ref"]; !hasRef {
		t.Error("$ref in response schema should be preserved in shallow mode")
	}
}

// OpenAPI 3.1 spec — sub-path discovery
func TestPaths_OpenAPI31SubPaths(t *testing.T) {
	out, err := executeCommand("--spec", testdataPath("openapi31-spec.json"), "paths", "/items")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "/items/{itemId}") {
		t.Error("expected /items/{itemId} as sub-path")
	}
}

func TestSwagger2_Paths(t *testing.T) {
	out, err := executeCommand("--spec", testdataPath("sample-v2-api.json"), "paths")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "/ping") {
		t.Error("expected /ping path from converted Swagger 2.0 spec")
	}
}

func TestPaths_ListShowsMethods(t *testing.T) {
	out, err := executeCommand("--spec", testdataPath("sample-api.json"), "paths", "--depth", "1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "GET") {
		t.Error("expected GET method in path listing")
	}
	if !strings.Contains(out, "/users") {
		t.Error("expected /users in path listing")
	}
}

func TestPaths_AutoDepthTruncates(t *testing.T) {
	out, err := executeCommand("--spec", testdataPath("many-paths.json"), "paths")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "Showing paths to depth") {
		t.Error("expected truncation notice")
	}
	if strings.Contains(out, "/a/1") {
		t.Error("auto-depth should not show /a/1 at depth 1")
	}
}

func TestPaths_AutoDepthNoTruncationWhenSmall(t *testing.T) {
	out, err := executeCommand("--spec", testdataPath("paths-test.json"), "paths")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.Contains(out, "Showing paths to depth") {
		t.Error("should not show truncation notice for small spec")
	}
	if !strings.Contains(out, "/project/tasks/{taskId}") {
		t.Error("expected full paths when under 50 lines")
	}
}

func TestPaths_ExplicitDepthOverridesAuto(t *testing.T) {
	out, err := executeCommand("--spec", testdataPath("many-paths.json"), "paths", "--depth", "2")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.Contains(out, "Showing paths to depth") {
		t.Error("explicit --depth should not show truncation notice")
	}
	if !strings.Contains(out, "/a/1") {
		t.Error("expected /a/1 with --depth 2")
	}
}

func TestPaths_FilterByPathSubstring(t *testing.T) {
	out, err := executeCommand("--spec", testdataPath("tags-test.json"), "paths", "--filter", "user")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "/users") {
		t.Error("expected /users to match --filter user")
	}
	if strings.Contains(out, "/admin") {
		t.Error("expected /admin to NOT match --filter user")
	}
}

func TestPaths_FilterByTag(t *testing.T) {
	out, err := executeCommand("--spec", testdataPath("tags-test.json"), "paths", "--filter", "Admin")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "/admin/settings") {
		t.Error("expected /admin/settings to match via Admin tag")
	}
}

func TestPaths_FilterBypassesAutoDepth(t *testing.T) {
	out, err := executeCommand("--spec", testdataPath("many-paths.json"), "paths", "--filter", "a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "/a/1") {
		t.Error("expected /a/1 at full depth with --filter")
	}
	if strings.Contains(out, "Showing paths to depth") {
		t.Error("--filter should not show truncation notice")
	}
}

func TestPaths_FilterCaseInsensitive(t *testing.T) {
	out, err := executeCommand("--spec", testdataPath("tags-test.json"), "paths", "--filter", "USERS")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "/users") {
		t.Error("expected case-insensitive match")
	}
}

func TestPaths_FilterNoMatch(t *testing.T) {
	out, err := executeCommand("--spec", testdataPath("tags-test.json"), "paths", "--filter", "zzz")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	lines := strings.Split(strings.TrimSpace(out), "\n")
	if len(lines) > 1 {
		t.Errorf("expected no matches, got: %s", out)
	}
}

func TestPaths_FilterShowsMethods(t *testing.T) {
	out, err := executeCommand("--spec", testdataPath("tags-test.json"), "paths", "--filter", "user")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "GET") {
		t.Error("expected methods shown with --filter results")
	}
}

func TestPaths_MethodsOnlyForConcretePaths(t *testing.T) {
	out, err := executeCommand("--spec", testdataPath("complex-endpoint.json"), "paths", "--depth", "1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	lines := strings.Split(strings.TrimSpace(out), "\n")
	for _, line := range lines {
		if strings.Contains(line, "/api") && !strings.HasPrefix(line, "#") {
			if strings.Contains(line, "GET") || strings.Contains(line, "POST") {
				t.Errorf("structural prefix should not show methods: %q", line)
			}
		}
	}
}
