package assetconditionapi

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/domain/assets/assetconditionapp"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/foundation/web"
)

type api struct {
	assetconditionapp *assetconditionapp.App
}

func newAPI(assetconditionapp *assetconditionapp.App) *api {
	return &api{
		assetconditionapp: assetconditionapp,
	}
}

func (api *api) create(ctx context.Context, r *http.Request) web.Encoder {
	var app assetconditionapp.NewAssetCondition
	if err := web.Decode(r, &app); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	assetCondition, err := api.assetconditionapp.Create(ctx, app)
	if err != nil {
		return errs.NewError(err)
	}

	return assetCondition
}

func (api *api) update(ctx context.Context, r *http.Request) web.Encoder {
	var app assetconditionapp.UpdateAssetCondition
	if err := web.Decode(r, &app); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	assetConditionID := web.Param(r, "asset_condition_id")
	parsed, err := uuid.Parse(assetConditionID)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	assetCondition, err := api.assetconditionapp.Update(ctx, app, parsed)
	if err != nil {
		return errs.NewError(err)
	}

	return assetCondition
}

func (api *api) delete(ctx context.Context, r *http.Request) web.Encoder {
	assetConditionID := web.Param(r, "asset_condition_id")

	parsed, err := uuid.Parse(assetConditionID)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	err = api.assetconditionapp.Delete(ctx, parsed)
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

	result, err := api.assetconditionapp.Query(ctx, qp)
	if err != nil {
		return errs.NewError(err)
	}

	return result
}

func (api *api) queryByID(ctx context.Context, r *http.Request) web.Encoder {
	assetConditionID := web.Param(r, "asset_condition_id")

	parsed, err := uuid.Parse(assetConditionID)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	result, err := api.assetconditionapp.QueryByID(ctx, parsed)
	if err != nil {
		return errs.NewError(err)
	}

	return result
}
