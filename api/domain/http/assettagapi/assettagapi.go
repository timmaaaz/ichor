package assettagapi

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/domain/assettagapp"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/foundation/web"
)

type api struct {
	assettagapp *assettagapp.App
}

func newAPI(assettagapp *assettagapp.App) *api {
	return &api{
		assettagapp: assettagapp,
	}
}

func (api *api) create(ctx context.Context, r *http.Request) web.Encoder {
	var app assettagapp.NewAssetTag
	if err := web.Decode(r, &app); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	assetTag, err := api.assettagapp.Create(ctx, app)
	if err != nil {
		return errs.NewError(err)
	}

	return assetTag
}

func (api *api) update(ctx context.Context, r *http.Request) web.Encoder {
	var app assettagapp.UpdateAssetTag
	if err := web.Decode(r, &app); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	assetTagID := web.Param(r, "asset_tag_id")
	parsed, err := uuid.Parse(assetTagID)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	assetTag, err := api.assettagapp.Update(ctx, app, parsed)
	if err != nil {
		return errs.NewError(err)
	}

	return assetTag
}

func (api *api) delete(ctx context.Context, r *http.Request) web.Encoder {
	assetTagID := web.Param(r, "asset_tag_id")

	parsed, err := uuid.Parse(assetTagID)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	err = api.assettagapp.Delete(ctx, parsed)
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

	assetTags, err := api.assettagapp.Query(ctx, qp)
	if err != nil {
		return errs.NewError(err)
	}

	return assetTags
}

func (api *api) queryByID(ctx context.Context, r *http.Request) web.Encoder {
	assetTagID := web.Param(r, "asset_tag_id")

	parsed, err := uuid.Parse(assetTagID)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	assetTag, err := api.assettagapp.QueryByID(ctx, parsed)
	if err != nil {
		return errs.NewError(err)
	}

	return assetTag
}
