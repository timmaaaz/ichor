// Package workflowsaveapi provides HTTP handlers for transactional workflow save operations.
package workflowsaveapi

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/domain/workflow/workflowsaveapp"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/app/sdk/mid"
	"github.com/timmaaaz/ichor/foundation/web"
)

// api holds dependencies for workflow save HTTP handlers.
type api struct {
	app *workflowsaveapp.App
}

// newAPI creates a new workflow save API handler.
func newAPI(app *workflowsaveapp.App) *api {
	return &api{app: app}
}

// save handles PUT /v1/workflow/rules/{id}/full
// Updates an existing workflow atomically (rule + actions + edges).
// Supports ?dry_run=true to validate without committing.
func (a *api) save(ctx context.Context, r *http.Request) web.Encoder {
	ruleID, err := uuid.Parse(web.Param(r, "id"))
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	var req workflowsaveapp.SaveWorkflowRequest
	if err := web.Decode(r, &req); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	if r.URL.Query().Get("dry_run") == "true" {
		return a.app.DryRunValidate(req)
	}

	resp, err := a.app.SaveWorkflow(ctx, ruleID, req)
	if err != nil {
		return errs.NewError(err)
	}

	return resp
}

// duplicate handles POST /v1/workflow/rules/{id}/duplicate
// Deep-copies a workflow (rule + active actions + edges) into a new editable copy.
func (a *api) duplicate(ctx context.Context, r *http.Request) web.Encoder {
	ruleID, err := uuid.Parse(web.Param(r, "id"))
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return errs.New(errs.Unauthenticated, err)
	}

	resp, err := a.app.DuplicateWorkflow(ctx, ruleID, userID)
	if err != nil {
		return errs.NewError(err)
	}

	return resp
}

// create handles POST /v1/workflow/rules/full
// Creates a new workflow atomically (rule + actions + edges).
// Supports ?dry_run=true to validate without committing.
func (a *api) create(ctx context.Context, r *http.Request) web.Encoder {
	var req workflowsaveapp.SaveWorkflowRequest
	if err := web.Decode(r, &req); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	if r.URL.Query().Get("dry_run") == "true" {
		return a.app.DryRunValidate(req)
	}

	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return errs.New(errs.Unauthenticated, err)
	}

	resp, err := a.app.CreateWorkflow(ctx, userID, req)
	if err != nil {
		return errs.NewError(err)
	}

	return resp
}
