package productuomapi

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/domain/products/productuomapp"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/foundation/web"
)

type api struct {
	productuomapp *productuomapp.App
}

func newAPI(productuomapp *productuomapp.App) *api {
	return &api{
		productuomapp: productuomapp,
	}
}

func (api *api) create(ctx context.Context, r *http.Request) web.Encoder {
	var app productuomapp.NewProductUOM
	if err := web.Decode(r, &app); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	uom, err := api.productuomapp.Create(ctx, app)
	if err != nil {
		return errs.NewError(err)
	}

	return uom
}

func (api *api) update(ctx context.Context, r *http.Request) web.Encoder {
	var app productuomapp.UpdateProductUOM
	if err := web.Decode(r, &app); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	uomID := web.Param(r, "uom_id")
	parsed, err := uuid.Parse(uomID)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	uom, err := api.productuomapp.Update(ctx, app, parsed)
	if err != nil {
		return errs.NewError(err)
	}

	return uom
}

func (api *api) delete(ctx context.Context, r *http.Request) web.Encoder {
	uomID := web.Param(r, "uom_id")
	parsed, err := uuid.Parse(uomID)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	if err := api.productuomapp.Delete(ctx, parsed); err != nil {
		return errs.NewError(err)
	}

	return nil
}

func (api *api) query(ctx context.Context, r *http.Request) web.Encoder {
	qp, err := parseQueryParams(r)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	uoms, err := api.productuomapp.Query(ctx, qp)
	if err != nil {
		return errs.NewError(err)
	}
	return uoms
}

func (api *api) queryByID(ctx context.Context, r *http.Request) web.Encoder {
	uomID := web.Param(r, "uom_id")

	parsed, err := uuid.Parse(uomID)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	uom, err := api.productuomapp.QueryByID(ctx, parsed)
	if err != nil {
		return errs.NewError(err)
	}

	return uom
}
