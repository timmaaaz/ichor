# Phase 3: Delegate Event Fix for `approvalrequestbus.Resolve()` [standalone bug fix]

## REVISED SCOPE

The `resolve_approval_request` workflow action handler has been **deferred**. It is only
justified for meta-orchestration (a second workflow resolving an approval that a first workflow
is waiting on). No such workflow exists today, and adding the handler speculatively risks
misuse — particularly the deadlock case where seek_approval and resolve_approval_request appear
in the same workflow.

**Part B (the handler) is deferred. Part A (the delegate event) remains — it is a standalone bug fix.**

## Objective

Fix the missing `delegate.Call()` in `approvalrequestbus.Business.Resolve()`. This is the only
method across all 51 bus domains that changes entity state without firing a delegate event.
Reactive workflows cannot currently trigger on approval resolution. This is a correctness gap
independent of whether a handler ever exists.

## Why Part A Is a Bug Fix

The pattern across all 51 entities: `delegate.Call()` fires on every state change. `approvalrequestbus.Resolve()` changes status (pending → approved/rejected/timed_out) but silently skips the delegate call. This means:
- Workflows CANNOT trigger off approval resolution
- Audit systems CANNOT observe resolutions
- The event bus has a blind spot for one of the most business-critical state transitions

## Part A: Add Delegate Event to Resolve()

**File**: `business/domain/workflow/approvalrequestbus/approvalrequestbus.go`

Pre-work: Read the file first. Find the `Resolve()` method. Confirm it does NOT call `delegate.Call()`.

Change: At the end of `Resolve()`, after updating the database record, add:
```go
if err := abus.delegate.Call(ctx, ActionUpdatedData(updatedRequest)); err != nil {
    return ApprovalRequest{}, fmt.Errorf("delegate[approval_request_updated]: %w", err)
}
```

This follows the identical pattern used in every other bus `Update()` method. No new model
types needed — `ActionUpdatedData` already exists in the package's `event.go`.

## Part B: `resolve_approval_request` Handler

**File**: `business/sdk/workflow/workflowactions/approval/resolve.go`

```
Package: approval

Handler struct: ResolveApprovalHandler
  - log *logger.Logger
  - db *sqlx.DB
  - approvalRequestBus *approvalrequestbus.Business

Config struct: ResolveApprovalConfig
  - ApprovalRequestID string `json:"approval_request_id"` (required)
  - Resolution string `json:"resolution"` (required: "approved" | "rejected")
  - Reason string `json:"reason,omitempty"` (audit trail — recommended)

Methods:
  - GetType() → "resolve_approval_request"
  - IsAsync() → true  (approval resolution triggers downstream effects — run async)
  - SupportsManualExecution() → true
  - GetDescription() → "Programmatically resolve an open approval request (approve or reject)"
  - Validate() — validate UUID, check resolution is "approved" or "rejected"
  - GetOutputPorts() → resolved_approved, resolved_rejected, not_found, already_resolved, failure
  - GetEntityModifications() → workflow.approval_requests (on_update, fields: [status])

Execute logic:
  1. Parse + validate config
  2. QueryByID → not_found if missing
  3. If request.Status != "pending" → already_resolved with current status
  4. Call bus.Resolve(ctx, id, execCtx.UserID, resolution, reason)
  5. Return resolved_approved or resolved_rejected based on resolution
  6. Error → failure
```

**Note on IsAsync**: set to `true` because Temporal's async activity mechanism is more appropriate
here — approval resolution typically triggers downstream workflow continuation (the `seek_approval`
handler is waiting on a Temporal signal). Using async avoids nested synchronous activity deadlock.

## Registration

**File**: `business/sdk/workflow/workflowactions/register.go`

Add to `RegisterAll()` (replace the existing seek_approval registration block):
```go
// Approval actions
registry.Register(approval.NewSeekApprovalHandler(config.Log, config.DB, config.Buses.ApprovalRequest, config.Buses.Alert))
registry.Register(approval.NewResolveApprovalHandler(config.Log, config.DB, config.Buses.ApprovalRequest))
```

`ApprovalRequest *approvalrequestbus.Business` already exists in BusDependencies ✓

## RegisterCoreActions Note

`RegisterCoreActions` uses `nil` buses for graceful degradation. Add a nil-guarded registration:
```go
registry.Register(approval.NewResolveApprovalHandler(log, db, nil))
```
The handler must handle `nil` approvalRequestBus by returning an immediate error in Execute().

## Testing

**Unit tests**: `business/sdk/workflow/workflowactions/approval/resolve_test.go`
- Config validation: missing ID, invalid UUID, invalid resolution value
- Execute routing: not_found, already_resolved, resolved_approved, resolved_rejected

**Integration test for Part A** (delegate event):
- After calling `approvalrequestbus.Resolve()`, verify a delegate event is captured
- Use the test delegate pattern established in other bus tests

## Verification

```bash
go build ./business/domain/workflow/approvalrequestbus/...
go build ./business/sdk/workflow/workflowactions/...
go build ./api/cmd/services/ichor/...
go test ./business/domain/workflow/approvalrequestbus/...
go test ./business/sdk/workflow/workflowactions/approval/...
```

## Definition of Done

- [ ] `delegate.Call()` added to `approvalrequestbus.Resolve()`
- [ ] `resolve_approval_request` handler file created
- [ ] Handler registered in RegisterAll() and RegisterCoreActions()
- [ ] Unit tests for resolve handler
- [ ] Test that Resolve() now fires delegate event
- [ ] `go build` passes on all affected packages
- [ ] `GET /v1/workflow/action-types` returns `resolve_approval_request`
