package executionapi

import (
	"net/http"

	"github.com/timmaaaz/ichor/api/sdk/http/mid"
	"github.com/timmaaaz/ichor/app/sdk/authclient"
	"github.com/timmaaaz/ichor/business/sdk/workflow"
	"github.com/timmaaaz/ichor/foundation/logger"
	"github.com/timmaaaz/ichor/foundation/web"
)

// Config holds the dependencies for the execution API routes.
type Config struct {
	Log         *logger.Logger
	WorkflowBus *workflow.Business
	AuthClient  *authclient.Client
}

// RouteTable is the table name used for permission checks.
const RouteTable = "workflow.automation_executions"

// Routes registers the execution history API routes.
func Routes(app *web.App, cfg Config) {
	const version = "v1"

	api := newAPI(cfg)
	authen := mid.Authenticate(cfg.AuthClient)

	// Execution history endpoints - authentication only, no special permissions needed
	// Users who can view rules can also view execution history
	app.HandlerFunc(http.MethodGet, version, "/workflow/executions", api.query, authen)
	app.HandlerFunc(http.MethodGet, version, "/workflow/executions/{id}", api.queryByID, authen)
}
