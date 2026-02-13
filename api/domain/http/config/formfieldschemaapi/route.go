package formfieldschemaapi

import (
	"net/http"

	"github.com/timmaaaz/ichor/api/sdk/http/mid"
	"github.com/timmaaaz/ichor/app/sdk/authclient"
	"github.com/timmaaaz/ichor/foundation/web"
)

// Config holds the dependencies for the form field schema API routes.
type Config struct {
	AuthClient *authclient.Client
}

// Routes registers the form field schema API routes.
func Routes(app *web.App, cfg Config) {
	const version = "v1"

	api := newAPI()
	authen := mid.Authenticate(cfg.AuthClient)

	app.HandlerFunc(http.MethodGet, version, "/config/form-field-types", api.queryFieldTypes, authen)
	app.HandlerFunc(http.MethodGet, version, "/config/form-field-types/{type}/schema", api.queryFieldTypeSchema, authen)
}
