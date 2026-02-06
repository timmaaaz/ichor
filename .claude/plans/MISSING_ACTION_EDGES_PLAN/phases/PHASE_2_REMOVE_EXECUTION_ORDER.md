# Phase 2: Remove execution_order Field

**Category**: Backend
**Status**: Pending
**Dependencies**: Phase 1 (Validation Layer Changes) - must be completed first

---

## Overview

Remove the `execution_order` field entirely from the workflow `rule_actions` table and all code that references it. With universal edge enforcement (Phase 1), execution order is determined by graph traversal of edges, making the `execution_order` column redundant and potentially misleading.

The `execution_order` field currently exists in the database schema, business layer models, database store operations, API layer models, app layer models, validation logic, and test utilities. This phase systematically removes it from all layers.

### Goals

1. Remove the `execution_order` column from the `workflow.rule_actions` CREATE TABLE definition
2. Remove `execution_order` from the `rule_actions_view` and from the inline JSON aggregation in `QueryAutomationRulesViewPaginated`
3. Remove `execution_order` from all Go structs, queries, conversions, and validations across every layer

### Why This Phase Matters

The `execution_order` field represents the old linear execution model. Keeping it alongside the new edge-based model creates confusion: developers might set `execution_order` thinking it controls execution order, when in reality edges determine the order. Removing it entirely enforces the "one way to do things" principle and prevents subtle bugs where linear and graph ordering disagree.

### order.By Evaluation

The `order.By` pattern in this codebase is used for user-facing paginated list endpoints (e.g., `QueryAutomationRulesViewPaginated` takes `order.By`). Action query methods are all internal lookups, not paginated endpoints:

- `QueryActionsByRule` - internal, called by `workflowsaveapp.go` for syncing
- `QueryRoleActionsViewByRuleID` - internal, called by `executor.go` for execution
- `QueryActionByID` / `QueryActionViewByID` - single-row lookups

Adding `order.By` to these methods would be unnecessary complexity. Instead, remove the hardcoded `ORDER BY execution_order ASC` from `QueryRoleActionsViewByRuleID` entirely (the graph executor determines order via edge traversal, not DB sort order).

### Transaction Infrastructure

No additional transaction work needed for this phase. The existing `workflowsaveapp.go` already uses proper transactions (`BeginTxx` + `NewWithTx` + deferred `Rollback` + `Commit` + post-commit delegate events). Since we are directly modifying the table definition in `migrate.sql` (not running ALTER TABLE at runtime), there is no migration transaction concern.

### Cache Invalidation

Not applicable. No `sturdyc` caching exists on `RuleAction` structs. The workflow domain uses direct DB access throughout.

---

## Prerequisites

Before starting this phase, ensure:

- [ ] Phase 1 (Validation Layer Changes) is completed
- [ ] Go development environment is ready (`go version` shows 1.23+)
- [ ] You can build the project: `go build ./...`

---

## Task Breakdown

### Task 1: Remove execution_order from Database Schema and Views

Since this project is not in production, directly modify the existing definitions rather than adding ALTER TABLE migrations.

**Files**:
- `business/sdk/migrate/sql/migrate.sql`

**Sub-task 1a: Remove column from CREATE TABLE**

Remove line 1029 from the `workflow.rule_actions` CREATE TABLE (version 1.69):

```sql
-- REMOVE this line:
execution_order INTEGER NOT NULL DEFAULT 1,
```

The table definition should go from:
```sql
CREATE TABLE workflow.rule_actions (
   id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
   automation_rules_id UUID NOT NULL REFERENCES workflow.automation_rules(id),
   name VARCHAR(100) NOT NULL,
   description TEXT,
   action_config JSONB NOT NULL,
   execution_order INTEGER NOT NULL DEFAULT 1,   -- DELETE THIS LINE
   is_active BOOLEAN DEFAULT TRUE,
   template_id UUID NULL REFERENCES workflow.action_templates(id),
   deactivated_by UUID NULL REFERENCES core.users(id)
);
```

To:
```sql
CREATE TABLE workflow.rule_actions (
   id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
   automation_rules_id UUID NOT NULL REFERENCES workflow.automation_rules(id),
   name VARCHAR(100) NOT NULL,
   description TEXT,
   action_config JSONB NOT NULL,
   is_active BOOLEAN DEFAULT TRUE,
   template_id UUID NULL REFERENCES workflow.action_templates(id),
   deactivated_by UUID NULL REFERENCES core.users(id)
);
```

**Sub-task 1b: Remove from rule_actions_view**

Remove `ra.execution_order,` from the `rule_actions_view` SELECT clause (line 1526):

```sql
-- BEFORE:
CREATE OR REPLACE VIEW workflow.rule_actions_view AS
   SELECT
      ra.id,
      ra.automation_rules_id,
      ra.name,
      ra.description,
      ra.action_config,
      ra.execution_order,        -- DELETE THIS LINE
      ra.is_active,
      ...

-- AFTER:
CREATE OR REPLACE VIEW workflow.rule_actions_view AS
   SELECT
      ra.id,
      ra.automation_rules_id,
      ra.name,
      ra.description,
      ra.action_config,
      ra.is_active,
      ...
```

**Note**: The `automation_rules_view` (lines 1475-1516) does NOT contain `execution_order`. The inline JSON aggregation that includes `execution_order` is in the Go code (`QueryAutomationRulesViewPaginated` in `workflowdb.go`), not in the SQL view definition. That is handled in Task 3.

**Validation**:
- [ ] `execution_order` no longer appears in the `CREATE TABLE workflow.rule_actions` statement
- [ ] `execution_order` no longer appears in the `rule_actions_view` definition
- [ ] No trailing or double commas after removal

---

### Task 2: Remove from Business Layer Models

**Files**:
- `business/sdk/workflow/models.go` - Remove `ExecutionOrder` from core structs
- `business/sdk/workflow/order.go` - Remove `ActionOrderByExecutionOrder` constant

**In `models.go`**, remove the `ExecutionOrder` field from all four structs:
1. `RuleAction` struct (line 341): Remove `ExecutionOrder int`
2. `NewRuleAction` struct (line 353): Remove `ExecutionOrder int`
3. `UpdateRuleAction` struct (line 363): Remove `ExecutionOrder *int`
4. `RuleActionView` struct (line 607): Remove `ExecutionOrder int`

**In `order.go`**, remove only this constant:
```go
// REMOVE this line:
ActionOrderByExecutionOrder = "execution_order"
```

**Do NOT remove** these unrelated constants (they are for `action_executions` table ordering):
- `ExecutionOrderByID`
- `ExecutionOrderByExecutedAt`
- `ExecutionOrderByStatus`
- `ExecutionOrderByRuleID`
- `ExecutionOrderByEntityType`
- `DefaultExecutionOrderBy`

**Validation**:
- [ ] No `ExecutionOrder` fields remain in `RuleAction`, `NewRuleAction`, `UpdateRuleAction`, `RuleActionView`
- [ ] `ActionOrderByExecutionOrder` constant is removed
- [ ] `ExecutionOrderBy*` constants for action_executions are preserved

---

### Task 3: Remove from Database Store Operations

**Files**:
- `business/sdk/workflow/stores/workflowdb/models.go` - Remove from DB structs and conversions
- `business/sdk/workflow/stores/workflowdb/workflowdb.go` - Remove from SQL queries

**In `models.go`** (5 changes):
1. Remove `ExecutionOrder int \`db:"execution_order"\`` from `ruleAction` struct (line 337)
2. Remove `ExecutionOrder: dbAction.ExecutionOrder` from `toCoreRuleAction()` (line 355)
3. Remove `ExecutionOrder: ra.ExecutionOrder` from `toDBRuleAction()` (line 388)
4. Remove `ExecutionOrder sql.NullInt32 \`db:"execution_order"\`` from `ruleActionView` struct (line 626)
5. Remove the entire `if dbView.ExecutionOrder.Valid { ... }` block from `toCoreRuleActionView()` (lines 653-655)

**In `workflowdb.go`** (7 changes):
1. **`CreateRuleAction`** (lines 656-661): Remove `execution_order` from both the INSERT column list and VALUES list. Check comma placement.
2. **`UpdateRuleAction`** (line 681): Remove `execution_order = :execution_order,` from the SET clause. Check comma placement.
3. **`QueryActionsByRule`** (line 741): Remove `execution_order,` from the SELECT column list.
4. **`QueryRoleActionsViewByRuleID`** (lines 1146-1153): Remove `execution_order,` from the SELECT column list AND remove the entire `ORDER BY execution_order ASC` clause. The graph executor determines order via edge traversal, so DB sort order is irrelevant.
5. **`QueryActionByID`** (line 1274): Remove `execution_order,` from the SELECT column list.
6. **`QueryActionViewByID`** (line 1300): Remove `execution_order,` from the SELECT column list.
7. **`QueryAutomationRulesViewPaginated`** (lines 1196-1209): In the inline JSON aggregation subquery:
   - Remove `'execution_order', ra.execution_order,` from the `json_build_object()` call (line 1202)
   - Change `ORDER BY ra.execution_order` to `ORDER BY ra.name` (line 1206) for alphabetical, deterministic ordering

**Note on `stores/workflowdb/order.go`**: This file contains only `automationRuleOrderByFields` and `executionOrderByFields` maps. Neither references `execution_order` from `rule_actions`. No changes needed in this file.

**Validation**:
- [ ] No references to `execution_order` in `workflowdb/models.go`
- [ ] No references to `execution_order` in `workflowdb/workflowdb.go` (except `ExecutionOrderBy*` for action_executions)
- [ ] SQL queries are syntactically correct after removal (check comma placement!)
- [ ] `QueryRoleActionsViewByRuleID` has no ORDER BY clause
- [ ] `QueryAutomationRulesViewPaginated` JSON agg uses `ORDER BY ra.name`

---

### Task 4: Remove from Business Layer Logic

**Files**:
- `business/sdk/workflow/workflowbus.go` - Remove from Create/Update logic

**2 changes**:

1. In `CreateRuleAction()` (line 636), remove:
```go
ExecutionOrder:   nra.ExecutionOrder,
```

2. In `UpdateRuleAction()` (lines 662-663), remove the entire block:
```go
if ura.ExecutionOrder != nil {
    action.ExecutionOrder = *ura.ExecutionOrder
}
```

**Validation**:
- [ ] No references to `ExecutionOrder` in `workflowbus.go`

---

### Task 5: Remove from API Layer

**Files**:
- `api/domain/http/workflow/ruleapi/action_model.go` - Structs and validation
- `api/domain/http/workflow/ruleapi/model.go` - Input/response structs
- `api/domain/http/workflow/ruleapi/validation.go` - Rule-level validation
- `api/domain/http/workflow/ruleapi/ruleapi.go` - Duplicate detection
- `api/domain/http/workflow/ruleapi/converters.go` - Conversion functions

**In `action_model.go`** (4 changes):
1. Remove `ExecutionOrder int \`json:"execution_order"\`` from `CreateActionRequest` (line 20)
2. Remove `ExecutionOrder *int \`json:"execution_order,omitempty"\`` from `UpdateActionRequest` (line 35)
3. Remove the `if req.ExecutionOrder < 0 { ... }` validation block from `ValidateCreateAction()` (lines 108-113)
4. Remove the `if req.ExecutionOrder != nil && *req.ExecutionOrder < 0 { ... }` validation block from `ValidateUpdateAction()` (lines 144-149)

**In `model.go`** (2 changes):
1. Remove `ExecutionOrder int \`json:"execution_order"\`` from `CreateActionInput` (line 50)
2. Remove `ExecutionOrder int \`json:"execution_order"\`` from `ActionResponse` (line 123)

**In `validation.go`** (1 change):
1. Remove the `if action.ExecutionOrder < 0 { ... }` validation block from `ValidateCreateRule()` (lines 86-91)

**In `ruleapi.go`** (1 change):
1. Remove the entire duplicate execution_order detection block from `validateRule()` (lines 585-597)

**In `converters.go`** (4 changes):
1. Remove `ExecutionOrder: input.ExecutionOrder,` from `toNewRuleAction()` (line 129)
2. Remove `ExecutionOrder: req.ExecutionOrder,` from `toNewRuleActionFromRequest()` (line 147)
3. Remove `ExecutionOrder: action.ExecutionOrder,` from `toActionResponse()` (line 63)
4. Remove `ExecutionOrder: req.ExecutionOrder,` from `toUpdateRuleAction()` (line 160)

**Validation**:
- [ ] No references to `ExecutionOrder` or `execution_order` in any file in the `ruleapi/` package
- [ ] Validation functions still work for remaining fields

---

### Task 6: Remove from App Layer

**Files**:
- `app/domain/workflow/workflowsaveapp/model.go` - Save request/response structs
- `app/domain/workflow/workflowsaveapp/workflowsaveapp.go` - Conversion logic

**In `model.go`** (2 changes):
1. Remove `ExecutionOrder int \`json:"execution_order" validate:"required,min=1"\`` from `SaveActionRequest` (line 46)
2. Remove `ExecutionOrder int \`json:"execution_order"\`` from `SaveActionResponse` (line 89)

**In `workflowsaveapp.go`** (3 changes):
1. In `syncActions()` (line 330), remove `ExecutionOrder: &reqAction.ExecutionOrder,` from the `UpdateRuleAction` struct
2. In `createAllActions()` (line 408), remove `ExecutionOrder: reqAction.ExecutionOrder,` from the `NewRuleAction` struct
3. In `buildResponse()` (line 539), remove `ExecutionOrder: action.ExecutionOrder,` from the `SaveActionResponse` struct

**Validation**:
- [ ] No references to `ExecutionOrder` in `workflowsaveapp/model.go`
- [ ] No references to `ExecutionOrder` in `workflowsaveapp/workflowsaveapp.go`

---

## Deliverables

- [ ] Modified `CREATE TABLE workflow.rule_actions` without `execution_order` column
- [ ] Modified `workflow.rule_actions_view` without `execution_order`
- [ ] Modified `QueryAutomationRulesViewPaginated` JSON aggregation: removed `execution_order`, changed ORDER BY to `ra.name`
- [ ] Cleaned business layer models (4 structs in `models.go`, 1 constant in `order.go`)
- [ ] Cleaned business layer logic (`workflowbus.go`)
- [ ] Cleaned database store models and queries (`workflowdb/models.go`, `workflowdb/workflowdb.go`)
- [ ] Cleaned API layer models, validation, handlers, and converters (`ruleapi/`)
- [ ] Cleaned app layer models and conversions (`workflowsaveapp/`)

---

## Validation Criteria

- [ ] Go compilation passes: `go build ./...`
- [ ] No references to `execution_order` remain in Go code (except `ExecutionOrderBy*` constants for action_executions)
- [ ] No references to `ExecutionOrder` struct field remain (except `ExecutionOrderBy*` constants)
- [ ] JSON API no longer includes `execution_order` in request/response bodies

---

## Testing Strategy

### What to Test

- **Compilation**: The primary test for this phase - all code must compile after field removal
- **Existing tests**: Many tests reference `ExecutionOrder` and will fail. This is expected and addressed in Phase 4 (Test Updates) and Phase 5 (Seed Data Updates)

### How to Test

```bash
# Build all packages to check for compilation errors
go build ./...

# Specifically build the affected packages
go build ./business/sdk/workflow/...
go build ./business/sdk/workflow/stores/workflowdb/...
go build ./api/domain/http/workflow/ruleapi/...
go build ./app/domain/workflow/workflowsaveapp/...

# Search for any remaining references (should return nothing except ExecutionOrderBy* and test files)
grep -rn "execution_order" business/sdk/workflow/ --include="*.go" | grep -v "_test.go" | grep -v "ExecutionOrderBy"
grep -rn "ExecutionOrder" business/sdk/workflow/ --include="*.go" | grep -v "_test.go" | grep -v "ExecutionOrderBy"
```

### Known Test Breakage

The following test files reference `ExecutionOrder` and will fail to compile after this phase. **This is expected** - they are fixed in Phase 4 and Phase 5:

- `business/sdk/workflow/testutil.go` (line 212)
- `business/sdk/workflow/executor_graph_test.go` (line 196)
- `business/sdk/workflow/executor_test.go` (lines 886, 898, 909, 1104)
- `business/sdk/workflow/workflow_crud_test.go` (lines 1132, 1142, 1178, 1191, 1226, 1235)
- `business/sdk/workflow/eventpublisher_integration_test.go` (lines 107, 297, 315, 417, 520, 736)
- `business/sdk/workflow/queue_test.go` (lines 420, 716)
- `business/sdk/workflow/workflowactions/data/updatefield_test.go` (line 207)

---

## Gotchas and Tips

- **Comma placement in SQL**: When removing `execution_order` from column lists, ensure you don't leave trailing or double commas. Check the column before and after the removed one.
- **Don't remove `ExecutionOrderBy*` constants**: The constants `ExecutionOrderByID`, `ExecutionOrderByExecutedAt`, `ExecutionOrderByStatus`, `ExecutionOrderByRuleID`, `ExecutionOrderByEntityType` (order.go lines 26-34) are for the `action_executions` table ordering. They have nothing to do with the `execution_order` field on `rule_actions`. **Leave them alone.**
- **Naming collision warning**: `ActionOrderByExecutionOrder` (REMOVE) vs `ExecutionOrderBy*` (KEEP) sound similar but are completely unrelated. The former is for rule_actions sorting, the latter is for action_executions history.
- **Don't remove `DefaultExecutionOrderBy`**: This is the default ordering for execution history queries (line 34 in order.go). Keep it.
- **Test files will break**: Don't try to fix test files in this phase. Phase 4 and 5 handle that. Focus on production code only.

---

## Reference

- Original plan: `.claude/plans/missing-action-edges.md`
- Progress tracking: `.claude/plans/MISSING_ACTION_EDGES_PLAN/PROGRESS.yaml`
- Key files affected:
  - `business/sdk/migrate/sql/migrate.sql` (table definition + view definition)
  - `business/sdk/workflow/models.go` (core structs)
  - `business/sdk/workflow/order.go` (order constant)
  - `business/sdk/workflow/workflowbus.go` (create/update logic)
  - `business/sdk/workflow/stores/workflowdb/models.go` (DB structs + conversions)
  - `business/sdk/workflow/stores/workflowdb/workflowdb.go` (SQL queries + JSON aggregation)
  - `api/domain/http/workflow/ruleapi/action_model.go` (API request structs + validation)
  - `api/domain/http/workflow/ruleapi/model.go` (API input/response structs)
  - `api/domain/http/workflow/ruleapi/validation.go` (rule-level validation)
  - `api/domain/http/workflow/ruleapi/ruleapi.go` (duplicate detection)
  - `api/domain/http/workflow/ruleapi/converters.go` (conversion functions)
  - `app/domain/workflow/workflowsaveapp/model.go` (save request/response structs)
  - `app/domain/workflow/workflowsaveapp/workflowsaveapp.go` (conversion logic)
