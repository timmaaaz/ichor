# Phase 14: Documentation Updates

**Category**: backend
**Status**: Pending
**Dependencies**: Phase 13 (Dead Code Removal & Temporal Rewiring)

---

## Overview

Update existing workflow documentation to reflect the Temporal-based architecture (replacing the old RabbitMQ/EventPublisher/QueueManager/Engine pipeline) and create 3 new Temporal-specific documents. Phase 13 removes the old engine; this phase ensures docs match the new reality.

**Scope**: 4 existing files to update + 3 new files to create = 7 total documentation files.

## Goals

1. Update 4 existing docs to remove old RabbitMQ/Engine references and describe Temporal architecture
2. Create 3 new Temporal-specific docs (integration overview, worker deployment, migration guide)
3. Ensure all architecture diagrams accurately reflect the Temporal event flow

## Prerequisites

- Phase 13 completed (old engine removed, TemporalDelegateHandler created, all.go rewired)
- Familiarity with the Temporal temporal package: `business/sdk/workflow/temporal/`
- Understanding of the new event flow: `delegate → TemporalDelegateHandler → WorkflowTrigger → Temporal → Worker → Activities`

### Key Architecture Reference (Post-Phase 13)

```
New Pipeline:
  Business Layer → delegate.Call() → TemporalDelegateHandler.handleEvent()
    → WorkflowTrigger.OnEntityEvent() → Temporal client.ExecuteWorkflow()
    → workflow-worker picks up task → ExecuteGraphWorkflow → Activities

Old Pipeline (REMOVED):
  Business Layer → delegate.Call() → DelegateHandler → EventPublisher
    → QueueManager → RabbitMQ → Engine → ActionExecutor
```

---

## Task Breakdown

### Task 1: Update `docs/workflow/README.md`

**Status**: Pending

**Description**: Replace the old architecture diagram and event flow with Temporal equivalents. Update the Key Files table. Remove references to EventPublisher, QueueManager, Engine.

**Current State** (174 lines):
- Line 23-28: Old architecture diagram (`EventPublisher → RabbitMQ → QueueManager → Engine → ActionExecutor`)
- Line 56-63: Old 7-step event flow referencing DelegateHandler → EventPublisher → RabbitMQ → QueueManager → Engine
- Lines 8-18: Quick Links table references `event-infrastructure.md` as "EventPublisher and delegate pattern"
- Lines 138-152: Key Files table lists `engine.go`, `queue.go`, `eventpublisher.go`, `delegatehandler.go`

**Changes Required**:

1. **Replace architecture diagram** (lines 22-29):
```
┌─────────────────────────────────────────────────────────────────────────────┐
│                           WORKFLOW ENGINE                                    │
│                                                                             │
│  DelegateHandler ──► WorkflowTrigger ──► Temporal ──► Worker ──► Activities │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

2. **Replace event flow** (lines 55-63) with new 7-step flow:
```
1. An entity is created/updated/deleted via API or formdata
2. The business layer fires a delegate event
3. TemporalDelegateHandler converts it to a TriggerEvent
4. WorkflowTrigger evaluates matching automation rules
5. For each matched rule, loads graph definition from DB
6. Dispatches workflow to Temporal (workflow-worker picks up)
7. GraphExecutor traverses action graph, executing each action as an activity
```

3. **Update Quick Links table** (line 17):
   - Change `event-infrastructure.md` description to "Delegate pattern and workflow dispatch"
   - Add links to new docs: `temporal.md`, `worker-deployment.md`, `migration-from-rabbitmq.md`

4. **Update Key Files table** (lines 138-152):
   - Remove: `engine.go`, `queue.go`, `eventpublisher.go`, `delegatehandler.go`
   - Add: `temporal/trigger.go`, `temporal/workflow.go`, `temporal/graph_executor.go`, `temporal/activities.go`, `temporal/delegatehandler.go`, `temporal/stores/edgedb/edgedb.go`
   - Keep: `models.go`, `trigger.go`, `template.go`, `workflowactions/`

**Files**:
- `docs/workflow/README.md` (MODIFY: ~30 lines changed)

---

### Task 2: Update `docs/workflow/architecture.md`

**Status**: Pending

**Description**: Major rewrite of the architecture doc. Remove EventPublisher, QueueManager, Engine, RabbitMQ sections. Replace with Temporal architecture covering GraphExecutor, activities, worker separation.

**Current State** (578 lines) - sections to change:
- Lines 9-61: System overview diagram (shows old EventPublisher → QueueManager → RabbitMQ → WorkflowEngine pipeline)
- Lines 63-95: EventPublisher section (~33 lines) - DELETE entirely
- Lines 97-119: DelegateHandler section (~23 lines) - REWRITE for TemporalDelegateHandler
- Lines 121-153: QueueManager section (~33 lines) - DELETE entirely
- Lines 155-188: Engine section (~34 lines) - DELETE entirely
- Lines 316-368: Complete Event Lifecycle (11-step flow with old pipeline) - REWRITE
- Lines 395-429: Initialization section (old all.go setup with Engine/QueueManager/EventPublisher) - REWRITE
- Lines 431-456: Error Handling section references RabbitMQ dead letter queue - REWRITE
- Lines 458-498: Performance section references RabbitMQ async processing and old executor - REWRITE
- Lines 500-569: Configuration section (RabbitMQ settings, queue types, circuit breakers) - DELETE entirely

**Sections to KEEP** (mostly unchanged):
- Lines 189-213: TriggerProcessor section (still used, already updated in Phase 9)
- Lines 215-260: ActionRegistry section
- Lines 261-295: ActionHandler and EntityModifier interfaces

**Section to DELETE**:
- Lines 296-314: AsyncActionHandler Interface section (replaced by Temporal's AsyncActivityHandler)

**New sections to ADD**:

1. **System Overview Diagram** (replace lines 9-61):
```
┌─────────────────────────────────────────────────────────────────────────────┐
│                           API Layer (ichor service)                          │
│  ordersapi / formdataapi / other apis...                                    │
└──────────────────────────────────┬──────────────────────────────────────────┘
                                   │ delegate.Call()
                                   ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                        Trigger Side (ichor service)                          │
│  TemporalDelegateHandler → WorkflowTrigger → TriggerProcessor               │
│                                   │                                          │
│                                   │ client.ExecuteWorkflow()                │
└───────────────────────────────────┼─────────────────────────────────────────┘
                                    │ (Temporal task queue)
                                    ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                      Execution Side (workflow-worker service)                │
│  ExecuteGraphWorkflow → GraphExecutor → Activities → ActionHandlers         │
│  ├── Linear: executeSingleAction                                            │
│  ├── Parallel: executeParallelWithConvergence                               │
│  └── Fire-and-forget: child workflows + PARENT_CLOSE_POLICY_ABANDON         │
└─────────────────────────────────────────────────────────────────────────────┘
```

2. **TemporalDelegateHandler** (replace old DelegateHandler section):
   - Location: `business/sdk/workflow/temporal/delegatehandler.go`
   - Bridges delegate events to WorkflowTrigger.OnEntityEvent()
   - Non-blocking goroutine dispatch (fail-open)
   - Event type mapping: created→on_create, updated→on_update, deleted→on_delete

3. **WorkflowTrigger** (NEW section):
   - Location: `business/sdk/workflow/temporal/trigger.go`
   - Loads graph definition from EdgeStore (DB)
   - Dispatches to Temporal with WorkflowInput
   - Workflow ID format: `workflow-{ruleID}-{entityID}-{executionID}`

4. **GraphExecutor** (NEW section):
   - Location: `business/sdk/workflow/temporal/graph_executor.go`
   - Deterministic graph traversal with sorted iteration
   - Convergence point detection (BFS-based)
   - Supports all 5 edge types

5. **Workflow Implementation** (NEW section):
   - Location: `business/sdk/workflow/temporal/workflow.go`
   - Continue-As-New at 10K history events
   - Parallel execution via child workflows + Selector
   - Fire-and-forget via PARENT_CLOSE_POLICY_ABANDON
   - Versioning with workflow.GetVersion()

6. **Activities** (NEW section):
   - Location: `business/sdk/workflow/temporal/activities.go`
   - Activities struct holds both Registry and AsyncRegistry
   - ExecuteActionActivity (sync) and ExecuteAsyncActionActivity (async)
   - buildExecContext populates ActionExecutionContext from input

7. **Initialization (all.go)** (replace old section):
   - Show new Temporal block: client.Dial → edgeStore → TriggerProcessor → WorkflowTrigger → DelegateHandler → RegisterDomain calls
   - Conditional on TemporalHostPort (log warning if empty)

8. **Error Handling** (REWRITE):
   - Non-blocking philosophy still applies (goroutine dispatch)
   - Temporal handles retries (configurable per activity type)
   - Regular actions: MaximumAttempts=3
   - Async/human actions: MaximumAttempts=1 (prevent duplicate side effects)

9. **Configuration** (replace RabbitMQ section):
   - `ICHOR_TEMPORAL_HOSTPORT` - Temporal server address
   - Task queue: `ichor-workflow`
   - Worker config in workflow-worker service

**Files**:
- `docs/workflow/architecture.md` (REWRITE: ~400 lines removed, ~350 lines added)

---

### Task 3: Update `docs/workflow/event-infrastructure.md`

**Status**: Pending

**Description**: Rewrite for Temporal. The entire "Two Entry Points" concept changes — FormData no longer uses EventPublisher blocking methods; everything goes through delegate → TemporalDelegateHandler now.

**Current State** (515 lines) - nearly ALL content references old pipeline:
- Lines 14-36: "Two Entry Points" (FormData → EventPublisher, Delegate → EventPublisher) - REWRITE
- Lines 38-148: EventPublisher section (methods, non-blocking design, entity extraction) - DELETE
- Lines 150-219: Old DelegateHandler section - REWRITE for TemporalDelegateHandler
- Lines 384-452: Event Flow Diagram (8-step old pipeline) - REWRITE
- Lines 454-476: Error Handling references EventPublisher/QueueManager - REWRITE
- Lines 506-515: Key Files table lists eventpublisher.go, delegatehandler.go - UPDATE

**Sections to KEEP** (still accurate):
- Lines 222-267: Domain Event Pattern (event.go structure) - KEEP entirely
- Lines 269-343: Integration in Business Layer (delegate.Call patterns) - KEEP entirely
- Lines 345-361: Registration in all.go - UPDATE to show TemporalDelegateHandler
- Lines 362-382: TriggerEvent Structure - KEEP
- Lines 477-504: Cascade Visualization and Condition Node Results - KEEP

**New content to write**:

1. **Overview**: Single entry point now — all events flow through delegate → TemporalDelegateHandler → Temporal
2. **TemporalDelegateHandler**: Structure, RegisterDomain, handleEvent (goroutine dispatch), extractEntityData
3. **Event Flow Diagram**: New 6-step flow (business → delegate → TemporalDelegateHandler → WorkflowTrigger → Temporal → Worker)
4. **Error Handling**: Non-blocking goroutine, Temporal handles retries, fail-open per rule
5. **Key Files**: Update table (remove eventpublisher.go, old delegatehandler.go; add temporal/delegatehandler.go, temporal/trigger.go)

**Files**:
- `docs/workflow/event-infrastructure.md` (REWRITE: ~200 lines removed, ~150 lines added)

---

### Task 4: Update `docs/workflow/testing.md`

**Status**: Pending

**Description**: Update test infrastructure descriptions. Remove RabbitMQ test container references and old Engine/QueueManager patterns. Add Temporal test patterns. Note: Phase 15 rewrites the actual test infrastructure code — this phase documents the new patterns.

**Current State** (822 lines) - sections to change:
- Lines 8-11: Overview references "Real RabbitMQ via testcontainers" - UPDATE
- Lines 28-60: Basic Setup pattern (RabbitMQ container, queue, engine, qm) - REWRITE
- Lines 62-92: "With Engine and Queue Manager" section - DELETE
- Lines 94-165: EventPublisher unit tests - DELETE (file being removed in Phase 13)
- Lines 168-273: Integration tests using EventPublisher/Engine/QueueManager - REWRITE
- Lines 378-437: Graph Execution Tests referencing executor_graph_test.go - DELETE (file removed)
- Lines 626-659: Delegate Handler Tests referencing old delegatehandler_test.go - DELETE
- Lines 661-704: Running Tests section references old test commands - UPDATE
- Lines 718-746: Manual Testing references RabbitMQ - UPDATE
- Lines 783-822: Troubleshooting references QueueManager/Engine - UPDATE

**Sections to KEEP** (still accurate):
- Lines 277-329: Alert Handler Tests (unchanged)
- Lines 478-528: Condition Handler Tests (unchanged)
- Lines 558-610: Edge API Tests (unchanged)
- Lines 612-624: Cascade API Tests (unchanged)
- Lines 754-780: Test Data Seeding (unchanged)

**New content to ADD**:

1. **Overview**: Update to reference Temporal test server instead of RabbitMQ
2. **Test Categories table**: Update with new file locations (Temporal package tests)
3. **Temporal Package Tests** (NEW section):
   - Unit tests: models, graph executor (determinism, edges, convergence), workflow, activities, trigger
   - Integration tests: workflow replay (real Temporal container), edge store (real DB)
   - 155+ tests total in `business/sdk/workflow/temporal/`
4. **Temporal Test Setup Pattern** (NEW section):
   - SDK TestWorkflowEnvironment for unit tests
   - TestActivityEnvironment for activity tests
   - Real Temporal container via `foundation/temporal/` for integration/replay tests
5. **Updated Running Tests commands**:
   - `go test ./business/sdk/workflow/temporal/...` (Temporal tests)
   - `go test -short ./business/sdk/workflow/temporal/...` (skip integration)
6. **Updated Manual Testing**: Reference Temporal UI at port 8280 instead of RabbitMQ at 15672

**Files**:
- `docs/workflow/testing.md` (REWRITE: ~250 lines removed, ~200 lines added)

---

### Task 5: Create `docs/workflow/temporal.md`

**Status**: Pending

**Description**: New comprehensive Temporal integration overview document.

**Sections to write**:

1. **Overview**: What Temporal provides (durability, visibility, retry, crash recovery) and why we chose it
2. **Architecture**: Two-service model (ichor triggers, workflow-worker executes)
3. **Task Queue**: `ichor-workflow` queue, single task queue design
4. **GraphExecutor Design**:
   - Deterministic graph traversal (sorted maps, UUID tie-breaking)
   - 5 edge types: start, sequence, true_branch, false_branch, always
   - Convergence point detection via BFS
   - Parallel execution patterns (convergence vs fire-and-forget)
5. **Activity Types**:
   - Sync actions (evaluate_condition, update_field, send_notification, create_alert, seek_approval)
   - Async actions (send_email, allocate_inventory) — MaximumAttempts=1
   - Human actions (manager_approval, manual_review, etc.) — MaximumAttempts=1
6. **Context Propagation**: MergedContext → template variable resolution → ActionExecutionContext
7. **Continue-As-New**: Triggers at HistoryLengthThreshold (10K events), preserves full MergedContext via ContinuationState
8. **Error Handling & Retry Policy**:
   - Regular: 3 retries, exponential backoff
   - Async/Human: 1 attempt (prevent duplicate side effects)
   - Per-rule fail-open: individual rule failures don't block others
9. **Workflow ID Format**: `workflow-{ruleID}-{entityID}-{executionID}` (deterministic prefix, unique suffix)
10. **Versioning**: `workflow.GetVersion(ctx, "graph-interpreter", DefaultVersion, 1)`

**Key files to reference**: All files in `business/sdk/workflow/temporal/`

**Files**:
- `docs/workflow/temporal.md` (CREATE: ~250 lines)

---

### Task 6: Create `docs/workflow/worker-deployment.md`

**Status**: Pending

**Description**: New document covering the workflow-worker service operations.

**Sections to write**:

1. **Overview**: Separate process from main API for independent scaling
2. **Configuration**:
   - `ICHOR_TEMPORAL_HOSTPORT` — Temporal server address (required)
   - `ICHOR_DB_*` — Database connection (same DB as main service)
   - No RabbitMQ dependency
3. **K8s Deployment**:
   - Base manifest: `zarf/k8s/base/workflow-worker/base-workflow-worker.yaml`
   - Dev overlay: `zarf/k8s/dev/workflow-worker/`
   - ConfigMap with environment variables
4. **Action Handler Registration**:
   - `RegisterCoreActions`: 5 handlers (evaluate_condition, update_field, seek_approval, send_email, send_notification)
   - `RegisterAll`: Full set (requires RabbitMQ + bus dependencies — future)
   - Empty AsyncRegistry (fails gracefully if async workflow dispatched)
5. **Scaling Considerations**:
   - Worker scales independently from API
   - Multiple workers can process the same task queue
   - Temporal handles task routing and load balancing
6. **Monitoring**:
   - Temporal UI (port 8280 via kubectl port-forward)
   - Worker logs: `make dev-logs-workflow-worker`
   - Describe pods: `make dev-describe-workflow-worker`
7. **Graceful Shutdown**: signal.NotifyContext + worker.InterruptCh()
8. **Makefile Targets**: dev-logs-workflow-worker, dev-describe-workflow-worker, temporal-ui

**Files**:
- `docs/workflow/worker-deployment.md` (CREATE: ~150 lines)

---

### Task 7: Create `docs/workflow/migration-from-rabbitmq.md`

**Status**: Pending

**Description**: Migration guide for developers familiar with the old RabbitMQ-based engine.

**Sections to write**:

1. **Why We Migrated**: Temporal provides durability (workflow state survives crashes), visibility (inspect running workflows), structured retries, parallel branch support, and eliminates custom engine/queue/circuit-breaker code
2. **What Changed**:
   - **Removed**: `engine.go`, `executor.go`, `dependency.go`, `queue.go`, `notificationQueue.go`, `eventpublisher.go`, `delegatehandler.go` + all their tests (~10K lines)
   - **Added**: `temporal/` package (~3K lines) + `workflow-worker` service
   - **Replaced**: `DelegateHandler` → `temporal.DelegateHandler`, `EventPublisher` → direct Temporal dispatch, `QueueManager` → Temporal task queue, `Engine` → `GraphExecutor` + Temporal workflows
3. **What Didn't Change**:
   - Same database tables (automation_rules, rule_actions, action_edges, etc.)
   - Same action handlers (all 7 still implement ActionHandler interface)
   - Same delegate event pattern in business layer
   - Same Save API validation (workflowsaveapp)
   - Same template variable resolution
4. **Configuration Changes**:
   - **New**: `ICHOR_TEMPORAL_HOSTPORT` (empty = disabled)
   - **Unchanged**: RabbitMQ still used for alert WebSocket delivery (separate concern)
   - **Removed**: Engine singleton, QueueManager config, circuit breakers
5. **Known Limitations** (post-migration):
   - Alert WebSocket delivery: alerts still write to DB but real-time WebSocket push stopped (REST polling works)
   - Async handler adapters: SendEmail and AllocateInventory StartAsync adapters needed for full async completion
   - FormData event publishing: FormData CRUD works; workflow triggers now fire via delegate events only
6. **Mapping Table**: Old concept → New concept for quick reference

**Files**:
- `docs/workflow/migration-from-rabbitmq.md` (CREATE: ~200 lines)

---

## Validation Criteria

- [ ] All `docs/workflow/*.md` files updated or created (7 files total)
- [ ] No references to `EventPublisher`, `QueueManager`, `workflow.NewEngine` in updated docs
- [ ] No references to `eventpublisher.go`, `queue.go`, `engine.go`, `delegatehandler.go` (old) in Key Files tables
- [ ] Architecture diagrams reflect Temporal flow in README.md and architecture.md
- [ ] New docs created: `temporal.md`, `worker-deployment.md`, `migration-from-rabbitmq.md`
- [ ] Links in `docs/workflow/README.md` Quick Links table include new docs
- [ ] RabbitMQ references only appear in migration guide (historical context) and alert WebSocket caveat
- [ ] `testing.md` references Temporal test patterns and `foundation/temporal/` test container

---

## Deliverables

- Updated `docs/workflow/README.md` (~30 lines changed)
- Updated `docs/workflow/architecture.md` (~400 lines removed, ~350 lines added)
- Updated `docs/workflow/event-infrastructure.md` (~200 lines removed, ~150 lines added)
- Updated `docs/workflow/testing.md` (~250 lines removed, ~200 lines added)
- New `docs/workflow/temporal.md` (~250 lines)
- New `docs/workflow/worker-deployment.md` (~150 lines)
- New `docs/workflow/migration-from-rabbitmq.md` (~200 lines)

---

## Gotchas & Tips

### Common Pitfalls

1. **Don't delete RabbitMQ references everywhere**: RabbitMQ is still used for alert WebSocket delivery (`alertws.AlertConsumer`). The migration guide should explain this. Only remove RabbitMQ references from the *workflow execution pipeline* docs.

2. **FormData entry point changed**: Old docs describe FormData as a separate entry point using `EventPublisher.PublishCreateEventsBlocking()`. Post-Phase 13, FormData fires delegate events like everything else. The "Two Entry Points" concept in `event-infrastructure.md` collapses to one.

3. **TriggerProcessor docs are already accurate**: The TriggerProcessor section in `architecture.md` (lines 189-213) was already updated during Phase 9. Don't overwrite it.

4. **Phase 15 hasn't happened yet**: When updating `testing.md`, document the *planned* test infrastructure (Temporal test server) but note that the actual test infrastructure rewrite happens in Phase 15. Don't document `apitest/workflow.go` as if it's already rewritten.

5. **Don't document internal implementation details**: Keep docs focused on architecture and usage patterns. The phase plan docs and code comments handle implementation specifics.

6. **AsyncActionHandler interface is gone**: The old `AsyncActionHandler` with `ProcessQueued` was removed in Phase 13. The new async pattern uses `AsyncActivityHandler` with `StartAsync` in the Temporal package. Update `architecture.md` accordingly.

### Tips

- Read each existing doc file fully before editing — some sections are still accurate
- Use `grep -r "EventPublisher\|QueueManager\|workflow.NewEngine" docs/workflow/` after all edits to verify cleanup
- The migration guide is the most valuable new doc for developers — prioritize making it clear
- Keep architecture diagrams simple — ASCII art, not complex nested boxes
- Cross-reference between docs (e.g., temporal.md links to worker-deployment.md for ops details)
- Task execution order doesn't matter — all 7 files are independent

---

## Testing Strategy

### Verification

This phase is pure documentation — no code changes, no compilation, no tests to run.

**Verification steps**:
1. Read each file to confirm accuracy against codebase
2. Grep for stale references: `grep -r "EventPublisher\|QueueManager\|workflow.NewEngine\|engine.go\|queue.go" docs/workflow/`
3. Verify all internal doc links resolve (check Quick Links table)
4. Verify new docs are linked from README.md

### Content Accuracy Checks

For each doc, verify key claims against actual code:
- Architecture diagram matches `all.go` initialization order
- Key Files tables list files that actually exist post-Phase 13
- Code examples compile (use `go build` on any Go snippets)
- Configuration env vars match `main.go` struct tags

---

## Plan Review Recommendations (Grade: A-)

**Reviewed**: 2026-02-10

### 1. Add Payload Size Documentation to `temporal.md`

The plan's Task 5 (temporal.md) should include a "Payload Management" or "Performance Considerations" subsection covering:

- Temporal's 2MB payload limit
- `MaxResultValueSize` (50KB) truncation via `sanitizeResult()`
- How `MergedContext` accumulates results and the truncation strategy
- Reference: `temporal/models.go:sanitizeResult()`, tested in `temporal/workflow_payload_test.go`

### 2. Add Technical Accuracy Verification Checklist

The "Content Accuracy Checks" section should include specific code references to verify:

```markdown
### README.md
- [ ] Architecture diagram matches all.go initialization order (lines 437-497)
- [ ] Event flow steps 1-7 match delegatehandler.go → trigger.go flow
- [ ] Key Files table lists only files that exist post-Phase 13

### architecture.md
- [ ] No references to engine.go, queue.go, eventpublisher.go (deleted)
- [ ] NewWorkflowTrigger signature matches temporal/trigger.go:61
- [ ] EdgeStore interface matches temporal/stores/edgedb/edgedb.go

### temporal.md
- [ ] TaskQueue constant matches temporal/workflow.go ("ichor-workflow")
- [ ] Continue-As-New threshold (10,000) matches temporal/workflow.go:22
- [ ] Activity retry policies match activityOptions() in workflow.go
```

### 3. Improve Validation Criteria with Expected Outputs

Current grep validation is good but should specify expected results:

```bash
# Expected: matches ONLY in migration-from-rabbitmq.md (historical context)
# and event-infrastructure.md (one historical context line)
# Zero matches in: README.md, architecture.md, testing.md, temporal.md, worker-deployment.md
grep -r "EventPublisher\|QueueManager\|workflow.NewEngine" docs/workflow/
```

### 4. Content Depth Guidance for New Docs (Tasks 5-7)

Tasks 5-7 provide section headings but could benefit from more depth guidance:

- **temporal.md** (~250 lines): GraphExecutor section should be the deepest (~60 lines) covering determinism, edge types, convergence detection, and parallel patterns. Activity Types section should include a table with all 3 categories and their retry policies.
- **worker-deployment.md** (~150 lines): Configuration section should include a table mapping env vars to their purpose and defaults. Action handler registration should list all 5 core handlers and explain why `RegisterAll` isn't used yet.
- **migration-from-rabbitmq.md** (~200 lines): The mapping table (old concept → new concept) is the most valuable artifact — prioritize clarity here. Include ~14 rows covering all replaced components.

---

## Commands Reference

```bash
# Start this phase
/workflow-temporal-next

# Validate this phase
/workflow-temporal-validate 14

# Review plan before implementing
/workflow-temporal-plan-review 14

# Review code after implementing
/workflow-temporal-review 14
```
