# Phase 12: Edge Case & Limit Tests

**Category**: Testing
**Status**: Pending
**Dependencies**: Phase 1 (test container), Phase 3 (models/sanitization), Phase 5 (workflow/Continue-As-New), Phase 6 (activities/async), Phase 11 (test helpers/fixtures)

---

## Overview

Test Continue-As-New context preservation, payload limits, error handling, and async completion. Phase 11 tested happy-path workflow execution and replay determinism. Phase 12 focuses on boundary conditions and failure modes: verifying ContinuationState preserves full MergedContext across Continue-As-New boundaries, testing payload sanitization at the 50KB per-value boundary, testing the full async activity lifecycle (ErrResultPending, task tokens, AsyncCompleter), and validating error propagation through the graph (activity failures, branch failures, retry policies).

### Code Under Test (Verified from Source)

| Area | Key Code | Threshold/Constant |
|------|----------|-------------------|
| Continue-As-New | `checkContinueAsNew()` in workflow.go | `HistoryLengthThreshold = 10,000` events (uses `>` not `>=`) |
| Payload truncation | `sanitizeResult()` in models.go | `MaxResultValueSize = 50KB` per value |
| Async activity | `ExecuteAsyncActionActivity()` → `activity.ErrResultPending` | Task token + external callback |
| Async completion | `AsyncCompleter.Complete/Fail()` | `client.CompleteActivity()` |
| Retry policy (regular) | `activityOptions()` | MaxAttempts=3, InitialInterval=1s, Backoff=2.0x |
| Retry policy (async/human) | `activityOptions()` | MaxAttempts=1 (no retries) |
| Regular timeout | `activityOptions()` | StartToClose=5min |
| Async timeout | `activityOptions()` | StartToClose=30min, Heartbeat=1min |
| Human timeout | `activityOptions()` | StartToClose=7 days, Heartbeat=1hr |

**Note**: `ContextSizeWarningBytes = 200KB` exists as a constant but is NOT currently enforced — `MergeResult()` does not log warnings at this threshold. The constant is reserved for future use.

## Goals

1. **Test Continue-As-New context preservation** — verify ContinuationState preserves full MergedContext (ActionResults + TriggerData + Flattened) across JSON round-trips, and extract `shouldContinueAsNew()` for threshold unit testing
2. **Verify payload sanitization boundary conditions** (50KB per value, exact boundary vs just over) and the **async activity lifecycle** (ErrResultPending, task tokens, AsyncCompleter Complete/Fail paths)
3. **Test error propagation** (activity failures fail workflow, branch failures fail parallel group, fire-and-forget errors isolated) and **retry policy enforcement** (3 attempts regular, 1 attempt async/human)

## Prerequisites

- Phase 1 complete — test container infrastructure for optional integration tests
- Phase 3 complete — `MergedContext`, `sanitizeResult`, `MaxResultValueSize`
- Phase 5 complete — `ExecuteGraphWorkflow`, `checkContinueAsNew`, `activityOptions`
- Phase 6 complete — `Activities` struct, `ExecuteAsyncActionActivity`, `AsyncCompleter`
- Phase 11 complete — test helpers (`testActionHandler`, `setupTestEnv`, `newTestRegistry`, `wfLinearGraph`, `wfConditionDiamond`, `wfParallelGraph`, `wfFireForgetGraph`)
- Docker available for optional integration tests (real Temporal container)

### Implementation Changes Required Before Tests

**1. Extract `shouldContinueAsNew` from `checkContinueAsNew`** (workflow.go):
```go
// shouldContinueAsNew is the testable core of the threshold check.
// Extracted so unit tests can verify boundary behavior without needing
// workflow.Context or mocked GetInfo().
func shouldContinueAsNew(currentHistoryLength int) bool {
    return currentHistoryLength > HistoryLengthThreshold
}

func checkContinueAsNew(ctx workflow.Context, input WorkflowInput, mergedCtx *MergedContext) error {
    info := workflow.GetInfo(ctx)
    if shouldContinueAsNew(int(info.GetCurrentHistoryLength())) {
        // ... existing logging and ContinueAsNewError logic ...
    }
    return nil
}
```

**2. Extract `ActivityCompleter` interface from `AsyncCompleter`** (async_completer.go):
```go
// ActivityCompleter is a narrow interface for completing async activities.
// Extracted from client.Client (~50 methods) for testability.
type ActivityCompleter interface {
    CompleteActivity(ctx context.Context, taskToken []byte, result any, err error) error
}

type AsyncCompleter struct {
    client ActivityCompleter  // Was client.Client
}

func NewAsyncCompleter(c ActivityCompleter) *AsyncCompleter {
    return &AsyncCompleter{client: c}
}
```

This change is backwards-compatible: `client.Client` implements `ActivityCompleter`.

### Key Signatures Reference (from Phase 11)

```go
// Phase 11 test helpers (same package, directly accessible):
type testActionHandler struct {
    actionType string
    result     any
    err        error
    called     int
}
func newTestRegistry(handlers ...*testActionHandler) *workflow.ActionRegistry
func setupTestEnv(t *testing.T, registry *workflow.ActionRegistry) *testsuite.TestWorkflowEnvironment
func wfLinearGraph(actionTypes ...string) (GraphDefinition, []uuid.UUID)
func wfConditionDiamond(trueArmType, falseArmType string) (GraphDefinition, wfDiamondIDs)
func wfParallelGraph(forkType, branchAType, branchBType, mergeType string) (GraphDefinition, wfParallelIDs)
func wfFireForgetGraph(forkType, branchAType, branchBType string) (GraphDefinition, wfFireForgetIDs)

// AsyncRegistry (activities_async.go):
func NewAsyncRegistry() *AsyncRegistry
func (r *AsyncRegistry) Register(actionType string, handler AsyncActivityHandler)  // TWO args
func (r *AsyncRegistry) Get(actionType string) (AsyncActivityHandler, bool)
```

---

## Test File Structure

```
business/sdk/workflow/temporal/
├── workflow_continueasnew_test.go    # Continue-As-New threshold + context preservation
├── workflow_payload_test.go          # Payload truncation boundary conditions + MergeResult integration
├── activities_async_test.go          # Async activity lifecycle and AsyncCompleter
├── workflow_errors_test.go           # Error handling and retry policies
└── ... (existing files unchanged, except workflow.go + async_completer.go minor refactors)
```

### Test Coverage Goals

| Area | Tests | Approach | Phase 3 Overlap? |
|------|-------|----------|------------------|
| shouldContinueAsNew threshold | 3 | Pure unit test on extracted function | No |
| ContinuationState preservation | 4 | Pure unit test (JSON round-trip) | No |
| Payload boundary (string) | 3 | Pure unit test on `sanitizeResult` | No (exact boundary) |
| Payload boundary (binary) | 2 | Pure unit test on `sanitizeResult` | No (exact boundary) |
| Payload boundary (object) | 2 | Pure unit test on `sanitizeResult` | No (exact boundary) |
| MergeResult + truncation | 2 | Pure unit test on `MergedContext.MergeResult` | No |
| MergedContext size stress | 1 | Pure unit test | No |
| Async activity start | 3 | Direct function call (bypass SDK) | No |
| AsyncCompleter Complete/Fail | 5 | Mock `ActivityCompleter` interface | No |
| Activity error → workflow failure | 3 | SDK test suite | Minimal (Phase 11 has 1) |
| Branch failure → parallel failure | 2 | SDK test suite | No |
| Fire-and-forget error isolation | 1 | SDK test suite | No |
| Retry policy enforcement | 4 | SDK test suite (handler state counting) | No |
| Handler not found | 1 | SDK test suite | Phase 11 has 1 |

**Total: ~36 tests** (all unique, no Phase 3 duplication)

---

## Task Breakdown

### Task 1: Continue-As-New Tests

**Status**: Pending

**Description**: Two concerns: (1) **Threshold logic** — verify `shouldContinueAsNew()` triggers at exactly `> HistoryLengthThreshold` (extracted pure function). (2) **Context preservation** — verify `ContinuationState` preserves full `MergedContext` structure through JSON round-trip, including edge cases (large maps, float64 precision, special characters, nil/empty states).

**Files**:
- `business/sdk/workflow/temporal/workflow.go` (minor refactor: extract `shouldContinueAsNew`)
- `business/sdk/workflow/temporal/workflow_continueasnew_test.go`

**Implementation Guide**:

```go
package temporal

import (
    "encoding/json"
    "fmt"
    "strings"
    "testing"

    "github.com/google/uuid"
    "github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// shouldContinueAsNew Threshold Tests (extracted pure function)
// ---------------------------------------------------------------------------

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

// ---------------------------------------------------------------------------
// ContinuationState JSON Round-Trip Preservation
// ---------------------------------------------------------------------------

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
            "key":             "value",
            "action1.result":  "data",
            "action1.count":   float64(42),
            "action2.nested":  map[string]any{"deep": "value"},
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
            "small_int":  float64(42),
            "large_int":  float64(9007199254740992), // 2^53
            "float_val":  float64(3.14159265358979),
            "negative":   float64(-1000),
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
    // ContinuationState with empty ActionResults should not cause nil pointer issues.
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
    // ActionResults may deserialize as nil (empty JSON object → empty map) or empty map
    // Either is fine as long as MergeResult handles both (it does — nil check in MergeResult).
}

func TestContinuationState_NilState_CreatesFreshContext(t *testing.T) {
    // First execution: ContinuationState is nil.
    // ExecuteGraphWorkflow should create fresh MergedContext from TriggerData.
    // This is the code path in workflow.go:
    //   if input.ContinuationState != nil { mergedCtx = input.ContinuationState }
    //   else { mergedCtx = NewMergedContext(input.TriggerData) }

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
            "unicode_名前":      {"result": "成功"},
        },
        Flattened: map[string]any{
            "key.with.dots":            "value",
            "action_with.dots.status":  "done",
            "unicode_名前.result":       "成功",
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
```

**Notes**:
- `shouldContinueAsNew` is extracted from `checkContinueAsNew` — pure function, no workflow.Context needed
- JSON round-trip tests cover the real serialization path Temporal uses
- `float64` is the expected type for all JSON numbers — use `float64` literals in expected values
- Phase 3 already tests `NewMergedContext` basics; these tests focus on JSON fidelity across continuation

---

### Task 2: Payload Size Limit Tests

**Status**: Pending

**Description**: Test `sanitizeResult` truncation at the exact 50KB boundary for strings, binary data, and complex objects. Focus on **boundary conditions** that Phase 3 does NOT cover (Phase 3 tests values 1000+ over limit; Phase 12 tests exact boundary). Also test `MergeResult` integration with large values and MergedContext under size stress.

**Phase 3 Already Covers** (DO NOT duplicate):
- `TestSanitizeResultTruncatesLargeString` — string 1000 over limit
- `TestSanitizeResultTruncatesLargeBytes` — bytes 1000 over limit
- `TestSanitizeResultTruncatesLargeObject` — large []string slice
- `TestSanitizeResultPreservesSmallValues` — string, int, bool, nil, map
- `TestSanitizeResultNilInput` — nil map → empty map
- `TestSanitizeResultDoesNotMutateInput` — immutability
- `TestSanitizeResultNumericFastPath` — int, int64, float32, float64, bool, uint
- `TestSanitizeResultPreservesNilValues` — nil field in map

**Files**:
- `business/sdk/workflow/temporal/workflow_payload_test.go`

**Implementation Guide**:

```go
package temporal

import (
    "fmt"
    "strings"
    "testing"

    "github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// Boundary Tests: String (exactly at limit vs just over)
// ---------------------------------------------------------------------------

func TestPayload_StringExactlyAtLimit(t *testing.T) {
    // MaxResultValueSize bytes: should NOT truncate (uses > not >=).
    result := map[string]any{
        "data": strings.Repeat("x", MaxResultValueSize),
    }
    sanitized, truncated := sanitizeResult(result)
    require.False(t, truncated, "exactly at limit should NOT truncate")
    require.Equal(t, result["data"], sanitized["data"])
    // No _truncated flag should exist
    _, hasTruncatedFlag := sanitized["data_truncated"]
    require.False(t, hasTruncatedFlag)
}

func TestPayload_StringOneOverLimit(t *testing.T) {
    // MaxResultValueSize+1 bytes: should truncate.
    result := map[string]any{
        "data": strings.Repeat("x", MaxResultValueSize+1),
    }
    sanitized, truncated := sanitizeResult(result)
    require.True(t, truncated)

    truncatedStr := sanitized["data"].(string)
    require.True(t, strings.HasSuffix(truncatedStr, "...[TRUNCATED]"))
    // Truncated prefix should be exactly MaxResultValueSize bytes
    require.Equal(t, MaxResultValueSize+len("...[TRUNCATED]"), len(truncatedStr))

    require.True(t, sanitized["data_truncated"].(bool))
}

func TestPayload_StringWellOverLimit(t *testing.T) {
    // 10x over limit: verify truncated string size is bounded.
    result := map[string]any{
        "data": strings.Repeat("x", MaxResultValueSize*10),
    }
    sanitized, truncated := sanitizeResult(result)
    require.True(t, truncated)

    truncatedStr := sanitized["data"].(string)
    maxExpected := MaxResultValueSize + len("...[TRUNCATED]")
    require.LessOrEqual(t, len(truncatedStr), maxExpected,
        "truncated string should be bounded to MaxResultValueSize + marker")
}

// ---------------------------------------------------------------------------
// Boundary Tests: Binary Data
// ---------------------------------------------------------------------------

func TestPayload_BinaryExactlyAtLimit(t *testing.T) {
    result := map[string]any{
        "data": make([]byte, MaxResultValueSize),
    }
    sanitized, truncated := sanitizeResult(result)
    require.False(t, truncated, "binary exactly at limit should NOT truncate")

    _, hasTruncatedFlag := sanitized["data_truncated"]
    require.False(t, hasTruncatedFlag)
}

func TestPayload_BinaryOneOverLimit(t *testing.T) {
    result := map[string]any{
        "data": make([]byte, MaxResultValueSize+1),
    }
    sanitized, truncated := sanitizeResult(result)
    require.True(t, truncated)
    require.Equal(t, "[BINARY_DATA_TRUNCATED]", sanitized["data"])
    require.True(t, sanitized["data_truncated"].(bool))
}

// ---------------------------------------------------------------------------
// Boundary Tests: Complex Objects (JSON serialization size)
// ---------------------------------------------------------------------------

func TestPayload_ObjectExactlyAtLimit(t *testing.T) {
    // Build an object whose JSON serialization is exactly at MaxResultValueSize.
    // This is approximate — JSON encoding adds overhead for keys and formatting.
    // Use a map with enough entries to get close to 50KB.
    smallMap := make(map[string]any)
    // Each entry: ~15 bytes overhead + value. Target ~49KB to stay under.
    for i := 0; i < 490; i++ {
        smallMap[fmt.Sprintf("k%03d", i)] = strings.Repeat("v", 90)
    }
    result := map[string]any{"data": smallMap}
    sanitized, truncated := sanitizeResult(result)
    require.False(t, truncated, "object under limit should NOT truncate")

    // Verify the object is preserved as-is (not replaced with string)
    _, isMap := sanitized["data"].(map[string]any)
    require.True(t, isMap, "object should be preserved as map")
}

func TestPayload_ObjectOverLimit(t *testing.T) {
    // Build an object whose JSON serialization exceeds MaxResultValueSize.
    largeMap := make(map[string]any)
    for i := 0; i < 1000; i++ {
        largeMap[fmt.Sprintf("key_%04d", i)] = strings.Repeat("v", 100)
    }
    result := map[string]any{"data": largeMap}
    sanitized, truncated := sanitizeResult(result)
    require.True(t, truncated)
    require.Equal(t, "[LARGE_OBJECT_TRUNCATED]", sanitized["data"])
    require.True(t, sanitized["data_truncated"].(bool))
}

// ---------------------------------------------------------------------------
// Multiple Large Fields
// ---------------------------------------------------------------------------

func TestPayload_MultipleLargeFields(t *testing.T) {
    // Two fields both over limit — both should be independently truncated.
    result := map[string]any{
        "field1": strings.Repeat("a", MaxResultValueSize+1),
        "field2": strings.Repeat("b", MaxResultValueSize+1),
    }
    sanitized, truncated := sanitizeResult(result)
    require.True(t, truncated)
    require.True(t, sanitized["field1_truncated"].(bool))
    require.True(t, sanitized["field2_truncated"].(bool))

    // Both fields should have truncation markers
    require.Contains(t, sanitized["field1"].(string), "...[TRUNCATED]")
    require.Contains(t, sanitized["field2"].(string), "...[TRUNCATED]")
}

// ---------------------------------------------------------------------------
// MergeResult Integration with Truncation
// ---------------------------------------------------------------------------

func TestPayload_MergeResult_LargeValueTruncated(t *testing.T) {
    ctx := NewMergedContext(map[string]any{"trigger": "data"})
    largeResult := map[string]any{
        "data": strings.Repeat("x", MaxResultValueSize+1),
    }
    ctx.MergeResult("test_action", largeResult)

    // Verify truncation happened in stored ActionResults
    storedResult := ctx.ActionResults["test_action"]
    require.Contains(t, storedResult["data"].(string), "...[TRUNCATED]")
    require.True(t, storedResult["data_truncated"].(bool))

    // Verify truncation flag propagated to Flattened map
    require.True(t, ctx.Flattened["test_action.data_truncated"].(bool))
}

func TestPayload_MergeResult_SmallValuePreserved(t *testing.T) {
    ctx := NewMergedContext(map[string]any{"trigger": "data"})
    smallResult := map[string]any{
        "data": strings.Repeat("x", MaxResultValueSize), // Exactly at limit
    }
    ctx.MergeResult("test_action", smallResult)

    // Value should be preserved without truncation
    storedResult := ctx.ActionResults["test_action"]
    require.Equal(t, smallResult["data"], storedResult["data"])

    // No truncation flag
    _, hasTruncatedFlag := storedResult["data_truncated"]
    require.False(t, hasTruncatedFlag)
}

// ---------------------------------------------------------------------------
// MergedContext Size Stress Test
// ---------------------------------------------------------------------------

func TestPayload_MergedContext_LargeContextStillFunctions(t *testing.T) {
    // Add many results to approach the 200KB warning threshold.
    // Verify MergedContext still functions correctly at large sizes.
    ctx := NewMergedContext(map[string]any{"trigger": "data"})

    // Add 20 actions with ~9KB each ≈ 180KB total
    for i := 0; i < 20; i++ {
        ctx.MergeResult(
            fmt.Sprintf("action_%d", i),
            map[string]any{"data": strings.Repeat("x", 9*1024)},
        )
    }

    require.Equal(t, 20, len(ctx.ActionResults))
    require.Equal(t, "data", ctx.Flattened["trigger"])

    // Verify first and last action results are intact
    require.NotEmpty(t, ctx.ActionResults["action_0"]["data"])
    require.NotEmpty(t, ctx.ActionResults["action_19"]["data"])

    // Verify Flattened map has entries for all actions
    require.NotNil(t, ctx.Flattened["action_0.data"])
    require.NotNil(t, ctx.Flattened["action_19.data"])
}
```

**Notes**:
- All tests focus on **exact boundary conditions** (50,000 vs 50,001 bytes) — Phase 3 only tests 1000+ over
- `TestPayload_ObjectExactlyAtLimit` is approximate due to JSON encoding overhead
- No Phase 3 test duplication — primitives, nil, empty, immutability all covered there
- `MergeResult` integration tests verify the full path: handler result → sanitize → ActionResults → Flattened

---

### Task 3: Async Activity Completion Tests

**Status**: Pending

**Description**: Test the async activity lifecycle: `ExecuteAsyncActionActivity` returning `ErrResultPending`, task token handling, and `AsyncCompleter.Complete/Fail` paths.

**Testing Strategy (Two-Tier)**:
- **Tier 1 (Unit)**: Call `ExecuteAsyncActionActivity` directly (bypass SDK TestActivityEnvironment which may not support `ErrResultPending`). Test `AsyncCompleter` with mock `ActivityCompleter` interface.
- **Tier 2 (Integration, optional)**: Full async round-trip via real Temporal container (start workflow → async activity → external complete → workflow resumes). Guarded with `testing.Short()`.

**Prerequisite**: Extract `ActivityCompleter` interface from `AsyncCompleter` (see Prerequisites section above).

**Files**:
- `business/sdk/workflow/temporal/async_completer.go` (minor refactor: extract interface)
- `business/sdk/workflow/temporal/activities_async_test.go`

**Implementation Guide**:

```go
package temporal

import (
    "context"
    "encoding/json"
    "errors"
    "testing"

    "github.com/google/uuid"
    "github.com/stretchr/testify/require"
    "github.com/timmaaaz/ichor/business/sdk/workflow"
    "go.temporal.io/sdk/activity"
)

// ---------------------------------------------------------------------------
// Mock Async Handler
// ---------------------------------------------------------------------------

type mockAsyncHandler struct {
    actionType     string
    startAsyncErr  error
    capturedToken  []byte
    capturedConfig json.RawMessage
    called         int
}

func (h *mockAsyncHandler) Execute(ctx context.Context, config json.RawMessage, execCtx workflow.ActionExecutionContext) (any, error) {
    return nil, errors.New("should not be called for async")
}

func (h *mockAsyncHandler) StartAsync(ctx context.Context, config json.RawMessage, execCtx workflow.ActionExecutionContext, taskToken []byte) error {
    h.called++
    h.capturedToken = taskToken
    h.capturedConfig = config
    return h.startAsyncErr
}

func (h *mockAsyncHandler) Validate(config json.RawMessage) error { return nil }
func (h *mockAsyncHandler) GetType() string                      { return h.actionType }
func (h *mockAsyncHandler) SupportsManualExecution() bool        { return false }
func (h *mockAsyncHandler) IsAsync() bool                        { return true }
func (h *mockAsyncHandler) GetDescription() string               { return "mock async" }

// ---------------------------------------------------------------------------
// Async Activity Tests (Direct Function Call — Tier 1)
// ---------------------------------------------------------------------------
//
// SDK TestActivityEnvironment may not support activity.ErrResultPending.
// Test ExecuteAsyncActionActivity directly. The function uses
// activity.GetInfo(ctx) for the task token, so we need an activity context.
// If direct call doesn't provide activity info, verify handler behavior
// via the mock's captured state.

func TestAsyncActivity_CallsStartAsync(t *testing.T) {
    handler := &mockAsyncHandler{actionType: "send_email"}
    asyncReg := NewAsyncRegistry()
    asyncReg.Register("send_email", handler) // TWO args: actionType, handler

    activities := &Activities{
        Registry:      workflow.NewActionRegistry(),
        AsyncRegistry: asyncReg,
    }

    input := ActionActivityInput{
        ActionID:    uuid.New(),
        ActionName:  "send_email_1",
        ActionType:  "send_email",
        Config:      json.RawMessage(`{"to": "user@example.com"}`),
        Context:     map[string]any{"entity_id": "123"},
        RuleID:      uuid.New(),
        ExecutionID: uuid.New(),
        RuleName:    "test-rule",
    }

    // Note: Calling directly requires activity context for GetInfo().TaskToken.
    // If TestActivityEnvironment supports it, use that. Otherwise, verify
    // the handler was called via mock state after SDK test environment call.
    // The key assertion: handler.called == 1 after activity executes.
    _ = activities
    _ = input
    // Implementation will determine which approach works.
    // See Gotchas section for TestActivityEnvironment + ErrResultPending notes.
}

func TestAsyncActivity_StartAsyncError_ReturnsError(t *testing.T) {
    handler := &mockAsyncHandler{
        actionType:    "send_email",
        startAsyncErr: errors.New("queue unavailable"),
    }
    asyncReg := NewAsyncRegistry()
    asyncReg.Register("send_email", handler)

    activities := &Activities{
        Registry:      workflow.NewActionRegistry(),
        AsyncRegistry: asyncReg,
    }

    input := ActionActivityInput{
        ActionID:    uuid.New(),
        ActionName:  "send_email_1",
        ActionType:  "send_email",
        Config:      json.RawMessage(`{}`),
        Context:     map[string]any{},
        RuleID:      uuid.New(),
        ExecutionID: uuid.New(),
        RuleName:    "test-rule",
    }

    // When StartAsync returns error, activity should return that error
    // (NOT ErrResultPending). The error wrapping should include action name.
    _ = activities
    _ = input
    // Verify: err != nil, err != activity.ErrResultPending
    // Verify: err.Error() contains "send_email_1" and "queue unavailable"
}

func TestAsyncActivity_UnknownAsyncType(t *testing.T) {
    asyncReg := NewAsyncRegistry()
    // Don't register any handlers

    activities := &Activities{
        Registry:      workflow.NewActionRegistry(),
        AsyncRegistry: asyncReg,
    }

    input := ActionActivityInput{
        ActionID:    uuid.New(),
        ActionName:  "unknown_1",
        ActionType:  "nonexistent_type",
        Config:      json.RawMessage(`{}`),
        Context:     map[string]any{},
        RuleID:      uuid.New(),
        ExecutionID: uuid.New(),
        RuleName:    "test-rule",
    }

    // Verify: error returned containing "no handler registered"
    _ = activities
    _ = input
}

// ---------------------------------------------------------------------------
// Mock ActivityCompleter (narrow interface from async_completer.go)
// ---------------------------------------------------------------------------

type mockActivityCompleter struct {
    completedToken  []byte
    completedResult any
    completedErr    error
    failedToken     []byte
    failedErr       error
    returnErr       error
}

func (m *mockActivityCompleter) CompleteActivity(ctx context.Context, taskToken []byte, result any, err error) error {
    if err != nil {
        // Fail path
        m.failedToken = taskToken
        m.failedErr = err
    } else {
        // Complete path
        m.completedToken = taskToken
        m.completedResult = result
    }
    return m.returnErr
}

// ---------------------------------------------------------------------------
// AsyncCompleter Tests (Mock ActivityCompleter Interface)
// ---------------------------------------------------------------------------

func TestAsyncCompleter_Complete_Success(t *testing.T) {
    mock := &mockActivityCompleter{}
    completer := NewAsyncCompleter(mock)

    token := []byte("test-token-123")
    result := ActionActivityOutput{
        Success: true,
        Result:  map[string]any{"status": "processed"},
    }

    err := completer.Complete(context.Background(), token, result)
    require.NoError(t, err)
    require.Equal(t, token, mock.completedToken)
    require.NotNil(t, mock.completedResult)
}

func TestAsyncCompleter_Complete_EmptyResult(t *testing.T) {
    mock := &mockActivityCompleter{}
    completer := NewAsyncCompleter(mock)

    err := completer.Complete(context.Background(), []byte("token"), ActionActivityOutput{})
    require.NoError(t, err)
    require.NotNil(t, mock.completedResult) // Zero-value struct still passed
}

func TestAsyncCompleter_Fail_Success(t *testing.T) {
    mock := &mockActivityCompleter{}
    completer := NewAsyncCompleter(mock)

    token := []byte("test-token-456")
    activityErr := errors.New("processing failed")

    err := completer.Fail(context.Background(), token, activityErr)
    require.NoError(t, err)
    require.Equal(t, token, mock.failedToken)
    require.Equal(t, activityErr, mock.failedErr)
}

func TestAsyncCompleter_Complete_ClientError(t *testing.T) {
    mock := &mockActivityCompleter{returnErr: errors.New("temporal unavailable")}
    completer := NewAsyncCompleter(mock)

    err := completer.Complete(context.Background(), []byte("token"), ActionActivityOutput{})
    require.Error(t, err)
    require.Contains(t, err.Error(), "temporal unavailable")
}

func TestAsyncCompleter_Fail_ClientError(t *testing.T) {
    mock := &mockActivityCompleter{returnErr: errors.New("temporal unavailable")}
    completer := NewAsyncCompleter(mock)

    err := completer.Fail(context.Background(), []byte("token"), errors.New("activity failed"))
    require.Error(t, err)
    require.Contains(t, err.Error(), "temporal unavailable")
}
```

**Notes**:
- `AsyncRegistry.Register` takes TWO arguments: `Register(actionType string, handler AsyncActivityHandler)`
- `mockActivityCompleter` implements the extracted `ActivityCompleter` interface (just `CompleteActivity`)
- `NewAsyncCompleter` constructor changes from `client.Client` to `ActivityCompleter` — backwards-compatible
- If `TestActivityEnvironment` supports `ErrResultPending`, use it; if not, test handler behavior via mock state
- The `CompleteActivity` mock distinguishes Complete vs Fail by checking the `err` parameter (nil = complete, non-nil = fail)

---

### Task 4: Error Handling Tests

**Status**: Pending

**Description**: Test error propagation through workflow execution: activity failures, branch failures in parallel execution, fire-and-forget error isolation, and retry policy enforcement for different action types. Uses Phase 11 helpers (`testActionHandler`, `setupTestEnv`, `wfLinearGraph`, etc.).

**Files**:
- `business/sdk/workflow/temporal/workflow_errors_test.go`

**Implementation Guide**:

```go
package temporal

import (
    "context"
    "encoding/json"
    "errors"
    "fmt"
    "testing"

    "github.com/google/uuid"
    "github.com/stretchr/testify/require"
    "github.com/timmaaaz/ichor/business/sdk/workflow"
)

// ---------------------------------------------------------------------------
// retryTestHandler: testActionHandler variant for retry testing
// ---------------------------------------------------------------------------

// retryTestHandler fails for the first failCount calls, then succeeds.
// Uses an atomic counter to track attempts across Temporal retry attempts.
type retryTestHandler struct {
    actionType string
    failCount  int
    result     any
    callCount  int
}

func (h *retryTestHandler) Execute(ctx context.Context, config json.RawMessage, execCtx workflow.ActionExecutionContext) (any, error) {
    h.callCount++
    if h.callCount <= h.failCount {
        return nil, fmt.Errorf("attempt %d/%d failed", h.callCount, h.failCount)
    }
    return h.result, nil
}

func (h *retryTestHandler) Validate(config json.RawMessage) error { return nil }
func (h *retryTestHandler) GetType() string                      { return h.actionType }
func (h *retryTestHandler) SupportsManualExecution() bool        { return false }
func (h *retryTestHandler) IsAsync() bool                        { return false }
func (h *retryTestHandler) GetDescription() string               { return "retry test" }

// ---------------------------------------------------------------------------
// Activity Failure → Workflow Failure
// ---------------------------------------------------------------------------

func TestError_ActivityFailure_FailsWorkflow(t *testing.T) {
    handler := &testActionHandler{
        actionType: "fail_action",
        err:        errors.New("action execution failed"),
    }
    env := setupTestEnv(t, newTestRegistry(handler))

    graph, _ := wfLinearGraph("fail_action")
    input := WorkflowInput{
        RuleID:      uuid.New(),
        RuleName:    "error-test",
        ExecutionID: uuid.New(),
        Graph:       graph,
        TriggerData: map[string]any{},
    }

    env.ExecuteWorkflow(ExecuteGraphWorkflow, input)
    require.True(t, env.IsWorkflowCompleted())
    err := env.GetWorkflowError()
    require.Error(t, err)
    // Error should contain action name for diagnostics
    require.Contains(t, err.Error(), "fail_action")
}

func TestError_ActivityFailure_MidChain_StopsExecution(t *testing.T) {
    // Graph: A (success) → B (fail) → C (success)
    // Verify: A runs, B fails, C does NOT run, workflow errors
    handlerA := &testActionHandler{actionType: "action_a", result: map[string]any{"ok": true}}
    handlerB := &testActionHandler{actionType: "action_b", err: errors.New("mid-chain failure")}
    handlerC := &testActionHandler{actionType: "action_c", result: map[string]any{"ok": true}}

    reg := newTestRegistry(handlerA, handlerB, handlerC)
    env := setupTestEnv(t, reg)

    graph, _ := wfLinearGraph("action_a", "action_b", "action_c")
    input := WorkflowInput{
        RuleID:      uuid.New(),
        RuleName:    "mid-chain-error",
        ExecutionID: uuid.New(),
        Graph:       graph,
        TriggerData: map[string]any{},
    }

    env.ExecuteWorkflow(ExecuteGraphWorkflow, input)
    require.Error(t, env.GetWorkflowError())
    require.Equal(t, 1, handlerA.called)  // Ran successfully
    require.Equal(t, 1, handlerB.called)  // Failed
    require.Equal(t, 0, handlerC.called)  // Never reached
}

func TestError_ActivityFailure_ErrorContainsActionName(t *testing.T) {
    handler := &testActionHandler{
        actionType: "specific_action",
        err:        errors.New("detailed error message"),
    }
    env := setupTestEnv(t, newTestRegistry(handler))

    graph, _ := wfLinearGraph("specific_action")
    input := WorkflowInput{
        RuleID:      uuid.New(),
        RuleName:    "error-detail-test",
        ExecutionID: uuid.New(),
        Graph:       graph,
        TriggerData: map[string]any{},
    }

    env.ExecuteWorkflow(ExecuteGraphWorkflow, input)
    err := env.GetWorkflowError()
    require.Error(t, err)
    // Error wrapping in executeSingleAction: "execute action %s (%s): %w"
    require.Contains(t, err.Error(), "specific_action")
    require.Contains(t, err.Error(), "detailed error message")
}

// ---------------------------------------------------------------------------
// Branch Failure in Parallel Execution
// ---------------------------------------------------------------------------

func TestError_ParallelBranchFailure_FailsWorkflow(t *testing.T) {
    // Parallel: fork → branchA (success) and branchB (fail) → convergence
    // Verify: workflow fails because branchB failed.
    // Note: SDK test suite executes child workflows synchronously.
    handlerFork := &testActionHandler{
        actionType: "fork_action",
        result:     map[string]any{"branch_taken": "true_branch"}, // Not used for parallel
    }
    handlerA := &testActionHandler{actionType: "branch_a", result: map[string]any{"ok": true}}
    handlerB := &testActionHandler{actionType: "branch_b", err: errors.New("branch B failed")}
    handlerMerge := &testActionHandler{actionType: "merge_action", result: map[string]any{"merged": true}}

    reg := newTestRegistry(handlerFork, handlerA, handlerB, handlerMerge)
    env := setupTestEnv(t, reg)

    graph, _ := wfParallelGraph("fork_action", "branch_a", "branch_b", "merge_action")
    input := WorkflowInput{
        RuleID:      uuid.New(),
        RuleName:    "parallel-error-test",
        ExecutionID: uuid.New(),
        Graph:       graph,
        TriggerData: map[string]any{},
    }

    env.ExecuteWorkflow(ExecuteGraphWorkflow, input)
    err := env.GetWorkflowError()
    require.Error(t, err)
    require.Contains(t, err.Error(), "branch B failed")
    require.Equal(t, 0, handlerMerge.called, "merge should NOT run when a branch fails")
}

func TestError_ParallelBothBranchesFail(t *testing.T) {
    handlerFork := &testActionHandler{actionType: "fork_action", result: map[string]any{}}
    handlerA := &testActionHandler{actionType: "branch_a", err: errors.New("branch A failed")}
    handlerB := &testActionHandler{actionType: "branch_b", err: errors.New("branch B failed")}
    handlerMerge := &testActionHandler{actionType: "merge_action", result: map[string]any{}}

    reg := newTestRegistry(handlerFork, handlerA, handlerB, handlerMerge)
    env := setupTestEnv(t, reg)

    graph, _ := wfParallelGraph("fork_action", "branch_a", "branch_b", "merge_action")
    input := WorkflowInput{
        RuleID:      uuid.New(),
        RuleName:    "both-branches-fail",
        ExecutionID: uuid.New(),
        Graph:       graph,
        TriggerData: map[string]any{},
    }

    env.ExecuteWorkflow(ExecuteGraphWorkflow, input)
    require.Error(t, env.GetWorkflowError())
    require.Equal(t, 0, handlerMerge.called)
}

// ---------------------------------------------------------------------------
// Fire-and-Forget Error Isolation
// ---------------------------------------------------------------------------

func TestError_FireAndForget_ErrorIsolation(t *testing.T) {
    // Fire-and-forget: fork → branchA (success) and branchB (fail)
    // Verify: parent workflow SUCCEEDS (fire-and-forget errors are logged, not propagated).
    // NOTE: SDK test suite may execute child workflows synchronously,
    // which could affect fire-and-forget behavior. If test fails due to
    // SDK limitations, document and guard with testing.Short().
    handlerFork := &testActionHandler{actionType: "fork_action", result: map[string]any{}}
    handlerA := &testActionHandler{actionType: "branch_a", result: map[string]any{"ok": true}}
    handlerB := &testActionHandler{actionType: "branch_b", err: errors.New("fire-forget error")}

    reg := newTestRegistry(handlerFork, handlerA, handlerB)
    env := setupTestEnv(t, reg)

    graph, _ := wfFireForgetGraph("fork_action", "branch_a", "branch_b")
    input := WorkflowInput{
        RuleID:      uuid.New(),
        RuleName:    "fire-forget-error",
        ExecutionID: uuid.New(),
        Graph:       graph,
        TriggerData: map[string]any{},
    }

    env.ExecuteWorkflow(ExecuteGraphWorkflow, input)
    // Parent should succeed — fire-and-forget branches run independently.
    // executeFireAndForget returns nil immediately.
    err := env.GetWorkflowError()
    require.NoError(t, err, "fire-and-forget branch errors should NOT fail parent")
}

// ---------------------------------------------------------------------------
// Handler Not Found
// ---------------------------------------------------------------------------

func TestError_HandlerNotFound(t *testing.T) {
    // Empty registry — no handlers registered.
    env := setupTestEnv(t, newTestRegistry()) // empty registry

    graph, _ := wfLinearGraph("nonexistent_type")
    input := WorkflowInput{
        RuleID:      uuid.New(),
        RuleName:    "missing-handler",
        ExecutionID: uuid.New(),
        Graph:       graph,
        TriggerData: map[string]any{},
    }

    env.ExecuteWorkflow(ExecuteGraphWorkflow, input)
    err := env.GetWorkflowError()
    require.Error(t, err)
    require.Contains(t, err.Error(), "no handler registered")
}

// ---------------------------------------------------------------------------
// Retry Policy Tests
// ---------------------------------------------------------------------------
//
// SDK test suite uses mock time, so retry backoff (1s → 2s → 4s) happens
// instantly. Tests verify retry COUNT via handler state, not timing.

func TestError_RetryPolicy_RegularAction_SucceedsAfterRetries(t *testing.T) {
    // Regular action: MaximumAttempts=3.
    // Handler fails on first 2 calls, succeeds on 3rd.
    handler := &retryTestHandler{
        actionType: "retry_action",
        failCount:  2,
        result:     map[string]any{"status": "ok"},
    }

    reg := workflow.NewActionRegistry()
    reg.Register(handler)
    env := setupTestEnv(t, reg)

    graph, _ := wfLinearGraph("retry_action")
    input := WorkflowInput{
        RuleID:      uuid.New(),
        RuleName:    "retry-success",
        ExecutionID: uuid.New(),
        Graph:       graph,
        TriggerData: map[string]any{},
    }

    env.ExecuteWorkflow(ExecuteGraphWorkflow, input)
    require.NoError(t, env.GetWorkflowError())
    require.Equal(t, 3, handler.callCount, "should be called 3 times (2 failures + 1 success)")
}

func TestError_RetryPolicy_RegularAction_ExhaustsRetries(t *testing.T) {
    // Regular action: MaximumAttempts=3.
    // Handler always fails. After 3 attempts, workflow should fail.
    handler := &retryTestHandler{
        actionType: "always_fail",
        failCount:  100, // Always fail
    }

    reg := workflow.NewActionRegistry()
    reg.Register(handler)
    env := setupTestEnv(t, reg)

    graph, _ := wfLinearGraph("always_fail")
    input := WorkflowInput{
        RuleID:      uuid.New(),
        RuleName:    "retry-exhausted",
        ExecutionID: uuid.New(),
        Graph:       graph,
        TriggerData: map[string]any{},
    }

    env.ExecuteWorkflow(ExecuteGraphWorkflow, input)
    require.Error(t, env.GetWorkflowError())
    require.Equal(t, 3, handler.callCount, "should attempt exactly 3 times")
}

func TestError_RetryPolicy_AsyncAction_NoRetry(t *testing.T) {
    // Async action (send_email): MaximumAttempts=1.
    // Action type "send_email" is in asyncActionTypes map, so activityOptions
    // sets MaximumAttempts=1. Activity fails → workflow fails immediately.
    handler := &testActionHandler{
        actionType: "send_email",
        err:        errors.New("email service unavailable"),
    }
    env := setupTestEnv(t, newTestRegistry(handler))

    graph, _ := wfLinearGraph("send_email")
    input := WorkflowInput{
        RuleID:      uuid.New(),
        RuleName:    "async-no-retry",
        ExecutionID: uuid.New(),
        Graph:       graph,
        TriggerData: map[string]any{},
    }

    env.ExecuteWorkflow(ExecuteGraphWorkflow, input)
    require.Error(t, env.GetWorkflowError())
    require.Equal(t, 1, handler.called, "async action should NOT retry (MaximumAttempts=1)")
}

func TestError_RetryPolicy_HumanAction_NoRetry(t *testing.T) {
    // Human action (manager_approval): MaximumAttempts=1.
    handler := &testActionHandler{
        actionType: "manager_approval",
        err:        errors.New("approval service down"),
    }
    env := setupTestEnv(t, newTestRegistry(handler))

    graph, _ := wfLinearGraph("manager_approval")
    input := WorkflowInput{
        RuleID:      uuid.New(),
        RuleName:    "human-no-retry",
        ExecutionID: uuid.New(),
        Graph:       graph,
        TriggerData: map[string]any{},
    }

    env.ExecuteWorkflow(ExecuteGraphWorkflow, input)
    require.Error(t, env.GetWorkflowError())
    require.Equal(t, 1, handler.called, "human action should NOT retry (MaximumAttempts=1)")
}
```

**Notes**:
- `retryTestHandler` uses a `callCount` field to track attempts — Temporal creates a new activity execution for each retry, but the handler struct is shared across retries in the SDK test suite
- Error messages should contain action name for diagnostics — verify with `require.Contains`
- Fire-and-forget error isolation may behave differently in SDK test suite vs real container — document if test needs adjustment
- `wfLinearGraph`, `wfParallelGraph`, `wfFireForgetGraph` are from Phase 11 (same package)
- Retry count verification: regular=3, async=1, human=1

---

## Validation Criteria

### Continue-As-New (5 criteria)
- [ ] `shouldContinueAsNew(10000)` returns `false` (threshold uses `>` not `>=`)
- [ ] `shouldContinueAsNew(10001)` returns `true`
- [ ] `shouldContinueAsNew(0)` returns `false`
- [ ] `ContinuationState` with 100+ ActionResults entries survives JSON round-trip without data loss
- [ ] Large integers (`>2^53`) in TriggerData preserve as `float64` after round-trip (documented limitation)

### Payload Limits (8 criteria)
- [ ] String at exactly 50,000 bytes does NOT truncate
- [ ] String at 50,001 bytes truncates with `...[TRUNCATED]` suffix and `_truncated` flag
- [ ] Binary data at exactly 50,000 bytes does NOT truncate
- [ ] Binary data at 50,001 bytes replaced with `[BINARY_DATA_TRUNCATED]` and `_truncated` flag
- [ ] Complex object JSON over 50KB replaced with `[LARGE_OBJECT_TRUNCATED]` and `_truncated` flag
- [ ] Two fields both over limit: both independently truncated with separate `_truncated` flags
- [ ] `MergeResult` with large value → truncation applied in `ActionResults` AND propagated to `Flattened`
- [ ] MergedContext with 20 actions (~180KB) still functions correctly

### Async Activities (5 criteria)
- [ ] `ExecuteAsyncActionActivity` calls `handler.StartAsync` with task token and config
- [ ] StartAsync error returns as activity error (NOT `ErrResultPending`)
- [ ] Unknown async action type returns error containing "no handler registered"
- [ ] `AsyncCompleter.Complete` forwards result and token to `ActivityCompleter`
- [ ] `AsyncCompleter.Fail` forwards error and token to `ActivityCompleter`

### Error Handling (6 criteria)
- [ ] Activity failure propagates to workflow error with action name in message
- [ ] Mid-chain failure (A→B→C, B fails): A runs, B fails, C does NOT run
- [ ] Parallel branch failure fails entire parallel group (merge not reached)
- [ ] Fire-and-forget branch failure does NOT fail parent workflow
- [ ] Handler-not-found returns error containing "no handler registered"
- [ ] Error wrapping includes action name and original error message

### Retry Policies (4 criteria)
- [ ] Regular action: handler called 3 times (2 failures + 1 success) → workflow succeeds
- [ ] Regular action: handler called 3 times (all failures) → workflow fails
- [ ] Async action (`send_email`): handler called exactly 1 time → workflow fails immediately
- [ ] Human action (`manager_approval`): handler called exactly 1 time → workflow fails immediately

### Build & Test (2 criteria)
- [ ] `go test ./business/sdk/workflow/temporal/... -count=1` passes (all tests including Phase 12)
- [ ] `go vet ./business/sdk/workflow/temporal/...` clean

---

## Deliverables

**New test files**:
- `business/sdk/workflow/temporal/workflow_continueasnew_test.go`
- `business/sdk/workflow/temporal/workflow_payload_test.go`
- `business/sdk/workflow/temporal/activities_async_test.go`
- `business/sdk/workflow/temporal/workflow_errors_test.go`

**Modified implementation files** (minor refactors for testability):
- `business/sdk/workflow/temporal/workflow.go` — extract `shouldContinueAsNew(int) bool`
- `business/sdk/workflow/temporal/async_completer.go` — extract `ActivityCompleter` interface

---

## Gotchas & Tips

### Common Pitfalls

- **SDK test suite and ErrResultPending**: The SDK `TestActivityEnvironment` may not support `activity.ErrResultPending` correctly. If `env.ExecuteActivity()` fails with ErrResultPending, test `ExecuteAsyncActionActivity` by calling it directly. Verify handler behavior via mock state (`handler.called`, `handler.capturedToken`). Document which approach works.
- **JSON float64 conversion**: When testing ContinuationState round-trips, ALL numbers become `float64` after JSON deserialization. An `int(42)` in the original becomes `float64(42)` after round-trip. Always use `float64` literals in expected values.
- **Phase 3 test overlap**: `models_test.go` already has 14 tests covering `sanitizeResult` basics (string, bytes, object, small values, nil, immutability, numeric fast-path). Phase 12 tests ONLY boundary conditions (exact limit vs one over) and `MergeResult` integration. Do NOT duplicate Phase 3 tests.
- **Retry timing in SDK test suite**: SDK test suite uses mock time, so retry backoff (1s → 2s → 4s) happens instantly. Verify retry COUNT via handler state counters, not timing.
- **`retryTestHandler` for retry tests**: Need per-call behavior (fail first N times, then succeed). `retryTestHandler` uses a `callCount` field and `failCount` threshold. Create this as a separate type in `workflow_errors_test.go` to avoid modifying Phase 11's `testActionHandler`.
- **AsyncRegistry.Register signature**: Takes TWO arguments: `Register(actionType string, handler AsyncActivityHandler)`. NOT `Register(handler)`.
- **sanitizeResult nil input**: Returns `make(map[string]any), false` — a non-nil empty map. NOT nil.
- **Fire-and-forget in SDK test suite**: Child workflows may execute synchronously in the SDK test suite. `executeFireAndForget` returns `nil` immediately (parent doesn't wait), but SDK may still propagate child errors. If the test fails due to SDK behavior, document the limitation and guard with `testing.Short()`.
- **ActivityCompleter interface extraction**: Must be done BEFORE writing AsyncCompleter tests. The extraction is backwards-compatible (`client.Client` satisfies `ActivityCompleter`).
- **shouldContinueAsNew extraction**: Use `int` parameter (not `int64`) since `GetCurrentHistoryLength()` returns `int`. Already unexported (lowercase) in code examples.

### Tips

- Start with **Task 2 (Payload)** — pure unit tests, fastest to implement, no dependencies
- Then **Task 1 (Continue-As-New)** — after extracting `shouldContinueAsNew`, also pure unit tests
- Then **Task 3 (Async)** — after extracting `ActivityCompleter` interface
- Finally **Task 4 (Errors)** — uses Phase 11 helpers, most setup required
- Use `require.ErrorContains(t, err, "expected substring")` for error message assertions
- All handlers and helpers from Phase 11 are in the same package (`temporal`) — directly accessible

### Expected Failure Modes

| Scenario | Expected Behavior |
|----------|-------------------|
| `shouldContinueAsNew(10000)` | `false` (uses `>` not `>=`) |
| `shouldContinueAsNew(10001)` | `true` (triggers Continue-As-New) |
| 50KB string truncation | String truncated at 50,000 chars + `...[TRUNCATED]`, `_truncated` flag set |
| Async handler StartAsync fails | Activity returns error (NOT `ErrResultPending`) |
| Activity exhausts all retries (3) | Workflow fails with last attempt's error |
| Branch failure in parallel | Entire parallel group fails, convergence never reached |
| Fire-and-forget failure | Error logged by child, parent workflow unaffected |

---

## Testing Strategy

### Unit Tests

**Pure function tests** (no Temporal infrastructure):
- `shouldContinueAsNew` — threshold boundary (extracted function)
- `sanitizeResult` — exact boundary conditions (50,000 vs 50,001 bytes)
- `MergedContext.MergeResult` — large value integration
- `MergedContext` JSON round-trip — ContinuationState preservation with various data shapes
- `AsyncCompleter` — Complete/Fail with mock `ActivityCompleter`

**SDK test suite** (in-process, no Docker):
- Activity error propagation to workflow
- Mid-chain failure stops execution
- Branch failure in parallel execution
- Fire-and-forget error isolation
- Retry policy enforcement (handler call counts)

### Integration Tests

**Real Temporal container** (Docker, guarded with `testing.Short()`):
- Full async activity completion round-trip (optional, advanced)
- Not planned for Phase 12 — covered in Phase 11 replay tests

### Test Execution

```bash
# Run all Phase 12 tests
go test -v ./business/sdk/workflow/temporal/... -run "TestShouldContinueAsNew|TestContinuationState|TestPayload|TestAsyncActivity|TestAsyncCompleter|TestError"

# Run only threshold + context tests (fast)
go test -v ./business/sdk/workflow/temporal/... -run "TestShouldContinueAsNew|TestContinuationState"

# Run only payload boundary tests (fast)
go test -v ./business/sdk/workflow/temporal/... -run "TestPayload"

# Run only async tests
go test -v ./business/sdk/workflow/temporal/... -run "TestAsyncActivity|TestAsyncCompleter"

# Run only error handling tests
go test -v ./business/sdk/workflow/temporal/... -run "TestError"

# Run all temporal package tests
go test -v ./business/sdk/workflow/temporal/...
```

---

## Scope Boundary

### In Scope (Phase 12)
- `shouldContinueAsNew` threshold logic (extracted pure function)
- ContinuationState JSON round-trip fidelity (action results, float64, special chars)
- Payload sanitization boundary conditions (exact 50KB limit)
- `MergeResult` integration with large values
- MergedContext size stress test (~200KB)
- Async activity handler integration (`StartAsync` calls, errors)
- `AsyncCompleter` lifecycle (Complete/Fail with mock interface)
- Error propagation (activity → workflow, branch → parallel, fire-and-forget isolation)
- Retry policy enforcement (3 attempts regular, 1 attempt async/human)
- Handler-not-found errors

### Out of Scope (Phase 11)
- Happy-path workflow execution (single, sequential, parallel)
- Replay determinism testing
- Basic test setup and graph builder helpers

### Out of Scope (Phase 3)
- Basic `sanitizeResult` behavior (string/bytes/object 1000+ over limit)
- Primitive pass-through, nil handling, input immutability
- `NewMergedContext`, `MergeResult` basics, `Clone` behavior

### Out of Scope (Not Unit-Testable)
- Actual Continue-As-New trigger at 10K events via SDK test suite (SDK limitation; `shouldContinueAsNew` extraction covers the logic)
- Temporal retry backoff timing (non-deterministic; verify attempt counts only)
- Activity heartbeat behavior (requires real worker)

### Out of Scope (Other)
- Graph executor unit tests (Phase 10)
- Production monitoring and alerting
- Performance benchmarks under load
- Multi-workflow concurrency stress testing

---

## Commands Reference

```bash
# Start this phase
/workflow-temporal-next

# Validate this phase
/workflow-temporal-validate 12

# Review plan before implementing
/workflow-temporal-plan-review 12

# Review code after implementing
/workflow-temporal-review 12
```
