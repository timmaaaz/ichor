package productcategoryapi

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/domain/inventory/core/productcategoryapp"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/foundation/web"
)

type api struct {
	productcategoryapp *productcategoryapp.App
}

func newAPI(productcategoryapp *productcategoryapp.App) *api {
	return &api{
		productcategoryapp: productcategoryapp,
	}
}

func (api *api) create(ctx context.Context, r *http.Request) web.Encoder {
	var app productcategoryapp.NewProductCategory
	if err := web.Decode(r, &app); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	pc, err := api.productcategoryapp.Create(ctx, app)
	if err != nil {
		return errs.NewError(err)
	}

	return pc
}

func (api *api) update(ctx context.Context, r *http.Request) web.Encoder {
	var app productcategoryapp.UpdateProductCategory
	if err := web.Decode(r, &app); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	pcID := web.Param(r, "product_category_id")
	parsed, err := uuid.Parse(pcID)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	pc, err := api.productcategoryapp.Update(ctx, app, parsed)
	if err != nil {
		return errs.NewError(err)
	}

	return pc
}

func (api *api) delete(ctx context.Context, r *http.Request) web.Encoder {
	pcID := web.Param(r, "product_category_id")

	parsed, err := uuid.Parse(pcID)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	err = api.productcategoryapp.Delete(ctx, parsed)
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

	pcs, err := api.productcategoryapp.Query(ctx, qp)
	if err != nil {
		return errs.NewError(err)
	}

	return pcs
}

func (api *api) queryByID(ctx context.Context, r *http.Request) web.Encoder {
	pcID := web.Param(r, "product_category_id")

	parsed, err := uuid.Parse(pcID)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	pc, err := api.productcategoryapp.QueryByID(ctx, parsed)
	if err != nil {
		return errs.NewError(err)
	}

	return pc
}
