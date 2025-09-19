package reportstoapi

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/domain/hr/reportstoapp"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/foundation/web"
)

type api struct {
	reportstoapp *reportstoapp.App
}

func newAPI(reportstoapp *reportstoapp.App) *api {
	return &api{
		reportstoapp: reportstoapp,
	}
}

func (api *api) create(ctx context.Context, r *http.Request) web.Encoder {
	var app reportstoapp.NewReportsTo
	if err := web.Decode(r, &app); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	reportsTo, err := api.reportstoapp.Create(ctx, app)
	if err != nil {
		return errs.NewError(err)
	}

	return reportsTo
}

func (api *api) update(ctx context.Context, r *http.Request) web.Encoder {
	var app reportstoapp.UpdateReportsTo
	if err := web.Decode(r, &app); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	reportsToID := web.Param(r, "reports_to_id")
	parsed, err := uuid.Parse(reportsToID)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	reportsTo, err := api.reportstoapp.Update(ctx, app, parsed)
	if err != nil {
		return errs.NewError(err)
	}

	return reportsTo
}

func (api *api) delete(ctx context.Context, r *http.Request) web.Encoder {
	reportsToID := web.Param(r, "reports_to_id")

	parsed, err := uuid.Parse(reportsToID)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	err = api.reportstoapp.Delete(ctx, parsed)
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

	result, err := api.reportstoapp.Query(ctx, qp)
	if err != nil {
		return errs.NewError(err)
	}

	return result
}

func (api *api) queryByID(ctx context.Context, r *http.Request) web.Encoder {
	reportsToID := web.Param(r, "reports_to_id")

	parsed, err := uuid.Parse(reportsToID)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	result, err := api.reportstoapp.QueryByID(ctx, parsed)
	if err != nil {
		return errs.NewError(err)
	}

	return result
}
