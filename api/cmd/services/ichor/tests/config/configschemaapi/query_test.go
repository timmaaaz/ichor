package configschema_test

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/timmaaaz/ichor/api/domain/http/config/configschemaapi"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
)

func queryTableConfigSchema200(sd SchemaSeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "basic",
			URL:        "/v1/config/schemas/table-config",
			Token:      sd.Users[0].Token,
			StatusCode: http.StatusOK,
			Method:     http.MethodGet,
			GotResp:    &configschemaapi.SchemaInfo{},
			ExpResp:    &configschemaapi.SchemaInfo{},
			CmpFunc: func(got any, exp any) string {
				gotResp, ok := got.(*configschemaapi.SchemaInfo)
				if !ok {
					return "error casting response"
				}
				if gotResp.Name == "" {
					return "schema name is empty"
				}
				if len(gotResp.Schema) == 0 {
					return "schema is empty"
				}

				var schema map[string]any
				if err := json.Unmarshal(gotResp.Schema, &schema); err != nil {
					return fmt.Sprintf("schema is not valid JSON: %v", err)
				}
				props, ok := schema["properties"].(map[string]any)
				if !ok {
					return "schema missing properties field"
				}
				if _, ok := props["data_source"]; !ok {
					return "schema missing data_source property"
				}
				if _, ok := props["visual_settings"]; !ok {
					return "schema missing visual_settings property"
				}

				return ""
			},
		},
	}
}

func queryLayoutSchema200(sd SchemaSeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "basic",
			URL:        "/v1/config/schemas/layout",
			Token:      sd.Users[0].Token,
			StatusCode: http.StatusOK,
			Method:     http.MethodGet,
			GotResp:    &configschemaapi.SchemaInfo{},
			ExpResp:    &configschemaapi.SchemaInfo{},
			CmpFunc: func(got any, exp any) string {
				gotResp, ok := got.(*configschemaapi.SchemaInfo)
				if !ok {
					return "error casting response"
				}
				if gotResp.Name == "" {
					return "schema name is empty"
				}
				if len(gotResp.Schema) == 0 {
					return "schema is empty"
				}

				var schema map[string]any
				if err := json.Unmarshal(gotResp.Schema, &schema); err != nil {
					return fmt.Sprintf("schema is not valid JSON: %v", err)
				}
				defs, ok := schema["$defs"].(map[string]any)
				if !ok {
					return "schema missing $defs"
				}
				if _, ok := defs["ResponsiveValue"]; !ok {
					return "schema missing ResponsiveValue definition"
				}

				return ""
			},
		},
	}
}

func queryContentTypes200(sd SchemaSeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "basic",
			URL:        "/v1/config/schemas/content-types",
			Token:      sd.Users[0].Token,
			StatusCode: http.StatusOK,
			Method:     http.MethodGet,
			GotResp:    &[]configschemaapi.ContentTypeInfo{},
			ExpResp:    &[]configschemaapi.ContentTypeInfo{},
			CmpFunc: func(got any, exp any) string {
				gotResp, ok := got.(*[]configschemaapi.ContentTypeInfo)
				if !ok {
					return "error casting response"
				}
				if len(*gotResp) != 6 {
					return fmt.Sprintf("expected 6 content types, got %d", len(*gotResp))
				}

				expectedTypes := map[string]bool{
					"table": true, "form": true, "chart": true,
					"tabs": true, "container": true, "text": true,
				}
				for _, ct := range *gotResp {
					delete(expectedTypes, ct.Type)
					if ct.Name == "" {
						return fmt.Sprintf("content type %q has empty name", ct.Type)
					}
				}
				if len(expectedTypes) > 0 {
					missing := make([]string, 0, len(expectedTypes))
					for t := range expectedTypes {
						missing = append(missing, t)
					}
					return fmt.Sprintf("missing content types: %v", missing)
				}

				return ""
			},
		},
	}
}

func querySchemas401(sd SchemaSeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "table-config-no-token",
			URL:        "/v1/config/schemas/table-config",
			Token:      "",
			StatusCode: http.StatusUnauthorized,
			Method:     http.MethodGet,
			GotResp:    &map[string]any{},
			ExpResp:    &map[string]any{},
			CmpFunc:    func(got any, exp any) string { return "" },
		},
		{
			Name:       "layout-no-token",
			URL:        "/v1/config/schemas/layout",
			Token:      "",
			StatusCode: http.StatusUnauthorized,
			Method:     http.MethodGet,
			GotResp:    &map[string]any{},
			ExpResp:    &map[string]any{},
			CmpFunc:    func(got any, exp any) string { return "" },
		},
		{
			Name:       "content-types-no-token",
			URL:        "/v1/config/schemas/content-types",
			Token:      "",
			StatusCode: http.StatusUnauthorized,
			Method:     http.MethodGet,
			GotResp:    &map[string]any{},
			ExpResp:    &map[string]any{},
			CmpFunc:    func(got any, exp any) string { return "" },
		},
	}
}
