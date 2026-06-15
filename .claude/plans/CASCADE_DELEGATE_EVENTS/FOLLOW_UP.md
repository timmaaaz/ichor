# FOLLOW_UP — Deferred / downstream tracks (cascade-delegate-events)

> Created 2026-06-10. Work intentionally deferred OUT of the **core plan** (loop guard + M2 `delegate.Call` + M1-synthesize + protected-list). Each item: what it is · why deferred · trigger to pick it up · key files.
> Core plan: [`DESIGN.md`](./DESIGN.md) (loop guard) + [`WRITE_PATH.md`](./WRITE_PATH.md) (write-path / protected-list).

---

## 1. Option B — bus-routing consolidation (the big one)
**What:** route generic `update_field`/`create_entity` writes through the domain bus via the FormData dispatcher, so they get validation + cascade natively and there is one write path per domain. Full shape in WRITE_PATH.md §9.
**Why deferred:** ~58-entity audit-and-migrate behind a one-time `EntityDispatcher` bridge (business→app dependency inversion). Correctness-positive for *plain* fields, but NOT a prereq (the core ships complete without it) and does NOT dissolve the protected-list — domain trace **Verdict a**: `bus.Update` bypasses the `Approve`/`Reject` guard just like raw SQL.
**Gating risk:** silent column-name↔JSON-tag drop — needs a per-column parity audit per entity (a non-matching/unexposed column drops silently on `json.Unmarshal`, worse than raw SQL).
**Trigger to pick up:** after the core ships + is stable; migrate entity-by-entity, deleting the raw+synthesize path per entity as it flips; once all migrated, delete the raw path entirely.
**Key files:** WRITE_PATH §5/§9; `app/sdk/formdataregistry/registry.go`; `formdataapp.go:436-488`; `api/cmd/services/ichor/build/all/formdata_registry.go`.

## 2. FormData latent tx bug
**What:** `UpsertFormData` opens a `tx` (`formdataapp.go:179`) but the registered closures write through base-pool app instances, not `tx`; each write auto-commits and each `delegate.Call` fires *before* the outer commit → no atomic multi-write.
**Why deferred:** only load-bearing if Option B is adopted (single-entity workflow writes barely care).
**Trigger:** Option B Phase 0, or sooner if FormData's own multi-entity ops need atomicity.

## 3. Reliability / ordering hardening (transactional outbox)
**What:** cascades are async / off-transaction / best-effort today — a cascaded rule can read the entity before the originating tx commits (read-your-writes hazard), and `delegate.Call` errors are swallowed (INVESTIGATION §13.1–13.2). v1 accepts this (DESIGN §10).
**Why deferred:** matches current behavior; the loop guard doesn't change it.
**Trigger:** a real cascade ordering/correctness bug, or a deliberate move to a transactional-outbox / read-after-commit model.

## 4. PR #176 doc salvage
**What:** #176's accurate `update_field` no-cascade limitation prose.
**Action:** when writing user-facing workflow docs for the shipped fix, fold in #176's accurate text; close #176 as **superseded**; never merge its broken `allocation_results` workaround.

## 5. Frontend Path B — real field-picker UX
**What:** convert `update_field`'s free-text `target_entity`/`target_field` into entity/field pickers that grey out protected fields with "requires the *Approve PO* action" messaging.
**Why deferred:** the core protected-list is backend-authoritative (rejection surfaces via the existing error toast = Small FE cost). The picker is a UX upgrade — Medium, because `DynamicForm` has no entity/field-picker type today (the pattern lives only in the condition path).
**Trigger:** after the backend protected-list ships, if authors want inline guidance.
**Key files:** `useActionConfigForms.ts:240-273`, `PropertyPanel.vue`, `SmartFieldPicker.vue`, introspection `Column` struct (`introspectionapp/model.go:52-62`).

## 6. Stale arch-doc cleanup
- `docs/arch/workflow-engine.md`: action set is **24**, not 21 (add the approve/reject/resolve handlers).
- `docs/workflow/cascade-visualization.md`: **19** `EntityModifier` implementors, not 1.
- Manifest drift fixes (the Item-5 consistency test will catch these once built): `resolve_approval_request` under-declares (`status` only; misses `resolved_by`/`resolution_reason`/`resolved_date`); approve/reject omit `updated_*`; `allocate` declares unqualified `allocation_results`.

## 7. Intentional-loop support (when a real case appears)
**What:** exit-clause authoring / bounded-loop semantics — the `maxReEntries` escape hatch on the visited-set.
**Why deferred:** decided (DESIGN §8) — forbid-by-default now; no concrete intentional-loop use case yet.
**Trigger:** a real time-delayed / externally-bounded loop requirement (e.g. "re-check stock hourly until available").

## 8. Cascade off alert / approval-request *creation*
**What:** let rules trigger on `create_alert` / `seek_approval` writes (add `delegate.Call` to `alertbus`/`approvalrequestbus.Create`).
**Why deferred:** decided silent for v1 (DESIGN §10 cascade-scope) — alerts are notification-terminal, and the meaningful approval event already cascades via `Resolve`. Cascading them adds event volume + loop surface (`create_alert` is very common) for hypothetical demand.
**Trigger:** a concrete "react to a critical alert being created" / "react to an approval being requested" workflow.

## 9. Protected-field registry + missing typed-action build list
> Derived from Probe 1/2 (2026-06-10) — **verify against code in build-phase 1** before wiring the block-list. Item-2 decision = block-now, build-on-demand (DESIGN §10). The registry itself is **core-plan** (the protected-list); the **NEEDS NEW ACTION** rows are the follow-up build list.

The ~30 guarded `(entity,field)` pairs, each blocked from the generic handlers (`update_field`/`create_entity`/`transition_status`); disposition = how a workflow legitimately writes it:

| Entity | Field(s) | Disposition |
|---|---|---|
| `purchase_orders` | status, approved_by/date/reason, rejected_by/date/reason | Existing action → `approve_/reject_purchase_order` |
| `purchase_orders` | priority | Validated → bus-routing (§1) covers it; enum, low-risk meanwhile |
| `inventory_adjustments` | approval_status, approved_by, approval_reason, rejected_by, rejection_reason | Existing action → `approve_/reject_inventory_adjustment` |
| `inventory_adjustments` | reason_code, quantity_change | Validated → bus-routing |
| `transfer_orders` | status (approved/rejected), approved_by_id, rejected_by_id, reasons | Existing action → `approve_/reject_transfer_order` |
| `transfer_orders` | **status (in_transit, completed)**, quantity, claimed_by, completed_by | **NEEDS NEW ACTION** → wrap bus `Claim`/`Execute` |
| `inventory_items` | quantity, reserved_quantity, allocated_quantity | Existing action → allocate/reserve/commit/release/receive |
| `inventory_transactions` | (whole table) | Protect-only — append-only ledger, never workflow-set |
| `orders` | priority | Validated → bus-routing |
| `orders` | fulfillment_status_id | Protect-only — recomputed by picking; direct set desyncs |
| `order_line_items` | picked_quantity, backordered_quantity, line_item_fulfillment_statuses_id, quantity | Protect-only — written transactionally by the picking ledger |
| `purchase_order_line_items` | quantity_received, quantity_cancelled | **NEEDS NEW ACTION** → wrap bus `ReceiveQuantity` (accumulate) |
| `users` | user_approval_status, approved_by, date_approved | **NEEDS NEW ACTION** → wrap userbus `Approve`/`Deny` |

**Follow-up build list (the NEEDS-NEW-ACTION rows — build each only when a real workflow needs it):**
1. `claim_transfer_order` / `execute_transfer_order` — wrap `transferorderbus.Claim`/`Execute` (in_transit→completed state machine + Execute's atomic stock move).
2. `receive_po_line_item` — wrap `purchaseorderlineitembus.ReceiveQuantity` (accumulate-not-replace).
3. `approve_user` / `deny_user` — wrap `userbus.Approve`/`Deny`.

**Being built now — `action-buttons-phase2` branch (2026-06-15):** build-list item 1 (`claim_transfer_order`/`execute_transfer_order`) **plus a NEW `release_to_picking` action** that emerged from the configurable-buttons Phase-2 work. `release_to_picking` flips `orders.fulfillment_status_id` PENDING/PROCESSING→PICKING **and** fans line items into `inventory.pick_tasks` (directed-work model) — making it the **sanctioned writer of `orders.fulfillment_status_id`**, so that row's disposition above changes from *Protect-only* → *routed to `release_to_picking` (entry transition only; pickingapp still owns the later PICKING→PACKING→SHIP recomputes)*. All three actions are wired to configurable `execute_action` buttons; the existing bare "Release to Picking" button (which used `transition_status`) is **re-pointed** at `release_to_picking` rather than kept alongside it. **Approval gate for release deferred → see §12.**

## 10. Cascade observability (OpenTelemetry distributed tracing) — committed
**What:** model a cascade chain as a distributed trace, reusing the existing HTTP OTel setup — **inject** the trace context (`traceparent`) on the event payload (the same carrier as the loop-guard visited-set), **extract** + start a child span on the dispatch side (`OnEntityEvent`), one span per hop. Optionally a dropped-cascade metric.
**Why deferred:** observability infrastructure, separate from the loop-guard/cascade work (v1 keeps the existing swallowed-error log line).
**Why it matters:** the whole `A→B→C` (and a cut `A→B→A`) becomes one connected trace — ties a dropped cascade to its chain, shows hop timing (surfaces the read-your-writes ordering), and gives the error/latency signal that makes the **F2 outbox** a data-driven decision instead of a guess.
**Trigger:** schedule alongside/just before F2; confirm the existing OTel propagator/exporter when wiring.
**Carrier note:** P1 builds the lineage carrier as a small *extensible* struct so `traceparent` slots in without re-plumbing.

**Note:** the "Validated → bus-routing" rows (priority, reason_code, etc.) are guarded only by field-level validation in `bus.Update`, so the bus-routing follow-up (§1) resolves them; until then they're low-risk (worst case an invalid enum, often also DB-CHECK-guarded). The "Protect-only" rows should never be workflow-settable at all — they're side-effects of picking/ledger ops, written by those flows, not by setting the field.

## 11. Whole-table protections — control plane / status defs / warehouse structure (SHIPPED 2026-06-12)
**What shipped:** 21 tables protected whole-table (`reg.ProtectEntity(table, "")` in `PopulateProtected`), extending the §9 field-level list. Categories: ENGINE (9 workflow.* tables), TABLE BUILDER (`config.table_configs`), RBAC (`core.roles`/`user_roles`/`table_access`), STATUS/REF DEFS (5 status-definition tables), WAREHOUSE STRUCTURE (`inventory.warehouses`/`zones`/`inventory_locations`). Full table + rationale in **WRITE_PATH §1b**. Workflow-protected only (not immutable — humans curate via the normal CRUD path). Tests: `Test_PopulateProtected_WholeTableInvariants` + DB-backed `Test_ProtectedFields_ResolveToRealColumns`.

**Deliberately left writable:** `core.users`, `hr.user_approval_comments`, and tie-in-free taxonomy tables (`products.brands`/`product_categories`, `assets.asset_conditions`/`asset_types`/`asset_tags`, `geography.*`, `hr.offices`/`titles`). Revisit taxonomy if code tie-ins ever appear.

**Two out-of-scope follow-ups surfaced here (NOT done):**
1. **Finding C5** — `ruleapi.create`/`update` can set `is_active` outside the activation-cascade gate. This is an HTTP author path (not the engine, not the generic handlers), so the protected registry does not cover it. Needs a separate gate at the ruleapi/ruleapp layer.
2. **FormData RBAC access** — the FormData write path (`formdata_registry`) has its own entity→bus dispatch and does not consult the protected registry; whether it should be blocked from RBAC tables is a separate question from the generic-workflow-handler block shipped here.

## 12. Approval gate before release-to-picking / order fulfillment
**What:** an optional approval step in front of the `release_to_picking` action (built in `action-buttons-phase2`, see §9 note). Releasing/fulfilling a customer order should optionally require sign-off before the order flips to PICKING and pick tasks are generated. Most natural shape: gate the release behind the existing `seek_approval` → `resolve_approval_request` flow (approve, then the resolve cascades to `release_to_picking`), or a rule that requires an approval record before the release action runs.
**Why deferred:** `action-buttons-phase2` ships the working `release_to_picking` capability first (button + workflow-triggerable action). Approval is an **additive gate**, not a prerequisite for the action to function — and the action being workflow-triggerable means an approval cascade can be layered on later without changing it. Decided 2026-06-15 (jake): "get the action working, worry about approval as a follow-up."
**Trigger:** a concrete requirement that fulfillment release needs sign-off (e.g. credit hold, manager approval for large/priority/high-value orders, allocation review).
**Key files (when picked up):** the `release_to_picking` handler (`business/sdk/workflow/workflowactions/inventory/release_to_picking.go`); approval primitives (`seek_approval`/`resolve_approval_request` handlers + `approvalrequestbus`); rule wiring in `business/sdk/dbtest/seed_workflow.go`. Note `SupportsManualExecution`/`action_permissions` already let you restrict *who* can release manually — the gate adds *workflow-enforced* sign-off on top.
