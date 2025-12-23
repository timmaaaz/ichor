package timezoneapi

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/domain/geography/timezoneapp"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/foundation/web"
)

type api struct {
	timezoneapp *timezoneapp.App
}

func newAPI(timezoneapp *timezoneapp.App) *api {
	return &api{
		timezoneapp: timezoneapp,
	}
}

func (api *api) create(ctx context.Context, r *http.Request) web.Encoder {
	var app timezoneapp.NewTimezone
	if err := web.Decode(r, &app); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	tz, err := api.timezoneapp.Create(ctx, app)
	if err != nil {
		return errs.NewError(err)
	}

	return tz
}

func (api *api) update(ctx context.Context, r *http.Request) web.Encoder {
	var app timezoneapp.UpdateTimezone
	if err := web.Decode(r, &app); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	timezoneID := web.Param(r, "timezone_id")
	parsed, err := uuid.Parse(timezoneID)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	tz, err := api.timezoneapp.Update(ctx, app, parsed)
	if err != nil {
		return errs.NewError(err)
	}

	return tz
}

func (api *api) delete(ctx context.Context, r *http.Request) web.Encoder {
	timezoneID := web.Param(r, "timezone_id")

	parsed, err := uuid.Parse(timezoneID)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	err = api.timezoneapp.Delete(ctx, parsed)
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

	tzs, err := api.timezoneapp.Query(ctx, qp)
	if err != nil {
		return errs.NewError(err)
	}

	return tzs
}

func (api *api) queryByID(ctx context.Context, r *http.Request) web.Encoder {
	timezoneID := web.Param(r, "timezone_id")

	parsed, err := uuid.Parse(timezoneID)

	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	tz, err := api.timezoneapp.QueryByID(ctx, parsed)
	if err != nil {
		return errs.NewError(err)
	}

	return tz
}
