# auth

[bus]=business [app]=application [api]=HTTP [db]=store [sdk]=shared [cache]=sturdyc
→=depends on ⊕=writes ⊗=reads ⚡=external [tx]=transaction [cache]=cached

---

## Pipeline

request → Authenticate middleware → Authorize/AuthorizeTable middleware → handler
                ↓                              ↓
         auth.Auth.Authenticate()     permissionsbus.QueryUserPermissions()
                ↓                              ↓
         Claims → context           UserPermissions (cached via rolecache + tableaccesscache)

key facts:
  - Two-layer auth: JWT validation (app/sdk/auth) + RBAC permission checks (permissionsbus)
  - permissionsbus does NOT live inside auth.Auth — table-level access checked separately via AuthorizeTable middleware

---

## Auth [sdk]

file: app/sdk/auth/auth.go
```go
type Auth struct {
    keyLookup KeyLookup
    userBus   *userbus.Business
    method    jwt.SigningMethod
    parser    *jwt.Parser
    issuer    string
}
```

  Authenticate(ctx context.Context, bearerToken string) (Claims, error)
  Authorize(ctx context.Context, claims Claims, userID uuid.UUID, rule string) error

imports: userbus, approvalbus

---

## PermissionsBus [bus]

file: business/domain/core/permissionsbus/permissionsbus.go
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

  NewBusiness(log, del, storer, urb *userrolebus.Business, tab *tableaccessbus.Business, rb *rolebus.Business) *Business
  NewWithTx(tx sqldb.CommitRollbacker) (*Business, error)
  QueryUserPermissions(ctx context.Context, userID uuid.UUID) (UserPermissions, error)
  ClearCache(ctx context.Context, data delegate.Data) error

Storer interface:
  NewWithTx(tx) (Storer, error)
  QueryUserPermissions(ctx, userID) (UserPermissions, error)
  ClearCache()

import scope: 89 files in api/ and app/

---

## RoleCache [cache]

file: business/domain/core/rolebus/stores/rolecache/rolecache.go
sturdyc config: TTL=60min  Capacity=10000  Shards=10

cached methods (cache-first, write-through):
  QueryByID(ctx, id) (Role, error)          ← cache hit → return; miss → DB + fill cache
  QueryByIDs(ctx, ids) ([]Role, error)       ← partial cache hits; fetch missing from DB

write-through (DB write → cache update):
  Create(ctx, role) (Role, error)
  Update(ctx, role) (Role, error)
  Delete(ctx, role) error

pass-through (no cache):
  Query(ctx, filter, orderBy, page) ([]Role, error)
  Count(ctx, filter) (int, error)
  QueryAll(ctx) ([]Role, error)
  NewWithTx(tx) (Storer, error)

---

## TableAccessCache [cache]

file: business/domain/core/tableaccessbus/stores/tableaccesscache/tableaccesscache.go
sturdyc config: TTL=60min  Capacity=10000  Shards=10

cached methods:
  QueryByID(ctx, id) (TableAccess, error)          ← cache-first
  QueryByRoleIDs(ctx, roleIDs) ([]TableAccess, error)  ← fetches DB, fills cache for all results
  QueryAll(ctx) ([]TableAccess, error)              ← fetches DB, fills entire cache

write-through:
  Create(ctx, ta) (TableAccess, error)
  Update(ctx, ta) (TableAccess, error)
  Delete(ctx, ta) error

pass-through (no cache):
  Query(ctx, filter, orderBy, page) ([]TableAccess, error)
  Count(ctx, filter) (int, error)
  NewWithTx(tx) (Storer, error)

---

## Middleware [app]

file: app/sdk/mid/
all files: authen.go, authorize.go, errors.go, logging.go, metrics.go, mid.go, otel.go, panics.go, restrictfields.go, transaction.go
mirror: api/sdk/http/mid/ (same files + logger.go)

### Authenticate variants

```go
// Remote auth service path
func Authenticate(ctx, client *authclient.Client, authorization string, next HandlerFunc) Encoder

// Local JWT path (same process)
func Bearer(ctx, ath *auth.Auth, authorization string, next HandlerFunc) Encoder

// HTTP Basic auth path
func Basic(ctx, ath *auth.Auth, userBus *userbus.Business, authorization string, next HandlerFunc) Encoder
```
context injection: userIDKey → uuid.UUID,  claimKey → auth.Claims
failure: errs.New(errs.Unauthenticated, err) → HTTP 401

### Authorize variants

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
context injection: userKey → userbus.User,  homeKey → homebus.Home
failure: errs.New(errs.Unauthenticated, err) → HTTP 401

### BeginCommitRollback

```go
func BeginCommitRollback(ctx context.Context, log *logger.Logger, bgn sqldb.Beginner, next HandlerFunc) Encoder
```
context injection: trKey → sqldb.CommitRollbacker
failure: Begin() fail → errs.Internal; Commit() fail → errs.Internal; Rollback() errors logged only (sql.ErrTxDone ignored)

---

## ⚠ Adding a new permission rule/action

  business/domain/core/permissionsbus/permissionsbus.go      (add rule constant if needed)
  app/sdk/mid/authorize.go                                    (new Authorize* variant if new resource type)
  api/domain/http/{area}/{entity}api/route.go                 (pass rule string to AuthorizeTable call)
  business/sdk/migrate/sql/migrate.sql                        (seed default role permissions if required)

## ⚠ Changing Sturdyc TTL or capacity (affects 2 caches independently)

  business/domain/core/rolebus/stores/rolecache/rolecache.go          (TTL=60min, Capacity=10000, Shards=10)
  business/domain/core/tableaccessbus/stores/tableaccesscache/tableaccesscache.go  (same values — change separately)
  Note: permissionsbus.ClearCache() is a delegate subscriber — it clears both caches on relevant events

## ⚠ Adding middleware to the route chain (ordering matters)

  api/domain/http/{area}/{entity}api/route.go    (append to handler chain — Authenticate must precede Authorize)
  app/sdk/mid/{middleware}.go                     (implement new middleware following HandlerFunc → Encoder pattern)
  app/sdk/mid/mid.go                              (register if it needs shared context key definitions)
  Note: Authenticate injects userIDKey + claimKey that Authorize reads — never reorder these two

## ⚠ Adding a new context key (from middleware)

  app/sdk/mid/mid.go         (define key type + Set*/Get* helpers)
  The new middleware file     (call setXxx(ctx, value) to inject)
  All callers that read it    (call GetXxx(ctx) to extract — compile error if missing)
