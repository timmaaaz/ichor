package mid

import (
	"context"
	"net/http"

	"github.com/timmaaaz/ichor/app/sdk/authclient"
	"github.com/timmaaaz/ichor/app/sdk/mid"
	"github.com/timmaaaz/ichor/business/domain/homebus"
	"github.com/timmaaaz/ichor/business/domain/permissions/permissionsbus"

	"github.com/timmaaaz/ichor/business/domain/users/userbus"
	"github.com/timmaaaz/ichor/foundation/web"
)

// Authorize validates authorization via the auth service with table information.
func Authorize(client *authclient.Client, permissionsBus *permissionsbus.Business, tableName string, action permissionsbus.Action, rule string) web.MidFunc {
	midFunc := func(ctx context.Context, r *http.Request, next mid.HandlerFunc) mid.Encoder {
		// Create table information
		tableInfo := &mid.TableInfo{
			Name:   tableName,
			Action: action,
		}

		// Call the standard Authorize middleware with the enhanced context
		return mid.AuthorizeTable(ctx, client, permissionsBus, tableInfo, rule, next)
	}

	return addMidFunc(midFunc)
}

// AuthorizeUser executes the specified role and extracts the specified
// user from the DB if a user id is specified in the call. Depending on the rule
// specified, the userid from the claims may be compared with the specified
// user id.
func AuthorizeUser(client *authclient.Client, userBus *userbus.Business, rule string) web.MidFunc {
	midFunc := func(ctx context.Context, r *http.Request, next mid.HandlerFunc) mid.Encoder {
		return mid.AuthorizeUser(ctx, client, userBus, rule, web.Param(r, "user_id"), next)
	}

	return addMidFunc(midFunc)
}

// AuthorizeHome executes the specified role and extracts the specified
// home from the DB if a home id is specified in the call. Depending on
// the rule specified, the userid from the claims may be compared with the
// specified user id from the home.
func AuthorizeHome(client *authclient.Client, homeBus *homebus.Business) web.MidFunc {
	midFunc := func(ctx context.Context, r *http.Request, next mid.HandlerFunc) mid.Encoder {
		return mid.AuthorizeHome(ctx, client, homeBus, web.Param(r, "home_id"), next)
	}

	return addMidFunc(midFunc)
}
