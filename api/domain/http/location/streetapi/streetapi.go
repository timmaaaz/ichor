package streetapi

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/domain/location/streetapp"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/foundation/web"
)

type api struct {
	streetapp *streetapp.App
}

func newAPI(streetapp *streetapp.App) *api {
	return &api{
		streetapp: streetapp,
	}
}

func (api *api) create(ctx context.Context, r *http.Request) web.Encoder {
	var app streetapp.NewStreet
	if err := web.Decode(r, &app); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	street, err := api.streetapp.Create(ctx, app)
	if err != nil {
		return errs.NewError(err)
	}

	return street
}

func (api *api) update(ctx context.Context, r *http.Request) web.Encoder {
	var app streetapp.UpdateStreet
	if err := web.Decode(r, &app); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	streetID := web.Param(r, "street_id")
	parsed, err := uuid.Parse(streetID)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	street, err := api.streetapp.Update(ctx, app, parsed)
	if err != nil {
		return errs.NewError(err)
	}

	return street
}

func (api *api) delete(ctx context.Context, r *http.Request) web.Encoder {
	streetID := web.Param(r, "street_id")

	parsed, err := uuid.Parse(streetID)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	err = api.streetapp.Delete(ctx, parsed)
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

	streets, err := api.streetapp.Query(ctx, qp)
	if err != nil {
		return errs.NewError(err)
	}

	return streets
}

func (api *api) queryByID(ctx context.Context, r *http.Request) web.Encoder {
	streetID := web.Param(r, "street_id")

	parsed, err := uuid.Parse(streetID)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	street, err := api.streetapp.QueryByID(ctx, parsed)
	if err != nil {
		return errs.NewError(err)
	}

	return street
}
