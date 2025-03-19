package supplierapi

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/domain/supplier/supplierapp"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/foundation/web"
)

type api struct {
	supplierapp *supplierapp.App
}

func newAPI(supplierapp *supplierapp.App) *api {
	return &api{
		supplierapp: supplierapp,
	}
}

func (api *api) create(ctx context.Context, r *http.Request) web.Encoder {
	var app supplierapp.NewSupplier
	if err := web.Decode(r, &app); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	supplier, err := api.supplierapp.Create(ctx, app)
	if err != nil {
		return errs.NewError(err)
	}

	return supplier
}

func (api *api) update(ctx context.Context, r *http.Request) web.Encoder {
	var app supplierapp.UpdateSupplier
	if err := web.Decode(r, &app); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	sID := web.Param(r, "supplier_id")
	parsed, err := uuid.Parse(sID)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	supplier, err := api.supplierapp.Update(ctx, app, parsed)
	if err != nil {
		return errs.NewError(err)
	}

	return supplier
}

func (api *api) delete(ctx context.Context, r *http.Request) web.Encoder {
	supplierID := web.Param(r, "supplier_id")

	parsed, err := uuid.Parse(supplierID)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	err = api.supplierapp.Delete(ctx, parsed)
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

	suppliers, err := api.supplierapp.Query(ctx, qp)
	if err != nil {
		return errs.NewError(err)
	}

	return suppliers
}

func (api *api) queryByID(ctx context.Context, r *http.Request) web.Encoder {
	supplierID := web.Param(r, "supplier_id")

	parsed, err := uuid.Parse(supplierID)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	supplier, err := api.supplierapp.QueryByID(ctx, parsed)
	if err != nil {
		return errs.NewError(err)
	}

	return supplier
}
