package userroleapi

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/domain/core/userroleapp"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/foundation/web"
)

type api struct {
	userroleapp *userroleapp.App
}

func newAPI(userroleapp *userroleapp.App) *api {
	return &api{
		userroleapp: userroleapp,
	}
}

func (api *api) create(ctx context.Context, r *http.Request) web.Encoder {
	var app userroleapp.NewUserRole
	if err := web.Decode(r, &app); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	// Validate

	userrole, err := api.userroleapp.Create(ctx, app)
	if err != nil {
		return errs.NewError(err)
	}

	return userrole
}

func (api *api) delete(ctx context.Context, r *http.Request) web.Encoder {
	userroleID := web.Param(r, "user_role_id")
	parsed, err := uuid.Parse(userroleID)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	err = api.userroleapp.Delete(ctx, parsed)
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

	userroles, err := api.userroleapp.Query(ctx, qp)
	if err != nil {
		return errs.NewError(err)
	}

	return userroles
}

func (api *api) queryByID(ctx context.Context, r *http.Request) web.Encoder {
	userroleID := web.Param(r, "user_role_id")
	parsed, err := uuid.Parse(userroleID)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	userrole, err := api.userroleapp.QueryByID(ctx, parsed)
	if err != nil {
		return errs.NewError(err)
	}

	return userrole
}
