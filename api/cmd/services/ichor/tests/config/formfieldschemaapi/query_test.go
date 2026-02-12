package formfieldschema_test

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/timmaaaz/ichor/api/domain/http/config/formfieldschemaapi"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
)

func queryFieldTypes200(sd FieldSchemaSeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "basic",
			URL:        "/v1/config/form-field-types",
			Token:      sd.Users[0].Token,
			StatusCode: http.StatusOK,
			Method:     http.MethodGet,
			GotResp:    &[]formfieldschemaapi.FieldTypeInfo{},
			ExpResp:    &[]formfieldschemaapi.FieldTypeInfo{},
			CmpFunc: func(got any, exp any) string {
				gotResp, ok := got.(*[]formfieldschemaapi.FieldTypeInfo)
				if !ok {
					return "error casting response"
				}

				if len(*gotResp) != 16 {
					return fmt.Sprintf("expected 16 field types, got %d", len(*gotResp))
				}

				for _, ft := range *gotResp {
					if ft.Type == "" {
						return "field type has empty type"
					}
					if ft.Name == "" {
						return fmt.Sprintf("field type %q has empty name", ft.Type)
					}
					if len(ft.ConfigSchema) == 0 {
						return fmt.Sprintf("field type %q has empty config_schema", ft.Type)
					}
				}

				return ""
			},
		},
	}
}

func queryFieldTypeSchema200(sd FieldSchemaSeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "dropdown",
			URL:        "/v1/config/form-field-types/dropdown/schema",
			Token:      sd.Users[0].Token,
			StatusCode: http.StatusOK,
			Method:     http.MethodGet,
			GotResp:    &formfieldschemaapi.FieldTypeInfo{},
			ExpResp:    &formfieldschemaapi.FieldTypeInfo{},
			CmpFunc: func(got any, exp any) string {
				gotResp, ok := got.(*formfieldschemaapi.FieldTypeInfo)
				if !ok {
					return "error casting response"
				}
				if gotResp.Type != "dropdown" {
					return fmt.Sprintf("expected type dropdown, got %q", gotResp.Type)
				}
				if len(gotResp.ConfigSchema) == 0 {
					return "config schema is empty"
				}

				var schema map[string]any
				if err := json.Unmarshal(gotResp.ConfigSchema, &schema); err != nil {
					return fmt.Sprintf("config schema is not valid JSON: %v", err)
				}
				props, ok := schema["properties"].(map[string]any)
				if !ok {
					return "schema missing properties"
				}
				for _, required := range []string{"entity", "label_column", "value_column"} {
					if _, ok := props[required]; !ok {
						return fmt.Sprintf("dropdown schema missing %q property", required)
					}
				}

				return ""
			},
		},
	}
}

func queryFieldTypeSchema404(sd FieldSchemaSeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "nonexistent",
			URL:        "/v1/config/form-field-types/nonexistent/schema",
			Token:      sd.Users[0].Token,
			StatusCode: http.StatusNotFound,
			Method:     http.MethodGet,
			GotResp:    &map[string]any{},
			ExpResp:    &map[string]any{},
			CmpFunc:    func(got any, exp any) string { return "" },
		},
	}
}

func queryFieldTypes401(sd FieldSchemaSeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "field-types-no-token",
			URL:        "/v1/config/form-field-types",
			Token:      "",
			StatusCode: http.StatusUnauthorized,
			Method:     http.MethodGet,
			GotResp:    &map[string]any{},
			ExpResp:    &map[string]any{},
			CmpFunc:    func(got any, exp any) string { return "" },
		},
		{
			Name:       "field-type-schema-no-token",
			URL:        "/v1/config/form-field-types/dropdown/schema",
			Token:      "",
			StatusCode: http.StatusUnauthorized,
			Method:     http.MethodGet,
			GotResp:    &map[string]any{},
			ExpResp:    &map[string]any{},
			CmpFunc:    func(got any, exp any) string { return "" },
		},
	}
}
