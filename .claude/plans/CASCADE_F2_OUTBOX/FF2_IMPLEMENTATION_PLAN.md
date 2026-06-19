# FF#2 — Path-A simple-write lost-cascade: Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make every cascade bus's entity write and its `outbox.Emit` commit atomically, so a simple HTTP write can never leave the entity committed with its cascade silently lost.

**Architecture:** Add a `sqldb.BeginOrJoin` primitive (begin a tx, or join the caller's if one is already on ctx) and a generic `outbox.WriteAtomic` wrapper hosted on the `outbox.Writer` (which already holds the base pool). Each cascade bus's `Create/Update/Delete` wraps its body in `WriteAtomic`, which rebinds the bus to the tx via `NewWithTx` and commits — unless a caller tx exists, in which case it JOINS and defers the commit. The begin-or-join authority lives in the business layer, next to the cascade emit; it never nests.

**Tech Stack:** Go 1.23, `github.com/jmoiron/sqlx`, PostgreSQL 16.4, Temporal (cascade relay), `dbtest` (Docker-backed integration tests).

**Design reference:** `.claude/plans/CASCADE_F2_OUTBOX/FF2_PATHA_LOSTCASCADE.md` (the approved Option B). Decision history: that doc's §7 + the F9_ATOMICITY resolved block.

## Global Constraints

- Module: `github.com/timmaaaz/ichor`. Go 1.23.
- **NEVER run `go test ./...`** — run only the packages you changed (CLAUDE.md). Use `go build -C <worktree> ./...` to trust the compiler over gopls.
- Worktree: `/Users/jaketimmer/src/work/superior/ichor/ichor-cascade-pathA-lostcascade`, branch `feature/cascade-pathA-lostcascade` off master `73e1c3bd`.
- Integration tests: each `dbtest.NewDatabase(t, ...)` gets an **isolated database** (confirm in Task 0); unique Temporal task queue per test (`"...-" + t.Name()`); scoped queries (never assert on global row counts).
- The fix must honor the **hard invariant**: when a caller tx is on ctx, JOIN it — never open a nested/inner tx.
- TDD: write the failing test, prove RED, implement to GREEN, commit. Frequent commits.
- **Do NOT push** (no `git push`, no PR) until the user confirms. Local commits only.

---

## File Structure

**New files:**
- `business/sdk/sqldb/beginorjoin.go` — the `BeginOrJoin` begin-or-join primitive (tx management only; no bus knowledge).
- `business/sdk/sqldb/beginorjoin_test.go` — DB-backed unit tests for join vs begin + commit ownership.
- `business/sdk/outbox/atomic.go` — `WriteAtomic` / `WriteAtomicVoid` generic wrappers (rebind bus + commit/defer).
- `business/sdk/outbox/atomic_test.go` — white-box unit tests using a fake bus + a temp probe table.
- `business/domain/core/currencybus/currencybus_atomicity_test.go` — the decisive RED-first test, the no-pool-warn test, and the A→B→A cascade-survives test (exemplar bus).

**Modified files:**
- Every cascade bus `*.go` (~62, full list in Task 5): `NewWithTx` → copy-then-override idiom; each method calling `b.outbox.Emit` → wrapped in `WriteAtomic`.
- `business/sdk/outbox/coverage_test.go` — extend to assert each cascade bus routes writes through `WriteAtomic` (enforcement).

**Out of scope (already atomic — do NOT wrap):**
- `workflowBus` (`.WithOutboxEmitter`, emits `allocation_results`) — only fired by `allocate`/`reserve` self-tx handlers; no simple route.
- The synthesized data handlers (`update_field`/`create_entity`/`transition_status`, `createputawaytask`) — worker-side, already `Begin*` + `sqldb.WithTx`.

---

## Task 0: Confirm test-DB isolation (no code)

**Files:** none (investigation gate).

- [ ] **Step 1: Confirm `dbtest.NewDatabase` gives each test its own database.**

Read `business/sdk/dbtest/dbtest.go` `NewDatabase` and `docs/arch/testing.md`. Confirm each test gets a uniquely-named, isolated database (so DDL like a poison trigger in one test cannot affect another). Expected: yes — a unique DB per test.

- [ ] **Step 2: If isolation is NOT per-test-DB, STOP and re-plan the poison mechanism.**

The decisive test (Task 3) installs a `BEFORE INSERT` trigger on `workflow.cascade_outbox`. That is only safe in an isolated DB. If tests share a DB, switch the poison to a connection-scoped mechanism or a dedicated DB and note it here before proceeding.

---

## Task 1: `sqldb.BeginOrJoin` primitive

**Files:**
- Create: `business/sdk/sqldb/beginorjoin.go`
- Test: `business/sdk/sqldb/beginorjoin_test.go`

**Interfaces:**
- Consumes: `sqldb.Beginner` (`tran.go:10`), `sqldb.CommitRollbacker` (`tran.go:15`), `sqldb.GetTx`/`WithCommitRollbacker` (`context.go:18`/`:28`).
- Produces: `func BeginOrJoin(ctx context.Context, bgn Beginner) (context.Context, CommitRollbacker, bool, error)` — returns `(ctx', tx, owned, err)`. `owned=false` ⇒ a caller already owns the tx (do NOT commit); `owned=true` ⇒ caller owns commit/rollback; on `owned=true` the returned ctx carries the new tx.

- [ ] **Step 1: Write the failing tests**

`business/sdk/sqldb/beginorjoin_test.go` (uses a dbtest pool for real `*sqlx.Tx` values):

```go
package sqldb_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/timmaaaz/ichor/business/sdk/dbtest"
	"github.com/timmaaaz/ichor/business/sdk/sqldb"
)

func Test_BeginOrJoin(t *testing.T) {
	db := dbtest.NewDatabase(t, "Test_BeginOrJoin")
	ctx := context.Background()

	t.Run("begins a fresh tx when ctx has none", func(t *testing.T) {
		newCtx, tx, owned, err := sqldb.BeginOrJoin(ctx, sqldb.NewBeginner(db.DB))
		require.NoError(t, err)
		require.True(t, owned, "no caller tx ⇒ this call owns the new tx")
		require.NotNil(t, tx)

		got, ok := sqldb.GetTx(newCtx)
		require.True(t, ok, "the new tx must be placed on the returned ctx")
		require.NotNil(t, got, "ctx carries the begun *sqlx.Tx so outbox.Emit can ride it")
		require.NoError(t, tx.Rollback())
	})

	t.Run("joins the caller's tx when one is on ctx, and reports not-owned", func(t *testing.T) {
		caller, err := db.DB.Beginx()
		require.NoError(t, err)
		t.Cleanup(func() { _ = caller.Rollback() })
		callerCtx := sqldb.WithTx(ctx, caller)

		newCtx, tx, owned, err := sqldb.BeginOrJoin(callerCtx, sqldb.NewBeginner(db.DB))
		require.NoError(t, err)
		require.False(t, owned, "a caller tx is in flight ⇒ JOIN, do not own/commit it")
		require.Equal(t, caller, tx, "must return the SAME caller tx, not a nested one")
		got, _ := sqldb.GetTx(newCtx)
		require.Equal(t, caller, got, "ctx still carries the caller's tx unchanged")
	})
}
```

- [ ] **Step 2: Run the tests to verify they fail**

Run: `go test -C /Users/jaketimmer/src/work/superior/ichor/ichor-cascade-pathA-lostcascade ./business/sdk/sqldb/ -run Test_BeginOrJoin -v`
Expected: FAIL to compile — `undefined: sqldb.BeginOrJoin`.

- [ ] **Step 3: Write the implementation**

`business/sdk/sqldb/beginorjoin.go`:

```go
package sqldb

import "context"

// BeginOrJoin returns a transaction to run a unit of work on, satisfying the
// begin-or-JOIN invariant. If the context already carries a transaction (a caller
// opened one) it is returned with owned=false — the caller MUST NOT commit it; the
// owner does. Otherwise a fresh transaction is begun on bgn, placed on the returned
// context (so ctx-tx readers such as outbox.Emit ride it), and returned with
// owned=true, in which case the caller owns Commit/Rollback.
//
// It never opens a nested transaction when one is already in flight: begin-or-JOIN,
// not begin-always.
func BeginOrJoin(ctx context.Context, bgn Beginner) (context.Context, CommitRollbacker, bool, error) {
	if tx, ok := GetTx(ctx); ok {
		return ctx, tx, false, nil
	}

	tx, err := bgn.Begin()
	if err != nil {
		return ctx, nil, false, err
	}

	return WithCommitRollbacker(ctx, tx), tx, true, nil
}
```

- [ ] **Step 4: Run the tests to verify they pass**

Run: `go test -C /Users/jaketimmer/src/work/superior/ichor/ichor-cascade-pathA-lostcascade ./business/sdk/sqldb/ -run Test_BeginOrJoin -v`
Expected: PASS (both subtests).

- [ ] **Step 5: Commit**

```bash
git -C /Users/jaketimmer/src/work/superior/ichor/ichor-cascade-pathA-lostcascade add business/sdk/sqldb/beginorjoin.go business/sdk/sqldb/beginorjoin_test.go
git -C /Users/jaketimmer/src/work/superior/ichor/ichor-cascade-pathA-lostcascade commit -m "feat(sqldb): add BeginOrJoin begin-or-join tx primitive (FF#2)"
```

---

## Task 2: `outbox.WriteAtomic` / `WriteAtomicVoid` wrappers

**Files:**
- Create: `business/sdk/outbox/atomic.go`
- Test: `business/sdk/outbox/atomic_test.go`

**Interfaces:**
- Consumes: `Writer.db` (`emit.go:35`, same package), `sqldb.BeginOrJoin`, `sqldb.NewBeginner`, `sqldb.CommitRollbacker`.
- Produces:
  - `func WriteAtomic[B any, T any](ctx, w *Writer, self B, newWithTx func(B, sqldb.CommitRollbacker) (B, error), fn func(context.Context, B) (T, error)) (T, error)`
  - `func WriteAtomicVoid[B any](ctx, w *Writer, self B, newWithTx func(B, sqldb.CommitRollbacker) (B, error), fn func(context.Context, B) error) error`

- [ ] **Step 1: Write the failing tests**

`business/sdk/outbox/atomic_test.go` (white-box — `package outbox` — so it can build a `*Writer` and reach `w.db`). It uses a temp probe table as the "entity" and a `fn` that does a real write then optionally errors, proving begin-path rollback and join-path deferral without needing a real bus:

```go
package outbox

import (
	"context"
	"errors"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"

	"github.com/timmaaaz/ichor/business/sdk/dbtest"
	"github.com/timmaaaz/ichor/business/sdk/sqldb"
)

// fakeBus stands in for a cascade *Business: it holds a tx-bindable executor and
// writes a probe row, exactly as a real storer.Create would.
type fakeBus struct{ exec sqlx.ExtContext }

func (b *fakeBus) NewWithTx(tx sqldb.CommitRollbacker) (*fakeBus, error) {
	ec, err := sqldb.GetExtContext(tx)
	if err != nil {
		return nil, err
	}
	return &fakeBus{exec: ec}, nil
}

func (b *fakeBus) writeProbe(ctx context.Context, id string) error {
	_, err := b.exec.ExecContext(ctx, `INSERT INTO atomic_probe (id) VALUES ($1)`, id)
	return err
}

func Test_WriteAtomic(t *testing.T) {
	db := dbtest.NewDatabase(t, "Test_WriteAtomic")
	ctx := context.Background()
	_, err := db.DB.Exec(`CREATE TABLE atomic_probe (id TEXT PRIMARY KEY)`)
	require.NoError(t, err)
	w := NewWriter(db.Log, db.DB, map[string]string{}, nil)
	count := func(id string) int {
		var n int
		require.NoError(t, db.DB.GetContext(ctx, &n, `SELECT count(*) FROM atomic_probe WHERE id=$1`, id))
		return n
	}

	t.Run("begin path commits on success", func(t *testing.T) {
		_, err := WriteAtomic(ctx, w, &fakeBus{exec: db.DB}, (*fakeBus).NewWithTx,
			func(ctx context.Context, b *fakeBus) (struct{}, error) {
				return struct{}{}, b.writeProbe(ctx, "ok")
			})
		require.NoError(t, err)
		require.Equal(t, 1, count("ok"), "begin path must commit the in-tx write")
	})

	t.Run("begin path rolls the in-tx write back when fn errors (the emit-failure shape)", func(t *testing.T) {
		boom := errors.New("emit failed")
		_, err := WriteAtomic(ctx, w, &fakeBus{exec: db.DB}, (*fakeBus).NewWithTx,
			func(ctx context.Context, b *fakeBus) (struct{}, error) {
				if e := b.writeProbe(ctx, "rollback"); e != nil {
					return struct{}{}, e
				}
				return struct{}{}, boom // emit fails AFTER the entity write
			})
		require.ErrorIs(t, err, boom)
		require.Equal(t, 0, count("rollback"), "begin path must roll the entity write back with the failed emit")
	})

	t.Run("join path uses the caller tx and does NOT commit (owner does)", func(t *testing.T) {
		caller, err := db.DB.Beginx()
		require.NoError(t, err)
		callerCtx := sqldb.WithTx(ctx, caller)

		_, err = WriteAtomic(callerCtx, w, &fakeBus{exec: db.DB}, (*fakeBus).NewWithTx,
			func(ctx context.Context, b *fakeBus) (struct{}, error) {
				return struct{}{}, b.writeProbe(ctx, "joined")
			})
		require.NoError(t, err)
		require.NoError(t, caller.Rollback(), "caller still owns the tx")
		require.Equal(t, 0, count("joined"), "join path must NOT commit; the caller's rollback removed the write")
	})

	t.Run("nil writer runs fn on the unmodified bus with no tx", func(t *testing.T) {
		_, err := WriteAtomic(ctx, (*Writer)(nil), &fakeBus{exec: db.DB}, (*fakeBus).NewWithTx,
			func(ctx context.Context, b *fakeBus) (struct{}, error) {
				return struct{}{}, b.writeProbe(ctx, "nilw")
			})
		require.NoError(t, err)
		require.Equal(t, 1, count("nilw"), "nil writer ⇒ pool write, no tx management")
	})
}
```

- [ ] **Step 2: Run to verify failure**

Run: `go test -C /Users/jaketimmer/src/work/superior/ichor/ichor-cascade-pathA-lostcascade ./business/sdk/outbox/ -run Test_WriteAtomic -v`
Expected: FAIL to compile — `undefined: WriteAtomic`.

- [ ] **Step 3: Write the implementation**

`business/sdk/outbox/atomic.go`:

```go
package outbox

import (
	"context"
	"fmt"

	"github.com/timmaaaz/ichor/business/sdk/sqldb"
)

// WriteAtomic runs a cascade bus's write (entity write + Emit) atomically. With no
// transaction on the context it begins one on the Writer's base pool, rebinds the bus
// to it via newWithTx, runs fn, and commits — rolling back on any error. When a caller
// transaction is already in flight it JOINS it: rebinds the bus onto it, runs fn, and
// does NOT commit (the caller owns the commit). A nil *Writer (a bus built without an
// outbox / pre-cutover) runs fn on the unmodified bus with no transaction management.
//
// newWithTx is the bus's own (*Business).NewWithTx method expression; it rebinds the
// storer to tx. The same tx is placed on ctx (by BeginOrJoin) so the bus's
// outbox.Emit rides it too — both writes share one transaction.
func WriteAtomic[B any, T any](
	ctx context.Context,
	w *Writer,
	self B,
	newWithTx func(B, sqldb.CommitRollbacker) (B, error),
	fn func(ctx context.Context, bus B) (T, error),
) (T, error) {
	var zero T

	if w == nil {
		return fn(ctx, self)
	}

	ctx, tx, owned, err := sqldb.BeginOrJoin(ctx, sqldb.NewBeginner(w.db))
	if err != nil {
		return zero, fmt.Errorf("outbox: begin-or-join: %w", err)
	}

	bus, err := newWithTx(self, tx)
	if err != nil {
		if owned {
			_ = tx.Rollback()
		}
		return zero, err
	}

	out, err := fn(ctx, bus)
	if err != nil {
		if owned {
			_ = tx.Rollback()
		}
		return zero, err
	}

	if owned {
		if err := tx.Commit(); err != nil {
			return zero, fmt.Errorf("outbox: commit: %w", err)
		}
	}
	return out, nil
}

// WriteAtomicVoid is WriteAtomic for bus methods that return only an error (Delete).
func WriteAtomicVoid[B any](
	ctx context.Context,
	w *Writer,
	self B,
	newWithTx func(B, sqldb.CommitRollbacker) (B, error),
	fn func(ctx context.Context, bus B) error,
) error {
	_, err := WriteAtomic(ctx, w, self, newWithTx,
		func(ctx context.Context, bus B) (struct{}, error) {
			return struct{}{}, fn(ctx, bus)
		})
	return err
}
```

- [ ] **Step 4: Run to verify pass**

Run: `go test -C /Users/jaketimmer/src/work/superior/ichor/ichor-cascade-pathA-lostcascade ./business/sdk/outbox/ -run Test_WriteAtomic -v`
Expected: PASS (all 4 subtests).

- [ ] **Step 5: Commit**

```bash
git -C /Users/jaketimmer/src/work/superior/ichor/ichor-cascade-pathA-lostcascade add business/sdk/outbox/atomic.go business/sdk/outbox/atomic_test.go
git -C /Users/jaketimmer/src/work/superior/ichor/ichor-cascade-pathA-lostcascade commit -m "feat(outbox): add WriteAtomic begin-or-join bus-write wrapper (FF#2)"
```

---

## Task 3: Wire the exemplar bus (currencybus) — the decisive RED-first test

This is the headline TDD task: prove the gap RED on the un-wrapped bus, then harden `NewWithTx` and wrap the methods to GREEN.

**Files:**
- Modify: `business/domain/core/currencybus/currencybus.go` (`NewWithTx` :67-80; `Create` :82-118; `Update` :120-166; `Delete` :168-187)
- Test: `business/domain/core/currencybus/currencybus_atomicity_test.go`

**Interfaces:**
- Consumes: `outbox.WriteAtomic`/`WriteAtomicVoid` (Task 2), `currencybus.NewBusiness(...).WithOutbox(w)`, `currencydb.NewStore`, `db.BusDomain.OutboxWriter`, `db.BusDomain.Delegate`.

- [ ] **Step 1: Write the decisive failing test + poison helper**

`business/domain/core/currencybus/currencybus_atomicity_test.go`:

```go
package currencybus_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/timmaaaz/ichor/business/domain/core/currencybus"
	"github.com/timmaaaz/ichor/business/domain/core/currencybus/stores/currencydb"
	"github.com/timmaaaz/ichor/business/sdk/dbtest"
)

// poisonOutbox makes any INSERT into workflow.cascade_outbox fail, so a cascade
// emit raises an error. Safe because dbtest gives each test an isolated database.
func poisonOutbox(t *testing.T, db *dbtest.Database) {
	t.Helper()
	for _, s := range []string{
		`CREATE OR REPLACE FUNCTION workflow.fail_outbox() RETURNS trigger
		   AS $$ BEGIN RAISE EXCEPTION 'poisoned outbox'; END; $$ LANGUAGE plpgsql`,
		`CREATE TRIGGER poison_outbox BEFORE INSERT ON workflow.cascade_outbox
		   FOR EACH ROW EXECUTE FUNCTION workflow.fail_outbox()`,
	} {
		_, err := db.DB.Exec(s)
		require.NoError(t, err)
	}
}

func currencyCount(t *testing.T, db *dbtest.Database, code string) int {
	t.Helper()
	var n int
	require.NoError(t, db.DB.GetContext(context.Background(), &n,
		`SELECT count(*) FROM core.currencies WHERE code = $1`, code))
	return n
}

// Test_Currencybus_Atomicity_BeginPath proves the FF#2 fix: on a simple write (no
// caller tx), a failed cascade emit rolls the entity write back with it.
func Test_Currencybus_Atomicity_BeginPath(t *testing.T) {
	db := dbtest.NewDatabase(t, "Test_Currencybus_Atomicity_BeginPath")
	ctx := context.Background()

	bus := currencybus.NewBusiness(db.Log, db.BusDomain.Delegate,
		currencydb.NewStore(db.Log, db.DB)).WithOutbox(db.BusDomain.OutboxWriter)

	poisonOutbox(t, db)

	uid := uuid.New()
	_, err := bus.Create(ctx, currencybus.NewCurrency{
		Code: "XTS", Name: "Atomicity Probe", Symbol: "¤", Locale: "en", DecimalPlaces: 2,
		IsActive: true, SortOrder: 1, CreatedBy: uid,
	})
	require.Error(t, err, "Create must fail when its cascade emit fails")

	require.Equal(t, 0, currencyCount(t, db, "XTS"),
		"FF#2: the entity write must roll back WITH the failed cascade emit (atomic) — "+
			"on master it is left committed (the lost-cascade gap)")
}
```

> Verify before running: (a) `core.currencies.created_by` has no FK to a users table (it is a plain audit uuid). If it DOES FK, seed a user via `db.BusDomain.User.Create(...)` and use its ID for `CreatedBy`. (b) `currencybus.NewCurrency` field names match `:82-118` (Code/Name/Symbol/Locale/DecimalPlaces/IsActive/SortOrder/CreatedBy).

- [ ] **Step 2: Run to verify RED on the un-wrapped bus**

Run: `go test -C /Users/jaketimmer/src/work/superior/ichor/ichor-cascade-pathA-lostcascade ./business/domain/core/currencybus/ -run Test_Currencybus_Atomicity_BeginPath -v`
Expected: FAIL — `currencyCount` returns `1`. On master, `storer.Create` autocommits the currency on the pool, then the poisoned `Emit` fails on the pool as a separate statement → the currency survives. This is the gap, demonstrated.

- [ ] **Step 3: Harden `NewWithTx` (copy-then-override) so the wrap won't drop `del`**

Replace `currencybus.go:67-80` with the name-agnostic copy-then-override idiom (this threads ALL fields, including the currently-dropped `del`):

```go
// NewWithTx constructs a new business value that will use the specified transaction
// in any store-related calls. It copies the receiver and overrides only the storer,
// so no field (delegate, outbox, log) is silently dropped (cf. commit 63f6b034).
func (b *Business) NewWithTx(tx sqldb.CommitRollbacker) (*Business, error) {
	storer, err := b.storer.NewWithTx(tx)
	if err != nil {
		return nil, err
	}

	nb := *b
	nb.storer = storer
	return &nb, nil
}
```

- [ ] **Step 4: Wrap `Create`, `Update`, `Delete` in `WriteAtomic`**

Keep the `otel.AddSpan`/`defer span.End()` lines OUTSIDE the wrapper (so the span covers the tx), then wrap the remaining body. The closure parameter is named `b`, shadowing the receiver, so the existing body runs on the tx-bound bus unchanged.

`Create` (`currencybus.go:82-118`) becomes:

```go
func (b *Business) Create(ctx context.Context, nc NewCurrency) (Currency, error) {
	ctx, span := otel.AddSpan(ctx, "business.currencybus.create")
	defer span.End()

	return outbox.WriteAtomic(ctx, b.outbox, b, (*Business).NewWithTx,
		func(ctx context.Context, b *Business) (Currency, error) {
			now := time.Now()

			currency := Currency{
				ID:            uuid.New(),
				Code:          nc.Code,
				Name:          nc.Name,
				Symbol:        nc.Symbol,
				Locale:        nc.Locale,
				DecimalPlaces: nc.DecimalPlaces,
				IsActive:      nc.IsActive,
				SortOrder:     nc.SortOrder,
				CreatedBy:     nc.CreatedBy,
				CreatedDate:   now,
				UpdatedBy:     nc.CreatedBy,
				UpdatedDate:   now,
			}

			if err := b.storer.Create(ctx, currency); err != nil {
				return Currency{}, fmt.Errorf("creating currency: %w", err)
			}

			evtData := ActionCreatedData(currency)
			if err := b.outbox.Emit(ctx, evtData); err != nil {
				return Currency{}, fmt.Errorf("emit cascade event: %w", err)
			}
			if err := b.del.Call(ctx, ActionCreatedData(currency)); err != nil {
				b.log.Error(ctx, "currencybus: delegate call failed", "action", ActionCreated, "err", err)
			}

			return currency, nil
		})
}
```

`Update` (`:120-166`): identical pattern — keep the AddSpan lines, then `return outbox.WriteAtomic(ctx, b.outbox, b, (*Business).NewWithTx, func(ctx context.Context, b *Business) (Currency, error) { <existing body from "before := currency" through "return currency, nil"> })`.

`Delete` (`:168-187`): returns only `error`, so use `WriteAtomicVoid`:

```go
func (b *Business) Delete(ctx context.Context, currency Currency) error {
	ctx, span := otel.AddSpan(ctx, "business.currencybus.delete")
	defer span.End()

	return outbox.WriteAtomicVoid(ctx, b.outbox, b, (*Business).NewWithTx,
		func(ctx context.Context, b *Business) error {
			if err := b.storer.Delete(ctx, currency); err != nil {
				return fmt.Errorf("deleting currency: %w", err)
			}
			evtData := ActionDeletedData(currency)
			if err := b.outbox.Emit(ctx, evtData); err != nil {
				return fmt.Errorf("emit cascade event: %w", err)
			}
			if err := b.del.Call(ctx, ActionDeletedData(currency)); err != nil {
				b.log.Error(ctx, "currencybus: delegate call failed", "action", ActionDeleted, "err", err)
			}
			return nil
		})
}
```

- [ ] **Step 5: Run the decisive test to verify GREEN**

Run: `go test -C /Users/jaketimmer/src/work/superior/ichor/ichor-cascade-pathA-lostcascade ./business/domain/core/currencybus/ -run Test_Currencybus_Atomicity_BeginPath -v`
Expected: PASS — `currencyCount("XTS") == 0`. `Create` self-began a tx; the poisoned emit rolled the currency back with it.

- [ ] **Step 6: Add the "no pool-fallback warn" companion test**

Append to `currencybus_atomicity_test.go` — proves the wrapped Create rides a tx (does NOT hit the `emit.go:118` pool fallback), and the cascade row is committed:

```go
func Test_Currencybus_Atomicity_NoPoolFallbackWarn(t *testing.T) {
	db := dbtest.NewDatabase(t, "Test_Currencybus_Atomicity_NoPoolFallbackWarn")
	ctx := context.Background()
	const warn = "outbox: no transaction on context"

	// db.Log writes to the test's captured buffer; assert via dbtest's log capture if
	// available, else build a bytes.Buffer-backed logger + Writer like outbox_test.go's
	// bufWriter (emit.go:45 NewWriter) and wire it into a fresh currencybus.
	bus := currencybus.NewBusiness(db.Log, db.BusDomain.Delegate,
		currencydb.NewStore(db.Log, db.DB)).WithOutbox(db.BusDomain.OutboxWriter)

	cur, err := bus.Create(ctx, currencybus.NewCurrency{
		Code: "XTT", Name: "OnTx", Symbol: "¤", Locale: "en", DecimalPlaces: 2,
		IsActive: true, SortOrder: 1, CreatedBy: uuid.New(),
	})
	require.NoError(t, err)

	var n int
	require.NoError(t, db.DB.GetContext(ctx, &n,
		`SELECT count(*) FROM workflow.cascade_outbox WHERE domain = $1 AND action = $2`,
		currencybus.DomainName, currencybus.ActionCreated))
	require.Equal(t, 1, n, "wrapped Create must commit exactly one cascade_outbox row on its own tx")
	_ = cur
}
```

> If `db.Log` does not expose captured output, use the `bufWriter` pattern from `outbox_test.go:315-318` (a `bytes.Buffer` logger) and assert `require.NotContains(buf.String(), warn)`. The committed-row assertion above is the primary signal; the warn assertion is the belt-and-suspenders.

- [ ] **Step 7: Run + verify the build of the whole bus package**

Run: `go test -C /Users/jaketimmer/src/work/superior/ichor/ichor-cascade-pathA-lostcascade ./business/domain/core/currencybus/... -v`
Then: `go build -C /Users/jaketimmer/src/work/superior/ichor/ichor-cascade-pathA-lostcascade ./...`
Expected: tests PASS; build clean.

- [ ] **Step 8: Commit**

```bash
git -C /Users/jaketimmer/src/work/superior/ichor/ichor-cascade-pathA-lostcascade add business/domain/core/currencybus/
git -C /Users/jaketimmer/src/work/superior/ichor/ichor-cascade-pathA-lostcascade commit -m "feat(currencybus): wrap cascade writes in WriteAtomic; harden NewWithTx (FF#2)"
```

---

## Task 4: A→B→A cascade survives the new atomic path

Prove a simple-write cascade (now written in the bus's self-begun tx) is durably delivered by the relay, and the A→B→A loop guard still stops. Reuse the synchronous `relay.ProcessBatch` rig from `cascade_ryw_test.go` (deterministic; no worker timing).

**Files:**
- Test: append to `business/domain/core/currencybus/currencybus_atomicity_test.go`

**Interfaces:**
- Consumes: `workflowtemporal.NewRelay(log, db.DB, dispatcher, RelayConfig{})` (`relay.go:93`), `(*Relay).ProcessBatch(ctx) (int, error)` (`relay.go:142`), `workflowtemporal.EventDispatcher` (`relay.go:55`).

- [ ] **Step 1: Write the test (a recorder dispatcher + commit-boundary drain)**

```go
type recorder struct{ seen int }

func (r *recorder) OnEntityEvent(_ context.Context, _ workflow.TriggerEvent) error {
	r.seen++
	return nil
}

// Test_Currencybus_Cascade_SurvivesBeginPath: a simple-write Create (no caller tx) now
// writes its cascade row inside the bus's own tx; once committed, the relay delivers it
// exactly once. Before commit it is invisible (proving it rode a real tx, not the pool).
func Test_Currencybus_Cascade_SurvivesBeginPath(t *testing.T) {
	db := dbtest.NewDatabase(t, "Test_Currencybus_Cascade_SurvivesBeginPath")
	ctx := context.Background()

	rec := &recorder{}
	relay := workflowtemporal.NewRelay(db.Log, db.DB, rec, workflowtemporal.RelayConfig{})

	bus := currencybus.NewBusiness(db.Log, db.BusDomain.Delegate,
		currencydb.NewStore(db.Log, db.DB)).WithOutbox(db.BusDomain.OutboxWriter)

	_, err := bus.Create(ctx, currencybus.NewCurrency{
		Code: "XCS", Name: "Cascade", Symbol: "¤", Locale: "en", DecimalPlaces: 2,
		IsActive: true, SortOrder: 1, CreatedBy: uuid.New(),
	})
	require.NoError(t, err, "Create committed its own tx (entity + cascade row together)")

	// The bus already committed; the relay finds and delivers the row exactly once.
	n, err := relay.ProcessBatch(ctx)
	require.NoError(t, err)
	require.GreaterOrEqual(t, n, 1, "relay delivers the committed simple-write cascade row")
	require.GreaterOrEqual(t, rec.seen, 1, "the cascade was dispatched (not lost on the pool)")
}
```

> Note: this asserts delivery of the simple-write cascade. The A→B→A *loop-guard-still-stops* property is already covered by `temporal/lineage_test.go` (`TestGuard_ABA_StopsAfterOneHop`) and the actionhandlers cascade suite, which are unaffected by FF#2 (the row content + lineage are unchanged; only the tx the row is written on changed). Do NOT rebuild loop-guard coverage here. If a fuller A→B→A is wanted, extend using `cascade_rig_test.go`'s `seedActiveRule`/`updateFieldByID`/`executionVisitedSets` helpers with a unique task queue.

- [ ] **Step 2: Run + verify**

Run: `go test -C /Users/jaketimmer/src/work/superior/ichor/ichor-cascade-pathA-lostcascade ./business/domain/core/currencybus/... -run Test_Currencybus_Cascade -v`
Expected: PASS.

- [ ] **Step 3: Commit**

```bash
git -C /Users/jaketimmer/src/work/superior/ichor/ichor-cascade-pathA-lostcascade add business/domain/core/currencybus/currencybus_atomicity_test.go
git -C /Users/jaketimmer/src/work/superior/ichor/ichor-cascade-pathA-lostcascade commit -m "test(currencybus): cascade survives the new atomic begin path (FF#2)"
```

---

## Task 5: Sweep — harden + wrap the remaining cascade buses

Apply the exact Task-3 transformation (harden `NewWithTx`; wrap every method that calls `b.outbox.Emit`) to every other cascade bus. **Mechanical and uniform** — the copy-then-override idiom is name-agnostic, and the wrap re-indents the existing body into a closure. Parallelizable by package (fan out via spawn-agents partitioned by area, per the codebase's broad-write convention).

**The per-bus recipe (apply to each bus below):**
1. **Harden `NewWithTx`:** replace its body with `storer, err := b.storer.NewWithTx(tx); if err != nil { return nil, err }; nb := *b; nb.storer = storer; return &nb, nil`. If the bus has **no** `NewWithTx`, add one with this body. (Confirm the storer field is named `storer`; if not, override the actual field name.)
2. **Wrap each emitting method:** for every method whose body contains `b.outbox.Emit(ctx,`, keep any leading `otel.AddSpan`/`defer span.End()` lines, then wrap the rest in `return outbox.WriteAtomic(ctx, b.outbox, b, (*Business).NewWithTx, func(ctx context.Context, b *Business) (<RetType>, error) { <existing body> })` (use `WriteAtomicVoid` + `func(ctx, b) error` for methods returning only `error`).
3. **Imports:** the bus already imports `outbox` and `sqldb` (it has the `outbox.Writer` field and the `sqldb.CommitRollbacker` param). No new imports.
4. **Build + per-package test:** `go build -C <worktree> ./business/domain/<area>/<bus>/...` then run that package's tests.

**Cascade bus list (from `dbtest.go:352-472` `.WithOutbox` + `all.go`).** Wrap every one EXCEPT the out-of-scope set named at the top. Confirm the live list with:
`grep -n '.WithOutbox(' business/sdk/dbtest/dbtest.go` (the authoritative wired set).

approvalbus, userbus, commentbus, homebus, citybus, streetbus, timezonebus, approvalstatusbus, fulfillmentstatusbus, assetconditionbus, assettypebus, validassetbus, tagbus, assettagbus, titlebus, reportstobus, officebus, userassetbus, assetbus, contactinfosbus, customersbus, brandbus, productcategorybus, productbus, physicalattributebus, productcostbus, costhistorybus, warehousebus, zonebus, inventorylocationbus, inventoryitembus, purchaseorderlineitemstatusbus, purchaseorderstatusbus, supplierbus, supplierproductbus, purchaseorderbus, purchaseorderlineitembus, metricsbus, inspectionbus, lottrackingsbus, serialnumberbus, lotlocationbus, rolebus, pagebus, paymenttermbus, rolepagebus, userrolebus, tableaccessbus, inventorytransactionbus, inventoryadjustmentbus, putawaytaskbus, picktaskbus, cyclecountsessionbus, cyclecountitembus, transferorderbus, orderfulfillmentstatusbus, lineitemfulfillmentstatusbus, ordersbus, orderlineitemsbus, formfieldbus, formbus, pagecontentbus, pageactionbus, pageconfigbus, scenariobus, labelbus.

(currencybus done in Task 3. `productuombus`/`settingsbus`/`approvalrequestbus` are NOT cascade buses — they are the documented coverage-test exclusions — skip them.)

**Per-bus watch-items (verify, do not assume):**
- **Cache-wrapped stores** (e.g. `currencybus`/`productbus` use a `*cache.NewStore` in dbtest). Confirm the cache store's `NewWithTx` threads the tx to its underlying db store (so the entity write rides the tx, not just the cache). If a cache `NewWithTx` is a no-op/returns itself, fix it the same copy-then-override way — flag any found.
- **Sibling-bus writes inside a method:** if an emitting method writes via *another* bus/store (not just `b.storer` + `b.outbox`), that inner write must also ride the tx (the inner bus would need rebinding). Grep each emitting method for other `*bus.`/`*.Create(`/`*.Update(` calls; flag any — they are a pre-existing atomicity question this wrap does not fix.
- **`homebus`/`scenariobus`/`labelbus`** have non-uniform `NewBusiness` constructor arities — that affects construction, NOT the `NewWithTx`/wrap recipe, which is uniform. No special handling needed for the wrap.

- [ ] **Step 1: Partition the bus list by area and fan out the mechanical edits**

Use spawn-agents (one agent per area: core, inventory, sales, procurement, assets, config, hr, finance). Each agent applies the recipe to its buses and runs `go build -C <worktree> ./business/domain/<area>/...`. Each returns: buses touched, any watch-item flags (cache no-op NewWithTx, sibling-bus writes, missing storer field).

- [ ] **Step 2: Integrate — full build**

Run: `go build -C /Users/jaketimmer/src/work/superior/ichor/ichor-cascade-pathA-lostcascade ./...`
Expected: clean. Fix any compile errors (most likely: a method with a different return signature, or a non-`storer` field name).

- [ ] **Step 3: Run the changed packages' tests (per area, NOT `./...`)**

Run each area's bus tests, e.g.:
`go test -C /Users/jaketimmer/src/work/superior/ichor/ichor-cascade-pathA-lostcascade ./business/domain/inventory/... ./business/domain/sales/... ./business/domain/core/...`
Expected: PASS. Investigate any failure as a real regression (CLAUDE.md testing rule).

- [ ] **Step 4: Commit (one commit per area, or one sweep commit)**

```bash
git -C /Users/jaketimmer/src/work/superior/ichor/ichor-cascade-pathA-lostcascade add business/domain/
git -C /Users/jaketimmer/src/work/superior/ichor/ichor-cascade-pathA-lostcascade commit -m "feat(cascade): wrap all cascade-bus writes in WriteAtomic; harden NewWithTx (FF#2)"
```

---

## Task 6: Coverage-test enforcement

Extend the existing every-cascade-bus-emits test so a future cascade bus that forgets the atomic wrap is caught RED. The current test (`coverage_test.go:37-121`) greps for `.outbox.Emit(ctx,`; add a guard that every emitting package also routes through `WriteAtomic`/`WriteAtomicVoid`.

**Files:**
- Modify: `business/sdk/outbox/coverage_test.go`

- [ ] **Step 1: Write the new guard (RED first)**

In the `WalkDir` callback (after the existing `emitting[pkg]=true` detection at `:80-82`), add detection of the wrapper, then a third guard after Guard 2:

```go
// (in the WalkDir callback, alongside the emitting/firing detection)
if strings.Contains(s, "outbox.WriteAtomic(") || strings.Contains(s, "outbox.WriteAtomicVoid(") {
	wrapped[pkg] = true // declare `wrapped := map[string]bool{}` next to `emitting`
}

// (new Guard 3, after Guard 2)
var unwrapped []string
for pkg := range emitting {
	if excluded[pkg] || wrapped[pkg] {
		continue
	}
	unwrapped = append(unwrapped, pkg)
}
sort.Strings(unwrapped)
if len(unwrapped) > 0 {
	t.Fatalf("FF#2 atomicity gap: these cascade buses emit to the outbox but do NOT wrap "+
		"their writes in outbox.WriteAtomic/WriteAtomicVoid — a simple-write emit failure would "+
		"silently lose the cascade. Wrap each emitting method:\n  %s", strings.Join(unwrapped, "\n  "))
}
```

- [ ] **Step 2: Run — expect GREEN now (Task 5 wrapped them all)**

Run: `go test -C /Users/jaketimmer/src/work/superior/ichor/ichor-cascade-pathA-lostcascade ./business/sdk/outbox/ -run TestCoverage -v`
Expected: PASS. If `unwrapped` is non-empty, a bus was missed in Task 5 — wrap it, do not weaken the guard.

- [ ] **Step 3: Prove the guard bites (temporary RED check)**

Temporarily revert ONE bus's wrap (e.g. comment the `WriteAtomic` on a small bus), re-run: expect FAIL naming that package. Restore the wrap. (This proves the guard is not vacuous — do not commit the revert.)

- [ ] **Step 4: Commit**

```bash
git -C /Users/jaketimmer/src/work/superior/ichor/ichor-cascade-pathA-lostcascade add business/sdk/outbox/coverage_test.go
git -C /Users/jaketimmer/src/work/superior/ichor/ichor-cascade-pathA-lostcascade commit -m "test(outbox): coverage guard — every cascade bus wraps writes atomically (FF#2)"
```

---

## Task 7: Regression + full verification

Prove the join path didn't regress the self-tx handlers, and the whole thing builds.

**Files:** none new (verification gate).

- [ ] **Step 1: Run the self-tx / phantom-rollback regressions (join path)**

These exercise cascade buses called UNDER a caller tx (the join path). They must stay GREEN — proving the wrap joins, never nests:

Run:
```
go test -C /Users/jaketimmer/src/work/superior/ichor/ichor-cascade-pathA-lostcascade \
  ./api/domain/http/inventory/transferorderapi/... \
  ./app/domain/sales/pickingapp/... \
  ./api/cmd/services/ichor/tests/workflow/actionhandlers/...
```
Expected: PASS — including the T3 phantom-rollback test (`transferorderapi/atomicity_test.go`) and the cascade composition/RYW tests. Any failure here means the join path nested or double-committed — STOP and fix `WriteAtomic`/`BeginOrJoin`.

- [ ] **Step 2: Run the outbox + sqldb suites**

Run: `go test -C /Users/jaketimmer/src/work/superior/ichor/ichor-cascade-pathA-lostcascade ./business/sdk/outbox/... ./business/sdk/sqldb/...`
Expected: PASS (Tasks 1, 2, 6 + existing poison/coverage tests).

- [ ] **Step 3: Full build**

Run: `go build -C /Users/jaketimmer/src/work/superior/ichor/ichor-cascade-pathA-lostcascade ./...`
Expected: clean.

- [ ] **Step 4: Final commit (if any verification fixups were needed)**

```bash
git -C /Users/jaketimmer/src/work/superior/ichor/ichor-cascade-pathA-lostcascade add -A
git -C /Users/jaketimmer/src/work/superior/ichor/ichor-cascade-pathA-lostcascade commit -m "test(cascade): FF#2 regression + verification green"
```

- [ ] **Step 5: STOP — do not push.** Report results to the user. Ship (github PR → rebase-merge → `git fetch github && git push origin github/master:master` → ff local master + `/merge-worktree`) only after the user confirms.

---

## Self-Review

**Spec coverage (vs `FF2_PATHA_LOSTCASCADE.md`):**
- Begin-or-join primitive (M2's missing accessor) → Task 1. ✓
- `WriteAtomic` wrapper hosted on `outbox.Writer` (no new bus wiring) → Task 2. ✓
- join-not-nest intrinsic (caller tx → JOIN, never commit) → Task 1 logic + Task 7 Step 1 regression. ✓
- Decisive RED-first: entity rolls back when emit fails on the begin path → Task 3 Steps 1–5. ✓
- A→B→A cascade survives the new path → Task 4. ✓
- All ~62 cascade buses (scope = "all") → Task 5. ✓
- Enforcement via the existing coverage test → Task 6. ✓
- Prerequisite surfaced during planning (the `NewWithTx` dropped-`del` class) → Task 3 Step 3 + Task 5 recipe Step 1. ✓

**Placeholder scan:** the two runtime-schema facts deferred (currencies.created_by FK; `db.Log` capture API) are flagged inline with concrete fallbacks, not left as TODOs. The Task 5 sweep uses a complete recipe + the full bus list (a mechanical repeat, not "similar to Task N").

**Type consistency:** `BeginOrJoin(ctx, Beginner) (ctx, CommitRollbacker, bool, error)` consumed by `WriteAtomic` exactly; `newWithTx func(B, sqldb.CommitRollbacker) (B, error)` matches the verified-uniform `(*Business).NewWithTx` shape; `WriteAtomicVoid` delegates to `WriteAtomic` with `struct{}`. The fake-bus test (Task 2) and real-bus wrap (Task 3) use the same method-expression form.

**Open risks carried into execution (flagged in-task, not hidden):** cache-store `NewWithTx` no-op; sibling-bus writes inside an emitting method; `created_by` FK; test-DB isolation (Task 0 gate). Each has a concrete check + fallback in its task.
