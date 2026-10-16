package translator

import (
	"encoding/json"
	"testing"
)

func TestCleanParametersSchema_PropertyNamesAsKeywords(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		checks func(t *testing.T, res map[string]any)
	}{
		{
			name: "preserves parameter named title, default, x-custom, const",
			input: `{
				"type": "object",
				"properties": {
					"title": { "type": "string", "description": "The document title" },
					"default": { "type": "boolean", "description": "Default flag" },
					"x-custom": { "type": "string", "description": "Custom metadata" },
					"const": { "type": "string", "description": "Constant value" }
				},
				"required": ["title"]
			}`,
			checks: func(t *testing.T, res map[string]any) {
				props, ok := res["properties"].(map[string]any)
				if !ok {
					t.Fatalf("expected properties map, got %T", res["properties"])
				}
				for _, key := range []string{"title", "default", "x-custom", "const"} {
					prop, exists := props[key].(map[string]any)
					if !exists {
						t.Errorf("expected property %q to be preserved", key)
						continue
					}
					if prop["type"] == nil {
						t.Errorf("expected property %q to retain its type", key)
					}
				}
				req, ok := res["required"].([]any)
				if !ok || len(req) != 1 || req[0] != "title" {
					t.Errorf("expected required to contain [title], got %v", res["required"])
				}
			},
		},
		{
			name: "preserves keyword-named parameters inside nested objects and arrays",
			input: `{
				"type": "object",
				"properties": {
					"nested": {
						"type": "object",
						"properties": {
							"properties": { "type": "string", "description": "Nested properties parameter" },
							"items": { "type": "string", "description": "Nested items parameter" }
						}
					},
					"list": {
						"type": "array",
						"items": {
							"type": "object",
							"properties": {
								"required": { "type": "string", "description": "Array element required parameter" },
								"enum": { "type": "string", "description": "Array element enum parameter" }
							}
						}
					}
				}
			}`,
			checks: func(t *testing.T, res map[string]any) {
				props, ok := res["properties"].(map[string]any)
				if !ok {
					t.Fatalf("expected properties map")
				}
				nested, ok := props["nested"].(map[string]any)
				if !ok {
					t.Fatalf("expected nested object")
				}
				nestedProps, ok := nested["properties"].(map[string]any)
				if !ok {
					t.Fatalf("expected nested properties")
				}
				if _, ok := nestedProps["properties"]; !ok {
					t.Errorf("expected property named 'properties' to be preserved")
				}
				if _, ok := nestedProps["items"]; !ok {
					t.Errorf("expected property named 'items' to be preserved")
				}

				list, ok := props["list"].(map[string]any)
				if !ok {
					t.Fatalf("expected list array")
				}
				itemsSchema, ok := list["items"].(map[string]any)
				if !ok {
					t.Fatalf("expected items schema")
				}
				itemProps, ok := itemsSchema["properties"].(map[string]any)
				if !ok {
					t.Fatalf("expected item properties")
				}
				if _, ok := itemProps["required"]; !ok {
					t.Errorf("expected property named 'required' to be preserved")
				}
				if _, ok := itemProps["enum"]; !ok {
					t.Errorf("expected property named 'enum' to be preserved")
				}
			},
		},
		{
			name: "strips unsupported keywords from subschemas without deleting keyword-named properties",
			input: `{
				"type": "object",
				"properties": {
					"errorMessage": {
						"type": "string",
						"minLength": 5,
						"errorMessage": "must be at least 5 chars"
					}
				}
			}`,
			checks: func(t *testing.T, res map[string]any) {
				props, ok := res["properties"].(map[string]any)
				if !ok {
					t.Fatalf("expected properties map")
				}
				errProp, ok := props["errorMessage"].(map[string]any)
				if !ok {
					t.Fatalf("expected errorMessage property to be preserved")
				}
				if errProp["type"] != "string" {
					t.Errorf("expected type string, got %v", errProp["type"])
				}
				if _, exists := errProp["minLength"]; exists {
					t.Errorf("expected minLength to be stripped")
				}
				if _, exists := errProp["errorMessage"]; exists {
					t.Errorf("expected unsupported errorMessage keyword to be stripped from subschema")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleaned := CleanParametersSchema([]byte(tt.input))
			var parsed map[string]any
			if err := json.Unmarshal(cleaned, &parsed); err != nil {
				t.Fatalf("failed to unmarshal cleaned schema: %v", err)
			}
			tt.checks(t, parsed)
		})
	}
}

func TestCleanParametersSchema_ConstToEnumAndKeywordPreservation(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		checks func(t *testing.T, res map[string]any)
	}{
		{
			name: "parameter named const with scalar const value becomes enum",
			input: `{
				"type": "object",
				"properties": {
					"const": {
						"const": "my-constant-val"
					}
				},
				"required": ["const"]
			}`,
			checks: func(t *testing.T, res map[string]any) {
				props, ok := res["properties"].(map[string]any)
				if !ok {
					t.Fatalf("expected properties map")
				}
				constProp, ok := props["const"].(map[string]any)
				if !ok {
					t.Fatalf("expected const property")
				}
				if _, exists := constProp["const"]; exists {
					t.Errorf("expected 'const' keyword to be deleted")
				}
				enumVal, exists := constProp["enum"]
				if !exists {
					t.Fatalf("expected 'enum' keyword to be added from const")
				}
				arr, ok := enumVal.([]any)
				if !ok || len(arr) != 1 || arr[0] != "my-constant-val" {
					t.Errorf("expected enum = ['my-constant-val'], got %v", enumVal)
				}
				// Verify required keeps all three (properties, required, etc.)
				req, ok := res["required"].([]any)
				if !ok || len(req) != 1 || req[0] != "const" {
					t.Errorf("expected required to keep 'const', got %v", res["required"])
				}
			},
		},
		{
			name: "const, enum, title parameters preserve exact schemas and required keeps them",
			input: `{
				"type": "object",
				"properties": {
					"const": { "type": "string", "title": "ConstTitle", "enum": ["a", "b"] },
					"enum": { "type": "string", "title": "EnumTitle" },
					"title": { "type": "string" }
				},
				"required": ["const", "enum", "title"]
			}`,
			checks: func(t *testing.T, res map[string]any) {
				props, ok := res["properties"].(map[string]any)
				if !ok {
					t.Fatalf("expected properties map")
				}
				for _, key := range []string{"const", "enum", "title"} {
					prop, ok := props[key].(map[string]any)
					if !ok {
						t.Errorf("expected property %q to be preserved", key)
						continue
					}
					if prop["type"] != "string" {
						t.Errorf("expected property %q type to be string (no type overwrite/placeholder), got %v", key, prop["type"])
					}
				}
				req, ok := res["required"].([]any)
				if !ok || len(req) != 3 {
					t.Errorf("expected required to keep all 3 parameters, got %v", res["required"])
				}
			},
		},
		{
			name: "nested keyword-named params through arrays and anyOf flattening",
			input: `{
				"type": "object",
				"properties": {
					"nestedArray": {
						"type": "array",
						"items": {
							"anyOf": [
								{
									"type": "object",
									"properties": {
										"const": { "type": "string", "const": "sub" },
										"enum": { "type": "string" }
									}
								}
							]
						}
					}
				}
			}`,
			checks: func(t *testing.T, res map[string]any) {
				props, ok := res["properties"].(map[string]any)
				if !ok {
					t.Fatalf("expected properties map")
				}
				arrSchema, ok := props["nestedArray"].(map[string]any)
				if !ok {
					t.Fatalf("expected nestedArray")
				}
				itemsSchema, ok := arrSchema["items"].(map[string]any)
				if !ok {
					t.Fatalf("expected items schema")
				}
				itemProps, ok := itemsSchema["properties"].(map[string]any)
				if !ok {
					t.Fatalf("expected item properties after anyOf flattening")
				}
				constSub, ok := itemProps["const"].(map[string]any)
				if !ok {
					t.Fatalf("expected property named 'const'")
				}
				if _, exists := constSub["const"]; exists {
					t.Errorf("expected const keyword inside subschema converted/stripped")
				}
				if _, exists := itemProps["enum"]; !exists {
					t.Errorf("expected property named 'enum' preserved")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleaned := CleanParametersSchema([]byte(tt.input))
			var parsed map[string]any
			if err := json.Unmarshal(cleaned, &parsed); err != nil {
				t.Fatalf("failed to unmarshal cleaned schema: %v", err)
			}
			tt.checks(t, parsed)
		})
	}
}
