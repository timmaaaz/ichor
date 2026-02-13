package tools_test

import (
	"testing"

	"github.com/timmaaaz/ichor/mcp/internal/tools"
)

func TestValidateTableConfig_Success(t *testing.T) {
	validationResult := `{"valid":true,"errors":[]}`

	session, ctx := setupToolTest(t,
		pathRouter(map[string]string{"/v1/data/validate": validationResult}),
		tools.RegisterValidationTools,
	)

	result := callTool(t, session, ctx, "validate_table_config", map[string]any{
		"config": map[string]any{
			"name": "test_table",
			"config": map[string]any{
				"data_sources": []any{},
			},
		},
	})

	if result.IsError {
		t.Fatalf("validate_table_config returned error: %s", getTextContent(t, result))
	}
	if text := getTextContent(t, result); text != validationResult {
		t.Errorf("got %q, want %q", text, validationResult)
	}
}

func TestValidateTableConfig_InvalidConfig(t *testing.T) {
	validationResult := `{"valid":false,"errors":["missing data_sources","invalid column types"]}`

	session, ctx := setupToolTest(t,
		pathRouter(map[string]string{"/v1/data/validate": validationResult}),
		tools.RegisterValidationTools,
	)

	result := callTool(t, session, ctx, "validate_table_config", map[string]any{
		"config": map[string]any{"name": "bad_config"},
	})

	if result.IsError {
		t.Fatal("validate_table_config should not return IsError for validation response")
	}
	if text := getTextContent(t, result); text != validationResult {
		t.Errorf("got %q, want %q", text, validationResult)
	}
}

func TestValidateTableConfig_APIError(t *testing.T) {
	session, ctx := setupToolTest(t,
		errorHandler(500),
		tools.RegisterValidationTools,
	)

	result := callTool(t, session, ctx, "validate_table_config", map[string]any{
		"config": map[string]any{"name": "test"},
	})
	if !result.IsError {
		t.Error("validate_table_config should return error on API failure")
	}
}
