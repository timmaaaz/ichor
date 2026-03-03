package settingsapi

import (
	"context"
	"net/http"

	"github.com/timmaaaz/ichor/app/domain/config/settingsapp"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/foundation/web"
)

type api struct {
	settingsapp *settingsapp.App
}

func newAPI(settingsapp *settingsapp.App) *api {
	return &api{
		settingsapp: settingsapp,
	}
}

func (api *api) create(ctx context.Context, r *http.Request) web.Encoder {
	var app settingsapp.NewSetting
	if err := web.Decode(r, &app); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	setting, err := api.settingsapp.Create(ctx, app)
	if err != nil {
		return errs.NewError(err)
	}

	return setting
}

func (api *api) update(ctx context.Context, r *http.Request) web.Encoder {
	var app settingsapp.UpdateSetting
	if err := web.Decode(r, &app); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	key := web.Param(r, "key")

	setting, err := api.settingsapp.Update(ctx, key, app)
	if err != nil {
		return errs.NewError(err)
	}

	return setting
}

func (api *api) delete(ctx context.Context, r *http.Request) web.Encoder {
	key := web.Param(r, "key")

	if err := api.settingsapp.Delete(ctx, key); err != nil {
		return errs.NewError(err)
	}

	return nil
}

func (api *api) query(ctx context.Context, r *http.Request) web.Encoder {
	qp, err := parseQueryParams(r)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	settings, err := api.settingsapp.Query(ctx, qp)
	if err != nil {
		return errs.NewError(err)
	}

	return settings
}

func (api *api) queryByKey(ctx context.Context, r *http.Request) web.Encoder {
	key := web.Param(r, "key")

	setting, err := api.settingsapp.QueryByKey(ctx, key)
	if err != nil {
		return errs.NewError(err)
	}

	return setting
}
