package mid

import (
	"context"
	"net/http"

	"github.com/timmaaaz/ichor/app/sdk/mid"
	"github.com/timmaaaz/ichor/business/domain/scenarios/scenariobus"
	"github.com/timmaaaz/ichor/foundation/web"
)

// ActiveScenario reads the active scenario once per request and populates
// the scenario context keys consumed by app handlers and floor-scoped
// repositories. Wired into the app-level middleware chain by build/all/all.go
// when cfg.ScenariosEnabled is true (phase 0d).
func ActiveScenario(bus *scenariobus.Business) web.MidFunc {
	midFunc := func(ctx context.Context, r *http.Request, next mid.HandlerFunc) mid.Encoder {
		return mid.ActiveScenario(ctx, bus, next)
	}

	return addMidFunc(midFunc)
}
