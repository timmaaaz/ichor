package ruleapi

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/business/sdk/workflow"
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
func (a *api) cascadeMap(ctx context.Context, r *http.Request) web.Encoder {
	ruleID, err := uuid.Parse(web.Param(r, "id"))
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	// Verify the action registry is available
	if a.registry == nil {
		return errs.New(errs.Internal, errors.New("action registry not configured"))
	}

	// Get rule
	rule, err := a.workflowBus.QueryRuleByID(ctx, ruleID)
	if err != nil {
		if errors.Is(err, workflow.ErrNotFound) {
			return errs.New(errs.NotFound, err)
		}
		return errs.Newf(errs.Internal, "query rule: %s", err)
	}

	// Get actions for this rule (using the view which includes template info)
	actions, err := a.workflowBus.QueryRoleActionsViewByRuleID(ctx, ruleID)
	if err != nil {
		return errs.Newf(errs.Internal, "query actions: %s", err)
	}

	// Build response
	response := CascadeResponse{
		RuleID:   rule.ID.String(),
		RuleName: rule.Name,
		Actions:  make([]ActionCascadeInfo, 0, len(actions)),
	}

	for _, action := range actions {
		info := ActionCascadeInfo{
			ActionID:            action.ID.String(),
			ActionName:          action.Name,
			ActionType:          action.TemplateActionType,
			DownstreamWorkflows: make([]DownstreamWorkflowInfo, 0),
		}

		// Check if handler implements EntityModifier
		handler, exists := a.registry.Get(action.TemplateActionType)
		if exists {
			if modifier, ok := handler.(workflow.EntityModifier); ok {
				mods := modifier.GetEntityModifications(action.ActionConfig)

				for _, mod := range mods {
					info.ModifiesEntity = mod.EntityName
					info.TriggersEvent = mod.EventType
					info.ModifiedFields = mod.Fields

					// Find active rules that listen to this entity mutation.
					downstreamRules, err := a.workflowBus.FindDownstreamRules(ctx, mod.EntityName, mod.EventType, ruleID)
					if err != nil {
						// Fail-soft: show no downstreams for this modification rather than
						// failing the whole endpoint.
						a.log.Error(ctx, "cascade map: find downstream rules", "entity", mod.EntityName, "event", mod.EventType, "error", err)
					}
					for _, rule := range downstreamRules {
						info.DownstreamWorkflows = append(info.DownstreamWorkflows, DownstreamWorkflowInfo{
							RuleID:            rule.ID.String(),
							RuleName:          rule.Name,
							TriggerConditions: rule.TriggerConditions,
							WillTriggerIf:     buildTriggerDescription(mod.EventType, mod.EntityName, rule.TriggerConditions),
						})
					}
				}
			}
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
