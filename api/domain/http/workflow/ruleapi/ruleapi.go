// Package ruleapi provides HTTP handlers for workflow automation rule CRUD operations.
package ruleapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/app/sdk/mid"
	"github.com/timmaaaz/ichor/app/sdk/query"
	"github.com/timmaaaz/ichor/business/sdk/order"
	"github.com/timmaaaz/ichor/business/sdk/page"
	"github.com/timmaaaz/ichor/business/sdk/workflow"
	"github.com/timmaaaz/ichor/foundation/logger"
	"github.com/timmaaaz/ichor/foundation/web"
)

// api holds dependencies for rule HTTP handlers.
type api struct {
	log         *logger.Logger
	workflowBus *workflow.Business
}

// newAPI creates a new rule API handler.
func newAPI(log *logger.Logger, workflowBus *workflow.Business) *api {
	return &api{
		log:         log,
		workflowBus: workflowBus,
	}
}

// query handles GET /v1/workflow/rules
func (a *api) query(ctx context.Context, r *http.Request) web.Encoder {
	qp := parseQueryParams(r)

	filter, err := parseFilter(qp)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	orderBy, err := order.Parse(orderByFields, qp.OrderBy, workflow.DefaultOrderBy)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	pg, err := page.Parse(qp.Page, qp.Rows)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	rules, err := a.workflowBus.QueryAutomationRulesViewPaginated(ctx, filter, orderBy, pg)
	if err != nil {
		return errs.Newf(errs.Internal, "query: %s", err)
	}

	total, err := a.workflowBus.CountAutomationRulesView(ctx, filter)
	if err != nil {
		return errs.Newf(errs.Internal, "count: %s", err)
	}

	return query.NewResult(toRuleResponses(rules), total, pg)
}

// queryByID handles GET /v1/workflow/rules/{id}
func (a *api) queryByID(ctx context.Context, r *http.Request) web.Encoder {
	id, err := uuid.Parse(web.Param(r, "id"))
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	// First try to get the rule (to verify it exists)
	_, err = a.workflowBus.QueryRuleByID(ctx, id)
	if err != nil {
		if errors.Is(err, workflow.ErrNotFound) {
			return errs.New(errs.NotFound, err)
		}
		return errs.Newf(errs.Internal, "query rule: %s", err)
	}

	// Get the view for the response (includes joined fields)
	// Note: There's no QueryAutomationRuleViewByID, so we use the paginated query with ID filter
	filter := workflow.AutomationRuleFilter{ID: &id}
	views, err := a.workflowBus.QueryAutomationRulesViewPaginated(ctx, filter, workflow.DefaultOrderBy, page.MustParse("1", "1"))
	if err != nil {
		return errs.Newf(errs.Internal, "query view: %s", err)
	}
	if len(views) == 0 {
		return errs.New(errs.NotFound, workflow.ErrNotFound)
	}

	// Get actions for this rule
	actions, err := a.workflowBus.QueryRoleActionsViewByRuleID(ctx, id)
	if err != nil {
		a.log.Error(ctx, "failed to query actions", "error", err, "rule_id", id)
		// Continue without actions rather than failing the request
	}

	response := toRuleResponse(views[0])
	response.Actions = toActionResponses(actions)

	return response
}

// create handles POST /v1/workflow/rules
func (a *api) create(ctx context.Context, r *http.Request) web.Encoder {
	var req CreateRuleRequest
	if err := web.Decode(r, &req); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	// Validate request
	if validationErr := ValidateCreateRule(req); validationErr != nil {
		return errs.New(errs.FailedPrecondition, validationErr)
	}

	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return errs.New(errs.Unauthenticated, err)
	}

	// Create the rule using business layer
	// Business layer method: CreateRule(ctx, NewAutomationRule) (AutomationRule, error)
	newRule := toNewAutomationRule(req, userID)
	rule, err := a.workflowBus.CreateRule(ctx, newRule)
	if err != nil {
		return errs.Newf(errs.Internal, "create rule: %s", err)
	}

	a.log.Info(ctx, "rule created", "rule_id", rule.ID, "created_by", userID)

	// Create embedded actions if provided
	// Business layer method: CreateRuleAction(ctx, NewRuleAction) (RuleAction, error)
	for _, actionInput := range req.Actions {
		newAction := toNewRuleAction(rule.ID, actionInput)
		action, err := a.workflowBus.CreateRuleAction(ctx, newAction)
		if err != nil {
			a.log.Error(ctx, "failed to create action", "error", err, "rule_id", rule.ID)
			// Note: Rule is created but action failed. Consider if this should rollback.
			// For now, continue and log the error.
			return errs.Newf(errs.Internal, "create action: %s", err)
		}
		a.log.Info(ctx, "action created", "action_id", action.ID, "rule_id", rule.ID)
	}

	// Fetch the view for response (includes joined fields)
	filter := workflow.AutomationRuleFilter{ID: &rule.ID}
	views, err := a.workflowBus.QueryAutomationRulesViewPaginated(ctx, filter, workflow.DefaultOrderBy, page.MustParse("1", "1"))
	if err != nil {
		return errs.Newf(errs.Internal, "query created rule: %s", err)
	}
	if len(views) == 0 {
		return errs.Newf(errs.Internal, "created rule not found in view")
	}

	response := toRuleResponse(views[0])
	if len(req.Actions) > 0 {
		actionsView, err := a.workflowBus.QueryRoleActionsViewByRuleID(ctx, rule.ID)
		if err != nil {
			a.log.Error(ctx, "failed to query actions for response", "error", err, "rule_id", rule.ID)
		}
		response.Actions = toActionResponses(actionsView)
	}

	return response
}

// update handles PUT /v1/workflow/rules/{id}
func (a *api) update(ctx context.Context, r *http.Request) web.Encoder {
	id, err := uuid.Parse(web.Param(r, "id"))
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	var req UpdateRuleRequest
	if err := web.Decode(r, &req); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	// Validate request
	if validationErr := ValidateUpdateRule(req); validationErr != nil {
		return errs.New(errs.FailedPrecondition, validationErr)
	}

	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return errs.New(errs.Unauthenticated, err)
	}

	// Fetch existing rule first (required by UpdateRule method signature)
	// Business layer method: QueryRuleByID(ctx, id) (AutomationRule, error)
	existingRule, err := a.workflowBus.QueryRuleByID(ctx, id)
	if err != nil {
		if errors.Is(err, workflow.ErrNotFound) {
			return errs.New(errs.NotFound, err)
		}
		return errs.Newf(errs.Internal, "query rule: %s", err)
	}

	// Update the rule
	// Business layer method: UpdateRule(ctx, AutomationRule, UpdateAutomationRule) (AutomationRule, error)
	updateRule := toUpdateAutomationRule(req, userID)
	updatedRule, err := a.workflowBus.UpdateRule(ctx, existingRule, updateRule)
	if err != nil {
		return errs.Newf(errs.Internal, "update rule: %s", err)
	}

	a.log.Info(ctx, "rule updated", "rule_id", updatedRule.ID, "updated_by", userID)

	// Fetch updated view for response
	filter := workflow.AutomationRuleFilter{ID: &id}
	views, err := a.workflowBus.QueryAutomationRulesViewPaginated(ctx, filter, workflow.DefaultOrderBy, page.MustParse("1", "1"))
	if err != nil {
		return errs.Newf(errs.Internal, "query updated rule: %s", err)
	}
	if len(views) == 0 {
		return errs.Newf(errs.Internal, "updated rule not found in view")
	}

	actionsView, err := a.workflowBus.QueryRoleActionsViewByRuleID(ctx, id)
	if err != nil {
		a.log.Error(ctx, "failed to query actions for response", "error", err, "rule_id", id)
	}

	response := toRuleResponse(views[0])
	response.Actions = toActionResponses(actionsView)

	return response
}

// delete handles DELETE /v1/workflow/rules/{id}
// Note: This is a soft-delete (deactivation), not a hard delete.
func (a *api) delete(ctx context.Context, r *http.Request) web.Encoder {
	id, err := uuid.Parse(web.Param(r, "id"))
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return errs.New(errs.Unauthenticated, err)
	}

	// Fetch existing rule first (required by DeactivateRule method signature)
	// Business layer method: QueryRuleByID(ctx, id) (AutomationRule, error)
	existingRule, err := a.workflowBus.QueryRuleByID(ctx, id)
	if err != nil {
		if errors.Is(err, workflow.ErrNotFound) {
			return errs.New(errs.NotFound, err)
		}
		return errs.Newf(errs.Internal, "query rule: %s", err)
	}

	// Deactivate the rule
	// Business layer method: DeactivateRule(ctx, AutomationRule) error
	if err := a.workflowBus.DeactivateRule(ctx, existingRule); err != nil {
		return errs.Newf(errs.Internal, "deactivate rule: %s", err)
	}

	a.log.Info(ctx, "rule deactivated", "rule_id", id, "deactivated_by", userID)

	return nil
}

// toggleActive handles PATCH /v1/workflow/rules/{id}/active
func (a *api) toggleActive(ctx context.Context, r *http.Request) web.Encoder {
	id, err := uuid.Parse(web.Param(r, "id"))
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	var req ToggleActiveRequest
	if err := web.Decode(r, &req); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return errs.New(errs.Unauthenticated, err)
	}

	// Fetch existing rule first (required by Activate/DeactivateRule method signatures)
	// Business layer method: QueryRuleByID(ctx, id) (AutomationRule, error)
	existingRule, err := a.workflowBus.QueryRuleByID(ctx, id)
	if err != nil {
		if errors.Is(err, workflow.ErrNotFound) {
			return errs.New(errs.NotFound, err)
		}
		return errs.Newf(errs.Internal, "query rule: %s", err)
	}

	// Toggle active status
	// Business layer methods:
	//   ActivateRule(ctx, AutomationRule) error
	//   DeactivateRule(ctx, AutomationRule) error
	if req.IsActive {
		err = a.workflowBus.ActivateRule(ctx, existingRule)
	} else {
		err = a.workflowBus.DeactivateRule(ctx, existingRule)
	}

	if err != nil {
		return errs.Newf(errs.Internal, "toggle active: %s", err)
	}

	a.log.Info(ctx, "rule active status toggled", "rule_id", id, "is_active", req.IsActive, "toggled_by", userID)

	// Fetch updated view for response
	filter := workflow.AutomationRuleFilter{ID: &id}
	views, err := a.workflowBus.QueryAutomationRulesViewPaginated(ctx, filter, workflow.DefaultOrderBy, page.MustParse("1", "1"))
	if err != nil {
		return errs.Newf(errs.Internal, "query rule: %s", err)
	}
	if len(views) == 0 {
		return errs.Newf(errs.Internal, "rule not found in view after toggle")
	}

	actionsView, err := a.workflowBus.QueryRoleActionsViewByRuleID(ctx, id)
	if err != nil {
		a.log.Error(ctx, "failed to query actions for response", "error", err, "rule_id", id)
	}

	response := toRuleResponse(views[0])
	response.Actions = toActionResponses(actionsView)

	return response
}

// ============================================================
// Phase 4C: Action CRUD Handlers
// ============================================================

// queryActions handles GET /v1/workflow/rules/{id}/actions
func (a *api) queryActions(ctx context.Context, r *http.Request) web.Encoder {
	ruleID, err := uuid.Parse(web.Param(r, "id"))
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	// Verify rule exists
	_, err = a.workflowBus.QueryRuleByID(ctx, ruleID)
	if err != nil {
		if errors.Is(err, workflow.ErrNotFound) {
			return errs.New(errs.NotFound, err)
		}
		return errs.Newf(errs.Internal, "query rule: %s", err)
	}

	// Note: Method name has "Role" not "Rule" - this is the actual implementation
	actions, err := a.workflowBus.QueryRoleActionsViewByRuleID(ctx, ruleID)
	if err != nil {
		return errs.Newf(errs.Internal, "query actions: %s", err)
	}

	return ActionList(toActionResponses(actions))
}

// createAction handles POST /v1/workflow/rules/{id}/actions
func (a *api) createAction(ctx context.Context, r *http.Request) web.Encoder {
	ruleID, err := uuid.Parse(web.Param(r, "id"))
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	var req CreateActionRequest
	if err := web.Decode(r, &req); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	// Validate request
	if validationErr := ValidateCreateAction(req); validationErr != nil {
		return errs.New(errs.InvalidArgument, validationErr)
	}

	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return errs.New(errs.Unauthenticated, err)
	}

	// Verify rule exists
	_, err = a.workflowBus.QueryRuleByID(ctx, ruleID)
	if err != nil {
		if errors.Is(err, workflow.ErrNotFound) {
			return errs.New(errs.NotFound, err)
		}
		return errs.Newf(errs.Internal, "query rule: %s", err)
	}

	// Convert request - NO userID parameter (business layer handles CreatedBy)
	newAction := toNewRuleActionFromRequest(ruleID, req)

	// Create action - NO time.Now() parameter (business layer generates timestamp)
	action, err := a.workflowBus.CreateRuleAction(ctx, newAction)
	if err != nil {
		return errs.Newf(errs.Internal, "create action: %s", err)
	}

	// Audit logging
	a.log.Info(ctx, "action created", "action_id", action.ID, "rule_id", ruleID, "created_by", userID)

	// Fetch view for response (includes template name if applicable)
	actionView, err := a.workflowBus.QueryActionViewByID(ctx, action.ID)
	if err != nil {
		return errs.Newf(errs.Internal, "query created action: %s", err)
	}

	return toActionResponse(actionView)
}

// updateAction handles PUT /v1/workflow/rules/{id}/actions/{action_id}
func (a *api) updateAction(ctx context.Context, r *http.Request) web.Encoder {
	ruleID, err := uuid.Parse(web.Param(r, "id"))
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	actionID, err := uuid.Parse(web.Param(r, "action_id"))
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	var req UpdateActionRequest
	if err := web.Decode(r, &req); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	// Validate request
	if validationErr := ValidateUpdateAction(req); validationErr != nil {
		return errs.New(errs.FailedPrecondition, validationErr)
	}

	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return errs.New(errs.Unauthenticated, err)
	}

	// Fetch existing action (required by UpdateRuleAction signature)
	existing, err := a.workflowBus.QueryActionByID(ctx, actionID)
	if err != nil {
		if errors.Is(err, workflow.ErrNotFound) {
			return errs.New(errs.NotFound, err)
		}
		return errs.Newf(errs.Internal, "query action: %s", err)
	}

	// Verify action belongs to the specified rule
	if existing.AutomationRuleID != ruleID {
		return errs.New(errs.NotFound, workflow.ErrActionNotInRule)
	}

	// Convert request - NO id or userID parameters
	updateAction := toUpdateRuleAction(req)

	// Update action - pass BOTH existing action AND update struct, NO time.Now()
	updatedAction, err := a.workflowBus.UpdateRuleAction(ctx, existing, updateAction)
	if err != nil {
		return errs.Newf(errs.Internal, "update action: %s", err)
	}

	// Audit logging
	a.log.Info(ctx, "action updated", "action_id", actionID, "rule_id", ruleID, "updated_by", userID)

	// Fetch view for response
	actionView, err := a.workflowBus.QueryActionViewByID(ctx, updatedAction.ID)
	if err != nil {
		return errs.Newf(errs.Internal, "query updated action: %s", err)
	}

	return toActionResponse(actionView)
}

// deleteAction handles DELETE /v1/workflow/rules/{id}/actions/{action_id}
func (a *api) deleteAction(ctx context.Context, r *http.Request) web.Encoder {
	ruleID, err := uuid.Parse(web.Param(r, "id"))
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	actionID, err := uuid.Parse(web.Param(r, "action_id"))
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return errs.New(errs.Unauthenticated, err)
	}

	// Fetch existing action (required by DeactivateRuleAction signature)
	existing, err := a.workflowBus.QueryActionByID(ctx, actionID)
	if err != nil {
		if errors.Is(err, workflow.ErrNotFound) {
			return errs.New(errs.NotFound, err)
		}
		return errs.Newf(errs.Internal, "query action: %s", err)
	}

	// Verify action belongs to the specified rule
	if existing.AutomationRuleID != ruleID {
		return errs.New(errs.NotFound, workflow.ErrActionNotInRule)
	}

	// Deactivate action - pass existing action STRUCT, not actionID, NO time.Now()
	err = a.workflowBus.DeactivateRuleAction(ctx, existing)
	if err != nil {
		return errs.Newf(errs.Internal, "deactivate action: %s", err)
	}

	// Audit logging
	a.log.Info(ctx, "action deactivated", "action_id", actionID, "rule_id", ruleID, "deactivated_by", userID)

	return nil
}

// ============================================================
// Phase 4C: Validation Handler
// ============================================================

// validateRule handles POST /v1/workflow/rules/{id}/validate
func (a *api) validateRule(ctx context.Context, r *http.Request) web.Encoder {
	ruleID, err := uuid.Parse(web.Param(r, "id"))
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	// Get rule
	rule, err := a.workflowBus.QueryRuleByID(ctx, ruleID)
	if err != nil {
		if errors.Is(err, workflow.ErrNotFound) {
			return errs.New(errs.NotFound, err)
		}
		return errs.Newf(errs.Internal, "query rule: %s", err)
	}

	// Get actions (returns []RuleAction for validation)
	actions, err := a.workflowBus.QueryActionsByRule(ctx, ruleID)
	if err != nil {
		return errs.Newf(errs.Internal, "query actions: %s", err)
	}

	var issues []ValidationIssue

	// Validate rule has at least one action
	if len(actions) == 0 {
		issues = append(issues, ValidationIssue{
			Level:   "warning",
			Field:   "actions",
			Message: "Rule has no actions configured",
		})
	}

	// Validate trigger conditions JSON - IMPORTANT: TriggerConditions is a pointer type
	if rule.TriggerConditions != nil && len(*rule.TriggerConditions) > 0 {
		var tc interface{}
		if err := json.Unmarshal(*rule.TriggerConditions, &tc); err != nil {
			issues = append(issues, ValidationIssue{
				Level:   "error",
				Field:   "trigger_conditions",
				Message: "trigger_conditions contains invalid JSON",
			})
		}
	}

	// Validate each action's config
	for i, action := range actions {
		if len(action.ActionConfig) > 0 {
			var ac interface{}
			if err := json.Unmarshal(action.ActionConfig, &ac); err != nil {
				issues = append(issues, ValidationIssue{
					Level:   "error",
					Field:   fmt.Sprintf("actions[%d].action_config", i),
					Message: "action_config contains invalid JSON",
				})
			}
		}
	}

	// Check for duplicate execution orders
	orderMap := make(map[int]int)
	for _, action := range actions {
		orderMap[action.ExecutionOrder]++
	}
	for order, count := range orderMap {
		if count > 1 {
			issues = append(issues, ValidationIssue{
				Level:   "warning",
				Field:   "execution_order",
				Message: fmt.Sprintf("Multiple actions have execution_order %d", order),
			})
		}
	}

	// Determine overall validity (errors make it invalid, warnings don't)
	isValid := true
	for _, issue := range issues {
		if issue.Level == "error" {
			isValid = false
			break
		}
	}

	return ValidateRuleResponse{
		Valid:  isValid,
		RuleID: ruleID,
		Issues: issues,
	}
}

// ============================================================
// Phase 5: Rule Execution History
// ============================================================

// queryRuleExecutions handles GET /v1/workflow/rules/{id}/executions
func (a *api) queryRuleExecutions(ctx context.Context, r *http.Request) web.Encoder {
	ruleID, err := uuid.Parse(web.Param(r, "id"))
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	// Verify rule exists
	_, err = a.workflowBus.QueryRuleByID(ctx, ruleID)
	if err != nil {
		if errors.Is(err, workflow.ErrNotFound) {
			return errs.New(errs.NotFound, err)
		}
		return errs.Newf(errs.Internal, "query rule: %s", err)
	}

	// Parse query parameters for pagination
	qp := parseExecutionQueryParams(r)

	orderBy, err := order.Parse(executionOrderByFields, qp.OrderBy, workflow.DefaultExecutionOrderBy)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	pg, err := page.Parse(qp.Page, qp.Rows)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	// Build filter with rule_id pre-populated
	filter := workflow.ExecutionFilter{
		RuleID: &ruleID,
	}

	// Add optional status filter
	if qp.Status != "" {
		status := workflow.ExecutionStatus(qp.Status)
		filter.Status = &status
	}

	executions, err := a.workflowBus.QueryExecutionsPaginated(ctx, filter, orderBy, pg)
	if err != nil {
		return errs.Newf(errs.Internal, "query: %s", err)
	}

	total, err := a.workflowBus.CountExecutions(ctx, filter)
	if err != nil {
		return errs.Newf(errs.Internal, "count: %s", err)
	}

	return query.NewResult(toRuleExecutionResponses(executions), total, pg)
}

// ExecutionQueryParams holds raw query parameter values for execution queries.
type ExecutionQueryParams struct {
	Page    string
	Rows    string
	OrderBy string
	Status  string
}

// parseExecutionQueryParams extracts execution query parameters from the HTTP request.
func parseExecutionQueryParams(r *http.Request) ExecutionQueryParams {
	return ExecutionQueryParams{
		Page:    r.URL.Query().Get("page"),
		Rows:    r.URL.Query().Get("rows"),
		OrderBy: r.URL.Query().Get("orderBy"),
		Status:  r.URL.Query().Get("status"),
	}
}

// executionOrderByFields maps API field names to business layer order constants.
var executionOrderByFields = map[string]string{
	"id":                  workflow.ExecutionOrderByID,
	"executed_at":         workflow.ExecutionOrderByExecutedAt,
	"status":              workflow.ExecutionOrderByStatus,
	"automation_rules_id": workflow.ExecutionOrderByRuleID,
	"entity_type":         workflow.ExecutionOrderByEntityType,
}

// RuleExecutionResponse is the response for a single execution in a rule's execution list.
type RuleExecutionResponse struct {
	ID              uuid.UUID  `json:"id"`
	Status          string     `json:"status"`
	ErrorMessage    string     `json:"error_message,omitempty"`
	ExecutionTimeMs int        `json:"execution_time_ms"`
	ExecutedAt      time.Time  `json:"executed_at"`
	TriggerSource   string     `json:"trigger_source"`
	ExecutedBy      *uuid.UUID `json:"executed_by,omitempty"`
}

// RuleExecutionList wraps a slice of rule executions for JSON encoding.
type RuleExecutionList []RuleExecutionResponse

// Encode implements web.Encoder for RuleExecutionList.
func (l RuleExecutionList) Encode() ([]byte, string, error) {
	data, err := json.Marshal(l)
	return data, "application/json", err
}

// toRuleExecutionResponse converts a business execution to an API response.
func toRuleExecutionResponse(exec workflow.AutomationExecution) RuleExecutionResponse {
	return RuleExecutionResponse{
		ID:              exec.ID,
		Status:          string(exec.Status),
		ErrorMessage:    exec.ErrorMessage,
		ExecutionTimeMs: exec.ExecutionTimeMs,
		ExecutedAt:      exec.ExecutedAt,
		TriggerSource:   exec.TriggerSource,
		ExecutedBy:      exec.ExecutedBy,
	}
}

// toRuleExecutionResponses converts a slice of business executions to API responses.
func toRuleExecutionResponses(executions []workflow.AutomationExecution) RuleExecutionList {
	resp := make(RuleExecutionList, len(executions))
	for i, exec := range executions {
		resp[i] = toRuleExecutionResponse(exec)
	}
	return resp
}
