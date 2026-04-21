// Package scenarioapp provides the app-layer orchestration for scenario
// management: CRUD, active-scenario query, Load (swap scenario data), and
// Reset (re-apply current scenario). Fixtures are seed-only and have no
// app-layer representation.
package scenarioapp

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/business/domain/scenarios/scenariobus"
	"github.com/timmaaaz/ichor/business/sdk/order"
	"github.com/timmaaaz/ichor/business/sdk/page"
)

// App manages the set of app-layer APIs for the scenario domain.
type App struct {
	bus *scenariobus.Business
}

// NewApp constructs a scenario app API for use.
func NewApp(bus *scenariobus.Business) *App {
	return &App{bus: bus}
}

// Create inserts a new scenario.
func (a *App) Create(ctx context.Context, app NewScenario) (Scenario, error) {
	s, err := a.bus.Create(ctx, toBusNewScenario(app))
	if err != nil {
		if errors.Is(err, scenariobus.ErrUniqueName) {
			return Scenario{}, errs.New(errs.Aborted, scenariobus.ErrUniqueName)
		}
		return Scenario{}, errs.Newf(errs.Internal, "create: %s", err)
	}
	return toAppScenario(s), nil
}

// Update applies a partial patch to an existing scenario.
func (a *App) Update(ctx context.Context, scenarioID uuid.UUID, app UpdateScenario) (Scenario, error) {
	s, err := a.bus.QueryByID(ctx, scenarioID)
	if err != nil {
		if errors.Is(err, scenariobus.ErrNotFound) {
			return Scenario{}, errs.New(errs.NotFound, scenariobus.ErrNotFound)
		}
		return Scenario{}, errs.Newf(errs.Internal, "querybyid: %s", err)
	}

	updated, err := a.bus.Update(ctx, s, toBusUpdateScenario(app))
	if err != nil {
		if errors.Is(err, scenariobus.ErrUniqueName) {
			return Scenario{}, errs.New(errs.Aborted, scenariobus.ErrUniqueName)
		}
		return Scenario{}, errs.Newf(errs.Internal, "update: %s", err)
	}
	return toAppScenario(updated), nil
}

// Delete removes a scenario.
func (a *App) Delete(ctx context.Context, scenarioID uuid.UUID) error {
	s, err := a.bus.QueryByID(ctx, scenarioID)
	if err != nil {
		if errors.Is(err, scenariobus.ErrNotFound) {
			return errs.New(errs.NotFound, scenariobus.ErrNotFound)
		}
		return errs.Newf(errs.Internal, "querybyid: %s", err)
	}

	if err := a.bus.Delete(ctx, s); err != nil {
		return errs.Newf(errs.Internal, "delete: %s", err)
	}
	return nil
}

// QueryByID returns the scenario with the given ID.
func (a *App) QueryByID(ctx context.Context, scenarioID uuid.UUID) (Scenario, error) {
	s, err := a.bus.QueryByID(ctx, scenarioID)
	if err != nil {
		if errors.Is(err, scenariobus.ErrNotFound) {
			return Scenario{}, errs.New(errs.NotFound, scenariobus.ErrNotFound)
		}
		return Scenario{}, errs.Newf(errs.Internal, "querybyid: %s", err)
	}
	return toAppScenario(s), nil
}

// Query returns scenarios matching the optional name-prefix filter.
func (a *App) Query(ctx context.Context, qp QueryParams) (Scenarios, error) {
	pg, err := page.Parse(qp.Page, qp.Rows)
	if err != nil {
		return nil, errs.NewFieldsError("page", err)
	}

	filter, err := parseFilter(qp)
	if err != nil {
		return nil, errs.NewFieldsError("filter", err)
	}

	orderBy, err := order.Parse(orderByFields, qp.OrderBy, defaultOrderBy)
	if err != nil {
		return nil, errs.NewFieldsError("orderby", err)
	}

	scenarios, err := a.bus.Query(ctx, filter, orderBy, pg)
	if err != nil {
		return nil, errs.Newf(errs.Internal, "query: %s", err)
	}
	return toAppScenarios(scenarios), nil
}

// Active returns the currently active scenario.
// Returns HTTP 404 if no scenario is active.
func (a *App) Active(ctx context.Context) (Scenario, error) {
	s, err := a.bus.Active(ctx)
	if err != nil {
		if errors.Is(err, scenariobus.ErrNotFound) {
			return Scenario{}, errs.New(errs.NotFound, scenariobus.ErrNotFound)
		}
		return Scenario{}, errs.Newf(errs.Internal, "active: %s", err)
	}
	return toAppScenario(s), nil
}

// Load swaps the active scenario to the specified ID, replacing scoped data
// atomically. Returns empty on success (HTTP 204).
func (a *App) Load(ctx context.Context, scenarioID uuid.UUID) error {
	if err := a.bus.Load(ctx, scenarioID); err != nil {
		if errors.Is(err, scenariobus.ErrNotFound) {
			return errs.New(errs.NotFound, scenariobus.ErrNotFound)
		}
		return errs.Newf(errs.Internal, "load: %s", err)
	}
	return nil
}

// Reset re-applies the current active scenario. Returns HTTP 409 if no
// scenario is currently active.
func (a *App) Reset(ctx context.Context) error {
	if err := a.bus.Reset(ctx); err != nil {
		if errors.Is(err, scenariobus.ErrNoActiveScenario) {
			return errs.New(errs.FailedPrecondition, scenariobus.ErrNoActiveScenario)
		}
		return errs.Newf(errs.Internal, "reset: %s", err)
	}
	return nil
}
