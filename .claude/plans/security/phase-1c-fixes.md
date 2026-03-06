# Security Fix Plan — Phase 1c: Session Management & OAuth

**Source:** `.claude/plans/SECURITY_AUDIT/findings/phase-1c.md` (in the Vue repo)
**Invariants to pass after fix:** INV-006, INV-016, INV-017
**Branch:** `security/phase-1c-fixes`

---

## Fixes

- [ ] **Fix 1 (CRITICAL):** Wire `Environment` field into `oauthCfg` in `main.go` — eliminates dev provider in production
- [ ] **Fix 2 (CRITICAL/related):** Add startup assertion — panic if dev provider would register in production
- [ ] **Fix 3 (HIGH):** Replace hardcoded Admin role with DB user lookup in `authCallback()`
- [ ] **Fix 4 (HIGH):** Replace JWT query-param redirect with URL fragment (`?token=` → `#token=`)
- [ ] **Fix 5 (MEDIUM):** Session cookie `Secure: true` and `SameSite: Lax`
- [ ] **Fix 6 (MEDIUM):** Reduce session cookie `MaxAge` from 30 days to 15 minutes
- [ ] **Fix 7 (MEDIUM):** JWT blocklist on OAuth logout
- [ ] **Fix 8 (LOW):** Rate limiting on OAuth endpoints
- [ ] **Fix 9 (LOW):** Validate provider name against allowlist

---

## Fix 1 — Wire Environment Field (CRITICAL, one line)

**Severity:** CRITICAL
**Risk:** Dev provider always registered in all environments including production — unauthenticated Admin JWT issuance
**File:** `ichor/api/cmd/services/ichor/main.go`

Find the `oauthCfg` struct literal (search for `oauthapi.Config{`). Add the `Environment` field:

```go
oauthCfg := oauthapi.Config{
    Auth:            oauthAuth,
    Log:             log,
    TokenKey:        cfg.OAuth.TokenKey,
    StoreKey:        cfg.OAuth.StoreKey,
    UIAdminRedirect: cfg.OAuth.UIAdminRedirect,
    UILoginRedirect: cfg.OAuth.UILoginRedirect,
    Environment:     cfg.OAuth.Environment,  // ADD THIS LINE
}
```

Also verify `cfg.OAuth` has an `Environment` field. Check the OAuth config struct definition in `main.go`. If it's missing:

```go
// In the cfg struct OAuth section:
Environment string `conf:"default:development"`
```

Using `default:development` means you must explicitly set `ICHOR_OAUTH_ENVIRONMENT=production` in prod — safer than a silent default.

---

## Fix 2 — Startup Assertion for Dev Provider in Production (CRITICAL defense-in-depth)

**Severity:** CRITICAL (defense-in-depth for Fix 1)
**File:** `ichor/api/domain/http/oauthapi/oauthapi.go`

Add a startup-time panic guard in `newAPI()`:

```go
func newAPI(cfg Config) *api {
    // Safety assertion: dev provider must never reach production
    if cfg.Environment == "production" {
        // This branch should only register Google — dev provider must not be present
        goth.UseProviders(
            google.New(cfg.GoogleKey, cfg.GoogleSecret, cfg.Callback),
        )
    } else {
        providers := []goth.Provider{
            NewDevelopmentProvider(cfg.Callback),
        }
        if cfg.GoogleKey != "" && cfg.GoogleSecret != "" {
            providers = append(providers,
                google.New(cfg.GoogleKey, cfg.GoogleSecret, cfg.Callback))
        }
        goth.UseProviders(providers...)
    }

    // Existing store + api setup below...
```

The `if cfg.Environment == "production"` guard was already present — Fix 1 is what makes it actually execute. This fix ensures no extra dev provider slips through.

Optionally, add an explicit check at the start of `newAPI()` for belt-and-suspenders:

```go
func newAPI(cfg Config) *api {
    if cfg.Environment == "production" && cfg.StoreKey == "dev-session-key-32-bytes-long!!!" {
        panic("oauthapi: production environment detected with development StoreKey — refusing to start")
    }
    // ...
}
```

---

## Fix 3 — Replace Hardcoded Admin Role with DB Lookup (HIGH)

**Severity:** HIGH
**Risk:** Every OAuth login grants Admin role regardless of actual user permissions
**File:** `ichor/api/domain/http/oauthapi/oauthapi.go`

### Step 1 — Add `UserBus` to the `api` struct and `Config`

```go
// In oauthapi.go:
type api struct {
    log             *logger.Logger
    auth            *auth.Auth
    store           sessions.Store
    tokenKey        string
    uiAdminRedirect string
    uiLoginRedirect string
    tokenExpiration time.Duration
    userBus         *userbus.Business  // ADD
}
```

```go
// In route.go Config struct:
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
    UserBus         *userbus.Business  // ADD
}
```

In `newAPI()`, assign: `userBus: cfg.UserBus`.

In `main.go`, pass `UserBus: userBus` when constructing `oauthCfg` (same `userBus` instance used by other API handlers).

### Step 2 — Replace hardcoded role in `authCallback()`

Current (line 121-129):
```go
claims := auth.Claims{
    RegisteredClaims: jwt.RegisteredClaims{
        Subject:   user.UserID,
        Issuer:    a.auth.Issuer(),
        ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(a.tokenExpiration)),
        IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
    },
    Roles: []string{userbus.Roles.Admin.String()}, // REMOVE THIS
}
```

Replace with a DB lookup:
```go
// Look up user by OAuth provider email
dbUser, err := a.userBus.QueryByEmail(r.Context(), user.Email)
if err != nil {
    a.log.Error(r.Context(), "oauth user not found in database: %s", user.Email)
    http.Error(w, "Unauthorized: user not registered", http.StatusForbidden)
    return
}

if !dbUser.Enabled {
    a.log.Error(r.Context(), "oauth login attempt by disabled user: %s", user.Email)
    http.Error(w, "Unauthorized: account disabled", http.StatusForbidden)
    return
}

claims := auth.Claims{
    RegisteredClaims: jwt.RegisteredClaims{
        Subject:   dbUser.ID.String(),  // use DB user ID, not OAuth provider ID
        Issuer:    a.auth.Issuer(),
        ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(a.tokenExpiration)),
        IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
    },
    Roles: dbUser.Roles,  // use actual DB roles
}
```

**Note:** Check what method `userbus.Business` exposes for lookup by email. The exact method name may be `QueryByEmail`, `QueryByFilter` with an email filter, or similar — look at how `basicauthapi` does its user lookup for the pattern to follow.

---

## Fix 4 — JWT in Redirect URL: Query Param → URL Fragment (HIGH)

**Severity:** HIGH
**Risk:** JWT exposed in server logs, browser history, and Referer header
**File:** `ichor/api/cmd/services/ichor/main.go`

Change the default for `UIAdminRedirect`:

```go
// Before:
UIAdminRedirect string `conf:"default:http://localhost:3001/admin?token="`

// After:
UIAdminRedirect string `conf:"default:http://localhost:3001/admin#token="`
```

The backend redirect code (`oauthapi.go:138`) already appends the token directly: `a.uiAdminRedirect+token`. Changing the default config is all that's needed on the backend.

**Frontend coordination required:** The admin Vue app must be updated to read the token from `window.location.hash` instead of `window.location.search`. Coordinate with the frontend team before deploying. This is tracked as a paired fix in the Vue repo's security findings.

---

## Fix 5 — Session Cookie Security Flags (MEDIUM — INV-006)

**Severity:** MEDIUM
**Risk:** Cookie sent over plain HTTP; no CSRF protection
**File:** `ichor/api/domain/http/oauthapi/oauthapi.go:50-55`

```go
// Before:
store.Options = &sessions.Options{
    Path:     "/",
    MaxAge:   86400 * 30,
    HttpOnly: true,
    Secure:   false,
}

// After:
store.Options = &sessions.Options{
    Path:     "/",
    MaxAge:   900, // 15 minutes — see Fix 6
    HttpOnly: true,
    Secure:   cfg.Environment == "production",
    SameSite: http.SameSiteLaxMode,
}
```

`SameSite: Lax` is required (not `Strict`) because the OAuth callback redirect arrives cross-site and must carry the state cookie for CSRF validation.

---

## Fix 6 — Reduce Session Cookie MaxAge (MEDIUM)

**Severity:** MEDIUM
**Risk:** 30-day stolen session cookie
**File:** `ichor/api/domain/http/oauthapi/oauthapi.go:52`

Applied in Fix 5 above: change `MaxAge: 86400 * 30` to `MaxAge: 900` (15 minutes).

The Gothic session cookie is only needed during the OAuth handshake (seconds). 15 minutes provides ample buffer for slow redirects while eliminating the 30-day exposure window. If there is any long-lived server-side state stored in the cookie beyond the OAuth handshake, identify and move it to a separate cookie or DB-backed session before applying this change.

---

## Fix 7 — JWT Blocklist on OAuth Logout (MEDIUM)

**Severity:** MEDIUM
**Risk:** JWT remains valid 20 minutes after logout
**File:** `ichor/api/domain/http/oauthapi/oauthapi.go:141-165`

This fix requires the JWT blocklist infrastructure from Phase 1b (rate limiting work) to be in place first, or implemented alongside it.

### Approach

Add a `jti` (JWT ID) claim to all issued tokens (both basic and OAuth), then check it on the blocklist middleware and add to blocklist on logout.

### Step 1 — Add `jti` to Claims (if not already present)

In `ichor/app/sdk/auth/auth.go`, ensure `GenerateToken` includes a `jti`:

```go
import "github.com/google/uuid"

// In GenerateToken():
claims.RegisteredClaims.ID = uuid.NewString()  // jti
```

### Step 2 — Create blocklist store

A simple in-memory blocklist is sufficient for a single-node deployment:

```go
// ichor/app/sdk/auth/blocklist.go
package auth

import (
    "sync"
    "time"
)

type BlocklistEntry struct {
    ExpiresAt time.Time
}

type Blocklist struct {
    mu      sync.RWMutex
    entries map[string]BlocklistEntry // jti → expiry
}

func NewBlocklist() *Blocklist {
    bl := &Blocklist{entries: make(map[string]BlocklistEntry)}
    go bl.cleanupLoop()
    return bl
}

func (bl *Blocklist) Add(jti string, expiresAt time.Time) {
    bl.mu.Lock()
    defer bl.mu.Unlock()
    bl.entries[jti] = BlocklistEntry{ExpiresAt: expiresAt}
}

func (bl *Blocklist) IsRevoked(jti string) bool {
    bl.mu.RLock()
    defer bl.mu.RUnlock()
    entry, ok := bl.entries[jti]
    return ok && time.Now().Before(entry.ExpiresAt)
}

func (bl *Blocklist) cleanupLoop() {
    for {
        time.Sleep(5 * time.Minute)
        bl.mu.Lock()
        for jti, entry := range bl.entries {
            if time.Now().After(entry.ExpiresAt) {
                delete(bl.entries, jti)
            }
        }
        bl.mu.Unlock()
    }
}
```

### Step 3 — Check blocklist in authen middleware

In `ichor/api/sdk/http/mid/authen.go`, after validating the JWT:

```go
if blocklist.IsRevoked(claims.ID) {
    return errs.Newf(errs.Unauthenticated, "token has been revoked")
}
```

### Step 4 — Add to blocklist in OAuth logout

In `oauthapi.go` `logout()`, extract and revoke the JWT from the `Authorization` header if present:

```go
func (a *api) logout(w http.ResponseWriter, r *http.Request) {
    // Revoke JWT if present
    if authHeader := r.Header.Get("Authorization"); authHeader != "" {
        if strings.HasPrefix(authHeader, "Bearer ") {
            token := strings.TrimPrefix(authHeader, "Bearer ")
            if claims, err := a.auth.ParseClaims(token); err == nil {
                a.blocklist.Add(claims.ID, claims.ExpiresAt.Time)
            }
        }
    }

    // Existing Gothic logout...
    sess.Values["user"] = nil
    gothic.Logout(w, r)
    http.Redirect(w, r, a.uiLoginRedirect, http.StatusFound)
}
```

**Note:** This is also applicable to `basicauthapi` logout for the same reason. Coordinate the blocklist as shared infrastructure passed to both API handlers via their `Config` structs.

---

## Fix 8 — Rate Limiting on OAuth Endpoints (LOW)

**Severity:** LOW (HIGH when combined with dev provider bypass)
**File:** `ichor/api/domain/http/oauthapi/route.go`

Reuse the `RateLimiter` infrastructure from Phase 1b Fix 1 (or implement it now if Phase 1b hasn't landed yet).

```go
// In route.go Routes():
import (
    "golang.org/x/time/rate"
    appmid "github.com/timmaaaz/ichor/app/sdk/mid"
    httpmid "github.com/timmaaaz/ichor/api/sdk/http/mid"
)

func Routes(app *web.App, cfg Config) {
    api := newAPI(cfg)

    // 5 auth initiations per minute per IP
    oauthLimiter := appmid.NewRateLimiter(rate.Every(12*time.Second), 3)

    app.RawHandlerFunc(http.MethodGet, "", "/api/auth/{provider}",
        httpmid.RateLimit(oauthLimiter)(api.authenticate))
    app.RawHandlerFunc(http.MethodGet, "", "/api/auth/{provider}/callback",
        httpmid.RateLimit(oauthLimiter)(api.authCallback))
    app.RawHandlerFunc(http.MethodGet, "", "/api/logout/{provider}", api.logout)
}
```

**Note:** `RawHandlerFunc` takes a plain `http.HandlerFunc`, not a `web.HandlerFunc`. Verify the adapter pattern works with `RawHandlerFunc` — may need to wrap differently than standard `HandlerFunc` routes. Check how `RawHandlerFunc` is implemented in `ichor/foundation/web/web.go`.

---

## Fix 9 — Provider Name Allowlist (LOW)

**Severity:** LOW
**File:** `ichor/api/domain/http/oauthapi/oauthapi.go:60-79`

```go
// Allowlist of known provider names
var validProviders = map[string]bool{
    "google":      true,
    "development": true,
}

gothic.GetProviderName = func(r *http.Request) (string, error) {
    segments := strings.Split(strings.Trim(r.URL.Path, "/"), "/")

    if len(segments) >= 3 && segments[0] == "api" && segments[1] == "auth" {
        provider := segments[2]
        if provider == "callback" && len(segments) >= 4 {
            provider = segments[3]
        }
        // Validate against allowlist
        if !validProviders[provider] {
            return "", fmt.Errorf("unknown oauth provider: %q", provider)
        }
        return provider, nil
    }

    // Remove the ?provider= fallback entirely — path-based extraction is sufficient
    return "", errors.New("provider not found in path")
}
```

---

## Verification Steps

After implementing all fixes:

```bash
# Build check
cd ichor && go build ./...

# Test suite
go test ./api/domain/http/oauthapi/... -v

# Manual verify CRITICAL fix: dev provider blocked in production
ICHOR_OAUTH_ENVIRONMENT=production go run ./api/cmd/services/ichor &
# Should still start (Google provider registered)
# Dev provider endpoint should return 404 or error:
curl -v http://localhost:8080/api/auth/development
# Expected: error from Goth ("unknown provider") — NOT a redirect to dev callback

# Manual verify MEDIUM fix: session cookie flags
curl -v http://localhost:8080/api/auth/google 2>&1 | grep -i "set-cookie"
# In production mode: should see Secure; SameSite=Lax

# Manual verify HIGH fix: role from DB, not hardcoded
# After OAuth callback, decode the JWT:
# https://jwt.io — verify Roles field is from DB, not ["admin"]
```

---

## Invariant Verification (post-fix)

Run in the Vue repo after merging:
```
/security-audit-phase regression
```

Expected:
- INV-006: PASS (`Secure: true` in production)
- INV-016: PASS (`Environment` field wired, dev provider blocked in production)
- INV-017: PASS (roles from DB lookup, not hardcoded)

---

## Files Modified

| File | Change |
|---|---|
| `ichor/api/cmd/services/ichor/main.go` | **MODIFY** — add `Environment` to oauthCfg; change `?token=` → `#token=` in UIAdminRedirect; add `UserBus` to oauthCfg |
| `ichor/api/domain/http/oauthapi/route.go` | **MODIFY** — add `UserBus` to Config; wire rate limiting on OAuth routes |
| `ichor/api/domain/http/oauthapi/oauthapi.go` | **MODIFY** — session cookie flags; DB role lookup; startup assertion; provider allowlist; JWT revocation on logout |
| `ichor/app/sdk/auth/blocklist.go` | **CREATE** — in-memory JWT blocklist |
| `ichor/app/sdk/auth/auth.go` | **MODIFY** — add `jti` to JWT claims in `GenerateToken` |
| `ichor/api/sdk/http/mid/authen.go` | **MODIFY** — check blocklist after JWT validation |
| `ichor/app/sdk/mid/ratelimit.go` | **CREATE** (if not done in Phase 1b) |
| `ichor/api/sdk/http/mid/ratelimit.go` | **CREATE** (if not done in Phase 1b) |

## Fix Priority

Fix 1 + 2 first — eliminates BOTH CRITICALs immediately with minimal code change.
Fix 3 + 4 next — HIGH severity, require more wiring but follow existing patterns.
Fix 5 + 6 together — two lines in the same struct literal.
Fix 7 last — depends on blocklist infrastructure; can be deferred to Phase 2 if Phase 1b blocklist isn't ready.
