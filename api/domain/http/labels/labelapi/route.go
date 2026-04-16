package labelapi

import (
	"net/http"

	"github.com/timmaaaz/ichor/api/sdk/http/mid"
	"github.com/timmaaaz/ichor/app/domain/labels/labelapp"
	"github.com/timmaaaz/ichor/app/sdk/authclient"
	"github.com/timmaaaz/ichor/foundation/logger"
	"github.com/timmaaaz/ichor/foundation/web"
)

// Config holds the dependencies for the label API routes. Matches the
// shape used by the other app-oriented APIs in this repo (directedworkapi
// pattern — bus-less, pre-constructed App) because labels do not yet
// have a permissions/table-access entry.
type Config struct {
	Log        *logger.Logger
	LabelApp   *labelapp.App
	AuthClient *authclient.Client
}

// Routes registers the three label endpoints behind Authenticate. No
// Authorize middleware is applied at Phase 0b — labels lack a row in
// core.table_access, and the plan explicitly calls for no role
// restrictions at this phase.
func Routes(app *web.App, cfg Config) {
	const version = "v1"

	authen := mid.Authenticate(cfg.AuthClient)
	api := newAPI(cfg.LabelApp)

	app.HandlerFunc(http.MethodGet, version, "/labels", api.query, authen)
	app.HandlerFunc(http.MethodPost, version, "/labels/print", api.print, authen)
	app.HandlerFunc(http.MethodPost, version, "/labels/render-print", api.renderPrint, authen)
}
