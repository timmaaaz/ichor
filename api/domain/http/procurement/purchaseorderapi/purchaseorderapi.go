package purchaseorderapi

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/domain/procurement/purchaseorderapp"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/foundation/web"
)

type api struct {
	purchaseorderapp *purchaseorderapp.App
}

func newAPI(purchaseorderapp *purchaseorderapp.App) *api {
	return &api{
		purchaseorderapp: purchaseorderapp,
	}
}

func (api *api) create(ctx context.Context, r *http.Request) web.Encoder {
	var app purchaseorderapp.NewPurchaseOrder
	if err := web.Decode(r, &app); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	po, err := api.purchaseorderapp.Create(ctx, app)
	if err != nil {
		return errs.NewError(err)
	}

	return po
}

func (api *api) update(ctx context.Context, r *http.Request) web.Encoder {
	var app purchaseorderapp.UpdatePurchaseOrder
	if err := web.Decode(r, &app); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	poID := web.Param(r, "purchase_order_id")
	parsed, err := uuid.Parse(poID)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	po, err := api.purchaseorderapp.Update(ctx, app, parsed)
	if err != nil {
		return errs.NewError(err)
	}

	return po
}

func (api *api) delete(ctx context.Context, r *http.Request) web.Encoder {
	poID := web.Param(r, "purchase_order_id")

	parsed, err := uuid.Parse(poID)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	err = api.purchaseorderapp.Delete(ctx, parsed)
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

	orders, err := api.purchaseorderapp.Query(ctx, qp)
	if err != nil {
		return errs.NewError(err)
	}

	return orders
}

func (api *api) queryByID(ctx context.Context, r *http.Request) web.Encoder {
	poID := web.Param(r, "purchase_order_id")

	parsed, err := uuid.Parse(poID)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	po, err := api.purchaseorderapp.QueryByID(ctx, parsed)
	if err != nil {
		return errs.NewError(err)
	}

	return po
}

func (api *api) queryByIDs(ctx context.Context, r *http.Request) web.Encoder {
	var app purchaseorderapp.QueryByIDsRequest
	if err := web.Decode(r, &app); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	orders, err := api.purchaseorderapp.QueryByIDs(ctx, app.IDs)
	if err != nil {
		return errs.NewError(err)
	}

	return orders
}

func (api *api) approve(ctx context.Context, r *http.Request) web.Encoder {
	var app purchaseorderapp.ApproveRequest
	if err := web.Decode(r, &app); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	poID := web.Param(r, "purchase_order_id")
	parsed, err := uuid.Parse(poID)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	approvedBy, err := uuid.Parse(app.ApprovedBy)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	po, err := api.purchaseorderapp.Approve(ctx, parsed, approvedBy)
	if err != nil {
		return errs.NewError(err)
	}

	return po
}
