package countryapi

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/domain/geography/countryapp"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/foundation/web"
)

type api struct {
	countryapp *countryapp.App
}

func newAPI(countryapp *countryapp.App) *api {
	return &api{
		countryapp: countryapp,
	}
}

func (api *api) query(ctx context.Context, r *http.Request) web.Encoder {
	qp, err := parseQueryParams(r)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	countries, err := api.countryapp.Query(ctx, qp)
	if err != nil {
		return errs.NewError(err)
	}

	return countries
}

func (api *api) queryByID(ctx context.Context, r *http.Request) web.Encoder {
	countryID := web.Param(r, "country_id")

	parsed, err := uuid.Parse(countryID)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	country, err := api.countryapp.QueryByID(ctx, parsed)
	if err != nil {
		return errs.NewError(err)
	}

	return country
}
