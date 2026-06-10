# Cascade Delegate Events — Investigation & Design Reference

> **Status:** Investigation complete; design not started.
> **Created:** 2026-06-09
> **Branch:** `feature/cascade-delegate-events` (worktree `../ichor-cascade-delegate-events`, base `master` @ `727d46ce`)
> **Origin:** Spawned from PR #176 (`docs/update-field-no-cascade`), which documents the `update_field` case in isolation. This investigation found the problem is **systemic**, spanning ~8 handlers via two mechanisms, and that **PR #176's recommended workaround is itself broken.**
> **Goal:** Decide whether — and how — workflow actions that mutate the DB should be able to trigger *other* automation rules (cascade), and how to guard against runtime cascade loops.

This doc is the single cross-chat source of truth. Every claim below is either ✅ **verified by direct read** (file:line confirmed in this session) or 🔎 **traced by subagent** (file:line reported, high confidence, not personally re-read). Confidence is marked inline.

---

## 1. Problem statement

A workflow action can change data in the database, but the workflow engine only "sees" a change if that change emits a **delegate event**. Several action handlers write to the DB *without* emitting a delegate event, so any automation rule that should fire in response to that write **silently never fires.** The data is correct in Postgres; the *event* simply never exists.

Concretely, this chain does NOT work today:

```
Rule A:  on_create X       →  update_field sets X.status = 'on_hold'
Rule B:  on_update X, status changed_to 'on_hold'  →  create_alert
```

Rule A's write lands in the DB; Rule B never fires. The same change made via an API `PUT` (which goes through the business layer) fires Rule B immediately. The workflow editor will happily let you wire this dead chain — and for some actions even *advertises* the (dead) cascade point in its UI.

**We have decided a cascade IS wanted** (per maintainer, 2026-06-09). The hard part is the loop problem (§7–8).

---

## 2. How cascading works — the listening mechanism

Delegate calls are the **entire** "listening" surface. Nothing polls the DB.

```
[bus].Create/Update/Delete ──(after DB write succeeds)──> delegate.Call(ctx, Data)
                                                               │
                                          (DelegateHandler is a registered subscriber)
                                                               ▼
   business/sdk/workflow/temporal/delegatehandler.go  ──> WorkflowTrigger.OnEntityEvent(TriggerEvent)
                                                               ▼
   TriggerProcessor.ProcessEvent  ──> matched rules  ──> Temporal.ExecuteWorkflow  ──> Activities
```

Key files (✅ from `docs/arch/delegate.md`, `docs/arch/workflow-engine.md`, both authoritative):
- `business/sdk/delegate/delegate.go` — `Delegate.Call(ctx, Data)`; subscribers register at startup in `all.go`.
- `business/sdk/workflow/temporal/delegatehandler.go` — the workflow's delegate subscriber; converts `delegate.Data` → `workflow.TriggerEvent` via reflection.
- `business/sdk/workflow/temporal/trigger.go` — `WorkflowTrigger.OnEntityEvent(ctx, TriggerEvent)` entry point.
- `business/sdk/workflow/trigger.go` — `TriggerProcessor.ProcessEvent` matches `TriggerConditions` JSON against `TriggerEvent.FieldChanges`.

**The invariant:** every domain `[bus]` `Create/Update/Delete` calls `delegate.Call(...)` after its write (205 call sites / 65 files per `delegate.md`). An action that writes through such a bus cascades. An action that writes raw SQL, or through a bus method that omits `delegate.Call`, does not.

**Corollary for any fix:** "make action X cascade" ALWAYS reduces to "make X's write emit a `TriggerEvent` onto `OnEntityEvent`" — whether by routing through a delegate-firing bus, or by synthesizing the event directly.

---

## 3. Complete action classification (the map)

All 21 registered action types, traced to their actual DB write. (Handler registry: `docs/arch/workflow-engine.md` lines 208–236.)

> **Independently re-verified 2026-06-09** by a second round of 4 parallel opus agents, scopes split orthogonally and each told to *refute* rather than confirm (see §13). **Every verdict below was CONFIRMED; no classification was overturned.** The 🔎 entries now have a second independent read.

### 🔴 Category A — Writes the DB, does NOT cascade (the gap)

| Action | Pkg / file | Mechanism | Write evidence | Confidence |
|---|---|---|---|---|
| `update_field` | `data/updatefield.go` | **M1** raw SQL `UPDATE` | `:248` `NamedExecContextWithCount`; query `:222`. FK auto-create branch also raw (`:348` `NamedQueryStruct`) | ✅ verified |
| `create_entity` | `data/create.go` | **M1** raw SQL `INSERT` | `:119` `NamedExecContextWithCount`; query `:114-117` | 🔎 traced |
| `transition_status` | `data/transition.go` | **M1** raw SQL `UPDATE` | `:155` `NamedExecContextWithCount`; query `:149` | 🔎 traced |
| `log_audit_entry` | `data/audit.go` | **M1** raw SQL `INSERT` → `workflow.audit_log` | `:145` `NamedExecContextWithCount`; query `:128-130` | 🔎 traced |
| `create_alert` | `communication/alert.go` | **M2** bus has **no delegate field at all** | `:187` `alertBus.Create`; bus `alertbus.go:81-90` no delegate | ✅ verified |
| `seek_approval` | `approval/seek.go` | **M2** bus `Create` skips delegate (only `Resolve` fires) | `:140`/`:201` `approvalRequestBus.Create`; bus `approvalrequestbus.go:72` (no delegate) vs `:143` (Resolve fires) | ✅ verified |
| `allocate_inventory` | `inventory/allocate.go` | **M2 partial** — `allocation_results` write is dead | `:474` `CreateAllocationResult` → `workflowbus.go:1037-1055` no delegate | ✅ verified |
| `reserve_inventory` | `inventory/reserve_inventory.go` | **M2 partial** — same dead `allocation_results` hook | `:382` `CreateAllocationResult` → `workflowbus.go:1050` no delegate | 🔎 traced |

> **`allocate_inventory` / `reserve_inventory` are "partial":** their *primary* `inventory_items` mutation DOES cascade (it goes through `inventoryitembus.Update`, see Category B). Only their *secondary* `allocation_results` side-write is dead. This matters because **PR #176 recommends the `allocation_results` path as the supported cascade workaround — and that exact path is the dead one** (see §8).

### 🟢 Category B — Writes the DB and DOES cascade (correct)

| Action | Bus(es) used | delegate fires at | Confidence |
|---|---|---|---|
| `commit_allocation` | `inventoryitembus.Update` (`commit_allocation.go:165`) | `inventoryitembus.go:157` `ActionUpdatedData` | 🔎 traced |
| `release_reservation` | `inventoryitembus.Update` (`release_reservation.go:160`) | `inventoryitembus.go:157` | 🔎 traced |
| `receive_inventory` | `inventoryitembus.Update` (`receive.go:215`) + `inventorytransactionbus.Create` (`receive.go:239`) | `inventoryitembus.go:157` / `inventorytransaction.go:101` | 🔎 traced |
| `create_put_away_task` | `putawaytaskbus.Create` (`createputawaytask.go:239`) | `putawaytaskbus.go:95` `ActionCreatedData` | 🔎 traced |
| `create_purchase_order` | `purchaseorderbus.Create` (`createpo.go:404`) + `purchaseorderlineitembus.Create` (`:427`) | `purchaseorderbus.go:135` / `purchaseorderlineitembus.go:108` `ActionCreatedData` | 🔎 traced |

> Note: `allocate_inventory` + `reserve_inventory` also belong here for their `inventory_items` write. Their `NewWithTx` bus variants propagate the delegate field, so the transactional path still fires it (verified for PO: `purchaseorderbus.go:75` `del: b.del`).

### ⚪ Category C — No DB write; cascade is n/a (correctly silent)

| Action | Why no cascade |
|---|---|
| `lookup_entity` (`data/lookup.go`) | read-only `SELECT` (`:131`) |
| `check_inventory` (`inventory/check_inventory.go`) | read-only `Query` (`:154`) |
| `check_reorder_point` (`inventory/check_reorder_point.go`) | read-only `Query` (`:154`) |
| `evaluate_condition` (`control/condition.go`) | pure in-memory branch eval; handler has no `db` |
| `delay` (`control/delay.go`) | Temporal `workflow.Sleep`; `Execute` is a non-prod fallback |
| `send_email` (`communication/email.go`) | external SMTP via `emailClient.Send` (`:112`); holds a `db` handle but **never uses it** |
| `send_notification` (`communication/notification.go`) | RabbitMQ → WebSocket (`:118`); comment says "ephemeral — no DB persistence" |
| `call_webhook` (`integration/webhook.go`) | outbound HTTP `httpClient.Do` (`:161`); no DB handle |

**Tally:** 8 write-but-don't-cascade (A) · 5 write-and-cascade (B, + 2 partial from A) · 8 no-write (C) = 21. ✅

---

## 4. The two failure mechanisms

The gap has **two distinct root causes** with different fix shapes:

### Mechanism 1 — Raw SQL, no bus exists
`update_field`, `create_entity`, `transition_status`, `log_audit_entry`.

These are **generic** actions: they take an arbitrary `target_entity` / `target_field` and hand-build `INSERT`/`UPDATE` strings (`fmt.Sprintf` + `sqldb.NamedExecContextWithCount`). There is no typed bus to call because the action is polymorphic over every table — so it bypasses the bus **by design**, and the delegate goes with it.
- **Fix shape:** synthetic event emission — the handler must build a `TriggerEvent` after a successful write and feed `OnEntityEvent` directly. This is PR #176's "option 1." Architectural; needs the loop guard (§7–8).
- Each has a hardcoded **table-name allowlist** (`IsValidTableName`) precisely because no bus guards them — evidence the raw-SQL bypass was a known tradeoff.

### Mechanism 2 — Bus path that omits `delegate.Call`
`create_alert`, `seek_approval`, and the `allocation_results` side-write inside `allocate_inventory` / `reserve_inventory`.

These DO go through a bus — but that specific bus/method never wired up the delegate:
- `alertbus` (`business/domain/workflow/alertbus/alertbus.go`) — **no delegate field at all** in the struct; `Create` (`:81-90`) just calls `b.storer.Create`. ✅ verified.
- `approvalrequestbus` (`business/domain/workflow/approvalrequestbus/approvalrequestbus.go`) — has a delegate field and uses it in `Resolve` (`:143`), but `Create` (`:72`) skips it. ✅ verified.
- `workflow.Business.CreateAllocationResult` (`business/sdk/workflow/workflowbus.go:1037-1055`) — has a delegate field (used for rule-change events elsewhere) and a `// TODO: Implement delegate stuff here` at `:21`, but the method writes via `b.storer.CreateAllocationResult` and returns with **no** `delegate.Call`. ✅ verified.
- **Fix shape:** local — add `delegate.Call(ctx, ActionCreatedData(...))` to each bus method (and wire a delegate field into `alertbus`, which has none). Low-risk, testable, and it makes PR #176's documented workaround actually work. **No loop guard needed beyond what M1 needs**, because these go through the normal delegate path that the engine already handles. *(But see §7: even normal delegate-path writes can form cross-rule loops, so the loop policy must cover both mechanisms.)*

---

## 5. Second dimension: validation + integrity bypass

The bus is where `delegate.Call` lives **and** where business-layer validation, FK enforcement, and invariant checks live. So Mechanism 1 doesn't only skip events — it skips **all** of that.
- `create_entity` only checks a table-name allowlist and auto-generates the `id` via `uuid.New()` — no domain validation.
- `update_field` has its own hardcoded allowlist precisely because nothing else guards it.

So "no cascade" and "no validation" are the same raw-SQL root cause wearing two hats. Any decision to keep Mechanism 1 as raw SQL (e.g. only bolt on synthetic events) leaves the validation gap open. Worth an explicit decision, not a silent one.

---

## 6. The editor advertises dead cascades

`update_field`, `create_entity`, and `transition_status` each implement `GetEntityModifications` returning `on_create`/`on_update` — which the workflow **editor** reads to present them as cascade points in the UI. So the editor tells users they can chain off three actions whose runtime never fires the cascade. PR #176 noted this for `update_field`; it's true for three actions.
- `update_field`: `updatefield.go:369-374`
- `create_entity`, `transition_status`: `GetEntityModifications` present (🔎 traced)
- `allocate_inventory`: `allocate.go:690-705` advertises `allocation_results / on_create` — aspirational, currently dead.
- Related existing doc: `docs/workflow/cascade-visualization.md` (the visualizer that consumes this metadata — check whether it implies cascades that don't fire).

---

## 7. Current loop protection — and why it does NOT cover this

**What exists:** `app/domain/workflow/workflowsaveapp/graph.go` runs **Kahn's topological sort** at **save time** to reject cycles. Its checks (per `docs/arch/workflow-save.md:107-117`): exactly one start edge, no cycles, all actions reachable, etc.

**Why it's insufficient for cascades:** Kahn's operates on a **single rule's** `Actions` + `Edges` (intra-rule DAG). Cascade loops are **cross-rule** and **runtime**:

```
Rule A (on_update X) → writes Y
Rule B (on_update Y) → writes X
                       └──────────> re-fires Rule A → re-fires Rule B → ∞
```

There is **no structure anywhere that models rule→rule edges.** Those "edges" exist only implicitly at runtime, when a write's delegate event happens to match another rule's `TriggerConditions` (e.g. `changed_to`). At save time you cannot even enumerate the cross-rule graph — it depends on runtime field *values*, not static topology. So:
- Kahn's (save-time, intra-rule) **cannot** detect or prevent cascade loops.
- A cascade loop guard must be **runtime** and must travel **with the event**.

**Today this is moot only because cascades don't fire from these actions at all.** The moment we make them cascade (the goal), we open the door to infinite A→B→A loops. **This is the core design problem to solve before shipping any cascade.**

---

## 8. Design options

### The cascade mechanism (two sub-fixes, can ship independently)
1. **Mechanism 2 fix (cheap, high-value, do first):** add `delegate.Call` to `alertbus.Create`, `approvalrequestbus.Create`, and `workflowbus.CreateAllocationResult` (wiring a delegate field into `alertbus`). This un-breaks PR #176's *own* documented workaround (`allocation_results`) and makes `create_alert` / `seek_approval` cascade. Local, ~few lines each + bus constructor wiring + tests.
2. **Mechanism 1 fix (architectural):** synthetic `TriggerEvent` emission from the generic `data/` handlers after a successful raw-SQL write. Bigger; also the natural place to decide the validation question (§5).

### The loop guard (REQUIRED before any cascade ships — see §7)
Candidate approaches (to be debated in a future chat):
- **Hop counter / cascade depth in the event.** Carry a `cascade_depth` (and/or an originating-event id chain) on `TriggerEvent` → `WorkflowInput`. Increment per hop; refuse to dispatch past a max depth. Simple, bounded, no false negatives on legitimate deep chains beyond the cap. PR #176's suggested guard.
- **Visited-set / cycle path on the event.** Carry the set of `(ruleID, entityID)` already fired in this cascade; refuse to re-enter one. Detects true cycles precisely; larger payload; needs care across Temporal Continue-As-New.
- **Per-entity-instance rate/idempotency.** Suppress re-fires of the same rule on the same entity within a window. Coarse; risks dropping legitimate rapid updates.
- **Hybrid:** depth cap as a hard backstop + visited-set for precise cycle breaks + structured logging when a cascade is cut.
- **Origin tagging / suppression:** mark workflow-action-originated writes (a flag on the delegate `Data`, on `ctx`, or on the entity payload) so the trigger pipeline can suppress or depth-limit re-cascade for automated writes. ⚠ **Today NO such flag exists** — buses call `delegate.Call` unconditionally with no human-vs-automated distinction (verified §13.3). Adding one is itself a cross-cutting change across ~85 domains' delegate calls or a single chokepoint in `DelegateHandler`/`OnEntityEvent`.

Open design questions for the guard:
- Where is depth incremented and checked — `DelegateHandler`, `WorkflowTrigger.OnEntityEvent`, or `TriggerProcessor`?
- How does the counter survive Continue-As-New (`ContinuationState`/`MergedContext`)?
- What is the policy when the cap is hit: drop silently, drop + audit log, or surface an alert?
- Does the guard apply to ALL cascades (including the already-working Category B chains, which can ALSO loop), or only the newly-enabled ones? (Almost certainly all — Category B can loop today and is simply unguarded.)

---

## 9. Open questions / decision surface

1. **Per-handler: should it cascade?** Not all Category A members are bugs. `log_audit_entry` writes to a terminal `workflow.audit_log` sink — arguably *should* stay silent. `create_entity` / `transition_status` / `update_field` advertising cascades they don't deliver clearly is wrong. Decide per handler before coding.
2. **Loop policy** (§8) — the gating design decision.
3. **Validation gap** (§5) — if M1 stays raw SQL, do we accept unvalidated writes, or route generic actions through a generic validation layer?
4. **Sequencing** — ship M2 fix first (cheap, un-breaks docs) behind the loop guard, then M1? Or design the loop guard first since both need it?
5. **PR #176 disposition** — it documents a workaround that's broken. Options: (a) hold #176 until M2 fix makes the workaround real, then merge together; (b) merge #176's accurate parts now but correct the `allocation_results` recommendation; (c) supersede #176 with this effort. **Recommend (a) or (b) — do not merge the broken-workaround claim as-is.**
6. **Cascade reliability & ordering** (surfaced in §13.1–13.2) — even Category-B cascades are best-effort (delegate errors swallowed) and dispatched on an async goroutine off the originating transaction. Does the cascade design need a transactional-outbox / read-after-commit guarantee, or is best-effort acceptable? This is orthogonal to the loop guard but is a real correctness question the decision should not leave implicit.

---

## 10. File reference index

**Action handlers** — `business/sdk/workflow/workflowactions/`
- `data/updatefield.go` · `data/create.go` · `data/transition.go` · `data/audit.go` · `data/lookup.go`
- `inventory/allocate.go` · `reserve_inventory.go` · `commit_allocation.go` · `receive.go` · `release_reservation.go` · `createputawaytask.go` · `check_inventory.go` · `check_reorder_point.go`
- `communication/alert.go` · `email.go` · `notification.go` · `approval/seek.go` · `integration/webhook.go` · `procurement/createpo.go` · `control/condition.go` · `delay.go`

**Buses missing/partial delegate (Mechanism 2 fix targets)**
- `business/domain/workflow/alertbus/alertbus.go` — `Create:81` (no delegate field)
- `business/domain/workflow/approvalrequestbus/approvalrequestbus.go` — `Create:72` (skips), `Resolve:143` (fires)
- `business/sdk/workflow/workflowbus.go` — `CreateAllocationResult:1037` (skips; TODO at `:21`)

**Delegate-firing buses (reference for correct pattern)**
- `inventoryitembus` `:157` · `inventorytransaction` `:101` · `putawaytaskbus` `:95` · `purchaseorderbus` `:135` · `purchaseorderlineitembus` `:108`

**Pipeline & loop guard surfaces**
- `business/sdk/delegate/delegate.go` — `Call:48`
- `business/sdk/workflow/temporal/delegatehandler.go` — delegate subscriber
- `business/sdk/workflow/temporal/trigger.go` — `OnEntityEvent`
- `business/sdk/workflow/trigger.go` — `TriggerProcessor.ProcessEvent`
- `business/sdk/workflow/temporal/models.go` — `WorkflowInput`, `TriggerEvent` (where a depth field would live)
- `app/domain/workflow/workflowsaveapp/graph.go` — Kahn's (save-time; does NOT cover runtime cascade loops)

**Authoritative arch docs**
- `docs/arch/delegate.md` · `docs/arch/workflow-engine.md` · `docs/arch/workflow-save.md`

---

## 11. Related existing docs (review before designing)
- `docs/workflow/event-infrastructure.md` — likely the closest existing description of the delegate/event flow.
- `docs/workflow/cascade-visualization.md` — the editor cascade visualizer (consumes `GetEntityModifications`; may show dead cascades).
- `docs/workflow/actions/update-field.md` — PR #176's target.
- `docs/workflow/actions/{create-alert,seek-approval,allocate-inventory}.md` — per-action docs likely needing updates once fixed.
- `docs/workflow/branching.md`, `docs/workflow/architecture.md`.

---

## 12. Verification log

**Round 1 (initial trace, 2026-06-09):** 3 parallel agents by handler directory + 2 keystones personally re-read (`CreateAllocationResult`, `alertbus`, `approvalrequestbus`, `updatefield`).

**Round 2 (independent adversarial re-verification, 2026-06-09):** 4 parallel opus agents, scopes split *orthogonally* (bus-delegate internals / data+control+comms+approval handler write-sites / inventory+procurement+integration handler write-sites / pipeline+loop-guard), each instructed to **refute**. Each cascade verdict was recomputed by composing two independently-audited facts: (a) does bus-method M emit delegate? (b) what does handler H actually write? **Result: every §3–§7 verdict CONFIRMED; nothing overturned.** See §13 for confirmations and newly surfaced facts.

---

## 13. Independent verification round (2026-06-09) — confirmations + new design-critical facts

### Confirmed (second independent read)
- `alertbus`: **no delegate field at all** (package doesn't even import `business/sdk/delegate`); every method calls only `b.storer.*`. ✅✅
- `approvalrequestbus`: `Create:72` no delegate; `Resolve:143` fires `ActionUpdatedData`. Selective-by-method confirmed. ✅✅
- `workflowbus.CreateAllocationResult:1037-1055`: no delegate; stale `// TODO: Implement delegate stuff here` at `workflowbus.go:21` (delegate IS used for rule lifecycle at `:490/:543/:562/:581`, just not allocation results). ✅✅
- The 5 "correct" buses all fire delegate post-write: `inventoryitembus.Update:157`, `inventorytransaction.Create:101`, `putawaytaskbus.Create:95`, `purchaseorderbus.Create:135` (field named `del`), `purchaseorderlineitembus.Create:108`. ✅✅
- **`NewWithTx` propagation — suspected bug does NOT exist:** all tx constructors copy the delegate field (`inventoryitembus:64`, `purchaseorderbus:75`, `purchaseorderlineitembus:65`, also `inventorytransactionbus:64`, `putawaytaskbus:64`). The cascade gap is exactly the 8 handlers in §3-A, not a flaky tx path. ✅✅
- Pipeline wired end-to-end: `delegate.Call:48` → `DelegateHandler` (registered in `all.go`) → `OnEntityEvent` (`trigger.go:102`) → `ProcessEvent` (`trigger.go:122`). ✅✅
- **No runtime loop guard exists** (Agent confidence ~98%). Kahn's (`workflowsaveapp/graph.go`) is save-time + single-rule + cannot see cross-rule relationships. ✅✅

### New facts that affect the DESIGN (not the verdicts)
1. **Cascades are best-effort, even the working ones.** Every `delegate.Call` failure is only logged (`b.log.Error`), never returned — the bus method still returns success. A Category-B cascade can silently fail to fire if the delegate handler errors. No cascade is guaranteed today.
2. **Dispatch is async + off-transaction.** `DelegateHandler.handleEvent` fires `OnEntityEvent` in a `go func()` goroutine (`delegatehandler.go:86`); `delegate.Call` runs with the shared `ctx`, NOT the handler's `tx`. Handlers that commit AFTER the bus write (e.g. `allocate.go:483` commits after the Update fired delegate at `inventoryitembus.go:157`) can have a cascaded rule dispatched that queries the DB **before the originating tx commits** — a read-your-writes / ordering hazard for any cascade that reads the entity it was triggered by.
3. **No automated-origin flag on writes.** Buses call `delegate.Call` unconditionally; nothing distinguishes a human write from a workflow-action write (no `fromWorkflow`/`isAutomated`/`suppress` anywhere). → A loop guard cannot currently "skip re-cascade for workflow-origin writes" because that origin is not recorded. Key constraint on §8's origin-tagging option.
4. **Temporal gives zero incidental loop protection.** Workflow ID = `workflow-{ruleID}-{entityID}-{executionID}` with `executionID = uuid.New()` fresh per dispatch (`trigger.go:170-178`); every re-fire is unique, so same-ID dedup can never coincidentally break a loop.
5. **Whole pipeline is gated on `cfg.TemporalClient != nil`** (`all.go:561`; else-branch logs "temporal: disabled"). No client → entire delegate→trigger wiring absent → cascades don't fire at all. ~85 domains registered unconditionally once a client exists. (Note: guard variable is `TemporalClient`, derived from the host:port config — not `TemporalHostPort` directly.)
6. **Handler write-detail refinements (verdicts unchanged):**
   - `seek_approval`: up to **3** writes on the async path (`approval_requests` + `alerts` + `alert_recipients` loop, `seek.go:140/256/274`) and only **1** on the sync `Execute` fallback (`approval_requests` only, `:201`, no alert). Both the approval-request and alert writes are dead-end (M2).
   - `create_alert`: up to **3** DB ops — alert INSERT (`:187`), `alert_recipients` INSERT-loop (`:192`), and a conditional `ResolveRelatedAlerts` UPDATE (`:200`). The UPDATE also doesn't cascade (alertbus is delegate-free).
   - `create_put_away_task`: the only write-handler using the **non-transactional** bus (no `NewWithTx`) — still cascades.
   - `send_email` and `seek_approval` carry a vestigial, never-dereferenced `db *sqlx.DB` field (dead handle).

### Round 3 (design-driven verification, 2026-06-10)
Surfaced while designing the loop guard (see `DESIGN.md`):
7. **Trigger matcher IS value-aware** — `evaluateFieldCondition` (`trigger.go:335-377`) supports `equals/not_equals/changed_from/changed_to/greater_than/less_than/contains/in`; `changed_to` = `current==value && prev!=value`. So edges can in principle be matched on values, not just fields.
8. **Mutation manifest is value-BLIND** — `EntityModification` (`interfaces.go:152-163`) = `{EntityName, EventType, Fields[]}` only, no produced value. So static edges built from it today would be field-level → would false-positive on state machines. (Design fix: extend with produced value.)
9. **`GetEntityModifications` coverage is ~19 handlers, not ~4** — correcting §3/§6. Implemented across data/inventory/procurement/approval handlers.
10. **The action set has grown past the "21" in `docs/arch/workflow-engine.md`** — new types seen: `approve_po, reject_po, approve_adjustment, approve_transfer_order, reject_transfer_order, reject_adjustment, resolve`. §3's enumeration is stale and should be re-counted before any cascade analysis ships.
11. **The manifest is consumed by NOTHING in Go** — no matcher / cascade graph / cycle detector exists; the "cascade visualization API" the interface comment promises is not implemented in the backend (frontend/planned?). The static-analysis layer is greenfield, though the per-action substrate exists.

### Round 4 (scenario grounding, 2026-06-10)
Surfaced while building code-grounded loop scenarios (see `DESIGN.md` §4a/§4b):
12. **`changed_to V` is a fixed-point latch** — `evaluateFieldCondition` requires `current==V && prev!=V` (`trigger.go:346-349`); an idempotent re-write does NOT re-fire. `on_update`/empty conditions auto-match *every* event (`trigger.go:270-289`). ✅ personally read.
13. **Real loop-prone seeded rules exist** (🔎 Explore-traced): `Self Trigger Rule` (literal `on_update`+`update_field` self-loop) + chain rules in `api/.../workflow/ruleapi/cascade_seed_test.go`; `Allocation Success - Update Line Items` (writes `order_line_items` status from `allocation_results`) in `business/sdk/dbtest/seed_workflow.go`. The `on_update`+no-conditions auto-match shape is the most common seeded trigger and several **active** rules use it.
14. **Refines item 11:** a cascade-detection/visualization layer reportedly already excludes self-loops *for display* (exercised in `cascade_seed_test.go`). So item 11's "consumed by nothing" is specific to the `GetEntityModifications` path — some rule-matching/cascade logic may exist via another path. `[OPEN]` locate it before sizing the static-half build.
