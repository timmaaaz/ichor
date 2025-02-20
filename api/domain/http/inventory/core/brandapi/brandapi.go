package brandapi

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/domain/inventory/core/brandapp"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/foundation/web"
)

type api struct {
	brandapp *brandapp.App
}

func newAPI(brandapp *brandapp.App) *api {
	return &api{
		brandapp: brandapp,
	}
}

func (api *api) create(ctx context.Context, r *http.Request) web.Encoder {
	var app brandapp.NewBrand
	if err := web.Decode(r, &app); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	brand, err := api.brandapp.Create(ctx, app)
	if err != nil {
		return errs.NewError(err)
	}

	return brand
}

func (api *api) update(ctx context.Context, r *http.Request) web.Encoder {
	var app brandapp.UpdateBrand
	if err := web.Decode(r, &app); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	brandID := web.Param(r, "brand_id")
	parsed, err := uuid.Parse(brandID)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	brand, err := api.brandapp.Update(ctx, app, parsed)
	if err != nil {
		return errs.NewError(err)
	}

	return brand
}

func (api *api) delete(ctx context.Context, r *http.Request) web.Encoder {
	brandID := web.Param(r, "brand_id")

	parsed, err := uuid.Parse(brandID)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	err = api.brandapp.Delete(ctx, parsed)
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

	brands, err := api.brandapp.Query(ctx, qp)
	if err != nil {
		return errs.NewError(err)
	}

	return brands
}

func (api *api) queryByID(ctx context.Context, r *http.Request) web.Encoder {
	brandID := web.Param(r, "brand_id")

	parsed, err := uuid.Parse(brandID)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	brand, err := api.brandapp.QueryByID(ctx, parsed)
	if err != nil {
		return errs.NewError(err)
	}

	return brand
}
