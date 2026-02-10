# Workflow Engine: Temporal Implementation

## Overview

This document provides the complete Temporal-based workflow engine implementation for Ichor. The engine interprets visual workflow graphs (nodes + edges stored in database) and executes them with full durability, parallel branch support, and async continuation.

**Decision**: Use Temporal from the start rather than building custom continuation infrastructure.

**Rationale**:
- Durability, crash recovery, and replay handled automatically
- Built-in visibility UI for debugging executions
- Workflow versioning for safe deployments with in-flight workflows
- Battle-tested at scale (Uber, Netflix, Stripe)
- Your action handlers become activities with minimal wrapping

---

## Temporal Determinism Requirements (Critical)

Temporal workflows must be **deterministic** - given the same input and history, they must produce the same sequence of commands. Violating this causes replay failures.

### What You MUST NOT Do in Workflow Code

| Forbidden | Why | Solution |
|-----------|-----|----------|
| `time.Now()` | Returns different value on replay | Use `workflow.Now(ctx)` |
| `rand.Int()` | Non-deterministic | Use `workflow.SideEffect()` |
| Direct HTTP/DB calls | External state changes | Move to Activities |
| Map iteration without sorting | Go maps iterate in random order | Sort keys first |
| `uuid.New()` | Non-deterministic | Use `workflow.SideEffect()` or pass as input |

### Map Iteration Safety

All map iterations in workflow code must be deterministic. **Always sort keys before iterating:**

```go
// WRONG - Non-deterministic
for nodeID, count := range reachable {
    // Order varies between runs!
}

// CORRECT - Deterministic
nodeIDs := make([]uuid.UUID, 0, len(reachable))
for nodeID := range reachable {
    nodeIDs = append(nodeIDs, nodeID)
}
sort.Slice(nodeIDs, func(i, j int) bool {
    return nodeIDs[i].String() < nodeIDs[j].String()
})
for _, nodeID := range nodeIDs {
    count := reachable[nodeID]
    // Now deterministic!
}
```

### What IS Safe in Workflow Code

- Reading workflow input (stored in history, same on replay)
- Branching based on input data
- Calling activities (results stored in history)
- Using `workflow.Now()`, `workflow.Sleep()`, `workflow.SideEffect()`

---

## Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                        Ichor Codebase                           │
├─────────────────────────────────────────────────────────────────┤
│  Visual Editor → rule_actions + action_edges → Postgres        │
│                                    │                            │
│                              Graph Config                       │
│                                    │                            │
│  Entity Events ──────────────────► WorkflowTrigger              │
└────────────────────────────────────┼────────────────────────────┘
                                     │
                                     ▼
┌─────────────────────────────────────────────────────────────────┐
│                      Temporal Server                            │
├─────────────────────────────────────────────────────────────────┤
│  ExecuteGraphWorkflow ──► Activities (your existing handlers)  │
│         │                            │                          │
│    Workflow State              Activity Results                 │
│    (Temporal owns)             (merged into context)            │
└─────────────────────────────────────────────────────────────────┘
```

---

## Graph Constraints (Must Preserve)

These are non-negotiable requirements from the visual workflow editor:

| Constraint | Description |
|------------|-------------|
| **Visual graph model** | Users draw workflows as nodes + edges stored in database |
| **Dynamic definitions** | No compilation; workflows created/modified at runtime via UI |
| **Parallel branches** | Multiple outgoing edges run concurrently |
| **Convergence detection** | Nodes with multiple incoming edges wait for all branches |
| **Edge types** | `start`, `sequence`, `true_branch`, `false_branch`, `always`, `parallel` |
| **Merged context** | Each action's result added to execution context for subsequent actions |
| **Template variables** | `{{action_name.field}}` syntax in action configs |

---

## Core Components

### 1. Models

**File**: `business/sdk/workflow/temporal/models.go`

```go
package temporal

import (
	"encoding/json"

	"github.com/google/uuid"
)

// WorkflowInput is passed when starting a workflow execution
type WorkflowInput struct {
	RuleID       uuid.UUID              `json:"rule_id"`
	ExecutionID  uuid.UUID              `json:"execution_id"`
	Graph        GraphDefinition        `json:"graph"`
	TriggerData  map[string]interface{} `json:"trigger_data"`
}

// GraphDefinition mirrors your database model (rule_actions + action_edges)
type GraphDefinition struct {
	Actions []ActionNode `json:"actions"`
	Edges   []ActionEdge `json:"edges"`
}

// ActionNode represents a single action in the workflow graph
type ActionNode struct {
	ID         uuid.UUID       `json:"id"`
	Name       string          `json:"name"`
	ActionType string          `json:"action_type"`
	Config     json.RawMessage `json:"action_config"`
}

// ActionEdge represents a directed edge between actions
type ActionEdge struct {
	ID             uuid.UUID  `json:"id"`
	SourceActionID *uuid.UUID `json:"source_action_id"` // nil for start edges
	TargetActionID uuid.UUID  `json:"target_action_id"`
	EdgeType       string     `json:"edge_type"` // start, sequence, true_branch, false_branch, always, parallel
	SortOrder      int        `json:"sort_order"`
}

// MergedContext accumulates results from all executed actions
// This is the key data structure for template variable resolution
type MergedContext struct {
	TriggerData   map[string]interface{}            `json:"trigger_data"`
	ActionResults map[string]map[string]interface{} `json:"action_results"` // action_name -> result
	Flattened     map[string]interface{}            `json:"flattened"`      // For template resolution
}

// NewMergedContext creates a context initialized with trigger data
func NewMergedContext(triggerData map[string]interface{}) *MergedContext {
	ctx := &MergedContext{
		TriggerData:   triggerData,
		ActionResults: make(map[string]map[string]interface{}),
		Flattened:     make(map[string]interface{}),
	}

	// Copy trigger data to flattened for initial template resolution
	for k, v := range triggerData {
		ctx.Flattened[k] = v
	}

	return ctx
}

// MaxResultValueSize limits individual result values to prevent payload bloat.
// Large values are truncated with a reference to external storage.
const MaxResultValueSize = 50 * 1024 // 50KB per value

// MergeResult adds an action's result to the context
// Supports template access patterns:
//   - {{action_name}} -> entire result map
//   - {{action_name.field}} -> specific field
//
// Large values are truncated to prevent exceeding Temporal's 2MB payload limit.
func (c *MergedContext) MergeResult(actionName string, result map[string]interface{}) {
	if c.ActionResults == nil {
		c.ActionResults = make(map[string]map[string]interface{})
	}
	if c.Flattened == nil {
		c.Flattened = make(map[string]interface{})
	}

	// Sanitize result to prevent payload size issues
	sanitized := sanitizeResult(result)

	// Store full result by action name
	c.ActionResults[actionName] = sanitized

	// Flatten for template access: "action_name.field" -> value
	for k, v := range sanitized {
		c.Flattened[actionName+"."+k] = v
	}

	// Also store action name pointing to full result for {{action_name}} access
	c.Flattened[actionName] = sanitized
}

// sanitizeResult truncates large values to prevent payload size issues
func sanitizeResult(result map[string]interface{}) map[string]interface{} {
	sanitized := make(map[string]interface{}, len(result))

	for k, v := range result {
		switch val := v.(type) {
		case string:
			if len(val) > MaxResultValueSize {
				// Truncate large strings, store reference note
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
		default:
			// For complex types, check serialized size
			if data, err := json.Marshal(val); err == nil && len(data) > MaxResultValueSize {
				sanitized[k] = "[LARGE_OBJECT_TRUNCATED]"
				sanitized[k+"_truncated"] = true
			} else {
				sanitized[k] = val
			}
		}
	}

	return sanitized
}

// Clone creates a deep copy for parallel branch execution
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

// BranchInput is passed to child workflows for parallel branch execution
type BranchInput struct {
	StartAction      ActionNode      `json:"start_action"`
	ConvergencePoint uuid.UUID       `json:"convergence_point"`
	Graph            GraphDefinition `json:"graph"`
	InitialContext   *MergedContext  `json:"initial_context"`
}

// BranchOutput is returned from child workflows
type BranchOutput struct {
	ActionResults map[string]map[string]interface{} `json:"action_results"`
}

// ActionActivityInput is passed to the action execution activity
type ActionActivityInput struct {
	ActionID   uuid.UUID              `json:"action_id"`
	ActionName string                 `json:"action_name"`
	ActionType string                 `json:"action_type"`
	Config     json.RawMessage        `json:"config"`
	Context    map[string]interface{} `json:"context"` // Merged context for template resolution
}

// ActionActivityOutput is returned from the action execution activity
type ActionActivityOutput struct {
	ActionID   uuid.UUID              `json:"action_id"`
	ActionName string                 `json:"action_name"`
	Result     map[string]interface{} `json:"result"`
	Success    bool                   `json:"success"`
}
```

---

### 2. Graph Executor

**File**: `business/sdk/workflow/temporal/graph_executor.go`

```go
package temporal

import (
	"sort"

	"github.com/google/uuid"
)

// GraphExecutor traverses the workflow graph respecting edge types
type GraphExecutor struct {
	graph         GraphDefinition
	actionsByID   map[uuid.UUID]ActionNode
	edgesBySource map[uuid.UUID][]ActionEdge // source_action_id -> outgoing edges
	incomingCount map[uuid.UUID]int          // For convergence detection
}

// NewGraphExecutor creates an executor from a graph definition
func NewGraphExecutor(graph GraphDefinition) *GraphExecutor {
	e := &GraphExecutor{
		graph:         graph,
		actionsByID:   make(map[uuid.UUID]ActionNode),
		edgesBySource: make(map[uuid.UUID][]ActionEdge),
		incomingCount: make(map[uuid.UUID]int),
	}

	// Index actions by ID
	for _, action := range graph.Actions {
		e.actionsByID[action.ID] = action
	}

	// Index edges by source and count incoming edges
	for _, edge := range graph.Edges {
		if edge.SourceActionID != nil {
			e.edgesBySource[*edge.SourceActionID] = append(e.edgesBySource[*edge.SourceActionID], edge)
		} else {
			// Start edges (source_action_id = nil) stored under nil UUID
			e.edgesBySource[uuid.Nil] = append(e.edgesBySource[uuid.Nil], edge)
		}
		e.incomingCount[edge.TargetActionID]++
	}

	// Sort edges by sort_order for deterministic execution
	for sourceID := range e.edgesBySource {
		edges := e.edgesBySource[sourceID]
		sort.Slice(edges, func(i, j int) bool {
			return edges[i].SortOrder < edges[j].SortOrder
		})
		e.edgesBySource[sourceID] = edges
	}

	return e
}

// GetStartActions returns all actions with start edges (source_action_id = nil)
func (e *GraphExecutor) GetStartActions() []ActionNode {
	startEdges := e.edgesBySource[uuid.Nil]
	actions := make([]ActionNode, 0, len(startEdges))

	for _, edge := range startEdges {
		if edge.EdgeType == "start" {
			if action, ok := e.actionsByID[edge.TargetActionID]; ok {
				actions = append(actions, action)
			}
		}
	}

	return actions
}

// GetNextActions returns the next actions to execute based on edge types and action result
//
// Edge type behavior:
//   - "sequence": Always follow
//   - "parallel": Always follow (concurrent with other parallel edges)
//   - "true_branch": Follow if result["branch_taken"] == "true_branch"
//   - "false_branch": Follow if result["branch_taken"] == "false_branch"
//   - "always": Always follow (parallel to condition branches)
func (e *GraphExecutor) GetNextActions(sourceActionID uuid.UUID, result map[string]interface{}) []ActionNode {
	edges := e.edgesBySource[sourceActionID]
	if len(edges) == 0 {
		return nil // End of path
	}

	var nextActions []ActionNode

	// Determine which edges to follow based on edge type and result
	for _, edge := range edges {
		shouldFollow := false

		switch edge.EdgeType {
		case "sequence", "parallel":
			// Always follow sequential and parallel edges
			shouldFollow = true

		case "true_branch":
			// Follow if condition evaluated to true
			if branch, ok := result["branch_taken"].(string); ok && branch == "true_branch" {
				shouldFollow = true
			}

		case "false_branch":
			// Follow if condition evaluated to false
			if branch, ok := result["branch_taken"].(string); ok && branch == "false_branch" {
				shouldFollow = true
			}

		case "always":
			// Always follow (used for actions that run regardless of condition result)
			shouldFollow = true

		case "start":
			// Start edges are only used for initial dispatch
			continue
		}

		if shouldFollow {
			if action, ok := e.actionsByID[edge.TargetActionID]; ok {
				nextActions = append(nextActions, action)
			}
		}
	}

	return nextActions
}

// FindConvergencePoint detects if multiple branches lead to the same node
// Returns the convergence node if found, nil if branches are fire-and-forget
//
// IMPORTANT: This function must be deterministic for Temporal replay.
// All map iterations are sorted by UUID to ensure consistent results.
func (e *GraphExecutor) FindConvergencePoint(branches []ActionNode) *ActionNode {
	if len(branches) <= 1 {
		return nil
	}

	// Track which nodes each branch can reach
	reachable := make(map[uuid.UUID]int) // node_id -> count of branches that can reach it

	for _, branch := range branches {
		visited := e.findReachableNodes(branch.ID, make(map[uuid.UUID]bool))
		for nodeID := range visited {
			reachable[nodeID]++
		}
	}

	// Find first node reachable by ALL branches (ordered by graph structure)
	// We want the closest convergence point
	//
	// DETERMINISM: Sort node IDs before iteration to ensure consistent results on replay
	nodeIDs := make([]uuid.UUID, 0, len(reachable))
	for nodeID := range reachable {
		nodeIDs = append(nodeIDs, nodeID)
	}
	sort.Slice(nodeIDs, func(i, j int) bool {
		return nodeIDs[i].String() < nodeIDs[j].String()
	})

	var convergencePoint *ActionNode
	minDepth := -1

	for _, nodeID := range nodeIDs {
		count := reachable[nodeID]
		if count == len(branches) {
			// This node is reachable by all branches
			depth := e.calculateMinDepth(branches[0].ID, nodeID)
			if minDepth == -1 || depth < minDepth {
				minDepth = depth
				action := e.actionsByID[nodeID]
				convergencePoint = &action
			}
		}
	}

	return convergencePoint
}

// findReachableNodes performs BFS to find all nodes reachable from start
func (e *GraphExecutor) findReachableNodes(startID uuid.UUID, visited map[uuid.UUID]bool) map[uuid.UUID]bool {
	if visited[startID] {
		return visited
	}
	visited[startID] = true

	// Follow all outgoing edges (assume condition results don't matter for reachability analysis)
	for _, edge := range e.edgesBySource[startID] {
		e.findReachableNodes(edge.TargetActionID, visited)
	}

	return visited
}

// calculateMinDepth calculates minimum edge count from source to target
func (e *GraphExecutor) calculateMinDepth(sourceID, targetID uuid.UUID) int {
	if sourceID == targetID {
		return 0
	}

	visited := make(map[uuid.UUID]bool)
	queue := []struct {
		id    uuid.UUID
		depth int
	}{{sourceID, 0}}

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		if visited[current.id] {
			continue
		}
		visited[current.id] = true

		for _, edge := range e.edgesBySource[current.id] {
			if edge.TargetActionID == targetID {
				return current.depth + 1
			}
			queue = append(queue, struct {
				id    uuid.UUID
				depth int
			}{edge.TargetActionID, current.depth + 1})
		}
	}

	return -1 // Not reachable
}

// PathLeadsTo checks if starting from actionID eventually reaches targetID
func (e *GraphExecutor) PathLeadsTo(actionID, targetID uuid.UUID) bool {
	reachable := e.findReachableNodes(actionID, make(map[uuid.UUID]bool))
	return reachable[targetID]
}

// Graph returns the underlying graph definition
func (e *GraphExecutor) Graph() GraphDefinition {
	return e.graph
}

// GetAction retrieves an action by ID
func (e *GraphExecutor) GetAction(id uuid.UUID) (ActionNode, bool) {
	action, ok := e.actionsByID[id]
	return action, ok
}

// HasMultipleIncoming returns true if the node has more than one incoming edge
// This indicates a potential convergence point
func (e *GraphExecutor) HasMultipleIncoming(actionID uuid.UUID) bool {
	return e.incomingCount[actionID] > 1
}
```

---

### 3. Main Workflow

**File**: `business/sdk/workflow/temporal/workflow.go`

```go
package temporal

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

const (
	TaskQueue = "ichor-workflow-queue"

	// HistoryLengthThreshold triggers Continue-As-New to prevent unbounded history growth.
	// Temporal has a 50K event limit; we reset well before that.
	HistoryLengthThreshold = 10000

	// ContextSizeWarningBytes logs a warning when merged context approaches payload limits.
	// Temporal has a 2MB hard limit with warnings at 256KB.
	ContextSizeWarningBytes = 200 * 1024 // 200KB
)

// ExecuteGraphWorkflow interprets any graph definition dynamically
// This is the core workflow that respects all graphing constraints
func ExecuteGraphWorkflow(ctx workflow.Context, input WorkflowInput) error {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting graph workflow",
		"rule_id", input.RuleID,
		"execution_id", input.ExecutionID,
	)

	// Initialize execution context
	mergedCtx := NewMergedContext(input.TriggerData)

	// Build graph executor
	executor := NewGraphExecutor(input.Graph)

	// Find start actions (edges with source_action_id = nil)
	startActions := executor.GetStartActions()
	if len(startActions) == 0 {
		logger.Info("Empty workflow - no start actions found")
		return nil
	}

	// Execute from start (may be multiple parallel start actions)
	return executeActions(ctx, executor, startActions, mergedCtx, input)
}

// checkContinueAsNew returns a ContinueAsNewError if history has grown too large.
// This prevents hitting Temporal's 50K event limit on long-running workflows.
func checkContinueAsNew(ctx workflow.Context, input WorkflowInput, mergedCtx *MergedContext) error {
	info := workflow.GetInfo(ctx)
	if info.GetCurrentHistoryLength() > HistoryLengthThreshold {
		logger := workflow.GetLogger(ctx)
		logger.Info("History threshold exceeded, continuing as new workflow",
			"history_length", info.GetCurrentHistoryLength(),
			"threshold", HistoryLengthThreshold,
		)

		// Preserve accumulated context in new workflow
		input.TriggerData = mergedCtx.Flattened

		return workflow.NewContinueAsNewError(ctx, ExecuteGraphWorkflow, input)
	}
	return nil
}

// executeActions handles both sequential and parallel execution
func executeActions(ctx workflow.Context, executor *GraphExecutor, actions []ActionNode, mergedCtx *MergedContext, input WorkflowInput) error {
	if len(actions) == 0 {
		return nil
	}

	// Check if we need to continue-as-new to avoid history size limits
	if err := checkContinueAsNew(ctx, input, mergedCtx); err != nil {
		return err
	}

	if len(actions) == 1 {
		// Sequential execution
		return executeSingleAction(ctx, executor, actions[0], mergedCtx, input)
	}

	// Parallel execution - check for convergence
	convergencePoint := executor.FindConvergencePoint(actions)

	if convergencePoint == nil {
		// Fire-and-forget parallel branches - no convergence
		return executeFireAndForget(ctx, executor, actions, mergedCtx, input)
	}

	// Parallel with convergence - must wait for all branches
	return executeParallelWithConvergence(ctx, executor, actions, convergencePoint, mergedCtx, input)
}

// executeSingleAction executes one action and continues to next
func executeSingleAction(ctx workflow.Context, executor *GraphExecutor, action ActionNode, mergedCtx *MergedContext, input WorkflowInput) error {
	logger := workflow.GetLogger(ctx)
	logger.Info("Executing action",
		"action_id", action.ID,
		"action_name", action.Name,
		"action_type", action.ActionType,
	)

	// Prepare activity input
	activityInput := ActionActivityInput{
		ActionID:   action.ID,
		ActionName: action.Name,
		ActionType: action.ActionType,
		Config:     action.Config,
		Context:    mergedCtx.Flattened,
	}

	// Configure activity options based on action type
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 5 * time.Minute,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    time.Minute,
			MaximumAttempts:    3,
		},
	}

	// Async actions get longer timeouts
	if isAsyncAction(action.ActionType) {
		ao.StartToCloseTimeout = 30 * time.Minute
		ao.HeartbeatTimeout = time.Minute
	}

	// Human-in-the-loop actions can take days
	if isHumanAction(action.ActionType) {
		ao.StartToCloseTimeout = 7 * 24 * time.Hour // 7 days
		ao.HeartbeatTimeout = time.Hour
	}

	activityCtx := workflow.WithActivityOptions(ctx, ao)

	// Execute the action as a Temporal activity
	var result ActionActivityOutput
	err := workflow.ExecuteActivity(activityCtx, ExecuteActionActivity, activityInput).Get(ctx, &result)
	if err != nil {
		logger.Error("Action failed",
			"action_id", action.ID,
			"action_name", action.Name,
			"error", err,
		)
		return err
	}

	// Merge result into context for subsequent actions
	mergedCtx.MergeResult(action.Name, result.Result)

	logger.Info("Action completed",
		"action_id", action.ID,
		"action_name", action.Name,
		"success", result.Success,
	)

	// Get next actions based on result and edge types
	nextActions := executor.GetNextActions(action.ID, result.Result)

	if len(nextActions) == 0 {
		// End of this path
		return nil
	}

	// Continue execution
	return executeActions(ctx, executor, nextActions, mergedCtx, input)
}

// executeFireAndForget launches parallel branches without waiting
func executeFireAndForget(ctx workflow.Context, executor *GraphExecutor, branches []ActionNode, mergedCtx *MergedContext, input WorkflowInput) error {
	logger := workflow.GetLogger(ctx)
	logger.Info("Executing fire-and-forget parallel branches",
		"branch_count", len(branches),
	)

	// Launch each branch concurrently using workflow.Go
	for i, branch := range branches {
		// Capture for closure
		branchAction := branch
		branchIndex := i

		workflow.Go(ctx, func(gCtx workflow.Context) {
			// Clone context for this branch
			branchCtx := mergedCtx.Clone()

			logger.Info("Starting fire-and-forget branch",
				"branch_index", branchIndex,
				"start_action", branchAction.Name,
			)

			// Execute this branch (errors logged but don't fail parent)
			if err := executeSingleAction(gCtx, executor, branchAction, branchCtx, input); err != nil {
				logger.Error("Fire-and-forget branch failed",
					"branch_index", branchIndex,
					"error", err,
				)
			}
		})
	}

	// Don't wait for fire-and-forget branches
	return nil
}

// executeParallelWithConvergence executes branches and waits for all at convergence
func executeParallelWithConvergence(
	ctx workflow.Context,
	executor *GraphExecutor,
	branches []ActionNode,
	convergencePoint *ActionNode,
	mergedCtx *MergedContext,
	input WorkflowInput,
) error {
	logger := workflow.GetLogger(ctx)
	logger.Info("Executing parallel branches with convergence",
		"branch_count", len(branches),
		"convergence_point", convergencePoint.Name,
	)

	// Create selector for waiting on all branches
	selector := workflow.NewSelector(ctx)
	branchResults := make([]BranchOutput, len(branches))
	branchErrors := make([]error, len(branches))

	for i, branch := range branches {
		// Capture for closure
		branchIndex := i
		branchAction := branch

		// Execute each branch as a child workflow
		childCtx := workflow.WithChildOptions(ctx, workflow.ChildWorkflowOptions{
			WorkflowID: fmt.Sprintf("%s-branch-%d-%s",
				workflow.GetInfo(ctx).WorkflowExecution.ID,
				branchIndex,
				branchAction.ID,
			),
		})

		future := workflow.ExecuteChildWorkflow(childCtx, ExecuteBranchUntilConvergence,
			BranchInput{
				StartAction:      branchAction,
				ConvergencePoint: convergencePoint.ID,
				Graph:            executor.Graph(),
				InitialContext:   mergedCtx,
			},
		)

		selector.AddFuture(future, func(f workflow.Future) {
			var output BranchOutput
			branchErrors[branchIndex] = f.Get(ctx, &output)
			branchResults[branchIndex] = output
		})
	}

	// Wait for all branches
	for i := 0; i < len(branches); i++ {
		selector.Select(ctx)
	}

	// Check for errors
	for i, err := range branchErrors {
		if err != nil {
			logger.Error("Branch failed",
				"branch_index", i,
				"error", err,
			)
			return fmt.Errorf("branch %d failed: %w", i, err)
		}
	}

	// Merge all branch results into context
	for _, br := range branchResults {
		for actionName, actionResult := range br.ActionResults {
			mergedCtx.MergeResult(actionName, actionResult)
		}
	}

	logger.Info("All branches converged, continuing from convergence point",
		"convergence_point", convergencePoint.Name,
	)

	// Continue from convergence point
	return executeSingleAction(ctx, executor, *convergencePoint, mergedCtx, input)
}

// ExecuteBranchUntilConvergence executes actions until reaching the convergence point
// This is a child workflow for each parallel branch
func ExecuteBranchUntilConvergence(ctx workflow.Context, input BranchInput) (BranchOutput, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting branch execution",
		"start_action", input.StartAction.Name,
		"convergence_point", input.ConvergencePoint,
	)

	executor := NewGraphExecutor(input.Graph)
	mergedCtx := input.InitialContext.Clone()

	currentAction := input.StartAction

	for {
		// Check if we've reached convergence point
		if currentAction.ID == input.ConvergencePoint {
			logger.Info("Branch reached convergence point")
			break
		}

		// Execute action
		activityInput := ActionActivityInput{
			ActionID:   currentAction.ID,
			ActionName: currentAction.Name,
			ActionType: currentAction.ActionType,
			Config:     currentAction.Config,
			Context:    mergedCtx.Flattened,
		}

		ao := workflow.ActivityOptions{
			StartToCloseTimeout: 5 * time.Minute,
			RetryPolicy: &temporal.RetryPolicy{
				InitialInterval:    time.Second,
				BackoffCoefficient: 2.0,
				MaximumAttempts:    3,
			},
		}

		if isAsyncAction(currentAction.ActionType) {
			ao.StartToCloseTimeout = 30 * time.Minute
		}

		if isHumanAction(currentAction.ActionType) {
			ao.StartToCloseTimeout = 7 * 24 * time.Hour
		}

		activityCtx := workflow.WithActivityOptions(ctx, ao)

		var result ActionActivityOutput
		err := workflow.ExecuteActivity(activityCtx, ExecuteActionActivity, activityInput).Get(ctx, &result)
		if err != nil {
			return BranchOutput{}, err
		}

		mergedCtx.MergeResult(currentAction.Name, result.Result)

		// Get next action (should be single path within a branch before convergence)
		nextActions := executor.GetNextActions(currentAction.ID, result.Result)
		if len(nextActions) == 0 {
			// Dead end before convergence - this shouldn't happen in well-formed graphs
			break
		}

		// Take first next action (branches within branches would need nested handling)
		currentAction = nextActions[0]
	}

	return BranchOutput{
		ActionResults: mergedCtx.ActionResults,
	}, nil
}

// isAsyncAction returns true for actions that involve external async operations
func isAsyncAction(actionType string) bool {
	asyncTypes := map[string]bool{
		"allocate_inventory":   true,
		"send_email":           true,
		"credit_check":         true,
		"fraud_detection":      true,
		"third_party_api_call": true,
		"reserve_shipping":     true,
	}
	return asyncTypes[actionType]
}

// isHumanAction returns true for actions that require human interaction
func isHumanAction(actionType string) bool {
	humanTypes := map[string]bool{
		"manager_approval":    true,
		"manual_review":       true,
		"human_verification":  true,
		"approval_request":    true,
	}
	return humanTypes[actionType]
}
```

---

### 3b. Workflow Versioning

When modifying workflow logic while workflows are in-flight, use Temporal's versioning to ensure safe replay.

**File**: `business/sdk/workflow/temporal/workflow.go` (add to ExecuteGraphWorkflow)

```go
// ExecuteGraphWorkflow with versioning support
func ExecuteGraphWorkflow(ctx workflow.Context, input WorkflowInput) error {
	logger := workflow.GetLogger(ctx)

	// Version the interpreter logic for safe deployments
	// Increment maxVersion when making breaking changes to execution logic
	v := workflow.GetVersion(ctx, "graph-interpreter", workflow.DefaultVersion, 1)

	logger.Info("Starting graph workflow",
		"rule_id", input.RuleID,
		"execution_id", input.ExecutionID,
		"interpreter_version", v,
	)

	// Version-specific logic (if needed in future)
	switch v {
	case workflow.DefaultVersion:
		// Original logic (v0)
		return executeGraphV0(ctx, input)
	case 1:
		// Current logic with Continue-As-New support
		return executeGraphV1(ctx, input)
	default:
		return executeGraphV1(ctx, input)
	}
}
```

**When to Increment Version:**
- Changing the order of activity calls
- Adding/removing activities in the execution path
- Changing how parallel branches are structured
- Modifying convergence detection logic

**Safe Changes (no version needed):**
- Bug fixes in activities (activities are versioned separately)
- Adding new action types to the registry
- Changing activity timeout values
- Logging changes

---

### 4. Activities

**File**: `business/sdk/workflow/temporal/activities.go`

```go
package temporal

import (
	"context"
	"encoding/json"
	"fmt"

	"go.temporal.io/sdk/activity"
)

// ActionHandler interface matches your existing handler structure
type ActionHandler interface {
	Execute(ctx context.Context, config json.RawMessage, execCtx ActionExecutionContext) (interface{}, error)
}

// ActionExecutionContext matches your existing context structure
type ActionExecutionContext struct {
	ActionID   string
	ActionName string
	RawData    map[string]interface{}
}

// ActionRegistry holds all registered action handlers
type ActionRegistry struct {
	handlers map[string]ActionHandler
}

// NewActionRegistry creates a new registry
func NewActionRegistry() *ActionRegistry {
	return &ActionRegistry{
		handlers: make(map[string]ActionHandler),
	}
}

// Register adds a handler for an action type
func (r *ActionRegistry) Register(actionType string, handler ActionHandler) {
	r.handlers[actionType] = handler
}

// GetHandler retrieves a handler by action type
func (r *ActionRegistry) GetHandler(actionType string) (ActionHandler, error) {
	h, ok := r.handlers[actionType]
	if !ok {
		return nil, fmt.Errorf("no handler registered for action type: %s", actionType)
	}
	return h, nil
}

// ActivityDependencies holds all dependencies needed by activities
type ActivityDependencies struct {
	ActionRegistry *ActionRegistry
}

var deps *ActivityDependencies

// SetActivityDependencies initializes activity dependencies (called during worker setup)
func SetActivityDependencies(d *ActivityDependencies) {
	deps = d
}

// ExecuteActionActivity dispatches to the appropriate handler based on action type
// This wraps your existing action handlers as Temporal activities
func ExecuteActionActivity(ctx context.Context, input ActionActivityInput) (ActionActivityOutput, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Executing action activity",
		"action_id", input.ActionID,
		"action_name", input.ActionName,
		"action_type", input.ActionType,
	)

	if deps == nil || deps.ActionRegistry == nil {
		return ActionActivityOutput{}, fmt.Errorf("activity dependencies not initialized")
	}

	// Build execution context matching your existing structure
	execCtx := ActionExecutionContext{
		ActionID:   input.ActionID.String(),
		ActionName: input.ActionName,
		RawData:    input.Context, // Merged context with all prior results
	}

	// Get handler from registry
	handler, err := deps.ActionRegistry.GetHandler(input.ActionType)
	if err != nil {
		return ActionActivityOutput{
			ActionID:   input.ActionID,
			ActionName: input.ActionName,
			Success:    false,
		}, fmt.Errorf("unknown action type %s: %w", input.ActionType, err)
	}

	// Execute the action using your existing handler
	result, err := handler.Execute(ctx, input.Config, execCtx)
	if err != nil {
		logger.Error("Action execution failed",
			"action_id", input.ActionID,
			"error", err,
		)
		return ActionActivityOutput{
			ActionID:   input.ActionID,
			ActionName: input.ActionName,
			Success:    false,
		}, err
	}

	// Convert result to map for context merging
	resultMap := toResultMap(result)

	logger.Info("Action execution succeeded",
		"action_id", input.ActionID,
		"action_name", input.ActionName,
	)

	return ActionActivityOutput{
		ActionID:   input.ActionID,
		ActionName: input.ActionName,
		Result:     resultMap,
		Success:    true,
	}, nil
}

// toResultMap converts any result type to a map for context merging
func toResultMap(result interface{}) map[string]interface{} {
	if result == nil {
		return map[string]interface{}{}
	}

	// If already a map, return directly
	if m, ok := result.(map[string]interface{}); ok {
		return m
	}

	// Marshal/unmarshal for struct types
	data, err := json.Marshal(result)
	if err != nil {
		return map[string]interface{}{"raw": result}
	}

	var m map[string]interface{}
	if err := json.Unmarshal(data, &m); err != nil {
		return map[string]interface{}{"raw": result}
	}

	return m
}
```

---

### 4b. Asynchronous Activity Completion (RabbitMQ Pattern)

For actions that publish to RabbitMQ and wait for an external response, use Temporal's async activity completion pattern instead of just extending timeouts.

**File**: `business/sdk/workflow/temporal/activities_async.go`

```go
package temporal

import (
	"context"
	"encoding/json"

	"go.temporal.io/sdk/activity"
)

// AsyncActionHandler handles actions that complete asynchronously
// (e.g., publish to RabbitMQ and wait for consumer response)
type AsyncActionHandler interface {
	// StartAsync initiates the async operation and returns immediately
	// The taskToken should be stored/forwarded for later completion
	StartAsync(ctx context.Context, config json.RawMessage, execCtx ActionExecutionContext, taskToken []byte) error
}

// ExecuteAsyncActionActivity handles actions that complete asynchronously
// The activity returns ErrResultPending; completion happens via CompleteActivity API
func ExecuteAsyncActionActivity(ctx context.Context, input ActionActivityInput) (ActionActivityOutput, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Starting async action activity",
		"action_id", input.ActionID,
		"action_name", input.ActionName,
		"action_type", input.ActionType,
	)

	if deps == nil || deps.AsyncRegistry == nil {
		return ActionActivityOutput{}, fmt.Errorf("async activity dependencies not initialized")
	}

	// Get task token for async completion
	activityInfo := activity.GetInfo(ctx)
	taskToken := activityInfo.TaskToken

	execCtx := ActionExecutionContext{
		ActionID:   input.ActionID.String(),
		ActionName: input.ActionName,
		RawData:    input.Context,
	}

	// Get async handler
	handler, err := deps.AsyncRegistry.GetAsyncHandler(input.ActionType)
	if err != nil {
		return ActionActivityOutput{
			ActionID:   input.ActionID,
			ActionName: input.ActionName,
			Success:    false,
		}, err
	}

	// Start the async operation (e.g., publish to RabbitMQ with taskToken as correlation ID)
	if err := handler.StartAsync(ctx, input.Config, execCtx, taskToken); err != nil {
		return ActionActivityOutput{}, err
	}

	// Return pending - activity does NOT complete yet
	// Completion will happen when external system calls CompleteActivity
	return ActionActivityOutput{}, activity.ErrResultPending
}
```

**File**: `business/sdk/workflow/temporal/async_completer.go`

```go
package temporal

import (
	"context"

	"go.temporal.io/sdk/client"
)

// AsyncCompleter completes activities from external systems (e.g., RabbitMQ consumers)
type AsyncCompleter struct {
	client client.Client
}

// NewAsyncCompleter creates a completer for async activities
func NewAsyncCompleter(c client.Client) *AsyncCompleter {
	return &AsyncCompleter{client: c}
}

// Complete finishes an async activity with a result
// Call this from your RabbitMQ consumer when a response arrives
func (c *AsyncCompleter) Complete(ctx context.Context, taskToken []byte, result ActionActivityOutput) error {
	return c.client.CompleteActivity(ctx, taskToken, result, nil)
}

// Fail fails an async activity with an error
func (c *AsyncCompleter) Fail(ctx context.Context, taskToken []byte, err error) error {
	return c.client.CompleteActivity(ctx, taskToken, nil, err)
}
```

**Example RabbitMQ Consumer Integration:**

```go
// In your RabbitMQ consumer service
func (c *Consumer) OnMessage(msg amqp.Delivery) {
	var response struct {
		TaskToken  []byte                 `json:"task_token"`
		ActionID   uuid.UUID              `json:"action_id"`
		ActionName string                 `json:"action_name"`
		Result     map[string]interface{} `json:"result"`
		Success    bool                   `json:"success"`
		Error      string                 `json:"error,omitempty"`
	}

	if err := json.Unmarshal(msg.Body, &response); err != nil {
		log.Error("Failed to unmarshal response", "error", err)
		msg.Nack(false, false)
		return
	}

	output := temporal.ActionActivityOutput{
		ActionID:   response.ActionID,
		ActionName: response.ActionName,
		Result:     response.Result,
		Success:    response.Success,
	}

	if response.Error != "" {
		c.completer.Fail(ctx, response.TaskToken, errors.New(response.Error))
	} else {
		c.completer.Complete(ctx, response.TaskToken, output)
	}

	msg.Ack(false)
}
```

---

### 5. Trigger Integration

**File**: `business/sdk/workflow/temporal/trigger.go`

```go
package temporal

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"go.temporal.io/sdk/client"

	"github.com/timmaaaz/ichor/business/domain/workflow/automationrulesbus"
	"github.com/timmaaaz/ichor/business/sdk/page"
	"github.com/timmaaaz/ichor/foundation/logger"
)

// EntityEvent represents an entity change event
type EntityEvent struct {
	EntityID   uuid.UUID
	EntityName string
	EventType  string // on_create, on_update, on_delete
	Data       map[string]interface{}
}

// EdgeStore interface for loading graph definitions from your database
type EdgeStore interface {
	QueryActionsByRule(ctx context.Context, ruleID uuid.UUID) ([]ActionNode, error)
	QueryEdgesByRule(ctx context.Context, ruleID uuid.UUID) ([]ActionEdge, error)
}

// WorkflowTrigger handles entity events and starts Temporal workflows
type WorkflowTrigger struct {
	log            *logger.Logger
	temporalClient client.Client
	rulesBus       *automationrulesbus.Business
	edgeStore      EdgeStore
}

// NewWorkflowTrigger creates a new trigger handler
func NewWorkflowTrigger(
	log *logger.Logger,
	tc client.Client,
	rb *automationrulesbus.Business,
	es EdgeStore,
) *WorkflowTrigger {
	return &WorkflowTrigger{
		log:            log,
		temporalClient: tc,
		rulesBus:       rb,
		edgeStore:      es,
	}
}

// OnEntityEvent is called when an entity event fires
// This replaces event-driven rule matching with direct Temporal workflow dispatch
func (t *WorkflowTrigger) OnEntityEvent(ctx context.Context, event EntityEvent) error {
	t.log.Info(ctx, "Processing entity event",
		"entity_name", event.EntityName,
		"event_type", event.EventType,
		"entity_id", event.EntityID,
	)

	// Find matching automation rules
	filter := automationrulesbus.QueryFilter{
		TriggerEntity: &event.EntityName,
		TriggerEvent:  &event.EventType,
		IsActive:      boolPtr(true),
	}

	rules, err := t.rulesBus.Query(ctx, filter, automationrulesbus.DefaultOrderBy, page.MustParse("1", "100"))
	if err != nil {
		return fmt.Errorf("query rules: %w", err)
	}

	t.log.Info(ctx, "Found matching rules",
		"count", len(rules),
	)

	for _, rule := range rules {
		// Load the graph definition from action_edges
		graph, err := t.loadGraphDefinition(ctx, rule.ID)
		if err != nil {
			t.log.Error(ctx, "Failed to load graph definition",
				"rule_id", rule.ID,
				"error", err,
			)
			continue
		}

		// Skip empty workflows
		if len(graph.Actions) == 0 {
			continue
		}

		// Generate unique execution ID
		executionID := uuid.New()

		// Create unique workflow ID
		workflowID := fmt.Sprintf("workflow-%s-%s-%s",
			rule.ID,
			event.EntityID,
			executionID,
		)

		// Start Temporal workflow
		workflowOptions := client.StartWorkflowOptions{
			ID:        workflowID,
			TaskQueue: TaskQueue,
		}

		we, err := t.temporalClient.ExecuteWorkflow(ctx, workflowOptions,
			ExecuteGraphWorkflow,
			WorkflowInput{
				RuleID:      rule.ID,
				ExecutionID: executionID,
				Graph:       graph,
				TriggerData: event.Data,
			},
		)

		if err != nil {
			t.log.Error(ctx, "Failed to start workflow",
				"rule_id", rule.ID,
				"workflow_id", workflowID,
				"error", err,
			)
			continue
		}

		t.log.Info(ctx, "Started workflow",
			"rule_id", rule.ID,
			"workflow_id", workflowID,
			"run_id", we.GetRunID(),
		)
	}

	return nil
}

func (t *WorkflowTrigger) loadGraphDefinition(ctx context.Context, ruleID uuid.UUID) (GraphDefinition, error) {
	// Load actions for this rule
	actions, err := t.edgeStore.QueryActionsByRule(ctx, ruleID)
	if err != nil {
		return GraphDefinition{}, fmt.Errorf("query actions: %w", err)
	}

	// Load edges for this rule
	edges, err := t.edgeStore.QueryEdgesByRule(ctx, ruleID)
	if err != nil {
		return GraphDefinition{}, fmt.Errorf("query edges: %w", err)
	}

	return GraphDefinition{
		Actions: actions,
		Edges:   edges,
	}, nil
}

func boolPtr(b bool) *bool {
	return &b
}
```

---

### 6. Edge Store Adapter

**File**: `business/sdk/workflow/temporal/stores/edgedb/edgedb.go`

```go
package edgedb

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"github.com/timmaaaz/ichor/business/sdk/workflow/temporal"
	"github.com/timmaaaz/ichor/foundation/logger"
)

// Store implements the EdgeStore interface
type Store struct {
	log *logger.Logger
	db  *sqlx.DB
}

// NewStore creates a new edge store
func NewStore(log *logger.Logger, db *sqlx.DB) *Store {
	return &Store{
		log: log,
		db:  db,
	}
}

// dbAction represents the database model for rule_actions
type dbAction struct {
	ID           uuid.UUID       `db:"id"`
	Name         string          `db:"name"`
	ActionType   string          `db:"action_type"`
	ActionConfig json.RawMessage `db:"action_config"`
}

// dbEdge represents the database model for action_edges
type dbEdge struct {
	ID             uuid.UUID  `db:"id"`
	SourceActionID *uuid.UUID `db:"source_action_id"`
	TargetActionID uuid.UUID  `db:"target_action_id"`
	EdgeType       string     `db:"edge_type"`
	SortOrder      int        `db:"sort_order"`
}

// QueryActionsByRule returns all actions for a rule
func (s *Store) QueryActionsByRule(ctx context.Context, ruleID uuid.UUID) ([]temporal.ActionNode, error) {
	const q = `
		SELECT id, name, action_type, action_config
		FROM workflow.rule_actions
		WHERE rule_id = $1
		ORDER BY sequence_order
	`

	var dbActions []dbAction
	if err := s.db.SelectContext(ctx, &dbActions, q, ruleID); err != nil {
		return nil, err
	}

	actions := make([]temporal.ActionNode, len(dbActions))
	for i, dba := range dbActions {
		actions[i] = temporal.ActionNode{
			ID:         dba.ID,
			Name:       dba.Name,
			ActionType: dba.ActionType,
			Config:     dba.ActionConfig,
		}
	}

	return actions, nil
}

// QueryEdgesByRule returns all edges for a rule
func (s *Store) QueryEdgesByRule(ctx context.Context, ruleID uuid.UUID) ([]temporal.ActionEdge, error) {
	const q = `
		SELECT id, source_action_id, target_action_id, edge_type, sort_order
		FROM workflow.action_edges
		WHERE rule_id = $1
		ORDER BY sort_order
	`

	var dbEdges []dbEdge
	if err := s.db.SelectContext(ctx, &dbEdges, q, ruleID); err != nil {
		return nil, err
	}

	edges := make([]temporal.ActionEdge, len(dbEdges))
	for i, dbe := range dbEdges {
		edges[i] = temporal.ActionEdge{
			ID:             dbe.ID,
			SourceActionID: dbe.SourceActionID,
			TargetActionID: dbe.TargetActionID,
			EdgeType:       dbe.EdgeType,
			SortOrder:      dbe.SortOrder,
		}
	}

	return edges, nil
}
```

---

### 7. Worker Entry Point

**File**: `api/cmd/services/workflow-worker/main.go`

```go
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/ardanlabs/conf/v3"
	"github.com/jmoiron/sqlx"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"

	"github.com/timmaaaz/ichor/business/sdk/sqldb"
	"github.com/timmaaaz/ichor/business/sdk/workflow/temporal"
	"github.com/timmaaaz/ichor/business/sdk/workflow/workflowactions/control"
	"github.com/timmaaaz/ichor/business/sdk/workflow/workflowactions/inventory"
	"github.com/timmaaaz/ichor/business/sdk/workflow/workflowactions/notification"
	"github.com/timmaaaz/ichor/business/sdk/workflow/workflowactions/persistence"
	"github.com/timmaaaz/ichor/foundation/logger"
)

func main() {
	log := logger.New(os.Stdout, logger.LevelInfo, "WORKFLOW-WORKER", func(context.Context) string { return "" })

	if err := run(log); err != nil {
		log.Error(context.Background(), "startup", "error", err)
		os.Exit(1)
	}
}

func run(log *logger.Logger) error {
	// Configuration
	cfg := struct {
		Temporal struct {
			HostPort  string `conf:"default:localhost:7233"`
			Namespace string `conf:"default:default"`
		}
		DB struct {
			User       string `conf:"default:postgres"`
			Password   string `conf:"default:postgres,mask"`
			Host       string `conf:"default:database-service.ichor-system.svc.cluster.local"`
			Name       string `conf:"default:postgres"`
			DisableTLS bool   `conf:"default:true"`
		}
	}{}

	if _, err := conf.Parse("ICHOR", &cfg); err != nil {
		return fmt.Errorf("parsing config: %w", err)
	}

	// Create database connection
	db, err := sqldb.Open(sqldb.Config{
		User:       cfg.DB.User,
		Password:   cfg.DB.Password,
		Host:       cfg.DB.Host,
		Name:       cfg.DB.Name,
		DisableTLS: cfg.DB.DisableTLS,
	})
	if err != nil {
		return fmt.Errorf("connecting to db: %w", err)
	}
	defer db.Close()

	// Create Temporal client
	c, err := client.Dial(client.Options{
		HostPort:  cfg.Temporal.HostPort,
		Namespace: cfg.Temporal.Namespace,
		Logger:    newTemporalLogger(log),
	})
	if err != nil {
		return fmt.Errorf("creating temporal client: %w", err)
	}
	defer c.Close()

	// Initialize action registry with all handlers
	actionRegistry := buildActionRegistry(log, db)

	// Set activity dependencies
	temporal.SetActivityDependencies(&temporal.ActivityDependencies{
		ActionRegistry: actionRegistry,
	})

	// Create worker
	w := worker.New(c, temporal.TaskQueue, worker.Options{
		MaxConcurrentActivityExecutionSize:     100,
		MaxConcurrentWorkflowTaskExecutionSize: 100,
	})

	// Register workflows
	w.RegisterWorkflow(temporal.ExecuteGraphWorkflow)
	w.RegisterWorkflow(temporal.ExecuteBranchUntilConvergence)

	// Register activities
	w.RegisterActivity(temporal.ExecuteActionActivity)

	log.Info(context.Background(), "Starting workflow worker",
		"task_queue", temporal.TaskQueue,
		"temporal_host", cfg.Temporal.HostPort,
	)

	// Handle shutdown
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Run worker
	errCh := make(chan error, 1)
	go func() {
		errCh <- w.Run(worker.InterruptCh())
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		log.Info(context.Background(), "Shutting down worker")
		return nil
	}
}

func buildActionRegistry(log *logger.Logger, db *sqlx.DB) *temporal.ActionRegistry {
	registry := temporal.NewActionRegistry()

	// Register all action handlers
	// These wrap your existing handlers

	// Control flow actions
	registry.Register("evaluate_condition", control.NewEvaluateConditionHandler(log))

	// Inventory actions
	registry.Register("allocate_inventory", inventory.NewAllocateInventoryHandler(log, db))

	// Notification actions
	registry.Register("send_email", notification.NewSendEmailHandler(log))
	registry.Register("create_alert", notification.NewCreateAlertHandler(log, db))

	// Persistence actions
	registry.Register("update_field", persistence.NewUpdateFieldHandler(log, db))
	registry.Register("create_record", persistence.NewCreateRecordHandler(log, db))

	// Add more handlers as needed...

	return registry
}

// temporalLogger adapts our logger to Temporal's logger interface
type temporalLogger struct {
	log *logger.Logger
}

func newTemporalLogger(log *logger.Logger) *temporalLogger {
	return &temporalLogger{log: log}
}

func (l *temporalLogger) Debug(msg string, keyvals ...interface{}) {
	l.log.Debug(context.Background(), msg, keyvals...)
}

func (l *temporalLogger) Info(msg string, keyvals ...interface{}) {
	l.log.Info(context.Background(), msg, keyvals...)
}

func (l *temporalLogger) Warn(msg string, keyvals ...interface{}) {
	l.log.Warn(context.Background(), msg, keyvals...)
}

func (l *temporalLogger) Error(msg string, keyvals ...interface{}) {
	l.log.Error(context.Background(), msg, keyvals...)
}
```

---

## Infrastructure

### Docker Compose (Development)

**File**: `zarf/compose/docker-compose-temporal.yml`

```yaml
version: "3.5"
services:
  temporal:
    image: temporalio/auto-setup:1.22
    ports:
      - "7233:7233"
    environment:
      - DB=postgresql
      - DB_PORT=5432
      - POSTGRES_USER=temporal
      - POSTGRES_PWD=temporal
      - POSTGRES_SEEDS=temporal-postgresql
    depends_on:
      - temporal-postgresql

  temporal-postgresql:
    image: postgres:16
    environment:
      POSTGRES_USER: temporal
      POSTGRES_PASSWORD: temporal
    ports:
      - "5433:5432"

  temporal-ui:
    image: temporalio/ui:2.21
    ports:
      - "8080:8080"
    environment:
      - TEMPORAL_ADDRESS=temporal:7233
    depends_on:
      - temporal

  workflow-worker:
    build:
      context: ../..
      dockerfile: zarf/docker/dockerfile.workflow-worker
    environment:
      - ICHOR_TEMPORAL_HOSTPORT=temporal:7233
      - ICHOR_DB_HOST=database:5432
    depends_on:
      - temporal
      - database
```

### Dockerfile

**File**: `zarf/docker/dockerfile.workflow-worker`

```dockerfile
FROM golang:1.23 AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o workflow-worker ./api/cmd/services/workflow-worker

FROM alpine:3.19
RUN apk --no-cache add ca-certificates
WORKDIR /app
COPY --from=builder /app/workflow-worker .
CMD ["./workflow-worker"]
```

### Kubernetes Deployment

**File**: `zarf/k8s/dev/workflow-worker/deployment.yaml`

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: workflow-worker
  namespace: ichor-system
spec:
  replicas: 2
  selector:
    matchLabels:
      app: workflow-worker
  template:
    metadata:
      labels:
        app: workflow-worker
    spec:
      containers:
        - name: workflow-worker
          image: ichor/workflow-worker:latest
          env:
            - name: ICHOR_TEMPORAL_HOSTPORT
              value: "temporal.temporal-system.svc.cluster.local:7233"
            - name: ICHOR_DB_HOST
              value: "database-service.ichor-system.svc.cluster.local:5432"
            - name: ICHOR_DB_USER
              valueFrom:
                secretKeyRef:
                  name: db-credentials
                  key: username
            - name: ICHOR_DB_PASSWORD
              valueFrom:
                secretKeyRef:
                  name: db-credentials
                  key: password
          resources:
            requests:
              cpu: "100m"
              memory: "128Mi"
            limits:
              cpu: "500m"
              memory: "512Mi"
```

---

## Makefile Targets

Add to `Makefile`:

```makefile
# =============================================================================
# Workflow Worker
# =============================================================================

workflow-worker-build:
	docker build \
		-f zarf/docker/dockerfile.workflow-worker \
		-t ichor/workflow-worker:latest \
		.

workflow-worker-run:
	go run ./api/cmd/services/workflow-worker

# =============================================================================
# Temporal
# =============================================================================

temporal-up:
	docker-compose -f zarf/compose/docker-compose-temporal.yml up -d

temporal-down:
	docker-compose -f zarf/compose/docker-compose-temporal.yml down

temporal-ui:
	@echo "Temporal UI: http://localhost:8080"
	open http://localhost:8080
```

---

## Implementation Phases

### Phase 1: Infrastructure Setup & Evaluation
1. **Evaluate temporalgraph library** - Review https://github.com/saltosystems/temporalgraph
   - Does it support your edge types (start, sequence, true_branch, false_branch, always, parallel)?
   - Does it handle convergence detection?
   - Can it work with runtime-defined graphs (not compile-time)?
   - Decision: Use temporalgraph OR proceed with custom implementation
2. Add Temporal Docker Compose configuration
3. Add workflow-worker Dockerfile
4. Add Makefile targets
5. Verify Temporal cluster starts correctly

### Phase 2: Core Implementation
1. Create `temporal/models.go` with context size limits
2. Create `temporal/graph_executor.go` with deterministic map iteration
3. Create `temporal/workflow.go` with Continue-As-New and versioning
4. Create `temporal/activities.go` and `activities_async.go`
5. Create `temporal/async_completer.go` for RabbitMQ integration
6. Write unit tests for graph executor (test determinism!)
7. **Determinism audit**: Review all map iterations, ensure sorted

### Phase 3: Integration
1. Create `temporal/trigger.go`
2. Create `temporal/stores/edgedb/edgedb.go`
3. Create `api/cmd/services/workflow-worker/main.go`
4. Wire action handlers into registry

### Phase 4: Testing
1. Create integration tests for workflow execution
2. Test parallel branch execution
3. Test convergence detection
4. Test async action handling (with mock RabbitMQ)
5. **Replay testing**: Record workflow history, replay with same input, verify identical commands
6. **Determinism tests**: Run same graph multiple times, verify convergence point selection is consistent
7. Test Continue-As-New triggers correctly at history threshold
8. Test context size truncation for large action results

### Phase 5: Deployment
1. Add Kubernetes manifests
2. Configure Temporal for production
3. Set up monitoring dashboards
4. Document operational procedures

---

## Files Summary

### New Files
- `business/sdk/workflow/temporal/models.go`
- `business/sdk/workflow/temporal/graph_executor.go`
- `business/sdk/workflow/temporal/workflow.go`
- `business/sdk/workflow/temporal/activities.go`
- `business/sdk/workflow/temporal/trigger.go`
- `business/sdk/workflow/temporal/stores/edgedb/edgedb.go`
- `api/cmd/services/workflow-worker/main.go`
- `zarf/compose/docker-compose-temporal.yml`
- `zarf/docker/dockerfile.workflow-worker`
- `zarf/k8s/dev/workflow-worker/deployment.yaml`

### Modified Files
- `Makefile` - Add workflow-worker and temporal targets
- `api/cmd/services/ichor/build/all/all.go` - Initialize Temporal client for trigger

---

## Key Design Decisions

| Decision | Rationale |
|----------|-----------|
| Generic interpreter workflow | Visual editor produces data, not code; one workflow interprets any graph |
| Child workflows for parallel branches | Cleaner than goroutines for long-running branches with convergence |
| Action registry pattern | Decouples Temporal from your action handler implementations |
| EdgeStore interface | Abstracts database access for testing and flexibility |
| Separate worker process | Can scale independently of main API service |
| Activity wraps handlers | Minimal changes to existing handler code |
| Sorted map iteration | **Required for determinism** - Go maps iterate randomly, breaking replay |
| Continue-As-New at 10K events | Prevents hitting Temporal's 50K event limit on long workflows |
| Context size truncation | Prevents exceeding Temporal's 2MB payload limit |
| Async activity completion | Proper pattern for RabbitMQ; avoids blocking workers |
| Workflow versioning | Enables safe deployments with in-flight workflows |

---

## Payload Size Limits (From Temporal Research)

| Limit | Threshold | Mitigation |
|-------|-----------|------------|
| Single payload | 2 MB hard limit (warning at 256 KB) | Truncate large action results |
| gRPC message | 4 MB | Keep merged context lean |
| Event history | 50 MB or 51,200 events | Use Continue-As-New |

**Best Practices:**
- Store only what's needed for subsequent actions in context
- Offload large data to external storage (S3, database), pass references
- Monitor context size growth in production

---

## Debugging & Visualization

### Temporal Web UI
- Shows workflow execution history
- Visualizes activity calls, timing, state
- Access at http://localhost:8080 (dev) via `make temporal-ui`

### Graph Visualization (Optional Enhancement)
Add a debug endpoint to visualize workflow graphs:

```go
// GET /api/v1/workflow/rules/{rule_id}/graph.mermaid
func (api *api) getGraphMermaid(ctx context.Context, r *http.Request) web.Encoder {
    graph, _ := api.edgeStore.LoadGraph(ctx, ruleID)
    return MermaidResponse{Content: graph.ToMermaid()}
}

func (g GraphDefinition) ToMermaid() string {
    var sb strings.Builder
    sb.WriteString("graph TD\n")
    for _, edge := range g.Edges {
        if edge.SourceActionID == nil {
            sb.WriteString(fmt.Sprintf("    START --> %s\n", edge.TargetActionID))
        } else {
            sb.WriteString(fmt.Sprintf("    %s -->|%s| %s\n",
                *edge.SourceActionID, edge.EdgeType, edge.TargetActionID))
        }
    }
    return sb.String()
}
```
