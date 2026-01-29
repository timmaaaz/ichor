# Phase 0: Business Layer Prerequisites

**Status**: Pending
**Category**: backend
**Dependencies**: None

---

## Overview

This phase addresses critical gaps in the existing workflow business layer that must be resolved before implementing API endpoints for workflow rule management. The code review identified that the planned API phases reference methods and patterns that don't exist in the current codebase.

## Rationale

The workflow business layer (`business/sdk/workflow/`) needs enhancements to support paginated, filterable, and orderable queries for automation rules. Currently:

1. **No Pagination Support** - `QueryAutomationRulesView` doesn't accept pagination parameters
2. **No Filtering Support** - No filter types exist for rule queries
3. **Missing Methods** - `QueryActionByID` doesn't exist in Storer interface
4. **Missing Error Types** - Need `ErrActionNotInRule` for better error handling

---

## LOC Breakdown

| Component | Estimated LOC |
|-----------|---------------|
| Filter types | ~50 |
| Order constants | ~30 |
| Storer interface updates | ~20 |
| Business layer methods | ~80 |
| Store implementation updates | ~120 |
| **Total** | **~300** |

---

## Tasks

### Task 1: Add Filter Type for Automation Rules

**File**: `business/sdk/workflow/filter.go` (create new file)

**Action**: Create a filter type that supports the query parameters needed by the API.

```go
package workflow

import (
    "github.com/google/uuid"
)

// AutomationRuleFilter provides query filtering for automation rules.
type AutomationRuleFilter struct {
    ID            *uuid.UUID
    Name          *string
    IsActive      *bool
    EntityID      *uuid.UUID
    EntityTypeID  *uuid.UUID
    TriggerTypeID *uuid.UUID
    CreatedBy     *uuid.UUID
}
```

**Deliverables**:
- `business/sdk/workflow/filter.go`

---

### Task 2: Add Order Constants for Automation Rules

**File**: `business/sdk/workflow/order.go` (create new file)

**Action**: Define ordering constants for rule queries following the pattern from `business/domain/workflow/alertbus/order.go`.

```go
package workflow

import "github.com/timmaaaz/ichor/business/sdk/order"

// DefaultOrderBy is the default ordering for automation rule queries.
var DefaultOrderBy = order.NewBy(OrderByCreatedDate, order.DESC)

// Order field constants for automation rules.
const (
    OrderByID          = "id"
    OrderByName        = "name"
    OrderByCreatedDate = "created_date"
    OrderByUpdatedDate = "updated_date"
    OrderByIsActive    = "is_active"
)

// Order field constants for rule actions.
const (
    ActionOrderByID             = "id"
    ActionOrderByExecutionOrder = "execution_order"
    ActionOrderByIsActive       = "is_active"
)
```

**Deliverables**:
- `business/sdk/workflow/order.go`

---

### Task 3: Update Storer Interface

**File**: `business/sdk/workflow/workflowbus.go`

**Action**: Add missing methods and update existing signatures to support pagination/filtering.

**Changes Required**:

Add these methods to the `Storer` interface:

```go
type Storer interface {
    // ... existing methods ...

    // UPDATED: Add pagination, filtering, ordering support
    QueryAutomationRulesViewPaginated(ctx context.Context, filter AutomationRuleFilter, orderBy order.By, page page.Page) ([]AutomationRuleView, error)
    CountAutomationRulesView(ctx context.Context, filter AutomationRuleFilter) (int, error)

    // NEW: Query single action by ID
    QueryActionByID(ctx context.Context, actionID uuid.UUID) (RuleAction, error)

    // NEW: Query single action view by ID (with template info)
    QueryActionViewByID(ctx context.Context, actionID uuid.UUID) (RuleActionView, error)
}
```

**Required imports** (add if not present):
- `"github.com/timmaaaz/ichor/business/sdk/order"`
- `"github.com/timmaaaz/ichor/business/sdk/page"`

**Deliverables**:
- Updates to `business/sdk/workflow/workflowbus.go`

---

### Task 4: Add New Error Types

**File**: `business/sdk/workflow/workflowbus.go`

**Action**: Add error types for better error handling in the API layer.

Add to existing error variables:

```go
var (
    // ... existing errors ...

    // ErrActionNotInRule indicates the action does not belong to the specified rule.
    ErrActionNotInRule = errors.New("action does not belong to specified rule")
)
```

**Deliverables**:
- Updates to `business/sdk/workflow/workflowbus.go`

---

### Task 5: Implement Business Layer Methods

**File**: `business/sdk/workflow/workflowbus.go`

**Action**: Add business layer wrapper methods for the new Storer methods.

```go
// QueryAutomationRulesViewPaginated retrieves a paginated view of automation rules.
func (b *Business) QueryAutomationRulesViewPaginated(
    ctx context.Context,
    filter AutomationRuleFilter,
    orderBy order.By,
    pg page.Page,
) ([]AutomationRuleView, error) {
    ctx, span := otel.AddSpan(ctx, "business.workflowbus.queryautomationrulesviewpaginated")
    defer span.End()

    rulesView, err := b.storer.QueryAutomationRulesViewPaginated(ctx, filter, orderBy, pg)
    if err != nil {
        return nil, fmt.Errorf("query: %w", err)
    }

    return rulesView, nil
}

// CountAutomationRulesView returns the total count of rules matching the filter.
func (b *Business) CountAutomationRulesView(ctx context.Context, filter AutomationRuleFilter) (int, error) {
    ctx, span := otel.AddSpan(ctx, "business.workflowbus.countautomationrulesview")
    defer span.End()

    count, err := b.storer.CountAutomationRulesView(ctx, filter)
    if err != nil {
        return 0, fmt.Errorf("count: %w", err)
    }

    return count, nil
}

// QueryActionByID retrieves a single rule action by ID.
func (b *Business) QueryActionByID(ctx context.Context, actionID uuid.UUID) (RuleAction, error) {
    ctx, span := otel.AddSpan(ctx, "business.workflowbus.queryactionbyid")
    defer span.End()

    action, err := b.storer.QueryActionByID(ctx, actionID)
    if err != nil {
        return RuleAction{}, fmt.Errorf("query: actionID[%s]: %w", actionID, err)
    }

    return action, nil
}

// QueryActionViewByID retrieves a single rule action view by ID (with template info).
func (b *Business) QueryActionViewByID(ctx context.Context, actionID uuid.UUID) (RuleActionView, error) {
    ctx, span := otel.AddSpan(ctx, "business.workflowbus.queryactionviewbyid")
    defer span.End()

    actionView, err := b.storer.QueryActionViewByID(ctx, actionID)
    if err != nil {
        return RuleActionView{}, fmt.Errorf("query: actionID[%s]: %w", actionID, err)
    }

    return actionView, nil
}
```

**Required imports** (add if not present):
- `"github.com/timmaaaz/ichor/business/sdk/order"`
- `"github.com/timmaaaz/ichor/business/sdk/page"`

**Deliverables**:
- Updates to `business/sdk/workflow/workflowbus.go`

---

### Task 6: Add Store Filter Implementation

**File**: `business/sdk/workflow/stores/workflowdb/filter.go` (create new file)

**Action**: Implement filter building for database queries following the pattern from `business/domain/workflow/alertbus/stores/alertdb/filter.go`.

```go
package workflowdb

import (
    "bytes"
    "strings"

    "github.com/timmaaaz/ichor/business/sdk/workflow"
)

func applyAutomationRuleFilter(filter workflow.AutomationRuleFilter, data map[string]any, buf *bytes.Buffer) {
    var wc []string

    if filter.ID != nil {
        data["id"] = *filter.ID
        wc = append(wc, "ar.id = :id")
    }

    if filter.Name != nil {
        data["name"] = "%" + *filter.Name + "%"
        wc = append(wc, "ar.name ILIKE :name")
    }

    if filter.IsActive != nil {
        data["is_active"] = *filter.IsActive
        wc = append(wc, "ar.is_active = :is_active")
    }

    if filter.EntityID != nil {
        data["entity_id"] = *filter.EntityID
        wc = append(wc, "ar.entity_id = :entity_id")
    }

    if filter.EntityTypeID != nil {
        data["entity_type_id"] = *filter.EntityTypeID
        wc = append(wc, "ar.entity_type_id = :entity_type_id")
    }

    if filter.TriggerTypeID != nil {
        data["trigger_type_id"] = *filter.TriggerTypeID
        wc = append(wc, "ar.trigger_type_id = :trigger_type_id")
    }

    if filter.CreatedBy != nil {
        data["created_by"] = *filter.CreatedBy
        wc = append(wc, "ar.created_by = :created_by")
    }

    if len(wc) > 0 {
        buf.WriteString(" WHERE ")
        buf.WriteString(strings.Join(wc, " AND "))
    }
}
```

**Deliverables**:
- `business/sdk/workflow/stores/workflowdb/filter.go`

---

### Task 7: Add Store Order Implementation

**File**: `business/sdk/workflow/stores/workflowdb/order.go` (create new file)

**Action**: Implement order mapping for database queries following the pattern from `business/domain/workflow/alertbus/stores/alertdb/order.go`.

```go
package workflowdb

import (
    "fmt"

    "github.com/timmaaaz/ichor/business/sdk/order"
    "github.com/timmaaaz/ichor/business/sdk/workflow"
)

var automationRuleOrderByFields = map[string]string{
    workflow.OrderByID:          "ar.id",
    workflow.OrderByName:        "ar.name",
    workflow.OrderByCreatedDate: "ar.created_date",
    workflow.OrderByUpdatedDate: "ar.updated_date",
    workflow.OrderByIsActive:    "ar.is_active",
}

func orderByClauseAutomationRule(orderBy order.By) (string, error) {
    byField, exists := automationRuleOrderByFields[orderBy.Field]
    if !exists {
        return "", fmt.Errorf("field %q does not exist", orderBy.Field)
    }

    return " ORDER BY " + byField + " " + orderBy.Direction, nil
}
```

**Deliverables**:
- `business/sdk/workflow/stores/workflowdb/order.go`

---

### Task 8: Update Store Implementation

**File**: `business/sdk/workflow/stores/workflowdb/workflowdb.go`

**Action**: Implement the new and updated Storer methods.

**New Methods to Add**:

```go
// QueryAutomationRulesViewPaginated retrieves rules with pagination, filtering, and ordering.
func (s *Store) QueryAutomationRulesViewPaginated(
    ctx context.Context,
    filter workflow.AutomationRuleFilter,
    orderBy order.By,
    pg page.Page,
) ([]workflow.AutomationRuleView, error) {
    data := map[string]any{
        "offset":   (pg.Number() - 1) * pg.RowsPerPage(),
        "rows_per_page": pg.RowsPerPage(),
    }

    const baseQuery = `
    SELECT
        ar.id,
        ar.name,
        ar.description,
        ar.entity_id,
        ar.trigger_conditions,
        ar.is_active,
        ar.created_date,
        ar.updated_date,
        ar.created_by,
        ar.updated_by,
        ar.trigger_type_id,
        COALESCE(tt.name, '') AS trigger_type_name,
        COALESCE(tt.description, '') AS trigger_type_description,
        ar.entity_type_id,
        COALESCE(et.name, '') AS entity_type_name,
        COALESCE(et.description, '') AS entity_type_description,
        COALESCE(e.name, '') AS entity_name,
        COALESCE(e.schema_name, '') AS entity_schema_name,
        COALESCE((
            SELECT json_agg(json_build_object(
                'id', ra.id,
                'name', ra.name,
                'description', ra.description,
                'action_config', ra.action_config,
                'execution_order', ra.execution_order,
                'is_active', ra.is_active,
                'template_id', ra.template_id,
                'deactivated_by', ra.deactivated_by
            ) ORDER BY ra.execution_order)
            FROM workflow.rule_actions ra
            WHERE ra.automation_rule_id = ar.id
        ), '[]'::json) AS actions
    FROM
        workflow.automation_rules ar
    LEFT JOIN
        workflow.trigger_types tt ON ar.trigger_type_id = tt.id
    LEFT JOIN
        workflow.entity_types et ON ar.entity_type_id = et.id
    LEFT JOIN
        workflow.entities e ON ar.entity_id = e.id`

    buf := bytes.NewBufferString(baseQuery)

    applyAutomationRuleFilter(filter, data, buf)

    orderByClause, err := orderByClauseAutomationRule(orderBy)
    if err != nil {
        return nil, fmt.Errorf("orderby: %w", err)
    }
    buf.WriteString(orderByClause)

    buf.WriteString(" LIMIT :rows_per_page OFFSET :offset")

    var dbRules []dbAutomationRuleView
    if err := sqldb.NamedQuerySlice(ctx, s.log, s.db, buf.String(), data, &dbRules); err != nil {
        return nil, fmt.Errorf("namedqueryslice: %w", err)
    }

    return toBusAutomationRulesView(dbRules), nil
}

// CountAutomationRulesView counts rules matching the filter.
func (s *Store) CountAutomationRulesView(
    ctx context.Context,
    filter workflow.AutomationRuleFilter,
) (int, error) {
    data := map[string]any{}

    const baseQuery = `
    SELECT COUNT(ar.id)
    FROM workflow.automation_rules ar`

    buf := bytes.NewBufferString(baseQuery)

    applyAutomationRuleFilter(filter, data, buf)

    var count int
    if err := sqldb.NamedQueryStruct(ctx, s.log, s.db, buf.String(), data, &count); err != nil {
        return 0, fmt.Errorf("namedquerystruct: %w", err)
    }

    return count, nil
}

// QueryActionByID retrieves a single action by its ID.
func (s *Store) QueryActionByID(ctx context.Context, actionID uuid.UUID) (workflow.RuleAction, error) {
    data := struct {
        ID uuid.UUID `db:"id"`
    }{
        ID: actionID,
    }

    const q = `
    SELECT
        id, automation_rule_id, name, description, action_config,
        execution_order, is_active, template_id, deactivated_by
    FROM workflow.rule_actions
    WHERE id = :id`

    var dbAction dbRuleAction
    if err := sqldb.NamedQueryStruct(ctx, s.log, s.db, q, data, &dbAction); err != nil {
        if errors.Is(err, sqldb.ErrDBNotFound) {
            return workflow.RuleAction{}, workflow.ErrNotFound
        }
        return workflow.RuleAction{}, fmt.Errorf("namedquerystruct: %w", err)
    }

    return toBusRuleAction(dbAction), nil
}

// QueryActionViewByID retrieves a single action view by its ID (with template info).
func (s *Store) QueryActionViewByID(ctx context.Context, actionID uuid.UUID) (workflow.RuleActionView, error) {
    data := struct {
        ID uuid.UUID `db:"id"`
    }{
        ID: actionID,
    }

    const q = `
    SELECT
        ra.id,
        ra.automation_rule_id AS automation_rules_id,
        ra.name,
        ra.description,
        ra.action_config,
        ra.execution_order,
        ra.is_active,
        ra.template_id,
        ra.deactivated_by,
        COALESCE(at.name, '') AS template_name,
        COALESCE(at.action_type, '') AS template_action_type,
        COALESCE(at.default_config, '{}'::json) AS template_default_config
    FROM workflow.rule_actions ra
    LEFT JOIN workflow.action_templates at ON ra.template_id = at.id
    WHERE ra.id = :id`

    var dbActionView dbRuleActionView
    if err := sqldb.NamedQueryStruct(ctx, s.log, s.db, q, data, &dbActionView); err != nil {
        if errors.Is(err, sqldb.ErrDBNotFound) {
            return workflow.RuleActionView{}, workflow.ErrNotFound
        }
        return workflow.RuleActionView{}, fmt.Errorf("namedquerystruct: %w", err)
    }

    return toBusRuleActionView(dbActionView), nil
}
```

**Required imports** (add if not present):
- `"bytes"`
- `"github.com/timmaaaz/ichor/business/sdk/order"`
- `"github.com/timmaaaz/ichor/business/sdk/page"`

**Deliverables**:
- Updates to `business/sdk/workflow/stores/workflowdb/workflowdb.go`

---

### Task 9: Update Store Models

**File**: `business/sdk/workflow/stores/workflowdb/models.go`

**Action**: Ensure database model structs exist for the view queries. Add if missing:

```go
// dbAutomationRuleView represents the database structure for the rules view.
type dbAutomationRuleView struct {
    ID                     uuid.UUID        `db:"id"`
    Name                   string           `db:"name"`
    Description            *string          `db:"description"`
    EntityID               *uuid.UUID       `db:"entity_id"`
    TriggerConditions      *json.RawMessage `db:"trigger_conditions"`
    Actions                json.RawMessage  `db:"actions"`
    IsActive               bool             `db:"is_active"`
    CreatedDate            time.Time        `db:"created_date"`
    UpdatedDate            time.Time        `db:"updated_date"`
    CreatedBy              uuid.UUID        `db:"created_by"`
    UpdatedBy              uuid.UUID        `db:"updated_by"`
    TriggerTypeID          *uuid.UUID       `db:"trigger_type_id"`
    TriggerTypeName        string           `db:"trigger_type_name"`
    TriggerTypeDescription string           `db:"trigger_type_description"`
    EntityTypeID           *uuid.UUID       `db:"entity_type_id"`
    EntityTypeName         string           `db:"entity_type_name"`
    EntityTypeDescription  string           `db:"entity_type_description"`
    EntityName             string           `db:"entity_name"`
    EntitySchemaName       string           `db:"entity_schema_name"`
}

// dbRuleActionView represents the database structure for action views.
type dbRuleActionView struct {
    ID                    uuid.UUID        `db:"id"`
    AutomationRulesID     *uuid.UUID       `db:"automation_rules_id"`
    Name                  string           `db:"name"`
    Description           string           `db:"description"`
    ActionConfig          json.RawMessage  `db:"action_config"`
    ExecutionOrder        int              `db:"execution_order"`
    IsActive              bool             `db:"is_active"`
    TemplateID            *uuid.UUID       `db:"template_id"`
    DeactivatedBy         uuid.UUID        `db:"deactivated_by"`
    TemplateName          string           `db:"template_name"`
    TemplateActionType    string           `db:"template_action_type"`
    TemplateDefaultConfig json.RawMessage  `db:"template_default_config"`
}

func toBusAutomationRulesView(dbRules []dbAutomationRuleView) []workflow.AutomationRuleView {
    busRules := make([]workflow.AutomationRuleView, len(dbRules))
    for i, r := range dbRules {
        busRules[i] = toBusAutomationRuleView(r)
    }
    return busRules
}

func toBusAutomationRuleView(dbRule dbAutomationRuleView) workflow.AutomationRuleView {
    return workflow.AutomationRuleView{
        ID:                     dbRule.ID,
        Name:                   dbRule.Name,
        Description:            dbRule.Description,
        EntityID:               dbRule.EntityID,
        TriggerConditions:      dbRule.TriggerConditions,
        Actions:                dbRule.Actions,
        IsActive:               dbRule.IsActive,
        CreatedDate:            dbRule.CreatedDate,
        UpdatedDate:            dbRule.UpdatedDate,
        CreatedBy:              dbRule.CreatedBy,
        UpdatedBy:              dbRule.UpdatedBy,
        TriggerTypeID:          dbRule.TriggerTypeID,
        TriggerTypeName:        dbRule.TriggerTypeName,
        TriggerTypeDescription: dbRule.TriggerTypeDescription,
        EntityTypeID:           dbRule.EntityTypeID,
        EntityTypeName:         dbRule.EntityTypeName,
        EntityTypeDescription:  dbRule.EntityTypeDescription,
        EntityName:             dbRule.EntityName,
        EntitySchemaName:       dbRule.EntitySchemaName,
    }
}

func toBusRuleActionView(dbAction dbRuleActionView) workflow.RuleActionView {
    return workflow.RuleActionView{
        ID:                    dbAction.ID,
        AutomationRulesID:     dbAction.AutomationRulesID,
        Name:                  dbAction.Name,
        Description:           dbAction.Description,
        ActionConfig:          dbAction.ActionConfig,
        ExecutionOrder:        dbAction.ExecutionOrder,
        IsActive:              dbAction.IsActive,
        TemplateID:            dbAction.TemplateID,
        DeactivatedBy:         dbAction.DeactivatedBy,
        TemplateName:          dbAction.TemplateName,
        TemplateActionType:    dbAction.TemplateActionType,
        TemplateDefaultConfig: dbAction.TemplateDefaultConfig,
    }
}
```

**Deliverables**:
- Updates to `business/sdk/workflow/stores/workflowdb/models.go`

---

### Task 10: Verify Database Schema

**Action**: Confirm that all required fields exist in the database schema.

**Checklist**:
- [ ] `workflow.automation_rules` has all fields from `AutomationRule` struct
- [ ] `workflow.rule_actions` has all fields from `RuleAction` struct
- [ ] `workflow.trigger_types` table exists
- [ ] `workflow.entity_types` table exists
- [ ] `workflow.entities` table exists
- [ ] `workflow.action_templates` table exists

**File to check**: `business/sdk/migrate/sql/migrate.sql`

---

## Validation Criteria

- [ ] `go build ./business/sdk/workflow/...` passes
- [ ] `go build ./business/sdk/workflow/stores/workflowdb/...` passes
- [ ] `AutomationRuleFilter` type exists with all filter fields
- [ ] Order constants defined for automation rules
- [ ] `Storer.QueryAutomationRulesViewPaginated` accepts filter, orderBy, page parameters
- [ ] `Storer.CountAutomationRulesView` exists for pagination totals
- [ ] `Storer.QueryActionByID` exists for single action lookup
- [ ] `Storer.QueryActionViewByID` exists for single action view lookup
- [ ] `ErrActionNotInRule` error type defined
- [ ] All business layer wrapper methods implemented
- [ ] Store implementations compile
- [ ] Database schema verified
- [ ] `make test` passes (no regressions)

---

## Deliverables

| Deliverable | File/Location | Status |
|-------------|---------------|--------|
| Filter types | `business/sdk/workflow/filter.go` | pending |
| Order constants | `business/sdk/workflow/order.go` | pending |
| Storer interface updates | `business/sdk/workflow/workflowbus.go` | pending |
| Error types | `business/sdk/workflow/workflowbus.go` | pending |
| Business methods | `business/sdk/workflow/workflowbus.go` | pending |
| Store filter impl | `business/sdk/workflow/stores/workflowdb/filter.go` | pending |
| Store order impl | `business/sdk/workflow/stores/workflowdb/order.go` | pending |
| Store query impl | `business/sdk/workflow/stores/workflowdb/workflowdb.go` | pending |
| Store models | `business/sdk/workflow/stores/workflowdb/models.go` | pending |

---

## Reference Files

**Existing Patterns to Follow**:
- `business/domain/workflow/alertbus/filter.go` - Filter pattern example
- `business/domain/workflow/alertbus/order.go` - Order pattern example
- `business/domain/workflow/alertbus/alertbus.go` - Business layer with pagination
- `business/domain/workflow/alertbus/stores/alertdb/filter.go` - Store filter implementation
- `business/domain/workflow/alertbus/stores/alertdb/order.go` - Store order implementation
- `business/sdk/order/order.go` - Order type utilities
- `business/sdk/page/page.go` - Pagination utilities

---

## Commands

```bash
# Verify compilation after changes
go build ./business/sdk/workflow/...
go build ./business/sdk/workflow/stores/workflowdb/...

# Run existing workflow tests
go test ./business/sdk/workflow/...

# Check for lint errors
make lint

# Run full test suite
make test
```

---

**Last Updated**: 2026-01-29
