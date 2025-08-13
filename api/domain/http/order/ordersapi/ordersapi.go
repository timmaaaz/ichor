package ordersapi

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/domain/order/ordersapp"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/foundation/web"
)

type api struct {
	ordersapp *ordersapp.App
}

func newAPI(ordersapp *ordersapp.App) *api {
	return &api{
		ordersapp: ordersapp,
	}
}

func (api *api) create(ctx context.Context, r *http.Request) web.Encoder {
	var app ordersapp.NewOrder
	if err := web.Decode(r, &app); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	status, err := api.ordersapp.Create(ctx, app)
	if err != nil {
		return errs.NewError(err)
	}

	return status
}

func (api *api) update(ctx context.Context, r *http.Request) web.Encoder {
	var app ordersapp.UpdateOrder
	if err := web.Decode(r, &app); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	statusID := web.Param(r, "orders_id")
	parsed, err := uuid.Parse(statusID)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	status, err := api.ordersapp.Update(ctx, app, parsed)
	if err != nil {
		return errs.NewError(err)
	}

	return status
}

func (api *api) delete(ctx context.Context, r *http.Request) web.Encoder {
	statusID := web.Param(r, "orders_id")

	parsed, err := uuid.Parse(statusID)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	err = api.ordersapp.Delete(ctx, parsed)
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

	pcs, err := api.ordersapp.Query(ctx, qp)
	if err != nil {
		return errs.NewError(err)
	}

	return pcs
}

func (api *api) queryByID(ctx context.Context, r *http.Request) web.Encoder {
	statusID := web.Param(r, "orders_id")

	parsed, err := uuid.Parse(statusID)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	status, err := api.ordersapp.QueryByID(ctx, parsed)
	if err != nil {
		return errs.NewError(err)
	}

	return status
}
