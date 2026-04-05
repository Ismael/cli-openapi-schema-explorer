package resolver

import (
	"encoding/json"
	"testing"
)

func TestResolveSimpleRef(t *testing.T) {
	components := map[string]any{
		"schemas": map[string]any{
			"User": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"id":   map[string]any{"type": "integer"},
					"name": map[string]any{"type": "string"},
				},
			},
		},
	}
	input := map[string]any{
		"$ref": "#/components/schemas/User",
	}

	r := New(components)
	result := r.ResolveFull(input)

	rm, ok := result.(map[string]any)
	if !ok {
		t.Fatalf("expected map, got %T", result)
	}
	if rm["type"] != "object" {
		t.Errorf("expected type 'object', got %v", rm["type"])
	}
	if _, hasRef := rm["$ref"]; hasRef {
		t.Error("$ref should be removed after resolution")
	}
}

func TestResolveNestedRef(t *testing.T) {
	components := map[string]any{
		"schemas": map[string]any{
			"User": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"id": map[string]any{"type": "integer"},
				},
			},
			"UserList": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"users": map[string]any{
						"type": "array",
						"items": map[string]any{
							"$ref": "#/components/schemas/User",
						},
					},
				},
			},
		},
	}
	input := map[string]any{
		"$ref": "#/components/schemas/UserList",
	}

	r := New(components)
	result := r.ResolveFull(input)

	rm := result.(map[string]any)
	props := rm["properties"].(map[string]any)
	users := props["users"].(map[string]any)
	items := users["items"].(map[string]any)
	if items["type"] != "object" {
		t.Errorf("expected nested User to be resolved, got %v", items)
	}
}

func TestResolveCircularRef(t *testing.T) {
	components := map[string]any{
		"schemas": map[string]any{
			"Tree": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"name": map[string]any{"type": "string"},
					"children": map[string]any{
						"type": "array",
						"items": map[string]any{
							"$ref": "#/components/schemas/Tree",
						},
					},
				},
			},
		},
	}
	input := map[string]any{
		"$ref": "#/components/schemas/Tree",
	}

	r := New(components)
	result := r.ResolveFull(input)

	rm := result.(map[string]any)
	if rm["type"] != "object" {
		t.Fatalf("expected top-level type 'object', got %v", rm["type"])
	}

	props := rm["properties"].(map[string]any)
	children := props["children"].(map[string]any)
	items := children["items"].(map[string]any)
	if items["type"] != "object" {
		t.Fatalf("expected first-level children to be resolved, got %v", items)
	}

	props2 := items["properties"].(map[string]any)
	children2 := props2["children"].(map[string]any)
	items2 := children2["items"].(map[string]any)
	if items2["$circular"] != "Tree" {
		t.Errorf("expected $circular marker, got %v", items2)
	}
}

func TestResolveExternalRef(t *testing.T) {
	components := map[string]any{}
	input := map[string]any{
		"$ref": "https://example.com/schemas/External.json",
	}

	r := New(components)
	result := r.ResolveFull(input)

	rm := result.(map[string]any)
	if rm["$ref"] != "https://example.com/schemas/External.json" {
		t.Errorf("external ref should be preserved, got %v", rm["$ref"])
	}
}

func TestResolveAllOf(t *testing.T) {
	components := map[string]any{
		"schemas": map[string]any{
			"Base": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"id": map[string]any{"type": "string"},
				},
			},
		},
	}
	input := map[string]any{
		"allOf": []any{
			map[string]any{"$ref": "#/components/schemas/Base"},
			map[string]any{
				"type": "object",
				"properties": map[string]any{
					"name": map[string]any{"type": "string"},
				},
			},
		},
	}

	r := New(components)
	result := r.ResolveFull(input)

	rm := result.(map[string]any)
	allOf := rm["allOf"].([]any)
	first := allOf[0].(map[string]any)
	if first["type"] != "object" {
		t.Errorf("expected allOf[0] ref to be resolved, got %v", first)
	}
}

// Converted from: reference-transform.test.ts — anyOf support
func TestResolveAnyOf(t *testing.T) {
	components := map[string]any{
		"schemas": map[string]any{
			"Cat": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"purrs": map[string]any{"type": "boolean"},
				},
			},
			"Dog": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"barks": map[string]any{"type": "boolean"},
				},
			},
		},
	}
	input := map[string]any{
		"anyOf": []any{
			map[string]any{"$ref": "#/components/schemas/Cat"},
			map[string]any{"$ref": "#/components/schemas/Dog"},
		},
	}

	r := New(components)
	result := r.ResolveFull(input)

	rm := result.(map[string]any)
	anyOf := rm["anyOf"].([]any)
	if len(anyOf) != 2 {
		t.Fatalf("expected 2 items in anyOf, got %d", len(anyOf))
	}
	cat := anyOf[0].(map[string]any)
	if cat["type"] != "object" {
		t.Errorf("expected anyOf[0] to be resolved, got %v", cat)
	}
	dog := anyOf[1].(map[string]any)
	if dog["type"] != "object" {
		t.Errorf("expected anyOf[1] to be resolved, got %v", dog)
	}
}

// Converted from: reference-transform.test.ts — oneOf support
func TestResolveOneOf(t *testing.T) {
	components := map[string]any{
		"schemas": map[string]any{
			"StringVal": map[string]any{"type": "string"},
			"IntVal":    map[string]any{"type": "integer"},
		},
	}
	input := map[string]any{
		"oneOf": []any{
			map[string]any{"$ref": "#/components/schemas/StringVal"},
			map[string]any{"$ref": "#/components/schemas/IntVal"},
		},
	}

	r := New(components)
	result := r.ResolveFull(input)

	rm := result.(map[string]any)
	oneOf := rm["oneOf"].([]any)
	if len(oneOf) != 2 {
		t.Fatalf("expected 2 items in oneOf, got %d", len(oneOf))
	}
	first := oneOf[0].(map[string]any)
	if first["type"] != "string" {
		t.Errorf("expected oneOf[0] type 'string', got %v", first["type"])
	}
	second := oneOf[1].(map[string]any)
	if second["type"] != "integer" {
		t.Errorf("expected oneOf[1] type 'integer', got %v", second["type"])
	}
}

// Converted from: reference-transform.test.ts — handles arrays properly (refs in array items)
func TestResolveRefInArrayItems(t *testing.T) {
	components := map[string]any{
		"schemas": map[string]any{
			"Task": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"id":    map[string]any{"type": "string"},
					"title": map[string]any{"type": "string"},
				},
			},
		},
	}
	input := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"items": map[string]any{
				"type": "array",
				"items": map[string]any{
					"$ref": "#/components/schemas/Task",
				},
			},
		},
	}

	r := New(components)
	result := r.ResolveFull(input)

	rm := result.(map[string]any)
	props := rm["properties"].(map[string]any)
	items := props["items"].(map[string]any)
	arrayItems := items["items"].(map[string]any)
	if arrayItems["type"] != "object" {
		t.Errorf("expected array items ref to be resolved, got %v", arrayItems)
	}
	if _, hasRef := arrayItems["$ref"]; hasRef {
		t.Error("$ref should be resolved in array items")
	}
}

// Converted from: reference-transform.test.ts — preserves non-reference values
func TestResolvePreservesNonRefValues(t *testing.T) {
	components := map[string]any{
		"schemas": map[string]any{
			"Test": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"name": map[string]any{"type": "string"},
				},
			},
		},
	}
	// Input without any $ref — should pass through unchanged
	input := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"name": map[string]any{"type": "string"},
		},
	}

	r := New(components)
	result := r.ResolveFull(input)

	rm := result.(map[string]any)
	if rm["type"] != "object" {
		t.Errorf("expected type 'object', got %v", rm["type"])
	}
	props := rm["properties"].(map[string]any)
	name := props["name"].(map[string]any)
	if name["type"] != "string" {
		t.Errorf("expected name type 'string', got %v", name["type"])
	}
}

// Converted from: reference-transform.test.ts — parameter references
func TestResolveParameterRef(t *testing.T) {
	components := map[string]any{
		"parameters": map[string]any{
			"LimitParam": map[string]any{
				"name": "limit",
				"in":   "query",
				"schema": map[string]any{
					"type": "integer",
				},
			},
		},
	}
	input := map[string]any{
		"$ref": "#/components/parameters/LimitParam",
	}

	r := New(components)
	result := r.ResolveFull(input)

	rm := result.(map[string]any)
	if rm["name"] != "limit" {
		t.Errorf("expected parameter name 'limit', got %v", rm["name"])
	}
	if rm["in"] != "query" {
		t.Errorf("expected parameter in 'query', got %v", rm["in"])
	}
}

// Converted from: reference-transform.test.ts — response references
func TestResolveResponseRef(t *testing.T) {
	components := map[string]any{
		"responses": map[string]any{
			"NotFound": map[string]any{
				"description": "Resource not found",
			},
		},
	}
	input := map[string]any{
		"$ref": "#/components/responses/NotFound",
	}

	r := New(components)
	result := r.ResolveFull(input)

	rm := result.(map[string]any)
	if rm["description"] != "Resource not found" {
		t.Errorf("expected description 'Resource not found', got %v", rm["description"])
	}
}

// Converted from: reference-transform.test.ts — nested refs in requestBody and response
func TestResolveNestedRefsInOperation(t *testing.T) {
	components := map[string]any{
		"schemas": map[string]any{
			"Task": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"id": map[string]any{"type": "string"},
				},
			},
		},
	}
	// Simulate an operation with refs in both requestBody and response
	input := map[string]any{
		"requestBody": map[string]any{
			"content": map[string]any{
				"application/json": map[string]any{
					"schema": map[string]any{
						"$ref": "#/components/schemas/Task",
					},
				},
			},
		},
		"responses": map[string]any{
			"201": map[string]any{
				"content": map[string]any{
					"application/json": map[string]any{
						"schema": map[string]any{
							"$ref": "#/components/schemas/Task",
						},
					},
				},
			},
		},
	}

	r := New(components)
	result := r.ResolveFull(input)

	rm := result.(map[string]any)

	// Check requestBody ref resolved
	rb := rm["requestBody"].(map[string]any)
	rbContent := rb["content"].(map[string]any)
	rbJSON := rbContent["application/json"].(map[string]any)
	rbSchema := rbJSON["schema"].(map[string]any)
	if rbSchema["type"] != "object" {
		t.Error("expected requestBody schema ref to be resolved")
	}

	// Check response ref resolved
	resp := rm["responses"].(map[string]any)
	resp201 := resp["201"].(map[string]any)
	respContent := resp201["content"].(map[string]any)
	respJSON := respContent["application/json"].(map[string]any)
	respSchema := respJSON["schema"].(map[string]any)
	if respSchema["type"] != "object" {
		t.Error("expected response schema ref to be resolved")
	}
}

func TestResolveProducesValidJSON(t *testing.T) {
	components := map[string]any{
		"schemas": map[string]any{
			"User": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"id": map[string]any{"type": "integer"},
				},
			},
		},
	}
	input := map[string]any{
		"$ref": "#/components/schemas/User",
	}

	r := New(components)
	result := r.ResolveFull(input)

	_, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("resolved result is not valid JSON: %v", err)
	}
}

func TestResolveShallow_PreservesResponseSchemaRef(t *testing.T) {
	components := map[string]any{
		"schemas": map[string]any{
			"Task": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"id": map[string]any{"type": "string"},
				},
			},
		},
	}
	input := map[string]any{
		"responses": map[string]any{
			"200": map[string]any{
				"content": map[string]any{
					"application/json": map[string]any{
						"schema": map[string]any{
							"$ref": "#/components/schemas/Task",
						},
					},
				},
			},
		},
	}

	r := New(components)
	result := r.Resolve(input)

	rm := result.(map[string]any)
	responses := rm["responses"].(map[string]any)
	resp200 := responses["200"].(map[string]any)
	content := resp200["content"].(map[string]any)
	appJSON := content["application/json"].(map[string]any)
	schema := appJSON["schema"].(map[string]any)

	if schema["$ref"] != "#/components/schemas/Task" {
		t.Errorf("expected $ref to be preserved in response schema, got %v", schema)
	}
	if _, hasType := schema["type"]; hasType {
		t.Error("schema should not be inlined in shallow mode")
	}
}

func TestResolveShallow_PreservesRequestBodySchemaRef(t *testing.T) {
	components := map[string]any{
		"schemas": map[string]any{
			"CreateTask": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"title": map[string]any{"type": "string"},
				},
			},
		},
	}
	input := map[string]any{
		"requestBody": map[string]any{
			"content": map[string]any{
				"application/json": map[string]any{
					"schema": map[string]any{
						"$ref": "#/components/schemas/CreateTask",
					},
				},
			},
		},
	}

	r := New(components)
	result := r.Resolve(input)

	rm := result.(map[string]any)
	rb := rm["requestBody"].(map[string]any)
	content := rb["content"].(map[string]any)
	appJSON := content["application/json"].(map[string]any)
	schema := appJSON["schema"].(map[string]any)

	if schema["$ref"] != "#/components/schemas/CreateTask" {
		t.Errorf("expected $ref to be preserved in requestBody schema, got %v", schema)
	}
	if _, hasType := schema["type"]; hasType {
		t.Error("schema should not be inlined in shallow mode")
	}
}

func TestResolveShallow_StillResolvesParameterRefs(t *testing.T) {
	components := map[string]any{
		"parameters": map[string]any{
			"LimitParam": map[string]any{
				"name": "limit",
				"in":   "query",
				"schema": map[string]any{
					"type": "integer",
				},
			},
		},
	}
	input := map[string]any{
		"parameters": []any{
			map[string]any{
				"$ref": "#/components/parameters/LimitParam",
			},
		},
	}

	r := New(components)
	result := r.Resolve(input)

	rm := result.(map[string]any)
	params := rm["parameters"].([]any)
	param := params[0].(map[string]any)

	if param["name"] != "limit" {
		t.Errorf("expected parameter $ref to be resolved inline, got %v", param)
	}
	if _, hasRef := param["$ref"]; hasRef {
		t.Error("$ref in parameter should be resolved in shallow mode")
	}
}

func TestResolveFull_StillResolvesEverything(t *testing.T) {
	components := map[string]any{
		"schemas": map[string]any{
			"Task": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"id": map[string]any{"type": "string"},
				},
			},
		},
	}
	input := map[string]any{
		"responses": map[string]any{
			"200": map[string]any{
				"content": map[string]any{
					"application/json": map[string]any{
						"schema": map[string]any{
							"$ref": "#/components/schemas/Task",
						},
					},
				},
			},
		},
	}

	r := New(components)
	result := r.ResolveFull(input)

	rm := result.(map[string]any)
	responses := rm["responses"].(map[string]any)
	resp200 := responses["200"].(map[string]any)
	content := resp200["content"].(map[string]any)
	appJSON := content["application/json"].(map[string]any)
	schema := appJSON["schema"].(map[string]any)

	if _, hasRef := schema["$ref"]; hasRef {
		t.Error("$ref should be inlined in ResolveFull mode")
	}
	if schema["type"] != "object" {
		t.Errorf("expected schema type 'object' after full resolution, got %v", schema["type"])
	}
}
