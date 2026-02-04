# Workflow Execution Tracking - Backend Enhancement Needs

**Created**: 2026-02-03
**Source**: Frontend Phase 13 Plan Review (Ichor Flow - Execution Order & History)
**Priority**: Future Enhancement (not blocking frontend MVP)

---

## Overview

During the plan review for Phase 13 of the Ichor Flow frontend (Execution Order & History), we identified several backend capabilities that would enhance the execution tracking system. The current backend implementation is sufficient for an MVP, but these enhancements would provide better debugging, performance analysis, and user experience.

---

## Current State (What Exists)

### Database Schema
- `workflow.automation_executions` table with:
  - Basic execution metadata (id, rule_id, status, timing)
  - `actions_executed` JSONB field containing array of `ActionResult` objects
  - `trigger_data` JSONB for trigger event information
  - Support for both automated and manual executions

### API Endpoints
- `GET /v1/workflow/executions` - List with pagination, filtering, sorting
- `GET /v1/workflow/executions/{id}` - Single execution with action results

### Data Tracking
- Per-action results: action_id, action_name, action_type, status, duration_ms, started_at, completed_at
- Condition branch tracking: `branch_taken` field ("true_branch" or "false_branch")
- Error tracking at execution and per-action level
- Manual vs. automation distinction via `trigger_source`

---

## Enhancement Requests

### 1. Separate Execution Steps Table (Medium Priority)

**Current Limitation**: Action results are stored as JSONB array in `actions_executed` column. This works but limits queryability.

**Proposed Enhancement**:
```sql
CREATE TABLE workflow.execution_steps (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    execution_id UUID NOT NULL REFERENCES workflow.automation_executions(id) ON DELETE CASCADE,
    action_id UUID REFERENCES workflow.rule_actions(id),
    step_order INTEGER NOT NULL,
    status VARCHAR(20) NOT NULL,
    started_at TIMESTAMP,
    completed_at TIMESTAMP,
    duration_ms INTEGER,
    input JSONB,
    output JSONB,
    error_message TEXT,
    branch_taken VARCHAR(20),
    created_date TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_execution_steps_execution_id ON workflow.execution_steps(execution_id);
CREATE INDEX idx_execution_steps_action_id ON workflow.execution_steps(action_id);
CREATE INDEX idx_execution_steps_status ON workflow.execution_steps(status);
```

**Benefits**:
- Query individual steps across all executions
- Better analytics (e.g., "which action fails most often?")
- Smaller payload for execution list (don't need to return all steps)
- Easier step-by-step timeline queries

**API Addition**:
```
GET /v1/workflow/executions/{id}/steps
```

---

### 2. Parallel Execution Batch Tracking (Low Priority)

**Current Limitation**: The executor processes actions in parallel batches, but batch grouping information is not persisted. Frontend cannot visualize which actions ran in parallel vs. sequential.

**Proposed Enhancement**:
```sql
CREATE TABLE workflow.execution_batches (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    execution_id UUID NOT NULL REFERENCES workflow.automation_executions(id) ON DELETE CASCADE,
    batch_order INTEGER NOT NULL,
    status VARCHAR(20) NOT NULL,
    started_at TIMESTAMP,
    completed_at TIMESTAMP,
    duration_ms INTEGER
);

-- Add batch_id to execution_steps
ALTER TABLE workflow.execution_steps ADD COLUMN batch_id UUID REFERENCES workflow.execution_batches(id);
```

**Alternative**: Add `batch_order` field to `ActionResult` in the JSONB:
```json
{
  "action_id": "...",
  "batch_order": 1,  // Actions with same batch_order ran in parallel
  ...
}
```

**Benefits**:
- Visualize parallel vs. sequential execution
- Performance debugging (identify slow batches)
- Better understanding of workflow execution flow

---

### 3. Per-Rule Execution Endpoint (Low Priority)

**Current Limitation**: To get executions for a specific rule, frontend must call the general executions endpoint with a filter. This works but could be optimized.

**Proposed Enhancement**:
```
GET /v1/workflow/rules/{rule_id}/executions
```

**Benefits**:
- Cleaner API for rule-specific history
- Potentially better query optimization
- More intuitive API design

**Workaround**: Frontend can use existing `GET /v1/workflow/executions?rule_id={id}` - this is sufficient for MVP.

---

### 4. Step Input/Output Tracking (Low Priority)

**Current Limitation**: The `ActionResult` structure has `result_data` but not explicit `input_data`. For debugging, it would be helpful to see what data was passed to each action.

**Proposed Enhancement**:
Add to `ActionResult`:
```go
type ActionResult struct {
    // ... existing fields
    InputData    map[string]interface{} `json:"input_data,omitempty"`   // NEW
    ResultData   map[string]interface{} `json:"result_data,omitempty"`
}
```

**Benefits**:
- Full input/output visibility for debugging
- Trace data flow through workflow
- Easier reproduction of issues

**Consideration**: Input data may be large. Consider:
- Truncation for large payloads
- Optional flag to include/exclude input data
- Separate endpoint for full step details

---

### 5. Real-Time Execution Updates (Future)

**Current Limitation**: Execution status is only available after completion. For long-running workflows, users cannot see progress.

**Proposed Enhancement Options**:

**Option A: WebSocket Updates**
- Extend existing alert WebSocket to include execution progress
- Push step completion events as they occur

**Option B: Polling-Friendly Status**
- Add `current_step` field to execution record
- Update as each step completes
- Frontend polls and shows progress

**Benefits**:
- Better UX for long-running workflows
- Real-time debugging capability
- User confidence (seeing progress)

**Note**: This is a larger undertaking and should be a separate phase if pursued.

---

## Priority Summary

| Enhancement | Priority | Effort | MVP Blocker? |
|-------------|----------|--------|--------------|
| Separate execution_steps table | Medium | Medium | No |
| Parallel batch tracking | Low | Low | No |
| Per-rule execution endpoint | Low | Low | No |
| Step input/output tracking | Low | Low | No |
| Real-time execution updates | Future | High | No |

---

## Current Workarounds (Frontend)

The frontend Phase 13 implementation will use these workarounds:

1. **No batch visualization** - Timeline shows actions sequentially based on `started_at` timestamps
2. **Parse JSONB for steps** - Extract step data from `actions_executed` array
3. **Filter client-side for per-rule** - Use general endpoint with `rule_id` filter parameter
4. **Polling for updates** - Poll every 3 seconds when executions are running

---

## Related Files

**Backend (this repo)**:
- `business/sdk/workflow/models.go` - AutomationExecution, ActionResult structs
- `business/sdk/workflow/executor.go` - Execution logic
- `business/sdk/workflow/workflowbus.go` - Business layer methods
- `api/domain/http/workflow/executionapi/` - API handlers
- `business/sdk/migrate/sql/migrate.sql` - Schema definitions

**Frontend (vue/ichor)**:
- `.claude/plans/COMPONENT_ARCHITECTURE_PLAN/phases/PHASE_13_EXECUTION_ORDER_HISTORY.md`
- `.claude/plans/COMPONENT_ARCHITECTURE_PLAN/PROGRESS.yaml`

---

## Notes

- These enhancements are documented for future planning, not immediate implementation
- The current backend implementation is sufficient for the frontend MVP
- Batch tracking would require changes to the executor to persist batch groupings
- Real-time updates would be a significant undertaking requiring WebSocket infrastructure changes
