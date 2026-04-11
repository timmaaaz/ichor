# CLAUDE.md

## Arch Files
docs/arch/ = authoritative. Read → answer directly. No codebase search to verify. ⚠ callouts are complete.

| File | MUST read when |
|------|---------------|
| `docs/arch/domain-template.md` | adding any new domain entity (7-layer checklist) |
| `docs/arch/delegate.md` | adding a new domain that fires events; modifying delegate registration |
| `docs/arch/sqldb.md` | modifying db stores, NamedQuery* error handling, connection config |
| `docs/arch/errs.md` | adding error codes, changing FieldError format, HTTP status mapping |
| `docs/arch/auth.md` | adding permissions, changing RBAC, modifying middleware chain, Sturdyc cache config |
| `docs/arch/seeding.md` | adding domain to BusDomain, new TestSeed* function, seed data changes, writing frontend or integration tests that assert against seeded data, understanding what users/products/orders/inventory exist in test DB |
| `docs/arch/table-builder.md` | modifying tablebuilder config types, adding aggregate/filter/interval type |
| `docs/arch/form-data.md` | touching formdataapp, formdataregistry, form fields, template processor |
| `docs/arch/workflow-engine.md` | adding ActionHandler, changing WorkflowInput, new edge type, Temporal infra |
| `docs/arch/workflow-save.md` | modifying DAG validation, adding edge types, changing graph topology rules |
| `docs/arch/workflow-alerts.md` | modifying WebSocket hub, approvalrequestbus, Temporal async completion |
| `docs/arch/picking.md` | touching pickingapp, FEFO logic, inventory transaction ledger, order status transitions |
| `docs/arch/inventory-ops.md` | touching putawaytask status machine, scan fan-out, inventory upsert logic |
| `docs/arch/agent-chat.md` | touching chatapi, agenttools executor, toolindex RAG, llm provider, toolcatalog |
| `docs/arch/mcp.md` | adding MCP tools, changing tool groups, modifying mcp client or prompts |
| `docs/arch/testing.md` | touching dbtest.go, docker container setup, NewDatabase, apitest, test container lifecycle, parallel test isolation |

## Conditional Reading

**MUST read the relevant guide before starting work in these areas:**

| When you are... | Read this first |
|---|---|
| Understanding layer architecture, domain organization, or naming conventions | `docs/guides/architecture.md` |
| Looking up make targets or dev commands | `docs/guides/commands.md` |
| Working with migrations, env vars, observability, or config | `docs/guides/quick-reference.md` |
| Adding a new domain from SQL schema | `docs/domain-implementation-guide.md` |
| Debugging a bug or test failure | `docs/debugging.md` |
| Working with layer patterns (Encoder/Decoder, Storer) | `docs/layer-patterns.md` |
| Working with financial/money calculations | `docs/financial-calculations.md` |
| Working with the MCP server | `mcp/README.md` |
| Working with the workflow engine | `docs/workflow/README.md` |
| Working with FormData multi-entity operations | `FORMDATA_IMPLEMENTATION.md` |
| Working with customer seed data or onboarding | `docs/arch/customer-seeding.md` |
| Investigating or fixing a bug file | `docs/bugs/README.md` |

---

## Core Rules

- **Never guess — verify.** Before referencing any column, function signature, file path, import, or wiring: read the actual code or grep for it. Do NOT infer from naming conventions. "It's probably called X" → grep for X first. Cost of verification < cost of a wrong guess.
- Do NOT make code changes unless explicitly asked. When the user says 'explore', 'analyze', 'investigate', or 'plan' — read and report only, never edit.
- Prefer targeted, minimal changes over broad refactors. Scope as narrowly as possible.
- When planning or implementing domain changes, all 7 layers must be addressed — see `docs/arch/domain-template.md`.

## Build & Verification

Go codebase with YAML (K8s/config) and TypeScript (Vue3 frontend). Always run `go build` on affected packages before reporting completion.

**NEVER run `go test ./...`** — the repo has hundreds of tests and many require a live database. Always run only the tests for the packages you actually changed. Example:
```bash
go test ./business/domain/sales/orderlineitemsbus/... ./app/domain/sales/pickingapp/...
```

## Testing

**Tests verify the app works. They are not the goal — they are the alarm.**

When a test fails:
1. **Default hypothesis: the production code is broken.** Investigate before touching the test.
2. Read the failing assertion. Trace the code path. Check recent changes to the code under test.
3. Only after you understand *why* it failed, decide what to fix.

**Second hypothesis (not first): the test is wrong.** Legitimate categories:
- Stale `ExpResp` after an intentional behavior change
- Hardcoded counts that drifted when an endpoint/feature was added or removed
- Race conditions in test setup

**Hard rule:** Before changing any `ExpResp`, hardcoded count, or test expectation, you must be able to state in one sentence what intentional behavior change made the old expectation wrong. If you can't, the test is right and the code is broken.

Update test assertions (especially hardcoded counts) when adding or removing endpoints/features.

## Git Workflow

Descriptive conventional commit messages. Include all relevant changed files. No confirmation needed unless the diff is unusually large.

## Infrastructure / K8s

Environment variables follow `ICHOR_*` pattern (`ICHOR_LLM_*` for LLM config). K8s secrets must be created before deployments that reference them. Verify env var naming matches between code, K8s manifests, and Makefile targets.

## Bug Fix Protocol

When the user describes a bug, **recommend `/investigate` before making any code changes**. Example: *"This sounds like a bug — want me to run `/investigate` first?"*

---

## Project Overview

Ichor is a production-grade ERP system built on the **Ardan Labs Service Starter Kit** (Domain-Driven, Data-Oriented Design). Covers HR, Assets, Inventory, Products, Procurement, Sales, and Workflow automation.

**Module**: `github.com/timmaaaz/ichor` | **Go**: 1.23 | **DB**: PostgreSQL 16.4 (multi-schema) | **Deploy**: Kubernetes (KIND for local)

## Important Notes

- **Never skip migrations** — always add new version, never edit existing
- **Business layer is source of truth** — all validation and logic goes here
- **Keep layers pure** — no business logic in API, no HTTP in business
- **Use delegate** — for UUID generation, timestamps (testing seams)
- **Use decimal for money math** — never use float64 for financial calculations
- **Integration tests are primary test strategy**
