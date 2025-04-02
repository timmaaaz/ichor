package serialnumberapi

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/domain/lots/serialnumberapp"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/foundation/web"
)

type api struct {
	serialnumberapp *serialnumberapp.App
}

func newAPI(serialnumberApp *serialnumberapp.App) *api {
	return &api{
		serialnumberapp: serialnumberApp,
	}
}

func (api *api) create(ctx context.Context, r *http.Request) web.Encoder {
	var app serialnumberapp.NewSerialNumber
	if err := web.Decode(r, &app); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	sn, err := api.serialnumberapp.Create(ctx, app)
	if err != nil {
		return errs.NewError(err)
	}

	return sn
}

func (api *api) update(ctx context.Context, r *http.Request) web.Encoder {
	var app serialnumberapp.UpdateSerialNumber
	if err := web.Decode(r, &app); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	snID := web.Param(r, "serial_id")
	parsed, err := uuid.Parse(snID)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	sn, err := api.serialnumberapp.Update(ctx, app, parsed)
	if err != nil {
		return errs.NewError(err)
	}

	return sn
}

func (api *api) delete(ctx context.Context, r *http.Request) web.Encoder {
	snID := web.Param(r, "serial_id")
	parsed, err := uuid.Parse(snID)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	err = api.serialnumberapp.Delete(ctx, parsed)
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

	sn, err := api.serialnumberapp.Query(ctx, qp)
	if err != nil {
		return errs.NewError(err)
	}

	return sn
}

func (api *api) queryByID(ctx context.Context, r *http.Request) web.Encoder {
	snID := web.Param(r, "serial_id")
	parsed, err := uuid.Parse(snID)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	sn, err := api.serialnumberapp.QueryByID(ctx, parsed)
	if err != nil {
		return errs.NewError(err)
	}

	return sn
}
