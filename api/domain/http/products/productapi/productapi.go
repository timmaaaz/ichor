package productapi

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/domain/products/productapp"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/foundation/web"
)

type api struct {
	productapp *productapp.App
}

func newAPI(productapp *productapp.App) *api {
	return &api{
		productapp: productapp,
	}
}

func (api *api) create(ctx context.Context, r *http.Request) web.Encoder {
	var app productapp.NewProduct
	if err := web.Decode(r, &app); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	product, err := api.productapp.Create(ctx, app)
	if err != nil {
		return errs.NewError(err)
	}

	return product
}

func (api *api) update(ctx context.Context, r *http.Request) web.Encoder {
	var app productapp.UpdateProduct
	if err := web.Decode(r, &app); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	productID := web.Param(r, "product_id")
	parsed, err := uuid.Parse(productID)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	product, err := api.productapp.Update(ctx, app, parsed)
	if err != nil {
		return errs.NewError(err)
	}

	return product
}

func (api *api) delete(ctx context.Context, r *http.Request) web.Encoder {
	productID := web.Param(r, "product_id")
	parsed, err := uuid.Parse(productID)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	if err := api.productapp.Delete(ctx, parsed); err != nil {
		return errs.NewError(err)
	}

	return nil
}

func (api *api) query(ctx context.Context, r *http.Request) web.Encoder {
	qp, err := parseQueryParams(r)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	products, err := api.productapp.Query(ctx, qp)
	if err != nil {
		return errs.NewError(err)
	}
	return products
}

func (api *api) queryByID(ctx context.Context, r *http.Request) web.Encoder {
	productID := web.Param(r, "product_id")

	parsed, err := uuid.Parse(productID)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	product, err := api.productapp.QueryByID(ctx, parsed)
	if err != nil {
		return errs.NewError(err)
	}

	return product
}
