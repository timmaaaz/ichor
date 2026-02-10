package temporal

import (
	"strings"
	"testing"

	"github.com/google/uuid"
)

func TestNewMergedContext(t *testing.T) {
	triggerData := map[string]any{
		"entity_id":   "abc-123",
		"entity_name": "orders",
		"event_type":  "on_create",
	}

	ctx := NewMergedContext(triggerData)

	// Trigger data should be in Flattened
	if ctx.Flattened["entity_id"] != "abc-123" {
		t.Errorf("expected entity_id in Flattened, got %v", ctx.Flattened["entity_id"])
	}

	// ActionResults should be empty
	if len(ctx.ActionResults) != 0 {
		t.Errorf("expected empty ActionResults, got %d", len(ctx.ActionResults))
	}
}

func TestMergeResult(t *testing.T) {
	ctx := NewMergedContext(map[string]any{"trigger": "data"})

	result := map[string]any{
		"status":  "sent",
		"message": "Email sent successfully",
	}
	ctx.MergeResult("send_email", result)

	// Check ActionResults
	if ctx.ActionResults["send_email"]["status"] != "sent" {
		t.Error("expected status in ActionResults")
	}

	// Check Flattened access patterns
	if ctx.Flattened["send_email.status"] != "sent" {
		t.Error("expected send_email.status in Flattened")
	}
	if ctx.Flattened["send_email.message"] != "Email sent successfully" {
		t.Error("expected send_email.message in Flattened")
	}

	// Check full result accessible by action name
	fullResult, ok := ctx.Flattened["send_email"].(map[string]any)
	if !ok {
		t.Fatal("expected send_email to be map in Flattened")
	}
	if fullResult["status"] != "sent" {
		t.Error("expected full result accessible by action name")
	}
}

func TestMergeResultMultipleActions(t *testing.T) {
	ctx := NewMergedContext(nil)

	ctx.MergeResult("action_a", map[string]any{"value": "first"})
	ctx.MergeResult("action_b", map[string]any{"value": "second"})

	if ctx.Flattened["action_a.value"] != "first" {
		t.Error("first action result missing")
	}
	if ctx.Flattened["action_b.value"] != "second" {
		t.Error("second action result missing")
	}
}

func TestMergeResultBranchTaken(t *testing.T) {
	// Condition actions store branch_taken in their result.
	// This is how the graph executor determines which edge to follow.
	ctx := NewMergedContext(nil)

	conditionResult := map[string]any{
		"evaluated":    true,
		"result":       true,
		"branch_taken": "true_branch",
	}
	ctx.MergeResult("check_status", conditionResult)

	// Graph executor reads this to decide true_branch vs false_branch edges
	if ctx.ActionResults["check_status"]["branch_taken"] != "true_branch" {
		t.Error("expected branch_taken in ActionResults")
	}
}

func TestClone(t *testing.T) {
	ctx := NewMergedContext(map[string]any{"trigger": "data"})
	ctx.MergeResult("action_a", map[string]any{"value": "original"})

	clone := ctx.Clone()

	// Modify clone
	clone.MergeResult("action_b", map[string]any{"value": "cloned"})
	clone.ActionResults["action_a"]["value"] = "modified"

	// Original should be unaffected
	if ctx.ActionResults["action_a"]["value"] != "original" {
		t.Error("clone modification affected original ActionResults")
	}
	if _, exists := ctx.ActionResults["action_b"]; exists {
		t.Error("clone's new action leaked to original")
	}
	if _, exists := ctx.Flattened["action_b.value"]; exists {
		t.Error("clone's flattened data leaked to original")
	}
}

func TestSanitizeResultTruncatesLargeString(t *testing.T) {
	largeString := strings.Repeat("x", MaxResultValueSize+1000)
	result := map[string]any{
		"large_field": largeString,
		"small_field": "hello",
	}

	sanitized, wasTruncated := sanitizeResult(result)

	if !wasTruncated {
		t.Error("expected wasTruncated to be true")
	}

	// Large string should be truncated
	truncated, ok := sanitized["large_field"].(string)
	if !ok {
		t.Fatal("expected truncated string")
	}
	if len(truncated) > MaxResultValueSize+50 { // +50 for "...[TRUNCATED]" suffix
		t.Errorf("string not truncated: len=%d", len(truncated))
	}
	if !strings.HasSuffix(truncated, "...[TRUNCATED]") {
		t.Error("expected TRUNCATED suffix")
	}

	// Truncation flag should be set
	if sanitized["large_field_truncated"] != true {
		t.Error("expected truncation flag")
	}

	// Small field should be unchanged
	if sanitized["small_field"] != "hello" {
		t.Error("small field should be unchanged")
	}
}

func TestSanitizeResultTruncatesLargeBytes(t *testing.T) {
	largeBytes := make([]byte, MaxResultValueSize+1000)
	result := map[string]any{
		"binary_data": largeBytes,
	}

	sanitized, _ := sanitizeResult(result)

	if sanitized["binary_data"] != "[BINARY_DATA_TRUNCATED]" {
		t.Error("expected binary data replacement")
	}
	if sanitized["binary_data_truncated"] != true {
		t.Error("expected truncation flag for binary")
	}
}

func TestSanitizeResultTruncatesLargeObject(t *testing.T) {
	// Create an object that serializes to > MaxResultValueSize
	largeSlice := make([]string, 0, 1000)
	for range 1000 {
		largeSlice = append(largeSlice, strings.Repeat("v", 100))
	}
	result := map[string]any{
		"large_object": largeSlice,
	}

	sanitized, _ := sanitizeResult(result)

	if sanitized["large_object"] != "[LARGE_OBJECT_TRUNCATED]" {
		t.Error("expected large object replacement")
	}
}

func TestSanitizeResultPreservesSmallValues(t *testing.T) {
	result := map[string]any{
		"string_val": "hello",
		"int_val":    42,
		"bool_val":   true,
		"nil_val":    nil,
		"map_val":    map[string]any{"nested": "value"},
	}

	sanitized, wasTruncated := sanitizeResult(result)

	if wasTruncated {
		t.Error("expected wasTruncated to be false for small values")
	}

	if sanitized["string_val"] != "hello" {
		t.Error("string_val changed")
	}
	if sanitized["int_val"] != 42 {
		t.Error("int_val changed")
	}
	if sanitized["bool_val"] != true {
		t.Error("bool_val changed")
	}
	if sanitized["nil_val"] != nil {
		t.Error("nil_val changed")
	}
}

func TestSanitizeResultNilInput(t *testing.T) {
	sanitized, wasTruncated := sanitizeResult(nil)

	if wasTruncated {
		t.Error("expected wasTruncated to be false for nil input")
	}

	if sanitized == nil {
		t.Fatal("expected non-nil map from nil input")
	}
	if len(sanitized) != 0 {
		t.Errorf("expected empty map, got %d entries", len(sanitized))
	}
}

func TestSanitizeResultDoesNotMutateInput(t *testing.T) {
	original := map[string]any{
		"large_field": strings.Repeat("x", MaxResultValueSize+1000),
		"small_field": "hello",
	}

	// Capture original keys count
	originalLen := len(original)

	_, _ = sanitizeResult(original)

	// Input map should not have been mutated (no _truncated keys added)
	if len(original) != originalLen {
		t.Errorf("input map was mutated: had %d keys, now has %d", originalLen, len(original))
	}
}

func TestSanitizeResultNumericFastPath(t *testing.T) {
	result := map[string]any{
		"int_val":     42,
		"int64_val":   int64(9999999999),
		"float32_val": float32(3.14),
		"float64_val": float64(2.718281828),
		"bool_val":    true,
		"uint_val":    uint(100),
	}

	sanitized, _ := sanitizeResult(result)

	if sanitized["int_val"] != 42 {
		t.Error("int_val changed")
	}
	if sanitized["int64_val"] != int64(9999999999) {
		t.Error("int64_val changed")
	}
	if sanitized["float64_val"] != float64(2.718281828) {
		t.Error("float64_val changed")
	}
	if sanitized["bool_val"] != true {
		t.Error("bool_val changed")
	}
}

func TestMergeResultNilResult(t *testing.T) {
	ctx := NewMergedContext(nil)

	// Should not panic
	ctx.MergeResult("action_a", nil)

	// Action should exist in ActionResults with empty map
	if _, exists := ctx.ActionResults["action_a"]; !exists {
		t.Error("expected action_a in ActionResults")
	}
}

func TestMergeResultOverwrite(t *testing.T) {
	ctx := NewMergedContext(nil)

	ctx.MergeResult("action_a", map[string]any{"value": "first"})
	ctx.MergeResult("action_a", map[string]any{"value": "second"})

	// Second call should overwrite
	if ctx.ActionResults["action_a"]["value"] != "second" {
		t.Error("expected overwrite of action_a result")
	}
	if ctx.Flattened["action_a.value"] != "second" {
		t.Error("expected overwrite of flattened key")
	}
}

func TestCloneSharedNestedReferences(t *testing.T) {
	ctx := NewMergedContext(nil)

	nestedMap := map[string]any{"nested_value": "original"}
	ctx.MergeResult("action_a", map[string]any{
		"nested": nestedMap,
	})

	clone := ctx.Clone()

	// Modify the nested map through the clone - values within result
	// maps are shared references (documented 2-level deep copy behavior).
	clonedNested := clone.ActionResults["action_a"]["nested"].(map[string]any)
	clonedNested["nested_value"] = "modified"

	// Original should also see the change (shared reference).
	originalNested := ctx.ActionResults["action_a"]["nested"].(map[string]any)
	if originalNested["nested_value"] != "modified" {
		t.Error("expected nested maps to share references after Clone (documented behavior)")
	}
}

func TestClonePreservesNilMaps(t *testing.T) {
	// A MergedContext with nil maps (e.g., from JSON deserialization)
	// should produce a clone that also has nil maps.
	ctx := &MergedContext{}

	clone := ctx.Clone()

	if clone.TriggerData != nil {
		t.Error("expected nil TriggerData in clone")
	}
	if clone.ActionResults != nil {
		t.Error("expected nil ActionResults in clone")
	}
	if clone.Flattened != nil {
		t.Error("expected nil Flattened in clone")
	}
}

func TestSanitizeResultPreservesNilValues(t *testing.T) {
	result := map[string]any{
		"nil_field":    nil,
		"string_field": "hello",
	}

	sanitized, wasTruncated := sanitizeResult(result)

	if wasTruncated {
		t.Error("expected wasTruncated to be false")
	}
	if sanitized["nil_field"] != nil {
		t.Error("nil value should be preserved as nil")
	}
	if sanitized["string_field"] != "hello" {
		t.Error("string field should be unchanged")
	}
}

// =============================================================================
// Validate Tests
// =============================================================================

func TestWorkflowInputValidate(t *testing.T) {
	validInput := WorkflowInput{
		RuleID:      uuid.New(),
		ExecutionID: uuid.New(),
		Graph: GraphDefinition{
			Actions: []ActionNode{{ID: uuid.New(), Name: "action1"}},
			Edges:   []ActionEdge{{ID: uuid.New(), EdgeType: EdgeTypeStart, TargetActionID: uuid.New()}},
		},
	}

	if err := validInput.Validate(); err != nil {
		t.Errorf("expected valid input to pass: %v", err)
	}

	// Missing RuleID.
	invalid := validInput
	invalid.RuleID = uuid.Nil
	if err := invalid.Validate(); err == nil {
		t.Error("expected error for missing RuleID")
	}

	// Missing ExecutionID.
	invalid = validInput
	invalid.ExecutionID = uuid.Nil
	if err := invalid.Validate(); err == nil {
		t.Error("expected error for missing ExecutionID")
	}

	// Empty actions.
	invalid = validInput
	invalid.Graph.Actions = nil
	if err := invalid.Validate(); err == nil {
		t.Error("expected error for empty actions")
	}

	// No start edge.
	invalid = validInput
	invalid.Graph.Edges = []ActionEdge{{ID: uuid.New(), EdgeType: EdgeTypeSequence, TargetActionID: uuid.New()}}
	if err := invalid.Validate(); err == nil {
		t.Error("expected error for missing start edge")
	}
}

func TestBranchInputValidate(t *testing.T) {
	validInput := BranchInput{
		StartAction:      ActionNode{ID: uuid.New(), Name: "action1"},
		ConvergencePoint: uuid.New(),
		Graph: GraphDefinition{
			Actions: []ActionNode{{ID: uuid.New(), Name: "action1"}},
		},
		InitialContext: NewMergedContext(nil),
	}

	if err := validInput.Validate(); err != nil {
		t.Errorf("expected valid input to pass: %v", err)
	}

	// Missing StartAction ID.
	invalid := validInput
	invalid.StartAction.ID = uuid.Nil
	if err := invalid.Validate(); err == nil {
		t.Error("expected error for missing StartAction ID")
	}

	// ConvergencePoint == uuid.Nil is valid (fire-and-forget branches).
	fireAndForget := validInput
	fireAndForget.ConvergencePoint = uuid.Nil
	if err := fireAndForget.Validate(); err != nil {
		t.Errorf("expected uuid.Nil convergence point to be valid (fire-and-forget): %v", err)
	}

	// Empty graph.
	invalid = validInput
	invalid.Graph.Actions = nil
	if err := invalid.Validate(); err == nil {
		t.Error("expected error for empty graph")
	}

	// Missing InitialContext.
	invalid = validInput
	invalid.InitialContext = nil
	if err := invalid.Validate(); err == nil {
		t.Error("expected error for missing InitialContext")
	}
}

func TestActionActivityInputValidate(t *testing.T) {
	validInput := ActionActivityInput{
		ActionID:   uuid.New(),
		ActionName: "send_email",
		ActionType: "notification",
	}

	if err := validInput.Validate(); err != nil {
		t.Errorf("expected valid input to pass: %v", err)
	}

	// Missing ActionID.
	invalid := validInput
	invalid.ActionID = uuid.Nil
	if err := invalid.Validate(); err == nil {
		t.Error("expected error for missing ActionID")
	}

	// Missing ActionName.
	invalid = validInput
	invalid.ActionName = ""
	if err := invalid.Validate(); err == nil {
		t.Error("expected error for missing ActionName")
	}

	// Missing ActionType.
	invalid = validInput
	invalid.ActionType = ""
	if err := invalid.Validate(); err == nil {
		t.Error("expected error for missing ActionType")
	}
}
