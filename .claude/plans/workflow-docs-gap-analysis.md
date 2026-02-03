# Workflow Documentation Gap Analysis - Phase 12

**Date**: 2026-02-03
**Purpose**: Comprehensive documentation updates for Phase 12 workflow functionality

---

## Executive Summary

Phase 12 has been **fully implemented** but the documentation has **significant gaps**. This plan identifies every documentation file that needs updating and exactly what's missing.

---

## Current Documentation Structure

```
docs/workflow/
├── README.md              # Main entry point
├── api-reference.md       # REST API documentation (ONLY covers Alert API)
├── architecture.md        # System overview
├── database-schema.md     # Schema documentation
├── event-infrastructure.md # Event flow details
├── adding-domains.md      # Domain integration guide
├── testing.md             # Testing guide
├── configuration/
│   ├── triggers.md
│   ├── rules.md
│   └── templates.md
└── actions/
    ├── overview.md
    ├── update-field.md
    ├── create-alert.md
    ├── send-email.md
    ├── send-notification.md
    ├── seek-approval.md
    └── allocate-inventory.md
```

**Missing Files:**
- `actions/evaluate-condition.md` - **DOES NOT EXIST**
- `branching.md` or `graph-execution.md` - **DOES NOT EXIST**
- `cascade-visualization.md` - **DOES NOT EXIST**

---

## Gap Analysis by File

### 1. NEW FILE: `actions/evaluate-condition.md`

**Status**: COMPLETELY MISSING

**Must document:**
- Action type: `evaluate_condition`
- Configuration structure:
  ```go
  type ConditionConfig struct {
      Conditions []FieldCondition `json:"conditions"`
      LogicType  string           `json:"logic_type"` // "and" (default) or "or"
  }
  ```
- All 10 operators: `equals`, `not_equals`, `changed_from`, `changed_to`, `greater_than`, `less_than`, `contains`, `in`, `is_null`, `is_not_null`
- `ConditionResult` return type with `BranchTaken` field
- Does NOT support manual execution
- Connection to edge system (true_branch/false_branch edges)
- Example configurations with branching scenarios

**Source**: `business/sdk/workflow/workflowactions/control/condition.go` (lines 1-303)

---

### 2. NEW FILE: `branching.md` (Graph-Based Execution)

**Status**: COMPLETELY MISSING

**Must document:**
- `action_edges` table and its purpose
- 5 edge types: `start`, `sequence`, `true_branch`, `false_branch`, `always`
- `ExecuteRuleActionsGraph()` method and BFS traversal
- Backwards compatibility (falls back to `execution_order` when no edges)
- Edge ordering with `edge_order` field
- Cycle prevention (visited nodes not re-executed)
- Visual examples of branching workflows:
  - Simple branch (condition → true/false paths)
  - Diamond pattern (converging branches)
  - Nested conditions

**Source**: `business/sdk/workflow/executor.go` (lines 211-402)

---

### 3. NEW FILE: `cascade-visualization.md`

**Status**: COMPLETELY MISSING

**Must document:**
- `EntityModifier` interface
- `GetEntityModifications()` method
- `EntityModification` struct (entity_name, event_type, fields)
- Which handlers implement it (currently only `UpdateFieldHandler`)
- `/workflow/rules/{id}/cascade-map` endpoint
- Response structure with `DownstreamWorkflowInfo`
- Self-trigger exclusion logic
- Human-readable trigger descriptions

**Source**: `api/domain/http/workflow/ruleapi/cascade.go` (lines 53-120)

---

### 4. `README.md` - Main Entry Point

**Status**: NEEDS UPDATE

**Specific line updates:**
- **Line 11**: Change "All 6 action types" → "All 7 action types"
- **Lines 40-47**: Add `evaluate_condition` row to Supported Actions table:
  ```
  | `evaluate_condition` | Evaluates conditions for branching |
  ```
- **Lines 91-102**: Add missing operators `is_null`, `is_not_null` to Condition Operators table
- **Lines 136-149**: Add `action_edges` to Database Tables list

**Missing content:**
- Quick link to branching/graph execution docs (after line 16)
- Quick link to cascade visualization docs
- Note about graph-based vs linear execution modes
- Mention of `BranchTaken` for conditional workflows

---

### 5. `architecture.md` - System Overview

**Status**: NEEDS MAJOR UPDATE

**Currently describes**: Linear sequential execution model

**Missing sections:**
- **Graph-Based Executor** section explaining:
  - `ActionEdge` model
  - BFS traversal algorithm
  - Edge type filtering logic
  - `ShouldFollowEdge()` function
- **Condition Nodes** section explaining:
  - How `evaluate_condition` integrates with edges
  - `BranchTaken` field in `ActionResult`
- **Cascade Visualization** section explaining:
  - `EntityModifier` interface
  - How downstream workflows are detected
- **Manual Execution** section (already partially implemented but undocumented)
- Update execution flow diagram to show branch points

---

### 6. `database-schema.md` - Schema Documentation

**Status**: NEEDS UPDATE

**Missing table (migration version 1.992):**
```sql
CREATE TABLE workflow.action_edges (
   id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
   rule_id UUID NOT NULL REFERENCES workflow.automation_rules(id) ON DELETE CASCADE,
   source_action_id UUID REFERENCES workflow.rule_actions(id) ON DELETE CASCADE,
   target_action_id UUID NOT NULL REFERENCES workflow.rule_actions(id) ON DELETE CASCADE,
   edge_type VARCHAR(20) NOT NULL CHECK (edge_type IN ('start', 'sequence', 'true_branch', 'false_branch', 'always')),
   edge_order INTEGER DEFAULT 0,
   created_date TIMESTAMP NOT NULL DEFAULT NOW(),
   CONSTRAINT unique_edge UNIQUE(source_action_id, target_action_id, edge_type)
);
```

**Missing indices:**
```sql
CREATE INDEX idx_action_edges_source ON workflow.action_edges(source_action_id);
CREATE INDEX idx_action_edges_target ON workflow.action_edges(target_action_id);
CREATE INDEX idx_action_edges_rule ON workflow.action_edges(rule_id);
```

---

### 7. `api-reference.md` - REST API

**Status**: NEEDS MAJOR UPDATE

**Current state**: ONLY documents Alert API (470 lines, lines 1-470)

**ENTIRELY MISSING sections:**

**Edge API (5 endpoints, from `edgeapi/route.go` lines 27-54):**
| Method | Route | Description |
|--------|-------|-------------|
| GET | `/v1/workflow/rules/{ruleID}/edges` | List all edges for rule |
| GET | `/v1/workflow/rules/{ruleID}/edges/{edgeID}` | Get single edge |
| POST | `/v1/workflow/rules/{ruleID}/edges` | Create new edge |
| DELETE | `/v1/workflow/rules/{ruleID}/edges/{edgeID}` | Delete single edge |
| DELETE | `/v1/workflow/rules/{ruleID}/edges-all` | Delete all edges for rule |

**Cascade API (1 endpoint):**
| Method | Route | Description |
|--------|-------|-------------|
| GET | `/v1/workflow/rules/{id}/cascade-map` | Get downstream workflows |

**Simulate API (1 endpoint, from `simulate.go`):**
| Method | Route | Description |
|--------|-------|-------------|
| POST | `/v1/workflow/rules/{id}/simulate` | Test rule execution (dry run) |

**Missing request/response models (need full documentation):**
- `CreateEdgeRequest` (source_action_id, target_action_id, edge_type, edge_order)
- `EdgeResponse` (id, rule_id, source_action_id, target_action_id, edge_type, edge_order, created_date)
- `CascadeResponse` (rule_id, rule_name, actions[])
- `ActionCascadeInfo` (action_id, action_name, action_type, modifies_entity, triggers_event, modified_fields, downstream_workflows[])
- `DownstreamWorkflowInfo` (rule_id, rule_name, trigger_conditions, will_trigger_if)
- `SimulateRequest` and `SimulateResponse`

---

### 8. `actions/overview.md` - Action Handler Overview

**Status**: NEEDS UPDATE

**Currently lists**: 6 action types

**Specific line updates:**
- **Lines 7-14**: Add `evaluate_condition` row to Available Actions table (7th type)
- **Lines 21-25**: ActionHandler interface is OUTDATED - missing:
  ```go
  SupportsManualExecution() bool
  IsAsync() bool
  GetDescription() string
  ```
- **Lines 30-36**: Methods table needs 3 new rows for above methods
- **Lines 49-61**: ActionExecutionContext missing `TriggerSource` field

**Missing sections:**
- `EntityModifier` interface documentation (after line 131)
- `BranchTaken` field in `ActionResult` explanation
- Section on async vs sync actions (after line 100)
- Which actions implement `EntityModifier` (for cascade detection)
- How condition actions return `ConditionResult`

---

### 9. `configuration/rules.md` - Rule Configuration

**Status**: NEEDS UPDATE

**Missing models:**
- `ActionEdge` model documentation
- `NewActionEdge` model for API creation
- Edge type explanations
- Best practices for when to use edges vs execution_order

---

### 10. `testing.md` - Testing Guide

**Status**: NEEDS UPDATE

**Missing test file references:**
- `business/sdk/workflow/executor_graph_test.go` - Graph execution tests
- `business/sdk/workflow/workflowactions/control/condition_test.go` - Condition tests
- `api/cmd/services/ichor/tests/workflow/edgeapi/` - Edge API tests
- `api/cmd/services/ichor/tests/workflow/ruleapi/cascade_test.go` - Cascade tests

**Missing test patterns:**
- Testing graph-based workflows
- Testing condition evaluation
- Testing cascade visualization
- Testing edge CRUD operations

---

### 11. `configuration/triggers.md` - Trigger Configuration

**Status**: MINOR UPDATE

**Needed:**
- Cross-reference note that same operators are used in `evaluate_condition` action

---

### 12. `event-infrastructure.md` - Event Flow

**Status**: MINOR UPDATE

**Needed:**
- Note about how cascade visualization leverages the event system
- How condition node results flow through the system

---

### 13. `adding-domains.md` - Domain Integration

**Status**: MINOR UPDATE

**Needed:**
- Note about `EntityModifier` interface for actions that modify entities
- How new domains automatically work with cascade visualization

---

## Implementation Priorities

### Critical (Must Have)
1. **Create `actions/evaluate-condition.md`** - Users cannot use conditions without this
2. **Create `branching.md`** - Core feature needs explanation
3. **Update `api-reference.md`** - APIs unusable without endpoint docs
4. **Update `database-schema.md`** - Schema reference incomplete

### High Priority
5. **Update `architecture.md`** - System understanding incomplete
6. **Create `cascade-visualization.md`** - New feature needs docs
7. **Update `actions/overview.md`** - Action list incomplete

### Medium Priority
8. **Update `configuration/rules.md`** - Edge models need docs
9. **Update `README.md`** - Navigation incomplete
10. **Update `testing.md`** - Test patterns needed

### Low Priority
11. **Update `configuration/triggers.md`** - Cross-reference
12. **Update `event-infrastructure.md`** - Minor note
13. **Update `adding-domains.md`** - Minor note

---

## Files to Create

| File | Purpose | Est. Lines |
|------|---------|------------|
| `actions/evaluate-condition.md` | Condition action documentation | ~250 |
| `branching.md` | Graph-based execution guide | ~350 |
| `cascade-visualization.md` | Cascade detection feature | ~200 |

---

## Files to Update

| File | Changes Needed | Est. Changes |
|------|----------------|--------------|
| `README.md` | Navigation links, action list, operators | ~20 lines |
| `architecture.md` | Graph executor, EntityModifier, execution model | ~150 lines |
| `database-schema.md` | action_edges table, indices | ~60 lines |
| `api-reference.md` | 7 new endpoints, models | ~200 lines |
| `actions/overview.md` | New action, EntityModifier, interface updates | ~80 lines |
| `configuration/rules.md` | Edge models | ~40 lines |
| `testing.md` | New test patterns | ~80 lines |
| `configuration/triggers.md` | Cross-reference | ~5 lines |
| `event-infrastructure.md` | Cascade note | ~10 lines |
| `adding-domains.md` | EntityModifier note | ~10 lines |

---

## Source Code References (for documentation content)

| Feature | Source File | Key Lines |
|---------|-------------|-----------|
| `action_edges` table | `business/sdk/migrate/sql/migrate.sql` | v1.992, lines 1914-1936 |
| `evaluate_condition` handler | `business/sdk/workflow/workflowactions/control/condition.go` | Lines 1-303 |
| `ConditionConfig` struct | `condition.go` | Lines 38-48 |
| `ConditionResult` struct | `business/sdk/workflow/models.go` | Lines 416-422 |
| `ActionEdge` model | `models.go` | Lines 395-414 |
| `BranchTaken` field | `models.go` | Line 122 |
| `ExecuteRuleActionsGraph()` | `business/sdk/workflow/executor.go` | Lines 211-402 |
| `ShouldFollowEdge()` | `executor.go` | Lines 346-360 |
| `EntityModifier` interface | `business/sdk/workflow/interfaces.go` | Lines 133-145 |
| `GetEntityModifications()` impl | `workflowactions/data/updatefield.go` | Lines 447-461 |
| Edge API routes | `api/domain/http/workflow/edgeapi/route.go` | Lines 27-54 |
| Cascade API handler | `api/domain/http/workflow/ruleapi/cascade.go` | Lines 53-120 |
| Cascade response models | `cascade.go` | Lines 14-51 |
| Graph execution tests | `business/sdk/workflow/executor_graph_test.go` | Full file |
| Cascade tests | `api/cmd/services/ichor/tests/workflow/ruleapi/cascade_test.go` | Full file |
| Edge API tests | `api/cmd/services/ichor/tests/workflow/edgeapi/` | Full directory |

---

## Verification Steps

After documentation updates:
1. All links in README.md resolve correctly
2. All API endpoints documented match implementation
3. All configuration examples are valid JSON
4. All code snippets compile
5. Cross-references between docs are valid

---

## Notes

- No breaking changes to existing documentation
- Graph execution is backwards compatible (falls back to execution_order)
- `action_edges` table is the core schema enabling branching
- `EntityModifier` is optional - only entity-modifying actions need it
- Cascade visualization is read-only/informational
- All Phase 12 features have comprehensive test coverage
