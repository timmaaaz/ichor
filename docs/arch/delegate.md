# delegate

[bus]=business [app]=application [api]=HTTP [db]=store [sdk]=shared
→=depends on ⊕=writes ⊗=reads ⚡=external [tx]=transaction [cache]=cached

---

## Delegate [sdk]

file: business/sdk/delegate/delegate.go
key facts:
  - Delegate struct: log *logger.Logger, funcs map[string]map[string][]Func
  - Thread-safe read after startup registration (no lock on Call path)
  - One Delegate instance per binary, shared across all domains in that process.
    The server wires its instance in all.go; the standalone Temporal worker
    (api/cmd/services/workflow-worker/main.go) wires its OWN instance.
  - THE GENERAL BEST-EFFORT CROSS-DOMAIN HOOK. Package doc, verbatim: "make function
    calls between different domain packages when an import is not possible." It is an
    indirect, in-process, best-effort function call whose sole purpose is dodging
    import cycles — NOT an event bus, durable log, or transaction participant.
  - Best-effort by contract: Call() logs any handler error and returns nil
    unconditionally (delegate.go:59). A subscriber that fails/panics never fails the
    originating write.
  - 205 call sites across 65 files in business/domain/ (verified 2026-03-09). Every
    [bus] Create/Update/Delete calls b.delegate.Call after its write. This is
    UNCHANGED by F2 (2026-06-17): the durable workflow cascade moved OFF the delegate
    (see below) but b.delegate.Call STAYS on every bus as the cross-domain extension
    point. "No subscriber today" ≠ dead code — a future domain can subscribe without
    re-touching the bus.

  // ✓ verified 2026-03-09
  Register(domainType string, actionType string, fn Func)
<!-- lsp:hover:48:21 -->
```go
func (d *Delegate) Call(ctx context.Context, data Data) error
```
<!-- lsp:refs:48:21 --> count=205 (excl. test mocks)

---

## ⚠ Cascade dispatch is NO LONGER a delegate subscriber (F2, 2026-06-17)

Before F2 the durable workflow cascade was a delegate subscriber: a workflow
DelegateHandler registered for every domain in a RegisterDomain loop (over
workflowdomains.Registrations()) in BOTH binaries. F2 moved the cascade OFF the
best-effort delegate onto a durable transactional outbox + polling relay (the full
seam is in workflow-engine.md):

  - A cascade-relevant bus now ALSO calls b.outbox.Emit(ctx, data) in the SAME tx as
    the entity write (business/sdk/outbox). temporal/relay.go drains that row and is
    the SOLE cascade dispatcher.
  - DELETED (F7.1): temporal/delegatehandler.go — the DelegateHandler struct,
    NewDelegateHandler, RegisterDomain, and the handleEvent goroutine.
  - RELOCATED: the TriggerEvent enrichment (extractEntityData / computeFieldChanges /
    extractIDViaReflection) → business/sdk/workflow/temporal/enrichment.go, now called
    by the relay instead of inline in the handler.
  - The b.delegate.Call on every cascade bus is KEPT as the best-effort hook; it
    simply has NO cascade subscriber listening anymore.

⚠ Do NOT describe the delegate as "cycle-break-only again" or "shrunk" — its purpose
(general best-effort cross-domain hook) is unchanged; only the cascade tenant left.

---

## Live delegate subscribers (after F2)

Three best-effort subscribers remain — each a PLATFORM INVARIANT universal across
ALL customers (see the governance rule below):

| Subscriber | file | Listens for | Best-effort reaction |
|---|---|---|---|
| permissions [bus] | business/domain/core/permissionsbus/permissionsbus.go:56-73 | role / tableaccess / userrole — updated + deleted | recompute permission cache |
| alertws [api] | api/domain/http/workflow/alertws/delegate.go:28-29 | userrole — created + deleted | WebSocket notify |
| rule-cache reload [sdk] | business/sdk/workflow/trigger.go:505 (RegisterCacheInvalidation) | rule lifecycle | reload trigger rule cache |

→ Making delegate.Call propagate errors GLOBALLY would wrongly couple a domain write
to a permission-cache or WebSocket hiccup. Best-effort is *correct* for these three.

---

## Three cross-domain mechanisms (choose by delivery guarantee)

When one domain's write must affect another domain, pick the mechanism by the
GUARANTEE you need — not by convenience. This is the spine of the cross-domain model:

| Mechanism | Guarantee | Where | Example |
|---|---|---|---|
| Direct app-layer orchestration (several buses, one tx) | must-happen **ATOMIC** | app layer | pickingapp; the Path-B workflow handlers (allocate/createpo: BeginTxx + NewWithTx) |
| Outbox → relay | must-happen **ASYNC** cascade (durable, at-least-once) | business/sdk/outbox + temporal/relay.go | domain write → workflow engine cascade |
| Delegate | **survivable-if-missed** best-effort reaction | business/sdk/delegate | the 3 subscribers above |

## ⚠ Governance — when to register a delegate subscriber

Register a delegate subscriber ONLY for functionality UNIVERSAL across ALL customers
(platform invariants: permission-cache recompute, WebSocket notify, rule-cache
reload). The delegate is hardcoded platform wiring identical for every customer, so
anything on it applies to all customers unconditionally; workflows are the
customer-configurable layer.

  - Customer-specific / configurable cross-domain reactions → WORKFLOWS (per-customer
    rule-configured), NEVER the delegate.
  - Must-happen-atomic cross-domain work → direct app-layer orchestration in one tx,
    NEVER the delegate.

Decision test — "is this needed for EVERY customer, always?"
  yes              → delegate subscriber
  customer-varying → workflow rule
  must-be-atomic   → app-layer orchestration

---

## Data [sdk]

file: business/sdk/delegate/model.go
```go
type Data struct {
    Domain    string
    Action    string
    RawParams []byte   // JSON-encoded event payload
}

type Func func(context.Context, Data) error
```

---

## StandardActionConstants [sdk]

Every domain package defines three action constants:
  ActionCreated  = "{entity}.created"
  ActionUpdated  = "{entity}.updated"
  ActionDeleted  = "{entity}.deleted"

Payloads (encoded in RawParams):
  ActionCreated  → { EntityID uuid, UserID uuid, Entity T }
  ActionUpdated  → { EntityID uuid, UserID uuid, Entity T, BeforeEntity T }
  ActionDeleted  → { EntityID uuid, UserID uuid, Entity T }

Every cascade-relevant [bus] Create/Update/Delete now does BOTH, after the DB write:
  b.outbox.Emit(ctx, ActionCreatedData(entity))   // durable cascade (return err)
  b.delegate.Call(ctx, ActionCreatedData(entity)) // best-effort hook (no cascade sub)
The SAME delegate.Data value feeds both — the outbox persists it (payload column),
the delegate fans it out in-process. Emit's error is propagated (return err) so
mid.BeginCommitRollback rolls back the entity row + the outbox row together; Call's
error is swallowed.

Beyond typed [bus] writes, two synthesized paths also produce cascade events; both
now emit to the outbox AND fire delegate (same dual write):
  - Synthesized events: generic data handlers (update_field / create_entity /
    transition_status) fire workflow.SyntheticEventData() after a confirmed raw-SQL
    write so it cascades like a bus write —
    business/sdk/workflow/workflowactions/data/synthesize.go. The handler resolves the
    target table → (domain, entity) via workflowdomains.ReverseMap(). The handler holds
    the outbox Writer via data.WithOutbox(...).
  - allocation_results: workflowbus fires workflow.ActionAllocationResultCreatedData()
    (domain workflow.AllocationResultDomainName, action "created") after writing an
    allocation_results row — business/sdk/workflow/event.go. workflowBus cannot hold an
    *outbox.Writer (outbox imports workflow → cycle), so it takes the bound Emit method
    via .WithOutboxEmitter(outboxWriter.Emit).

---

## ⚠ Adding a new domain that should CASCADE (fire workflow events)

The cascade now flows through the outbox, NOT the delegate subscriber. To make a new
domain cascade:

  business/domain/{area}/{entity}bus/{entity}bus.go               (define ActionCreated/Updated/Deleted consts + ActionCreatedData/etc helpers)
  business/domain/{area}/{entity}bus/{entity}bus.go               (in Create/Update/Delete: call b.outbox.Emit(ctx, data) and `return err`; keep b.delegate.Call as the best-effort hook)
  business/domain/{area}/{entity}bus/{entity}bus.go               (add the WithOutbox(*outbox.Writer) option + b.outbox field — mirror an existing cascade bus)
  business/sdk/workflowdomains/workflowdomains.go                 (add the (schema, domain, entity) registration — drives the outbox entity-name map + relay + the ReverseMap)
  api/cmd/services/ichor/build/all/all.go                         (append .WithOutbox(outboxWriter) to the bus construction)
  api/cmd/services/workflow-worker/main.go                        (if the worker constructs this bus: append .WithOutbox(outboxWriter))
  business/sdk/dbtest/dbtest.go                                   (append .WithOutbox(outboxWriter) so the integration suite exercises the live relay path — F8 parity)
  verify: the Registrations() coverage test (every cascade bus emits) goes RED if a bus has WithOutbox but forgets to call Emit

## ⚠ Registering a new best-effort delegate subscriber (Register() call)

Only for platform invariants universal across ALL customers (governance rule above).
NOTE: workflowdomains.Registrations() no longer drives a delegate RegisterDomain loop
(that was the cascade subscriber, deleted at the F2 cutover) — it drives the outbox.

  {subscriber_package}/{file}.go                                  (implement Func: func(context.Context, Data) error; call del.Register(domain, action, fn))
  composition root that owns the subscriber                        (server: all.go; worker: workflow-worker/main.go — wire the Register call where that subscriber's deps are built, e.g. permissionsbus does it in its own constructor)

## ⚠ Changing Data struct shape

  business/sdk/delegate/model.go                                   (Data struct)
  ALL 205 [bus] files that call delegate.Call() with ActionCreatedData/etc helpers
  business/sdk/workflow/temporal/enrichment.go                     (relay decodes RawParams via reflection — extractEntityData; was delegatehandler.go pre-F2)
  business/sdk/outbox/emit.go + model.go                           (payload = JSON-marshalled Data; eventTypeForAction maps Action → event_type)
  verify: findReferences(business/sdk/delegate/delegate.go:48:21) — confirm exact call site count before mass edit
