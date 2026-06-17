# domain/{area}/{entity}

[bus]=business [app]=application [api]=HTTP [db]=store [sdk]=shared
→=depends on ⊕=writes ⊗=reads ⚡=external [tx]=transaction [cache]=cached

Reference domain: sales/orders (ordersbus)

---

## {Entity}Bus [bus]

file: business/domain/{area}/{entity}bus/{entity}bus.go
imports: {entity}db.Store[db], delegate.Delegate, outbox.Writer[sdk], logger.Logger
key facts:
  - Business struct: log *logger.Logger, storer Storer, delegate *delegate.Delegate, outbox *outbox.Writer
  - Constructor: NewBusiness(log, delegate, storer) *Business  (outbox injected separately, NOT via the constructor)
  - WithOutbox(w *outbox.Writer) *Business — returns a copy wired to the cascade outbox; the composition
    root (all.go / worker / dbtest) calls this. A nil Writer makes Emit a no-op (inert until F2 cutover)
  - NewWithTx(tx sqldb.CommitRollbacker) (*Business, error) — creates tx-scoped copy
  - CRUD methods: Create, Update, Delete, Query, Count, QueryByID
  - Create/Update/Delete: after the DB write, build `evtData := Action{X}Data(...)` ONCE, then
      b.outbox.Emit(ctx, evtData)  — durable cascade, SAME tx; RETURN its error (rolls the write back)
      b.delegate.Call(ctx, ...)    — best-effort cross-domain hook; error is logged, NOT returned
    The cascade now flows via the outbox+relay, NOT the delegate (see delegate.md / workflow-engine.md).
  - filter: {Entity}Filter with queryable fields

cascade events fired (the SAME Action{X}Data payload feeds BOTH b.outbox.Emit and b.delegate.Call):
  Create → ActionCreatedData(entity)          — params: EntityID, UserID, Entity
  Update → ActionUpdatedData(before, entity)  — params: EntityID, UserID, Entity, BeforeEntity
  Delete → ActionDeletedData(entity)          — params: EntityID, UserID, Entity
  → outbox.Emit persists it to workflow.cascade_outbox (durable; the relay dispatches the cascade)
  → delegate.Call fans it out best-effort in-process (no cascade subscriber today — see delegate.md governance)

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

## ⚠ Adding a second writer to a shared payload (twin-site drift)

rule: two sites feeding the same `map[string]any` consumer → use a typed struct or shared builder, never inline literals

existing shared builders:
  api/domain/http/{area}/{entity}api/{entity}api.go : toApp{Entity}                 (HTTP response shape)
  business/domain/{area}/{entity}bus/event.go : ActionCreated/Updated/DeletedData   (cascade event payload — fed to BOTH outbox.Emit + delegate.Call)
  api/domain/http/workflow/approvalapi/approvalapi.go : buildResolveResult           (Temporal activity Result)
  business/sdk/workflow/workflowactions/communication/alert.go : BuildAlertPayload   (WebSocket alert shape)

incident: PR #126 — retry Temporal path dropped resolved_by/reason
audit: tasks/twin-site-audit.md

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

  if the domain must CASCADE (fire workflow events) — additionally:
  business/domain/{area}/{entity}bus/{entity}bus.go               (add `outbox *outbox.Writer` field + WithOutbox option; b.outbox.Emit(ctx, evtData) + return err in each CRUD)
  business/domain/{area}/{entity}bus/event.go                     (Action{Created,Updated,Deleted}Data payload builders)
  business/sdk/workflowdomains/workflowdomains.go                 (add the (schema, domain, entity) Registrations() entry — drives the outbox entity-name map + relay)
  api/cmd/services/ichor/build/all/all.go                         (append .WithOutbox(outboxWriter) to the bus)
  api/cmd/services/workflow-worker/main.go                        (append .WithOutbox(outboxWriter) IF the worker constructs this bus)
  business/sdk/dbtest/dbtest.go                                   (append .WithOutbox(outboxWriter) so integration tests exercise the cascade — F8 parity)
  → delegate.Call alone no longer cascades; the durable path is the outbox. See delegate.md + workflow-engine.md.
