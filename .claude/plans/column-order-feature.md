# Plan: Add Column Order Field to TableBuilder

---

## Frontend Developer Synopsis

### What's Changing

The API will return a new `order` field in `meta.columns` that controls display order. The `meta.columns` array will be **pre-sorted** by this order value.

### API Response Changes

**New field in `ColumnMetadata`:**
```json
{
  "meta": {
    "columns": [
      {
        "field": "product_name",
        "header": "Product",
        "order": 10,
        ...
      },
      {
        "field": "quantity",
        "header": "Qty",
        "order": 20,
        ...
      }
    ]
  }
}
```

### Key Points for Frontend

1. **Array is pre-sorted** - The `meta.columns` array arrives sorted by `order` value. Frontend can render columns in array order without additional sorting.

2. **Order values** - Lower values appear first. Negative values allowed (e.g., `-10` appears before `10`).

3. **Hidden columns at end** - Columns marked `hidden: true` are placed at the end of the array regardless of their `order` value.

4. **Backward compatible** - If `order` is `0` or missing, columns appear in their original processing order (main → foreign → computed).

5. **No frontend changes required initially** - Current hardcoded headers continue to work. This enables future dynamic column rendering.

### Example Usage (Future Enhancement)

```javascript
// Render columns in order from API
const columns = response.meta.columns
  .filter(col => !col.hidden)
  .map(col => ({
    field: col.field,
    header: col.header,
    // columns already sorted by order
  }));
```

---

## Summary

Add an `Order int` field to `ColumnConfig` to allow explicit column ordering in table configurations, giving control over display order via `VisualSettings.Columns` without reordering data source definitions.

## Problem

- Column order is currently determined by the order columns appear in `DataSource.Select.Columns`, `ForeignTables`, then `ClientComputedColumns`
- `VisualSettings.Columns` is a map (no guaranteed order)
- Users want to display `product_name` first but it comes from a `ForeignTable` which is processed after main columns

## Solution

Add `Order int` field to `ColumnConfig` and sort metadata by this field before returning.

---

## Changes

### 1. model.go - Add Order field to ColumnConfig (~line 193)

```go
type ColumnConfig struct {
    // ... existing fields ...
    Order        int             `json:"order,omitempty"`  // Display order (lower = earlier)
    // ... rest of fields ...
}
```

### 2. model.go - Add Order field to ColumnMetadata (~line 318)

```go
type ColumnMetadata struct {
    // ... existing fields ...
    Order        int    `json:"order,omitempty"`  // Display order for frontend
    // ... rest of fields ...
}
```

### 3. store.go - Update buildColumnMetadata() (~line 455)

After building the metadata slice, before returning:

```go
import (
    // ... existing imports ...
    "sort"
)
```

```go
// Copy Order values from VisualSettings to metadata
for i := range metadata {
    if vs, ok := config.VisualSettings.Columns[metadata[i].Field]; ok {
        metadata[i].Order = vs.Order
    }
}

// Sort by Order (only if any column has explicit order)
hasExplicitOrder := false
for _, m := range metadata {
    if m.Order != 0 {
        hasExplicitOrder = true
        break
    }
}

if hasExplicitOrder {
    // Filter out hidden columns before sorting
    var visibleMetadata []ColumnMetadata
    var hiddenMetadata []ColumnMetadata
    for _, m := range metadata {
        if vs, ok := config.VisualSettings.Columns[m.Field]; ok && vs.Hidden {
            hiddenMetadata = append(hiddenMetadata, m)
        } else {
            visibleMetadata = append(visibleMetadata, m)
        }
    }

    sort.SliceStable(visibleMetadata, func(i, j int) bool {
        // Primary sort by Order
        if visibleMetadata[i].Order != visibleMetadata[j].Order {
            return visibleMetadata[i].Order < visibleMetadata[j].Order
        }
        // Secondary sort by Field name for determinism when Order is equal
        return visibleMetadata[i].Field < visibleMetadata[j].Field
    })

    // Rebuild metadata with hidden columns at end
    metadata = append(visibleMetadata, hiddenMetadata...)
}

### 3b. validation.go - Add strict Order validation

In `validateVisualSettings()`:

```go
import (
    // ... existing imports ...
    "sort"
    "strings"
)
```

```go
// Check for strict Order enforcement
var hasOrder, missingOrder []string
for name, col := range c.VisualSettings.Columns {
    if col.Hidden {
        // Warn if hidden column has Order set (likely mistake)
        if col.Order != 0 {
            result.AddWarning(
                fmt.Sprintf("visual_settings.columns[%s].order", name),
                fmt.Sprintf("column '%s' is hidden but has Order value set - this will be ignored", name),
            )
        }
        continue // Skip hidden columns for strict order check
    }

    // Validate Order bounds
    if col.Order < -1000 || col.Order > 1000 {
        result.AddError(
            fmt.Sprintf("visual_settings.columns[%s].order", name),
            fmt.Sprintf("order value %d is out of reasonable range [-1000, 1000]", col.Order),
            "INVALID_VALUE",
        )
    }

    if col.Order != 0 {
        hasOrder = append(hasOrder, name)
    } else {
        missingOrder = append(missingOrder, name)
    }
}

// Sort for deterministic error messages (Go maps have non-deterministic iteration)
sort.Strings(hasOrder)
sort.Strings(missingOrder)

if len(hasOrder) > 0 && len(missingOrder) > 0 {
    result.AddError(
        "visual_settings.columns",
        fmt.Sprintf(
            "Mixed explicit and implicit column ordering detected. "+
            "When any visible column has an explicit Order value, ALL visible columns must have Order values. "+
            "Columns with Order: [%s]. Columns without Order: [%s]. "+
            "Tip: Use Order values like 10, 20, 30 to allow easy insertions later.",
            strings.Join(hasOrder, ", "),
            strings.Join(missingOrder, ", "),
        ),
        "STRICT_ORDER",
    )
}
```

### 4. validation_test.go - Add tests

```go
func TestValidateConfig_ColumnOrderStrict(t *testing.T) {
    tests := []struct {
        name         string
        columns      map[string]tablebuilder.ColumnConfig
        wantErrors   []string
        wantWarnings []string
    }{
        {
            name: "all_columns_with_order",
            columns: map[string]tablebuilder.ColumnConfig{
                "col1": {Name: "col1", Type: "string", Order: 1},
                "col2": {Name: "col2", Type: "string", Order: 2},
            },
            wantErrors: nil,
        },
        {
            name: "all_columns_with_order_zero",
            columns: map[string]tablebuilder.ColumnConfig{
                "col1": {Name: "col1", Type: "string", Order: 0},
                "col2": {Name: "col2", Type: "string", Order: 0},
            },
            wantErrors: nil, // No explicit order, so no strict enforcement
        },
        {
            name: "mixed_order_and_no_order",
            columns: map[string]tablebuilder.ColumnConfig{
                "col1": {Name: "col1", Type: "string", Order: 1},
                "col2": {Name: "col2", Type: "string", Order: 0}, // Missing
            },
            wantErrors: []string{"STRICT_ORDER"},
        },
        {
            name: "hidden_column_with_order_warns",
            columns: map[string]tablebuilder.ColumnConfig{
                "col1": {Name: "col1", Type: "string", Order: 1},
                "col2": {Name: "col2", Type: "string", Order: 2},
                "col3": {Name: "col3", Type: "string", Hidden: true, Order: 5},
            },
            wantErrors:   nil,
            wantWarnings: []string{"is hidden but has Order value"},
        },
        {
            name: "negative_order_values",
            columns: map[string]tablebuilder.ColumnConfig{
                "col1": {Name: "col1", Type: "string", Order: -10},
                "col2": {Name: "col2", Type: "string", Order: 5},
            },
            wantErrors: nil,
        },
        {
            name: "duplicate_order_values",
            columns: map[string]tablebuilder.ColumnConfig{
                "col1": {Name: "col1", Type: "string", Order: 5},
                "col2": {Name: "col2", Type: "string", Order: 5},
                "col3": {Name: "col3", Type: "string", Order: 10},
            },
            wantErrors: nil, // Allowed, stable sort with secondary key applies
        },
        {
            name: "order_out_of_bounds_high",
            columns: map[string]tablebuilder.ColumnConfig{
                "col1": {Name: "col1", Type: "string", Order: 1001},
                "col2": {Name: "col2", Type: "string", Order: 2},
            },
            wantErrors: []string{"INVALID_VALUE"},
        },
        {
            name: "order_out_of_bounds_low",
            columns: map[string]tablebuilder.ColumnConfig{
                "col1": {Name: "col1", Type: "string", Order: -1001},
                "col2": {Name: "col2", Type: "string", Order: 2},
            },
            wantErrors: []string{"INVALID_VALUE"},
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            config := &tablebuilder.Config{
                Title:      "Test",
                WidgetType: "table",
                DataSource: []tablebuilder.DataSource{
                    {Source: "test", Select: tablebuilder.SelectConfig{
                        Columns: []tablebuilder.ColumnDefinition{{Name: "col1"}, {Name: "col2"}},
                    }},
                },
                VisualSettings: tablebuilder.VisualSettings{
                    Columns: tt.columns,
                },
            }

            result := config.ValidateConfig()

            // Check errors
            for _, wantErr := range tt.wantErrors {
                found := false
                for _, err := range result.Errors {
                    if strings.Contains(err.Code, wantErr) {
                        found = true
                        break
                    }
                }
                if !found {
                    t.Errorf("Expected error with code %q, but not found", wantErr)
                }
            }

            // Check warnings
            for _, wantWarn := range tt.wantWarnings {
                found := false
                for _, warn := range result.Warnings {
                    if strings.Contains(warn.Message, wantWarn) {
                        found = true
                        break
                    }
                }
                if !found {
                    t.Errorf("Expected warning containing %q, but not found", wantWarn)
                }
            }
        })
    }
}

func TestStore_BuildColumnMetadata_Ordering(t *testing.T) {
    // Test that metadata is returned in Order sequence
    config := &tablebuilder.Config{
        DataSource: []tablebuilder.DataSource{{
            Select: tablebuilder.SelectConfig{
                Columns: []tablebuilder.ColumnDefinition{
                    {Name: "id"},
                    {Name: "name"},
                    {Name: "email"},
                },
            },
        }},
        VisualSettings: tablebuilder.VisualSettings{
            Columns: map[string]tablebuilder.ColumnConfig{
                "id":    {Order: 30, Type: "string"},
                "name":  {Order: 10, Type: "string"},
                "email": {Order: 20, Type: "string"},
            },
        },
    }

    store := tablebuilder.NewStore()
    metadata := store.BuildColumnMetadata(config.DataSource[0].Select.Columns, config, nil)

    // Verify order: name (10), email (20), id (30)
    if len(metadata) != 3 {
        t.Fatalf("Expected 3 columns, got %d", len(metadata))
    }

    expected := []string{"name", "email", "id"}
    for i, exp := range expected {
        if metadata[i].Field != exp {
            t.Errorf("Expected column %d to be %q, got %q", i, exp, metadata[i].Field)
        }
    }
}

func TestStore_BuildColumnMetadata_DeterministicOrder(t *testing.T) {
    // Run same config multiple times to verify deterministic ordering
    config := &tablebuilder.Config{
        DataSource: []tablebuilder.DataSource{{
            Select: tablebuilder.SelectConfig{
                Columns: []tablebuilder.ColumnDefinition{
                    {Name: "a"}, {Name: "b"}, {Name: "c"}, {Name: "d"}, {Name: "e"},
                },
            },
        }},
        VisualSettings: tablebuilder.VisualSettings{
            Columns: map[string]tablebuilder.ColumnConfig{
                "a": {Order: 5, Type: "string"},
                "b": {Order: 5, Type: "string"}, // Duplicate order - should sort by field name
                "c": {Order: 5, Type: "string"}, // Duplicate order
                "d": {Order: 10, Type: "string"},
                "e": {Order: 1, Type: "string"},
            },
        },
    }

    store := tablebuilder.NewStore()
    var firstResult []string

    // Run 100 times to catch non-determinism
    for i := 0; i < 100; i++ {
        metadata := store.BuildColumnMetadata(config.DataSource[0].Select.Columns, config, nil)
        var fields []string
        for _, m := range metadata {
            fields = append(fields, m.Field)
        }

        if i == 0 {
            firstResult = fields
        } else {
            for j, f := range fields {
                if f != firstResult[j] {
                    t.Fatalf("Non-deterministic order detected on iteration %d: got %v, expected %v", i, fields, firstResult)
                }
            }
        }
    }

    // Verify expected order: e (1), a/b/c (5, sorted by field name), d (10)
    expected := []string{"e", "a", "b", "c", "d"}
    for i, exp := range expected {
        if firstResult[i] != exp {
            t.Errorf("Expected column %d to be %q, got %q", i, exp, firstResult[i])
        }
    }
}

func TestStore_BuildColumnMetadata_HiddenColumnsAtEnd(t *testing.T) {
    config := &tablebuilder.Config{
        DataSource: []tablebuilder.DataSource{{
            Select: tablebuilder.SelectConfig{
                Columns: []tablebuilder.ColumnDefinition{
                    {Name: "visible1"},
                    {Name: "hidden1"},
                    {Name: "visible2"},
                },
            },
        }},
        VisualSettings: tablebuilder.VisualSettings{
            Columns: map[string]tablebuilder.ColumnConfig{
                "visible1": {Order: 20, Type: "string"},
                "hidden1":  {Order: 1, Type: "string", Hidden: true}, // Hidden with low order
                "visible2": {Order: 10, Type: "string"},
            },
        },
    }

    store := tablebuilder.NewStore()
    metadata := store.BuildColumnMetadata(config.DataSource[0].Select.Columns, config, nil)

    // Verify order: visible2 (10), visible1 (20), hidden1 (hidden, at end)
    expected := []string{"visible2", "visible1", "hidden1"}
    for i, exp := range expected {
        if metadata[i].Field != exp {
            t.Errorf("Expected column %d to be %q, got %q", i, exp, metadata[i].Field)
        }
    }
}
```

**Test Coverage Summary:**
- All columns with explicit Order - sorts correctly
- All columns with Order: 0 - preserves implicit order
- Mixed explicit/implicit Order - triggers STRICT_ORDER error
- Hidden columns with Order - warns about ignored Order value
- Negative Order values - allowed and sorts correctly
- Duplicate Order values - uses secondary sort by field name
- Order out of bounds (>1000 or <-1000) - triggers INVALID_VALUE error
- Deterministic ordering - run 100 times to catch map iteration issues
- Hidden columns - placed at end regardless of Order value

### 5. tables.go - Update example config

```go
"product_name": {
    Name:   "product_name",
    Header: "Product",
    Order:  1,  // Display first
    // ...
},
"current_quantity": {
    Order:  2,  // Display second
    // ...
},
```

---

## Files to Modify

1. [model.go](business/sdk/tablebuilder/model.go) - Add Order to ColumnConfig and ColumnMetadata
2. [store.go](business/sdk/tablebuilder/store.go) - Add sorting logic to buildColumnMetadata()
3. [validation.go](business/sdk/tablebuilder/validation.go) - Add strict Order validation
4. [validation_test.go](business/sdk/tablebuilder/validation_test.go) - Add Order field tests
5. [chart.go](business/sdk/tablebuilder/chart.go) - Fix detectCategoricalColumns()
6. [tables.go](business/sdk/dbtest/seedmodels/tables.go) - Update ComplexConfig with Order values

---

## Design Notes

- **Strict enforcement**: If ANY non-hidden column has `Order != 0`, ALL non-hidden columns must have `Order`
- **Validation error**: Single comprehensive error with list of columns with/without Order values
- **Bounds validation**: Order must be in range [-1000, 1000] to catch configuration errors
- **Negative values**: Allowed - useful for "always first" columns
- **Duplicate values**: Allowed - stable sort with secondary key (field name) ensures determinism
- **Secondary sort key**: When Order values are equal, sort alphabetically by Field name for deterministic output
- **Hidden columns**:
  - Excluded from strict Order enforcement
  - Warn if hidden column has Order set (likely configuration mistake)
  - Placed at end of metadata array regardless of Order value
- **Deterministic validation**: Sort column names before producing error messages (Go maps have non-deterministic iteration)
- **LLM-friendly**: Strict rules work better for LLM-generated configs (clear constraints, validation feedback)

---

## Ramifications Analysis (Explored)

### Frontend Impact ✅ SAFE

- **Finding**: Frontend does NOT consume dynamic column metadata from API
- Headers are hardcoded per-route (e.g., `Users.js` has static `UsersTableHeaders`)
- API returns `meta.columns` but frontend ignores it completely
- **Conclusion**: Adding `Order` field won't break anything - frontend will ignore it until enhanced

### Other TableBuilder Features ⚠️ ATTENTION NEEDED

#### Charts (chart.go) - RISK IDENTIFIED

1. **Column detection uses map iteration** (lines 632-650):
   - `detectCategoricalColumns()` iterates `range row` on a map
   - Map iteration in Go is **non-deterministic**
   - First string column becomes category, numeric columns become values
   - **This is a pre-existing bug**, not caused by Order field

2. **SeriesConfig index matching** (lines 281-293):
   - Combo charts match `valueColumns[i]` to `SeriesConfig[i]` by index
   - If column order changes, series configs will mismatch
   - **Recommendation**: Chart configs should reference columns by name, not position

#### Computed Columns - ORDER MATTERS

- Later computed columns can reference earlier ones in expressions
- `buildColumnMetadata()` processes in order: main → foreign → computed
- **Our change**: Sorting happens AFTER metadata is built, so computed column evaluation is unaffected

#### SQL Column Selection - SAFE

- `buildSelectColumns()` iterates config in order
- Our sorting only affects the metadata array returned to frontend
- SQL query order is unchanged

#### Export - LIKELY AFFECTED

- No explicit export code found, but Actions support `"export"` type
- Column order in metadata array likely determines CSV/Excel column order
- **This is actually desired behavior** - Order field will control export column order

---

## Verification

1. Run specific tests, not `make test` to ensure all tests pass
2. Run `make lint` for linting
3. Test manually that columns appear in expected order in API response

---

---

## Chart Column Detection Fix (Additional)

### Problem

`detectCategoricalColumns()` in [chart.go:632-650](business/sdk/tablebuilder/chart.go#L632-L650) iterates over `map[string]any` which has **non-deterministic order** in Go. This causes:

- Random category column selection
- Random value column ordering
- Inconsistent chart rendering between requests

### Solution

Use `data.Meta.Columns` (ordered `[]ColumnMetadata`) instead of iterating `data.Data[0]` map.

### 6. chart.go - Fix detectCategoricalColumns() (~line 632)

**Before:**

```go
func (ct *ChartTransformer) detectCategoricalColumns(data *TableData) (string, []string) {
    if len(data.Data) == 0 {
        return "", nil
    }

    var categoryCol string
    var valueColumns []string

    row := data.Data[0]
    for key, val := range row {  // ❌ Non-deterministic map iteration
        if ct.isNumericValue(val) {
            valueColumns = append(valueColumns, key)
        } else if categoryCol == "" {
            categoryCol = key
        }
    }

    return categoryCol, valueColumns
}
```

**After:**

```go
func (ct *ChartTransformer) detectCategoricalColumns(data *TableData) (string, []string) {
    if len(data.Data) == 0 {
        return "", nil
    }

    var categoryCol string
    var valueColumns []string

    row := data.Data[0]

    // Use metadata columns for deterministic ordering if available
    if len(data.Meta.Columns) > 0 {
        for _, col := range data.Meta.Columns {
            key := col.Field
            val, exists := row[key]
            if !exists {
                continue
            }

            if ct.isNumericValue(val) {
                valueColumns = append(valueColumns, key)
            } else if categoryCol == "" {
                categoryCol = key
            }
        }
    } else {
        // Fallback to deterministic map iteration using sorted keys
        // This handles edge cases where metadata isn't populated
        keys := make([]string, 0, len(row))
        for key := range row {
            keys = append(keys, key)
        }
        sort.Strings(keys)

        for _, key := range keys {
            val := row[key]
            if ct.isNumericValue(val) {
                valueColumns = append(valueColumns, key)
            } else if categoryCol == "" {
                categoryCol = key
            }
        }
    }

    return categoryCol, valueColumns
}
```

**Note:** Add `"sort"` to imports in chart.go if not already present.

### Files to Modify (Updated)

1. [model.go](business/sdk/tablebuilder/model.go) - Add Order to ColumnConfig and ColumnMetadata
2. [store.go](business/sdk/tablebuilder/store.go) - Add sorting logic to buildColumnMetadata()
3. [validation.go](business/sdk/tablebuilder/validation.go) - Add strict Order validation
4. [validation_test.go](business/sdk/tablebuilder/validation_test.go) - Add Order field tests
5. [chart.go](business/sdk/tablebuilder/chart.go) - Fix detectCategoricalColumns() to use metadata
6. [tables.go](business/sdk/dbtest/seedmodels/tables.go) - Update ComplexConfig with Order values

---

## Review Summary

This plan was reviewed by the go-service-reviewer agent. The following fixes were incorporated:

### Critical Fixes Applied

1. **Non-deterministic map iteration** - Added `sort.Strings()` calls before producing validation error messages to ensure consistent output across runs.

2. **Secondary sort key** - Added Field name as secondary sort key when Order values are equal, preventing arbitrary ordering.

3. **Hidden column handling** - Added logic to filter hidden columns before sorting and place them at end of metadata array, with warnings for hidden columns that have Order set.

4. **Chart fix defensive code** - Added fallback to sorted map keys when `data.Meta.Columns` is not populated.

5. **Bounds validation** - Added range check [-1000, 1000] for Order values.

6. **Comprehensive error message** - Changed from per-column errors to single error with all affected columns listed.

### Test Coverage Expanded

Added 8 test scenarios:
- All columns with explicit Order
- All columns with Order: 0 (no strict enforcement)
- Mixed explicit/implicit Order (STRICT_ORDER error)
- Hidden columns with Order (warning)
- Negative Order values
- Duplicate Order values (secondary sort by field name)
- Order out of bounds
- Deterministic ordering (100 iteration test)
- Hidden columns placed at end
