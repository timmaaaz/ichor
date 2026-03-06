package mid

import (
	"context"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"golang.org/x/time/rate"

	"github.com/timmaaaz/ichor/app/sdk/errs"
	appmid "github.com/timmaaaz/ichor/app/sdk/mid"
	"github.com/timmaaaz/ichor/foundation/web"
)

// =============================================================================
// IP Extraction
// =============================================================================

// IPExtractor extracts the client IP from an HTTP request for rate limiting.
//
// Two implementations are provided:
//   - RemoteAddrExtractor: use when the service is accessed directly (no reverse proxy).
//   - TrustedProxyExtractor: use when behind a verified reverse proxy chain.
//
// Do NOT switch to TrustedProxyExtractor without also configuring the proxy to
// overwrite X-Forwarded-For on inbound requests. Using XFF without a verified
// proxy chain allows clients to spoof their IP and bypass rate limits entirely.
type IPExtractor func(r *http.Request) string

// RemoteAddrExtractor extracts the client IP from r.RemoteAddr.
//
// This is the correct extractor when the service is accessed directly, without
// a reverse proxy. r.RemoteAddr is the actual TCP connection address and cannot
// be spoofed by the client.
var RemoteAddrExtractor IPExtractor = func(r *http.Request) string {
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return ip
}

// TrustedProxyExtractor returns an IPExtractor that reads X-Forwarded-For and
// returns the rightmost IP that does not belong to any of the trusted proxy CIDRs.
//
// Security model (rightmost-untrusted algorithm):
//  1. If RemoteAddr is not itself a trusted proxy, XFF cannot be trusted — fall
//     back to RemoteAddr. This prevents a direct client from faking XFF headers.
//  2. Collect all X-Forwarded-For IPs and scan right-to-left.
//  3. Skip IPs that match a trusted CIDR (these are your proxies).
//  4. Return the first untrusted IP — this is the real client.
//  5. If XFF is absent or all IPs are trusted proxies, return RemoteAddr.
//
// NEVER use the leftmost XFF IP for security decisions — it is trivially
// spoofable and has been the root cause of multiple rate-limit bypass CVEs.
//
// Usage: wire this in route.go when a reverse proxy (nginx/traefik/ALB) sits
// in front of the service and is configured to overwrite XFF from clients.
//
//	cidrs := httpmid.ParseTrustedCIDRs(cfg.TrustedProxyCIDRs) // e.g. "10.0.0.0/8"
//	httpmid.RateLimit(limiter, httpmid.TrustedProxyExtractor(cidrs), 12)
func TrustedProxyExtractor(trustedCIDRs []*net.IPNet) IPExtractor {
	return func(r *http.Request) string {
		remoteIP, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			remoteIP = r.RemoteAddr
		}

		// If the direct connection is not from a trusted proxy, XFF cannot
		// be trusted — a client could have forged it.
		if !isTrustedIP(remoteIP, trustedCIDRs) {
			return remoteIP
		}

		// Collect all XFF IPs into a flat slice. Multiple XFF headers and
		// comma-separated values are both valid per RFC 7239.
		var ips []string
		for _, h := range r.Header.Values("X-Forwarded-For") {
			for _, part := range strings.Split(h, ",") {
				ip := strings.TrimSpace(part)
				if ip != "" {
					ips = append(ips, ip)
				}
			}
		}

		// Scan right-to-left: each trusted proxy appends to the right, so the
		// rightmost untrusted IP is the actual client. Validate with net.ParseIP
		// before returning — an injected non-IP string would otherwise become an
		// arbitrary rate-limit bucket key, bypassing per-IP limiting entirely.
		for i := len(ips) - 1; i >= 0; i-- {
			if !isTrustedIP(ips[i], trustedCIDRs) {
				if net.ParseIP(ips[i]) != nil {
					return ips[i]
				}
				return remoteIP
			}
		}

		// All XFF IPs were trusted proxies — fall back to RemoteAddr.
		return remoteIP
	}
}

// ParseTrustedCIDRs parses a comma-separated list of CIDR strings into
// *net.IPNet values suitable for passing to TrustedProxyExtractor.
// Invalid or empty entries are silently skipped.
//
// Example: "10.0.0.0/8,172.16.0.0/12,192.168.0.0/16"
func ParseTrustedCIDRs(cidrList string) []*net.IPNet {
	if cidrList == "" {
		return nil
	}
	var nets []*net.IPNet
	for _, s := range strings.Split(cidrList, ",") {
		s = strings.TrimSpace(s)
		if s == "" {
			continue
		}
		_, ipNet, err := net.ParseCIDR(s)
		if err == nil {
			nets = append(nets, ipNet)
		}
	}
	return nets
}

func isTrustedIP(ipStr string, cidrs []*net.IPNet) bool {
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return false
	}
	for _, cidr := range cidrs {
		if cidr.Contains(ip) {
			return true
		}
	}
	return false
}

// =============================================================================
// RateLimiter
// =============================================================================

type ipLimiter struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

// RateLimiter manages per-IP token bucket rate limiters. It is safe for
// concurrent use. A background goroutine purges stale entries every 5 minutes
// to prevent unbounded memory growth. Call Stop to terminate the goroutine on
// graceful shutdown. Stop is safe to call multiple times.
type RateLimiter struct {
	mu       sync.Mutex
	limiters map[string]*ipLimiter
	r        rate.Limit
	burst    int
	done     chan struct{}
	once     sync.Once
}

// NewRateLimiter creates a per-IP token bucket rate limiter.
//
//	r     — refill rate, e.g. rate.Every(12*time.Second) ≈ 5 req/min
//	burst — max burst (tokens available immediately at start or after idle)
//
// Call Stop to terminate the background cleanup goroutine on shutdown.
func NewRateLimiter(r rate.Limit, burst int) *RateLimiter {
	rl := &RateLimiter{
		limiters: make(map[string]*ipLimiter),
		r:        r,
		burst:    burst,
		done:     make(chan struct{}),
	}
	go rl.cleanupLoop()
	return rl
}

// Allow reports whether the given IP is within its rate limit, consuming one
// token if it is. It is safe for concurrent use.
func (rl *RateLimiter) Allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	entry, exists := rl.limiters[ip]
	if !exists {
		entry = &ipLimiter{limiter: rate.NewLimiter(rl.r, rl.burst)}
		rl.limiters[ip] = entry
	}
	entry.lastSeen = time.Now()
	return entry.limiter.Allow()
}

// Stop terminates the background cleanup goroutine. Call this during graceful
// shutdown to avoid goroutine leaks in tests and short-lived processes.
func (rl *RateLimiter) Stop() {
	rl.once.Do(func() { close(rl.done) })
}

// cleanupLoop removes IPs not seen for 10 minutes, running every 5 minutes.
func (rl *RateLimiter) cleanupLoop() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			rl.mu.Lock()
			for ip, entry := range rl.limiters {
				if time.Since(entry.lastSeen) > 10*time.Minute {
					delete(rl.limiters, ip)
				}
			}
			rl.mu.Unlock()
		case <-rl.done:
			return
		}
	}
}

// =============================================================================
// RateLimit middleware
// =============================================================================

// RateLimit returns a per-IP rate limiting middleware.
//
//	rl              — rate limiter instance; create one per endpoint.
//	extract         — IP extraction strategy; use RemoteAddrExtractor
//	                  for direct deployments, TrustedProxyExtractor for
//	                  deployments behind a verified reverse proxy.
//	retryAfterSecs  — value written to the Retry-After header on 429 responses;
//	                  set to ceil(1/rate_per_second), e.g. 12 for rate.Every(12s).
func RateLimit(rl *RateLimiter, extract IPExtractor, retryAfterSecs int) web.MidFunc {
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
