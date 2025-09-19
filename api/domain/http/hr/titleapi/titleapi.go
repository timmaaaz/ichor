package titleapi

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/domain/hr/titleapp"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/foundation/web"
)

type api struct {
	titleapp *titleapp.App
}

func newAPI(titleapp *titleapp.App) *api {
	return &api{
		titleapp: titleapp,
	}
}

func (api *api) create(ctx context.Context, r *http.Request) web.Encoder {
	var app titleapp.NewTitle
	if err := web.Decode(r, &app); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	title, err := api.titleapp.Create(ctx, app)
	if err != nil {
		return errs.NewError(err)
	}

	return title
}

func (api *api) update(ctx context.Context, r *http.Request) web.Encoder {
	var app titleapp.UpdateTitle
	if err := web.Decode(r, &app); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	titleID := web.Param(r, "title_id")
	parsed, err := uuid.Parse(titleID)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	title, err := api.titleapp.Update(ctx, app, parsed)
	if err != nil {
		return errs.NewError(err)
	}

	return title
}

func (api *api) delete(ctx context.Context, r *http.Request) web.Encoder {
	titleID := web.Param(r, "title_id")

	parsed, err := uuid.Parse(titleID)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	err = api.titleapp.Delete(ctx, parsed)
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

	titles, err := api.titleapp.Query(ctx, qp)
	if err != nil {
		return errs.NewError(err)
	}

	return titles
}

func (api *api) queryByID(ctx context.Context, r *http.Request) web.Encoder {
	titleID := web.Param(r, "title_id")

	parsed, err := uuid.Parse(titleID)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	title, err := api.titleapp.QueryByID(ctx, parsed)
	if err != nil {
		return errs.NewError(err)
	}

	return title
}
