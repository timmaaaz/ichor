package temporal

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"time"

	"github.com/google/uuid"
	enumspb "go.temporal.io/api/enums/v1"
	"go.temporal.io/api/serviceerror"
	"go.temporal.io/sdk/client"

	"github.com/timmaaaz/ichor/business/sdk/workflow"
	"github.com/timmaaaz/ichor/foundation/logger"
)

// =============================================================================
// Interfaces
// =============================================================================

// ExecutionStore persists workflow execution records.
// Implemented by stores/workflowdb.Store.
type ExecutionStore interface {
	CreateExecution(ctx context.Context, exec workflow.AutomationExecution) error

	// DeleteExecution removes an execution record by id. The trigger writes the record
	// before starting the workflow (so activities can reference its FK); if the start
	// fails or is rejected as a duplicate, the orphaned record is reclaimed via this so
	// execution-record dedup is record-level, not just run-level.
	DeleteExecution(ctx context.Context, id uuid.UUID) error

	// QueryExecutionByID loads a single execution record by id. Used by RerunExecution
	// to recover an execution's originating rule + persisted trigger_data so the rule
	// can be re-fired against a freshly reconstructed event. Backed by workflowdb.Store
	// (it also backs workflow.Business.QueryExecutionByID).
	QueryExecutionByID(ctx context.Context, id uuid.UUID) (workflow.AutomationExecution, error)
}

// EdgeStore loads graph definitions (actions + edges) from the database.
// Implemented by stores/edgedb.Store in Phase 8.
type EdgeStore interface {
	// QueryActionsByRule returns all action nodes for a given automation rule.
	QueryActionsByRule(ctx context.Context, ruleID uuid.UUID) ([]ActionNode, error)

	// QueryEdgesByRule returns all action edges for a given automation rule.
	QueryEdgesByRule(ctx context.Context, ruleID uuid.UUID) ([]ActionEdge, error)
}

// RuleMatcher matches entity events against automation rules.
// Implemented by workflow.TriggerProcessor.
type RuleMatcher interface {
	ProcessEvent(ctx context.Context, event workflow.TriggerEvent) (*workflow.ProcessingResult, error)
}

// WorkflowStarter starts Temporal workflows. This narrow interface
// is satisfied by client.Client and enables testing without mocking
// the full Temporal Client interface (50+ methods).
type WorkflowStarter interface {
	ExecuteWorkflow(ctx context.Context, options client.StartWorkflowOptions, workflow any, args ...any) (client.WorkflowRun, error)
}

// =============================================================================
// WorkflowTrigger
// =============================================================================

// WorkflowTrigger handles entity events and dispatches Temporal workflows.
//
// This replaces the existing Engine.ExecuteWorkflow() entry point.
// Rule matching is delegated to the RuleMatcher interface (workflow.TriggerProcessor).
// Graph loading is delegated to the EdgeStore interface.
// Workflow execution is delegated to Temporal via WorkflowStarter.ExecuteWorkflow.
type WorkflowTrigger struct {
	log            *logger.Logger
	starter        WorkflowStarter
	ruleMatcher    RuleMatcher
	edgeStore      EdgeStore
	executionStore ExecutionStore
	taskQueue      string
}

// NewWorkflowTrigger creates a new trigger handler.
// The starter parameter accepts client.Client (which satisfies WorkflowStarter).
// Workflows are dispatched to the default TaskQueue constant. Use WithTaskQueue
// to override (e.g., in tests where each test has a unique worker task queue).
func NewWorkflowTrigger(
	log *logger.Logger,
	starter WorkflowStarter,
	rm RuleMatcher,
	es EdgeStore,
	exs ExecutionStore,
) *WorkflowTrigger {
	return &WorkflowTrigger{
		log:            log,
		starter:        starter,
		ruleMatcher:    rm,
		edgeStore:      es,
		executionStore: exs,
		taskQueue:      TaskQueue,
	}
}

// WithTaskQueue overrides the Temporal task queue used when dispatching workflows.
// Returns the trigger for chaining. Primarily used in tests to route workflows
// to the per-test worker task queue instead of the global production queue.
func (t *WorkflowTrigger) WithTaskQueue(tq string) *WorkflowTrigger {
	t.taskQueue = tq
	return t
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

	// Match automation rules using RuleMatcher (TriggerProcessor).
	result, err := t.ruleMatcher.ProcessEvent(ctx, event)
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

	// Cascade lineage carried from the originating write (empty for a human /
	// non-workflow write, which starts a fresh chain). See lineage.go.
	parentLineage := lineageFromContext(ctx)

	// Start a Temporal workflow for each matched rule.
	for _, rm := range result.MatchedRules {
		if !rm.Matched {
			continue
		}

		// Runtime loop guard (P1): refuse to dispatch a (rule, entity) that has
		// already fired in this cascade chain. This is the universal backstop that
		// stops A->B->A loops; it keys on the (rule, entity) pair, so the same rule
		// firing for a *different* entity is allowed (e.g. a rule processing a batch).
		if parentLineage.Contains(rm.Rule.ID, event.EntityID) {
			t.log.Warn(ctx, "cascade loop prevented: rule already fired for entity in this chain",
				"rule_id", rm.Rule.ID,
				"rule_name", rm.Rule.Name,
				"entity_name", event.EntityName,
				"entity_id", event.EntityID,
			)
			continue
		}

		// Seed the next generation: parent set extended with this (rule, entity).
		childLineage := parentLineage.With(rm.Rule.ID, event.EntityID)

		if _, err := t.startWorkflowForRule(ctx, event, rm, childLineage); err != nil {
			t.log.Error(ctx, "Failed to start workflow for rule",
				"rule_id", rm.Rule.ID,
				"rule_name", rm.Rule.Name,
				"error", err,
			)
			// Continue to next rule - don't fail the entire event.
			continue
		}
	}

	return nil
}

// ErrExecutionNotRerunnable is returned when an execution has no automation
// rule to re-fire (e.g. a manual execution).
var ErrExecutionNotRerunnable = errors.New("execution has no automation rule to re-run")

// reconstructTriggerEvent reverses buildTriggerData from a persisted execution's
// trigger_data. EventID is left zero so the dispatch mints a fresh dedup key
// (clearing the Temporal workflow-id REJECT_DUPLICATE wall), and Timestamp is
// re-stamped (the stored value is a time.Time.String() rendering, not RFC3339,
// so it cannot be round-tripped into a time.Time — re-stamping is correct for a
// replay anyway).
func reconstructTriggerEvent(triggerData json.RawMessage) (workflow.TriggerEvent, error) {
	var m map[string]any
	if err := json.Unmarshal(triggerData, &m); err != nil {
		return workflow.TriggerEvent{}, fmt.Errorf("parse trigger_data: %w", err)
	}

	ev := workflow.TriggerEvent{Timestamp: time.Now()}
	if s, ok := m["event_type"].(string); ok {
		ev.EventType = s
	}
	if s, ok := m["entity_name"].(string); ok {
		ev.EntityName = s
	}
	if s, ok := m["entity_id"].(string); ok {
		if id, err := uuid.Parse(s); err == nil {
			ev.EntityID = id
		}
	}
	if s, ok := m["user_id"].(string); ok {
		if id, err := uuid.Parse(s); err == nil {
			ev.UserID = id
		}
	}
	if fcRaw, ok := m["field_changes"].(map[string]any); ok {
		fc := make(map[string]workflow.FieldChange, len(fcRaw))
		for field, v := range fcRaw {
			if inner, ok := v.(map[string]any); ok {
				fc[field] = workflow.FieldChange{OldValue: inner["old_value"], NewValue: inner["new_value"]}
			}
		}
		if len(fc) > 0 {
			ev.FieldChanges = fc
		}
	}

	// RawData = everything except the metadata + cascade-lineage keys.
	raw := make(map[string]any, len(m))
	for k, v := range m {
		switch k {
		case "event_type", "entity_name", "entity_id", "user_id", "timestamp", "field_changes", CascadeLineageKey:
			continue
		}
		raw[k] = v
	}
	if len(raw) > 0 {
		ev.RawData = raw
	}
	return ev, nil
}

// RerunExecution re-fires the single rule that produced executionID against a
// fresh event reconstructed from its persisted trigger_data, minting a brand-new
// execution id. The fresh id clears all three dedup walls: the allocation_results
// idempotency key, the Temporal workflow-id REJECT_DUPLICATE guard (workflow id is
// keyed on a fresh dedup key because the reconstructed event has a zero EventID),
// and the execution-record upsert. Returns the new execution id.
//
// A fresh cascade lineage (empty WorkflowLineage) is seeded so the replay starts a
// clean chain rooted at the new execution. Returns ErrExecutionNotRerunnable when
// the execution has no originating rule (e.g. a manual execution).
func (t *WorkflowTrigger) RerunExecution(ctx context.Context, executionID uuid.UUID) (uuid.UUID, error) {
	exec, err := t.executionStore.QueryExecutionByID(ctx, executionID)
	if err != nil {
		return uuid.Nil, fmt.Errorf("load execution %s: %w", executionID, err)
	}
	if exec.AutomationRuleID == nil {
		return uuid.Nil, ErrExecutionNotRerunnable
	}

	event, err := reconstructTriggerEvent(exec.TriggerData)
	if err != nil {
		return uuid.Nil, fmt.Errorf("reconstruct event: %w", err)
	}

	rm := workflow.RuleMatchResult{
		Rule: workflow.AutomationRuleView{ID: *exec.AutomationRuleID, Name: exec.RuleName},
	}

	// startWorkflowForRule mints a fresh executionID and returns it. A fresh
	// (empty) lineage seeds a clean cascade chain rooted at the new execution.
	newID, err := t.startWorkflowForRule(ctx, event, rm, WorkflowLineage{})
	if err != nil {
		return uuid.Nil, fmt.Errorf("dispatch rerun: %w", err)
	}
	return newID, nil
}

// startWorkflowForRule loads the graph definition and starts a Temporal workflow
// for a single matched rule.
//
// On success it returns the freshly minted execution id (uuid.New()), which lets
// RerunExecution surface the new id to its caller; OnEntityEvent ignores it.
func (t *WorkflowTrigger) startWorkflowForRule(
	ctx context.Context,
	event workflow.TriggerEvent,
	rm workflow.RuleMatchResult,
	lineage WorkflowLineage,
) (uuid.UUID, error) {
	// Load graph definition from database.
	graph, err := t.loadGraphDefinition(ctx, rm.Rule.ID)
	if err != nil {
		return uuid.Nil, fmt.Errorf("load graph for rule %s: %w", rm.Rule.ID, err)
	}

	// Skip rules with empty graphs (no actions configured).
	if len(graph.Actions) == 0 {
		t.log.Warn(ctx, "Skipping rule with empty graph",
			"rule_id", rm.Rule.ID,
			"rule_name", rm.Rule.Name,
		)
		return uuid.Nil, nil
	}

	// Generate unique execution ID.
	executionID := uuid.New()

	// Stamp the chain root once (first dispatched hop), then preserve it.
	if lineage.OriginatingExecutionID == uuid.Nil {
		lineage.OriginatingExecutionID = executionID
	}

	// Deterministic, dedup-keyed workflow ID. Format: workflow-{ruleID}-{dedupKey}.
	//
	// When the event was drained from the outbox relay (F2), event.EventID is the
	// durable outbox row id, so the id is stable across an at-least-once re-publish of
	// that row; the REJECT_DUPLICATE reuse policy below then collapses the retry into
	// a single execution (at-least-once emission -> effectively-once execution).
	//
	// Direct OnEntityEvent callers (human writes pre-cutover, tests) carry a zero
	// EventID and fall back to the per-dispatch random executionID, preserving the
	// pre-F2 behavior of a unique id per dispatch.
	dedupKey := event.EventID
	if dedupKey == uuid.Nil {
		dedupKey = executionID
	}
	workflowID := fmt.Sprintf("workflow-%s-%s", rm.Rule.ID, dedupKey)

	// Build trigger data map (reused for both the execution record and workflow input).
	triggerData := buildTriggerData(event)

	// Carry the seeded cascade lineage onto the workflow input (rides TriggerData,
	// Continue-As-New-safe). It flows into the action activity context, then through
	// the handler's bus write / delegate.Call to seed the next cascade generation
	// (P1 runtime loop guard — see lineage.go). It is also persisted on the execution
	// record's TriggerData JSON below, giving each execution free cascade provenance.
	triggerData[CascadeLineageKey] = lineage

	// Persist the execution record before starting the workflow so that
	// activities which write to approval_requests (FK → automation_executions)
	// can reference the row immediately.
	triggerDataJSON, err := json.Marshal(triggerData)
	if err != nil {
		return uuid.Nil, fmt.Errorf("marshal trigger data: %w", err)
	}

	if err := t.executionStore.CreateExecution(ctx, workflow.AutomationExecution{
		ID:               executionID,
		AutomationRuleID: &rm.Rule.ID,
		EntityType:       event.EntityName,
		TriggerData:      triggerDataJSON,
		Status:           workflow.StatusPending,
		TriggerSource:    workflow.TriggerSourceAutomation,
	}); err != nil {
		return uuid.Nil, fmt.Errorf("creating execution record: %w", err)
	}

	// If ExecuteWorkflow below fails — or is rejected as a duplicate (a relay re-delivery
	// of the same outbox row, surfaced as an error by WorkflowExecutionErrorWhenAlreadyStarted
	// + REJECT_DUPLICATE) — the record created above is an orphan at StatusPending with no
	// corresponding workflow run. The error path reclaims it via DeleteExecution, making dedup
	// RECORD-level, not just run-level (Temporal already dedups the workflow run itself).

	// Build workflow input with trigger data from the event.
	input := WorkflowInput{
		RuleID:      rm.Rule.ID,
		RuleName:    rm.Rule.Name,
		ExecutionID: executionID,
		Graph:       graph,
		TriggerData: triggerData,
	}

	// Start Temporal workflow. REJECT_DUPLICATE makes a re-dispatch of the same
	// outbox event (same workflow id) a no-op at the server, giving effectively-once
	// execution atop the relay's at-least-once delivery (DESIGN §5).
	workflowOptions := client.StartWorkflowOptions{
		ID:                    workflowID,
		TaskQueue:             t.taskQueue,
		WorkflowIDReusePolicy: enumspb.WORKFLOW_ID_REUSE_POLICY_REJECT_DUPLICATE,
		// Surface a duplicate start as an error instead of the SDK default (which silently
		// returns a handle to the existing run). The trigger needs that signal to reclaim the
		// execution record it wrote before the start; without it a relay re-delivery leaves a
		// second orphaned StatusPending row. Run-level dedup is unaffected — no second workflow
		// ever runs; only the client now learns of the rejection.
		WorkflowExecutionErrorWhenAlreadyStarted: true,
	}

	we, err := t.starter.ExecuteWorkflow(ctx, workflowOptions,
		ExecuteGraphWorkflow,
		input,
	)
	if err != nil {
		// The start produced no new workflow run, so the record written above is an orphan.
		// Reclaim it (best-effort) — this is what makes execution-record dedup record-level.
		if delErr := t.executionStore.DeleteExecution(ctx, executionID); delErr != nil {
			t.log.Error(ctx, "failed to delete orphaned execution record after workflow start error",
				"execution_id", executionID,
				"workflow_id", workflowID,
				"delete_error", delErr,
			)
		}

		// A duplicate rejection is an expected, benign re-delivery: the original workflow is
		// already running (or already ran), so once the orphan record is reclaimed this is an
		// idempotent no-op. Any other error is a real dispatch failure.
		var alreadyStarted *serviceerror.WorkflowExecutionAlreadyStarted
		if errors.As(err, &alreadyStarted) {
			t.log.Warn(ctx, "duplicate workflow start rejected (re-delivery); orphaned execution record reclaimed",
				"rule_id", rm.Rule.ID,
				"workflow_id", workflowID,
			)
			return uuid.Nil, nil
		}
		return uuid.Nil, fmt.Errorf("execute workflow: %w", err)
	}

	t.log.Info(ctx, "Started Temporal workflow",
		"rule_id", rm.Rule.ID,
		"rule_name", rm.Rule.Name,
		"workflow_id", workflowID,
		"run_id", we.GetRunID(),
	)

	return executionID, nil
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

	// Copy raw entity data first so metadata keys take precedence
	// if RawData contains conflicting keys (e.g., "event_type").
	maps.Copy(data, event.RawData)

	// Include event metadata (overwrites any conflicting RawData keys).
	data["event_type"] = event.EventType
	data["entity_name"] = event.EntityName
	data["entity_id"] = event.EntityID.String()
	data["user_id"] = event.UserID.String()
	data["timestamp"] = event.Timestamp.String()

	// Include field changes for update events.
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
