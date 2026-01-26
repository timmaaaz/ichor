# Table Configuration Validator - Complete Field Reference

> **Note:** This document should be copied to `business/sdk/tablebuilder/docs/config-validator-spec.md` for permanent storage within the package.

This document provides a comprehensive mapping of all fields in the tablebuilder configuration system, intended to serve as the foundation for building a robust configuration validator.

---

## Implementation Plan

### Phase 1: Copy Documentation
1. Create `business/sdk/tablebuilder/docs/` directory
2. Copy this file to `business/sdk/tablebuilder/docs/config-validator-spec.md`

### Phase 2: Build Validator (Optional Follow-up)
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
| `type` | `string` | Cond. | Data type | Required unless `hidden=true`; see valid column types |
| `hidden` | `bool` | No | Hide column from display | If true, column is selected but not rendered |
| `sortable` | `bool` | No | Enable sorting | - |
| `filterable` | `bool` | No | Enable filtering | - |
| `cell_template` | `string` | No | Custom renderer | - |
| `format` | `*FormatConfig` | No | Value formatting | - |
| `editable` | `*EditableConfig` | No | Inline editing | - |
| `link` | `*LinkConfig` | No | Clickable link | - |
| `lookup` | `*LookupConfig` | No | Dropdown config | Required when `type="lookup"` |

**Source:** [model.go:187-201](business/sdk/tablebuilder/model.go#L187-L201)

**Hidden Columns:**
- When `hidden: true`, the column is selected in the query (data is available in row)
- The column is exempt from `type` validation requirement
- Frontend should skip rendering this column in the table
- Use case: columns needed for lookup labels or other data purposes but not displayed directly

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
| `label` | `string` | Cond. | Static link text | Either `label` or `label_column` required |
| `label_column` | `string` | Cond. | Column name to use as dynamic label | Takes precedence over `label` |

**Source:** [model.go:216-221](business/sdk/tablebuilder/model.go#L216-L221)

**URL Template Variables:**
- `{field_name}` - replaced with row value

**Label Behavior:**
- If `label_column` is set, frontend uses that column's value from the row as the link text
- If only `label` is set, use the static text
- `label_column` takes precedence over `label` if both are set

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
