package purchaseorderlineitemstatusapi

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/domain/procurement/purchaseorderlineitemstatusapp"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/foundation/web"
)

type api struct {
	purchaseorderlineitemstatusapp *purchaseorderlineitemstatusapp.App
}

func newAPI(purchaseorderlineitemstatusapp *purchaseorderlineitemstatusapp.App) *api {
	return &api{
		purchaseorderlineitemstatusapp: purchaseorderlineitemstatusapp,
	}
}

func (api *api) create(ctx context.Context, r *http.Request) web.Encoder {
	var app purchaseorderlineitemstatusapp.NewPurchaseOrderLineItemStatus
	if err := web.Decode(r, &app); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	polis, err := api.purchaseorderlineitemstatusapp.Create(ctx, app)
	if err != nil {
		return errs.NewError(err)
	}

	return polis
}

func (api *api) update(ctx context.Context, r *http.Request) web.Encoder {
	var app purchaseorderlineitemstatusapp.UpdatePurchaseOrderLineItemStatus
	if err := web.Decode(r, &app); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	polisID := web.Param(r, "purchase_order_line_item_status_id")
	parsed, err := uuid.Parse(polisID)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	polis, err := api.purchaseorderlineitemstatusapp.Update(ctx, app, parsed)
	if err != nil {
		return errs.NewError(err)
	}

	return polis
}

func (api *api) delete(ctx context.Context, r *http.Request) web.Encoder {
	polisID := web.Param(r, "purchase_order_line_item_status_id")

	parsed, err := uuid.Parse(polisID)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	err = api.purchaseorderlineitemstatusapp.Delete(ctx, parsed)
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

	statuses, err := api.purchaseorderlineitemstatusapp.Query(ctx, qp)
	if err != nil {
		return errs.NewError(err)
	}

	return statuses
}

func (api *api) queryByID(ctx context.Context, r *http.Request) web.Encoder {
	polisID := web.Param(r, "purchase_order_line_item_status_id")

	parsed, err := uuid.Parse(polisID)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	polis, err := api.purchaseorderlineitemstatusapp.QueryByID(ctx, parsed)
	if err != nil {
		return errs.NewError(err)
	}

	return polis
}

func (api *api) queryByIDs(ctx context.Context, r *http.Request) web.Encoder {
	var app purchaseorderlineitemstatusapp.QueryByIDsRequest
	if err := web.Decode(r, &app); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	statuses, err := api.purchaseorderlineitemstatusapp.QueryByIDs(ctx, app.IDs)
	if err != nil {
		return errs.NewError(err)
	}

	return statuses
}

func (api *api) queryAll(ctx context.Context, r *http.Request) web.Encoder {
	statuses, err := api.purchaseorderlineitemstatusapp.QueryAll(ctx)
	if err != nil {
		return errs.NewError(err)
	}

	return statuses
}
