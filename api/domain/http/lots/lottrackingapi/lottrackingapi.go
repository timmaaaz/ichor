package lottrackingapi

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/domain/lots/lottrackingapp"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/foundation/web"
)

type api struct {
	lottrackingapp *lottrackingapp.App
}

func newAPI(lotTrackingApp *lottrackingapp.App) *api {
	return &api{
		lottrackingapp: lotTrackingApp,
	}
}

func (api *api) create(ctx context.Context, r *http.Request) web.Encoder {
	var app lottrackingapp.NewLotTracking
	if err := web.Decode(r, &app); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	lotTracking, err := api.lottrackingapp.Create(ctx, app)
	if err != nil {
		return errs.NewError(err)
	}

	return lotTracking
}

func (api *api) update(ctx context.Context, r *http.Request) web.Encoder {
	var app lottrackingapp.UpdateLotTracking
	if err := web.Decode(r, &app); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	lotTrackingID := web.Param(r, "lot_id")
	parsed, err := uuid.Parse(lotTrackingID)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	lotTracking, err := api.lottrackingapp.Update(ctx, app, parsed)
	if err != nil {
		return errs.NewError(err)
	}

	return lotTracking
}

func (api *api) delete(ctx context.Context, r *http.Request) web.Encoder {
	lotTrackingID := web.Param(r, "lot_id")
	parsed, err := uuid.Parse(lotTrackingID)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	err = api.lottrackingapp.Delete(ctx, parsed)
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

	lotTrackings, err := api.lottrackingapp.Query(ctx, qp)
	if err != nil {
		return errs.NewError(err)
	}

	return lotTrackings
}

func (api *api) queryByID(ctx context.Context, r *http.Request) web.Encoder {
	lotTrackingID := web.Param(r, "lot_id")
	parsed, err := uuid.Parse(lotTrackingID)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	lotTracking, err := api.lottrackingapp.QueryByID(ctx, parsed)
	if err != nil {
		return errs.NewError(err)
	}

	return lotTracking
}
