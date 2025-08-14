# Table Builder Package - Developer Guide

## Table of Contents

- [Overview](#overview)
- [Installation](#installation)
- [Core Concepts](#core-concepts)
- [Basic Usage](#basic-usage)
- [Configuration Structure](#configuration-structure)
- [Data Sources](#data-sources)
- [Filtering and Sorting](#filtering-and-sorting)
- [Computed Columns](#computed-columns)
- [Foreign Tables and Joins](#foreign-tables-and-joins)
- [Pagination](#pagination)
- [Visual Settings](#visual-settings)
- [Configuration Storage](#configuration-storage)
- [Error Handling](#error-handling)
- [Performance Considerations](#performance-considerations)
- [Migration Guide](#migration-guide)
- [API Reference](#api-reference)

## Overview

The Table Builder package (`tablebuilder`) is a dynamic SQL query builder for PostgreSQL that enables runtime configuration of complex table queries with support for joins, filters, computed columns, and nested data structures. It's designed to work seamlessly with the Ardan Labs architecture pattern and integrates with existing `sqldb`, `page`, and `logger` packages.

### Key Features

- **Dynamic Query Building**: Build complex SQL queries from JSON configurations
- **Type Safety**: Uses goqu for SQL injection prevention
- **Computed Columns**: Runtime expression evaluation with govaluate
- **Foreign Relationships**: Support for nested joins and related data
- **Configuration Storage**: Store and retrieve configurations as JSONB
- **Pagination Support**: Integrates with existing page.Page type
- **View Support**: Works with PostgreSQL views and functions

## Installation

### Prerequisites

```bash
# Required dependencies
go get github.com/doug-martin/goqu/v9
go get github.com/Knetic/govaluate
go get github.com/jmoiron/sqlx
go get github.com/jackc/pgx/v5
```

### Database Setup

Run the migration to create the configuration storage table:

```sql
-- Run migration from tablebuilder-migration.sql
CREATE TABLE IF NOT EXISTS table_configs (
    id UUID PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    description TEXT,
    config JSONB NOT NULL,
    created_by UUID NOT NULL,
    updated_by UUID NOT NULL,
    created_date TIMESTAMP NOT NULL,
    updated_date TIMESTAMP NOT NULL
);
```

## Core Concepts

### Store Types

The package provides two main store types:

1. **`Store`** - Executes table queries based on configurations
2. **`ConfigStore`** - Manages saving/loading configurations from database

### Configuration Flow

```
JSON Config → Validation → Query Building → Execution → Transformation → Result
```

### Data Source Types

- **`query`** - Direct table queries
- **`view`** - PostgreSQL view queries
- **`viewcount`** - Count queries for related data
- **`rpc`** - Stored procedure/function calls

## Basic Usage

### Simple Query Example

```go
package main

import (
    "context"
    "github.com/timmaaaz/ichor/business/sdk/tablebuilder"
    "github.com/timmaaaz/ichor/foundation/logger"
)

func main() {
    // Initialize dependencies
    log := logger.New("app", logger.LevelInfo)
    db, _ := sqldb.Open(cfg)

    // Create store
    store := tablebuilder.NewStore(log, db)

    // Define configuration
    config := &tablebuilder.Config{
        Title:         "Product List",
        WidgetType:    "table",
        Visualization: "table",
        DataSource: []tablebuilder.DataSource{
            {
                Type:   "query",
                Source: "products",
                Select: tablebuilder.SelectConfig{
                    Columns: []tablebuilder.ColumnDefinition{
                        {Name: "id", Path: "id"},
                        {Name: "name", Path: "name"},
                        {Name: "sku", Path: "sku"},
                        {Name: "price", Path: "price"},
                    },
                },
                Filters: []tablebuilder.Filter{
                    {
                        Column:   "is_active",
                        Operator: "eq",
                        Value:    true,
                    },
                },
                Sort: []tablebuilder.Sort{
                    {
                        Column:    "name",
                        Direction: "asc",
                    },
                },
                Limit: 50,
            },
        },
        VisualSettings: tablebuilder.VisualSettings{
            // Visual configuration
        },
        Permissions: tablebuilder.Permissions{
            Roles:   []string{"admin", "user"},
            Actions: []string{"view"},
        },
    }

    // Execute query
    params := tablebuilder.QueryParams{
        Page:  1,
        Limit: 25,
    }

    result, err := store.FetchTableData(context.Background(), config, params)
    if err != nil {
        log.Error("Failed to fetch data", "error", err)
        return
    }

    // Process results
    for _, row := range result.Data {
        // Process each row
    }
}
```

## Configuration Structure

### Complete Configuration Example

```go
type Config struct {
    Title            string          `json:"title"`
    WidgetType       string          `json:"widget_type"`       // "table", "chart", etc.
    Visualization    string          `json:"visualization"`     // "table", "grid", etc.
    PositionX        int             `json:"position_x"`
    PositionY        int             `json:"position_y"`
    Width            int             `json:"width"`
    Height           int             `json:"height"`
    DataSource       []DataSource    `json:"data_source"`
    RefreshInterval  int             `json:"refresh_interval"`  // seconds
    RefreshMode      string          `json:"refresh_mode"`      // "polling", "manual"
    VisualSettings   VisualSettings  `json:"visual_settings"`
    Permissions      Permissions     `json:"permissions"`
}
```

### JSON Configuration Example

```json
{
  "title": "Inventory Dashboard",
  "widget_type": "table",
  "visualization": "table",
  "position_x": 0,
  "position_y": 0,
  "width": 12,
  "height": 6,
  "data_source": [
    {
      "type": "query",
      "source": "inventory_items",
      "select": {
        "columns": [
          { "name": "id", "path": "id" },
          { "name": "quantity", "path": "quantity", "alias": "current_stock" }
        ]
      },
      "filters": [{ "column": "quantity", "operator": "gt", "value": 0 }],
      "limit": 100
    }
  ],
  "refresh_interval": 300,
  "refresh_mode": "polling",
  "visual_settings": {},
  "permissions": {
    "roles": ["admin"],
    "actions": ["view", "export"]
  }
}
```

## Data Sources

### Query Type

Standard table queries:

```go
DataSource{
    Type:   "query",
    Source: "products",  // Table name
    Select: SelectConfig{
        Columns: []ColumnDefinition{
            {Name: "id", Path: "id"},
            {Name: "name", Path: "name"},
        },
    },
}
```

### View Type

Query PostgreSQL views:

```go
DataSource{
    Type:   "view",
    Source: "orders_base",  // View name
    Select: SelectConfig{
        Columns: []ColumnDefinition{
            {Name: "order_id", Path: "order_id"},
            {Name: "customer_name", Path: "customer_name"},
        },
    },
}
```

### ViewCount Type

Count related records:

```go
DataSource{
    Type:         "viewcount",
    Source:       "order_line_items",
    ParentSource: "orders.order_id",
    SelectBy:     "order_id",
    Select: SelectConfig{
        Columns: []ColumnDefinition{
            {Name: "id", Path: "id"},
        },
    },
}
```

### RPC Type

Call stored procedures/functions:

```go
DataSource{
    Type:   "rpc",
    Source: "get_inventory_summary",
    Args: map[string]any{
        "warehouse_id": "uuid-here",
    },
}
```

## Filtering and Sorting

### Filter Operators

- `eq` - Equal
- `neq` - Not equal
- `gt` - Greater than
- `gte` - Greater than or equal
- `lt` - Less than
- `lte` - Less than or equal
- `in` - In array
- `like` - SQL LIKE
- `ilike` - Case-insensitive LIKE
- `is_null` - IS NULL
- `is_not_null` - IS NOT NULL

### Static Filters

```go
Filters: []tablebuilder.Filter{
    {
        Column:   "status",
        Operator: "eq",
        Value:    "active",
    },
    {
        Column:   "price",
        Operator: "gte",
        Value:    100.00,
    },
}
```

### Dynamic Filters

```go
// In configuration
Filters: []tablebuilder.Filter{
    {
        Column:   "category_id",
        Operator: "eq",
        Dynamic:  true,  // Value comes from params
    },
}

// At runtime
params := tablebuilder.QueryParams{
    Dynamic: map[string]any{
        "category_id": "uuid-here",
    },
}
```

### Sorting

```go
Sort: []tablebuilder.Sort{
    {
        Column:    "created_date",
        Direction: "desc",
        Priority:  1,
    },
    {
        Column:    "name",
        Direction: "asc",
        Priority:  2,
    },
}
```

## Computed Columns

### Basic Computed Column

```go
ClientComputedColumns: []tablebuilder.ComputedColumn{
    {
        Name:       "total_value",
        Expression: "quantity * price",
    },
    {
        Name:       "status",
        Expression: "quantity > reorder_point ? 'ok' : 'low'",
    },
}
```

### Complex Expressions

```go
ClientComputedColumns: []tablebuilder.ComputedColumn{
    {
        Name:       "margin",
        Expression: "(selling_price - cost) / selling_price * 100",
    },
    {
        Name:       "days_until_reorder",
        Expression: "quantity / avg_daily_usage",
    },
}
```

### Conditional Logic

```go
ClientComputedColumns: []tablebuilder.ComputedColumn{
    {
        Name: "risk_level",
        Expression: `
            quantity <= 0 ? 'out_of_stock' :
            quantity <= safety_stock ? 'critical' :
            quantity <= reorder_point ? 'low' :
            'normal'
        `,
    },
}
```

## Foreign Tables and Joins

### Simple Join

```go
ForeignTables: []tablebuilder.ForeignTable{
    {
        Table:        "categories",
        Relationship: "category_id",
        JoinType:     "left",
        Columns: []tablebuilder.ColumnDefinition{
            {Name: "name", Path: "categories.name", Alias: "category_name"},
        },
    },
}
```

### Nested Joins

```go
ForeignTables: []tablebuilder.ForeignTable{
    {
        Table:        "products",
        Relationship: "product_id",
        Columns: []tablebuilder.ColumnDefinition{
            {Name: "name", Path: "products.name"},
        },
        ForeignTables: []tablebuilder.ForeignTable{
            {
                Table:        "brands",
                Relationship: "brand_id",
                Columns: []tablebuilder.ColumnDefinition{
                    {Name: "name", Path: "products.brands.name", Alias: "brand_name"},
                },
            },
        },
    },
}
```

## Pagination

### Using with page.Page

```go
import "github.com/timmaaaz/ichor/business/sdk/page"

// Create page
pg := page.NewPage(1, 25)

// Query with pagination
result, err := store.QueryByPage(ctx, config, pg)

// Access pagination metadata
fmt.Printf("Page %d of %d\n", result.Meta.Page, result.Meta.TotalPages)
fmt.Printf("Total records: %d\n", result.Meta.Total)
```

### Manual Pagination

```go
params := tablebuilder.QueryParams{
    Page:  2,
    Limit: 50,
}

result, err := store.FetchTableData(ctx, config, params)
```

## Visual Settings

### Column Configuration

```go
VisualSettings: tablebuilder.VisualSettings{
    Columns: map[string]tablebuilder.ColumnConfig{
        "product_name": {
            Name:       "product_name",
            Header:     "Product",
            Width:      250,
            Align:      "left",
            Sortable:   true,
            Filterable: true,
            Format: &tablebuilder.FormatConfig{
                Type: "text",
            },
        },
        "price": {
            Name:   "price",
            Header: "Price",
            Width:  120,
            Align:  "right",
            Format: &tablebuilder.FormatConfig{
                Type:      "currency",
                Currency:  "USD",
                Precision: 2,
            },
        },
        "quantity": {
            Name:   "quantity",
            Header: "Stock",
            Width:  100,
            Align:  "right",
            Format: &tablebuilder.FormatConfig{
                Type:      "number",
                Precision: 0,
            },
        },
    },
}
```

### Conditional Formatting

```go
ConditionalFormatting: []tablebuilder.ConditionalFormat{
    {
        Column:     "status",
        Condition:  "eq",
        Value:      "critical",
        Color:      "#ff0000",
        Background: "#ffeeee",
        Icon:       "alert",
    },
    {
        Column:     "quantity",
        Condition:  "lt",
        Value:      10,
        Color:      "#ff9900",
        Background: "#fff5ee",
    },
}
```

### Actions

```go
RowActions: map[string]tablebuilder.Action{
    "edit": {
        Name:       "edit",
        Label:      "Edit",
        Icon:       "edit",
        ActionType: "modal",
        Component:  "EditForm",
        Params: map[string]any{
            "id": "{id}",
        },
    },
    "view": {
        Name:       "view",
        Label:      "View Details",
        Icon:       "eye",
        ActionType: "link",
        URL:        "/items/{id}",
    },
},
TableActions: map[string]tablebuilder.Action{
    "export": {
        Name:       "export",
        Label:      "Export CSV",
        Icon:       "download",
        ActionType: "export",
        Format:     "csv",
    },
}
```

## Configuration Storage

### Saving Configurations

```go
configStore := tablebuilder.NewConfigStore(log, db)

// Create and save configuration
stored, err := configStore.Create(
    ctx,
    "inventory_dashboard",           // name
    "Main inventory dashboard view", // description
    config,                          // configuration
    userID,                          // created by
)

if err != nil {
    return fmt.Errorf("save config: %w", err)
}
```

### Loading Configurations

```go
// Load by ID
config, err := configStore.LoadConfig(ctx, configID)

// Load by name
config, err := configStore.LoadConfigByName(ctx, "inventory_dashboard")

// Get all configs for a user
configs, err := configStore.QueryByUser(ctx, userID)
```

### Updating Configurations

```go
stored, err := configStore.Update(
    ctx,
    configID,
    "inventory_dashboard_v2",
    "Updated dashboard with new features",
    updatedConfig,
    userID,
)
```

### Deleting Configurations

```go
err := configStore.Delete(ctx, configID)
```

## Error Handling

### Package Errors

```go
var (
    // Configuration errors
    ErrInvalidConfig     = errors.New("invalid table configuration")
    ErrNoDataSource      = errors.New("no data source specified")
    ErrInvalidDataSource = errors.New("invalid data source configuration")

    // Column errors
    ErrColumnNotFound    = errors.New("column not found")
    ErrInvalidColumn     = errors.New("invalid column configuration")

    // Query errors
    ErrInvalidQuery      = errors.New("invalid query")
    ErrQueryFailed       = errors.New("query execution failed")

    // Database errors
    ErrNotFound          = errors.New("record not found")
    ErrDatabaseError     = errors.New("database error")
)
```

### Error Handling Pattern

```go
result, err := store.FetchTableData(ctx, config, params)
if err != nil {
    switch {
    case errors.Is(err, tablebuilder.ErrInvalidConfig):
        // Handle invalid configuration
        return fmt.Errorf("configuration error: %w", err)

    case errors.Is(err, tablebuilder.ErrNotFound):
        // Handle not found
        return fmt.Errorf("data not found: %w", err)

    default:
        // Handle other errors
        return fmt.Errorf("fetch table data: %w", err)
    }
}
```

## Performance Considerations

### Query Optimization

1. **Use Views**: Pre-defined views are faster than complex joins
2. **Limit Results**: Always use pagination for large datasets
3. **Index Columns**: Ensure filtered and sorted columns are indexed
4. **Computed Columns**: Complex computations should be done in the database when possible

### Caching Strategy

```go
// Example caching wrapper (not included in package)
type CachedStore struct {
    store *tablebuilder.Store
    cache map[string]*tablebuilder.TableData
    ttl   time.Duration
}

func (cs *CachedStore) FetchTableData(ctx context.Context, config *tablebuilder.Config, params tablebuilder.QueryParams) (*tablebuilder.TableData, error) {
    key := generateCacheKey(config, params)

    if cached, ok := cs.cache[key]; ok {
        return cached, nil
    }

    result, err := cs.store.FetchTableData(ctx, config, params)
    if err != nil {
        return nil, err
    }

    cs.cache[key] = result
    // Set TTL...

    return result, nil
}
```

### Batch Operations

For multiple related queries, use multiple data sources:

```go
DataSource: []tablebuilder.DataSource{
    {
        Type:   "query",
        Source: "orders",
        // Primary query
    },
    {
        Type:         "viewcount",
        Source:       "order_line_items",
        ParentSource: "orders.id",
        // Related count
    },
}
```

## Migration Guide

### From TypeScript to Go

#### TypeScript Configuration

```typescript
const config = {
  data_source: [
    {
      type: "query",
      source: "products",
      select: {
        columns: [{ name: "id", path: "id" }],
      },
    },
  ],
};
```

#### Go Configuration

```go
config := &tablebuilder.Config{
    DataSource: []tablebuilder.DataSource{
        {
            Type:   "query",
            Source: "products",
            Select: tablebuilder.SelectConfig{
                Columns: []tablebuilder.ColumnDefinition{
                    {Name: "id", Path: "id"},
                },
            },
        },
    },
}
```

### Key Differences

1. **No alias mapping complexity** - goqu handles this automatically
2. **Type safety** - Compile-time type checking vs runtime
3. **Performance** - ~10x faster execution than TypeScript
4. **Expression syntax** - Uses govaluate syntax instead of JavaScript

## API Reference

### Store Methods

#### FetchTableData

```go
func (s *Store) FetchTableData(ctx context.Context, config *Config, params QueryParams) (*TableData, error)
```

Executes the table configuration and returns the data.

**Parameters:**

- `ctx`: Context for cancellation and tracing
- `config`: Table configuration
- `params`: Runtime query parameters

**Returns:**

- `*TableData`: Query results with metadata
- `error`: Error if query fails

#### QueryByPage

```go
func (s *Store) QueryByPage(ctx context.Context, config *Config, pg page.Page) (*TableData, error)
```

Queries table data with page.Page support.

### ConfigStore Methods

#### Create

```go
func (s *ConfigStore) Create(ctx context.Context, name, description string, config *Config, userID uuid.UUID) (*StoredConfig, error)
```

Saves a new table configuration.

#### LoadConfig

```go
func (s *ConfigStore) LoadConfig(ctx context.Context, id uuid.UUID) (*Config, error)
```

Loads a configuration by ID.

#### LoadConfigByName

```go
func (s *ConfigStore) LoadConfigByName(ctx context.Context, name string) (*Config, error)
```

Loads a configuration by name.

### QueryBuilder Methods

#### BuildQuery

```go
func (qb *QueryBuilder) BuildQuery(ds *DataSource, params QueryParams, isPrimary bool) (string, map[string]interface{}, error)
```

Builds a SQL query from a data source configuration.

#### BuildCountQuery

```go
func (qb *QueryBuilder) BuildCountQuery(ds *DataSource, params QueryParams) (string, map[string]interface{}, error)
```

Builds a count query for pagination.

### Evaluator Methods

#### Evaluate

```go
func (e *Evaluator) Evaluate(expression string, row TableRow) (interface{}, error)
```

Evaluates an expression with the given row context.

#### CompileExpressions

```go
func (e *Evaluator) CompileExpressions(expressions []ComputedColumn) error
```

Pre-compiles a list of expressions for validation.

## Best Practices

### 1. Configuration Validation

Always validate configurations before use:

```go
if err := config.Validate(); err != nil {
    return fmt.Errorf("invalid config: %w", err)
}
```

### 2. Use Views for Complex Queries

Create PostgreSQL views for commonly used complex queries:

```sql
CREATE VIEW orders_summary AS
SELECT o.*, c.name as customer_name
FROM orders o
JOIN customers c ON o.customer_id = c.id;
```

### 3. Limit Result Sets

Always use pagination for potentially large result sets:

```go
config.DataSource[0].Limit = 100  // Maximum rows
```

### 4. Handle Errors Gracefully

Use proper error wrapping and logging:

```go
if err != nil {
    log.Error("query failed",
        "config", config.Title,
        "error", err,
    )
    return fmt.Errorf("fetch %s: %w", config.Title, err)
}
```

### 5. Secure Dynamic Filters

Validate dynamic filter values:

```go
if userInput != "" {
    params.Dynamic["user_id"] = validateUUID(userInput)
}
```

### 6. Optimize Computed Columns

Keep expressions simple and consider database computed columns for complex logic:

```sql
ALTER TABLE inventory_items
ADD COLUMN stock_value GENERATED ALWAYS AS (quantity * unit_price) STORED;
```

## Troubleshooting

### Common Issues

#### "invalid table configuration"

- Check that `Title` is not empty
- Ensure at least one `DataSource` is defined
- Verify `Source` is specified in each data source

#### "column not found"

- Verify column exists in the table/view
- Check column name spelling and case
- Ensure proper table prefix for joined columns

#### "expression evaluation failed"

- Validate expression syntax
- Ensure all referenced fields exist in the row
- Check for division by zero or null values

#### "query execution failed"

- Check database connection
- Verify table/view exists
- Ensure user has proper permissions
- Check for SQL syntax errors in filters

### Debug Mode

Enable debug logging to see generated SQL:

```go
log.SetLevel(logger.LevelDebug)
// Queries will be logged with generated SQL
```

## Examples Repository

For complete working examples, see:

- `examples/simple_query.go` - Basic table query
- `examples/complex_joins.go` - Multi-table joins
- `examples/computed_columns.go` - Expression evaluation
- `examples/stored_configs.go` - Configuration management
- `examples/pagination.go` - Pagination examples
- `examples/dynamic_filters.go` - Runtime filtering

## Support

For issues, questions, or contributions:

1. Check this documentation
2. Review error messages and logs
3. Consult the API reference
4. Check existing issues in the repository
5. Create a new issue with:
   - Configuration causing the issue
   - Error messages
   - Expected vs actual behavior
   - Database schema if relevant
