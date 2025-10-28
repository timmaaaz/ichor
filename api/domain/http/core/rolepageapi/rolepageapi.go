package rolepageapi

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/domain/core/rolepageapp"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/foundation/web"
)

type api struct {
	rolepageapp *rolepageapp.App
}

func newAPI(rolepageapp *rolepageapp.App) *api {
	return &api{
		rolepageapp: rolepageapp,
	}
}

func (api *api) create(ctx context.Context, r *http.Request) web.Encoder {
	var app rolepageapp.NewRolePage
	if err := web.Decode(r, &app); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	rolePage, err := api.rolepageapp.Create(ctx, app)
	if err != nil {
		return errs.NewError(err)
	}

	return rolePage
}

func (api *api) update(ctx context.Context, r *http.Request) web.Encoder {
	var app rolepageapp.UpdateRolePage
	if err := web.Decode(r, &app); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	rolePageID := web.Param(r, "role_page_id")
	parsed, err := uuid.Parse(rolePageID)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	rolePage, err := api.rolepageapp.Update(ctx, app, parsed)
	if err != nil {
		return errs.NewError(err)
	}

	return rolePage
}

func (api *api) delete(ctx context.Context, r *http.Request) web.Encoder {
	rolePageID := web.Param(r, "role_page_id")
	parsed, err := uuid.Parse(rolePageID)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	err = api.rolepageapp.Delete(ctx, parsed)
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

	rolePage, err := api.rolepageapp.Query(ctx, qp)
	if err != nil {
		return errs.NewError(err)
	}

	return rolePage
}

func (api *api) queryByID(ctx context.Context, r *http.Request) web.Encoder {
	rolePageID := web.Param(r, "role_page_id")
	parsed, err := uuid.Parse(rolePageID)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	rolePage, err := api.rolepageapp.QueryByID(ctx, parsed)
	if err != nil {
		return errs.NewError(err)
	}

	return rolePage
}
