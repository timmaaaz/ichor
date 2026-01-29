package actionapi

import (
	"net/http"

	"github.com/timmaaaz/ichor/api/sdk/http/mid"
	"github.com/timmaaaz/ichor/app/domain/workflow/actionapp"
	"github.com/timmaaaz/ichor/app/sdk/authclient"
	"github.com/timmaaaz/ichor/business/domain/core/userrolebus"
	"github.com/timmaaaz/ichor/business/domain/workflow/actionpermissionsbus"
	"github.com/timmaaaz/ichor/business/sdk/workflow"
	"github.com/timmaaaz/ichor/foundation/logger"
	"github.com/timmaaaz/ichor/foundation/web"
)

// Config holds the dependencies for the action API routes.
type Config struct {
	Log           *logger.Logger
	ActionService *workflow.ActionService
	ActionPermBus *actionpermissionsbus.Business
	UserRoleBus   *userrolebus.Business
	AuthClient    *authclient.Client
}

// Routes registers the manual action execution API routes.
func Routes(app *web.App, cfg Config) {
	const version = "v1"

	actionApp := actionapp.NewApp(cfg.ActionService, cfg.ActionPermBus)
	api := newAPI(actionApp, cfg.UserRoleBus)
	authen := mid.Authenticate(cfg.AuthClient)

	// List available actions for the authenticated user (filtered by permissions)
	app.HandlerFunc(http.MethodGet, version, "/workflow/actions", api.list, authen)

	// Execute an action manually
	// Permission checking is done in the app layer based on action_permissions table
	app.HandlerFunc(http.MethodPost, version, "/workflow/actions/{actionType}/execute", api.execute, authen)

	// Get execution status (for tracking async actions)
	app.HandlerFunc(http.MethodGet, version, "/workflow/executions/{executionId}", api.getExecutionStatus, authen)
}
