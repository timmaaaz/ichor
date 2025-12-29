# Convert Package Refactor: Master Plan

## Overview

Replace reflection-based `convert.PopulateSameTypes` and `convert.PopulateTypesFromStrings` with explicit, type-safe conversion functions across the codebase.

## Problem Statement

The `business/sdk/convert` package uses reflection to convert between structs at runtime. While this reduces boilerplate, it:
- Bypasses compile-time type safety
- Silently fails on field mismatches
- Hides conversion logic from developers
- Adds runtime performance overhead
- Is not idiomatic Go

## Scope

| Function | Current Usages | Layer | Replacement Pattern |
|----------|---------------|-------|---------------------|
| `PopulateSameTypes` | 32+ | Business (`*bus`) | Explicit field assignment in `Update()` methods |
| `PopulateTypesFromStrings` | 38+ | App (`*app`) | Explicit `toBusNew*()` / `toBusUpdate*()` functions |

## Goals

1. **Type safety** - All conversions checked at compile time
2. **Explicit code** - Conversion logic visible and debuggable
3. **Better errors** - Field-level error messages on parse failures
4. **Performance** - Direct assignment vs reflection overhead
5. **Maintainability** - Changes to structs surface as compile errors

## Progress

- [x] Phase 1: Business Layer - `PopulateSameTypes` Removal (COMPLETED)
- [x] Phase 2: App Layer - `PopulateTypesFromStrings` Removal (COMPLETED)
- [x] Phase 3: Cleanup - Delete convert package (COMPLETED)

## Strategy

### Phase 1: Business Layer - `PopulateSameTypes` Removal (COMPLETED)
Replace reflection-based Update conversions with explicit field assignment in `*bus` packages.

**Pattern Before:**
```go
func (b *Business) Update(ctx context.Context, entity Entity, update UpdateEntity) (Entity, error) {
    if err := convert.PopulateSameTypes(update, &entity); err != nil {
        return Entity{}, err
    }
    // ...
}
```

**Pattern After:**
```go
func (b *Business) Update(ctx context.Context, entity Entity, update UpdateEntity) (Entity, error) {
    if update.Name != nil {
        entity.Name = *update.Name
    }
    if update.Description != nil {
        entity.Description = *update.Description
    }
    // ... explicit field assignments
}
```

### Phase 2: App Layer - `PopulateTypesFromStrings` Removal
Replace reflection-based string parsing with explicit conversion functions in `*app` packages.

**Pattern Before:**
```go
func toBusNewRole(app NewRole) (rolebus.NewRole, error) {
    dest := rolebus.NewRole{}
    err := convert.PopulateTypesFromStrings(app, &dest)
    return dest, err
}
```

**Pattern After:**
```go
func toBusNewRole(app NewRole) (rolebus.NewRole, error) {
    bus := rolebus.NewRole{
        Name:        app.Name,
        Description: app.Description,
    }
    return bus, nil
}
```

### Phase 3: Cleanup
- Remove `business/sdk/convert` package
- Update CLAUDE.md documentation
- Run tests to verify no regressions

## Success Criteria

- [x] Zero imports of `business/sdk/convert` remain
- [ ] All tests pass (`make test`)
- [ ] No new linting errors (`make lint`)
- [ ] CLAUDE.md updated to reflect new patterns

## Risk Mitigation

1. **Incremental approach** - Refactor one domain at a time, run tests after each
2. **Preserve behavior** - Explicit code must handle nil pointers identically to reflection
3. **Review nullable semantics** - Ensure zero-value vs nil distinction is maintained

---

# Phase 1: Business Layer - Remove `PopulateSameTypes` (COMPLETED)

## Overview
Replace `convert.PopulateSameTypes(update, &entity)` with explicit field assignment in all `*bus` package Update methods.

## Status: COMPLETED

All 32+ business layer files have been refactored to use explicit field assignments.

---

# Phase 2: App Layer - Remove `PopulateTypesFromStrings` (COMPLETED)

## Overview
Replace `convert.PopulateTypesFromStrings(app, &dest)` with explicit conversion functions in all `*app` package `toBusNew*` and `toBusUpdate*` functions.

## Status: COMPLETED

All 27 app layer files have been refactored to use explicit type-safe conversions.

## Pattern Transformation

**Before (roleapp/model.go:65):**
```go
func toBusNewRole(app NewRole) (rolebus.NewRole, error) {
    dest := rolebus.NewRole{}
    err := convert.PopulateTypesFromStrings(app, &dest)
    return dest, err
}
```

**After:**
```go
func toBusNewRole(app NewRole) (rolebus.NewRole, error) {
    bus := rolebus.NewRole{
        Name:        app.Name,
        Description: app.Description,
    }
    return bus, nil
}
```

**For UUID fields:**
```go
func toBusNewUserAsset(app NewUserAsset) (userassetbus.NewUserAsset, error) {
    userID, err := uuid.Parse(app.UserID)
    if err != nil {
        return userassetbus.NewUserAsset{}, fmt.Errorf("parse userID: %w", err)
    }

    assetID, err := uuid.Parse(app.AssetID)
    if err != nil {
        return userassetbus.NewUserAsset{}, fmt.Errorf("parse assetID: %w", err)
    }

    bus := userassetbus.NewUserAsset{
        UserID:  userID,
        AssetID: assetID,
        // ...
    }
    return bus, nil
}
```

**For nullable/Update fields (pointers):**
```go
func toBusUpdateRole(app UpdateRole) (rolebus.UpdateRole, error) {
    bus := rolebus.UpdateRole{
        Name:        app.Name,        // *string → *string, direct copy
        Description: app.Description, // *string → *string, direct copy
    }
    return bus, nil
}
```

## Files to Modify (25 app layer files)

### Core Domain
| File | Lines | Functions |
|------|-------|-----------|
| `app/domain/core/roleapp/model.go` | 67, 94 | toBusNewRole, toBusUpdateRole |
| `app/domain/core/userroleapp/model.go` | 67 | toBusNewUserRole |
| `app/domain/core/rolepageapp/model.go` | 71, 97 | toBusNewRolePage, toBusUpdateRolePage |
| `app/domain/core/contactinfosapp/model.go` | 118, 162 | toBusNewContactInfo, toBusUpdateContactInfo |

### Assets Domain
| File | Lines | Functions |
|------|-------|-----------|
| `app/domain/assets/userassetapp/model.go` | 95, 129 | toBusNewUserAsset, toBusUpdateUserAsset |
| `app/domain/assets/assetapp/model.go` | 78, 106 | toBusNewAsset, toBusUpdateAsset |

### HR Domain
| File | Lines | Functions |
|------|-------|-----------|
| `app/domain/hr/commentapp/model.go` | 75 | toBusNewComment |

### Products Domain
| File | Lines | Functions |
|------|-------|-----------|
| `app/domain/products/productapp/model.go` | 114, 150 | toBusNewProduct, toBusUpdateProduct |
| `app/domain/products/physicalattributeapp/model.go` | 136, 205 | toBusNewPhysicalAttribute, toBusUpdatePhysicalAttribute |
| `app/domain/products/brandapp/model.go` | 76, 102 | toBusNewBrand, toBusUpdateBrand |
| `app/domain/products/costhistoryapp/model.go` | 94, 132 | toBusNewCostHistory, toBusUpdateCostHistory |

### Inventory Domain
| File | Lines | Functions |
|------|-------|-----------|
| `app/domain/inventory/serialnumberapp/model.go` | 86, 113 | toBusNewSerialNumber, toBusUpdateSerialNumber |
| `app/domain/inventory/lottrackingsapp/model.go` | 94, 125 | toBusNewLotTracking, toBusUpdateLotTracking |

### Procurement Domain
| File | Lines | Functions |
|------|-------|-----------|
| `app/domain/procurement/supplierproductapp/model.go` | 103, 143 | toBusNewSupplierProduct, toBusUpdateSupplierProduct |
| `app/domain/procurement/purchaseorderlineitemapp/model.go` | 129, 169 | toBusNewPurchaseOrderLineItem, toBusUpdatePurchaseOrderLineItem |

### Sales Domain
| File | Lines | Functions |
|------|-------|-----------|
| `app/domain/sales/lineitemfulfillmentstatusapp/model.go` | 66, 92 | toBusNewLineItemFulfillmentStatus, toBusUpdateLineItemFulfillmentStatus |
| `app/domain/sales/customersapp/model.go` | 92, 121 | toBusNewCustomer, toBusUpdateCustomer |
| `app/domain/sales/orderfulfillmentstatusapp/model.go` | 66, 92 | toBusNewOrderFulfillmentStatus, toBusUpdateOrderFulfillmentStatus |
| `app/domain/sales/orderlineitemsapp/model.go` | 94, 124 | toBusNewOrderLineItem, toBusUpdateOrderLineItem |
| `app/domain/sales/ordersapp/model.go` | 93, 122 | toBusNewOrder, toBusUpdateOrder |

## Type Conversion Patterns

| App Type | Bus Type | Conversion |
|----------|----------|------------|
| `string` | `string` | Direct assignment |
| `string` | `uuid.UUID` | `uuid.Parse(app.Field)` |
| `string` | `time.Time` | `time.Parse(format, app.Field)` |
| `string` | `int` | `strconv.Atoi(app.Field)` |
| `string` | `bool` | `strconv.ParseBool(app.Field)` |
| `*string` | `*string` | Direct assignment |
| `*string` | `*uuid.UUID` | Parse if non-nil |
| `*string` | `*time.Time` | Parse if non-nil |

## Implementation Steps

1. For each file listed above:
   - Read the app model fields and corresponding bus model fields
   - Replace `convert.PopulateTypesFromStrings(app, &dest)` with explicit struct construction
   - Add UUID/time parsing with error handling where needed
   - Remove `convert` import
   - Run `go build ./...` to verify

2. After completing all files in a domain, run `make test` for that domain's tests

---

# Phase 3: Cleanup & Documentation (COMPLETED)

## Status: COMPLETED

The `business/sdk/convert` package has been deleted.

## Files Deleted
- `business/sdk/convert/convert.go`
- `business/sdk/convert/convert_test.go`

## Files to Update

### CLAUDE.md
Remove references to convert package. The existing documentation already shows explicit conversion patterns for complex models like User.

## Final Verification

```bash
# Ensure no remaining references
grep -r "business/sdk/convert" --include="*.go"

# Run full test suite
make test

# Run linting
make lint

# Build to verify
go build ./...
```

## Estimated Effort

| Phase | Files | Est. Time |
|-------|-------|-----------|
| Phase 1 | 28 business files | COMPLETED |
| Phase 2 | 27 app files | COMPLETED |
| Phase 3 | 2 files deleted | COMPLETED |
