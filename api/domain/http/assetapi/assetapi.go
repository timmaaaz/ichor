package assetapi

import (
	"context"
	"net/http"

	"github.com/google/uuid"

	"github.com/timmaaaz/ichor/app/domain/assetapp"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/foundation/web"
)

type api struct {
	assetapp *assetapp.App
}

func newAPI(assetapp *assetapp.App) *api {
	return &api{
		assetapp: assetapp,
	}
}

func (api *api) create(ctx context.Context, r *http.Request) web.Encoder {
	var app assetapp.NewAsset
	if err := web.Decode(r, &app); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	asset, err := api.assetapp.Create(ctx, app)
	if err != nil {
		return errs.NewError(err)
	}

	return asset
}

func (api *api) update(ctx context.Context, r *http.Request) web.Encoder {
	var app assetapp.UpdateAsset
	if err := web.Decode(r, &app); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	assetID := web.Param(r, "asset_id")
	parsed, err := uuid.Parse(assetID)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	asset, err := api.assetapp.Update(ctx, app, parsed)
	if err != nil {
		return errs.NewError(err)
	}

	return asset
}

func (api *api) delete(ctx context.Context, r *http.Request) web.Encoder {
	assetID := web.Param(r, "asset_id")

	parsed, err := uuid.Parse(assetID)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	err = api.assetapp.Delete(ctx, parsed)
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

	assets, err := api.assetapp.Query(ctx, qp)
	if err != nil {
		return errs.NewError(err)
	}

	return assets
}

func (api *api) queryByID(ctx context.Context, r *http.Request) web.Encoder {
	assetID := web.Param(r, "asset_id")

	parsed, err := uuid.Parse(assetID)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	asset, err := api.assetapp.QueryByID(ctx, parsed)
	if err != nil {
		return errs.NewError(err)
	}

	return asset
}
