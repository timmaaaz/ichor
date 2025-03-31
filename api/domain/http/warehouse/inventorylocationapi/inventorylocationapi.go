package inventorylocationapi

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/domain/warehouse/inventorylocationapp"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/foundation/web"
)

type api struct {
	inventorylocationapp *inventorylocationapp.App
}

func newAPI(location *inventorylocationapp.App) *api {
	return &api{inventorylocationapp: location}
}

func (api *api) create(ctx context.Context, r *http.Request) web.Encoder {
	var app inventorylocationapp.NewInventoryLocation
	if err := web.Decode(r, &app); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	location, err := api.inventorylocationapp.Create(ctx, app)
	if err != nil {
		return errs.NewError(err)
	}

	return location
}

func (api *api) update(ctx context.Context, r *http.Request) web.Encoder {
	var app inventorylocationapp.UpdateInventoryLocation
	if err := web.Decode(r, &app); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	locationID := web.Param(r, "location_id")
	parsed, err := uuid.Parse(locationID)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	location, err := api.inventorylocationapp.Update(ctx, app, parsed)
	if err != nil {
		return errs.NewError(err)
	}

	return location
}

func (api *api) delete(ctx context.Context, r *http.Request) web.Encoder {
	locationID := web.Param(r, "location_id")
	parsed, err := uuid.Parse(locationID)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	err = api.inventorylocationapp.Delete(ctx, parsed)
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

	locations, err := api.inventorylocationapp.Query(ctx, qp)
	if err != nil {
		return errs.NewError(err)
	}

	return locations
}

func (api *api) queryByID(ctx context.Context, r *http.Request) web.Encoder {
	locationID := web.Param(r, "location_id")
	parsed, err := uuid.Parse(locationID)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	location, err := api.inventorylocationapp.QueryByID(ctx, parsed)
	if err != nil {
		return errs.NewError(err)
	}

	return location
}
