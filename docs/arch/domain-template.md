# domain/{area}/{entity}

[bus]=business [app]=application [api]=HTTP [db]=store [sdk]=shared
→=depends on ⊕=writes ⊗=reads ⚡=external [tx]=transaction [cache]=cached

Reference domain: sales/orders (ordersbus)

---

## {Entity}Bus [bus]

file: business/domain/{area}/{entity}bus/{entity}bus.go
imports: {entity}db.Store[db], delegate.Delegate, logger.Logger
key facts:
  - Business struct: log *logger.Logger, storer Storer, delegate *delegate.Delegate
  - Constructor: NewBusiness(log, delegate, storer) *Business
  - NewWithTx(tx sqldb.CommitRollbacker) (*Business, error) — creates tx-scoped copy
  - CRUD methods: Create, Update, Delete, Query, Count, QueryByID
  - Create/Update/Delete each call delegate.Call() after DB write (workflow trigger hook)
  - filter: {Entity}Filter with queryable fields

delegate events fired:
  Create → delegate.Call(ctx, ActionCreatedData(entity))    — params: EntityID, UserID, Entity
  Update → delegate.Call(ctx, ActionUpdatedData(before, entity)) — params: EntityID, UserID, Entity, BeforeEntity
  Delete → delegate.Call(ctx, ActionDeletedData(entity))    — params: EntityID, UserID, Entity

⊗ {schema}.{table}
⊕ {schema}.{table}

---

## {Entity}DB [db]

file: business/domain/{area}/{entity}bus/stores/{entity}db/{entity}db.go
imports: sqldb, internal db{Entity} null-type model
key facts:
  - db{Entity} uses sql.NullString/NullTime for nullable string/time columns
  - nullable UUID foreign keys use *uuid.UUID (pointer, not NullString)
  - NamedQuerySlice → returns nil (NOT ErrDBNotFound) for empty results
  - NamedQueryStruct → returns ErrDBNotFound for missing rows

method → function:
  Query(ctx, filter, orderBy, page) ([]Entity, error)  → NamedQuerySlice
  Count(ctx, filter) (int, error)                       → NamedQueryStruct
  QueryByID(ctx, id) (Entity, error)                    → NamedQueryStruct
  Create(ctx, entity) error                             → NamedQueryStruct or ExecContext
  Update(ctx, entity) error                             → ExecContext
  Delete(ctx, entity) error                             → ExecContext

⊗ {schema}.{table}
⊕ {schema}.{table}

---

## {Entity}App [app]

file: app/domain/{area}/{entity}app/{entity}app.go
imports: {entity}bus.Business[bus], auth.Auth
key facts:
  - App struct: {entity}bus *{entity}bus.Business, auth *auth.Auth
  - Constructors: NewApp(bus) *App, NewAppWithAuth(bus, auth) *App
  - NewApp/UpdateApp validation (required fields, format checks)
  - Conversions: toBusNew{Entity}(app) ({entity}bus.New{Entity}, error)
                 toBusUpdate{Entity}(app) ({entity}bus.Update{Entity}, error)
                 ToApp{Entity}(bus) {Entity}
                 ToApp{Entities}(bus []) []{Entity}
  - QueryParams parsed via parseFilter() → {Entity}Filter

---

## {Entity}API [api]

file: api/domain/http/{area}/{entity}api/{entity}api.go
imports: {entity}app.App[app], mid.Authenticate, mid.Authorize
routes:
  GET    /v1/{area}/{entities}          [Authenticate, Authorize(Read)]
  GET    /v1/{area}/{entities}/{id}     [Authenticate, Authorize(Read)]
  POST   /v1/{area}/{entities}          [Authenticate, Authorize(Create)]
  PUT    /v1/{area}/{entities}/{id}     [Authenticate, Authorize(Update)]
  DELETE /v1/{area}/{entities}/{id}     [Authenticate, Authorize(Delete)]
  POST   /v1/{area}/{entities}/{id}/{action}  [Authenticate, Authorize(Update)]  ← action-verb routes (optional)

---

## Migration [mig]

file: business/sdk/migrate/sql/migrate.sql
version: X.YY
table: {schema}.{table}

---

## ⚠ Adding a filter field to {Entity}Filter

  business/domain/{area}/{entity}bus/{entity}bus.go               (add field to QueryFilter struct)
  business/domain/{area}/{entity}bus/stores/{entity}db/{entity}db.go   (add WHERE clause + param binding)
  app/domain/{area}/{entity}app/{entity}app.go                    (add to parseFilter() + QueryParams struct)
  api/cmd/services/ichor/tests/{area}/{entity}api/{entity}_query_test.go   (update test assertions)

## ⚠ Adding a column to {schema}.{table}

  business/sdk/migrate/sql/migrate.sql                            (new version, ALTER TABLE — never edit existing)
  business/domain/{area}/{entity}bus/stores/{entity}db/{entity}db.go   (db{Entity} model + SELECT/INSERT/UPDATE query)
  business/domain/{area}/{entity}bus/{entity}bus.go               (bus model if exposed in API)
  app/domain/{area}/{entity}app/model.go                          (app model + toBus*/ToApp* conversions)

## ⚠ Adding a new domain entity (full 7-layer checklist)

  business/domain/{area}/{entity}bus/{entity}bus.go               (Business struct + CRUD methods)
  business/domain/{area}/{entity}bus/stores/{entity}db/{entity}db.go   (DB store)
  app/domain/{area}/{entity}app/{entity}app.go                    (App layer)
  app/domain/{area}/{entity}app/model.go                          (models + conversions)
  api/domain/http/{area}/{entity}api/{entity}api.go               (HTTP handlers)
  api/domain/http/{area}/{entity}api/route.go                     (route registration)
  api/cmd/services/ichor/build/all/all.go                         (wire bus + register routes)
  business/sdk/migrate/sql/migrate.sql                            (new table migration)
  business/sdk/dbtest/dbtest.go                                   (add to BusDomain)
  api/cmd/services/ichor/tests/{area}/{entity}api/                (integration tests)
  api/cmd/services/ichor/tests/{area}/{entity}api/seed_test.go    (seed data)
