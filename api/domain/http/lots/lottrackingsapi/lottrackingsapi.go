package lottrackingsapi

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/domain/lots/lottrackingsapp"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/foundation/web"
)

type api struct {
	lottrackingsapp *lottrackingsapp.App
}

func newAPI(lotTrackingsApp *lottrackingsapp.App) *api {
	return &api{
		lottrackingsapp: lotTrackingsApp,
	}
}

func (api *api) create(ctx context.Context, r *http.Request) web.Encoder {
	var app lottrackingsapp.NewLotTrackings
	if err := web.Decode(r, &app); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	lotTrackings, err := api.lottrackingsapp.Create(ctx, app)
	if err != nil {
		return errs.NewError(err)
	}

	return lotTrackings
}

func (api *api) update(ctx context.Context, r *http.Request) web.Encoder {
	var app lottrackingsapp.UpdateLotTrackings
	if err := web.Decode(r, &app); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	lotTrackingsID := web.Param(r, "lot_id")
	parsed, err := uuid.Parse(lotTrackingsID)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	lotTrackings, err := api.lottrackingsapp.Update(ctx, app, parsed)
	if err != nil {
		return errs.NewError(err)
	}

	return lotTrackings
}

func (api *api) delete(ctx context.Context, r *http.Request) web.Encoder {
	lotTrackingsID := web.Param(r, "lot_id")
	parsed, err := uuid.Parse(lotTrackingsID)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	err = api.lottrackingsapp.Delete(ctx, parsed)
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

	lotTrackingss, err := api.lottrackingsapp.Query(ctx, qp)
	if err != nil {
		return errs.NewError(err)
	}

	return lotTrackingss
}

func (api *api) queryByID(ctx context.Context, r *http.Request) web.Encoder {
	lotTrackingsID := web.Param(r, "lot_id")
	parsed, err := uuid.Parse(lotTrackingsID)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	lotTrackings, err := api.lottrackingsapp.QueryByID(ctx, parsed)
	if err != nil {
		return errs.NewError(err)
	}

	return lotTrackings
}
