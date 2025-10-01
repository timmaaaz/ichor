package dataapi

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/domain/dataapp"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/foundation/web"
)

type api struct {
	dataapp *dataapp.App
}

func newAPI(dataapp *dataapp.App) *api {
	return &api{
		dataapp: dataapp,
	}
}

func (api *api) create(ctx context.Context, r *http.Request) web.Encoder {
	var app dataapp.NewTableConfig
	if err := web.Decode(r, &app); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	tConfig, err := api.dataapp.Create(ctx, app)
	if err != nil {
		return errs.NewError(err)
	}

	return tConfig
}

func (api *api) update(ctx context.Context, r *http.Request) web.Encoder {
	var app dataapp.UpdateTableConfig
	if err := web.Decode(r, &app); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	id := web.Param(r, "table_config_id")
	parsed, err := uuid.Parse(id)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	tConfig, err := api.dataapp.Update(ctx, parsed, app)
	if err != nil {
		return errs.NewError(err)
	}

	return tConfig
}

func (api *api) delete(ctx context.Context, r *http.Request) web.Encoder {
	id := web.Param(r, "table_config_id")
	parsed, err := uuid.Parse(id)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	if err := api.dataapp.Delete(ctx, parsed); err != nil {
		return errs.NewError(err)
	}

	return nil
}

func (api *api) queryByID(ctx context.Context, r *http.Request) web.Encoder {
	id := web.Param(r, "table_config_id")
	parsed, err := uuid.Parse(id)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	tableConfig, err := api.dataapp.QueryByID(ctx, parsed)
	if err != nil {
		return errs.NewError(err)
	}

	return tableConfig
}

func (api *api) queryByName(ctx context.Context, r *http.Request) web.Encoder {
	name := web.Param(r, "name")

	tableConfig, err := api.dataapp.QueryByName(ctx, name)
	if err != nil {
		return errs.NewError(err)
	}

	return tableConfig
}

func (api *api) queryByUser(ctx context.Context, r *http.Request) web.Encoder {
	id := web.Param(r, "user_id")
	parsed, err := uuid.Parse(id)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	tableConfigs, err := api.dataapp.QueryByUser(ctx, parsed)
	if err != nil {
		return errs.NewError(err)
	}

	return tableConfigs
}

func (api *api) executeQuery(ctx context.Context, r *http.Request) web.Encoder {
	id := web.Param(r, "table_config_id")
	parsed, err := uuid.Parse(id)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	// Query filters/pagination stuff
	var app dataapp.TableQuery
	if err := web.Decode(r, &app); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	tableData, err := api.dataapp.ExecuteQuery(ctx, parsed, app)
	if err != nil {
		return errs.NewError(err)
	}

	return tableData
}

func (api *api) executeQueryByName(ctx context.Context, r *http.Request) web.Encoder {
	name := web.Param(r, "name")

	var app dataapp.TableQuery
	if err := web.Decode(r, &app); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	tableData, err := api.dataapp.ExecuteQueryByName(ctx, name, app)
	if err != nil {
		return errs.NewError(err)
	}

	return tableData
}

func (api *api) validateConfig(ctx context.Context, r *http.Request) web.Encoder {
	var app dataapp.NewTableConfig
	if err := web.Decode(r, &app); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	if err := api.dataapp.ValidateConfig(ctx, app); err != nil {
		return errs.NewError(err)
	}

	return nil
}
