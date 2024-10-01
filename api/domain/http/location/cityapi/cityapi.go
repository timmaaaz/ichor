package cityapi

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/domain/location/cityapp"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/foundation/web"
)

type api struct {
	cityapp *cityapp.App
}

func newAPI(cityapp *cityapp.App) *api {
	return &api{
		cityapp: cityapp,
	}
}

func (api *api) create(ctx context.Context, r *http.Request) web.Encoder {
	var app cityapp.NewCity
	if err := web.Decode(r, &app); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	city, err := api.cityapp.Create(ctx, app)
	if err != nil {
		return errs.NewError(err)
	}

	return city
}

func (api *api) update(ctx context.Context, r *http.Request) web.Encoder {
	var app cityapp.UpdateCity
	if err := web.Decode(r, &app); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	cityID := web.Param(r, "city_id")
	parsed, err := uuid.Parse(cityID)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	city, err := api.cityapp.Update(ctx, app, parsed)
	if err != nil {
		return errs.NewError(err)
	}

	return city
}

func (api *api) delete(ctx context.Context, r *http.Request) web.Encoder {
	cityID := web.Param(r, "city_id")

	parsed, err := uuid.Parse(cityID)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	err = api.cityapp.Delete(ctx, parsed)
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

	cities, err := api.cityapp.Query(ctx, qp)
	if err != nil {
		return errs.NewError(err)
	}

	return cities
}

func (api *api) queryByID(ctx context.Context, r *http.Request) web.Encoder {
	cityID := web.Param(r, "city_id")

	parsed, err := uuid.Parse(cityID)

	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	city, err := api.cityapp.QueryByID(ctx, parsed)
	if err != nil {
		return errs.NewError(err)
	}

	return city
}
