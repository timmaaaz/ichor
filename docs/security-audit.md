# Security Audit — Open Issues

Findings from a security audit of the Ichor backend (2026-03). Each issue includes the exact location, current behavior, risk, and a concrete fix. Issues are ordered by priority within each severity tier.

Phases 1–2 of a planned 13-phase audit have been completed. Additional issues will be appended as later phases are audited.

---

## CRITICAL

### [C-1] CORS default accepts all origins (`*`)

- **File:** `api/cmd/services/ichor/main.go:90`
- **Current code:**
  ```go
  CORSAllowedOrigins []string `conf:"default:*"`
  ```
- **Risk:** Any origin (including attacker-controlled sites) can make credentialed cross-origin requests to the API. Browsers will include cookies and auth headers on those requests.
- **Fix:** Change the default to an explicit allowlist. Empty default forces the operator to set it explicitly:
  ```go
  CORSAllowedOrigins []string `conf:"default:https://yourdomain.com"`
  ```
  For local development, set `ICHOR_WEB_CORSALLOWEDORIGINS=http://localhost:3001` in your `.env`. Add a startup assertion that rejects `*` when `cfg.Environment == "production"`.

---

## HIGH

### [H-1] Unbounded request body — memory exhaustion

- **File:** `foundation/web/request.go` (the `Decode` function)
- **Current code:**
  ```go
  data, err := io.ReadAll(r.Body)
  ```
- **Risk:** An attacker can send an arbitrarily large request body (e.g., multi-GB) and exhaust server memory, causing an OOM crash or extreme latency.
- **Fix:** Wrap `r.Body` with `http.MaxBytesReader` before reading. Suggest 10 MB default, configurable per route for SSE/file upload endpoints:
  ```go
  const defaultMaxBodyBytes = 10 << 20 // 10 MB
  r.Body = http.MaxBytesReader(w, r.Body, defaultMaxBodyBytes)
  data, err := io.ReadAll(r.Body)
  if err != nil {
      // http.MaxBytesReader returns *http.MaxBytesError on overflow
      var maxErr *http.MaxBytesError
      if errors.As(err, &maxErr) {
          return fmt.Errorf("request: body exceeds maximum size of %d bytes", defaultMaxBodyBytes)
      }
      return fmt.Errorf("request: unable to read payload: %w", err)
  }
  ```
  Note: `Decode` currently takes `*http.Request` but not `http.ResponseWriter`. Either thread `w` through or apply `MaxBytesReader` earlier in the middleware chain.

---

### [H-2] JWT delivered as query parameter in OAuth redirect URL

- **File:** `api/domain/http/oauthapi/oauthapi.go:180`, `api/cmd/services/ichor/main.go:123`
- **Current code:**
  ```go
  // main.go
  UIAdminRedirect string `conf:"default:http://localhost:3001/admin?token="`

  // oauthapi.go
  http.Redirect(w, r, a.uiAdminRedirect+token, http.StatusFound)
  // Result: http://localhost:3001/admin?token=eyJhbGci...
  ```
- **Risk:** The JWT appears in: (1) server access logs via the `Location` response header, (2) browser history, (3) `Referer` headers sent to any third-party resource the admin page loads, (4) proxy/CDN logs. The 20-minute production token lifetime limits exposure but logs retain the JWT indefinitely.
- **Fix (option A — minimal change):** Use a URL fragment instead of a query parameter. Fragments are never sent to the server and do not appear in `Referer` headers:
  ```go
  // main.go default
  UIAdminRedirect string `conf:"default:http://localhost:3001/admin#token="`
  ```
  The frontend reads `window.location.hash` instead of `window.location.search`. No backend logic changes needed beyond the config default.
- **Fix (option B — more secure):** One-time code exchange. Generate a random opaque code, store `code → jwt` in memory (or Redis) with a 60-second TTL, redirect to `?code=<random>`, and add a `GET /api/auth/token-exchange?code=<code>` endpoint that returns the JWT once and deletes the code. The JWT never appears in a URL.

---

## MEDIUM

### [M-1] Query parameters logged raw — JWT in WebSocket URL written to access log

- **File:** `app/sdk/mid/logging.go:17-18`
- **Current code:**
  ```go
  if rawQuery != "" {
      path = fmt.Sprintf("%s?%s", path, rawQuery)
  }
  ```
- **Risk:** The WebSocket upgrade request for alerts uses `?token=<jwt>` in the URL (see `useAlertWebSocket.ts`). This full URL — including the JWT — is written to every access log entry. Logs are often stored and forwarded to observability platforms where the JWT is then queryable.
- **Fix:** Scrub known sensitive query parameters before logging:
  ```go
  func scrubQuery(rawQuery string) string {
      values, err := url.ParseQuery(rawQuery)
      if err != nil {
          return "[unparseable]"
      }
      for _, key := range []string{"token", "access_token", "jwt"} {
          if values.Has(key) {
              values.Set(key, "[redacted]")
          }
      }
      return values.Encode()
  }

  if rawQuery != "" {
      path = fmt.Sprintf("%s?%s", path, scrubQuery(rawQuery))
  }
  ```

---

### [M-2] `reauthenticate()` email is client-controlled — not bound to the expired session

- **File:** `api/domain/http/basicauthapi/basicauthapi.go` (new endpoint needed)
- **Current behavior:** The frontend sends `{ email, password }` to `/api/auth/basic/login` during re-authentication. The email comes from Pinia store state, which any same-origin code can mutate. The backend has no way to verify the re-auth is for the session that expired.
- **Risk:** Code running in the same browser origin (a compromised dependency, browser extension) can change `authStore.user.email` before re-auth fires, causing the server to issue a token for a different account.
- **Fix:** Add a dedicated `POST /api/auth/basic/reauth` endpoint that accepts only `{ password }`. The email is derived server-side from the expired (but still parseable) token in the `Authorization: Bearer <expired-jwt>` header:
  ```go
  func (a *api) reauth(ctx context.Context, r *http.Request) web.Encoder {
      // Parse the expired token WITHOUT validating expiry to get the subject
      claims, err := a.auth.ParseExpired(r.Header.Get("Authorization"))
      if err != nil {
          return errs.New(errs.Unauthenticated, err)
      }

      var req struct {
          Password string `json:"password" validate:"required"`
      }
      if err := web.Decode(r, &req); err != nil {
          return errs.New(errs.InvalidArgument, err)
      }

      // Look up user by subject from the expired token — client cannot influence this
      user, err := a.userBus.QueryByID(ctx, uuid.MustParse(claims.Subject))
      if err != nil {
          return errs.New(errs.Unauthenticated, errors.New("invalid credentials"))
      }

      if err := userbus.Authenticate(user, req.Password); err != nil {
          return errs.New(errs.Unauthenticated, errors.New("invalid credentials"))
      }

      return a.issueToken(ctx, user)
  }
  ```
  Update the frontend `reauthenticate()` in `src/stores/auth.ts` to call `POST /api/auth/basic/reauth` with the expired token in the header and only `{ password }` in the body.

---

### [M-3] JWT blocklist — logout doesn't invalidate tokens

- **File:** `api/domain/http/basicauthapi/basicauthapi.go:212-230`, `api/domain/http/oauthapi/oauthapi.go:141-165`
- **Current behavior:** Both logout handlers clear the session/cookie client-side but do not revoke the JWT. A captured token stays valid for up to 20 minutes after logout.
- **Risk:** Stolen tokens (from logs — see H-2, M-1 — or XSS) remain usable until natural expiry. Role demotions also don't take effect until token expiry (mitigated partially by the per-request `isUserEnabled` DB check, which covers account disables but not role changes).
- **Fix:** Add a JTI-based blocklist. Requires adding a `jti` (JWT ID) claim to issued tokens:
  1. In `auth.go`, add `ID: uuid.NewString()` to `RegisteredClaims` when issuing tokens.
  2. Create an in-memory (or Redis-backed) blocklist store with TTL equal to `TokenExpiration`.
  3. In the `Authenticate` middleware, after signature verification, check if the token's `jti` is in the blocklist — reject if present.
  4. In both logout handlers, add the token's `jti` to the blocklist before responding.
  - **Simpler alternative:** If Redis is not available, an in-memory `sync.Map` with periodic cleanup is sufficient for single-instance deployments. For multi-node: use Redis.

---

## LOW

### [L-1] Query parameters logged raw (see M-1 for the WebSocket/JWT angle)

Already covered in M-1. The general principle: any sensitive value that appears in a query param (API keys, tokens) will be logged. The scrubbing fix in M-1 addresses this globally.

---

### [L-2] `RestrictFields` middleware is a no-op

- **File:** `app/sdk/mid/restrictfields.go`
- **Current code:**
  ```go
  func RestrictFields(ctx context.Context, next HandlerFunc) Encoder {
      resp := next(ctx)
      return resp
  }
  ```
- **Risk:** No security risk as-is (it's a pass-through). Risk is that future code may assume it is active and rely on it for field filtering.
- **Fix:** Either implement the intended field-restriction logic, or delete the function and remove all call sites. Dead code with a security-sounding name is a maintenance hazard.

---

### [L-3] OAuth provider name not validated against allowlist

- **File:** `api/domain/http/oauthapi/oauthapi.go:62-79`
- **Current code:** The `provider` string extracted from the URL path segment is passed directly to Goth's registry without validation. An unknown provider causes a 500, but unsanitized input reaching library code is a defense-in-depth gap.
- **Fix:**
  ```go
  validProviders := map[string]bool{"google": true}
  if cfg.Environment != "production" {
      validProviders["development"] = true
  }

  if !validProviders[provider] {
      return "", errs.Newf(errs.InvalidArgument, "unknown provider: %s", provider)
  }
  ```
  Also remove the `?provider=` query parameter fallback — it creates two manipulation surfaces for the same value.

---

### [L-4] Dev/prod token lifetime inconsistency

- **File:** `api/cmd/services/ichor/main.go:387-401`
- **Current behavior:** OAuth login issues 8-hour tokens in dev; basic auth login always issues 20-minute tokens regardless of environment.
- **Risk:** Developer confusion. Devs using basic auth in dev environments encounter frequent re-logins that don't reflect production behavior.
- **Fix:** Apply the same environment check to basic auth config:
  ```go
  basicAuthExpiration := cfg.OAuth.TokenExpiration // 20m
  if cfg.OAuth.Environment != "production" {
      basicAuthExpiration = cfg.OAuth.DevTokenExpiration // 8h
  }
  basicAuthCfg := basicauthapi.Config{
      TokenExpiration: basicAuthExpiration,
      ...
  }
  ```

---

## Phases Not Yet Audited

The following areas have known risk indicators from planning but have not been formally audited. They are listed here so they are not forgotten.

| Area | Known Risk | Relevant Files |
|------|------------|----------------|
| WebSocket auth (Phase 2c) | JWT passed as `?token=` in WebSocket upgrade URL — appears in server logs | `api/sdk/http/mid/authen.go:51` |
| RBAC cache (Phase 4) | `accessiblePages` not invalidated on server-side role revoke | frontend + `business/domain/core/` |
| Form data (Phase 5) | Template variable resolution — cross-entity ownership boundary check | `app/domain/xapp/formdataapp/` |
| Table builder SQL (Phase 6) | Config JSON → SQL translation — verify parameterized queries throughout | `app/domain/xapp/dataapp/` |
| Table builder expressions (Phase 6) | `client_computed_columns.expression` — verify not server-side eval | `app/domain/xapp/dataapp/` |
| Alerts bulk ops (Phase 8) | Dismiss/acknowledge endpoints — verify scoped to requesting user's alerts | `app/domain/xapp/alertapp/` |
| Assistant chat (Phase 9) | `conversation_id` — verify user A cannot access user B's conversation | `app/domain/xapp/agentapp/` |
| Floor transactions (Phase 11b) | `transaction_type` — verify validated server-side, not client-trusted | `app/domain/xapp/inventoryapp/` |

---

## Already Fixed (for reference)

| Issue | Fix | Commit/PR |
|-------|-----|-----------|
| Dev OAuth provider always active in production | `Environment` field wired in `main.go`; production guard + startup panic | backend |
| All OAuth users get hardcoded Admin role | `userbus.ParseRolesToString(dbUser.Roles)` — DB lookup | backend |
| Session cookie `Secure: false` | `Secure: cfg.Environment == "production"` | backend |
| Session cookie missing `SameSite` | `SameSite: http.SameSiteLaxMode` | backend |
| Session cookie 30-day expiry | Reduced to `MaxAge: 900` (15 minutes) | backend |
| No rate limiting on auth endpoints | Token bucket per-IP in `api/sdk/http/mid/ratelimit.go` | backend |
| bcrypt cost factor 10 | `const bcryptCost = 12` in `userbus.go` | backend |
| Refresh window bug — tokens always eligible for refresh | `refreshWindow = 5 * time.Minute`; logic flipped | backend |
| Admin components using wrong localStorage key (`'token'`) | Migrated to `'auth_token'` | frontend PR #76 |
| Direct localStorage reads bypassing authStore | `src/utils/getToken.ts` centralized utility; all sites migrated | frontend `eed41f4` |
| No CSP / security headers | Added to `vite.config.ts`; production requires reverse proxy config | frontend `c2b457c` |
| No `eslint-plugin-security` | Added to `eslint.config.ts` | frontend `c2b457c` |
| 15 npm vulnerabilities | `npm audit fix` applied | frontend `c2b457c` |
