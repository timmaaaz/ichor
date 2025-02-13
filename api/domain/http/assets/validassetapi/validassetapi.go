package validassetapi

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/domain/assets/validassetapp"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/foundation/web"
)

type api struct {
	validassetapp *validassetapp.App
}

func newAPI(validassetapp *validassetapp.App) *api {
	return &api{
		validassetapp: validassetapp,
	}
}

func (api *api) create(ctx context.Context, r *http.Request) web.Encoder {
	var app validassetapp.NewValidAsset
	if err := web.Decode(r, &app); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	asset, err := api.validassetapp.Create(ctx, app)
	if err != nil {
		return errs.NewError(err)
	}

	return asset
}

func (api *api) update(ctx context.Context, r *http.Request) web.Encoder {
	var app validassetapp.UpdateValidAsset
	if err := web.Decode(r, &app); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	assetID := web.Param(r, "valid_asset_id")
	parsed, err := uuid.Parse(assetID)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	asset, err := api.validassetapp.Update(ctx, app, parsed)
	if err != nil {
		return errs.NewError(err)
	}

	return asset
}

func (api *api) delete(ctx context.Context, r *http.Request) web.Encoder {
	assetID := web.Param(r, "valid_asset_id")

	parsed, err := uuid.Parse(assetID)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	err = api.validassetapp.Delete(ctx, parsed)
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

	assets, err := api.validassetapp.Query(ctx, qp)
	if err != nil {
		return errs.NewError(err)
	}

	return assets
}

func (api *api) queryByID(ctx context.Context, r *http.Request) web.Encoder {
	assetID := web.Param(r, "valid_asset_id")

	parsed, err := uuid.Parse(assetID)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	asset, err := api.validassetapp.QueryByID(ctx, parsed)
	if err != nil {
		return errs.NewError(err)
	}

	return asset
}
