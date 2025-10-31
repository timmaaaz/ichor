package purchaseorderstatusapi

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/domain/procurement/purchaseorderstatusapp"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/foundation/web"
)

type api struct {
	purchaseorderstatusapp *purchaseorderstatusapp.App
}

func newAPI(purchaseorderstatusapp *purchaseorderstatusapp.App) *api {
	return &api{
		purchaseorderstatusapp: purchaseorderstatusapp,
	}
}

func (api *api) create(ctx context.Context, r *http.Request) web.Encoder {
	var app purchaseorderstatusapp.NewPurchaseOrderStatus
	if err := web.Decode(r, &app); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	pos, err := api.purchaseorderstatusapp.Create(ctx, app)
	if err != nil {
		return errs.NewError(err)
	}

	return pos
}

func (api *api) update(ctx context.Context, r *http.Request) web.Encoder {
	var app purchaseorderstatusapp.UpdatePurchaseOrderStatus
	if err := web.Decode(r, &app); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	posID := web.Param(r, "purchase_order_status_id")
	parsed, err := uuid.Parse(posID)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	pos, err := api.purchaseorderstatusapp.Update(ctx, app, parsed)
	if err != nil {
		return errs.NewError(err)
	}

	return pos
}

func (api *api) delete(ctx context.Context, r *http.Request) web.Encoder {
	posID := web.Param(r, "purchase_order_status_id")

	parsed, err := uuid.Parse(posID)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	err = api.purchaseorderstatusapp.Delete(ctx, parsed)
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

	statuses, err := api.purchaseorderstatusapp.Query(ctx, qp)
	if err != nil {
		return errs.NewError(err)
	}

	return statuses
}

func (api *api) queryByID(ctx context.Context, r *http.Request) web.Encoder {
	posID := web.Param(r, "purchase_order_status_id")

	parsed, err := uuid.Parse(posID)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	pos, err := api.purchaseorderstatusapp.QueryByID(ctx, parsed)
	if err != nil {
		return errs.NewError(err)
	}

	return pos
}

func (api *api) queryByIDs(ctx context.Context, r *http.Request) web.Encoder {
	var app purchaseorderstatusapp.QueryByIDsRequest
	if err := web.Decode(r, &app); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	statuses, err := api.purchaseorderstatusapp.QueryByIDs(ctx, app.IDs)
	if err != nil {
		return errs.NewError(err)
	}

	return statuses
}

func (api *api) queryAll(ctx context.Context, r *http.Request) web.Encoder {
	statuses, err := api.purchaseorderstatusapp.QueryAll(ctx)
	if err != nil {
		return errs.NewError(err)
	}

	return statuses
}
