# PR #124 Scenario Subsystem Retrospective Fixes

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Fix the 6 non-false-positive issues surfaced by the retrospective review of PR #124 — runtime correctness bugs in the scenario subsystem, plus test-infrastructure gaps flagged against `docs/arch/domain-template.md` and `docs/arch/seeding.md`.

**Architecture:** The scenario subsystem filters and tags rows by a ctx-scoped `scenario_id`. A single helper (`sqldb.ApplyScenarioFilter`) is called inside 55 production `Query`/`QueryByID`/`Count` methods. The critical bug is a detection flaw in that helper that breaks every `QueryByID` when a scenario is active. Fixes are targeted: a helper-internal change for the critical bug (no call-site ripple), three one-line additions for missing `Count` filters, a one-line rollback, a doc-comment clarification, and new test files mirroring the `labels` domain pattern.

**Tech Stack:** Go 1.23, PostgreSQL 16.4 (multi-schema), `github.com/jmoiron/sqlx` via `business/sdk/sqldb`, integration tests via `business/sdk/dbtest` + `api/sdk/http/apitest`.

---

## Phases

- **Phase 1 — Runtime correctness** (issues #1, #2, #5): 4 tasks. Low LOC, high signal. Merge first.
- **Phase 2 — Convention alignment** (issue #6): 1 task. Doc-comment only; no behavior change.
- **Phase 3 — Test infrastructure** (issues #4, #3): 2 tasks. Substantially larger; can be deferred to a follow-up branch.

## File Structure

**Phase 1 — modify:**
- `business/sdk/sqldb/scenariofilter.go` — fix `WHERE` detection (the critical bug)
- `business/sdk/sqldb/scenariofilter_test.go` — add regression tests for multi-line SQL
- `business/domain/inventory/transferorderbus/stores/transferorderdb/transferorderdb.go` — add filter to `Count`
- `business/domain/sales/orderfulfillmentstatusbus/stores/orderfulfillmentstatusdb/orderfulfillmentstatusdb.go` — add filter to `Count`
- `business/domain/sales/orderlineitemsbus/stores/orderlineitemsdb/orderlineitemsdb.go` — add filter to `Count`
- `business/domain/scenarios/scenariobus/scenariobus.go` — rollback on commit failure in `Load`

**Phase 2 — modify:**
- `business/sdk/sqldb/scenariofilter.go` — clarify `GetScenarioFilter` doc comment

**Phase 3 — create:**
- `business/domain/scenarios/scenariobus/testutil.go` — `TestSeedScenarios` helper
- `api/cmd/services/ichor/tests/scenarios/scenarioapi/main_test.go` — `Test_Scenarios` entry point
- `api/cmd/services/ichor/tests/scenarios/scenarioapi/seed_test.go` — `insertSeedData`
- `api/cmd/services/ichor/tests/scenarios/scenarioapi/query_test.go` — GET happy/auth cases
- `api/cmd/services/ichor/tests/scenarios/scenarioapi/active_test.go` — `/scenarios/active` smoke test

---

## Phase 1 — Runtime Correctness

### Task 1: Fix `ApplyScenarioFilter` WHERE detection (issue #1, score 92)

**Files:**
- Modify: `business/sdk/sqldb/scenariofilter.go`
- Test: `business/sdk/sqldb/scenariofilter_test.go`

**Root cause:** Detection uses `strings.Contains(buf.String(), " WHERE ")` (space-padded). Every `QueryByID` across 18 stores builds SQL as `"\n    WHERE\n        id = :id"` (whitespace-bounded). Detection returns `false`, helper emits a second ` WHERE (...)`, producing invalid double-`WHERE` SQL whenever a scenario is active.

**Fix:** Replace the space-padded substring check with a `\b`-bounded regex. This is internal to the helper — the public signature stays stable and no caller needs to change.

- [ ] **Step 1: Write the failing test**

Append to `business/sdk/sqldb/scenariofilter_test.go`:

```go
func TestApplyScenarioFilter_MultilineWhere(t *testing.T) {
	ctx := sqldb.SetScenarioFilter(context.Background(), uuid.New())

	// SQL formatted like real QueryByID: WHERE bounded by newlines/tabs, not spaces.
	const q = `
    SELECT id, product_id
    FROM inventory.inventory_items
    WHERE
        id = :id
    `
	buf := bytes.NewBufferString(q)
	data := map[string]any{"id": "abc"}

	sqldb.ApplyScenarioFilter(ctx, buf, data)

	got := buf.String()
	if strings.Count(strings.ToUpper(got), "WHERE") != 1 {
		t.Fatalf("expected exactly one WHERE clause, got:\n%s", got)
	}
	if !strings.Contains(got, " AND (scenario_id IS NULL OR scenario_id = :scenario_id)") {
		t.Fatalf("expected AND-prefixed scenario clause, got:\n%s", got)
	}
}
```

- [ ] **Step 2: Run test and confirm it fails**

```bash
go test ./business/sdk/sqldb/ -run TestApplyScenarioFilter_MultilineWhere -v
```
Expected: FAIL with "expected exactly one WHERE clause" — the buffer contains two `WHERE` keywords.

- [ ] **Step 3: Fix the detection**

In `business/sdk/sqldb/scenariofilter.go`, add a package-level regex near the top of the file (after imports):

```go
// hasWhereRe matches a WHERE keyword bounded by non-identifier characters.
// This handles multi-line SQL where WHERE is surrounded by newlines/tabs
// rather than spaces — the case that broke the prior substring check.
var hasWhereRe = regexp.MustCompile(`(?i)\bWHERE\b`)
```

Add `"regexp"` to the imports.

Then replace lines 53-57 (the `strings.Contains` branch):

```go
	if hasWhereRe.MatchString(buf.String()) {
		buf.WriteString(" AND (scenario_id IS NULL OR scenario_id = :scenario_id)")
		return
	}
	buf.WriteString(" WHERE (scenario_id IS NULL OR scenario_id = :scenario_id)")
```

- [ ] **Step 4: Run the new test and the existing suite**

```bash
go test ./business/sdk/sqldb/ -v
```
Expected: PASS for `TestApplyScenarioFilter_MultilineWhere` and all pre-existing tests in that package.

- [ ] **Step 5: Build the package**

```bash
go build ./business/sdk/sqldb/...
```
Expected: no output (success).

- [ ] **Step 6: Commit**

```bash
git add business/sdk/sqldb/scenariofilter.go business/sdk/sqldb/scenariofilter_test.go
git commit -m "fix(sqldb): detect WHERE as keyword, not space-padded substring

The prior strings.Contains(\" WHERE \") check missed the multi-line SQL
used in every QueryByID (WHERE bounded by newlines), causing
ApplyScenarioFilter to emit a second WHERE clause when a scenario was
active — invalid SQL on every by-ID fetch across 18 scoped stores.

Replace with a regexp.\\b WHERE \\b match that is whitespace- and
case-tolerant. Signature unchanged; no call-site ripple."
```

---

### Task 2: Add `ApplyScenarioFilter` to the 3 `Count` methods that miss it (issue #2, score 75)

**Files:**
- Modify: `business/domain/inventory/transferorderbus/stores/transferorderdb/transferorderdb.go:~172`
- Modify: `business/domain/sales/orderfulfillmentstatusbus/stores/orderfulfillmentstatusdb/orderfulfillmentstatusdb.go:~152`
- Modify: `business/domain/sales/orderlineitemsbus/stores/orderlineitemsdb/orderlineitemsdb.go:~168`

**Root cause:** `Query()` in each of these stores applies the scenario filter, but `Count()` does not. When a scenario is active, pagination math divides by a global row total while the page payload is scenario-scoped. Reference for the correct pattern: `inventoryitemdb.go:381-398` (Count applies the filter before `NamedQueryStruct`).

**Approach:** Each `Count` currently builds a query directly into a `q` variable and calls `NamedQueryStruct(ctx, s.log, s.db, q, data, &count)`. We need to convert `q` into a `*bytes.Buffer`, apply the filter, then dispatch. Use the pattern from `inventoryitemdb.Count`.

- [ ] **Step 1: Read the current `Count` method in each store**

Get current shape for each file:

```bash
sed -n '159,180p' business/domain/inventory/transferorderbus/stores/transferorderdb/transferorderdb.go
sed -n '139,160p' business/domain/sales/orderfulfillmentstatusbus/stores/orderfulfillmentstatusdb/orderfulfillmentstatusdb.go
sed -n '155,175p' business/domain/sales/orderlineitemsbus/stores/orderlineitemsdb/orderlineitemsdb.go
```

Each should look like:
```go
func (s *Store) Count(ctx context.Context, filter ...QueryFilter) (int, error) {
	data := map[string]any{}
	const q = `SELECT COUNT(*) AS count FROM ...`
	buf := bytes.NewBufferString(q)
	s.applyFilter(filter, data, buf)

	var count struct{ Count int `db:"count"` }
	if err := sqldb.NamedQueryStruct(ctx, s.log, s.db, q, data, &count); err != nil {
		return 0, fmt.Errorf("namedquerystruct: %w", err)
	}
	return count.Count, nil
}
```

- [ ] **Step 2: Edit `transferorderdb.go` Count method**

Insert `sqldb.ApplyScenarioFilter(ctx, buf, data)` immediately after `s.applyFilter(...)` and change the `NamedQueryStruct` call to use `buf.String()` instead of `q`. Exact edit — find this block:

```go
	s.applyFilter(filter, data, buf)

	var count struct {
		Count int `db:"count"`
	}
	if err := sqldb.NamedQueryStruct(ctx, s.log, s.db, q, data, &count); err != nil {
```

Replace with:

```go
	s.applyFilter(filter, data, buf)
	sqldb.ApplyScenarioFilter(ctx, buf, data)

	var count struct {
		Count int `db:"count"`
	}
	if err := sqldb.NamedQueryStruct(ctx, s.log, s.db, buf.String(), data, &count); err != nil {
```

- [ ] **Step 3: Apply the same edit to `orderfulfillmentstatusdb.go` and `orderlineitemsdb.go`**

Exact same two-line change: insert the `ApplyScenarioFilter` call after `applyFilter`, and swap `q` for `buf.String()` in the `NamedQueryStruct` call. Verify the surrounding code actually uses `buf := bytes.NewBufferString(q)` — if any store passes `q` directly without a buffer, first wrap it in a buffer.

- [ ] **Step 4: Build the affected packages**

```bash
go build ./business/domain/inventory/transferorderbus/... \
         ./business/domain/sales/orderfulfillmentstatusbus/... \
         ./business/domain/sales/orderlineitemsbus/...
```
Expected: no output.

- [ ] **Step 5: Run targeted unit tests if present**

```bash
go test ./business/domain/inventory/transferorderbus/... \
        ./business/domain/sales/orderfulfillmentstatusbus/... \
        ./business/domain/sales/orderlineitemsbus/...
```
Expected: PASS. If any test uses a DB container and can't run locally, note that CI will verify.

- [ ] **Step 6: Commit**

```bash
git add business/domain/inventory/transferorderbus/stores/transferorderdb/transferorderdb.go \
        business/domain/sales/orderfulfillmentstatusbus/stores/orderfulfillmentstatusdb/orderfulfillmentstatusdb.go \
        business/domain/sales/orderlineitemsbus/stores/orderlineitemsdb/orderlineitemsdb.go
git commit -m "fix(stores): apply scenario filter in Count for 3 missed stores

Query() in transferorderdb, orderfulfillmentstatusdb, and orderlineitemsdb
all applied ApplyScenarioFilter, but their Count() methods did not. With
a scenario active, paginators divided page size by a global row total
while the page payload was scenario-scoped — off-by-N pagination.

Mirrors the pattern already used by inventoryitemdb.Count()."
```

---

### Task 3: Roll back transaction on `Load` commit failure (issue #5, score 65)

**Files:**
- Modify: `business/domain/scenarios/scenariobus/scenariobus.go:295-297`

**Root cause:** Every intermediate-error branch in `Load` explicitly calls `tx.Rollback()`, but the `tx.Commit()` failure path just returns the error. `database/sql` auto-releases the conn on commit failure, so this is not a resource leak — but it is inconsistent with the surrounding code and leaves a path where no `Rollback` is logged. `Reset` delegates to `Load` so fixing `Load` covers both.

- [ ] **Step 1: Edit the commit block**

In `business/domain/scenarios/scenariobus/scenariobus.go`, find:

```go
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("load commit: %w", err)
	}
```

Replace with:

```go
	if err := tx.Commit(); err != nil {
		tx.Rollback()
		return fmt.Errorf("load commit: %w", err)
	}
```

(`Rollback` after a failed `Commit` returns `sql.ErrTxDone` as a no-op in `database/sql`; the call is safe and documents intent.)

- [ ] **Step 2: Build**

```bash
go build ./business/domain/scenarios/...
```
Expected: no output.

- [ ] **Step 3: Commit**

```bash
git add business/domain/scenarios/scenariobus/scenariobus.go
git commit -m "fix(scenarios): rollback transaction on Load commit failure

Every intermediate-error branch in Load already calls tx.Rollback(); the
commit-failure branch was the only exception. database/sql makes this a
no-op in practice, but the inconsistency obscured intent and left the
commit path uncovered by the repo's rollback-on-error convention."
```

---

## Phase 2 — Convention Alignment

### Task 4: Clarify `GetScenarioFilter` doc comment (issue #6, score 50)

**Files:**
- Modify: `business/sdk/sqldb/scenariofilter.go:23-26`

**Root cause:** `GetScenarioFilter`'s doc comment says it's for "callers that need to TAG writes (not filter reads)." But `inventoryitemdb` uses it directly to filter reads in 4 functions (`QueryWithLocationDetails`, `QueryItemsWithProductAtLocation`, `QueryAvailableForAllocation`, `queryFEFO`) — because those queries use table aliases (`ii.scenario_id`) that `ApplyScenarioFilter` can't emit. The deviation is pragmatic; the right fix is to update the comment to acknowledge the carve-out, not to rewrite the call sites (which would require a larger helper API).

- [ ] **Step 1: Update the comment**

Replace the existing `GetScenarioFilter` doc comment with:

```go
// GetScenarioFilter returns the active scenario id from ctx and a bool
// indicating presence. A zero uuid.UUID is treated as absent.
//
// Typical use: write-tagging, e.g. bus.Create populating ScenarioID on
// new rows. Aliased-read queries (where ApplyScenarioFilter's unaliased
// scenario_id column would be ambiguous — e.g. inventoryitemdb multi-join
// reads) also call this directly and hand-build their alias-qualified
// filter. Prefer ApplyScenarioFilter for single-table reads.
```

- [ ] **Step 2: Commit**

```bash
git add business/sdk/sqldb/scenariofilter.go
git commit -m "docs(sqldb): acknowledge aliased-read carve-out for GetScenarioFilter

The comment previously reserved GetScenarioFilter for write-tagging only,
but inventoryitemdb reads call it directly because ApplyScenarioFilter
hard-codes an unaliased scenario_id column that is ambiguous in
multi-join queries. Document the pattern rather than forcing a helper
redesign."
```

---

## Phase 3 — Test Infrastructure (optional, larger scope)

> The two tasks below bring the scenario subsystem into compliance with the 7-layer checklist in `docs/arch/domain-template.md` and the `testutil.go` convention in `docs/arch/seeding.md`. Neither blocks runtime correctness, and both scored 75 (just under the review threshold). Execute if you want full convention compliance; defer if you want to ship Phase 1 + 2 first.

### Task 5: Create `scenariobus/testutil.go` with `TestSeedScenarios` (issue #4, score 75)

**Files:**
- Create: `business/domain/scenarios/scenariobus/testutil.go`

**Reference pattern:** `business/domain/labels/labelbus/testutil.go` (PR #122).

- [ ] **Step 1: Create the file**

```go
package scenariobus

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

// TestNewScenarios generates n NewScenario values with deterministic names.
func TestNewScenarios(n int) []NewScenario {
	scenarios := make([]NewScenario, n)
	for i := range n {
		scenarios[i] = NewScenario{
			ID:          uuid.New(),
			Name:        fmt.Sprintf("test-scenario-%04d", i+1),
			Description: fmt.Sprintf("Test scenario %d", i+1),
		}
	}
	return scenarios
}

// TestSeedScenarios inserts n scenarios via the bus and returns them. The
// results are unsorted; callers that need deterministic order should sort
// by ID. Intended for integration tests that need scenario rows present
// without going through the YAML loader.
func TestSeedScenarios(ctx context.Context, n int, api *Business) ([]Scenario, error) {
	newScenarios := TestNewScenarios(n)

	scenarios := make([]Scenario, len(newScenarios))
	for i, ns := range newScenarios {
		s, err := api.Create(ctx, ns)
		if err != nil {
			return nil, fmt.Errorf("seeding scenario %d: %w", i, err)
		}
		scenarios[i] = s
	}
	return scenarios, nil
}
```

> **Verify before writing:** Read `business/domain/scenarios/scenariobus/model.go` to confirm the exact field names on `NewScenario` (ID, Name, Description). If fields differ, adjust the literal above to match — the structure stays the same.

- [ ] **Step 2: Build**

```bash
go build ./business/domain/scenarios/...
```
Expected: no output.

- [ ] **Step 3: Commit**

```bash
git add business/domain/scenarios/scenariobus/testutil.go
git commit -m "test(scenarios): add testutil.go with TestSeedScenarios helper

Per docs/arch/seeding.md, TestSeed* helpers live in {entity}bus/testutil.go
so integration tests can seed without circular imports. Mirrors
labelbus/testutil.go exactly. Enables scenarioapi integration tests
(Task 6)."
```

---

### Task 6: Minimal `scenarioapi` integration tests (issue #3, score 75)

**Files:**
- Create: `api/cmd/services/ichor/tests/scenarios/scenarioapi/main_test.go`
- Create: `api/cmd/services/ichor/tests/scenarios/scenarioapi/seed_test.go`
- Create: `api/cmd/services/ichor/tests/scenarios/scenarioapi/query_test.go`
- Create: `api/cmd/services/ichor/tests/scenarios/scenarioapi/active_test.go`

**Reference pattern:** `api/cmd/services/ichor/tests/labels/labelapi/` (PR #122).

**Scope for retrospective:** Cover the read-only happy paths and a 401 per endpoint family. `Load`/`Reset`/`Create`/`Update`/`Delete` cases require scenario-fixture YAML plumbing and admin/table-access role flips — defer those to a follow-up PR. The goal here is to (a) satisfy the 7-layer convention, (b) exercise the middleware's ctx-propagation of `scenario_id`, (c) give future contributors a foothold.

- [ ] **Step 1: Write `main_test.go`**

```go
package scenarioapi_test

import (
	"testing"

	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
)

func Test_Scenarios(t *testing.T) {
	t.Parallel()

	test := apitest.StartTest(t, "Test_Scenarios")

	sd, err := insertSeedData(test.DB, test.Auth)
	if err != nil {
		t.Fatalf("Seeding error: %s", err)
	}

	test.Run(t, query200(sd), "query-200")
	test.Run(t, query401(sd), "query-401")
	test.Run(t, queryByID200(sd), "query-by-id-200")
	test.Run(t, queryByID401(sd), "query-by-id-401")
	test.Run(t, queryByID404(sd), "query-by-id-404")
	test.Run(t, active200(sd), "active-200")
	test.Run(t, active401(sd), "active-401")
}
```

- [ ] **Step 2: Write `seed_test.go`**

Mirror `labelapi/seed_test.go`, substituting `TestSeedScenarios` for `TestSeedLabels` and `inventory.scenarios` for `core.label_catalog` in the `table_access` downgrade loop. Include the `SeedData` struct with `Scenarios []scenariobus.Scenario` and the standard admin / regular-user setup.

> **Verify first:** Read `labelapi/seed_test.go` end-to-end and mirror it. Adjust only the resource-specific bits (domain, table name, seed helper).

- [ ] **Step 3: Write `query_test.go`**

Create five `apitest.Table` definitions covering: `query200`, `query401`, `queryByID200`, `queryByID401`, `queryByID404`. Each returns an `[]apitest.Table` slice. Follow the shape in `labelapi/query_test.go` exactly — the only differences are:
- URL: `/v1/scenarios`
- Response type: `scenarioapp.Scenario` (confirm exact type name in `app/domain/scenarios/scenarioapp/model.go`)
- Seed data: `sd.Scenarios` from `SeedData`

> **Verify first:** Open `labelapi/query_test.go` and confirm the return-type contract before writing.

- [ ] **Step 4: Write `active_test.go`**

Two cases — `active200` and `active401`. Hit `GET /v1/scenarios/active` with and without a valid token. Without any scenario loaded, `active` should return a 200 with an empty/null body (or however the handler indicates no-active). Confirm the exact shape by reading the handler at `api/domain/http/scenarios/scenarioapi/`.

- [ ] **Step 5: Build + run**

```bash
go build ./api/cmd/services/ichor/tests/scenarios/...
go test ./api/cmd/services/ichor/tests/scenarios/scenarioapi/ -v -run Test_Scenarios
```
Expected: PASS on all 7 sub-tests. If any fail, fix incrementally — read the error, check the referenced handler's response shape, adjust.

- [ ] **Step 6: Commit**

```bash
git add api/cmd/services/ichor/tests/scenarios/
git commit -m "test(scenarios): add minimal scenarioapi integration tests

Covers happy + 401 cases for Query, QueryByID, and /scenarios/active to
satisfy the 7-layer checklist in docs/arch/domain-template.md. Defers
Create/Update/Delete/Load/Reset cases to a follow-up — those require
scenario-fixture YAML plumbing and are out of scope for the retrospective
fix batch."
```

---

## Self-Review

**Spec coverage:**
- Issue #1 (score 92) → Task 1 ✓
- Issue #2 (score 75) → Task 2 ✓
- Issue #5 (score 65) → Task 3 ✓
- Issue #6 (score 50) → Task 4 ✓
- Issue #4 (score 75) → Task 5 ✓
- Issue #3 (score 75) → Task 6 ✓

**Placeholders:** Task 6 Steps 2-4 reference `labelapi`/`scenarioapp` shapes instead of inlining full test code. This is justified because (a) the labelapi files are ~80-200 lines each and duplicating them in the plan would be larger than the code itself, and (b) the "mirror `labelapi/X`" instruction plus an explicit `Verify first:` step pins the reader to a concrete source. The critical path fixes (Tasks 1-3) do inline all code.

**Type consistency:** `TestSeedScenarios` signature in Task 5 matches the call shape implied by Task 6's seed_test.go. Field names in `NewScenario` are explicitly flagged as needing verification against `model.go` before writing.

---

## Execution notes

- Each task commits independently — frequent commits per global CLAUDE.md.
- Commit messages use conventional prefix (`fix`, `docs`, `test`) per the repo convention (see `git log --oneline` for recent examples).
- No migration changes are required — all column and table work was shipped in PR #124 and is not being modified.
- `go test ./...` is explicitly forbidden per project CLAUDE.md. Each task runs targeted tests for affected packages only.
