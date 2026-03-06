# Security Fix Plan — Phase 1b: Login & Token Issuance

**Source:** `.claude/plans/SECURITY_AUDIT/findings/phase-1b.md` (in the Vue repo)
**Invariants to pass after fix:** INV-011, INV-012
**Branch:** `security/phase-1b-fixes`

---

## Fixes

- [ ] **Fix 1 (HIGH):** Rate limiting on `/api/auth/basic/login` and `/refresh`
- [ ] **Fix 2 (MEDIUM):** bcrypt cost factor 10 → 12
- [ ] **Fix 3 (MEDIUM):** Invert refresh eligibility guard
- [ ] **Fix 4 (LOW):** JWT in UIAdminRedirect URL query param

---

## Fix 1 — Rate Limiting on Auth Endpoints

**Severity:** HIGH
**Risk:** Unlimited login attempts allow brute force / credential stuffing

### Architecture

This codebase uses a two-layer middleware pattern:
- `ichor/app/sdk/mid/` — pure business logic, no HTTP types
- `ichor/api/sdk/http/mid/` — HTTP adapters wrapping the inner layer via `addMidFunc()`
- `golang.org/x/time/rate` is already vendored (`go.mod`: `v0.5.0 // indirect`)

### Step 1 — Create core rate limiter: `ichor/app/sdk/mid/ratelimit.go`

```go
package mid

import (
    "context"
    "net"
    "sync"
    "time"

    "github.com/timmaaaz/ichor/app/sdk/errs"
    "golang.org/x/time/rate"
)

// ipLimiter holds per-IP rate limiter state.
type ipLimiter struct {
    limiter  *rate.Limiter
    lastSeen time.Time
}

// RateLimiter holds a map of per-IP limiters and cleans up stale entries.
type RateLimiter struct {
    mu       sync.Mutex
    limiters map[string]*ipLimiter
    r        rate.Limit
    burst    int
}

// NewRateLimiter creates a per-IP token bucket limiter.
// r = events per second (e.g. rate.Every(12*time.Second) for 5/min)
// burst = max burst size
func NewRateLimiter(r rate.Limit, burst int) *RateLimiter {
    rl := &RateLimiter{
        limiters: make(map[string]*ipLimiter),
        r:        r,
        burst:    burst,
    }
    go rl.cleanupLoop()
    return rl
}

func (rl *RateLimiter) getLimiter(ip string) *rate.Limiter {
    rl.mu.Lock()
    defer rl.mu.Unlock()

    entry, exists := rl.limiters[ip]
    if !exists {
        entry = &ipLimiter{limiter: rate.NewLimiter(rl.r, rl.burst)}
        rl.limiters[ip] = entry
    }
    entry.lastSeen = time.Now()
    return entry.limiter
}

// cleanupLoop removes IPs not seen for 10 minutes to prevent unbounded growth.
func (rl *RateLimiter) cleanupLoop() {
    for {
        time.Sleep(5 * time.Minute)
        rl.mu.Lock()
        for ip, entry := range rl.limiters {
            if time.Since(entry.lastSeen) > 10*time.Minute {
                delete(rl.limiters, ip)
            }
        }
        rl.mu.Unlock()
    }
}

// RateLimit is the core middleware function. Extracts the client IP and
// checks the per-IP token bucket.
func RateLimit(ctx context.Context, remoteAddr string, rl *RateLimiter, next HandlerFunc) Encoder {
    ip, _, err := net.SplitHostPort(remoteAddr)
    if err != nil {
        ip = remoteAddr // fallback if no port
    }

    if !rl.getLimiter(ip).Allow() {
        return errs.Newf(errs.ResourceExhausted, "too many requests")
    }

    return next(ctx)
}
```

**Note:** `errs.ResourceExhausted` must map to HTTP 429. Verify in `ichor/app/sdk/errs/` — if it doesn't exist, add it or use the closest match and update the errors.go status code map.

### Step 2 — Create HTTP adapter: `ichor/api/sdk/http/mid/ratelimit.go`

```go
package mid

import (
    "context"
    "net/http"

    appmid "github.com/timmaaaz/ichor/app/sdk/mid"
    "github.com/timmaaaz/ichor/foundation/web"
)

// RateLimit wraps the core rate limiter as a per-route web.MidFunc.
func RateLimit(rl *appmid.RateLimiter) web.MidFunc {
    midFunc := func(ctx context.Context, r *http.Request, next appmid.HandlerFunc) appmid.Encoder {
        return appmid.RateLimit(ctx, r.RemoteAddr, rl, next)
    }
    return addMidFunc(midFunc)
}
```

### Step 3 — Check errs package for 429 mapping

File: `ichor/app/sdk/errs/errs.go` (or similar)

- Verify `ResourceExhausted` exists and maps to `http.StatusTooManyRequests` (429)
- If missing, add the constant and the mapping entry
- Also verify `ichor/app/sdk/mid/errors.go` maps it correctly

### Step 4 — Wire into basicauth routes: `ichor/api/domain/http/basicauthapi/route.go`

```go
import (
    "golang.org/x/time/rate"
    appmid "github.com/timmaaaz/ichor/app/sdk/mid"
    httpmid "github.com/timmaaaz/ichor/api/sdk/http/mid"
)

func Routes(app *web.App, cfg Config) {
    api := NewAPI(cfg)

    // 5 requests per minute per IP, burst of 3
    loginLimiter := appmid.NewRateLimiter(rate.Every(12*time.Second), 3)
    // 10 requests per minute per IP, burst of 5
    refreshLimiter := appmid.NewRateLimiter(rate.Every(6*time.Second), 5)

    app.HandlerFunc(http.MethodPost, "", "/api/auth/basic/login", api.login,
        httpmid.RateLimit(loginLimiter))
    app.HandlerFunc(http.MethodPost, "", "/api/auth/basic/refresh", api.refresh,
        httpmid.RateLimit(refreshLimiter))
    app.HandlerFunc(http.MethodPost, "", "/api/auth/basic/logout", api.logout)
}
```

### Step 5 — Add `Retry-After` header on 429

In `ichor/api/sdk/http/mid/ratelimit.go`, set the header before returning:
```go
// In the midFunc, after Allow() returns false:
w := web.GetWriter(ctx) // or however the response writer is accessed
w.Header().Set("Retry-After", "12")
```

Check `web.GetWriter` API in `ichor/foundation/web/` — may need a different approach depending on context key access.

---

## Fix 2 — bcrypt Cost Factor

**Severity:** MEDIUM
**Files:** 2 changes

### Change 1: `ichor/business/domain/core/userbus/userbus.go`

Find all `bcrypt.GenerateFromPassword([]byte(...), bcrypt.DefaultCost)` calls.

Replace with:
```go
const bcryptCost = 12

// Then in the function:
hash, err := bcrypt.GenerateFromPassword([]byte(nu.Password), bcryptCost)
```

Apply to all occurrences (user create + user update password paths).

### Change 2: `ichor/api/domain/http/basicauthapi/basicauthapi.go`

The `HashPassword` helper at line 238:
```go
// Before:
hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

// After:
const bcryptCost = 12
hash, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
```

**Note on existing passwords:** Existing bcrypt hashes store the cost in the hash itself. On next login, `bcrypt.CompareHashAndPassword` will succeed (comparison still works). To silently upgrade old hashes to cost 12 on login:

```go
// In login handler, after successful CompareHashAndPassword:
if cost, err := bcrypt.Cost([]byte(user.PasswordHash)); err == nil && cost < bcryptCost {
    if newHash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcryptCost); err == nil {
        _ = a.userBus.UpdatePassword(ctx, user.ID, string(newHash)) // best-effort
    }
}
```

This is optional but recommended. Only implement if `userBus.UpdatePassword` (or equivalent) already exists.

---

## Fix 3 — Refresh Eligibility Guard

**Severity:** MEDIUM
**File:** `ichor/api/domain/http/basicauthapi/basicauthapi.go:162`

Current (broken — window 30m > token life 20m, always allows refresh):
```go
if timeUntilExpiry > 30*time.Minute {
    return errs.Newf(errs.FailedPrecondition, "token not eligible for refresh yet")
}
```

Fix — only allow refresh in the last 5 minutes of token life:
```go
const refreshWindow = 5 * time.Minute

if timeUntilExpiry > refreshWindow {
    return errs.Newf(errs.FailedPrecondition, "token not eligible for refresh yet")
}
```

This means:
- Token issued, 20 min valid
- Refresh blocked for first 15 minutes
- Refresh allowed in last 5 minutes before expiry
- Client should detect `sessionExpired` / 401 and trigger refresh proactively

**Frontend implication:** The Vue `authStore` auto-refresh triggers 30 seconds before expiry. This is compatible with a 5-minute window.

---

## Fix 4 — JWT in Admin Redirect URL

**Severity:** LOW
**File:** `ichor/api/cmd/services/ichor/main.go`

Current:
```go
UIAdminRedirect string `conf:"default:http://localhost:3001/admin?token="`
```

Fix — use URL fragment (never sent to server):
```go
UIAdminRedirect string `conf:"default:http://localhost:3001/admin#token="`
```

**Frontend implication:** The admin app must read the token from `window.location.hash` instead of `window.location.search`. Verify where this redirect is consumed in the admin frontend before merging.

Also check: `UILoginRedirect string \`conf:"default:http://localhost:3001/login"\`` — confirm this does NOT append the token as a query param anywhere in the OAuth handler.

---

## Verification Steps

After implementing all fixes:

```bash
# Build check
cd ichor && go build ./...

# Test suite
go test ./api/domain/http/basicauthapi/... -v
go test ./app/sdk/mid/... -v
go test ./foundation/keystore/... -v

# Manual verify rate limiting (should 429 after burst):
for i in $(seq 1 6); do
    curl -s -o /dev/null -w "%{http_code}\n" \
    -X POST http://localhost:8080/api/auth/basic/login \
    -H "Content-Type: application/json" \
    -d '{"email":"test@test.com","password":"wrong"}'
done
# Expected: 401 401 401 [first burst] then 429 429 429
```

---

## Invariant Verification (post-fix)

Run in the Vue repo after merging:
```
/security-audit-phase regression
```

Expected:
- INV-011: PASS (bcrypt cost >= 12)
- INV-012: PASS (refresh window 5m < token life 20m)

---

## Files Modified

| File | Change |
|---|---|
| `ichor/app/sdk/mid/ratelimit.go` | **CREATE** — core per-IP token bucket |
| `ichor/api/sdk/http/mid/ratelimit.go` | **CREATE** — HTTP adapter |
| `ichor/app/sdk/errs/` | **MODIFY** (if needed) — add ResourceExhausted → 429 |
| `ichor/api/domain/http/basicauthapi/route.go` | **MODIFY** — wire rate limiters |
| `ichor/business/domain/core/userbus/userbus.go` | **MODIFY** — bcrypt cost 12 |
| `ichor/api/domain/http/basicauthapi/basicauthapi.go` | **MODIFY** — bcrypt cost 12 + refresh window |
| `ichor/api/cmd/services/ichor/main.go` | **MODIFY** — `#token=` fragment |
