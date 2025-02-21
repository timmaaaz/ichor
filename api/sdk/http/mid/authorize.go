package mid

import (
	"context"
	"net/http"

	"github.com/timmaaaz/ichor/app/sdk/authclient"
	"github.com/timmaaaz/ichor/app/sdk/mid"
	"github.com/timmaaaz/ichor/business/domain/homebus"
	"github.com/timmaaaz/ichor/business/domain/productbus"
	"github.com/timmaaaz/ichor/business/domain/users/userbus"
	"github.com/timmaaaz/ichor/foundation/web"
)

// Authorize validates authorization via the auth service.
func Authorize(client *authclient.Client, rule string) web.MidFunc {
	midFunc := func(ctx context.Context, r *http.Request, next mid.HandlerFunc) mid.Encoder {
		return mid.Authorize(ctx, client, rule, next)
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

func AuthorizeCheckPermissions(client *authclient.Client, userBus *userbus.Business, rule string) web.MidFunc {
	midFunc := func(ctx context.Context, r *http.Request, next mid.HandlerFunc) mid.Encoder {
		return mid.AuthorizeCheckPermissions(ctx, client, userBus, web.Param(r, "user_id"), next)
	}

	return addMidFunc(midFunc)
}

// AuthorizeProduct executes the specified role and extracts the specified
// product from the DB if a product id is specified in the call. Depending on
// the rule specified, the userid from the claims may be compared with the
// specified user id from the product.
func AuthorizeProduct(client *authclient.Client, productBus *productbus.Business) web.MidFunc {
	midFunc := func(ctx context.Context, r *http.Request, next mid.HandlerFunc) mid.Encoder {
		return mid.AuthorizeProduct(ctx, client, productBus, web.Param(r, "product_id"), next)
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
