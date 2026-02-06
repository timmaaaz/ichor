package ruleapi

import (
	"encoding/json"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/sdk/workflow"
)

// toRuleResponse converts a workflow.AutomationRuleView to RuleResponse.
func toRuleResponse(rule workflow.AutomationRuleView) RuleResponse {
	resp := RuleResponse{
		ID:               rule.ID,
		Name:             rule.Name,
		EntityName:       rule.EntityName,
		EntitySchemaName: rule.EntitySchemaName,
		EntityTypeName:   rule.EntityTypeName,
		TriggerTypeName:  rule.TriggerTypeName,
		CanvasLayout:     rule.CanvasLayout,
		IsActive:         rule.IsActive,
		CreatedBy:        rule.CreatedBy,
		UpdatedBy:        rule.UpdatedBy,
		CreatedDate:      rule.CreatedDate,
		UpdatedDate:      rule.UpdatedDate,
	}

	// Handle optional pointer fields from the view
	if rule.Description != nil {
		resp.Description = *rule.Description
	}
	if rule.EntityID != nil {
		resp.EntityID = *rule.EntityID
	}
	if rule.EntityTypeID != nil {
		resp.EntityTypeID = *rule.EntityTypeID
	}
	if rule.TriggerTypeID != nil {
		resp.TriggerTypeID = *rule.TriggerTypeID
	}
	if rule.TriggerConditions != nil {
		resp.TriggerConditions = *rule.TriggerConditions
	}

	return resp
}

// toRuleResponses converts a slice of workflow.AutomationRuleView to RuleList.
func toRuleResponses(rules []workflow.AutomationRuleView) RuleList {
	responses := make(RuleList, len(rules))
	for i, rule := range rules {
		responses[i] = toRuleResponse(rule)
	}
	return responses
}

// toActionResponse converts a workflow.RuleActionView to ActionResponse.
func toActionResponse(action workflow.RuleActionView) ActionResponse {
	resp := ActionResponse{
		ID:                 action.ID,
		Name:               action.Name,
		Description:        action.Description,
		ActionConfig:       action.ActionConfig,
		IsActive:           action.IsActive,
		TemplateID:         action.TemplateID,
		TemplateName:       action.TemplateName,
		TemplateActionType: action.TemplateActionType,
	}

	// RuleActionView uses AutomationRulesID (note the 's')
	if action.AutomationRulesID != nil {
		resp.RuleID = *action.AutomationRulesID
	}

	return resp
}

// toActionResponses converts a slice of workflow.RuleActionView to []ActionResponse.
func toActionResponses(actions []workflow.RuleActionView) []ActionResponse {
	responses := make([]ActionResponse, len(actions))
	for i, action := range actions {
		responses[i] = toActionResponse(action)
	}
	return responses
}

// toNewAutomationRule converts CreateRuleRequest to workflow.NewAutomationRule.
func toNewAutomationRule(req CreateRuleRequest, createdBy uuid.UUID) workflow.NewAutomationRule {
	// Convert json.RawMessage (value) to *json.RawMessage (pointer) for business layer
	var triggerConditions *json.RawMessage
	if len(req.TriggerConditions) > 0 {
		triggerConditions = &req.TriggerConditions
	}

	return workflow.NewAutomationRule{
		Name:              req.Name,
		Description:       req.Description,
		EntityID:          req.EntityID,
		EntityTypeID:      req.EntityTypeID,
		TriggerTypeID:     req.TriggerTypeID,
		TriggerConditions: triggerConditions,
		CanvasLayout:      req.CanvasLayout,
		IsActive:          req.IsActive,
		CreatedBy:         createdBy,
	}
}

// toUpdateAutomationRule converts UpdateRuleRequest to workflow.UpdateAutomationRule.
func toUpdateAutomationRule(req UpdateRuleRequest, updatedBy uuid.UUID) workflow.UpdateAutomationRule {
	return workflow.UpdateAutomationRule{
		Name:              req.Name,
		Description:       req.Description,
		EntityID:          req.EntityID,
		EntityTypeID:      req.EntityTypeID,
		TriggerTypeID:     req.TriggerTypeID,
		TriggerConditions: req.TriggerConditions,
		CanvasLayout:      req.CanvasLayout,
		IsActive:          req.IsActive,
		UpdatedBy:         &updatedBy,
	}
}

// toNewRuleAction converts CreateActionInput to workflow.NewRuleAction.
func toNewRuleAction(ruleID uuid.UUID, input CreateActionInput) workflow.NewRuleAction {
	return workflow.NewRuleAction{
		AutomationRuleID: ruleID,
		Name:             input.Name,
		Description:      input.Description,
		ActionConfig:     input.ActionConfig,
		IsActive:         input.IsActive,
		TemplateID:       input.TemplateID,
	}
}

// ============================================================
// Phase 4C: Action Request Converters
// ============================================================

// toNewRuleActionFromRequest converts CreateActionRequest to workflow.NewRuleAction.
// Note: No CreatedBy parameter - business layer handles audit fields internally.
func toNewRuleActionFromRequest(ruleID uuid.UUID, req CreateActionRequest) workflow.NewRuleAction {
	return workflow.NewRuleAction{
		AutomationRuleID: ruleID,
		Name:             req.Name,
		Description:      req.Description,
		ActionConfig:     req.ActionConfig,
		IsActive:         req.IsActive,
		TemplateID:       req.TemplateID,
	}
}

// toUpdateRuleAction converts UpdateActionRequest to workflow.UpdateRuleAction.
// Note: No ID or UpdatedBy parameters - these are handled separately.
func toUpdateRuleAction(req UpdateActionRequest) workflow.UpdateRuleAction {
	return workflow.UpdateRuleAction{
		Name:         req.Name,
		Description:  req.Description,
		ActionConfig: req.ActionConfig,
		IsActive:     req.IsActive,
		TemplateID:     req.TemplateID,
	}
}
