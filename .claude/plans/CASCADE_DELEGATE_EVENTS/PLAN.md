# PLAN — Cascade Delegate Events: overarching build roadmap

> Created 2026-06-10. The master sequence for the whole effort: the **CORE plan** (decided design → shippable) + the **FOLLOW-UP tracks** (committed vs on-demand), so nothing gets skipped.
> Design is fully decided: [`DESIGN.md`](./DESIGN.md) (loop guard), [`WRITE_PATH.md`](./WRITE_PATH.md) (write-path), [`FOLLOW_UP.md`](./FOLLOW_UP.md) (deferred detail), [`INVESTIGATION.md`](./INVESTIGATION.md) (facts).
> When execution starts, mirror this into a `PROGRESS.yaml` (repo convention) for phase tracking. Status: ☐ not started · ◐ in progress · ☑ done.

---

## Tracking table — the "don't skip anything" view

| # | Track | Type | Status | Gate / trigger |
|---|---|---|---|---|
| P0 | Foundations (shared edge fn, value-aware manifest, consistency test) | core | ☐ | — |
| P1 | Runtime loop guard (visited-set) | core | ☐ | after P0 |
| P2 | Static loop detector (graph + SCC + 3-tier) | core | ☐ | after P0 |
| P3 | Protected-list (validation fix) | core | ☐ | after P0 |
| PG | **Guard Verification** (cascades OFF — prove the guard correct in isolation) | core | ☐ | after P0–P3; **GREEN = the §0 precondition that opens P4** |
| P4 | **Cascade enablement (THE GATE)** | core | ☐ | **requires P1+P2+P3 built AND PG green** |
| PL | **Live-System Verification** (cascades ON — prove the live guarded pipeline) | core | ☐ | after P4 |
| P5 | Ship — supersede PR #176 + arch-doc fixes | core | ☐ | after PL |
| F1 | Bus-routing consolidation (Option B) | **committed** | ☐ | after core, per-entity |
| F2 | **Reliability hardening — transactional outbox** | **committed** | ☐ | after core — DO NOT SKIP |
| F3 | Missing typed actions (claim/execute_transfer_order, receive_po_line_item, approve/deny_user) | on-demand | ☐ | a workflow needs the field |
| F4 | Frontend field-picker UX (Path B) | on-demand | ☐ | authors want inline guidance |
| F5 | Arch-doc cleanup + manifest-drift fixes | committed | ☐ | folded into P5 where possible |
| F6 | Alert/approval-creation cascade | on-demand | ☐ | concrete "react to alert/approval" case |
| F7 | Intentional-loop support (`maxReEntries`) | on-demand | ☐ | a real bounded-loop requirement |
| F8 | Cascade observability (OTel distributed tracing + drop metric) | **committed** | ☐ | pair with F2 — makes the outbox data-driven |

---

## CORE PLAN

> Hard rule (DESIGN §0): the loop guard **must be complete and verified before P4 opens the cascade gate.** The broken state is the safe state; turning cascades on is the last mile.
>
> **Construction is two pieces, each with its own testing phase — the gate (P4) is the seam:**
> - **Piece 1 — Build the guard (cascades stay OFF):** P0 → P1 → P2 → P3, then **PG (Guard Verification)**. Each build phase carries its own inline tests; PG is the consolidated adversarial/property/cross-cutting pass that proves the *assembled* guard in isolation. **PG green is the concrete, auditable form of the §0 precondition — it gates P4.**
> - **Piece 2 — Turn it on (cascades ON):** P4 enablement, then **PL (Live-System Verification)** — the end-to-end tests that are *structurally impossible* before P4 (a real loop being stopped, fan-out storms, read-after-commit ordering), then P5 ship/cleanup.
>
> Why split testing per piece: the two pieces test different things under different preconditions. Piece 1's tests run now with cascades OFF; Piece 2's tests *require* cascades ON, so they can't move earlier. Making each a phase turns "verify before the gate" from a sentence into a checkpoint.

### P0 — Foundations / build-prep
**Goal:** shared substrate all later phases need.
- Lift `findDownstreamRules` (`api/.../ruleapi/cascade.go:125`) into a shared bus/sdk function (consumed by P2's detector AND the existing cascade-map endpoint — one implementation).
- Extend `EntityModification` (`interfaces.go:152`) with produced value + operator (value-extension): populate for the 4 enum-const handlers + ~4 config-literal handlers; mark dynamic/templated as **indeterminate**.
- Stand up the consistency-test harness (declared mutations == delegate events that actually fire).
**Done when:** shared edge fn in place; manifest carries values; consistency test green on current handlers.

### P1 — Runtime loop guard (visited-set) — the universal backstop
**Goal:** stop cross-rule re-entry at dispatch time.
- Carry the visited-set on `WorkflowInput.TriggerData` (Continue-As-New-safe), inside a small **extensible `WorkflowLineage` struct** (`{ visitedSet, originatingExecutionID, … }`) — room for `traceparent`/correlation later (F8) without re-plumbing. Stamp the activity ctx once in `activities.go`; read it in `DelegateHandler.handleEvent` **before** the `:86` goroutine; seed the next generation = parent ∪ {(thisRule, entityID)}.
- Re-entry check in the matched-rule loop (`trigger.go:128`) before `ExecuteWorkflow`.
**Tests (integration):** A→B→A stops after one hop · A→B→C→done progresses · convergent sync runs once · visited-set survives CAN.
**Done when:** guard verified end-to-end (no cascades enabled yet).

### P2 — Static loop detector
**Goal:** block provable loops at authoring; surface every cascade as info.
- Build the inter-rule graph over **active** rules via the P0 edge fn + value-aware edges; Tarjan SCC; re-armability check (DESIGN §4a — `changed_to V` fixed-point self-terminates).
- Single-rule auto-match self-loop **hard-block**.
- Enforce at BOTH `prepareRequest` (`workflowsaveapp.go:62`) and `ActivateRule`/`toggleActive` (`workflowbus.go:571`/`ruleapi.go:272`).
- **Three-tier output:** error (block) · warn (indeterminate) · info datapoint (any cascade edge).
**Tests:** provable loop → block w/ path · indeterminate → warn · convergent sync → allowed · self-loop → block · info datapoints surfaced.
**Done when:** detector enforces at both hooks, active-only.

### P3 — Protected-list (the validation/integrity fix)
**Goal:** close the guarded-field bypass.
- Build the block-list from **domain-declared `protected` tags collected at startup** (the `delegate.Register` pattern), unioned with typed-action manifest claims — **no central hand-list** (verify against FOLLOW_UP §9 / code first; pick the exact tag-vs-method form here).
- Enforce in `update_field`/`create_entity`/`transition_status` — block guarded fields with a clear "needs typed action X" error; route where an action exists; `transition_status` is itself subject to the block on invariant-status.
- FE surfaces the rejection (Path A — backend-authoritative; existing error toast).
**Tests:** guarded write blocked + clear error · plain write allowed · `transition_status` blocked on invariant status.
**Done when:** protected-list enforced + FE shows the rejection.

### PG — Guard Verification (Piece 1 exit — cascades stay OFF)
**Goal:** prove the *assembled* loop guard is correct in isolation, before anything is allowed to turn cascades on. This is the §0 hard rule made into an auditable checkpoint. (P1/P2/P3 each ship with their own inline tests; PG is the consolidated cross-cutting pass over the whole guard.)
- **Static detector (P2):** differential test (static `evalGate`/`classifyEdge` ⟺ runtime `evaluateFieldCondition` agree on every decidable verdict — pins static==runtime, fails on drift); property test (Tarjan SCC vs brute-force oracle over random graphs); must-not-block corpus (legitimate state machines / fan-out / convergent syncs → zero blocks); adversarial loop corpus; full operator-matrix coverage. *(Started in P2.4.)*
- **Runtime guard (P1):** visited-set unit + the 4 DESIGN scenarios, Continue-As-New survival, goroutine-ctx propagation. *(Built in P1.3.)*
- **Protected-list (P3):** guarded-field block + clear "needs typed action X" error; `transition_status` blocked on invariant status; plain writes allowed.
- **Coverage discipline:** no silent caps; assert the indeterminate band is *correctly* indeterminate (warns, not blocks).
**Done when:** the guard is provably correct with cascades OFF → **this green is the precondition that unlocks P4.** No flaky/skipped guard tests.

### P4 — Cascade enablement (THE GATE)
**Precondition:** P1 + P2 + P3 built **and PG green**. The only phase that turns events on.
- M1 **synthesize** (manifest-driven) for `update_field`/`create_entity`/`transition_status`; thread the visited-set lineage onto the synthesized event.
- M2: add `delegate.Call` to `workflowbus.CreateAllocationResult` + fix the unqualified `allocation_results` entity name.
- Keep the three sinks (`log_audit_entry`, `create_alert`, `seek_approval`) silent.
**Tests (inline smoke):** the 4 advertised handlers cascade end-to-end · sinks stay silent · consistency test still green. (The exhaustive end-to-end pass is PL.)
**Done when:** cascades live behind the guard.

### PL — Live-System Verification (Piece 2 exit — cascades ON)
**Goal:** prove the *enabled, guarded* pipeline end-to-end — the tests that are structurally impossible before P4 (nothing loops while M1/M2 are off). This is where "does this critical thing actually work" gets answered against a live system (real Temporal + real DB).
- **The decisive test:** a real `A→B→A` across *synthesized* events is actually stopped by the runtime visited-set (the loop that could not fire in P1). Plus `A→B→C→done` progresses.
- **Fan-out / storm:** one write → N rows → N cascades; confirm bounded behavior (static analysis is blind to this; depth-cap was HELD in P1 — re-evaluate here with real fan-out).
- **Ordering / read-after-commit:** a cascaded rule that reads the entity it was triggered by (DESIGN §9 best-effort accepted for v1 — assert the actual behavior; flag if the read-your-writes hazard bites, → F2 outbox).
- **Idempotency / soak:** repeated/duplicate delegate fires don't double-cascade beyond intent; convergent syncs self-terminate live.
- **Full pipeline integration sweep:** the 4 handlers, sinks stay silent, consistency test green, guard never over-blocks a legitimate live chain.
- **M2 allocation_results cascade (PL.M2 — ADDED 2026-06-12, was unowned):** the M2 event is currently non-functional — it passes the tagless `AllocationResult` struct, so `status` (buried in the `AllocationData` blob) never reaches a trigger condition and the seeded Allocation-Success/Failed rules can never fire (zero prod impact: both `IsActive:false`, no prod rules). Make M2 real: (1) flatten the blob into a map Entity in `event.go ActionAllocationResultCreatedData` to surface `status` (retires the M-b note); (2) add `reference_id` to the `allocate.go`/`reserve_inventory.go` result structs so `{{reference_id}}` resolves; (3) activate the rules + prove the live `allocate→order_line_items` cascade — which also unblocks the read-after-commit bullet above for the M2 path. Steps 1–2 are a low-risk event-contract fix that could alternatively land in PR #182. Full detail in `PROGRESS.yaml` PL.M2.
**Done when:** the live guarded pipeline is proven end-to-end; the decisive loop-stopped test is green; the M2 allocation_results cascade is proven live.

### P5 — Ship + cleanup
- Supersede PR #176 (close it; fold its accurate `update_field` prose into user docs; drop the broken `allocation_results` workaround).
- Fix stale arch docs (workflow-engine.md → 24 handlers; cascade-visualization.md → 19 implementors).
**Done when:** PL green + docs honest → ship.

---

## FOLLOW-UP TRACKS (after core — committed ≠ optional)

**Committed (scheduled; do not skip):**
- **F1 — Bus-routing consolidation (Option B):** the ~58-entity audit-and-migrate behind the `EntityDispatcher` bridge. WRITE_PATH §9.
- **F2 — Reliability hardening (transactional outbox):** guarantee each cascade fires exactly once, after commit — closes the read-your-writes + swallowed-error gaps accepted in v1. FOLLOW_UP §3. **This is the one to actively schedule so best-effort isn't permanent.**
- **F5 — Arch-doc cleanup + manifest-drift fixes:** fold into P5 where possible; otherwise a standalone pass. FOLLOW_UP §6.
- **F8 — Cascade observability (OTel distributed tracing):** extend the HTTP OTel pattern across the cascade (trace context on the event payload, span per hop) + optional drop metric. Pair with F2 — it's what makes the outbox a data-driven call. FOLLOW_UP §10.

**On-demand (build when a concrete need appears — tracked, not forgotten):**
- **F3 — Missing typed actions:** `claim_/execute_transfer_order`, `receive_po_line_item`, `approve_/deny_user` (FOLLOW_UP §9 build list).
- **F4 — Frontend field-picker UX (Path B):** FOLLOW_UP §5.
- **F6 — Alert/approval-creation cascade:** FOLLOW_UP §8.
- **F7 — Intentional-loop support:** FOLLOW_UP §7.
