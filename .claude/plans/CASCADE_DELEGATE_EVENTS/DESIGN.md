# DESIGN — Cascade loop prevention (working notes)

> **Status:** active design exploration — NOT final. Decisions tagged `[DECIDED]` / `[LEANING]` / `[OPEN]`.
> **Created:** 2026-06-10 · **Branch:** `feature/cascade-delegate-events`
> **Companion:** [`INVESTIGATION.md`](./INVESTIGATION.md) holds the *verified facts*; this doc holds the *solution space* — ideas, tradeoffs, why we leaned the way we did.

This captures the design conversation so the reasoning (not just the conclusion) is recoverable. Read INVESTIGATION first for the bug; read this for what to do about it.

---

## 0. Sequencing — the one decision everything hangs on

`[DECIDED]` **Cascade IS wanted.** `[DECIDED]` **The loop guard is a prerequisite to enabling cascades, not follow-up work.**

The tempting order — "fix M1/M2 so delegates fire everywhere, then solve loops" — is backwards. Today the *only* thing preventing infinite loops from the 8 broken handlers is that their delegates don't fire; the broken state is the safe state. Turning delegates on *arms* every A→B→A chain into a pipeline we verified has **zero** backstop. And the guard needs the *same* write→delegate→event plumbing you'd be enabling — so if you turn cascades on first, there's no channel left to inject the guard. **Net: design the guard + lineage plumbing first; turning delegates on is the last mile.**

---

## 1. Verified facts this design rests on (see INVESTIGATION §13)

- **Trigger matcher is value-aware** — `changed_to` = `current==value && prev!=value`; also `equals/not_equals/changed_from/greater_than/less_than/contains/in`. (`trigger.go:335-377`)
- **Mutation manifest is value-BLIND** — `EntityModification` = `{EntityName, EventType, Fields[]}`, no value. (`interfaces.go:152-163`)
- **`GetEntityModifications` is on ~19 handlers** (not the ~4 originally thought), incl. action types beyond the arch doc's "21": `approve_po, reject_po, approve_adjustment, approve_transfer_order, reject_transfer_order, reject_adjustment, resolve`.
- **The manifest is currently consumed by NOTHING** in Go — no matcher, no cascade graph, no cycle detection. The static analysis is **greenfield, but the per-action substrate already exists**.
- **No carried lineage/depth** on `TriggerEvent`. But `ActionExecutionContext` already has `TriggerSource` ("automation"/"manual") + `RuleID`/`ExecutionID` — **partial plumbing exists**.
- **No runtime loop guard.** Temporal workflow IDs are unique per fire (`uuid.New()` suffix) → zero dedup. Delegate dispatch is **async (goroutine), off-transaction, best-effort** (errors swallowed).
- **`changed_to V` is a fixed-point latch** — fires only `current==V && prev!=V` (`trigger.go:346-349`); an idempotent re-write does NOT re-fire. `on_update`/empty-conditions auto-matches *every* event (no latch) (`trigger.go:270-289`).
- **Real loop-prone seeded rules exist** — `Self Trigger Rule` (literal `on_update`+`update_field` self-loop, `cascade_seed_test.go`) and `Allocation Success - Update Line Items` (writes `order_line_items` status from `allocation_results`, `seed_workflow.go`). The `on_update`+no-conditions auto-match shape is the **most common** seeded trigger and several **active** rules use it.
- **A cascade-detection / visualization layer already excludes self-loops for *display*** but does NOT enforce at runtime — possible substrate for the static half; `[OPEN]` locate it.

---

## 2. The core idea — cycle detection at TWO times

The two halves are the same cycle-detection idea applied at different moments. They cover each other's blind spots, so we want **both**, not a choice between them.

| | **Static half** (save/authoring time) | **Runtime half** (dispatch time) |
|---|---|---|
| Question | "Is there a cycle in the *rule graph*?" | "Has this *cascade chain* already fired (rule, entity)?" |
| Effect | **Prevents** designed-in loops before they run | **Catches** loops as they happen |
| Strength | Zero side effects; great UX; shows the path | Sees actual values + entity instances; exact |
| Blind to | Dynamic/templated values (→ indeterminate); manifest drift; fan-out storms | Can't prevent — only stops after ≥1 hop |
| Needs | value-aware manifest + cycle detector (greenfield) | carried lineage token (plumbing) |

`[LEANING]` Build both. Static = primary prevention; runtime = backstop for everything static can't prove.

---

## 3. Static half — "sweep triggers" inter-rule cycle detection

**Origin:** user proposal — sweep each workflow's action nodes; if a node's mutation aligns with any *other* rule's trigger, treat the connected rules as one graph and cycle-check; warn/visual-cue at authoring when a node can trigger another workflow.

Refinements from the discussion:

- **It's a SECOND graph, not the existing Kahn's.** Today's `workflowsaveapp/graph.go` Kahn's is *intra-rule* (action DAG). This is a new *inter-rule* graph: nodes = rules (or rule-actions), **edge `X→Y` iff X produces a value that satisfies Y's trigger.** Different node type, different graph. Don't conflate them.
- **Detection model:** a node's *incoming* side is its **trigger**; its *outgoing* side is its **action's mutations**. The back-edge `B→A` is created when **B's output satisfies A's trigger** — i.e., "A's trigger is what flags the loop," but only because B's *specific value* matches it.
- **Value-awareness is REQUIRED, plus a re-armability check** (see §4 / §4a). Field-level matching false-positives badly; even value-aware edges over-flag *convergent* cycles unless you also check whether the cycle re-arms the gate.
- **Use SCC / Tarjan over the whole graph**, not pairwise A↔B — catches `A→B→C→A`.
- **Re-run on every rule SAVE and every ACTIVATION** — either edit can close a loop, and activating a dormant rule can close one against the active set. `[OPEN]` check against *active-only* rules or *all* rules? (active-only = more permissive, makes activation a load-bearing check).
- **Three-way verdict, not boolean:** `provably-safe` / `provably-looping` / `indeterminate`. Block on proven loop, **warn (don't block)** on indeterminate, explain *why* indeterminate.
- **Visual cue** in the editor: a node that can trigger another workflow gets a marker; closing a cycle surfaces the exact path. (Independently valuable — today the editor surfaces nothing.)

---

## 4. Edge semantics — value-awareness, then the fixed-point (re-armability) refinement

`[DECIDED]` static edges must match **produced value ⊨ trigger condition value**, not just field equality.

```
Rule A: trigger (status changed_to 'pending')   → set status = 'approved'
Rule B: trigger (status changed_to 'approved')  → set status = 'shipped'
```
Runtime: `pending → A → approved → B → shipped → (nothing) → done.` Linear, terminates.

- **Field-level edges** (all today's value-blind manifest can build): "A & B both touch `status`" → draws `A→A`, `A→B`, `B→A`, `B→B`. **A pile of false cycles — blocks a normal state machine.**
- **Value-level edges:** A produces `'approved'` → matches only B's `changed_to 'approved'` → `A→B`. B produces `'shipped'` → nothing triggers on it → no edge. **Graph = `A→B`, acyclic. Correct.**

This is the user's "modifying the same field shouldn't flag a loop — only a true loop-back should." It's right, and it *requires* the manifest to carry the produced value.

`[OPEN]` **Manifest gap to close:** extend `EntityModification` with the produced value (+ operator). Where statically known (`transition_status` knows its to-status at `transition.go:560` and currently discards it; `update_field` with a literal) → include it. Where dynamic/templated (`{{trigger.x}}`) → mark **indeterminate** → conservative edge + defer to runtime guard.

### 4a — value-awareness alone over-flags: the fixed-point (re-armability) refinement

Value-aware edges are necessary but **not sufficient** — they over-flag *convergent* cycles. `changed_to V` fires only on `current==V && prev!=V` (`trigger.go:346-349`), so **V is a fixed-point latch**: once the field sits at V the trigger is disarmed until something moves it back off V.

Convergent-sync false positive — why naive value-edges aren't enough:

```
Rule A: on_update line_items, status changed_to 'ALLOCATED' → set orders.status = 'PROCESSING'
Rule B: on_update orders,     status changed_to 'PROCESSING' → set line_items.status = 'ALLOCATED'
```

Value-aware edges draw `A→B` *and* `B→A` (a cycle) — but at runtime it runs **once and stops**: the 2nd time B sets line_items=`'ALLOCATED'` they're *already* ALLOCATED → `changed_to` fails (prev==new) → A never re-arms. Naive cycle detection would wrongly **block this legitimate sync**.

`[DECIDED]` **The flag = cycle + a re-armable (non-convergent) gate**, not a bare cycle. A cycle only loops if it can move a gated field *off* its trigger value (re-arm the latch). Convergent cycles have a built-in exit = the fixed point. ("Not the exact value — the *mutability* of the gated field off its fixed point is what flags it.")

| Cycle edge's trigger gate | Fixed point? | Loops when… |
|---|---|---|
| `on_update` / no conditions (auto-match) | none | **always** — any write re-arms (no latch to converge to) |
| `changed_to V` / `equals V` | yes (F=V disarms) | the cycle **also writes F to ≠V** (resets the latch); else converges → **safe** |

Precision / soundness of the re-armability check:
- Needs **literal** write values. Dynamic/templated writes → can't prove non-reset → **warn, don't block** (conservative) + defer to runtime guard.
- The **runtime visited-set is exact here** — a convergent cycle never re-fires the rule, so it never re-enters `(rule, entity)`; the visited-set never false-trips. This subtlety is a *static-analysis* precision concern only.
- The `changed_to` fixed-point is a **sound, decidable form of an "exit condition"** (structural, not an arbitrary user predicate) — it rehabilitates a restricted version of policy B (§7) without the halting-problem trap.

### 4b — code-grounded unintentional-loop scenarios (real seeded rules)

Both shapes exist (or are one edit away) in the seed: `business/sdk/dbtest/seed_workflow.go`, `api/.../workflow/ruleapi/cascade_seed_test.go`.

**S1 — single-rule self-loop (value-blind; the common one).** Shape of the real *active* rule `Order Updated - Notify Operations` (`on_update orders`, no conditions) with a mutating action:

```
trigger on_update orders   (no conditions → auto-match EVERY update)
action  update_field set orders.<field>
```

The write is itself an `on_update orders` → re-matches → forever. Literally seeded as `Self Trigger Rule` (cascade_seed_test.go); only the *visualization* layer excludes it, **not** the engine. → `[LEANING]` **hard block** — no legitimate single-rule self-loop exists.

**S2 — parent↔child status sync (two rules, A→B→A).** Built on the real `Allocation Success - Update Line Items` rule (writes `order_line_items` status from `allocation_results`) + the natural reciprocal. **Loops iff written with `on_update`/auto-match; self-terminates with `changed_to` + convergent values** (§4a). The danger is the trigger shape, not the two-rule topology.

`[LEANING]` **Highest-leverage authoring lever:** warn whenever an auto-match `on_update` rule's entity is written by *any* rule (incl. itself). The Explore pass found `on_update`+no-conditions is the *most common* seeded trigger shape and several **active** rules use it — so this one heuristic catches most real unintentional loops with no graph math.

---

## 5. Runtime half — the carried token

`[LEANING]` Carry a **visited-set of `(ruleID, entityID)`** on the event; before dispatching rule R on entity E, if `(R,E)` is already in the chain's set → loop → stop.

- **Why visited-set over a bare depth counter:** it naturally distinguishes progression from loop *without* value-awareness. `A→B→C→done` visits distinct `(rule, entity)` pairs → never false-stops. `A→B→A` re-enters `(A, order#5)` → caught exactly. The depth counter is the cruder universal fallback (and the right tool if bounded loops are ever allowed — see §7).
- **Plumbing** `[OPEN]`: the token rides `TriggerEvent → WorkflowInput → ActionActivityInput`, gets injected into the ctx the action passes to its bus write, propagates through `delegate.Call`, and `DelegateHandler` seeds the next event's set = parent ∪ `(thisRule, thisEntity)`. Lives on the **event payload**, not Go `ctx` alone, because it must cross the async/Temporal boundary. Partial infra exists (`ActionExecutionContext.TriggerSource`/`RuleID`).
- **Also covers fan-out storms** (acyclic but 1 write → 1000 rows → 1000 delegates) via a depth/rate component — something static cycle detection can't see at all.

---

## 6. The linchpin — one manifest, three consumers

`[LEANING]` Make "what each action mutates" a **single declared manifest** that drives all three:

```
   ActionMutationManifest (entity, event, field, →value)
        │              │               │
   runtime delegate    static edges    editor cue
   (fire exactly       (build graph    ("this node can
    these)              + cycle-check)   trigger Workflow X")
```

Why this matters: the whole scheme's soundness reduces to the user's condition "delegates on point" — runtime must fire a delegate for *exactly* the mutations the manifest declares. If one declaration drives both the runtime emit AND the static edges, **drift becomes impossible by construction**, and a consistency test (`declared mutations == mutations that fire a delegate in an integration test`) lets CI enforce it. This also **reframes the M1/M2 fix**: fixing M1/M2 = making the manifest authoritative (the investigation already showed the current manifest has drifted/is aspirational).

---

## 7. Policy — block (A) vs exit-clause (B)

User proposed: **A)** don't let users create loops, or **B)** force an exit clause. User leans A ("an unintended loop is a logic flaw").

`[LEANING] A as the backbone — for the provable band.` Agreed: block a *proven* cycle at authoring with a clear "here's the path A→B→A and the trigger that closes it" error. But:

- **A is only the authoring-time half.** It catches what static analysis can *prove*. The **indeterminate** band (dynamic values) still needs the runtime guard. Don't let "we block at save" imply runtime is covered — it isn't.
- **Must warn-not-block on indeterminate**, or you block legitimate dynamic workflows and the feature gets disabled.
- **The trap in B:** an arbitrary user exit clause ("loop until status='done'") is **unsound** — proving the body ever reaches the exit is the halting problem. The only *sound* exit clause is a **bounded variant** (max-iterations / decreasing counter) = the runtime depth cap surfaced as an explicit opt-in.

**The clean reframing — policy choice = runtime mechanism choice:**

| Policy | Runtime mechanism | Behavior on re-entry |
|---|---|---|
| **A** (no loops) | visited-set `(ruleID, entityID)` | refuse any re-entry (precise cycle break) |
| **B** (bounded loop) | depth / iteration cap | allow bounded re-entry up to N |

`[LEANING]` **Visited-set with a per-rule `maxReEntries = N` override (default 0).** Gives strict-no-loops by default (policy A) *and* a sound, explicit, bounded opt-in (policy B) if an intentional-loop use case ever appears — without rework.

---

## 8. Intentional loops

`[OPEN]` — separate *intent* from *danger*. The hazard is a **tight, unbounded** loop, not intent. Real intentional loops in an ERP are usually:
- **time-delayed** ("re-check stock hourly until available" — `delay` action; rate-limited, not a tight spin), or
- **real-world-bounded** (dunning until paid; escalation) — terminate on external state.

`[LEANING]` **Default-forbid (A); do NOT build intentional-loop support yet** — no evidence of a concrete need, and speculative loop-support is complexity to earn with evidence. The visited-set-with-N-override mechanism (§7) keeps the door open without committing now.

---

## 9. Cross-cutting tradeoffs to keep in view

- **False positives (over-approx)** block valid work → feature gets turned off. Mitigation: value-aware edges + warn-not-block on indeterminate.
- **False negatives (manifest drift / dynamic targets)** give *false confidence* (a loop certified safe). Mitigation: single-source manifest + consistency test + runtime backstop.
- **Reliability/ordering (orthogonal to loops):** delegate is best-effort (errors swallowed) and async off-transaction — a cascaded rule can read before the originating tx commits. `[OPEN]` does the design need a transactional outbox / read-after-commit guarantee, or is best-effort acceptable?

---

## 10. Decisions ledger

**`[DECIDED]`**
- Cascade is wanted.
- Loop guard precedes enabling cascades (sequencing §0).
- Static edges must be value-aware, not field-level (§4).
- Static detector = **cycle + a re-armable (non-convergent) gate**, not a bare cycle; `changed_to` fixed-points self-terminate (§4a).

**`[LEANING]`**
- Two-half design: static prevention + runtime backstop (§2).
- Single mutation manifest as source of truth for runtime + static + UI (§6).
- Policy A (block provable loops) as default; warn on indeterminate (§7).
- Runtime mechanism = visited-set `(ruleID, entityID)` + per-rule `maxReEntries` override default 0 (§7).
- Defer intentional-loop support (§8).
- Single-rule self-loops (`on_update`+same-entity write) → hard block; primary authoring warning = auto-match `on_update` rule whose entity any rule writes (§4b).

**`[OPEN]`**
- Activation check scope: active-only vs all rules (§3).
- Exact token-threading mechanism across the Temporal/async boundary (§5).
- Manifest value-extension shape + which handlers can supply static values vs dynamic (§4).
- Reliability/ordering guarantees needed, if any (§9).
- PR #176 disposition (INVESTIGATION §9).

---

## 11. Next — evidence to gather before committing

1. **Do cycles — or plausibly-intentional loops — exist in the currently seeded/active rules?** Zero → policy A ships with no migration pain. Some → learn bug-vs-intentional before picking the mechanism.
2. **`GetEntityModifications` coverage detail:** which of the ~19 return real mods vs nil; which produced-values are statically knowable vs dynamic (sizes the §4 manifest work).
3. **Locate the cascade-detection / visualization layer** — the Explore pass found it *does* exist and already excludes self-loops *for display* (exercised in `cascade_seed_test.go`), but does NOT enforce at runtime. How does it match rules (value-aware? `GetEntityModifications`-based? a different path?)? It may already be a large chunk of the static-half substrate — sizing the build depends on this.
