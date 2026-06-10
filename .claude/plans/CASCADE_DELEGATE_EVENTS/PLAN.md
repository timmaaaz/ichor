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
| P4 | **Cascade enablement (THE GATE)** | core | ☐ | **requires P1+P2+P3 verified** |
| P5 | Verification + supersede PR #176 + arch-doc fixes | core | ☐ | after P4 |
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

> Hard rule (DESIGN §0): the loop guard (P1+P2) **must be complete and verified before P4 opens the cascade gate.** The broken state is the safe state; turning cascades on is the last mile.

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

### P4 — Cascade enablement (THE GATE)
**Precondition:** P1 + P2 + P3 complete & verified. The only phase that turns events on.
- M1 **synthesize** (manifest-driven) for `update_field`/`create_entity`/`transition_status`; thread the visited-set lineage onto the synthesized event.
- M2: add `delegate.Call` to `workflowbus.CreateAllocationResult` + fix the unqualified `allocation_results` entity name.
- Keep the three sinks (`log_audit_entry`, `create_alert`, `seek_approval`) silent.
**Tests:** the 4 advertised handlers cascade end-to-end · loop guard catches an A→B→A across synthesized events · sinks stay silent · consistency test still green.
**Done when:** cascades live behind the guard.

### P5 — Verification + cleanup
- Full integration sweep across the pipeline.
- Supersede PR #176 (close it; fold its accurate `update_field` prose into user docs; drop the broken `allocation_results` workaround).
- Fix stale arch docs (workflow-engine.md → 24 handlers; cascade-visualization.md → 19 implementors).
**Done when:** green + docs honest → ship.

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
