package tools_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/timmaaaz/ichor/mcp/internal/tools"
)

// ===== analyze_workflow =====

func TestAnalyzeWorkflow_Success(t *testing.T) {
	// 4 actions, none with evaluate_condition/create_alert/log_audit_entry → triggers 3 suggestions.
	// 2 edges from same source → branching factor 2 (no high branching suggestion).
	mockRoutes := map[string]string{
		"/v1/workflow/rules/wf-1": `{"id":"wf-1","name":"test rule","trigger_type":"on_create"}`,
		"/v1/workflow/rules/wf-1/actions": `[
			{"id":"a1","action_type":"send_email"},
			{"id":"a2","action_type":"update_field"},
			{"id":"a3","action_type":"send_notification"},
			{"id":"a4","action_type":"delay"}
		]`,
		"/v1/workflow/rules/wf-1/edges": `[
			{"id":"e1","source_action_id":"a1","target_action_id":"a2"},
			{"id":"e2","source_action_id":"a1","target_action_id":"a3"},
			{"id":"e3","source_action_id":"a3","target_action_id":"a4"}
		]`,
	}

	session, ctx := setupToolTest(t,
		pathRouter(mockRoutes),
		tools.RegisterAnalysisTools,
	)

	result := callTool(t, session, ctx, "analyze_workflow", map[string]any{"id": "wf-1"})

	if result.IsError {
		t.Fatalf("returned error: %s", getTextContent(t, result))
	}

	text := getTextContent(t, result)
	var analysis map[string]any
	if err := json.Unmarshal([]byte(text), &analysis); err != nil {
		t.Fatalf("failed to parse analysis: %v", err)
	}

	// Verify metrics.
	if ac, ok := analysis["action_count"].(float64); !ok || ac != 4 {
		t.Errorf("action_count = %v, want 4", analysis["action_count"])
	}
	if ec, ok := analysis["edge_count"].(float64); !ok || ec != 3 {
		t.Errorf("edge_count = %v, want 3", analysis["edge_count"])
	}
	if mb, ok := analysis["max_branching"].(float64); !ok || mb != 2 {
		t.Errorf("max_branching = %v, want 2", analysis["max_branching"])
	}

	// Verify suggestions include expected recommendations.
	suggestions, ok := analysis["suggestions"].([]any)
	if !ok {
		t.Fatalf("suggestions not an array: %T", analysis["suggestions"])
	}

	suggestionText := ""
	for _, s := range suggestions {
		suggestionText += s.(string) + "\n"
	}

	// With 4 actions and no evaluate_condition → should suggest adding it.
	if !strings.Contains(suggestionText, "evaluate_condition") {
		t.Error("expected suggestion about evaluate_condition")
	}
	// No create_alert → should suggest it.
	if !strings.Contains(suggestionText, "create_alert") {
		t.Error("expected suggestion about create_alert")
	}
	// No log_audit_entry → should suggest it.
	if !strings.Contains(suggestionText, "log_audit_entry") {
		t.Error("expected suggestion about log_audit_entry")
	}
}

func TestAnalyzeWorkflow_LargeWorkflow_SplitSuggestion(t *testing.T) {
	// 11 actions → triggers "consider splitting" suggestion.
	actions := `[`
	for i := 0; i < 11; i++ {
		if i > 0 {
			actions += ","
		}
		actions += `{"id":"a` + string(rune('0'+i)) + `","action_type":"send_email"}`
	}
	actions += `]`

	mockRoutes := map[string]string{
		"/v1/workflow/rules/wf-big":         `{"id":"wf-big","name":"big workflow"}`,
		"/v1/workflow/rules/wf-big/actions":  actions,
		"/v1/workflow/rules/wf-big/edges":    `[]`,
	}

	session, ctx := setupToolTest(t,
		pathRouter(mockRoutes),
		tools.RegisterAnalysisTools,
	)

	result := callTool(t, session, ctx, "analyze_workflow", map[string]any{"id": "wf-big"})

	if result.IsError {
		t.Fatalf("returned error: %s", getTextContent(t, result))
	}

	text := getTextContent(t, result)
	if !strings.Contains(text, "splitting") {
		t.Error("expected 'splitting' suggestion for >10 actions")
	}
}

func TestAnalyzeWorkflow_MissingID(t *testing.T) {
	session, ctx := setupToolTest(t,
		staticHandler(`{}`),
		tools.RegisterAnalysisTools,
	)

	// Pass empty string for id to bypass SDK required check but trigger handler validation.
	result := callTool(t, session, ctx, "analyze_workflow", map[string]any{"id": ""})
	if !result.IsError {
		t.Error("analyze_workflow should return error when id is empty")
	}
}

func TestAnalyzeWorkflow_APIError(t *testing.T) {
	session, ctx := setupToolTest(t,
		errorHandler(500),
		tools.RegisterAnalysisTools,
	)

	result := callTool(t, session, ctx, "analyze_workflow", map[string]any{"id": "wf-1"})
	if !result.IsError {
		t.Error("analyze_workflow should return error on API failure")
	}
}

// ===== suggest_templates =====

func TestSuggestTemplates_Success(t *testing.T) {
	mockRoutes := map[string]string{
		"/v1/workflow/templates/active": `[{"id":"t1","name":"email template","is_active":true}]`,
		"/v1/workflow/action-types":     `[{"type":"send_email","category":"notification"}]`,
	}

	session, ctx := setupToolTest(t,
		pathRouter(mockRoutes),
		tools.RegisterAnalysisTools,
	)

	result := callTool(t, session, ctx, "suggest_templates", map[string]any{
		"use_case": "notify when inventory is low",
	})

	if result.IsError {
		t.Fatalf("returned error: %s", getTextContent(t, result))
	}

	text := getTextContent(t, result)
	var resp map[string]any
	if err := json.Unmarshal([]byte(text), &resp); err != nil {
		t.Fatalf("failed to parse result: %v", err)
	}

	if resp["use_case"] != "notify when inventory is low" {
		t.Errorf("use_case = %v, want 'notify when inventory is low'", resp["use_case"])
	}
	if resp["guidance"] == nil {
		t.Error("expected guidance text in response")
	}
	if resp["available_templates"] == nil {
		t.Error("expected available_templates in response")
	}
	if resp["available_action_types"] == nil {
		t.Error("expected available_action_types in response")
	}
}

func TestSuggestTemplates_MissingUseCase(t *testing.T) {
	session, ctx := setupToolTest(t,
		staticHandler(`{}`),
		tools.RegisterAnalysisTools,
	)

	// Pass empty string to bypass SDK required check but trigger handler validation.
	result := callTool(t, session, ctx, "suggest_templates", map[string]any{"use_case": ""})
	if !result.IsError {
		t.Error("suggest_templates should return error when use_case is empty")
	}
}

// ===== show_cascade =====

func TestShowCascade_WithMatches(t *testing.T) {
	rulesJSON := `[
		{"id":"r1","name":"order notify","entity_type":"orders","is_active":true,"trigger_type":"on_create"},
		{"id":"r2","name":"product update","entity_type":"products","is_active":true,"trigger_type":"on_update"},
		{"id":"r3","name":"order approve","entity_type":"orders","is_active":true,"trigger_type":"on_create"}
	]`

	session, ctx := setupToolTest(t,
		pathRouter(map[string]string{"/v1/workflow/rules": rulesJSON}),
		tools.RegisterAnalysisTools,
	)

	result := callTool(t, session, ctx, "show_cascade", map[string]any{"entity": "orders"})

	if result.IsError {
		t.Fatalf("returned error: %s", getTextContent(t, result))
	}

	text := getTextContent(t, result)
	var resp map[string]any
	if err := json.Unmarshal([]byte(text), &resp); err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	if resp["entity"] != "orders" {
		t.Errorf("entity = %v, want 'orders'", resp["entity"])
	}
	if count, ok := resp["rule_count"].(float64); !ok || count != 2 {
		t.Errorf("rule_count = %v, want 2", resp["rule_count"])
	}
	rules, ok := resp["triggered_rules"].([]any)
	if !ok || len(rules) != 2 {
		t.Errorf("triggered_rules has %d items, want 2", len(rules))
	}
}

func TestShowCascade_NoMatches(t *testing.T) {
	rulesJSON := `[
		{"id":"r1","name":"order notify","entity_type":"orders","is_active":true}
	]`

	session, ctx := setupToolTest(t,
		pathRouter(map[string]string{"/v1/workflow/rules": rulesJSON}),
		tools.RegisterAnalysisTools,
	)

	result := callTool(t, session, ctx, "show_cascade", map[string]any{"entity": "users"})

	if result.IsError {
		t.Fatalf("returned error: %s", getTextContent(t, result))
	}

	text := getTextContent(t, result)
	var resp map[string]any
	if err := json.Unmarshal([]byte(text), &resp); err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	if count, ok := resp["rule_count"].(float64); !ok || count != 0 {
		t.Errorf("rule_count = %v, want 0", resp["rule_count"])
	}
	if !strings.Contains(resp["message"].(string), "No workflows") {
		t.Error("expected 'No workflows' message for zero matches")
	}
}

func TestShowCascade_MissingEntity(t *testing.T) {
	session, ctx := setupToolTest(t,
		staticHandler(`{}`),
		tools.RegisterAnalysisTools,
	)

	// Pass empty string to bypass SDK required check but trigger handler validation.
	result := callTool(t, session, ctx, "show_cascade", map[string]any{"entity": ""})
	if !result.IsError {
		t.Error("show_cascade should return error when entity is empty")
	}
}

func TestShowCascade_APIError(t *testing.T) {
	session, ctx := setupToolTest(t,
		errorHandler(500),
		tools.RegisterAnalysisTools,
	)

	result := callTool(t, session, ctx, "show_cascade", map[string]any{"entity": "orders"})
	if !result.IsError {
		t.Error("show_cascade should return error on API failure")
	}
}
