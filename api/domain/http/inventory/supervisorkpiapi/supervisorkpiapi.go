package supervisorkpiapi

import (
	"context"
	"net/http"

	"github.com/timmaaaz/ichor/app/domain/inventory/supervisorkpiapp"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/foundation/web"
)

type api struct {
	supervisorKPIApp *supervisorkpiapp.App
}

func newAPI(app *supervisorkpiapp.App) *api {
	return &api{supervisorKPIApp: app}
}

func (a *api) query(ctx context.Context, r *http.Request) web.Encoder {
	kpis, err := a.supervisorKPIApp.Query(ctx)
	if err != nil {
		return errs.NewError(err)
	}

	return kpis
}
