# Table Builder Testing Gaps — Research & Plan

**Date**: 2026-03-10
**Status**: Research complete, implementation not started
**Estimated phases**: 4
**Orchestration model**: Opus (this doc) + spawn-agents per phase

---

## Executive Summary

The dynamic table builder (`business/sdk/tablebuilder/`) has strong coverage of its **declarative layer** (config validation, expression evaluation, type mapping) but near-zero **isolated unit tests for SQL generation**. Most SQL behavior is only exercised indirectly via integration tests that require a live database.

Additionally, agent analysis uncovered **9 likely bugs** in the production code that should be fixed before or alongside tests.

---

## Current Coverage Assessment

### Well-Covered (do not need focus)
| Area | Test File | Subtests |
|------|-----------|----------|
| Config validation | `validation_test.go` | 40+ |
| Expression evaluation | `evaluator_test.go` | 34 |
| Chart transformation | `tablebuilder_test.go` | 25 |
| Multi-dimensional GroupBy | `multi_groupby_test.go` | 15 |
| PostgreSQL type mapping | `typemapper_test.go` | 4 |
| PageConfig CRUD (HTTP) | `pageconfigapi/` | 15+ |
| Data CRUD + execute (HTTP) | `tests/data/` | full 23 routes |

### Coverage Gaps (focus areas)
| Area | Current State | Risk |
|------|---------------|------|
| Filter WHERE clause generation (all 11 operators) | Only `gt` in integration tests | HIGH |
| JOIN condition building | Implicitly tested | HIGH |
| Pagination (`params.Rows`, `ds.Rows` fallback) | Zero isolated tests | MEDIUM |
| Sort priority / param override | Zero isolated tests | MEDIUM |
| Metric aggregation SQL (SUM/COUNT/AVG etc.) | Implicitly tested | HIGH |
| GroupBy interval SQL (`DATE_TRUNC`) | Implicitly tested | MEDIUM |
| `BuildCountQuery` ForeignTable join miss | Zero tests + likely bug | HIGH |
| `mergeData` / secondary source logic | Implicitly tested | MEDIUM |
| ConfigStore protection rules | Zero tests | HIGH |
| Evaluator edge cases (ternary colon, array index >9) | Zero tests | HIGH |

---

## Confirmed Bugs (Found During Research)

These were identified by static analysis of `store.go`/`builder.go`/`configstore.go`. Each should be verified and fixed as part of the test phases.

### BUG-1: `is_null`/`is_not_null` filters silently dropped
**File**: `business/sdk/tablebuilder/store.go` (filter building)
**Problem**: The nil-value early-return in `buildFilterExpression` runs BEFORE the operator switch. An `is_null` filter with `Value: nil` (correct usage) is silently eaten and never applied to the query.
**Severity**: HIGH — business logic bug, filters produce wrong results
**Fix**: Check for `is_null`/`is_not_null` operators before the nil-value guard.

### BUG-2: `Sort.Priority` field never used
**File**: `business/sdk/tablebuilder/store.go` (`applySorting`)
**Problem**: Sorts are applied in slice order. The `Priority` field on `Sort` is defined in the model but never read. Callers who set priority expecting deterministic multi-sort behavior are silently wrong.
**Severity**: MEDIUM — silent semantic bug
**Fix**: Sort `ds.Sort` by `Priority` field before applying, or document Priority as deprecated.

### BUG-3: `BuildCountQuery` missing ForeignTable joins
**File**: `business/sdk/tablebuilder/store.go` (`BuildCountQuery`)
**Problem**: `BuildCountQuery` applies `ds.Joins` but NOT `ds.Select.ForeignTables`. If any filter references a column only available via a ForeignTable join, the count query will fail at runtime while the data query succeeds.
**Severity**: HIGH — runtime error in paginated queries with ForeignTable filters
**Fix**: Apply ForeignTable joins in `BuildCountQuery` the same way they're applied in `BuildQuery`.

### BUG-4: `executeRPC` likely broken
**File**: `business/sdk/tablebuilder/store.go` (`executeRPC`)
**Problem**: Generates `SELECT * FROM funcName(:args)` where `:args` is a map. sqlx named query substitution requires struct/map with named keys, not a single `:args` map literal. This entire code path is likely non-functional.
**Severity**: HIGH (if RPC type is used), LOW (if no configs use it)
**Fix**: Verify if any seeded configs use `type: "rpc"`. If yes, fix the param handling. If no, add a validation error for unsupported type.

### BUG-5: `BuildJoinCondition` raw SQL fallback
**File**: `business/sdk/tablebuilder/store.go`
**Problem**: When a `Join.On` condition doesn't contain `=`, it falls back to `goqu.L(condition)` — raw SQL passed as a literal. `ValidateConfig` only checks that `On` is non-empty, not that it's a safe column reference pattern.
**Severity**: MEDIUM — potential injection surface (admin-controlled, but still)
**Fix**: Apply `isValidColumnReference` to both sides of join conditions in `ValidateConfig`.

### BUG-6: `WithIntrospection` dead option
**File**: `business/sdk/tablebuilder/store.go`
**Problem**: `WithIntrospection` stores `introspectionbus.Business` but no code in `store.go` ever reads `s.introspection`. The option is wired but does nothing.
**Severity**: LOW — dead code, not a functional bug
**Fix**: Either implement introspection usage or remove the option.

### BUG-7: `CreatePageConfig`/`UpdatePageConfig` silently overwrites UserID
**File**: `business/sdk/tablebuilder/configstore.go`
**Problem**: When `IsDefault: true`, the code zeroes `UserID` without returning an error or warning. A caller passing a real `UserID` with `IsDefault: true` will have their UserID silently discarded.
**Severity**: HIGH — data integrity bug
**Fix**: Either document this as intended behavior (with a comment), or return an error if both `IsDefault: true` and `UserID != uuid.Nil` are set.

### BUG-8: `replaceTernary` breaks on colons in string literals
**File**: `business/sdk/tablebuilder/evaluator.go`
**Problem**: `replaceTernary` uses `strings.Split(rest, ":")` which will split incorrectly on any ternary whose false-branch contains a colon (e.g., time strings `"12:00"`, or label strings `"Active: Yes"`).
**Severity**: MEDIUM — expression evaluation fails silently for affected patterns
**Fix**: Use a smarter split that respects quoted strings.

### BUG-9: Two duplicate `collectExemptColumns` methods
**File**: `business/sdk/tablebuilder/model.go` and `validation.go`
**Problem**: `collectExemptColumnsForValidation` (model.go) and `collectExemptColumns` (validation.go) implement the same logic independently. If one is updated without the other, `Validate()` and `ValidateConfig()` will disagree on which columns are exempt.
**Severity**: LOW now, HIGH in future — divergence risk
**Fix**: Deduplicate to one shared method.

---

## Phase Plan

### Phase 1 — Bug Fixes (Pre-Test)
**Goal**: Fix confirmed bugs before writing tests, so tests validate correct behavior.

Bugs to fix:
- [ ] BUG-1: `is_null`/`is_not_null` filter nil-guard ordering
- [ ] BUG-3: `BuildCountQuery` ForeignTable joins
- [ ] BUG-7: `CreatePageConfig` silent UserID overwrite
- [ ] BUG-8: `replaceTernary` colon in false-branch
- [ ] BUG-9: Deduplicate `collectExemptColumns`

Deferred (need more investigation):
- BUG-2: `Sort.Priority` — decide: fix or deprecate
- BUG-4: `executeRPC` — check if any configs use it first
- BUG-5: Join condition injection — assess real risk
- BUG-6: `WithIntrospection` dead code — remove or implement

Files to modify:
- `business/sdk/tablebuilder/store.go`
- `business/sdk/tablebuilder/configstore.go`
- `business/sdk/tablebuilder/evaluator.go`
- `business/sdk/tablebuilder/model.go` or `validation.go`

Tests to run after:
```bash
go test ./business/sdk/tablebuilder/...
```

---

### Phase 2 — SQL Generation Unit Tests
**Goal**: Add isolated unit tests for `store.go`/`builder.go` SQL generation without requiring a live DB.

**Approach**: Test `QueryBuilder.BuildQuery()` directly and inspect the generated SQL string. Use table-driven tests.

#### Filter tests (HIGH priority)
Test every operator in `buildFilterExpression`:
- `eq`, `neq`, `gt`, `gte`, `lt`, `lte`
- `in` (slice), `in` (scalar — different SQL)
- `like`, `ilike` (verify `%` wrapping)
- `is_null` (after BUG-1 fix — verify filter IS applied)
- `is_not_null` (same)
- Unknown operator → `eq` fallback
- Dynamic filter: key present in `params.Dynamic`
- Dynamic filter: key absent → falls back to static Value
- Combined static + dynamic filters

#### JOIN tests (HIGH priority)
- `BuildJoinCondition` with `=` separator
- `BuildJoinCondition` without `=` → raw SQL fallback
- ForeignTable: `table.column` RelationshipFrom/To
- ForeignTable: bare column (fallback paths)
- ForeignTable: Schema + Alias
- ForeignTable: JoinType `left`/`right`/`full`/`inner`
- Nested ForeignTables (depth 2+)

#### Sort tests (MEDIUM priority)
- `params.Sort` overrides `ds.Sort` when non-empty
- `ds.Sort` used when `params.Sort` empty
- Multiple sort columns
- `asc`/`desc` directions

#### Pagination tests (MEDIUM priority)
- `params.Rows` used when set
- `ds.Rows` used when `params.Rows = 0`
- Neither set → no LIMIT clause
- Page 1 → offset 0
- Page N → correct offset calculation
- Pagination not applied to secondary data sources

#### Metric query tests (HIGH priority)
- Each aggregate function: `sum`, `count`, `avg`, `min`, `max`, `count_distinct`
- Arithmetic expression: each operator `multiply`, `add`, `subtract`, `divide`
- `buildArithmeticExpression` with invalid column reference → blocked by whitelist

#### GroupBy SQL tests (MEDIUM priority)
- Simple column GroupBy
- GroupBy with Interval → verify `DATE_TRUNC` SQL output
- GroupBy with Expression → raw SQL allowed
- Multiple GroupBy columns (3-dimensional)

Files to create:
- `business/sdk/tablebuilder/builder_test.go` (new file)

---

### Phase 3 — ConfigStore & Metadata Unit Tests
**Goal**: Test ConfigStore protection rules and column metadata building without HTTP.

#### ConfigStore tests
- `Create` with duplicate name → `ErrDBDuplicatedEntry`
- `Delete` on system config → `ErrSystemConfigProtected`
- `Delete` on non-existent ID → `ErrNotFound`
- `CreatePageConfig` with `IsDefault: true` → UserID zeroed (after BUG-7 fix: verify error or document)
- `QueryPageContentWithChildren` nesting: verify children attached to correct parents
- `QueryPageContentWithChildren` with orphaned child → silently dropped
- `LoadConfig` with corrupt JSON → unmarshal error

#### Column metadata tests (`buildColumnMetadata`)
- Primary key detection: source `*_base` view strips suffix
- Foreign key `*_id` columns → hidden by default
- LabelColumn columns → hidden
- `VisualSettings.Hidden: true` overrides FK/PK hidden logic
- `VisualSettings.Hidden: false` un-hides FK column
- Order sorting: hasExplicitOrder → stable sort by Order, hidden at end

#### `mergeData` tests
- Missing `ParentSource` → error
- Missing `SelectBy` → error
- `parseParentKey` with `table.column` → returns `column`
- `parseParentKey` with bare column → as-is
- Match found: `type = "viewcount"` → sets count field
- No match in secondary → primary row unchanged

Files to create:
- `business/sdk/tablebuilder/configstore_test.go` (new file, requires DB)
- `business/sdk/tablebuilder/metadata_test.go` (new file)

---

### Phase 4 — Evaluator Edge Cases & Security Tests
**Goal**: Fill evaluator gaps and explicitly test security-critical whitelist enforcement at the builder level.

#### Evaluator edge cases
- Expression > 500 chars → error
- `flattenValue` with `map[string]any` → flattened with `_` separator
- `flattenValue` with `[]any` → indexed `_0`, `_1`, capped at 10
- `replaceArrayAccess` with index ≥ 10 → not replaced (document as limitation or fix)
- `replaceTernary` with colon in false-branch (after BUG-8 fix)
- `replaceTernary` with nested ternary
- Cache eviction at 100+ entries

#### Security whitelist tests at builder level
These are separate from validation tests — they verify the builder enforces whitelists even if `ValidateConfig` is bypassed:
- `AllowedAggregateFunctions` blocks unknown function in `buildMetricExpression`
- `AllowedOperators` blocks unknown operator in `buildArithmeticExpression`
- `AllowedIntervals` blocks unknown interval in `buildGroupBySelectExpression`
- `isValidColumnReference` blocks SQL injection patterns when called from `buildArithmeticExpression`
- `BuildJoinCondition` raw SQL fallback — document as approved injection surface or add sanitization

#### `collectExemptColumns` divergence test (after BUG-9 fix)
- Verify `Validate()` and `ValidateConfig()` agree on exempt columns for same config

Files to modify:
- `business/sdk/tablebuilder/evaluator_test.go` (add subtests)
- `business/sdk/tablebuilder/validation_test.go` (add security whitelist tests)

---

## File Map

```
business/sdk/tablebuilder/
├── store.go                    ← BUG-1, BUG-2, BUG-3, BUG-5, BUG-6 fixes
├── configstore.go              ← BUG-7 fix
├── evaluator.go                ← BUG-8 fix
├── model.go / validation.go    ← BUG-9 fix
│
├── builder_test.go             ← NEW: Phase 2 SQL generation tests
├── configstore_test.go         ← NEW: Phase 3 ConfigStore tests (requires DB)
├── metadata_test.go            ← NEW: Phase 3 metadata tests
├── evaluator_test.go           ← EXTEND: Phase 4 edge cases
└── validation_test.go          ← EXTEND: Phase 4 security whitelist tests
```

---

## Investigation Still Needed (Before Implementation)

1. **Does any seeded config use `type: "rpc"`?** Search `seed_tablebuilder.go` and `seed_config_pages.go` for `"rpc"`. If yes, BUG-4 is HIGH priority.

2. **Is `WithIntrospection` intentionally incomplete?** Check git log for the commit that added it — was introspection meant to be wired later?

3. **Does `executeCount` (type `"viewcount"`) work correctly?** The positional arg approach differs from the rest of the named-param code. Check if any integration tests actually exercise this path end-to-end.

4. **What format does `ds.Args` take for filter injection into count queries?** The `Args` map handling across `FetchTableData`, `GetCount`, and `FetchTableDataCount` is inconsistent.

---

## Suggested Phase Order

```
Phase 1 (Bug Fixes) → Phase 2 (SQL Unit Tests) → Phase 3 (ConfigStore/Metadata) → Phase 4 (Evaluator/Security)
```

Phase 1 must come first — tests should validate correct behavior, not document bugs.
Phases 2–4 are independent and could be parallelized with spawn-agents.

Run tests after each phase:
```bash
go test ./business/sdk/tablebuilder/...
```

For Phase 3 (ConfigStore tests), a live DB is required:
```bash
go test ./business/sdk/tablebuilder/... -run TestConfigStore
```
