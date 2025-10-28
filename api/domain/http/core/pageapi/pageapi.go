package pageapi

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/domain/core/pageapp"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/foundation/web"
)

type api struct {
	pageapp *pageapp.App
}

func newAPI(pageapp *pageapp.App) *api {
	return &api{
		pageapp: pageapp,
	}
}

func (api *api) create(ctx context.Context, r *http.Request) web.Encoder {
	var app pageapp.NewPage
	if err := web.Decode(r, &app); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	page, err := api.pageapp.Create(ctx, app)
	if err != nil {
		return errs.NewError(err)
	}

	return page
}

func (api *api) update(ctx context.Context, r *http.Request) web.Encoder {
	var app pageapp.UpdatePage
	if err := web.Decode(r, &app); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	pageID := web.Param(r, "page_id")
	parsed, err := uuid.Parse(pageID)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	page, err := api.pageapp.Update(ctx, app, parsed)
	if err != nil {
		return errs.NewError(err)
	}

	return page
}

func (api *api) delete(ctx context.Context, r *http.Request) web.Encoder {
	pageID := web.Param(r, "page_id")
	parsed, err := uuid.Parse(pageID)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	err = api.pageapp.Delete(ctx, parsed)
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

	page, err := api.pageapp.Query(ctx, qp)
	if err != nil {
		return errs.NewError(err)
	}

	return page
}

func (api *api) queryByID(ctx context.Context, r *http.Request) web.Encoder {
	pageID := web.Param(r, "page_id")
	parsed, err := uuid.Parse(pageID)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	page, err := api.pageapp.QueryByID(ctx, parsed)
	if err != nil {
		return errs.NewError(err)
	}

	return page
}
