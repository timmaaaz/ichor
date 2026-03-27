# Progress Summary: domain-template.md

## Overview
This is the **7-layer domain architecture checklist** for Ichor. It defines the canonical structure for adding any new domain entity (e.g., orders, shipments, vendors).

## Key Concepts

### Layer Abbreviations (notation key)
- `[bus]` = business layer (core logic)
- `[app]` = application layer (validation, authorization)
- `[api]` = HTTP/REST layer (endpoints)
- `[db]` = database store layer
- `[sdk]` = shared utilities
- `→` = depends on
- `⊕` = writes to
- `⊗` = reads from
- `⚡` = external call
- `[tx]` = transaction boundary
- `[cache]` = cached result

### Reference Domain
Uses `sales/orders` (ordersbus) as the standard example throughout.

## The 7 Layers

### 1. {Entity}Bus [bus] — `business/domain/{area}/{entity}bus/{entity}bus.go`
**Core business logic and workflow triggers.**

Struct:
- `log *logger.Logger`
- `storer Storer` (reads/writes database)
- `delegate *delegate.Delegate` (fires events for workflow engine)

CRUD methods: Create, Update, Delete, Query, Count, QueryByID

Delegate events fired:
- `Create` → fires `ActionCreated` with EntityID, UserID, Entity
- `Update` → fires `ActionUpdated` with EntityID, UserID, Entity, BeforeEntity
- `Delete` → fires `ActionDeleted` with EntityID, UserID, Entity

Each mutation calls `delegate.Call()` AFTER the database write (crucial for ordering).

### 2. {Entity}DB [db] — `business/domain/{area}/{entity}bus/stores/{entity}db/{entity}db.go`
**Database persistence layer.**

Key patterns:
- Uses `sql.NullString` / `sql.NullTime` for nullable columns
- Nullable UUIDs use `*uuid.UUID` (pointer, not NullString)
- `NamedQuerySlice` returns `nil` (NOT `ErrDBNotFound`) when no rows
- `NamedQueryStruct` returns `ErrDBNotFound` when row missing

Standard methods map to SQL operations (Query, Count, QueryByID, Create, Update, Delete).

### 3. {Entity}App [app] — `app/domain/{area}/{entity}app/{entity}app.go`
**Application layer: validation, authorization, conversions.**

Struct:
- `{entity}bus *{entity}bus.Business`
- `auth *auth.Auth`

Responsibility: parse query params, validate inputs, convert between API/bus models via `toBus*` and `ToApp*` converters.

### 4. {Entity}API [api] — `api/domain/http/{area}/{entity}api/{entity}api.go`
**HTTP endpoint handlers.**

Standard RESTful routes with auth middleware:
- `GET /v1/{area}/{entities}` — list
- `GET /v1/{area}/{entities}/{id}` — read one
- `POST /v1/{area}/{entities}` — create
- `PUT /v1/{area}/{entities}/{id}` — update
- `DELETE /v1/{area}/{entities}/{id}` — delete
- `POST /v1/{area}/{entities}/{id}/{action}` — optional action verbs

All authenticated and authorized.

### 5. Migration [mig] — `business/sdk/migrate/sql/migrate.sql`
**Database schema.**

Each new table requires a new migration version (never edit existing migrations).

### 6. Wire & Registration
- `api/cmd/services/ichor/build/all/all.go` — wire business bus + register routes
- `business/sdk/dbtest/dbtest.go` — add entity to `BusDomain` enum for testing

### 7. Tests
- `api/cmd/services/ichor/tests/{area}/{entity}api/` — integration test directory
- `seed_test.go` — seed data setup

## Change Patterns

### ⚠ Adding a Filter Field
Affects 4 files:
1. `{entity}bus.go` — add field to QueryFilter struct
2. `{entity}db.go` — add WHERE clause + param binding
3. `{entity}app.go` — parse from QueryParams struct
4. `{entity}_query_test.go` — update test assertions

### ⚠ Adding a Column
Affects 4 files:
1. `migrate.sql` — new migration version (ALTER TABLE, never edit existing)
2. `{entity}db.go` — db{Entity} model + SELECT/INSERT/UPDATE
3. `{entity}bus.go` — if exposed in API
4. `{entity}app.go` + `model.go` — app model + conversions

### ⚠ Adding a Complete New Domain (7-Layer Checklist)
Must create/modify exactly these 11 items:
1. `{entity}bus.go` — business struct + CRUD
2. `{entity}db.go` — database store
3. `{entity}app.go` — application layer
4. `model.go` — app models + conversions
5. `{entity}api.go` — HTTP handlers
6. `route.go` — route registration
7. `all.go` — wire + register routes
8. `migrate.sql` — create table migration
9. `dbtest.go` — add to BusDomain enum
10. `tests/{entity}api/` — integration test directory
11. `seed_test.go` — seed data

## Critical Points
- Delegate events are ALWAYS fired AFTER database writes (workflow engine dependency)
- Filter queries use `NamedQuerySlice` (nil on empty, not error)
- Single missing row queries use `NamedQueryStruct` (error on missing)
- Nullable UUIDs use pointer, not NullString
- All routes require auth middleware (Authenticate + Authorize)
- Never edit an existing migration; always create a new version

## Notes for Future Development
This file is the authoritative checklist when adding any new domain. Should be consulted before implementing new entities.
