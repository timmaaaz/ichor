package mid

import (
	"context"
	"net/http"

	"github.com/timmaaaz/ichor/app/sdk/auth"
	"github.com/timmaaaz/ichor/app/sdk/authclient"
	"github.com/timmaaaz/ichor/app/sdk/mid"
	"github.com/timmaaaz/ichor/business/domain/users/userbus"
	"github.com/timmaaaz/ichor/foundation/web"
)

// Authenticate validates authentication via the auth service.
func Authenticate(client *authclient.Client) web.MidFunc {
	midFunc := func(ctx context.Context, r *http.Request, next mid.HandlerFunc) mid.Encoder {
		return mid.Authenticate(ctx, client, r.Header.Get("authorization"), next)
	}

	return addMidFunc(midFunc)
}

// Bearer processes JWT authentication logic.
func Bearer(ath *auth.Auth) web.MidFunc {
	midFunc := func(ctx context.Context, r *http.Request, next mid.HandlerFunc) mid.Encoder {
		return mid.Bearer(ctx, ath, r.Header.Get("authorization"), next)
	}

	return addMidFunc(midFunc)
}

// Basic processes basic authentication logic.
func Basic(userBus *userbus.Business, ath *auth.Auth) web.MidFunc {
	midFunc := func(ctx context.Context, r *http.Request, next mid.HandlerFunc) mid.Encoder {
		return mid.Basic(ctx, ath, userBus, r.Header.Get("authorization"), next)
	}

	return addMidFunc(midFunc)
}
