# Plan: Support Hidden Columns in Table Config Validation

## Summary for Frontend Review

**What's changing:** Adding a `hidden` field to `ColumnConfig` in `visual_settings.columns`.

**JSON structure change:**
```json
{
  "visual_settings": {
    "columns": {
      "customer_name": {
        "hidden": true
      }
    }
  }
}
```

**Frontend impact:**
- When rendering table columns, check if `column.hidden === true` and skip rendering that column
- The column data will still be present in the row data (for use by other columns like lookups)
- This is consistent with the existing `hidden` field already present in `ColumnMetadata` responses

---

## Problem Statement

The `OrdersTableConfig` validation is failing with:
```
❌ OrdersTableConfig:
   • data_source[0].select.foreign_tables[0].columns[0]: column "customer_name" (from sales.customers) missing from visual_settings.columns
```

**Root Cause:** The `customer_name` column is selected from a foreign table (sales.customers) because it's needed for display in the lookup field for `orders.customer_id`. However, `customer_name` itself doesn't need to be shown as a separate visible column - it's a "support" column that provides data but shouldn't appear in the table.

Currently, the validation logic only exempts columns used as `LabelColumn` in `LinkConfig`, but doesn't have a concept of "hidden" columns that are selected for data purposes but shouldn't require full visual settings.

## Current State

1. **In `tables.go` (lines 477, 573):**
   - Foreign table column: `{Name: "name", Alias: "customer_name", TableColumn: "customers.name"}`
   - Visual settings entry: `"customer_name": {}` (empty struct - no Type defined)

2. **In `model.go` (line 306):**
   - `ColumnMetadata` already has a `Hidden bool` field used in the store layer
   - But `ColumnConfig` (visual settings) has no `Hidden` field

3. **In `validation.go`:**
   - `validateSelectedColumnsHaveVisualSettings()` checks that all selected columns have visual settings with a valid `Type`
   - `collectLabelColumns()` only exempts columns used as `LabelColumn` in `LinkConfig`

## Solution

Add a `Hidden bool` field to `ColumnConfig` in visual settings. Hidden columns:
- Are selected in the query (so data is available)
- Have minimal visual settings (just `Hidden: true`)
- Are exempt from the `Type` requirement in validation
- Are not displayed as table columns in the UI

This approach is:
- Explicit - clear what the column is for
- Consistent with existing `ColumnMetadata.Hidden` pattern in store layer
- Useful for frontend - can use this to skip rendering the column

## Implementation Steps

### Step 1: Update `ColumnConfig` model
**File:** `business/sdk/tablebuilder/model.go`

Add `Hidden` field to `ColumnConfig`:
```go
type ColumnConfig struct {
    Name         string          `json:"name"`
    Header       string          `json:"header"`
    Width        int             `json:"width,omitempty"`
    Align        string          `json:"align,omitempty"`
    Type         string          `json:"type,omitempty"`
    Hidden       bool            `json:"hidden,omitempty"`  // NEW: Column selected but not displayed
    Sortable     bool            `json:"sortable,omitempty"`
    // ... rest unchanged
}
```

### Step 2: Update validation to skip hidden columns
**File:** `business/sdk/tablebuilder/validation.go`

Update `collectLabelColumns()` to also collect hidden columns (or create a new function `collectExemptColumns()`):

```go
// collectExemptColumns returns a set of column names that are exempt from Type validation.
// This includes:
// 1. Columns used as LabelColumn in LinkConfig (display purposes in links)
// 2. Columns marked as Hidden (selected for data but not displayed)
func (c *Config) collectExemptColumns() map[string]bool {
    exempt := make(map[string]bool)
    for name, colConfig := range c.VisualSettings.Columns {
        // Exempt LabelColumn references
        if colConfig.Link != nil && colConfig.Link.LabelColumn != "" {
            exempt[colConfig.Link.LabelColumn] = true
        }
        // Exempt hidden columns
        if colConfig.Hidden {
            exempt[name] = true
        }
    }
    return exempt
}
```

Update `validateSelectedColumnsHaveVisualSettings()` and `validateForeignTableColumnsHaveVisualSettings()` to use this new function.

### Step 3: Update model.go validation helper
**File:** `business/sdk/tablebuilder/model.go`

Update `collectLabelColumnsForValidation()` to also include hidden columns, or rename and expand it.

### Step 4: Update test data
**File:** `business/sdk/dbtest/seedmodels/tables.go`

Update the `customer_name` entry in `OrdersTableConfig`:
```go
"customer_name": {
    Hidden: true,
},
```

### Step 5: Update spec documentation
**File:** `business/sdk/tablebuilder/docs/tablebuilder-config-validator-spec.md`

Add `Hidden` field to ColumnConfig documentation in section 14.

## Files to Modify

1. `business/sdk/tablebuilder/model.go` - Add `Hidden` field to `ColumnConfig`, update `collectLabelColumnsForValidation()`
2. `business/sdk/tablebuilder/validation.go` - Update `collectLabelColumns()` or create new function
3. `business/sdk/dbtest/seedmodels/tables.go` - Update `OrdersTableConfig` to mark `customer_name` as hidden
4. `business/sdk/tablebuilder/docs/tablebuilder-config-validator-spec.md` - Document the new field

## Testing

After implementation, run the table config validator CLI to verify the fix:
```bash
go run api/cmd/tooling/admin/main.go validate-table-configs
```

This should show `OrdersTableConfig` passing (no longer showing the `customer_name` error).

## Notes

- The `store.go` already sets `meta.Hidden = true` in several places (lines 368, 376, 386, 470), showing this pattern is already used at the metadata level
- This change aligns the configuration layer with the metadata layer's existing concept of hidden columns
