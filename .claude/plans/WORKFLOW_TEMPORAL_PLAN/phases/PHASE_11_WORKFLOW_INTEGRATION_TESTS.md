# Phase 11: Workflow Integration Tests

**Category**: Testing
**Status**: Pending
**Dependencies**: Phase 1 (test container infra), Phase 3 (models), Phase 4 (graph executor), Phase 5 (workflow.go), Phase 6 (activities)

---

## Overview

End-to-end workflow execution tests with the Temporal test framework. Phase 10 tested the graph executor in isolation (pure struct/method tests). Phase 11 executes **real workflows against a real Temporal server** via the Docker test container from Phase 1. Tests register `ExecuteGraphWorkflow`, `ExecuteBranchUntilConvergence`, and `Activities` on a worker, then dispatch workflows via `client.ExecuteWorkflow()` with various graph shapes and validate outcomes.

### Two Test Approaches Available

| Approach | Infrastructure | Best For |
|----------|---------------|----------|
| **SDK Test Suite** (`testsuite.WorkflowTestSuite`) | No server needed, mocked time/activities | Unit-level workflow logic, fast |
| **Real Temporal Container** (`temporal.GetTestContainer`) | Docker container, full gRPC stack | True end-to-end integration, replay |

Phase 11 uses **both**: SDK test suite for unit-level workflow tests (fast, isolated) and real containers for integration/replay tests (thorough, realistic).

## Goals

1. **Test full workflow execution** via real Temporal container — single action, sequential chains, parallel branches with convergence, fire-and-forget branches
2. **Verify Temporal replay safety** with recorded-then-replayed workflow histories — confirm deterministic command sequences
3. **Validate activity dispatch, handler integration, and MergedContext propagation** — verify results flow through graph steps, condition branches resolve correctly, parallel results merge at convergence

## Prerequisites

- Phase 1 complete — `foundation/temporal/temporal.go` test container infrastructure
- Phase 3 complete — `WorkflowInput`, `GraphDefinition`, `MergedContext` models
- Phase 4 complete — `GraphExecutor` for graph traversal
- Phase 5 complete — `ExecuteGraphWorkflow`, `ExecuteBranchUntilConvergence` workflow functions
- Phase 6 complete — `Activities` struct with `ExecuteActionActivity`/`ExecuteAsyncActionActivity`
- Docker available for test containers

### Key Signatures from Implementation (reference for test code)

```go
// models.go - WorkflowInput
type WorkflowInput struct {
    RuleID            uuid.UUID       `json:"rule_id"`
    RuleName          string          `json:"rule_name"`
    ExecutionID       uuid.UUID       `json:"execution_id"`
    Graph             GraphDefinition `json:"graph"`
    TriggerData       map[string]any  `json:"trigger_data"`
    ContinuationState *MergedContext  `json:"continuation_state,omitempty"`
}

// models.go - ActionNode fields
type ActionNode struct {
    ID            uuid.UUID       `json:"id"`
    Name          string          `json:"name"`
    Description   string          `json:"description"`
    ActionType    string          `json:"action_type"`
    Config        json.RawMessage `json:"action_config"`
    IsActive      bool            `json:"is_active"`
    DeactivatedBy uuid.UUID       `json:"deactivated_by"`
}

// models.go - ActionEdge (SourceActionID is *uuid.UUID, nil for start edges)
type ActionEdge struct {
    ID             uuid.UUID  `json:"id"`
    SourceActionID *uuid.UUID `json:"source_action_id"`
    TargetActionID uuid.UUID  `json:"target_action_id"`
    EdgeType       string     `json:"edge_type"`
    SortOrder      int        `json:"sort_order"`
}

// workflow.go - Workflow functions
func ExecuteGraphWorkflow(ctx workflow.Context, input WorkflowInput) error
func ExecuteBranchUntilConvergence(ctx workflow.Context, input BranchInput) (BranchOutput, error)

// activities.go - Activities struct
type Activities struct {
    Registry      *workflow.ActionRegistry
    AsyncRegistry *AsyncRegistry
}

// activities_async.go - NewAsyncRegistry
func NewAsyncRegistry() *AsyncRegistry

// foundation/temporal - Test container
func GetTestContainer(t *testing.T) Container   // Container has HostPort string field
func NewTestClient(hostPort string) (client.Client, error)

// workflow interfaces.go - ActionHandler (6 methods)
// Execute, Validate, GetType, SupportsManualExecution, IsAsync, GetDescription
// ActionRegistry: Register(handler ActionHandler), Get(actionType string) (ActionHandler, bool)
```

---

## Test File Structure

```
business/sdk/workflow/temporal/
├── workflow_test.go                 # SDK test suite: unit-level workflow tests + shared helpers
├── workflow_parallel_test.go        # SDK test suite: parallel/convergence tests
├── workflow_replay_test.go          # Real container: integration + replay determinism tests
└── ... (existing files unchanged)
```

### Test Coverage Goals

| Area | Target | Validation Criteria | Approach |
|------|--------|---------------------|----------|
| Single action execution | Activity called, result in context | `handler.called == 1`, workflow completes without error | SDK test suite |
| Sequential chain (3+ actions) | Results propagate, MergedContext grows | All handlers called, each successive handler's context includes prior results | SDK test suite |
| Condition branching | true_branch/false_branch dispatched correctly | Only the correct branch handler called; other branch handler `called == 0` | SDK test suite |
| Parallel with convergence | Child workflows, results merged at convergence | Both branch results present in MergedContext at convergence point action | SDK test suite |
| Fire-and-forget | Child workflows, parent completes independently | Parent workflow returns nil error; child workflows launched (verify via handler.called) | SDK test suite |
| Error propagation | Activity error fails workflow | `env.GetWorkflowError()` returns non-nil; error message contains action name | SDK test suite |
| End-to-end with real server | Full gRPC round-trip, actual worker execution | `run.Get(ctx, nil)` returns no error; handler.called verified | Real container |
| Replay determinism | Same history replayed without non-determinism errors | `replayer.ReplayWorkflowHistory()` returns nil error | Real container |

---

## Task Breakdown

### Task 1: Set Up Temporal Test Environment

**Status**: Pending

**Description**: Create test helpers and fixture builders that both `workflow_test.go` and `workflow_parallel_test.go` share. This includes graph builder helpers, mock action handlers, and the SDK test suite setup. All helpers are defined in `workflow_test.go` and automatically visible to `workflow_parallel_test.go` (same package).

**Files**:
- `business/sdk/workflow/temporal/workflow_test.go` (test helpers + basic tests in same file)

**Implementation Guide**:

```go
package temporal

import (
    "context"
    "encoding/json"
    "fmt"
    "testing"

    "github.com/google/uuid"
    "github.com/stretchr/testify/require"
    "github.com/timmaaaz/ichor/business/sdk/workflow"
    "go.temporal.io/sdk/testsuite"
)

// =============================================================================
// Test Action Handler (implements all 6 ActionHandler methods)
// =============================================================================

// testActionHandler is a configurable mock action handler for testing.
// Tracks call count so tests can verify which handlers were invoked.
type testActionHandler struct {
    actionType string
    result     any
    err        error
    called     int
}

func (h *testActionHandler) Execute(ctx context.Context, config json.RawMessage, execCtx workflow.ActionExecutionContext) (any, error) {
    h.called++
    if h.err != nil {
        return nil, h.err
    }
    return h.result, nil
}

func (h *testActionHandler) Validate(config json.RawMessage) error { return nil }
func (h *testActionHandler) GetType() string                      { return h.actionType }
func (h *testActionHandler) SupportsManualExecution() bool        { return false }
func (h *testActionHandler) IsAsync() bool                        { return false }
func (h *testActionHandler) GetDescription() string               { return "test handler" }

// newTestRegistry creates an ActionRegistry pre-populated with test handlers.
// ActionRegistry.Register uses handler.GetType() as the key.
func newTestRegistry(handlers ...*testActionHandler) *workflow.ActionRegistry {
    reg := workflow.NewActionRegistry()
    for _, h := range handlers {
        reg.Register(h)
    }
    return reg
}

// =============================================================================
// SDK Test Suite Setup
// =============================================================================

// setupTestEnv creates a Temporal test environment with registered workflows/activities.
// AsyncRegistry is initialized empty — async actions will fail gracefully with
// "no handler registered" error. Async completion is tested in Phase 12.
func setupTestEnv(t *testing.T, registry *workflow.ActionRegistry) *testsuite.TestWorkflowEnvironment {
    t.Helper()
    suite := &testsuite.WorkflowTestSuite{}
    env := suite.NewTestWorkflowEnvironment()

    // Register both workflow functions.
    env.RegisterWorkflow(ExecuteGraphWorkflow)
    env.RegisterWorkflow(ExecuteBranchUntilConvergence)

    // Register activities via Activities struct (pointer).
    // Temporal resolves struct method names by string ("ExecuteActionActivity").
    activities := &Activities{
        Registry:      registry,
        AsyncRegistry: NewAsyncRegistry(), // Empty — no async handlers for unit tests
    }
    env.RegisterActivity(activities)

    return env
}

// =============================================================================
// Graph Builder Helpers
// =============================================================================

// buildLinearGraph creates a linear workflow: start -> action[0] -> action[1] -> ... -> action[N-1]
// Each action gets a deterministic UUID based on its index for test reproducibility.
// Returns the graph definition and the ordered list of action IDs.
//
// Edge structure:
//   - 1 start edge (source=nil, target=action[0], type=start)
//   - N-1 sequence edges (source=action[i-1], target=action[i], type=sequence)
func buildLinearGraph(actionTypes ...string) (GraphDefinition, []uuid.UUID) {
    if len(actionTypes) == 0 {
        return GraphDefinition{}, nil
    }

    actions := make([]ActionNode, len(actionTypes))
    ids := make([]uuid.UUID, len(actionTypes))

    for i, at := range actionTypes {
        ids[i] = uuid.New()
        actions[i] = ActionNode{
            ID:         ids[i],
            Name:       fmt.Sprintf("action_%d", i),
            ActionType: at,
            Config:     json.RawMessage(`{}`),
            IsActive:   true,
        }
    }

    // Start edge: nil source -> first action
    edges := make([]ActionEdge, 0, len(actionTypes))
    edges = append(edges, ActionEdge{
        ID:             uuid.New(),
        SourceActionID: nil, // Start edge
        TargetActionID: ids[0],
        EdgeType:       EdgeTypeStart,
        SortOrder:      1,
    })

    // Sequence edges: action[i-1] -> action[i]
    for i := 1; i < len(actionTypes); i++ {
        src := ids[i-1]
        edges = append(edges, ActionEdge{
            ID:             uuid.New(),
            SourceActionID: &src,
            TargetActionID: ids[i],
            EdgeType:       EdgeTypeSequence,
            SortOrder:      1,
        })
    }

    return GraphDefinition{Actions: actions, Edges: edges}, ids
}

// diamondIDs holds action IDs for a diamond-shaped condition graph.
type diamondIDs struct {
    Condition uuid.UUID
    TrueArm   uuid.UUID
    FalseArm  uuid.UUID
    Merge     uuid.UUID
}

// buildDiamondGraph creates a condition diamond:
//
//   start -> condition --(true_branch)--> trueArm --(sequence)--> merge
//                      \--(false_branch)-> falseArm --(sequence)--> merge
//
// The condition action should use "evaluate_condition" type.
// trueArm and falseArm use the provided action types.
// merge uses "merge_action" type.
func buildDiamondGraph(trueArmType, falseArmType string) (GraphDefinition, diamondIDs) {
    condID := uuid.New()
    trueID := uuid.New()
    falseID := uuid.New()
    mergeID := uuid.New()

    actions := []ActionNode{
        {ID: condID, Name: "condition", ActionType: "evaluate_condition", Config: json.RawMessage(`{}`), IsActive: true},
        {ID: trueID, Name: "true_arm", ActionType: trueArmType, Config: json.RawMessage(`{}`), IsActive: true},
        {ID: falseID, Name: "false_arm", ActionType: falseArmType, Config: json.RawMessage(`{}`), IsActive: true},
        {ID: mergeID, Name: "merge", ActionType: "merge_action", Config: json.RawMessage(`{}`), IsActive: true},
    }

    edges := []ActionEdge{
        // Start -> condition
        {ID: uuid.New(), SourceActionID: nil, TargetActionID: condID, EdgeType: EdgeTypeStart, SortOrder: 1},
        // Condition -> true_arm (true_branch)
        {ID: uuid.New(), SourceActionID: &condID, TargetActionID: trueID, EdgeType: EdgeTypeTrueBranch, SortOrder: 1},
        // Condition -> false_arm (false_branch)
        {ID: uuid.New(), SourceActionID: &condID, TargetActionID: falseID, EdgeType: EdgeTypeFalseBranch, SortOrder: 2},
        // true_arm -> merge (sequence)
        {ID: uuid.New(), SourceActionID: &trueID, TargetActionID: mergeID, EdgeType: EdgeTypeSequence, SortOrder: 1},
        // false_arm -> merge (sequence)
        {ID: uuid.New(), SourceActionID: &falseID, TargetActionID: mergeID, EdgeType: EdgeTypeSequence, SortOrder: 1},
    }

    return GraphDefinition{Actions: actions, Edges: edges}, diamondIDs{
        Condition: condID,
        TrueArm:   trueID,
        FalseArm:  falseID,
        Merge:     mergeID,
    }
}

// parallelIDs holds action IDs for a parallel convergence graph.
type parallelIDs struct {
    Fork    uuid.UUID
    BranchA uuid.UUID
    BranchB uuid.UUID
    Merge   uuid.UUID
}

// buildParallelGraph creates a parallel fork/join:
//
//   start -> fork --(sequence)--> branchA --(sequence)--> merge
//                 \--(sequence)--> branchB --(sequence)--> merge
//
// fork action executes first, then branchA and branchB run as parallel child
// workflows (detected by GraphExecutor because fork has 2 outgoing sequence edges).
// Both converge at merge (detected by FindConvergencePoint).
func buildParallelGraph(branchAType, branchBType string) (GraphDefinition, parallelIDs) {
    forkID := uuid.New()
    branchAID := uuid.New()
    branchBID := uuid.New()
    mergeID := uuid.New()

    actions := []ActionNode{
        {ID: forkID, Name: "fork", ActionType: "fork_action", Config: json.RawMessage(`{}`), IsActive: true},
        {ID: branchAID, Name: "branch_a", ActionType: branchAType, Config: json.RawMessage(`{}`), IsActive: true},
        {ID: branchBID, Name: "branch_b", ActionType: branchBType, Config: json.RawMessage(`{}`), IsActive: true},
        {ID: mergeID, Name: "merge", ActionType: "merge_action", Config: json.RawMessage(`{}`), IsActive: true},
    }

    edges := []ActionEdge{
        // Start -> fork
        {ID: uuid.New(), SourceActionID: nil, TargetActionID: forkID, EdgeType: EdgeTypeStart, SortOrder: 1},
        // Fork -> branchA (sequence) — two outgoing sequence edges trigger parallel execution
        {ID: uuid.New(), SourceActionID: &forkID, TargetActionID: branchAID, EdgeType: EdgeTypeSequence, SortOrder: 1},
        // Fork -> branchB (sequence)
        {ID: uuid.New(), SourceActionID: &forkID, TargetActionID: branchBID, EdgeType: EdgeTypeSequence, SortOrder: 2},
        // BranchA -> merge (sequence) — both branches converge here
        {ID: uuid.New(), SourceActionID: &branchAID, TargetActionID: mergeID, EdgeType: EdgeTypeSequence, SortOrder: 1},
        // BranchB -> merge (sequence)
        {ID: uuid.New(), SourceActionID: &branchBID, TargetActionID: mergeID, EdgeType: EdgeTypeSequence, SortOrder: 1},
    }

    return GraphDefinition{Actions: actions, Edges: edges}, parallelIDs{
        Fork:    forkID,
        BranchA: branchAID,
        BranchB: branchBID,
        Merge:   mergeID,
    }
}

// fireForgetIDs holds action IDs for a fire-and-forget graph.
type fireForgetIDs struct {
    Fork    uuid.UUID
    BranchA uuid.UUID
    BranchB uuid.UUID
}

// buildFireForgetGraph creates parallel branches with NO convergence point:
//
//   start -> fork --(sequence)--> branchA
//                 \--(sequence)--> branchB
//
// No merge action. FindConvergencePoint returns nil, triggering executeFireAndForget.
func buildFireForgetGraph(branchAType, branchBType string) (GraphDefinition, fireForgetIDs) {
    forkID := uuid.New()
    branchAID := uuid.New()
    branchBID := uuid.New()

    actions := []ActionNode{
        {ID: forkID, Name: "fork", ActionType: "fork_action", Config: json.RawMessage(`{}`), IsActive: true},
        {ID: branchAID, Name: "branch_a", ActionType: branchAType, Config: json.RawMessage(`{}`), IsActive: true},
        {ID: branchBID, Name: "branch_b", ActionType: branchBType, Config: json.RawMessage(`{}`), IsActive: true},
    }

    edges := []ActionEdge{
        {ID: uuid.New(), SourceActionID: nil, TargetActionID: forkID, EdgeType: EdgeTypeStart, SortOrder: 1},
        {ID: uuid.New(), SourceActionID: &forkID, TargetActionID: branchAID, EdgeType: EdgeTypeSequence, SortOrder: 1},
        {ID: uuid.New(), SourceActionID: &forkID, TargetActionID: branchBID, EdgeType: EdgeTypeSequence, SortOrder: 2},
    }

    return GraphDefinition{Actions: actions, Edges: edges}, fireForgetIDs{
        Fork:    forkID,
        BranchA: branchAID,
        BranchB: branchBID,
    }
}
```

**Notes**:
- `testsuite.WorkflowTestSuite` doesn't need Docker — runs fully in-process
- Mock handlers allow configurable results/errors per action type
- Graph builder helpers prevent test boilerplate and make graph shapes clear
- `setupTestEnv` centralizes workflow/activity registration
- All helpers in `workflow_test.go` are visible to `workflow_parallel_test.go` (same package `temporal`)
- `ActionEdge.SourceActionID` is `*uuid.UUID` — must take address of local variable for non-start edges (e.g., `&forkID`)

---

### Task 2: Simple Workflow Execution Tests

**Status**: Pending

**Description**: Test core workflow execution patterns: single action, sequential chain, condition branching, MergedContext propagation, validation errors, deactivated actions, unknown action types, and error propagation.

**Files**:
- `business/sdk/workflow/temporal/workflow_test.go`

**Implementation Guide**:

```go
// =============================================================================
// Single Action Tests
// =============================================================================

func TestWorkflow_SingleAction(t *testing.T) {
    handler := &testActionHandler{
        actionType: "test_action",
        result:     map[string]any{"status": "done"},
    }
    env := setupTestEnv(t, newTestRegistry(handler))

    graph, _ := buildLinearGraph("test_action")
    input := WorkflowInput{
        RuleID:      uuid.New(),
        RuleName:    "test-rule",
        ExecutionID: uuid.New(),
        Graph:       graph,
        TriggerData: map[string]any{"entity_id": "123"},
    }

    env.ExecuteWorkflow(ExecuteGraphWorkflow, input)
    require.True(t, env.IsWorkflowCompleted())
    require.NoError(t, env.GetWorkflowError())
    require.Equal(t, 1, handler.called)
}

// =============================================================================
// Sequential Chain Tests
// =============================================================================

func TestWorkflow_SequentialChain(t *testing.T) {
    // 3 actions in sequence, each returns a unique result.
    handler0 := &testActionHandler{actionType: "step_0", result: map[string]any{"val": "a"}}
    handler1 := &testActionHandler{actionType: "step_1", result: map[string]any{"val": "b"}}
    handler2 := &testActionHandler{actionType: "step_2", result: map[string]any{"val": "c"}}

    env := setupTestEnv(t, newTestRegistry(handler0, handler1, handler2))

    graph, _ := buildLinearGraph("step_0", "step_1", "step_2")
    input := WorkflowInput{
        RuleID:      uuid.New(),
        RuleName:    "chain-test",
        ExecutionID: uuid.New(),
        Graph:       graph,
        TriggerData: map[string]any{"entity_id": "456"},
    }

    env.ExecuteWorkflow(ExecuteGraphWorkflow, input)
    require.True(t, env.IsWorkflowCompleted())
    require.NoError(t, env.GetWorkflowError())

    // Verify all 3 handlers were called exactly once.
    require.Equal(t, 1, handler0.called)
    require.Equal(t, 1, handler1.called)
    require.Equal(t, 1, handler2.called)
}

// =============================================================================
// Condition Branching Tests
// =============================================================================

func TestWorkflow_ConditionBranch_True(t *testing.T) {
    // Condition handler returns branch_taken = true_branch.
    condHandler := &testActionHandler{
        actionType: "evaluate_condition",
        result:     map[string]any{"branch_taken": "true_branch"},
    }
    trueHandler := &testActionHandler{actionType: "true_action", result: map[string]any{"path": "true"}}
    falseHandler := &testActionHandler{actionType: "false_action", result: map[string]any{"path": "false"}}
    mergeHandler := &testActionHandler{actionType: "merge_action", result: map[string]any{"merged": true}}

    env := setupTestEnv(t, newTestRegistry(condHandler, trueHandler, falseHandler, mergeHandler))

    graph, _ := buildDiamondGraph("true_action", "false_action")
    input := WorkflowInput{
        RuleID:      uuid.New(),
        RuleName:    "cond-true-test",
        ExecutionID: uuid.New(),
        Graph:       graph,
        TriggerData: map[string]any{},
    }

    env.ExecuteWorkflow(ExecuteGraphWorkflow, input)
    require.True(t, env.IsWorkflowCompleted())
    require.NoError(t, env.GetWorkflowError())

    // True branch taken: condition + trueArm + merge called. falseArm NOT called.
    require.Equal(t, 1, condHandler.called)
    require.Equal(t, 1, trueHandler.called)
    require.Equal(t, 0, falseHandler.called, "false_branch handler should NOT be called")
    require.Equal(t, 1, mergeHandler.called)
}

func TestWorkflow_ConditionBranch_False(t *testing.T) {
    // Condition handler returns branch_taken = false_branch.
    condHandler := &testActionHandler{
        actionType: "evaluate_condition",
        result:     map[string]any{"branch_taken": "false_branch"},
    }
    trueHandler := &testActionHandler{actionType: "true_action", result: map[string]any{"path": "true"}}
    falseHandler := &testActionHandler{actionType: "false_action", result: map[string]any{"path": "false"}}
    mergeHandler := &testActionHandler{actionType: "merge_action", result: map[string]any{"merged": true}}

    env := setupTestEnv(t, newTestRegistry(condHandler, trueHandler, falseHandler, mergeHandler))

    graph, _ := buildDiamondGraph("true_action", "false_action")
    input := WorkflowInput{
        RuleID:      uuid.New(),
        RuleName:    "cond-false-test",
        ExecutionID: uuid.New(),
        Graph:       graph,
        TriggerData: map[string]any{},
    }

    env.ExecuteWorkflow(ExecuteGraphWorkflow, input)
    require.True(t, env.IsWorkflowCompleted())
    require.NoError(t, env.GetWorkflowError())

    // False branch taken: condition + falseArm + merge called. trueArm NOT called.
    require.Equal(t, 1, condHandler.called)
    require.Equal(t, 0, trueHandler.called, "true_branch handler should NOT be called")
    require.Equal(t, 1, falseHandler.called)
    require.Equal(t, 1, mergeHandler.called)
}

// =============================================================================
// Validation Error Tests
// =============================================================================

func TestWorkflow_InvalidInput_NoRuleID(t *testing.T) {
    env := setupTestEnv(t, newTestRegistry())

    // Missing RuleID (uuid.Nil) — Validate() returns "rule_id is required"
    input := WorkflowInput{
        ExecutionID: uuid.New(),
        Graph:       GraphDefinition{Actions: []ActionNode{{ID: uuid.New()}}, Edges: []ActionEdge{{EdgeType: EdgeTypeStart}}},
    }
    env.ExecuteWorkflow(ExecuteGraphWorkflow, input)
    require.True(t, env.IsWorkflowCompleted())
    require.Error(t, env.GetWorkflowError())
    require.Contains(t, env.GetWorkflowError().Error(), "rule_id is required")
}

func TestWorkflow_InvalidInput_NoStartEdges(t *testing.T) {
    env := setupTestEnv(t, newTestRegistry())

    // Graph with actions and a sequence edge but no start edge.
    actionID := uuid.New()
    input := WorkflowInput{
        RuleID:      uuid.New(),
        ExecutionID: uuid.New(),
        Graph: GraphDefinition{
            Actions: []ActionNode{{ID: actionID, Name: "orphan", ActionType: "test", Config: json.RawMessage(`{}`), IsActive: true}},
            Edges:   []ActionEdge{{ID: uuid.New(), SourceActionID: &actionID, TargetActionID: uuid.New(), EdgeType: EdgeTypeSequence}},
        },
    }
    env.ExecuteWorkflow(ExecuteGraphWorkflow, input)
    require.True(t, env.IsWorkflowCompleted())
    require.Error(t, env.GetWorkflowError())
    require.Contains(t, env.GetWorkflowError().Error(), "start edge")
}

func TestWorkflow_InvalidInput_EmptyGraph(t *testing.T) {
    env := setupTestEnv(t, newTestRegistry())

    // Graph with no actions — Validate() returns "graph must contain at least one action"
    input := WorkflowInput{
        RuleID:      uuid.New(),
        ExecutionID: uuid.New(),
        Graph:       GraphDefinition{},
    }
    env.ExecuteWorkflow(ExecuteGraphWorkflow, input)
    require.True(t, env.IsWorkflowCompleted())
    require.Error(t, env.GetWorkflowError())
    require.Contains(t, env.GetWorkflowError().Error(), "at least one action")
}

// =============================================================================
// Error Propagation Tests
// =============================================================================

func TestWorkflow_ActionError_FailsWorkflow(t *testing.T) {
    // Activity returns error -> workflow should fail with wrapped error containing action name.
    failHandler := &testActionHandler{
        actionType: "fail_action",
        err:        fmt.Errorf("simulated failure"),
    }
    env := setupTestEnv(t, newTestRegistry(failHandler))

    graph, _ := buildLinearGraph("fail_action")
    input := WorkflowInput{
        RuleID:      uuid.New(),
        RuleName:    "error-test",
        ExecutionID: uuid.New(),
        Graph:       graph,
        TriggerData: map[string]any{},
    }

    env.ExecuteWorkflow(ExecuteGraphWorkflow, input)
    require.True(t, env.IsWorkflowCompleted())
    require.Error(t, env.GetWorkflowError())
    // Error should contain the action name from the wrapping in activities.go
    require.Contains(t, env.GetWorkflowError().Error(), "action_0")
}

func TestWorkflow_UnknownActionType(t *testing.T) {
    // Registry has no handler for "unknown_type" — activity returns "no handler registered" error.
    env := setupTestEnv(t, newTestRegistry()) // Empty registry

    graph, _ := buildLinearGraph("unknown_type")
    input := WorkflowInput{
        RuleID:      uuid.New(),
        RuleName:    "unknown-test",
        ExecutionID: uuid.New(),
        Graph:       graph,
        TriggerData: map[string]any{},
    }

    env.ExecuteWorkflow(ExecuteGraphWorkflow, input)
    require.True(t, env.IsWorkflowCompleted())
    require.Error(t, env.GetWorkflowError())
    require.Contains(t, env.GetWorkflowError().Error(), "no handler registered")
}

// =============================================================================
// Deactivated Action Tests
// =============================================================================

// NOTE: Current implementation does NOT skip deactivated actions — the activity
// dispatches regardless of IsActive. If skip-deactivated is a future requirement,
// this test documents the expected behavior. For now, verify the workflow still
// executes the action (handler.called == 1). Update this test when skip logic
// is added to executeSingleAction or ExecuteActionActivity.
func TestWorkflow_DeactivatedAction_StillExecutes(t *testing.T) {
    handler := &testActionHandler{
        actionType: "test_action",
        result:     map[string]any{"status": "done"},
    }
    env := setupTestEnv(t, newTestRegistry(handler))

    // Build graph manually with IsActive=false on the action.
    actionID := uuid.New()
    graph := GraphDefinition{
        Actions: []ActionNode{
            {ID: actionID, Name: "deactivated", ActionType: "test_action", Config: json.RawMessage(`{}`), IsActive: false, DeactivatedBy: uuid.New()},
        },
        Edges: []ActionEdge{
            {ID: uuid.New(), SourceActionID: nil, TargetActionID: actionID, EdgeType: EdgeTypeStart, SortOrder: 1},
        },
    }
    input := WorkflowInput{
        RuleID:      uuid.New(),
        RuleName:    "deactivated-test",
        ExecutionID: uuid.New(),
        Graph:       graph,
        TriggerData: map[string]any{},
    }

    env.ExecuteWorkflow(ExecuteGraphWorkflow, input)
    require.True(t, env.IsWorkflowCompleted())
    // Current behavior: activity still executes for deactivated actions.
    // Future: may change to skip. Update test accordingly.
    require.NoError(t, env.GetWorkflowError())
    require.Equal(t, 1, handler.called)
}
```

**Notes**:
- SDK test suite runs synchronously — `env.ExecuteWorkflow()` blocks until complete
- `env.GetWorkflowError()` returns the workflow's return error
- Test both happy paths and error paths
- `require.Contains` validates error wrapping from activities.go includes action name
- Deactivated action test documents current behavior (no skip logic) — update when skip is added

---

### Task 3: Parallel Execution Tests

**Status**: Pending

**Description**: Test parallel branch execution, convergence, fire-and-forget, and result merging at convergence points. Also directly test `ExecuteBranchUntilConvergence` child workflow.

**Files**:
- `business/sdk/workflow/temporal/workflow_parallel_test.go`

**Implementation Guide**:

```go
package temporal

import (
    "testing"

    "github.com/google/uuid"
    "github.com/stretchr/testify/require"
)

// =============================================================================
// Parallel with Convergence
// =============================================================================

func TestWorkflow_ParallelBranches_Convergence(t *testing.T) {
    // Graph: start -> fork -> branchA, branchB -> merge
    // fork executes first (single start action), then its two outgoing sequence
    // edges trigger parallel execution. Both branches converge at merge.
    forkHandler := &testActionHandler{actionType: "fork_action", result: map[string]any{"forked": true}}
    branchAHandler := &testActionHandler{actionType: "branch_a_type", result: map[string]any{"price": 100}}
    branchBHandler := &testActionHandler{actionType: "branch_b_type", result: map[string]any{"quantity": 5}}
    mergeHandler := &testActionHandler{actionType: "merge_action", result: map[string]any{"merged": true}}

    env := setupTestEnv(t, newTestRegistry(forkHandler, branchAHandler, branchBHandler, mergeHandler))

    graph, _ := buildParallelGraph("branch_a_type", "branch_b_type")
    input := WorkflowInput{
        RuleID:      uuid.New(),
        RuleName:    "parallel-test",
        ExecutionID: uuid.New(),
        Graph:       graph,
        TriggerData: map[string]any{},
    }

    env.ExecuteWorkflow(ExecuteGraphWorkflow, input)
    require.True(t, env.IsWorkflowCompleted())
    require.NoError(t, env.GetWorkflowError())

    // All 4 handlers should be called: fork + branchA + branchB + merge.
    require.Equal(t, 1, forkHandler.called)
    require.Equal(t, 1, branchAHandler.called)
    require.Equal(t, 1, branchBHandler.called)
    require.Equal(t, 1, mergeHandler.called, "merge action at convergence should be called")
}

// =============================================================================
// Fire-and-Forget
// =============================================================================

func TestWorkflow_FireAndForget(t *testing.T) {
    // Graph: start -> fork -> branchA, branchB (no convergence).
    // Parent workflow should complete immediately after launching branches.
    forkHandler := &testActionHandler{actionType: "fork_action", result: map[string]any{"forked": true}}
    branchAHandler := &testActionHandler{actionType: "branch_a_type", result: map[string]any{"a": 1}}
    branchBHandler := &testActionHandler{actionType: "branch_b_type", result: map[string]any{"b": 2}}

    env := setupTestEnv(t, newTestRegistry(forkHandler, branchAHandler, branchBHandler))

    graph, _ := buildFireForgetGraph("branch_a_type", "branch_b_type")
    input := WorkflowInput{
        RuleID:      uuid.New(),
        RuleName:    "fire-forget-test",
        ExecutionID: uuid.New(),
        Graph:       graph,
        TriggerData: map[string]any{},
    }

    env.ExecuteWorkflow(ExecuteGraphWorkflow, input)
    require.True(t, env.IsWorkflowCompleted())
    require.NoError(t, env.GetWorkflowError())

    // Fork handler called. Branch handlers are launched as child workflows
    // with PARENT_CLOSE_POLICY_ABANDON.
    // NOTE: In SDK test suite, child workflows execute synchronously, so both
    // branch handlers WILL be called. In production, they run independently.
    require.Equal(t, 1, forkHandler.called)
    // Branch handlers should be called (SDK test suite runs children synchronously)
    require.Equal(t, 1, branchAHandler.called)
    require.Equal(t, 1, branchBHandler.called)
}

// =============================================================================
// Direct Child Workflow Tests (ExecuteBranchUntilConvergence)
// =============================================================================

func TestWorkflow_BranchUntilConvergence_SingleBranch(t *testing.T) {
    // Direct test of ExecuteBranchUntilConvergence as a standalone workflow.
    // Linear chain of 3 actions, stop before convergence point.
    handler0 := &testActionHandler{actionType: "step_0", result: map[string]any{"v": 0}}
    handler1 := &testActionHandler{actionType: "step_1", result: map[string]any{"v": 1}}
    handler2 := &testActionHandler{actionType: "step_2", result: map[string]any{"v": 2}}

    env := setupTestEnv(t, newTestRegistry(handler0, handler1, handler2))

    // Build a linear graph: step_0 -> step_1 -> step_2
    graph, ids := buildLinearGraph("step_0", "step_1", "step_2")

    // Branch starts at step_0, converges at step_2 (so step_2 should NOT execute).
    branchInput := BranchInput{
        StartAction:      graph.Actions[0], // step_0
        ConvergencePoint: ids[2],           // step_2 is the convergence point
        Graph:            graph,
        InitialContext:   NewMergedContext(map[string]any{"trigger": "test"}),
        RuleID:           uuid.New(),
        ExecutionID:      uuid.New(),
        RuleName:         "branch-test",
    }

    env.ExecuteWorkflow(ExecuteBranchUntilConvergence, branchInput)
    require.True(t, env.IsWorkflowCompleted())
    require.NoError(t, env.GetWorkflowError())

    // step_0 and step_1 executed, step_2 (convergence) NOT executed.
    require.Equal(t, 1, handler0.called)
    require.Equal(t, 1, handler1.called)
    require.Equal(t, 0, handler2.called, "convergence point should NOT be executed by branch")
}

func TestWorkflow_BranchUntilConvergence_FireAndForget(t *testing.T) {
    // ConvergencePoint = uuid.Nil (fire-and-forget). All actions should execute.
    handler0 := &testActionHandler{actionType: "step_0", result: map[string]any{"v": 0}}
    handler1 := &testActionHandler{actionType: "step_1", result: map[string]any{"v": 1}}

    env := setupTestEnv(t, newTestRegistry(handler0, handler1))

    graph, _ := buildLinearGraph("step_0", "step_1")
    branchInput := BranchInput{
        StartAction:      graph.Actions[0],
        ConvergencePoint: uuid.Nil, // Fire-and-forget: run until end
        Graph:            graph,
        InitialContext:   NewMergedContext(map[string]any{}),
        RuleID:           uuid.New(),
        ExecutionID:      uuid.New(),
        RuleName:         "fire-forget-branch",
    }

    env.ExecuteWorkflow(ExecuteBranchUntilConvergence, branchInput)
    require.True(t, env.IsWorkflowCompleted())
    require.NoError(t, env.GetWorkflowError())

    // Both actions executed (no convergence to stop at).
    require.Equal(t, 1, handler0.called)
    require.Equal(t, 1, handler1.called)
}
```

**Notes**:
- SDK test suite handles child workflows automatically — `ExecuteBranchUntilConvergence` will be called as a real child workflow in the test environment
- For fire-and-forget, child workflows with `PARENT_CLOSE_POLICY_ABANDON` execute synchronously in SDK test suite (both branch handlers will be called), unlike production where they run independently. Tests verify the branch handlers are invoked.
- Result merge verification: check handler call counts to verify convergence behavior
- Test `ExecuteBranchUntilConvergence` both as child (from `ExecuteGraphWorkflow`) and directly
- If fire-and-forget tests are flaky in SDK test suite, move them to integration tests in `workflow_replay_test.go`

---

### Task 4: Replay Testing

**Status**: Pending

**Description**: Execute workflows against a real Temporal container, then replay the recorded histories to verify determinism. Non-determinism errors (NDE) during replay indicate code changes broke backward compatibility with in-flight workflows. Histories are generated dynamically per test run (no golden files in Phase 11).

**Files**:
- `business/sdk/workflow/temporal/workflow_replay_test.go`

**Implementation Guide**:

```go
package temporal

import (
    "context"
    "fmt"
    "testing"
    "time"

    "github.com/google/uuid"
    "github.com/stretchr/testify/require"
    temporalclient "go.temporal.io/sdk/client"
    "go.temporal.io/sdk/worker"

    "github.com/timmaaaz/ichor/business/sdk/workflow"
    foundationtemporal "github.com/timmaaaz/ichor/foundation/temporal"
)

// =============================================================================
// Integration Test Helpers
// =============================================================================

// startTestWorker creates a Temporal worker for integration tests.
// Uses a unique task queue per test to prevent cross-test interference.
func startTestWorker(t *testing.T, tc temporalclient.Client, taskQueue string, handlers ...*testActionHandler) worker.Worker {
    t.Helper()
    reg := newTestRegistry(handlers...)

    w := worker.New(tc, taskQueue, worker.Options{})
    w.RegisterWorkflow(ExecuteGraphWorkflow)
    w.RegisterWorkflow(ExecuteBranchUntilConvergence)
    w.RegisterActivity(&Activities{
        Registry:      reg,
        AsyncRegistry: NewAsyncRegistry(),
    })
    require.NoError(t, w.Start())
    return w
}

// =============================================================================
// Integration Tests (Real Temporal Container)
// =============================================================================

func TestIntegration_SimpleWorkflow_RealServer(t *testing.T) {
    if testing.Short() {
        t.Skip("skipping integration test in short mode")
    }

    // Get shared test container (singleton pattern from foundation/temporal).
    c := foundationtemporal.GetTestContainer(t)
    tc, err := foundationtemporal.NewTestClient(c.HostPort)
    require.NoError(t, err)
    defer tc.Close()

    // Create and start worker with unique task queue.
    handler := &testActionHandler{
        actionType: "test_action",
        result:     map[string]any{"status": "done"},
    }
    taskQueue := fmt.Sprintf("test-integration-%s", t.Name())
    w := startTestWorker(t, tc, taskQueue, handler)
    defer w.Stop()

    // Build and execute workflow.
    graph, _ := buildLinearGraph("test_action")
    input := WorkflowInput{
        RuleID:      uuid.New(),
        RuleName:    "integration-test-rule",
        ExecutionID: uuid.New(),
        Graph:       graph,
        TriggerData: map[string]any{"entity_id": "abc"},
    }

    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    run, err := tc.ExecuteWorkflow(ctx, temporalclient.StartWorkflowOptions{
        TaskQueue: taskQueue,
    }, ExecuteGraphWorkflow, input)
    require.NoError(t, err)

    // Wait for completion — full gRPC round-trip.
    require.NoError(t, run.Get(ctx, nil))

    // Verify handler was called.
    require.Equal(t, 1, handler.called)
}

func TestIntegration_SequentialChain_RealServer(t *testing.T) {
    if testing.Short() {
        t.Skip("skipping integration test in short mode")
    }

    c := foundationtemporal.GetTestContainer(t)
    tc, err := foundationtemporal.NewTestClient(c.HostPort)
    require.NoError(t, err)
    defer tc.Close()

    handler0 := &testActionHandler{actionType: "step_0", result: map[string]any{"val": "a"}}
    handler1 := &testActionHandler{actionType: "step_1", result: map[string]any{"val": "b"}}
    handler2 := &testActionHandler{actionType: "step_2", result: map[string]any{"val": "c"}}

    taskQueue := fmt.Sprintf("test-integration-%s", t.Name())
    w := startTestWorker(t, tc, taskQueue, handler0, handler1, handler2)
    defer w.Stop()

    graph, _ := buildLinearGraph("step_0", "step_1", "step_2")
    input := WorkflowInput{
        RuleID:      uuid.New(),
        RuleName:    "chain-integration",
        ExecutionID: uuid.New(),
        Graph:       graph,
        TriggerData: map[string]any{},
    }

    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    run, err := tc.ExecuteWorkflow(ctx, temporalclient.StartWorkflowOptions{
        TaskQueue: taskQueue,
    }, ExecuteGraphWorkflow, input)
    require.NoError(t, err)
    require.NoError(t, run.Get(ctx, nil))

    require.Equal(t, 1, handler0.called)
    require.Equal(t, 1, handler1.called)
    require.Equal(t, 1, handler2.called)
}

// =============================================================================
// Replay Determinism Tests
// =============================================================================

func TestReplay_SimpleSequential(t *testing.T) {
    if testing.Short() {
        t.Skip("skipping integration test in short mode")
    }

    c := foundationtemporal.GetTestContainer(t)
    tc, err := foundationtemporal.NewTestClient(c.HostPort)
    require.NoError(t, err)
    defer tc.Close()

    // Step 1: Execute workflow to generate history.
    handler := &testActionHandler{
        actionType: "test_action",
        result:     map[string]any{"status": "replayed"},
    }
    taskQueue := fmt.Sprintf("test-replay-%s", t.Name())
    w := startTestWorker(t, tc, taskQueue, handler)
    defer w.Stop()

    graph, _ := buildLinearGraph("test_action")
    input := WorkflowInput{
        RuleID:      uuid.New(),
        RuleName:    "replay-test",
        ExecutionID: uuid.New(),
        Graph:       graph,
        TriggerData: map[string]any{"entity_id": "replay-entity"},
    }

    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    run, err := tc.ExecuteWorkflow(ctx, temporalclient.StartWorkflowOptions{
        TaskQueue: taskQueue,
    }, ExecuteGraphWorkflow, input)
    require.NoError(t, err)
    require.NoError(t, run.Get(ctx, nil))

    // Step 2: Fetch completed workflow history.
    // GetWorkflowHistory returns an iterator (not a slice).
    // The false param means non-long-poll (fetch completed history immediately).
    // The 0 param is the event filter type (0 = all events).
    iter := tc.GetWorkflowHistory(ctx, run.GetID(), run.GetRunID(), false, 0)

    // Step 3: Replay with same workflow definitions.
    // Only workflow functions need registration (NOT activities — they're in the history).
    replayer := worker.NewWorkflowReplayer()
    replayer.RegisterWorkflow(ExecuteGraphWorkflow)
    replayer.RegisterWorkflow(ExecuteBranchUntilConvergence)

    // ReplayWorkflowHistory returns error if workflow code produces different
    // Temporal commands than what's recorded in the history (non-determinism error).
    err = replayer.ReplayWorkflowHistory(nil, iter)
    require.NoError(t, err, "replay should not produce non-determinism error")
}

func TestReplay_SequentialChain(t *testing.T) {
    if testing.Short() {
        t.Skip("skipping integration test in short mode")
    }

    c := foundationtemporal.GetTestContainer(t)
    tc, err := foundationtemporal.NewTestClient(c.HostPort)
    require.NoError(t, err)
    defer tc.Close()

    handler0 := &testActionHandler{actionType: "step_0", result: map[string]any{"v": 0}}
    handler1 := &testActionHandler{actionType: "step_1", result: map[string]any{"v": 1}}
    handler2 := &testActionHandler{actionType: "step_2", result: map[string]any{"v": 2}}

    taskQueue := fmt.Sprintf("test-replay-%s", t.Name())
    w := startTestWorker(t, tc, taskQueue, handler0, handler1, handler2)
    defer w.Stop()

    graph, _ := buildLinearGraph("step_0", "step_1", "step_2")
    input := WorkflowInput{
        RuleID:      uuid.New(),
        RuleName:    "replay-chain",
        ExecutionID: uuid.New(),
        Graph:       graph,
        TriggerData: map[string]any{},
    }

    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    run, err := tc.ExecuteWorkflow(ctx, temporalclient.StartWorkflowOptions{
        TaskQueue: taskQueue,
    }, ExecuteGraphWorkflow, input)
    require.NoError(t, err)
    require.NoError(t, run.Get(ctx, nil))

    iter := tc.GetWorkflowHistory(ctx, run.GetID(), run.GetRunID(), false, 0)

    replayer := worker.NewWorkflowReplayer()
    replayer.RegisterWorkflow(ExecuteGraphWorkflow)
    replayer.RegisterWorkflow(ExecuteBranchUntilConvergence)

    err = replayer.ReplayWorkflowHistory(nil, iter)
    require.NoError(t, err, "sequential chain replay should be deterministic")
}
```

**Notes**:
- `worker.NewWorkflowReplayer()` is the SDK's built-in replay tool
- `ReplayWorkflowHistory` returns error if workflow code produces different commands than history
- Only workflow functions need registration for replay (NOT activities — they're in the history)
- Use `testing.Short()` guard — integration tests are slower (~5s each)
- Real container tests verify the full gRPC round-trip that SDK test suite skips
- `GetWorkflowHistory(ctx, workflowID, runID, false, 0)` — `false` = non-long-poll, `0` = all events
- Replay tests are the gold standard for Temporal deployment safety
- Histories are generated dynamically per test run. Phase 12 may add golden file regression tests in `testdata/` for cross-version compatibility checks.
- Each test uses `fmt.Sprintf("test-replay-%s", t.Name())` for unique task queues

---

## Validation Criteria

- [ ] All SDK test suite tests pass (`go test -run "TestWorkflow_" ./business/sdk/workflow/temporal/...`)
- [ ] All integration tests pass with real container (`go test -run "TestIntegration_" ./business/sdk/workflow/temporal/...`)
- [ ] Replay tests verify no non-determinism errors (`go test -run "TestReplay_" ./business/sdk/workflow/temporal/...`)
- [ ] Single action, sequential chain, condition branching all verified
- [ ] Parallel branches with convergence + fire-and-forget both verified
- [ ] MergedContext propagation confirmed (action N sees results from action N-1)
- [ ] Invalid input rejected with clear errors (rule_id, start edge, empty graph)
- [ ] Error propagation: activity errors surface in workflow error with action name
- [ ] Unknown action type returns "no handler registered" error
- [ ] `go test ./business/sdk/workflow/temporal/... -count=1` passes (all new + existing tests)
- [ ] `go vet ./business/sdk/workflow/temporal/...` clean
- [ ] `go test -race ./business/sdk/workflow/temporal/...` clean
- [ ] Integration tests skipped in `go test -short` mode

---

## Deliverables

- `business/sdk/workflow/temporal/workflow_test.go`
- `business/sdk/workflow/temporal/workflow_parallel_test.go`
- `business/sdk/workflow/temporal/workflow_replay_test.go`

---

## Gotchas & Tips

### Common Pitfalls

- **SDK test suite vs real container**: `TestWorkflowEnvironment` does NOT use gRPC. Some behaviors differ:
  - Child workflow execution is synchronous in test suite (parallel timing not realistic)
  - `activity.GetInfo().TaskToken` is empty in test suite (async activity completion won't work — test async in Phase 12)
  - Context propagation may differ between test suite and real server
- **No `t.Parallel()` with SDK test suite**: `TestWorkflowEnvironment` is NOT goroutine-safe. Do NOT use `t.Parallel()` on SDK test suite tests. Each test must run serially or create its own `setupTestEnv` instance (which they do — each test gets a fresh env).
- **Existing `workflow_test.go` conflict**: Phase 4 tests are in `graph_executor_test.go` and `models_test.go`. There is NO existing `workflow_test.go` — safe to create it.
- **Activity registration**: When using SDK test suite, register `&Activities{...}` (struct pointer), NOT individual methods. The test suite resolves activity functions by struct method name string (e.g., `"ExecuteActionActivity"`).
- **Fire-and-forget child workflows**: `PARENT_CLOSE_POLICY_ABANDON` may behave differently in SDK test suite — child workflows execute synchronously instead of running independently. If fire-and-forget tests are flaky in test suite, move them to integration tests with real container.
- **Replay history format**: `GetWorkflowHistory` returns an iterator, not a slice. `ReplayWorkflowHistory` accepts the iterator directly. Only workflow functions need registration for replay (activities are recorded in the history).
- **Test isolation with unique task queues**: Each integration test MUST use a unique task queue name (`fmt.Sprintf("test-%s", t.Name())`) to prevent interference when tests share the same Temporal container.
- **Docker dependency**: Integration tests need Docker running. Guard with `testing.Short()`.
- **Workflow versioning**: `workflow.go` uses `workflow.GetVersion(ctx, "graph-interpreter", DefaultVersion, 1)`. Replay tests will pass as long as the version doesn't change between record and replay. If version is incremented in a future phase, existing replay tests will need updated histories.
- **ActionEdge.SourceActionID is `*uuid.UUID`**: When building test graphs, take the address of a local variable (`src := id; edge.SourceActionID = &src`). Using `&ids[i]` directly indexes into a slice and may cause issues if the slice is reallocated.
- **Deactivated actions**: Current implementation does NOT skip deactivated actions — `executeSingleAction` dispatches regardless of `IsActive`. Tests document this behavior. If skip logic is added later, update tests.

### Tips

- Start with SDK test suite tests (Tasks 1-3) — they're fast and give 80% of the confidence
- Use `env.OnActivity().Return(result, nil)` to mock activities instead of real handlers when testing workflow logic specifically
- Table-driven tests work great for condition branching (true/false/always x different results)
- `require.Contains(t, err.Error(), "substring")` is useful for validation error tests
- The SDK test suite's `env.IsWorkflowCompleted()` is useful to verify workflows didn't hang
- Replay tests generate histories dynamically each run (no `testdata/` golden files needed in Phase 11)

---

## Mock Data and Fixtures

### Test Action Handlers

All handlers use `testActionHandler` with configurable `actionType`, `result`, `err`, and `called` counter. Create per-test instances for isolation.

| Usage Pattern | ActionType | Returns | Use Case |
|--------------|-----------|---------|----------|
| result handler | `"test_action"` / `"step_N"` | `map[string]any{"status": "done"}` | Basic execution, sequential chains |
| condition handler | `"evaluate_condition"` | `map[string]any{"branch_taken": "true_branch"}` or `"false_branch"` | Condition branching |
| error handler | `"fail_action"` | `.err = fmt.Errorf("simulated failure")` | Error propagation |
| fork handler | `"fork_action"` | `map[string]any{"forked": true}` | Parallel execution trigger |
| merge handler | `"merge_action"` | `map[string]any{"merged": true}` | Convergence point |

### Standard Test Graphs (via builder helpers)

| Builder | Shape | Actions | Edges | Tests |
|---------|-------|---------|-------|-------|
| `buildLinearGraph("t")` | `start -> A` | 1 | 1 start | Single action |
| `buildLinearGraph("t","t","t")` | `start -> A -> B -> C` | 3 | 1 start + 2 sequence | Sequential chain |
| `buildDiamondGraph("t","f")` | `start -> cond -> A/B -> merge` | 4 | 1 start + 2 branch + 2 sequence | Condition branching |
| `buildParallelGraph("a","b")` | `start -> fork -> A,B -> merge` | 4 | 1 start + 2 seq + 2 seq | Parallel convergence |
| `buildFireForgetGraph("a","b")` | `start -> fork -> A,B` | 3 | 1 start + 2 sequence | No convergence |

---

## Scope Boundary

### In Scope (Phase 11)
- Workflow execution tests (SDK test suite + real container)
- Activity dispatch verification (handler called, result returned)
- MergedContext propagation across actions
- Condition branching (true/false)
- Parallel execution + convergence
- Fire-and-forget branches
- Replay determinism testing
- Input validation error tests (missing rule_id, no start edges, empty graph)
- Error propagation (activity error -> workflow error)
- Unknown action type handling
- Direct `ExecuteBranchUntilConvergence` child workflow testing

### Out of Scope (Phase 12)
- Continue-As-New tests (history length threshold crossing) — Continue-As-New LOGIC exists in workflow.go but testing workflows large enough to trigger the 10K event threshold is a Phase 12 concern
- Payload size limit / truncation tests
- Async activity completion (ErrResultPending + external callback)
- Activity retry / timeout behavior
- Error handling edge cases (handler panics, partial branch failures in parallel)

### Out of Scope (Other Phases)
- Graph executor unit tests (Phase 10, complete)
- Edge store database tests (Phase 8, complete)
- Kubernetes deployment tests (Phase 13)

---

## Commands Reference

```bash
# Start this phase
/workflow-temporal-next

# Validate this phase
/workflow-temporal-validate 11

# Review plan before implementing
/workflow-temporal-plan-review 11

# Review code after implementing
/workflow-temporal-review 11
```
