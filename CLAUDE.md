# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Core Rules

Do NOT make code changes unless explicitly asked. When the user says 'explore', 'analyze', 'investigate', or 'plan', only read and report — never edit files.

## Code Changes

Prefer targeted, minimal changes over broad refactors. When fixing a bug or making a change, scope it as narrowly as possible unless the user explicitly requests a larger refactor.

## Build & Verification

This is primarily a Go codebase with YAML (K8s/config) and TypeScript (Vue3 frontend). When making changes, always run `go build` on the affected packages before reporting completion. For frontend changes, run the appropriate build/lint command.

**NEVER run `go test ./...`** — the repo has hundreds of tests and many require a live database. Always run only the tests for the packages you actually changed. Example:
```bash
go test ./business/domain/sales/orderlineitemsbus/... ./app/domain/sales/pickingapp/...
```

## Testing

When implementing a plan that adds or removes tools/endpoints, update test assertions (especially tool counts) as part of the same change. Check for hardcoded counts in test files.

## Planning & Implementation

Before proposing an implementation plan, thoroughly explore the existing codebase structure first. Do not assume architecture — check for existing routing systems, naming conventions, and patterns already in use.

## Git Workflow

When the user asks to commit and push, use a descriptive conventional commit message and include all relevant changed files. Do not ask for confirmation unless the diff is unusually large.

## Infrastructure / K8s

Environment variables in this project follow the pattern `ICHOR_LLM_*`. K8s secrets must be created before deployments that reference them. Always verify env var naming matches between code, K8s manifests, and Makefile targets.

## Project Overview

Ichor is a production-grade ERP system built using the **Ardan Labs Service Starter Kit** architecture. It implements Domain Driven, Data Oriented Design patterns with full Kubernetes deployment support. The project is a fork/adaptation of the Ardan Labs service architecture specifically for ERP use cases covering HR, Assets, Inventory, Products, Procurement, Sales, and Workflow automation.

**Module**: `github.com/timmaaaz/ichor`
**Go Version**: 1.23
**Architecture**: Ardan Labs Domain-Driven, Data-Oriented Design
**Deployment**: Kubernetes (KIND for local development)
**Database**: PostgreSQL 16.4 with multi-schema design

## Essential Commands

### Development Setup
```bash
make dev-gotooling    # Install Go tooling
make dev-brew         # Install Homebrew dependencies (kind, kubectl, kustomize, pgcli, watch)
make dev-docker       # Pull Docker images
```

### Testing
```bash
make test             # Run all tests with linting and vulnerability checks
make test-race        # Run tests with race detector
make test-only        # Run only tests (no linting)
make lint             # Lint code
make vuln-check       # Check for vulnerabilities
make test-down        # Shutdown test containers
```

### Local Kubernetes Development
```bash
make dev-up           # Start KIND cluster with all services
make dev-update-apply # Build containers and deploy to KIND
make dev-logs         # View logs (formatted)
make dev-logs-auth    # View auth service logs
make dev-logs-init    # View init container logs
make dev-update       # Restart deployments (after code changes)
make dev-status       # Check pod status
make dev-down         # Shutdown cluster
```

### Database Operations
```bash
make migrate          # Run migrations
make seed             # Seed database with test data
make seed-frontend    # Seed frontend configuration
make pgcli            # Access PostgreSQL CLI
make dev-database-recreate  # Recreate database (deletes all data!)
```

### Docker Compose (Alternative to Kubernetes)
```bash
make compose-up       # Start with existing images
make compose-build-up # Build and start
make compose-logs     # View logs
make compose-down     # Shutdown
```

### Running Locally (Without Containers)
```bash
make run              # Run main service locally
make run-help         # Run with help output
make admin            # Run admin tooling
```

### Authentication & API Testing
```bash
make token            # Get authentication token
export TOKEN=<TOKEN>  # Export token for subsequent requests
make users            # Test users endpoint
make curl-create      # Create new user
make live             # Test liveness probe
make ready            # Test readiness probe
make load             # Run load test (100 concurrent, 1000 requests)
```

## Architecture

### Ardan Labs Layer Architecture

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

For detailed patterns (Encoder/Decoder interfaces, Storer pattern, model conversion), see [Layer Patterns](docs/layer-patterns.md).

### Domain Organization

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

### Naming Conventions

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

### Service Architecture

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

## Quick References

### Testing

**Integration Tests** are located at:
```
api/cmd/services/ichor/tests/{domain}/{entityapi}/
```

**Test Helpers**:
- `business/sdk/unitest` - Unit test helpers for business layer
- `business/sdk/dbtest` - Database test setup
- Use `apitest.Table` pattern for HTTP integration tests

### Database Migrations

**Location**: `business/sdk/migrate/sql/migrate.sql`

**Format**:
```sql
-- Version: X.YY
-- Description: What this migration does
CREATE TABLE schema.table_name (
    id UUID PRIMARY KEY,
    ...
);
```

**Schemas** are versioned separately:
- Version 0.xx - Schema creation
- Version 1.xx - Core tables
- Version 2.xx - Configuration tables

### Authentication & Authorization

**JWT-based** with RSA keys:
- Keys stored in `zarf/keys/` or environment variable `ICHOR_KEYS`
- OAuth providers: Google, Development (for testing)
- Basic auth also supported for testing

**Permission System**:
- Role-based with table-level permissions
- Cached in `permissionsbus`, `rolecache`, `tableaccesscache`
- Check permissions via `PermissionsBus.CheckAccess()`

**Middleware**:
- `mid.Authenticate()` - Validates JWT
- `mid.Authorize()` - Checks permissions

### Caching Strategy

**Sturdyc** used for business layer caching:
- User cache: 1 minute TTL
- Role cache: 60 minutes TTL
- Table access cache: 60 minutes TTL
- Permissions cache: 60 minutes TTL

### Configuration

**Environment Variables** (prefixed with `ICHOR_`):
- `ICHOR_DB_HOST` - Database host
- `ICHOR_DB_USER` - Database user
- `ICHOR_DB_PASSWORD` - Database password
- `ICHOR_KEYS` - RSA keys for JWT (multiline)
- `ICHOR_WEB_API_HOST` - API listen address

**Config Parsing**: Uses `github.com/ardanlabs/conf/v3`

### Observability

**Tracing**: OpenTelemetry → Tempo (Grafana stack)
- Configured in `foundation/otel/`
- 5% sampling by default

**Metrics**: Exposed via `/metrics` endpoint
- View with: `make metrics-view`
- Prometheus format

**Logging**: Structured JSON logs
- `foundation/logger/`
- Trace ID injection
- Format tool: `api/cmd/tooling/logfmt/`

**Visualization**:
- Grafana: `make grafana` (http://localhost:3100)
- Statsviz: `make statsviz` (http://localhost:3010/debug/statsviz)

## Development Workflow

### Adding a New Domain Entity

1. **Create Business Layer** (`business/domain/{area}/{entity}bus/`)
   - Define `{entity}bus.go` with Business struct and methods
   - Define `stores/{entity}db/{entity}db.go` for database operations
   - Define models: `Entity`, `NewEntity`, `UpdateEntity`

2. **Create Application Layer** (`app/domain/{area}/{entity}app/`)
   - Define `{entity}app.go` with App struct
   - Define `model.go` with API models and validation
   - Implement conversion functions: `toBus*()`, `toApp*()`

3. **Create API Layer** (`api/domain/http/{area}/{entity}api/`)
   - Define `{entity}api.go` with HTTP handlers
   - Define `route.go` with route configuration
   - Implement CRUD handlers: create, query, queryByID, update, delete

4. **Wire Dependencies** (`api/cmd/services/ichor/build/all/all.go`)
   - Instantiate business layer: `entityBus := entitybus.NewBusiness(...)`
   - Call `entityapi.Routes(app, entityapi.Config{...})`

5. **Add Tests** (`api/cmd/services/ichor/tests/{area}/{entity}api/`)
   - Create test files using `apitest.Table` pattern
   - Seed test data in `seed_test.go`

6. **Create Migration** (`business/sdk/migrate/sql/migrate.sql`)
   - Add table creation with appropriate schema
   - Follow version numbering convention

### Adding Domain from SQL Schema

Run `/add-domain <sql-schema>` for interactive guided implementation, or see [Domain Implementation Guide](docs/domain-implementation-guide.md) for the full step-by-step walkthrough.

### Debugging

See [Debugging Guide](docs/debugging.md) for troubleshooting common issues.

## Specialized Topics

These systems have their own detailed documentation:

- **FormData System** → [FORMDATA_IMPLEMENTATION.md](FORMDATA_IMPLEMENTATION.md) — Multi-entity transactional operations
- **Workflow Engine** → [docs/workflow/README.md](docs/workflow/README.md) — Automation, Temporal, actions, triggers
- **MCP Server** → [mcp/README.md](mcp/README.md) — LLM agent integration via Model Context Protocol
- **MCP Architecture** → [docs/mcp-architecture.md](docs/mcp-architecture.md) — How MCP fits into the Ichor system
- **Agent Infrastructure** → [docs/agent-infrastructure.md](docs/agent-infrastructure.md) — Catalog, schemas, discovery endpoints
- **Financial Calculations** → [docs/financial-calculations.md](docs/financial-calculations.md) — Decimal arithmetic for money
- **Layer Patterns** → [docs/layer-patterns.md](docs/layer-patterns.md) — Encoder/Decoder interfaces, Storer pattern

### Agent Chat (Conversational AI)

The in-app agent chat (`api/domain/http/agentapi/chatapi/`) uses **Gemini Flash 2.5**. When writing or modifying system prompts, tool descriptions, or chat logic, optimize for Gemini Flash 2.5's strengths and limitations.

**Agent Tool Design Principles** (apply when adding/modifying tools):
- **Minimize tool count per call** — 3-6 tools per LLM call, not 15+. Use intent-based routing to send only relevant tools.
- **Eliminate tools the LLM should never call** — if preview-first means `create_workflow` is a trap, don't send it.
- **Consolidate identical-signature tools** — 3 discovery tools with no params → 1 tool with a category enum.
- **Treat tool descriptions like onboarding docs** — explain to the LLM as if it's a new team member. Make implicit knowledge explicit.
- **Control response size** — add `response_format` params or server-side summarization to prevent tools from flooding context.
- **Tool count directly degrades selection accuracy in non-frontier models** — every extra tool is decision surface the model can get wrong.

**Reference Sources for Agent Tool Best Practices**:
- [Anthropic: Writing Tools for Agents](https://www.anthropic.com/engineering/writing-tools-for-agents) — consolidation, descriptions, thoughtful design over quantity
- [Anthropic: Advanced Tool Use](https://www.anthropic.com/engineering/advanced-tool-use) — tool search tool pattern for large tool sets
- [Anthropic: Tool Use Docs](https://docs.anthropic.com/en/docs/agents-and-tools/tool-use/overview) — implementation patterns, parallel execution
- [Google: Gemini Function Calling](https://ai.google.dev/gemini-api/docs/function-calling) — enum params, description quality, tool count guidance (10-20 max)
- [Tool RAG: Scalable AI Agents (Red Hat)](https://next.redhat.com/2025/11/26/tool-rag-the-next-breakthrough-in-scalable-ai-agents/) — dynamic tool retrieval triples accuracy, halves prompt length
- [LangGraph: Handling Many Tools](https://langchain-ai.github.io/langgraph/how-tos/many-tools/) — routing patterns, tool subsets
- [Optimizing Tool Calling (Paragon)](https://www.useparagon.com/learn/rag-best-practices-optimizing-tool-calling/) — tool selection impact by model tier

## Important Notes

- **Never skip migrations** - Always add new version, never edit existing
- **Business layer is source of truth** - All validation and logic goes here
- **Keep layers pure** - No business logic in API, no HTTP in business
- **Use delegate** - For UUID generation, timestamps (testing seams)
- **Cache carefully** - Only cache read-heavy, infrequently changing data
- **Test everything** - Integration tests are primary test strategy
- **Use decimal for money math** - Never use float64 for financial calculations

## Bug Fix Protocol

When the user describes a bug, error, or unexpected behavior, **recommend `/investigate` before making any code changes**. Look for signals: error messages, "broken", "failing", "wrong", "not working", stack traces, test failures, or unexpected output. A quick suggestion is sufficient — don't block if they decline. Example: *"This sounds like a bug — want me to run `/investigate` first to diagnose the root cause before making changes?"*

## Additional Resources

- **Ardan Labs Course**: https://github.com/ardanlabs/service/wiki
- **Makefile Help**: `make help` for all available commands
