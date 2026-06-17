# F9 — results, §8 coverage closure, findings & fast-follows

> Implements `F9_ATOMICITY_AND_TEST_GAPS.md`. Branch `feature/cascade-f2-atomicity` off master `5bb92ba2`.
> All work verified: `go build ./...` clean; changed-package suites green (`business/sdk/outbox`,
> `api/.../transferorderapi`, the cascade tests in `api/.../actionhandlers`). NEVER ran `go test ./...`.

## What shipped

**Part A — atomicity fix (the phantom-cascade-on-rollback bug).** Added `ctx = sqldb.WithTx(ctx, tx)`
immediately after `BeginTxx` in the 8 confirmed cascade-emitting self-tx handlers so the cascade
`outbox.Emit` rides the handler's transaction (commit/rollback together) instead of the base pool:
`transferorderapp.Execute`, `inventoryadjustmentapp.Approve`, `pickingapp.PickQuantity`/`ShortPick`,
`cyclecountsessionapp.complete`, `picktaskapp.complete`, `putawaytaskapp.complete`,
`pageactionapp.BatchCreate`, `inspectionapp.Fail`. (Entity writes already rode the tx via the storer;
only the Emit path changed. Sole other ctx-tx consumer, `mid.GetTran`, is request-lifecycle and
unaffected.)

**Part B — DESIGN §8 test gaps.** The three GAPs are now DIRECT, and two COMPOSED cases were upgraded:
- **T3 on-a-tx / phantom-rollback (P1)** — `transferorderapi/atomicity_test.go`. Proven RED before the
  fix (+3 phantom rows), GREEN after. This is also Part A's regression guard.
- **I2 tx-poison backstop (P2)** — `outbox/outbox_test.go`. Forced outbox INSERT failure on a shared tx
  → COMMIT downgraded to ROLLBACK → co-tx write gone.
- **I3 decisive read-your-writes (P3)** — `actionhandlers/cascade_ryw_test.go`. Synchronous
  `relay.ProcessBatch` around the commit boundary: pre-commit row invisible → 0 dispatched; post-commit
  → 1 dispatched reading the committed value. Replaces `TestCascade_M2_LiveCascade`'s timing-luck.
- **I1 (P4)** — `actionhandlers/cascade_composition_test.go`: rollback a real bus write → 0 rows AND no
  cascade, both legs in one test.
- **I6 (P4)** — same file: two updates to one entity → dispatched in seq order end-to-end.
- **#2 (P5)** — `outbox/outbox_test.go`: the loud pool-fallback warn is now asserted (fires on no-tx,
  not on tx).
- **#17 (P5)** — `outbox/coverage_test.go` re-driven off `workflowdomains.Registrations()` (per-package
  detection + registry-derived floor; kills the frozen `~64` and stops `excluded` masking a registered miss).

### §8 coverage matrix at close
- Unit: 17/17 DIRECT (was 15 DIRECT + 2 PARTIAL; #2 and #17 upgraded).
- Integration: I1/I3/I5/I7/I8/T1/T2/T3 DIRECT; I2 DIRECT; I6 DIRECT. I4 + path-breadth: see below.

## P4 items NOT built (evidence-based, not silent caps)

- **I4 (dedup re-feed → exactly one execution): NOT built; already covered + a finding.** Both halves are
  DIRECT today: `TestRelay_DrainsInSeqOrderThenDeletes` asserts `EventID == outbox row id`;
  `TestOnEntityEvent_DedupWorkflowID` asserts same-eventID → same `workflow-{ruleID}-{eventID}` id +
  `REJECT_DUPLICATE`. **Finding:** the spec's literal "exactly ONE `automation_executions` row on re-feed"
  does NOT hold — the execution RECORD is created at `temporal/trigger.go:234` *before* `ExecuteWorkflow`
  (:268), with a TODO at `:245`. So `REJECT_DUPLICATE` gives effectively-once **workflow execution**
  (Temporal), but a rare relay re-delivery would create a second StatusPending **record**. Dedup is
  run-level, not record-level — a separate, pre-existing concern, out of F9 scope (see fast-follow 3).
- **Live-relay path breadth (`create_purchase_order` Path B; `create_entity`/`transition_status` Path C):
  NOT built.** All three cascade PATHS already have a DIRECT live-relay test (A: `cascade_outbox` human
  trigger; B: `approve_transfer_order` bus emit; C: `update_field` synthesize via `cascade_m2`). The
  remaining per-action breadth exercises identical relay plumbing and is recorder-covered in
  `manifest_consistency_test.go` → COMPOSED. Rebuilding per action adds cost without new coverage of the
  cascade machinery.

## Fast-follows (separate changes, separately tested — NOT in F9)

1. **`formdataapp.UpsertFormData` non-atomic multi-entity submit.** Confirmed: it opens a tx + `defer
   Rollback`/`Commit` but creates ZERO tx-bound buses — all entity writes route through
   `api/.../build/all/formdata_registry.go` closures bound to POOL app instances, so the tx wraps
   nothing. This is a broader non-atomicity gap (a bad FK mid-submit leaves prior writes committed), NOT
   the phantom-on-rollback bug. Fix = thread tx-bound buses/apps through the registry (or make the
   registry exec layer honor ctx-tx). Must carry its own atomicity test.

2. **Path-A simple-write lost-cascade** (fail-SAFE; resolved DESIGN DECISION — do NOT wire tx middleware
   in F9). Close it in a scoped, separately-tested change. **M1/M2 entry-gate measurements (done):**
   - **M1 = "ever" (measured):** cascade-emitting bus methods ARE called under a caller-supplied tx — by
     the 8 F9 self-tx handlers and the 9 workflow action handlers, all via `bus.NewWithTx(tx)`. Per the
     spec's decision tree, this rules OUT an unconditional bus-local tx (it would nest: the inner commit
     makes the cascade durable before the outer write commits → a phantom-shaped split). The fix must be a
     **single begin-or-join authority** — `mid.BeginCommitRollback` scoped to the cascade-emitting routes
     (the mechanism `outbox/emit.go:60-63` already names), or a bus-layer begin-or-join primitive.
   - **M2 = no accessor today:** `sqldb` exposes `WithTx`/`GetTx`/`GetTxExecutor`/`WithCommitRollbacker`
     but no "join the ctx-tx if present, else open one" primitive. The begin-or-join fix would add one (or
     live in the scoped middleware).
   - Test bar: inject an Emit failure → prove the ENTITY write rolls back with it, + an A→B→A cascade
     through the new atomic path.

3. **Execution-record dedup (run-level vs record-level).** Per the I4 finding, `trigger.go:234`
   `CreateExecution` runs before `ExecuteWorkflow`; a deduped (REJECT_DUPLICATE) re-dispatch still writes
   a second StatusPending record (orphaned per the `:245` TODO). Low frequency (needs a relay re-delivery
   after a failed delete-on-publish), but worth closing: delete the orphaned row on `ExecuteWorkflow`
   error, or move record creation after a successful start / make it idempotent on the workflow id.

4. **`nontx_buses` (`workflowactions/inventory/createputawaytask.go:239`).** `putawaytaskbus.Create` on a
   tx-less ctx (pool fallback). Risk is a lost/orphaned event if the process dies mid-write (fail-safe),
   not phantom-on-rollback. Wrap the write+emit in a tx + `WithTx`. Deferred.
