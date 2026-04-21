package mid

import (
	"context"
	"errors"

	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/business/domain/scenarios/scenariobus"
	"github.com/timmaaaz/ichor/business/sdk/sqldb"
)

// activeReader is the minimal scenariobus surface the ActiveScenario
// middleware depends on. External callers pass *scenariobus.Business, which
// satisfies this interface; tests use a hand-rolled fake.
type activeReader interface {
	Active(ctx context.Context) (scenariobus.Scenario, error)
}

// ActiveScenario reads the singleton scenarios_active row once per request
// and populates both the mid scenario key (read via GetScenario) and the
// paired sqldb key (read via sqldb.GetScenarioFilter). When no scenario is
// active, both keys stay unset and downstream reads/writes proceed unfiltered.
func ActiveScenario(ctx context.Context, bus activeReader, next HandlerFunc) Encoder {
	active, err := bus.Active(ctx)
	if err != nil {
		if errors.Is(err, scenariobus.ErrNotFound) {
			return next(ctx)
		}
		return errs.New(errs.Internal, err)
	}

	ctx = setScenario(ctx, active.ID)
	ctx = sqldb.SetScenarioFilter(ctx, active.ID)

	return next(ctx)
}
