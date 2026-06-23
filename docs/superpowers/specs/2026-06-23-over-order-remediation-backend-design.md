# Over-Order Remediation — Backend Design (Phase 1)

- **Date:** 2026-06-23
- **Branch:** `feature/over-order-remediation` (based on `633d8cd1`)
- **Status:** Approved design — ready for implementation plan
- **Scope:** Backend / workflow engine only. Frontend (`../vue/ichor`) is a separate repo and a later session; see "Frontend follow-up".

---

## 1. Problem

An "over-order" is an order line whose requested quantity exceeds available inventory. The workflow engine should make every over-order **operator-actionable end to end**: detected, surfaced, held for a human decision, and recoverable once stock is fixed.

The enabling primitive already exists. After commit `633d8cd1`:

- `check_inventory` is **quantity-aware**: it gates on `available >= max(threshold, requestedQty)` (only on the `source_from_line_item` path; `check_inventory.go:137-142,178`). Ports: `sufficient` (default), `insufficient`.
- `reserve_inventory` declares ports `{success, partial, insufficient_stock, failure}` (`reserve_inventory.go:151-158`) and **soft-fails** on a true shortfall: it returns a result map with `output="insufficient_stock"` and a `nil` error *before* `tx.Commit`, so the deferred rollback undoes any greedy partial reserves (`reserve_inventory.go:369-383`). Infra failures still hard-error (Temporal retries). A shortfall is therefore a **routable decision point**, not a crash.

Two gaps remain:

1. **The default graph doesn't route the shortfall.** In the only active inventory graph (seed Rule 5 "Granular Inventory Pipeline"), `reserve_inventory`'s sole outgoing edge is *unconditional* (`SourceOutput=nil`) to a success-alert. Its `insufficient_stock` / `partial` / `failure` ports have **no edge** — a race-condition shortfall dead-ends silently. (`check_inventory`'s `insufficient` port *is* wired to an alert; reserve's is not.)
2. **No way to re-run a failed/insufficient execution.** `executionapi` is read-only (two GET routes). Once an over-order routes to an alert/approval, there is no backend path to re-attempt the reservation after the operator fixes stock.

Both `seed_workflow.go` and the execution surface are **production** code: `seed_workflow.go` ships the default graphs to real customers via `InsertPlatformConfig` / `make seed-platform` (it is not test-only despite living under `dbtest`).

---

## 2. Goals / Non-Goals

### Goals
- Wire a sensible **default over-order remediation graph** into the production seed.
- Add a **re-run-execution** endpoint — the one real backend gap — that genuinely re-attempts (not replays a cached failure).
- Give over-orders a first-class **`over_order` alert type** that surfaces in the existing Exception Inbox, deep-linkable to its execution.
- Fix one **verified correctness bug** the work depends on (approval heartbeat timeout).

### Non-Goals (deferred — evidence shows they're already expressible or out of scope)
- First-class `backorder` action — a customer domain-modeling choice (new status/entity), not an engine primitive.
- First-class `partial_fulfillment` action — `reserve_inventory` already has a `partial` port + `allow_partial` config.
- In-line `wait_for_event` — wait-for-restock already works as a *second* rule triggered on `receive_inventory`'s `inventory_items` on_update (quantity) cascade.
- Generic `retry` action — the re-run endpoint *is* the operator-driven retry; Temporal already auto-retries infra errors 3×.
- `seek_approval` `timed_out` durable timer — see §8 (documented fast-follow; the loop closes without it).

---

## 3. Decisions

| # | Decision | Rationale |
|---|----------|-----------|
| Q1 | Scope = config + default graph + re-run + `over_order` alert. Defer absent first-class actions. | The configurable architecture already exists; the deferred actions are already expressible (see Non-Goals). |
| Q2 | Default wiring = over-order **alert + approval hold** (both). | An over-order is a business exception worth a human decision; the alert guarantees visibility even if the approval is never acted on. |
| Q2b | Approval fix = remove the 1h heartbeat timeout. Defer the `timed_out` durable timer. | Heartbeat is a verified correctness bug and load-bearing for the default (see §6.D). The timer is a larger, riskier engine change the default doesn't need. |
| Q3 | Re-run keyed on **`execution_id`**, single-rule replay, fresh execution-id. | Unambiguous "re-run this run." The id is available from both the executions page and the `approval_requests` row. Avoids the surprising "re-evaluate every matching rule" scope of the entity+rule path. |
| Q4 | `over_order` alert type (no migration), `high`/`critical` severity, `execution_id` in `context`. | `alert_type` and `context` are free-form (no migration). `high`/`critical` already differentiate; a new severity value would need an `ALTER TYPE` migration for no real gain. |
| ★ | Implement "both" **sequentially** (alert → approval), not as a parallel fan-out. | A parallel fan-out has no convergence point, so the executor runs the branches as **fire-and-forget child workflows with `PARENT_CLOSE_POLICY_ABANDON`** (`workflow.go:286-337`): the parent completes immediately (status `completed`) while the approval lives in a detached child — bad for tracking and re-run. Sequential keeps it one clean run that holds at the approval (status `running`), which re-run-by-execution-id maps onto directly. The alert still fires first, preserving the "immediate visibility" intent. |

---

## 4. Default Over-Order Graph (Deliverable A)

Extend the active seed Rule 5 "Granular Inventory Pipeline" (`business/sdk/dbtest/seed_workflow.go`). Sequential chain:

```
reserve_inventory
  --[insufficient_stock]--> over_order_alert   (create_alert: type=over_order, severity=high)
        --[success]-------> approval_hold       (seek_approval: approvers, approval_type=any, timeout_hours)
              --[approved]--> approved_alert     (create_alert: "approved by {{resolved_by}}")
              --[rejected]--> rejected_alert     (create_alert: "rejected — hold/cancel line")
              --[timed_out]--> (UNWIRED — item-2 follow-up; see §8)
  --[failure]-------------> failure_alert        (create_alert: severity=critical)
```

New nodes added to Rule 5: `over_order_alert`, `approval_hold`, `approved_alert`, `rejected_alert`, `failure_alert`. New edges (all `EdgeType="sequence"`):

- `reserve_inventory --[insufficient_stock]--> over_order_alert`
- `over_order_alert --[success]--> approval_hold`
- `approval_hold --[approved]--> approved_alert`
- `approval_hold --[rejected]--> rejected_alert`
- `reserve_inventory --[failure]--> failure_alert`

Edges are added with the existing bus API — `busDomain.Workflow.CreateActionEdge(workflow.NewActionEdge{...})` — taking the address of a local port-string variable for `SourceOutput` (Go can't address a literal). The bus layer does not validate `SourceOutput` against declared ports (validation lives on the save/API path), so the seed accepts these edges; all the port strings used here are real declared ports.

**Config notes:**
- `seek_approval` config: `approvers` (a seeded role/admin user UUID — reuse the platform `admin_gopher` or a supervisor role already in the seed), `approval_type: "any"` (the only implemented mode; `all`/`majority` silently degrade to `any`), `timeout_hours` (set, but inert until the item-2 follow-up — documented).
- `over_order_alert` recipients: a seeded supervisor role (≥1 recipient is required by `AlertConfig.Validate`).
- `partial` port is intentionally left unwired in the default — it cannot fire while `allow_partial:false` (the default reserve config). A customer enabling `allow_partial` would wire it in their own graph (configurable principle).

---

## 5. Re-Run Execution Endpoint (Deliverable B — the main gap)

### 5.1 The three dedup walls (why a fresh execution-id is mandatory)
Re-running with the *original* execution-id is a no-op against three independent guards; **minting a fresh `uuid.New()` execution-id clears all three**:

1. **Action idempotency** — `reserve_inventory`/`allocate` key on `idempotencyKey = "{executionID}_{ruleID}_{actionType}"` against `workflow.allocation_results` (`reserve_inventory.go:203`). Same id → cached result returned, no re-attempt.
2. **Temporal run dedup** — workflow id is `workflow-{ruleID}-{dedupKey}` with `REJECT_DUPLICATE` + `WorkflowExecutionErrorWhenAlreadyStarted=true` (`trigger.go:218-222,271-281`). `dedupKey = event.EventID` or, when `EventID==uuid.Nil`, the fresh per-dispatch executionID.
3. **Execution-record** — `CreateExecution` insert + `recordExecution` upsert keyed on `id`. Reusing the id collides/overwrites history.

### 5.2 Mechanism
A re-run reconstructs the original event and dispatches **only the one rule** that produced the execution, with a fresh execution-id and `EventID=uuid.Nil`. This reuses the existing `startWorkflowForRule`, which already mints `executionID := uuid.New()` (`trigger.go:201`) and, with `EventID==uuid.Nil`, derives `dedupKey` from that fresh id — clearing walls 1 and 2; the fresh id also clears wall 3.

`trigger_data` (persisted on the execution row at dispatch, `trigger.go:242-251`) is sufficient to rebuild the event: `buildTriggerData` (`trigger.go:344-371`) copies the **entire** `event.RawData` plus `entity_id`, `entity_name`, `event_type`, `user_id`, `field_changes`. Reconstruction reverses this. Crucially, `reserve_inventory` reads **live** inventory at execution time — `trigger_data` only says *what* to reserve — so a re-run after a restock genuinely succeeds rather than replaying the stale shortfall.

### 5.3 Layers
- **HTTP** — `api/domain/http/workflow/executionapi/route.go`: add `POST /v1/workflow/executions/{id}/rerun` (route table is GET-only today). `mid.Authenticate` + `mid.Authorize` (mutating write on `workflow.automation_executions`; exact rule confirmed against `docs/arch/auth.md` during impl). Handler in `executionapi.go` parses `{id}`, calls the app layer, returns the new execution id.
- **App** — new thin `app/domain/workflow/executionapp/` (mirrors `actionapp`): `Rerun(ctx, executionID, userID) (newID, error)` — permission/authorization mapping + orchestration; `errs.*` mapping (e.g. `ErrNotFound` → 404, "no rule" → 400/422).
- **Dispatch** — new exported method on `WorkflowTrigger` (`business/sdk/workflow/temporal/trigger.go`): `RerunExecution(ctx, executionID) (uuid.UUID, error)`:
  1. Load the original execution by id (add `QueryExecutionByID` to the trigger's execution-store interface if absent).
  2. If `AutomationRuleID == nil` → error (a manual/ruleless execution has no rule to re-fire).
  3. Load the rule (name/id) to build the `RuleMatchResult`.
  4. `reconstructTriggerEvent(execution)` — reverse `buildTriggerData`: parse `entity_id`/`user_id` UUIDs, pull `entity_name`/`event_type`, rebuild `FieldChanges` from `field_changes`, and treat the remaining keys (minus the metadata + cascade-lineage keys) as `RawData`. Set `EventID = uuid.Nil`, fresh lineage (a re-run starts a new cascade chain).
  5. Call `startWorkflowForRule` → fresh execution-id, dispatch, return the new id.
- **Wiring** — `api/cmd/services/ichor/build/all/all.go`: thread the already-built `WorkflowTrigger` (constructed inside `if cfg.TemporalClient != nil`) into `executionapi.Config`. When Temporal is disabled (trigger nil), the rerun route returns a clear "unavailable" error; the GET routes are unaffected.
- **Test infra** — `api/sdk/http/apitest/workflow.go`: expose the trigger so integration tests can exercise rerun.

### 5.4 Out of scope for v1 (deferred)
- A re-run does **not** resolve/cancel the original pending approval (stale-approval-on-rerun) — see §8.
- No `parent_execution_id` provenance column — add only if lineage tracking is later required.

---

## 6. Supporting changes

### Deliverable C — `over_order` alert + `execution_id` enrichment
- New `alert_type` value `over_order` — free `VARCHAR(100)`, **no migration**, **no new severity**. Surfaces in the existing Exception Inbox (`/workflow/alerts/mine`, filterable `alertType=over_order`).
- Small `create_alert` handler change (`communication/alert.go`): expose `execCtx.ExecutionID` (and `RuleID`) so it is available in the alert's `context` JSON and/or as a template var for `action_url` (e.g. `/workflow/executions/{{execution_id}}`). Today only `Title`/`Message`/`ActionURL` are templated, from `execCtx.RawData` (`alert.go:131-144`), and `Context` is stored verbatim (`alert.go:114-118,133`); `execution_id` is not in `RawData`, so the handler must inject it. This benefits **every** alert (each becomes deep-linkable to its execution), not just over-orders — keep the change general.
- `context` payload for the over-order alert: `execution_id`, `product_id`, `requested_qty`, `available_qty`, `order_ref`.

### Deliverable D — Approval heartbeat fix (verified bug)
Delete `ao.HeartbeatTimeout = time.Hour` for human actions (`workflow.go:657`). The async-completion path returns `activity.ErrResultPending` immediately with **no heartbeating goroutine** (`activities.go:210`), so a 1h heartbeat timeout (enforced server-side from activity start) kills any hold longer than ~1h; with `MaxAttempts=1` the activity then fails and the held approval is orphaned. Async-completion activities should not set a heartbeat timeout — the 7-day `StartToCloseTimeout` is the real bound. This is load-bearing: the default graph holds at `seek_approval`, so without the fix the shipped default breaks after an hour. It is also a general correctness fix for every approval.

---

## 7. Testing (integration-primary; run only changed packages — never `go test ./...`)

- **Re-run (core):** seed an over-order execution → `POST .../rerun` → assert (a) a **new, distinct** execution-id is returned, (b) a fresh Temporal run starts (no `REJECT_DUPLICATE` collision), (c) `reserve_inventory` actually re-attempts (not the cached `allocation_results` result), (d) after a simulated restock the re-run **succeeds**. Negative: re-run of a ruleless execution → error; unknown id → 404.
- **Heartbeat fix:** an approval left pending past the old 1h window still resolves and resumes the run (locks the fix; guards regression).
- **Default seed graph:** load Rule 5 after seeding → assert the new nodes/edges exist with the correct `source_output` strings; the graph validates on the save path.
- **Approval happy-path:** drive `seek_approval` → resolve `approved` then (separately) `rejected` → assert routing to `approved_alert` / `rejected_alert` through the sequential chain.
- **Alert enrichment:** assert the `over_order` alert carries `execution_id` (+ product/qty/order) in `context` and is returned by `/workflow/alerts/mine?alertType=over_order`.

Packages touched (tests scoped to these): `business/sdk/dbtest`, `business/sdk/workflow/temporal`, `business/sdk/workflow/workflowactions/communication`, `app/domain/workflow/executionapp`, `api/domain/http/workflow/executionapi`, and the relevant `api/cmd/services/ichor/tests/workflow/...` integration harness.

---

## 8. Deferred / Follow-ups (engine)

- **`timed_out` durable timer (item 2):** make the `timed_out` port routable via a `workflow.Selector` racing the approval activity against a `workflow.NewTimer(timeout_hours)`, plus deterministic config parse, DB marking of the expired approval, and cancellation of the orphaned activity. Larger, shared-executor-surface change with real regression risk; the default is safe without it because the `over_order` alert fires first (sequentially, before the hold), so the exception is surfaced regardless of the approval's fate. Documented as a fast-follow.
- **Stale-approval-on-rerun cleanup:** optionally have re-run resolve/cancel the prior pending approval for the same entity. Independent of the timer.
- **`parent_execution_id` provenance:** only if re-run lineage/chain tracking is later needed.

---

## 9. Frontend follow-up (seeds the later `../vue/ichor` prompt)

The backend contracts above are shaped to fit this work; none of it is in this PR.

- **Re-run button** on `ExecutionDetailPanel.vue` → `POST /v1/workflow/executions/{id}/rerun`; on success, navigate to the returned new execution id. Show for failed/insufficient/awaiting-approval executions.
- **WS approval push:** push a typed `approval_request` (created) event over the WebSocket to close the created-vs-resolved asymmetry — today only `approval_resolved` is typed; new approvals arrive only as a generic `alert`. Handle it in `composables/useAlertWebSocket.ts` and the approvals Pinia store. (Backend follow-up to emit the typed message is noted; the FE work consumes it.)
- **Exception Inbox surfacing:** filter `alertType=over_order` to give over-orders their own lane; deep-link an alert to its execution via `context.execution_id` (→ the Re-run button above).
- **Editor PropertyPanel:** expose `check_inventory.threshold` / `reserve_inventory` config (e.g. `allow_partial`, `reservation_duration_hours`) so authors can tune the over-order thresholds without editing JSON.

---

## 10. Key files

- `business/sdk/dbtest/seed_workflow.go` — default graph (Deliverable A); ships via `make seed-platform`.
- `business/sdk/workflow/temporal/trigger.go` — `startWorkflowForRule`, `buildTriggerData`, new `RerunExecution` + `reconstructTriggerEvent` (Deliverable B).
- `api/domain/http/workflow/executionapi/{route.go,executionapi.go,model.go}` — rerun route + handler (B).
- `app/domain/workflow/executionapp/` — new app layer for rerun (B).
- `business/sdk/workflow/workflowactions/communication/alert.go` — `execution_id` enrichment (Deliverable C).
- `business/sdk/workflow/temporal/workflow.go:657` — delete human-action heartbeat timeout (Deliverable D).
- `api/cmd/services/ichor/build/all/all.go` — thread `WorkflowTrigger` into `executionapi.Config`.
- `api/sdk/http/apitest/workflow.go` — test infra for rerun.

Reference (read before implementing): `docs/arch/workflow-engine.md`, `docs/arch/workflow-alerts.md`, `docs/arch/domain-template.md`, `docs/arch/auth.md`.
