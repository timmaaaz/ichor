// api/domain/http/oauthapi/route.go
package oauthapi

import (
	"math"
	"net/http"
	"time"

	httpmid "github.com/timmaaaz/ichor/api/sdk/http/mid"
	"github.com/timmaaaz/ichor/app/sdk/auth"
	"github.com/timmaaaz/ichor/business/domain/core/userbus"
	"github.com/timmaaaz/ichor/foundation/logger"
	"github.com/timmaaaz/ichor/foundation/web"
	"golang.org/x/time/rate"
)

// Config contains all the configuration for the OAuth app.
type Config struct {
	Auth              *auth.Auth
	Log               *logger.Logger
	TokenKey          string
	GoogleKey         string
	GoogleSecret      string
	Callback          string
	StoreKey          string
	UIAdminRedirect   string
	UILoginRedirect   string
	Environment       string
	EnableDevAuth     bool
	TokenExpiration   time.Duration
	UserBus           *userbus.Business // DB role lookup in authCallback
	Blocklist         *auth.Blocklist   // JWT revocation on logout
	TrustedProxyCIDRs string            // see httpmid.TrustedProxyExtractor; empty = use RemoteAddr
}

// Routes adds the OAuth routes to the web.App and returns a cleanup function
// that stops the background rate limiter goroutines. The caller must call the
// returned function during graceful shutdown to avoid goroutine leaks.
func Routes(app *web.App, cfg Config) func() {
	api := newAPI(cfg)

	// IP extractor: RemoteAddrExtractor by default (direct deployment).
	// Automatically upgrades to XFF-aware extraction when TrustedProxyCIDRs
	// is set — e.g. when nginx or a cloud load balancer is placed in front.
	extract := httpmid.IPExtractor(httpmid.RemoteAddrExtractor)
	if cfg.TrustedProxyCIDRs != "" {
		extract = httpmid.TrustedProxyExtractor(httpmid.ParseTrustedCIDRs(cfg.TrustedProxyCIDRs))
	}

	// Separate rate limiters for initiate vs callback:
	// - Initiate: user-driven, strict limit (5/min per IP, burst 3).
	// - Callback: IdP-driven redirect; user cannot retry directly so a higher
	//   burst is appropriate (burst 5, ~10/min) to absorb provider retries.
	const initiateInterval = 12 * time.Second
	const callbackInterval = 6 * time.Second

	initLimiter := httpmid.NewRateLimiter(rate.Every(initiateInterval), 3)
	callbackLimiter := httpmid.NewRateLimiter(rate.Every(callbackInterval), 5)

	initiateRetry := int(math.Ceil(initiateInterval.Seconds()))
	callbackRetry := int(math.Ceil(callbackInterval.Seconds()))

	app.RawHandlerFunc(http.MethodGet, "", "/api/auth/{provider}", api.authenticate,
		httpmid.RateLimit(initLimiter, extract, initiateRetry))
	app.RawHandlerFunc(http.MethodGet, "", "/api/auth/{provider}/callback", api.authCallback,
		httpmid.RateLimit(callbackLimiter, extract, callbackRetry))
	app.RawHandlerFunc(http.MethodGet, "", "/api/logout/{provider}", api.logout)

	return func() {
		initLimiter.Stop()
		callbackLimiter.Stop()
	}
}
