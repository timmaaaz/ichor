# Form Validation Remediation Plan

## Overview

The `make validate-forms-deep` test (`TestFormConfigsAgainstSchema`) is failing with **18 errors across 12 forms**. These errors are caused by mismatches between form field configurations and the actual database schema.

## Error Summary

| Error Type | Count | Description |
|------------|-------|-------------|
| Column name mismatch | 2 | Form uses wrong column name |
| Missing columns | 3 | Schema doesn't have the referenced column |
| Missing table | 2 | Referenced entity doesn't exist |
| Wrong value_column | 4 | Using prefixed ID (`supplier_id`) instead of `id` |
| Qualified column names | 2 | Dropdown uses `table.column` format instead of `column` |
| Wrong display_field case | 3 | Using camelCase instead of snake_case |

## Files to Modify

1. [tableforms.go](business/sdk/dbtest/seedmodels/tableforms.go)
2. [forms.go](business/sdk/dbtest/seedmodels/forms.go)

---

## Fix 1: sales.orders - Wrong Column Name

**Error**: `column "fulfillment_status_id" not found in table sales.orders`

**File**: [tableforms.go:908](business/sdk/dbtest/seedmodels/tableforms.go#L908)

**Current**:
```go
Name: "fulfillment_status_id"
```

**Fix**:
```go
Name: "order_fulfillment_status_id"
```

**Reason**: Schema defines `order_fulfillment_status_id` (line 875 of migrate.sql)

---

## Fix 2: line_item_fulfillment_statuses Dropdown - Qualified Column Names

**Error**: `label column "line_item_fulfillment_statuses.name" not found`

**File**: [forms.go:381-383](business/sdk/dbtest/seedmodels/forms.go#L381-L383)

**Current**:
```go
DropdownConfig: &formfieldbus.DropdownConfig{
    Entity:      "sales.line_item_fulfillment_statuses",
    LabelColumn: "line_item_fulfillment_statuses.name",
    ValueColumn: "line_item_fulfillment_statuses.id",
}
```

**Fix**:
```go
DropdownConfig: &formfieldbus.DropdownConfig{
    Entity:      "sales.line_item_fulfillment_statuses",
    LabelColumn: "name",
    ValueColumn: "id",
}
```

**Reason**: Column names should not include table prefix

---

## Fix 3: procurement.suppliers - Wrong value_column

**Errors** (3 locations):
- `GetPurchaseOrderFormFields` - value_column "supplier_id" not found
- `GetSupplierProductFormFields` - value_column "supplier_id" not found
- `GetFullPurchaseOrderFormFields` (inherits from GetPurchaseOrderFormFields)

**File**: [tableforms.go:958](business/sdk/dbtest/seedmodels/tableforms.go#L958)

**Current**:
```go
Config: json.RawMessage(`{"entity": "procurement.suppliers", "display_field": "name", "value_column": "supplier_id", ...}`)
```

**Fix**:
```go
Config: json.RawMessage(`{"entity": "procurement.suppliers", "display_field": "name", "value_column": "id", ...}`)
```

**File**: [tableforms.go:1014](business/sdk/dbtest/seedmodels/tableforms.go#L1014)

Same fix - change `"value_column": "supplier_id"` to `"value_column": "id"`

**Reason**: Primary key is `id`, not `supplier_id`

---

## Fix 4: procurement.supplier_products - Wrong value_column

**Error**: `value column "supplier_product_id" not found in table procurement.supplier_products`

**File**: [tableforms.go:974](business/sdk/dbtest/seedmodels/tableforms.go#L974)

**Current**:
```go
Config: json.RawMessage(`{"entity": "procurement.supplier_products", ..., "value_column": "supplier_product_id", ...}`)
```

**Fix**:
```go
Config: json.RawMessage(`{"entity": "procurement.supplier_products", ..., "value_column": "id", ...}`)
```

**Reason**: Primary key is `id`, not `supplier_product_id`

---

## Fix 5: geography.timezones - Wrong display_field Case

**Error**: `display column "displayName" not found in table geography.timezones`

**Affected forms** (3 locations):
- `GetContactInfoFormFields`
- `GetFullSupplierFormFields` (inherits from GetContactInfoFormFields)
- `GetFullCustomerFormFields` (inherits from GetContactInfoFormFields)

**File**: [tableforms.go:471](business/sdk/dbtest/seedmodels/tableforms.go#L471)

**Current**:
```go
Config: json.RawMessage(`{"entity": "geography.timezones", "display_field": "displayName"}`)
```

**Fix**:
```go
Config: json.RawMessage(`{"entity": "geography.timezones", "display_field": "display_name"}`)
```

**Reason**: Schema uses snake_case (`display_name`), not camelCase

---

## Fix 6: core.contact_types - Missing Table

**Error**: `dropdown entity core.contact_types does not exist in database`

**File**: [tableforms.go:472](business/sdk/dbtest/seedmodels/tableforms.go#L472)

**Current**:
```go
Config: json.RawMessage(`{"entity": "core.contact_types", "display_field": "name"}`)
```

**Fix**: Change to use PostgreSQL enum type instead of dropdown:
```go
Name: "preferred_contact_type", Label: "Preferred Contact Method", FieldType: "enum", FieldOrder: 10, Required: true, Config: json.RawMessage(`{"enum_name": "contact_type"}`)
```

**Reason**: `preferred_contact_type` is a PostgreSQL ENUM (`contact_type`), not a FK to another table. The enum values are: `'phone', 'email', 'mail', 'fax'`

---

## Fix 7: Status Tables - Missing description Column

**Errors**:
- `hr.user_approval_status` missing `description`
- `assets.approval_status` missing `description`
- `assets.fulfillment_status` missing `description`

**Files**:
- [tableforms.go:130-141](business/sdk/dbtest/seedmodels/tableforms.go#L130-L141) - GetUserApprovalStatusFormFields
- [tableforms.go:161-172](business/sdk/dbtest/seedmodels/tableforms.go#L161-L172) - GetAssetApprovalStatusFormFields
- [tableforms.go:192-203](business/sdk/dbtest/seedmodels/tableforms.go#L192-L203) - GetAssetFulfillmentStatusFormFields

**Fix**: Remove the `description` field entries from all three functions.

**Reason**: These status tables don't have a `description` column. They have: `id`, `icon_id`, `name`, `primary_color`, `secondary_color`, `icon`

---

## Implementation Checklist

### tableforms.go Changes

| Line | Function | Change |
|------|----------|--------|
| 130-141 | `GetUserApprovalStatusFormFields` | Remove description field |
| 161-172 | `GetAssetApprovalStatusFormFields` | Remove description field |
| 192-203 | `GetAssetFulfillmentStatusFormFields` | Remove description field |
| 471 | `GetContactInfoFormFields` | Change `displayName` to `display_name` |
| 472 | `GetContactInfoFormFields` | Change smart-combobox to enum type |
| 908 | `GetSalesOrderFormFields` | Change `fulfillment_status_id` to `order_fulfillment_status_id` |
| 958 | `GetPurchaseOrderFormFields` | Change `supplier_id` to `id` in value_column |
| 974 | `GetPurchaseOrderLineItemFormFields` | Change `supplier_product_id` to `id` in value_column |
| 1014 | `GetSupplierProductFormFields` | Change `supplier_id` to `id` in value_column |

### forms.go Changes

| Line | Function | Change |
|------|----------|--------|
| 381-383 | `GetFullSalesOrderFormFields` | Remove table prefix from LabelColumn and ValueColumn |

---

## Verification

After making changes, run:
```bash
make validate-forms-deep
```

**Expected result**: All 55 forms pass with 0 errors.

---

## Notes

- The composite forms (`GetFull*`) inherit errors from the base forms they call
- Fixing the base form automatically fixes the composite form
- The `enum` field type should work like other enum fields in the codebase (see `discount_type` at line 371 in forms.go for reference)
