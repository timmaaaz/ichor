package supplierproductapi

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/domain/procurement/supplierproductapp"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/foundation/web"
)

type api struct {
	supplierproductapp *supplierproductapp.App
}

func newAPI(supplierproductapp *supplierproductapp.App) *api {
	return &api{
		supplierproductapp: supplierproductapp,
	}
}

func (api *api) create(ctx context.Context, r *http.Request) web.Encoder {
	var app supplierproductapp.NewSupplierProduct
	if err := web.Decode(r, &app); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	supplier, err := api.supplierproductapp.Create(ctx, app)
	if err != nil {
		return errs.NewError(err)
	}

	return supplier
}

func (api *api) update(ctx context.Context, r *http.Request) web.Encoder {
	var app supplierproductapp.UpdateSupplierProduct
	if err := web.Decode(r, &app); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	sID := web.Param(r, "supplier_product_id")
	parsed, err := uuid.Parse(sID)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	supplier, err := api.supplierproductapp.Update(ctx, app, parsed)
	if err != nil {
		return errs.NewError(err)
	}

	return supplier
}

func (api *api) delete(ctx context.Context, r *http.Request) web.Encoder {
	supplierID := web.Param(r, "supplier_product_id")

	parsed, err := uuid.Parse(supplierID)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	err = api.supplierproductapp.Delete(ctx, parsed)
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

	suppliers, err := api.supplierproductapp.Query(ctx, qp)
	if err != nil {
		return errs.NewError(err)
	}

	return suppliers
}

func (api *api) queryByID(ctx context.Context, r *http.Request) web.Encoder {
	supplierID := web.Param(r, "supplier_product_id")

	parsed, err := uuid.Parse(supplierID)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	supplier, err := api.supplierproductapp.QueryByID(ctx, parsed)
	if err != nil {
		return errs.NewError(err)
	}

	return supplier
}

func (api *api) queryByIDs(ctx context.Context, r *http.Request) web.Encoder {
	var app supplierproductapp.QueryByIDsRequest
	if err := web.Decode(r, &app); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	supplierProducts, err := api.supplierproductapp.QueryByIDs(ctx, app.IDs)
	if err != nil {
		return errs.NewError(err)
	}

	return supplierProducts
}
