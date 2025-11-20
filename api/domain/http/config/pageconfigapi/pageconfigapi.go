package pageconfigapi

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/domain/config/pageconfigapp"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/foundation/web"
)

type api struct {
	pageConfigApp *pageconfigapp.App
}

func newAPI(pageConfigApp *pageconfigapp.App) *api {
	return &api{
		pageConfigApp: pageConfigApp,
	}
}

// create adds a new page configuration to the system.
func (api *api) create(ctx context.Context, r *http.Request) web.Encoder {
	var app pageconfigapp.NewPageConfig
	if err := web.Decode(r, &app); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	config, err := api.pageConfigApp.Create(ctx, app)
	if err != nil {
		return errs.NewError(err)
	}

	return config
}

// update modifies an existing page configuration.
func (api *api) update(ctx context.Context, r *http.Request) web.Encoder {
	var app pageconfigapp.UpdatePageConfig
	if err := web.Decode(r, &app); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	configID, err := uuid.Parse(web.Param(r, "config_id"))
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	config, err := api.pageConfigApp.Update(ctx, app, configID)
	if err != nil {
		return errs.NewError(err)
	}

	return config
}

// delete removes a page configuration from the system.
func (api *api) delete(ctx context.Context, r *http.Request) web.Encoder {
	configID, err := uuid.Parse(web.Param(r, "config_id"))
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	if err := api.pageConfigApp.Delete(ctx, configID); err != nil {
		return errs.NewError(err)
	}

	return nil
}

// queryByID retrieves a single page configuration by ID.
func (api *api) queryByID(ctx context.Context, r *http.Request) web.Encoder {
	configID, err := uuid.Parse(web.Param(r, "config_id"))
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	config, err := api.pageConfigApp.QueryByID(ctx, configID)
	if err != nil {
		return errs.NewError(err)
	}

	return config
}

// queryByName retrieves the default page configuration by name.
func (api *api) queryByName(ctx context.Context, r *http.Request) web.Encoder {
	name := web.Param(r, "name")

	config, err := api.pageConfigApp.QueryByName(ctx, name)
	if err != nil {
		return errs.NewError(err)
	}

	return config
}

// queryAll retrieves all page configurations from the system.
func (api *api) queryAll(ctx context.Context, r *http.Request) web.Encoder {
	configs, err := api.pageConfigApp.QueryAll(ctx)
	if err != nil {
		return errs.NewError(err)
	}

	return configs
}
