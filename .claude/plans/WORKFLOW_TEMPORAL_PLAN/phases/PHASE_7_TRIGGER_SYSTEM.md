# Phase 7: Trigger System

**Category**: backend
**Status**: Pending
**Dependencies**: Phase 3 (Core Models & Context - COMPLETED, provides `WorkflowInput`, `GraphDefinition`, `ActionNode`, `ActionEdge`), Phase 5 (Workflow Implementation - provides `ExecuteGraphWorkflow`), Phase 6 (Activities & Async - provides `ExecuteActionActivity`, `ExecuteAsyncActionActivity`)

---

## Overview

Implement the bridge between entity events and Temporal workflow execution. The `WorkflowTrigger` replaces the existing `QueueManager.processWorkflowEvent()` → `Engine.ExecuteWorkflow()` entry point with a Temporal-native dispatch: entity event → rule matching → graph loading → `client.ExecuteWorkflow()`. This phase produces `trigger.go` (the trigger dispatcher) and `trigger_test.go` (unit tests with mocked dependencies).

The trigger system reuses the existing `TriggerProcessor` for rule matching (no reimplementation) and introduces the `EdgeStore` interface for loading graph definitions from PostgreSQL (implemented in Phase 8). The key difference from the existing system: instead of executing workflow logic in-process, the trigger starts a Temporal workflow that runs in the workflow-worker service.

## Goals

1. **Implement `WorkflowTrigger` that receives entity events and dispatches Temporal workflows** via `client.ExecuteWorkflow`, with rule matching via the existing `TriggerProcessor` and `AutomationRuleView` matching
2. **Define `EdgeStore` interface for loading graph definitions from PostgreSQL** - two methods (`QueryActionsByRule`, `QueryEdgesByRule`) that return `temporal.ActionNode`/`temporal.ActionEdge` types (implemented by `edgedb.Store` in Phase 8)
3. **Write unit tests with mock `EdgeStore` and mock Temporal client** for trigger dispatch logic, rule matching, error handling, and workflow ID generation

## Prerequisites

- Phase 3 complete: `WorkflowInput`, `GraphDefinition`, `ActionNode`, `ActionEdge`, `TaskQueue` constant
- Phase 5 complete: `ExecuteGraphWorkflow` function (referenced by name in `client.ExecuteWorkflow`)
- Phase 6 complete: Activities registered (needed for workflow execution, though not directly imported by trigger)
- Existing `workflow.TriggerProcessor` in `business/sdk/workflow/trigger.go` (rule matching)
- Existing `workflow.TriggerEvent` in `business/sdk/workflow/models.go` (event structure)
- Existing `workflow.AutomationRuleView` in `business/sdk/workflow/models.go` (matched rule data)
- Temporal SDK available in `go.mod`/`vendor` (from Phase 1)

---

## Go Package Structure

```
business/sdk/workflow/temporal/
    models.go              <- Phase 3 (COMPLETED)
    models_test.go         <- Phase 3 (COMPLETED)
    graph_executor.go      <- Phase 4 (COMPLETED)
    graph_executor_test.go <- Phase 4 (COMPLETED)
    workflow.go            <- Phase 5
    activities.go          <- Phase 6
    activities_async.go    <- Phase 6
    async_completer.go     <- Phase 6
    trigger.go             <- THIS PHASE (Task 1)
    trigger_test.go        <- THIS PHASE (Task 2)
```

---

## Existing Systems (DO NOT MODIFY)

### Current Trigger Flow (Being Replaced)

```
Entity Change
    → delegate event fires (non-blocking)
    → DelegateHandler bridges to EventPublisher
    → EventPublisher queues to RabbitMQ
    → QueueManager consumer picks up event
    → QueueManager.processWorkflowEvent()
    → Engine.ExecuteWorkflow(ctx, event)         ← WE REPLACE THIS
        → TriggerProcessor.ProcessEvent()        ← WE REUSE THIS
        → ActionExecutor.ExecuteRuleActionsGraph() ← REPLACED BY TEMPORAL
```

### New Trigger Flow (This Phase)

```
Entity Change
    → delegate event fires (non-blocking)
    → DelegateHandler bridges to EventPublisher
    → EventPublisher queues to RabbitMQ
    → QueueManager consumer picks up event
    → WorkflowTrigger.OnEntityEvent(ctx, event)  ← THIS PHASE
        → TriggerProcessor.ProcessEvent()         ← REUSED (rule matching)
        → EdgeStore.QueryActionsByRule()           ← NEW INTERFACE (Phase 8 impl)
        → EdgeStore.QueryEdgesByRule()             ← NEW INTERFACE (Phase 8 impl)
        → client.ExecuteWorkflow()                 ← TEMPORAL DISPATCH
            → ExecuteGraphWorkflow()               ← Phase 5
```

### Existing `TriggerProcessor` (Reused As-Is)

From `business/sdk/workflow/trigger.go`:

```go
func (tp *TriggerProcessor) ProcessEvent(ctx context.Context, event TriggerEvent) (*ProcessingResult, error)
```

Returns `ProcessingResult` containing:
```go
type ProcessingResult struct {
    TriggerEvent        TriggerEvent
    TotalRulesEvaluated int
    MatchedRules        []RuleMatchResult
    ProcessingTime      time.Duration
    Errors              []string
}

type RuleMatchResult struct {
    Rule             AutomationRuleView
    Matched          bool
    TriggerEvent     TriggerEvent
    ConditionResults []ConditionEvaluationResult
    MatchReason      string
    ExecutionContext map[string]interface{}
}
```

### Existing `TriggerEvent` Structure

From `business/sdk/workflow/models.go`:

```go
type TriggerEvent struct {
    EventType    string                 `json:"event_type"`      // on_create, on_update, on_delete
    EntityName   string                 `json:"entity_name"`
    EntityID     uuid.UUID              `json:"entity_id,omitempty"`
    FieldChanges map[string]FieldChange `json:"field_changes,omitempty"`
    Timestamp    time.Time              `json:"timestamp"`
    RawData      map[string]any         `json:"raw_data,omitempty"`
    UserID       uuid.UUID              `json:"user_id,omitempty"`
}
```

**Important**: The implementation plan reference uses a simplified `EntityEvent` struct. We should use the **existing `workflow.TriggerEvent`** directly instead of creating a new struct. This avoids data loss (FieldChanges, Timestamp, UserID would be lost) and maintains compatibility with `TriggerProcessor.ProcessEvent()`.

---

## Task Breakdown

### Task 1: Implement trigger.go

**Status**: Pending

**Description**: Implement the trigger dispatcher that receives entity events, matches them against automation rules via the existing `TriggerProcessor`, loads graph definitions from the `EdgeStore`, and starts Temporal workflows via `client.ExecuteWorkflow`. The `WorkflowTrigger` struct is the Temporal replacement for the existing `Engine.ExecuteWorkflow()` entry point.

**Notes**:
- `EdgeStore` interface - abstracts graph loading from PostgreSQL (implemented in Phase 8)
- `WorkflowTrigger` struct - holds logger, Temporal client, TriggerProcessor reference, EdgeStore
- `OnEntityEvent` - receives `workflow.TriggerEvent`, matches rules, loads graphs, starts workflows
- `loadGraphDefinition` - calls EdgeStore to build `GraphDefinition` from actions + edges
- Uses existing `workflow.TriggerEvent` (NOT a new EntityEvent struct) to preserve all event data
- Workflow ID format: `workflow-{ruleID}-{entityID}-{executionID}` for deduplication and traceability
- Continues to next rule on individual rule failures (fail-open per rule, not per event)

**Files**:
- `business/sdk/workflow/temporal/trigger.go`

**Implementation Guide**:

```go
package temporal

import (
    "context"
    "fmt"

    "github.com/google/uuid"
    "go.temporal.io/sdk/client"

    "github.com/timmaaaz/ichor/business/sdk/workflow"
    "github.com/timmaaaz/ichor/foundation/logger"
)

// =============================================================================
// EdgeStore Interface
// =============================================================================

// EdgeStore loads graph definitions (actions + edges) from the database.
// Implemented by stores/edgedb.Store in Phase 8.
type EdgeStore interface {
    // QueryActionsByRule returns all action nodes for a given automation rule.
    QueryActionsByRule(ctx context.Context, ruleID uuid.UUID) ([]ActionNode, error)

    // QueryEdgesByRule returns all action edges for a given automation rule.
    QueryEdgesByRule(ctx context.Context, ruleID uuid.UUID) ([]ActionEdge, error)
}

// =============================================================================
// WorkflowTrigger
// =============================================================================

// WorkflowTrigger handles entity events and dispatches Temporal workflows.
//
// This replaces the existing Engine.ExecuteWorkflow() entry point.
// Rule matching is delegated to the existing TriggerProcessor.
// Graph loading is delegated to the EdgeStore interface.
// Workflow execution is delegated to Temporal via client.ExecuteWorkflow.
//
// Usage (wired in Phase 9):
//
//     trigger := temporal.NewWorkflowTrigger(log, temporalClient, triggerProcessor, edgeStore)
//     // In QueueManager consumer or DelegateHandler:
//     trigger.OnEntityEvent(ctx, event)
type WorkflowTrigger struct {
    log              *logger.Logger
    temporalClient   client.Client
    triggerProcessor *workflow.TriggerProcessor
    edgeStore        EdgeStore
}

// NewWorkflowTrigger creates a new trigger handler.
func NewWorkflowTrigger(
    log *logger.Logger,
    tc client.Client,
    tp *workflow.TriggerProcessor,
    es EdgeStore,
) *WorkflowTrigger {
    return &WorkflowTrigger{
        log:              log,
        temporalClient:   tc,
        triggerProcessor: tp,
        edgeStore:        es,
    }
}

// OnEntityEvent processes an entity event by matching automation rules
// and starting Temporal workflows for each matched rule.
//
// Individual rule failures are logged and skipped (fail-open per rule).
// Returns an error only if rule matching itself fails.
func (t *WorkflowTrigger) OnEntityEvent(ctx context.Context, event workflow.TriggerEvent) error {
    t.log.Info(ctx, "Processing entity event for Temporal dispatch",
        "entity_name", event.EntityName,
        "event_type", event.EventType,
        "entity_id", event.EntityID,
    )

    // Match automation rules using existing TriggerProcessor
    result, err := t.triggerProcessor.ProcessEvent(ctx, event)
    if err != nil {
        return fmt.Errorf("process event: %w", err)
    }

    matchedCount := 0
    for _, rm := range result.MatchedRules {
        if rm.Matched {
            matchedCount++
        }
    }

    t.log.Info(ctx, "Rule matching complete",
        "total_evaluated", result.TotalRulesEvaluated,
        "matched", matchedCount,
    )

    // Start a Temporal workflow for each matched rule
    for _, rm := range result.MatchedRules {
        if !rm.Matched {
            continue
        }

        if err := t.startWorkflowForRule(ctx, event, rm); err != nil {
            t.log.Error(ctx, "Failed to start workflow for rule",
                "rule_id", rm.Rule.ID,
                "rule_name", rm.Rule.Name,
                "error", err,
            )
            // Continue to next rule - don't fail the entire event
            continue
        }
    }

    return nil
}

// startWorkflowForRule loads the graph definition and starts a Temporal workflow
// for a single matched rule.
func (t *WorkflowTrigger) startWorkflowForRule(
    ctx context.Context,
    event workflow.TriggerEvent,
    rm workflow.RuleMatchResult,
) error {
    // Load graph definition from database
    graph, err := t.loadGraphDefinition(ctx, rm.Rule.ID)
    if err != nil {
        return fmt.Errorf("load graph for rule %s: %w", rm.Rule.ID, err)
    }

    // Skip rules with empty graphs (no actions configured)
    if len(graph.Actions) == 0 {
        t.log.Info(ctx, "Skipping rule with empty graph",
            "rule_id", rm.Rule.ID,
            "rule_name", rm.Rule.Name,
        )
        return nil
    }

    // Generate unique execution ID
    executionID := uuid.New()

    // Create deterministic workflow ID for deduplication and traceability.
    // Format: workflow-{ruleID}-{entityID}-{executionID}
    // The executionID ensures uniqueness even for the same rule+entity combination.
    workflowID := fmt.Sprintf("workflow-%s-%s-%s",
        rm.Rule.ID,
        event.EntityID,
        executionID,
    )

    // Build workflow input with trigger data from the event
    input := WorkflowInput{
        RuleID:      rm.Rule.ID,
        ExecutionID: executionID,
        Graph:       graph,
        TriggerData: buildTriggerData(event),
    }

    // Start Temporal workflow
    workflowOptions := client.StartWorkflowOptions{
        ID:        workflowID,
        TaskQueue: TaskQueue,
    }

    we, err := t.temporalClient.ExecuteWorkflow(ctx, workflowOptions,
        ExecuteGraphWorkflow,
        input,
    )
    if err != nil {
        return fmt.Errorf("execute workflow: %w", err)
    }

    t.log.Info(ctx, "Started Temporal workflow",
        "rule_id", rm.Rule.ID,
        "rule_name", rm.Rule.Name,
        "workflow_id", workflowID,
        "run_id", we.GetRunID(),
    )

    return nil
}

// loadGraphDefinition loads actions and edges from the EdgeStore
// and assembles them into a GraphDefinition.
func (t *WorkflowTrigger) loadGraphDefinition(ctx context.Context, ruleID uuid.UUID) (GraphDefinition, error) {
    actions, err := t.edgeStore.QueryActionsByRule(ctx, ruleID)
    if err != nil {
        return GraphDefinition{}, fmt.Errorf("query actions: %w", err)
    }

    edges, err := t.edgeStore.QueryEdgesByRule(ctx, ruleID)
    if err != nil {
        return GraphDefinition{}, fmt.Errorf("query edges: %w", err)
    }

    return GraphDefinition{
        Actions: actions,
        Edges:   edges,
    }, nil
}

// buildTriggerData converts a TriggerEvent into a map suitable for
// WorkflowInput.TriggerData. This populates the initial MergedContext
// in the workflow, making event data available for template resolution.
func buildTriggerData(event workflow.TriggerEvent) map[string]any {
    data := make(map[string]any)

    // Include event metadata
    data["event_type"] = event.EventType
    data["entity_name"] = event.EntityName
    data["entity_id"] = event.EntityID.String()
    data["user_id"] = event.UserID.String()
    data["timestamp"] = event.Timestamp.String()

    // Include raw entity data (the entity snapshot)
    for k, v := range event.RawData {
        data[k] = v
    }

    // Include field changes for update events
    if len(event.FieldChanges) > 0 {
        changes := make(map[string]any, len(event.FieldChanges))
        for field, change := range event.FieldChanges {
            changes[field] = map[string]any{
                "old_value": change.OldValue,
                "new_value": change.NewValue,
            }
        }
        data["field_changes"] = changes
    }

    return data
}
```

**Design Decisions**:

1. **Uses existing `workflow.TriggerEvent`** - NOT a new `EntityEvent` struct. The implementation plan reference uses a simplified `EntityEvent` but that would lose FieldChanges, Timestamp, and UserID data. Using the existing type preserves all event data and maintains compatibility with `TriggerProcessor.ProcessEvent()`.

2. **Uses existing `workflow.TriggerProcessor`** - Rule matching logic is already implemented and battle-tested. No reimplementation needed. The trigger just calls `ProcessEvent()` and iterates matched rules.

3. **Fail-open per rule** - If one rule's graph fails to load or workflow fails to start, other rules still execute. This matches the existing `Engine.ExecuteWorkflow` behavior where individual rule failures don't block others.

4. **`buildTriggerData` flattens event** - Converts the typed `TriggerEvent` into a `map[string]any` for `WorkflowInput.TriggerData`. This makes all event data available in the `MergedContext` for template variable resolution (e.g., `{{entity_id}}`, `{{event_type}}`).

5. **Deterministic workflow ID** - Format `workflow-{ruleID}-{entityID}-{executionID}` enables:
   - Searching in Temporal UI by rule or entity
   - Preventing duplicate executions (Temporal rejects duplicate workflow IDs)
   - Tracing from entity event to workflow execution

6. **`EdgeStore` interface** - Clean boundary for database access. The trigger only knows about `ActionNode` and `ActionEdge` types (from Phase 3 models). The Phase 8 adapter translates from database rows.

---

### Task 2: Write Unit Tests for Trigger Logic

**Status**: Pending

**Description**: Write comprehensive unit tests for the trigger system using mock implementations of `EdgeStore` and Temporal's `client.Client`. Tests should cover: successful dispatch, empty graphs, rule matching with no matches, EdgeStore errors, Temporal client errors, multiple matched rules with partial failures, and `buildTriggerData` conversion.

**Notes**:
- Mock `EdgeStore` - returns configurable actions/edges per rule ID
- Mock Temporal client - use `go.temporal.io/sdk/mocks` package (`mocks.Client`)
- Test `buildTriggerData` independently (pure function, no mocks needed)
- Test `OnEntityEvent` end-to-end flow with mocked dependencies
- Mock `TriggerProcessor` - needs interface or test helper; check if existing code has a mock

**Files**:
- `business/sdk/workflow/temporal/trigger_test.go`

**Implementation Guide**:

```go
package temporal_test

import (
    "context"
    "encoding/json"
    "errors"
    "testing"
    "time"

    "github.com/google/uuid"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
    "github.com/stretchr/testify/require"

    sdkmocks "go.temporal.io/sdk/mocks"

    "github.com/timmaaaz/ichor/business/sdk/workflow"
    "github.com/timmaaaz/ichor/business/sdk/workflow/temporal"
)

// =============================================================================
// Mock EdgeStore
// =============================================================================

type mockEdgeStore struct {
    actions map[uuid.UUID][]temporal.ActionNode
    edges   map[uuid.UUID][]temporal.ActionEdge
    err     error // If set, all calls return this error
}

func newMockEdgeStore() *mockEdgeStore {
    return &mockEdgeStore{
        actions: make(map[uuid.UUID][]temporal.ActionNode),
        edges:   make(map[uuid.UUID][]temporal.ActionEdge),
    }
}

func (m *mockEdgeStore) QueryActionsByRule(ctx context.Context, ruleID uuid.UUID) ([]temporal.ActionNode, error) {
    if m.err != nil {
        return nil, m.err
    }
    return m.actions[ruleID], nil
}

func (m *mockEdgeStore) QueryEdgesByRule(ctx context.Context, ruleID uuid.UUID) ([]temporal.ActionEdge, error) {
    if m.err != nil {
        return nil, m.err
    }
    return m.edges[ruleID], nil
}

// =============================================================================
// Tests
// =============================================================================

// Test buildTriggerData conversion
func TestBuildTriggerData(t *testing.T) {
    t.Run("basic event", func(t *testing.T) {
        // Test that all event fields are present in trigger data
    })

    t.Run("update event with field changes", func(t *testing.T) {
        // Test that field_changes are properly nested
    })

    t.Run("empty raw data", func(t *testing.T) {
        // Test with nil/empty RawData
    })
}

// Test OnEntityEvent with successful dispatch
func TestOnEntityEvent_Success(t *testing.T) {
    // Setup: mock EdgeStore with valid graph, mock Temporal client
    // Action: call OnEntityEvent with matching rule
    // Assert: client.ExecuteWorkflow called with correct WorkflowInput
}

// Test OnEntityEvent with no matched rules
func TestOnEntityEvent_NoMatches(t *testing.T) {
    // Setup: TriggerProcessor returns no matched rules
    // Action: call OnEntityEvent
    // Assert: client.ExecuteWorkflow NOT called, no error returned
}

// Test OnEntityEvent with empty graph
func TestOnEntityEvent_EmptyGraph(t *testing.T) {
    // Setup: EdgeStore returns no actions for matched rule
    // Action: call OnEntityEvent
    // Assert: client.ExecuteWorkflow NOT called (skipped), no error returned
}

// Test OnEntityEvent with EdgeStore error
func TestOnEntityEvent_EdgeStoreError(t *testing.T) {
    // Setup: EdgeStore returns error for one rule
    // Action: call OnEntityEvent with two matched rules
    // Assert: first rule skipped with error, second rule still dispatched
}

// Test OnEntityEvent with Temporal client error
func TestOnEntityEvent_TemporalError(t *testing.T) {
    // Setup: Temporal client returns error on ExecuteWorkflow
    // Action: call OnEntityEvent
    // Assert: error logged, next rules still attempted
}

// Test OnEntityEvent with multiple matched rules
func TestOnEntityEvent_MultipleRules(t *testing.T) {
    // Setup: Two matched rules with valid graphs
    // Action: call OnEntityEvent
    // Assert: Two separate workflows started with unique IDs
}

// Test workflow ID format
func TestWorkflowIDFormat(t *testing.T) {
    // Assert: workflow ID matches "workflow-{ruleID}-{entityID}-{executionID}"
}

// Test that TriggerProcessor.ProcessEvent error is returned
func TestOnEntityEvent_ProcessEventError(t *testing.T) {
    // Setup: TriggerProcessor.ProcessEvent returns error
    // Action: call OnEntityEvent
    // Assert: error propagated (not swallowed)
}
```

**Testing Challenges**:

1. **Mocking `TriggerProcessor`** - The existing `TriggerProcessor` is a concrete struct, not an interface. Options:
   - Extract a `TriggerMatcher` interface from `WorkflowTrigger` (preferred - clean dependency injection)
   - Use the existing `TriggerProcessor` with a mocked `workflowBus` (harder, deeper mock chain)
   - Create a wrapper that satisfies an interface

   **Recommended approach**: Define a `RuleMatcher` interface in `trigger.go`:
   ```go
   // RuleMatcher matches entity events against automation rules.
   // Implemented by workflow.TriggerProcessor.
   type RuleMatcher interface {
       ProcessEvent(ctx context.Context, event workflow.TriggerEvent) (*workflow.ProcessingResult, error)
   }
   ```
   Then `WorkflowTrigger` depends on `RuleMatcher` instead of `*workflow.TriggerProcessor`. This enables clean mocking in tests while keeping the real `TriggerProcessor` as the production implementation.

2. **Mocking Temporal Client** - Use `go.temporal.io/sdk/mocks` package which provides `mocks.Client` with full testify mock support:
   ```go
   mockClient := &sdkmocks.Client{}
   mockClient.On("ExecuteWorkflow", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
       Return(&sdkmocks.WorkflowRun{}, nil)
   ```

---

## Validation Criteria

- [ ] `go build ./business/sdk/workflow/temporal/...` passes
- [ ] `trigger.go` compiles with correct imports (`go.temporal.io/sdk/client`, `business/sdk/workflow`)
- [ ] `EdgeStore` interface defined with `QueryActionsByRule` and `QueryEdgesByRule`
- [ ] `WorkflowTrigger` accepts existing `workflow.TriggerEvent` (not a new event type)
- [ ] `OnEntityEvent` calls `TriggerProcessor.ProcessEvent` (or `RuleMatcher` interface) for rule matching
- [ ] `OnEntityEvent` starts one Temporal workflow per matched rule
- [ ] Individual rule failures don't block other rules (fail-open per rule)
- [ ] `buildTriggerData` preserves event metadata AND raw entity data AND field changes
- [ ] Workflow ID format is `workflow-{ruleID}-{entityID}-{executionID}`
- [ ] `go test ./business/sdk/workflow/temporal/...` passes (all trigger tests)
- [ ] No import cycles between `temporal` package and `workflow` package
- [ ] Mock EdgeStore and mock Temporal client used in tests (no real DB or Temporal needed)

---

## Deliverables

- `business/sdk/workflow/temporal/trigger.go`
- `business/sdk/workflow/temporal/trigger_test.go`

---

## Gotchas & Tips

### Common Pitfalls

- **Don't create a new `EntityEvent` struct** - The implementation plan reference uses a simplified `EntityEvent` with only `EntityID`, `EntityName`, `EventType`, `Data`. But the existing `workflow.TriggerEvent` has more fields (FieldChanges, Timestamp, UserID). Use `workflow.TriggerEvent` directly to avoid data loss and maintain compatibility with `TriggerProcessor.ProcessEvent()`.

- **`TriggerProcessor` is a concrete struct, not an interface** - You can't directly mock it in tests. Extract a `RuleMatcher` interface (see Task 2 notes). This is a minor deviation from the implementation plan but enables proper unit testing.

- **`uuid.New()` in workflow trigger is fine** - Unlike workflow code (Phase 5), trigger code runs outside Temporal's deterministic sandbox. It's OK to use `uuid.New()`, `time.Now()`, etc.

- **Don't import `workflow.Engine`** - The trigger replaces the engine's entry point, it doesn't wrap it. The trigger should only depend on `TriggerProcessor` (or `RuleMatcher`) and `EdgeStore`, not the full engine.

- **Temporal `ExecuteWorkflow` is fire-and-forget** - `client.ExecuteWorkflow` returns a `WorkflowRun` handle but doesn't wait for completion. The trigger dispatches workflows and moves on. The workflow runs independently in the worker.

- **Workflow ID uniqueness** - Temporal rejects duplicate workflow IDs. The executionID (UUID) ensures uniqueness, but be aware that re-processing the same event (e.g., RabbitMQ retry) with the same executionID would fail. Consider whether the trigger should use event-derived IDs for idempotency.

### Tips

- Start with Task 1 (`trigger.go`) - it's straightforward once the interface is defined
- The `buildTriggerData` function is a pure function - test it first (no mocks needed)
- When writing tests, use `go.temporal.io/sdk/mocks` for the Temporal client mock
- The `RuleMatcher` interface pattern is a standard Go testing practice - keeps production code clean while enabling test isolation
- Check if `go.temporal.io/sdk/mocks` is already in vendor from Phase 1; if not, you'll need to `go mod vendor` after adding the test dependency

### Relationship to Existing System

The trigger system doesn't immediately replace the existing `Engine.ExecuteWorkflow()`. Both can coexist during migration:

```
Existing Path (still works):
    QueueManager → Engine.ExecuteWorkflow() → in-process execution

New Path (this phase):
    QueueManager → WorkflowTrigger.OnEntityEvent() → Temporal workflow

Phase 9 wiring switches the QueueManager to call WorkflowTrigger instead of Engine.
```

This allows incremental migration: individual rules can be migrated to Temporal while others continue using the existing engine.

---

## Testing Strategy

### Unit Tests (This Phase)

The unit tests use mock implementations for all external dependencies:

1. **Mock `RuleMatcher`** (interface extracted from `TriggerProcessor`) - returns configurable matched rules
2. **Mock `EdgeStore`** - returns configurable actions/edges per rule
3. **Mock Temporal Client** (`go.temporal.io/sdk/mocks.Client`) - captures workflow start calls

Test scenarios:
- Successful single-rule dispatch
- Multiple matched rules
- No matched rules (no workflows started)
- Empty graph (workflow skipped)
- EdgeStore error (rule skipped, others continue)
- Temporal client error (rule skipped, others continue)
- TriggerProcessor error (error returned immediately)
- buildTriggerData with all event types (create, update, delete)
- Workflow ID format verification

### Integration Tests (Deferred to Phase 11)

Full end-to-end trigger tests with real Temporal test container and seeded database are deferred to Phase 11 (Workflow Integration Tests).

---

## Commands Reference

```bash
# Start this phase
/workflow-temporal-next

# Validate this phase
/workflow-temporal-validate 7

# Review plan before implementing
/workflow-temporal-plan-review 7

# Review code after implementing
/workflow-temporal-review 7
```
