package lineitemfulfillmentstatusapi

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/domain/sales/lineitemfulfillmentstatusapp"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/foundation/web"
)

type api struct {
	lineitemfulfillmentstatusapp *lineitemfulfillmentstatusapp.App
}

func newAPI(lineitemfulfillmentstatusapp *lineitemfulfillmentstatusapp.App) *api {
	return &api{
		lineitemfulfillmentstatusapp: lineitemfulfillmentstatusapp,
	}
}

func (api *api) create(ctx context.Context, r *http.Request) web.Encoder {
	var app lineitemfulfillmentstatusapp.NewLineItemFulfillmentStatus
	if err := web.Decode(r, &app); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	status, err := api.lineitemfulfillmentstatusapp.Create(ctx, app)
	if err != nil {
		return errs.NewError(err)
	}

	return status
}

func (api *api) update(ctx context.Context, r *http.Request) web.Encoder {
	var app lineitemfulfillmentstatusapp.UpdateLineItemFulfillmentStatus
	if err := web.Decode(r, &app); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	statusID := web.Param(r, "line_item_fulfillment_status_id")
	parsed, err := uuid.Parse(statusID)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	status, err := api.lineitemfulfillmentstatusapp.Update(ctx, app, parsed)
	if err != nil {
		return errs.NewError(err)
	}

	return status
}

func (api *api) delete(ctx context.Context, r *http.Request) web.Encoder {
	statusID := web.Param(r, "line_item_fulfillment_status_id")

	parsed, err := uuid.Parse(statusID)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	err = api.lineitemfulfillmentstatusapp.Delete(ctx, parsed)
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

	pcs, err := api.lineitemfulfillmentstatusapp.Query(ctx, qp)
	if err != nil {
		return errs.NewError(err)
	}

	return pcs
}

func (api *api) queryByID(ctx context.Context, r *http.Request) web.Encoder {
	statusID := web.Param(r, "line_item_fulfillment_status_id")

	parsed, err := uuid.Parse(statusID)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	status, err := api.lineitemfulfillmentstatusapp.QueryByID(ctx, parsed)
	if err != nil {
		return errs.NewError(err)
	}

	return status
}
