# Cascade Delegate Events — Call-Chain Diagrams

Companion to [`INVESTIGATION.md`](./INVESTIGATION.md). All file:line refs verified 2026-06-09 (two rounds, see INVESTIGATION §12–13).

**The one idea:** a cascade is a chain whose only fragile link is the `delegate.Call` right after a DB write. Every other link (storer → DB → Temporal pipeline → next rule) is robust and shared. The bug is always "the chain is severed before `delegate.Call`." The fixes restore/splice that link; the loop guard adds a gate downstream of it.

---

## 1. The working cascade chain (Category B does this today)

```
TRIGGER  (API PUT, or a prior rule's bus write)
   │
   ▼
[bus].Update(ctx, before, item)                    inventoryitembus.go:110
   ├─ b.storer.Update(ctx, ip)            :152 ───────────────► DB write (inventory_items)
   └─ b.delegate.Call(ctx, ActionUpdatedData(before, ip))  :157   ★ THE ONE FRAGILE LINK
        │                                                          (errors here are SWALLOWED — best-effort)
        ▼
delegate.Call(ctx, Data)                           delegate.go:48
   │   in-memory, synchronous dispatch to registered Func(s)
   ▼
DelegateHandler.handleEvent(ctx, Data)             temporal/delegatehandler.go:49
   │   delegate.Data ──reflection──► workflow.TriggerEvent
   │   go func() { …                                 :86   ⚠ ASYNC, off the originating tx
   ▼
WorkflowTrigger.OnEntityEvent(ctx, event)          temporal/trigger.go:102
   ▼
TriggerProcessor.ProcessEvent(ctx, event)          workflow/trigger.go:122
   │   match EntityName + TriggerType + FieldChanges  vs each rule's TriggerConditions
   ▼
startWorkflowForRule → starter.ExecuteWorkflow      temporal/trigger.go:133, 222
   │   workflowID = workflow-{ruleID}-{entityID}-{uuid.New()}   :170  ← unique EVERY fire (no dedup)
   ▼
⚡ Temporal → Worker → GraphExecutor → Activities → ActionHandler.Execute(…)
   │
   └─► the next rule's action runs … which may write again ──► (back to TRIGGER)
```

Reliability caveats on this "working" path (INVESTIGATION §13.1–13.2): the `★` link is best-effort (delegate errors are logged, not returned), and dispatch is async on a goroutine using the shared `ctx` not the handler's `tx` — so a cascaded rule can query the DB before the originating transaction commits.

---

## 2. Where the chain breaks — the two mechanisms

```
 ── MECHANISM 1: raw SQL (update_field, create_entity, transition_status, log_audit_entry) ──

 ActionHandler.Execute(…)
    └─ executeUpdate(…)                        data/updatefield.go:220
         ├─ query := "UPDATE %s SET …"          :222
         └─ sqldb.NamedExecContextWithCount(…)  :248 ───► DB write
              ✗ no bus · no delegate.Call
              ╳━━━ CHAIN SEVERED — engine never hears about it


 ── MECHANISM 2: bus exists but never emits (create_alert, seek_approval, allocation_results) ──

 ActionHandler.Execute(…)
    └─ alertBus.Create(ctx, alert)             communication/alert.go:187
         └─ [bus].Create(ctx, alert)           alertbus/alertbus.go:81
              └─ b.storer.Create(ctx, alert)   :85 ───► DB write (workflow.alerts)
                   ✗ struct has NO delegate field; returns nil
                   ╳━━━ CHAIN SEVERED — engine never hears about it
```

Both severs are at the SAME link as the working chain's `★` — M1 because no bus holds the link, M2 because the bus dropped it.
- M2 variants: `approvalrequestbus.Create:72` (delegate field exists but only `Resolve:143` fires it); `workflowbus.CreateAllocationResult:1050` (delegate field exists, used elsewhere, TODO at :21).

---

## 3. The loop problem (the chain feeding itself, once cascades fire)

```
        ┌──────────────────────────────────────────────────────────┐
        │                                                          │
        ▼                                                          │
   Rule A (on_update X) ──► writes Y ──► delegate ──► Rule B matches
                                                           │
                                                           ▼
   Rule A matches ◄── delegate ◄── writes X ◄──── Rule B (on_update Y)
        │                                                          ▲
        └──────────────────────────────────────────────────────────┘

   Guards present today:  depth counter ✗   visited set ✗   origin flag ✗   Temporal dedup ✗
   → A↔B re-fires forever.   (Kahn's only checks INSIDE one rule's DAG, at save time —
     workflowsaveapp/graph.go; structurally blind to cross-rule relationships.)
```

---

## 4. The same chain, annotated with every fix option

```
Handler.Execute ─►[write]─► delegate.Call ─► DelegateHandler ─► OnEntityEvent ─► ProcessEvent ─► Temporal ─► next rule
                    │             │                │                                                  │
   ┌── M1 FIX ──────┘             │                │                                                  │
   │   synthesize a TriggerEvent  │                │                                                  │
   │   right after the raw-SQL    │                │                                                  │
   │   write, feed OnEntityEvent  │                │                                                  │
   │                              │                │                                                  │
   │   ┌── M2 FIX ────────────────┘                │                                                  │
   │   │   add the missing delegate.Call to        │                                                  │
   │   │   alertbus / approvalrequestbus /         │                                                  │
   │   │   workflowbus.CreateAllocationResult      │                                                  │
   │   │                                           │                                                  │
   │   │                       ┌── LOOP GUARD ─────┴──────────────────────────────────────────────────┘
   │   │                       │   (A) depth counter on TriggerEvent/WorkflowInput — increment per hop, cap
   │   │                       │   (B) visited-set {(ruleID,entityID)} carried on the event — refuse re-entry
   │   │                       │   (C) origin tag on the write — suppress/limit re-cascade for automated writes
   │   │                       │   ▸ best enforced at ONE chokepoint: DelegateHandler or OnEntityEvent
   ▼   ▼                       ▼
 (restore / splice the link)  (add a gate before dispatch)
```

Why the chokepoint matters: every cascade — M1-synthesized or M2-restored — flows through `DelegateHandler`/`OnEntityEvent`. Enforcing the loop guard there covers all producers uniformly, instead of re-implementing it per handler. The guard is MANDATORY before shipping either fix, because both add new event producers to a pipe that currently has no backstop.
