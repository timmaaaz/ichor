# Adjustments Reason Code Validation Implementation Plan

> **For agentic workers:** REQUIRED: Use superpowers:subagent-driven-development (if subagents available) or superpowers:executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.
>
> **Worktree:** Create a worktree before executing: `create a worktree for blocker-006-reason-codes and execute this plan`

**Goal:** Validate inventory adjustment reason codes at all three layers (business constants, app validate tags, DB CHECK constraint) so invalid codes are rejected.

**Architecture:** Option A (lightweight): Add ValidReasonCodes map in business layer, oneof tag in app model, and CHECK constraint migration. No new domain needed.

**Tech Stack:** Go 1.23, PostgreSQL, Ardan Labs service architecture

---

## Valid Reason Codes

```
damaged | theft | data_entry_error | receiving_error | picking_error | found_stock | other
```

---

## Step 1: Business layer constants and validation

- [ ] **Add reason code constants and validation to `inventoryadjustmentbus.go`**

**File:** `business/domain/inventory/inventoryadjustmentbus/inventoryadjustmentbus.go`

Add after the `ApprovalStatus*` constants block (after line 32):

```go
// Valid reason code values for inventory adjustments.
const (
	ReasonCodeDamaged        = "damaged"
	ReasonCodeTheft          = "theft"
	ReasonCodeDataEntryError = "data_entry_error"
	ReasonCodeReceivingError = "receiving_error"
	ReasonCodePickingError   = "picking_error"
	ReasonCodeFoundStock     = "found_stock"
	ReasonCodeOther          = "other"
)

// ValidReasonCodes is the set of known reason codes.
var ValidReasonCodes = map[string]bool{
	ReasonCodeDamaged:        true,
	ReasonCodeTheft:          true,
	ReasonCodeDataEntryError: true,
	ReasonCodeReceivingError: true,
	ReasonCodePickingError:   true,
	ReasonCodeFoundStock:     true,
	ReasonCodeOther:          true,
}
```

Add a new error variable in the existing `var` block (line 19-25):

```go
ErrInvalidReasonCode = errors.New("invalid reason code")
```

- [ ] **Add validation in `Create()` method**

**File:** `business/domain/inventory/inventoryadjustmentbus/inventoryadjustmentbus.go`

Insert before the `ia := InventoryAdjustment{` line (before line 84):

```go
if !ValidReasonCodes[nia.ReasonCode] {
    return InventoryAdjustment{}, fmt.Errorf("create: %w", ErrInvalidReasonCode)
}
```

- [ ] **Add validation in `Update()` method**

**File:** `business/domain/inventory/inventoryadjustmentbus/inventoryadjustmentbus.go`

Replace lines 146-148 with:

```go
if u.ReasonCode != nil {
    if !ValidReasonCodes[*u.ReasonCode] {
        return InventoryAdjustment{}, fmt.Errorf("update: %w", ErrInvalidReasonCode)
    }
    ia.ReasonCode = *u.ReasonCode
}
```

- [ ] **Build check**

```bash
go build ./business/domain/inventory/inventoryadjustmentbus/...
```

---

## Step 2: App layer validate tags

- [ ] **Add `oneof` tag to `NewInventoryAdjustment.ReasonCode`**

**File:** `app/domain/inventory/inventoryadjustmentapp/model.go`, line 104

Change:
```go
ReasonCode     string `json:"reason_code" validate:"required"`
```
To:
```go
ReasonCode     string `json:"reason_code" validate:"required,oneof=damaged theft data_entry_error receiving_error picking_error found_stock other"`
```

- [ ] **Add `oneof` tag to `UpdateInventoryAdjustment.ReasonCode`**

**File:** `app/domain/inventory/inventoryadjustmentapp/model.go`, line 175

Change:
```go
ReasonCode     *string `json:"reason_code" validate:"omitempty"`
```
To:
```go
ReasonCode     *string `json:"reason_code" validate:"omitempty,oneof=damaged theft data_entry_error receiving_error picking_error found_stock other"`
```

- [ ] **Build check**

```bash
go build ./app/domain/inventory/inventoryadjustmentapp/...
```

---

## Step 3: Database CHECK constraint migration

- [ ] **Add migration version 2.15**

**File:** `business/sdk/migrate/sql/migrate.sql`

Append after the last line (line 2259):

```sql

-- Version: 2.15
-- Description: Add CHECK constraint on reason_code for inventory adjustments.
UPDATE inventory.inventory_adjustments SET reason_code = 'other' WHERE reason_code NOT IN ('damaged', 'theft', 'data_entry_error', 'receiving_error', 'picking_error', 'found_stock', 'other');
ALTER TABLE inventory.inventory_adjustments ADD CONSTRAINT inventory_adjustments_reason_code_check CHECK (reason_code IN ('damaged', 'theft', 'data_entry_error', 'receiving_error', 'picking_error', 'found_stock', 'other'));
```

The `UPDATE` statement ensures existing rows with invalid reason codes (e.g., from old seed data) are migrated to `'other'` before the CHECK is applied, preventing migration failure on populated databases.

---

## Step 4: Fix test data

All existing test data uses invalid reason codes (`"Test Reason"`, `"Purchase"`, `"Adjustment"`). These must be updated to valid values.

- [ ] **Fix business layer test helper**

**File:** `business/domain/inventory/inventoryadjustmentbus/testutil.go`, line 25

Change:
```go
ReasonCode:     "Test Reason",
```
To:
```go
ReasonCode:     "other",
```

- [ ] **Fix business layer unit test — create**

**File:** `business/domain/inventory/inventoryadjustmentbus/inventoryadjustmentbus_test.go`

Line 232 — change `"Purchase"` to `"damaged"`:
```go
ReasonCode:     "damaged",
```

Line 243 — change `"Purchase"` to `"damaged"`:
```go
ReasonCode:     "damaged",
```

- [ ] **Fix business layer unit test — update**

**File:** `business/domain/inventory/inventoryadjustmentbus/inventoryadjustmentbus_test.go`

Line 284 — change `"Adjustment"` to `"theft"`:
```go
ReasonCode:            "theft",
```

Line 296 — change `"Adjustment"` to `"theft"`:
```go
ReasonCode:     dbtest.StringPointer("theft"),
```

- [ ] **Fix integration test — create_test.go**

**File:** `api/cmd/services/ichor/tests/inventory/inventoryadjustmentapi/create_test.go`

All occurrences of `ReasonCode: "Purchase"` (lines 34, 46, 82, 103, 124, 145, 166, 209, 230, 252, 274, 296, 318, 348, 370, 392, 414) must be changed to a valid reason code:
```go
ReasonCode:     "damaged",
```

Use find-and-replace across the file: `"Purchase"` -> `"damaged"` for all ReasonCode assignments.

- [ ] **Fix integration test — update_test.go**

**File:** `api/cmd/services/ichor/tests/inventory/inventoryadjustmentapi/update_test.go`

Line 34 — change:
```go
ReasonCode:     dbtest.StringPointer("Adjustment"),
```
To:
```go
ReasonCode:     dbtest.StringPointer("theft"),
```

Line 46 — change:
```go
ReasonCode:            "Adjustment",
```
To:
```go
ReasonCode:            "theft",
```

- [ ] **Update the create_test.go validation error test case**

**File:** `api/cmd/services/ichor/tests/inventory/inventoryadjustmentapi/create_test.go`, around line 192

The existing test expects a "required field" error for missing reason_code. The error message format will now include the `oneof` constraint. Verify that the existing test case where `ReasonCode` is omitted still produces the `"reason_code is a required field"` error (the `required` tag fires before `oneof`, so this should be unchanged). No change expected here, but verify.

---

## Step 5: Add business layer validation test

- [ ] **Add invalid reason code test case to unit tests**

**File:** `business/domain/inventory/inventoryadjustmentbus/inventoryadjustmentbus_test.go`

Add a test case inside the `create` function that attempts to create with an invalid reason code and asserts `ErrInvalidReasonCode`:

```go
// Test invalid reason code
invalidNIA := inventoryadjustmentbus.NewInventoryAdjustment{
    ProductID:      sd.Products[0].ID,
    LocationID:     sd.InventoryLocations[0].LocationID,
    AdjustedBy:     sd.Admins[0].ID,
    QuantityChange: 10,
    ReasonCode:     "invalid_code",
    Notes:          "Test Notes",
    AdjustmentDate: time.Now(),
}
_, err := busDomain.InventoryAdjustment.Create(ctx, invalidNIA)
if !errors.Is(err, inventoryadjustmentbus.ErrInvalidReasonCode) {
    t.Fatalf("expected ErrInvalidReasonCode, got: %v", err)
}
```

The exact seed data field names depend on the test's `unitest.SeedData` structure — adapt the field references to match the existing `create` test function's pattern.

---

## Step 6: Build and test

- [ ] **Full build of affected packages**

```bash
go build ./business/domain/inventory/inventoryadjustmentbus/... ./app/domain/inventory/inventoryadjustmentapp/...
```

Expected: clean build, no errors.

- [ ] **Run business layer unit tests**

```bash
go test ./business/domain/inventory/inventoryadjustmentbus/...
```

Expected: all tests pass, including the new invalid reason code test.

- [ ] **Run integration tests** (requires live database)

```bash
go test ./api/cmd/services/ichor/tests/inventory/inventoryadjustmentapi/...
```

Expected: all tests pass with updated reason code values.

---

## Files Modified Summary

| File | Change |
|------|--------|
| `business/domain/inventory/inventoryadjustmentbus/inventoryadjustmentbus.go` | Add constants, ValidReasonCodes map, ErrInvalidReasonCode, validation in Create/Update |
| `app/domain/inventory/inventoryadjustmentapp/model.go` | Add `oneof` validate tag to New and Update models |
| `business/sdk/migrate/sql/migrate.sql` | Version 2.15: UPDATE + CHECK constraint |
| `business/domain/inventory/inventoryadjustmentbus/testutil.go` | Fix `"Test Reason"` -> `"other"` |
| `business/domain/inventory/inventoryadjustmentbus/inventoryadjustmentbus_test.go` | Fix test reason codes + add invalid code test |
| `api/cmd/services/ichor/tests/inventory/inventoryadjustmentapi/create_test.go` | Fix `"Purchase"` -> `"damaged"` |
| `api/cmd/services/ichor/tests/inventory/inventoryadjustmentapi/update_test.go` | Fix `"Adjustment"` -> `"theft"` |
