# Phase 8: Edge Store Adapter

**Category**: backend
**Status**: Pending
**Dependencies**: Phase 3 (Core Models & Context - COMPLETED, provides `ActionNode`, `ActionEdge`, `GraphDefinition` types), Phase 7 (Trigger System - COMPLETED, defines `EdgeStore` interface in `trigger.go`)

---

## Overview

Implement the database adapter that satisfies the `EdgeStore` interface (defined in Phase 7) by loading graph definitions from PostgreSQL. The adapter queries `workflow.rule_actions` (joined with `workflow.action_templates` for ActionType resolution) and `workflow.action_edges` to build the `temporal.ActionNode` and `temporal.ActionEdge` types used by the trigger system and graph executor.

This is a thin database translation layer following the existing `stores/` pattern in the codebase (e.g., `stores/workflowdb/`, `stores/roledb/`). It uses `sqldb.NamedQuerySlice` for query execution and handles NULL columns (`source_action_id`, `template_id`, `deactivated_by`) with `sql.NullString`.

### Relationship to Existing Store

This store is intentionally separate from `business/sdk/workflow/stores/workflowdb/`. The existing store:
- Returns `workflow.RuleAction` and `workflow.ActionEdge` (business layer types)
- Is used by the existing workflow engine
- Has many more methods (CRUD operations, views, etc.)

The new edgedb store:
- Returns `temporal.ActionNode` and `temporal.ActionEdge` (Temporal layer types)
- Is used only by the Temporal trigger system
- Has exactly 2 read-only methods
- Can coexist with the existing store during migration

## Goals

1. **Implement EdgeStore database adapter** that loads `ActionNode` and `ActionEdge` from PostgreSQL using the `rule_actions` table joined with `action_templates` (for ActionType resolution) and the `action_edges` table
2. **Write SQL queries using the existing `sqldb.NamedQuerySlice` pattern** with proper NULL handling for start edges (`source_action_id` is NULL) and nullable template fields (`template_id`, `deactivated_by`)
3. **Write integration tests using the `dbtest` package** that verify graph loading against seeded workflow data in a real PostgreSQL test container

## Prerequisites

- Phase 3 complete: `temporal.ActionNode`, `temporal.ActionEdge` types (target conversion types)
- Phase 7 complete: `EdgeStore` interface definition in `trigger.go`
- Database schema: `workflow.rule_actions`, `workflow.action_edges`, `workflow.action_templates` tables
- Existing patterns: `sqldb.NamedQuerySlice`, `sql.NullString`, named parameter queries
- Seeded workflow data available in test database (via `dbtest` / migration seed)

---

## Go Package Structure

```
business/sdk/workflow/temporal/
    models.go              <- Phase 3 (COMPLETED)
    models_test.go         <- Phase 3 (COMPLETED)
    graph_executor.go      <- Phase 4 (COMPLETED)
    graph_executor_test.go <- Phase 4 (COMPLETED)
    workflow.go            <- Phase 5 (COMPLETED)
    activities.go          <- Phase 6 (COMPLETED)
    activities_async.go    <- Phase 6 (COMPLETED)
    async_completer.go     <- Phase 6 (COMPLETED)
    trigger.go             <- Phase 7 (COMPLETED) - defines EdgeStore interface
    trigger_test.go        <- Phase 7 (COMPLETED)
    stores/
        edgedb/
            edgedb.go      <- THIS PHASE (Task 2)
            edgedb_test.go <- THIS PHASE (Task 3)
```

---

## Database Schema Reference

### workflow.rule_actions Table

```sql
CREATE TABLE workflow.rule_actions (
    id                  UUID PRIMARY KEY,
    automation_rules_id UUID NOT NULL REFERENCES workflow.automation_rules(id),
    name                VARCHAR(100) NOT NULL,
    description         TEXT,
    action_config       JSONB NOT NULL,
    is_active           BOOLEAN DEFAULT TRUE,
    template_id         UUID NULL REFERENCES workflow.action_templates(id),
    deactivated_by      UUID NULL REFERENCES core.users(id)
);
```

**Key columns for EdgeStore**:
- `id`, `name`, `description`, `action_config`, `is_active`, `deactivated_by` → map directly to `ActionNode`
- `template_id` → JOIN to `action_templates.action_type` → `ActionNode.ActionType`

### workflow.action_templates Table

```sql
CREATE TABLE workflow.action_templates (
    id             UUID PRIMARY KEY,
    name           VARCHAR(100) NOT NULL UNIQUE,
    description    TEXT,
    action_type    VARCHAR(50) NOT NULL,  -- ← THIS IS ActionNode.ActionType
    default_config JSONB NOT NULL,
    ...
);
```

### workflow.action_edges Table

```sql
CREATE TABLE workflow.action_edges (
    id               UUID PRIMARY KEY,
    rule_id          UUID NOT NULL REFERENCES workflow.automation_rules(id) ON DELETE CASCADE,
    source_action_id UUID NULL REFERENCES workflow.rule_actions(id) ON DELETE CASCADE,  -- NULL for start edges
    target_action_id UUID NOT NULL REFERENCES workflow.rule_actions(id) ON DELETE CASCADE,
    edge_type        VARCHAR(20) NOT NULL CHECK (edge_type IN ('start', 'sequence', 'true_branch', 'false_branch', 'always')),
    edge_order       INTEGER DEFAULT 0,
    created_date     TIMESTAMP NOT NULL DEFAULT NOW()
);
```

### workflow.rule_actions_view (Existing)

```sql
CREATE OR REPLACE VIEW workflow.rule_actions_view AS
   SELECT
      ra.id, ra.automation_rules_id, ra.name, ra.description,
      ra.action_config, ra.is_active, ra.template_id,
      at.name as template_name,
      at.action_type as template_action_type,
      at.default_config as template_default_config
   FROM workflow.rule_actions ra
   LEFT JOIN workflow.action_templates at ON ra.template_id = at.id;
```

**Note**: The existing `rule_actions_view` doesn't include `deactivated_by` and is used by other parts of the system. Rather than modifying a shared view, this store writes a focused query that selects exactly what `ActionNode` needs (`deactivated_by` + `template_action_type`). See Design Decision #1.

---

## EdgeStore Interface (From Phase 7)

```go
// EdgeStore loads graph definitions (actions + edges) from the database.
type EdgeStore interface {
    QueryActionsByRule(ctx context.Context, ruleID uuid.UUID) ([]ActionNode, error)
    QueryEdgesByRule(ctx context.Context, ruleID uuid.UUID) ([]ActionEdge, error)
}
```

---

## Task Breakdown

### Task 1: Create stores/edgedb Directory

**Status**: Pending

**Description**: Create the `business/sdk/workflow/temporal/stores/edgedb/` directory. This follows the existing store naming convention (e.g., `stores/workflowdb/`, `stores/roledb/`).

**Files**:
- `business/sdk/workflow/temporal/stores/edgedb/` (directory)

---

### Task 2: Implement edgedb.go

**Status**: Pending

**Description**: Implement the `Store` struct that satisfies the `EdgeStore` interface. The store has two SQL queries: one for loading actions (joined with templates for ActionType) and one for loading edges. Each query uses `sqldb.NamedQuerySlice` with named parameters and struct-based parameter binding. Database models (`dbAction`, `dbEdge`) handle NULL columns with `sql.NullString`, and conversion functions map to the `temporal.ActionNode`/`temporal.ActionEdge` types.

**Notes**:
- `Store` struct with `*logger.Logger` and `sqlx.ExtContext` (matches existing patterns)
- `dbAction` struct - database model for joined query (includes `template_action_type`)
- `dbEdge` struct - database model for action_edges
- `QueryActionsByRule` - JOINs `rule_actions` with `action_templates` for ActionType
- `QueryEdgesByRule` - queries `action_edges` ordered by `edge_order`
- `toActionNode` / `toActionEdge` - conversion helpers
- Uses `sql.NullString` for nullable UUID columns (`source_action_id`, `deactivated_by`)
- Uses `sqldb.NamedQuerySlice` (NOT raw `db.SelectContext`) to follow codebase conventions

**Files**:
- `business/sdk/workflow/temporal/stores/edgedb/edgedb.go`

**Implementation Guide**:

```go
// Package edgedb implements the EdgeStore interface for loading graph
// definitions from PostgreSQL. Used by the Temporal trigger system
// to build GraphDefinition from rule_actions and action_edges tables.
package edgedb

import (
    "context"
    "database/sql"
    "encoding/json"
    "fmt"

    "github.com/google/uuid"
    "github.com/jmoiron/sqlx"

    "github.com/timmaaaz/ichor/business/sdk/sqldb"
    "github.com/timmaaaz/ichor/business/sdk/workflow/temporal"
    "github.com/timmaaaz/ichor/foundation/logger"
)

// Store implements the temporal.EdgeStore interface by loading
// graph definitions from PostgreSQL.
type Store struct {
    log *logger.Logger
    db  sqlx.ExtContext
}

// NewStore creates a new edge store.
func NewStore(log *logger.Logger, db sqlx.ExtContext) *Store {
    return &Store{
        log: log,
        db:  db,
    }
}

// =============================================================================
// Database Models
// =============================================================================

// dbAction represents the database model for rule_actions joined with
// action_templates to resolve ActionType.
//
// Uses sql.NullString for nullable columns:
//   - deactivated_by: NULL when action is active
//   - template_action_type: NULL when no template is linked (template_id is NULL)
type dbAction struct {
    ID                 string          `db:"id"`
    Name               string          `db:"name"`
    Description        sql.NullString  `db:"description"`
    ActionConfig       json.RawMessage `db:"action_config"`
    IsActive           bool            `db:"is_active"`
    DeactivatedBy      sql.NullString  `db:"deactivated_by"`
    TemplateActionType sql.NullString  `db:"template_action_type"`
}

// dbEdge represents the database model for action_edges.
//
// Uses sql.NullString for source_action_id which is NULL for start edges.
type dbEdge struct {
    ID             string         `db:"id"`
    SourceActionID sql.NullString `db:"source_action_id"`
    TargetActionID string         `db:"target_action_id"`
    EdgeType       string         `db:"edge_type"`
    EdgeOrder      int            `db:"edge_order"`
}

// =============================================================================
// EdgeStore Interface Implementation
// =============================================================================

// QueryActionsByRule returns all action nodes for a given automation rule,
// with ActionType resolved from the linked action_template.
func (s *Store) QueryActionsByRule(ctx context.Context, ruleID uuid.UUID) ([]temporal.ActionNode, error) {
    data := struct {
        RuleID string `db:"automation_rules_id"`
    }{
        RuleID: ruleID.String(),
    }

    const q = `
    SELECT
        ra.id,
        ra.name,
        ra.description,
        ra.action_config,
        ra.is_active,
        ra.deactivated_by,
        at.action_type AS template_action_type
    FROM
        workflow.rule_actions ra
    LEFT JOIN
        workflow.action_templates at ON ra.template_id = at.id
    WHERE
        ra.automation_rules_id = :automation_rules_id`

    var dbActions []dbAction
    if err := sqldb.NamedQuerySlice(ctx, s.log, s.db, q, data, &dbActions); err != nil {
        return nil, fmt.Errorf("namedqueryslice[actions]: %w", err)
    }

    actions := make([]temporal.ActionNode, len(dbActions))
    for i, dba := range dbActions {
        actions[i] = toActionNode(dba)
    }

    return actions, nil
}

// QueryEdgesByRule returns all action edges for a given automation rule,
// ordered by edge_order for deterministic graph traversal.
func (s *Store) QueryEdgesByRule(ctx context.Context, ruleID uuid.UUID) ([]temporal.ActionEdge, error) {
    data := struct {
        RuleID string `db:"rule_id"`
    }{
        RuleID: ruleID.String(),
    }

    const q = `
    SELECT
        id,
        source_action_id,
        target_action_id,
        edge_type,
        edge_order
    FROM
        workflow.action_edges
    WHERE
        rule_id = :rule_id
    ORDER BY
        edge_order ASC`

    var dbEdges []dbEdge
    if err := sqldb.NamedQuerySlice(ctx, s.log, s.db, q, data, &dbEdges); err != nil {
        return nil, fmt.Errorf("namedqueryslice[edges]: %w", err)
    }

    edges := make([]temporal.ActionEdge, len(dbEdges))
    for i, dbe := range dbEdges {
        edges[i] = toActionEdge(dbe)
    }

    return edges, nil
}

// =============================================================================
// Conversion Functions
// =============================================================================

// toActionNode converts a database action row to a temporal.ActionNode.
func toActionNode(dba dbAction) temporal.ActionNode {
    node := temporal.ActionNode{
        ID:       uuid.MustParse(dba.ID),
        Name:     dba.Name,
        Config:   dba.ActionConfig,
        IsActive: dba.IsActive,
    }

    if dba.Description.Valid {
        node.Description = dba.Description.String
    }

    if dba.DeactivatedBy.Valid {
        node.DeactivatedBy = uuid.MustParse(dba.DeactivatedBy.String)
    }

    if dba.TemplateActionType.Valid {
        node.ActionType = dba.TemplateActionType.String
    }

    return node
}

// toActionEdge converts a database edge row to a temporal.ActionEdge.
func toActionEdge(dbe dbEdge) temporal.ActionEdge {
    edge := temporal.ActionEdge{
        ID:             uuid.MustParse(dbe.ID),
        TargetActionID: uuid.MustParse(dbe.TargetActionID),
        EdgeType:       dbe.EdgeType,
        SortOrder:      dbe.EdgeOrder,
    }

    // source_action_id is NULL for start edges
    if dbe.SourceActionID.Valid {
        id := uuid.MustParse(dbe.SourceActionID.String)
        edge.SourceActionID = &id
    }

    return edge
}
```

**Design Decisions**:

1. **Custom query instead of existing view** - The `rule_actions_view` doesn't include `deactivated_by`. Rather than modifying the shared view (which other code depends on), we write a focused query that selects exactly what `ActionNode` needs.

2. **LEFT JOIN for ActionType** - Actions may not have a template (`template_id` is nullable). The LEFT JOIN ensures actions without templates still appear in results (with `ActionType` as empty string). Phase 6's activity dispatcher should handle missing ActionType gracefully.

3. **`sqlx.ExtContext` instead of `*sqlx.DB`** - Matches existing store patterns. `ExtContext` is satisfied by both `*sqlx.DB` and `*sqlx.Tx`, enabling the store to work within transactions if needed.

4. **`uuid.MustParse` for ID conversion** - Database IDs are validated at write time (UUID PRIMARY KEY constraint), so parsing cannot fail for valid data. If corrupt data exists, the panic is preferable to silently skipping rows.

5. **`edge_order` → `SortOrder` mapping** - The database column is `edge_order` but the temporal model uses `SortOrder`. The conversion function handles this rename.

6. **No `created_date` on edges** - The `ActionEdge` temporal model doesn't need `created_date` (it's metadata, not execution-relevant). We intentionally omit it from the SELECT.

---

### Task 3: Write Integration Tests

**Status**: Pending

**Description**: Write integration tests that run against a real PostgreSQL test container with seeded workflow data. Tests should verify both queries return correct data, handle empty results, and properly convert NULL columns.

**Notes**:
- Use `dbtest` package for database setup
- Need seeded workflow data: automation rule with rule_actions, action_templates, and action_edges
- Test `QueryActionsByRule` with seeded data
- Test `QueryEdgesByRule` with seeded data
- Test with non-existent rule ID (empty results)
- Test NULL handling: start edges (source_action_id NULL), actions without templates (template_id NULL)
- Test `edge_order` ordering
- Consider using existing seed data or inserting test-specific data

**Files**:
- `business/sdk/workflow/temporal/stores/edgedb/edgedb_test.go`

**Implementation Guide**:

The integration tests use `dbtest.NewDatabase` for a PostgreSQL test container, then `workflow.TestSeedFullWorkflow` to seed automation rules, action templates, rule actions, and action edges. This seeder creates 5 rules with 10 actions (distributed across rules), each rule getting a start edge + sequence edge chain.

**Key types from `dbtest`**:
- `dbtest.NewDatabase(t, name)` → `*Database` with `.DB` (`*sqlx.DB`), `.Log` (`*logger.Logger`), `.BusDomain`
- `dbtest.BusDomain.Workflow` → `*workflow.Business` (for seeding)
- `workflow.TestSeedFullWorkflow(ctx, userID, api)` → `*TestWorkflowData` with `.AutomationRules`, `.RuleActions`, `.ActionTemplates`

**Data volumes**: `TestSeedFullWorkflow` creates 5 rules, 3 templates, 10 actions distributed across rules. Each rule gets a start edge + sequence edges (so 2 actions per rule = 2 edges: 1 start + 1 sequence).

```go
package edgedb_test

import (
    "context"
    "encoding/json"
    "testing"

    "github.com/google/uuid"

    "github.com/timmaaaz/ichor/business/domain/core/userbus"
    "github.com/timmaaaz/ichor/business/sdk/dbtest"
    "github.com/timmaaaz/ichor/business/sdk/workflow"
    "github.com/timmaaaz/ichor/business/sdk/workflow/temporal"
    "github.com/timmaaaz/ichor/business/sdk/workflow/temporal/stores/edgedb"
)

func Test_EdgeDB(t *testing.T) {
    db := dbtest.NewDatabase(t, "Test_EdgeDB")

    // Seed a user (needed for TestSeedFullWorkflow's createdBy FK)
    ctx := context.Background()
    users, err := userbus.TestSeedUsersWithNoFKs(ctx, 1, userbus.Roles.Admin, db.BusDomain.User)
    if err != nil {
        t.Fatalf("Seeding users: %s", err)
    }

    // Seed full workflow data: 5 rules, 3 templates, 10 actions, edges
    wfData, err := workflow.TestSeedFullWorkflow(ctx, users[0].ID, db.BusDomain.Workflow)
    if err != nil {
        t.Fatalf("Seeding workflow: %s", err)
    }

    store := edgedb.NewStore(db.Log, db.DB)

    // Pick a rule that has actions seeded
    ruleID := wfData.AutomationRules[0].ID

    // -------------------------------------------------------------------------
    // QueryActionsByRule - success
    t.Run("query-actions-success", func(t *testing.T) {
        actions, err := store.QueryActionsByRule(ctx, ruleID)
        if err != nil {
            t.Fatalf("QueryActionsByRule: %s", err)
        }
        if len(actions) == 0 {
            t.Fatal("expected at least one action")
        }
        for _, a := range actions {
            if a.ID == uuid.Nil {
                t.Error("action ID should not be nil")
            }
            if a.Name == "" {
                t.Error("action Name should not be empty")
            }
            // ActionType should be resolved from template (non-empty for seeded data)
            if a.ActionType == "" {
                t.Log("warning: ActionType empty (action may lack template)")
            }
            // Config should be valid JSON
            if !json.Valid(a.Config) {
                t.Error("action Config is not valid JSON")
            }
        }
    })

    // -------------------------------------------------------------------------
    // QueryActionsByRule - not found (random UUID)
    t.Run("query-actions-not-found", func(t *testing.T) {
        actions, err := store.QueryActionsByRule(ctx, uuid.New())
        if err != nil {
            t.Fatalf("QueryActionsByRule: expected nil error, got %s", err)
        }
        if len(actions) != 0 {
            t.Fatalf("expected empty slice, got %d actions", len(actions))
        }
    })

    // -------------------------------------------------------------------------
    // QueryEdgesByRule - success with start edge and ordering
    t.Run("query-edges-success", func(t *testing.T) {
        edges, err := store.QueryEdgesByRule(ctx, ruleID)
        if err != nil {
            t.Fatalf("QueryEdgesByRule: %s", err)
        }
        if len(edges) == 0 {
            t.Fatal("expected at least one edge")
        }

        // Verify at least one start edge (nil SourceActionID)
        hasStart := false
        for _, e := range edges {
            if e.EdgeType == temporal.EdgeTypeStart {
                hasStart = true
                if e.SourceActionID != nil {
                    t.Error("start edge should have nil SourceActionID")
                }
            }
        }
        if !hasStart {
            t.Error("expected at least one start edge")
        }

        // Verify ordering: SortOrder should be non-decreasing
        for i := 1; i < len(edges); i++ {
            if edges[i].SortOrder < edges[i-1].SortOrder {
                t.Errorf("edges not ordered: [%d].SortOrder=%d < [%d].SortOrder=%d",
                    i, edges[i].SortOrder, i-1, edges[i-1].SortOrder)
            }
        }
    })

    // -------------------------------------------------------------------------
    // QueryEdgesByRule - not found (random UUID)
    t.Run("query-edges-not-found", func(t *testing.T) {
        edges, err := store.QueryEdgesByRule(ctx, uuid.New())
        if err != nil {
            t.Fatalf("QueryEdgesByRule: expected nil error, got %s", err)
        }
        if len(edges) != 0 {
            t.Fatalf("expected empty slice, got %d edges", len(edges))
        }
    })

    // -------------------------------------------------------------------------
    // Round-trip: load full graph and verify referential integrity
    t.Run("round-trip", func(t *testing.T) {
        actions, err := store.QueryActionsByRule(ctx, ruleID)
        if err != nil {
            t.Fatalf("QueryActionsByRule: %s", err)
        }
        edges, err := store.QueryEdgesByRule(ctx, ruleID)
        if err != nil {
            t.Fatalf("QueryEdgesByRule: %s", err)
        }

        // Build action ID set
        actionIDs := make(map[uuid.UUID]bool, len(actions))
        for _, a := range actions {
            actionIDs[a.ID] = true
        }

        // Verify all edge targets reference valid actions
        for _, e := range edges {
            if !actionIDs[e.TargetActionID] {
                t.Errorf("edge %s targets unknown action %s", e.ID, e.TargetActionID)
            }
            if e.SourceActionID != nil && !actionIDs[*e.SourceActionID] {
                t.Errorf("edge %s sources unknown action %s", e.ID, *e.SourceActionID)
            }
        }
    })
}
```

**Testing Approach**:

The integration tests use `workflow.TestSeedFullWorkflow` which creates complete workflow data including automation rules, action templates, rule actions, and action edges (with start + sequence edges per rule). This is the recommended approach because:

1. **Seeding is already implemented** - `TestSeedFullWorkflow` handles all FK constraints and creates realistic data
2. **Edges are guaranteed** - The edge enforcement project (completed 2026-02-06) ensured `TestSeedRuleActions` always creates edge chains
3. **Templates are linked** - Seeded actions have `template_id` set, so `ActionType` resolves via the LEFT JOIN
4. **No raw SQL needed** - Business layer seeding handles all schema constraints

**Discovering seed data volumes**:
```bash
# After running tests, check what was seeded
make pgcli
SELECT r.id as rule_id, COUNT(ra.id) as action_count
FROM workflow.automation_rules r
LEFT JOIN workflow.rule_actions ra ON r.id = ra.automation_rules_id
GROUP BY r.id LIMIT 5;

SELECT rule_id, COUNT(*) as edge_count
FROM workflow.action_edges
GROUP BY rule_id LIMIT 5;
```

---

## Validation Criteria

- [ ] `go build ./business/sdk/workflow/temporal/...` passes
- [ ] `go build ./business/sdk/workflow/temporal/stores/edgedb/...` passes
- [ ] `edgedb.go` compiles with correct imports (`sqldb`, `temporal`, `sql`, `json`)
- [ ] `Store` satisfies `temporal.EdgeStore` interface (compile-time check)
- [ ] `QueryActionsByRule` JOINs with `action_templates` for ActionType resolution
- [ ] `QueryActionsByRule` includes `deactivated_by` column
- [ ] `QueryEdgesByRule` orders by `edge_order ASC`
- [ ] `QueryEdgesByRule` handles NULL `source_action_id` (start edges → nil pointer)
- [ ] `toActionNode` handles NULL `description`, `deactivated_by`, `template_action_type`
- [ ] Integration tests pass against test database
- [ ] Non-existent rule ID returns empty slice (not error)
- [ ] No import cycles between `edgedb`, `temporal`, and `workflow` packages

---

## Deliverables

- `business/sdk/workflow/temporal/stores/edgedb/edgedb.go`
- `business/sdk/workflow/temporal/stores/edgedb/edgedb_test.go`

---

## Gotchas & Tips

### Common Pitfalls

- **ActionType comes from `action_templates`, NOT `rule_actions`** - The `rule_actions` table does NOT have an `action_type` column. It must be resolved via the LEFT JOIN with `action_templates` on `template_id`. If `template_id` is NULL (no linked template), `ActionType` will be an empty string, which is valid - Phase 6's activity dispatcher should handle missing ActionType gracefully.

- **The existing `rule_actions_view` is missing `deactivated_by`** - Don't use the view directly. Write a custom query that includes `deactivated_by` from `rule_actions` and `action_type` from `action_templates`.

- **Column name is `edge_order`, not `sequence_order`** - The `action_edges` table uses `edge_order` (confirmed in migration DDL). Some older plan documents may reference `sequence_order` which does not exist. The queries in this plan already use the correct column name.

- **`sql.NullString` for UUID columns** - The existing codebase pattern uses `sql.NullString` (not `*uuid.UUID`) for nullable UUID columns. Follow the same pattern in `dbAction` and `dbEdge`.

- **`uuid.MustParse` safety** - Using `MustParse` assumes database UUIDs are always valid. This is safe because PostgreSQL enforces UUID type constraints. If you're concerned, use `uuid.Parse` with error handling instead, but the existing codebase uses `MustParse` freely.

- **Named parameters use struct tags** - `sqldb.NamedQuerySlice` uses the `db:"field_name"` struct tag on the parameter struct to bind `:field_name` in SQL. Make sure the tag matches the SQL parameter name exactly.

- **`NamedQuerySlice` returns empty slices for zero rows, NOT an error** - Unlike `NamedQueryStruct` (which returns `ErrDBNotFound` for single-row queries), `NamedQuerySlice` simply sets the destination slice to `nil` when no rows match. This is correct behavior for `EdgeStore` - empty results are valid, not errors. No special error handling is needed. (Confirmed by reading `sqldb.namedQuerySlice` implementation: it iterates `rows.Next()`, appends to a local slice, and sets `*dest = slice` - an empty loop yields `nil`.)

### Tips

- Start by reading `business/sdk/workflow/stores/workflowdb/workflowdb.go` for the existing `QueryActionsByRule` and `QueryEdgesByRuleID` implementations - these are your closest reference.
- The `sqlx.ExtContext` parameter type allows the store to be used with both `*sqlx.DB` and `*sqlx.Tx`. Use this instead of `*sqlx.DB`.
- Add a compile-time interface check: `var _ temporal.EdgeStore = (*Store)(nil)` at the top of the file.
- For integration tests, use `dbtest.NewDatabase` + `workflow.TestSeedFullWorkflow` (see Task 3 implementation guide for complete example). The `dbtest.BusDomain.Workflow` field provides the `*workflow.Business` needed for seeding.
- The `edge_order` column defaults to 0, so edges without explicit ordering will have `SortOrder: 0`.
- If a linked `action_template` is deleted (but `template_id` still references it), the LEFT JOIN returns NULL for `template_action_type`. This is handled correctly by `sql.NullString` - `ActionType` becomes empty string.

### Expected Data Volumes

Typical rules have 5-20 actions, each action having 1-3 outgoing edges. Start edges: typically 1 per rule (single entry point). The `TestSeedFullWorkflow` helper creates 5 rules, 3 templates, 10 actions (distributed across rules), with start + sequence edge chains per rule.

---

## Testing Strategy

### Integration Tests (This Phase)

Integration tests run against a real PostgreSQL test container:

1. **Test database setup** - Use `dbtest` package to get a database connection with seeded data
2. **Query verification** - Verify actions and edges load correctly for seeded rule IDs
3. **NULL handling** - Verify start edges (nil SourceActionID) and templateless actions
4. **Empty results** - Verify non-existent rule IDs return empty slices
5. **Ordering** - Verify edges are ordered by edge_order

### Compile-Time Checks

```go
// Verify Store implements EdgeStore interface at compile time
var _ temporal.EdgeStore = (*Store)(nil)
```

### No Unit Tests Needed

This is a thin database adapter with no business logic. Integration tests against the real database are sufficient. Mocking the database would add no value.

---

## Commands Reference

```bash
# Start this phase
/workflow-temporal-next

# Validate this phase
/workflow-temporal-validate 8

# Review plan before implementing
/workflow-temporal-plan-review 8

# Review code after implementing
/workflow-temporal-review 8
```
