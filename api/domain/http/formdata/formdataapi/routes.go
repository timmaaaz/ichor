package formdataapi

import (
	"net/http"

	"github.com/timmaaaz/ichor/api/sdk/http/mid"
	"github.com/timmaaaz/ichor/app/domain/formdata/formdataapp"
	"github.com/timmaaaz/ichor/app/sdk/auth"
	"github.com/timmaaaz/ichor/app/sdk/authclient"
	"github.com/timmaaaz/ichor/business/domain/core/permissionsbus"
	"github.com/timmaaaz/ichor/foundation/web"
)

// Config contains dependencies needed for formdata routes.
type Config struct {
	FormdataApp    *formdataapp.App
	AuthClient     *authclient.Client
	PermissionsBus *permissionsbus.Business
}

// Routes binds formdata endpoints to the web application.
//
// Endpoints:
//
//	POST /v1/formdata/:form_id/upsert - Multi-entity transactional upsert
//	POST /v1/formdata/:form_id/validate - Validate form configuration
//
// All endpoints require authentication. Authorization rules can be
// customized based on business requirements.
func Routes(app *web.App, cfg Config) {
	const version = "v1"
	const tableName = "config.formdata"

	authen := mid.Authenticate(cfg.AuthClient)

	api := newAPI(cfg.FormdataApp)

	// POST /v1/formdata/:form_id/upsert
	// Handles dynamic multi-entity create/update operations
	app.HandlerFunc(http.MethodPost, version, "/formdata/{form_id}/upsert", api.upsert, authen,
		mid.Authorize(cfg.AuthClient, cfg.PermissionsBus, tableName, permissionsbus.Actions.Create, auth.RuleAny))

	// POST /v1/formdata/:form_id/validate
	// Validates that a form has all required fields for specified operations
	app.HandlerFunc(http.MethodPost, version, "/formdata/{form_id}/validate", api.validate, authen,
		mid.Authorize(cfg.AuthClient, cfg.PermissionsBus, tableName, permissionsbus.Actions.Read, auth.RuleAny))

}
