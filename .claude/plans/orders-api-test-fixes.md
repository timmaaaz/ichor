# Orders API Test Fixes Plan

## Problem Analysis

The `Test_Order` tests are failing due to a mismatch between test expectations and actual API behavior. After analyzing the codebase, I've identified **two categories of issues**:

### Issue 1: Missing `order_date` in Test Inputs

**Root Cause**: The `NewOrder` model has `OrderDate` marked as `validate:"required"`, but the test inputs don't provide it.

**Evidence**:
- [model.go:125](app/domain/sales/ordersapp/model.go#L125): `OrderDate string json:"order_date" validate:"required"`
- The `create200` "basic" test doesn't include `OrderDate` in the input
- All `create400` tests also omit `OrderDate`, causing validation to fail on **two** fields instead of one

**Design Decision Needed**: Should `OrderDate` be:
1. **Required (current)**: Client must provide the order date
2. **Auto-populated**: System defaults to current date if not provided (like `created_date`)

Looking at how this is used:
- `ordersbus.TestSeedOrders` always provides `OrderDate` explicitly
- The business layer doesn't have auto-population logic for `OrderDate` (unlike `CreatedDate`)
- Similar pattern exists in `purchaseorderapp` which also marks `order_date` as required

**Recommendation**: Keep `OrderDate` as required. The order date represents when the order was placed, which is a business-meaningful date that should be explicitly provided (it might differ from the current date for backdated orders, imports, etc.).

### Issue 2: Incomplete `ExpResp` in Update Test

**Root Cause**: The `update200` "basic" test only specifies a few fields in `ExpResp`, but the API returns the full `Order` object with all fields populated.

**Evidence from test output**:
```diff
OrderDate:           "",      // ExpResp (empty)
OrderDate:           "2026-01-19",  // Got (actual value)
Subtotal:            "",      // ExpResp (empty)
Subtotal:            "189.00",     // Got (actual value)
... (many more fields)
```

The test only sets these fields in `ExpResp`:
- `ID`, `Number`, `CustomerID`, `FulfillmentStatusID`
- `CreatedBy`, `UpdatedBy`, `DueDate`
- `CreatedDate`, `UpdatedDate`

But the response includes: `OrderDate`, `Subtotal`, `TaxRate`, `TaxAmount`, `ShippingCost`, `TotalAmount`, `Currency`, `Notes`

---

## Fix Plan

### Step 1: Fix Create Tests (add `OrderDate` to all test inputs)

**File**: [create_test.go](api/cmd/services/ichor/tests/sales/ordersapi/create_test.go)

**Changes**:

1. **create200 "basic" test** (lines 21-27): Add `OrderDate` to input
   ```go
   Input: &ordersapp.NewOrder{
       Number:              "ORD-12345",
       CustomerID:          sd.Customers[0].ID,
       FulfillmentStatusID: sd.OrderFulfillmentStatuses[0].ID,
       CreatedBy:           sd.Admins[0].ID.String(),
       DueDate:             time.Now().Add(3 * 24 * time.Hour).Format("2006-01-02"),
       OrderDate:           time.Now().Format("2006-01-02"),  // ADD THIS
   },
   ```

2. **create200 ExpResp** (lines 29-36): Add `OrderDate` to expected response
   ```go
   ExpResp: &ordersapp.Order{
       Number:              "ORD-12345",
       CustomerID:          sd.Customers[0].ID,
       FulfillmentStatusID: sd.OrderFulfillmentStatuses[0].ID,
       CreatedBy:           sd.Admins[0].ID.String(),
       UpdatedBy:           sd.Admins[0].ID.String(),
       DueDate:             time.Now().Add(3 * 24 * time.Hour).Format("2006-01-02"),
       OrderDate:           time.Now().Format("2006-01-02"),  // ADD THIS
   },
   ```

3. **create400 "missing number"** (lines 64-68): Add `OrderDate` to input
4. **create400 "missing customer id"** (lines 86-90): Add `OrderDate` to input
5. **create400 "missing due date"** (lines 108-113): Add `OrderDate` to input
6. **create400 "missing fulfillment status id"** (lines 130-134): Add `OrderDate` to input
7. **create400 "missing created by"** (lines 152-156): Add `OrderDate` to input

### Step 2: Fix Update Test (complete the `ExpResp`)

**File**: [update_test.go](api/cmd/services/ichor/tests/sales/ordersapi/update_test.go)

**Changes**:

1. **update200 "basic" ExpResp** (lines 26-35): Add all the missing fields from `sd.Orders[0]`
   ```go
   ExpResp: &ordersapp.Order{
       ID:                  sd.Orders[0].ID,
       Number:              sd.Orders[0].Number,
       CustomerID:          sd.Customers[1].ID,  // This is updated
       FulfillmentStatusID: sd.Orders[0].FulfillmentStatusID,
       OrderDate:           sd.Orders[0].OrderDate,      // ADD
       Subtotal:            sd.Orders[0].Subtotal,       // ADD
       TaxRate:             sd.Orders[0].TaxRate,        // ADD
       TaxAmount:           sd.Orders[0].TaxAmount,      // ADD
       ShippingCost:        sd.Orders[0].ShippingCost,   // ADD
       TotalAmount:         sd.Orders[0].TotalAmount,    // ADD
       Currency:            sd.Orders[0].Currency,       // ADD
       Notes:               sd.Orders[0].Notes,          // ADD
       CreatedBy:           sd.Orders[0].CreatedBy,
       UpdatedBy:           sd.Orders[0].UpdatedBy,
       DueDate:             sd.Orders[0].DueDate,
       CreatedDate:         sd.Orders[0].CreatedDate,
       UpdatedDate:         sd.Orders[0].UpdatedDate,
   },
   ```

---

## Summary of Changes

| File | Test Case | Change |
|------|-----------|--------|
| `create_test.go` | `create200` basic | Add `OrderDate` to Input and ExpResp |
| `create_test.go` | `create400` missing number | Add `OrderDate` to Input |
| `create_test.go` | `create400` missing customer id | Add `OrderDate` to Input |
| `create_test.go` | `create400` missing due date | Add `OrderDate` to Input |
| `create_test.go` | `create400` missing fulfillment status id | Add `OrderDate` to Input |
| `create_test.go` | `create400` missing created by | Add `OrderDate` to Input |
| `update_test.go` | `update200` basic | Add missing fields to ExpResp |

## API Behavior Verification

These changes ensure tests match the actual API behavior:
- **Create**: Requires `order_date` as a mandatory field (intentional design)
- **Update**: Returns complete `Order` object with all fields (standard behavior)

No changes to the API implementation are needed - the tests simply need to match the existing behavior.
