package pageactionapi

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/domain/config/pageactionapp"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/foundation/web"
)

type api struct {
	pageactionapp *pageactionapp.App
}

func newAPI(pageactionapp *pageactionapp.App) *api {
	return &api{
		pageactionapp: pageactionapp,
	}
}

// =============================================================================
// Create Handlers
// =============================================================================

func (api *api) createButton(ctx context.Context, r *http.Request) web.Encoder {
	var app pageactionapp.NewButtonAction
	if err := web.Decode(r, &app); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	action, err := api.pageactionapp.CreateButton(ctx, app)
	if err != nil {
		return errs.NewError(err)
	}

	return action
}

func (api *api) createDropdown(ctx context.Context, r *http.Request) web.Encoder {
	var app pageactionapp.NewDropdownAction
	if err := web.Decode(r, &app); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	action, err := api.pageactionapp.CreateDropdown(ctx, app)
	if err != nil {
		return errs.NewError(err)
	}

	return action
}

func (api *api) createSeparator(ctx context.Context, r *http.Request) web.Encoder {
	var app pageactionapp.NewSeparatorAction
	if err := web.Decode(r, &app); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	action, err := api.pageactionapp.CreateSeparator(ctx, app)
	if err != nil {
		return errs.NewError(err)
	}

	return action
}

// =============================================================================
// Update Handlers
// =============================================================================

func (api *api) updateButton(ctx context.Context, r *http.Request) web.Encoder {
	var app pageactionapp.UpdateButtonAction
	if err := web.Decode(r, &app); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	actionID := web.Param(r, "action_id")
	parsed, err := uuid.Parse(actionID)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	action, err := api.pageactionapp.UpdateButton(ctx, app, parsed)
	if err != nil {
		return errs.NewError(err)
	}

	return action
}

func (api *api) updateDropdown(ctx context.Context, r *http.Request) web.Encoder {
	var app pageactionapp.UpdateDropdownAction
	if err := web.Decode(r, &app); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	actionID := web.Param(r, "action_id")
	parsed, err := uuid.Parse(actionID)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	action, err := api.pageactionapp.UpdateDropdown(ctx, app, parsed)
	if err != nil {
		return errs.NewError(err)
	}

	return action
}

func (api *api) updateSeparator(ctx context.Context, r *http.Request) web.Encoder {
	var app pageactionapp.UpdateSeparatorAction
	if err := web.Decode(r, &app); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	actionID := web.Param(r, "action_id")
	parsed, err := uuid.Parse(actionID)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	action, err := api.pageactionapp.UpdateSeparator(ctx, app, parsed)
	if err != nil {
		return errs.NewError(err)
	}

	return action
}

// =============================================================================
// Delete Handler
// =============================================================================

func (api *api) delete(ctx context.Context, r *http.Request) web.Encoder {
	actionID := web.Param(r, "action_id")
	parsed, err := uuid.Parse(actionID)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	err = api.pageactionapp.Delete(ctx, parsed)
	if err != nil {
		return errs.NewError(err)
	}

	return nil
}

// =============================================================================
// Query Handlers
// =============================================================================

func (api *api) query(ctx context.Context, r *http.Request) web.Encoder {
	qp, err := parseQueryParams(r)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	actions, err := api.pageactionapp.Query(ctx, qp)
	if err != nil {
		return errs.NewError(err)
	}

	return actions
}

func (api *api) queryByID(ctx context.Context, r *http.Request) web.Encoder {
	actionID := web.Param(r, "action_id")
	parsed, err := uuid.Parse(actionID)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	action, err := api.pageactionapp.QueryByID(ctx, parsed)
	if err != nil {
		return errs.NewError(err)
	}

	return action
}

func (api *api) queryByPageConfigID(ctx context.Context, r *http.Request) web.Encoder {
	pageConfigID := web.Param(r, "page_config_id")
	parsed, err := uuid.Parse(pageConfigID)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	actions, err := api.pageactionapp.QueryByPageConfigID(ctx, parsed)
	if err != nil {
		return errs.NewError(err)
	}

	return actions
}

// =============================================================================
// Batch Operations
// =============================================================================

func (api *api) batchCreate(ctx context.Context, r *http.Request) web.Encoder {
	var app pageactionapp.BatchCreateRequest
	if err := web.Decode(r, &app); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	actions, err := api.pageactionapp.BatchCreate(ctx, app)
	if err != nil {
		return errs.NewError(err)
	}

	return actions
}
