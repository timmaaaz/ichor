# Refactor Conversion Fixes Plan

This document outlines the test failures discovered after executing the convert-refactor-master-plan.md and provides fixes for each.

## Summary of Issues

| Package | Issue | Root Cause | Fix Required |
|---------|-------|------------|--------------|
| userassetapi | create-200 and update-200 failing (400 status) | Date format mismatch: tests send Go `time.Time.String()` format but app layer expects RFC3339 | Fix test seed data to format dates as RFC3339 |
| warehouseapi | update-200 failing (400 status) | `UpdateWarehouse` struct doesn't handle optional fields correctly | Fix the `toBusUpdateWarehouse` function to handle empty/zero values |
| lottrackingsapi | Error message mismatch in malformed date tests | Tests expect old error message format, but refactored code uses new format | Update test expected error messages |
| data | execute-200 test data mismatch | Non-deterministic test data - sorted results off by 2 items | Not a conversion issue - test needs better isolation or comparison |
| tablebuilder | funnel chart returns "pie" type | `transformFunnel` calls `transformPie` which returns `ChartTypePie` | Fix `transformFunnel` to return correct type |

---

## 1. userassetapi Fixes

**Files to modify:**
- `app/domain/assets/userassetapp/model.go`
- `api/cmd/services/ichor/tests/assets/userassetapi/seed_test.go`

**Problem:**
The test seed data calls `userassetapp.ToAppUserAssets(userAssets)` which converts `time.Time` values to strings using `bus.DateReceived.String()` format:
```
2025-11-02 15:48:07.834276 +0000 UTC
```

But the `toBusNewUserAsset` and `toBusUpdateUserAsset` functions expect RFC3339 format:
```
2006-01-02T15:04:05Z07:00
```

**Root Cause (model.go lines 52-53):**
```go
DateReceived:        bus.DateReceived.String(),  // Uses Go's default String() format
LastMaintenance:     bus.LastMaintenance.String(),
```

**Fix:**
Change `ToAppUserAsset` to format dates as RFC3339:

```go
func ToAppUserAsset(bus userassetbus.UserAsset) UserAsset {
    return UserAsset{
        ID:                  bus.ID.String(),
        UserID:              bus.UserID.String(),
        AssetID:             bus.AssetID.String(),
        ApprovedBy:          bus.ApprovedBy.String(),
        ApprovalStatusID:    bus.ApprovalStatusID.String(),
        FulfillmentStatusID: bus.FulfillmentStatusID.String(),
        DateReceived:        bus.DateReceived.Format(time.RFC3339),  // Changed
        LastMaintenance:     bus.LastMaintenance.Format(time.RFC3339),  // Changed
    }
}
```

---

## 2. warehouseapi Fixes

**Files to modify:**
- `app/domain/inventory/warehouseapp/model.go`

**Problem:**
The `UpdateWarehouse` struct has non-pointer fields for optional values, and the conversion function doesn't handle empty/optional strings correctly.

Current struct (model.go lines 111-117):
```go
type UpdateWarehouse struct {
    Code      string `json:"code" validate:"omitempty"`
    StreetID  string `json:"street_id" validate:"omitempty,uuid"`
    Name      string `json:"name" validate:"omitempty"`
    IsActive  bool   `json:"is_active" validate:"omitempty"`
    UpdatedBy string `json:"updated_by" validate:"required,uuid"`
}
```

The `toBusUpdateWarehouse` function tries to parse empty strings as UUIDs which fails.

**Fix:**
Change `UpdateWarehouse` to use pointer types for optional fields (standard pattern):

```go
type UpdateWarehouse struct {
    Code      *string `json:"code"`
    StreetID  *string `json:"street_id" validate:"omitempty,uuid"`
    Name      *string `json:"name"`
    IsActive  *bool   `json:"is_active"`
    UpdatedBy string  `json:"updated_by" validate:"required,uuid"`
}
```

And update `toBusUpdateWarehouse`:

```go
func toBusUpdateWarehouse(app UpdateWarehouse) (warehousebus.UpdateWarehouse, error) {
    bus := warehousebus.UpdateWarehouse{}

    if app.Code != nil {
        bus.Code = app.Code
    }

    if app.Name != nil {
        bus.Name = app.Name
    }

    if app.IsActive != nil {
        bus.IsActive = app.IsActive
    }

    if app.StreetID != nil && *app.StreetID != "" {
        streetID, err := uuid.Parse(*app.StreetID)
        if err != nil {
            return warehousebus.UpdateWarehouse{}, errs.Newf(errs.InvalidArgument, "parse streetID: %s", err)
        }
        bus.StreetID = &streetID
    }

    updatedBy, err := uuid.Parse(app.UpdatedBy)
    if err != nil {
        return warehousebus.UpdateWarehouse{}, errs.Newf(errs.InvalidArgument, "parse updatedBy: %s", err)
    }
    bus.UpdatedBy = &updatedBy

    return bus, nil
}
```

Also update the test file to use pointers:
- `api/cmd/services/ichor/tests/inventory/warehouseapi/update_test.go`

```go
func update200(sd apitest.SeedData) []apitest.Table {
    name := "Updated Warehouse"
    isActive := false
    return []apitest.Table{
        {
            Name:       "basic",
            URL:        "/v1/inventory/warehouses/" + sd.Warehouses[0].ID,
            Token:      sd.Admins[0].Token,
            Method:     "PUT",
            StatusCode: 200,
            Input: &warehouseapp.UpdateWarehouse{
                Name:      &name,
                IsActive:  &isActive,
                UpdatedBy: sd.Admins[0].ID.String(),
            },
            // ... rest unchanged
        },
    }
}
```

---

## 3. lottrackingsapi Fixes

**Files to modify:**
- `api/cmd/services/ichor/tests/inventory/lottrackingsapi/create_test.go`

**Problem:**
The tests expect error messages with the old format:
```
toBusNewLotTrackings: failed to parse time: parsing time "..."
```

But the refactored code returns:
```
parse manufactureDate: parsing time "..."
```

**Fix:**
Update the expected error messages in `create_test.go` lines 230, 262, 294:

Change from:
```go
ExpResp: errs.Newf(errs.InvalidArgument, `toBusNewLotTrackings: failed to parse time: parsing time "`+md.Format(time.RFC1123)+`" as "`+timeutil.FORMAT+`": cannot parse "`+md.Format(time.RFC1123)+`" as "2006"`)
```

To:
```go
ExpResp: errs.Newf(errs.InvalidArgument, `parse manufactureDate: parsing time "`+md.Format(time.RFC1123)+`" as "`+timeutil.FORMAT+`": cannot parse "`+md.Format(time.RFC1123)+`" as "2006"`)
```

Similarly for `expirationDate` (line 262) and `receivedDate` (line 294).

---

## 4. data Tests

**Files to potentially modify:**
- `api/cmd/services/ichor/tests/data/execute_test.go`
- `api/cmd/services/ichor/tests/data/seed_test.go`

**Problem:**
The test is comparing expected data based on sorted inventory items, but there's a 2-item offset in the results. This appears to be a test data consistency issue rather than a conversion issue.

The expected data starts at current_stock 999-992 but the API returns 1001-992. Two extra inventory items with higher quantities exist in the actual DB that weren't seeded by the test.

**Analysis:**
This is likely NOT a conversion issue. It could be:
1. Additional data from another test or migration
2. Race condition in test data setup
3. The test seed is creating fewer items than expected

**Fix Options:**
1. **Option A (Recommended):** Make the comparison more flexible by only checking that returned data matches a subset
2. **Option B:** Investigate seed_test.go to ensure exactly 30 items are created
3. **Option C:** Use `cmpopts.IgnoreSliceElements` in the comparison to skip first N items

For now, mark this as **needing investigation** - not a conversion issue.

---

## 5. tablebuilder Fixes

**Files to modify:**
- `business/sdk/tablebuilder/chart.go`

**Problem:**
The `transformFunnel` function (line 349-352) calls `transformPie` which returns `ChartTypePie`:

```go
func (ct *ChartTransformer) transformFunnel(data *TableData, settings *ChartVisualSettings) (*ChartResponse, error) {
    // Funnel is similar to pie - name/value pairs
    return ct.transformPie(data, settings)  // Returns ChartTypePie!
}
```

**Fix:**
Modify `transformFunnel` to set the correct type:

```go
func (ct *ChartTransformer) transformFunnel(data *TableData, settings *ChartVisualSettings) (*ChartResponse, error) {
    // Funnel is similar to pie - name/value pairs
    result, err := ct.transformPie(data, settings)
    if err != nil {
        return nil, err
    }
    result.Type = ChartTypeFunnel
    return result, nil
}
```

---

## Execution Order

1. **tablebuilder** - Quick single-line fix
2. **lottrackingsapi** - Simple test expectation updates (3 changes)
3. **userassetapi** - Change date formatting in model.go
4. **warehouseapi** - Structural change to use pointer types
5. **data** - Investigate separately (not a conversion issue)

---

## Verification Commands

After each fix, run the respective tests:

```bash
# tablebuilder
go test -v ./business/sdk/tablebuilder/...

# lottrackingsapi
go test -v ./api/cmd/services/ichor/tests/inventory/lottrackingsapi/...

# userassetapi
go test -v ./api/cmd/services/ichor/tests/assets/userassetapi/...

# warehouseapi
go test -v ./api/cmd/services/ichor/tests/inventory/warehouseapi/...

# data
go test -v ./api/cmd/services/ichor/tests/data/...

# All together
make test
```
