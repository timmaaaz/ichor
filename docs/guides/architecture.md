# Architecture Overview

## Ardan Labs Layer Architecture

This codebase follows the **Ardan Labs Service Starter Kit** architecture (Domain-Driven, Data-Oriented Design).

**Layer rules** (higher imports lower, NEVER reverse):
```
┌─────────────────────────────────────────────────────┐
│  api/             HTTP handlers, routes, tests      │
│  ├── domain/http/  Domain-specific HTTP APIs        │
│  ├── cmd/services/  Service entry points            │
│  └── sdk/http/      HTTP framework utilities        │
└─────────────────────────────────────────────────────┘
                        ↓
┌─────────────────────────────────────────────────────┐
│  app/             Application layer                 │
│  ├── domain/       Domain apps (validation, conv)   │
│  └── sdk/          App-level utilities              │
└─────────────────────────────────────────────────────┘
                        ↓
┌─────────────────────────────────────────────────────┐
│  business/        Business logic layer              │
│  ├── domain/       Domain business packages (*bus)  │
│  └── sdk/          Business utilities, migration    │
└─────────────────────────────────────────────────────┘
                        ↓
┌─────────────────────────────────────────────────────┐
│  foundation/      Framework-agnostic utilities      │
│                   (logger, keystore, otel, web)     │
└─────────────────────────────────────────────────────┘
```

**Key Rules:**
- Higher layers can import lower layers, NEVER the reverse
- Business layer contains ALL domain logic
- App layer validates and transforms between API ↔ Business models
- API layer handles HTTP concerns only (routing, middleware, serialization)

**Package naming**: `*bus` (business), `*app` (application), `*api` (api layer)

For detailed patterns (Encoder/Decoder interfaces, Storer pattern, model conversion), see [Layer Patterns](../layer-patterns.md).

## Domain Organization

Domains are organized by business area using PostgreSQL schemas:

- **core/** - Users, roles, permissions, contact info (`core.*` tables)
- **hr/** - Offices, titles, reports-to, homes, approval (`hr.*` tables)
- **geography/** - Countries, regions, cities, streets (`geography.*` tables)
- **assets/** - Asset types, conditions, valid assets, user assets (`assets.*` tables)
- **inventory/** - Warehouses, zones, locations, items, tracking (`inventory.*` tables)
- **products/** - Products, brands, categories, costs, attributes (`products.*` tables)
- **procurement/** - Suppliers, supplier products (`procurement.*` tables)
- **sales/** - Customers, orders, line items, fulfillment (`sales.*` tables)
- **config/** - Table configs, page configs, forms (`config.*` tables)
- **workflow/** - Automation rules, actions, entities (`workflow.*` tables)

## Naming Conventions

**Business Layer Packages** end in `bus`:
- `userbus`, `assetbus`, `productbus`, etc.
- Located in `business/domain/{area}/{entity}bus/`
- Data stores in `business/domain/{area}/{entity}bus/stores/{entity}db/`

**Application Layer Packages** end in `app`:
- `userapp`, `assetapp`, `productapp`, etc.
- Located in `app/domain/{area}/{entity}app/`

**API Layer Packages** end in `api`:
- `userapi`, `assetapi`, `productapi`, etc.
- Located in `api/domain/http/{area}/{entity}api/`

**Model Naming**:
- Creation: `New{Entity}` (e.g., `NewUser`, `NewAsset`)
- Update: `Update{Entity}` (e.g., `UpdateUser`, `UpdateAsset`)
- Response: `{Entity}` (e.g., `User`, `Asset`)

## Service Architecture

**Main Service**: `api/cmd/services/ichor/main.go`
- Entry point for the Ichor API service
- Configures: Database, Auth, OAuth, Tracing, CORS
- Routes can be built in different configurations:
  - `all` - All routes (default)
  - `crud` - Transactional endpoints only
  - `reporting` - Reporting endpoints only

**Auth Service**: `api/cmd/services/auth/` (separate microservice)

**Metrics Service**: `api/cmd/services/metrics/` (observability)

**Route Binding**: `api/cmd/services/ichor/build/`
- `all/all.go` - Binds all domain routes
- `crud/crud.go` - CRUD-only routes
- `reporting/reporting.go` - Reporting-only routes
