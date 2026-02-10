package temporal

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// =============================================================================
// shouldContinueAsNew Threshold Tests
// =============================================================================

func TestShouldContinueAsNew_AtThreshold(t *testing.T) {
	// Threshold is 10,000. Implementation uses > (not >=).
	require.False(t, shouldContinueAsNew(HistoryLengthThreshold),
		"exactly at threshold should NOT trigger (uses > not >=)")
}

func TestShouldContinueAsNew_OverThreshold(t *testing.T) {
	require.True(t, shouldContinueAsNew(HistoryLengthThreshold+1),
		"one over threshold should trigger")
}

func TestShouldContinueAsNew_UnderThreshold(t *testing.T) {
	require.False(t, shouldContinueAsNew(0), "zero should not trigger")
	require.False(t, shouldContinueAsNew(100), "small value should not trigger")
	require.False(t, shouldContinueAsNew(HistoryLengthThreshold-1), "one under should not trigger")
}

// =============================================================================
// ContinuationState JSON Round-Trip Preservation
// =============================================================================

func TestContinuationState_ActionResultsRoundTrip(t *testing.T) {
	// Verify ActionResults map survives JSON round-trip through ContinuationState.
	// This is the critical path for Continue-As-New: Temporal serializes WorkflowInput
	// to JSON, stores it, then deserializes for the new workflow execution.
	original := &MergedContext{
		TriggerData: map[string]any{"key": "value"},
		ActionResults: map[string]map[string]any{
			"action1": {"result": "data", "count": float64(42)},
			"action2": {"nested": map[string]any{"deep": "value"}},
		},
		Flattened: map[string]any{
			"key":            "value",
			"action1.result": "data",
			"action1.count":  float64(42),
			"action2.nested": map[string]any{"deep": "value"},
		},
	}

	data, err := json.Marshal(original)
	require.NoError(t, err)

	var restored MergedContext
	require.NoError(t, json.Unmarshal(data, &restored))

	require.Equal(t, original.TriggerData["key"], restored.TriggerData["key"])
	require.Equal(t, original.ActionResults["action1"]["result"], restored.ActionResults["action1"]["result"])
	require.Equal(t, original.ActionResults["action1"]["count"], restored.ActionResults["action1"]["count"])
	require.Equal(t, original.Flattened["action1.result"], restored.Flattened["action1.result"])
}

func TestContinuationState_LargeActionResultsMap(t *testing.T) {
	// Simulate a long-running workflow with 100+ completed actions.
	original := &MergedContext{
		TriggerData:   map[string]any{"entity_id": "test-123"},
		ActionResults: make(map[string]map[string]any),
		Flattened:     make(map[string]any),
	}
	original.Flattened["entity_id"] = "test-123"

	for i := 0; i < 100; i++ {
		name := fmt.Sprintf("action_%d", i)
		original.ActionResults[name] = map[string]any{
			"status": "done",
			"index":  float64(i),
			"data":   strings.Repeat("x", 100),
		}
		original.Flattened[name+".status"] = "done"
		original.Flattened[name+".index"] = float64(i)
	}

	data, err := json.Marshal(original)
	require.NoError(t, err)

	var restored MergedContext
	require.NoError(t, json.Unmarshal(data, &restored))

	require.Equal(t, 100, len(restored.ActionResults))
	require.Equal(t, "done", restored.ActionResults["action_0"]["status"])
	require.Equal(t, "done", restored.ActionResults["action_99"]["status"])
	require.Equal(t, float64(99), restored.ActionResults["action_99"]["index"])
}

func TestContinuationState_Float64Precision(t *testing.T) {
	// JSON deserializes all numbers as float64. Verify large integers
	// (>2^53) survive as float64 (known precision limitation, documented).
	original := &MergedContext{
		TriggerData: map[string]any{
			"small_int": float64(42),
			"large_int": float64(9007199254740992), // 2^53
			"float_val": float64(3.14159265358979),
			"negative":  float64(-1000),
		},
		ActionResults: make(map[string]map[string]any),
		Flattened:     make(map[string]any),
	}

	data, err := json.Marshal(original)
	require.NoError(t, err)

	var restored MergedContext
	require.NoError(t, json.Unmarshal(data, &restored))

	require.Equal(t, float64(42), restored.TriggerData["small_int"])
	require.Equal(t, float64(9007199254740992), restored.TriggerData["large_int"])
	require.InDelta(t, 3.14159265358979, restored.TriggerData["float_val"].(float64), 1e-15)
	require.Equal(t, float64(-1000), restored.TriggerData["negative"])
}

func TestContinuationState_EmptyActionResults(t *testing.T) {
	// Workflow just started, no actions completed yet.
	original := &MergedContext{
		TriggerData:   map[string]any{"entity_id": "123"},
		ActionResults: map[string]map[string]any{},
		Flattened:     map[string]any{"entity_id": "123"},
	}

	data, err := json.Marshal(original)
	require.NoError(t, err)

	var restored MergedContext
	require.NoError(t, json.Unmarshal(data, &restored))

	require.NotNil(t, restored.TriggerData)
	require.Equal(t, "123", restored.TriggerData["entity_id"])
}

func TestContinuationState_NilState_CreatesFreshContext(t *testing.T) {
	// First execution: ContinuationState is nil.
	// ExecuteGraphWorkflow creates fresh MergedContext from TriggerData.
	triggerData := map[string]any{"entity_id": "abc", "event_type": "on_create"}
	ctx := NewMergedContext(triggerData)

	require.Equal(t, "abc", ctx.Flattened["entity_id"])
	require.Equal(t, "on_create", ctx.Flattened["event_type"])
	require.Empty(t, ctx.ActionResults)
}

func TestContinuationState_SpecialCharactersInKeys(t *testing.T) {
	// Action names with dots, unicode, or special chars in keys.
	original := &MergedContext{
		TriggerData: map[string]any{"key.with.dots": "value"},
		ActionResults: map[string]map[string]any{
			"action_with.dots": {"status": "done"},
			"unicode_名前":       {"result": "成功"},
		},
		Flattened: map[string]any{
			"key.with.dots":           "value",
			"action_with.dots.status": "done",
			"unicode_名前.result":      "成功",
		},
	}

	data, err := json.Marshal(original)
	require.NoError(t, err)

	var restored MergedContext
	require.NoError(t, json.Unmarshal(data, &restored))

	require.Equal(t, "value", restored.TriggerData["key.with.dots"])
	require.Equal(t, "done", restored.ActionResults["action_with.dots"]["status"])
	require.Equal(t, "成功", restored.ActionResults["unicode_名前"]["result"])
}
