package physicalattributeapi

import (
	"net/http"

	"github.com/timmaaaz/ichor/api/sdk/http/mid"
	"github.com/timmaaaz/ichor/app/domain/inventory/core/physicalattributeapp"
	"github.com/timmaaaz/ichor/app/sdk/auth"
	"github.com/timmaaaz/ichor/app/sdk/authclient"
	"github.com/timmaaaz/ichor/business/domain/inventory/core/physicalattributebus"
	"github.com/timmaaaz/ichor/business/domain/permissions/permissionsbus"
	"github.com/timmaaaz/ichor/foundation/logger"
	"github.com/timmaaaz/ichor/foundation/web"
)

type Config struct {
	Log                  *logger.Logger
	PhysicalAttributeBus *physicalattributebus.Business
	AuthClient           *authclient.Client
	PermissionsBus       *permissionsbus.Business
}

const TableName = "physical_attributes"

func Routes(app *web.App, cfg Config) {
	const version = "v1"

	authen := mid.Authenticate(cfg.AuthClient)

	api := newAPI(physicalattributeapp.NewApp(cfg.PhysicalAttributeBus))
	app.HandlerFunc(http.MethodGet, version, "/inventory/core/physical-attributes", api.query, authen,
		mid.Authorize(cfg.AuthClient, cfg.PermissionsBus, TableName, permissionsbus.Actions.Read, auth.RuleAny))

	app.HandlerFunc(http.MethodGet, version, "/inventory/core/physical-attributes/{attribute_id}", api.queryByID, authen,
		mid.Authorize(cfg.AuthClient, cfg.PermissionsBus, TableName, permissionsbus.Actions.Read, auth.RuleAny))

	app.HandlerFunc(http.MethodPost, version, "/inventory/core/physical-attributes", api.create, authen,
		mid.Authorize(cfg.AuthClient, cfg.PermissionsBus, TableName, permissionsbus.Actions.Create, auth.RuleAny))

	app.HandlerFunc(http.MethodPut, version, "/inventory/core/physical-attributes/{attribute_id}", api.update, authen,
		mid.Authorize(cfg.AuthClient, cfg.PermissionsBus, TableName, permissionsbus.Actions.Update, auth.RuleAny))

	app.HandlerFunc(http.MethodDelete, version, "/inventory/core/physical-attributes/{attribute_id}", api.delete, authen,
		mid.Authorize(cfg.AuthClient, cfg.PermissionsBus, TableName, permissionsbus.Actions.Delete, auth.RuleAny))

}
