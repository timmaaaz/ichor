package orderlineitemsapi

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/domain/sales/orderlineitemsapp"
	"github.com/timmaaaz/ichor/app/domain/sales/pickingapp"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/foundation/web"
)

type api struct {
	orderlineitemsapp *orderlineitemsapp.App
	pickingApp        *pickingapp.App
}

func newAPI(orderlineitemsapp *orderlineitemsapp.App, pickingApp *pickingapp.App) *api {
	return &api{
		orderlineitemsapp: orderlineitemsapp,
		pickingApp:        pickingApp,
	}
}

func (api *api) create(ctx context.Context, r *http.Request) web.Encoder {
	var app orderlineitemsapp.NewOrderLineItem
	if err := web.Decode(r, &app); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	status, err := api.orderlineitemsapp.Create(ctx, app)
	if err != nil {
		return errs.NewError(err)
	}

	return status
}

func (api *api) update(ctx context.Context, r *http.Request) web.Encoder {
	var app orderlineitemsapp.UpdateOrderLineItem
	if err := web.Decode(r, &app); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	statusID := web.Param(r, "order_line_items_id")
	parsed, err := uuid.Parse(statusID)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	status, err := api.orderlineitemsapp.Update(ctx, app, parsed)
	if err != nil {
		return errs.NewError(err)
	}

	return status
}

func (api *api) delete(ctx context.Context, r *http.Request) web.Encoder {
	statusID := web.Param(r, "order_line_items_id")

	parsed, err := uuid.Parse(statusID)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	err = api.orderlineitemsapp.Delete(ctx, parsed)
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

	pcs, err := api.orderlineitemsapp.Query(ctx, qp)
	if err != nil {
		return errs.NewError(err)
	}

	return pcs
}

func (api *api) queryByID(ctx context.Context, r *http.Request) web.Encoder {
	statusID := web.Param(r, "order_line_items_id")

	parsed, err := uuid.Parse(statusID)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	status, err := api.orderlineitemsapp.QueryByID(ctx, parsed)
	if err != nil {
		return errs.NewError(err)
	}

	return status
}

func (api *api) pickQuantity(ctx context.Context, r *http.Request) web.Encoder {
	var req pickingapp.PickQuantityRequest
	if err := web.Decode(r, &req); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	lineItemID, err := uuid.Parse(web.Param(r, "order_line_items_id"))
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	result, err := api.pickingApp.PickQuantity(ctx, lineItemID, req)
	if err != nil {
		return errs.NewError(err)
	}

	return result
}

func (api *api) shortPick(ctx context.Context, r *http.Request) web.Encoder {
	var req pickingapp.ShortPickRequest
	if err := web.Decode(r, &req); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	lineItemID, err := uuid.Parse(web.Param(r, "order_line_items_id"))
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	result, err := api.pickingApp.ShortPick(ctx, lineItemID, req)
	if err != nil {
		return errs.NewError(err)
	}

	return result
}
