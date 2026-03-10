# Code Review Fixes — PR #87 Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Fix 9 non-false-positive issues found during code review of PR #87 (workflow action handlers + audit trail).

**Architecture:** Fixes span the bus layer (error handling, constants), workflow action handlers (race protection), app layer (reason field threading, UpdatePurchaseOrder cleanup), and docs (README catalog update). No new packages. No schema changes.

**Tech Stack:** Go 1.23, Temporal workflow handlers, Ardan Labs layer architecture.

---

## Task 1: Export status constants in `transferorderbus` and `inventoryadjustmentbus`

**Why:** Handler files use raw string literals (`"approved"`, `"rejected"`, `"pending"`) that silently diverge if canonical values change. Issue #10.

**Files:**
- Modify: `business/domain/inventory/transferorderbus/transferorderbus.go:18-25`
- Modify: `business/domain/inventory/inventoryadjustmentbus/inventoryadjustmentbus.go:18-25`

**Step 1: Add status constants to `transferorderbus.go`**

After the `var (...)` error block (line 25), add:

```go
// Transfer order status values.
const (
	StatusPending  = "pending"
	StatusApproved = "approved"
	StatusRejected = "rejected"
)
```

**Step 2: Update the string literals in `transferorderbus.go` to use the constants**

In `Approve()` (line 206):
```go
// Before:
if to.Status == "approved" || to.Status == "rejected" {
// After:
if to.Status == StatusApproved || to.Status == StatusRejected {
```

In `Reject()` (line 234), same change.

**Step 3: Add status constants to `inventoryadjustmentbus.go`**

After the `var (...)` error block (line 25), add:

```go
// Inventory adjustment approval status values.
const (
	ApprovalStatusPending  = "pending"
	ApprovalStatusApproved = "approved"
	ApprovalStatusRejected = "rejected"
)
```

**Step 4: Update string literals in `inventoryadjustmentbus.go` to use the constants**

In `Approve()` (line 186):
```go
// Before:
if ia.ApprovalStatus != "pending" {
// After:
if ia.ApprovalStatus != ApprovalStatusPending {
```

In `Reject()` (line 214), same change.

Also update where status is assigned:
```go
// In Approve():
ia.ApprovalStatus = ApprovalStatusApproved
// In Reject():
ia.ApprovalStatus = ApprovalStatusRejected
```

**Step 5: Build**

```bash
go build ./business/domain/inventory/transferorderbus/... ./business/domain/inventory/inventoryadjustmentbus/...
```
Expected: no errors.

**Step 6: Commit**

```bash
git add business/domain/inventory/transferorderbus/transferorderbus.go business/domain/inventory/inventoryadjustmentbus/inventoryadjustmentbus.go
git commit -m "refactor(bus): export status constants for transfer order and inventory adjustment"
```

---

## Task 2: Update inventory action handlers to use status constants

**Why:** Removes string literals from 4 handler files. Part of Issue #10.

**Files:**
- Modify: `business/sdk/workflow/workflowactions/inventory/approve_adjustment.go:99-103`
- Modify: `business/sdk/workflow/workflowactions/inventory/reject_adjustment.go:102-106`
- Modify: `business/sdk/workflow/workflowactions/inventory/approve_transfer_order.go:99-103`
- Modify: `business/sdk/workflow/workflowactions/inventory/reject_transfer_order.go:102-106`

**Step 1: Update `approve_adjustment.go` switch**

```go
// Before (lines 99-103):
switch ia.ApprovalStatus {
case "approved":
    return map[string]any{"output": "already_approved", "adjustment_id": cfg.AdjustmentID}, nil
case "rejected":
    return map[string]any{"output": "already_rejected", "adjustment_id": cfg.AdjustmentID}, nil
}

// After:
switch ia.ApprovalStatus {
case inventoryadjustmentbus.ApprovalStatusApproved:
    return map[string]any{"output": "already_approved", "adjustment_id": cfg.AdjustmentID}, nil
case inventoryadjustmentbus.ApprovalStatusRejected:
    return map[string]any{"output": "already_rejected", "adjustment_id": cfg.AdjustmentID}, nil
}
```

**Step 2: Update `reject_adjustment.go` switch** — same pattern, same replacement.

**Step 3: Update `approve_transfer_order.go` switch**

```go
// Before (lines 99-103):
switch to.Status {
case "approved":
    return map[string]any{"output": "already_approved", "transfer_order_id": cfg.TransferOrderID}, nil
case "rejected":
    return map[string]any{"output": "already_rejected", "transfer_order_id": cfg.TransferOrderID}, nil
}

// After:
switch to.Status {
case transferorderbus.StatusApproved:
    return map[string]any{"output": "already_approved", "transfer_order_id": cfg.TransferOrderID}, nil
case transferorderbus.StatusRejected:
    return map[string]any{"output": "already_rejected", "transfer_order_id": cfg.TransferOrderID}, nil
}
```

**Step 4: Update `reject_transfer_order.go` switch** — same pattern.

**Step 5: Build**

```bash
go build ./business/sdk/workflow/workflowactions/...
```
Expected: no errors.

**Step 6: Commit**

```bash
git add business/sdk/workflow/workflowactions/inventory/
git commit -m "refactor(workflowactions): replace status string literals with exported bus constants"
```

---

## Task 3: Add race-condition protection to inventory adjustment handlers

**Why:** If status changes between QueryByID and Approve/Reject (race), the bus returns `ErrInvalidApprovalStatus` which currently propagates as a Temporal retry instead of clean output. Issues #1, #2.

**Files:**
- Modify: `business/sdk/workflow/workflowactions/inventory/approve_adjustment.go:106-109`
- Modify: `business/sdk/workflow/workflowactions/inventory/reject_adjustment.go:109-112`

**Step 1: Update error handling in `approve_adjustment.go`**

```go
// Before (lines 106-109):
approved, err := h.inventoryAdjustmentBus.Approve(ctx, ia, execCtx.UserID, cfg.ApprovalReason)
if err != nil {
    return nil, fmt.Errorf("approve adjustment: %w", err)
}

// After:
approved, err := h.inventoryAdjustmentBus.Approve(ctx, ia, execCtx.UserID, cfg.ApprovalReason)
if err != nil {
    if errors.Is(err, inventoryadjustmentbus.ErrInvalidApprovalStatus) {
        return map[string]any{"output": "failure", "error": "adjustment status changed concurrently"}, nil
    }
    return nil, fmt.Errorf("approve adjustment: %w", err)
}
```

**Step 2: Update error handling in `reject_adjustment.go`**

```go
// Before (lines 109-112):
rejected, err := h.inventoryAdjustmentBus.Reject(ctx, ia, execCtx.UserID, cfg.RejectionReason)
if err != nil {
    return nil, fmt.Errorf("reject adjustment: %w", err)
}

// After:
rejected, err := h.inventoryAdjustmentBus.Reject(ctx, ia, execCtx.UserID, cfg.RejectionReason)
if err != nil {
    if errors.Is(err, inventoryadjustmentbus.ErrInvalidApprovalStatus) {
        return map[string]any{"output": "failure", "error": "adjustment status changed concurrently"}, nil
    }
    return nil, fmt.Errorf("reject adjustment: %w", err)
}
```

**Step 3: Build**

```bash
go build ./business/sdk/workflow/workflowactions/...
```
Expected: no errors.

**Step 4: Commit**

```bash
git add business/sdk/workflow/workflowactions/inventory/approve_adjustment.go business/sdk/workflow/workflowactions/inventory/reject_adjustment.go
git commit -m "fix(workflowactions): route ErrInvalidApprovalStatus to failure port instead of Temporal retry"
```

---

## Task 4: Add race-condition protection to transfer order handlers

**Why:** Same pattern as Task 3 but for transfer orders. Issues #1 (transfer order variant).

**Files:**
- Modify: `business/sdk/workflow/workflowactions/inventory/approve_transfer_order.go:106-109`
- Modify: `business/sdk/workflow/workflowactions/inventory/reject_transfer_order.go:109-112`

**Step 1: Update error handling in `approve_transfer_order.go`**

```go
// Before (lines 106-109):
approved, err := h.transferOrderBus.Approve(ctx, to, execCtx.UserID, cfg.ApprovalReason)
if err != nil {
    return nil, fmt.Errorf("approve transfer order: %w", err)
}

// After:
approved, err := h.transferOrderBus.Approve(ctx, to, execCtx.UserID, cfg.ApprovalReason)
if err != nil {
    if errors.Is(err, transferorderbus.ErrInvalidTransferStatus) {
        return map[string]any{"output": "failure", "error": "transfer order status changed concurrently"}, nil
    }
    return nil, fmt.Errorf("approve transfer order: %w", err)
}
```

**Step 2: Update error handling in `reject_transfer_order.go`**

```go
// Before (lines 109-112):
rejected, err := h.transferOrderBus.Reject(ctx, to, execCtx.UserID, cfg.RejectionReason)
if err != nil {
    return nil, fmt.Errorf("reject transfer order: %w", err)
}

// After:
rejected, err := h.transferOrderBus.Reject(ctx, to, execCtx.UserID, cfg.RejectionReason)
if err != nil {
    if errors.Is(err, transferorderbus.ErrInvalidTransferStatus) {
        return map[string]any{"output": "failure", "error": "transfer order status changed concurrently"}, nil
    }
    return nil, fmt.Errorf("reject transfer order: %w", err)
}
```

**Step 3: Build**

```bash
go build ./business/sdk/workflow/workflowactions/...
```
Expected: no errors.

**Step 4: Commit**

```bash
git add business/sdk/workflow/workflowactions/inventory/approve_transfer_order.go business/sdk/workflow/workflowactions/inventory/reject_transfer_order.go
git commit -m "fix(workflowactions): route ErrInvalidTransferStatus to failure port instead of Temporal retry"
```

---

## Task 5: Fix `purchaseorderbus.Approve()` missing RejectedBy guard + `approve_po.go` dead code

**Why:** `purchaseorderbus.Approve()` does not check `RejectedBy`, so a concurrent rejection between QueryByID and the Approve call silently overwrites a rejected PO with approval. Issues #3, #5.

**Fix strategy:** Add `ErrAlreadyRejected` check to `Approve()`, then add the matching post-call catch to `approve_po.go`. The pre-flight guards remain (they short-circuit for the common path and avoid a write); the post-call catches now handle the race window.

**Files:**
- Modify: `business/domain/procurement/purchaseorderbus/purchaseorderbus.go:261-289`
- Modify: `business/sdk/workflow/workflowactions/procurement/approve_po.go:108-114`

**Step 1: Add `RejectedBy` guard to `purchaseorderbus.Approve()`**

```go
// Before (lines 266-268):
if po.ApprovedBy != nil {
    return PurchaseOrder{}, fmt.Errorf("approve: %w", ErrAlreadyApproved)
}

// After:
if po.ApprovedBy != nil {
    return PurchaseOrder{}, fmt.Errorf("approve: %w", ErrAlreadyApproved)
}
if po.RejectedBy != nil {
    return PurchaseOrder{}, fmt.Errorf("approve: %w", ErrAlreadyRejected)
}
```

**Step 2: Add `ErrAlreadyRejected` post-call catch to `approve_po.go`**

```go
// Before (lines 108-114):
approved, err := h.purchaseOrderBus.Approve(ctx, po, execCtx.UserID, cfg.ApprovalReason)
if err != nil {
    if errors.Is(err, purchaseorderbus.ErrAlreadyApproved) {
        return map[string]any{"output": "already_approved", "purchase_order_id": cfg.PurchaseOrderID}, nil
    }
    return nil, fmt.Errorf("approve purchase order: %w", err)
}

// After:
approved, err := h.purchaseOrderBus.Approve(ctx, po, execCtx.UserID, cfg.ApprovalReason)
if err != nil {
    if errors.Is(err, purchaseorderbus.ErrAlreadyApproved) {
        return map[string]any{"output": "already_approved", "purchase_order_id": cfg.PurchaseOrderID}, nil
    }
    if errors.Is(err, purchaseorderbus.ErrAlreadyRejected) {
        return map[string]any{"output": "already_rejected", "purchase_order_id": cfg.PurchaseOrderID}, nil
    }
    return nil, fmt.Errorf("approve purchase order: %w", err)
}
```

**Step 3: Build**

```bash
go build ./business/domain/procurement/purchaseorderbus/... ./business/sdk/workflow/workflowactions/...
```
Expected: no errors.

**Step 4: Run existing purchase order bus tests**

```bash
go test ./business/domain/procurement/purchaseorderbus/...
```
Expected: PASS (existing tests unaffected — the new guard only blocks a previously-allowed invalid state).

**Step 5: Commit**

```bash
git add business/domain/procurement/purchaseorderbus/purchaseorderbus.go business/sdk/workflow/workflowactions/procurement/approve_po.go
git commit -m "fix(purchaseorderbus): guard Approve() against already-rejected POs; add matching handler catch"
```

---

## Task 6: Thread reason field through HTTP approve/reject for purchase orders

**Why:** `ApproveRequest` and `RejectRequest` have no reason field, so the HTTP endpoints silently pass `""` to the bus even though the DB column and workflow handlers support reasons. Issue #4.

**Files:**
- Modify: `app/domain/procurement/purchaseorderapp/model.go:523-557`
- Modify: `app/domain/procurement/purchaseorderapp/purchaseorderapp.go:156-196`
- Modify: `api/domain/http/procurement/purchaseorderapi/purchaseorderapi.go:117-165`

**Step 1: Add reason fields to `ApproveRequest` and `RejectRequest` in `model.go`**

```go
// Before (lines 523-526):
type ApproveRequest struct {
    ApprovedBy string `json:"approved_by" validate:"required,uuid"`
}

// After:
type ApproveRequest struct {
    ApprovedBy     string `json:"approved_by" validate:"required,uuid"`
    ApprovalReason string `json:"approval_reason"`
}

// Before (lines 541-544):
type RejectRequest struct {
    RejectedBy string `json:"rejected_by" validate:"required,uuid"`
}

// After:
type RejectRequest struct {
    RejectedBy      string `json:"rejected_by" validate:"required,uuid"`
    RejectionReason string `json:"rejection_reason"`
}
```

**Step 2: Update `App.Approve` and `App.Reject` signatures in `purchaseorderapp.go`**

```go
// Before (line 157):
func (a *App) Approve(ctx context.Context, id uuid.UUID, approvedBy uuid.UUID) (PurchaseOrder, error) {

// After:
func (a *App) Approve(ctx context.Context, id uuid.UUID, approvedBy uuid.UUID, reason string) (PurchaseOrder, error) {
```

Inside `Approve`, change (line 166):
```go
// Before:
approvedPO, err := a.purchaseorderbus.Approve(ctx, po, approvedBy, "")
// After:
approvedPO, err := a.purchaseorderbus.Approve(ctx, po, approvedBy, reason)
```

```go
// Before (line 178):
func (a *App) Reject(ctx context.Context, id uuid.UUID, rejectedBy uuid.UUID) (PurchaseOrder, error) {

// After:
func (a *App) Reject(ctx context.Context, id uuid.UUID, rejectedBy uuid.UUID, reason string) (PurchaseOrder, error) {
```

Inside `Reject`, change (line 187):
```go
// Before:
rejectedPO, err := a.purchaseorderbus.Reject(ctx, po, rejectedBy, "")
// After:
rejectedPO, err := a.purchaseorderbus.Reject(ctx, po, rejectedBy, reason)
```

**Step 3: Update API handlers to pass reason**

In `purchaseorderapi.go`, update `reject` handler (around line 134):
```go
// Before:
po, err := api.purchaseorderapp.Reject(ctx, parsed, rejectedBy)
// After:
po, err := api.purchaseorderapp.Reject(ctx, parsed, rejectedBy, app.RejectionReason)
```

Update `approve` handler (around line 159):
```go
// Before:
po, err := api.purchaseorderapp.Approve(ctx, parsed, approvedBy)
// After:
po, err := api.purchaseorderapp.Approve(ctx, parsed, approvedBy, app.ApprovalReason)
```

**Step 4: Build the affected packages**

```bash
go build ./app/domain/procurement/purchaseorderapp/... ./api/domain/http/procurement/purchaseorderapi/...
```
Expected: no errors.

**Step 5: Commit**

```bash
git add app/domain/procurement/purchaseorderapp/model.go app/domain/procurement/purchaseorderapp/purchaseorderapp.go api/domain/http/procurement/purchaseorderapi/purchaseorderapi.go
git commit -m "feat(purchaseorderapp): thread approval/rejection reason through HTTP endpoints"
```

---

## Task 7: Remove `approved_by`/`approved_date` bypass from app `UpdatePurchaseOrder`

**Why:** The general PUT endpoint exposes `approved_by` and `approved_date` as settable fields, bypassing the state guards in `Approve()`. The dedicated `/approve` endpoint now exists for this. Issue #9.

**Files:**
- Modify: `app/domain/procurement/purchaseorderapp/model.go:313-334` (UpdatePurchaseOrder struct)
- Modify: `app/domain/procurement/purchaseorderapp/model.go:349-494` (toBusUpdatePurchaseOrder func)

**Step 1: Remove `ApprovedBy` and `ApprovedDate` from `UpdatePurchaseOrder` struct**

```go
// Remove these two lines from UpdatePurchaseOrder (lines 329-330):
ApprovedBy   *string `json:"approved_by" validate:"omitempty,uuid"`
ApprovedDate *string `json:"approved_date" validate:"omitempty"`
```

**Step 2: Remove `ApprovedBy` and `ApprovedDate` handling from `toBusUpdatePurchaseOrder`**

Remove the entire blocks (lines 461-475):
```go
// Remove:
if app.ApprovedBy != nil {
    approvedBy, err := uuid.Parse(*app.ApprovedBy)
    if err != nil {
        return purchaseorderbus.UpdatePurchaseOrder{}, errs.NewFieldsError("approvedBy", err)
    }
    dest.ApprovedBy = &approvedBy
}

if app.ApprovedDate != nil {
    approvedDate, err := parseFlexibleDate(*app.ApprovedDate)
    if err != nil {
        return purchaseorderbus.UpdatePurchaseOrder{}, errs.NewFieldsError("approvedDate", err)
    }
    dest.ApprovedDate = &approvedDate
}
```

**Step 3: Build**

```bash
go build ./app/domain/procurement/purchaseorderapp/...
```
Expected: no errors.

**Step 4: Check for any tests that set approved_by/approved_date via the update endpoint**

```bash
grep -r "approved_by\|approved_date\|ApprovedBy\|ApprovedDate" api/cmd/services/ichor/tests/ --include="*.go" -l
```

If any test uses `UpdatePurchaseOrder{ApprovedBy: ...}`, update it to use the `Approve()` method instead.

**Step 5: Commit**

```bash
git add app/domain/procurement/purchaseorderapp/model.go
git commit -m "fix(purchaseorderapp): remove approved_by/approved_date from UpdatePurchaseOrder to enforce guarded path"
```

---

## Task 8: Add unit tests for `Approve()` and `Reject()` in `purchaseorderbus_test.go`

**Why:** CLAUDE.md 7-layer checklist requires unit tests for new business methods. Issue #8.

**Files:**
- Modify: `business/domain/procurement/purchaseorderbus/purchaseorderbus_test.go`

**Step 1: Add test runner calls to `Test_PurchaseOrder`**

```go
// Add after existing unitest.Run calls (after line 39):
unitest.Run(t, approve(db.BusDomain, sd), "approve")
unitest.Run(t, reject(db.BusDomain, sd), "reject")
```

**Step 2: Add `approve()` test function**

Add after the `delete()` function:

```go
func approve(busDomain dbtest.BusDomain, sd unitest.SeedData) []unitest.Table {
    approverID := sd.Admins[0].ID
    reason := "Approved by automated test"

    return []unitest.Table{
        {
            Name: "Approve",
            ExpResp: purchaseorderbus.PurchaseOrder{
                ID:        sd.PurchaseOrders[1].ID,
                ApprovedBy: &approverID,
                ApprovalReason: reason,
            },
            ExcFunc: func(ctx context.Context) any {
                got, err := busDomain.PurchaseOrder.Approve(ctx, sd.PurchaseOrders[1], approverID, reason)
                if err != nil {
                    return err
                }
                return got
            },
            CmpFunc: func(got, exp any) string {
                gotResp, exists := got.(purchaseorderbus.PurchaseOrder)
                if !exists {
                    return fmt.Sprintf("expected purchaseorderbus.PurchaseOrder, got %T", got)
                }
                expResp, exists := exp.(purchaseorderbus.PurchaseOrder)
                if !exists {
                    return fmt.Sprintf("expected purchaseorderbus.PurchaseOrder, got %T", exp)
                }

                if gotResp.ApprovedBy == nil {
                    return "expected ApprovedBy to be set"
                }
                if *gotResp.ApprovedBy != *expResp.ApprovedBy {
                    return fmt.Sprintf("ApprovedBy mismatch: got %s, exp %s", gotResp.ApprovedBy, expResp.ApprovedBy)
                }
                if gotResp.ApprovalReason != expResp.ApprovalReason {
                    return cmp.Diff(gotResp.ApprovalReason, expResp.ApprovalReason)
                }
                if gotResp.ApprovedDate.IsZero() {
                    return "expected ApprovedDate to be set"
                }
                return ""
            },
        },
        {
            Name: "ApproveAlreadyApproved",
            ExpResp: purchaseorderbus.ErrAlreadyApproved,
            ExcFunc: func(ctx context.Context) any {
                // First approve it
                po, err := busDomain.PurchaseOrder.Approve(ctx, sd.PurchaseOrders[2], approverID, reason)
                if err != nil {
                    return err
                }
                // Try to approve again — should fail
                _, err = busDomain.PurchaseOrder.Approve(ctx, po, approverID, reason)
                return err
            },
            CmpFunc: func(got, exp any) string {
                gotErr, ok := got.(error)
                if !ok {
                    return fmt.Sprintf("expected error, got %T", got)
                }
                expErr := exp.(error)
                if !errors.Is(gotErr, expErr) {
                    return fmt.Sprintf("expected %v, got %v", expErr, gotErr)
                }
                return ""
            },
        },
        {
            Name: "ApproveAlreadyRejected",
            ExpResp: purchaseorderbus.ErrAlreadyRejected,
            ExcFunc: func(ctx context.Context) any {
                // First reject it
                po, err := busDomain.PurchaseOrder.Reject(ctx, sd.PurchaseOrders[3], approverID, "rejecting first")
                if err != nil {
                    return err
                }
                // Try to approve a rejected PO — should fail
                _, err = busDomain.PurchaseOrder.Approve(ctx, po, approverID, reason)
                return err
            },
            CmpFunc: func(got, exp any) string {
                gotErr, ok := got.(error)
                if !ok {
                    return fmt.Sprintf("expected error, got %T", got)
                }
                expErr := exp.(error)
                if !errors.Is(gotErr, expErr) {
                    return fmt.Sprintf("expected %v, got %v", expErr, gotErr)
                }
                return ""
            },
        },
    }
}
```

**Step 3: Add `reject()` test function**

```go
func reject(busDomain dbtest.BusDomain, sd unitest.SeedData) []unitest.Table {
    rejectorID := sd.Admins[0].ID
    reason := "Rejected by automated test"

    return []unitest.Table{
        {
            Name: "Reject",
            ExpResp: purchaseorderbus.PurchaseOrder{
                ID:              sd.PurchaseOrders[4].ID,
                RejectedBy:      &rejectorID,
                RejectionReason: reason,
            },
            ExcFunc: func(ctx context.Context) any {
                got, err := busDomain.PurchaseOrder.Reject(ctx, sd.PurchaseOrders[4], rejectorID, reason)
                if err != nil {
                    return err
                }
                return got
            },
            CmpFunc: func(got, exp any) string {
                gotResp, exists := got.(purchaseorderbus.PurchaseOrder)
                if !exists {
                    return fmt.Sprintf("expected purchaseorderbus.PurchaseOrder, got %T", got)
                }
                expResp, exists := exp.(purchaseorderbus.PurchaseOrder)
                if !exists {
                    return fmt.Sprintf("expected purchaseorderbus.PurchaseOrder, got %T", exp)
                }

                if gotResp.RejectedBy == nil {
                    return "expected RejectedBy to be set"
                }
                if *gotResp.RejectedBy != *expResp.RejectedBy {
                    return fmt.Sprintf("RejectedBy mismatch: got %s, exp %s", gotResp.RejectedBy, expResp.RejectedBy)
                }
                if gotResp.RejectionReason != expResp.RejectionReason {
                    return cmp.Diff(gotResp.RejectionReason, expResp.RejectionReason)
                }
                if gotResp.RejectedDate.IsZero() {
                    return "expected RejectedDate to be set"
                }
                return ""
            },
        },
        {
            Name: "RejectAlreadyApproved",
            ExpResp: purchaseorderbus.ErrAlreadyApproved,
            ExcFunc: func(ctx context.Context) any {
                // PurchaseOrders[2] was approved in the approve tests — but since tests run in isolation
                // per unitest.Run, re-approve a fresh PO here.
                // Note: sd.PurchaseOrders[2] is fresh at the start of this sub-run.
                po, err := busDomain.PurchaseOrder.Approve(ctx, sd.PurchaseOrders[2], rejectorID, "approving first")
                if err != nil {
                    return err
                }
                // Try to reject an already-approved PO
                _, err = busDomain.PurchaseOrder.Reject(ctx, po, rejectorID, reason)
                return err
            },
            CmpFunc: func(got, exp any) string {
                gotErr, ok := got.(error)
                if !ok {
                    return fmt.Sprintf("expected error, got %T", got)
                }
                expErr := exp.(error)
                if !errors.Is(gotErr, expErr) {
                    return fmt.Sprintf("expected %v, got %v", expErr, gotErr)
                }
                return ""
            },
        },
    }
}
```

**Note on test isolation:** The `unitest.Run` framework runs each table entry independently. `sd.PurchaseOrders` indices used here are:
- [1] — for approve happy path
- [2] — for approve-already-approved and reject-already-approved
- [3] — for approve-already-rejected
- [4] — for reject happy path

The seed produces 5 POs. If test ordering causes state conflicts, use `sd.PurchaseOrders[4]` for approve happy path and adjust accordingly. The CmpFunc-based comparison style (checking specific fields rather than full struct diff) avoids timestamp issues.

**Step 4: Add missing `errors` import if not present**

Check that `"errors"` is in the test file imports. If not, add it.

**Step 5: Run tests**

```bash
go test ./business/domain/procurement/purchaseorderbus/... -run Test_PurchaseOrder -v
```
Expected: all subtests pass including new `approve` and `reject`.

**Step 6: Commit**

```bash
git add business/domain/procurement/purchaseorderbus/purchaseorderbus_test.go
git commit -m "test(purchaseorderbus): add unit tests for Approve() and Reject() business methods"
```

---

## Task 9: Update `docs/workflow/README.md` handler catalog

**Why:** The README still references "All 7 action types" and the Supported Actions table is missing the 6 new handlers. Issue #7.

**Files:**
- Modify: `docs/workflow/README.md:11,43-53`

**Step 1: Update the quick links table entry**

```markdown
<!-- Before (line 11): -->
| [Actions](actions/) | All 7 action types and their configuration |

<!-- After: -->
| [Actions](actions/) | All 13 action types and their configuration |
```

**Step 2: Add the 6 new action types to the Supported Actions table**

```markdown
<!-- Add after the evaluate_condition row (after line 53): -->
| `approve_inventory_adjustment` | Approves a pending inventory adjustment, recording approver and reason |
| `reject_inventory_adjustment` | Rejects a pending inventory adjustment, recording rejector and reason |
| `approve_transfer_order` | Approves a pending transfer order, recording approver and reason |
| `reject_transfer_order` | Rejects a pending transfer order, recording rejector and reason |
| `approve_purchase_order` | Approves a purchase order, recording approver and reason for audit trail |
| `reject_purchase_order` | Rejects a purchase order, recording rejector and reason for audit trail |
```

**Step 3: Commit**

```bash
git add docs/workflow/README.md
git commit -m "docs(workflow): add 6 new approval/rejection action types to README catalog"
```

---

## Final: Build all affected packages

```bash
go build ./business/domain/inventory/transferorderbus/... ./business/domain/inventory/inventoryadjustmentbus/... ./business/domain/procurement/purchaseorderbus/... ./app/domain/procurement/purchaseorderapp/... ./api/domain/http/procurement/purchaseorderapi/... ./business/sdk/workflow/workflowactions/...
```
Expected: no errors.
