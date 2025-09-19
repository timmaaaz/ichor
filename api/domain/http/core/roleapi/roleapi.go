package roleapi

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/domain/core/roleapp"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/foundation/web"
)

type api struct {
	roleapp *roleapp.App
}

func newAPI(roleapp *roleapp.App) *api {
	return &api{
		roleapp: roleapp,
	}
}

func (api *api) create(ctx context.Context, r *http.Request) web.Encoder {
	var app roleapp.NewRole
	if err := web.Decode(r, &app); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	role, err := api.roleapp.Create(ctx, app)
	if err != nil {
		return errs.NewError(err)
	}

	return role
}

func (api *api) update(ctx context.Context, r *http.Request) web.Encoder {
	var app roleapp.UpdateRole
	if err := web.Decode(r, &app); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	roleID := web.Param(r, "role_id")
	parsed, err := uuid.Parse(roleID)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	role, err := api.roleapp.Update(ctx, app, parsed)
	if err != nil {
		return errs.NewError(err)
	}

	return role
}

func (api *api) delete(ctx context.Context, r *http.Request) web.Encoder {
	roleID := web.Param(r, "role_id")
	parsed, err := uuid.Parse(roleID)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	err = api.roleapp.Delete(ctx, parsed)
	if err != nil {
		return errs.NewError(err)
	}

	return nil
}

func (api *api) query(ctx context.Context, r *http.Request) web.Encoder {
	qp, err := parseQueryParams(r)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	role, err := api.roleapp.Query(ctx, qp)
	if err != nil {
		return errs.NewError(err)
	}

	return role
}

func (api *api) queryByID(ctx context.Context, r *http.Request) web.Encoder {
	roleID := web.Param(r, "role_id")
	parsed, err := uuid.Parse(roleID)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	role, err := api.roleapp.QueryByID(ctx, parsed)
	if err != nil {
		return errs.NewError(err)
	}

	return role
}
