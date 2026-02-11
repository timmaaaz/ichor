package referenceapi

import (
	"net/http"

	"github.com/timmaaaz/ichor/api/sdk/http/mid"
	"github.com/timmaaaz/ichor/app/sdk/authclient"
	"github.com/timmaaaz/ichor/business/domain/core/permissionsbus"
	"github.com/timmaaaz/ichor/business/sdk/workflow"
	"github.com/timmaaaz/ichor/foundation/logger"
	"github.com/timmaaaz/ichor/foundation/web"
)

// Config holds the dependencies for the reference API routes.
type Config struct {
	Log            *logger.Logger
	WorkflowBus    *workflow.Business
	AuthClient     *authclient.Client
	PermissionsBus *permissionsbus.Business
	ActionRegistry *workflow.ActionRegistry
}

// RouteTable is the table name used for permission checks (consistency with other packages).
const RouteTable = "workflow.reference_data"

// Routes registers the reference data API routes.
func Routes(app *web.App, cfg Config) {
	const version = "v1"

	api := newAPI(cfg)
	authen := mid.Authenticate(cfg.AuthClient)

	// Reference data endpoints - authentication only, no special permissions needed
	app.HandlerFunc(http.MethodGet, version, "/workflow/trigger-types", api.queryTriggerTypes, authen)
	app.HandlerFunc(http.MethodGet, version, "/workflow/entity-types", api.queryEntityTypes, authen)
	app.HandlerFunc(http.MethodGet, version, "/workflow/entities", api.queryEntities, authen)
	app.HandlerFunc(http.MethodGet, version, "/workflow/action-types", api.queryActionTypes, authen)
	app.HandlerFunc(http.MethodGet, version, "/workflow/action-types/{type}/schema", api.queryActionTypeSchema, authen)
	app.HandlerFunc(http.MethodGet, version, "/workflow/templates", api.queryTemplates, authen)
	app.HandlerFunc(http.MethodGet, version, "/workflow/templates/active", api.queryActiveTemplates, authen)
}
