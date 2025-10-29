package purchaseorderlineitemapi

import (
	"context"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/domain/procurement/purchaseorderlineitemapp"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/foundation/web"
)

type api struct {
	purchaseorderlineitemapp *purchaseorderlineitemapp.App
}

func newAPI(purchaseorderlineitemapp *purchaseorderlineitemapp.App) *api {
	return &api{
		purchaseorderlineitemapp: purchaseorderlineitemapp,
	}
}

func (api *api) create(ctx context.Context, r *http.Request) web.Encoder {
	var app purchaseorderlineitemapp.NewPurchaseOrderLineItem
	if err := web.Decode(r, &app); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	poli, err := api.purchaseorderlineitemapp.Create(ctx, app)
	if err != nil {
		return errs.NewError(err)
	}

	return poli
}

func (api *api) update(ctx context.Context, r *http.Request) web.Encoder {
	var app purchaseorderlineitemapp.UpdatePurchaseOrderLineItem
	if err := web.Decode(r, &app); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	poliID := web.Param(r, "purchase_order_line_item_id")
	parsed, err := uuid.Parse(poliID)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	poli, err := api.purchaseorderlineitemapp.Update(ctx, app, parsed)
	if err != nil {
		return errs.NewError(err)
	}

	return poli
}

func (api *api) delete(ctx context.Context, r *http.Request) web.Encoder {
	poliID := web.Param(r, "purchase_order_line_item_id")

	parsed, err := uuid.Parse(poliID)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	err = api.purchaseorderlineitemapp.Delete(ctx, parsed)
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

	items, err := api.purchaseorderlineitemapp.Query(ctx, qp)
	if err != nil {
		return errs.NewError(err)
	}

	return items
}

func (api *api) queryByID(ctx context.Context, r *http.Request) web.Encoder {
	poliID := web.Param(r, "purchase_order_line_item_id")

	parsed, err := uuid.Parse(poliID)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	poli, err := api.purchaseorderlineitemapp.QueryByID(ctx, parsed)
	if err != nil {
		return errs.NewError(err)
	}

	return poli
}

func (api *api) queryByIDs(ctx context.Context, r *http.Request) web.Encoder {
	var app purchaseorderlineitemapp.QueryByIDsRequest
	if err := web.Decode(r, &app); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	items, err := api.purchaseorderlineitemapp.QueryByIDs(ctx, app.IDs)
	if err != nil {
		return errs.NewError(err)
	}

	return items
}

func (api *api) queryByPurchaseOrderID(ctx context.Context, r *http.Request) web.Encoder {
	poID := web.Param(r, "purchase_order_id")

	parsed, err := uuid.Parse(poID)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	items, err := api.purchaseorderlineitemapp.QueryByPurchaseOrderID(ctx, parsed)
	if err != nil {
		return errs.NewError(err)
	}

	return items
}

func (api *api) receiveQuantity(ctx context.Context, r *http.Request) web.Encoder {
	var app purchaseorderlineitemapp.ReceiveQuantityRequest
	if err := web.Decode(r, &app); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	poliID := web.Param(r, "purchase_order_line_item_id")
	parsed, err := uuid.Parse(poliID)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	quantity, err := strconv.Atoi(app.Quantity)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	receivedBy, err := uuid.Parse(app.ReceivedBy)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	poli, err := api.purchaseorderlineitemapp.ReceiveQuantity(ctx, parsed, quantity, receivedBy)
	if err != nil {
		return errs.NewError(err)
	}

	return poli
}