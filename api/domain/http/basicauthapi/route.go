package basicauthapi

import (
	"math"
	"net/http"
	"time"

	"github.com/jmoiron/sqlx"
	httpmid "github.com/timmaaaz/ichor/api/sdk/http/mid"
	"github.com/timmaaaz/ichor/app/sdk/auth"
	"github.com/timmaaaz/ichor/business/domain/core/userbus"
	"github.com/timmaaaz/ichor/foundation/logger"
	"github.com/timmaaaz/ichor/foundation/web"
	"golang.org/x/time/rate"
)

// RateLimitConfig holds rate limiting parameters for the basic auth endpoints.
// All fields are configurable via environment variables (prefix: ICHOR_RATELIMIT_).
//
// Shop-floor tuning guide: if many workers share a single NAT IP (e.g. factory
// WiFi), lower LoginInterval and raise LoginBurst so a full shift can log in
// simultaneously without hitting 429. Example for 30 workers:
//
//	ICHOR_RATELIMIT_LOGININTERVAL=2s
//	ICHOR_RATELIMIT_LOGINBURST=30
type RateLimitConfig struct {
	LoginInterval     time.Duration
	LoginBurst        int
	RefreshInterval   time.Duration
	RefreshBurst      int
	TrustedProxyCIDRs string // see httpmid.TrustedProxyExtractor; empty = use RemoteAddr
}

// Config contains all the mandatory systems required by handlers.
type Config struct {
	Log             *logger.Logger
	Auth            *auth.Auth
	DB              *sqlx.DB
	TokenKey        string
	TokenExpiration time.Duration
	UserBus         *userbus.Business
	Blocklist       *auth.Blocklist
	RateLimit       RateLimitConfig
}

// Routes registers the basic auth HTTP routes and returns a cleanup function
// that stops the background rate limiter goroutines. The caller must call the
// returned function during graceful shutdown to avoid goroutine leaks.
func Routes(app *web.App, cfg Config) func() {
	api := NewAPI(cfg)

	rl := cfg.RateLimit

	loginLimiter := httpmid.NewRateLimiter(rate.Every(rl.LoginInterval), rl.LoginBurst)
	refreshLimiter := httpmid.NewRateLimiter(rate.Every(rl.RefreshInterval), rl.RefreshBurst)

	// Retry-After is the interval ceiling in seconds — how long until the next
	// token is available. Derived from config so it stays in sync automatically.
	loginRetry := int(math.Ceil(rl.LoginInterval.Seconds()))
	refreshRetry := int(math.Ceil(rl.RefreshInterval.Seconds()))

	// IP extractor: RemoteAddrExtractor by default (direct deployment).
	// Automatically upgrades to XFF-aware extraction when TrustedProxyCIDRs
	// is set — e.g. when nginx or a cloud load balancer is placed in front.
	extract := httpmid.IPExtractor(httpmid.RemoteAddrExtractor)
	if rl.TrustedProxyCIDRs != "" {
		extract = httpmid.TrustedProxyExtractor(httpmid.ParseTrustedCIDRs(rl.TrustedProxyCIDRs))
	}

	app.HandlerFunc(http.MethodPost, "", "/api/auth/basic/login", api.login,
		httpmid.RateLimit(loginLimiter, extract, loginRetry))
	app.HandlerFunc(http.MethodPost, "", "/api/auth/basic/refresh", api.refresh,
		httpmid.RateLimit(refreshLimiter, extract, refreshRetry))
	app.HandlerFunc(http.MethodPost, "", "/api/auth/basic/logout", api.logout)

	return func() {
		loginLimiter.Stop()
		refreshLimiter.Stop()
	}
}
