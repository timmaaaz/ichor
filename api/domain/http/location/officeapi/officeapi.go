package officeapi

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/domain/location/officeapp"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/foundation/web"
)

type api struct {
	officeapp *officeapp.App
}

func newAPI(officeapp *officeapp.App) *api {
	return &api{
		officeapp: officeapp,
	}
}

func (api *api) create(ctx context.Context, r *http.Request) web.Encoder {
	var app officeapp.NewOffice
	if err := web.Decode(r, &app); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	assetTag, err := api.officeapp.Create(ctx, app)
	if err != nil {
		return errs.NewError(err)
	}

	return assetTag
}

func (api *api) update(ctx context.Context, r *http.Request) web.Encoder {
	var app officeapp.UpdateOffice
	if err := web.Decode(r, &app); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	assetTagID := web.Param(r, "office_id")
	parsed, err := uuid.Parse(assetTagID)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	assetTag, err := api.officeapp.Update(ctx, app, parsed)
	if err != nil {
		return errs.NewError(err)
	}

	return assetTag
}

func (api *api) delete(ctx context.Context, r *http.Request) web.Encoder {
	assetTagID := web.Param(r, "office_id")

	parsed, err := uuid.Parse(assetTagID)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	err = api.officeapp.Delete(ctx, parsed)
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

	assetTags, err := api.officeapp.Query(ctx, qp)
	if err != nil {
		return errs.NewError(err)
	}

	return assetTags
}

func (api *api) queryByID(ctx context.Context, r *http.Request) web.Encoder {
	assetTagID := web.Param(r, "office_id")

	parsed, err := uuid.Parse(assetTagID)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	assetTag, err := api.officeapp.QueryByID(ctx, parsed)
	if err != nil {
		return errs.NewError(err)
	}

	return assetTag
}
