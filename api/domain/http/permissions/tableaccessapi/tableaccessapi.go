package tableaccessapi

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/domain/permissions/tableaccessapp"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/foundation/web"
)

type api struct {
	tableaccessapp *tableaccessapp.App
}

func newAPI(tableaccessapp *tableaccessapp.App) *api {
	return &api{
		tableaccessapp: tableaccessapp,
	}
}

func (api *api) create(ctx context.Context, r *http.Request) web.Encoder {
	var app tableaccessapp.NewTableAccess
	if err := web.Decode(r, &app); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	tableaccess, err := api.tableaccessapp.Create(ctx, app)
	if err != nil {
		return errs.NewError(err)
	}

	return tableaccess

}

func (api *api) update(ctx context.Context, r *http.Request) web.Encoder {
	var app tableaccessapp.UpdateTableAccess
	if err := web.Decode(r, &app); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	tableaccessID := web.Param(r, "table_access_id")
	parsed, err := uuid.Parse(tableaccessID)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	tableaccess, err := api.tableaccessapp.Update(ctx, app, parsed)
	if err != nil {
		return errs.NewError(err)
	}

	return tableaccess
}

func (api *api) delete(ctx context.Context, r *http.Request) web.Encoder {
	tableaccessID := web.Param(r, "table_access_id")
	parsed, err := uuid.Parse(tableaccessID)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	err = api.tableaccessapp.Delete(ctx, parsed)
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

	tableaccesses, err := api.tableaccessapp.Query(ctx, qp)
	if err != nil {
		return errs.NewError(err)
	}

	return tableaccesses
}

func (api *api) queryByID(ctx context.Context, r *http.Request) web.Encoder {
	tableaccessID := web.Param(r, "table_access_id")
	parsed, err := uuid.Parse(tableaccessID)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	tableaccess, err := api.tableaccessapp.QueryByID(ctx, parsed)
	if err != nil {
		return errs.NewError(err)
	}

	return tableaccess
}
