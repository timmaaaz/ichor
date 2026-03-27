# Progress Summary: table-builder.md

## Overview
Dynamic SQL query builder and visualization config system. Driven entirely by JSON config; no hand-written queries. Powers tables, charts, KPIs, and heatmaps in the frontend page builder.

## Config [sdk] — `business/sdk/tablebuilder/model.go`

**Responsibility:** Define table, chart, and KPI configuration structure.

### Main Config Struct
```go
type Config struct {
    Title           string
    WidgetType      string
    Visualization   string
    PositionX       int
    PositionY       int
    Width           int
    Height          int
    DataSource      []DataSource
    RefreshInterval int
    RefreshMode     string
    VisualSettings  VisualSettings
    Permissions     Permissions
}
```

### DataSource Struct
```go
type DataSource struct {
    Type         string
    Source       string
    Schema       string
    Select       SelectConfig
    Args         map[string]any
    SelectBy     string
    ParentSource string
    Joins        []Join
    Filters      []Filter
    Sort         []Sort
    Rows         int
    Metrics      []MetricConfig
    GroupBy      []GroupByConfig
}
```

### ForeignTable (Recursive Joins)
```go
type ForeignTable struct {
    Table            string
    Alias            string
    Schema           string
    RelationshipFrom string
    RelationshipTo   string
    JoinType         string
    Columns          []ColumnDefinition
    ForeignTables    []ForeignTable   // ← recursive (nested joins)
}
```

### MetricConfig (Aggregation)
```go
type MetricConfig struct {
    Name       string
    Function   string            // must be in AllowedAggregateFunctions
    Column     string
    Expression *ExpressionConfig // optional; multi-column arithmetic
}

type ExpressionConfig struct {
    Operator string   // multiply, add, subtract, divide
    Columns  []string // ["order_line_items.quantity", "product_costs.selling_price"]
}
```

### GroupByConfig (Temporal Grouping)
```go
type GroupByConfig struct {
    Column     string
    Interval   string // must be in AllowedIntervals
    Alias      string
    Expression bool   // if true, Column is treated as raw SQL (e.g. EXTRACT(DOW FROM created_date))
}
```

### Security Whitelists

All values validated against whitelists before SQL generation:

```go
AllowedAggregateFunctions: { sum, count, count_distinct, avg, min, max }
AllowedOperators:          { multiply, add, subtract, divide }
AllowedIntervals:          { day, week, month, quarter, year }
```

**Injection-safe:** No string interpolation; all values parameterized.

### Supporting Types
- ColumnDefinition, SelectConfig, ComputedColumn
- Join, Filter, Sort
- VisualSettings, ColumnConfig, FormatConfig, EditableConfig, LinkConfig, LookupConfig
- ConditionalFormat, Action, PaginationConfig, Permissions
- TableData, TableRow, MetaData, ColumnMetadata, RelationshipInfo
- QueryParams, StoredConfig
- PageConfig, UpdatePageConfig, PageContent, LayoutConfig, ResponsiveValue
- ChartResponse, SeriesData, KPIData, ChartMeta, HeatmapData, GanttData, TreemapData
- ChartVisualSettings, SeriesConfig, AxisConfig, LegendConfig, TooltipConfig, KPIConfig

## Store [sdk] — `business/sdk/tablebuilder/store.go`

**Responsibility:** Execute queries from Config at runtime.

### Methods
```go
func NewStore(log *logger.Logger, db *sqlx.DB, options ...StoreOption) *Store
func WithIntrospection(introspection *introspectionbus.Business) StoreOption

func (s *Store) FetchTableData(ctx context.Context, config *Config, params QueryParams) (*TableData, error)
func (s *Store) FetchTableDataCount(ctx context.Context, config *Config, params QueryParams) (int, error)
func (s *Store) GetCount(ctx context.Context, ds *DataSource, params QueryParams) (int, error)
func (s *Store) QueryByPage(ctx context.Context, config *Config, pg page.Page) (*TableData, error)
```

### Key Facts
- **Validates config against whitelists** — before execution
- **Generates parameterized SQL** — no string interpolation
- **Optional introspection** — via WithIntrospection option

## ConfigStore [sdk] — `business/sdk/tablebuilder/configstore.go`

**Responsibility:** CRUD for persisted table, page, and content configs.

### Data Sources
- ⊕⊗ config.table_configs
- ⊕⊗ config.page_configs
- ⊕⊗ config.page_contents

### StoredConfig CRUD
```go
Create(ctx, name, description string, config *Config, userID uuid.UUID) (*StoredConfig, error)
Update(ctx, id uuid.UUID, name, description string, config *Config, userID uuid.UUID) (*StoredConfig, error)
Delete(ctx, id uuid.UUID) error
QueryByID(ctx, id uuid.UUID) (*StoredConfig, error)
QueryByName(ctx, name string) (*StoredConfig, error)
QueryByUser(ctx, userID uuid.UUID) ([]StoredConfig, error)
QueryAll(ctx) ([]StoredConfig, error)
LoadConfig(ctx, id uuid.UUID) (*Config, error)
LoadConfigByName(ctx, name string) (*Config, error)
ValidateStoredConfig(ctx, id uuid.UUID) error
```

### PageConfig CRUD
```go
CreatePageConfig(ctx, pc PageConfig) (*PageConfig, error)
UpdatePageConfig(ctx, pc PageConfig) (*PageConfig, error)
DeletePageConfig(ctx, id uuid.UUID) error
QueryPageByName(ctx, name string) (*PageConfig, error)
QueryPageByNameAndUserID(ctx, name string, userID uuid.UUID) (*PageConfig, error)
QueryPageByID(ctx, id uuid.UUID) (*PageConfig, error)
```

### PageContent CRUD
```go
CreatePageContent(ctx, content PageContent) (*PageContent, error)
UpdatePageContent(ctx, content PageContent) (*PageContent, error)
DeletePageContent(ctx, id uuid.UUID) error
QueryPageContentByID(ctx, id uuid.UUID) (*PageContent, error)
QueryPageContentByConfigID(ctx, pageConfigID uuid.UUID) ([]PageContent, error)
QueryPageContentWithChildren(ctx, pageConfigID uuid.UUID) ([]PageContent, error)
```

## Change Patterns

### ⚠ Adding a New Aggregate Function
Affects 3 areas:
1. `business/sdk/tablebuilder/model.go` — add to AllowedAggregateFunctions whitelist
2. `business/sdk/tablebuilder/store.go` — handle new function in SQL generation logic
3. `api/cmd/services/ichor/tests/...` — add integration test for new function

### ⚠ Adding a New DataSource Type
Affects 2 areas:
1. `business/sdk/tablebuilder/model.go` — new Type constant
2. `business/sdk/tablebuilder/store.go` — dispatch on Type in FetchTableData

### ⚠ Adding a New Visualization Type
Affects 3 areas:
1. `business/sdk/tablebuilder/model.go` — new WidgetType/Visualization constant + supporting struct if needed
2. `business/sdk/tablebuilder/store.go` — handle data shape for new vis in FetchTableData / QueryByPage
3. **Frontend page builder** — render new visualization type

### ⚠ Adding a Column to config.table_configs or config.page_configs
Affects 3 areas:
1. `business/sdk/migrate/sql/migrate.sql` — new migration version (ALTER TABLE)
2. `business/sdk/tablebuilder/configstore.go` — StoredConfig/PageConfig model + affected queries
3. `business/sdk/tablebuilder/model.go` — StoredConfig/PageConfig struct if field exposed

## Testing Patterns

### Unit Tests (No DB)
- `business/sdk/tablebuilder/builder_test.go` — SQL generation: use NewQueryBuilder() + BuildQuery(), assert with assertSQL/assertNoSQL helpers
- `business/sdk/tablebuilder/multi_groupby_test.go` — GroupBy SQL generation and validation error cases
- `business/sdk/tablebuilder/chart_test.go` — ChartTransformer unit tests

### Integration Tests (Require DB via dbtest.NewDatabase)
- `business/sdk/tablebuilder/tablebuilder_test.go` — Store.FetchTableData, QueryByPage, ConfigStore CRUD, PageConfig CRUD

### JSON Comparison
StoredConfig.Config is json.RawMessage; Postgres normalizes key order on round-trip.

**Use `dbtest.NormalizeJSONFields(got, want)` before cmp.Diff** to align byte representation without losing semantic equality checking.

See: `business/sdk/dbtest/dbtest.go` — NormalizeJSONFields (exported), normalizeJSON (internal)

### ClientComputedColumns
Expressions like `"qty <= reorder ? 'low' : 'normal'"` are JavaScript, evaluated on the frontend. Server returns nil for these columns in TableRow.

**Tests should check key presence in row, not value.**

## Critical Points
- **No hand-written queries** — all SQL generated from Config
- **Whitelist-based security** — injection-safe via parameterization + whitelist validation
- **Nested joins supported** — recursive ForeignTable struct for complex relationships
- **Temporal grouping** — Interval + Expression for date bucketing
- **Client-computed columns** — server returns nil (frontend evaluates)
- **JSON persistence** — StoredConfig.Config is json.RawMessage (opaque to backend)

## Notes for Future Development
TableBuilder is a sophisticated config-driven query system. Most changes will be:
- Adding new visualization types (moderate, requires frontend support)
- Adding new aggregate functions (low-risk, just add to whitelist)
- Adding new DataSource types (moderate, requires dispatch logic)
- Changing config column schema (risky, affects persistence layer)

The whitelist approach prevents injection while remaining extensible for legitimate use cases.
