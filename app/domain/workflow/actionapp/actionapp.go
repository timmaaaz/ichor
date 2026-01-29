// Package actionapp provides the application layer for manual action execution.
package actionapp

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/business/domain/workflow/actionpermissionsbus"
	"github.com/timmaaaz/ichor/business/sdk/workflow"
)

// App manages the set of app layer API functions for manual action execution.
type App struct {
	actionService *workflow.ActionService
	actionPermBus *actionpermissionsbus.Business
}

// NewApp constructs an action app API for use.
func NewApp(actionService *workflow.ActionService, actionPermBus *actionpermissionsbus.Business) *App {
	return &App{
		actionService: actionService,
		actionPermBus: actionPermBus,
	}
}

// Execute performs a manual action execution with permission checking.
func (a *App) Execute(ctx context.Context, actionType string, req ExecuteRequest, userID uuid.UUID, userRoles []uuid.UUID) (ExecuteResponse, error) {
	// Check if user has permission to execute this action
	canExecute, err := a.actionPermBus.CanUserExecuteAction(ctx, userID, actionType, userRoles)
	if err != nil {
		return ExecuteResponse{}, errs.Newf(errs.Internal, "check permission: %s", err)
	}

	if !canExecute {
		return ExecuteResponse{}, errs.New(errs.PermissionDenied, errors.New("user does not have permission to execute this action"))
	}

	// Parse entity ID if provided
	var entityID *uuid.UUID
	if req.EntityID != nil && *req.EntityID != "" {
		id, err := uuid.Parse(*req.EntityID)
		if err != nil {
			return ExecuteResponse{}, errs.NewFieldsError("entityId", err)
		}
		entityID = &id
	}

	// Build business layer request
	busReq := workflow.ExecuteRequest{
		ActionType: actionType,
		Config:     req.Config,
		EntityID:   entityID,
		EntityName: req.EntityName,
		RawData:    req.RawData,
		UserID:     userID,
	}

	// Execute the action
	result, err := a.actionService.Execute(ctx, busReq, workflow.TriggerSourceManual)
	if err != nil {
		// Check for specific action service errors
		if errors.Is(err, workflow.ErrActionNotFound) {
			return ExecuteResponse{}, errs.New(errs.NotFound, workflow.ErrActionNotFound)
		}
		if errors.Is(err, workflow.ErrManualExecutionNotSupported) {
			return ExecuteResponse{}, errs.New(errs.FailedPrecondition, workflow.ErrManualExecutionNotSupported)
		}
		// For validation errors, return as InvalidArgument
		return toAppExecuteResponse(result), errs.New(errs.InvalidArgument, err)
	}

	return toAppExecuteResponse(result), nil
}

// ListAvailable returns all actions that the user can execute manually.
// This filters the list based on the user's role permissions.
func (a *App) ListAvailable(ctx context.Context, userID uuid.UUID, userRoles []uuid.UUID) (AvailableActions, error) {
	// Get all actions that support manual execution
	allActions := a.actionService.ListManuallyExecutableActions()

	// Get allowed action types for the user's roles
	allowedTypes, err := a.actionPermBus.GetAllowedActionsForRoles(ctx, userRoles)
	if err != nil {
		return nil, errs.Newf(errs.Internal, "get allowed actions: %s", err)
	}

	// Build a set of allowed action types for fast lookup
	allowedSet := make(map[string]bool)
	for _, t := range allowedTypes {
		allowedSet[t] = true
	}

	// Filter actions to only those the user can execute
	var permitted []workflow.ActionInfo
	for _, action := range allActions {
		if allowedSet[action.Type] {
			permitted = append(permitted, action)
		}
	}

	return toAppAvailableActions(permitted), nil
}

// GetExecutionStatus retrieves the status of an action execution by ID.
func (a *App) GetExecutionStatus(ctx context.Context, executionID uuid.UUID) (ExecutionStatus, error) {
	status, err := a.actionService.GetExecutionStatus(ctx, executionID)
	if err != nil {
		return ExecutionStatus{}, errs.Newf(errs.NotFound, "execution not found: %s", err)
	}

	return toAppExecutionStatus(status), nil
}
