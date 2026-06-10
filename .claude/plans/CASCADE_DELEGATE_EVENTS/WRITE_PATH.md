# WRITE_PATH — How generic workflow writes should happen (validation + cascade + bus-routing)

> **Status:** active design. Domain-trace probe (2026-06-10) **corrected** the bus-routing thesis — see §8. **Correctness-weighted** per maintainer.
> **Companions:** [`INVESTIGATION.md`](./INVESTIGATION.md) (verified facts), [`DESIGN.md`](./DESIGN.md) (loop-guard solution). This doc owns the *write-path / validation* dimension — orthogonal to the loop guard, surfaced while deciding the M1 fix.
> **Decision criterion (maintainer, 2026-06-10):** heavily weight the architecturally CORRECT option over the incrementally cheaper one.

---

## 0. Why this dimension exists (extends INVESTIGATION §5)

The generic raw-SQL actions (`update_field`, `create_entity`, `transition_status`) bypass the domain bus, so they skip not just delegate events but **all** business-layer validation / invariants / side-effects. Synthesizing events (the M1 fix) closes the *event* gap only. The *validation* gap is arguably the more dangerous: a missed cascade fails to fire something; an unvalidated raw write can put an entity into a domain-illegal state — corrupt data — directly.

## 1. The protected-field reality (probes) — it's SMALL

- Real write invariants live in dedicated **action-verb** bus methods (`Approve`/`Reject`/`Claim`/`Execute`/`ReceiveQuantity`/`UpsertQuantity`), **NOT** in the generic `Update` (a plain field-set). Only **6 of 77 buses** carry any guard.
- PROTECTED ≈ **9 entities / ~30 (entity,field) pairs**: `purchase_orders`, `inventory_adjustments`, `transfer_orders`, `orders`, `order_line_items`, `inventory_items`, `purchase_order_line_items`, `users` (+ `inventory_transactions` ledger by convention). Load-bearing fields = **status/state fields + a few quantity/ledger fields**.
- **FK columns drop out** — Postgres enforces existence (`REFERENCES … ON DELETE RESTRICT`); no app-layer protection needed.
- **Money fields mostly drop out** — pass-through; no bus recomputes totals from line items.
- Protected unit is **(entity, field)**, not whole table.

## 2. The bus-routing mechanism — FormData IS a generic entity→bus dispatcher

- `app/sdk/formdataregistry/registry.go` + `app/domain/formdata/formdataapp/formdataapp.go` (`executeCreate:437`, `executeUpdate:458`): generic entity-name → typed `NewXxx`/`UpdateXxx` decode → `app.Create/Update`.
- A FormData write runs `app.Validate()` (struct `validate` tags) and goes **app → bus → `delegate.Call`** (verified `assetbus.go:85`, `ordersbus.go:217`).
- **What it closes:** the cascade gap (delegate fires) + **field-level** validation (type/enum/format tags).
- **What it does NOT close:** the **state-machine invariants** — those live in the action-verb methods, which `bus.Update` does not call (see §8 Verdict a).
- **Coverage:** 58 entities registered (vs ~68 in `tables.go`); only 3 have `QueryByNameFunc` (FK-by-name).
- Decode is **fully generic** (identical 3-line `json.Unmarshal`+`Validate` closures) — one dispatcher works for all entities.

## 3. Codebase-consolidation upside (real, but narrower than first thought)

For **plain** fields, routing through the bus is structural simplification: deletes the raw-SQL query construction (incl. FK auto-create branch `updatefield.go:342`), avoids writing synthesize-event code (bus fires the delegate natively), consolidates two "what's writable" lists (`tables.go` 68 + formdata 58) toward one, and restores "all writes for domain X go through `Xbus`." **But** (per §8) it does **NOT** shrink the protected-list, and it adds a per-entity audit + a silent-drop hazard. Net: a worthwhile *consolidation* for plain fields, not a cure-all.

## 4. The fork — where generic writes live

| | **A — stay in `business`** (raw SQL + synthesize + protected-list) | **B — move to `app`** (route through FormData dispatcher) | **C — hybrid** (dispatcher for plain fields, protected-list for guarded) |
|---|---|---|---|
| Cascade gap | fixed by synthesize | fixed free (bus fires it) | both |
| Field-level validation | open | fixed free | fixed where routed |
| **State-machine invariants** | **protected-list → typed action** | **STILL bypassed by `bus.Update`** | **protected-list → typed action** |
| Protected-list | needed (~30 pairs) | **STILL needed** (guarded fields bypass `bus.Update` too) | needed |
| Silent column/tag-drop risk | none (raw writes the column) | **yes — per-entity audit** | yes, per migrated entity |
| Layering work | none | L restructure (business→app interface) | L for the dispatcher half |
| End state | two write paths | one path for plain fields | converges on one for plain fields |

**Key correction:** bus-routing does NOT eliminate the protected-list — see §8. The protected-list is the real fix for guarded fields under *every* option.

## 5. Layering blocker

Handlers are `business/`-layer, import only `business/domain/*/bus`; registry + `app.Create/Update` are `app/`; `business` never imports `app` (verified). Minimal-violation path = a narrow dispatcher **interface in `business`**:
```go
type EntityDispatcher interface {
    Create(ctx, entity string, data json.RawMessage) (any, error)
    Update(ctx, entity string, id uuid.UUID, data json.RawMessage) (any, error)
}
```
implemented in `app/` (lifting `executeCreate`/`executeUpdate`), injected at `all.go`. Per-entity wiring ≈ 30–35 lines + 1 import + 1 constructor param; arg-order varies (`Update(ctx,model,id)` vs `Update(ctx,id,model)`).

## 6. Cross-cutting flags

- **FormData latent tx bug:** `UpsertFormData` opens `tx` (`formdataapp.go:179`) but closures write through base-pool app instances, not `tx`; each write auto-commits and each `delegate.Call` fires pre-outer-commit. No atomic multi-write today. (Reliability/ordering, INVESTIGATION §9 Q6.)
- **Frontend:** `update_field` config is **free-text** (`useActionConfigForms.ts:240-273`), NOT a table-list picker — protected-model FE cost is Small (backend-authoritative rejection) to Medium (real picker). `create_entity` not on FE.
- **Manifest drift confirmed:** `resolve_approval_request` under-declares; approve/reject omit `updated_*`; `allocate` unqualified `allocation_results` → Item-5 consistency test earns its place regardless.

## 7. Decision (DECIDED 2026-06-10, corrected by the domain trace)

- `[DECIDED]` **The protected-list is the must-do correctness core** — block generic writes on the ~30 guarded `(entity,field)` pairs and route them to the typed action-verbs (`approve_po`, `transition_status`-with-guards, etc.). Small, high-value (closes the data-corruption hole), needed under every option, and the only thing that actually fixes the guarded-field bypass. Independent of bus-routing.
- `[DECIDED]` **Cascade for plain-field generic writes = synthesize now; bus-routing (Option B) is a per-entity consolidation FOLLOW-UP** (→ FOLLOW_UP.md §1), NOT a prerequisite and NOT a protected-list eliminator. Adopt incrementally after the core ships.
- Loop guard (DESIGN §10 Items 1–4) is **unaffected** and remains the gating prerequisite.

## 8. Worked example (domain trace, 2026-06-10) — the correction

- **Plain field `orders.priority`:** clean. `{"id":..,"priority":"high"}` → generic `json.Unmarshal` into `UpdateOrder` (`ordersapp/model.go:312`, tag `priority`) + `Validate()` (`oneof=low medium high critical`) → `ordersbus.Update:146` → `delegate.Call:217`. DB column `priority` == JSON tag `priority` → **MATCH, safe.**
- **Protected field `purchase_orders.status` → VERDICT (a): bus-routing relocates the bypass, does not fix it.** `UpdatePurchaseOrder` exposes `PurchaseOrderStatusID` as a settable pointer (`model.go:328`); `purchaseorderbus.Update:155-157` sets it with no guard; the `ErrAlreadyApproved` guard exists ONLY in `Approve:296-301`/`Reject:329-334`. Routing `update_field` for PO status → `app.Update` → `bus.Update` sets status with no state-machine check — same bypass as raw SQL today (not a regression, but explicitly not a mitigation). → **the protected-list cannot be replaced by bus-routing.**
- **Silent column/tag-drop hazard (real):** workflow uses raw DB column names; app models decode by JSON tag; a non-matching/unexposed column drops silently on `json.Unmarshal` (no error, zero effect) — worse than raw SQL. Requires a per-column tag/column audit across all 58 entities before reroute. `updateOrderTotals` (`formdataapp.go:807`) is the working precedent AND the trap (works only because its columns match tags).

## 9. What Option B (full bus-routing) actually is — and where it sits (2026-06-10)

**B is a FOLLOW-UP, not a prereq.** It is not a feature; it is a refactor of *how the two generic data actions write*. Shape:

- **Bridge (one-time):** define `EntityDispatcher` (§5) in `business/sdk/workflow`; implement in `app/` by lifting `executeCreate`/`executeUpdate`; inject at `all.go`; fix/scope the FormData tx bug (§6).
- **Per-entity migration (×58 — the bulk, repeated):** audit every writable column vs the app model's JSON tags (the gating silent-drop risk); register entity if absent (~10 of 68 whitelisted aren't; ~30 lines each); reconcile `Update` arg-order; add `QueryByNameFunc` for FK-by-name on `create_entity` paths (3/58 today); test write→bus→validate→delegate.
- **Flip + cleanup:** route migrated entities to the dispatcher; once all migrated, delete the raw-SQL + synthesize code.

**Why follow-up, not prereq:** the core cascade plan ships complete *without* B — M1 cascades via **synthesize**, M2 buses get `delegate.Call`, the **loop guard** makes it safe, the **protected-list** routes guarded fields to typed actions. Making B a prereq would gate the whole feature behind a 58-entity migration (against §0). Even correctness-weighted, B-as-prereq is *less* correct: the dangerous hole (guarded-field bypass) is already closed by the protected-list; B's remaining gain (one write path for plain fields) is consolidation best done incrementally — rushing 58 entities invites the silent-drop bugs.

**B touches none of:** the loop guard, the M2 fixes, or the protected-list. It only swaps the M1 generic-write *implementation* (raw+synthesize → bus-routing). → core plan = loop guard + M2 + M1-synthesize + protected-list; **B = later consolidation track.**
- **Mechanicals:** decode fully generic; per-entity wiring ~30–35 lines; FK-by-name needs `QueryByNameFunc` (3/58 today); a `business`-side `EntityDispatcher` interface (§5) resolves the layering.
