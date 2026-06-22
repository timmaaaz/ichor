# workflow-engine

[bus]=business [app]=application [api]=HTTP [db]=store [sdk]=shared
→=depends on ⊕=writes ⊗=reads ⚡=external [tx]=transaction [cache]=cached

---

## StateMachine

trigger-event → rule-match → workflow-dispatch → graph-traverse → action-execute
                                                                  ├─ sync: activity returns result
                                                                  └─ async: StartAsync + CompleteActivity
invariant: each rule execution gets unique executionID; workflow ID deterministic prefix

---

## Pipeline (cascade dispatch — F2 outbox + relay)

bus write (Path A/B/C) → b.outbox.Emit(ctx, data) [SAME tx] → workflow.cascade_outbox
  → Relay[sdk] poll → buildEvent (enrichment.go) → contextWithLineage → WorkflowTrigger[sdk]
  → Temporal[⚡] → Worker → Activities[sdk]

The Relay is the SOLE cascade dispatcher after the F2 cutover (2026-06-17) — the old
delegate DelegateHandler subscriber is DELETED (see delegate.md). A cascade-relevant
write persists ONE outbox row in its own tx; a single polling relay drains it. The
3-mechanism cross-domain model (orchestration / outbox / delegate) lives in delegate.md.

event sources → b.outbox.Emit(ctx, data) (alongside the KEPT best-effort b.delegate.Call):
  - all cascade-relevant domain [bus] Create/Update/Delete (Path A human + Path B handlers)
  - generic data handlers (update_field/create_entity/transition_status) fire
    SyntheticEventData after a confirmed raw-SQL write (synthesize.go) — Path C, what
    makes a worker activity's write cascade to downstream rules
  - workflowbus fires ActionAllocationResultCreatedData (domain "allocation_result")

Composition roots wire ONE outbox.Writer each (built identically all three places):
  - server: all.go — Writer injected into 67 cascade buses + workflowBus; STARTS the relay
  - worker: workflow-worker/main.go — Writer injected into its ~5 buses; does NOT start a
    relay (server-only v1). Worker activity writes (Path B/C ORIGINATE here) Emit rows the
    SERVER relay drains
  - tests: dbtest.go — Writer injected into the same buses so the integration suite
    exercises the live relay path (F8 parity), not a retired transport

---

## Outbox [sdk]

file: business/sdk/outbox/{model,store,emit}.go
imports: delegate, sqldb, workflow (NOT temporal — would cycle via the relay)
key facts:
  - Writer.Emit(ctx, delegate.Data) persists ONE workflow.cascade_outbox row on the
    originating tx (sqldb.GetTxExecutor(ctx)); RETURNS its error so the bus propagates
    it (return err) → mid.BeginCommitRollback rolls back entity row + outbox row together
  - nil *Writer = no-op (inert until the F5 cutover injects a real Writer — DESIGN §6;
    this is what lets the Emit calls be added across buses ahead of an atomic flip)
  - No tx on ctx → falls back to the base pool with a LOUD warn (NOT atomic; the
    on-a-tx trip-wire flags any covered path that lands here)
  - Two facts not in delegate.Data are injected by the composition root (keeps the pkg
    api/temporal-free): entityForDomain map[domain]entity (from
    workflowdomains.Registrations()) + lineage func(ctx)[]byte (temporal.MarshalLineageFromContext)
  - Store: Insert / FetchPending (ORDER BY seq, FOR UPDATE SKIP LOCKED) /
    DeletePublished / MarkAttempt / Reap — each takes the sqlx.ExtContext explicitly
  - // bouncer: marker at the emit seam reserves a volume-guard pre-filter (follow-up)
⊕ workflow.cascade_outbox

---

## Relay [sdk]

file: business/sdk/workflow/temporal/relay.go
imports: outbox.Store, workflow, delegate, EventDispatcher (← *WorkflowTrigger)
key facts:
  - The SOLE cascade dispatcher post-F2. Run(ctx) polls every PollInterval (500ms
    default), claims a batch FOR UPDATE SKIP LOCKED ORDER BY seq, dispatches in seq order
  - dispatchRow: buildEvent(row) → contextWithLineage(decodeLineage(row.Lineage)) →
    dispatcher.OnEntityEvent → success: DeletePublished (delete-on-publish); error:
    MarkAttempt, dead after MaxAttempts (5) so it never head-of-line blocks
  - buildEvent (enrichment.go: extractEntityData / computeFieldChanges, relocated from
    the deleted DelegateHandler) rebuilds workflow.TriggerEvent from row.Payload and
    sets event.EventID = row.ID (the dedup key)
  - EventDispatcher interface (OnEntityEvent) extracted so the relay is testable with a
    fake — no Temporal stack. ProcessBatch/Reap exported for deterministic test draining
  - RelayConfig: PollInterval 500ms / BatchSize 100 / MaxAttempts 5 / DeadRowWindow 7d /
    ReapInterval 1h (zero-value fields fall back to these; ICHOR_* at the call site)
  - SERVER-ONLY v1: all.go starts it in a goroutine inside `if cfg.TemporalClient != nil`;
    the worker does not. A LISTEN/NOTIFY fast-path or a broker is a swap behind the same
    table + the same OnEntityEvent boundary (DESIGN §2)
⊗ workflow.cascade_outbox (FetchPending) ⊕ delete / mark-attempt / reap

---

## workflowdomains [sdk]

file: business/sdk/workflowdomains/workflowdomains.go
  (relocated api→business in F8.1 so dbtest can build the outbox Writer without a
   business→api import)
key facts:
  - Single source of (schema, domain, entity) registrations. Post-F2 it drives the
    OUTBOX, not a delegate subscriber loop: the composition roots build the
    entityForDomain map (delegate domain → workflow entity name) for outbox.NewWriter
    from Registrations(); the generic-write handler builds its reverse map from it
  - Registrations() → []EntityReg; references bus DomainName/EntityName consts so a
    drifted name fails the build, not silently at runtime
  - ReverseMap() → map[schema.table]EntityRef for generic-write synthesis; gated by
    data.IsValidTableName so a non-writable / empty-Schema entry (e.g. allocation_results)
    is absent from the map
  - The Registrations()-coverage test asserts every cascade bus actually emits — a bus
    that gets WithOutbox but forgets Emit goes RED (its cascade would silently vanish)

---

## WorkflowTrigger [sdk]

file: business/sdk/workflow/temporal/trigger.go
imports: RuleMatcher[sdk], WorkflowStarter (narrow client.Client interface), EdgeStore[db]
key facts:
  - OnEntityEvent(ctx, TriggerEvent) — entry point; post-F2 the caller is the Relay
  - RuleMatcher.ProcessEvent() → []MatchedRule
  - Loads GraphDefinition per matched rule from EdgeStore
  - Fails open: individual rule failure logged + skipped; other rules proceed
  - Workflow ID: "workflow-{ruleID}-{dedupKey}" (trigger.go:200-214). dedupKey =
    event.EventID (= outbox row id) when drained by the relay, else a per-dispatch
    random executionID fallback (zero-EventID direct test callers). With
    WorkflowIDReusePolicy=REJECT_DUPLICATE (trigger.go:265) a re-published row → same id
    → Temporal rejects the dup: at-least-once emission, effectively-once execution
  - TriggerEvent.EventID added (models.go:29) — the relay sets it = row.ID
  - Task queue: "ichor-workflow-queue" (models.go:18) (tests: "test-workflow-{t.Name()}")

⚡ Temporal.ExecuteWorkflow

---

## TriggerProcessor / RuleMatcher [sdk]

file: business/sdk/workflow/trigger.go  (TriggerProcessor — NOT in temporal/)
imports: workflow.Business[bus]
key facts:
  - RuleMatcher interface extracted for unit test isolation
  - <!-- lsp:hover:88:29 -->
    ```go
    func (tp *TriggerProcessor) Initialize(ctx context.Context) error
    ```
    Initialize loads active rules (NOT LoadRules())
  - ProcessEvent(ctx, TriggerEvent) → MatchResult{MatchedRules[]}
  - Condition evaluation: TriggerConditions JSON matched against TriggerEvent.FieldChanges

---

## TemporalWorkflow [sdk]

file: business/sdk/workflow/temporal/workflow.go
imports: GraphExecutor[sdk], temporal.workflow SDK
key facts:
  - Receives WorkflowInput, delegates execution to GraphExecutor.Execute()
  - Continue-As-New triggered at ~10,000 history events
  - ContinuationState (*MergedContext) preserves accumulated results across CAN

```go
type WorkflowInput struct {
    RuleID            uuid.UUID
    RuleName          string
    ExecutionID       uuid.UUID
    Graph             GraphDefinition
    TriggerData       map[string]any
    ContinuationState *MergedContext
}

type GraphDefinition struct {
    Actions []ActionNode
    Edges   []ActionEdge
}

type ActionNode struct {
    ID            uuid.UUID
    Name          string
    Description   string
    ActionType    string
    Config        json.RawMessage
    IsActive      bool
    DeactivatedBy uuid.UUID
}

type ActionEdge struct {
    ID             uuid.UUID
    SourceActionID *uuid.UUID
    TargetActionID uuid.UUID
    EdgeType       string
    SourceOutput   *string
    SortOrder      int
}

type MergedContext struct {
    TriggerData   map[string]any
    ActionResults map[string]map[string]any
    Flattened     map[string]any
}
```

---

## GraphExecutor [sdk]

file: business/sdk/workflow/temporal/graph_executor.go
key facts:
  - BFS traversal of ActionNode/ActionEdge graph
  - <!-- lsp:refs:37:1 --> names=[EdgeTypeStart,EdgeTypeSequence,EdgeTypeAlways] count=3
    Edge types: start, sequence, always (only 3 types — no true_branch/false_branch)
  - Parallel branches: fire concurrent activities, merge at convergence point
  - BranchInput.ConvergencePoint = uuid.Nil → fire-and-forget (parent close = abandon)
  - activityOptions(): MaxAttempts=3 sync, MaxAttempts=1 async/human (no duplicate queue)

```go
type BranchInput struct {
    StartAction       ActionNode
    ConvergencePoint  uuid.UUID
    Graph             GraphDefinition
    InitialContext    *MergedContext
    RuleID            uuid.UUID
    ExecutionID       uuid.UUID
    RuleName          string
}
```

---

## Activities [sdk]

file: business/sdk/workflow/temporal/activities.go
imports: workflow.ActionRegistry[sdk], AsyncRegistry[sdk]
key facts:
  - Activities struct holds Registry (sync) + AsyncRegistry (async)
  - ExecuteActionActivity: sync handler path
  - ExecuteAsyncActionActivity: async handler StartAsync path
  - selectActivityFunc returns string name ("ExecuteActionActivity" / "ExecuteAsyncActionActivity")
  - toResultMap: handles nil/map/struct via JSON roundtrip (int64→float64 lossy >2^53)

```go
type ActionActivityInput struct {
    ActionID    uuid.UUID
    ActionName  string
    ActionType  string
    Config      json.RawMessage
    Context     map[string]any
    RuleID      uuid.UUID
    ExecutionID uuid.UUID
    RuleName    string
}

type ActionActivityOutput struct {
    ActionID   uuid.UUID
    ActionName string
    Result     map[string]any
    Success    bool
}
```

---

## ActionHandler Interface [sdk]

file: business/sdk/workflow/interfaces.go
<!-- lsp:hover:39:6 -->
```go
type ActionHandler interface {
	// Execute performs the action with the given configuration and context.
	// Returns the result data (type varies by handler) and any error encountered.
	Execute(ctx context.Context, config json.RawMessage, context ActionExecutionContext) (any, error)

	// Validate validates the action configuration before execution.
	Validate(config json.RawMessage) error

	// GetType returns the unique identifier for this action type (e.g., "allocate_inventory").
	GetType() string

	// SupportsManualExecution returns true if this action can be triggered manually via API.
	// Returns false for actions like update_field that should only run via automation.
	SupportsManualExecution() bool

	// IsAsync returns true if this action queues work for async processing.
	// Async actions return immediately with tracking info; sync actions complete inline.
	IsAsync() bool

	// GetDescription returns a human-readable description for discovery APIs.
	GetDescription() string
}
```
<!-- lsp:refs:39:6 --> count=28 (production, excl. test mocks)

registered action types (28, verified 2026-06-12):
  seek_approval, evaluate_condition, delay, update_field, create_entity,
  lookup_entity, transition_status, log_audit_entry, check_inventory,
  allocate_inventory, reserve_inventory, release_reservation, check_reorder_point,
  receive_inventory, commit_allocation, send_email, send_notification,
  create_alert, create_purchase_order, call_webhook, create_put_away_task,
  resolve_approval_request, approve_purchase_order, reject_purchase_order,
  approve_inventory_adjustment, reject_inventory_adjustment,
  approve_transfer_order, reject_transfer_order

Implementors (production only — 3 test mocks excluded):
  business/sdk/workflow/workflowactions/approval/seek.go              (seek_approval)
  business/sdk/workflow/workflowactions/communication/alert.go        (create_alert)
  business/sdk/workflow/workflowactions/communication/email.go        (send_email)
  business/sdk/workflow/workflowactions/communication/notification.go (send_notification)
  business/sdk/workflow/workflowactions/control/condition.go          (evaluate_condition)
  business/sdk/workflow/workflowactions/control/delay.go              (delay)
  business/sdk/workflow/workflowactions/data/audit.go                 (log_audit_entry)
  business/sdk/workflow/workflowactions/data/create.go                (create_entity)
  business/sdk/workflow/workflowactions/data/lookup.go                (lookup_entity)
  business/sdk/workflow/workflowactions/data/transition.go            (transition_status)
  business/sdk/workflow/workflowactions/data/updatefield.go           (update_field)
  business/sdk/workflow/workflowactions/integration/webhook.go        (call_webhook)
  business/sdk/workflow/workflowactions/inventory/allocate.go         (allocate_inventory)
  business/sdk/workflow/workflowactions/inventory/check_inventory.go  (check_inventory)
  business/sdk/workflow/workflowactions/inventory/check_reorder_point.go (check_reorder_point)
  business/sdk/workflow/workflowactions/inventory/commit_allocation.go (commit_allocation)
  business/sdk/workflow/workflowactions/inventory/receive.go          (receive_inventory)
  business/sdk/workflow/workflowactions/inventory/release_reservation.go (release_reservation)
  business/sdk/workflow/workflowactions/inventory/createputawaytask.go (create_put_away_task)
  business/sdk/workflow/workflowactions/inventory/reserve_inventory.go (reserve_inventory)
  business/sdk/workflow/workflowactions/procurement/createpo.go       (create_purchase_order)
  business/sdk/workflow/workflowactions/approval/resolve.go           (resolve_approval_request)
  business/sdk/workflow/workflowactions/procurement/approve_po.go     (approve_purchase_order)
  business/sdk/workflow/workflowactions/procurement/reject_po.go      (reject_purchase_order)
  business/sdk/workflow/workflowactions/inventory/approve_adjustment.go (approve_inventory_adjustment)
  business/sdk/workflow/workflowactions/inventory/reject_adjustment.go (reject_inventory_adjustment)
  business/sdk/workflow/workflowactions/inventory/approve_transfer_order.go (approve_transfer_order)
  business/sdk/workflow/workflowactions/inventory/reject_transfer_order.go (reject_transfer_order)

---

## EdgeStore [db]

file: business/sdk/workflow/temporal/stores/edgedb/edgedb.go
key facts:
  - Read-only: 2 methods (LoadActions, LoadEdges)
  - Custom query (NOT rule_actions_view — view lacks deactivated_by column)
  - LEFT JOIN action_templates for ActionType (NULL template_id → falls back to
    inline action_config "action_type"/legacy "type"; empty string if neither)
  - workflowbus.CreateRuleAction/UpdateRuleAction reject the empty-on-both state
    at write time (validateActionExecutable), so a node here is never unexecutable
  - sql.NullString for nullable UUIDs (deactivated_by, source_action_id)
  - NamedQuerySlice returns nil for empty (NOT ErrDBNotFound)
⊗ workflow.rule_actions
⊗ workflow.action_edges
⊗ workflow.action_templates

---

## WorkflowBus [bus]

file: business/sdk/workflow/workflowbus/ (workflow.Business)
key facts:
  - CRUD for: AutomationRules, RuleActions, ActionEdges, ActionTemplates, Executions
  - Does NOT handle trigger/dispatch — that is TriggerProcessor + WorkflowTrigger
⊗⊕ workflow.automation_rules
⊗⊕ workflow.rule_actions
⊗⊕ workflow.action_edges
⊗⊕ workflow.action_templates
⊗⊕ workflow.automation_executions

---

## DBSchema

workflow.automation_rules       — rule definitions (entity_id, trigger_type, conditions)
workflow.rule_actions           — action nodes attached to rules
workflow.action_edges           — directed edges (type: start / sequence / always)
workflow.action_templates       — reusable action type configs
workflow.automation_executions  — execution history
workflow.cascade_outbox         — F2 transactional outbox (migration 2.42): one row per
                                  cascade event, drained by the relay. id = dedup key,
                                  seq = total order, lineage = loop-guard visited-set
                                  (also F8's traceparent slot); partial index on pending rows

---

## ⚠ evaluate_condition action config JSON tags

read config struct json tags before writing any evaluate_condition seed/test data
file: business/sdk/workflow/workflowactions/control/condition.go
silent fail: wrong json key → field evaluates to zero value → condition = false, no error logged

## ⚠ alert source_rule_id propagation

file: business/sdk/workflow/workflowactions/communication/alert.go — Execute() assigns SourceRuleID
SourceRuleID = uuid.Nil for manual executions (nil execCtx.RuleID)
test isolation: always scope alert queries — alertbus.QueryFilter{SourceRuleID: &ruleID}
never count global alert totals in workflow tests — concurrent subtests pollute the count

---

## ⚠ Execute() MUST return map[string]any — never a typed struct

why:
  1. Temporal deserializes activity results to map — concrete types erased at SDK boundary
  2. MergedContext.ActionResults is map[string]map[string]any — template resolution needs {{action_name.field}}
  3. GraphExecutor reads result["output"] for edge routing — must coexist with result data

required key "output": string matched against ActionEdge.SourceOutput
  if missing: activities.go injects "success" default (silent misroute risk)

typed structs fine internally — only Execute() return must be map
tests: assert to map[string]any, never concrete struct

## ⚠ Execute() failure reporting — soft "failure" output vs hard error (retry semantics)

two ways to report a failure from Execute(), with DIFFERENT engine behavior:
  - hard error (`return nil, err`): Temporal treats the activity as failed → RETRIES it
    (activityOptions MaximumAttempts=3 for normal actions; longRunning/human differ — temporal/workflow.go)
  - soft failure (`return map{"output":"failure", ...}, nil`): no error → NO retry; GraphExecutor
    just routes to the "failure" output edge (activities.go: nil err → Success:true, output preserved)

choose by whether the handler's write is SAFE TO RETRY (idempotent):
  - idempotent / safe to re-run → hard error is fine; it auto-recovers transient blips. Most self-tx
    inventory handlers (receive, allocate, …) hard-error on tx/commit failure.
  - NON-idempotent (creates a fresh row with a new uuid + no dedup key) → a retry makes a DUPLICATE;
    use soft "failure" so a transient failure does NOT auto-duplicate.

worst case is a Commit failure: it is in-doubt (server may have committed after the client saw a
network error) — retrying a non-idempotent create then is the classic duplicate-write hazard.

deliberate divergence in tree: create_put_away_task (createputawaytask.go) returns soft "failure" on
BeginTxx/NewWithTx/Commit failure because putawaytaskbus.Create mints a fresh uuid with no dedup →
retry = duplicate task. Do NOT "simplify" it to match the hard-error siblings. Trade-off accepted: a
transient failure leaves the task MISSING (visible, recoverable) rather than DUPLICATED (silent,
corrupts inventory). Proper long-term fix = idempotency key on Create, after which retry becomes safe
and the whole handler family could go hard.

## ⚠ Adding a new ActionHandler

  business/sdk/workflow/interfaces.go                              (confirm ActionHandler interface satisfied)
  business/sdk/workflow/temporal/activities.go                     (AsyncRegistry vs Registry decision)
  api/cmd/services/ichor/build/all/all.go                          (Register() call in ActionRegistry setup)
  business/sdk/dbtest/seedmodels/                                   (new test seed if handler needs domain data)
  docs/workflow/README.md                                           (update handler catalog)
  decide Execute() failure contract: soft "failure" output vs hard error — see ⚠ above (non-idempotent writes MUST use soft; retry = duplicate)
  verify: goToImplementation(business/sdk/workflow/interfaces.go:39:6) — confirm existing 28 implementors; register new handler alongside them in all.go

## ⚠ Adding a new Edge type

  business/sdk/workflow/temporal/models.go                         (new EdgeType const — task queue file)
  business/sdk/workflow/temporal/graph_executor.go                 (handle in BFS traversal)
  app/domain/workflow/workflowsaveapp/graph.go                     (cycle/validation logic)
  app/domain/workflow/workflowsaveapp/model.go                     (allowed edge types — see workflow-save.md)
  api/cmd/services/ichor/tests/workflow/                            (integration test update)

## ⚠ Changing WorkflowInput shape

  business/sdk/workflow/temporal/models.go                         (WorkflowInput struct)
  business/sdk/workflow/temporal/workflow.go                       (unpack new fields)
  business/sdk/workflow/temporal/trigger.go                        (populate new fields when dispatching)
  business/sdk/workflow/temporal/graph_executor.go                 (consume new fields if needed)
  apitest/workflow.go                                               (update test infra if WorkflowInfra changes)

## ⚠ Cascade path (outbox → relay) — files & reliability tests

seam: bus write → b.outbox.Emit(ctx, delegate.Data) INSERT workflow.cascade_outbox
(SAME tx; migration 2.42) → relay.go polls (FOR UPDATE SKIP LOCKED, ORDER BY seq) →
buildEvent (enrichment.go) → contextWithLineage → WorkflowTrigger.OnEntityEvent.

LINEAGE (the A→B→A loop guard): WorkflowLineage{Visited[], OriginatingExecutionID}
(lineage.go) is serialized into the row's lineage column at emit
(temporal.MarshalLineageFromContext — nil for human/non-workflow writes, so the column
stays NULL and a fresh chain starts) and rehydrated before dispatch (decodeLineage →
contextWithLineage). The guard SURVIVES the serialize→DB→rehydrate round-trip. F2
changed DELIVERY only (best-effort → durable at-least-once + dedup); the guard logic,
which rules match, and which events fire are unchanged.

DEDUP: workflow id = workflow-{ruleID}-{eventID}, eventID = outbox row id,
REJECT_DUPLICATE (see WorkflowTrigger above).

  business/sdk/outbox/{model,store,emit}.go                        (table model + Store + Writer.Emit)
  business/sdk/workflow/temporal/relay.go                          (the polling relay — sole dispatcher)
  business/sdk/workflow/temporal/enrichment.go                     (buildEvent helpers; was delegatehandler.go)
  business/sdk/workflow/temporal/lineage.go                        (WorkflowLineage; MarshalLineageFromContext; contextWithLineage)
  business/sdk/sqldb/context.go                                    (WithTx / GetTx / GetTxExecutor — the tx-on-ctx carrier Emit reads)
  business/sdk/migrate/sql/migrate.sql                             (Version: 2.42 — workflow.cascade_outbox)
  api/cmd/services/ichor/build/all/all.go                          (Writer built + injected into 67 buses; relay STARTED)
  api/cmd/services/workflow-worker/main.go                         (Writer injected; NO relay)
  business/sdk/dbtest/dbtest.go                                    (Writer injected — F8 test parity; BusDomain.OutboxWriter)

reliability tests (changed-pkg only — NEVER go test ./...):
  api/cmd/services/ichor/tests/workflow/actionhandlers/cascade_outbox_test.go (e2e: human → ruleA activity → Emit → relay → ruleB; lineage survives)
  business/sdk/workflow/temporal/relay_test.go                     (TestRelay_* — drain order, retry/dead, reap, buildEvent enrichment, decode lineage)
  business/sdk/workflow/temporal/{lineage_test,trigger_test}.go    (TestGuard_ABA_StopsAfterOneHop, TestCascade_LoopGuard, dedup id)
  business/sdk/outbox/{outbox_test,coverage_test}.go               (Emit atomicity / RYW / pool fallback; every-cascade-bus-emits coverage)

