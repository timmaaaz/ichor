# Progress Summary: auth.md

## Overview
Architecture for Ichor's two-layer authentication and authorization system. Covers JWT validation, RBAC via permissions, token revocation, and caching.

## Pipeline

```
request
   ↓
Authenticate middleware → auth.Auth.Authenticate()
   ↓
Claims → context
   ↓
Authorize/AuthorizeTable middleware → permissionsbus.QueryUserPermissions()
   ↓
UserPermissions (cached via rolecache + tableaccesscache)
   ↓
handler
```

### Key Facts
- **Two-layer model:** JWT validation (app/sdk/auth) + RBAC permission checks (permissionsbus)
- **Separate concerns:** permissionsbus is NOT inside auth.Auth; table-level access checked via separate AuthorizeTable middleware

## Auth [sdk] — `app/sdk/auth/auth.go`

### Struct
```go
type Config struct {
    // ...
    Blocklist *Blocklist // optional; revoked JTIs rejected in Authenticate
}

type Auth struct {
    keyLookup KeyLookup
    userBus   *userbus.Business
    method    jwt.SigningMethod
    parser    *jwt.Parser
    issuer    string
    blocklist *Blocklist // nil = no revocation check
}
```

### Methods
- `Authenticate(ctx context.Context, bearerToken string) (Claims, error)` — validates JWT + user enabled + blocklist check (if configured)
- `Authorize(ctx context.Context, claims Claims, userID uuid.UUID, rule string) error` — basic rule check (deprecated, use Authorize middleware instead)
- `GenerateToken(kid string, claims Claims) (string, error)` — auto-assigns claims.ID (jti) via uuid.NewString() if empty

### Dependencies
- userbus, approvalbus

## Blocklist [sdk] — `app/sdk/auth/blocklist.go`

**Responsibility:** In-memory store of revoked JWT token IDs (JTI).

### Key Facts
- **Data structure:** map[jti]expiresAt; thread-safe (sync.Mutex)
- **TTL cleanup:** Background goroutine purges expired entries every 5 minutes
- **Multi-node deployments:** Replace with shared store (Redis / DB)
- **Lifecycle:** NewBlocklist() starts cleanup; call Stop() on graceful shutdown
- **No-op for empty jti:** Both Add() and IsRevoked() handle nil/empty safely

### Methods
- `Add(jti string, expiresAt time.Time)` — records revoked token
- `IsRevoked(jti string) bool` — returns true only if jti present AND not yet expired

### Integration
- Wired into auth.Config.Blocklist before auth.New()
- Authenticate() checks it on every request

## PermissionsBus [bus] — `business/domain/core/permissionsbus/permissionsbus.go`

**Responsibility:** Query user's permissions (roles, table access).

### Struct
```go
type Business struct {
    log            *logger.Logger
    del            *delegate.Delegate
    storer         Storer
    RolesBus       *rolebus.Business
    UserRolesBus   *userrolebus.Business
    TableAccessBus *tableaccessbus.Business
}
```

### Methods
- `NewBusiness(log, del, storer, urb *userrolebus.Business, tab *tableaccessbus.Business, rb *rolebus.Business) *Business`
- `NewWithTx(tx sqldb.CommitRollbacker) (*Business, error)` — transaction-scoped copy
- `QueryUserPermissions(ctx context.Context, userID uuid.UUID) (UserPermissions, error)` — returns cached permissions
- `ClearCache(ctx context.Context, data delegate.Data) error` — delegate subscriber for cache invalidation

### Storer Interface
- `NewWithTx(tx) (Storer, error)`
- `QueryUserPermissions(ctx, userID) (UserPermissions, error)`
- `ClearCache()`

## RoleCache [cache] — `business/domain/core/rolebus/stores/rolecache/rolecache.go`

**Responsibility:** Cache role lookups via Sturdyc.

### Sturdyc Config
- TTL: 60 minutes
- Capacity: 10,000 entries
- Shards: 10

### Cached Methods (cache-first, write-through)
- `QueryByID(ctx, id) (Role, error)` — hit → return; miss → DB + fill cache
- `QueryByIDs(ctx, ids) ([]Role, error)` — partial hits; fetch missing from DB

### Write-Through Methods (DB write → cache update)
- `Create(ctx, role) (Role, error)`
- `Update(ctx, role) (Role, error)`
- `Delete(ctx, role) error`

### Pass-Through Methods (no cache)
- `Query(ctx, filter, orderBy, page) ([]Role, error)`
- `Count(ctx, filter) (int, error)`
- `QueryAll(ctx) ([]Role, error)`
- `NewWithTx(tx) (Storer, error)`

## TableAccessCache [cache] — `business/domain/core/tableaccessbus/stores/tableaccesscache/tableaccesscache.go`

**Responsibility:** Cache table-level access rules via Sturdyc.

### Sturdyc Config
- TTL: 60 minutes
- Capacity: 10,000 entries
- Shards: 10 (same as RoleCache)

### Cached Methods
- `QueryByID(ctx, id) (TableAccess, error)` — cache-first
- `QueryByRoleIDs(ctx, roleIDs) ([]TableAccess, error)` — fetches DB, fills cache for all results
- `QueryAll(ctx) ([]TableAccess, error)` — fetches DB, fills entire cache

### Write-Through Methods
- `Create(ctx, ta) (TableAccess, error)`
- `Update(ctx, ta) (TableAccess, error)`
- `Delete(ctx, ta) error`

### Pass-Through Methods
- `Query(ctx, filter, orderBy, page) ([]TableAccess, error)`
- `Count(ctx, filter) (int, error)`
- `NewWithTx(tx) (Storer, error)`

## Middleware [app] — `app/sdk/mid/` + `api/sdk/http/mid/`

**All files:** authen.go, authorize.go, errors.go, logging.go, metrics.go, mid.go, otel.go, panics.go, restrictfields.go, transaction.go

### Authenticate Variants

Three authentication paths (all inject context):

```go
// Remote auth service path
func Authenticate(ctx, client *authclient.Client, authorization string, next HandlerFunc) Encoder

// Local JWT path (same process)
func Bearer(ctx, ath *auth.Auth, authorization string, next HandlerFunc) Encoder

// HTTP Basic auth path
func Basic(ctx, ath *auth.Auth, userBus *userbus.Business, authorization string, next HandlerFunc) Encoder
```

- **Context injection:** `userIDKey → uuid.UUID`, `claimKey → auth.Claims`
- **Failure:** `errs.New(errs.Unauthenticated, err) → HTTP 401`

### Authorize Variants

Four authorization patterns (all require prior Authenticate):

```go
// Rule-only check via auth service
func Authorize(ctx, client *authclient.Client, rule string, next HandlerFunc) Encoder

// Table-level RBAC check (most common for domain routes)
func AuthorizeTable(ctx, client *authclient.Client, permissionsBus *permissionsbus.Business, tableInfo *TableInfo, rule string, next HandlerFunc) Encoder

// Ownership check (caller must own the resource)
func AuthorizeUser(ctx, client *authclient.Client, userBus *userbus.Business, rule string, id string, next HandlerFunc) Encoder

// Home-specific resource ownership
func AuthorizeHome(ctx, client *authclient.Client, homeBus *homebus.Business, id string, next HandlerFunc) Encoder
```

- **Context injection:** `userKey → userbus.User`, `homeKey → homebus.Home`
- **Failure:** `errs.New(errs.Unauthenticated, err) → HTTP 401`

### BeginCommitRollback

```go
func BeginCommitRollback(ctx context.Context, log *logger.Logger, bgn sqldb.Beginner, next HandlerFunc) Encoder
```

- **Context injection:** `trKey → sqldb.CommitRollbacker`
- **Failure:** Begin() fail → `errs.Internal`; Commit() fail → `errs.Internal`; Rollback() errors logged only (sql.ErrTxDone ignored)

## UserBus [bus] — `business/domain/core/userbus/userbus.go`

### Password Security
- **bcryptCost = 12** (constant; higher than bcrypt.DefaultCost=10)
- Applied in both Create and Update when a new password is set

## Change Patterns

### ⚠ Revoking a Token (Logout)
Affects 2 areas:
1. `app/sdk/auth/blocklist.go` — Blocklist.Add() to record jti + expiry
2. `api/cmd/services/ichor/main.go` — Blocklist created before auth.New(); passed as auth.Config.Blocklist
3. **Note:** Logout handlers call auth.Authenticate (not ParseClaims) to verify signature before blocklisting

### ⚠ Adding a New Permission Rule/Action
Affects 3-4 areas:
1. `business/domain/core/permissionsbus/permissionsbus.go` — add rule constant if needed
2. `app/sdk/mid/authorize.go` — new Authorize* variant if new resource type
3. `api/domain/http/{area}/{entity}api/route.go` — pass rule string to AuthorizeTable call
4. `business/sdk/migrate/sql/migrate.sql` — seed default role permissions if required

### ⚠ Changing Sturdyc TTL or Capacity (Affects 2 Caches Independently)
Both caches have identical defaults but can be tuned separately:
1. `business/domain/core/rolebus/stores/rolecache/rolecache.go` — TTL=60min, Capacity=10,000, Shards=10
2. `business/domain/core/tableaccessbus/stores/tableaccesscache/tableaccesscache.go` — same values; change separately
3. **Note:** permissionsbus.ClearCache() is a delegate subscriber; clears both caches on relevant events

### ⚠ Adding Middleware to the Route Chain (Ordering Matters)
Affects 3 areas:
1. `api/domain/http/{area}/{entity}api/route.go` — append to handler chain (Authenticate must precede Authorize)
2. `app/sdk/mid/{middleware}.go` — implement new middleware following HandlerFunc → Encoder pattern
3. `app/sdk/mid/mid.go` — register if it needs shared context key definitions
4. **Critical:** Authenticate injects userIDKey + claimKey that Authorize reads; never reorder these two

### ⚠ Adding a New Context Key (from Middleware)
Affects 2-3 areas:
1. `app/sdk/mid/mid.go` — define key type + Set*/Get* helpers
2. The new middleware file — call setXxx(ctx, value) to inject
3. All callers that read it — call GetXxx(ctx) to extract (compile error if missing)

## Critical Points
- Authenticate and Authorize are **separate middleware steps**; order matters
- Both caches use Sturdyc with identical defaults but are cached independently
- Token revocation (Blocklist) is **per-instance only** (in-memory); multi-node deployments need Redis/DB
- bcrypt cost is intentionally higher (12 vs 10) for security
- Context keys must be defined in mid.go for consistency

## Notes for Future Development
Authentication is multi-layered (JWT signature + user enabled + blocklist) for security. Permission checks leverage two independent caches for performance. Most changes will be adding new permission rules (straightforward) rather than auth pipeline modifications (risky).
