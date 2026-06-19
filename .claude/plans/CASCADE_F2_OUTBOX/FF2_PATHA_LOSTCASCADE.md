# FF#2 — Path-A simple-write lost-cascade (the last cascade reliability gap)

> Design doc for sign-off. Branch `feature/cascade-pathA-lostcascade` off master `73e1c3bd`
> (post-F9 + FF#1/#3/#4/#5). **No implementation code is written yet** — this doc proposes the
> mechanism, analyzes the fork, recommends one, and lays out the test plan. Wire nothing until approved.
>
> Implements fast-follow #2 of `F9_RESULTS.md`; honors the settled decision in
> `F9_ATOMICITY_AND_TEST_GAPS.md` Part A (the `✅ DESIGN DECISION — RESOLVED 2026-06-17` block).

---

## 0. TL;DR

- **The gap (fail-SAFE, real):** a cascade bus's entity write and its `outbox.Emit` are two **separate
  statements**. On a Path-A human HTTP single-bus write there is no surrounding tx, so both autocommit
  independently on the base pool. If `Emit` fails *after* the entity commits, the cascade is **silently
  and permanently lost** (a missing allocation / status-transition / inventory side-effect, with nothing
  left to reconcile). It is never a phantom or a wrong write — that is why it was deferred out of F9.
- **Decisive new evidence:** db-stores bind their executor at construction; they do **not** read the tx
  from ctx. So merely putting a tx on ctx (what HTTP middleware does) routes **only the emit** onto the
  tx — the **entity write stays on the pool** unless the bus is rebound via `NewWithTx(tx)`. And the
  Ardan `executeUnderTransaction` rebind pattern **does not exist** in this codebase. This makes
  **Option A (scoped middleware) far larger than the `emit.go:62` comment implies.**
- **Recommendation: Option B — a `sqldb` begin-or-join primitive + a small bus-write wrapper, with the
  begin-or-join authority hosted on the `outbox.Writer`** (which already holds the base pool and is
  injected into exactly the cascade buses). join-not-nest is **intrinsic** (a `GetTx(ctx)` check), it
  composes correctly with all 18 self-tx callers (they set ctx-tx → detected → **joined**, never
  nested), and it is **enforced by the existing `Registrations()` coverage test**. No route enumeration,
  no exclusion list, no HTTP-layer change.
- **Test bar (RED-first, decisive):** inject an `Emit` failure on the new atomic path → prove the
  **entity write rolls back with it** (RED on master: entity committed + cascade lost; GREEN after);
  plus an A→B→A cascade through the new path proving the simple-write cascade now survives; plus the
  existing T3 phantom-rollback test stays GREEN (proves the join branch didn't regress the self-tx path).

---

## 1. The gap — verified mechanism (why it is real)

A cascade-relevant bus `Create/Update/Delete` does three things after building the entity:

```
b.storer.Create(ctx, ent)     // entity write  → through the storer's BOUND executor
b.outbox.Emit(ctx, data)      // cascade row    → through sqldb.GetTxExecutor(ctx), else base pool
b.delegate.Call(ctx, data)    // best-effort hook (swallows error) — unchanged
```

The two writes use **different tx mechanisms** (this is the whole gap):

| Write | How it gets its DB handle | file:line |
|---|---|---|
| entity (`storer.Create`) | the store's **bound** `db sqlx.ExtContext` field, set at `NewStore`/`NewWithTx`; never reads ctx | `…/stores/<e>db/<e>db.go` Store struct + `Create` pass `s.db` (e.g. `inventorytransactiondb.go:18-21,53`) |
| cascade (`outbox.Emit`) | `sqldb.GetTxExecutor(ctx)`; **no tx on ctx → base pool + loud warn** | `business/sdk/outbox/emit.go:116-121` |

**On a Path-A simple write** (e.g. `POST /core/currencies` → `currencyapp.Create` → pool-bound
`currencybus.Create`): the bus is pool-bound, so `storer.Create` autocommits the entity on the pool, and
`Emit` finds no ctx-tx and autocommits the outbox row on the pool as a **second, independent statement**
(`emit.go:118` warn fires). If that second statement fails — DB hiccup, constraint, connection drop —
the **entity is already durably committed and the cascade is gone forever.**

**Why F9 didn't close it:** F9 fixed the *phantom* (fail-DANGEROUS) direction by enrolling the tx on ctx
in the 10 self-tx app handlers (`ctx = sqldb.WithTx(ctx, tx)`), so their cascade emits ride the handler
tx and roll back together. Those handlers already hold a tx; Path-A simple writes do **not**. The
remaining lost-cascade is fail-SAFE (a *missed* cascade, never a phantom) and was explicitly deferred.

> Pre-F2 framing (do NOT call this an F2 regression): pre-F2 the cascade fired from a detached
> `go func()` on `context.Background()`, so a missed write was *also* possible — best-effort, often
> dropped. F2 made delivery durable; this fast-follow extends that durability to the one write shape
> (no-caller-tx) the F9 pass left on the pool.

---

## 2. Settled constraints (recap — not re-opened)

From the resolved DESIGN DECISION (`F9_ATOMICITY_AND_TEST_GAPS.md:69-98`) and the task brief:

- **REJECTED — the global flip:** wiring `mid.BeginCommitRollback` on all ~152 write routes. Over-broad
  behavior change; also splits entity-write/emit across two txns in the self-tx handlers (worse).
- **REJECTED — an *unconditional* bus-local tx** in the cascade bus methods. **M1 is measured "ever":**
  cascade-emitting bus methods ARE called under a caller-supplied tx — by the 10 self-tx app handlers
  AND the 10 workflow action handlers, all via `bus.NewWithTx(tx)` (see §4). An unconditional inner tx
  there **nests**: the inner commit makes the cascade durable BEFORE the outer write commits →
  re-creates a phantom-shaped split → makes the dangerous bug worse.
- **HARD INVARIANT:** when a caller tx is already on ctx, the write+emit must **JOIN it, never open a
  nested/inner tx. begin-or-JOIN, not begin-always.**

Per the spec's own decision tree: **M1="ever" ⇒ the fix must be a single begin-or-join authority** —
either `mid.BeginCommitRollback` scoped to the cascade routes (Option A) **or** a bus-layer begin-or-join
primitive (Option B). **M2 is measured "no accessor today":** `sqldb` exposes
`WithTx`/`GetTx`/`GetTxExecutor`/`WithCommitRollbacker` but **no begin-or-join** — confirmed absent
(`sqldb/context.go`, `sqldb/tran.go`). Whichever option wins must add one.

---

## 3. The decisive code findings (verified on `73e1c3bd`)

These four facts drive the whole recommendation:

**F-1 — Stores bind their executor; ctx-tx alone does NOT route the entity write.**
The store holds `db sqlx.ExtContext` set at `NewStore`/`NewWithTx` and passes `s.db` straight to
`NamedExecContext` (`inventorytransactiondb.go:18-21, 30-40, 53`). It never calls `GetTx`/`GetTxExecutor`.
→ **For the entity row and the outbox row to share one tx, the bus must be rebound via `NewWithTx(tx)`
(storer) AND the same `*sqlx.Tx` must be on ctx (`WithTx`) for `Emit`.** A ctx-only tx is insufficient.

**F-2 — One tx carrier already, but no begin-or-join.**
`sqldb` uses a single `txKey{}` (`context.go:10`). `mid.GetTran`, `sqldb.GetTx`, and `Emit`'s
`GetTxExecutor` all read it (the old separate mid `trKey` was removed in F2/§7.3). `BeginCommitRollback`
begins a `*sqlx.Tx` and publishes it via `setTran → WithCommitRollbacker → WithTx` (`app/sdk/mid/
transaction.go:18,36`; `mid.go:118-120`). **No `BeginOrJoin`/`EnsureTx` helper exists** — `sqldb.Begin`
(`tran.go:36-38`) unconditionally opens a new tx and ignores ctx.

**F-3 — The middleware-rebind pattern does NOT exist.** `grep executeUnderTransaction` → 0 hits.
`mid.GetTran` has **no production caller** (dead plumbing). The *only* ctx-tx→rebind site is
`app/sdk/formdataregistry/txbind.go:27-33` (`TxBind`, used solely by `formdataapp`). Every tx-using
handler is **self-tx**: `db.BeginTxx` + its own `bus.NewWithTx(tx)`. → **Scoped middleware (Option A)
cannot rely on any existing rebind; it would have to introduce per-handler rebinding across the simple
CRUD apps.**

**F-4 — `BeginCommitRollback` is wired to ZERO routes** (only its own def + the `emit.go:62` doc comment).
The global chain (`mux.go:112-116`) is `Otel/Logger/Errors/Metrics/Panics` — no tx middleware.

### Cascade-bus / route inventory (the seam surface)

- `workflowdomains.Registrations()` returns `[]EntityReg{Schema,Domain,Entity}` — **68 entries**
  (`workflowdomains.go:98-193`); a hand-coded literal keyed on bus `DomainName`/`EntityName` consts.
- **Cascade buses = `.WithOutbox(...)`-constructed buses** (+ `workflowBus` via `.WithOutboxEmitter`):
  ~62–65 domain buses + the `workflowBus` emitter. Wired identically in `all.go`,
  `workflow-worker/main.go`, and `dbtest.go`. (Pin the exact count at implementation from the
  `.WithOutbox(` call sites — it does not change the design.)
- **Route coverage:** **every domain cascade bus has a direct HTTP write route** (`api/domain/http/
  <area>/<entity>api/`). The **only route-less** cascade emitter is `allocation_results` (the
  `workflowBus` emitter), fired exclusively by the `allocate`/`reserve` workflow handlers — always under
  a self-tx → **already atomic, needs nothing.**
- **Two synthesized non-bus cascade paths** (`synthesize.go` for `update_field`/`create_entity`/
  `transition_status`; `allocation_results`) originate inside worker activity handlers that already
  `Begin*` + `sqldb.WithTx` → **already atomic.** The gap is **only** the Path-A human HTTP simple write.

### The 10 self-tx app handlers (HTTP-reachable — would be Option A's exclusion list)

| Handler | `BeginTxx` | `WithTx` | Route |
|---|---|---|---|
| `transferorderapp.Execute` | :284 | :292 | `inventory/transferorderapi` POST …/execute |
| `inventoryadjustmentapp.Approve` | :215 | :223 | `inventory/inventoryadjustmentapi` POST …/approve |
| `pickingapp.PickQuantity` / `ShortPick` | :125 / :306 | :133 / :314 | `sales/orderlineitemsapi` POST …/pick-quantity, …/short-pick |
| `cyclecountsessionapp.complete` | :128 | :136 | `inventory/cyclecountsessionapi` PUT (status→complete) |
| `picktaskapp.complete` | :164 | :172 | `inventory/picktaskapi` PUT (status→complete) |
| `putawaytaskapp.complete` | :145 | :153 | `inventory/putawaytaskapi` PUT (status→complete) |
| `pageactionapp.BatchCreate` | :294 | :302 | `config/pageactionapi` POST …/batch |
| `inspectionapp.Fail` | :214 | :222 | `inventory/inspectionapi` POST …/fail |
| `formdataapp.UpsertFormData` | :180 | :192 | `formdata/formdataapi` (FF#1: binds tx-bound buses via `TxBind`) |

Plus 10 workflow action handlers calling `sqldb.WithTx` (`data/{create,transition,updatefield}.go`,
`inventory/{allocate,commit_allocation,receive,release_reservation,reserve_inventory,createputawaytask}.go`,
`procurement/createpo.go`) — worker-side, **not** HTTP-reachable.

---

## 4. Option A — scoped `mid.BeginCommitRollback` on the cascade-emitting routes

**Mechanism.** Apply `mid.BeginCommitRollback` as a per-route middleware on the write routes (POST/PUT/
DELETE) of the cascade route packages. The middleware begins a request tx, publishes it on the sqldb
`txKey{}` (so `Emit` joins it), and commits iff the handler returns no error.

**Enumeration.** Derive the route set from the cascade buses: `Registrations()` / the `.WithOutbox`
call sites → the ~62–65 `<entity>api` route packages. (All domain cascade buses have a route; the
route-less `allocation_results` is excluded automatically.)

**What it actually costs (given F-1/F-3) — bigger than the `emit.go:62` comment implies:**

1. **Per-handler rebind is mandatory.** Middleware only puts the tx on ctx. Because stores bind their
   executor (F-1) and no app reads the ctx-tx to rebind (F-3), the **entity write would still autocommit
   on the pool** — middleware alone leaves the emit on the tx and the entity on the pool, which does not
   achieve atomicity. Each of the ~62 simple CRUD apps' `Create/Update/Delete` must gain a `TxBind`-style
   rebind (`if tx, ok := sqldb.GetTx(ctx); ok { bus = bus.NewWithTx(tx) }`) — i.e. introduce the
   `executeUnderTransaction` pattern the codebase has deliberately never used. **~3 layers × ~62 packages.**
2. **Config threading.** Each route package's `Config` must carry a `sqldb.Beginner` (none do today) to
   construct the middleware — touches every cascade `route.go` + the `build/all` wiring.
3. **Exclusion list (the nesting hazard).** The 10 self-tx routes (table in §3) must be **excluded**, or
   the request tx double-ups with the handler's own `BeginTxx`. Three of them (`cyclecountsession`/
   `picktask`/`putawaytask` `complete`) share their plain PUT update route, so the *whole* update route
   must be excluded — a hand-maintained allow/deny list with no compile-time backing.
4. **No enforcement.** A newly added cascade route silently misses the middleware+rebind; the existing
   bus coverage test won't catch it. A bespoke **route-coverage test** must be written and maintained.
5. **Doesn't cover the synthesized/worker paths** (acceptable — already atomic), but it means the
   guarantee lives in two different layers (HTTP for human writes, self-tx for worker writes).

**Verdict.** Honors join-not-nest only via a manually-maintained exclusion list. Five distinct kinds of
change spanning api + app layers, introducing a cross-cutting pattern against the codebase grain, with
no enforcement. The word "scoped" is misleading once F-1/F-3 are accounted for.

---

## 5. Option B — a `sqldb` begin-or-join primitive + a bus-write wrapper (RECOMMENDED)

**Where the authority lives:** the **business layer, on/next to the `outbox.Writer`** — the same place
the cascade emit lives, and the only seam that sees *both* the entity write and the emit.

**Two new pieces:**

**(a) The primitive (M2's missing accessor), in `sqldb`:**
```go
// BeginOrJoin returns the in-flight ctx tx if present (joined=true; caller must NOT commit it —
// the owner does), otherwise begins a fresh tx on bgn and puts it on ctx (joined=false; caller owns
// commit/rollback). This is the begin-or-JOIN authority the hard invariant requires.
func BeginOrJoin(ctx context.Context, bgn Beginner) (context.Context, CommitRollbacker, bool /*joined*/, error)
```
~12 lines, pure tx management — no knowledge of buses/storers. Satisfies join-not-nest by construction.

**(b) The bus-write wrapper, hosted by `outbox` (it holds the base pool — `Writer.db *sqlx.DB`, and is
injected into exactly the cascade buses):**
```go
// In business/sdk/outbox. Generic over the concrete bus type B and its result T.
// On a JOIN, the storer must also ride the ctx tx, so we rebind B via its NewWithTx even on the join
// path (robust against a cascade bus that calls another cascade bus without rebinding).
func WriteAtomic[B any, T any](
    ctx  context.Context,
    w    *Writer,                                   // nil Writer (pre-cutover/tests) → fn(ctx, self), no tx
    self B,
    newWithTx func(B, sqldb.CommitRollbacker) (B, error),
    fn   func(ctx context.Context, bus B) (T, error),
) (T, error)
```
Behavior: `sqldb.BeginOrJoin(ctx, w.Beginner())` → publish ctx-tx → `bus := newWithTx(self, tx)` (rebinds
the storer onto the tx) → `fn(ctx, bus)` (does `storer.Create` + `Emit`, both now on the tx) → **commit
iff we began it; rollback on error; never commit a joined tx.** `w.Beginner()` is a new one-liner on the
Writer (`return sqldb.NewBeginner(w.db)`).

**Each cascade bus `Create/Update/Delete` becomes (mechanical, ~1 wrap):**
```go
func (b *Business) Create(ctx context.Context, na NewEntity) (Entity, error) {
    return outbox.WriteAtomic(ctx, b.outbox, b, (*Business).NewWithTx,
        func(ctx context.Context, b *Business) (Entity, error) {
            return b.create(ctx, na)   // the current method body, factored into an unexported method
        })
}
```

**Why this is the right shape:**

- **join-not-nest is intrinsic (the hard invariant, for free).** The 10 self-tx app handlers + 10
  workflow handlers set ctx-tx (`sqldb.WithTx`) before calling the bus → `BeginOrJoin` detects it →
  **joins** → no inner tx, no phantom-shaped split. No exclusion list. This is *exactly* the "bus-layer
  begin-or-join primitive" branch the spec's M1="ever" decision tree points to.
- **Robust to nested cascade-bus calls.** Because the wrapper rebinds `self` via `NewWithTx` even on the
  join path, a cascade bus that internally calls another cascade bus can't accidentally leave the inner
  storer on the pool. (Verify during impl whether any such nested calls exist; the wrapper is safe either
  way.)
- **Enforced by an existing test.** `coverage_test.go` already enumerates every cascade bus from
  `Registrations()` and asserts it emits. Extend it (or add a sibling) to assert a **no-caller-tx write
  never lands on the pool fallback** — i.e. every cascade bus is wrapped. Compile-time-ish safety net; a
  new cascade bus that forgets the wrap goes RED. Option A has no equivalent.
- **One layer, minimal wiring.** No `Config` threading, no HTTP middleware, no `Beginner` field on 62
  buses — the wrapper reaches the pool through `b.outbox`, which the buses **already** hold. The only new
  wiring is the existing `.WithOutbox(...)` injection (unchanged call sites).
- **Covers the gap uniformly** (every cascade bus, any caller) and keeps the atomicity guarantee in the
  same layer as the cascade itself (business/sdk), consistent with the worker/synthesize paths.

**Honest cost.** ~62–65 buses × 3 methods get a one-line wrap + a body factored into an unexported
method (~190 mechanical edits). Large but **uniform, mechanical, parallelizable** (fan out by package),
and **enforced**. Contrast Option A's heterogeneous, unenforced, cross-layer churn. The churn here is
"touch each cascade bus once in a uniform way"; Option A's is "touch api Config + app rebind + route mw +
exclusion list + new test, each slightly different per domain."

**Minor, accepted behavior change.** On the *begin* path a simple write now runs inside an explicit tx
(vs autocommit) and `b.delegate.Call` fires pre-commit — which is **already** how it behaves on every
self-tx handler path today, so this makes simple writes *consistent* with tx writes, not novel. The
best-effort subscribers (permission cache / alertws / rule-reload) are unaffected by the microsecond
shift; none depend on a post-commit read.

---

## 6. Option C — make db-stores read ctx-tx (considered, REJECTED)

Make every store resolve its executor as `sqldb.GetTxExecutor(ctx)` else its bound pool. Then ctx-tx
becomes the single carrier for *both* storer and emit, and a begin-or-join anywhere upstream makes the
whole write atomic (and `NewWithTx` becomes redundant). **Rejected:** it changes executor resolution in
**~75 db store files** and alters write semantics for **all** writes, not just cascade ones — precisely
the "over-broad behavior change" the spec rejected for the global flip. F9 deliberately avoided it. Out
of scope for a fast-follow.

---

## 7. Recommendation & justification

**Adopt Option B.** Scored against the brief's axes:

| Axis | Option A (scoped middleware) | **Option B (begin-or-join, recommended)** |
|---|---|---|
| Blast radius | api `Config` threading + app rebind + route mw + exclusion list + new test, across ~62 pkgs × 2 layers | ~190 uniform one-line wraps in business layer + 1 primitive + 1 Writer method |
| Where the authority lives | HTTP transport (can't cover worker/synthesize) | business layer, next to the cascade emit (covers every caller) |
| join-not-nest (hard invariant) | manual exclusion list of 10 self-tx routes | **intrinsic** — `GetTx` check joins, never nests |
| Composition with self-tx handlers | must exclude them | they set ctx-tx → **auto-joined** |
| Enforcement | bespoke route-coverage test, easy to miss a new route | **existing `Registrations()` coverage test** extended; new bus → RED |
| Against/with the grain | introduces `executeUnderTransaction` the repo never used | reuses `NewWithTx` + the existing `.WithOutbox` injection |
| Satisfies spec decision tree | one of the two named M1="ever" remedies | the other — and the cleaner given F-1/F-3 |

Option B is the precise "bus-layer begin-or-join primitive" the resolved decision named, it satisfies the
hard invariant structurally rather than by hand-maintained exclusions, and the decisive code evidence
(F-1 store binding, F-3 no rebind plumbing) turns Option A's apparent "small middleware" advantage into a
large, cross-layer, unenforced change. The authority belongs where the cascade lives.

**Scoping question for you (§10):** wrap *all* cascade buses (complete, future-proof, enforced) vs only
the highest-risk subset first. Recommendation: **all** — begin-or-join is a no-op join when a tx is
present, so universal application is safe, removes the "which subset" judgment call, and the coverage
test only enforces a complete sweep.

---

## 8. Route / seam inventory (for the test surface and the sweep)

- **In scope (get the wrap):** all ~62–65 `.WithOutbox` domain cascade buses. Each reachable from its
  `<entity>api` write route with no caller tx (the gap) AND from self-tx handlers with a caller tx (the
  join path). Both exercised by the wrapper.
- **Already atomic (no change needed):** `allocation_results` (route-less, always under `allocate`/
  `reserve` self-tx); the synthesized paths `update_field`/`create_entity`/`transition_status` (worker
  handlers already `Begin*`+`WithTx`).
- **Must stay GREEN (the join branch must not regress them):** the 10 self-tx app handlers (§3 table) —
  especially the T3 phantom-rollback guard in `transferorderapi/atomicity_test.go`.

---

## 9. Test plan (TDD, RED-first, decisive — changed packages only, never `go test ./...`)

Reuse the existing rigs: `outbox/outbox_test.go` (tx + **poison** Emit forms; pool-fallback warn
assertion), `outbox/coverage_test.go` (`Registrations()`-derived every-cascade-bus-emits),
`actionhandlers/cascade_outbox_test.go` + `startCascadeRig` (human→ruleA→Emit→relay→ruleB), and
`transferorderapi/atomicity_test.go` (T3 phantom-rollback). Unique Temporal task queue per test; scoped
queries; `go build -C <worktree> ./...` before trusting gopls.

1. **★ Decisive RED-first — entity rolls back with a failed Emit on the *begin* path.**
   Construct a real cascade bus (a simple one, e.g. `currencybus`/`labelbus`) with a **poison `Writer`**
   (Emit returns an error / INSERT targets a missing table — reuse `outbox_test.go`'s poison form). Call
   `bus.Create(ctx, …)` on a **pool ctx (no caller tx)**.
   - **RED on master:** entity row committed (+1) AND error from Emit → lost cascade, split write.
   - **GREEN after Option B:** `bus.Create` self-began a tx; Emit failure rolls it back → **entity row
     absent (0)** AND error returned. This is the literal test bar from the brief.
2. **A→B→A cascade survives through the new atomic path.** Via `startCascadeRig`: a no-caller-tx simple
   write whose outbox row is now written in the bus's self-tx, drained by the relay, fires ruleA → … →
   and the A→B→A loop guard still **stops** after one hop. Proves the simple-write cascade is now durable
   AND the guard is unaffected.
3. **Join-not-nest regression (no double-tx, no phantom).** The existing T3
   (`transferorderapi/atomicity_test.go`) and `cascade_composition_test.go` (rollback → 0 rows AND no
   cascade) stay GREEN — the self-tx handler's bus calls **join** the handler tx (one tx, one commit),
   they don't nest. Add an assertion that a self-tx handler path commits exactly once (no inner commit
   ahead of the outer).
4. **Coverage enforcement.** Extend `coverage_test.go`: every cascade bus, written with no caller tx,
   **never hits the `emit.go:118` pool-fallback warn** (i.e. is wrapped). New cascade bus that forgets
   the wrap → RED.
5. **`BeginOrJoin` unit tests (`sqldb`):** present-tx → returns it, `joined=true`, commit is a no-op for
   the caller; absent-tx → begins, `joined=false`, ctx now carries the tx, commit/rollback owned.

---

## 10. Resolution (sign-off) — RESOLVED 2026-06-19

**Approved: Option B (begin-or-join in the bus), scope = ALL cascade buses.** Justification refined
in review: it is the older-Ardan `WithinTran` idiom (business-layer-owned tx) + a join branch, chosen
because the outbox row is the bus's own *invisible delivery infrastructure* (not a caller-orchestrated
business write), and because begin-or-join composes correctly with **every** tx owner — no caller (it
begins), an app self-tx (it joins), or a future newer-Ardan `BeginCommitRollback` middleware (it joins)
— with zero exclusion lists and no rework. Wrapper home = `business/sdk/outbox`; `sqldb` gets only the
`BeginOrJoin` primitive. **Prerequisite discovered during planning:** the `NewWithTx` silent
dropped-field class (`currencybus.NewWithTx` drops `del`; `delegate.Call` is not nil-safe) must be
hardened first via the copy-then-override idiom, or the rebind panics — shared by any rebind-based fix.
Implementation plan: **`FF2_IMPLEMENTATION_PLAN.md`** (7 tasks, TDD, RED-first decisive test).

Original open questions (now answered):
1. **Approve Option B** (begin-or-join primitive + `outbox.WriteAtomic` wrapper on the cascade buses)? — **YES.**
2. **Scope:** wrap **all** ~62–65 cascade buses now (recommended — complete, enforced, no judgment call),
   or stage to a high-risk subset first?
3. **Wrapper home:** `business/sdk/outbox` (recommended — it already holds the pool and is cascade-only)
   vs a new `business/sdk/busatomic`. Either keeps `sqldb` pure (it gets only `BeginOrJoin`).
4. **Primitive shape:** `BeginOrJoin(ctx, Beginner) (ctx, CommitRollbacker, joined, error)` as written,
   or a closure form `sqldb.InTx(ctx, bgn, func(ctx) error)`? (The generic bus wrapper needs the raw tx
   to call `NewWithTx`, so the explicit-return form composes more cleanly.)

**On approval:** implement TDD — test #1 RED on master first, then the primitive + wrapper, then sweep
the buses (parallelizable by package), GREEN the suite, `go build ./...` clean, run only changed
packages. Ship via the dual-remote flow (github PR → rebase-merge → `git fetch github && git push origin
github/master:master` → ff local master + `/merge-worktree`). **No push until you confirm.**

---

## 11. Known limitations / follow-ups (surfaced during implementation — NOT FF#2 regressions)

Both predate FF#2 or are out of its stated scope (single-bus "state + its own event" atomicity). FF#2
made each entity's write+cascade atomic; neither item below is introduced by it. Tracked here so they're
not rediscovered.

1. **`formbus.ImportForms` / `pageconfigbus.ImportPageConfigs` are non-atomic ACROSS entities.** These
   orchestration methods call the public `b.Create` (its own begin+commit tx) then loop
   `b.formFieldBus.Create` / sub-entity creates (each its own separate tx) — so a form commits, then its
   fields land in separate transactions; a mid-loop failure leaves the form orphaned. This is the same
   *multi-entity orchestration* class deliberately excluded from FF#2 (cf. `formdataapp` in F9), reachable
   only from their own CRUD route (no caller tx). Fix = an app-layer self-tx around the whole import (mirror
   the F9 self-tx handlers / FF#1 `formdataapp.UpsertFormData` `TxBind` pattern). Deferred.

2. **`warehousebus.Create` retry is begin-path-only.** Its unique-violation retry runs each attempt in its
   own `WriteAtomic` tx — correct today (no caller-tx call site; only `warehouseapp`, pool-bound). Under a
   future caller tx it would JOIN, and a unique violation would poison the caller's tx instead of rolling
   back. Flagged with an in-code comment at the retry loop; a future self-tx/workflow handler that creates a
   warehouse under its tx must not rely on the retry.
