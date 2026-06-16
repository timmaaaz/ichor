# DESIGN — F2: Transactional outbox for cascade delegate events

> Created 2026-06-16. Branch `feature/cascade-f2-outbox` off master `f98c1b6b`.
> Pays down the v1-accepted debt named in CASCADE_DELEGATE_EVENTS: FOLLOW_UP §3,
> DESIGN §10 (decision §211 "best-effort for v1"), INVESTIGATION §13.1–13.2.
> Hardens **delivery guarantees only** — cascade *semantics* (which events fire,
> which rules match, the loop guard) are unchanged.

---

## 1. The problem (verified, file:line)

A cascade event is a **dual write** with no atomicity between the two halves:

1. A bus method writes the entity inside a tx.
2. `delegate.Call(ctx, data)` fans out in-process → `DelegateHandler.handleEvent`
   → a **detached `go func()` on `context.Background()`** (`delegatehandler.go:94-100`)
   → `WorkflowTrigger.OnEntityEvent` (`trigger.go:102`) → `ExecuteWorkflow`.

Two failure modes (INVESTIGATION §13.1–13.2):

- **Swallowed errors.** `Delegate.Call` (`delegate.go:48`) logs any handler error and
  **`return nil` unconditionally (`delegate.go:59`)**. The real loss point is deeper: the
  dispatch goroutine's `OnEntityEvent` error is only logged (`delegatehandler.go:97`). A crash
  or error between commit and dispatch silently drops the cascade; the write still "succeeded."
- **Read-your-writes.** The goroutine runs on a fresh `context.Background()`, racing the commit.
  A cascaded rule can query the DB **before the originating tx commits** (cited:
  `allocate.go:488` commits *after* the bus write at `inventoryitembus.go:157` fired the event).

### The three write paths (where the tx lives differs — drives the carrier work)

| Path | Trigger | tx situation | emit timing |
|------|---------|--------------|-------------|
| **A** human HTTP | `mid.BeginCommitRollback` (`transaction.go:18→36→45`): `Begin → setTran(ctx,tx) → next() → Commit`. Bus (`ordersbus.go:134→139`) `storer.Create` then `delegate.Call`. | tx on ctx (under `mid`'s private key) | inside tx, before commit |
| **B** workflow handlers | 6 handlers: `BeginTxx → NewWithTx(tx) → bus.Create → Commit` (`createpo.go:385/393/444`, `receive.go`, `reserve_inventory.go`, `release_reservation.go`, `allocate.go:425/488`, `commit_allocation.go`) | tx in the `NewWithTx` bus, **not on ctx** | inside tx, before commit |
| **C** synthesized generic writes | `create.go:138 / updatefield.go:276 / transition.go:172` write on raw `*sqlx.DB`, then `synthesize.go:42` | **no tx at all** (pooled auto-commit) | strictly after the auto-committed write |

### The delegate is a CYCLE-BREAKER, not an event bus (the principle that shapes this design)
`business/sdk/delegate/delegate.go` package doc, verbatim:
> *"Package delegate provides the ability to make function calls between different domain
> packages when an import is not possible."*

It is an indirect, **best-effort, in-process** function call whose sole purpose is dodging import
cycles. It is **not** a message bus, durable event log, or transaction participant. It has **four
kinds of subscribers**, most of them legitimately best-effort:

| Subscriber | Listens for | Contract |
|---|---|---|
| `temporal/delegatehandler.go:35-41` | all domains | cascade dispatch (the one F2 moves) |
| `core/permissionsbus/permissionsbus.go:56-73` | role / tableaccess / userrole | best-effort permission-cache recompute |
| `http/workflow/alertws/delegate.go:28` | userrole created | best-effort WebSocket notify |
| `sdk/workflow/trigger.go:514-518` | rule lifecycle | best-effort rule-cache reload |

→ Making `delegate.Call` propagate errors **globally** would wrongly couple a write to a
permission-cache or WebSocket hiccup. The delegate's best-effort contract is *correct* for those.

### The carrier that must survive
The loop guard rides a Go `context.Value`: `WorkflowLineage{Visited[], OriginatingExecutionID}`
(`lineage.go:49`), captured at `delegatehandler.go:91`. It JSON-round-trips already
(`trigger.go:215`, `lineage.go:118`, key `CascadeLineageKey` `lineage.go:41`). The outbox must
**serialize lineage into the row and re-hydrate it before dispatch**, or the guard dies.

### Precedent
No outbox / poller / `SKIP LOCKED` / `pg_notify` anywhere in app code. Greenfield.

---

## 2. Decisions ledger

**`[DECIDED]` with the user (2026-06-16):**

- **Build the outbox** (not a broker). Rabbit/Kafka is a deliberate "future day" swap — the
  relay→`OnEntityEvent` boundary stays clean so a broker drops in as a different relay against the
  same table + same `OnEntityEvent`, no rewrite.
- **Scope = full uniform.** Every cascade-relevant domain write emits to the outbox (human Path A +
  workflow Path B/C). Default everything through; narrow only with a concrete reason.
- **Relay = polling publisher.** Simplest, textbook, zero new infra. Sub-second latency is
  irrelevant for an async engine. LISTEN/NOTIFY is a latency optimization deferrable without
  touching the table, carrier, or seam.
- **Dedup = Temporal workflow-ID keyed on the outbox event id.** At-least-once *emission* +
  effectively-once *execution*.
- **★ Placement = the Ardan-principled one (revised 2026-06-16).** The outbox write is **explicit
  persistence in the unit of work**, NOT buried inside the best-effort `delegate.Call`. The
  delegate stays **untouched** as the cycle-breaker it is (its permissions/alertws/rule-reload
  subscribers are correct best-effort uses); only its **cascade subscriber is removed**. Atomicity
  is expressed the way the whole codebase expresses it — the operation **returns an error** and
  `mid.BeginCommitRollback` rolls back. Postgres tx-poison (§4) is a *backstop*, not the mechanism.
- **Volume guard = v1 write-always (simple).** Write a row for every cascade-relevant event; rely
  on delete-on-publish. A coarse "bouncer" pre-filter (skip if no active rule listens for
  `(domain,event_type)`) is a perf optimization **deferred until churn is observed** — captured in
  follow_up + a `// bouncer:` marker at the emit seam. Correctness identical either way.
- **Cleanup is a first-class deliverable** (§7). **Testing is TDD-first / prove-RED** (§8).
- Sub-decisions (recommended, not vetoed): relay **server-only** v1; **delete-on-publish** +
  **reaper** (7-day dead-row window); poll **500ms**, `ICHOR_*`-configurable.

**`[OPEN]`** — none blocking. LISTEN/NOTIFY fast-path, worker-side relay (HA), and the bouncer
pre-filter are post-v1 knobs.

---

## 3. Architecture

```
WRITE SIDE (cascade-relevant bus Create/Update/Delete, any of the 3 tx paths)
  storer.Create/Update/Delete(ctx, entity)                       ← entity write
  if err := b.outbox.Emit(ctx, data); err != nil { return ... }  ← EXPLICIT persistence, same tx,
        tx := sqldb.GetTx(ctx)                                       error PROPAGATED
        INSERT workflow.cascade_outbox (...)
  delegate.Call(ctx, data)   ← UNCHANGED, best-effort; only fires for domains that still
                               have a best-effort subscriber (permissions/alertws/rule-reload)
  → tx.Commit()  — entity row + outbox row commit together, or both roll back (returned error)

RELAY SIDE (one polling loop, server)
  every ~500ms:
    SELECT ... WHERE published_at IS NULL AND NOT dead ORDER BY seq LIMIT N FOR UPDATE SKIP LOCKED
    for each row in seq order:
       event   = rebuild workflow.TriggerEvent from row (the extractEntityData/computeFieldChanges
                 logic RELOCATED here from delegatehandler.go); event.EventID = row.id
       lineage = decode(row.lineage); ctx = contextWithLineage(ctx, lineage)
       err := WorkflowTrigger.OnEntityEvent(ctx, event)   ← the ONLY dispatcher now
       on success: DELETE row     on error: attempts++, last_error; dead=true after N
    reaper: DELETE dead rows older than the window
```

**What moves:** the cascade subscriber comes **off** the delegate; the `TriggerEvent` enrichment
relocates from `DelegateHandler` into the relay. **What stays:** the delegate (best-effort
cycle-breaker), `OnEntityEvent` and everything downstream, the loop guard semantics.

---

## 4. Atomicity — the idiomatic mechanism + the backstop

**Mechanism (idiomatic):** `outbox.Emit` writes the row on the originating tx and **returns its
error**; the bus method propagates it (`return err`); `mid.BeginCommitRollback` commits only if
`isError(resp)` is nil (`transaction.go:40-47`) → on error it rolls back **both** writes. This is
the exact path every other write in the codebase uses. Honest, explicit, type-enforced.

**Backstop (defense in depth):** because the `INSERT` shares the tx, even if a caller ever failed
to propagate, a failed `INSERT` poisons the tx in Postgres → the eventual `COMMIT` downgrades to
`ROLLBACK`. So atomicity holds via two independent guarantees. We do **not** rely on this as the
primary mechanism, and we do **not** change `delegate.Call`'s 197-site contract.

**Cost (inherent to any transactional outbox):** the outbox table moves onto the **critical path
of every cascade-relevant write** — if it's unwritable, those writes fail. It's a trivial local
`INSERT` that fails only under serious DB trouble (which would fail the entity write anyway).

### The tx-on-ctx carrier (layer-pure)
`sqldb.WithTx(ctx, *sqlx.Tx)` / `sqldb.GetTx(ctx)` (`sqldb/context.go:13/18`) already exist in the
**business** layer but are **dead code**. We **activate** them as the carrier `outbox.Emit` reads
(a `*sqlx.Tx` *is* an `sqlx.ExtContext`) — no `business → app/sdk/mid` import. Populate at the 3 tx
origins:
- **Path A** (`mid/transaction.go`): `bgn.Begin()`'s `CommitRollbacker` is concretely `*sqlx.Tx`
  (`tran.go:37`) → type-assert + one `sqldb.WithTx` line.
- **Path B** (6 `BeginTxx` handlers): `ctx = sqldb.WithTx(ctx, tx)` after `BeginTxx`.
- **Path C** (3 synthesize sites): wrap write+emit in a tx, set `WithTx`.

`sqldb.GetTxExecutor(ctx) (sqlx.ExtContext, bool)` convenience; if no tx, `Emit` falls back to the
base pool **with a loud warning** (degraded; never expected for the 3 covered paths) — the §8
"on-a-tx" trip-wire catches a path that forgets to populate it.

### What failures are visible (F2 surfaces more, hides nothing)
- **Source write + outbox INSERT can't commit** → tx rolls back → the user's action **fails
  visibly** (returned error); they retry. (Today: action succeeds, cascade silently vanishes.)
- **Event committed but dispatch keeps failing** → row **retried**, then `dead=true` + `last_error`
  + loud log — a durable, queryable record. (Today: silently dropped.)
- **Workflow runs, an action inside fails** → *execution* outcome, tracked where it lives:
  Temporal + `workflow.automation_executions`. F2 only guarantees the workflow *gets started*.
F2 owns **delivery**; rich user-facing surfacing of execution outcomes is **F8** + execution
history (§9).

---

## 5. Outbox table + dedup

### Migration 2.42 (append-only; master=2.40, buttons branch=2.41 — not in this worktree)
```sql
-- Version: 2.42
-- Description: F2 transactional outbox for cascade delegate events. One row per cascade-relevant
--   domain event, written in the SAME tx as the entity write; a polling relay drains it into
--   WorkflowTrigger.OnEntityEvent at-least-once. id doubles as the Temporal workflow-id dedup key;
--   seq gives total order; lineage carries the loop-guard visited-set (and is F8's traceparent slot).
CREATE TABLE workflow.cascade_outbox (
    id            UUID        PRIMARY KEY,        -- durable event id = Temporal dedup key
    seq           BIGSERIAL   NOT NULL UNIQUE,    -- total order for the relay
    domain        TEXT        NOT NULL,           -- delegate.Data.Domain
    action        TEXT        NOT NULL,           -- delegate.Data.Action (e.g. orders.updated)
    event_type    TEXT        NOT NULL,           -- on_create | on_update | on_delete
    entity_name   TEXT        NOT NULL,
    payload       JSONB       NOT NULL,           -- delegate.Data (Domain/Action/RawParams) for relay enrichment
    lineage       JSONB,                          -- serialized WorkflowLineage (loop guard; F8 traceparent)
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    attempts      INT         NOT NULL DEFAULT 0,
    last_error    TEXT,
    published_at  TIMESTAMPTZ,                    -- NULL = pending
    dead          BOOLEAN     NOT NULL DEFAULT false
);
CREATE INDEX idx_cascade_outbox_pending
    ON workflow.cascade_outbox (seq)
    WHERE published_at IS NULL AND dead = false;
```

### Dedup mechanics
- Add `EventID uuid.UUID` to `workflow.TriggerEvent`; the relay sets it to `row.id`.
- `trigger.go` (~`:170`) derives the workflow ID `workflow-{ruleID}-{eventID}` (replacing the
  random `executionID`), `WorkflowIDReusePolicy = REJECT_DUPLICATE`. Re-published row → same ID →
  Temporal rejects the duplicate. Zero `EventID` (direct `OnEntityEvent` test callers) →
  `uuid.New()` fallback (no test rewrite forced). **Verify nothing else parses the random id first.**

### Ordering / failure / retention
- **Ordering:** single relay, `ORDER BY seq` → per-entity causal order. `SKIP LOCKED` keeps a 2nd
  process safe.
- **Failure/poison:** `attempts` + `last_error`; `dead=true` after N → skip (no head-of-line block).
- **Retention:** delete-on-publish; reaper deletes `dead` rows past the window (7d). `pending-age`
  gauge surfaces a stuck relay.

---

## 6. The cutover (atomic gate — no double-dispatch window)

Two cascade paths must NEVER be live at once (the old delegate dispatch + the new relay) — they'd
double-fire (different event ids → dedup won't collapse them). So `outbox.Emit` is **inert until
cutover**: the buses get the `Emit` call early (parallel-safe), but the injected `outbox.Writer` is
nil/disabled until one wiring change flips everything together:

**Cutover commit (`all.go` + worker):** inject the real `outbox.Writer` (Emit goes live) **+** start
the relay **+** remove the cascade subscriber from the delegate registration — atomically, in one
change. Mirrors the parent plan's P4 gate ("no armed-but-undrained window"). Before cutover:
`Emit` no-ops, cascades flow via the delegate as today. After: cascades flow via outbox+relay only.

---

## 7. Cleanup (first-class deliverable)

1. **Remove the cascade subscriber from the delegate** (the `RegisterDomain` cascade closures) at
   cutover — the delegate returns to being *only* the cycle-breaker its doc describes. Its
   best-effort subscribers (permissions/alertws/rule-reload) are untouched.
2. **Relocate** `extractEntityData`/`computeFieldChanges` from `DelegateHandler` into the relay;
   delete now-dead `DelegateHandler` dispatch code (goroutine, `WorkflowTrigger` field).
3. **Reduce tx-plumbing duplication** — back `mid.GetTran` with `sqldb.GetTx` (one source of truth)
   after verifying `mid.GetTran`'s callers, so F2 *reduces* tx-on-ctx duplication, not adds.
4. **Table self-cleans** — delete-on-publish + reaper (§5).
5. **Final dead-code sweep** — grep for the old goroutine / handler-side `OnEntityEvent` / dropped
   field; confirm zero residue before ship.

---

## 8. Testing strategy (TDD-first, prove-RED)

Reliability guarantees are only real if the prevented failure is demonstrable. Write the
reliability tests first, prove RED on master, implement to GREEN. Changed packages only —
**never `go test ./...`** (CLAUDE.md). Unique Temporal task queue per test; scoped queries
(`arch/testing.md`); trust `go build -C <worktree>` over gopls.

### Unit (`business/sdk/outbox/*_test.go`)
- **Emit:** writes on the ctx tx; pool fallback + warn when absent; returns real error (propagated).
- **Store:** insert; fetch-pending ordered by `seq`; delete-on-publish; mark-dead; reaper; SKIP
  LOCKED non-double-processing.
- **Relay enrichment:** rebuilds `TriggerEvent` from a stored row identically to the old
  `DelegateHandler` path (golden compare); lineage round-trips.
- **Dedup:** deterministic workflow id from `(ruleID, eventID)`; zero `EventID` → `uuid.New()`.

### Integration (DB + Temporal; extend the `cascade_lineage_test.go` manual-worker rig)
- **Atomicity (both directions):** (a) a bus write whose tx rolls back leaves **zero** outbox rows
  and fires no cascade; (b) a forced `Emit` error makes the bus return error → the entity write
  rolls back (explicit propagation; the poison backstop is also asserted).
- **Read-your-writes:** a Path-B `allocate` cascade reads the **committed** entity. RED on master, GREEN after.
- **At-least-once + dedup:** re-process the same row id → **exactly one** workflow runs.
- **Lineage / loop-guard survives the round-trip:** A→B→A **stops** across serialize+rehydrate
  (build on `cascade_lineage_test.go:63-105`). The single most important regression.
- **Ordering:** two updates to one entity → cascades fire in `seq` order.
- **End-to-end Path A:** human write → committed row → relay → cascade runs → row deleted.
- **Reaper:** aged `dead` rows swept; younger retained.

### Trip-wires (prove-RED discipline)
- **No double-dispatch:** after cutover, assert the delegate no longer dispatches cascade (only the
  relay does, only after commit).
- **Coverage:** every domain in `workflowdomains.Registrations()` emits to the outbox — a
  consistency test (mirrors `manifest_consistency_test.go`) so a bus that gets the cascade
  subscriber removed but **forgets `outbox.Emit`** goes RED (its cascade would silently vanish).
- **On-a-tx:** each of the 3 tx paths lands the row on a tx, not the pool fallback.

---

## 9. F8 hook (observability — sibling follow-up, do NOT build here)

The `lineage JSONB` column IS the extensible carrier FOLLOW_UP §10 reserved for F8's `traceparent`.
F8 will: add `traceparent` to `WorkflowLineage`, inject at the emit seam, extract + start a child
span at the relay's re-hydration point (one span/hop) + a dropped/dead-row metric. F2 leaves both
seams in place; the relay's `pending-age` gauge is the first such metric.

---

## 10. File manifest

**New**
- `business/sdk/migrate/sql/migrate.sql` — append `-- Version: 2.42`. *(modify-append)*
- `business/sdk/outbox/model.go` — `Outbox` row + `delegate.Data`/lineage (de)serialize.
- `business/sdk/outbox/store.go` (or `stores/outboxdb/`) — insert / fetch-pending / delete / mark-dead / reap.
- `business/sdk/outbox/emit.go` — `Writer.Emit(ctx, data) error` (tx-from-ctx; nil/disabled = no-op until cutover).
- `business/sdk/workflow/temporal/relay.go` — polling loop + reaper + TriggerEvent enrichment + dispatch via `WorkflowTrigger`.

**Modify — serial spine**
- `business/sdk/sqldb/context.go` — `GetTxExecutor(ctx)` convenience.
- `app/sdk/mid/transaction.go` (+ `mid.go`) — Path A `sqldb.WithTx`; back `GetTran` with `sqldb.GetTx`.
- `business/sdk/workflow/temporal/delegatehandler.go` — relocate enrichment to relay; delete goroutine + `WorkflowTrigger` dep; remove cascade `RegisterDomain` at cutover.
- `business/sdk/workflow/temporal/trigger.go` + `TriggerEvent` model — `EventID` + deterministic workflow id + reuse policy.
- `api/cmd/services/ichor/build/all/all.go` — build `outbox.Writer`, inject into buses; **cutover**: enable Writer + start relay + drop cascade subscriber registration.
- `api/cmd/services/workflow-worker/main.go` — inject `outbox.Writer` into the worker's buses; drop cascade subscriber registration at cutover.

**Modify — parallel fan-out (spawn-agents, partitioned by package)**
- The ~58 cascade-relevant domain buses (the `workflowdomains.Registrations()` set):
  `b.outbox.Emit(ctx, data)` + `return err` in Create/Update/Delete; keep `delegate.Call` only
  where a best-effort subscriber exists (role / tableaccess / userrole), else it becomes dead → drop.
- `business/sdk/workflow/workflowactions/{procurement/createpo, inventory/receive, reserve_inventory, release_reservation, allocate, commit_allocation}.go` — Path B `sqldb.WithTx`.
- `business/sdk/workflow/workflowactions/data/{create,updatefield,transition}.go` (+ `synthesize.go`) — Path C tx wrap + `WithTx`.
- `docs/arch/delegate.md` + `docs/arch/workflow-engine.md` — document the outbox seam, relay, dedup, and that the delegate is cycle-break-only again.

---

## 11. Execution model (serial spine → parallel fan-out)

The work splits into a small interdependent spine and a large embarrassingly-parallel tail
(per the standing preference: don't serialize dozens of file edits — fan out via `/spawn-agents`).

1. **Serial spine (main context):** migration → `sqldb` helper → `business/sdk/outbox` (store +
   Emit + model) → relay → `trigger.go` dedup → the `mid`/Path-A carrier. Build + unit-test as one
   coherent unit. Everything below depends on these seams existing.
2. **Parallel fan-out (`/spawn-agents`, opus, one package group per agent):** add `outbox.Emit` +
   error propagation across the ~58 cascade buses; Path-B (6 handlers) and Path-C (3 handlers)
   `WithTx`. Each agent builds its own packages. Inert until cutover (Writer not yet injected).
3. **Cutover gate (main context, one commit):** inject Writer + start relay + remove cascade
   subscriber. Then integrating build + the §8 reliability/trip-wire sweep.

---

## 12. Risks / watch-items

- **Double-dispatch window** — mitigated by the inert-Emit + atomic cutover (§6); the no-double-
  dispatch trip-wire (§8) guards it.
- **Forgot-to-Emit a cascade domain** → its cascade silently vanishes after the subscriber is
  removed → the §8 coverage consistency test (over `Registrations()`) catches it RED.
- **gopls phantom errors in worktrees** — trust `go build -C <worktree>`/`vet`/`test`.
- **Migration** — outbox is **2.42**; never edit existing; `darwin` orders by `-- Version:`.
- **`trigger.go` workflow-ID change** — verify nothing parses the random `executionID` first.
- **Non-tx bus paths** (e.g. `create_put_away_task`) — Emit hits the pool fallback (not atomic);
  the on-a-tx trip-wire flags them; making them transactional is a contained follow-up if needed.
- **Worker relay absence** — server down → rows accumulate (correct, at-least-once), drain on
  return; worker relay is a post-v1 HA knob.
