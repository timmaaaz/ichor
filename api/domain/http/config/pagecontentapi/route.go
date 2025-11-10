package pagecontentapi

import (
	"net/http"

	"github.com/timmaaaz/ichor/api/sdk/http/mid"
	"github.com/timmaaaz/ichor/app/domain/config/pagecontentapp"
	"github.com/timmaaaz/ichor/app/sdk/auth"
	"github.com/timmaaaz/ichor/app/sdk/authclient"
	"github.com/timmaaaz/ichor/business/domain/config/pagecontentbus"
	"github.com/timmaaaz/ichor/business/domain/core/permissionsbus"
	"github.com/timmaaaz/ichor/foundation/logger"
	"github.com/timmaaaz/ichor/foundation/web"
)

// Config contains all the mandatory systems required by handlers.
type Config struct {
	Log            *logger.Logger
	PageContentBus *pagecontentbus.Business
	AuthClient     *authclient.Client
	PermissionsBus *permissionsbus.Business
}

const (
	RouteTable = "page_content"
)

// Routes adds specific routes for this group.
func Routes(app *web.App, cfg Config) {
	const version = "v1"

	pageContentApp := pagecontentapp.NewApp(cfg.PageContentBus)
	api := newAPI(pageContentApp)
	authen := mid.Authenticate(cfg.AuthClient)

	// Public routes (read-only, authenticated users)
	app.HandlerFunc(http.MethodGet, version, "/config/page-content/{content_id}", api.queryByID, authen,
		mid.Authorize(cfg.AuthClient, cfg.PermissionsBus, "page_content", permissionsbus.Actions.Read, auth.RuleAny))

	app.HandlerFunc(http.MethodGet, version, "/config/page-configs/content/{page_config_id}", api.queryByPageConfigID, authen,
		mid.Authorize(cfg.AuthClient, cfg.PermissionsBus, "page_content", permissionsbus.Actions.Read, auth.RuleAny))

	app.HandlerFunc(http.MethodGet, version, "/config/page-configs/content/children/{page_config_id}", api.queryWithChildren, authen,
		mid.Authorize(cfg.AuthClient, cfg.PermissionsBus, "page_content", permissionsbus.Actions.Read, auth.RuleAny))

	// Admin-only routes (create, update, delete)
	app.HandlerFunc(http.MethodPost, version, "/config/page-content", api.create, authen,
		mid.Authorize(cfg.AuthClient, cfg.PermissionsBus, "page_content", permissionsbus.Actions.Create, auth.RuleAdminOnly))

	app.HandlerFunc(http.MethodPut, version, "/config/page-content/{content_id}", api.update, authen,
		mid.Authorize(cfg.AuthClient, cfg.PermissionsBus, "page_content", permissionsbus.Actions.Update, auth.RuleAdminOnly))

	app.HandlerFunc(http.MethodDelete, version, "/config/page-content/{content_id}", api.delete, authen,
		mid.Authorize(cfg.AuthClient, cfg.PermissionsBus, "page_content", permissionsbus.Actions.Delete, auth.RuleAdminOnly))
}
