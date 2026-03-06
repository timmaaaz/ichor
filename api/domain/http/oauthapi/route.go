// api/domain/http/oauthapi/route.go
package oauthapi

import (
	"math"
	"net/http"
	"time"

	"github.com/timmaaaz/ichor/app/sdk/auth"
	appmid "github.com/timmaaaz/ichor/app/sdk/mid"
	"github.com/timmaaaz/ichor/business/domain/core/userbus"
	"github.com/timmaaaz/ichor/foundation/logger"
	"github.com/timmaaaz/ichor/foundation/web"
	httpmid "github.com/timmaaaz/ichor/api/sdk/http/mid"
	"golang.org/x/time/rate"
)

// Config contains all the configuration for the OAuth app.
type Config struct {
	Auth            *auth.Auth
	Log             *logger.Logger
	TokenKey        string
	GoogleKey       string
	GoogleSecret    string
	Callback        string
	StoreKey        string
	UIAdminRedirect string
	UILoginRedirect string
	Environment     string
	EnableDevAuth   bool
	TokenExpiration time.Duration
	UserBus         *userbus.Business // Fix 3: DB role lookup in authCallback
	Blocklist       *auth.Blocklist   // Fix 7: JWT revocation on logout
}

// Routes adds the OAuth routes to the web.App.
func Routes(app *web.App, cfg Config) {
	api := newAPI(cfg)

	// Fix 8: Rate limit OAuth initiation and callback endpoints.
	// 5 requests per minute per IP (burst 3) — mirrors the login limiter.
	const oauthInterval = 12 * time.Second
	oauthLimiter := appmid.NewRateLimiter(rate.Every(oauthInterval), 3)
	retryAfter := int(math.Ceil(oauthInterval.Seconds()))

	app.RawHandlerFunc(http.MethodGet, "", "/api/auth/{provider}", api.authenticate,
		httpmid.RateLimit(oauthLimiter, appmid.RemoteAddrExtractor, retryAfter))
	app.RawHandlerFunc(http.MethodGet, "", "/api/auth/{provider}/callback", api.authCallback,
		httpmid.RateLimit(oauthLimiter, appmid.RemoteAddrExtractor, retryAfter))
	app.RawHandlerFunc(http.MethodGet, "", "/api/logout/{provider}", api.logout)
}
