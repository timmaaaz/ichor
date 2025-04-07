package transferorderapi

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/domain/movement/transferorderapp"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/foundation/web"
)

type api struct {
	transferorderapp *transferorderapp.App
}

func newAPI(transferorderapp *transferorderapp.App) *api {
	return &api{
		transferorderapp: transferorderapp,
	}
}

func (api *api) create(ctx context.Context, r *http.Request) web.Encoder {
	var app transferorderapp.NewTransferOrder
	if err := web.Decode(r, &app); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	to, err := api.transferorderapp.Create(ctx, app)
	if err != nil {
		return errs.NewError(err)
	}

	return to
}

func (api *api) update(ctx context.Context, r *http.Request) web.Encoder {
	var app transferorderapp.UpdateTransferOrder
	if err := web.Decode(r, &app); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	toID := web.Param(r, "transfer_id")
	parsed, err := uuid.Parse(toID)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	to, err := api.transferorderapp.Update(ctx, parsed, app)
	if err != nil {
		return errs.NewError(err)
	}

	return to
}

func (api *api) delete(ctx context.Context, r *http.Request) web.Encoder {
	toID := web.Param(r, "transfer_id")
	parsed, err := uuid.Parse(toID)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	if err := api.transferorderapp.Delete(ctx, parsed); err != nil {
		return errs.NewError(err)
	}

	return nil
}

func (api *api) query(ctx context.Context, r *http.Request) web.Encoder {
	qp, err := parseQueryParams(r)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	tos, err := api.transferorderapp.Query(ctx, qp)
	if err != nil {
		return errs.NewError(err)
	}

	return tos
}

func (api *api) queryByID(ctx context.Context, r *http.Request) web.Encoder {
	toID := web.Param(r, "transfer_id")
	parsed, err := uuid.Parse(toID)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	to, err := api.transferorderapp.QueryByID(ctx, parsed)
	if err != nil {
		return errs.NewError(err)
	}

	return to
}
