package lotlocationapi

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/domain/inventory/lotlocationapp"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/foundation/web"
)

type api struct {
	lotlocationapp *lotlocationapp.App
}

func newAPI(lotLocationApp *lotlocationapp.App) *api {
	return &api{
		lotlocationapp: lotLocationApp,
	}
}

func (api *api) create(ctx context.Context, r *http.Request) web.Encoder {
	var app lotlocationapp.NewLotLocation
	if err := web.Decode(r, &app); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	lotLocation, err := api.lotlocationapp.Create(ctx, app)
	if err != nil {
		return errs.NewError(err)
	}

	return lotLocation
}

func (api *api) update(ctx context.Context, r *http.Request) web.Encoder {
	var app lotlocationapp.UpdateLotLocation
	if err := web.Decode(r, &app); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	lotLocationID := web.Param(r, "lot_location_id")
	parsed, err := uuid.Parse(lotLocationID)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	lotLocation, err := api.lotlocationapp.Update(ctx, app, parsed)
	if err != nil {
		return errs.NewError(err)
	}

	return lotLocation
}

func (api *api) delete(ctx context.Context, r *http.Request) web.Encoder {
	lotLocationID := web.Param(r, "lot_location_id")
	parsed, err := uuid.Parse(lotLocationID)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	if err := api.lotlocationapp.Delete(ctx, parsed); err != nil {
		return errs.NewError(err)
	}

	return nil
}

func (api *api) query(ctx context.Context, r *http.Request) web.Encoder {
	qp, err := parseQueryParams(r)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	lotLocations, err := api.lotlocationapp.Query(ctx, qp)
	if err != nil {
		return errs.NewError(err)
	}

	return lotLocations
}

func (api *api) queryByID(ctx context.Context, r *http.Request) web.Encoder {
	lotLocationID := web.Param(r, "lot_location_id")
	parsed, err := uuid.Parse(lotLocationID)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	lotLocation, err := api.lotlocationapp.QueryByID(ctx, parsed)
	if err != nil {
		return errs.NewError(err)
	}

	return lotLocation
}
