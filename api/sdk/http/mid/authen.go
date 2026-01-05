package mid

import (
	"context"
	"net/http"

	"github.com/timmaaaz/ichor/app/sdk/auth"
	"github.com/timmaaaz/ichor/app/sdk/authclient"
	"github.com/timmaaaz/ichor/app/sdk/mid"
	"github.com/timmaaaz/ichor/business/domain/core/userbus"
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

// BearerQueryParam creates middleware that validates JWT from query param or header.
// This is specifically for WebSocket connections where browsers cannot set custom headers.
// Security Note: Query param tokens appear in server logs - use short-lived tokens.
func BearerQueryParam(ath *auth.Auth) web.MidFunc {
	midFunc := func(ctx context.Context, r *http.Request, next mid.HandlerFunc) mid.Encoder {
		// Try Authorization header first (standard flow)
		token := r.Header.Get("authorization")

		// Fall back to query parameter for WebSocket upgrade requests
		if token == "" {
			if qToken := r.URL.Query().Get("token"); qToken != "" {
				token = "Bearer " + qToken
			}
		}

		return mid.Bearer(ctx, ath, token, next)
	}

	return addMidFunc(midFunc)
}
