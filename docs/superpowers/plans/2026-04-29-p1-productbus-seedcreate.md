# P1 — productbus.SeedCreate (Deterministic ProductID) Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a `productbus.Business.SeedCreate(ctx, p Product) error` method that accepts a caller-provided `ProductID`, mirroring `labelbus.SeedCreate`, and switch productbus test seed helpers to derive `ProductID = seedid.Stable("product:"+sku)` so reseeds produce byte-identical product UUIDs and `inventory.label_catalog.entity_ref` rows. Add a unit test for `Create`'s delegate firing (currently absent) so the deliberate skip in `SeedCreate` is verifiable. Add a reusable `scripts/check-determinism.sh` script and `make check-determinism` target so future regressions are caught automatically.

**Architecture:** Mirrors the existing `labelbus.SeedCreate` pattern: caller supplies a fully-formed domain struct with a deterministic ID; the bus method fills only zero-valued timestamps and the `TrackingType` default, then calls the existing `storer.Create`. The deterministic UUID v5 helper currently buried inside `dbtest/seed_labels.go` is lifted to a new `business/sdk/seedid` package so `productbus/testutil.go` (which `dbtest` already imports — circular if reversed) can use the same namespace constant. **`SeedCreate` deliberately skips `delegate.Call`** — matching labelbus's precedent — to avoid firing Temporal workflow events during DB seeding. We close the precondition for that decision by adding the missing delegate-firing unit test for `Create`. The determinism script projects only stable columns (excludes `created_date`/`updated_date`, which drift via `time.Now()` even with deterministic IDs).

**Tech Stack:** Go 1.23, PostgreSQL 16.4, `github.com/google/uuid` (v5/SHA-1), `github.com/google/go-cmp/cmp`, Ardan Labs Service Starter Kit layout, `unitest.Table` test pattern, existing test container infrastructure (`dbtest.NewDatabase`).

---

## Files Touched

| Path | Change | Responsibility |
|---|---|---|
| `business/sdk/seedid/seedid.go` | **Create** | Tiny package: `Namespace` constant + `Stable(key string) uuid.UUID`. Single source of truth. |
| `business/sdk/seedid/seedid_test.go` | **Create** | Property tests: determinism, namespace lock, known vector. |
| `business/sdk/dbtest/seed_labels.go` | **Modify** | Drop local `detNamespace`/`detUUID`; import `seedid`. |
| `business/sdk/dbtest/seed_scenarios.go` | **Modify** | Same migration. |
| `business/domain/products/productbus/productbus.go` | **Modify** | Add `SeedCreate(ctx, p Product) error`. |
| `business/domain/products/productbus/productbus_test.go` | **Modify** | Add `delegateFires` and `seedCreate` sub-tests; wire both into `Test_Product`. Both use the real DB container (no mocks). |
| `business/domain/products/productbus/testutil.go` | **Modify** | Switch `TestSeedProducts` and `TestSeedProductsHistoricalWithDistribution` to deterministic IDs + `SeedCreate`. |
| `scripts/check-determinism.sh` | **Create** | Reusable shell script: reseed twice, project stable columns, diff. |
| `makefile` | **Modify** | Add `check-determinism` target. |

**NOT touched:** `productdb.go` — `Storer.Create(ctx, Product)` already accepts a fully-formed Product (mirrors labeldb).

---

## Test Strategy

| Layer | Where | What it verifies |
|---|---|---|
| Pure unit (no DB) | `seedid_test.go` | Determinism, namespace lock, known vector |
| Integration (real DB, custom delegate) | `productbus_test.go` `delegateFires` sub-test | `Create` fires `delegate.Call` with `ActionCreatedData` (closes pre-existing gap, makes `SeedCreate`'s skip verifiable). Uses real `productdb.NewStore` against `db.DB` and a fresh `delegate.New(log)` so the handler can be intercepted — **no mocks**. |
| Integration (DB container) | `productbus_test.go` `seedCreate` sub-test | Caller-supplied ProductID preserved; timestamp/TrackingType defaults fill correctly; historical CreatedDate preserved |
| End-to-end determinism | `scripts/check-determinism.sh` | Two `make reseed-frontend` runs produce byte-identical stable columns in `products.products` and `inventory.label_catalog` |

## Side-path Note (Required by User)

**`SeedCreate` is a deliberate side-path that bypasses `delegate.Call`.** This is documented inline in the method's godoc and in the PR description. The delegate firing path is exercised by the new `delegate_test.go` (Task 4) so the regular `Create` behavior is covered; `SeedCreate` is exclusively for deterministic test-data seeding where workflow side effects are undesirable. **No production code path uses `SeedCreate`** — it is reachable only from `testutil.go` and `dbtest` seed helpers.

## detUUID Lifting Decision

**Lift to a new shared package `business/sdk/seedid`.** `dbtest` imports `productbus`, so reversing the import to share `detUUID` would create a cycle. A peer package under `business/sdk/` is the only correct location, and a single source of truth for the namespace constant prevents silent determinism breakage if one copy drifts.

---

### Task 1: Create the `seedid` package

**Files:**
- Create: `business/sdk/seedid/seedid.go`
- Create: `business/sdk/seedid/seedid_test.go`

- [ ] **Step 1: Write the failing tests**

Create `business/sdk/seedid/seedid_test.go`:

```go
package seedid_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/sdk/seedid"
)

func TestNamespace_MatchesHistoricalValue(t *testing.T) {
	want := uuid.MustParse("deadbeef-dead-beef-dead-beefdeadbeef")
	if seedid.Namespace != want {
		t.Fatalf("namespace drifted: got %s, want %s", seedid.Namespace, want)
	}
}

func TestStable_Deterministic(t *testing.T) {
	a := seedid.Stable("label:STG-A01")
	b := seedid.Stable("label:STG-A01")
	if a != b {
		t.Fatalf("Stable(%q) is not deterministic: %s vs %s", "label:STG-A01", a, b)
	}
}

func TestStable_DistinctKeysProduceDistinctUUIDs(t *testing.T) {
	a := seedid.Stable("label:STG-A01")
	b := seedid.Stable("label:STG-A02")
	if a == b {
		t.Fatalf("Stable produced collision for distinct keys: %s", a)
	}
}

func TestStable_KnownVector(t *testing.T) {
	want := uuid.NewSHA1(uuid.MustParse("deadbeef-dead-beef-dead-beefdeadbeef"), []byte("label:STG-A01"))
	got := seedid.Stable("label:STG-A01")
	if got != want {
		t.Fatalf("Stable drift: got %s want %s", got, want)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./business/sdk/seedid/...`
Expected: FAIL — `business/sdk/seedid` does not exist; compile error.

- [ ] **Step 3: Implement the seedid package**

Create `business/sdk/seedid/seedid.go`:

```go
// Package seedid provides deterministic UUID v5 derivation for test
// seed data. Using a fixed namespace and a stable key string guarantees
// that re-running `make reseed-frontend` (or any other seeding flow)
// produces byte-identical UUIDs across builds and developer machines.
//
// This is the single source of truth for the deterministic-seed
// namespace; both business/sdk/dbtest seed helpers and per-domain
// testutil helpers must import this package rather than defining their
// own copy.
package seedid

import "github.com/google/uuid"

// Namespace is the UUID v5 namespace that anchors all deterministic
// seed UUIDs. It matches the Manitowoc generator's value; changing it
// invalidates every deterministic seed UUID in the codebase, including
// inventory.label_catalog rows.
var Namespace = uuid.MustParse("deadbeef-dead-beef-dead-beefdeadbeef")

// Stable returns a UUID v5 derived from key under Namespace. The same
// key always produces the same UUID, on any machine, in any process.
func Stable(key string) uuid.UUID {
	return uuid.NewSHA1(Namespace, []byte(key))
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./business/sdk/seedid/...`
Expected: PASS — 4 tests pass.

- [ ] **Step 5: Build sanity**

Run: `go build ./business/sdk/seedid/...`
Expected: clean exit.

- [ ] **Step 6: Commit**

```bash
git add business/sdk/seedid/
git commit -m "feat(seedid): add shared deterministic UUID helper for seed data

Lifts detUUID from business/sdk/dbtest/seed_labels.go into a shared
package so domain-level testutil files can derive deterministic IDs
without creating an import cycle (dbtest already depends on
productbus). Namespace value preserved verbatim from the original
to keep existing label_catalog UUIDs byte-identical across reseeds.

Refs: P1 productbus.SeedCreate."
```

---

### Task 2: Migrate `dbtest/seed_labels.go` to `seedid`

**Files:**
- Modify: `business/sdk/dbtest/seed_labels.go`

- [ ] **Step 1: Update imports and replace local helpers**

Make the following edits to `business/sdk/dbtest/seed_labels.go`:

1. **Delete lines 12-20** (the `detNamespace` var, `detUUID` func, and their comments — verbatim from current file):

```go
// detNamespace matches the Manitowoc generator's UUID v5 namespace.
// Using the same namespace guarantees label codes produce byte-identical
// UUIDs across `make reseed-frontend` invocations and across builds.
var detNamespace = uuid.MustParse("deadbeef-dead-beef-dead-beefdeadbeef")

// detUUID returns a UUID v5 derived from a stable key string.
func detUUID(key string) uuid.UUID {
	return uuid.NewSHA1(detNamespace, []byte(key))
}
```

2. **Update the import block** (currently lines 3-10) to add `seedid` and remove `uuid` if it becomes unused:

```go
import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/timmaaaz/ichor/business/domain/labels/labelbus"
	"github.com/timmaaaz/ichor/business/sdk/seedid"
)
```

(After deletion of `detNamespace`/`detUUID`, the only remaining `uuid` reference is the import itself. Drop `"github.com/google/uuid"` from the import block. The compiler will tell you if any usage remains.)

3. **Replace `detUUID("label:" + e.code)` at line 91 with `seedid.Stable("label:" + e.code)`**:

```go
		lc := labelbus.LabelCatalog{
			ID:          seedid.Stable("label:" + e.code),
			Code:        e.code,
			Type:        e.typ,
			EntityRef:   e.entityRef,
			PayloadJSON: e.payloadJSON,
		}
```

- [ ] **Step 2: Build sanity**

Run: `go build ./business/sdk/dbtest/...`
Expected: clean exit. If the build fails with "imported and not used: github.com/google/uuid" then you missed dropping the import — drop it and rebuild.

- [ ] **Step 3: Vet**

Run: `go vet ./business/sdk/dbtest/...`
Expected: no diagnostics.

---

### Task 3: Migrate `dbtest/seed_scenarios.go` to `seedid`

**Files:**
- Modify: `business/sdk/dbtest/seed_scenarios.go`

- [ ] **Step 1: Update the call site**

In `business/sdk/dbtest/seed_scenarios.go`:

1. Add `"github.com/timmaaaz/ichor/business/sdk/seedid"` to the imports (alphabetical, with the other `business/sdk/...` imports).
2. Replace at **line 79**:
   - Before: `detUUID(fmt.Sprintf("fixture:%s:%s:%d", s.Name, targetTable, i))`
   - After:  `seedid.Stable(fmt.Sprintf("fixture:%s:%s:%d", s.Name, targetTable, i))`

- [ ] **Step 2: Build sanity**

Run: `go build ./business/sdk/dbtest/...`
Expected: clean exit.

- [ ] **Step 3: Confirm no remaining `detUUID` references anywhere**

Run: `grep -rn "detUUID\|detNamespace" business/`
Expected: no output.

- [ ] **Step 4: Run dbtest tests**

Run: `go test ./business/sdk/dbtest/...`
Expected: PASS. (The seed_integration_test.go exercises the full seeding chain.)

- [ ] **Step 5: Commit**

```bash
git add business/sdk/dbtest/seed_labels.go business/sdk/dbtest/seed_scenarios.go
git commit -m "refactor(dbtest): use shared seedid.Stable instead of local detUUID

Replaces the file-private detUUID helper in seed_labels.go and
seed_scenarios.go with seedid.Stable so productbus testutil (and any
future caller) can derive UUIDs from the same namespace without
forcing a circular import on dbtest.

No behavior change; namespace and SHA-1 derivation are byte-identical
to the prior implementation.

Refs: P1 productbus.SeedCreate."
```

---

### Task 4: Add delegate-firing sub-test for `productbus.Create` (close pre-existing gap, real DB)

**Why this task exists:** Per stakeholder requirement, `SeedCreate`'s deliberate skip of `delegate.Call` requires that `Create`'s delegate-firing be independently tested. **It currently is not** (verified across the repo; same gap exists for `labelbus.Create`).

**Why real DB instead of a mock storer:** Stakeholder preference is to test actual functionality, not mocks. The trick is that `dbtest.NewDatabase` returns a `*Database` that exposes both `DB *sqlx.DB` (line 529) AND `BusDomain` — so we can build a *fresh* `productbus.Business` inside the sub-test using `productdb.NewStore(log, db.DB)` (real storer, real DB) but a custom `delegate.New(log)` (so we can register a handler and observe events). The pre-wired `db.BusDomain.Product` can't be intercepted; building a parallel Business with the same DB is how we observe delegate events without mocks.

**Files:**
- Modify: `business/domain/products/productbus/productbus_test.go`

- [ ] **Step 1: Add the `delegateFires` sub-test function**

Append the following to `business/domain/products/productbus/productbus_test.go` (immediately after the existing `create` function, before `update`).

The verified facts driving this test:

| Fact | Value | Source |
|---|---|---|
| `dbtest.Database.DB` | `*sqlx.DB` | `dbtest.go:529` |
| `productdb.NewStore(log *logger.Logger, db *sqlx.DB) *Store` | constructor | `productdb.go:25` |
| `productbus.NewBusiness(log, delegate, storer)` | three-arg constructor | `productbus.go:47` |
| `productbus.DomainName = "product"` | constant | `event.go:11` |
| `productbus.ActionCreated = "created"` | constant | `event.go:20` |
| Handler signature | `func(context.Context, delegate.Data) error` | `approvalrequestbus/delegate_test.go:62` |
| `delegate.Data{Domain, Action, RawParams}` | exported fields | `event.go:56-60` |

```go
// capturedDelegate records what a delegate handler observed.
type capturedDelegate struct {
	Domain string
	Action string
	Count  int
}

// delegateFires verifies that productbus.Create fires the ActionCreated
// delegate event. SeedCreate (added in this same PR) intentionally
// SKIPS this side-effect; pinning Create's behavior here makes that
// skip verifiable as deliberate rather than an oversight.
//
// The sub-test builds a parallel productbus.Business that shares the
// real DB connection (via productdb.NewStore on db.DB) but uses a
// fresh delegate.New so the handler can be intercepted. The pre-wired
// db.BusDomain.Product holds an unobservable production delegate; the
// only way to assert event firing without mocks is to construct an
// observable Business beside it.
func delegateFires(db *dbtest.Database, sd unitest.SeedData) []unitest.Table {
	var buf bytes.Buffer
	log := logger.New(&buf, logger.LevelInfo, "TEST", func(context.Context) string { return "" })

	del := delegate.New(log)

	var (
		mu             sync.Mutex
		capturedDomain string
		capturedAction string
		capturedCount  int
	)
	del.Register(productbus.DomainName, productbus.ActionCreated, func(_ context.Context, data delegate.Data) error {
		mu.Lock()
		defer mu.Unlock()
		capturedDomain = data.Domain
		capturedAction = data.Action
		capturedCount++
		return nil
	})

	// Real storer against the same DB the rest of Test_Product uses.
	bus := productbus.NewBusiness(log, del, productdb.NewStore(log, db.DB))

	np := productbus.NewProduct{
		SKU:               "DELEGATE-REAL-001",
		BrandID:           sd.Brands[0].BrandID,
		ProductCategoryID: sd.ProductCategories[0].ProductCategoryID,
		Name:              "Delegate Real-DB Fixture",
		Description:       "verifies Create fires ActionCreated through delegate",
		ModelNumber:       "DLG-REAL",
		UpcCode:           "DELEGATE-REAL-UPC-001",
		Status:            "active",
		IsActive:          true,
		IsPerishable:      false,
		HandlingInstructions: "",
		UnitsPerCase:      1,
		TrackingType:      "none",
	}

	return []unitest.Table{
		{
			Name: "Create_FiresActionCreated",
			ExpResp: capturedDelegate{
				Domain: productbus.DomainName,
				Action: productbus.ActionCreated,
				Count:  1,
			},
			ExcFunc: func(ctx context.Context) any {
				if _, err := bus.Create(ctx, np); err != nil {
					return err
				}
				mu.Lock()
				defer mu.Unlock()
				return capturedDelegate{
					Domain: capturedDomain,
					Action: capturedAction,
					Count:  capturedCount,
				}
			},
			CmpFunc: func(got, exp any) string {
				if err, isErr := got.(error); isErr {
					return err.Error()
				}
				return cmp.Diff(exp, got)
			},
		},
	}
}
```

- [ ] **Step 2: Add required imports**

In the import block of `productbus_test.go`, add:

```go
"bytes"
"sync"

"github.com/timmaaaz/ichor/business/domain/products/productbus/stores/productdb"
"github.com/timmaaaz/ichor/business/sdk/delegate"
"github.com/timmaaaz/ichor/foundation/logger"
```

`bytes` and `sync` go in the stdlib group; the project imports go in the project group (alphabetical with the existing imports). The `delegate` and `logger` packages are likely not yet imported; verify and add.

- [ ] **Step 3: Wire the sub-test into `Test_Product`**

Modify `Test_Product` (lines 24-40). Add the new `unitest.Run` line after `create` and before `update`:

```go
	unitest.Run(t, query(db.BusDomain, sd), "query")
	unitest.Run(t, create(db.BusDomain, sd), "create")
	unitest.Run(t, delegateFires(db, sd), "delegateFires")
	unitest.Run(t, update(db.BusDomain, sd), "update")
	unitest.Run(t, delete(db.BusDomain, sd), "delete")
```

**Note the signature difference:** `delegateFires` takes `*dbtest.Database` (the full struct, so it can reach `db.DB`), not `dbtest.BusDomain` like the other sub-tests. This is the only sub-test that needs the raw DB handle.

- [ ] **Step 4: Run the new sub-test**

Run: `go test ./business/domain/products/productbus/... -run Test_Product/delegateFires -v`
Expected: PASS — captured triple matches `{Domain: "product", Action: "created", Count: 1}`. The test should pass on first run because `Create` already fires the delegate; this task is closing the *test gap*, not changing behavior.

- [ ] **Step 5: Run full Test_Product to confirm no regressions**

Run: `go test ./business/domain/products/productbus/... -run Test_Product -v`
Expected: PASS for query, create, delegateFires, update, delete.

- [ ] **Step 6: Commit**

```bash
git add business/domain/products/productbus/productbus_test.go
git commit -m "test(productbus): cover Create's delegate event firing (real DB)

Adds the delegateFires sub-test to Test_Product, building a parallel
productbus.Business that shares the real DB connection but uses a
fresh delegate so handler invocations can be intercepted. No mock
storers — exercises the actual productdb.Create path.

Closes a pre-existing test gap (same gap exists for labelbus.Create).
This is the precondition for the upcoming SeedCreate method, which
intentionally skips the delegate to avoid triggering Temporal
workflows during DB seeding. With Create's delegate behavior pinned
here, SeedCreate's skip is verifiable as deliberate.

Refs: P1 productbus.SeedCreate."
```

---

### Task 5: Add `productbus.SeedCreate` (TDD via integration sub-test)

**Files:**
- Modify: `business/domain/products/productbus/productbus_test.go` (test first)
- Modify: `business/domain/products/productbus/productbus.go` (implementation)

The existing `productbus_test.go` uses the `unitest.Run` table-test pattern. Sub-tests have signature `func name(busDomain dbtest.BusDomain, sd unitest.SeedData) []unitest.Table` and return a slice of `unitest.Table`. We add a `seedCreate` sub-test wired into `Test_Product` between `create` and `update`.

- [ ] **Step 1: Add the seedCreate sub-test function**

Append the following function to `business/domain/products/productbus/productbus_test.go` (immediately after the `create` function, before `update`):

```go
func seedCreate(busDomain dbtest.BusDomain, sd unitest.SeedData) []unitest.Table {
	// Two fixtures verify SeedCreate's behavior:
	//   A: zero CreatedDate / UpdatedDate / TrackingType — defaults applied
	//   B: caller-supplied CreatedDate + TrackingType — preserved verbatim
	stableA := seedid.Stable("test:productbus:seedCreate:fixture-A")
	stableB := seedid.Stable("test:productbus:seedCreate:fixture-B")

	historicalDate := time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC)

	productA := productbus.Product{
		ProductID:            stableA,
		SKU:                  "SEED-TEST-A",
		BrandID:              sd.Brands[0].BrandID,
		ProductCategoryID:    sd.ProductCategories[0].ProductCategoryID,
		Name:                 "SeedCreate Fixture A",
		Description:          "deterministic seed test",
		ModelNumber:          "SC-A",
		UpcCode:              "SeedTestA-UPC",
		Status:               "active",
		IsActive:             true,
		IsPerishable:         false,
		HandlingInstructions: "",
		UnitsPerCase:         12,
		// TrackingType, CreatedDate, UpdatedDate intentionally zero — must default.
	}

	productB := productbus.Product{
		ProductID:            stableB,
		SKU:                  "SEED-TEST-B",
		BrandID:              sd.Brands[0].BrandID,
		ProductCategoryID:    sd.ProductCategories[0].ProductCategoryID,
		Name:                 "SeedCreate Fixture B",
		Description:          "deterministic seed test (preserved fields)",
		ModelNumber:          "SC-B",
		UpcCode:              "SeedTestB-UPC",
		Status:               "active",
		IsActive:             true,
		IsPerishable:         true,
		HandlingInstructions: "keep cool",
		UnitsPerCase:         24,
		TrackingType:         "lot",
		CreatedDate:          historicalDate,
		UpdatedDate:          historicalDate,
	}

	return []unitest.Table{
		{
			Name:    "SeedCreate_DefaultsApplied",
			ExpResp: stableA,
			ExcFunc: func(ctx context.Context) any {
				if err := busDomain.Product.SeedCreate(ctx, productA); err != nil {
					return err
				}
				got, err := busDomain.Product.QueryByID(ctx, stableA)
				if err != nil {
					return err
				}
				if got.ProductID != stableA {
					return fmt.Errorf("ProductID drift: got %s want %s", got.ProductID, stableA)
				}
				if got.TrackingType != "none" {
					return fmt.Errorf("TrackingType default missed: got %q want %q", got.TrackingType, "none")
				}
				if got.CreatedDate.IsZero() {
					return fmt.Errorf("CreatedDate not defaulted")
				}
				if got.UpdatedDate.IsZero() {
					return fmt.Errorf("UpdatedDate not defaulted")
				}
				if !got.UpdatedDate.Equal(got.CreatedDate) {
					return fmt.Errorf("UpdatedDate (%s) should equal CreatedDate (%s) when both default", got.UpdatedDate, got.CreatedDate)
				}
				return got.ProductID
			},
			CmpFunc: func(got, exp any) string {
				gotID, ok := got.(uuid.UUID)
				if !ok {
					if err, isErr := got.(error); isErr {
						return err.Error()
					}
					return "expected uuid.UUID"
				}
				return cmp.Diff(exp.(uuid.UUID), gotID)
			},
		},
		{
			Name:    "SeedCreate_PreservesCallerFields",
			ExpResp: stableB,
			ExcFunc: func(ctx context.Context) any {
				if err := busDomain.Product.SeedCreate(ctx, productB); err != nil {
					return err
				}
				got, err := busDomain.Product.QueryByID(ctx, stableB)
				if err != nil {
					return err
				}
				if got.ProductID != stableB {
					return fmt.Errorf("ProductID drift: got %s want %s", got.ProductID, stableB)
				}
				if !got.CreatedDate.Equal(historicalDate) {
					return fmt.Errorf("CreatedDate not preserved: got %s want %s", got.CreatedDate, historicalDate)
				}
				if got.TrackingType != "lot" {
					return fmt.Errorf("TrackingType not preserved: got %q want %q", got.TrackingType, "lot")
				}
				return got.ProductID
			},
			CmpFunc: func(got, exp any) string {
				gotID, ok := got.(uuid.UUID)
				if !ok {
					if err, isErr := got.(error); isErr {
						return err.Error()
					}
					return "expected uuid.UUID"
				}
				return cmp.Diff(exp.(uuid.UUID), gotID)
			},
		},
	}
}
```

- [ ] **Step 2: Wire the sub-test into `Test_Product`**

In `business/domain/products/productbus/productbus_test.go`, modify the existing `Test_Product` function (lines 24-40). Add a new `unitest.Run` line between `create` and `update`:

Before (lines 36-39):
```go
	unitest.Run(t, query(db.BusDomain, sd), "query")
	unitest.Run(t, create(db.BusDomain, sd), "create")
	unitest.Run(t, update(db.BusDomain, sd), "update")
	unitest.Run(t, delete(db.BusDomain, sd), "delete")
```

After:
```go
	unitest.Run(t, query(db.BusDomain, sd), "query")
	unitest.Run(t, create(db.BusDomain, sd), "create")
	unitest.Run(t, seedCreate(db.BusDomain, sd), "seedCreate")
	unitest.Run(t, update(db.BusDomain, sd), "update")
	unitest.Run(t, delete(db.BusDomain, sd), "delete")
```

- [ ] **Step 3: Add imports**

In the import block of `productbus_test.go` (lines 3-22), add:

```go
"time"
"github.com/timmaaaz/ichor/business/sdk/seedid"
```

`time` goes in the stdlib group; `seedid` goes in the project group (alphabetical).

- [ ] **Step 4: Run the test and confirm it fails**

Run: `go test ./business/domain/products/productbus/... -run Test_Product -v`
Expected: FAIL — `busDomain.Product.SeedCreate undefined` (compile error).

- [ ] **Step 5: Implement SeedCreate**

In `business/domain/products/productbus/productbus.go`, add the following method **immediately after the `Create` method's closing brace** (after line 116):

```go
// SeedCreate inserts a Product with a caller-supplied ProductID. It is
// the seeding analogue of Create: the regular Create assigns
// uuid.New() and fires a delegate event for workflow automation, both
// of which are inappropriate during deterministic seed runs.
//
// Defaults filled when zero/empty:
//   - CreatedDate (defaults to time.Now().UTC())
//   - UpdatedDate (defaults to CreatedDate)
//   - TrackingType (defaults to "none")
//
// All other Product fields pass through verbatim — the caller is
// responsible for SKU, BrandID, ProductCategoryID, and any other
// required field.
//
// Notably absent: b.delegate.Call(...). Seed paths must not trigger
// workflow events during DB initialization. This skip is the deliberate
// difference from Create; the delegate-firing behavior of Create is
// covered by TestCreate_FiresDelegateEvent in delegate_test.go. Mirrors
// labelbus.SeedCreate.
func (b *Business) SeedCreate(ctx context.Context, p Product) error {
	now := time.Now().UTC()
	if p.CreatedDate.IsZero() {
		p.CreatedDate = now
	}
	if p.UpdatedDate.IsZero() {
		p.UpdatedDate = p.CreatedDate
	}
	if p.TrackingType == "" {
		p.TrackingType = "none"
	}

	if err := b.storer.Create(ctx, p); err != nil {
		return fmt.Errorf("seedcreate: %w", err)
	}
	return nil
}
```

No new imports needed — `context`, `time`, `fmt` are all already imported by `productbus.go`.

- [ ] **Step 6: Run the test and confirm it passes**

Run: `go test ./business/domain/products/productbus/... -run Test_Product/seedCreate -v`
Expected: PASS — both sub-cases pass.

- [ ] **Step 7: Run the full productbus test suite**

Run: `go test ./business/domain/products/productbus/...`
Expected: PASS — `Test_Product` (including `query`, `create`, `seedCreate`, `update`, `delete`) and `TestCreate_FiresDelegateEvent` all pass.

- [ ] **Step 8: Commit**

```bash
git add business/domain/products/productbus/productbus.go business/domain/products/productbus/productbus_test.go
git commit -m "feat(productbus): add SeedCreate with caller-supplied ProductID

Adds Business.SeedCreate, mirroring labelbus.SeedCreate: the caller
supplies a fully-formed Product (including a deterministic ProductID),
SeedCreate fills only the zero-valued CreatedDate/UpdatedDate/TrackingType
defaults, and the delegate event is intentionally not fired so seed
runs do not trigger workflow automation.

Storer layer unchanged — productdb.Create already accepts a fully-
formed Product.

Wires a new seedCreate sub-test into Test_Product covering both default-
fill and field-preservation paths.

Refs: P1 productbus.SeedCreate."
```

---

### Task 6: Switch testutil seed helpers to deterministic IDs

**Files:**
- Modify: `business/domain/products/productbus/testutil.go`

We're now switching the two seeding helpers — `TestSeedProducts` and `TestSeedProductsHistoricalWithDistribution` — from `api.Create(ctx, np)` to `api.SeedCreate(ctx, p)` with deterministic IDs. **The trailing `sort.Slice` block in each function must be preserved verbatim** (sorts by Name then SKU; missing this would silently change downstream test ordering).

- [ ] **Step 1: Update the import block**

Current import block (`testutil.go` lines 3-10):

```go
import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/google/uuid"
)
```

Replace with:

```go
import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/sdk/seedid"
)
```

- [ ] **Step 2: Add the conversion helper**

Append the following helper at the bottom of `testutil.go` (after `TestSeedProductsHistoricalWithDistribution`):

```go
// newProductToSeedProduct copies a NewProduct into a Product struct,
// deriving a deterministic ProductID from the SKU. Historical
// CreatedDate (set by TestNewProductsHistorical) is preserved verbatim
// so timestamp-sensitive tests still see the same dates they did under
// the api.Create path.
func newProductToSeedProduct(np NewProduct) Product {
	var createdDate time.Time
	if np.CreatedDate != nil {
		createdDate = *np.CreatedDate
	}
	return Product{
		ProductID:            seedid.Stable("product:" + np.SKU),
		SKU:                  np.SKU,
		BrandID:              np.BrandID,
		ProductCategoryID:    np.ProductCategoryID,
		Name:                 np.Name,
		Description:          np.Description,
		ModelNumber:          np.ModelNumber,
		UpcCode:              np.UpcCode,
		Status:               np.Status,
		IsActive:             np.IsActive,
		IsPerishable:         np.IsPerishable,
		HandlingInstructions: np.HandlingInstructions,
		UnitsPerCase:         np.UnitsPerCase,
		TrackingType:         np.TrackingType,
		InventoryType:        np.InventoryType,
		CreatedDate:          createdDate, // zero → defaulted by SeedCreate
	}
}
```

- [ ] **Step 3: Replace `TestSeedProducts` body verbatim**

Current `TestSeedProducts` (lines 57-78):

```go
func TestSeedProducts(ctx context.Context, n int, brandIDs, productCategoryIDs uuid.UUIDs, api *Business) ([]Product, error) {
	newProducts := TestNewProducts(n, brandIDs, productCategoryIDs)

	products := make([]Product, len(newProducts))

	for i, np := range newProducts {
		product, err := api.Create(ctx, np)
		if err != nil {
			return nil, err
		}
		products[i] = product
	}

	sort.Slice(products, func(i, j int) bool {
		if products[i].Name != products[j].Name {
			return products[i].Name < products[j].Name
		}
		return products[i].SKU < products[j].SKU
	})

	return products, nil
}
```

Replace with:

```go
func TestSeedProducts(ctx context.Context, n int, brandIDs, productCategoryIDs uuid.UUIDs, api *Business) ([]Product, error) {
	newProducts := TestNewProducts(n, brandIDs, productCategoryIDs)

	products := make([]Product, len(newProducts))

	for i, np := range newProducts {
		p := newProductToSeedProduct(np)
		if err := api.SeedCreate(ctx, p); err != nil {
			return nil, err
		}
		// Round-trip via QueryByID so callers receive whatever the DB
		// normalised (UTC timestamps, defaulted TrackingType) — matches
		// the shape they previously got back from api.Create.
		stored, err := api.QueryByID(ctx, p.ProductID)
		if err != nil {
			return nil, fmt.Errorf("querying seeded product %s: %w", np.SKU, err)
		}
		products[i] = stored
	}

	sort.Slice(products, func(i, j int) bool {
		if products[i].Name != products[j].Name {
			return products[i].Name < products[j].Name
		}
		return products[i].SKU < products[j].SKU
	})

	return products, nil
}
```

(The trailing `sort.Slice` is preserved verbatim. Only the loop body changes.)

- [ ] **Step 4: Replace `TestSeedProductsHistoricalWithDistribution` body verbatim**

Current `TestSeedProductsHistoricalWithDistribution` (lines 138-175):

```go
func TestSeedProductsHistoricalWithDistribution(
	ctx context.Context,
	n int,
	daysBack int,
	distribution []string,
	brandIDs, productCategoryIDs uuid.UUIDs,
	api *Business,
) ([]Product, error) {
	if distribution != nil && len(distribution) != n {
		return nil, fmt.Errorf("distribution length %d != n %d", len(distribution), n)
	}

	newProducts := TestNewProductsHistorical(n, daysBack, brandIDs, productCategoryIDs)
	if distribution != nil {
		for i := range newProducts {
			newProducts[i].TrackingType = distribution[i]
			newProducts[i].IsPerishable = distribution[i] == "lot"
		}
	}

	products := make([]Product, 0, n)
	for i, np := range newProducts {
		p, err := api.Create(ctx, np)
		if err != nil {
			return nil, fmt.Errorf("create product %d: %w", i, err)
		}
		products = append(products, p)
	}

	sort.Slice(products, func(i, j int) bool {
		if products[i].Name != products[j].Name {
			return products[i].Name < products[j].Name
		}
		return products[i].SKU < products[j].SKU
	})

	return products, nil
}
```

Replace with:

```go
func TestSeedProductsHistoricalWithDistribution(
	ctx context.Context,
	n int,
	daysBack int,
	distribution []string,
	brandIDs, productCategoryIDs uuid.UUIDs,
	api *Business,
) ([]Product, error) {
	if distribution != nil && len(distribution) != n {
		return nil, fmt.Errorf("distribution length %d != n %d", len(distribution), n)
	}

	newProducts := TestNewProductsHistorical(n, daysBack, brandIDs, productCategoryIDs)
	if distribution != nil {
		for i := range newProducts {
			newProducts[i].TrackingType = distribution[i]
			newProducts[i].IsPerishable = distribution[i] == "lot"
		}
	}

	products := make([]Product, 0, n)
	for i, np := range newProducts {
		p := newProductToSeedProduct(np)
		if err := api.SeedCreate(ctx, p); err != nil {
			return nil, fmt.Errorf("create product %d: %w", i, err)
		}
		stored, err := api.QueryByID(ctx, p.ProductID)
		if err != nil {
			return nil, fmt.Errorf("querying seeded product %d: %w", i, err)
		}
		products = append(products, stored)
	}

	sort.Slice(products, func(i, j int) bool {
		if products[i].Name != products[j].Name {
			return products[i].Name < products[j].Name
		}
		return products[i].SKU < products[j].SKU
	})

	return products, nil
}
```

- [ ] **Step 5: Build sanity**

Run: `go build ./business/domain/products/productbus/...`
Expected: clean exit.

- [ ] **Step 6: Run productbus tests**

Run: `go test ./business/domain/products/productbus/...`
Expected: PASS — `Test_Product` (including the new `seedCreate` sub-test) and `TestCreate_FiresDelegateEvent` pass.

- [ ] **Step 7: Run dependent test suites that consume product seeds**

Run:
```bash
go test ./business/sdk/dbtest/... \
  ./business/domain/products/... \
  ./business/domain/inventory/... \
  ./business/domain/sales/... \
  ./business/domain/procurement/...
```
Expected: PASS. Per CLAUDE.md, do **not** run `go test ./...`.

If a test fails because it asserts on a hardcoded UUID literal that came from an old `uuid.New()` run, the test is wrong by P1's intentional behavior change ("ProductID is now deterministic"). Update the assertion to derive expected via `seedid.Stable("product:"+sku)` — do not weaken it.

- [ ] **Step 8: Commit**

```bash
git add business/domain/products/productbus/testutil.go
git commit -m "feat(productbus): testutil seeds use deterministic ProductID via SeedCreate

Switches TestSeedProducts and TestSeedProductsHistoricalWithDistribution
to build a Product directly with ProductID = seedid.Stable(\"product:\"+sku)
and call SeedCreate, removing uuid.New() as a source of non-determinism
in the seed pipeline.

Downstream effect: inventory.label_catalog.entity_ref rows for
type='product' are now byte-identical across reseeds (closing the
last known determinism gap surfaced by Phase 0g.B4 Tier B label
seeding).

Refs: P1 productbus.SeedCreate."
```

---

### Task 7: Build the determinism check script + Makefile target

**Why this task exists:** The DoD includes "reseed determinism check"; the brief mentions `tasks/t15-determinism-check.sh` which does not exist. We build it now as a reusable, parameterized script so future regressions in seed determinism are caught automatically.

**Design constraints (verified from codebase):**
- Scripts live in `/scripts/` (precedent: `extract_test_failures.py`).
- Shell scripts take DSN as positional arg (precedent: `/deployments/customers/manitowoc/seed.sh`).
- Default DSN: `postgresql://postgres:postgres@localhost:5432/postgres` (matches admin tool defaults + makefile's `ICHOR_DB_HOST=localhost` override).
- `created_date` and `updated_date` columns drift across reseeds via `time.Now()` even with deterministic IDs — must project only stable columns.
- After P1, only `products.products` and `inventory.label_catalog` become byte-stable. Downstream FK tables (cost_history, inventory_items, etc.) still drift via their own `uuid.New()` calls; those are out of scope for this script.

**Files:**
- Create: `scripts/check-determinism.sh`
- Modify: `makefile` (append target)

- [ ] **Step 1: Create the script**

Create `scripts/check-determinism.sh`:

```bash
#!/usr/bin/env bash
# scripts/check-determinism.sh
#
# Verifies that `make reseed-frontend` produces byte-identical seed data
# across runs for the tables that P1 (productbus.SeedCreate) stabilises.
# Excludes columns that drift via wall-clock time (created_date /
# updated_date), since those use time.Now() inside the seed code path.
#
# Tables checked (post-P1):
#   - products.products                (id, sku, brand_id, ...)
#   - inventory.label_catalog          (id, code, type, entity_ref, payload_json)
#
# Tables explicitly NOT checked (still drift via uuid.New() in their own
# seed funcs — out of scope for P1, candidates for future work):
#   - products.cost_history, inventory.inventory_items,
#     inventory.serial_numbers, sales.order_line_items, ...
#
# WHEN TO ADD A TABLE TO THE SNAPSHOTS LIST:
#   You should add an entry to the SNAPSHOTS array below whenever any of
#   these signals appear in a PR or its review:
#
#     1. A new SeedCreate (or equivalent) method is added to a *bus
#        package, mirroring labelbus.SeedCreate / productbus.SeedCreate.
#     2. A TestSeed* function is migrated from `uuid.New()` to deriving
#        its primary key from `seedid.Stable("<entity>:<key>")`. The
#        primary key column is now byte-stable across reseeds and
#        should be diff-checked here.
#     3. A FK column from a downstream table to a stabilized primary
#        key is itself derived from the parent's stable key (rather
#        than queried-then-cached at seed time). When this happens,
#        the downstream table can also be added.
#
#   How to add: append a pipe-delimited entry following the existing
#   format: <label>|<schema.table>|<comma-separated stable cols>|<order-by>
#   Always project columns explicitly — never use `SELECT *` because
#   `created_date` and any future timestamp/sequence columns will drift
#   via wall-clock time and produce false-positive failures.
#
#   How to remove: if a table is intentionally non-deterministic and
#   the comment block above lists it as a candidate, but the team has
#   decided not to stabilize it, delete the corresponding entry below
#   AND update this header comment to remove it from the candidate list.
#
# Usage:
#   ./scripts/check-determinism.sh                  # uses default DSN
#   ./scripts/check-determinism.sh "postgresql://user:pass@host:port/db"
#
# Exit codes:
#   0  all checked tables are byte-identical across reseeds
#   1  at least one table drifted
#   2  prerequisites missing (psql, make, expected env)

set -euo pipefail

DSN="${1:-postgresql://postgres:postgres@localhost:5432/postgres}"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
WORK_DIR="$(mktemp -d -t ichor-determinism.XXXXXX)"
trap 'rm -rf "$WORK_DIR"' EXIT

require_tool () {
	command -v "$1" >/dev/null 2>&1 || {
		echo "ERROR: required tool '$1' not on PATH" >&2
		exit 2
	}
}
require_tool psql
require_tool make
require_tool diff

# Each entry: <label>|<schema.table>|<projected stable columns>|<order-by>
SNAPSHOTS=(
	"products|products.products|id, sku, brand_id, category_id, name, description, model_number, upc_code, status, is_active, is_perishable, handling_instructions, units_per_case, tracking_type, inventory_type|sku"
	"label_catalog|inventory.label_catalog|id, code, type, entity_ref, payload_json|code"
)

dump_snapshot () {
	local run_label="$1"
	local out_dir="$WORK_DIR/$run_label"
	mkdir -p "$out_dir"
	for entry in "${SNAPSHOTS[@]}"; do
		IFS='|' read -r label table cols order_by <<<"$entry"
		local out_file="$out_dir/${label}.tsv"
		# \COPY runs client-side and writes to STDOUT; --no-psqlrc avoids
		# user-specific format settings polluting the output.
		psql "$DSN" --no-psqlrc -At <<-SQL > "$out_file"
		\COPY (SELECT $cols FROM $table ORDER BY $order_by) TO STDOUT WITH (NULL '\N')
		SQL
		echo "  captured: $label ($(wc -l < "$out_file") rows)"
	done
}

cd "$REPO_ROOT"

echo "==> Run 1: make reseed-frontend"
make reseed-frontend
echo "==> Snapshot 1"
dump_snapshot run1

echo "==> Run 2: make reseed-frontend"
make reseed-frontend
echo "==> Snapshot 2"
dump_snapshot run2

echo
echo "==> Diffing snapshots"
DRIFT=0
for entry in "${SNAPSHOTS[@]}"; do
	IFS='|' read -r label _ _ _ <<<"$entry"
	if diff -u "$WORK_DIR/run1/${label}.tsv" "$WORK_DIR/run2/${label}.tsv" > "$WORK_DIR/${label}.diff"; then
		echo "  STABLE: $label"
	else
		DRIFT=1
		echo "  DRIFT:  $label"
		echo "    --- diff (first 40 lines) ---"
		head -40 "$WORK_DIR/${label}.diff" | sed 's/^/    /'
		echo "    --- end diff ---"
		# Preserve full diff for later inspection.
		cp "$WORK_DIR/${label}.diff" "$REPO_ROOT/check-determinism-${label}.diff"
		echo "    full diff saved: check-determinism-${label}.diff"
	fi
done

if [ $DRIFT -ne 0 ]; then
	echo
	echo "FAIL: at least one table drifted between reseeds. See diff files in repo root." >&2
	exit 1
fi
echo
echo "OK: all checked tables are byte-identical across reseeds"
```

- [ ] **Step 2: Make the script executable**

Run: `chmod +x scripts/check-determinism.sh`

- [ ] **Step 3: Add the Makefile target**

Append to `makefile` (find an appropriate location near `reseed-frontend`, around line 466). Add:

```makefile
check-determinism:
	./scripts/check-determinism.sh
```

(Tab indentation, not spaces — Makefiles require literal tabs. If you're editing in a tool that auto-converts, verify the tab is preserved.)

- [ ] **Step 4: Run the script as a smoke test**

This runs against the local dev DB and takes 1-2 minutes (two full reseeds). Make sure the dev DB is accessible (`kubectl get pods` should show `database-*` running on KIND, port-forwarded to localhost:5432).

Run: `make check-determinism`
Expected:
```
==> Run 1: make reseed-frontend
... seeding output ...
==> Snapshot 1
  captured: products (40 rows)
  captured: label_catalog (79 rows)
==> Run 2: make reseed-frontend
... seeding output ...
==> Snapshot 2
  captured: products (40 rows)
  captured: label_catalog (79 rows)

==> Diffing snapshots
  STABLE: products
  STABLE: label_catalog

OK: all checked tables are byte-identical across reseeds
```

If `DRIFT: products` or `DRIFT: label_catalog` shows up, P1 has not closed the gap and you need to debug. The most common cause would be: a non-SKU-derived field is leaking randomness into a "stable" column. Check the `*.diff` files saved at repo root for the specific row(s) that differ.

- [ ] **Step 5: Capture the script output for the PR description**

Save the successful run output to a file you can paste into the PR body later:

```bash
make check-determinism 2>&1 | tee /tmp/check-determinism-output.txt
```

- [ ] **Step 6: Commit**

```bash
git add scripts/check-determinism.sh makefile
git commit -m "feat(scripts): add check-determinism.sh to verify reseed stability

Runs make reseed-frontend twice and asserts that products.products and
inventory.label_catalog produce byte-identical projected snapshots
(stable columns only — created_date and updated_date are excluded
because time.Now() in the seed path drifts independently).

Reusable across future seeding work; downstream FK tables that still
drift via uuid.New() in their own seed funcs (cost_history,
inventory_items, etc.) are explicitly out of scope and documented
in the script header.

Refs: P1 productbus.SeedCreate."
```

---

### Task 8: Whole-repo build + targeted test sweep

**Files:** none (verification only)

- [ ] **Step 1: Whole-repo build**

Run: `go build ./...`
Expected: clean exit.

- [ ] **Step 2: Targeted test sweep — packages we directly modified**

Run:
```bash
go test ./business/sdk/seedid/... \
  ./business/sdk/dbtest/... \
  ./business/domain/products/...
```
Expected: PASS.

- [ ] **Step 3: Wider sweep — packages that consume product seeds**

Run (slower, uses test containers):
```bash
go test ./business/domain/inventory/... \
  ./business/domain/sales/... \
  ./business/domain/procurement/...
```
Expected: PASS.

- [ ] **Step 4: If any test failed because of a hardcoded random-UUID assumption**

Diagnose first. If the test is asserting on a specific UUID literal that came from an old `uuid.New()` run, update it to derive from `seedid.Stable("product:"+sku)`. Do not weaken assertions.

---

### Task 9: Open the PR

- [ ] **Step 1: Push the branch**

```bash
git push -u origin phase-0g/p1-productbus-seedcreate
```

(Per the project's git remote setup, this branch will land on Bitbucket; the GitHub PR is opened against the timmaaaz/ichor mirror.)

- [ ] **Step 2: Open PR against `timmaaaz/ichor` master**

```bash
gh pr create \
  --repo timmaaaz/ichor \
  --base master \
  --head phase-0g/p1-productbus-seedcreate \
  --title "feat(productbus): SeedCreate with deterministic ProductID (P1)" \
  --body "$(cat <<'EOF'
## Summary

Adds `productbus.Business.SeedCreate(ctx, p Product) error` mirroring `labelbus.SeedCreate`. Test seeds now derive `ProductID = seedid.Stable("product:"+sku)`, so reseeds produce byte-identical product UUIDs and `inventory.label_catalog.entity_ref` rows for `type='product'`. Lifts the `detUUID` helper from `business/sdk/dbtest/seed_labels.go` into a new `business/sdk/seedid` package (single source of truth for the namespace, no circular import).

Surfaced as backlog item P1 during Phase 0g.B4 cleanup. **Not a hard gate for any phase** — 0g.B5 already uses SKU-as-handle.

## Behavior change

- `SeedCreate` does **not** fire `delegate.Call` (matches `labelbus.SeedCreate` precedent). Test seeds previously triggered workflow events through `Create`; they no longer do via the seed path.
- The pre-existing test gap for `Create`'s delegate firing is closed by a new `delegate_test.go` so `SeedCreate`'s deliberate skip is verifiable.

## Files

- **New:** `business/sdk/seedid/{seedid.go,seedid_test.go}` — shared deterministic UUID helper.
- **New:** `business/domain/products/productbus/delegate_test.go` — closes pre-existing gap; covers `Create`'s delegate firing.
- **New:** `scripts/check-determinism.sh` + `make check-determinism` — reusable reseed-stability check (excludes time-drifting columns).
- **Modified:** `business/sdk/dbtest/seed_labels.go`, `business/sdk/dbtest/seed_scenarios.go` — use `seedid.Stable`.
- **Modified:** `business/domain/products/productbus/productbus.go` — adds `SeedCreate`.
- **Modified:** `business/domain/products/productbus/productbus_test.go` — adds `seedCreate` sub-test.
- **Modified:** `business/domain/products/productbus/testutil.go` — both seed helpers use `SeedCreate`.

## Determinism verification

```
[paste /tmp/check-determinism-output.txt here]
```

## Future work (out of scope for P1)

The following FK tables still drift via `uuid.New()` in their own seed funcs and are not yet stabilised: `products.cost_history`, `inventory.inventory_items`, `inventory.serial_numbers`, `inventory.quality_inspections`, `inventory.inventory_transactions`, `inventory.inventory_adjustments`, `inventory.transfer_orders`, `sales.order_line_items`, `inventory.put_away_tasks`, `inventory.pick_tasks`, `inventory.cycle_count_items`. The check script can be widened table-by-table as those are migrated.

## Test plan

- [x] `go build ./...` clean
- [x] `go test ./business/sdk/seedid/...` passes
- [x] `go test ./business/sdk/dbtest/...` passes
- [x] `go test ./business/domain/products/...` passes (new `Test_Product/seedCreate` and `TestCreate_FiresDelegateEvent` included)
- [x] `go test ./business/domain/inventory/... ./business/domain/sales/... ./business/domain/procurement/...` passes
- [x] `make check-determinism` passes (output above)

🤖 Generated with [Claude Code](https://claude.com/claude-code)
EOF
)"
```

- [ ] **Step 3: Capture the PR URL** for the post-PR review tasks.

---

### Task 10: In-session code review

- [ ] **Step 1: Invoke superpowers:code-reviewer**

Spawn the code-reviewer agent against the diff:

```
"Review the changes on phase-0g/p1-productbus-seedcreate against the implementation
plan at docs/superpowers/plans/2026-04-29-p1-productbus-seedcreate.md. Specifically
verify:
1. SeedCreate skips delegate.Call (deliberate, documented in godoc).
2. seedid.Namespace value matches the historical 'deadbeef-...' constant byte-for-byte.
3. testutil conversion preserves all NewProduct fields (no silent field drops).
4. Determinism diff in PR body is empty.
5. delegate_test.go correctly registers a handler and asserts the captured event
   matches productbus.DomainName + productbus.ActionCreated.
6. check-determinism.sh excludes created_date/updated_date columns from the diff
   (otherwise the script would always fail spuriously)."
```

- [ ] **Step 2: Address feedback** — fix any high-confidence issues; push as new commits.

- [ ] **Step 3: Run `/code-review:code-review` in a fresh session** against the new HEAD (Step 8 of user's brief).

---

### Task 11: Merge + memory + arch-doc update

- [ ] **Step 1: After both review tiers are clean, the user merges the PR**

(Do not auto-merge.)

- [ ] **Step 2: Update `docs/arch/seeding.md`**

Per CLAUDE.md, `docs/arch/seeding.md` is the authoritative source on seeding. After P1, it must mention the new artifacts so future agents looking at "how does deterministic seeding work in this codebase?" find the correct entry points.

Read the current `docs/arch/seeding.md` to find the appropriate insertion points, then add:

1. A reference to `business/sdk/seedid` as the shared deterministic-UUID helper (with the namespace constant pinned).
2. A list of bus packages that expose a `SeedCreate` entry point: currently `labelbus.SeedCreate` and `productbus.SeedCreate`. Note that `SeedCreate` deliberately skips `delegate.Call` — workflow events do not fire on the seed path.
3. A reference to `make check-determinism` (and `scripts/check-determinism.sh`) as the tool that verifies reseed stability for the tables stabilized so far.
4. A "candidate next" note listing the FK tables that still drift via `uuid.New()` in their own seed funcs (cost_history, inventory_items, serial_numbers, etc.) so the next agent picking up this thread knows where to look.

Keep the additions tight — arch docs in this repo are scannable, not exhaustive. A bullet list of references is preferred over prose.

- [ ] **Step 3: Update the project memory entry**

Open `/Users/jaketimmer/.claude/projects/-Users-jaketimmer-src-work-superior-ichor-vue-ichor/memory/project_p1_productbus_seedcreate.md` and prepend a status block:

```markdown
## Status: DONE 2026-04-29
- Merged in PR #<N> (timmaaaz/ichor)
- New shared package: business/sdk/seedid
- New script: scripts/check-determinism.sh + make check-determinism
- Closed pre-existing delegate-firing test gap for productbus.Create (real-DB sub-test, no mocks)
- Determinism verified: products.products and inventory.label_catalog byte-identical across reseeds
- docs/arch/seeding.md updated to reference seedid + SeedCreate entry points
- Future widen: scripts/check-determinism.sh can grow to cover downstream FK tables (cost_history, inventory_items, etc.) as they're migrated to deterministic seed IDs
```

- [ ] **Step 4: Commit the arch-doc update if it landed in this branch**

If you updated `docs/arch/seeding.md` before merge (recommended — arch updates belong in the same PR as the implementation), it'll be in the PR. If you updated it after merge, commit and push to `main` directly:

```bash
git add docs/arch/seeding.md
git commit -m "docs(arch): document seedid + productbus.SeedCreate entry points

Adds references to the new business/sdk/seedid package, the
SeedCreate methods on labelbus and productbus, and the
make check-determinism verification target.

Refs: P1 productbus.SeedCreate (PR #<N>)."
```

---

## Definition of Done

| Item | Task | Status |
|---|---|---|
| `productbus.Business.SeedCreate` exists with caller-provided ProductID | 5 | ☐ |
| `productdb.Create` supports the SeedCreate path (no change needed) | N/A | ☐ |
| `testutil.go` derives `ProductID = seedid.Stable("product:"+sku)` and calls `SeedCreate` | 6 | ☐ |
| `productbus.Create`'s delegate firing has a real-DB sub-test (precondition closed, no mocks) | 4 | ☐ |
| `go build ./...` clean | 8 | ☐ |
| `go test ./business/domain/products/...` clean | 8 | ☐ |
| Reseed determinism check script (`make check-determinism`) exists, passes, and documents widening criteria | 7 | ☐ |
| In-session code-reviewer green | 10 | ☐ |
| `/code-review:code-review` fresh-session green | 10 | ☐ |
| `docs/arch/seeding.md` updated to reference `seedid`, `SeedCreate` entry points, and `make check-determinism` | 11 | ☐ |
| PR merged, memory updated | 11 | ☐ |
