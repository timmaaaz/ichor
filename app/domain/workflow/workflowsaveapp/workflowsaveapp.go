// Package workflowsaveapp provides the application layer for transactional workflow save operations.
package workflowsaveapp

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/business/sdk/delegate"
	"github.com/timmaaaz/ichor/business/sdk/workflow"
	"github.com/timmaaaz/ichor/foundation/logger"
)

// App manages the set of app layer API functions for workflow save operations.
type App struct {
	log         *logger.Logger
	db          *sqlx.DB
	workflowBus *workflow.Business
	delegate    *delegate.Delegate
	registry    *workflow.ActionRegistry
}

// NewApp constructs a workflow save app API for use.
func NewApp(log *logger.Logger, db *sqlx.DB, workflowBus *workflow.Business, del *delegate.Delegate, registry *workflow.ActionRegistry) *App {
	return &App{
		log:         log,
		db:          db,
		workflowBus: workflowBus,
		delegate:    del,
		registry:    registry,
	}
}

// prepareRequest validates the request and prepares it for saving. When actions
// are present, action configs and graph structure are validated. When no actions
// are present, IsActive is forced to false (draft workflow).
func (a *App) prepareRequest(req *SaveWorkflowRequest) error {
	if err := req.Validate(); err != nil {
		return err
	}

	if len(req.Actions) > 0 {
		if err := ValidateActionConfigs(req.Actions); err != nil {
			return errs.Newf(errs.InvalidArgument, "action config: %s", err)
		}

		if err := ValidateGraph(req.Actions, req.Edges); err != nil {
			return errs.Newf(errs.InvalidArgument, "graph: %s", err)
		}

		if a.registry != nil {
			if err := validateOutputPorts(req.Actions, req.Edges, a.registry); err != nil {
				return errs.Newf(errs.InvalidArgument, "output port: %s", err)
			}
		}
	} else {
		req.IsActive = false
	}

	return nil
}

// DryRunValidate runs full validation on the workflow request without committing
// to the database. Returns a ValidationResult indicating success or errors.
func (a *App) DryRunValidate(req SaveWorkflowRequest) ValidationResult {
	var validationErrors []string

	if err := req.Validate(); err != nil {
		validationErrors = append(validationErrors, err.Error())
	}

	if len(req.Actions) > 0 {
		if err := ValidateActionConfigs(req.Actions); err != nil {
			validationErrors = append(validationErrors, fmt.Sprintf("action config: %s", err))
		}

		if err := ValidateGraph(req.Actions, req.Edges); err != nil {
			validationErrors = append(validationErrors, fmt.Sprintf("graph: %s", err))
		}

		if a.registry != nil {
			if err := validateOutputPorts(req.Actions, req.Edges, a.registry); err != nil {
				validationErrors = append(validationErrors, fmt.Sprintf("output port: %s", err))
			}
		}
	}

	return ValidationResult{
		Valid:       len(validationErrors) == 0,
		Errors:      validationErrors,
		ActionCount: len(req.Actions),
		EdgeCount:   len(req.Edges),
	}
}

// SaveWorkflow updates an existing workflow atomically (rule + actions + edges).
// This performs all operations within a single database transaction.
func (a *App) SaveWorkflow(ctx context.Context, ruleID uuid.UUID, req SaveWorkflowRequest) (SaveWorkflowResponse, error) {
	if err := a.prepareRequest(&req); err != nil {
		return SaveWorkflowResponse{}, err
	}

	// Begin transaction
	tx, err := a.db.BeginTxx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return SaveWorkflowResponse{}, errs.Newf(errs.Internal, "begin tx: %s", err)
	}
	defer func() {
		_ = tx.Rollback()
	}()

	// 5. Get transaction-aware business
	txBus, err := a.workflowBus.NewWithTx(tx)
	if err != nil {
		return SaveWorkflowResponse{}, errs.Newf(errs.Internal, "new with tx: %s", err)
	}

	// 6. Fetch and update rule
	rule, err := a.updateRule(ctx, txBus, ruleID, req)
	if err != nil {
		return SaveWorkflowResponse{}, err
	}

	// 7. Sync actions (create/update/delete)
	actionIDMap, savedActions, err := a.syncActions(ctx, txBus, ruleID, req.Actions)
	if err != nil {
		return SaveWorkflowResponse{}, err
	}

	// 8. Delete existing edges and recreate
	if err := txBus.DeleteEdgesByRuleID(ctx, ruleID); err != nil {
		return SaveWorkflowResponse{}, errs.Newf(errs.Internal, "delete edges: %s", err)
	}

	savedEdges, err := a.createEdges(ctx, txBus, ruleID, req.Edges, actionIDMap)
	if err != nil {
		return SaveWorkflowResponse{}, err
	}

	// 9. Commit transaction
	if err := tx.Commit(); err != nil {
		return SaveWorkflowResponse{}, errs.Newf(errs.Internal, "commit: %s", err)
	}

	// 10. Fire delegate event AFTER commit to invalidate cache
	if a.delegate != nil {
		a.log.Info(ctx, "workflowsaveapp: transaction committed, firing delegate event", "ruleID", ruleID, "action", "updated")
		if err := a.delegate.Call(ctx, workflow.ActionRuleChangedData(workflow.ActionRuleUpdated, ruleID)); err != nil {
			a.log.Error(ctx, "workflowsaveapp: delegate call failed", "action", workflow.ActionRuleUpdated, "err", err)
		}
	}

	return buildResponse(rule, savedActions, savedEdges, req.CanvasLayout), nil
}

// CreateWorkflow creates a new workflow atomically (rule + actions + edges).
// This performs all operations within a single database transaction.
func (a *App) CreateWorkflow(ctx context.Context, userID uuid.UUID, req SaveWorkflowRequest) (SaveWorkflowResponse, error) {
	if err := a.prepareRequest(&req); err != nil {
		return SaveWorkflowResponse{}, err
	}

	// Begin transaction
	tx, err := a.db.BeginTxx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return SaveWorkflowResponse{}, errs.Newf(errs.Internal, "begin tx: %s", err)
	}
	defer func() {
		_ = tx.Rollback()
	}()

	// 5. Get transaction-aware business
	txBus, err := a.workflowBus.NewWithTx(tx)
	if err != nil {
		return SaveWorkflowResponse{}, errs.Newf(errs.Internal, "new with tx: %s", err)
	}

	// 6. Create rule
	rule, err := a.createRule(ctx, txBus, userID, req)
	if err != nil {
		return SaveWorkflowResponse{}, err
	}

	// 7. Create all actions
	actionIDMap, savedActions, err := a.createAllActions(ctx, txBus, rule.ID, req.Actions)
	if err != nil {
		return SaveWorkflowResponse{}, err
	}

	// 8. Create edges
	savedEdges, err := a.createEdges(ctx, txBus, rule.ID, req.Edges, actionIDMap)
	if err != nil {
		return SaveWorkflowResponse{}, err
	}

	// 9. Commit transaction
	if err := tx.Commit(); err != nil {
		return SaveWorkflowResponse{}, errs.Newf(errs.Internal, "commit: %s", err)
	}

	// 10. Fire delegate event AFTER commit to invalidate cache
	if a.delegate != nil {
		a.log.Info(ctx, "workflowsaveapp: transaction committed, firing delegate event", "ruleID", rule.ID, "action", "created")
		if err := a.delegate.Call(ctx, workflow.ActionRuleChangedData(workflow.ActionRuleCreated, rule.ID)); err != nil {
			a.log.Error(ctx, "workflowsaveapp: delegate call failed", "action", workflow.ActionRuleCreated, "err", err)
		}
	}

	return buildResponse(rule, savedActions, savedEdges, req.CanvasLayout), nil
}

// updateRule updates the automation rule metadata.
func (a *App) updateRule(ctx context.Context, bus *workflow.Business, ruleID uuid.UUID, req SaveWorkflowRequest) (workflow.AutomationRule, error) {
	// Fetch existing rule
	rule, err := bus.QueryRuleByID(ctx, ruleID)
	if err != nil {
		return workflow.AutomationRule{}, errs.Newf(errs.NotFound, "rule not found: %s", err)
	}

	// Default workflows cannot be modified via save — use duplicate instead.
	if rule.IsDefault {
		return workflow.AutomationRule{}, errs.Newf(errs.PermissionDenied, "cannot modify default workflow: use POST /v1/workflow/rules/{id}/duplicate to create an editable copy")
	}

	// Parse entity ID
	entityID, err := uuid.Parse(req.EntityID)
	if err != nil {
		return workflow.AutomationRule{}, errs.Newf(errs.InvalidArgument, "invalid entity_id: %s", err)
	}

	// Parse trigger type ID
	triggerTypeID, err := uuid.Parse(req.TriggerTypeID)
	if err != nil {
		return workflow.AutomationRule{}, errs.Newf(errs.InvalidArgument, "invalid trigger_type_id: %s", err)
	}

	// Prepare trigger conditions
	var triggerConditions *json.RawMessage
	if len(req.TriggerConditions) > 0 {
		triggerConditions = &req.TriggerConditions
	}

	// Update the rule
	update := workflow.UpdateAutomationRule{
		Name:              &req.Name,
		Description:       &req.Description,
		EntityID:          &entityID,
		TriggerTypeID:     &triggerTypeID,
		TriggerConditions: triggerConditions,
		CanvasLayout:      &req.CanvasLayout,
		IsActive:          &req.IsActive,
	}

	updatedRule, err := bus.UpdateRule(ctx, rule, update)
	if err != nil {
		return workflow.AutomationRule{}, errs.Newf(errs.Internal, "update rule: %s", err)
	}

	return updatedRule, nil
}

// createRule creates a new automation rule.
func (a *App) createRule(ctx context.Context, bus *workflow.Business, userID uuid.UUID, req SaveWorkflowRequest) (workflow.AutomationRule, error) {
	// Parse entity ID
	entityID, err := uuid.Parse(req.EntityID)
	if err != nil {
		return workflow.AutomationRule{}, errs.Newf(errs.InvalidArgument, "invalid entity_id: %s", err)
	}

	// Parse trigger type ID
	triggerTypeID, err := uuid.Parse(req.TriggerTypeID)
	if err != nil {
		return workflow.AutomationRule{}, errs.Newf(errs.InvalidArgument, "invalid trigger_type_id: %s", err)
	}

	// Prepare trigger conditions
	var triggerConditions *json.RawMessage
	if len(req.TriggerConditions) > 0 {
		triggerConditions = &req.TriggerConditions
	}

	// We need to lookup the entity to get EntityTypeID
	// For now, we'll need to get entity type from somewhere - assume it's passed or we query it
	// The business layer requires EntityTypeID, so we need to fetch the entity first
	// This is a simplification - in practice we might need to query the entity
	entity, err := a.getEntityType(ctx, bus, entityID)
	if err != nil {
		return workflow.AutomationRule{}, err
	}

	newRule := workflow.NewAutomationRule{
		Name:              req.Name,
		Description:       req.Description,
		EntityID:          entityID,
		EntityTypeID:      entity.EntityTypeID,
		TriggerTypeID:     triggerTypeID,
		TriggerConditions: triggerConditions,
		CanvasLayout:      req.CanvasLayout,
		IsActive:          req.IsActive,
		CreatedBy:         userID,
	}

	rule, err := bus.CreateRule(ctx, newRule)
	if err != nil {
		return workflow.AutomationRule{}, errs.Newf(errs.Internal, "create rule: %s", err)
	}

	return rule, nil
}

// getEntityType fetches entity info to get the EntityTypeID.
// This is needed because CreateRule requires EntityTypeID.
func (a *App) getEntityType(ctx context.Context, bus *workflow.Business, entityID uuid.UUID) (workflow.Entity, error) {
	entities, err := bus.QueryEntities(ctx)
	if err != nil {
		return workflow.Entity{}, errs.Newf(errs.Internal, "query entities: %s", err)
	}

	for _, e := range entities {
		if e.ID == entityID {
			return e, nil
		}
	}

	return workflow.Entity{}, errs.Newf(errs.NotFound, "entity not found: %s", entityID)
}

// syncActions synchronizes actions: creates new, updates existing, deletes removed.
// Returns a map from temp:N or existing UUID to the final UUID.
func (a *App) syncActions(ctx context.Context, bus *workflow.Business, ruleID uuid.UUID, actions []SaveActionRequest) (map[string]uuid.UUID, []workflow.RuleAction, error) {
	// Get existing actions for this rule
	existingActions, err := bus.QueryActionsByRule(ctx, ruleID)
	if err != nil {
		return nil, nil, errs.Newf(errs.Internal, "query existing actions: %s", err)
	}

	// Build map of existing action IDs
	existingMap := make(map[uuid.UUID]workflow.RuleAction)
	for _, action := range existingActions {
		existingMap[action.ID] = action
	}

	// Track which actions are referenced in the request
	referencedIDs := make(map[uuid.UUID]bool)
	actionIDMap := make(map[string]uuid.UUID)
	var savedActions []workflow.RuleAction

	for i, reqAction := range actions {
		tempKey := fmt.Sprintf("temp:%d", i)

		if reqAction.ID != nil && *reqAction.ID != "" {
			// Update existing action
			actionID, err := uuid.Parse(*reqAction.ID)
			if err != nil {
				return nil, nil, errs.Newf(errs.InvalidArgument, "invalid action id: %s", err)
			}

			existing, exists := existingMap[actionID]
			if !exists {
				return nil, nil, errs.Newf(errs.InvalidArgument, "action %s does not belong to this rule", actionID)
			}

			// Determine action type from config
			actionType := getActionTypeFromConfig(reqAction.ActionConfig)

			update := workflow.UpdateRuleAction{
				Name:         &reqAction.Name,
				Description:  &reqAction.Description,
				ActionConfig: &reqAction.ActionConfig,
				IsActive:     &reqAction.IsActive,
			}

			// Store action type in config if not already present
			if actionType == "" {
				actionType = reqAction.ActionType
			}
			configWithType, err := ensureActionTypeInConfig(reqAction.ActionConfig, reqAction.ActionType)
			if err == nil {
				update.ActionConfig = &configWithType
			}

			updated, err := bus.UpdateRuleAction(ctx, existing, update)
			if err != nil {
				return nil, nil, errs.Newf(errs.Internal, "update action: %s", err)
			}

			referencedIDs[actionID] = true
			actionIDMap[tempKey] = actionID
			actionIDMap[actionID.String()] = actionID
			savedActions = append(savedActions, updated)
		} else {
			// Create new action
			configWithType, err := ensureActionTypeInConfig(reqAction.ActionConfig, reqAction.ActionType)
			if err != nil {
				return nil, nil, errs.Newf(errs.InvalidArgument, "prepare action config: %s", err)
			}

			newAction := workflow.NewRuleAction{
				AutomationRuleID: ruleID,
				Name:             reqAction.Name,
				Description:      reqAction.Description,
				ActionConfig:     configWithType,
				IsActive:         reqAction.IsActive,
			}

			created, err := bus.CreateRuleAction(ctx, newAction)
			if err != nil {
				return nil, nil, errs.Newf(errs.Internal, "create action: %s", err)
			}

			actionIDMap[tempKey] = created.ID
			savedActions = append(savedActions, created)
		}
	}

	// Delete actions that are no longer referenced
	for id := range existingMap {
		if !referencedIDs[id] {
			if err := bus.DeactivateRuleAction(ctx, existingMap[id]); err != nil {
				return nil, nil, errs.Newf(errs.Internal, "delete action: %s", err)
			}
		}
	}

	return actionIDMap, savedActions, nil
}

// createAllActions creates all actions for a new workflow.
func (a *App) createAllActions(ctx context.Context, bus *workflow.Business, ruleID uuid.UUID, actions []SaveActionRequest) (map[string]uuid.UUID, []workflow.RuleAction, error) {
	actionIDMap := make(map[string]uuid.UUID)
	var savedActions []workflow.RuleAction

	for i, reqAction := range actions {
		tempKey := fmt.Sprintf("temp:%d", i)

		configWithType, err := ensureActionTypeInConfig(reqAction.ActionConfig, reqAction.ActionType)
		if err != nil {
			return nil, nil, errs.Newf(errs.InvalidArgument, "prepare action config: %s", err)
		}

		newAction := workflow.NewRuleAction{
			AutomationRuleID: ruleID,
			Name:             reqAction.Name,
			Description:      reqAction.Description,
			ActionConfig:     configWithType,
			IsActive:         reqAction.IsActive,
		}

		created, err := bus.CreateRuleAction(ctx, newAction)
		if err != nil {
			return nil, nil, errs.Newf(errs.Internal, "create action: %s", err)
		}

		actionIDMap[tempKey] = created.ID
		savedActions = append(savedActions, created)
	}

	return actionIDMap, savedActions, nil
}

// createEdges creates action edges with temp ID resolution.
func (a *App) createEdges(ctx context.Context, bus *workflow.Business, ruleID uuid.UUID, edges []SaveEdgeRequest, actionIDMap map[string]uuid.UUID) ([]workflow.ActionEdge, error) {
	var savedEdges []workflow.ActionEdge

	for i, reqEdge := range edges {
		// Resolve target action ID
		targetID, err := resolveActionID(reqEdge.TargetActionID, actionIDMap)
		if err != nil {
			return nil, errs.Newf(errs.InvalidArgument, "edge[%d]: %s", i, err)
		}

		// Resolve source action ID (can be nil for start edges)
		var sourceID *uuid.UUID
		if reqEdge.SourceActionID != "" {
			resolved, err := resolveActionID(reqEdge.SourceActionID, actionIDMap)
			if err != nil {
				return nil, errs.Newf(errs.InvalidArgument, "edge[%d]: %s", i, err)
			}
			sourceID = &resolved
		}

		var sourceOutput *string
		if reqEdge.SourceOutput != "" {
			s := reqEdge.SourceOutput
			sourceOutput = &s
		}

		newEdge := workflow.NewActionEdge{
			RuleID:         ruleID,
			SourceActionID: sourceID,
			TargetActionID: targetID,
			EdgeType:       reqEdge.EdgeType,
			SourceOutput:   sourceOutput,
			EdgeOrder:      reqEdge.EdgeOrder,
		}

		created, err := bus.CreateActionEdge(ctx, newEdge)
		if err != nil {
			return nil, errs.Newf(errs.Internal, "create edge: %s", err)
		}

		savedEdges = append(savedEdges, created)
	}

	return savedEdges, nil
}

// resolveActionID converts a temp:N reference or UUID string to a real UUID.
func resolveActionID(ref string, actionIDMap map[string]uuid.UUID) (uuid.UUID, error) {
	// Check if it's a temp reference
	if strings.HasPrefix(ref, "temp:") {
		id, exists := actionIDMap[ref]
		if !exists {
			// Try to parse the index
			indexStr := strings.TrimPrefix(ref, "temp:")
			index, err := strconv.Atoi(indexStr)
			if err != nil {
				return uuid.Nil, fmt.Errorf("invalid temp reference: %s", ref)
			}
			return uuid.Nil, fmt.Errorf("temp index %d not found in action map", index)
		}
		return id, nil
	}

	// Try to parse as UUID
	id, err := uuid.Parse(ref)
	if err != nil {
		return uuid.Nil, fmt.Errorf("invalid action reference: %s", ref)
	}

	// Check if UUID is in map (for existing actions)
	if mapped, exists := actionIDMap[ref]; exists {
		return mapped, nil
	}

	return id, nil
}

// ensureActionTypeInConfig adds action_type to the config JSON if not present.
func ensureActionTypeInConfig(config json.RawMessage, actionType string) (json.RawMessage, error) {
	var configMap map[string]interface{}
	if err := json.Unmarshal(config, &configMap); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}

	if _, exists := configMap["action_type"]; !exists {
		configMap["action_type"] = actionType
	}

	result, err := json.Marshal(configMap)
	if err != nil {
		return nil, fmt.Errorf("marshal config: %w", err)
	}

	return result, nil
}

// getActionTypeFromConfig extracts the action_type field from config JSON.
func getActionTypeFromConfig(config json.RawMessage) string {
	var configMap map[string]interface{}
	if err := json.Unmarshal(config, &configMap); err != nil {
		return ""
	}

	if actionType, ok := configMap["action_type"].(string); ok {
		return actionType
	}

	return ""
}

// DuplicateWorkflow deep-copies an existing workflow (rule + active actions + edges).
// The duplicate gets name "{original}-DUPLICATE", is_active=false, is_default=false,
// and is owned by the requesting user.
func (a *App) DuplicateWorkflow(ctx context.Context, ruleID uuid.UUID, userID uuid.UUID) (SaveWorkflowResponse, error) {
	// 1. Fetch source rule
	sourceRule, err := a.workflowBus.QueryRuleByID(ctx, ruleID)
	if err != nil {
		return SaveWorkflowResponse{}, errs.Newf(errs.NotFound, "rule not found: %s", err)
	}

	// 2. Fetch active actions only
	allActions, err := a.workflowBus.QueryActionsByRule(ctx, ruleID)
	if err != nil {
		return SaveWorkflowResponse{}, errs.Newf(errs.Internal, "query actions: %s", err)
	}

	var activeActions []workflow.RuleAction
	for _, action := range allActions {
		if action.IsActive {
			activeActions = append(activeActions, action)
		}
	}

	// 3. Fetch edges
	edges, err := a.workflowBus.QueryEdgesByRuleID(ctx, ruleID)
	if err != nil {
		return SaveWorkflowResponse{}, errs.Newf(errs.Internal, "query edges: %s", err)
	}

	// 4. Begin transaction
	tx, err := a.db.BeginTxx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return SaveWorkflowResponse{}, errs.Newf(errs.Internal, "begin tx: %s", err)
	}
	defer func() {
		_ = tx.Rollback()
	}()

	txBus, err := a.workflowBus.NewWithTx(tx)
	if err != nil {
		return SaveWorkflowResponse{}, errs.Newf(errs.Internal, "new with tx: %s", err)
	}

	// 5. Create duplicate rule
	newRule := workflow.NewAutomationRule{
		Name:              sourceRule.Name + "-DUPLICATE",
		Description:       sourceRule.Description,
		EntityID:          sourceRule.EntityID,
		EntityTypeID:      sourceRule.EntityTypeID,
		TriggerTypeID:     sourceRule.TriggerTypeID,
		TriggerConditions: sourceRule.TriggerConditions,
		CanvasLayout:      sourceRule.CanvasLayout,
		IsActive:          false,
		IsDefault:         false,
		CreatedBy:         userID,
	}

	rule, err := txBus.CreateRule(ctx, newRule)
	if err != nil {
		return SaveWorkflowResponse{}, errs.Newf(errs.Internal, "create rule: %s", err)
	}

	// 6. Create actions and build oldID→newID map
	oldToNewID := make(map[uuid.UUID]uuid.UUID)
	var savedActions []workflow.RuleAction

	for _, srcAction := range activeActions {
		newAction := workflow.NewRuleAction{
			AutomationRuleID: rule.ID,
			Name:             srcAction.Name,
			Description:      srcAction.Description,
			ActionConfig:     srcAction.ActionConfig,
			IsActive:         srcAction.IsActive,
			TemplateID:       srcAction.TemplateID,
		}

		created, err := txBus.CreateRuleAction(ctx, newAction)
		if err != nil {
			return SaveWorkflowResponse{}, errs.Newf(errs.Internal, "create action: %s", err)
		}

		oldToNewID[srcAction.ID] = created.ID
		savedActions = append(savedActions, created)
	}

	// 7. Create edges, remapping IDs; skip edges whose source or target was inactive
	var savedEdges []workflow.ActionEdge

	for _, edge := range edges {
		// Remap target — skip if target action was inactive
		newTargetID, ok := oldToNewID[edge.TargetActionID]
		if !ok {
			continue
		}

		// Remap source (nil for start edges)
		var newSourceID *uuid.UUID
		if edge.SourceActionID != nil {
			mapped, ok := oldToNewID[*edge.SourceActionID]
			if !ok {
				continue // source action was inactive
			}
			newSourceID = &mapped
		}

		newEdge := workflow.NewActionEdge{
			RuleID:         rule.ID,
			SourceActionID: newSourceID,
			TargetActionID: newTargetID,
			EdgeType:       edge.EdgeType,
			SourceOutput:   edge.SourceOutput,
			EdgeOrder:      edge.EdgeOrder,
		}

		created, err := txBus.CreateActionEdge(ctx, newEdge)
		if err != nil {
			return SaveWorkflowResponse{}, errs.Newf(errs.Internal, "create edge: %s", err)
		}

		savedEdges = append(savedEdges, created)
	}

	// 8. Commit transaction
	if err := tx.Commit(); err != nil {
		return SaveWorkflowResponse{}, errs.Newf(errs.Internal, "commit: %s", err)
	}

	// 9. Fire delegate event AFTER commit
	if a.delegate != nil {
		a.log.Info(ctx, "workflowsaveapp: duplicate committed, firing delegate event", "ruleID", rule.ID, "action", "created")
		if err := a.delegate.Call(ctx, workflow.ActionRuleChangedData(workflow.ActionRuleCreated, rule.ID)); err != nil {
			a.log.Error(ctx, "workflowsaveapp: delegate call failed", "action", workflow.ActionRuleCreated, "err", err)
		}
	}

	// 10. Return response
	return buildResponse(rule, savedActions, savedEdges, sourceRule.CanvasLayout), nil
}

// buildResponse constructs the SaveWorkflowResponse from business layer objects.
func buildResponse(rule workflow.AutomationRule, actions []workflow.RuleAction, edges []workflow.ActionEdge, canvasLayout json.RawMessage) SaveWorkflowResponse {
	// Convert actions
	actionResponses := make([]SaveActionResponse, len(actions))
	for i, action := range actions {
		actionResponses[i] = SaveActionResponse{
			ID:             action.ID.String(),
			Name:           action.Name,
			Description:    action.Description,
			ActionType:   getActionTypeFromConfig(action.ActionConfig),
			ActionConfig: action.ActionConfig,
			IsActive:     action.IsActive,
		}
	}

	// Convert edges
	edgeResponses := make([]SaveEdgeResponse, len(edges))
	for i, edge := range edges {
		sourceID := ""
		if edge.SourceActionID != nil {
			sourceID = edge.SourceActionID.String()
		}

		sourceOutput := ""
		if edge.SourceOutput != nil {
			sourceOutput = *edge.SourceOutput
		}

		edgeResponses[i] = SaveEdgeResponse{
			ID:             edge.ID.String(),
			SourceActionID: sourceID,
			TargetActionID: edge.TargetActionID.String(),
			EdgeType:       edge.EdgeType,
			SourceOutput:   sourceOutput,
			EdgeOrder:      edge.EdgeOrder,
		}
	}

	// Handle trigger conditions
	var triggerConditions json.RawMessage
	if rule.TriggerConditions != nil {
		triggerConditions = *rule.TriggerConditions
	}

	// Use passed canvas layout or rule's canvas layout
	layout := canvasLayout
	if len(layout) == 0 {
		layout = rule.CanvasLayout
	}

	return SaveWorkflowResponse{
		ID:                rule.ID.String(),
		Name:              rule.Name,
		Description:       rule.Description,
		IsActive:          rule.IsActive,
		EntityID:          rule.EntityID.String(),
		TriggerTypeID:     rule.TriggerTypeID.String(),
		TriggerConditions: triggerConditions,
		Actions:           actionResponses,
		Edges:             edgeResponses,
		CanvasLayout:      layout,
		CreatedDate:       rule.CreatedDate.Format(time.RFC3339),
		UpdatedDate:       rule.UpdatedDate.Format(time.RFC3339),
	}
}
