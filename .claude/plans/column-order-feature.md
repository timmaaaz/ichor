# Plan: Add Column Order Field to TableBuilder

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
    sort.SliceStable(metadata, func(i, j int) bool {
        return metadata[i].Order < metadata[j].Order
    })
}
```

Add `"sort"` to imports if not present.

### 3b. validation.go - Add strict Order validation
In `validateVisualSettings()`:

```go
// Check for strict Order enforcement
var hasOrder, missingOrder []string
for name, col := range c.VisualSettings.Columns {
    if col.Hidden {
        continue // Skip hidden columns
    }
    if col.Order != 0 {
        hasOrder = append(hasOrder, name)
    } else {
        missingOrder = append(missingOrder, name)
    }
}

if len(hasOrder) > 0 && len(missingOrder) > 0 {
    for _, name := range missingOrder {
        result.AddError(
            fmt.Sprintf("visual_settings.columns[%s].order", name),
            fmt.Sprintf("column '%s' missing Order value - when using explicit ordering, all visible columns must have Order specified", name),
            "STRICT_ORDER",
        )
    }
}
```

### 4. validation_test.go - Add tests
- Test explicit Order sorts correctly
- Test default (0) preserves implicit order
- Test negative values work
- Test duplicate Order values use stable sort

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
- **Validation error**: Clear message like `"column 'stock_status' missing Order value - when using explicit ordering, all visible columns must have Order specified"`
- **Negative values**: Allowed - useful for "always first" columns
- **Duplicate values**: Allowed - stable sort preserves relative position
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
1. Run `make test` to ensure all tests pass
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

    // Use metadata columns for deterministic ordering
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

    return categoryCol, valueColumns
}
```

### Files to Modify (Updated)
1. [model.go](business/sdk/tablebuilder/model.go) - Add Order to ColumnConfig and ColumnMetadata
2. [store.go](business/sdk/tablebuilder/store.go) - Add sorting logic to buildColumnMetadata()
3. [validation.go](business/sdk/tablebuilder/validation.go) - Add strict Order validation
4. [validation_test.go](business/sdk/tablebuilder/validation_test.go) - Add Order field tests
5. [chart.go](business/sdk/tablebuilder/chart.go) - Fix detectCategoricalColumns() to use metadata
6. [tables.go](business/sdk/dbtest/seedmodels/tables.go) - Update ComplexConfig with Order values
