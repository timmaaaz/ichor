package presenceapi

import (
	"net/http"

	"github.com/timmaaaz/ichor/api/sdk/http/mid"
	"github.com/timmaaaz/ichor/app/sdk/authclient"
	"github.com/timmaaaz/ichor/business/domain/core/permissionsbus"
	"github.com/timmaaaz/ichor/foundation/logger"
	"github.com/timmaaaz/ichor/foundation/web"
)

// Config holds the dependencies for the presence API routes.
type Config struct {
	Log            *logger.Logger
	AlertHub       ActiveWorkersHub
	AuthClient     *authclient.Client
	PermissionsBus *permissionsbus.Business
}

// Routes registers the floor presence endpoints.
func Routes(app *web.App, cfg Config) {
	const version = "v1"

	authen := mid.Authenticate(cfg.AuthClient)
	a := newAPI(cfg.AlertHub)

	// activeWorkers is intentionally open to all authenticated roles —
	// floor supervisors, managers, and ops staff all need presence visibility.
	app.HandlerFunc(http.MethodGet, version, "/floor/active-workers", a.activeWorkers, authen)
}
