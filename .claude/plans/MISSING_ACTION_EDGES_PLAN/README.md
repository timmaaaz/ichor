# Universal Action Edge Enforcement Implementation Plan

**Project**: Ichor ERP System
**Goal**: Require action edges for all workflow rules that have actions, eliminating the dual execution mode (linear vs graph)
**Timeline**: 8 phases
**Status**: Planning Complete - Ready for Phase Execution

---

## Table of Contents

1. [Executive Summary](#executive-summary)
2. [Decision](#decision)
3. [Implementation Strategy](#implementation-strategy)
4. [Phase Overview](#phase-overview)
5. [Technology Stack](#technology-stack)
6. [Success Criteria](#success-criteria)
7. [Quick Reference Guide](#quick-reference-guide)

---

## Executive Summary

### The Goal

Require action edges for all workflow rules that have actions. This eliminates the dual execution mode (linear vs graph) and simplifies the workflow engine.

### The Approach

1. Add validation to require edges when actions exist
2. Remove the `execution_order` field entirely
3. Remove the linear executor fallback
4. Update all tests to use graph execution
5. Fix seed data to include edges
6. Update documentation

### Why This Matters

- **Single mental model** - One way to represent workflow execution flow
- **Consistent visual representation** - All rules render correctly in the workflow editor
- **Simpler observability** - One query pattern for execution path analysis
- **Easier feature development** - No "what if no edges?" edge cases
- **Better enterprise scalability** - Consistent auditing and monitoring

---

## Decision

**Option B - Require Edges Universally**

Key decisions made:
1. **`execution_order` field:** Remove entirely - clean break from linear execution mode
2. **Validation layer:** App layer only (in `workflowsaveapp`) - keeps business layer flexible

---

## Implementation Strategy

### Approach

Sequential implementation through 8 phases, starting with validation and ending with documentation updates.

### Key Principles

1. **Backwards Compatibility During Transition** - Add validation first, then remove old code
2. **Test-Driven Changes** - Update tests before removing functionality
3. **Comprehensive Coverage** - All seed data and test fixtures updated

### Phase Dependencies

- Phase 1 (Validation) is independent
- Phases 2-3 can run in parallel after Phase 1
- Phase 4 (Tests) depends on Phases 2-3
- Phase 5 (Seeds) can run in parallel with Phase 4
- Phase 6 (Docs) runs last

---

## Phase Overview

| Phase | Name | Category | Description | Key Deliverables |
|-------|------|----------|-------------|------------------|
| 1 | Validation Layer Changes | backend | Enforce that rules with actions must have edges | Validation in workflowsaveapp |
| 2 | Remove execution_order Field | backend | Remove field from entire codebase | Migration, model changes |
| 3 | Remove Linear Executor | backend | Remove linear execution fallback | executor.go, engine.go changes |
| 4 | Test Updates | testing | Remove obsolete tests, update others | Test file modifications |
| 5 | Seed Data Updates | backend | All seeded rules must have proper edges | Seed function updates |
| 6 | Documentation Updates | documentation | Remove linear execution mode references | Doc file updates |

---

## Technology Stack

### Backend
- **Language/Runtime**: Go 1.23
- **Framework**: Ardan Labs Service Architecture
- **Database**: PostgreSQL 16.4
- **Additional**: sqlc, pgx

### Frontend
- **Framework**: Vue 3
- **Language**: TypeScript
- **UI Library**: shadcn-vue
- **State Management**: Pinia

### Tools
- **Development**: make, KIND (Kubernetes)
- **Testing**: go test, integration tests
- **Build**: Docker, Kubernetes

---

## Success Criteria

### Functional Requirements
- All rules with actions require edges (validation enforced)
- `execution_order` field completely removed
- Linear executor removed, only graph execution remains

### Non-Functional Requirements
- All existing tests pass (with updates)
- No performance regression
- Backwards compatibility for existing data (migration handles it)

### Quality Metrics
- **Code Quality**: All lint checks pass
- **Testing**: `make test` passes
- **Performance**: No execution time regression

---

## Quick Reference Guide

### Commands

Execute phases:
```bash
/missing-action-edges-status      # View current progress
/missing-action-edges-next        # Execute next pending phase
/missing-action-edges-phase N     # Jump to specific phase
/missing-action-edges-validate    # Run validation checks
```

Code review:
```bash
/missing-action-edges-review N    # Manual code review for phase N
```

Planning:
```bash
/missing-action-edges-build-phase # Generate next phase documentation
/missing-action-edges-summary     # Generate executive summary
/missing-action-edges-dependencies # Show cross-plan dependencies
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

1. [Phase 1: Validation Layer Changes](./phases/PHASE_1_VALIDATION_LAYER.md)
2. [Phase 2: Remove execution_order Field](./phases/PHASE_2_REMOVE_EXECUTION_ORDER.md)
3. [Phase 3: Remove Linear Executor](./phases/PHASE_3_REMOVE_LINEAR_EXECUTOR.md)
4. [Phase 4: Test Updates](./phases/PHASE_4_TEST_UPDATES.md)
5. [Phase 5: Seed Data Updates](./phases/PHASE_5_SEED_DATA_UPDATES.md)
6. [Phase 6: Documentation Updates](./phases/PHASE_6_DOCUMENTATION_UPDATES.md)

---

## Related Documentation

- [Workflow Engine Documentation](../../../docs/workflow/README.md)
- [Branching Documentation](../../../docs/workflow/branching.md)
- [Database Schema](../../../docs/workflow/database-schema.md)

---

## Summary Statistics

| Category | Files | Changes |
|----------|-------|---------|
| Validation | 1 | Add edge requirement check |
| Remove execution_order | ~15 | Models, queries, validation, migrations |
| Remove linear executor | 2 | executor.go, engine.go |
| Tests to delete | 1 | 3 test functions |
| Tests to modify | 1 | 2 test functions |
| Seeds to fix | 5 | Add edge creation |
| Tests to review | 5 | Inline rule creation |
| Documentation | 7 | Remove linear mode references |

**Total: ~30 files affected**

---

## Notes

### Assumptions
- Existing rules in production already use edges (or will be migrated)
- No external systems depend on linear execution behavior

### Constraints
- Must maintain API compatibility for create/update operations
- Migration must handle existing data gracefully

### Future Enhancements
- Consider adding edge validation warnings in the UI editor
- Add execution path visualization in workflow monitor

---

**Last Updated**: 2026-02-05
**Created By**: Claude Code
**Status**: Planning
