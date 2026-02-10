# Universal Action Edge Enforcement

## Decision: Option B - Require Edges Universally

After analysis, we've decided to **require action edges for all workflow rules** that have actions. This eliminates the dual execution mode (linear vs graph) and provides:

- **Single mental model** - One way to represent workflow execution flow
- **Consistent visual representation** - All rules render correctly in the workflow editor
- **Simpler observability** - One query pattern for execution path analysis
- **Easier feature development** - No "what if no edges?" edge cases
- **Better enterprise scalability** - Consistent auditing and monitoring

## Decisions Made

1. **`execution_order` field:** Remove entirely - clean break from linear execution mode
2. **Validation layer:** App layer only (in `workflowsaveapp`) - keeps business layer flexible

---

## Phase 1: Validation Layer Changes

**Goal:** Enforce that rules with actions must have edges connecting them.

### Files to Modify

| File | Change |
|------|--------|
| [app/domain/workflow/workflowsaveapp/graph.go](app/domain/workflow/workflowsaveapp/graph.go) | Add check: if `len(actions) > 0 && len(edges) == 0` return error |

### Validation Logic

```
ValidateGraph() called
  → if len(actions) > 0 && len(edges) == 0: return error "rules with actions require edges"
  → if len(actions) == 0 && len(edges) == 0: return nil (trigger-only rule is valid)
  → validate start edge, cycles, reachability (existing logic)
```

---

## Phase 2: Remove `execution_order` Field

**Goal:** Remove the `execution_order` field entirely from the codebase.

### Database Migration

| File | Lines | Change |
|------|-------|--------|
| [business/sdk/migrate/sql/migrate.sql](business/sdk/migrate/sql/migrate.sql) | 1029 | Add new migration to DROP COLUMN `execution_order` from `workflow.rule_actions` |
| [business/sdk/migrate/sql/migrate.sql](business/sdk/migrate/sql/migrate.sql) | 1519-1533 | Update `workflow.rule_actions_view` to remove `execution_order` from SELECT and ORDER BY |

### Business Layer Models

| File | Lines | Change |
|------|-------|--------|
| [business/sdk/workflow/stores/workflowdb/models.go](business/sdk/workflow/stores/workflowdb/models.go) | 337 | Remove `ExecutionOrder` from `ruleAction` struct |
| [business/sdk/workflow/stores/workflowdb/models.go](business/sdk/workflow/stores/workflowdb/models.go) | 626 | Remove `ExecutionOrder` from `ruleActionView` struct |
| [business/sdk/workflow/order.go](business/sdk/workflow/order.go) | 20 | Remove `ActionOrderByExecutionOrder` constant |

### Database Store Operations

| File | Lines | Change |
|------|-------|--------|
| [business/sdk/workflow/stores/workflowdb/workflowdb.go](business/sdk/workflow/stores/workflowdb/workflowdb.go) | 654-668 | Remove from CREATE INSERT |
| [business/sdk/workflow/stores/workflowdb/workflowdb.go](business/sdk/workflow/stores/workflowdb/workflowdb.go) | 672-691 | Remove from UPDATE SET |
| [business/sdk/workflow/stores/workflowdb/workflowdb.go](business/sdk/workflow/stores/workflowdb/workflowdb.go) | 1140-1160 | Remove from SELECT and ORDER BY |
| [business/sdk/workflow/stores/workflowdb/workflowdb.go](business/sdk/workflow/stores/workflowdb/workflowdb.go) | 1196-1209 | Remove from view query JSON aggregate |
| [business/sdk/workflow/stores/workflowdb/workflowdb.go](business/sdk/workflow/stores/workflowdb/workflowdb.go) | 1275, 1301 | Remove from additional queries |

### API Layer Models

| File | Lines | Change |
|------|-------|--------|
| [api/domain/http/workflow/ruleapi/action_model.go](api/domain/http/workflow/ruleapi/action_model.go) | 20 | Remove from `CreateActionRequest` |
| [api/domain/http/workflow/ruleapi/action_model.go](api/domain/http/workflow/ruleapi/action_model.go) | 35 | Remove from `UpdateActionRequest` |
| [api/domain/http/workflow/ruleapi/action_model.go](api/domain/http/workflow/ruleapi/action_model.go) | 108-113 | Remove validation in `ValidateCreateAction()` |
| [api/domain/http/workflow/ruleapi/action_model.go](api/domain/http/workflow/ruleapi/action_model.go) | 144-149 | Remove validation in `ValidateUpdateAction()` |
| [api/domain/http/workflow/ruleapi/model.go](api/domain/http/workflow/ruleapi/model.go) | 50 | Remove from `CreateActionInput` |
| [api/domain/http/workflow/ruleapi/model.go](api/domain/http/workflow/ruleapi/model.go) | 123 | Remove from `ActionResponse` |
| [api/domain/http/workflow/ruleapi/validation.go](api/domain/http/workflow/ruleapi/validation.go) | 86-91 | Remove validation in `ValidateCreateRule()` |
| [api/domain/http/workflow/ruleapi/ruleapi.go](api/domain/http/workflow/ruleapi/ruleapi.go) | 585-597 | Remove duplicate execution_order detection warning |

### App Layer Models

| File | Lines | Change |
|------|-------|--------|
| [app/domain/workflow/workflowsaveapp/model.go](app/domain/workflow/workflowsaveapp/model.go) | 46 | Remove from `SaveActionRequest` |
| [app/domain/workflow/workflowsaveapp/model.go](app/domain/workflow/workflowsaveapp/model.go) | 89 | Remove from `SaveActionResponse` |

---

## Phase 3: Remove Linear Executor

**Goal:** Remove the linear execution fallback path, keep only graph execution.

### Executor Changes

| File | Lines | Change |
|------|-------|--------|
| [business/sdk/workflow/executor.go](business/sdk/workflow/executor.go) | 112-209 | DELETE `ExecuteRuleActions()` function entirely |
| [business/sdk/workflow/executor.go](business/sdk/workflow/executor.go) | 222-227 | CHANGE fallback to return error instead of calling linear executor |
| [business/sdk/workflow/executor.go](business/sdk/workflow/executor.go) | 264-268 | CHANGE fallback to return error instead of calling linear executor |

### Engine Changes

| File | Lines | Change |
|------|-------|--------|
| [business/sdk/workflow/engine.go](business/sdk/workflow/engine.go) | 422 | CHANGE from `ExecuteRuleActions()` to `ExecuteRuleActionsGraph()` |

### New Error Handling

Replace fallback logic with error:
```go
if len(edges) == 0 {
    return BatchExecutionResult{}, fmt.Errorf("rule %s has no edges - all rules with actions require edges", ruleID)
}
```

---

## Phase 4: Test Updates

**Goal:** Remove obsolete tests, update others to include edges.

### Tests to DELETE (linear fallback tests)

| File | Lines | Test Name |
|------|-------|-----------|
| [business/sdk/workflow/executor_graph_test.go](business/sdk/workflow/executor_graph_test.go) | 246-293 | `TestGraphExec_NoEdges_FallsBackToLinear` |
| [business/sdk/workflow/executor_graph_test.go](business/sdk/workflow/executor_graph_test.go) | 295-328 | `TestGraphExec_EmptyEdges_FallsBackToLinear` |
| [business/sdk/workflow/executor_graph_test.go](business/sdk/workflow/executor_graph_test.go) | 421-459 | `TestGraphExec_NoStartEdge_FallsBackToLinear` |

### Tests to MODIFY (use graph executor)

| File | Lines | Test Name | Change |
|------|-------|-----------|--------|
| [business/sdk/workflow/executor_test.go](business/sdk/workflow/executor_test.go) | 765-977 | `TestActionExecutor_Stats` | Update to use `ExecuteRuleActionsGraph()` and include edges |
| [business/sdk/workflow/executor_test.go](business/sdk/workflow/executor_test.go) | 979-1166 | `TestActionExecutor_ExecutionHistory` | Update to use `ExecuteRuleActionsGraph()` and include edges |

### Tests to ADD

| File | Test Name | Purpose |
|------|-----------|---------|
| [api/cmd/services/ichor/tests/workflow/workflowsaveapi/validation_test.go](api/cmd/services/ichor/tests/workflow/workflowsaveapi/validation_test.go) | `Test_ActionsWithoutEdges_Fails` | Verify validation rejects actions without edges |

---

## Phase 5: Seed Data Updates

**Goal:** All seeded workflow rules must have proper edges.

### Seed Functions MISSING Edge Creation

| File | Function | Lines | Actions Created | Fix Required |
|------|----------|-------|-----------------|--------------|
| [business/sdk/workflow/testutil.go](business/sdk/workflow/testutil.go) | `TestSeedRuleActions()` | 229-243 | Creates actions | Add edge creation after each action |
| [business/sdk/workflow/testutil.go](business/sdk/workflow/testutil.go) | `TestSeedFullWorkflow()` | 401-511 | Seeds 10 actions | Add edge creation |
| [api/cmd/services/ichor/tests/workflow/ruleapi/seed_test.go](api/cmd/services/ichor/tests/workflow/ruleapi/seed_test.go) | `insertSeedData()` | 29-143 | 5 actions | Add edge creation |
| [api/cmd/services/ichor/tests/workflow/ruleapi/cascade_seed_test.go](api/cmd/services/ichor/tests/workflow/ruleapi/cascade_seed_test.go) | `insertCascadeSeedData()` | 58-387 | 8 actions across 6 rules | Add edge creation |
| [api/cmd/services/ichor/tests/workflow/executionapi/seed_test.go](api/cmd/services/ichor/tests/workflow/executionapi/seed_test.go) | `insertSeedData()` | 32-146 | 3 actions | Add edge creation |

### Seed Functions Already Creating Edges (Good Examples)

| File | Function | Pattern to Follow |
|------|----------|-------------------|
| [api/cmd/services/ichor/tests/workflow/edgeapi/seed_test.go](api/cmd/services/ichor/tests/workflow/edgeapi/seed_test.go) | `insertSeedData()` | Uses `CreateActionEdge()` after creating actions |
| [api/cmd/services/ichor/tests/workflow/workflowsaveapi/seed_test.go](api/cmd/services/ichor/tests/workflow/workflowsaveapi/seed_test.go) | `seedEdgesForRule()` | Helper function for edge creation |
| [api/cmd/services/ichor/tests/workflow/workflowsaveapi/execution_seed_test.go](api/cmd/services/ichor/tests/workflow/workflowsaveapi/execution_seed_test.go) | `createSimpleWorkflowDirect()` | Creates edges inline |

### Test Files with Inline Rule Creation (Need Review)

| File | Lines | Notes |
|------|-------|-------|
| [api/cmd/services/ichor/tests/sales/ordersapi/workflow_test.go](api/cmd/services/ichor/tests/sales/ordersapi/workflow_test.go) | 87-163 | 3 rules created inline |
| [api/cmd/services/ichor/tests/formdata/formdataapi/workflow_test.go](api/cmd/services/ichor/tests/formdata/formdataapi/workflow_test.go) | 49-181 | 2 rules created inline |
| [api/cmd/services/ichor/tests/workflow/workflowsaveapi/actions_test.go](api/cmd/services/ichor/tests/workflow/workflowsaveapi/actions_test.go) | 70-1062 | 10+ rules created inline |
| [api/cmd/services/ichor/tests/workflow/workflowsaveapi/errors_test.go](api/cmd/services/ichor/tests/workflow/workflowsaveapi/errors_test.go) | 70-752 | Multiple rules for error tests |
| [api/cmd/services/ichor/tests/workflow/workflowsaveapi/trigger_test.go](api/cmd/services/ichor/tests/workflow/workflowsaveapi/trigger_test.go) | 151-444 | 3 rules in trigger tests |

---

## Phase 6: Documentation Updates

**Goal:** Remove references to linear execution mode, document edge requirement.

### Files to Update

| File | Sections | Change |
|------|----------|--------|
| [docs/workflow/README.md](docs/workflow/README.md) | Line 89+ | Remove "linear execution mode" description |
| [docs/workflow/branching.md](docs/workflow/branching.md) | Lines 7, 14, 94-108 | Remove backwards compatibility sections |
| [docs/workflow/database-schema.md](docs/workflow/database-schema.md) | Lines 117, 171, 316 | Remove `execution_order` references |
| [docs/workflow/architecture.md](docs/workflow/architecture.md) | Lines 470-512 | Remove parallel execution and backwards compat sections |
| [docs/workflow/configuration/rules.md](docs/workflow/configuration/rules.md) | Lines 112, 121, 154, 425, 447, 460 | Remove `execution_order` from examples |
| [docs/workflow/actions/overview.md](docs/workflow/actions/overview.md) | Lines 182, 194, 207 | Update action ordering documentation |
| [docs/workflow/api-reference.md](docs/workflow/api-reference.md) | Line 625 | Remove from API reference |

---

## Verification Plan

After implementation, verify:

1. **Run tests**: `make test` - all workflow tests pass
2. **Reseed database**: `make dev-database-recreate && make seed`
3. **Visual editor**: Open any seeded rule in `/workflow/editor/:id` - edges should render
4. **Validation**: Try to save a rule with actions but no edges via API - should fail
5. **Execution**: Trigger a workflow rule - should execute via graph traversal

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
