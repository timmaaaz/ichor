# Trigger Uniqueness Analysis

## Question
Should we restrict to one trigger per entity/event type combination (e.g., one "on_create" for "orders") with multiple actions branching from that single trigger?

---

## Current State

### Database Schema (No Uniqueness Constraint)
```sql
-- workflow.automation_rules (lines 960-987 of migrate.sql)
CREATE TABLE workflow.automation_rules (
   id UUID PRIMARY KEY,
   entity_id UUID NOT NULL,
   trigger_type_id UUID NOT NULL REFERENCES workflow.trigger_types(id),
   trigger_conditions JSONB NULL,
   -- ... other fields
   -- NO UNIQUE constraint on (entity_id, trigger_type_id)
);
```

**Currently**: Multiple rules can be created for the same entity + trigger type combination.

### How Triggers Are Processed

From [trigger.go:461-472](business/sdk/workflow/trigger.go#L461-L472):
```go
func (tp *TriggerProcessor) getRulesForEntity(entityName string) []AutomationRuleView {
    rules := make([]AutomationRuleView, 0)
    for _, rule := range tp.activeRules {
        if rule.EntityName == entityName && rule.IsActive {
            rules = append(rules, rule)
        }
    }
    return rules
}
```

Then in `ProcessEvent()` (lines 121-156):
- Gets ALL rules for the entity
- Evaluates EACH rule against the event
- Returns ALL matched rules
- **All matching rules execute** (potentially in parallel if same execution_order)

### Trigger Condition Differentiation

Rules can have different `trigger_conditions` for the same entity/trigger type:

```json
// Rule 1: All orders
{ "entity": "orders", "trigger_type": "on_create", "trigger_conditions": null }

// Rule 2: High-value orders only
{ "entity": "orders", "trigger_type": "on_create",
  "trigger_conditions": { "field_conditions": [{ "field_name": "total", "operator": "greater_than", "value": 10000 }] } }

// Rule 3: Priority orders only
{ "entity": "orders", "trigger_type": "on_create",
  "trigger_conditions": { "field_conditions": [{ "field_name": "is_priority", "operator": "equals", "value": true }] } }
```

All three rules could fire for a high-value priority order.

---

## Proposed Change: One Trigger Per Entity/Event

### What This Would Mean

1. **Single Automation Rule** per (entity, trigger_type) combination
2. **All actions** for that trigger belong to that one rule
3. **Conditions** would move to action-level or use `evaluate_condition` branching

### Database Change Required

```sql
-- Add unique constraint
ALTER TABLE workflow.automation_rules
ADD CONSTRAINT unique_entity_trigger UNIQUE(entity_id, trigger_type_id);
```

### Migration Considerations

Existing data might have:
- Multiple rules for same entity/trigger (need to merge)
- Rules with overlapping trigger_conditions (need to convert to action-level conditions)

---

## Analysis: Pros and Cons

### Arguments FOR Single Trigger Per Entity/Event

| Benefit | Explanation |
|---------|-------------|
| **Simpler mental model** | Users see one "on_create for orders" workflow, not N scattered rules |
| **Easier to visualize** | One graph per trigger shows all possible paths |
| **No accidental duplicates** | Prevents creating duplicate rules that do the same thing |
| **Centralized management** | All "order created" logic in one place |
| **Clearer cascade visualization** | The cascade map feature is simpler when there's one rule per trigger |
| **Consistent with graph-based model** | Current graph execution already supports branching within a single rule |

### Arguments AGAINST Single Trigger Per Entity/Event

| Concern | Explanation |
|---------|-------------|
| **Loss of flexibility** | Can't have completely separate workflows for different use cases |
| **Complex rules become unwieldy** | One rule with 50 actions is harder to manage than 5 rules with 10 actions |
| **No conditional triggers** | Can't have "Rule A fires if total > $10k, Rule B fires if priority=true" at trigger level |
| **Template reuse harder** | Currently rules can reference templates; merging rules complicates this |
| **Breaking change** | Existing rules need migration; may not map cleanly |
| **Department ownership** | Finance team's order workflow vs Operations team's workflow can't be separate |

---

## Alternative Approaches

### Option A: Keep Current (Multiple Rules Allowed)

**Description**: No change - multiple rules per entity/trigger allowed

**Pros**: Maximum flexibility, no migration needed
**Cons**: Can lead to scattered, hard-to-understand configurations

### Option B: Strict Uniqueness (One Rule Per Trigger)

**Description**: Add `UNIQUE(entity_id, trigger_type_id)` constraint

**Pros**: Simple model, clear ownership
**Cons**: Forces all conditions into action-level branching, loses trigger-level filtering

### Option C: Uniqueness by Name/Purpose (Soft Constraint)

**Description**: No database constraint, but UI/API encourages consolidation
- Warning when creating rule for existing entity/trigger
- "View existing rules" button before creating new
- Lint/validation that flags potential duplicates

**Pros**: Flexibility preserved, guides users toward best practices
**Cons**: Doesn't prevent duplicates, requires UI work

### Option D: Hybrid - Unique Trigger + Trigger Conditions as Selector

**Description**:
- One rule per (entity, trigger_type) BUT...
- `trigger_conditions` becomes a **selector** for which actions fire
- Actions gain a `when` field that references condition sets

```json
{
  "entity": "orders",
  "trigger_type": "on_create",
  "condition_sets": {
    "high_value": { "field_conditions": [{ "field_name": "total", "operator": "greater_than", "value": 10000 }] },
    "priority": { "field_conditions": [{ "field_name": "is_priority", "operator": "equals", "value": true }] },
    "default": null
  },
  "actions": [
    { "name": "Notify VIP Team", "when": "high_value", ... },
    { "name": "Create Alert", "when": "priority", ... },
    { "name": "Send Confirmation", "when": "default", ... }
  ]
}
```

**Pros**: Single rule, but preserves conditional logic at trigger level
**Cons**: More complex schema, migration work

---

## Current Graph Execution Already Supports Branching

The existing `evaluate_condition` action + `action_edges` table supports conditional branching within a rule:

```
[on_create orders]
       |
       v
[evaluate_condition: total > 10000?]
       |
   /       \
  true     false
   |         |
   v         v
[VIP Alert] [Standard Email]
```

This means **Option B** (strict uniqueness) is viable because:
1. All conditional logic can move to `evaluate_condition` actions
2. Graph execution handles branching
3. Multiple "paths" within one rule replaces multiple rules

---

## Key Questions to Consider

1. **How many rules currently exist per entity/trigger?**
   - If most are 1:1 already, migration is easy
   - If many duplicates exist, need careful merging strategy

2. **Do teams need separate ownership?**
   - If Finance and Ops each "own" their order workflows, merging is political
   - If workflows are centrally managed, single rule is cleaner

3. **How complex are existing trigger_conditions?**
   - Simple conditions can easily become `evaluate_condition` actions
   - Complex nested conditions may be harder to migrate

4. **What does the UI workflow builder expect?**
   - If building a visual editor, single-rule-per-trigger is much simpler
   - Multiple rules require a "rule selector" before showing the graph

---

## Recommendation

**Start with Option C (Soft Constraint)** as a first step:
1. Add API validation that warns when creating duplicate entity/trigger rules
2. Add a "merge into existing rule" suggestion in the UI
3. Collect usage data on how many duplicates exist

**Then evaluate Option B (Strict Uniqueness)** based on:
- If usage shows most cases are 1:1, migrate to strict constraint
- If legitimate multi-rule patterns emerge, keep flexible model

This gives you data before committing to a breaking change.

---

## Files That Would Need Changes (If Implementing Option B)

| Layer | File | Change |
|-------|------|--------|
| **Migration** | `business/sdk/migrate/sql/migrate.sql` | Add UNIQUE constraint |
| **Business** | `business/sdk/workflow/trigger.go` | Logic unchanged (already handles single rule) |
| **API** | `api/domain/http/workflow/ruleapi/ruleapi.go` | Add validation on create |
| **App** | `app/domain/workflow/ruleapp/model.go` | Add validation error |
| **Tests** | `api/cmd/services/ichor/tests/workflow/ruleapi/` | Update tests for constraint |
| **Docs** | `docs/workflow/configuration/rules.md` | Document uniqueness requirement |

---

## Next Steps

1. **Query current data**: How many (entity, trigger_type) combinations have multiple rules?
2. **Review UI plans**: Does the workflow builder assume single or multiple rules?
3. **Discuss with stakeholders**: Is there a business need for separate rule ownership?
4. **Decide on approach**: Soft constraint vs hard constraint vs no change
