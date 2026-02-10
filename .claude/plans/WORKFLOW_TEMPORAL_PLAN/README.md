# Workflow Temporal Implementation Plan

**Project**: Ichor ERP System
**Goal**: Implement Temporal-based workflow engine that interprets visual workflow graphs with full durability, parallel branch support, and async continuation
**Timeline**: 15 phases
**Status**: Planning Complete - Ready for Phase Execution

---

## Table of Contents

1. [Executive Summary](#executive-summary)
2. [Current Architecture](#current-architecture)
3. [Target Architecture](#target-architecture)
4. [Implementation Strategy](#implementation-strategy)
5. [Phase Overview](#phase-overview)
6. [Technology Stack](#technology-stack)
7. [Success Criteria](#success-criteria)
8. [Quick Reference Guide](#quick-reference-guide)

---

## Executive Summary

### The Goal

Implement a production-grade workflow engine using Temporal that:
- Interprets visual workflow graphs (nodes + edges stored in database)
- Provides full durability and crash recovery
- Supports parallel branch execution with convergence detection
- Handles async continuation for long-running operations (RabbitMQ pattern)
- Wraps existing action handlers with minimal modification

### The Approach

Use Temporal from the start rather than building custom continuation infrastructure:
- **Durability**: Temporal handles crash recovery and replay automatically
- **Visibility**: Built-in UI for debugging workflow executions
- **Versioning**: Safe deployments with in-flight workflows
- **Battle-tested**: Used at scale by Uber, Netflix, Stripe

### Why This Matters

The visual workflow editor produces dynamic graph definitions at runtime. Users draw workflows as nodes and edges, which are stored in the database. This implementation interprets those graphs while preserving:
- Parallel branches run concurrently
- Nodes with multiple incoming edges wait for all branches (convergence)
- Action results merge into execution context for template variable resolution
- Edge types: `start`, `sequence`, `true_branch`, `false_branch`, `always`, `parallel`

---

## Current Architecture

### Backend

**Current Workflow Structure**:
```
business/domain/workflow/
├── automationrulesbus/     # Automation rules business logic
├── ruleactionsbus/         # Rule actions business logic
└── actionedgesbus/         # Action edges (graph connectivity)

business/sdk/workflow/
└── workflowactions/        # Action handlers (control, inventory, notification, persistence)
```

**Key Components**:
- Visual editor → `rule_actions` + `action_edges` → PostgreSQL
- Entity events fire workflow triggers
- Action handlers execute business logic

**Current Patterns**:
- Event-driven rule matching
- Sequential action execution
- Template variable resolution (`{{action_name.field}}`)

### Database

**Current Schema** (workflow.* tables):
```sql
workflow.automation_rules   -- Trigger definitions
workflow.rule_actions       -- Actions (nodes in graph)
workflow.action_edges       -- Edges connecting actions
```

---

## Target Architecture

### What Changes

- **New**: Temporal server for workflow orchestration
- **New**: Workflow worker service (separate from main API)
- **New**: Graph executor for interpreting visual graphs
- **Modified**: Entity events dispatch to Temporal instead of direct execution
- **Wrapped**: Existing action handlers become Temporal activities

### What Stays the Same

- Visual workflow editor and graph storage in PostgreSQL
- Action handler business logic
- Template variable syntax (`{{action_name.field}}`)
- Edge types and graph semantics

### New Components/Systems

```
┌─────────────────────────────────────────────────────────────────┐
│                        Ichor Codebase                           │
├─────────────────────────────────────────────────────────────────┤
│  Visual Editor → rule_actions + action_edges → Postgres        │
│                                    │                            │
│                              Graph Config                       │
│                                    │                            │
│  Entity Events ──────────────────► WorkflowTrigger              │
└────────────────────────────────────┼────────────────────────────┘
                                     │
                                     ▼
┌─────────────────────────────────────────────────────────────────┐
│                      Temporal Server                            │
├─────────────────────────────────────────────────────────────────┤
│  ExecuteGraphWorkflow ──► Activities (existing handlers)       │
│         │                            │                          │
│    Workflow State              Activity Results                 │
│    (Temporal owns)             (merged into context)            │
└─────────────────────────────────────────────────────────────────┘
```

---

## Implementation Strategy

### Approach

Build incrementally with clear phase boundaries:
1. Infrastructure first (Temporal cluster, worker skeleton)
2. Evaluate existing libraries before custom implementation
3. Core logic in isolation (models, graph executor, workflow)
4. Integration layer (triggers, database adapters)
5. Comprehensive testing (determinism is critical!)

### Key Principles

1. **Determinism is Non-Negotiable** - Temporal workflows must produce identical commands on replay. All map iterations sorted, no `time.Now()` or `rand` in workflow code.
2. **Minimal Handler Changes** - Existing action handlers become activities with thin wrapper. Business logic stays intact.
3. **Single-Prompt Phases** - Each phase scoped to complete in one focused session, reducing context loss and errors.

### Phase Dependencies

**CRITICAL**: Phase 1 (Infrastructure) must complete first. We cannot test ANY code without `make dev-bounce` working with Temporal and workflow-worker.

```
┌─────────────────────────────────────────────────────────────────────┐
│                    Phase 1 (Infrastructure)                         │
│             FOUNDATIONAL - Blocks ALL other phases                  │
│        (Temporal + workflow-worker integrated into dev-bounce)      │
└────────────────────────────────┬────────────────────────────────────┘
                                 │
         ┌───────────────────────┴───────────────────────┐
         │                                               │
   (can run in parallel - research only)                 │
         │                                               │
  ┌──────▼──────┐                                        │
  │   Phase 2   │                                        │
  │ Evaluation  │────────────────────┐                   │
  │ (research)  │                    │                   │
  └─────────────┘                    │                   │
                                     ▼                   │
                              ┌─────────────┐            │
                              │   Phase 3   │◄───────────┘
                              │Core Models  │
                              └──────┬──────┘
                                     │
                              ┌──────▼──────┐
                              │   Phase 4   │────────────────────┐
                              │Graph Exec.  │                    │
                              └──────┬──────┘                    │
                                     │                    ┌──────▼──────┐
                              ┌──────▼──────┐             │  Phase 10   │
                              │   Phase 5   │             │ Unit Tests  │
                              │  Workflow   │             └──────┬──────┘
                              └──────┬──────┘                    │
                                     │                           │
                              ┌──────▼──────┐                    │
                              │   Phase 6   │                    │
                              │ Activities  │                    │
                              └──────┬──────┘                    │
                                     │                           │
                         ┌───────────┴───────────┐               │
                         │                       │               │
                  ┌──────▼──────┐         ┌──────▼──────┐        │
                  │   Phase 7   │         │   Phase 8   │        │
                  │   Trigger   │         │ Edge Store  │        │
                  └──────┬──────┘         └──────┬──────┘        │
                         │                       │               │
                         └───────────┬───────────┘               │
                                     │                           │
                              ┌──────▼──────┐                    │
                              │   Phase 9   │                    │
                              │Worker Wire  │                    │
                              └──────┬──────┘                    │
                                     │                           │
                         ┌───────────┼───────────┐               │
                         │           │           │               │
                  ┌──────▼──────┐    │    ┌──────▼──────┐        │
                  │  Phase 11   │    │    │  Phase 12   │        │
                  │ Int. Tests  │    │    │ Edge Cases  │        │
                  └──────┬──────┘    │    └──────┬──────┘        │
                         │           │           │               │
                         └───────────┼───────────┘               │
                                     │                           │
                                     └───────────────────────────┘
                                     │
                              ┌──────▼──────┐
                              │  Phase 13   │
                              │ Dead Code   │
                              │ Removal &   │
                              │Temporal Wire│
                              └──────┬──────┘
                                     │
                              ┌──────▼──────┐
                              │  Phase 14   │
                              │    Docs     │
                              └──────┬──────┘
                                     │
                              ┌──────▼──────┐
                              │  Phase 15   │
                              │Integration  │
                              │Verification │
                              └─────────────┘
```

---

## Phase Overview

| Phase | Name | Category | Description | Key Deliverables |
|-------|------|----------|-------------|------------------|
| 1 | **Infrastructure Setup** | infrastructure | **FOUNDATIONAL**: Temporal + workflow-worker in dev-bounce + test container infrastructure. MUST complete first. | K8s manifests, Dockerfile, Makefile targets, `foundation/temporal/temporal.go` test container |
| 2 | Temporalgraph Evaluation | research | Evaluate temporalgraph library for compatibility | Decision document: use library or custom |
| 3 | Core Models & Context | backend | WorkflowInput, GraphDefinition, MergedContext models | `temporal/models.go` |
| 4 | Graph Executor | backend | Graph traversal with deterministic iteration | `temporal/graph_executor.go`, unit tests |
| 5 | Workflow Implementation | backend | Main workflow with Continue-As-New, versioning | `temporal/workflow.go` |
| 6 | Activities & Async | backend | Activity wrappers, async completion pattern | `temporal/activities.go`, `activities_async.go`, `async_completer.go` |
| 7 | Trigger System | backend | Entity event to workflow dispatch | `temporal/trigger.go` |
| 8 | Edge Store Adapter | backend | Database adapter for loading graphs | `temporal/stores/edgedb/edgedb.go` |
| 9 | Worker Service & Wiring | backend | Worker entry point, action registry wiring | `workflow-worker/main.go`, `all/all.go` modifications |
| 10 | Graph Executor Unit Tests | testing | Determinism tests, convergence consistency | Test files for graph executor |
| 11 | Workflow Integration Tests | testing | End-to-end workflow tests, replay testing | Integration test suite |
| 12 | Edge Case & Limit Tests | testing | Continue-As-New, context truncation, mock RabbitMQ | Edge case test coverage |
| 13 | Dead Code Removal & Temporal Rewiring | backend | Remove old engine, wire Temporal as sole path | TemporalDelegateHandler, rewired `all.go`, ~4600 lines removed |
| 14 | Documentation Updates | backend | Update workflow docs, create Temporal docs | Updated + new docs in `docs/workflow/` |
| 15 | Integration Verification | testing | Rewrite test infra, verify end-to-end | Rewritten `apitest/workflow.go`, health checks |

---

## Technology Stack

### Backend
- **Language/Runtime**: Go 1.23
- **Framework**: Ardan Labs Service Architecture
- **Database**: PostgreSQL 16.4 (workflow.* schema)
- **Workflow Engine**: Temporal 1.22+
- **Message Queue**: RabbitMQ (for async activity completion)

### Infrastructure
- **Containers**: Docker, KIND (local K8s)
- **Orchestration**: Kubernetes
- **Temporal UI**: temporalio/ui:2.21

### Tools
- **Development**: Make, Docker Compose
- **Testing**: Go test, Temporal test framework
- **Build**: Docker multi-stage builds

---

## Success Criteria

### Functional Requirements
- ✅ Workflow graphs execute with full durability (crash recovery via Temporal)
- ✅ Parallel branches execute concurrently with convergence detection
- ✅ Async actions complete via RabbitMQ callback pattern
- ✅ Template variables resolve from merged action context
- ✅ All edge types respected: start, sequence, true_branch, false_branch, always, parallel

### Non-Functional Requirements
- ✅ Deterministic execution (replay produces identical commands)
- ✅ History limits handled via Continue-As-New (< 50K events)
- ✅ Payload limits handled via result truncation (< 2MB)
- ✅ Worker scales independently from main API service

### Quality Metrics
- **Code Quality**: Follows Ardan Labs patterns, idiomatic Go
- **Testing**: Determinism tests pass, replay tests pass, integration tests cover all edge types
- **Performance**: Workflows handle 100+ concurrent actions per worker

---

## Quick Reference Guide

### Commands

Execute phases:
```bash
/workflow-temporal-status      # View current progress
/workflow-temporal-next        # Execute next pending phase
/workflow-temporal-phase N     # Jump to specific phase
/workflow-temporal-validate    # Run validation checks
```

Code review:
```bash
/workflow-temporal-review      # Code review AFTER implementation
/workflow-temporal-plan-review # Plan review BEFORE implementation
```

Planning:
```bash
/workflow-temporal-build-phase # Generate next phase documentation
/workflow-temporal-summary     # Generate executive summary
/workflow-temporal-dependencies # Show cross-plan dependencies
/workflow-temporal-quick-status # Compact phase overview table
```

### Phase Documentation

Each phase has detailed documentation in `phases/PHASE_N_NAME.md` with:
- Overview and goals
- Task breakdown
- Implementation guide with code examples
- Validation criteria
- Testing strategy

### PROGRESS.yaml

Track progress in [PROGRESS.yaml](./PROGRESS.yaml):
- Real-time phase and task status
- Context (current focus, next task)
- Blockers and decisions
- Deliverables tracking

---

## Phase Documentation

1. [Phase 1: Infrastructure Setup](./phases/PHASE_1_INFRASTRUCTURE_SETUP.md)
2. [Phase 2: Temporalgraph Evaluation](./phases/PHASE_2_TEMPORALGRAPH_EVALUATION.md)
3. [Phase 3: Core Models & Context](./phases/PHASE_3_CORE_MODELS.md)
4. [Phase 4: Graph Executor](./phases/PHASE_4_GRAPH_EXECUTOR.md)
5. [Phase 5: Workflow Implementation](./phases/PHASE_5_WORKFLOW.md)
6. [Phase 6: Activities & Async](./phases/PHASE_6_ACTIVITIES.md)
7. [Phase 7: Trigger System](./phases/PHASE_7_TRIGGER.md)
8. [Phase 8: Edge Store Adapter](./phases/PHASE_8_EDGE_STORE.md)
9. [Phase 9: Worker Service & Wiring](./phases/PHASE_9_WORKER_WIRING.md)
10. [Phase 10: Graph Executor Unit Tests](./phases/PHASE_10_UNIT_TESTS.md)
11. [Phase 11: Workflow Integration Tests](./phases/PHASE_11_INTEGRATION_TESTS.md)
12. [Phase 12: Edge Case & Limit Tests](./phases/PHASE_12_EDGE_CASE_TESTS.md)
13. [Phase 13: Dead Code Removal & Temporal Rewiring](./phases/PHASE_13_DEAD_CODE_REMOVAL.md)
14. [Phase 14: Documentation Updates](./phases/PHASE_14_DOCUMENTATION_UPDATES.md)
15. [Phase 15: Integration Verification](./phases/PHASE_15_INTEGRATION_VERIFICATION.md)

---

## Related Documentation

- [Workflow Temporal Implementation Design](./.claude/plans/workflow-temporal-implementation.md)
- [Workflow Engine Documentation](docs/workflow/README.md)
- [Ardan Labs Service Architecture](https://github.com/ardanlabs/service/wiki)

---

## Notes

### Assumptions
- Temporal server will be deployed separately (not embedded)
- RabbitMQ is available for async activity completion
- Existing action handlers are functional and tested

### Constraints
- Temporal determinism requirements (no maps iteration without sorting, no time.Now())
- 2MB payload limit per Temporal message
- 50K event limit per workflow history

### Testing Infrastructure Requirement

**CRITICAL**: Temporal must be buildable for testing in the same way that RabbitMQ is, enabling both integration and unit tests.

The project follows a test container pattern established in `foundation/rabbitmq/rabbitmq.go`:
- `StartTemporal()` - starts a Temporal dev server container
- `GetTestContainer(t *testing.T)` - returns a shared container for tests (singleton pattern)
- `NewTestClient(url string)` - creates a test client bypassing singleton for test isolation
- Container struct with connection URL

This infrastructure must be created in **Phase 1** alongside K8s manifests to enable:
1. Unit tests for graph executor and models (can use Temporal test framework)
2. Integration tests for workflow execution (require running Temporal)
3. Replay tests for determinism verification (require Temporal history)
4. End-to-end tests with real workflow triggers

See `foundation/rabbitmq/rabbitmq.go` for the reference implementation pattern.

### Future Enhancements (Separate Plan)
- Production Temporal configuration
- Monitoring dashboards
- Operational runbooks
- Multi-tenancy support

---

**Last Updated**: 2026-02-09
**Created By**: Claude Code
**Status**: Planning Complete
