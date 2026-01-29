// Package actionapi provides HTTP handlers for manual workflow action execution.
package actionapi

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/domain/workflow/actionapp"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/app/sdk/mid"
	"github.com/timmaaaz/ichor/business/domain/core/userrolebus"
	"github.com/timmaaaz/ichor/foundation/web"
)

type api struct {
	actionApp   *actionapp.App
	userRoleBus *userrolebus.Business
}

func newAPI(actionApp *actionapp.App, userRoleBus *userrolebus.Business) *api {
	return &api{
		actionApp:   actionApp,
		userRoleBus: userRoleBus,
	}
}

// list returns all actions that the authenticated user can execute manually.
// This filters based on the user's role permissions.
func (a *api) list(ctx context.Context, r *http.Request) web.Encoder {
	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return errs.New(errs.Unauthenticated, err)
	}

	roleIDs, err := a.getUserRoleIDs(ctx, userID)
	if err != nil {
		return errs.Newf(errs.Internal, "get user roles: %s", err)
	}

	actions, err := a.actionApp.ListAvailable(ctx, userID, roleIDs)
	if err != nil {
		return errs.NewError(err)
	}

	return actions
}

// execute manually triggers a workflow action.
// The action type is specified in the URL path, and the configuration is in the request body.
func (a *api) execute(ctx context.Context, r *http.Request) web.Encoder {
	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return errs.New(errs.Unauthenticated, err)
	}

	roleIDs, err := a.getUserRoleIDs(ctx, userID)
	if err != nil {
		return errs.Newf(errs.Internal, "get user roles: %s", err)
	}

	actionType := web.Param(r, "actionType")
	if actionType == "" {
		return errs.New(errs.InvalidArgument, errs.NewFieldsError("actionType", nil))
	}

	var req actionapp.ExecuteRequest
	if err := web.Decode(r, &req); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	if err := req.Validate(); err != nil {
		return errs.NewError(err)
	}

	response, err := a.actionApp.Execute(ctx, actionType, req, userID, roleIDs)
	if err != nil {
		return errs.NewError(err)
	}

	return response
}

// getExecutionStatus retrieves the status of an action execution.
// This is useful for tracking async actions.
func (a *api) getExecutionStatus(ctx context.Context, r *http.Request) web.Encoder {
	_, err := mid.GetUserID(ctx)
	if err != nil {
		return errs.New(errs.Unauthenticated, err)
	}

	executionIDStr := web.Param(r, "executionId")
	executionID, err := uuid.Parse(executionIDStr)
	if err != nil {
		return errs.New(errs.InvalidArgument, errs.NewFieldsError("executionId", err))
	}

	status, err := a.actionApp.GetExecutionStatus(ctx, executionID)
	if err != nil {
		return errs.NewError(err)
	}

	return status
}

// getUserRoleIDs fetches role IDs for the current user.
func (a *api) getUserRoleIDs(ctx context.Context, userID uuid.UUID) ([]uuid.UUID, error) {
	userRoles, err := a.userRoleBus.QueryByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	roleIDs := make([]uuid.UUID, len(userRoles))
	for i, ur := range userRoles {
		roleIDs[i] = ur.RoleID
	}
	return roleIDs, nil
}
