# Phase 3: Core Models & Context

**Category**: backend
**Status**: Pending
**Dependencies**: Phase 1 (Infrastructure Setup - COMPLETED), Phase 2 (Temporalgraph Evaluation - COMPLETED, decision: custom GraphExecutor)

---

## Overview

Implement the foundational data types for the Temporal workflow engine: `WorkflowInput`, `GraphDefinition`, `ActionNode`, `ActionEdge`, `MergedContext`, and supporting input/output types. These models form the contract between the graph executor (Phase 4), workflow orchestration (Phase 5), and activity execution (Phase 6).

This is the first code phase that produces Go source files in the new `business/sdk/workflow/temporal/` package. All models must be compatible with the existing workflow types in `business/sdk/workflow/models.go` (specifically `ActionEdge`, `RuleAction`, `RuleActionView`, `ActionResult`, and `ConditionResult`).

## Goals

1. **Define all core Temporal workflow data types compatible with existing workflow models** - `ActionNode`, `ActionEdge`, `GraphDefinition`, `WorkflowInput`, `BranchInput`/`BranchOutput`, `ActionActivityInput`/`ActionActivityOutput` that map cleanly to/from the existing `business/sdk/workflow` types
2. **Implement MergedContext with result accumulation, cloning for parallel branches, and payload size sanitization** - Thread-safe context that supports `{{action_name.field}}` template variable resolution, deep cloning for parallel branches, and truncation of large values to stay within Temporal's 2MB payload limit
3. **Validate models with unit tests ensuring MergeResult, Clone, and truncation behavior are correct** - Comprehensive tests for context merging, cloning isolation, and sanitization edge cases

## Prerequisites

- Phase 1 complete: Temporal SDK dependency available in `go.mod`/`vendor`
- Phase 2 complete: Custom GraphExecutor confirmed (no temporalgraph dependency)
- Understanding of existing workflow models in `business/sdk/workflow/models.go`

---

## Go Package Structure

```
business/sdk/workflow/temporal/
    models.go          <- THIS PHASE
    models_test.go     <- THIS PHASE
    graph_executor.go  <- Phase 4
    workflow.go        <- Phase 5
    activities.go      <- Phase 6
    activities_async.go <- Phase 6
    async_completer.go  <- Phase 6
    trigger.go         <- Phase 7
    stores/
        edgedb/        <- Phase 8
```

---

## Task Breakdown

### Task 1: Create Temporal Package Directory Structure

**Status**: Pending

**Description**: Create the `business/sdk/workflow/temporal/` package directory. This package sits alongside the existing `business/sdk/workflow/` package and contains all Temporal-specific workflow types and logic.

**Notes**:
- Package name: `temporal` (import path: `github.com/timmaaaz/ichor/business/sdk/workflow/temporal`)
- Lives under `business/sdk/` because it's reusable workflow infrastructure, not domain-specific
- Follows the same pattern as `business/sdk/workflow/` which already exists

**Files**:
- `business/sdk/workflow/temporal/` (directory)

**Implementation Guide**:

The directory is created implicitly when writing the first file (`models.go`). No separate action needed unless you want to create an empty package doc file first.

---

### Task 2: Implement models.go

**Status**: Pending

**Description**: Create all core data types for the Temporal workflow engine. These types define the contract between all subsequent phases (graph executor, workflow, activities, trigger).

**Notes**:
- `WorkflowInput` - Top-level input passed when starting a Temporal workflow
- `GraphDefinition` - Container for actions (nodes) and edges, loaded from database
- `ActionNode` - Single action in the workflow graph (maps from `workflow.RuleAction` + `RuleActionView`)
- `ActionEdge` - Directed edge between actions (maps from `workflow.ActionEdge` with `EdgeOrder` -> `SortOrder`)
- `MergedContext` - Accumulates action results for template variable resolution
- `BranchInput`/`BranchOutput` - For parallel branch child workflow execution
- `ActionActivityInput`/`ActionActivityOutput` - Activity input/output types
- `MaxResultValueSize` constant (50KB) for payload truncation
- `sanitizeResult` for truncating large values before merging into context
- Constants: `TaskQueue`, `HistoryLengthThreshold`, `ContextSizeWarningBytes`

**Files**:
- `business/sdk/workflow/temporal/models.go`

**Implementation Guide**:

```go
package temporal

import (
    "encoding/json"

    "github.com/google/uuid"
)

// =============================================================================
// Constants
// =============================================================================

const (
    // TaskQueue is the Temporal task queue name for workflow workers.
    TaskQueue = "ichor-workflow-queue"

    // HistoryLengthThreshold triggers Continue-As-New to prevent unbounded history.
    HistoryLengthThreshold = 10_000

    // ContextSizeWarningBytes logs a warning when merged context approaches limits.
    ContextSizeWarningBytes = 200 * 1024 // 200KB

    // MaxResultValueSize limits individual result values to prevent payload bloat.
    // Temporal has a 2MB payload limit; this keeps individual values manageable.
    MaxResultValueSize = 50 * 1024 // 50KB per value
)

// =============================================================================
// Graph Definition Types
// =============================================================================

// WorkflowInput is passed when starting a workflow execution via Temporal.
type WorkflowInput struct {
    RuleID      uuid.UUID              `json:"rule_id"`
    ExecutionID uuid.UUID              `json:"execution_id"`
    Graph       GraphDefinition        `json:"graph"`
    TriggerData map[string]interface{} `json:"trigger_data"`
}

// GraphDefinition mirrors the database model (rule_actions + action_edges).
// It is loaded from PostgreSQL and passed as workflow input.
type GraphDefinition struct {
    Actions []ActionNode `json:"actions"`
    Edges   []ActionEdge `json:"edges"`
}

// ActionNode represents a single action in the workflow graph.
// Maps from workflow.RuleAction + RuleActionView (business layer).
//
// Field selection rationale:
//   - ID, Name, ActionType, Config: Core execution fields needed by the graph
//     executor and activity dispatcher.
//   - Description: Valuable for Temporal UI visibility, activity logging, and
//     error messages (debugging "Update inventory levels" vs action ID "a3f2...").
//   - IsActive, DeactivatedBy: Enable runtime enforcement. Long-running workflows
//     (e.g., human approval taking days) may encounter actions deactivated after
//     the workflow started. Activities can check IsActive and skip/fail gracefully
//     with a clear audit trail (DeactivatedBy).
//
// Intentionally omitted:
//   - AutomationRuleID: Already on WorkflowInput.RuleID - all actions in a graph
//     belong to the same rule, so carrying it on each node is redundant.
//   - TemplateID: ActionType is already resolved from the template by the Phase 8
//     adapter. If template-specific behavior is needed later (versioning, default
//     config reload), it can be added without breaking the model.
type ActionNode struct {
    ID            uuid.UUID       `json:"id"`
    Name          string          `json:"name"`
    Description   string          `json:"description"`
    ActionType    string          `json:"action_type"`
    Config        json.RawMessage `json:"action_config"`
    IsActive      bool            `json:"is_active"`
    DeactivatedBy uuid.UUID       `json:"deactivated_by"`
}

// ActionEdge represents a directed edge between actions in the workflow graph.
// Maps from workflow.ActionEdge with EdgeOrder -> SortOrder.
// SourceActionID is nil for start edges (entry points into the graph).
type ActionEdge struct {
    ID             uuid.UUID  `json:"id"`
    SourceActionID *uuid.UUID `json:"source_action_id"` // nil for start edges
    TargetActionID uuid.UUID  `json:"target_action_id"`
    EdgeType       string     `json:"edge_type"` // start, sequence, true_branch, false_branch, always
    SortOrder      int        `json:"sort_order"`
}

// =============================================================================
// Execution Context
// =============================================================================

// MergedContext accumulates results from all executed actions.
// It supports template variable resolution via the Flattened map:
//   - {{action_name}} -> entire result map
//   - {{action_name.field}} -> specific field from an action's result
type MergedContext struct {
    TriggerData   map[string]interface{}            `json:"trigger_data"`
    ActionResults map[string]map[string]interface{} `json:"action_results"` // action_name -> result
    Flattened     map[string]interface{}            `json:"flattened"`      // For template resolution
}

// NewMergedContext creates a context initialized with trigger data.
// Trigger data is copied to the Flattened map for immediate template access.
func NewMergedContext(triggerData map[string]interface{}) *MergedContext {
    ctx := &MergedContext{
        TriggerData:   triggerData,
        ActionResults: make(map[string]map[string]interface{}),
        Flattened:     make(map[string]interface{}),
    }

    for k, v := range triggerData {
        ctx.Flattened[k] = v
    }

    return ctx
}

// MergeResult adds an action's result to the context.
// Large values are sanitized (truncated) to prevent exceeding Temporal's 2MB payload limit.
// Results are indexed both by action name (for ActionResults) and as
// flattened "action_name.field" keys (for template resolution).
func (c *MergedContext) MergeResult(actionName string, result map[string]interface{}) {
    if c.ActionResults == nil {
        c.ActionResults = make(map[string]map[string]interface{})
    }
    if c.Flattened == nil {
        c.Flattened = make(map[string]interface{})
    }

    sanitized := sanitizeResult(result)

    c.ActionResults[actionName] = sanitized

    for k, v := range sanitized {
        c.Flattened[actionName+"."+k] = v
    }

    c.Flattened[actionName] = sanitized
}

// Clone creates a 2-level deep copy for parallel branch execution.
// TriggerData, ActionResults, and Flattened maps are cloned, and each
// action's result map is copied. However, values WITHIN result maps
// (e.g., a nested map[string]interface{}) are shared references.
// This is acceptable because action results are treated as immutable
// after MergeResult - activities produce results, MergeResult stores
// them, and nothing mutates individual result values afterward.
func (c *MergedContext) Clone() *MergedContext {
    clone := &MergedContext{
        TriggerData:   make(map[string]interface{}),
        ActionResults: make(map[string]map[string]interface{}),
        Flattened:     make(map[string]interface{}),
    }

    for k, v := range c.TriggerData {
        clone.TriggerData[k] = v
    }
    for k, v := range c.ActionResults {
        resultCopy := make(map[string]interface{})
        for rk, rv := range v {
            resultCopy[rk] = rv
        }
        clone.ActionResults[k] = resultCopy
    }
    for k, v := range c.Flattened {
        clone.Flattened[k] = v
    }

    return clone
}

// sanitizeResult truncates large values to prevent payload size issues.
// Returns a new map - the input map is never mutated.
// String values > MaxResultValueSize are truncated with a "[TRUNCATED]" marker.
// Binary data > MaxResultValueSize is replaced entirely.
// Complex objects > MaxResultValueSize (when serialized) are replaced.
// Primitive types (bool, int, float) bypass JSON marshaling (known small).
func sanitizeResult(result map[string]interface{}) map[string]interface{} {
    if result == nil {
        return make(map[string]interface{})
    }

    sanitized := make(map[string]interface{}, len(result))

    for k, v := range result {
        switch val := v.(type) {
        case string:
            if len(val) > MaxResultValueSize {
                sanitized[k] = val[:MaxResultValueSize] + "...[TRUNCATED]"
                sanitized[k+"_truncated"] = true
            } else {
                sanitized[k] = val
            }
        case []byte:
            if len(val) > MaxResultValueSize {
                sanitized[k] = "[BINARY_DATA_TRUNCATED]"
                sanitized[k+"_truncated"] = true
            } else {
                sanitized[k] = val
            }
        case bool, int, int8, int16, int32, int64,
            uint, uint8, uint16, uint32, uint64,
            float32, float64:
            sanitized[k] = val // Known small types, skip marshaling
        default:
            data, err := json.Marshal(val)
            if err != nil {
                sanitized[k] = "[MARSHAL_ERROR]"
                sanitized[k+"_truncated"] = true
            } else if len(data) > MaxResultValueSize {
                sanitized[k] = "[LARGE_OBJECT_TRUNCATED]"
                sanitized[k+"_truncated"] = true
            } else {
                sanitized[k] = val
            }
        }
    }

    return sanitized
}

// =============================================================================
// Parallel Execution Types
// =============================================================================

// BranchInput is passed to child workflows for parallel branch execution.
// Each parallel branch runs as a separate Temporal child workflow.
type BranchInput struct {
    StartAction      ActionNode      `json:"start_action"`
    ConvergencePoint uuid.UUID       `json:"convergence_point"`
    Graph            GraphDefinition `json:"graph"`
    InitialContext   *MergedContext  `json:"initial_context"`
}

// BranchOutput is returned from child workflows.
// Contains all action results accumulated during the branch execution.
type BranchOutput struct {
    ActionResults map[string]map[string]interface{} `json:"action_results"`
}

// =============================================================================
// Activity Types
// =============================================================================

// ActionActivityInput is passed to the action execution activity.
// Config contains the action's JSON configuration (possibly with template variables).
// Context provides the merged execution context for template variable resolution.
type ActionActivityInput struct {
    ActionID   uuid.UUID              `json:"action_id"`
    ActionName string                 `json:"action_name"`
    ActionType string                 `json:"action_type"`
    Config     json.RawMessage        `json:"config"`
    Context    map[string]interface{} `json:"context"`
}

// ActionActivityOutput is returned from the action execution activity.
// Result contains the action handler's output (e.g., BranchTaken for conditions).
// Success indicates whether the action completed without error.
type ActionActivityOutput struct {
    ActionID   uuid.UUID              `json:"action_id"`
    ActionName string                 `json:"action_name"`
    Result     map[string]interface{} `json:"result"`
    Success    bool                   `json:"success"`
}
```

**Template Variable Resolution Strategy**:

The `MergedContext.Flattened` map supports two access patterns:
- `Flattened["action_name"]` -> entire result map for that action
- `Flattened["action_name.field"]` -> specific field value from an action's result

This is preparation for the Phase 6 activity implementation, where template variables in action configs (e.g., `{{send_email.status}}`) will be resolved using this flattened structure. The existing `resolveTemplateVars` function in `workflowactions/communication/alert.go` uses a regex-based approach (`templateVarPattern`) against a flat `map[string]interface{}`. The Flattened map is designed to be passed directly to that same pattern - no new resolver needed, just a richer context map.

**Phase 3 Scope**: Store results in flattened format (this phase).
**Phase 6 Scope**: Pass `Flattened` map as the template resolution context when executing activities.

**Compatibility Mapping with Existing Types**:

| Temporal Type | Existing Type | Mapping Notes |
|---------------|---------------|---------------|
| `ActionNode` | `workflow.RuleAction` + `RuleActionView` | `ActionType` from `TemplateActionType` (preferred) or `ActionConfig["action_type"]` (fallback). `Description`, `IsActive`, `DeactivatedBy` mapped directly. `AutomationRuleID` and `TemplateID` omitted (see field rationale in type comment). |
| `ActionEdge` | `workflow.ActionEdge` | `EdgeOrder` -> `SortOrder`, `RuleID` omitted (in `WorkflowInput`) |
| `ActionActivityOutput.Result` | `workflow.ActionResult.ResultData` | Direct map; `BranchTaken` stored as `result["branch_taken"]` |
| `MergedContext.TriggerData` | `workflow.TriggerEvent.RawData` | Trigger event data flattened for template access |

**ActionType Extraction (Phase 8 Adapter Responsibility)**:

The `ActionNode.ActionType` field is resolved during graph loading in the Phase 8 Edge Store adapter, NOT in Phase 3 models. The extraction priority is:

1. **`RuleActionView.TemplateActionType`** (preferred) - Set when the action uses a template. This is the primary source for most actions.
2. **`ActionConfig["action_type"]`** (fallback) - For actions without a template, or where the template doesn't define an action type, the adapter falls back to extracting it from the JSON config.
3. **Error** - If neither source provides an action type, the adapter returns an error and the action is not included in the graph.

Phase 3 defines the `ActionType` field on `ActionNode` as a plain string. Phase 8 is responsible for populating it correctly during the `workflow.RuleAction` -> `temporal.ActionNode` conversion.

---

### Task 3: Write Unit Tests for MergedContext

**Status**: Pending

**Description**: Write comprehensive unit tests for `MergedContext` covering `MergeResult`, `Clone`, and `sanitizeResult` behavior. These tests ensure the context accumulation and payload protection work correctly before building the graph executor and workflow on top.

**Notes**:
- Test `MergeResult` - Verify results are stored by action name and flattened correctly
- Test `Clone` - Verify deep copy isolation (modifying clone doesn't affect original)
- Test `sanitizeResult` truncation - Verify large strings, byte arrays, and complex objects are truncated
- Test `NewMergedContext` - Verify trigger data is copied to Flattened map
- Test multiple `MergeResult` calls - Verify sequential accumulation works correctly
- Test `BranchTaken` compatibility - Verify condition results with `branch_taken` field are stored correctly
- Test `sanitizeResult` nil input - Returns empty map, not panic
- Test `sanitizeResult` does not mutate input map
- Test `sanitizeResult` numeric fast-path - Primitives bypass JSON marshaling
- Test `MergeResult` nil result - Should not panic, stores empty map
- Test `MergeResult` overwrite - Second call for same action name overwrites first

**Files**:
- `business/sdk/workflow/temporal/models_test.go`

**Implementation Guide**:

```go
package temporal

import (
    "strings"
    "testing"
)

func TestNewMergedContext(t *testing.T) {
    triggerData := map[string]interface{}{
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
    ctx := NewMergedContext(map[string]interface{}{"trigger": "data"})

    result := map[string]interface{}{
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
    fullResult, ok := ctx.Flattened["send_email"].(map[string]interface{})
    if !ok {
        t.Fatal("expected send_email to be map in Flattened")
    }
    if fullResult["status"] != "sent" {
        t.Error("expected full result accessible by action name")
    }
}

func TestMergeResultMultipleActions(t *testing.T) {
    ctx := NewMergedContext(nil)

    ctx.MergeResult("action_a", map[string]interface{}{"value": "first"})
    ctx.MergeResult("action_b", map[string]interface{}{"value": "second"})

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

    conditionResult := map[string]interface{}{
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
    ctx := NewMergedContext(map[string]interface{}{"trigger": "data"})
    ctx.MergeResult("action_a", map[string]interface{}{"value": "original"})

    clone := ctx.Clone()

    // Modify clone
    clone.MergeResult("action_b", map[string]interface{}{"value": "cloned"})
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
    result := map[string]interface{}{
        "large_field": largeString,
        "small_field": "hello",
    }

    sanitized := sanitizeResult(result)

    // Large string should be truncated
    truncated, ok := sanitized["large_field"].(string)
    if !ok {
        t.Fatal("expected truncated string")
    }
    if len(truncated) > MaxResultValueSize+50 { // +50 for "[TRUNCATED]" suffix
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
    result := map[string]interface{}{
        "binary_data": largeBytes,
    }

    sanitized := sanitizeResult(result)

    if sanitized["binary_data"] != "[BINARY_DATA_TRUNCATED]" {
        t.Error("expected binary data replacement")
    }
    if sanitized["binary_data_truncated"] != true {
        t.Error("expected truncation flag for binary")
    }
}

func TestSanitizeResultTruncatesLargeObject(t *testing.T) {
    // Create an object that serializes to > MaxResultValueSize
    largeMap := make(map[string]string)
    for i := 0; i < 10000; i++ {
        largeMap[strings.Repeat("k", 10)] = strings.Repeat("v", 100)
    }
    result := map[string]interface{}{
        "large_object": largeMap,
    }

    sanitized := sanitizeResult(result)

    if sanitized["large_object"] != "[LARGE_OBJECT_TRUNCATED]" {
        t.Error("expected large object replacement")
    }
}

func TestSanitizeResultPreservesSmallValues(t *testing.T) {
    result := map[string]interface{}{
        "string_val": "hello",
        "int_val":    42,
        "bool_val":   true,
        "nil_val":    nil,
        "map_val":    map[string]interface{}{"nested": "value"},
    }

    sanitized := sanitizeResult(result)

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
    sanitized := sanitizeResult(nil)

    if sanitized == nil {
        t.Fatal("expected non-nil map from nil input")
    }
    if len(sanitized) != 0 {
        t.Errorf("expected empty map, got %d entries", len(sanitized))
    }
}

func TestSanitizeResultDoesNotMutateInput(t *testing.T) {
    original := map[string]interface{}{
        "large_field": strings.Repeat("x", MaxResultValueSize+1000),
        "small_field": "hello",
    }

    // Capture original keys count
    originalLen := len(original)

    _ = sanitizeResult(original)

    // Input map should not have been mutated (no _truncated keys added)
    if len(original) != originalLen {
        t.Errorf("input map was mutated: had %d keys, now has %d", originalLen, len(original))
    }
}

func TestSanitizeResultNumericFastPath(t *testing.T) {
    result := map[string]interface{}{
        "int_val":     42,
        "int64_val":   int64(9999999999),
        "float32_val": float32(3.14),
        "float64_val": float64(2.718281828),
        "bool_val":    true,
        "uint_val":    uint(100),
    }

    sanitized := sanitizeResult(result)

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

    ctx.MergeResult("action_a", map[string]interface{}{"value": "first"})
    ctx.MergeResult("action_a", map[string]interface{}{"value": "second"})

    // Second call should overwrite
    if ctx.ActionResults["action_a"]["value"] != "second" {
        t.Error("expected overwrite of action_a result")
    }
    if ctx.Flattened["action_a.value"] != "second" {
        t.Error("expected overwrite of flattened key")
    }
}
```

---

## Validation Criteria

- [ ] `go build ./business/sdk/workflow/temporal/...` passes with no errors
- [ ] `go test ./business/sdk/workflow/temporal/...` passes - all unit tests green
- [ ] Large string values (>50KB) are truncated with `...[TRUNCATED]` suffix
- [ ] Large byte arrays are replaced with `[BINARY_DATA_TRUNCATED]`
- [ ] Large serialized objects are replaced with `[LARGE_OBJECT_TRUNCATED]`
- [ ] `_truncated` flag is set for all truncated values
- [ ] `Clone()` produces isolated copies (modifying clone doesn't affect original)
- [ ] `MergeResult()` populates both `ActionResults` and `Flattened` maps correctly
- [ ] Template variable patterns work: `Flattened["action_name.field"]` returns correct value
- [ ] `BranchTaken` from condition results is accessible in `ActionResults` for graph traversal
- [ ] `sanitizeResult(nil)` returns empty map, not panic
- [ ] `sanitizeResult` does NOT mutate input map (no `_truncated` keys added to original)
- [ ] Numeric/bool types bypass JSON marshaling in `sanitizeResult` (fast-path)
- [ ] `Clone()` produces isolated copies at 2-level depth (modifying clone's action result values doesn't affect original)
- [ ] `MergeResult` with nil result does not panic
- [ ] `MergeResult` with duplicate action name overwrites previous result
- [ ] No import cycles with existing `business/sdk/workflow` package
- [ ] `go vet ./business/sdk/workflow/temporal/...` passes

---

## Deliverables

- `business/sdk/workflow/temporal/models.go` - All core data types and constants
- `business/sdk/workflow/temporal/models_test.go` - Unit tests for MergedContext

---

## Gotchas & Tips

### Common Pitfalls

- **Import cycle with `business/sdk/workflow`**: The temporal package should NOT import `business/sdk/workflow` directly. The models here are intentionally separate types that get converted to/from via adapter functions in Phase 8 (Edge Store). If you add an import to `business/sdk/workflow`, you'll create a cycle when that package eventually imports temporal.
- **`EdgeOrder` vs `SortOrder` naming**: The existing `workflow.ActionEdge` uses `EdgeOrder`. The temporal `ActionEdge` uses `SortOrder`. The adapter (Phase 8) handles this mapping. Don't rename either - they serve different serialization contexts.
- **`json.RawMessage` is `[]byte`**: When comparing `json.RawMessage` values in tests, use `string()` conversion or `bytes.Equal()`, not `==`.
- **Nil maps in MergedContext**: `MergeResult` defensively initializes `ActionResults` and `Flattened` if nil. This is important because JSON deserialization of a `MergedContext` (e.g., after Continue-As-New) may leave these as nil if the source JSON had `null`.
- **Template variable format**: The flattened key format is `"action_name.field"` (dot-separated). This must match the template resolution system in the existing workflow engine. Don't change this format.
- **Parallel edge type**: The implementation plan mentions a `parallel` edge type, but the existing codebase only validates `start`, `sequence`, `true_branch`, `false_branch`, `always`. For Phase 3, use only the 5 existing edge types. Parallel execution is determined by multiple outgoing edges from the same source, not by a special edge type.

### Tips

- Start with `models.go` - get it compiling first, then write tests
- The implementation plan document (`.claude/plans/workflow-temporal-implementation.md` lines 111-298) has the complete reference implementation for all types
- Run `go vet` after writing models to catch common issues (unused imports, shadowed variables)
- The `sanitizeResult` function handles `nil` map input explicitly (returns empty map). `nil` values within the map fall through to the `default` case - `json.Marshal(nil)` returns `"null"` (4 bytes), well under the limit
- Keep the package clean: no database imports, no Temporal SDK imports in models.go. This file is pure Go types that can be tested without any external dependencies.

---

## Testing Strategy

### Unit Tests

Tests in `models_test.go` cover:

1. **NewMergedContext** - Trigger data initialization and Flattened population
2. **MergeResult** - Single action, multiple actions, flattened key format, full result access
3. **MergeResult with BranchTaken** - Condition action results stored correctly for graph traversal
4. **MergeResult edge cases** - Nil result map (no panic, stores empty), overwrite existing action name
5. **Clone isolation** - 2-level copy prevents cross-branch data leakage
6. **sanitizeResult** - String truncation, byte array replacement, large object replacement, small value preservation
7. **sanitizeResult edge cases** - Nil input returns empty map, input map not mutated, numeric fast-path, marshal errors

### What's NOT Tested Here

- Graph traversal logic (Phase 4)
- Temporal workflow determinism (Phase 5)
- Activity execution (Phase 6)
- Database loading/conversion (Phase 8)

These are intentionally deferred to their respective phases. Phase 3 tests validate only the data types and context accumulation logic.

---

## Temporal-Specific Patterns

### Payload Size Management

Temporal enforces a 2MB payload limit per workflow input/output. The `MergedContext` grows with each action's result, so `sanitizeResult` ensures no individual value exceeds 50KB. The `ContextSizeWarningBytes` constant (200KB) is used in Phase 5 to log warnings when the context is approaching dangerous sizes.

### Continue-As-New Compatibility

`MergedContext` is JSON-serializable because it's passed as workflow input during Continue-As-New. All fields use JSON tags and basic Go types (`map[string]interface{}`). No function references, channels, or other non-serializable types.

### Determinism Requirements

The models themselves don't need to be deterministic (that's Phase 4's concern). However, `SortOrder` on `ActionEdge` is critical - it ensures the graph executor processes edges in a deterministic order during replay.

---

## Existing Code Reference

### Edge Types (`business/sdk/workflow/models.go:387-393`)
```go
const (
    EdgeTypeStart       = "start"
    EdgeTypeSequence    = "sequence"
    EdgeTypeTrueBranch  = "true_branch"
    EdgeTypeFalseBranch = "false_branch"
    EdgeTypeAlways      = "always"
)
```

### ActionEdge (`business/sdk/workflow/models.go:397-405`)
```go
type ActionEdge struct {
    ID             uuid.UUID
    RuleID         uuid.UUID
    SourceActionID *uuid.UUID
    TargetActionID uuid.UUID
    EdgeType       string
    EdgeOrder      int
    CreatedDate    time.Time
}
```

### RuleActionView (`business/sdk/workflow/models.go:598-611`)
```go
type RuleActionView struct {
    ID                    uuid.UUID
    AutomationRulesID     *uuid.UUID
    Name                  string
    Description           string
    ActionConfig          json.RawMessage
    IsActive              bool
    TemplateID            *uuid.UUID
    TemplateName          string
    TemplateActionType    string          // <- Primary source for ActionNode.ActionType
    TemplateDefaultConfig json.RawMessage
    DeactivatedBy         uuid.UUID
}
```

### ActionResult (`business/sdk/workflow/models.go:112-123`)
```go
type ActionResult struct {
    ActionID     uuid.UUID
    ActionName   string
    ActionType   string
    Status       string
    ResultData   map[string]interface{}  // -> ActionActivityOutput.Result
    ErrorMessage string
    Duration     time.Duration
    StartedAt    time.Time
    CompletedAt  *time.Time
    BranchTaken  string  // "true_branch" or "false_branch" for conditions
}
```

---

## Commands Reference

```bash
# Start this phase
/workflow-temporal-next

# Validate this phase
/workflow-temporal-validate 3

# Review plan before implementing
/workflow-temporal-plan-review 3

# Review code after implementing
/workflow-temporal-review 3
```
