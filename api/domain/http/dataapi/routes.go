package dataapi

import (
	"net/http"

	"github.com/timmaaaz/ichor/api/sdk/http/mid"
	"github.com/timmaaaz/ichor/app/domain/config/pageactionapp"
	"github.com/timmaaaz/ichor/app/domain/dataapp"
	"github.com/timmaaaz/ichor/app/sdk/auth"
	"github.com/timmaaaz/ichor/app/sdk/authclient"
	"github.com/timmaaaz/ichor/business/domain/core/permissionsbus"
	"github.com/timmaaaz/ichor/business/sdk/tablebuilder"
	"github.com/timmaaaz/ichor/foundation/logger"
	"github.com/timmaaaz/ichor/foundation/web"
)

// TODO: Need to work permissions in here to be based on the contents of the
// configs, not the route table. These are amalgamations of different tables
// and data across the system and should be handled differently. We also are
// going to need to take into account different types of returns like tables,
// graphs, reports, whatever and these should probably be categorized by
// subdomains within data.

type Config struct {
	Log            *logger.Logger
	ConfigStore    *tablebuilder.ConfigStore
	TableStore     *tablebuilder.Store
	PageActionApp  *pageactionapp.App
	AuthClient     *authclient.Client
	PermissionsBus *permissionsbus.Business
}

const (
	RouteTable = "table_configs"
)

func Routes(app *web.App, cfg Config) {

	const version = "v1"
	authen := mid.Authenticate(cfg.AuthClient)
	api := newAPI(dataapp.NewApp(cfg.ConfigStore, cfg.TableStore, cfg.PageActionApp))

	// configstore
	app.HandlerFunc(http.MethodPost, version, "/data", api.create, authen,
		mid.Authorize(cfg.AuthClient, cfg.PermissionsBus, RouteTable, permissionsbus.Actions.Create, auth.RuleAny))

	app.HandlerFunc(http.MethodPut, version, "/data/{table_config_id}", api.update, authen,
		mid.Authorize(cfg.AuthClient, cfg.PermissionsBus, RouteTable, permissionsbus.Actions.Update, auth.RuleAny))

	app.HandlerFunc(http.MethodDelete, version, "/data/{table_config_id}", api.delete, authen,
		mid.Authorize(cfg.AuthClient, cfg.PermissionsBus, RouteTable, permissionsbus.Actions.Delete, auth.RuleAny))

	app.HandlerFunc(http.MethodGet, version, "/data/id/{table_config_id}", api.queryByID, authen,
		mid.Authorize(cfg.AuthClient, cfg.PermissionsBus, RouteTable, permissionsbus.Actions.Read, auth.RuleAny))

	app.HandlerFunc(http.MethodGet, version, "/data/name/{name}", api.queryByName, authen,
		mid.Authorize(cfg.AuthClient, cfg.PermissionsBus, RouteTable, permissionsbus.Actions.Read, auth.RuleAny))

	app.HandlerFunc(http.MethodGet, version, "/data/user/{user_id}", api.queryByUser, authen,
		mid.Authorize(cfg.AuthClient, cfg.PermissionsBus, RouteTable, permissionsbus.Actions.Read, auth.RuleAny))

	// store
	app.HandlerFunc(http.MethodPost, version, "/data/execute/{table_config_id}", api.executeQuery, authen,
		mid.Authorize(cfg.AuthClient, cfg.PermissionsBus, RouteTable, permissionsbus.Actions.Read, auth.RuleAny))

	app.HandlerFunc(http.MethodPost, version, "/data/execute/name/{name}", api.executeQueryByName, authen,
		mid.Authorize(cfg.AuthClient, cfg.PermissionsBus, RouteTable, permissionsbus.Actions.Read, auth.RuleAny))

	// Count endpoints
	app.HandlerFunc(http.MethodPost, version, "/data/execute/count/{table_config_id}", api.executeQueryCountByID, authen,
		mid.Authorize(cfg.AuthClient, cfg.PermissionsBus, RouteTable, permissionsbus.Actions.Read, auth.RuleAny))

	app.HandlerFunc(http.MethodPost, version, "/data/execute/name/count/{name}", api.executeQueryCountByName, authen,
		mid.Authorize(cfg.AuthClient, cfg.PermissionsBus, RouteTable, permissionsbus.Actions.Read, auth.RuleAny))

	app.HandlerFunc(http.MethodPost, version, "/data/validate", api.validateConfig, authen,
		mid.Authorize(cfg.AuthClient, cfg.PermissionsBus, RouteTable, permissionsbus.Actions.Read, auth.RuleAny))

	// PageConfig routes
	app.HandlerFunc(http.MethodPost, version, "/data/page", api.createPageConfig, authen,
		mid.Authorize(cfg.AuthClient, cfg.PermissionsBus, RouteTable, permissionsbus.Actions.Create, auth.RuleAny))

	app.HandlerFunc(http.MethodPut, version, "/data/page/{page_config_id}", api.updatePageConfig, authen,
		mid.Authorize(cfg.AuthClient, cfg.PermissionsBus, RouteTable, permissionsbus.Actions.Update, auth.RuleAny))

	app.HandlerFunc(http.MethodDelete, version, "/data/page/{page_config_id}", api.deletePageConfig, authen,
		mid.Authorize(cfg.AuthClient, cfg.PermissionsBus, RouteTable, permissionsbus.Actions.Delete, auth.RuleAny))

	app.HandlerFunc(http.MethodGet, version, "/data/page/name/{name}", api.queryFullPageByName, authen,
		mid.Authorize(cfg.AuthClient, cfg.PermissionsBus, RouteTable, permissionsbus.Actions.Read, auth.RuleAny))

	app.HandlerFunc(http.MethodGet, version, "/data/page/name/{name}/user/{user_id}", api.queryFullPageByNameAndUserID, authen,
		mid.Authorize(cfg.AuthClient, cfg.PermissionsBus, RouteTable, permissionsbus.Actions.Read, auth.RuleAny))

	app.HandlerFunc(http.MethodGet, version, "/data/page/id/{page_config_id}", api.queryFullPageByID, authen,
		mid.Authorize(cfg.AuthClient, cfg.PermissionsBus, RouteTable, permissionsbus.Actions.Read, auth.RuleAny))

	// PageTabConfig routes
	app.HandlerFunc(http.MethodPost, version, "/data/page/tab", api.createPageTabConfig, authen,
		mid.Authorize(cfg.AuthClient, cfg.PermissionsBus, RouteTable, permissionsbus.Actions.Create, auth.RuleAny))

	app.HandlerFunc(http.MethodPut, version, "/data/page/tab/{page_tab_config_id}", api.updatePageTabConfig, authen,
		mid.Authorize(cfg.AuthClient, cfg.PermissionsBus, RouteTable, permissionsbus.Actions.Update, auth.RuleAny))

	app.HandlerFunc(http.MethodDelete, version, "/data/page/tab/{page_tab_config_id}", api.deletePageTabConfig, authen,
		mid.Authorize(cfg.AuthClient, cfg.PermissionsBus, RouteTable, permissionsbus.Actions.Delete, auth.RuleAny))
}
