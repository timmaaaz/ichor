# Table Configuration Validator - Complete Field Reference

> **Note:** This document should be copied to `business/sdk/tablebuilder/docs/config-validator-spec.md` for permanent storage within the package.

This document provides a comprehensive mapping of all fields in the tablebuilder configuration system, intended to serve as the foundation for building a robust configuration validator.

---

## Quick Start (For Fresh Prompts)

### What This Document Contains
- Complete field reference for all 22+ config types (Sections 1-23)
- All allowed values/whitelists (Section 24)
- Existing validation rules (Section 25)
- Proposed enhancements (Section 26)
- Full implementation code (Section 27)
- Seed validation integration (Section 28)

### Key Files to Modify
1. **`business/sdk/tablebuilder/whitelists.go`** - NEW: Create with all allowed values (see Section 27, Step 1)
2. **`business/sdk/tablebuilder/validation.go`** - ENHANCE: Add `ValidateConfig()` method (see Section 27, Step 2)
3. **`business/sdk/tablebuilder/configstore.go`** - UPDATE: Use new validation (optional, see Section 28)

### How Validation Integrates with Seeding
Validation is **already wired** into the seed flow:
```
seedFrontend.go → configStore.Create() → config.Validate() → DB insert
```
Enhancing `Config.Validate()` automatically validates all seed configs in `seedmodels/tables.go`.

### To Test After Implementation
```bash
make seed-frontend  # Validates all seed configs
make test           # Also validates via test seeding
```

---

## Implementation Plan

### Phase 1: Copy Documentation
1. Create `business/sdk/tablebuilder/docs/` directory
2. Copy this file to `business/sdk/tablebuilder/docs/config-validator-spec.md`

### Phase 2: Build Validator
Based on this specification, implement enhanced validation in `validation.go`

---

## Table of Contents

1. [Config (Root)](#1-config-root)
2. [DataSource](#2-datasource)
3. [SelectConfig](#3-selectconfig)
4. [ColumnDefinition](#4-columndefinition)
5. [ForeignTable](#5-foreigntable)
6. [ComputedColumn](#6-computedcolumn)
7. [Join](#7-join)
8. [Filter](#8-filter)
9. [Sort](#9-sort)
10. [MetricConfig (Charts)](#10-metricconfig-charts)
11. [ExpressionConfig](#11-expressionconfig)
12. [GroupByConfig](#12-groupbyconfig)
13. [VisualSettings](#13-visualsettings)
14. [ColumnConfig](#14-columnconfig)
15. [FormatConfig](#15-formatconfig)
16. [EditableConfig](#16-editableconfig)
17. [LinkConfig](#17-linkconfig)
18. [LookupConfig](#18-lookupconfig)
19. [ConditionalFormat](#19-conditionalformat)
20. [Action](#20-action)
21. [PaginationConfig](#21-paginationconfig)
22. [Permissions](#22-permissions)
23. [Chart Visual Settings](#23-chart-visual-settings)
24. [Allowed Values (Whitelists)](#24-allowed-values-whitelists)
25. [Existing Validation Rules](#25-existing-validation-rules)
26. [Proposed Validator Enhancements](#26-proposed-validator-enhancements)

---

## 1. Config (Root)

The top-level configuration object.

| Field | Type | Required | Description | Validation Rules |
|-------|------|----------|-------------|------------------|
| `title` | `string` | **Yes** | Widget display title | Must be non-empty |
| `widget_type` | `string` | No | Type of widget | `"table"` or `"chart"` |
| `visualization` | `string` | No | Visualization subtype | Depends on widget_type |
| `position_x` | `int` | No | Dashboard X position | >= 0 |
| `position_y` | `int` | No | Dashboard Y position | >= 0 |
| `width` | `int` | No | Widget width (grid units) | > 0, typically 1-12 |
| `height` | `int` | No | Widget height (grid units) | > 0 |
| `data_source` | `[]DataSource` | **Yes** | Data source configurations | At least one required |
| `refresh_interval` | `int` | No | Auto-refresh seconds | >= 0 (0 = disabled) |
| `refresh_mode` | `string` | No | Refresh behavior | `"polling"` or `"manual"` |
| `visual_settings` | `VisualSettings` | No | Visual configuration | Required for table widgets |
| `permissions` | `Permissions` | No | Access control | - |

**Source:** [model.go:22-35](business/sdk/tablebuilder/model.go#L22-L35)

---

## 2. DataSource

Defines where data comes from and how to query it.

| Field | Type | Required | Description | Validation Rules |
|-------|------|----------|-------------|------------------|
| `type` | `string` | No | Query type | `"query"`, `"view"`, `"viewcount"`, `"rpc"` |
| `source` | `string` | **Yes** | Table/view/function name | Must be non-empty, valid identifier |
| `schema` | `string` | No | Database schema | Valid schema name |
| `select` | `SelectConfig` | No | Column selection | Required for query/view types |
| `args` | `map[string]any` | No | Query arguments | For RPC calls |
| `select_by` | `string` | No | Selection criteria | - |
| `parent_source` | `string` | No | Parent data source | For nested queries |
| `joins` | `[]Join` | No | Explicit table joins | - |
| `filters` | `[]Filter` | No | WHERE conditions | - |
| `sort` | `[]Sort` | No | ORDER BY clauses | - |
| `rows` | `int` | No | Row limit | > 0 |
| `metrics` | `[]MetricConfig` | No | Chart aggregations | For chart widgets |
| `group_by` | `[]GroupByConfig` | No | Chart grouping | For chart widgets |

**Source:** [model.go:38-54](business/sdk/tablebuilder/model.go#L38-L54)

---

## 3. SelectConfig

Defines what columns to select from the data source.

| Field | Type | Required | Description | Validation Rules |
|-------|------|----------|-------------|------------------|
| `columns` | `[]ColumnDefinition` | No | Direct column selections | - |
| `foreign_tables` | `[]ForeignTable` | No | Related table joins | - |
| `client_computed_columns` | `[]ComputedColumn` | No | Client-side calculations | - |

**Source:** [model.go:113-118](business/sdk/tablebuilder/model.go#L113-L118)

---

## 4. ColumnDefinition

Defines a single column selection.

| Field | Type | Required | Description | Validation Rules |
|-------|------|----------|-------------|------------------|
| `name` | `string` | **Yes** | Column name in database | Valid identifier |
| `alias` | `string` | No | Output alias | Valid identifier |
| `table_column` | `string` | No | Qualified name (`table.column`) | Format: `table.column` |

**Source:** [model.go:121-125](business/sdk/tablebuilder/model.go#L121-L125)

**Notes:**
- The output field name is determined by: `alias` > `table_column` > `name`
- Must have corresponding entry in `VisualSettings.Columns` with valid `Type`

---

## 5. ForeignTable

Defines a related table to join.

| Field | Type | Required | Description | Validation Rules |
|-------|------|----------|-------------|------------------|
| `table` | `string` | **Yes** | Related table name | Valid identifier |
| `alias` | `string` | No | Table alias | Valid identifier; required for self-joins |
| `schema` | `string` | No | Table schema | Defaults to `public` |
| `relationship_from` | `string` | **Yes** | Join from column | Format: `table.column` |
| `relationship_to` | `string` | **Yes** | Join to column | Format: `table.column` |
| `join_type` | `string` | No | Join type | `"inner"`, `"left"`, `"right"`, `"full"` |
| `columns` | `[]ColumnDefinition` | No | Columns to select | - |
| `foreign_tables` | `[]ForeignTable` | No | Nested foreign tables | Recursive structure |
| `relationship_direction` | `string` | No | Direction hint | `"parent_to_child"`, `"child_to_parent"` |

**Source:** [model.go:128-138](business/sdk/tablebuilder/model.go#L128-L138)

---

## 6. ComputedColumn

Defines a client-side computed column.

| Field | Type | Required | Description | Validation Rules |
|-------|------|----------|-------------|------------------|
| `name` | `string` | **Yes** | Output column name | Valid identifier |
| `expression` | `string` | **Yes** | JavaScript-like expression | Valid expression syntax |

**Source:** [model.go:141-144](business/sdk/tablebuilder/model.go#L141-L144)

**Expression Examples:**
- `current_quantity <= reorder_point ? 'low' : 'normal'`
- `(current_quantity / maximum_stock) * 100`
- `daysUntil(due_date)`

---

## 7. Join

Defines an explicit table join (alternative to ForeignTable).

| Field | Type | Required | Description | Validation Rules |
|-------|------|----------|-------------|------------------|
| `table` | `string` | **Yes** | Table to join | Valid identifier |
| `schema` | `string` | No | Table schema | Defaults to `public` |
| `type` | `string` | **Yes** | Join type | `"inner"`, `"left"`, `"right"`, `"full"` |
| `on` | `string` | **Yes** | Join condition | SQL ON clause |

**Source:** [model.go:147-152](business/sdk/tablebuilder/model.go#L147-L152)

---

## 8. Filter

Defines a WHERE clause condition.

| Field | Type | Required | Description | Validation Rules |
|-------|------|----------|-------------|------------------|
| `column` | `string` | **Yes** | Column to filter | Valid column reference |
| `operator` | `string` | **Yes** | Comparison operator | See allowed operators |
| `value` | `any` | **Yes** | Filter value | Type must match column |
| `dynamic` | `bool` | No | Is runtime filter | - |
| `label` | `string` | No | UI label | - |
| `control_type` | `string` | No | UI control type | - |

**Source:** [model.go:155-162](business/sdk/tablebuilder/model.go#L155-L162)

**Allowed Operators:**
- `eq` - equals
- `neq` - not equals
- `gt` - greater than
- `gte` - greater than or equal
- `lt` - less than
- `lte` - less than or equal
- `in` - in array
- `like` - pattern match (case-sensitive)
- `ilike` - pattern match (case-insensitive)

---

## 9. Sort

Defines ORDER BY configuration.

| Field | Type | Required | Description | Validation Rules |
|-------|------|----------|-------------|------------------|
| `column` | `string` | **Yes** | Column to sort | Valid column reference |
| `direction` | `string` | **Yes** | Sort direction | `"asc"` or `"desc"` |
| `priority` | `int` | No | Sort order | >= 0 |
| `custom_order` | `[]string` | No | Custom sort values | For CASE-based sorting |

**Source:** [model.go:165-170](business/sdk/tablebuilder/model.go#L165-L170)

---

## 10. MetricConfig (Charts)

Defines an aggregated metric for charts.

| Field | Type | Required | Description | Validation Rules |
|-------|------|----------|-------------|------------------|
| `name` | `string` | **Yes** | Output alias | Non-empty, valid identifier |
| `function` | `string` | **Yes** | Aggregate function | See allowed functions |
| `column` | `string` | Cond. | Column to aggregate | Required if no expression |
| `expression` | `*ExpressionConfig` | Cond. | Multi-column math | Required if no column |

**Source:** [model.go:61-66](business/sdk/tablebuilder/model.go#L61-L66)

**Validation:**
- Name is required
- Function must be in allowed list
- Either `column` OR `expression` (not both, not neither)
- Column must pass `columnRefPattern` regex

---

## 11. ExpressionConfig

Defines multi-column arithmetic for metrics.

| Field | Type | Required | Description | Validation Rules |
|-------|------|----------|-------------|------------------|
| `operator` | `string` | **Yes** | Math operator | `"multiply"`, `"add"`, `"subtract"`, `"divide"` |
| `columns` | `[]string` | **Yes** | Columns in expression | At least 2 columns |

**Source:** [model.go:69-72](business/sdk/tablebuilder/model.go#L69-L72)

**Validation:**
- Operator must be in allowed list
- At least 2 columns required
- Each column must pass `columnRefPattern` regex

---

## 12. GroupByConfig

Defines grouping for charts.

| Field | Type | Required | Description | Validation Rules |
|-------|------|----------|-------------|------------------|
| `column` | `string` | **Yes** | Column to group by | Valid column reference |
| `interval` | `string` | No | Time interval | `"day"`, `"week"`, `"month"`, `"quarter"`, `"year"` |
| `alias` | `string` | No | Output name | Valid identifier |
| `expression` | `bool` | No | Is raw SQL | If true, column can be SQL expression |

**Source:** [model.go:75-80](business/sdk/tablebuilder/model.go#L75-L80)

**Validation:**
- Column is required
- If `expression=false`, column must pass regex
- Alias (if present) must pass regex

---

## 13. VisualSettings

Contains all visual configuration.

| Field | Type | Required | Description | Validation Rules |
|-------|------|----------|-------------|------------------|
| `columns` | `map[string]ColumnConfig` | **Yes** | Column configurations | Keys must match select columns |
| `conditional_formatting` | `[]ConditionalFormat` | No | Conditional styling | - |
| `row_actions` | `map[string]Action` | No | Per-row actions | - |
| `table_actions` | `map[string]Action` | No | Table-wide actions | - |
| `pagination` | `*PaginationConfig` | No | Pagination settings | - |
| `theme` | `string` | No | Theme name | - |

**Source:** [model.go:177-184](business/sdk/tablebuilder/model.go#L177-L184)

---

## 14. ColumnConfig

Visual configuration for a single column.

| Field | Type | Required | Description | Validation Rules |
|-------|------|----------|-------------|------------------|
| `name` | `string` | **Yes** | Field name | Must match a selected column |
| `header` | `string` | No | Display header | - |
| `width` | `int` | No | Column width (px) | > 0 |
| `align` | `string` | No | Text alignment | `"left"`, `"center"`, `"right"` |
| `type` | `string` | **Yes** | Data type | See valid column types |
| `sortable` | `bool` | No | Enable sorting | - |
| `filterable` | `bool` | No | Enable filtering | - |
| `cell_template` | `string` | No | Custom renderer | - |
| `format` | `*FormatConfig` | No | Value formatting | - |
| `editable` | `*EditableConfig` | No | Inline editing | - |
| `link` | `*LinkConfig` | No | Clickable link | - |
| `lookup` | `*LookupConfig` | No | Dropdown config | Required when `type="lookup"` |

**Source:** [model.go:187-200](business/sdk/tablebuilder/model.go#L187-L200)

---

## 15. FormatConfig

Defines value formatting.

| Field | Type | Required | Description | Validation Rules |
|-------|------|----------|-------------|------------------|
| `type` | `string` | **Yes** | Format type | `"number"`, `"currency"`, `"date"`, `"datetime"`, `"boolean"`, `"percent"` |
| `precision` | `int` | No | Decimal places | >= 0 |
| `currency` | `string` | No | Currency code | ISO 4217 code |
| `format` | `string` | No | Date format | Go time format string |

**Source:** [model.go:203-208](business/sdk/tablebuilder/model.go#L203-L208)

---

## 16. EditableConfig

Defines inline editing configuration.

| Field | Type | Required | Description | Validation Rules |
|-------|------|----------|-------------|------------------|
| `type` | `string` | **Yes** | Input type | `"text"`, `"number"`, `"checkbox"`, `"select"`, `"date"`, `"textarea"` |
| `placeholder` | `string` | No | Input placeholder | - |

**Source:** [model.go:211-214](business/sdk/tablebuilder/model.go#L211-L214)

---

## 17. LinkConfig

Defines clickable link configuration.

| Field | Type | Required | Description | Validation Rules |
|-------|------|----------|-------------|------------------|
| `url` | `string` | **Yes** | Link URL | Valid URL template |
| `label` | `string` | **Yes** | Link text | - |

**Source:** [model.go:217-220](business/sdk/tablebuilder/model.go#L217-L220)

**URL Template Variables:**
- `{field_name}` - replaced with row value

---

## 18. LookupConfig

Defines lookup dropdown configuration.

| Field | Type | Required | Description | Validation Rules |
|-------|------|----------|-------------|------------------|
| `entity` | `string` | **Yes** | Source entity | Format: `schema.table` |
| `label_column` | `string` | **Yes** | Display column | Format: `table.column` |
| `value_column` | `string` | **Yes** | Value column | Format: `table.column` |
| `display_columns` | `[]string` | No | Additional display cols | - |

**Source:** [model.go:225-230](business/sdk/tablebuilder/model.go#L225-L230)

---

## 19. ConditionalFormat

Defines conditional styling rules.

| Field | Type | Required | Description | Validation Rules |
|-------|------|----------|-------------|------------------|
| `column` | `string` | **Yes** | Target column | Must exist in config |
| `condition` | `string` | **Yes** | Comparison operator | `"eq"`, `"neq"`, `"gt"`, `"lt"`, etc. |
| `value` | `any` | **Yes** | Comparison value | - |
| `condition2` | `string` | No | Second condition | For range checks |
| `value2` | `any` | No | Second value | Required with condition2 |
| `color` | `string` | No | Text color | Hex color or CSS color |
| `background` | `string` | No | Background color | Hex color or CSS color |
| `icon` | `string` | No | Icon name | - |

**Source:** [model.go:233-242](business/sdk/tablebuilder/model.go#L233-L242)

---

## 20. Action

Defines an action button/link.

| Field | Type | Required | Description | Validation Rules |
|-------|------|----------|-------------|------------------|
| `name` | `string` | **Yes** | Action identifier | - |
| `label` | `string` | **Yes** | Display label | - |
| `icon` | `string` | No | Icon name | - |
| `action_type` | `string` | **Yes** | Action type | `"link"`, `"modal"`, `"export"`, `"print"`, `"custom"` |
| `url` | `string` | Cond. | Target URL | Required for link type |
| `component` | `string` | No | Component name | For modal type |
| `params` | `map[string]any` | No | Action parameters | - |
| `format` | `string` | No | Export format | For export type |

**Source:** [model.go:245-254](business/sdk/tablebuilder/model.go#L245-L254)

---

## 21. PaginationConfig

Defines pagination settings.

| Field | Type | Required | Description | Validation Rules |
|-------|------|----------|-------------|------------------|
| `enabled` | `bool` | **Yes** | Pagination enabled | - |
| `page_sizes` | `[]int` | No | Available sizes | All > 0 |
| `default_page_size` | `int` | No | Default size | Must be in page_sizes |

**Source:** [model.go:257-261](business/sdk/tablebuilder/model.go#L257-L261)

---

## 22. Permissions

Defines access control.

| Field | Type | Required | Description | Validation Rules |
|-------|------|----------|-------------|------------------|
| `roles` | `[]string` | No | Allowed roles | Valid role names |
| `actions` | `[]string` | No | Allowed actions | `"view"`, `"edit"`, `"delete"`, `"export"` |

**Source:** [model.go:264-267](business/sdk/tablebuilder/model.go#L264-L267)

---

## 23. Chart Visual Settings

Extended settings for chart widgets.

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `chart_type` | `string` | **Yes** | Chart visualization type |
| `category_column` | `string` | No | X-axis category column |
| `value_columns` | `[]string` | No | Y-axis value columns |
| `aggregate_function` | `string` | No | Default aggregation |
| `group_by` | `string` | No | Grouping column |
| `x_axis` | `*AxisConfig` | No | X-axis configuration |
| `y_axis` | `*AxisConfig` | No | Y-axis configuration |
| `y2_axis` | `*AxisConfig` | No | Secondary Y-axis |
| `colors` | `[]string` | No | Color palette |
| `legend` | `*LegendConfig` | No | Legend configuration |
| `tooltip` | `*TooltipConfig` | No | Tooltip configuration |
| `kpi` | `*KPIConfig` | No | KPI card settings |
| `series_config` | `[]SeriesConfig` | No | Per-series settings |

**Source:** [model.go:792-828](business/sdk/tablebuilder/model.go#L792-L828)

---

## 24. Allowed Values (Whitelists)

### Valid Column Types
```go
"string"   // Text, varchar, char types
"number"   // Integer, decimal, numeric types
"datetime" // Date, time, timestamp types
"boolean"  // Boolean type
"uuid"     // UUID type
"status"   // Enum/status fields
"computed" // Client-computed columns
"lookup"   // Foreign key lookups
```
**Source:** [typemapper.go:7-16](business/sdk/tablebuilder/typemapper.go#L7-L16)

### Aggregate Functions
```go
"sum"            // SUM
"count"          // COUNT
"count_distinct" // COUNT(DISTINCT ...)
"avg"            // AVG
"min"            // MIN
"max"            // MAX
```
**Source:** [model.go:87-94](business/sdk/tablebuilder/model.go#L87-L94)

### Expression Operators
```go
"multiply" // *
"add"      // +
"subtract" // -
"divide"   // /
```
**Source:** [model.go:97-102](business/sdk/tablebuilder/model.go#L97-L102)

### Time Intervals
```go
"day"
"week"
"month"
"quarter"
"year"
```
**Source:** [model.go:105-111](business/sdk/tablebuilder/model.go#L105-L111)

### Chart Types
```go
"line"
"bar"
"stacked-bar"
"stacked-area"
"combo"
"kpi"
"gauge"
"pie"
"waterfall"
"funnel"
"heatmap"
"treemap"
"gantt"
```
**Source:** [model.go:703-717](business/sdk/tablebuilder/model.go#L703-L717)

### Column Reference Pattern
```regex
^[a-zA-Z_][a-zA-Z0-9_]*(\.[a-zA-Z_][a-zA-Z0-9_]*)*$
```
Allows: `column`, `table.column`, `schema.table.column`
**Source:** [validation.go:13](business/sdk/tablebuilder/validation.go#L13)

---

## 25. Existing Validation Rules

Current validation in `Config.Validate()`:

1. **Title required** - Config must have non-empty title
2. **DataSource required** - At least one data source
3. **Source required** - Each DataSource must have a source
4. **Column types required** - All columns must have valid Type in VisualSettings (table widgets only)
5. **Recursive foreign table validation** - Nested ForeignTables also validated

Current validation for charts:

1. **MetricConfig.Name required**
2. **MetricConfig.Function must be in whitelist**
3. **MetricConfig must have Column OR Expression (not both)**
4. **ExpressionConfig.Operator must be in whitelist**
5. **ExpressionConfig must have at least 2 columns**
6. **GroupByConfig.Column required**
7. **GroupByConfig.Interval must be in whitelist (if present)**
8. **All column references must match regex pattern**

**Source:** [validation.go](business/sdk/tablebuilder/validation.go), [model.go:365-466](business/sdk/tablebuilder/model.go#L365-L466)

---

## 26. Proposed Validator Enhancements

Based on the field analysis, here are recommended additional validations:

### High Priority

1. **DataSource.Type validation** - Ensure type is in allowed list
2. **Join type validation** - `inner`, `left`, `right`, `full`
3. **Filter operator validation** - Validate against allowed operators
4. **Sort direction validation** - `asc` or `desc`
5. **ColumnConfig.align validation** - `left`, `center`, `right`
6. **FormatConfig.type validation** - Validate format types
7. **EditableConfig.type validation** - Validate input types
8. **Action.action_type validation** - Validate action types

### Medium Priority

1. **VisualSettings.Columns key matching** - Keys must match selected columns
2. **LookupConfig required for lookup type** - If type="lookup", lookup config required
3. **Link URL validation** - Basic URL format check
4. **PaginationConfig consistency** - default_page_size in page_sizes
5. **ForeignTable relationship validation** - Format check for from/to columns
6. **Width/Height positive** - Grid dimensions > 0
7. **RefreshInterval non-negative** - >= 0

### Low Priority (Schema Validation)

1. **Schema existence** - Verify schema exists in database
2. **Table existence** - Verify tables exist
3. **Column existence** - Verify columns exist in tables
4. **Foreign key validity** - Verify FK relationships match DB schema

### Security Validations

1. **SQL injection prevention** - Already handled via columnRefPattern
2. **Expression sanitization** - Validate computed column expressions
3. **Role validation** - Verify roles exist in system

---

## Implementation Notes

### Validation Function Signature
```go
func (c *Config) Validate() error
func (c *Config) ValidateWithSchema(db *sqlx.DB) error // Deep validation
```

### Error Types
From [errors.go](business/sdk/tablebuilder/errors.go):
- `ErrInvalidConfig` - General config error
- `ErrNoDataSource` - Missing data source
- `ErrInvalidDataSource` - Bad data source config
- `ErrColumnNotFound` - Missing column
- `ErrInvalidColumn` - Bad column config
- `ErrMissingColumnType` - Column lacks type
- `ErrInvalidFilter` - Bad filter config
- `ErrInvalidSort` - Bad sort config
- `ErrInvalidJoin` - Bad join config
- `ErrInvalidExpression` - Bad expression

### Validation Approach
1. Validate structure (required fields, types)
2. Validate values (whitelists, formats)
3. Validate relationships (column references match)
4. Optional: Validate against database schema

---

## 27. Implementation Guide

### File Structure

```
business/sdk/tablebuilder/
├── model.go           # Data structures (existing)
├── errors.go          # Error definitions (existing)
├── validation.go      # Core validation (enhance)
├── whitelists.go      # NEW: Centralized allowed values
├── validator/         # NEW: Modular validators
│   ├── config.go      # Config-level validation
│   ├── datasource.go  # DataSource validation
│   ├── columns.go     # Column & VisualSettings validation
│   ├── filters.go     # Filter validation
│   ├── sorts.go       # Sort validation
│   ├── joins.go       # Join & ForeignTable validation
│   ├── charts.go      # Chart-specific validation
│   └── schema.go      # Database schema validation (optional)
└── docs/
    └── config-validator-spec.md  # This document
```

### Step 1: Create Centralized Whitelists

**File:** `business/sdk/tablebuilder/whitelists.go`

```go
package tablebuilder

// AllowedDataSourceTypes defines valid data source types
var AllowedDataSourceTypes = map[string]bool{
    "query":     true,
    "view":      true,
    "viewcount": true,
    "rpc":       true,
}

// AllowedJoinTypes defines valid SQL join types
var AllowedJoinTypes = map[string]bool{
    "inner": true,
    "left":  true,
    "right": true,
    "full":  true,
}

// AllowedFilterOperators defines valid filter comparison operators
var AllowedFilterOperators = map[string]bool{
    "eq":    true,
    "neq":   true,
    "gt":    true,
    "gte":   true,
    "lt":    true,
    "lte":   true,
    "in":    true,
    "like":  true,
    "ilike": true,
}

// AllowedSortDirections defines valid sort directions
var AllowedSortDirections = map[string]bool{
    "asc":  true,
    "desc": true,
}

// AllowedAlignments defines valid text alignments
var AllowedAlignments = map[string]bool{
    "left":   true,
    "center": true,
    "right":  true,
}

// AllowedFormatTypes defines valid format types
var AllowedFormatTypes = map[string]bool{
    "number":   true,
    "currency": true,
    "date":     true,
    "datetime": true,
    "boolean":  true,
    "percent":  true,
}

// AllowedEditableTypes defines valid editable input types
var AllowedEditableTypes = map[string]bool{
    "text":     true,
    "number":   true,
    "checkbox": true,
    "select":   true,
    "date":     true,
    "textarea": true,
}

// AllowedActionTypes defines valid action types
var AllowedActionTypes = map[string]bool{
    "link":   true,
    "modal":  true,
    "export": true,
    "print":  true,
    "custom": true,
}

// AllowedWidgetTypes defines valid widget types
var AllowedWidgetTypes = map[string]bool{
    "table": true,
    "chart": true,
}

// AllowedRefreshModes defines valid refresh modes
var AllowedRefreshModes = map[string]bool{
    "polling": true,
    "manual":  true,
}

// AllowedPermissionActions defines valid permission actions
var AllowedPermissionActions = map[string]bool{
    "view":   true,
    "edit":   true,
    "delete": true,
    "export": true,
}
```

### Step 2: Enhance Config.Validate()

**File:** `business/sdk/tablebuilder/validation.go`

```go
package tablebuilder

import (
    "fmt"
    "regexp"
)

// ValidationResult contains all validation errors
type ValidationResult struct {
    Errors   []ValidationError
    Warnings []ValidationWarning
}

type ValidationError struct {
    Field   string
    Message string
    Code    string
}

type ValidationWarning struct {
    Field   string
    Message string
}

func (r *ValidationResult) HasErrors() bool {
    return len(r.Errors) > 0
}

func (r *ValidationResult) AddError(field, message, code string) {
    r.Errors = append(r.Errors, ValidationError{
        Field:   field,
        Message: message,
        Code:    code,
    })
}

func (r *ValidationResult) AddWarning(field, message string) {
    r.Warnings = append(r.Warnings, ValidationWarning{
        Field:   field,
        Message: message,
    })
}

// ValidateConfig performs comprehensive validation
func (c *Config) ValidateConfig() *ValidationResult {
    result := &ValidationResult{}

    // 1. Root-level validation
    c.validateRoot(result)

    // 2. DataSource validation
    for i, ds := range c.DataSource {
        c.validateDataSource(result, ds, fmt.Sprintf("data_source[%d]", i))
    }

    // 3. VisualSettings validation (for table widgets)
    if c.WidgetType != "chart" {
        c.validateVisualSettings(result)
    }

    // 4. Permissions validation
    c.validatePermissions(result)

    return result
}

func (c *Config) validateRoot(result *ValidationResult) {
    if c.Title == "" {
        result.AddError("title", "title is required", "REQUIRED")
    }

    if c.WidgetType != "" && !AllowedWidgetTypes[c.WidgetType] {
        result.AddError("widget_type", fmt.Sprintf("invalid widget type: %s", c.WidgetType), "INVALID_VALUE")
    }

    if c.RefreshMode != "" && !AllowedRefreshModes[c.RefreshMode] {
        result.AddError("refresh_mode", fmt.Sprintf("invalid refresh mode: %s", c.RefreshMode), "INVALID_VALUE")
    }

    if c.RefreshInterval < 0 {
        result.AddError("refresh_interval", "refresh_interval must be >= 0", "INVALID_VALUE")
    }

    if c.Width < 0 {
        result.AddError("width", "width must be >= 0", "INVALID_VALUE")
    }

    if c.Height < 0 {
        result.AddError("height", "height must be >= 0", "INVALID_VALUE")
    }

    if len(c.DataSource) == 0 {
        result.AddError("data_source", "at least one data source is required", "REQUIRED")
    }
}

func (c *Config) validateDataSource(result *ValidationResult, ds DataSource, prefix string) {
    if ds.Source == "" {
        result.AddError(prefix+".source", "source is required", "REQUIRED")
    }

    if ds.Type != "" && !AllowedDataSourceTypes[ds.Type] {
        result.AddError(prefix+".type", fmt.Sprintf("invalid type: %s", ds.Type), "INVALID_VALUE")
    }

    // Validate filters
    for i, f := range ds.Filters {
        c.validateFilter(result, f, fmt.Sprintf("%s.filters[%d]", prefix, i))
    }

    // Validate sorts
    for i, s := range ds.Sort {
        c.validateSort(result, s, fmt.Sprintf("%s.sort[%d]", prefix, i))
    }

    // Validate joins
    for i, j := range ds.Joins {
        c.validateJoin(result, j, fmt.Sprintf("%s.joins[%d]", prefix, i))
    }

    // Validate foreign tables
    for i, ft := range ds.Select.ForeignTables {
        c.validateForeignTable(result, ft, fmt.Sprintf("%s.select.foreign_tables[%d]", prefix, i))
    }

    // Validate metrics (for charts)
    for i, m := range ds.Metrics {
        c.validateMetric(result, m, fmt.Sprintf("%s.metrics[%d]", prefix, i))
    }

    // Validate group_by (for charts)
    for i, g := range ds.GroupBy {
        c.validateGroupBy(result, g, fmt.Sprintf("%s.group_by[%d]", prefix, i))
    }
}

func (c *Config) validateFilter(result *ValidationResult, f Filter, prefix string) {
    if f.Column == "" {
        result.AddError(prefix+".column", "column is required", "REQUIRED")
    } else if !isValidColumnReference(f.Column) {
        result.AddError(prefix+".column", fmt.Sprintf("invalid column reference: %s", f.Column), "INVALID_FORMAT")
    }

    if f.Operator == "" {
        result.AddError(prefix+".operator", "operator is required", "REQUIRED")
    } else if !AllowedFilterOperators[f.Operator] {
        result.AddError(prefix+".operator", fmt.Sprintf("invalid operator: %s", f.Operator), "INVALID_VALUE")
    }
}

func (c *Config) validateSort(result *ValidationResult, s Sort, prefix string) {
    if s.Column == "" {
        result.AddError(prefix+".column", "column is required", "REQUIRED")
    } else if !isValidColumnReference(s.Column) {
        result.AddError(prefix+".column", fmt.Sprintf("invalid column reference: %s", s.Column), "INVALID_FORMAT")
    }

    if s.Direction == "" {
        result.AddError(prefix+".direction", "direction is required", "REQUIRED")
    } else if !AllowedSortDirections[s.Direction] {
        result.AddError(prefix+".direction", fmt.Sprintf("invalid direction: %s", s.Direction), "INVALID_VALUE")
    }
}

func (c *Config) validateJoin(result *ValidationResult, j Join, prefix string) {
    if j.Table == "" {
        result.AddError(prefix+".table", "table is required", "REQUIRED")
    }

    if j.Type == "" {
        result.AddError(prefix+".type", "type is required", "REQUIRED")
    } else if !AllowedJoinTypes[j.Type] {
        result.AddError(prefix+".type", fmt.Sprintf("invalid join type: %s", j.Type), "INVALID_VALUE")
    }

    if j.On == "" {
        result.AddError(prefix+".on", "on condition is required", "REQUIRED")
    }
}

func (c *Config) validateForeignTable(result *ValidationResult, ft ForeignTable, prefix string) {
    if ft.Table == "" {
        result.AddError(prefix+".table", "table is required", "REQUIRED")
    }

    if ft.RelationshipFrom == "" {
        result.AddError(prefix+".relationship_from", "relationship_from is required", "REQUIRED")
    } else if !isValidColumnReference(ft.RelationshipFrom) {
        result.AddError(prefix+".relationship_from", fmt.Sprintf("invalid column reference: %s", ft.RelationshipFrom), "INVALID_FORMAT")
    }

    if ft.RelationshipTo == "" {
        result.AddError(prefix+".relationship_to", "relationship_to is required", "REQUIRED")
    } else if !isValidColumnReference(ft.RelationshipTo) {
        result.AddError(prefix+".relationship_to", fmt.Sprintf("invalid column reference: %s", ft.RelationshipTo), "INVALID_FORMAT")
    }

    if ft.JoinType != "" && !AllowedJoinTypes[ft.JoinType] {
        result.AddError(prefix+".join_type", fmt.Sprintf("invalid join type: %s", ft.JoinType), "INVALID_VALUE")
    }

    // Recursively validate nested foreign tables
    for i, nested := range ft.ForeignTables {
        c.validateForeignTable(result, nested, fmt.Sprintf("%s.foreign_tables[%d]", prefix, i))
    }
}

func (c *Config) validateVisualSettings(result *ValidationResult) {
    prefix := "visual_settings"

    // Validate each column config
    for name, col := range c.VisualSettings.Columns {
        colPrefix := fmt.Sprintf("%s.columns[%s]", prefix, name)

        if col.Type == "" {
            result.AddError(colPrefix+".type", "type is required", "REQUIRED")
        } else if !IsValidColumnType(col.Type) {
            result.AddError(colPrefix+".type", fmt.Sprintf("invalid column type: %s", col.Type), "INVALID_VALUE")
        }

        if col.Align != "" && !AllowedAlignments[col.Align] {
            result.AddError(colPrefix+".align", fmt.Sprintf("invalid alignment: %s", col.Align), "INVALID_VALUE")
        }

        // Validate format config
        if col.Format != nil {
            c.validateFormatConfig(result, col.Format, colPrefix+".format")
        }

        // Validate editable config
        if col.Editable != nil {
            c.validateEditableConfig(result, col.Editable, colPrefix+".editable")
        }

        // Validate lookup config when type is "lookup"
        if col.Type == "lookup" && col.Lookup == nil {
            result.AddError(colPrefix+".lookup", "lookup config required when type is 'lookup'", "REQUIRED")
        }

        if col.Lookup != nil {
            c.validateLookupConfig(result, col.Lookup, colPrefix+".lookup")
        }
    }

    // Validate pagination config
    if c.VisualSettings.Pagination != nil {
        c.validatePaginationConfig(result, c.VisualSettings.Pagination, prefix+".pagination")
    }

    // Validate conditional formatting
    for i, cf := range c.VisualSettings.ConditionalFormatting {
        c.validateConditionalFormat(result, cf, fmt.Sprintf("%s.conditional_formatting[%d]", prefix, i))
    }

    // Validate row actions
    for name, action := range c.VisualSettings.RowActions {
        c.validateAction(result, action, fmt.Sprintf("%s.row_actions[%s]", prefix, name))
    }

    // Validate table actions
    for name, action := range c.VisualSettings.TableActions {
        c.validateAction(result, action, fmt.Sprintf("%s.table_actions[%s]", prefix, name))
    }
}

func (c *Config) validateFormatConfig(result *ValidationResult, f *FormatConfig, prefix string) {
    if f.Type != "" && !AllowedFormatTypes[f.Type] {
        result.AddError(prefix+".type", fmt.Sprintf("invalid format type: %s", f.Type), "INVALID_VALUE")
    }

    if f.Precision < 0 {
        result.AddError(prefix+".precision", "precision must be >= 0", "INVALID_VALUE")
    }
}

func (c *Config) validateEditableConfig(result *ValidationResult, e *EditableConfig, prefix string) {
    if e.Type == "" {
        result.AddError(prefix+".type", "type is required", "REQUIRED")
    } else if !AllowedEditableTypes[e.Type] {
        result.AddError(prefix+".type", fmt.Sprintf("invalid editable type: %s", e.Type), "INVALID_VALUE")
    }
}

func (c *Config) validateLookupConfig(result *ValidationResult, l *LookupConfig, prefix string) {
    if l.Entity == "" {
        result.AddError(prefix+".entity", "entity is required", "REQUIRED")
    }

    if l.LabelColumn == "" {
        result.AddError(prefix+".label_column", "label_column is required", "REQUIRED")
    }

    if l.ValueColumn == "" {
        result.AddError(prefix+".value_column", "value_column is required", "REQUIRED")
    }
}

func (c *Config) validatePaginationConfig(result *ValidationResult, p *PaginationConfig, prefix string) {
    for i, size := range p.PageSizes {
        if size <= 0 {
            result.AddError(fmt.Sprintf("%s.page_sizes[%d]", prefix, i), "page size must be > 0", "INVALID_VALUE")
        }
    }

    if p.DefaultPageSize > 0 && len(p.PageSizes) > 0 {
        found := false
        for _, size := range p.PageSizes {
            if size == p.DefaultPageSize {
                found = true
                break
            }
        }
        if !found {
            result.AddWarning(prefix+".default_page_size", "default_page_size should be in page_sizes")
        }
    }
}

func (c *Config) validateConditionalFormat(result *ValidationResult, cf ConditionalFormat, prefix string) {
    if cf.Column == "" {
        result.AddError(prefix+".column", "column is required", "REQUIRED")
    }

    if cf.Condition == "" {
        result.AddError(prefix+".condition", "condition is required", "REQUIRED")
    } else if !AllowedFilterOperators[cf.Condition] {
        result.AddError(prefix+".condition", fmt.Sprintf("invalid condition: %s", cf.Condition), "INVALID_VALUE")
    }

    if cf.Condition2 != "" && !AllowedFilterOperators[cf.Condition2] {
        result.AddError(prefix+".condition2", fmt.Sprintf("invalid condition2: %s", cf.Condition2), "INVALID_VALUE")
    }
}

func (c *Config) validateAction(result *ValidationResult, a Action, prefix string) {
    if a.Name == "" {
        result.AddError(prefix+".name", "name is required", "REQUIRED")
    }

    if a.Label == "" {
        result.AddError(prefix+".label", "label is required", "REQUIRED")
    }

    if a.ActionType == "" {
        result.AddError(prefix+".action_type", "action_type is required", "REQUIRED")
    } else if !AllowedActionTypes[a.ActionType] {
        result.AddError(prefix+".action_type", fmt.Sprintf("invalid action_type: %s", a.ActionType), "INVALID_VALUE")
    }

    // URL is required for link type
    if a.ActionType == "link" && a.URL == "" {
        result.AddError(prefix+".url", "url is required for link action type", "REQUIRED")
    }
}

func (c *Config) validateMetric(result *ValidationResult, m MetricConfig, prefix string) {
    if err := ValidateMetricConfig(m); err != nil {
        result.AddError(prefix, err.Error(), "INVALID_CONFIG")
    }
}

func (c *Config) validateGroupBy(result *ValidationResult, g GroupByConfig, prefix string) {
    if err := ValidateGroupByConfig(&g); err != nil {
        result.AddError(prefix, err.Error(), "INVALID_CONFIG")
    }
}

func (c *Config) validatePermissions(result *ValidationResult) {
    prefix := "permissions"

    for i, action := range c.Permissions.Actions {
        if !AllowedPermissionActions[action] {
            result.AddError(fmt.Sprintf("%s.actions[%d]", prefix, i), fmt.Sprintf("invalid action: %s", action), "INVALID_VALUE")
        }
    }
}
```

### Step 3: Add Validation to API Layer

**File:** `app/domain/dataapp/dataapp.go` (enhance existing)

```go
// ValidateConfig validates a configuration without saving
func (a *App) ValidateConfig(ctx context.Context, config tablebuilder.Config) (*tablebuilder.ValidationResult, error) {
    result := config.ValidateConfig()
    return result, nil
}

// Create validates before saving
func (a *App) Create(ctx context.Context, app NewTableConfig) (TableConfig, error) {
    // Parse the config
    var config tablebuilder.Config
    if err := json.Unmarshal(app.Config, &config); err != nil {
        return TableConfig{}, errs.Newf(errs.InvalidArgument, "invalid config JSON: %s", err)
    }

    // Validate
    result := config.ValidateConfig()
    if result.HasErrors() {
        return TableConfig{}, errs.Newf(errs.InvalidArgument, "config validation failed: %v", result.Errors)
    }

    // Continue with create...
}
```

### Step 4: Testing Strategy

**File:** `business/sdk/tablebuilder/validation_test.go`

```go
package tablebuilder_test

import (
    "testing"

    "github.com/timmaaaz/ichor/business/sdk/tablebuilder"
)

func TestValidateConfig_RequiredFields(t *testing.T) {
    tests := []struct {
        name        string
        config      tablebuilder.Config
        expectError bool
        errorField  string
    }{
        {
            name:        "empty title",
            config:      tablebuilder.Config{},
            expectError: true,
            errorField:  "title",
        },
        {
            name: "missing data source",
            config: tablebuilder.Config{
                Title: "Test",
            },
            expectError: true,
            errorField:  "data_source",
        },
        {
            name: "valid minimal config",
            config: tablebuilder.Config{
                Title: "Test",
                DataSource: []tablebuilder.DataSource{
                    {Source: "test_table"},
                },
            },
            expectError: false,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := tt.config.ValidateConfig()

            if tt.expectError && !result.HasErrors() {
                t.Errorf("expected error for field %s, got none", tt.errorField)
            }

            if !tt.expectError && result.HasErrors() {
                t.Errorf("unexpected errors: %v", result.Errors)
            }
        })
    }
}

func TestValidateConfig_Whitelists(t *testing.T) {
    // Test each whitelist validation...
}

func TestValidateConfig_NestedStructures(t *testing.T) {
    // Test foreign tables, filters, sorts, etc.
}
```

### Migration Path

1. **Phase 1:** Add `whitelists.go` with centralized allowed values
2. **Phase 2:** Add `ValidateConfig()` method alongside existing `Validate()`
3. **Phase 3:** Add tests for new validation
4. **Phase 4:** Update API layer to use new validation
5. **Phase 5:** Deprecate old `Validate()` in favor of `ValidateConfig()`
6. **Phase 6:** Add optional schema validation (requires DB connection)

---

## 28. Seed Validation Integration

### How Validation Works with Seeding

The validation is **already wired into the seeding flow**. Here's the chain:

```
seedFrontend.go (InsertSeedData)
    ↓
configStore.Create(ctx, name, desc, config, userID)   ← business/sdk/tablebuilder/configstore.go:31
    ↓
config.Validate()                                      ← business/sdk/tablebuilder/model.go:365
    ↓
[If validation passes] → json.Marshal → DB INSERT
[If validation fails]  → Returns error, seeding stops
```

**Key file:** `business/sdk/tablebuilder/configstore.go`
```go
// Line 31-35
func (s *ConfigStore) Create(ctx context.Context, name, description string, config *Config, userID uuid.UUID) (*StoredConfig, error) {
    // Validate the configuration
    if err := config.Validate(); err != nil {
        return nil, fmt.Errorf("validate config: %w", err)
    }
    // ... continues to marshal and insert
}
```

### What This Means for Implementation

**No changes needed to seedFrontend.go** - simply enhancing `Config.Validate()` (or replacing it with `ValidateConfig()`) will automatically validate all seed configs.

When you run the seed:
- Each config in `seedmodels/tables.go` goes through `configStore.Create()`
- `configStore.Create()` calls `config.Validate()`
- If any validation fails, you get an error like:
  ```
  seed frontend: create table config "Products List": validate config: invalid filter operator: badop
  ```

### Seed Config Files

All seed configs are defined in:
- **Table configs:** `business/sdk/dbtest/seedmodels/tables.go`
- **Chart configs:** `business/sdk/dbtest/seedmodels/tables.go` (ChartConfigs array)

The seeding is triggered from:
- **Production seed:** `api/cmd/tooling/admin/commands/seedFrontend.go`
- **Test seed:** `business/sdk/dbtest/seedFrontend.go`

### Enhanced Error Messages for Seeding

To make debugging seed validation easier, consider this pattern in `configstore.go`:

```go
func (s *ConfigStore) Create(ctx context.Context, name, description string, config *Config, userID uuid.UUID) (*StoredConfig, error) {
    // Use enhanced validation
    result := config.ValidateConfig()
    if result.HasErrors() {
        // Format errors nicely for seed debugging
        var msgs []string
        for _, err := range result.Errors {
            msgs = append(msgs, fmt.Sprintf("%s: %s", err.Field, err.Message))
        }
        return nil, fmt.Errorf("validate config %q: %s", name, strings.Join(msgs, "; "))
    }
    // ... continues
}
```

This would produce errors like:
```
seed frontend: create table config "Products List": validate config "Products List":
  data_source[0].filters[0].operator: invalid operator: badop;
  visual_settings.columns[name].type: invalid column type: wrongtype
```

### Running Seed Validation

To test your seed configs after implementing the validator:

```bash
# Run the admin seed command (validates all configs)
make seed-frontend

# Or run tests which also seed configs
make test
```

Any invalid configs will cause the seed to fail with detailed error messages.
