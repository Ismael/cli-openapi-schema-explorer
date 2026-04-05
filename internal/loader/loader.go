package loader

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/getkin/kin-openapi/openapi2"
	"github.com/getkin/kin-openapi/openapi2conv"
	"github.com/getkin/kin-openapi/openapi3"
)

// Load loads an OpenAPI spec from a file path or URL.
// Automatically detects and converts Swagger 2.0 specs to OpenAPI 3.x.
func Load(source string) (*openapi3.T, error) {
	data, err := readSource(source)
	if err != nil {
		return nil, fmt.Errorf("failed to read spec: %w", err)
	}

	if isSwagger2(data) {
		return loadSwagger2(data)
	}

	doc, err := loadOpenAPI3(data)
	if err != nil {
		// Retry with sanitized data to handle common incompatibilities
		// (e.g. numeric exclusiveMinimum/exclusiveMaximum from OpenAPI 3.1 / JSON Schema 2020-12)
		sanitized, sanitizeErr := sanitizeSpec(data)
		if sanitizeErr != nil {
			return nil, fmt.Errorf("failed to parse spec: %w", err)
		}
		doc, err = loadOpenAPI3(sanitized)
		if err != nil {
			return nil, fmt.Errorf("failed to parse spec: %w", err)
		}
	}

	return doc, nil
}

func readSource(source string) ([]byte, error) {
	if strings.HasPrefix(source, "http://") || strings.HasPrefix(source, "https://") {
		return fetchURL(source)
	}
	return os.ReadFile(source)
}

func fetchURL(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch URL: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d from %s", resp.StatusCode, url)
	}

	return io.ReadAll(resp.Body)
}

func loadOpenAPI3(data []byte) (*openapi3.T, error) {
	loader := openapi3.NewLoader()
	loader.IsExternalRefsAllowed = false
	return loader.LoadFromData(data)
}

// sanitizeSpec fixes common incompatibilities between OpenAPI 3.1/JSON Schema 2020-12
// and kin-openapi's OpenAPI 3.0 parser. Specifically, it converts numeric
// exclusiveMinimum/exclusiveMaximum to their boolean equivalents.
func sanitizeSpec(data []byte) ([]byte, error) {
	var raw any
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, err
	}
	sanitizeValue(raw)
	return json.Marshal(raw)
}

func sanitizeValue(v any) {
	switch val := v.(type) {
	case map[string]any:
		sanitizeMap(val)
	case []any:
		for _, item := range val {
			sanitizeValue(item)
		}
	}
}

func sanitizeMap(m map[string]any) {
	// Convert numeric exclusiveMinimum to boolean + adjust minimum
	if exMin, ok := m["exclusiveMinimum"]; ok {
		if n, isNum := exMin.(float64); isNum {
			m["exclusiveMinimum"] = true
			if _, hasMin := m["minimum"]; !hasMin {
				m["minimum"] = n
			}
		}
	}
	// Convert numeric exclusiveMaximum to boolean + adjust maximum
	if exMax, ok := m["exclusiveMaximum"]; ok {
		if n, isNum := exMax.(float64); isNum {
			m["exclusiveMaximum"] = true
			if _, hasMax := m["maximum"]; !hasMax {
				m["maximum"] = n
			}
		}
	}
	for _, v := range m {
		sanitizeValue(v)
	}
}

func isSwagger2(data []byte) bool {
	var probe struct {
		Swagger string `json:"swagger" yaml:"swagger"`
	}
	if err := json.Unmarshal(data, &probe); err != nil {
		return false
	}
	return strings.HasPrefix(probe.Swagger, "2.")
}

func loadSwagger2(data []byte) (*openapi3.T, error) {
	var swagger openapi2.T
	if err := json.Unmarshal(data, &swagger); err != nil {
		return nil, fmt.Errorf("failed to parse Swagger 2.0 spec: %w", err)
	}
	doc, err := openapi2conv.ToV3(&swagger)
	if err != nil {
		return nil, fmt.Errorf("failed to convert Swagger 2.0 to OpenAPI 3.0: %w", err)
	}
	return doc, nil
}

// Doc is a type alias for the loaded OpenAPI document.
type Doc = openapi3.T
