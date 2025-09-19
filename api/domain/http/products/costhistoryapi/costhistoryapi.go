package costhistoryapi

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/domain/products/costhistoryapp"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/foundation/web"
)

type api struct {
	costhistoryapp *costhistoryapp.App
}

func newAPI(costhistoryapp *costhistoryapp.App) *api {
	return &api{
		costhistoryapp: costhistoryapp,
	}
}

func (api *api) create(ctx context.Context, r *http.Request) web.Encoder {
	var app costhistoryapp.NewCostHistory
	if err := web.Decode(r, &app); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	costHistory, err := api.costhistoryapp.Create(ctx, app)
	if err != nil {
		return errs.NewError(err)
	}

	return costHistory
}

func (api *api) update(ctx context.Context, r *http.Request) web.Encoder {
	var app costhistoryapp.UpdateCostHistory
	if err := web.Decode(r, &app); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	chID := web.Param(r, "cost_history_id")
	parsed, err := uuid.Parse(chID)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	productCost, err := api.costhistoryapp.Update(ctx, app, parsed)
	if err != nil {
		return errs.NewError(err)
	}

	return productCost
}

func (api *api) delete(ctx context.Context, r *http.Request) web.Encoder {
	chID := web.Param(r, "cost_history_id")

	parsed, err := uuid.Parse(chID)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	err = api.costhistoryapp.Delete(ctx, parsed)
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

	costs, err := api.costhistoryapp.Query(ctx, qp)
	if err != nil {
		return errs.NewError(err)
	}

	return costs
}

func (api *api) queryByID(ctx context.Context, r *http.Request) web.Encoder {
	chID := web.Param(r, "cost_history_id")

	parsed, err := uuid.Parse(chID)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	ch, err := api.costhistoryapp.QueryByID(ctx, parsed)
	if err != nil {
		return errs.NewError(err)
	}

	return ch
}
