package resolver

import (
	"strings"
)

// Resolver resolves $ref pointers inline with circular reference detection.
type Resolver struct {
	components map[string]any
}

// New creates a Resolver with the spec's components section.
func New(components map[string]any) *Resolver {
	return &Resolver{components: components}
}

// Resolve recursively resolves $ref pointers in the given value (shallow mode).
// $ref inside responses.*.content.*.schema and requestBody.content.*.schema are
// preserved as-is. All other internal refs (#/components/...) are inlined.
// External refs are left as-is. Circular refs are replaced with {"$circular": "SchemaName"}.
func (r *Resolver) Resolve(v any) any {
	seen := make(map[string]bool)
	return r.resolve(v, seen, false, "")
}

// ResolveFull recursively resolves all $ref pointers inline (full mode).
// Internal refs (#/components/...) are inlined. External refs are left as-is.
// Circular refs are replaced with {"$circular": "SchemaName"} on second encounter.
func (r *Resolver) ResolveFull(v any) any {
	seen := make(map[string]bool)
	return r.resolve(v, seen, true, "")
}

func (r *Resolver) resolve(v any, seen map[string]bool, full bool, parentKey string) any {
	switch val := v.(type) {
	case map[string]any:
		return r.resolveMap(val, seen, full, parentKey)
	case []any:
		return r.resolveSlice(val, seen, full, parentKey)
	default:
		return v
	}
}

func (r *Resolver) resolveMap(m map[string]any, seen map[string]bool, full bool, parentKey string) any {
	if ref, ok := m["$ref"].(string); ok {
		// In shallow mode, preserve $ref if we're at a body schema position
		if !full && parentKey == "schema" {
			return map[string]any{"$ref": ref}
		}
		return r.resolveRef(ref, seen, full)
	}

	result := make(map[string]any, len(m))
	for k, v := range m {
		result[k] = r.resolve(v, seen, full, effectiveKey(parentKey, k))
	}
	return result
}

func (r *Resolver) resolveSlice(s []any, seen map[string]bool, full bool, parentKey string) any {
	result := make([]any, len(s))
	for i, v := range s {
		result[i] = r.resolve(v, seen, full, parentKey)
	}
	return result
}

func (r *Resolver) resolveRef(ref string, seen map[string]bool, full bool) any {
	if !strings.HasPrefix(ref, "#/components/") {
		return map[string]any{"$ref": ref}
	}

	// Parse: #/components/{type}/{name}
	parts := strings.SplitN(strings.TrimPrefix(ref, "#/components/"), "/", 2)
	if len(parts) != 2 {
		return map[string]any{"$ref": ref}
	}
	compType, compName := parts[0], parts[1]

	typeMap, ok := r.components[compType].(map[string]any)
	if !ok {
		return map[string]any{"$ref": ref}
	}
	schema, ok := typeMap[compName]
	if !ok {
		return map[string]any{"$ref": ref}
	}

	if seen[ref] {
		// Already seen in this chain — resolve one more level but cap all
		// nested refs to the same target with $circular markers.
		return capCircular(deepCopy(schema), ref, compName)
	}

	// Mark as seen, resolve, then unmark (per-chain, not global)
	seen[ref] = true
	resolved := r.resolve(deepCopy(schema), seen, full, "")
	delete(seen, ref)

	return resolved
}

// effectiveKey returns the key to pass as parentKey when descending into a child.
// It tracks the chain: responses/requestBody → status/body → content → mediaType → schema.
func effectiveKey(parentKey, childKey string) string {
	switch parentKey {
	case "":
		// Top-level: start tracking if entering responses or requestBody
		if childKey == "responses" || childKey == "requestBody" {
			return childKey
		}
		return ""
	case "responses":
		// status code keys (e.g. "200", "404")
		return "response_status"
	case "response_status":
		if childKey == "content" {
			return "content"
		}
		return ""
	case "requestBody":
		if childKey == "content" {
			return "content"
		}
		return ""
	case "content":
		// media type keys (e.g. "application/json")
		return "media_type"
	case "media_type":
		if childKey == "schema" {
			return "schema"
		}
		return ""
	default:
		return ""
	}
}

// capCircular resolves a schema one level deep, replacing any $ref to the
// circular target with a $circular marker instead of recursing further.
func capCircular(v any, ref string, name string) any {
	switch val := v.(type) {
	case map[string]any:
		if r, ok := val["$ref"].(string); ok && r == ref {
			return map[string]any{"$circular": name}
		}
		result := make(map[string]any, len(val))
		for k, v := range val {
			result[k] = capCircular(v, ref, name)
		}
		return result
	case []any:
		result := make([]any, len(val))
		for i, v := range val {
			result[i] = capCircular(v, ref, name)
		}
		return result
	default:
		return v
	}
}

// deepCopy creates a deep copy of a value to avoid mutating the original components.
func deepCopy(v any) any {
	switch val := v.(type) {
	case map[string]any:
		m := make(map[string]any, len(val))
		for k, v := range val {
			m[k] = deepCopy(v)
		}
		return m
	case []any:
		s := make([]any, len(val))
		for i, v := range val {
			s[i] = deepCopy(v)
		}
		return s
	default:
		return v
	}
}
