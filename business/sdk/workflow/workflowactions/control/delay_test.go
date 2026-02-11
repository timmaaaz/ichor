package control_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/timmaaaz/ichor/business/sdk/workflow"
	"github.com/timmaaaz/ichor/business/sdk/workflow/workflowactions/control"
)

// =============================================================================
// Validation Tests
// =============================================================================

func TestDelay_Validate_ValidDurations(t *testing.T) {
	handler := control.NewDelayHandler(newTestLogger())

	durations := []string{"30s", "5m", "24h", "168h", "1h30m"}

	for _, d := range durations {
		t.Run(d, func(t *testing.T) {
			config := json.RawMessage(`{"duration": "` + d + `"}`)
			if err := handler.Validate(config); err != nil {
				t.Errorf("Validate(%s) should succeed, got error: %v", d, err)
			}
		})
	}
}

func TestDelay_Validate_ZeroDuration(t *testing.T) {
	handler := control.NewDelayHandler(newTestLogger())

	config := json.RawMessage(`{"duration": "0s"}`)
	err := handler.Validate(config)

	if err == nil {
		t.Error("Validate() should return error for zero duration")
	}
}

func TestDelay_Validate_NegativeDuration(t *testing.T) {
	handler := control.NewDelayHandler(newTestLogger())

	config := json.RawMessage(`{"duration": "-1h"}`)
	err := handler.Validate(config)

	if err == nil {
		t.Error("Validate() should return error for negative duration")
	}
}

func TestDelay_Validate_ExceedsCap(t *testing.T) {
	handler := control.NewDelayHandler(newTestLogger())

	config := json.RawMessage(`{"duration": "721h"}`)
	err := handler.Validate(config)

	if err == nil {
		t.Error("Validate() should return error for duration exceeding 30-day cap")
	}
}

func TestDelay_Validate_InvalidFormat(t *testing.T) {
	handler := control.NewDelayHandler(newTestLogger())

	cases := []string{"abc", "", "1 hour", "1d"}
	for _, d := range cases {
		t.Run(d, func(t *testing.T) {
			config := json.RawMessage(`{"duration": "` + d + `"}`)
			if err := handler.Validate(config); err == nil {
				t.Errorf("Validate(%q) should return error for invalid format", d)
			}
		})
	}
}

func TestDelay_Validate_InvalidJSON(t *testing.T) {
	handler := control.NewDelayHandler(newTestLogger())

	config := json.RawMessage(`{invalid json}`)
	err := handler.Validate(config)

	if err == nil {
		t.Error("Validate() should return error for invalid JSON")
	}
}

// =============================================================================
// Handler Metadata Tests
// =============================================================================

func TestDelay_GetType(t *testing.T) {
	handler := control.NewDelayHandler(newTestLogger())

	if got := handler.GetType(); got != "delay" {
		t.Errorf("GetType() = %s, want delay", got)
	}
}

func TestDelay_SupportsManualExecution(t *testing.T) {
	handler := control.NewDelayHandler(newTestLogger())

	if handler.SupportsManualExecution() {
		t.Error("SupportsManualExecution() should return false for delay handler")
	}
}

func TestDelay_IsAsync(t *testing.T) {
	handler := control.NewDelayHandler(newTestLogger())

	if handler.IsAsync() {
		t.Error("IsAsync() should return false for delay handler")
	}
}

func TestDelay_GetDescription(t *testing.T) {
	handler := control.NewDelayHandler(newTestLogger())

	desc := handler.GetDescription()
	if desc == "" {
		t.Error("GetDescription() should return a non-empty string")
	}
}

// =============================================================================
// Execute Fallback Test
// =============================================================================

func TestDelay_Execute_Fallback(t *testing.T) {
	handler := control.NewDelayHandler(newTestLogger())
	ctx := context.Background()

	config := json.RawMessage(`{"duration": "5m"}`)
	execCtx := workflow.ActionExecutionContext{
		EntityName: "orders",
		EventType:  "on_create",
	}

	result, err := handler.Execute(ctx, config, execCtx)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	resultMap, ok := result.(map[string]any)
	if !ok {
		t.Fatalf("Execute() result is not map[string]any, got %T", result)
	}

	if resultMap["delayed"] != true {
		t.Error("Execute() result should have delayed=true")
	}

	if resultMap["duration"] != "5m" {
		t.Errorf("Execute() result duration = %v, want 5m", resultMap["duration"])
	}
}

// =============================================================================
// ParseDuration Tests
// =============================================================================

func TestParseDuration_BoundaryValues(t *testing.T) {
	// Exactly at max cap (720h = 30 days) should succeed
	if _, err := control.ParseDuration("720h"); err != nil {
		t.Errorf("ParseDuration(720h) should succeed, got: %v", err)
	}

	// Just above max cap should fail
	if _, err := control.ParseDuration("720h1s"); err == nil {
		t.Error("ParseDuration(720h1s) should fail")
	}

	// 1 second should succeed
	if _, err := control.ParseDuration("1s"); err != nil {
		t.Errorf("ParseDuration(1s) should succeed, got: %v", err)
	}
}
