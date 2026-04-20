// Package scenarioapi exposes HTTP routes for the scenario subsystem:
// CRUD on scenarios, active scenario query, Load (swap), and Reset.
package scenarioapi

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/domain/scenarios/scenarioapp"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/foundation/web"
)

type api struct {
	scenarioapp *scenarioapp.App
}

func newAPI(app *scenarioapp.App) *api {
	return &api{scenarioapp: app}
}

// create handles POST /v1/scenarios.
func (a *api) create(ctx context.Context, r *http.Request) web.Encoder {
	var app scenarioapp.NewScenario
	if err := web.Decode(r, &app); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	s, err := a.scenarioapp.Create(ctx, app)
	if err != nil {
		return errs.NewError(err)
	}
	return s
}

// update handles PUT /v1/scenarios/{id}.
func (a *api) update(ctx context.Context, r *http.Request) web.Encoder {
	var app scenarioapp.UpdateScenario
	if err := web.Decode(r, &app); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	id, err := uuid.Parse(web.Param(r, "id"))
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	s, err := a.scenarioapp.Update(ctx, id, app)
	if err != nil {
		return errs.NewError(err)
	}
	return s
}

// delete handles DELETE /v1/scenarios/{id}.
func (a *api) delete(ctx context.Context, r *http.Request) web.Encoder {
	id, err := uuid.Parse(web.Param(r, "id"))
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	if err := a.scenarioapp.Delete(ctx, id); err != nil {
		return errs.NewError(err)
	}
	return nil
}

// queryByID handles GET /v1/scenarios/{id}.
func (a *api) queryByID(ctx context.Context, r *http.Request) web.Encoder {
	id, err := uuid.Parse(web.Param(r, "id"))
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	s, err := a.scenarioapp.QueryByID(ctx, id)
	if err != nil {
		return errs.NewError(err)
	}
	return s
}

// query handles GET /v1/scenarios.
func (a *api) query(ctx context.Context, r *http.Request) web.Encoder {
	qp, err := parseQueryParams(r)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	scenarios, err := a.scenarioapp.Query(ctx, qp)
	if err != nil {
		return errs.NewError(err)
	}
	return scenarios
}

// active handles GET /v1/scenarios/active.
func (a *api) active(ctx context.Context, r *http.Request) web.Encoder {
	s, err := a.scenarioapp.Active(ctx)
	if err != nil {
		return errs.NewError(err)
	}
	return s
}

// load handles POST /v1/scenarios/{id}/load.
func (a *api) load(ctx context.Context, r *http.Request) web.Encoder {
	id, err := uuid.Parse(web.Param(r, "id"))
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	if err := a.scenarioapp.Load(ctx, id); err != nil {
		return errs.NewError(err)
	}
	return nil
}

// reset handles POST /v1/scenarios/active/reset.
func (a *api) reset(ctx context.Context, r *http.Request) web.Encoder {
	if err := a.scenarioapp.Reset(ctx); err != nil {
		return errs.NewError(err)
	}
	return nil
}
