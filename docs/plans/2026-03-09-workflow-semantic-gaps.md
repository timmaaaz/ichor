# Workflow Semantic Gaps Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Close four categories of workflow semantic gaps: approval handler pairs with full audit trails, a delegate event bug fix in approvalrequestbus, and a field schema discovery API.

**Architecture:** Three approval domains (inventory adjustment, transfer order, purchase order) each need DB migrations for audit trail columns, bus method updates, and paired approve/reject workflow action handlers. The approvalrequestbus needs delegate support added from scratch (it currently has no delegate field), then a `resolve_approval_request` handler. The field schema discovery API adds a static enum registry and a new endpoint in referenceapi.

**Tech Stack:** Go 1.23, PostgreSQL 16.4, sqlx, Temporal (for resolve handler IsAsync=true), `business/sdk/workflow` action handler interface

---

> **Note:** Phase 1 (putaway completion) is already implemented. `putawaytaskapp.complete()` exists at `app/domain/inventory/putawaytaskapp/putawaytaskapp.go:134`, `UpsertQuantity()` exists in inventoryitembus, and `TestUpdate200Complete` covers the 3-way atomic transaction in `api/cmd/services/ichor/tests/inventory/putawaytaskapi/update_test.go`. Start at Task 1.

---

## Task 1: Inventory Adjustment — DB Migration (audit trail columns)

**Files:**
- Modify: `business/sdk/migrate/sql/migrate.sql`

**Step 1: Append the migration**

Add to the end of `migrate.sql`:

```sql
-- Version: 2.09
-- Description: Add audit trail columns to inventory.inventory_adjustments for workflow approval handlers.
ALTER TABLE inventory.inventory_adjustments
    ADD COLUMN rejected_by     UUID        NULL REFERENCES core.users(id),
    ADD COLUMN rejection_reason TEXT       NULL,
    ADD COLUMN approval_reason  TEXT       NULL;
```

**Step 2: Verify SQL is valid**

```bash
cd /Users/jaketimmer/src/work/superior/ichor/ichor
grep -A5 "Version: 2.09" business/sdk/migrate/sql/migrate.sql
```

Expected: the three ADD COLUMN lines appear.

**Step 3: Commit**

```bash
git add business/sdk/migrate/sql/migrate.sql
git commit -m "feat(inventory): add audit trail columns to inventory_adjustments"
```

---

## Task 2: Inventory Adjustment — Bus Model + DB Model

**Files:**
- Modify: `business/domain/inventory/inventoryadjustmentbus/model.go`
- Modify: `business/domain/inventory/inventoryadjustmentbus/stores/inventoryadjustmentdb/model.go`

**Read first:**
- `business/domain/inventory/inventoryadjustmentbus/model.go` — see current struct fields
- `business/domain/inventory/inventoryadjustmentbus/stores/inventoryadjustmentdb/model.go` — see db model pattern

**Step 1: Add audit trail fields to bus model**

In `inventoryadjustmentbus/model.go`, add three fields to `InventoryAdjustment`:

```go
type InventoryAdjustment struct {
    // ... existing fields ...
    ApprovedBy            *uuid.UUID `json:"approved_by"`
    RejectedBy            *uuid.UUID `json:"rejected_by"`       // add
    ApprovalStatus        string     `json:"approval_status"`
    ApprovalReason        string     `json:"approval_reason"`   // add
    RejectionReason       string     `json:"rejection_reason"`  // add
    // ... rest of existing fields ...
}
```

Add to `UpdateInventoryAdjustment`:

```go
type UpdateInventoryAdjustment struct {
    // ... existing fields ...
    RejectedBy      *uuid.UUID `json:"rejected_by,omitempty"`      // add
    ApprovalReason  *string    `json:"approval_reason,omitempty"`  // add
    RejectionReason *string    `json:"rejection_reason,omitempty"` // add
}
```

**Step 2: Update DB model to match**

In `inventoryadjustmentdb/model.go`, add fields to the `inventoryAdjustment` struct:

```go
type inventoryAdjustment struct {
    // ... existing fields ...
    ApprovedBy      *uuid.UUID `db:"approved_by"`
    RejectedBy      *uuid.UUID `db:"rejected_by"`       // add
    ApprovalStatus  string     `db:"approval_status"`
    ApprovalReason  string     `db:"approval_reason"`   // add
    RejectionReason string     `db:"rejection_reason"`  // add
    // ... rest ...
}
```

Update both `toBusInventoryAdjustment` and `toDBInventoryAdjustment` to map the three new fields.

**Step 3: Build check**

```bash
go build ./business/domain/inventory/inventoryadjustmentbus/...
```

Expected: no errors.

**Step 4: Commit**

```bash
git add business/domain/inventory/inventoryadjustmentbus/
git commit -m "feat(inventory): add audit trail fields to InventoryAdjustment models"
```

---

## Task 3: Inventory Adjustment — Fix Bus Methods (Reject + Approve signatures)

**Files:**
- Modify: `business/domain/inventory/inventoryadjustmentbus/inventoryadjustmentbus.go`

**Read first:**
- `business/domain/inventory/inventoryadjustmentbus/inventoryadjustmentbus.go` — see current `Approve()` and `Reject()` signatures

**Step 1: Update `Approve()` to accept a reason parameter**

Change signature from `Approve(ctx, ia, approvedBy uuid.UUID)` to include reason:

```go
func (b *Business) Approve(ctx context.Context, ia InventoryAdjustment, approvedBy uuid.UUID, reason string) (InventoryAdjustment, error) {
    ctx, span := otel.AddSpan(ctx, "business.inventoryadjustmentbus.approve")
    defer span.End()

    if ia.ApprovalStatus != "pending" {
        return InventoryAdjustment{}, fmt.Errorf("approve: %w", ErrInvalidApprovalStatus)
    }

    before := ia
    now := time.Now()
    ia.ApprovedBy = &approvedBy
    ia.ApprovalStatus = "approved"
    ia.ApprovalReason = reason   // new
    ia.UpdatedDate = now

    if err := b.storer.Update(ctx, ia); err != nil {
        return InventoryAdjustment{}, fmt.Errorf("approve: %w", err)
    }

    if err := b.delegate.Call(ctx, ActionUpdatedData(before, ia)); err != nil {
        b.log.Error(ctx, "inventoryadjustmentbus: delegate call failed", "action", ActionUpdated, "err", err)
    }

    return ia, nil
}
```

**Step 2: Fix `Reject()` to capture rejectedBy and reason**

Change signature from `Reject(ctx, ia)` to include rejectedBy and reason:

```go
func (b *Business) Reject(ctx context.Context, ia InventoryAdjustment, rejectedBy uuid.UUID, reason string) (InventoryAdjustment, error) {
    ctx, span := otel.AddSpan(ctx, "business.inventoryadjustmentbus.reject")
    defer span.End()

    if ia.ApprovalStatus != "pending" {
        return InventoryAdjustment{}, fmt.Errorf("reject: %w", ErrInvalidApprovalStatus)
    }

    before := ia
    ia.RejectedBy = &rejectedBy       // new
    ia.ApprovalStatus = "rejected"
    ia.RejectionReason = reason        // new
    ia.UpdatedDate = time.Now()

    if err := b.storer.Update(ctx, ia); err != nil {
        return InventoryAdjustment{}, fmt.Errorf("reject: %w", err)
    }

    if err := b.delegate.Call(ctx, ActionUpdatedData(before, ia)); err != nil {
        b.log.Error(ctx, "inventoryadjustmentbus: delegate call failed", "action", ActionUpdated, "err", err)
    }

    return ia, nil
}
```

**Step 3: Fix all callers of Reject()**

Search for callers:

```bash
grep -r "\.Reject(" /Users/jaketimmer/src/work/superior/ichor/ichor/app --include="*.go" -l
grep -r "\.Reject(" /Users/jaketimmer/src/work/superior/ichor/ichor/api --include="*.go" -l
```

Update each caller to pass `rejectedBy uuid.UUID` (from middleware user ID) and `reason string` (from request body or empty string).

**Step 4: Build check**

```bash
go build ./business/domain/inventory/inventoryadjustmentbus/...
go build ./app/...
go build ./api/...
```

Expected: no errors.

**Step 5: Commit**

```bash
git add business/domain/inventory/inventoryadjustmentbus/inventoryadjustmentbus.go
git commit -m "feat(inventory): add audit trail fields to Approve/Reject bus methods"
```

---

## Task 4: Transfer Order — DB Migration

**Files:**
- Modify: `business/sdk/migrate/sql/migrate.sql`

**Step 1: Append the migration**

```sql
-- Version: 2.10
-- Description: Add audit trail columns to inventory.transfer_orders for workflow approval handlers.
ALTER TABLE inventory.transfer_orders
    ADD COLUMN rejected_by_id   UUID NULL REFERENCES core.users(id),
    ADD COLUMN rejection_reason TEXT NULL,
    ADD COLUMN approval_reason  TEXT NULL;
```

**Step 2: Build check**

```bash
grep -A5 "Version: 2.10" business/sdk/migrate/sql/migrate.sql
```

**Step 3: Commit**

```bash
git add business/sdk/migrate/sql/migrate.sql
git commit -m "feat(inventory): add audit trail columns to transfer_orders"
```

---

## Task 5: Transfer Order — Bus Model + DB Model + Reject() Method

**Files:**
- Modify: `business/domain/inventory/transferorderbus/model.go`
- Modify: `business/domain/inventory/transferorderbus/stores/transferorderdb/model.go`
- Modify: `business/domain/inventory/transferorderbus/transferorderbus.go`

**Read first:**
- `business/domain/inventory/transferorderbus/model.go` — see current fields
- `business/domain/inventory/transferorderbus/stores/transferorderdb/model.go` — see db model pattern
- `business/domain/inventory/transferorderbus/transferorderbus.go` — see existing `Approve()` method

**Step 1: Add audit trail fields to `TransferOrder` bus model**

```go
type TransferOrder struct {
    // ... existing fields ...
    ApprovedByID   *uuid.UUID `json:"approved_by_id"`
    RejectedByID   *uuid.UUID `json:"rejected_by_id"`    // add
    ApprovalReason string     `json:"approval_reason"`   // add
    RejectionReason string    `json:"rejection_reason"`  // add
    Status         string     `json:"status"`
    // ... rest ...
}
```

Add to `UpdateTransferOrder`:
```go
RejectedByID    *uuid.UUID `json:"rejected_by_id,omitempty"`
ApprovalReason  *string    `json:"approval_reason,omitempty"`
RejectionReason *string    `json:"rejection_reason,omitempty"`
```

**Step 2: Update DB model to match**

Follow the exact same pattern as Task 2 for the `transferorderdb/model.go`.
Add `RejectedByID *uuid.UUID`, `ApprovalReason string`, `RejectionReason string` with `db:"rejected_by_id"`, `db:"approval_reason"`, `db:"rejection_reason"` tags. Update both conversion functions.

**Step 3: Update `Approve()` to accept reason**

Change `Approve(ctx, to, approvedBy uuid.UUID)` to `Approve(ctx, to, approvedBy uuid.UUID, reason string)` and set `to.ApprovalReason = reason`.

**Step 4: Add `Reject()` method and `ErrInvalidTransferStatus` error**

Add the error var:
```go
var (
    // ... existing errors ...
    ErrInvalidTransferStatus = errors.New("transfer order is not in a state that allows approval/rejection")
)
```

Add the method (place after `Approve()`):
```go
func (b *Business) Reject(ctx context.Context, to TransferOrder, rejectedBy uuid.UUID, reason string) (TransferOrder, error) {
    ctx, span := otel.AddSpan(ctx, "business.transferorderbus.reject")
    defer span.End()

    if to.Status == "approved" || to.Status == "rejected" {
        return TransferOrder{}, fmt.Errorf("reject: %w", ErrInvalidTransferStatus)
    }

    before := to
    now := time.Now()
    to.RejectedByID = &rejectedBy
    to.Status = "rejected"
    to.RejectionReason = reason
    to.UpdatedDate = now

    if err := b.storer.Update(ctx, to); err != nil {
        return TransferOrder{}, fmt.Errorf("reject: %w", err)
    }

    if err := b.delegate.Call(ctx, ActionUpdatedData(before, to)); err != nil {
        b.log.Error(ctx, "transferorderbus: delegate call failed", "action", ActionUpdated, "err", err)
    }

    return to, nil
}
```

**Step 5: Fix callers of `Approve()`**

```bash
grep -r "\.Approve(" /Users/jaketimmer/src/work/superior/ichor/ichor/app --include="*.go" -l
grep -r "\.Approve(" /Users/jaketimmer/src/work/superior/ichor/ichor/api --include="*.go" -l
```

Update each caller to pass `reason string` as the new last argument.

**Step 6: Build check**

```bash
go build ./business/domain/inventory/transferorderbus/...
go build ./app/...
go build ./api/...
```

**Step 7: Commit**

```bash
git add business/domain/inventory/transferorderbus/
git commit -m "feat(inventory): add audit trail fields and Reject() to transferorderbus"
```

---

## Task 6: Purchase Order — DB Migration

**Files:**
- Modify: `business/sdk/migrate/sql/migrate.sql`

**Step 1: Append the migration**

```sql
-- Version: 2.11
-- Description: Add audit trail columns to procurement.purchase_orders for workflow approval handlers.
ALTER TABLE procurement.purchase_orders
    ADD COLUMN rejected_by      UUID         NULL REFERENCES core.users(id),
    ADD COLUMN rejected_date    TIMESTAMPTZ  NULL,
    ADD COLUMN rejection_reason TEXT         NULL,
    ADD COLUMN approval_reason  TEXT         NULL;
```

**Step 2: Commit**

```bash
git add business/sdk/migrate/sql/migrate.sql
git commit -m "feat(procurement): add audit trail columns to purchase_orders"
```

---

## Task 7: Purchase Order — Bus Model + DB Model + Reject() Method

**Files:**
- Modify: `business/domain/procurement/purchaseorderbus/model.go`
- Modify: `business/domain/procurement/purchaseorderbus/stores/purchaseorderdb/model.go`
- Modify: `business/domain/procurement/purchaseorderbus/purchaseorderbus.go`

**Read first:**
- `business/domain/procurement/purchaseorderbus/model.go` — see current PurchaseOrder fields (ApprovedBy is `uuid.UUID`, not pointer)
- `business/domain/procurement/purchaseorderbus/stores/purchaseorderdb/model.go` — see db model
- `business/domain/procurement/purchaseorderbus/purchaseorderbus.go` — see existing `Approve()` method

**Step 1: Add fields to `PurchaseOrder` bus model**

```go
type PurchaseOrder struct {
    // ... existing fields ...
    ApprovedBy      uuid.UUID `json:"approved_by"`
    ApprovedDate    time.Time `json:"approved_date"`
    RejectedBy      uuid.UUID `json:"rejected_by"`       // add (zero UUID = not rejected)
    RejectedDate    time.Time `json:"rejected_date"`     // add
    ApprovalReason  string    `json:"approval_reason"`   // add
    RejectionReason string    `json:"rejection_reason"`  // add
    // ... rest ...
}
```

Add to `UpdatePurchaseOrder`:
```go
ApprovalReason  *string    `json:"approval_reason,omitempty"`
RejectionReason *string    `json:"rejection_reason,omitempty"`
RejectedBy      *uuid.UUID `json:"rejected_by,omitempty"`
RejectedDate    *time.Time `json:"rejected_date,omitempty"`
```

**Step 2: Update DB model**

In `purchaseorderdb/model.go`, find the db struct (it likely uses `sql.NullString`/`sql.NullTime` for nullable fields or just `uuid.UUID` for non-nullable). Follow the existing pattern for nullable UUID columns (check how `delivery_street_id` handles optional UUIDs). Add:
- `RejectedBy uuid.UUID` (maps to nullable `rejected_by` — use zero UUID as sentinel, or `*uuid.UUID` if the pattern supports it)
- `RejectedDate time.Time` (maps to nullable `rejected_date` — use zero time as sentinel)
- `ApprovalReason string` (maps to `approval_reason`, nullable text → empty string)
- `RejectionReason string` (maps to `rejection_reason`, nullable text → empty string)

Update both conversion functions to map the four new fields.

**Step 3: Update `Approve()` to accept reason and add `ErrAlreadyApproved`**

Add error var:
```go
var (
    // ... existing errors ...
    ErrAlreadyApproved = errors.New("purchase order is already approved")
    ErrAlreadyRejected = errors.New("purchase order is already rejected")
)
```

Change `Approve()` signature:
```go
func (b *Business) Approve(ctx context.Context, po PurchaseOrder, approvedBy uuid.UUID, reason string) (PurchaseOrder, error) {
    // Guard: if ApprovedBy is non-zero, already approved
    if po.ApprovedBy != (uuid.UUID{}) {
        return PurchaseOrder{}, fmt.Errorf("approve: %w", ErrAlreadyApproved)
    }

    before := po
    now := time.Now().UTC()
    po.ApprovedBy = approvedBy
    po.ApprovedDate = now
    po.ApprovalReason = reason  // new
    po.UpdatedBy = approvedBy
    po.UpdatedDate = now

    // ... rest unchanged ...
}
```

**Step 4: Add `Reject()` method**

```go
func (b *Business) Reject(ctx context.Context, po PurchaseOrder, rejectedBy uuid.UUID, reason string) (PurchaseOrder, error) {
    ctx, span := otel.AddSpan(ctx, "business.purchaseorderbus.reject")
    defer span.End()

    if po.ApprovedBy != (uuid.UUID{}) {
        return PurchaseOrder{}, fmt.Errorf("reject: %w", ErrAlreadyApproved)
    }
    if po.RejectedBy != (uuid.UUID{}) {
        return PurchaseOrder{}, fmt.Errorf("reject: %w", ErrAlreadyRejected)
    }

    before := po
    now := time.Now().UTC()
    po.RejectedBy = rejectedBy
    po.RejectedDate = now
    po.RejectionReason = reason
    po.UpdatedBy = rejectedBy
    po.UpdatedDate = now

    if err := b.storer.Update(ctx, po); err != nil {
        return PurchaseOrder{}, fmt.Errorf("reject: %w", err)
    }

    if err := b.del.Call(ctx, ActionUpdatedData(before, po)); err != nil {
        b.log.Error(ctx, "purchaseorderbus: delegate call failed", "action", ActionUpdated, "err", err)
    }

    return po, nil
}
```

**Step 5: Fix callers of `Approve()`**

```bash
grep -rn "\.Approve(" /Users/jaketimmer/src/work/superior/ichor/ichor --include="*.go" | grep -v "_test.go"
```

Update each caller to pass `reason string` (pass empty string `""` if the caller has no reason to provide).

**Step 6: Build check**

```bash
go build ./business/domain/procurement/purchaseorderbus/...
go build ./app/...
go build ./api/...
```

**Step 7: Commit**

```bash
git add business/domain/procurement/purchaseorderbus/
git commit -m "feat(procurement): add audit trail fields and Reject() to purchaseorderbus"
```

---

## Task 8: Approve/Reject Inventory Adjustment — Action Handlers

**Files:**
- Create: `business/sdk/workflow/workflowactions/inventory/approve_adjustment.go`
- Create: `business/sdk/workflow/workflowactions/inventory/reject_adjustment.go`
- Create: `business/sdk/workflow/workflowactions/inventory/approve_adjustment_test.go`
- Create: `business/sdk/workflow/workflowactions/inventory/reject_adjustment_test.go`

**Read first:**
- `business/sdk/workflow/workflowactions/inventory/receive.go` — canonical handler pattern to follow
- `business/sdk/workflow/workflowactions/inventory/check_inventory_test.go` — test pattern

**Step 1: Write failing test for approve_adjustment (validate cases)**

```go
// approve_adjustment_test.go
package inventory_test

func Test_ApproveInventoryAdjustment(t *testing.T) {
    db := dbtest.NewDatabase(t, "Test_ApproveInventoryAdjustment")

    sd, err := insertApprovalSeedData(db.BusDomain)
    if err != nil {
        t.Fatalf("seeding: %v", err)
    }

    var buf bytes.Buffer
    log := logger.New(&buf, logger.LevelInfo, "TEST", func(context.Context) string {
        return otel.GetTraceID(context.Background())
    })

    sd.Handler = inventory.NewApproveInventoryAdjustmentHandler(log, db.BusDomain.InventoryAdjustment)

    unitest.Run(t, approveAdjustmentValidateTests(sd), "validate")
    unitest.Run(t, approveAdjustmentExecuteTests(db.BusDomain, sd), "execute")
}
```

**Step 2: Run test to verify it fails**

```bash
go test ./business/sdk/workflow/workflowactions/inventory/... -run Test_ApproveInventoryAdjustment -v
```

Expected: compile error — `NewApproveInventoryAdjustmentHandler` does not exist.

**Step 3: Create `approve_adjustment.go`**

```go
package inventory

import (
    "context"
    "encoding/json"
    "fmt"

    "github.com/google/uuid"
    "github.com/timmaaaz/ichor/business/domain/inventory/inventoryadjustmentbus"
    "github.com/timmaaaz/ichor/business/sdk/workflow"
    "github.com/timmaaaz/ichor/foundation/logger"
)

type ApproveInventoryAdjustmentConfig struct {
    AdjustmentID   string `json:"adjustment_id"`
    ApprovalReason string `json:"approval_reason,omitempty"`
}

type ApproveInventoryAdjustmentHandler struct {
    log                    *logger.Logger
    inventoryAdjustmentBus *inventoryadjustmentbus.Business
}

func NewApproveInventoryAdjustmentHandler(log *logger.Logger, inventoryAdjustmentBus *inventoryadjustmentbus.Business) *ApproveInventoryAdjustmentHandler {
    return &ApproveInventoryAdjustmentHandler{log: log, inventoryAdjustmentBus: inventoryAdjustmentBus}
}

func (h *ApproveInventoryAdjustmentHandler) GetType() string        { return "approve_inventory_adjustment" }
func (h *ApproveInventoryAdjustmentHandler) IsAsync() bool          { return false }
func (h *ApproveInventoryAdjustmentHandler) SupportsManualExecution() bool { return true }
func (h *ApproveInventoryAdjustmentHandler) GetDescription() string {
    return "Approve a pending inventory adjustment, capturing approver and reason for audit trail"
}

func (h *ApproveInventoryAdjustmentHandler) Validate(config json.RawMessage) error {
    var cfg ApproveInventoryAdjustmentConfig
    if err := json.Unmarshal(config, &cfg); err != nil {
        return fmt.Errorf("invalid config: %w", err)
    }
    if cfg.AdjustmentID == "" {
        return fmt.Errorf("adjustment_id is required")
    }
    if _, err := uuid.Parse(cfg.AdjustmentID); err != nil {
        return fmt.Errorf("invalid adjustment_id: %w", err)
    }
    return nil
}

func (h *ApproveInventoryAdjustmentHandler) GetOutputPorts() []workflow.OutputPort {
    return []workflow.OutputPort{
        {Name: "approved", Description: "Adjustment approved successfully", IsDefault: true},
        {Name: "not_found", Description: "Adjustment not found"},
        {Name: "already_approved", Description: "Adjustment was already approved (idempotent)"},
        {Name: "already_rejected", Description: "Adjustment was already rejected — cannot approve"},
        {Name: "failure", Description: "Unexpected error"},
    }
}

func (h *ApproveInventoryAdjustmentHandler) GetEntityModifications(config json.RawMessage) []workflow.EntityModification {
    return []workflow.EntityModification{
        {EntityName: "inventory.inventory_adjustments", EventType: "on_update", Fields: []string{"approval_status", "approved_by", "approval_reason"}},
    }
}

func (h *ApproveInventoryAdjustmentHandler) Execute(ctx context.Context, config json.RawMessage, execCtx workflow.ActionExecutionContext) (any, error) {
    var cfg ApproveInventoryAdjustmentConfig
    if err := json.Unmarshal(config, &cfg); err != nil {
        return map[string]any{"output": "failure", "error": err.Error()}, nil
    }

    id, err := uuid.Parse(cfg.AdjustmentID)
    if err != nil {
        return map[string]any{"output": "failure", "error": "invalid adjustment_id"}, nil
    }

    ia, err := h.inventoryAdjustmentBus.QueryByID(ctx, id)
    if err != nil {
        if errors.Is(err, inventoryadjustmentbus.ErrNotFound) {
            return map[string]any{"output": "not_found", "adjustment_id": cfg.AdjustmentID}, nil
        }
        return nil, fmt.Errorf("query adjustment: %w", err)
    }

    switch ia.ApprovalStatus {
    case "approved":
        return map[string]any{"output": "already_approved", "adjustment_id": cfg.AdjustmentID}, nil
    case "rejected":
        return map[string]any{"output": "already_rejected", "adjustment_id": cfg.AdjustmentID}, nil
    }

    approved, err := h.inventoryAdjustmentBus.Approve(ctx, ia, execCtx.UserID, cfg.ApprovalReason)
    if err != nil {
        return nil, fmt.Errorf("approve adjustment: %w", err)
    }

    return map[string]any{
        "output":          "approved",
        "adjustment_id":   approved.InventoryAdjustmentID.String(),
        "approved_by":     execCtx.UserID.String(),
        "approval_reason": cfg.ApprovalReason,
    }, nil
}
```

**Step 4: Create `reject_adjustment.go`**

Mirror of approve handler with:
- Type: `"reject_inventory_adjustment"`
- Config: `AdjustmentID string`, `RejectionReason string` (required — non-empty)
- Validate: also check `rejection_reason != ""`
- Output ports: `rejected`, `not_found`, `already_approved`, `already_rejected`, `failure`
- Execute: calls `bus.Reject(ctx, ia, execCtx.UserID, cfg.RejectionReason)`

**Step 5: Write tests for both handlers**

For each handler, write a `unitest.Table` slice covering:

Validate tests (use `json.RawMessage`):
- Missing `adjustment_id` → error
- Invalid UUID format → error
- Missing `rejection_reason` on reject handler → error
- Valid config → nil error

Execute tests (need real DB with seeded adjustment in pending state):
- Valid pending adjustment → `approved`/`rejected` output
- Already approved → `already_approved` output
- Already rejected → `already_rejected` output
- Unknown ID → `not_found` output

**Step 6: Run tests**

```bash
go test ./business/sdk/workflow/workflowactions/inventory/... -run "Test_ApproveInventoryAdjustment|Test_RejectInventoryAdjustment" -v
```

Expected: PASS.

**Step 7: Build check**

```bash
go build ./business/sdk/workflow/workflowactions/inventory/...
```

**Step 8: Commit**

```bash
git add business/sdk/workflow/workflowactions/inventory/approve_adjustment.go \
        business/sdk/workflow/workflowactions/inventory/reject_adjustment.go \
        business/sdk/workflow/workflowactions/inventory/approve_adjustment_test.go \
        business/sdk/workflow/workflowactions/inventory/reject_adjustment_test.go
git commit -m "feat(workflow): add approve/reject inventory adjustment action handlers"
```

---

## Task 9: Approve/Reject Transfer Order — Action Handlers

**Files:**
- Create: `business/sdk/workflow/workflowactions/inventory/approve_transfer_order.go`
- Create: `business/sdk/workflow/workflowactions/inventory/reject_transfer_order.go`
- Create: `business/sdk/workflow/workflowactions/inventory/approve_transfer_order_test.go`
- Create: `business/sdk/workflow/workflowactions/inventory/reject_transfer_order_test.go`

**Pattern:** Identical to Task 8. Follow the same structure with:
- Approve type: `"approve_transfer_order"`, config field: `TransferOrderID`
- Reject type: `"reject_transfer_order"`, config field: `TransferOrderID`, required `RejectionReason`
- `ErrInvalidTransferStatus` guards in bus → map to `already_approved`/`already_rejected` output ports
- Entity modification: `inventory.transfer_orders`, fields: `[status, approved_by_id, rejection_reason]`
- Bus: `transferorderbus.Approve(ctx, to, execCtx.UserID, reason)` / `transferorderbus.Reject(ctx, to, execCtx.UserID, reason)`
- BusDependencies field needed: `TransferOrder *transferorderbus.Business` (must be added in Task 11)

Follow the exact same step structure as Task 8 (write test → fail → implement → pass → commit).

**Run tests:**

```bash
go test ./business/sdk/workflow/workflowactions/inventory/... -run "Test_ApproveTransferOrder|Test_RejectTransferOrder" -v
```

**Step N: Commit**

```bash
git add business/sdk/workflow/workflowactions/inventory/approve_transfer_order.go \
        business/sdk/workflow/workflowactions/inventory/reject_transfer_order.go \
        business/sdk/workflow/workflowactions/inventory/approve_transfer_order_test.go \
        business/sdk/workflow/workflowactions/inventory/reject_transfer_order_test.go
git commit -m "feat(workflow): add approve/reject transfer order action handlers"
```

---

## Task 10: Approve/Reject Purchase Order — Action Handlers

**Files:**
- Create: `business/sdk/workflow/workflowactions/procurement/approve_po.go`
- Create: `business/sdk/workflow/workflowactions/procurement/reject_po.go`
- Create: `business/sdk/workflow/workflowactions/procurement/approve_po_test.go`
- Create: `business/sdk/workflow/workflowactions/procurement/reject_po_test.go`

**Read first:**
- `business/sdk/workflow/workflowactions/procurement/createpo.go` — existing procurement handler pattern
- `business/domain/procurement/purchaseorderbus/purchaseorderbus.go` — `Approve()` and new `Reject()` signatures

**Pattern:** Same handler structure as Tasks 8–9. Key differences:
- Package: `procurement` (not `inventory`)
- Approve type: `"approve_purchase_order"`, config: `PurchaseOrderID string`, `ApprovalReason string` (optional)
- Reject type: `"reject_purchase_order"`, config: `PurchaseOrderID string`, `RejectionReason string` (required)
- Guard: `if po.ApprovedBy != (uuid.UUID{})` → `already_approved`; `if po.RejectedBy != (uuid.UUID{})` → `already_rejected`
- Entity modification: `procurement.purchase_orders`, fields: `[approved_by, approved_date, rejected_by, rejected_date]`
- BusDependencies field `PurchaseOrder` already exists in `register.go` ✓

Follow the exact same step structure as Task 8.

**Run tests:**

```bash
go test ./business/sdk/workflow/workflowactions/procurement/... -run "Test_ApprovePurchaseOrder|Test_RejectPurchaseOrder" -v
```

**Commit:**

```bash
git add business/sdk/workflow/workflowactions/procurement/approve_po.go \
        business/sdk/workflow/workflowactions/procurement/reject_po.go \
        business/sdk/workflow/workflowactions/procurement/approve_po_test.go \
        business/sdk/workflow/workflowactions/procurement/reject_po_test.go
git commit -m "feat(workflow): add approve/reject purchase order action handlers"
```

---

## Task 11: Register All 6 New Handlers + Wire Dependencies

**Files:**
- Modify: `business/sdk/workflow/workflowactions/register.go`
- Modify: `api/cmd/services/ichor/build/all/all.go`

**Read first:**
- `business/sdk/workflow/workflowactions/register.go` — current `BusDependencies` and `RegisterGranularInventoryActions`, `RegisterProcurementActions`
- `api/cmd/services/ichor/build/all/all.go` — how `ActionConfig.Buses` is populated

**Step 1: Add missing bus fields to `BusDependencies`**

In `register.go`, add to `BusDependencies`:

```go
type BusDependencies struct {
    // Inventory domain
    InventoryItem          *inventoryitembus.Business
    InventoryLocation      *inventorylocationbus.Business
    InventoryTransaction   *inventorytransactionbus.Business
    InventoryAdjustment    *inventoryadjustmentbus.Business  // ADD
    TransferOrder          *transferorderbus.Business        // ADD
    Product                *productbus.Business
    Workflow               *workflow.Business

    // ... rest unchanged ...
}
```

Add the imports:
```go
"github.com/timmaaaz/ichor/business/domain/inventory/inventoryadjustmentbus"
"github.com/timmaaaz/ichor/business/domain/inventory/transferorderbus"
```

**Step 2: Register new handlers in `RegisterGranularInventoryActions`**

```go
func RegisterGranularInventoryActions(registry *workflow.ActionRegistry, config ActionConfig) {
    // ... existing registrations ...

    if config.Buses.InventoryAdjustment != nil {
        registry.Register(inventory.NewApproveInventoryAdjustmentHandler(config.Log, config.Buses.InventoryAdjustment))
        registry.Register(inventory.NewRejectInventoryAdjustmentHandler(config.Log, config.Buses.InventoryAdjustment))
    }

    if config.Buses.TransferOrder != nil {
        registry.Register(inventory.NewApproveTransferOrderHandler(config.Log, config.Buses.TransferOrder))
        registry.Register(inventory.NewRejectTransferOrderHandler(config.Log, config.Buses.TransferOrder))
    }
}
```

**Step 3: Register PO handlers in `RegisterProcurementActions`**

```go
func RegisterProcurementActions(registry *workflow.ActionRegistry, config ActionConfig) {
    if config.Buses.PurchaseOrder != nil {
        registry.Register(procurement.NewCreatePurchaseOrderHandler(...)) // existing
        registry.Register(procurement.NewApprovePurchaseOrderHandler(config.Log, config.Buses.PurchaseOrder)) // ADD
        registry.Register(procurement.NewRejectPurchaseOrderHandler(config.Log, config.Buses.PurchaseOrder))  // ADD
    }
}
```

**Step 4: Wire new buses in `all.go`**

In `all.go`, find where `ActionConfig.Buses` is populated and add:

```go
Buses: workflowactions.BusDependencies{
    // ... existing fields ...
    InventoryAdjustment: inventoryAdjustmentBus,  // ADD (verify variable name)
    TransferOrder:       transferOrderBus,          // ADD (verify variable name)
},
```

Check the variable names by searching:
```bash
grep -n "inventoryAdjustmentBus\|transferOrderBus\|InventoryAdjustment\|TransferOrder" \
    api/cmd/services/ichor/build/all/all.go
```

**Step 5: Build check**

```bash
go build ./business/sdk/workflow/workflowactions/...
go build ./api/cmd/services/ichor/...
```

Expected: no errors.

**Step 6: Commit**

```bash
git add business/sdk/workflow/workflowactions/register.go \
        api/cmd/services/ichor/build/all/all.go
git commit -m "feat(workflow): register approval handlers and wire new bus dependencies"
```

---

## Task 12: Add Delegate Support to approvalrequestbus

**Files:**
- Modify: `business/domain/workflow/approvalrequestbus/approvalrequestbus.go`
- Create: `business/domain/workflow/approvalrequestbus/event.go`

**Read first:**
- `business/domain/workflow/approvalrequestbus/approvalrequestbus.go` — current struct (no delegate field)
- `business/domain/inventory/inventoryadjustmentbus/approvalrequestbus.go` does NOT have delegate — compare with a bus that does, e.g. `business/domain/inventory/putawaytaskbus/putawaytaskbus.go` for the full delegate pattern

**Why this is needed:** `approvalrequestbus.Business` currently has no `*delegate.Delegate` field and no `event.go`. Adding `delegate.Call()` to `Resolve()` requires adding the full delegate infrastructure first.

**Step 1: Create `event.go`**

Create `business/domain/workflow/approvalrequestbus/event.go` following the exact pattern of any other `event.go` in the codebase (e.g., `business/domain/inventory/putawaytaskbus/event.go`):

```go
package approvalrequestbus

import "github.com/timmaaaz/ichor/business/sdk/delegate"

// Action event constants.
const (
    ActionUpdated = "approval_request_updated"
)

// ActionUpdatedData creates the delegate data for an approval request update.
func ActionUpdatedData(before, after ApprovalRequest) delegate.Data {
    return delegate.Data{
        Domain: "workflow",
        Action: ActionUpdated,
        RawData: map[string]any{
            "before": before,
            "after":  after,
        },
    }
}
```

**Important:** Read an existing `event.go` first (e.g., `business/domain/inventory/putawaytaskbus/event.go`) to match the exact `delegate.Data` construction pattern used in this codebase.

**Step 2: Add `delegate` field to `Business` struct**

```go
import "github.com/timmaaaz/ichor/business/sdk/delegate"

type Business struct {
    log      *logger.Logger
    storer   Storer
    delegate *delegate.Delegate  // ADD
}

func NewBusiness(log *logger.Logger, del *delegate.Delegate, storer Storer) *Business {
    return &Business{
        log:      log,
        delegate: del,    // ADD
        storer:   storer,
    }
}
```

**Step 3: Find and update all callers of `approvalrequestbus.NewBusiness`**

```bash
grep -rn "approvalrequestbus.NewBusiness" /Users/jaketimmer/src/work/superior/ichor/ichor --include="*.go"
```

For each caller, add the delegate parameter. In `all.go` and `dbtest.go`, the delegate is typically the shared `delegate.New()` instance. Check how other buses receive the delegate in those files.

**Step 4: Build check**

```bash
go build ./business/domain/workflow/approvalrequestbus/...
go build ./business/sdk/dbtest/...
go build ./api/cmd/services/ichor/...
```

**Step 5: Commit**

```bash
git add business/domain/workflow/approvalrequestbus/
git commit -m "feat(workflow): add delegate support infrastructure to approvalrequestbus"
```

---

## Task 13: Add `delegate.Call()` to `approvalrequestbus.Resolve()`

**Files:**
- Modify: `business/domain/workflow/approvalrequestbus/approvalrequestbus.go`

**Step 1: Write failing test**

In a new file `business/domain/workflow/approvalrequestbus/approvalrequestbus_test.go` (or add to existing), write a test that:
1. Creates an approval request
2. Calls `Resolve()`
3. Verifies a delegate event was captured (use the delegate capture pattern — check how other bus tests capture delegate events, e.g. via a test delegate)

**Step 2: Run test to verify it fails**

```bash
go test ./business/domain/workflow/approvalrequestbus/... -run Test_Resolve_FiresDelegateEvent -v
```

Expected: FAIL — no delegate event fired.

**Step 3: Add `delegate.Call()` to `Resolve()`**

```go
func (b *Business) Resolve(ctx context.Context, id, resolvedBy uuid.UUID, status, reason string) (ApprovalRequest, error) {
    ctx, span := otel.AddSpan(ctx, "business.approvalrequestbus.resolve")
    defer span.End()

    req, err := b.storer.Resolve(ctx, id, resolvedBy, status, reason)
    if err != nil {
        return ApprovalRequest{}, fmt.Errorf("resolve approval request: %w", err)
    }

    // Fire delegate event — needed for reactive workflow triggers on approval resolution.
    if err := b.delegate.Call(ctx, ActionUpdatedData(ApprovalRequest{ID: id}, req)); err != nil {
        b.log.Error(ctx, "approvalrequestbus: delegate call failed on resolve", "err", err)
    }

    return req, nil
}
```

Note: `ActionUpdatedData` takes `before, after`. Since `Resolve()` only returns the post-state, pass a minimal before with just the ID. Alternatively, do a `QueryByID` before the resolve to get the full before state — check what other bus `Update()` patterns do.

**Step 4: Run test to verify it passes**

```bash
go test ./business/domain/workflow/approvalrequestbus/... -run Test_Resolve_FiresDelegateEvent -v
```

Expected: PASS.

**Step 5: Build check**

```bash
go build ./business/domain/workflow/approvalrequestbus/...
```

**Step 6: Commit**

```bash
git add business/domain/workflow/approvalrequestbus/approvalrequestbus.go
git commit -m "fix(workflow): fire delegate event in approvalrequestbus.Resolve() — closes silent event gap"
```

---

## Task 14: `resolve_approval_request` Action Handler

**Files:**
- Create: `business/sdk/workflow/workflowactions/approval/resolve.go`
- Create: `business/sdk/workflow/workflowactions/approval/resolve_test.go`

**Read first:**
- `business/sdk/workflow/workflowactions/approval/` — see existing `seek_approval` handler for package conventions
- `business/domain/workflow/approvalrequestbus/approvalrequestbus.go` — `Resolve(ctx, id, resolvedBy, status, reason)` signature
- `business/domain/workflow/approvalrequestbus/model.go` — `ApprovalRequest` struct and status constants

**Step 1: Create `resolve.go`**

```go
package approval

import (
    "context"
    "encoding/json"
    "errors"
    "fmt"

    "github.com/google/uuid"
    "github.com/timmaaaz/ichor/business/domain/workflow/approvalrequestbus"
    "github.com/timmaaaz/ichor/business/sdk/workflow"
    "github.com/timmaaaz/ichor/foundation/logger"
)

type ResolveApprovalConfig struct {
    ApprovalRequestID string `json:"approval_request_id"` // required
    Resolution        string `json:"resolution"`           // required: "approved" | "rejected"
    Reason            string `json:"reason,omitempty"`     // optional audit trail
}

type ResolveApprovalHandler struct {
    log                 *logger.Logger
    approvalRequestBus  *approvalrequestbus.Business
}

func NewResolveApprovalHandler(log *logger.Logger, approvalRequestBus *approvalrequestbus.Business) *ResolveApprovalHandler {
    return &ResolveApprovalHandler{log: log, approvalRequestBus: approvalRequestBus}
}

func (h *ResolveApprovalHandler) GetType() string        { return "resolve_approval_request" }
func (h *ResolveApprovalHandler) IsAsync() bool          { return true }
func (h *ResolveApprovalHandler) SupportsManualExecution() bool { return true }
func (h *ResolveApprovalHandler) GetDescription() string {
    return "Programmatically resolve an open approval request — enables cross-workflow orchestration where one workflow closes an approval another workflow is waiting on"
}

func (h *ResolveApprovalHandler) Validate(config json.RawMessage) error {
    var cfg ResolveApprovalConfig
    if err := json.Unmarshal(config, &cfg); err != nil {
        return fmt.Errorf("invalid config: %w", err)
    }
    if cfg.ApprovalRequestID == "" {
        return fmt.Errorf("approval_request_id is required")
    }
    if _, err := uuid.Parse(cfg.ApprovalRequestID); err != nil {
        return fmt.Errorf("invalid approval_request_id: %w", err)
    }
    if cfg.Resolution != "approved" && cfg.Resolution != "rejected" {
        return fmt.Errorf("resolution must be 'approved' or 'rejected', got %q", cfg.Resolution)
    }
    return nil
}

func (h *ResolveApprovalHandler) GetOutputPorts() []workflow.OutputPort {
    return []workflow.OutputPort{
        {Name: "resolved_approved", Description: "Approval request was resolved as approved"},
        {Name: "resolved_rejected", Description: "Approval request was resolved as rejected"},
        {Name: "not_found", Description: "Approval request not found"},
        {Name: "already_resolved", Description: "Approval request was already resolved"},
        {Name: "failure", Description: "Unexpected error"},
    }
}

func (h *ResolveApprovalHandler) GetEntityModifications(config json.RawMessage) []workflow.EntityModification {
    return []workflow.EntityModification{
        {EntityName: "workflow.approval_requests", EventType: "on_update", Fields: []string{"status"}},
    }
}

func (h *ResolveApprovalHandler) Execute(ctx context.Context, config json.RawMessage, execCtx workflow.ActionExecutionContext) (any, error) {
    if h.approvalRequestBus == nil {
        return map[string]any{"output": "failure", "error": "approval request bus not available"}, nil
    }

    var cfg ResolveApprovalConfig
    if err := json.Unmarshal(config, &cfg); err != nil {
        return map[string]any{"output": "failure", "error": err.Error()}, nil
    }

    id, _ := uuid.Parse(cfg.ApprovalRequestID)

    req, err := h.approvalRequestBus.QueryByID(ctx, id)
    if err != nil {
        if errors.Is(err, approvalrequestbus.ErrNotFound) {
            return map[string]any{"output": "not_found", "approval_request_id": cfg.ApprovalRequestID}, nil
        }
        return nil, fmt.Errorf("query approval request: %w", err)
    }

    if req.Status != approvalrequestbus.StatusPending {
        return map[string]any{
            "output":              "already_resolved",
            "approval_request_id": cfg.ApprovalRequestID,
            "current_status":      string(req.Status),
        }, nil
    }

    resolved, err := h.approvalRequestBus.Resolve(ctx, id, execCtx.UserID, cfg.Resolution, cfg.Reason)
    if err != nil {
        if errors.Is(err, approvalrequestbus.ErrAlreadyResolved) {
            return map[string]any{"output": "already_resolved", "approval_request_id": cfg.ApprovalRequestID}, nil
        }
        return nil, fmt.Errorf("resolve approval request: %w", err)
    }

    outputPort := "resolved_approved"
    if cfg.Resolution == "rejected" {
        outputPort = "resolved_rejected"
    }

    return map[string]any{
        "output":              outputPort,
        "approval_request_id": resolved.ID.String(),
        "resolution":          cfg.Resolution,
    }, nil
}
```

**Step 2: Write and run tests**

In `resolve_test.go`, write `unitest.Table` covering:

Validate tests:
- Missing `approval_request_id` → error
- Invalid UUID → error
- Invalid resolution value → error
- Valid `approved` → nil
- Valid `rejected` → nil

Execute tests (need seeded approval request in pending state):
- Pending request + `approved` → `resolved_approved` port
- Pending request + `rejected` → `resolved_rejected` port
- Unknown ID → `not_found` port
- Already-resolved request → `already_resolved` port

```bash
go test ./business/sdk/workflow/workflowactions/approval/... -run Test_ResolveApprovalRequest -v
```

Expected: PASS.

**Step 3: Build check**

```bash
go build ./business/sdk/workflow/workflowactions/approval/...
```

**Step 4: Commit**

```bash
git add business/sdk/workflow/workflowactions/approval/resolve.go \
        business/sdk/workflow/workflowactions/approval/resolve_test.go
git commit -m "feat(workflow): add resolve_approval_request action handler"
```

---

## Task 15: Register `resolve_approval_request` Handler

**Files:**
- Modify: `business/sdk/workflow/workflowactions/register.go`

**Step 1: Add to `RegisterAll`**

In the Approval actions block:
```go
// Approval actions
registry.Register(approval.NewSeekApprovalHandler(config.Log, config.DB, config.Buses.ApprovalRequest, config.Buses.Alert))
registry.Register(approval.NewResolveApprovalHandler(config.Log, config.Buses.ApprovalRequest))  // ADD
```

**Step 2: Add to `RegisterCoreActions` (nil-guarded)**

```go
// Approval actions - nil buses for core path (graceful degradation)
registry.Register(approval.NewSeekApprovalHandler(log, db, nil, nil))
registry.Register(approval.NewResolveApprovalHandler(log, nil))  // ADD — nil bus returns failure gracefully
```

**Step 3: Build check**

```bash
go build ./business/sdk/workflow/workflowactions/...
go build ./api/cmd/services/ichor/...
```

**Step 4: Verify handler appears in action types**

```bash
# In a running dev environment:
make token
curl -H "Authorization: Bearer $TOKEN" http://localhost:3000/v1/workflow/action-types | jq '.[] | select(.type == "resolve_approval_request")'
```

**Step 5: Commit**

```bash
git add business/sdk/workflow/workflowactions/register.go
git commit -m "feat(workflow): register resolve_approval_request in all and core action sets"
```

---

## Task 16: Field Schema Registry

**Files:**
- Create: `business/sdk/workflow/fieldschema/registry.go`

**Step 1: Create the registry**

```go
package fieldschema

// FieldSchema describes a single field on an entity that has known discrete values.
type FieldSchema struct {
    Name        string   `json:"name"`
    Type        string   `json:"type"`             // "enum", "string", "uuid", etc.
    Values      []string `json:"values,omitempty"` // only for type="enum"
    Description string   `json:"description,omitempty"`
}

// EntitySchema groups the field schemas for one entity.
type EntitySchema struct {
    Entity string        `json:"entity"`
    Fields []FieldSchema `json:"fields"`
}

// KnownEnumFields maps DB entity names (schema.table format) to their registered
// enum field schemas. This is the authoritative list of discoverable enum values
// for workflow trigger field_conditions.
//
// NOTE: When adding new status/enum fields to any bus model, also add an entry here.
var KnownEnumFields = map[string][]FieldSchema{
    "inventory.put_away_tasks": {
        {Name: "status", Type: "enum", Values: []string{"pending", "in_progress", "completed", "cancelled"}, Description: "Lifecycle state of the putaway task"},
    },
    "inventory.inventory_adjustments": {
        {Name: "approval_status", Type: "enum", Values: []string{"pending", "approved", "rejected"}, Description: "Approval state of the inventory adjustment"},
    },
    "inventory.lot_trackings": {
        {Name: "quality_status", Type: "enum", Values: []string{"good", "on_hold", "quarantined", "released", "expired"}, Description: "Quality control state of the lot"},
    },
    "workflow.alerts": {
        {Name: "status", Type: "enum", Values: []string{"active", "acknowledged", "dismissed", "resolved"}, Description: "Alert lifecycle state"},
        {Name: "severity", Type: "enum", Values: []string{"low", "medium", "high", "critical"}, Description: "Alert severity level"},
    },
    "workflow.approval_requests": {
        {Name: "status", Type: "enum", Values: []string{"pending", "approved", "rejected", "timed_out", "expired"}, Description: "Approval request resolution state"},
        {Name: "approval_type", Type: "enum", Values: []string{"any", "all", "majority"}, Description: "Required approval quorum type"},
    },
}

// GetEntitySchema returns the field schemas for the given entity, or nil if unknown.
func GetEntitySchema(entity string) ([]FieldSchema, bool) {
    fields, ok := KnownEnumFields[entity]
    return fields, ok
}
```

**Step 2: Build check**

```bash
go build ./business/sdk/workflow/fieldschema/...
```

**Step 3: Commit**

```bash
git add business/sdk/workflow/fieldschema/
git commit -m "feat(workflow): add field schema registry for enum discovery"
```

---

## Task 17: Field Schema Discovery Endpoint

**Files:**
- Modify: `api/domain/http/workflow/referenceapi/referenceapi.go`
- Modify: `api/domain/http/workflow/referenceapi/route.go`

**Read first:**
- `api/domain/http/workflow/referenceapi/referenceapi.go` — see existing handler pattern (e.g., `queryTriggerTypes`)
- `api/domain/http/workflow/referenceapi/route.go` — see route registration pattern
- `business/sdk/workflow/workflow.go` — see if `Business` has an entity catalog method, or check how `queryEntities` works in referenceapi.go to understand how entity names are validated

**Step 1: Add handler to `referenceapi.go`**

```go
// queryEntityFields handles GET /v1/workflow/entities/{entity}/fields
// Returns the known enum field schemas for the given entity.
// 200 with fields array — entity known (may be empty if no registered enums)
// 404 — entity not registered in the workflow entity catalog at all
func (a *api) queryEntityFields(ctx context.Context, r *http.Request) web.Encoder {
    entityName := web.Param(r, "entity")

    // Validate: check entity exists in the workflow entity catalog.
    // (Look at how queryEntities works to find the entity lookup method.)
    // If entity does not exist in catalog → return 404.
    // ... entity validation logic here ...

    fields, _ := fieldschema.GetEntitySchema(entityName)
    if fields == nil {
        fields = []fieldschema.FieldSchema{} // empty array, not null
    }

    return fieldschema.EntitySchema{
        Entity: entityName,
        Fields: fields,
    }
}
```

**Important:** Read the existing `queryEntities` handler first to understand how entity validation works (checking if an entity exists in the catalog). Use the same mechanism to validate `{entity}` before returning field schemas.

**Step 2: Add import**

```go
import "github.com/timmaaaz/ichor/business/sdk/workflow/fieldschema"
```

**Step 3: Register route in `route.go`**

```go
app.HandlerFunc(http.MethodGet, version, "/workflow/entities/{entity}/fields", api.queryEntityFields, authen)
```

Place it alongside the existing `/workflow/entities` route.

**Step 4: Build check**

```bash
go build ./api/domain/http/workflow/referenceapi/...
go build ./api/cmd/services/ichor/...
```

**Step 5: Commit**

```bash
git add api/domain/http/workflow/referenceapi/referenceapi.go \
        api/domain/http/workflow/referenceapi/route.go
git commit -m "feat(workflow): add GET /workflow/entities/{entity}/fields discovery endpoint"
```

---

## Task 18: Field Schema Discovery — Integration Tests

**Files:**
- Create: `api/cmd/services/ichor/tests/workflow/referenceapi/fields_test.go` (or add to existing test file)

**Read first:**
- Look for existing test files in `api/cmd/services/ichor/tests/workflow/` to understand the test structure
- `api/cmd/services/ichor/tests/inventory/putawaytaskapi/query_test.go` — canonical apitest.Table pattern

**Step 1: Write the tests**

```go
func fieldSchema200(sd apitest.SeedData) []apitest.Table {
    return []apitest.Table{
        {
            Name:       "known-enum-entity",
            URL:        "/v1/workflow/entities/inventory.put_away_tasks/fields",
            Token:      sd.Admins[0].Token,
            Method:     http.MethodGet,
            StatusCode: http.StatusOK,
            GotResp:    &fieldschema.EntitySchema{},
            ExpResp: &fieldschema.EntitySchema{
                Entity: "inventory.put_away_tasks",
                Fields: []fieldschema.FieldSchema{
                    {Name: "status", Type: "enum", Values: []string{"pending", "in_progress", "completed", "cancelled"}},
                },
            },
            CmpFunc: func(got, exp any) string {
                // Compare entity and field names; ignore description for brevity
                return cmp.Diff(got, exp, cmpopts.IgnoreFields(fieldschema.FieldSchema{}, "Description"))
            },
        },
        {
            Name:       "entity-with-no-registered-enums",
            URL:        "/v1/workflow/entities/inventory.inventory_items/fields",
            Token:      sd.Admins[0].Token,
            Method:     http.MethodGet,
            StatusCode: http.StatusOK,
            GotResp:    &fieldschema.EntitySchema{},
            ExpResp:    &fieldschema.EntitySchema{Entity: "inventory.inventory_items", Fields: []fieldschema.FieldSchema{}},
            CmpFunc:    func(got, exp any) string { return cmp.Diff(got, exp) },
        },
    }
}

func fieldSchema404(sd apitest.SeedData) []apitest.Table {
    return []apitest.Table{
        {
            Name:       "nonexistent-entity",
            URL:        "/v1/workflow/entities/nonexistent.table/fields",
            Token:      sd.Admins[0].Token,
            Method:     http.MethodGet,
            StatusCode: http.StatusNotFound,
            GotResp:    &errs.Error{},
            ExpResp:    errs.Newf(errs.NotFound, "entity not found"),
            CmpFunc:    func(got, exp any) string { return cmp.Diff(got, exp) },
        },
    }
}
```

**Step 2: Run tests**

```bash
go test ./api/cmd/services/ichor/tests/workflow/... -run TestFieldSchema -v
```

Expected: PASS.

**Step 3: Commit**

```bash
git add api/cmd/services/ichor/tests/workflow/referenceapi/
git commit -m "test(workflow): add integration tests for entity field schema discovery endpoint"
```

---

## Final Verification

**Step 1: Build everything**

```bash
go build ./...
```

**Step 2: Run all changed package tests**

```bash
go test ./business/domain/inventory/inventoryadjustmentbus/...
go test ./business/domain/inventory/transferorderbus/...
go test ./business/domain/procurement/purchaseorderbus/...
go test ./business/domain/workflow/approvalrequestbus/...
go test ./business/sdk/workflow/workflowactions/inventory/...
go test ./business/sdk/workflow/workflowactions/procurement/...
go test ./business/sdk/workflow/workflowactions/approval/...
go test ./business/sdk/workflow/fieldschema/...
go test ./api/cmd/services/ichor/tests/workflow/...
```

**Step 3: Verify action types endpoint**

In a running dev environment, confirm all 7 new action types are discoverable:
- `approve_inventory_adjustment`, `reject_inventory_adjustment`
- `approve_transfer_order`, `reject_transfer_order`
- `approve_purchase_order`, `reject_purchase_order`
- `resolve_approval_request`

```bash
make token
curl -s -H "Authorization: Bearer $TOKEN" http://localhost:3000/v1/workflow/action-types | jq '[.[] | select(.type | startswith("approve_") or startswith("reject_") or . == "resolve_approval_request")] | length'
```

Expected: `7`

**Step 4: Verify field schema endpoint**

```bash
curl -s -H "Authorization: Bearer $TOKEN" \
    "http://localhost:3000/v1/workflow/entities/inventory.put_away_tasks/fields" | jq .
```

Expected: status field with 4 enum values.
