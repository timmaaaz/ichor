package tools_test

import (
	"encoding/json"
	"testing"

	"github.com/timmaaaz/ichor/mcp/internal/tools"
)

// ===== search_database_schema (progressive modes) =====

func TestSearchDB_NoArgs_ListSchemas(t *testing.T) {
	schemasJSON := `["core","hr","inventory","products","sales","workflow"]`

	session, ctx := setupToolTest(t,
		pathRouter(map[string]string{"/v1/introspection/schemas": schemasJSON}),
		tools.RegisterSearchTools,
	)

	result := callTool(t, session, ctx, "search_database_schema", nil)

	if result.IsError {
		t.Fatalf("returned error: %s", getTextContent(t, result))
	}
	if text := getTextContent(t, result); text != schemasJSON {
		t.Errorf("got %q, want %q", text, schemasJSON)
	}
}

func TestSearchDB_SchemaOnly_ListTables(t *testing.T) {
	tablesJSON := `["users","roles","permissions","pages"]`

	session, ctx := setupToolTest(t,
		pathRouter(map[string]string{"/v1/introspection/schemas/core/tables": tablesJSON}),
		tools.RegisterSearchTools,
	)

	result := callTool(t, session, ctx, "search_database_schema", map[string]any{
		"schema": "core",
	})

	if result.IsError {
		t.Fatalf("returned error: %s", getTextContent(t, result))
	}
	if text := getTextContent(t, result); text != tablesJSON {
		t.Errorf("got %q, want %q", text, tablesJSON)
	}
}

func TestSearchDB_SchemaAndTable_ColumnsAndRelationships(t *testing.T) {
	columnsJSON := `[{"name":"id","type":"uuid"},{"name":"email","type":"text"}]`
	relsJSON := `[{"column":"role_id","foreign_table":"roles"}]`

	session, ctx := setupToolTest(t,
		pathRouter(map[string]string{
			"/v1/introspection/tables/core/users/columns":       columnsJSON,
			"/v1/introspection/tables/core/users/relationships": relsJSON,
		}),
		tools.RegisterSearchTools,
	)

	result := callTool(t, session, ctx, "search_database_schema", map[string]any{
		"schema": "core",
		"table":  "users",
	})

	if result.IsError {
		t.Fatalf("returned error: %s", getTextContent(t, result))
	}

	// Verify merged JSON contains both keys.
	text := getTextContent(t, result)
	var merged map[string]json.RawMessage
	if err := json.Unmarshal([]byte(text), &merged); err != nil {
		t.Fatalf("failed to parse result: %v", err)
	}
	if _, ok := merged["columns"]; !ok {
		t.Error("result missing 'columns' key")
	}
	if _, ok := merged["relationships"]; !ok {
		t.Error("result missing 'relationships' key")
	}
}

func TestSearchDB_APIError(t *testing.T) {
	session, ctx := setupToolTest(t,
		errorHandler(500),
		tools.RegisterSearchTools,
	)

	// Test all three modes return errors.
	tests := []struct {
		name string
		args map[string]any
	}{
		{"no_args", nil},
		{"schema_only", map[string]any{"schema": "core"}},
		{"schema_and_table", map[string]any{"schema": "core", "table": "users"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := callTool(t, session, ctx, "search_database_schema", tt.args)
			if !result.IsError {
				t.Error("should return error on API failure")
			}
		})
	}
}

// ===== search_enums =====

func TestSearchEnums_SchemaOnly_ListTypes(t *testing.T) {
	enumsJSON := `["role_type","trigger_type","action_status"]`

	session, ctx := setupToolTest(t,
		pathRouter(map[string]string{"/v1/introspection/enums/core": enumsJSON}),
		tools.RegisterSearchTools,
	)

	result := callTool(t, session, ctx, "search_enums", map[string]any{
		"schema": "core",
	})

	if result.IsError {
		t.Fatalf("returned error: %s", getTextContent(t, result))
	}
	if text := getTextContent(t, result); text != enumsJSON {
		t.Errorf("got %q, want %q", text, enumsJSON)
	}
}

func TestSearchEnums_SchemaAndName_GetValues(t *testing.T) {
	optionsJSON := `[{"value":"admin","label":"Administrator"},{"value":"user","label":"Standard User"}]`

	session, ctx := setupToolTest(t,
		pathRouter(map[string]string{"/v1/config/enums/core/role_type/options": optionsJSON}),
		tools.RegisterSearchTools,
	)

	result := callTool(t, session, ctx, "search_enums", map[string]any{
		"schema": "core",
		"name":   "role_type",
	})

	if result.IsError {
		t.Fatalf("returned error: %s", getTextContent(t, result))
	}
	if text := getTextContent(t, result); text != optionsJSON {
		t.Errorf("got %q, want %q", text, optionsJSON)
	}
}

func TestSearchEnums_MissingSchema(t *testing.T) {
	session, ctx := setupToolTest(t,
		staticHandler(`{}`),
		tools.RegisterSearchTools,
	)

	// Pass empty string for schema to bypass SDK required check but trigger handler validation.
	result := callTool(t, session, ctx, "search_enums", map[string]any{"schema": ""})
	if !result.IsError {
		t.Error("search_enums should return error when schema is empty")
	}
}

func TestSearchEnums_APIError(t *testing.T) {
	session, ctx := setupToolTest(t,
		errorHandler(500),
		tools.RegisterSearchTools,
	)

	tests := []struct {
		name string
		args map[string]any
	}{
		{"schema_only", map[string]any{"schema": "core"}},
		{"schema_and_name", map[string]any{"schema": "core", "name": "role_type"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := callTool(t, session, ctx, "search_enums", tt.args)
			if !result.IsError {
				t.Error("should return error on API failure")
			}
		})
	}
}
