package warehouseapi

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/domain/warehouse/warehouseapp"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/foundation/web"
)

type api struct {
	warehouseapp *warehouseapp.App
}

func newAPI(warehouseapp *warehouseapp.App) *api {
	return &api{
		warehouseapp: warehouseapp,
	}
}

func (api *api) create(ctx context.Context, r *http.Request) web.Encoder {
	var app warehouseapp.NewWarehouse
	if err := web.Decode(r, &app); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	warehouse, err := api.warehouseapp.Create(ctx, app)
	if err != nil {
		return errs.NewError(err)
	}

	return warehouse
}

func (api *api) update(ctx context.Context, r *http.Request) web.Encoder {
	var app warehouseapp.UpdateWarehouse
	if err := web.Decode(r, &app); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	warehouseID := web.Param(r, "warehouse_id")
	parsed, err := uuid.Parse(warehouseID)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	warehouse, err := api.warehouseapp.Update(ctx, app, parsed)
	if err != nil {
		return errs.NewError(err)
	}

	return warehouse
}

func (api *api) delete(ctx context.Context, r *http.Request) web.Encoder {
	warehouseID := web.Param(r, "warehouse_id")
	parsed, err := uuid.Parse(warehouseID)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	err = api.warehouseapp.Delete(ctx, parsed)
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

	result, err := api.warehouseapp.Query(ctx, qp)
	if err != nil {
		return errs.NewError(err)
	}

	return result
}

func (api *api) queryByID(ctx context.Context, r *http.Request) web.Encoder {
	warehouseID := web.Param(r, "warehouse_id")
	parsed, err := uuid.Parse(warehouseID)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	warehouse, err := api.warehouseapp.QueryByID(ctx, parsed)
	if err != nil {
		return errs.NewError(err)
	}

	return warehouse
}
