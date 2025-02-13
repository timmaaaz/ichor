package userassetapi

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/domain/assets/userassetapp"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/foundation/web"
)

type api struct {
	userassetapp *userassetapp.App
}

func newAPI(userassetapp *userassetapp.App) *api {
	return &api{
		userassetapp: userassetapp,
	}
}

func (api *api) create(ctx context.Context, r *http.Request) web.Encoder {
	var app userassetapp.NewUserAsset
	if err := web.Decode(r, &app); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	asset, err := api.userassetapp.Create(ctx, app)
	if err != nil {
		return errs.NewError(err)
	}

	return asset
}

func (api *api) update(ctx context.Context, r *http.Request) web.Encoder {
	var app userassetapp.UpdateUserAsset
	if err := web.Decode(r, &app); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	assetID := web.Param(r, "user_asset_id")
	parsed, err := uuid.Parse(assetID)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	asset, err := api.userassetapp.Update(ctx, app, parsed)
	if err != nil {
		return errs.NewError(err)
	}

	return asset
}

func (api *api) delete(ctx context.Context, r *http.Request) web.Encoder {
	assetID := web.Param(r, "user_asset_id")

	parsed, err := uuid.Parse(assetID)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	err = api.userassetapp.Delete(ctx, parsed)
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

	assets, err := api.userassetapp.Query(ctx, qp)
	if err != nil {
		return errs.NewError(err)
	}

	return assets
}

func (api *api) queryByID(ctx context.Context, r *http.Request) web.Encoder {
	assetID := web.Param(r, "user_asset_id")

	parsed, err := uuid.Parse(assetID)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	asset, err := api.userassetapp.QueryByID(ctx, parsed)
	if err != nil {
		return errs.NewError(err)
	}

	return asset
}
