# Phase 2: `approve_purchase_order` Action Handler

## REVISED SCOPE

After critical review, only `approve_purchase_order` moves forward in this phase.

The other proposed handlers have been **deferred**:
- `approve_inventory_adjustment` / `reject_inventory_adjustment` — marginal value; `transition_status` covers the status flip (except delegate event). Build when a concrete workflow chain requires the delegate event.
- `approve_transfer_order` — same marginal value profile. Deferred.

**Why `approve_purchase_order` clears the threshold** and the others don't:

The PO status field is a UUID foreign key into `procurement.purchase_order_statuses` — a lookup
table. Generic actions cannot abstract this: `transition_status` works on string status fields,
and `update_field` would require rule authors to hardcode a deployment-specific UUID for
"approved". That's structurally broken. Additionally, `purchaseorderbus.Approve()` performs a
4-field atomic write (status, approved_by, approved_date, updated_by, updated_date) that cannot
be expressed as a single generic action node without splitting into multiple nodes with a race
window between them.

**The marginal handlers** (adjustment, transfer order) have string status fields that
`transition_status` can handle. Their only gap is delegate event propagation — worth building
only when a downstream reactive rule concretely requires distinguishing workflow-originated
approvals from UI-originated approvals.

## Objective

Add `approve_purchase_order` to the procurement action handler package.

## Pre-Work: Read Before Implementing

1. Read `business/domain/inventory/inventoryadjustmentbus/inventoryadjustmentbus.go` — Approve() and Reject() signatures
2. Read `business/domain/inventory/transferorderbus/transferorderbus.go` — Approve() signature
3. Read `business/domain/procurement/purchaseorderbus/purchaseorderbus.go` — Approve() signature
4. Note: all three currently fire `ActionUpdatedData` delegate events after approval — the handlers do not need to re-fire them (bus already does it)

## Handler Specifications

### Handler 1: `approve_inventory_adjustment`

**File**: `business/sdk/workflow/workflowactions/inventory/approve_adjustment.go`

```
Config: ApproveAdjustmentConfig
  - AdjustmentID string `json:"adjustment_id"` (required)
  - ApprovalReason string `json:"approval_reason,omitempty"`

GetType() → "approve_inventory_adjustment"
IsAsync() → false
SupportsManualExecution() → true
GetDescription() → "Approve a pending inventory adjustment (stock correction)"
Output ports: approved, not_found, already_approved, rejected_state, failure
GetEntityModifications() → inventory.inventory_adjustments (on_update, fields: [approval_status])

Execute logic:
  1. Parse + validate adjustment_id UUID
  2. QueryByID → not_found if missing
  3. If approval_status == "approved" → already_approved (idempotent)
  4. If approval_status == "rejected" → rejected_state (cannot re-approve)
  5. Call bus.Approve(ctx, adjustment, execCtx.UserID) → approved
  6. Error → failure
```

### Handler 2: `reject_inventory_adjustment`

**File**: `business/sdk/workflow/workflowactions/inventory/reject_adjustment.go`

```
Config: RejectAdjustmentConfig
  - AdjustmentID string `json:"adjustment_id"` (required)
  - RejectionReason string `json:"rejection_reason"` (required — audit trail)

GetType() → "reject_inventory_adjustment"
IsAsync() → false
SupportsManualExecution() → true
GetDescription() → "Reject a pending inventory adjustment (stock correction)"
Output ports: rejected, not_found, already_rejected, approved_state, failure
GetEntityModifications() → inventory.inventory_adjustments (on_update, fields: [approval_status])

Execute logic: mirror of approve handler with Reject() call
```

### Handler 3: `approve_transfer_order`

**File**: `business/sdk/workflow/workflowactions/inventory/approve_transfer_order.go`

```
Config: ApproveTransferOrderConfig
  - TransferOrderID string `json:"transfer_order_id"` (required)
  - ApprovalReason string `json:"approval_reason,omitempty"`

GetType() → "approve_transfer_order"
IsAsync() → false
SupportsManualExecution() → true
GetDescription() → "Approve an inter-location inventory transfer order"
Output ports: approved, not_found, already_approved, failure
GetEntityModifications() → inventory.transfer_orders (on_update, fields: [status])

BusDependency: transferorderbus.Business (new — currently not in BusDependencies)
```

### Handler 4: `approve_purchase_order`

**File**: `business/sdk/workflow/workflowactions/procurement/approve_po.go`

```
Config: ApprovePurchaseOrderConfig
  - PurchaseOrderID string `json:"purchase_order_id"` (required)
  - ApprovalReason string `json:"approval_reason,omitempty"`

GetType() → "approve_purchase_order"
IsAsync() → false
SupportsManualExecution() → true
GetDescription() → "Approve a purchase order, allowing goods receipt to proceed"
Output ports: approved, not_found, already_approved, failure
GetEntityModifications() → procurement.purchase_orders (on_update)

BusDependency: purchaseorderbus.Business (already in BusDependencies ✓)
```

## Registration Changes

**File**: `business/sdk/workflow/workflowactions/register.go`

Add to `BusDependencies`:
```go
// Inventory domain (new)
TransferOrder *transferorderbus.Business
// InventoryAdjustment *inventoryadjustmentbus.Business  (need to check if it exists)
```

Add to `RegisterGranularInventoryActions()`:
```go
if config.Buses.InventoryAdjustment != nil {
    registry.Register(inventory.NewApproveAdjustmentHandler(...))
    registry.Register(inventory.NewRejectAdjustmentHandler(...))
}
if config.Buses.TransferOrder != nil {
    registry.Register(inventory.NewApproveTransferOrderHandler(...))
}
```

Add to `RegisterProcurementActions()`:
```go
if config.Buses.PurchaseOrder != nil {
    // existing create_purchase_order registration...
    registry.Register(procurement.NewApprovePurchaseOrderHandler(
        config.Log,
        config.Buses.PurchaseOrder,
    ))
}
```

## all.go Wiring

**File**: `api/cmd/services/ichor/build/all/all.go`

Add to ActionConfig.Buses:
- `InventoryAdjustment: inventoryAdjustmentBus` (if not already present)
- `TransferOrder: transferOrderBus` (if not already present)
- `PurchaseOrder` is already wired ✓

## Unit Tests

For each handler, create a `*_test.go` file testing:
- Config validation (missing ID, invalid UUID format)
- Execute → correct output port routing (approved, not_found, already_approved, failure)
- Use mock bus or table-driven unit tests

## Verification

```bash
go build ./business/sdk/workflow/workflowactions/...
go build ./api/cmd/services/ichor/...
go test ./business/sdk/workflow/workflowactions/inventory/...
go test ./business/sdk/workflow/workflowactions/procurement/...
```

## Definition of Done

- [ ] 4 handler files created and compile
- [ ] BusDependencies updated with TransferOrder (and InventoryAdjustment if needed)
- [ ] All handlers registered in appropriate Register*() functions
- [ ] Wired in all.go
- [ ] Unit tests for all 4 handlers
- [ ] `GET /v1/workflow/action-types` returns all 4 new types
- [ ] `go build` passes on all affected packages
