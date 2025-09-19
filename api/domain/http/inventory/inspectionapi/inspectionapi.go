package inspectionapi

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/domain/inventory/inspectionapp"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/foundation/web"
)

type api struct {
	inspectionapp *inspectionapp.App
}

func newAPI(inspectionapp *inspectionapp.App) *api {
	return &api{
		inspectionapp: inspectionapp,
	}
}

func (api *api) create(ctx context.Context, r *http.Request) web.Encoder {
	var app inspectionapp.NewInspection
	if err := web.Decode(r, &app); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	inspection, err := api.inspectionapp.Create(ctx, app)
	if err != nil {
		return errs.NewError(err)
	}

	return inspection
}

func (api *api) update(ctx context.Context, r *http.Request) web.Encoder {
	var app inspectionapp.UpdateInspection
	if err := web.Decode(r, &app); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	inspectionID := web.Param(r, "inspection_id")
	parsed, err := uuid.Parse(inspectionID)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	inspection, err := api.inspectionapp.Update(ctx, app, parsed)
	if err != nil {
		return errs.NewError(err)
	}

	return inspection
}

func (api *api) delete(ctx context.Context, r *http.Request) web.Encoder {
	inspectionID := web.Param(r, "inspection_id")
	parsed, err := uuid.Parse(inspectionID)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	if err := api.inspectionapp.Delete(ctx, parsed); err != nil {
		return errs.NewError(err)
	}

	return nil
}

func (api *api) query(ctx context.Context, r *http.Request) web.Encoder {
	qp, err := parseQueryParams(r)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	inspections, err := api.inspectionapp.Query(ctx, qp)
	if err != nil {
		return errs.NewError(err)
	}

	return inspections
}

func (api *api) queryByID(ctx context.Context, r *http.Request) web.Encoder {
	inspectionID := web.Param(r, "inspection_id")
	parsed, err := uuid.Parse(inspectionID)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	inspection, err := api.inspectionapp.QueryByID(ctx, parsed)
	if err != nil {
		return errs.NewError(err)
	}

	return inspection
}
