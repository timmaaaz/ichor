package temporal

import (
	"encoding/json"
	"errors"
	"fmt"
	"maps"

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
// Edge Type Constants
// =============================================================================

const (
	// EdgeTypeStart marks an entry point into the graph (SourceActionID is nil).
	EdgeTypeStart = "start"

	// EdgeTypeSequence connects actions in a linear chain.
	EdgeTypeSequence = "sequence"

	// EdgeTypeTrueBranch follows when a condition evaluates to true.
	EdgeTypeTrueBranch = "true_branch"

	// EdgeTypeFalseBranch follows when a condition evaluates to false.
	EdgeTypeFalseBranch = "false_branch"

	// EdgeTypeAlways follows regardless of the source action's result.
	EdgeTypeAlways = "always"
)

// =============================================================================
// Graph Definition Types
// =============================================================================

// WorkflowInput is passed when starting a workflow execution via Temporal.
//
// ContinuationState preserves the full MergedContext across Continue-As-New
// boundaries. On initial execution it is nil; after Continue-As-New it carries
// the accumulated ActionResults, TriggerData, and Flattened maps so no
// structural information is lost.
type WorkflowInput struct {
	RuleID            uuid.UUID       `json:"rule_id"`
	RuleName          string          `json:"rule_name"`
	ExecutionID       uuid.UUID       `json:"execution_id"`
	Graph             GraphDefinition `json:"graph"`
	TriggerData       map[string]any  `json:"trigger_data"`
	ContinuationState *MergedContext  `json:"continuation_state,omitempty"`
}

// Validate checks that WorkflowInput has the required fields for execution.
func (wi WorkflowInput) Validate() error {
	if wi.RuleID == uuid.Nil {
		return errors.New("rule_id is required")
	}
	if wi.ExecutionID == uuid.Nil {
		return errors.New("execution_id is required")
	}
	if len(wi.Graph.Actions) == 0 {
		return errors.New("graph must contain at least one action")
	}

	hasStartEdge := false
	for _, edge := range wi.Graph.Edges {
		if edge.EdgeType == EdgeTypeStart {
			hasStartEdge = true
			break
		}
	}
	if !hasStartEdge {
		return errors.New("graph must contain at least one start edge")
	}

	return nil
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
	DeactivatedBy uuid.UUID       `json:"deactivated_by"` // uuid.Nil if not deactivated
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
	TriggerData   map[string]any            `json:"trigger_data"`
	ActionResults map[string]map[string]any `json:"action_results"` // action_name -> result
	Flattened     map[string]any            `json:"flattened"`      // For template resolution
}

// NewMergedContext creates a context initialized with trigger data.
// Trigger data is copied to the Flattened map for immediate template access.
func NewMergedContext(triggerData map[string]any) *MergedContext {
	ctx := &MergedContext{
		TriggerData:   triggerData,
		ActionResults: make(map[string]map[string]any),
		Flattened:     make(map[string]any),
	}

	maps.Copy(ctx.Flattened, triggerData)

	return ctx
}

// MergeResult adds an action's result to the context.
// Large values are sanitized (truncated) to prevent exceeding Temporal's 2MB payload limit.
// Results are indexed both by action name (for ActionResults) and as
// flattened "action_name.field" keys (for template resolution).
func (c *MergedContext) MergeResult(actionName string, result map[string]any) {
	if c.ActionResults == nil {
		c.ActionResults = make(map[string]map[string]any)
	}
	if c.Flattened == nil {
		c.Flattened = make(map[string]any)
	}

	sanitized, _ := sanitizeResult(result)
	// NOTE: wasTruncated bool available for future logging in Phase 6
	// when workflow.GetLogger() is available in activities.

	c.ActionResults[actionName] = sanitized

	for k, v := range sanitized {
		c.Flattened[actionName+"."+k] = v
	}

	c.Flattened[actionName] = sanitized
}

// Clone creates a 2-level deep copy for parallel branch execution.
// TriggerData, ActionResults, and Flattened maps are cloned, and each
// action's result map is copied. However, values WITHIN result maps
// (e.g., a nested map[string]any) are shared references.
// This is acceptable because action results are treated as immutable
// after MergeResult - activities produce results, MergeResult stores
// them, and nothing mutates individual result values afterward.
func (c *MergedContext) Clone() *MergedContext {
	clone := &MergedContext{}

	if c.TriggerData != nil {
		clone.TriggerData = make(map[string]any, len(c.TriggerData))
		maps.Copy(clone.TriggerData, c.TriggerData)
	}

	if c.ActionResults != nil {
		clone.ActionResults = make(map[string]map[string]any, len(c.ActionResults))
		for k, v := range c.ActionResults {
			if v != nil {
				resultCopy := make(map[string]any, len(v))
				maps.Copy(resultCopy, v)
				clone.ActionResults[k] = resultCopy
			}
		}
	}

	if c.Flattened != nil {
		clone.Flattened = make(map[string]any, len(c.Flattened))
		maps.Copy(clone.Flattened, c.Flattened)
	}

	return clone
}

// sanitizeResult truncates large values to prevent payload size issues.
// Returns a new map and a bool indicating whether any values were truncated.
// The input map is never mutated.
// String values > MaxResultValueSize are truncated with a "[TRUNCATED]" marker.
// Binary data > MaxResultValueSize is replaced entirely.
// Complex objects > MaxResultValueSize (when serialized) are replaced.
// Nil values are preserved as-is.
// Primitive types (bool, int, float) bypass JSON marshaling (known small).
func sanitizeResult(result map[string]any) (map[string]any, bool) {
	if result == nil {
		return make(map[string]any), false
	}

	sanitized := make(map[string]any, len(result))
	wasTruncated := false

	for k, v := range result {
		if v == nil {
			sanitized[k] = nil
			continue
		}

		switch val := v.(type) {
		case string:
			if len(val) > MaxResultValueSize {
				sanitized[k] = val[:MaxResultValueSize] + "...[TRUNCATED]"
				sanitized[k+"_truncated"] = true
				wasTruncated = true
			} else {
				sanitized[k] = val
			}
		case []byte:
			if len(val) > MaxResultValueSize {
				sanitized[k] = "[BINARY_DATA_TRUNCATED]"
				sanitized[k+"_truncated"] = true
				wasTruncated = true
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
				wasTruncated = true
			} else if len(data) > MaxResultValueSize {
				sanitized[k] = "[LARGE_OBJECT_TRUNCATED]"
				sanitized[k+"_truncated"] = true
				wasTruncated = true
			} else {
				sanitized[k] = val
			}
		}
	}

	return sanitized, wasTruncated
}

// =============================================================================
// Parallel Execution Types
// =============================================================================

// BranchInput is passed to child workflows for parallel branch execution.
// Each parallel branch runs as a separate Temporal child workflow.
//
// ConvergencePoint is the action ID where this branch should stop execution.
// For fire-and-forget branches (no convergence), set to uuid.Nil - the branch
// executes until there are no more next actions.
type BranchInput struct {
	StartAction      ActionNode      `json:"start_action"`
	ConvergencePoint uuid.UUID       `json:"convergence_point"` // uuid.Nil for fire-and-forget
	Graph            GraphDefinition `json:"graph"`
	InitialContext   *MergedContext  `json:"initial_context"`
	RuleID           uuid.UUID       `json:"rule_id"`
	ExecutionID      uuid.UUID       `json:"execution_id"`
	RuleName         string          `json:"rule_name"`
}

// Validate checks that BranchInput has the required fields for execution.
// ConvergencePoint may be uuid.Nil for fire-and-forget branches.
func (bi BranchInput) Validate() error {
	if bi.StartAction.ID == uuid.Nil {
		return errors.New("start_action.id is required")
	}
	if len(bi.Graph.Actions) == 0 {
		return errors.New("graph must contain at least one action")
	}
	if bi.InitialContext == nil {
		return errors.New("initial_context is required")
	}
	return nil
}

// BranchOutput is returned from child workflows.
// Contains all action results accumulated during the branch execution.
type BranchOutput struct {
	ActionResults map[string]map[string]any `json:"action_results"`
}

// =============================================================================
// Activity Types
// =============================================================================

// ActionActivityInput is passed to the action execution activity.
// Config contains the action's JSON configuration (possibly with template variables).
// Context provides the merged execution context for template variable resolution.
type ActionActivityInput struct {
	ActionID    uuid.UUID       `json:"action_id"`
	ActionName  string          `json:"action_name"`
	ActionType  string          `json:"action_type"`
	Config      json.RawMessage `json:"config"`
	Context     map[string]any  `json:"context"`
	RuleID      uuid.UUID       `json:"rule_id"`
	ExecutionID uuid.UUID       `json:"execution_id"`
	RuleName    string          `json:"rule_name"`
}

// Validate checks that ActionActivityInput has the required fields for execution.
func (aai ActionActivityInput) Validate() error {
	if aai.ActionID == uuid.Nil {
		return fmt.Errorf("action_id is required")
	}
	if aai.ActionName == "" {
		return fmt.Errorf("action_name is required")
	}
	if aai.ActionType == "" {
		return fmt.Errorf("action_type is required")
	}
	return nil
}

// ActionActivityOutput is returned from the action execution activity.
// Result contains the action handler's output (e.g., BranchTaken for conditions).
// Success indicates whether the action completed without error.
type ActionActivityOutput struct {
	ActionID   uuid.UUID              `json:"action_id"`
	ActionName string                 `json:"action_name"`
	Result     map[string]any `json:"result"`
	Success    bool                   `json:"success"`
}
