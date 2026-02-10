package temporal

import (
	"context"
	"fmt"
	"maps"

	"github.com/google/uuid"
	"go.temporal.io/sdk/client"

	"github.com/timmaaaz/ichor/business/sdk/workflow"
	"github.com/timmaaaz/ichor/foundation/logger"
)

// =============================================================================
// Interfaces
// =============================================================================

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
	log         *logger.Logger
	starter     WorkflowStarter
	ruleMatcher RuleMatcher
	edgeStore   EdgeStore
}

// NewWorkflowTrigger creates a new trigger handler.
// The starter parameter accepts client.Client (which satisfies WorkflowStarter).
func NewWorkflowTrigger(
	log *logger.Logger,
	starter WorkflowStarter,
	rm RuleMatcher,
	es EdgeStore,
) *WorkflowTrigger {
	return &WorkflowTrigger{
		log:         log,
		starter:     starter,
		ruleMatcher: rm,
		edgeStore:   es,
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

	// Start a Temporal workflow for each matched rule.
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
			// Continue to next rule - don't fail the entire event.
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
	// Load graph definition from database.
	graph, err := t.loadGraphDefinition(ctx, rm.Rule.ID)
	if err != nil {
		return fmt.Errorf("load graph for rule %s: %w", rm.Rule.ID, err)
	}

	// Skip rules with empty graphs (no actions configured).
	if len(graph.Actions) == 0 {
		t.log.Warn(ctx, "Skipping rule with empty graph",
			"rule_id", rm.Rule.ID,
			"rule_name", rm.Rule.Name,
		)
		return nil
	}

	// Generate unique execution ID.
	executionID := uuid.New()

	// Deterministic workflow ID for deduplication and traceability.
	// Format: workflow-{ruleID}-{entityID}-{executionID}
	workflowID := fmt.Sprintf("workflow-%s-%s-%s",
		rm.Rule.ID,
		event.EntityID,
		executionID,
	)

	// Build workflow input with trigger data from the event.
	input := WorkflowInput{
		RuleID:      rm.Rule.ID,
		RuleName:    rm.Rule.Name,
		ExecutionID: executionID,
		Graph:       graph,
		TriggerData: buildTriggerData(event),
	}

	// Start Temporal workflow.
	workflowOptions := client.StartWorkflowOptions{
		ID:        workflowID,
		TaskQueue: TaskQueue,
	}

	we, err := t.starter.ExecuteWorkflow(ctx, workflowOptions,
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
