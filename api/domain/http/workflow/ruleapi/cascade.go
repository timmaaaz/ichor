package ruleapi

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/foundation/web"
)

// ============================================================
// Cascade Visualization Response Types (Phase 12.8)
// ============================================================

// CascadeResponse represents downstream workflows for a rule.
// Shows users which workflows will be triggered when this rule's actions execute.
type CascadeResponse struct {
	RuleID   string              `json:"rule_id"`
	RuleName string              `json:"rule_name"`
	Actions  []ActionCascadeInfo `json:"actions"`
}

// Encode implements web.Encoder for CascadeResponse.
func (r CascadeResponse) Encode() ([]byte, string, error) {
	data, err := json.Marshal(r)
	return data, "application/json", err
}

// ActionCascadeInfo describes an action and any downstream workflows it may trigger.
type ActionCascadeInfo struct {
	ActionID            string                   `json:"action_id"`
	ActionName          string                   `json:"action_name"`
	ActionType          string                   `json:"action_type"`
	ModifiesEntity      string                   `json:"modifies_entity,omitempty"`
	TriggersEvent       string                   `json:"triggers_event,omitempty"`
	ModifiedFields      []string                 `json:"modified_fields,omitempty"`
	DownstreamWorkflows []DownstreamWorkflowInfo `json:"downstream_workflows"`
}

// DownstreamWorkflowInfo describes a workflow that may be triggered by an action.
type DownstreamWorkflowInfo struct {
	RuleID            string           `json:"rule_id"`
	RuleName          string           `json:"rule_name"`
	TriggerConditions *json.RawMessage `json:"trigger_conditions,omitempty"`
	WillTriggerIf     string           `json:"will_trigger_if"`
}

// cascadeMap handles GET /v1/workflow/rules/{id}/cascade-map
// Returns all downstream workflows that could be triggered by this rule's actions.
// The cascade data is gathered by the app layer (ruleApp.CascadeMap); this handler only
// shapes it into the response DTO (including the human-readable WillTriggerIf text).
func (a *api) cascadeMap(ctx context.Context, r *http.Request) web.Encoder {
	ruleID, err := uuid.Parse(web.Param(r, "id"))
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	rule, entries, err := a.ruleApp.CascadeMap(ctx, ruleID)
	if err != nil {
		return errs.NewError(err)
	}

	response := CascadeResponse{
		RuleID:   rule.ID.String(),
		RuleName: rule.Name,
		Actions:  make([]ActionCascadeInfo, 0, len(entries)),
	}

	for _, e := range entries {
		info := ActionCascadeInfo{
			ActionID:            e.ActionID.String(),
			ActionName:          e.ActionName,
			ActionType:          e.ActionType,
			ModifiesEntity:      e.ModifiesEntity,
			TriggersEvent:       e.TriggersEvent,
			ModifiedFields:      e.ModifiedFields,
			DownstreamWorkflows: make([]DownstreamWorkflowInfo, 0, len(e.Downstream)),
		}

		for _, d := range e.Downstream {
			info.DownstreamWorkflows = append(info.DownstreamWorkflows, DownstreamWorkflowInfo{
				RuleID:            d.RuleID.String(),
				RuleName:          d.RuleName,
				TriggerConditions: d.TriggerConditions,
				WillTriggerIf:     buildTriggerDescription(d.EventType, d.EntityName, d.TriggerConditions),
			})
		}

		response.Actions = append(response.Actions, info)
	}

	return response
}

// buildTriggerDescription creates a human-readable description of when the rule will trigger.
func buildTriggerDescription(eventType, entityName string, conditions *json.RawMessage) string {
	base := ""
	switch eventType {
	case "on_create":
		base = "any " + entityName + " creation"
	case "on_update":
		base = "any " + entityName + " update"
	case "on_delete":
		base = "any " + entityName + " deletion"
	default:
		base = eventType + " on " + entityName
	}

	// If there are conditions, note that
	if conditions != nil && len(*conditions) > 2 { // len > 2 to exclude empty "{}"
		return base + " (with conditions)"
	}

	return base
}
