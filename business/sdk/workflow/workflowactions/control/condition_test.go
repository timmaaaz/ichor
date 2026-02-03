// Package control_test contains unit tests for control flow action handlers.
package control_test

import (
	"context"
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/timmaaaz/ichor/business/sdk/workflow"
	"github.com/timmaaaz/ichor/business/sdk/workflow/workflowactions/control"
	"github.com/timmaaaz/ichor/foundation/logger"
	"github.com/timmaaaz/ichor/foundation/otel"
)

/*
Package control_test tests the EvaluateConditionHandler component.

WHAT THIS TESTS:
- Validation of condition configurations
- All supported operators (equals, not_equals, greater_than, less_than, contains, in, is_null, is_not_null, changed_from, changed_to)
- Logic combinations (AND/OR)
- Branch result generation (true_branch/false_branch)
- Edge cases (nil data, missing fields, type mismatches, json.Number handling)
- Handler metadata (GetType, GetDescription, SupportsManualExecution, IsAsync)

WHAT THIS DOES NOT TEST:
- Integration with the workflow execution engine
- Database operations
- RabbitMQ messaging
- Actual workflow branching behavior (covered in executor_graph_test.go)

NOTE: These are pure unit tests that don't require any external services.
Each test creates its own handler instance, so they could theoretically run
in parallel, but we follow the package convention of sequential execution.
*/

// =============================================================================
// Test Setup
// =============================================================================

func newTestLogger() *logger.Logger {
	return logger.New(os.Stdout, logger.LevelInfo, "TEST", func(context.Context) string {
		return otel.GetTraceID(context.Background())
	})
}

func newTestHandler() *control.EvaluateConditionHandler {
	return control.NewEvaluateConditionHandler(newTestLogger())
}

// =============================================================================
// Validation Tests (12.9.1)
// =============================================================================

func TestValidate_EmptyConditions(t *testing.T) {
	handler := newTestHandler()

	config := json.RawMessage(`{"conditions": []}`)
	err := handler.Validate(config)

	if err == nil {
		t.Error("Validate() should return error when no conditions provided")
	}
	if err.Error() != "at least one condition is required" {
		t.Errorf("Validate() error = %q, want %q", err.Error(), "at least one condition is required")
	}
}

func TestValidate_MissingFieldName(t *testing.T) {
	handler := newTestHandler()

	config := json.RawMessage(`{
		"conditions": [
			{"field_name": "", "operator": "equals", "value": "test"}
		]
	}`)
	err := handler.Validate(config)

	if err == nil {
		t.Error("Validate() should return error when field_name is empty")
	}
}

func TestValidate_MissingOperator(t *testing.T) {
	handler := newTestHandler()

	config := json.RawMessage(`{
		"conditions": [
			{"field_name": "status", "operator": "", "value": "test"}
		]
	}`)
	err := handler.Validate(config)

	if err == nil {
		t.Error("Validate() should return error when operator is empty")
	}
}

func TestValidate_InvalidOperator(t *testing.T) {
	handler := newTestHandler()

	config := json.RawMessage(`{
		"conditions": [
			{"field_name": "status", "operator": "invalid_op", "value": "test"}
		]
	}`)
	err := handler.Validate(config)

	if err == nil {
		t.Error("Validate() should return error for unknown operator")
	}
}

func TestValidate_InvalidLogicType(t *testing.T) {
	handler := newTestHandler()

	config := json.RawMessage(`{
		"conditions": [
			{"field_name": "status", "operator": "equals", "value": "active"}
		],
		"logic_type": "invalid"
	}`)
	err := handler.Validate(config)

	if err == nil {
		t.Error("Validate() should return error for invalid logic_type")
	}
}

func TestValidate_ValidConfig(t *testing.T) {
	handler := newTestHandler()

	tests := []struct {
		name   string
		config json.RawMessage
	}{
		{
			name: "single equals condition",
			config: json.RawMessage(`{
				"conditions": [
					{"field_name": "status", "operator": "equals", "value": "active"}
				]
			}`),
		},
		{
			name: "with and logic",
			config: json.RawMessage(`{
				"conditions": [
					{"field_name": "status", "operator": "equals", "value": "active"}
				],
				"logic_type": "and"
			}`),
		},
		{
			name: "with or logic",
			config: json.RawMessage(`{
				"conditions": [
					{"field_name": "status", "operator": "equals", "value": "active"}
				],
				"logic_type": "or"
			}`),
		},
		{
			name: "multiple conditions",
			config: json.RawMessage(`{
				"conditions": [
					{"field_name": "status", "operator": "equals", "value": "active"},
					{"field_name": "amount", "operator": "greater_than", "value": 100}
				],
				"logic_type": "and"
			}`),
		},
		{
			name: "all operators",
			config: json.RawMessage(`{
				"conditions": [
					{"field_name": "f1", "operator": "equals", "value": "v1"},
					{"field_name": "f2", "operator": "not_equals", "value": "v2"},
					{"field_name": "f3", "operator": "greater_than", "value": 10},
					{"field_name": "f4", "operator": "less_than", "value": 20},
					{"field_name": "f5", "operator": "contains", "value": "sub"},
					{"field_name": "f6", "operator": "in", "value": ["a", "b"]},
					{"field_name": "f7", "operator": "is_null"},
					{"field_name": "f8", "operator": "is_not_null"},
					{"field_name": "f9", "operator": "changed_from", "previous_value": "old"},
					{"field_name": "f10", "operator": "changed_to", "value": "new"}
				],
				"logic_type": "or"
			}`),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := handler.Validate(tt.config)
			if err != nil {
				t.Errorf("Validate() should succeed for valid config, got error: %v", err)
			}
		})
	}
}

func TestValidate_InvalidJSON(t *testing.T) {
	handler := newTestHandler()

	config := json.RawMessage(`{invalid json}`)
	err := handler.Validate(config)

	if err == nil {
		t.Error("Validate() should return error for invalid JSON")
	}
}

// =============================================================================
// Operator Tests (12.9.2)
// =============================================================================

func TestOperator_Equals_Match(t *testing.T) {
	handler := newTestHandler()
	ctx := context.Background()

	config := json.RawMessage(`{
		"conditions": [{"field_name": "status", "operator": "equals", "value": "active"}]
	}`)

	execCtx := workflow.ActionExecutionContext{
		EntityName: "orders",
		EventType:  "on_create",
		RawData:    map[string]interface{}{"status": "active"},
	}

	result, err := handler.Execute(ctx, config, execCtx)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	condResult, ok := result.(workflow.ConditionResult)
	if !ok {
		t.Fatalf("Execute() result is not ConditionResult, got %T", result)
	}

	if !condResult.Result {
		t.Error("equals operator should match 'active' == 'active'")
	}
	if condResult.BranchTaken != workflow.EdgeTypeTrueBranch {
		t.Errorf("BranchTaken = %s, want %s", condResult.BranchTaken, workflow.EdgeTypeTrueBranch)
	}
}

func TestOperator_Equals_NoMatch(t *testing.T) {
	handler := newTestHandler()
	ctx := context.Background()

	config := json.RawMessage(`{
		"conditions": [{"field_name": "status", "operator": "equals", "value": "active"}]
	}`)

	execCtx := workflow.ActionExecutionContext{
		EntityName: "orders",
		EventType:  "on_create",
		RawData:    map[string]interface{}{"status": "inactive"},
	}

	result, err := handler.Execute(ctx, config, execCtx)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	condResult := result.(workflow.ConditionResult)
	if condResult.Result {
		t.Error("equals operator should not match 'inactive' == 'active'")
	}
	if condResult.BranchTaken != workflow.EdgeTypeFalseBranch {
		t.Errorf("BranchTaken = %s, want %s", condResult.BranchTaken, workflow.EdgeTypeFalseBranch)
	}
}

func TestOperator_NotEquals_Match(t *testing.T) {
	handler := newTestHandler()
	ctx := context.Background()

	config := json.RawMessage(`{
		"conditions": [{"field_name": "status", "operator": "not_equals", "value": "inactive"}]
	}`)

	execCtx := workflow.ActionExecutionContext{
		EntityName: "orders",
		EventType:  "on_create",
		RawData:    map[string]interface{}{"status": "active"},
	}

	result, err := handler.Execute(ctx, config, execCtx)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	condResult := result.(workflow.ConditionResult)
	if !condResult.Result {
		t.Error("not_equals operator should match 'active' != 'inactive'")
	}
}

func TestOperator_NotEquals_NoMatch(t *testing.T) {
	handler := newTestHandler()
	ctx := context.Background()

	config := json.RawMessage(`{
		"conditions": [{"field_name": "status", "operator": "not_equals", "value": "active"}]
	}`)

	execCtx := workflow.ActionExecutionContext{
		EntityName: "orders",
		EventType:  "on_create",
		RawData:    map[string]interface{}{"status": "active"},
	}

	result, err := handler.Execute(ctx, config, execCtx)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	condResult := result.(workflow.ConditionResult)
	if condResult.Result {
		t.Error("not_equals operator should not match 'active' != 'active'")
	}
}

func TestOperator_GreaterThan_Numeric(t *testing.T) {
	handler := newTestHandler()
	ctx := context.Background()

	tests := []struct {
		name   string
		data   map[string]interface{}
		value  float64
		expect bool
	}{
		{"150 > 100", map[string]interface{}{"amount": float64(150)}, 100, true},
		{"100 > 100", map[string]interface{}{"amount": float64(100)}, 100, false},
		{"50 > 100", map[string]interface{}{"amount": float64(50)}, 100, false},
		{"int 150 > 100", map[string]interface{}{"amount": 150}, 100, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := json.RawMessage(`{
				"conditions": [{"field_name": "amount", "operator": "greater_than", "value": ` + jsonNumber(tt.value) + `}]
			}`)

			execCtx := workflow.ActionExecutionContext{
				EntityName: "orders",
				EventType:  "on_create",
				RawData:    tt.data,
			}

			result, err := handler.Execute(ctx, config, execCtx)
			if err != nil {
				t.Fatalf("Execute() error = %v", err)
			}

			condResult := result.(workflow.ConditionResult)
			if condResult.Result != tt.expect {
				t.Errorf("greater_than result = %v, want %v", condResult.Result, tt.expect)
			}
		})
	}
}

func TestOperator_GreaterThan_String(t *testing.T) {
	handler := newTestHandler()
	ctx := context.Background()

	config := json.RawMessage(`{
		"conditions": [{"field_name": "grade", "operator": "greater_than", "value": "a"}]
	}`)

	execCtx := workflow.ActionExecutionContext{
		EntityName: "students",
		EventType:  "on_create",
		RawData:    map[string]interface{}{"grade": "b"},
	}

	result, err := handler.Execute(ctx, config, execCtx)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	condResult := result.(workflow.ConditionResult)
	if !condResult.Result {
		t.Error("greater_than operator should match 'b' > 'a' (string comparison)")
	}
}

func TestOperator_LessThan_Numeric(t *testing.T) {
	handler := newTestHandler()
	ctx := context.Background()

	tests := []struct {
		name   string
		amount float64
		limit  float64
		expect bool
	}{
		{"50 < 100", 50, 100, true},
		{"100 < 100", 100, 100, false},
		{"150 < 100", 150, 100, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := json.RawMessage(`{
				"conditions": [{"field_name": "amount", "operator": "less_than", "value": ` + jsonNumber(tt.limit) + `}]
			}`)

			execCtx := workflow.ActionExecutionContext{
				EntityName: "orders",
				EventType:  "on_create",
				RawData:    map[string]interface{}{"amount": tt.amount},
			}

			result, err := handler.Execute(ctx, config, execCtx)
			if err != nil {
				t.Fatalf("Execute() error = %v", err)
			}

			condResult := result.(workflow.ConditionResult)
			if condResult.Result != tt.expect {
				t.Errorf("less_than result = %v, want %v", condResult.Result, tt.expect)
			}
		})
	}
}

func TestOperator_Contains_Match(t *testing.T) {
	handler := newTestHandler()
	ctx := context.Background()

	config := json.RawMessage(`{
		"conditions": [{"field_name": "description", "operator": "contains", "value": "world"}]
	}`)

	execCtx := workflow.ActionExecutionContext{
		EntityName: "items",
		EventType:  "on_create",
		RawData:    map[string]interface{}{"description": "hello world"},
	}

	result, err := handler.Execute(ctx, config, execCtx)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	condResult := result.(workflow.ConditionResult)
	if !condResult.Result {
		t.Error("contains operator should match 'hello world' contains 'world'")
	}
}

func TestOperator_Contains_NoMatch(t *testing.T) {
	handler := newTestHandler()
	ctx := context.Background()

	config := json.RawMessage(`{
		"conditions": [{"field_name": "description", "operator": "contains", "value": "world"}]
	}`)

	execCtx := workflow.ActionExecutionContext{
		EntityName: "items",
		EventType:  "on_create",
		RawData:    map[string]interface{}{"description": "hello there"},
	}

	result, err := handler.Execute(ctx, config, execCtx)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	condResult := result.(workflow.ConditionResult)
	if condResult.Result {
		t.Error("contains operator should not match 'hello there' contains 'world'")
	}
}

func TestOperator_In_Match(t *testing.T) {
	handler := newTestHandler()
	ctx := context.Background()

	config := json.RawMessage(`{
		"conditions": [{"field_name": "color", "operator": "in", "value": ["red", "green", "blue"]}]
	}`)

	execCtx := workflow.ActionExecutionContext{
		EntityName: "items",
		EventType:  "on_create",
		RawData:    map[string]interface{}{"color": "red"},
	}

	result, err := handler.Execute(ctx, config, execCtx)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	condResult := result.(workflow.ConditionResult)
	if !condResult.Result {
		t.Error("in operator should match 'red' in ['red', 'green', 'blue']")
	}
}

func TestOperator_In_NoMatch(t *testing.T) {
	handler := newTestHandler()
	ctx := context.Background()

	config := json.RawMessage(`{
		"conditions": [{"field_name": "color", "operator": "in", "value": ["red", "green", "blue"]}]
	}`)

	execCtx := workflow.ActionExecutionContext{
		EntityName: "items",
		EventType:  "on_create",
		RawData:    map[string]interface{}{"color": "yellow"},
	}

	result, err := handler.Execute(ctx, config, execCtx)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	condResult := result.(workflow.ConditionResult)
	if condResult.Result {
		t.Error("in operator should not match 'yellow' in ['red', 'green', 'blue']")
	}
}

func TestOperator_IsNull_True(t *testing.T) {
	handler := newTestHandler()
	ctx := context.Background()

	config := json.RawMessage(`{
		"conditions": [{"field_name": "optional_field", "operator": "is_null"}]
	}`)

	execCtx := workflow.ActionExecutionContext{
		EntityName: "items",
		EventType:  "on_create",
		RawData:    map[string]interface{}{"optional_field": nil},
	}

	result, err := handler.Execute(ctx, config, execCtx)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	condResult := result.(workflow.ConditionResult)
	if !condResult.Result {
		t.Error("is_null operator should match nil value")
	}
}

func TestOperator_IsNull_False(t *testing.T) {
	handler := newTestHandler()
	ctx := context.Background()

	config := json.RawMessage(`{
		"conditions": [{"field_name": "optional_field", "operator": "is_null"}]
	}`)

	execCtx := workflow.ActionExecutionContext{
		EntityName: "items",
		EventType:  "on_create",
		RawData:    map[string]interface{}{"optional_field": "value"},
	}

	result, err := handler.Execute(ctx, config, execCtx)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	condResult := result.(workflow.ConditionResult)
	if condResult.Result {
		t.Error("is_null operator should not match non-nil value")
	}
}

func TestOperator_IsNotNull_True(t *testing.T) {
	handler := newTestHandler()
	ctx := context.Background()

	config := json.RawMessage(`{
		"conditions": [{"field_name": "required_field", "operator": "is_not_null"}]
	}`)

	execCtx := workflow.ActionExecutionContext{
		EntityName: "items",
		EventType:  "on_create",
		RawData:    map[string]interface{}{"required_field": "value"},
	}

	result, err := handler.Execute(ctx, config, execCtx)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	condResult := result.(workflow.ConditionResult)
	if !condResult.Result {
		t.Error("is_not_null operator should match non-nil value")
	}
}

func TestOperator_IsNotNull_False(t *testing.T) {
	handler := newTestHandler()
	ctx := context.Background()

	config := json.RawMessage(`{
		"conditions": [{"field_name": "required_field", "operator": "is_not_null"}]
	}`)

	execCtx := workflow.ActionExecutionContext{
		EntityName: "items",
		EventType:  "on_create",
		RawData:    map[string]interface{}{"required_field": nil},
	}

	result, err := handler.Execute(ctx, config, execCtx)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	condResult := result.(workflow.ConditionResult)
	if condResult.Result {
		t.Error("is_not_null operator should not match nil value")
	}
}

func TestOperator_ChangedFrom_Match(t *testing.T) {
	handler := newTestHandler()
	ctx := context.Background()

	config := json.RawMessage(`{
		"conditions": [{"field_name": "status", "operator": "changed_from", "previous_value": "draft"}]
	}`)

	execCtx := workflow.ActionExecutionContext{
		EntityName: "orders",
		EventType:  "on_update",
		RawData:    map[string]interface{}{"status": "pending"},
		FieldChanges: map[string]workflow.FieldChange{
			"status": {OldValue: "draft", NewValue: "pending"},
		},
	}

	result, err := handler.Execute(ctx, config, execCtx)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	condResult := result.(workflow.ConditionResult)
	if !condResult.Result {
		t.Error("changed_from operator should match when old value was 'draft'")
	}
}

func TestOperator_ChangedFrom_NoUpdate(t *testing.T) {
	handler := newTestHandler()
	ctx := context.Background()

	config := json.RawMessage(`{
		"conditions": [{"field_name": "status", "operator": "changed_from", "previous_value": "draft"}]
	}`)

	// On create event, changed_from should return false
	execCtx := workflow.ActionExecutionContext{
		EntityName: "orders",
		EventType:  "on_create",
		RawData:    map[string]interface{}{"status": "pending"},
	}

	result, err := handler.Execute(ctx, config, execCtx)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	condResult := result.(workflow.ConditionResult)
	if condResult.Result {
		t.Error("changed_from operator should return false on on_create event")
	}
}

func TestOperator_ChangedTo_Match(t *testing.T) {
	handler := newTestHandler()
	ctx := context.Background()

	config := json.RawMessage(`{
		"conditions": [{"field_name": "status", "operator": "changed_to", "value": "shipped"}]
	}`)

	execCtx := workflow.ActionExecutionContext{
		EntityName: "orders",
		EventType:  "on_update",
		RawData:    map[string]interface{}{"status": "shipped"},
		FieldChanges: map[string]workflow.FieldChange{
			"status": {OldValue: "pending", NewValue: "shipped"},
		},
	}

	result, err := handler.Execute(ctx, config, execCtx)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	condResult := result.(workflow.ConditionResult)
	if !condResult.Result {
		t.Error("changed_to operator should match when new value is 'shipped'")
	}
}

func TestOperator_ChangedTo_SameValue(t *testing.T) {
	handler := newTestHandler()
	ctx := context.Background()

	config := json.RawMessage(`{
		"conditions": [{"field_name": "status", "operator": "changed_to", "value": "shipped"}]
	}`)

	// When old and new are same, changed_to should return false
	execCtx := workflow.ActionExecutionContext{
		EntityName: "orders",
		EventType:  "on_update",
		RawData:    map[string]interface{}{"status": "shipped"},
		FieldChanges: map[string]workflow.FieldChange{
			"status": {OldValue: "shipped", NewValue: "shipped"},
		},
	}

	result, err := handler.Execute(ctx, config, execCtx)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	condResult := result.(workflow.ConditionResult)
	if condResult.Result {
		t.Error("changed_to operator should return false when old value equals new value")
	}
}

// =============================================================================
// Logic Combination Tests (12.9.3)
// =============================================================================

func TestLogic_And_AllTrue(t *testing.T) {
	handler := newTestHandler()
	ctx := context.Background()

	config := json.RawMessage(`{
		"conditions": [
			{"field_name": "status", "operator": "equals", "value": "active"},
			{"field_name": "amount", "operator": "greater_than", "value": 100}
		],
		"logic_type": "and"
	}`)

	execCtx := workflow.ActionExecutionContext{
		EntityName: "orders",
		EventType:  "on_create",
		RawData: map[string]interface{}{
			"status": "active",
			"amount": float64(150),
		},
	}

	result, err := handler.Execute(ctx, config, execCtx)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	condResult := result.(workflow.ConditionResult)
	if !condResult.Result {
		t.Error("AND logic: all conditions true should return true")
	}
}

func TestLogic_And_OneFalse(t *testing.T) {
	handler := newTestHandler()
	ctx := context.Background()

	config := json.RawMessage(`{
		"conditions": [
			{"field_name": "status", "operator": "equals", "value": "active"},
			{"field_name": "amount", "operator": "greater_than", "value": 100}
		],
		"logic_type": "and"
	}`)

	execCtx := workflow.ActionExecutionContext{
		EntityName: "orders",
		EventType:  "on_create",
		RawData: map[string]interface{}{
			"status": "inactive", // This condition will fail
			"amount": float64(150),
		},
	}

	result, err := handler.Execute(ctx, config, execCtx)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	condResult := result.(workflow.ConditionResult)
	if condResult.Result {
		t.Error("AND logic: one false condition should return false")
	}
}

func TestLogic_Or_AllFalse(t *testing.T) {
	handler := newTestHandler()
	ctx := context.Background()

	config := json.RawMessage(`{
		"conditions": [
			{"field_name": "status", "operator": "equals", "value": "active"},
			{"field_name": "amount", "operator": "greater_than", "value": 200}
		],
		"logic_type": "or"
	}`)

	execCtx := workflow.ActionExecutionContext{
		EntityName: "orders",
		EventType:  "on_create",
		RawData: map[string]interface{}{
			"status": "inactive",   // fails
			"amount": float64(100), // fails (not > 200)
		},
	}

	result, err := handler.Execute(ctx, config, execCtx)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	condResult := result.(workflow.ConditionResult)
	if condResult.Result {
		t.Error("OR logic: all false conditions should return false")
	}
}

func TestLogic_Or_OneTrue(t *testing.T) {
	handler := newTestHandler()
	ctx := context.Background()

	config := json.RawMessage(`{
		"conditions": [
			{"field_name": "status", "operator": "equals", "value": "active"},
			{"field_name": "amount", "operator": "greater_than", "value": 200}
		],
		"logic_type": "or"
	}`)

	execCtx := workflow.ActionExecutionContext{
		EntityName: "orders",
		EventType:  "on_create",
		RawData: map[string]interface{}{
			"status": "active",     // true
			"amount": float64(100), // false
		},
	}

	result, err := handler.Execute(ctx, config, execCtx)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	condResult := result.(workflow.ConditionResult)
	if !condResult.Result {
		t.Error("OR logic: one true condition should return true")
	}
}

func TestLogic_DefaultIsAnd(t *testing.T) {
	handler := newTestHandler()
	ctx := context.Background()

	// No logic_type specified, should default to AND
	config := json.RawMessage(`{
		"conditions": [
			{"field_name": "status", "operator": "equals", "value": "active"},
			{"field_name": "amount", "operator": "greater_than", "value": 100}
		]
	}`)

	// One condition false - should fail with AND logic
	execCtx := workflow.ActionExecutionContext{
		EntityName: "orders",
		EventType:  "on_create",
		RawData: map[string]interface{}{
			"status": "inactive",   // false
			"amount": float64(150), // true
		},
	}

	result, err := handler.Execute(ctx, config, execCtx)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	condResult := result.(workflow.ConditionResult)
	if condResult.Result {
		t.Error("Default logic should be AND: one false condition should return false")
	}
}

// =============================================================================
// Branch Result Tests (12.9.4)
// =============================================================================

func TestExecute_ReturnsConditionResult(t *testing.T) {
	handler := newTestHandler()
	ctx := context.Background()

	config := json.RawMessage(`{
		"conditions": [{"field_name": "status", "operator": "equals", "value": "active"}]
	}`)

	execCtx := workflow.ActionExecutionContext{
		EntityName: "orders",
		EventType:  "on_create",
		RawData:    map[string]interface{}{"status": "active"},
	}

	result, err := handler.Execute(ctx, config, execCtx)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	condResult, ok := result.(workflow.ConditionResult)
	if !ok {
		t.Fatalf("Execute() should return ConditionResult, got %T", result)
	}

	if !condResult.Evaluated {
		t.Error("ConditionResult.Evaluated should be true")
	}
}

func TestExecute_TrueBranch(t *testing.T) {
	handler := newTestHandler()
	ctx := context.Background()

	config := json.RawMessage(`{
		"conditions": [{"field_name": "status", "operator": "equals", "value": "active"}]
	}`)

	execCtx := workflow.ActionExecutionContext{
		EntityName: "orders",
		EventType:  "on_create",
		RawData:    map[string]interface{}{"status": "active"},
	}

	result, err := handler.Execute(ctx, config, execCtx)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	condResult := result.(workflow.ConditionResult)
	if condResult.BranchTaken != workflow.EdgeTypeTrueBranch {
		t.Errorf("BranchTaken = %s, want %s", condResult.BranchTaken, workflow.EdgeTypeTrueBranch)
	}
	if !condResult.Result {
		t.Error("Result should be true when branch is true_branch")
	}
}

func TestExecute_FalseBranch(t *testing.T) {
	handler := newTestHandler()
	ctx := context.Background()

	config := json.RawMessage(`{
		"conditions": [{"field_name": "status", "operator": "equals", "value": "active"}]
	}`)

	execCtx := workflow.ActionExecutionContext{
		EntityName: "orders",
		EventType:  "on_create",
		RawData:    map[string]interface{}{"status": "inactive"},
	}

	result, err := handler.Execute(ctx, config, execCtx)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	condResult := result.(workflow.ConditionResult)
	if condResult.BranchTaken != workflow.EdgeTypeFalseBranch {
		t.Errorf("BranchTaken = %s, want %s", condResult.BranchTaken, workflow.EdgeTypeFalseBranch)
	}
	if condResult.Result {
		t.Error("Result should be false when branch is false_branch")
	}
}

// =============================================================================
// Edge Cases (12.9.5)
// =============================================================================

func TestEdgeCase_NilData(t *testing.T) {
	handler := newTestHandler()
	ctx := context.Background()

	config := json.RawMessage(`{
		"conditions": [{"field_name": "status", "operator": "equals", "value": "active"}]
	}`)

	execCtx := workflow.ActionExecutionContext{
		EntityName: "orders",
		EventType:  "on_create",
		RawData:    nil, // nil data
	}

	result, err := handler.Execute(ctx, config, execCtx)
	if err != nil {
		t.Fatalf("Execute() should handle nil data gracefully, got error: %v", err)
	}

	condResult := result.(workflow.ConditionResult)
	// With nil data, the field value will be nil, which won't equal "active"
	if condResult.Result {
		t.Error("With nil data, condition should evaluate to false")
	}
}

func TestEdgeCase_MissingField(t *testing.T) {
	handler := newTestHandler()
	ctx := context.Background()

	config := json.RawMessage(`{
		"conditions": [{"field_name": "missing_field", "operator": "equals", "value": "test"}]
	}`)

	execCtx := workflow.ActionExecutionContext{
		EntityName: "orders",
		EventType:  "on_create",
		RawData:    map[string]interface{}{"other_field": "value"},
	}

	result, err := handler.Execute(ctx, config, execCtx)
	if err != nil {
		t.Fatalf("Execute() should handle missing field gracefully, got error: %v", err)
	}

	condResult := result.(workflow.ConditionResult)
	// Missing field should be treated as nil, which won't equal "test"
	if condResult.Result {
		t.Error("With missing field, condition should evaluate to false")
	}
}

func TestEdgeCase_TypeMismatch(t *testing.T) {
	handler := newTestHandler()
	ctx := context.Background()

	// String value in config but numeric in data
	config := json.RawMessage(`{
		"conditions": [{"field_name": "count", "operator": "equals", "value": "100"}]
	}`)

	execCtx := workflow.ActionExecutionContext{
		EntityName: "orders",
		EventType:  "on_create",
		RawData:    map[string]interface{}{"count": 100}, // numeric
	}

	result, err := handler.Execute(ctx, config, execCtx)
	if err != nil {
		t.Fatalf("Execute() should handle type mismatch gracefully, got error: %v", err)
	}

	condResult := result.(workflow.ConditionResult)
	// The comparison uses fmt.Sprintf to convert both to strings, so "100" == "100" should match
	if !condResult.Result {
		t.Error("Type mismatch should still match when string representations are equal")
	}
}

func TestEdgeCase_JsonNumber(t *testing.T) {
	handler := newTestHandler()
	ctx := context.Background()

	config := json.RawMessage(`{
		"conditions": [{"field_name": "amount", "operator": "greater_than", "value": 100}]
	}`)

	// Simulate JSON unmarshaling which produces json.Number
	var rawData map[string]interface{}
	jsonData := `{"amount": 150}`
	dec := json.NewDecoder(strings.NewReader(jsonData))
	dec.UseNumber()
	if err := dec.Decode(&rawData); err != nil {
		t.Fatalf("Failed to decode JSON: %v", err)
	}

	execCtx := workflow.ActionExecutionContext{
		EntityName: "orders",
		EventType:  "on_create",
		RawData:    rawData,
	}

	result, err := handler.Execute(ctx, config, execCtx)
	if err != nil {
		t.Fatalf("Execute() should handle json.Number, got error: %v", err)
	}

	condResult := result.(workflow.ConditionResult)
	if !condResult.Result {
		t.Error("json.Number comparison should work correctly")
	}
}

func TestEdgeCase_EmptyString(t *testing.T) {
	handler := newTestHandler()
	ctx := context.Background()

	config := json.RawMessage(`{
		"conditions": [{"field_name": "name", "operator": "equals", "value": ""}]
	}`)

	execCtx := workflow.ActionExecutionContext{
		EntityName: "orders",
		EventType:  "on_create",
		RawData:    map[string]interface{}{"name": ""},
	}

	result, err := handler.Execute(ctx, config, execCtx)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	condResult := result.(workflow.ConditionResult)
	if !condResult.Result {
		t.Error("Empty string should equal empty string")
	}
}

func TestEdgeCase_ContainsEmptyString(t *testing.T) {
	handler := newTestHandler()
	ctx := context.Background()

	config := json.RawMessage(`{
		"conditions": [{"field_name": "text", "operator": "contains", "value": ""}]
	}`)

	execCtx := workflow.ActionExecutionContext{
		EntityName: "items",
		EventType:  "on_create",
		RawData:    map[string]interface{}{"text": "hello"},
	}

	result, err := handler.Execute(ctx, config, execCtx)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	condResult := result.(workflow.ConditionResult)
	// Any string contains the empty string
	if !condResult.Result {
		t.Error("Any string should contain the empty string")
	}
}

func TestEdgeCase_ContainsNonString(t *testing.T) {
	handler := newTestHandler()
	ctx := context.Background()

	// Non-string value in field
	config := json.RawMessage(`{
		"conditions": [{"field_name": "count", "operator": "contains", "value": "5"}]
	}`)

	execCtx := workflow.ActionExecutionContext{
		EntityName: "items",
		EventType:  "on_create",
		RawData:    map[string]interface{}{"count": 150},
	}

	result, err := handler.Execute(ctx, config, execCtx)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	condResult := result.(workflow.ConditionResult)
	// Non-string values should return false for contains
	if condResult.Result {
		t.Error("Contains with non-string field value should return false")
	}
}

func TestEdgeCase_InWithEmptyArray(t *testing.T) {
	handler := newTestHandler()
	ctx := context.Background()

	config := json.RawMessage(`{
		"conditions": [{"field_name": "color", "operator": "in", "value": []}]
	}`)

	execCtx := workflow.ActionExecutionContext{
		EntityName: "items",
		EventType:  "on_create",
		RawData:    map[string]interface{}{"color": "red"},
	}

	result, err := handler.Execute(ctx, config, execCtx)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	condResult := result.(workflow.ConditionResult)
	// Nothing is "in" an empty array
	if condResult.Result {
		t.Error("In with empty array should return false")
	}
}

func TestEdgeCase_IsNullMissingField(t *testing.T) {
	handler := newTestHandler()
	ctx := context.Background()

	config := json.RawMessage(`{
		"conditions": [{"field_name": "missing", "operator": "is_null"}]
	}`)

	execCtx := workflow.ActionExecutionContext{
		EntityName: "items",
		EventType:  "on_create",
		RawData:    map[string]interface{}{"other": "value"},
	}

	result, err := handler.Execute(ctx, config, execCtx)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	condResult := result.(workflow.ConditionResult)
	// Missing field should be treated as nil
	if !condResult.Result {
		t.Error("is_null should return true for missing field")
	}
}

// =============================================================================
// Handler Metadata Tests
// =============================================================================

func TestHandler_GetType(t *testing.T) {
	handler := newTestHandler()

	if got := handler.GetType(); got != "evaluate_condition" {
		t.Errorf("GetType() = %s, want evaluate_condition", got)
	}
}

func TestHandler_GetDescription(t *testing.T) {
	handler := newTestHandler()

	desc := handler.GetDescription()
	if desc == "" {
		t.Error("GetDescription() should return a non-empty string")
	}
}

func TestHandler_SupportsManualExecution(t *testing.T) {
	handler := newTestHandler()

	if handler.SupportsManualExecution() {
		t.Error("SupportsManualExecution() should return false for condition handler")
	}
}

func TestHandler_IsAsync(t *testing.T) {
	handler := newTestHandler()

	if handler.IsAsync() {
		t.Error("IsAsync() should return false for condition handler")
	}
}

// =============================================================================
// Helper Functions
// =============================================================================

func jsonNumber(n float64) string {
	b, _ := json.Marshal(n)
	return string(b)
}
