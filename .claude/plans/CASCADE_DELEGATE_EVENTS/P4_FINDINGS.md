# P4 FINDINGS — code-grounded map for the cascade-enablement gate

> **Created 2026-06-12** from a 3-agent parallel exploration of the current code (branch `feature/cascade-delegate-events`). Purpose: let a **fresh session execute P4 without re-exploring**. Read this + `DESIGN.md` §5–§6/§10 + `PLAN.md` P4 — that's enough to build.
> **Status:** research done, decisions OPEN (see §E). Cascades still OFF. Piece 1 (guard) complete + committed (`decdcfd2`).
> **Every claim below is grounded in `file:line` as of this commit — re-verify a line before editing if the file has moved.**

P4 = the ONLY phase that turns cascades ON. Two mechanisms + sinks stay silent:
- **M1 synthesize:** `update_field` / `create_entity` / `transition_status` fire a manifest-driven delegate event after their raw-SQL write, threading the P1 visited-set lineage.
- **M2 allocation_results:** `workflowbus.CreateAllocationResult` fires a `delegate.Call`; reconcile the `allocation_results` entity name.
- **Sinks stay silent:** `log_audit_entry`, `create_alert`, `seek_approval`.

---

## ⚠ TWO CORRECTIONS TO THE PLAN-OF-RECORD (read first)

1. **`allocation_results` must stay BARE/UNQUALIFIED for the runtime delegate event.** PLAN.md:84 says "fix the unqualified `allocation_results` entity name" implying schema-qualification. The code says the opposite: the trigger matches `rule.EntityName == event.EntityName` on the **bare** `workflow.entities.name` column (`trigger.go:209`; the name column is `c.relname` = bare table name, `seed.sql:653-717`, schema is a *separate* column). `inventoryitembus.EntityName = "inventory_items"` (unqualified) with an explicit comment that entities are stored table-name-only (`inventoryitembus/event.go:16`). The seeded consumer rule "Allocation Success - Update Line Items" looks up its entity via `QueryEntityByName(ctx, "allocation_results")` (`seed_workflow.go:48`). **→ the delegate event's `EntityName` must be bare `"allocation_results"`.** A qualified `"workflow.allocation_results"` would never match. The *manifest* (`GetEntityModifications`) is a separate consumer and may want a different string — see §C.

2. **The Temporal worker has NO delegate subscriber.** `workflow-worker/main.go:162` does `del := delegate.New(log)` but registers **no** `DelegateHandler` / `RegisterDomain` / `WorkflowTrigger` on it (the worker uses `RegisterAll`, `main.go:189`). Activities (where handler `Execute` runs during a cascade) run on the worker. So a synthesized `delegate.Call` there reaches **zero subscribers and silently no-ops** — the cascade dies at hop 1. **P4 must wire a `WorkflowTrigger` + `DelegateHandler` + `RegisterDomain(...)` into the worker**, mirroring `all.go:583-664` / `apitest/workflow.go:93-107`. This is a structural prerequisite PLAN.md does not surface. (The in-process `all.go` binary DOES wire the subscriber; the standalone worker does not — `workflow-engine.md:19-21` documents only the in-process pipeline, a doc gap.)

---

## A. M1 — the three generic handlers (current state + recommended mechanism)

All three live in `business/sdk/workflow/workflowactions/data/`, share an identical dep shape, and **none imports `sdk/delegate` or holds a bus today.**

| Handler | File | Struct deps | Manual-exec? | Manifest (`GetEntityModifications`) |
|---|---|---|---|---|
| `update_field` | `updatefield.go` | log, db, templateProc, protected | **false** (automation-only, `:63-67`) | `EntityName=cfg.TargetEntity` (schema-qualified), `on_update`, `Fields=[TargetField]`, `Changes=[ProducedChange]` (`:374-408`) |
| `create_entity` | `create.go` | log, db, templateProc, protected | true (`:50-52`) | `EntityName=cfg.TargetEntity`, `on_create`, no Fields/Changes (`:153-163`) |
| `transition_status` | `transition.go` | log, db, templateProc, protected | true (`:50-52`) | `EntityName=cfg.TargetEntity`, `on_update`, `Fields=[StatusField]`, `Changes` (`:190-210`) |

Execute ends (where synthesis hooks in, AFTER a successful write):
- `update_field`: returns `result` with `status:"success"`, `records_affected` (can be **0** on no match) (`:200-229`).
- `create_entity`: returns `created_id` (= `processedFields["id"]`, auto-gen at `:112-114`), `status:"success"` (`:133-149`).
- `transition_status`: **invalid-transition early-exit at `:150-157` does NO write** (must NOT fire); success branch writes + returns `transitioned:true` (`:181-186`). It reads `currentStatus` first (`:139`) so it has BOTH old+new for `on_update`.

### The normal delegate→cascade channel a synthesized event must enter
`bus.Update/Create` → `delegate.Call(ctx, ActionXxxData(...))` → `DelegateHandler.handleEvent` → `WorkflowTrigger.OnEntityEvent` → Temporal.
- `delegate.Call(ctx, data)` invokes subscribers **synchronously, passing `ctx` straight through** (`delegate/delegate.go:52`). **This ctx pass-through is the load-bearing fact for lineage.**
- `delegate.Data = {Domain, Action, RawParams}`; `RawParams` → `workflow.DelegateEventParams{EntityID, UserID, Entity, BeforeEntity}` (`workflow/event.go:34-39`).
- `RegisterDomain(del, domainName, entityName)` registers 3 funcs `(domain,"created"/"updated"/"deleted")` → `handleEvent("on_create"/...)` which sets `TriggerEvent.EntityName = entityName` (`delegatehandler.go:34-47,58-64`).
- Canonical working emit to copy: `inventoryitembus.go:102-104` `delegate.Call(ctx, ActionCreatedData(item))`; constructor `inventoryitembus/event.go:43-61`; consts `DomainName="inventoryitem"`, `EntityName="inventory_items"`, `ActionCreated="created"` (`event.go:11,16,20-22`).

### Lineage threads FOR FREE if synthesis routes through `delegate.Call(ctx)`
- P1 stamps lineage on the Go ctx in the activity **before** `handler.Execute`: `ctx = contextWithLineage(ctx, lineageFromContextMap(input.Context))` (`temporal/activities.go:64` sync / `:143` async; Execute at `:66`).
- `DelegateHandler` reads it back via `lineageFromContext(ctx)` before the dispatch goroutine and re-stamps onto `context.Background()` (`delegatehandler.go:91,95`).
- So when a generic handler calls `h.delegate.Call(ctx, ...)` with its **received ctx**, the lineage propagates with **zero extra code**. (If synthesis bypassed `delegate.Call` and hit `OnEntityEvent` directly, lineage would need manual threading AND the consistency-test spy — which hooks `delegate.Call` — wouldn't see it.)
- Lineage struct: `WorkflowLineage{Visited []string "ruleID:entityID", OriginatingExecutionID}` (`temporal/lineage.go:49-65`); payload key `CascadeLineageKey="__cascade_lineage"` (`:41`); rides `WorkflowInput.TriggerData` (`trigger.go:215`).

### RECOMMENDED mechanism — Option A (inject `*delegate.Delegate`, synthesize via `delegate.Call`)
Aligns with every DECIDED note (single channel for M1+M2, lineage-via-ctx, manifest-consistency CI spy). Mirror the protected-registry injection template exactly:
- Add `WithDelegate(del *delegate.Delegate) Option` to `data/protected.go` (or new `data/options.go`); store `delegate` on each handler struct.
- In each `Execute`, after a confirmed write, build `delegate.Data` from the manifest + call `h.delegate.Call(ctx, data)`.
- Thread `del` through **both** registration entry points: worker `RegisterAll` (`register.go:84-87`, called `main.go:189`) AND `all.go` upgrade block (`all.go:565-567`, where the 3 handlers already get `WithProtectedRegistry`; `delegate` is in scope at `all.go:381/397`). `RegisterCoreActions` stays delegate-free by design (`register.go:188-193`).

Rejected: **Option B** (inject DelegateHandler, call OnEntityEvent directly) — import-cycle risk (`data`→`temporal`), breaks single-channel symmetry, invisible to the consistency spy. **Option C** (bus-routing) — explicitly a FOLLOW-UP (F1), out of scope.

---

## B. M2 — allocation_results (smaller than it looks; bus already has the delegate field)

- **`CreateAllocationResult` fires nothing today** (`workflowbus.go:1036-1055`) — writes one row, no `delegate.Call`.
- **No struct/constructor churn needed:** `Business` ALREADY holds `delegate *delegate.Delegate` (`workflowbus.go:128-132`), constructed in `NewBusiness` (`:135-141`), propagated through `NewWithTx` (`:145-158`). The bus already fires rule-CRUD delegates correctly: `if b.delegate != nil { b.delegate.Call(ctx, ActionRuleChangedData(...)) }` (`:489-493` etc.). **M2 mirrors this exact nil-guarded, error-log-not-fail pattern in `CreateAllocationResult`.**
- **NEW delegate namespace required:** workflowbus already owns `DomainName="workflow_rule"` for rule-CRUD (`workflow/event.go:12`) — allocation_results needs a **separate** new `DomainName` (e.g. `"allocation_result"`) + an `ActionAllocationResultCreatedData(...)` constructor with `Action="created"` (none exists today — grep `AllocationResultCreated` = empty). The `Action` MUST be the const `"created"` because `RegisterDomain` registers under `workflow.ActionCreated` (`delegatehandler.go:35`). Only `on_create` is meaningful (insert-only table; rule triggers on_create).
- **Register the domain in all.go:** add `delegateHandler.RegisterDomain(delegate, <AllocationResultDomainName>, "allocation_results")` inside the `if cfg.TemporalHostPort != ""` block alongside `all.go:594-681`. (Currently absent — that's why even a wired `delegate.Call` wouldn't reach anything.)
- **Both allocate + reserve covered by one bus fix:** `allocate.go:474-477` and `reserve_inventory.go:382-385` both call `txWorkflowBus.CreateAllocationResult` (the `NewWithTx` variant carries the same delegate). Note: fires **pre-commit** (inside the handler tx; commit at `allocate.go:483` / `reserve:390`). Downstream rule reads `order_line_items` (different table), so the read-your-writes hazard is mild — flag for PL ordering test.
- **Reserve manifest gap:** `allocate.go:690-706` declares both `inventory.inventory_items` AND `allocation_results`; `reserve_inventory.go:408-416` declares ONLY `inventory_items` — add the `allocation_results` modification to reserve's manifest for declared==fired consistency.

---

## C. THE LOAD-BEARING DECISION — schema-qualified entity name → delegate domain mapping

Shared by M1 and M2. **There is no mechanical transform** between the three name forms:
- Manifest declares **schema-qualified** table: `"inventory.inventory_items"`, `"sales.orders"` (e.g. `allocate.go:696`, generic handlers' `cfg.TargetEntity`).
- Delegate fires under the bus's short **DomainName**: `"inventoryitem"`, `"order"` (NOT derivable by stripping the schema — `inventoryitem` ≠ `inventory_items`).
- Trigger matches on the bare **EntityName**: `"inventory_items"`, `"orders"`.

The only existing bridge is hard-coded: `all.go:594-681` (~60–90 `RegisterDomain(domainName, entityName)` lines) and the test's 7-entry `entityDomain` map (`actionhandlers/manifest_consistency_test.go:121-129`). The generic handlers accept an **arbitrary** `target_entity` from the whitelist (`data/tables.go`), so synthesis needs to resolve `schema.table` → `(domainName, bare entityName)` for the **full writable-table set**, and **unmapped targets must degrade safely** (no panic, no false cascade — just no event, logged).

**Options to decide (§E.2):** (a) a central registry/reverse-lookup keyed by qualified name (single source, reusable by the test); (b) emit only `entity_name`+`event_type`+`field_changes` and route by bare entity name, sidestepping DomainName; (c) per-table mapping table maintained alongside the whitelist. Option (a) or (b) preferred — (b) is attractive because the trigger only needs the bare EntityName, but `delegate.Call` is keyed by `(Domain, Action)`, so a synthesized event still needs *a* domain to register/emit under. A dedicated synthetic domain (e.g. `"generic_write"`) that carries the bare entity name in the payload + a `RegisterDomain` per writable entity is a possible clean shape — design in the fresh session.

---

## D. Tests — exactly ONE file flips by design

- **Structural test** `workflowactions/manifest_consistency_test.go` — asserts DECLARED side only, no delegate spy, no knownSilent. **Stays green at P4.**
- **DB-backed "declared==fired"** `api/cmd/services/ichor/tests/workflow/actionhandlers/manifest_consistency_test.go` — THE trip-wire file:
  - `knownSilentEntities = {"allocation_results": true}` (`:131-134`) → **remove at M2**, add declared==fired for `allocation_results`/`on_create`.
  - `update_field_known_silent` subtest (`:545-570`) asserts NO `inventoryitem.updated` fires; RED-by-design error string at `:568`. → **delete subtest at M1**, add positive declared==fired. Mapping `"inventory.inventory_items"→inventoryitembus.DomainName` already present (`:127`) — no map extension for this case.
  - `create_entity`/`transition_status` are **not executed here today** → add NEW positive subtests; ensure their chosen `target_entity` keys exist in `entityDomain` (`sales.orders`, `inventory.inventory_transactions` are NOT in the map — pick mapped targets or extend).
- **Everything else stays green** (verified orthogonal): `cascade_lineage_test.go` (one-hop A→B via already-firing Category-B handler — the PL bridge, not a casualty), `inventory_test.go`, `comms_test.go` (sinks), `ruleapi/cascade*_test.go` (static cascade-MAP), `actionapi/execute_test.go` (manual-execute asserts 400/403 before any write — but **glance** during impl: a successful manual `create_entity`/`transition_status` will begin firing once synthesis is wired), loopdetect/cascade_detect/lineage/protected tests. No cascade-off execution-count assertions found.

### PL (live verification, the phase AFTER P4) — reusable harness
- `apitest/workflow.go InitWorkflowInfra(t,db)` — shared Temporal container, worker on `"test-workflow-"+t.Name()`, auto-cleanup (`testing.md:109-129`). Caveat: must `Delegate.Register(...)` explicitly.
- **The real PL template = the manual-worker rig** in `actionhandlers/cascade_lineage_test.go:63-105` (worker + custom ActionRegistry + `RegisterActivity` + `NewDelegateHandler` + `RegisterDomain` + `seedRule` + poll `automation_executions.trigger_data` for `CascadeLineageKey`). Extend rule B to write back into A's entity (via a synthesized M1 handler) → assert the 3rd hop is visited-set-blocked = **the decisive A→B→A loop-stopped test**.
- Shared seed `seedConsistencyBase` (`manifest_consistency_test.go:196-330`); delegate spy `delegateRecorder` (`:73-108`) for fan-out/idempotency counts.

---

## E. OPEN DECISIONS (resolve at the top of the fresh session)

1. **Synthesis channel** — `delegate.Call(ctx)` (Option A, recommended) vs direct `OnEntityEvent`. *Recommend A.* Everything else hangs on this.
2. **schema.table → (domain, bare entity) mapping** (§C) — central registry vs synthetic domain vs per-table map; MUST degrade safely for unmapped/whitelisted-but-no-bus targets. *Load-bearing; design in fresh session.*
3. **Worker delegate-subscriber wiring** (⚠#2) — confirm we add `WorkflowTrigger+DelegateHandler+RegisterDomain` to `workflow-worker/main.go`. *Structural prereq; not optional.*
4. **`allocation_results` name** — delegate event = bare `"allocation_results"` (⚠#1). Decide separately whether the *manifest* line (`allocate.go:701`) gets qualified for cascade-map/static-edge consistency, and whether the consistency test compares manifest-name vs event-name (reconcile if so).
5. **Fire-only-on-real-write guards** — `transition_status` invalid-transition (no write) must NOT fire (`transition.go:150`); `update_field` `records_affected==0` — suppress event? *Recommend yes (avoid phantom cascades).*
6. **`FieldChange.OldValue` contract per handler** — `update_field` does a blind UPDATE (no prior value); `transition_status` has `currentStatus`; `create_entity` is on_create (no diff). Decide what `FieldChanges`/`BeforeEntity` the synthesized event carries (downstream field-condition rules depend on it).
7. **Manual-execution cascade semantics** — `create_entity`/`transition_status` can run via `POST .../execute` (no activity → empty lineage → fresh chain). Should a manual run cascade at all, or only automation-triggered? *Recommend: cascade with a fresh chain (correct per `lineage.go:104-110`); confirm.*
8. **`RegisterCoreActions` stays delegate-free** — confirm synthesis dep is wired only in `RegisterAll`/`all.go` (like the protected registry), so cascade-visualization core-path tests are unaffected.

---

## F. Proposed build order (fresh session)
1. Resolve §E.1–E.2 (channel + mapping) — they shape all the code.
2. Wire the worker delegate subscriber (§E.3) — prove a real bus write on the worker path cascades (de-risk the structural gap first).
3. M2 (smaller, isolated, the bus already has the delegate field) — `CreateAllocationResult` `delegate.Call` + new event constructor + `RegisterDomain` + reserve manifest + flip the `allocation_results` knownSilent. Verify with the DB consistency test.
4. M1 (the bigger one) — inject delegate into the 3 handlers, synthesize per the §E mapping, fire-only-on-write guards, lineage-via-ctx. Flip the `update_field` trip-wire; add create/transition positive subtests.
5. Inline smoke (PLAN P4 "Done when"): 4 advertised handlers cascade, sinks silent, both consistency tests green.
6. Hand off to **PL** (separate phase) for the decisive live A→B→A-stopped + fan-out + ordering + idempotency tests.
</content>
</invoke>
