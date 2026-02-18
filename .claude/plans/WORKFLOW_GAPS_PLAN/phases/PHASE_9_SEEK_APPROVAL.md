# Phase 5: Implement seek_approval

**Category**: Backend + Database
**Status**: Pending
**Dependencies**: None (but Phase 1 should be done first if approvals need to update entity status)
**Effort**: High

---

## Overview

`seek_approval` is the most complex stub. Currently:
- `Execute()` returns a fake `approval_id` with `"output": "approved"` hardcoded — the workflow always takes the approved branch
- No approval record is persisted
- No approvers are notified
- No mechanism exists for approvers to submit decisions

The full implementation requires:
1. A new `workflow.approval_requests` database table
2. A `approvalrequestbus` business package
3. `SeekApprovalHandler.StartAsync()` implementation (pauses Temporal workflow)
4. An API endpoint for approvers to submit decisions (which completes the Temporal activity)

---

## Goals

1. Persist approval requests to DB so they survive restarts
2. Notify approvers when a request is created (via `create_alert`)
3. Pause the Temporal workflow until an approver responds
4. Expose an API endpoint for approvers to resolve requests

---

## Task Breakdown

### Task 1: Create workflow.approval_requests Migration

**File**: `business/sdk/migrate/sql/migrate.sql`

Append a new version:
```sql
-- Version: X.XX
-- Description: Add workflow approval requests table

CREATE TABLE workflow.approval_requests (
    approval_request_id UUID NOT NULL,
    execution_id        UUID NOT NULL REFERENCES workflow.automation_executions(automation_execution_id),
    rule_id             UUID NOT NULL REFERENCES workflow.automation_rules(automation_rule_id),
    action_name         VARCHAR(100) NOT NULL,
    approvers           UUID[] NOT NULL,
    approval_type       VARCHAR(20) NOT NULL CHECK (approval_type IN ('any', 'all', 'majority')),
    status              VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'approved', 'rejected', 'timed_out', 'expired')),
    timeout_hours       INT NOT NULL DEFAULT 72,
    task_token          TEXT,  -- Temporal task token for async completion (base64)
    approval_message    TEXT,
    resolved_by         UUID REFERENCES core.users(user_id),
    resolution_reason   TEXT,
    created_date        TIMESTAMP NOT NULL DEFAULT NOW(),
    resolved_date       TIMESTAMP,

    PRIMARY KEY (approval_request_id)
);

CREATE INDEX idx_approval_requests_execution ON workflow.approval_requests(execution_id);
CREATE INDEX idx_approval_requests_status ON workflow.approval_requests(status) WHERE status = 'pending';
```

Check the current highest version in migrate.sql before writing the version number.

### Task 2: Create approvalrequestbus

**New directory**: `business/domain/workflow/approvalrequestbus/`

**model.go**:
```go
type ApprovalRequest struct {
    ApprovalRequestID uuid.UUID
    ExecutionID       uuid.UUID
    RuleID            uuid.UUID
    ActionName        string
    Approvers         []uuid.UUID
    ApprovalType      string  // "any" | "all" | "majority"
    Status            string  // "pending" | "approved" | "rejected" | "timed_out"
    TimeoutHours      int
    TaskToken         string  // base64-encoded Temporal task token
    ApprovalMessage   string
    ResolvedBy        uuid.UUID
    ResolutionReason  string
    CreatedDate       time.Time
    ResolvedDate      time.Time
}

type NewApprovalRequest struct {
    ApprovalRequestID uuid.UUID
    ExecutionID       uuid.UUID
    RuleID            uuid.UUID
    ActionName        string
    Approvers         []uuid.UUID
    ApprovalType      string
    TimeoutHours      int
    TaskToken         string
    ApprovalMessage   string
}

type UpdateApprovalRequest struct {
    Status           *string
    ResolvedBy       *uuid.UUID
    ResolutionReason *string
    ResolvedDate     *time.Time
}
```

**approvalrequestbus.go**: Standard Ardan Labs Business struct with Create, QueryByID, Update.

**DB store**: `stores/approvalrequestdb/` with standard sqlx query patterns. Use `pq.Array()` for the `approvers UUID[]` column.

### Task 3: Implement SeekApprovalHandler.StartAsync

`StartAsync` implements `AsyncActivityHandler`. Temporal calls this when the activity runs; the method should initiate the approval but NOT complete the task token. The workflow remains paused until `AsyncCompleter.Complete()` is called.

**File**: `business/sdk/workflow/workflowactions/approval/seek.go`

The handler needs new dependencies:

```go
type SeekApprovalHandler struct {
    log                *logger.Logger
    db                 *sqlx.DB
    approvalRequestBus *approvalrequestbus.Business
    alertBus           *alertbus.Business  // for notifying approvers
    asyncCompleter     temporal.AsyncCompleter // for timeout expiry (optional, future)
}
```

`StartAsync` implementation:

```go
func (h *SeekApprovalHandler) StartAsync(ctx context.Context, config json.RawMessage, execCtx workflow.ActionExecutionContext, taskToken []byte) error {
    var cfg struct {
        Approvers       []string `json:"approvers"`
        ApprovalType    string   `json:"approval_type"`
        TimeoutHours    int      `json:"timeout_hours"`
        ApprovalMessage string   `json:"approval_message"`
    }
    if err := json.Unmarshal(config, &cfg); err != nil {
        return fmt.Errorf("parse seek_approval config: %w", err)
    }

    if cfg.TimeoutHours == 0 {
        cfg.TimeoutHours = 72
    }

    // Parse approver UUIDs
    var approvers []uuid.UUID
    for _, a := range cfg.Approvers {
        uid, err := uuid.Parse(a)
        if err != nil {
            return fmt.Errorf("invalid approver UUID %q: %w", a, err)
        }
        approvers = append(approvers, uid)
    }

    ruleID := uuid.Nil
    if execCtx.RuleID != nil {
        ruleID = *execCtx.RuleID
    }

    // Persist approval request
    req := approvalrequestbus.NewApprovalRequest{
        ApprovalRequestID: uuid.New(),
        ExecutionID:       execCtx.ExecutionID,
        RuleID:            ruleID,
        ActionName:        execCtx.ActionName, // need to add ActionName to ActionExecutionContext
        Approvers:         approvers,
        ApprovalType:      cfg.ApprovalType,
        TimeoutHours:      cfg.TimeoutHours,
        TaskToken:         base64.StdEncoding.EncodeToString(taskToken),
        ApprovalMessage:   cfg.ApprovalMessage,
    }
    if err := h.approvalRequestBus.Create(ctx, req); err != nil {
        return fmt.Errorf("create approval request: %w", err)
    }

    // Notify each approver via create_alert
    if h.alertBus != nil {
        for _, approverID := range approvers {
            alert := alertbus.Alert{
                ID:               uuid.New(),
                AlertType:        "approval_request",
                Severity:         alertbus.SeverityHigh,
                Title:            fmt.Sprintf("Approval Required: %s", execCtx.RuleName),
                Message:          fmt.Sprintf("Your approval is required for %s. Request ID: %s", cfg.ApprovalMessage, req.ApprovalRequestID),
                SourceEntityID:   execCtx.EntityID,
                SourceEntityName: execCtx.EntityName,
                SourceRuleID:     ruleID,
                Status:           alertbus.StatusActive,
                CreatedDate:      time.Now(),
                UpdatedDate:      time.Now(),
            }
            if err := h.alertBus.Create(ctx, alert); err != nil {
                h.log.Error(ctx, "failed to create approval alert", "approver", approverID, "error", err)
            }
        }
    }

    h.log.Info(ctx, "seek_approval: approval request created",
        "approval_id", req.ApprovalRequestID,
        "approvers", len(approvers),
        "timeout_hours", cfg.TimeoutHours)

    return nil
}
```

**Register in `humanActionTypes` map** in `business/sdk/workflow/temporal/workflow.go`:
```go
var humanActionTypes = map[string]bool{
    "seek_approval": true,
}
```

Check the current `humanActionTypes` definition — it may already include `seek_approval`.

### Task 4: Add Approval Resolution API Endpoint

**New package**: `api/domain/http/workflow/approvalapi/`

**approvalapi.go** — handler for resolving an approval:

```go
// POST /workflow/approvals/{approval_request_id}/resolve
// Body: {"resolution": "approved"|"rejected", "reason": "optional reason"}
func (h *Handlers) Resolve(ctx context.Context, r *web.Request) web.Encoder {
    approvalID, err := uuid.Parse(web.Param(r, "approval_request_id"))
    // ... parse body ...
    // 1. Load approval request (verify it's pending)
    // 2. Update status in DB
    // 3. Decode task token from base64
    // 4. Call asyncCompleter.Complete(ctx, taskToken, map{"output": resolution})
    // 5. Return 200
}
```

The `AsyncCompleter` is defined in `business/sdk/workflow/temporal/async_completer.go`. Wire it through the handler config.

**route.go**: Mount at `POST /workflow/approvals/{approval_request_id}/resolve`

**Wire in all.go**: Instantiate approvalrequestbus, wire AsyncCompleter.

---

## Validation

```bash
go build ./...

# Verify approval_requests table created
make migrate && psql -c "\d workflow.approval_requests"

# Integration test: trigger a workflow with seek_approval, verify:
# 1. approval_request created in DB
# 2. Alert created for approver
# 3. Temporal workflow paused (execution stays in_progress)
# 4. POST /workflow/approvals/{id}/resolve resumes workflow
# 5. approved/rejected branches taken correctly
```

---

## Gotchas

- **`ActionName` in `ActionExecutionContext`** — the context may not currently have an `ActionName` field. Check `business/sdk/workflow/models.go`. If missing, add it when building `ActionActivityInput` in `activities.go`.
- **Task token is `[]byte` (binary)** — base64-encode before storing in DB TEXT column. Decode before calling `AsyncCompleter.Complete()`.
- **`majority` approval type** — requires tracking individual votes, not just a single decision. For initial implementation, treat `majority` like `any` (first approval wins). Document this simplification.
- **Timeout handling** — the `timeout_hours` config sets when the approval expires, but automatic expiry requires a separate Temporal timer or cron job. For initial implementation, approvals don't automatically time out — a future phase can add the expiry timer.
- **`humanActionTypes` vs `asyncActionTypes`** — check which map `seek_approval` belongs to in `workflow.go`. "Human" actions use `ExecuteAsyncActionActivity` with `StartAsync`; "async" actions use a different path.
