package zoneapi

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/domain/inventory/zoneapp"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/foundation/web"
)

type api struct {
	zoneapp *zoneapp.App
}

func newAPI(zone *zoneapp.App) *api {
	return &api{zoneapp: zone}
}

func (api *api) create(ctx context.Context, r *http.Request) web.Encoder {
	var app zoneapp.NewZone
	if err := web.Decode(r, &app); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	zone, err := api.zoneapp.Create(ctx, app)
	if err != nil {
		return errs.NewError(err)
	}

	return zone
}

func (api *api) update(ctx context.Context, r *http.Request) web.Encoder {
	var app zoneapp.UpdateZone
	if err := web.Decode(r, &app); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	zoneID := web.Param(r, "zone_id")
	parsed, err := uuid.Parse(zoneID)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	zone, err := api.zoneapp.Update(ctx, app, parsed)
	if err != nil {
		return errs.NewError(err)
	}

	return zone
}

func (api *api) delete(ctx context.Context, r *http.Request) web.Encoder {
	zoneID := web.Param(r, "zone_id")
	parsed, err := uuid.Parse(zoneID)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	err = api.zoneapp.Delete(ctx, parsed)
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

	zones, err := api.zoneapp.Query(ctx, qp)
	if err != nil {
		return errs.NewError(err)
	}

	return zones
}

func (api *api) queryByID(ctx context.Context, r *http.Request) web.Encoder {
	zoneID := web.Param(r, "zone_id")
	parsed, err := uuid.Parse(zoneID)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	zone, err := api.zoneapp.QueryByID(ctx, parsed)
	if err != nil {
		return errs.NewError(err)
	}

	return zone
}
