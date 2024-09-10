package regionapi

import (
	"context"
	"net/http"

	"bitbucket.org/superiortechnologies/ichor/app/domain/location/regionapp"
	"bitbucket.org/superiortechnologies/ichor/app/sdk/errs"
	"bitbucket.org/superiortechnologies/ichor/foundation/web"
	"github.com/google/uuid"
)

type api struct {
	regionapp *regionapp.App
}

func newAPI(regionapp *regionapp.App) *api {
	return &api{
		regionapp: regionapp,
	}
}

func (api *api) query(ctx context.Context, r *http.Request) web.Encoder {
	qp, err := parseQueryParams(r)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	regions, err := api.regionapp.Query(ctx, qp)
	if err != nil {
		return errs.NewError(err)
	}

	return regions
}

func (api *api) queryByID(ctx context.Context, r *http.Request) web.Encoder {
	regionID := web.Param(r, "region_id")

	parsed, err := uuid.Parse(regionID)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	region, err := api.regionapp.QueryByID(ctx, parsed)
	if err != nil {
		errs.New(errs.InvalidArgument, err)
	}

	return region
}
