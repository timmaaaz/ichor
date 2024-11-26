package assettypeapi

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/domain/assettypeapp"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/foundation/web"
)

type api struct {
	assettypeapp *assettypeapp.App
}

func newAPI(assettypeapp *assettypeapp.App) *api {
	return &api{
		assettypeapp: assettypeapp,
	}
}

func (api *api) create(ctx context.Context, r *http.Request) web.Encoder {
	var app assettypeapp.NewAssetType
	if err := web.Decode(r, &app); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	status, err := api.assettypeapp.Create(ctx, app)
	if err != nil {
		return errs.NewError(err)
	}

	return status
}

func (api *api) update(ctx context.Context, r *http.Request) web.Encoder {
	var app assettypeapp.UpdateAssetType
	if err := web.Decode(r, &app); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	statusID := web.Param(r, "asset_type_id")
	parsed, err := uuid.Parse(statusID)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	status, err := api.assettypeapp.Update(ctx, app, parsed)
	if err != nil {
		return errs.NewError(err)
	}

	return status
}

func (api *api) delete(ctx context.Context, r *http.Request) web.Encoder {
	approvalStatusID := web.Param(r, "asset_type_id")

	parsed, err := uuid.Parse(approvalStatusID)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	err = api.assettypeapp.Delete(ctx, parsed)
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

	statuses, err := api.assettypeapp.Query(ctx, qp)
	if err != nil {
		return errs.NewError(err)
	}

	return statuses
}

func (api *api) queryByID(ctx context.Context, r *http.Request) web.Encoder {
	assetTypeID := web.Param(r, "asset_type_id")

	parsed, err := uuid.Parse(assetTypeID)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	aprvlStatus, err := api.assettypeapp.QueryByID(ctx, parsed)
	if err != nil {
		return errs.NewError(err)
	}

	return aprvlStatus
}
