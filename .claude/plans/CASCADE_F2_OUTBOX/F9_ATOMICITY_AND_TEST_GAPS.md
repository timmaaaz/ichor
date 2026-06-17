# F9 — Atomicity hardening + DESIGN §8 test gaps  (CASCADE_F2_OUTBOX follow-up)

> Authored 2026-06-17 as a fresh-window handoff. Self-contained: execute WITHOUT
> re-deriving — every claim below is verified against committed code on master `d3cff6f5`.
> Source of truth: `DESIGN.md` §4 (atomicity) + §8 (testing); `PROGRESS.yaml` follow_up
> (`apptx_withtx`, `nontx_buses`); the merged F2 code on master.

## Status / context
F2 (transactional outbox for cascade delegate events) shipped to master `d3cff6f5` (PR #185).
Cascades flow: bus write → `b.outbox.Emit(ctx, data)` INSERT `workflow.cascade_outbox` (same tx) →
`temporal/relay.go` polls → `WorkflowTrigger.OnEntityEvent`. A post-ship audit found a latent
atomicity bug AND the DESIGN §8 test that was supposed to catch it was never written. **These are
one root cause** — fix the bug and write its guard together (TDD: guard RED first).

---

## THE ROOT CAUSE (one sentence)
The DESIGN §8 "on-a-tx trip-wire" (T3) was never written, so the app-layer handlers that begin their
own tx but never enroll it on the ctx shipped with a silent **phantom-cascade-on-rollback** bug — no
test goes red when a cascade row commits on the pool instead of the entity's transaction.

---

## PART A — Atomicity fix: app handlers must enroll their tx on ctx

### Mechanism (verified)
`bus.NewWithTx(tx)` swaps only the storer; it does **NOT** call `sqldb.WithTx(ctx, tx)`. So when that
bus cascades, `outbox.Emit` reads `sqldb.GetTxExecutor(ctx)` (`emit.go:116`), finds no tx, and falls
back to the **base pool** with a loud warn (`emit.go:118`) — the row commits **immediately, independent
of the handler's tx**. If the handler then rolls back on a business error, the entity write vanishes but
the outbox row survives → the relay fires a **phantom cascade** for a rolled-back operation.

Verified: `git grep 'sqldb\.WithTx' -- '*.go' ':!*_test.go'` → the ONLY 9 setters are the workflow
action handlers (`workflowactions/data/{create,transition,updatefield}.go`,
`workflowactions/inventory/{allocate,commit_allocation,receive,release_reservation,reserve_inventory}.go`,
`workflowactions/procurement/createpo.go`). **Zero** under `app/domain/`.

### The fix (per handler)
Immediately after `db.BeginTxx(...)` and BEFORE the `NewWithTx` bus calls, add:
```go
ctx = sqldb.WithTx(ctx, tx)   // enroll the tx so cascade Emits ride it (atomic with the entity write)
```
Then verify the handler passes that SAME `ctx` into every bus call. Result: the bus Emits land on the
handler's tx → commit/rollback together → no phantom on rollback. (Import `business/sdk/sqldb` if absent.)

### Handlers to fix — verify each anchor on the live tree first (from master `d3cff6f5`)
Priority = reachability of a rollback AFTER the first cascade emit during normal use.

| Priority | Handler (method) | File anchor | Cascade buses written | Rollback trigger |
|---|---|---|---|---|
| HIGH | `transferorderapp.Execute` | `app/domain/inventory/transferorderapp/transferorderapp.go:245` (BeginTxx `:259`) | transferorderbus Update + inventorytransactionbus ×2 | insufficient source stock |
| HIGH | `inventoryadjustmentapp.Approve` | `app/domain/inventory/inventoryadjustmentapp/...:170` | inventoryadjustmentbus Approve/Update + inventorytransactionbus | zero `QuantityChange` (user-supplied) |
| HIGH | `pickingapp.PickQuantity` / `ShortPick` | `app/domain/sales/pickingapp/pickingapp.go:124` / `:300` | inventoryitembus, inventorytransactionbus, orderlineitemsbus, ordersbus | insufficient / substitute stock |
| HIGH | `formdataapp.UpsertFormData` | `app/domain/.../formdataapp/...:179` | generic registry → orders/orderlineitems/products/suppliers… | multi-entity submit, one bad FK ⚠ ALSO: confirm whether `NewWithTx` is even applied to its entity writes — agent flagged a broader pre-existing atomicity gap here; investigate before fixing |
| MED | `cyclecountsessionapp.complete` | `:119` | cyclecountsessionbus + inventoryadjustmentbus (per item) | per-item adjustment error |
| MED | `picktaskapp.complete` | `:146` | picktaskbus + inventorytransactionbus | DecrementQuantity error |
| MED | `putawaytaskapp.complete` | `:134` | putawaytaskbus + inventorytransactionbus | Upsert/FK error |
| MED | `pageactionapp.BatchCreate` | `:287` | pageactionbus (CreateButton/Dropdown/Separator) | per-action unique/FK error |
| LOW | `inspectionapp.Fail` | `:158` | inspectionbus + lottrackingsbus | quarantine error |

Also re-run the discovery to catch any sibling: `git grep -l 'BeginTxx\|NewWithTx' -- 'app/domain/**/*.go'`.

### SEPARATE (the `nontx_buses` follow-up, lower priority)
`business/sdk/workflow/workflowactions/inventory/createputawaytask.go:239` calls `putawaytaskbus.Create`
on a **tx-less** ctx — pool fallback, but there's NO surrounding tx / no rollback step, so its risk is
non-atomicity (a lost/orphaned event if the process dies mid-write), NOT phantom-on-rollback. Fix =
wrap the write+emit in a tx + `WithTx`. Defer unless it matters.

### ✅ DESIGN DECISION — RESOLVED 2026-06-17 via /debate (do NOT re-open)
`mid.BeginCommitRollback` is wired to **zero routes** (verified: only its own def +
`app/sdk/mid/transaction.go`). So Path-A human HTTP writes carry no tx → single-bus writes Emit on the
pool. Exposure = a non-atomic dual write: if Emit fails AFTER the entity commits, the cascade is
**silently and permanently LOST** (a missing allocation / status-transition / inventory side-effect —
fail-SAFE, not a phantom, but real and unrecoverable: there is nothing left to reconcile against).
The "wire it on all ~152 write routes" option is **REJECTED** (over-broad behavior change; and it would
SPLIT entity-write/emit across two txns in the self-tx handlers — making the dangerous bug worse).

RESOLUTION:
- **In F9 (this pass): ship ONLY the per-handler `sqldb.WithTx(ctx, tx)` bind-line above.** That kills
  100% of the fail-DANGEROUS phantom cascades — F9's actual mandate. The Path-A simple-write lost-cascade
  is fail-SAFE and is **explicitly OUT of F9 scope**. Do NOT wire any tx middleware in F9. Record the
  fast-follow below.
- **Fast-follow (separate change, scoped, separately tested):** close the simple-write lost-cascade. It
  MUST carry its own cascade-atomicity test (inject an Emit failure → prove the ENTITY write rolls back
  with it, + an A→B→A cascade through the new atomic path). Never a free rider on a green build.
- **Mechanism for the fast-follow — decide by MEASUREMENT, not principle. Run both first:**
  - **M1:** for each cascade-emitting bus method, can it EVER be called under a caller-supplied tx (from a
    self-tx / orchestration handler via `NewWithTx`)?  `git grep` the call sites.
  - **M2:** does the F3.3-unified tx key already expose a clean "join the ctx-tx if present, else open one"
    accessor (a ~3-line read of the same key `emit.go` already reads)?
  - **M1 = "never" for all emitting buses → bus-local tx**: wrap `storer.Create` + `Emit` in a tx inside
    the bus method (atomic at the source, NO ambient-tx contract, touches only cascade buses). Strongest if M2 = yes.
  - **M1 = "ever" for any → that subset needs a SINGLE begin-or-join authority**: `mid.BeginCommitRollback`
    scoped to the cascade-emitting routes (the mechanism `emit.go:60-63` already names), OR a bus-layer
    begin-or-join primitive. An UNCONDITIONAL bus-local tx there silently NESTS (the inner commit makes the
    cascade durable BEFORE the outer write commits) → re-creates a phantom-shaped split. Avoid.
  - (Debate record: the global flip is dead; lost-cascade is real-but-fail-safe; middleware and the
    per-handler WithTx bind-line are complementary halves of one guarantee, not rivals.)

### Pre-F2 framing (do NOT call this an F2 regression)
Pre-F2 the delegate dispatched cascades via a detached `go func()` on `context.Background()`, so a
phantom ALSO fired on rollback — best-effort, often dropped. F2 didn't create the phantom; it made it
(1) loudly logged and (2) one-line fixable. The nuance that makes it worth fixing NOW: F2's reliable
delivery means a phantom that used to be silently dropped will now **reliably execute a workflow**.

---

## PART B — Test gaps (DESIGN §8). Prioritized.

### P1 — on-a-tx trip-wire  (GAP; this is also PART A's regression guard) ★ do first, TDD
DESIGN §8 T3: "each of the 3 tx paths lands the row on a tx, not the pool fallback … catches a path
that forgets to populate it." It does not exist. Write it so it goes **RED against current app handlers
(pool fallback) BEFORE the PART A fix, GREEN after**. Pick the decisive form:
  - **Phantom-prevention (best):** run a real self-tx handler (e.g. `transferorderapp.Execute` forced
    into its insufficient-stock rollback, or wrap a handler call in a tx the test rolls back) and assert
    **zero `workflow.cascade_outbox` rows** afterward. RED today (pool-committed row survives), GREEN
    after WithTx.
  - **No-pool-fallback (complements):** capture the logger and assert the `"no transaction on context —
    emitting on base pool"` warn (`emit.go:118`) does NOT fire for the covered paths. (This also closes
    unit partial #2 — the warn is currently never asserted.)

### P2 — tx-poison backstop  (GAP)
DESIGN §8 explicitly required "the poison backstop is also asserted." Nothing does (grep `poison` in
tests = 0 hits). Force an outbox INSERT failure on a shared tx, attempt `COMMIT`, assert Postgres
downgraded it to ROLLBACK → BOTH the entity row and the outbox row are absent.

### P3 — read-your-writes, decisive  (GAP)
The only RYW evidence (`cascade_m2_test.go` `TestCascade_M2_LiveCascade`) passes by timing luck — its
own comments (`:181-183`,`:208-209`) concede the M2 fire is pre-commit and works only because worker lag
> commit. Write the deterministic Path-B test §8 asked for ("RED on master, GREEN after"): the cascaded
rule must read the COMMITTED entity. The relay only dispatches committed rows, so the outbox closes the
race — make a test that would be RED under a pre-commit dispatch and GREEN via the relay.

### P4 — composition → decisive upgrades (lower)
Currently proven only by combining smaller tests; add one decisive test each:
  - **I1**: roll back a real BUS write → assert 0 outbox rows AND no cascade (both legs, one test).
  - **I4**: re-feed the SAME outbox row id through the relay → assert exactly ONE
    `workflow.automation_executions` row (dedup decisive; today only implied by EventID=row.id +
    REJECT_DUPLICATE).
  - **I6**: two updates to ONE entity → assert cascades fire in `seq` order end-to-end (today only
    relay-unit with a fake dispatcher + 3 distinct domains).
  - **Path breadth**: live relay E2E for Path-B `create_purchase_order` and Path-C `create_entity` /
    `transition_status` (today only `allocate` + `update_field` get a live relay cascade; the rest lean
    on `manifest_consistency_test.go`, which asserts against a delegate RECORDER and never touches the relay).

### P5 — unit partials
  - **#2** loud-warn assertion — folded into P1's log-capture form.
  - **#17** `TestCoverage_EveryCascadeBusEmits` is a heuristic source scan of `business/domain` with a
    hand-maintained `excluded` set. Re-drive it off `workflowdomains.Registrations()` so it cannot miss a
    cascade bus registered outside `business/domain` or pass vacuously when `excluded` drifts.

---

## Sequencing (TDD)
1. Write the P1 trip-wire / phantom-rollback test → **prove RED** (current pool fallback). Commit the RED proof note.
2. Apply PART A `WithTx` fixes → P1 **GREEN**. (Bug + guard land together.)
3. Add P2 (poison), P3 (RYW), then P4 decisive tests.
4. Refactor the coverage trip-wire (P5 #17); fold in P5 #2.
5. Record the Path-A simple-write lost-cascade FAST-FOLLOW (decision already made — see the resolved
   DESIGN DECISION in Part A). Do NOT wire any tx middleware in F9; do NOT re-open the debate. Capture the
   M1/M2 measurements as the fast-follow's entry gate.

## Verification
- `go build ./...` clean (run via `go build -C <worktree>` if gopls squiggles).
- Run ONLY changed packages — **NEVER `go test ./...`**: each fixed `app/domain/...` handler pkg, `business/sdk/outbox/...`, and `api/cmd/services/ichor/tests/workflow/actionhandlers/...` (the cascade suite + the new trip-wire). Unique Temporal task queue per test; scoped queries.

## Branch / ship (dual-remote: github=source, Bitbucket=mirror)
- Create a fresh worktree off master `d3cff6f5` (e.g. `feature/cascade-f2-atomicity`) via the `/worktree` skill. First action in the worktree: copy/commit THIS spec into the branch (it's untracked in the main repo).
- Ship: PR on github → rebase-merge → `git fetch github && git push origin github/master:master` → ff the main repo + clean up the worktree (`/merge-worktree`).

---

## Evidence appendix (so you don't re-derive)

**Verified anchors (master `d3cff6f5`):**
- `business/sdk/outbox/emit.go:116-121` — `GetTxExecutor(ctx)` → pool fallback + warn when no tx.
- `sqldb.WithTx` setters: ONLY the 9 workflow action handlers listed above; ZERO in `app/domain/`.
- `mid.BeginCommitRollback`: defined (`api/sdk/http/mid/transaction.go:14`, `app/sdk/mid/transaction.go:14`), wired to no route.
- `transferorderapp.Execute`: func `:245`, `BeginTxx :259`, `NewWithTx ×3 :266/:280/:316`.
- Relay = sole dispatcher (`relay.go`, started `all.go:633` inside `if cfg.TemporalClient != nil`).

**Existing tests to extend (don't start from scratch):**
- `business/sdk/outbox/outbox_test.go` — `TestEmit` already has the on-tx / rollback→0-rows / pool-fallback / error-propagation unit patterns; the P1/P2 forms build directly on these.
- `api/cmd/services/ichor/tests/workflow/actionhandlers/cascade_outbox_test.go` + `cascade_rig_test.go` — the live integration rig (worker + WorkflowTrigger + relay, NO DelegateHandler). P1 phantom-prevention + P3 RYW + P4 build on this.
- `business/sdk/workflow/temporal/relay_test.go` — fake-dispatcher relay tests; I4/I6 decisive forms extend these.
- `business/sdk/outbox/coverage_test.go` — the P5 #17 trip-wire to refactor.

**§8 coverage matrix at handoff (DIRECT = decisive single test / COMPOSED = implied by combination / GAP = none):**
- Unit: 17/17 covered (15 DIRECT, 2 PARTIAL: #2 warn-not-asserted, #17 heuristic).
- Integration: I5 (A→B→A stop) DIRECT; I7 (E2E Path A) DIRECT; I8/T1/T2 DIRECT (unit/by-construction).
  I1/I4/I6 COMPOSED. **I2 (poison) GAP. T3 (on-a-tx) GAP. I3 (RYW decisive) GAP.**
  Path A DIRECT; Path B DIRECT(allocate)/COMPOSED(create_po + 4 others); Path C DIRECT(update_field)/COMPOSED(create_entity, transition_status).
