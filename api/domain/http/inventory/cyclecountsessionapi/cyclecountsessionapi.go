package cyclecountsessionapi

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/domain/inventory/cyclecountsessionapp"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/foundation/web"
)

type api struct {
	cyclecountsessionapp *cyclecountsessionapp.App
}

func newAPI(cyclecountsessionapp *cyclecountsessionapp.App) *api {
	return &api{
		cyclecountsessionapp: cyclecountsessionapp,
	}
}

func (api *api) create(ctx context.Context, r *http.Request) web.Encoder {
	var app cyclecountsessionapp.NewCycleCountSession
	if err := web.Decode(r, &app); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	session, err := api.cyclecountsessionapp.Create(ctx, app)
	if err != nil {
		return errs.NewError(err)
	}

	return session
}

func (api *api) update(ctx context.Context, r *http.Request) web.Encoder {
	var app cyclecountsessionapp.UpdateCycleCountSession
	if err := web.Decode(r, &app); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	sessionID := web.Param(r, "session_id")
	parsed, err := uuid.Parse(sessionID)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	session, err := api.cyclecountsessionapp.Update(ctx, parsed, app)
	if err != nil {
		return errs.NewError(err)
	}

	return session
}

func (api *api) delete(ctx context.Context, r *http.Request) web.Encoder {
	sessionID := web.Param(r, "session_id")
	parsed, err := uuid.Parse(sessionID)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	if err := api.cyclecountsessionapp.Delete(ctx, parsed); err != nil {
		return errs.NewError(err)
	}

	return nil
}

func (api *api) query(ctx context.Context, r *http.Request) web.Encoder {
	qp, err := parseQueryParams(r)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	sessions, err := api.cyclecountsessionapp.Query(ctx, qp)
	if err != nil {
		return errs.NewError(err)
	}

	return sessions
}

func (api *api) queryByID(ctx context.Context, r *http.Request) web.Encoder {
	sessionID := web.Param(r, "session_id")
	parsed, err := uuid.Parse(sessionID)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	session, err := api.cyclecountsessionapp.QueryByID(ctx, parsed)
	if err != nil {
		return errs.NewError(err)
	}

	return session
}
