package fulfillmentstatusapi

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/domain/assets/fulfillmentstatusapp"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/foundation/web"
)

type api struct {
	fulfillmentstatusapp *fulfillmentstatusapp.App
}

func newAPI(fulfillmentstatusapp *fulfillmentstatusapp.App) *api {
	return &api{
		fulfillmentstatusapp: fulfillmentstatusapp,
	}
}

func (api *api) create(ctx context.Context, r *http.Request) web.Encoder {
	var app fulfillmentstatusapp.NewFulfillmentStatus
	if err := web.Decode(r, &app); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	status, err := api.fulfillmentstatusapp.Create(ctx, app)
	if err != nil {
		return errs.NewError(err)
	}

	return status
}

func (api *api) update(ctx context.Context, r *http.Request) web.Encoder {
	var app fulfillmentstatusapp.UpdateFulfillmentStatus
	if err := web.Decode(r, &app); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	statusID := web.Param(r, "fulfillment_status_id")
	parsed, err := uuid.Parse(statusID)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	status, err := api.fulfillmentstatusapp.Update(ctx, app, parsed)
	if err != nil {
		return errs.NewError(err)
	}

	return status
}

func (api *api) delete(ctx context.Context, r *http.Request) web.Encoder {
	fulfillmentStatusID := web.Param(r, "fulfillment_status_id")

	parsed, err := uuid.Parse(fulfillmentStatusID)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	err = api.fulfillmentstatusapp.Delete(ctx, parsed)
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

	statuses, err := api.fulfillmentstatusapp.Query(ctx, qp)
	if err != nil {
		return errs.NewError(err)
	}

	return statuses
}

func (api *api) queryByID(ctx context.Context, r *http.Request) web.Encoder {
	fulfillmentStatusID := web.Param(r, "fulfillment_status_id")

	parsed, err := uuid.Parse(fulfillmentStatusID)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	fulfillmentStatus, err := api.fulfillmentstatusapp.QueryByID(ctx, parsed)
	if err != nil {
		return errs.NewError(err)
	}

	return fulfillmentStatus
}
