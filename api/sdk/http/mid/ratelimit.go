package mid

import (
	"context"
	"net/http"
	"strconv"

	"github.com/timmaaaz/ichor/app/sdk/errs"
	appmid "github.com/timmaaaz/ichor/app/sdk/mid"
	"github.com/timmaaaz/ichor/foundation/web"
)

// RateLimit returns a per-IP rate limiting middleware.
//
//	rl              — rate limiter instance; create one per endpoint.
//	extract         — IP extraction strategy; use appmid.RemoteAddrExtractor
//	                  for direct deployments, appmid.TrustedProxyExtractor for
//	                  deployments behind a verified reverse proxy.
//	retryAfterSecs  — value written to the Retry-After header on 429 responses;
//	                  set to ceil(1/rate_per_second), e.g. 12 for rate.Every(12s).
func RateLimit(rl *appmid.RateLimiter, extract appmid.IPExtractor, retryAfterSecs int) web.MidFunc {
	retryAfter := strconv.Itoa(retryAfterSecs)

	midFunc := func(ctx context.Context, r *http.Request, next appmid.HandlerFunc) appmid.Encoder {
		ip := extract(r)
		if !rl.Allow(ip) {
			if w := web.GetWriter(ctx); w != nil {
				w.Header().Set("Retry-After", retryAfter)
			}
			return errs.Newf(errs.ResourceExhausted, "too many requests")
		}
		return next(ctx)
	}

	return addMidFunc(midFunc)
}
