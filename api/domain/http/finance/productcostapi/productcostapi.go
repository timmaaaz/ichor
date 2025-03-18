package productcostapi

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/domain/finance/productcostapp"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/foundation/web"
)

type api struct {
	productcostapp *productcostapp.App
}

func newAPI(productcostapp *productcostapp.App) *api {
	return &api{
		productcostapp: productcostapp,
	}
}

func (api *api) create(ctx context.Context, r *http.Request) web.Encoder {
	var app productcostapp.NewProductCost
	if err := web.Decode(r, &app); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	brand, err := api.productcostapp.Create(ctx, app)
	if err != nil {
		return errs.NewError(err)
	}

	return brand
}

func (api *api) update(ctx context.Context, r *http.Request) web.Encoder {
	var app productcostapp.UpdateProductCost
	if err := web.Decode(r, &app); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	pcID := web.Param(r, "product_cost_id")
	parsed, err := uuid.Parse(pcID)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	productCost, err := api.productcostapp.Update(ctx, app, parsed)
	if err != nil {
		return errs.NewError(err)
	}

	return productCost
}

func (api *api) delete(ctx context.Context, r *http.Request) web.Encoder {
	pcID := web.Param(r, "product_cost_id")

	parsed, err := uuid.Parse(pcID)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	err = api.productcostapp.Delete(ctx, parsed)
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

	costs, err := api.productcostapp.Query(ctx, qp)
	if err != nil {
		return errs.NewError(err)
	}

	return costs
}

func (api *api) queryByID(ctx context.Context, r *http.Request) web.Encoder {
	productCostID := web.Param(r, "product_cost_id")

	parsed, err := uuid.Parse(productCostID)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	productCost, err := api.productcostapp.QueryByID(ctx, parsed)
	if err != nil {
		return errs.NewError(err)
	}

	return productCost
}
