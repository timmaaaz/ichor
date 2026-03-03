# table-builder

[sdk]=shared SDK [bus]=business layer [db]=store
→=depends on ⊕=writes ⊗=reads

---

## Overview

Dynamic SQL query builder + visualization config system driven entirely by JSON config.
No hand-written queries. Config persisted in config.table_configs and config.page_configs;
loaded by frontend page builder to drive tables, charts, KPIs, and heatmaps.

Two components:
  Store       — executes queries from a Config at runtime
  ConfigStore — CRUD for persisting/loading Config + PageConfig + PageContent

---

## Config [sdk]

file: business/sdk/tablebuilder/model.go

### Core types

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

type MetricConfig struct {
    Name       string
    Function   string    // must be in AllowedAggregateFunctions
    Column     string
    Expression string
}

type GroupByConfig struct {
    Column     string
    Interval   string    // must be in AllowedIntervals
    Alias      string
    Expression string
}
```

### Security whitelists (model.go)

```go
AllowedAggregateFunctions: { sum, count, count_distinct, avg, min, max }
AllowedOperators:          { multiply, add, subtract, divide }
AllowedIntervals:          { day, week, month, quarter, year }
```

All values validated against whitelists before SQL generation — injection-safe.

### Supporting types (model.go)

ColumnDefinition, SelectConfig, ExpressionConfig, ComputedColumn, Join, Filter, Sort,
VisualSettings, ColumnConfig, FormatConfig, EditableConfig, LinkConfig, LookupConfig,
ConditionalFormat, Action, PaginationConfig, Permissions, TableData, TableRow,
MetaData, ColumnMetadata, RelationshipInfo, QueryParams, StoredConfig,
PageConfig, UpdatePageConfig, PageContent, LayoutConfig, ResponsiveValue,
ChartResponse, SeriesData, KPIData, ChartMeta, HeatmapData, GanttData, TreemapData,
ChartVisualSettings, SeriesConfig, AxisConfig, LegendConfig, TooltipConfig, KPIConfig

---

## Store [sdk]

file: business/sdk/tablebuilder/store.go
key facts:
  - Validates config against whitelists before execution
  - Generates parameterized SQL (no string interpolation)
  - Optional introspection via WithIntrospection option

```go
func NewStore(log *logger.Logger, db *sqlx.DB, options ...StoreOption) *Store
func WithIntrospection(introspection *introspectionbus.Business) StoreOption

func (s *Store) FetchTableData(ctx context.Context, config *Config, params QueryParams) (*TableData, error)
func (s *Store) FetchTableDataCount(ctx context.Context, config *Config, params QueryParams) (int, error)
func (s *Store) GetCount(ctx context.Context, ds *DataSource, params QueryParams) (int, error)
func (s *Store) QueryByPage(ctx context.Context, config *Config, pg page.Page) (*TableData, error)
```

---

## ConfigStore [sdk]

file: business/sdk/tablebuilder/configstore.go
persistence: ⊕⊗ config.table_configs, ⊕⊗ config.page_configs, ⊕⊗ config.page_contents

StoredConfig CRUD:
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

PageConfig CRUD:
  CreatePageConfig(ctx, pc PageConfig) (*PageConfig, error)
  UpdatePageConfig(ctx, pc PageConfig) (*PageConfig, error)
  DeletePageConfig(ctx, id uuid.UUID) error
  QueryPageByName(ctx, name string) (*PageConfig, error)
  QueryPageByNameAndUserID(ctx, name string, userID uuid.UUID) (*PageConfig, error)
  QueryPageByID(ctx, id uuid.UUID) (*PageConfig, error)

PageContent CRUD:
  CreatePageContent(ctx, content PageContent) (*PageContent, error)
  UpdatePageContent(ctx, content PageContent) (*PageContent, error)
  DeletePageContent(ctx, id uuid.UUID) error
  QueryPageContentByID(ctx, id uuid.UUID) (*PageContent, error)
  QueryPageContentByConfigID(ctx, pageConfigID uuid.UUID) ([]PageContent, error)
  QueryPageContentWithChildren(ctx, pageConfigID uuid.UUID) ([]PageContent, error)

---

## ⚠ Adding a new aggregate function

  business/sdk/tablebuilder/model.go          (add to AllowedAggregateFunctions whitelist)
  business/sdk/tablebuilder/store.go          (handle new function in SQL generation logic)
  api/cmd/services/ichor/tests/...            (add integration test for new function)

## ⚠ Adding a new DataSource type

  business/sdk/tablebuilder/model.go          (new Type constant)
  business/sdk/tablebuilder/store.go          (dispatch on Type in FetchTableData)

## ⚠ Adding a new visualization type

  business/sdk/tablebuilder/model.go          (new WidgetType/Visualization constant + supporting struct if needed)
  business/sdk/tablebuilder/store.go          (handle data shape for new vis in FetchTableData / QueryByPage)
  Frontend page builder                        (render new visualization type)

## ⚠ Adding a column to config.table_configs or config.page_configs

  business/sdk/migrate/sql/migrate.sql        (new version, ALTER TABLE)
  business/sdk/tablebuilder/configstore.go    (StoredConfig/PageConfig model + affected queries)
  business/sdk/tablebuilder/model.go          (StoredConfig/PageConfig struct if field exposed)
