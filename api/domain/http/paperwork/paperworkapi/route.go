// Package paperworkapi maintains the web-based API for paperwork.
package paperworkapi

import (
	"net/http"

	"github.com/timmaaaz/ichor/api/sdk/http/mid"
	"github.com/timmaaaz/ichor/app/domain/paperwork/paperworkapp"
	"github.com/timmaaaz/ichor/app/sdk/authclient"
	"github.com/timmaaaz/ichor/business/domain/paperwork/paperworkbus"
	"github.com/timmaaaz/ichor/foundation/logger"
	"github.com/timmaaaz/ichor/foundation/web"
)

// Config carries the dependencies for the paperwork API.
//
// Phase 0g.B2 has no PermissionsBus: B2 wires Authenticate only (handlers
// return 501; RBAC has no semantic content yet). Phase 0g.B3 adds Authorize
// once handlers do real work.
type Config struct {
	Log          *logger.Logger
	PaperworkBus *paperworkbus.Business
	AuthClient   *authclient.Client
}

// Routes registers paperwork endpoints behind Authenticate. All three
// handlers return 501 Not Implemented during B2.
func Routes(app *web.App, cfg Config) {
	const version = "v1"

	authen := mid.Authenticate(cfg.AuthClient)
	api := newAPI(paperworkapp.NewApp(cfg.PaperworkBus))

	app.HandlerFunc(http.MethodGet, version, "/paperwork/pick-sheet", api.pickSheet, authen)
	app.HandlerFunc(http.MethodGet, version, "/paperwork/receive-cover", api.receiveCover, authen)
	app.HandlerFunc(http.MethodGet, version, "/paperwork/transfer-sheet", api.transferSheet, authen)
}
