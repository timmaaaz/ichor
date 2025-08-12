package customersapi

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/domain/core/customersapp"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/foundation/web"
)

type api struct {
	customersapp *customersapp.App
}

func newAPI(customersapp *customersapp.App) *api {
	return &api{
		customersapp: customersapp,
	}
}

func (api *api) create(ctx context.Context, r *http.Request) web.Encoder {
	var app customersapp.NewCustomers
	if err := web.Decode(r, &app); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	customer, err := api.customersapp.Create(ctx, app)
	if err != nil {
		return errs.NewError(err)
	}

	return customer
}

func (api *api) update(ctx context.Context, r *http.Request) web.Encoder {
	var app customersapp.UpdateCustomers
	if err := web.Decode(r, &app); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	ciID := web.Param(r, "customers_id")
	parsed, err := uuid.Parse(ciID)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	customer, err := api.customersapp.Update(ctx, app, parsed)
	if err != nil {
		return errs.NewError(err)
	}

	return customer
}

func (api *api) delete(ctx context.Context, r *http.Request) web.Encoder {
	ciID := web.Param(r, "customers_id")

	parsed, err := uuid.Parse(ciID)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	err = api.customersapp.Delete(ctx, parsed)
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

	customers, err := api.customersapp.Query(ctx, qp)
	if err != nil {
		return errs.NewError(err)
	}

	return customers
}

func (api *api) queryByID(ctx context.Context, r *http.Request) web.Encoder {
	ciID := web.Param(r, "customers_id")

	parsed, err := uuid.Parse(ciID)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	ci, err := api.customersapp.QueryByID(ctx, parsed)
	if err != nil {
		return errs.NewError(err)
	}

	return ci
}
