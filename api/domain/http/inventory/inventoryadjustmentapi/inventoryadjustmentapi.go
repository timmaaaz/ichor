package inventoryadjustmentapi

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/domain/inventory/inventoryadjustmentapp"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/foundation/web"
)

type api struct {
	inventoryadjustmentapp *inventoryadjustmentapp.App
}

func newAPI(inventoryadjustmentapp *inventoryadjustmentapp.App) *api {
	return &api{
		inventoryadjustmentapp: inventoryadjustmentapp,
	}
}

func (api *api) create(ctx context.Context, r *http.Request) web.Encoder {
	var app inventoryadjustmentapp.NewInventoryAdjustment
	if err := web.Decode(r, &app); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	adjustment, err := api.inventoryadjustmentapp.Create(ctx, app)
	if err != nil {
		return errs.NewError(err)
	}

	return adjustment
}

func (api *api) update(ctx context.Context, r *http.Request) web.Encoder {
	var app inventoryadjustmentapp.UpdateInventoryAdjustment
	if err := web.Decode(r, &app); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	adjustmentID := web.Param(r, "adjustment_id")
	parsed, err := uuid.Parse(adjustmentID)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	adjustment, err := api.inventoryadjustmentapp.Update(ctx, parsed, app)
	if err != nil {
		return errs.NewError(err)
	}

	return adjustment
}

func (api *api) delete(ctx context.Context, r *http.Request) web.Encoder {
	adjustmentID := web.Param(r, "adjustment_id")
	parsed, err := uuid.Parse(adjustmentID)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	err = api.inventoryadjustmentapp.Delete(ctx, parsed)
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

	adjustments, err := api.inventoryadjustmentapp.Query(ctx, qp)
	if err != nil {
		return errs.NewError(err)
	}

	return adjustments
}

func (api *api) approve(ctx context.Context, r *http.Request) web.Encoder {
	var app inventoryadjustmentapp.ApproveRequest
	if err := web.Decode(r, &app); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	adjustmentID := web.Param(r, "adjustment_id")
	parsed, err := uuid.Parse(adjustmentID)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	approvedBy, err := uuid.Parse(app.ApprovedBy)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	adjustment, err := api.inventoryadjustmentapp.Approve(ctx, parsed, approvedBy)
	if err != nil {
		return errs.NewError(err)
	}

	return adjustment
}

func (api *api) reject(ctx context.Context, r *http.Request) web.Encoder {
	adjustmentID := web.Param(r, "adjustment_id")
	parsed, err := uuid.Parse(adjustmentID)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	adjustment, err := api.inventoryadjustmentapp.Reject(ctx, parsed)
	if err != nil {
		return errs.NewError(err)
	}

	return adjustment
}

func (api *api) queryByID(ctx context.Context, r *http.Request) web.Encoder {
	adjustmentID := web.Param(r, "adjustment_id")
	parsed, err := uuid.Parse(adjustmentID)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	adjustment, err := api.inventoryadjustmentapp.QueryByID(ctx, parsed)
	if err != nil {
		return errs.NewError(err)
	}

	return adjustment
}
