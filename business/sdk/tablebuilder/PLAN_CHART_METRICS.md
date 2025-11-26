# Chart Metrics & Time Grouping Implementation Plan

## Overview

Add structured, safe aggregation and time-grouping support to the tablebuilder for chart queries. This replaces the unsafe raw SQL `TableColumn` approach with a validated, whitelist-based system.

## Problem Statement

Current chart seed data uses `TableColumn` with raw SQL expressions:
```go
{Name: "total_revenue", Alias: "total_revenue", TableColumn: "SUM(order_line_items.quantity * product_costs.selling_price)"}
```

Issues:
1. `TableColumn` is **completely ignored** by `buildSelectColumns()` in builder.go
2. Raw SQL injection risk if configs are created via API
3. No support for `GROUP BY` with time intervals (daily, monthly, etc.)

## Solution

Add structured `Metrics` and `GroupBy` fields to `DataSource` with whitelist validation.

---

## Phase 1: Model Changes (`model.go`)

### Add New Types

```go
// MetricConfig defines an aggregated value for charts
type MetricConfig struct {
    Name       string            `json:"name"`                  // Output alias: "total_revenue"
    Function   string            `json:"function"`              // sum, count, avg, min, max, count_distinct
    Column     string            `json:"column,omitempty"`      // Simple column: "quantity"
    Expression *ExpressionConfig `json:"expression,omitempty"`  // For multi-column math
}

// ExpressionConfig for multi-column arithmetic
type ExpressionConfig struct {
    Operator string   `json:"operator"` // multiply, add, subtract, divide
    Columns  []string `json:"columns"`  // ["order_line_items.quantity", "product_costs.selling_price"]
}

// GroupByConfig for time-series and categorical grouping
type GroupByConfig struct {
    Column   string `json:"column"`              // "orders.created_date" or "categories.name"
    Interval string `json:"interval,omitempty"`  // day, week, month, quarter, year (dates only)
    Alias    string `json:"alias,omitempty"`     // Output name: "month"
}
```

### Modify DataSource

```go
type DataSource struct {
    // ... existing fields ...

    // Chart aggregation (safe, structured)
    Metrics []MetricConfig `json:"metrics,omitempty"`
    GroupBy *GroupByConfig `json:"group_by,omitempty"`
}
```

### Add Validation Constants

```go
// Allowed aggregate functions (whitelist)
var AllowedAggregateFunctions = map[string]string{
    "sum":            "SUM",
    "count":          "COUNT",
    "count_distinct": "COUNT(DISTINCT %s)",
    "avg":            "AVG",
    "min":            "MIN",
    "max":            "MAX",
}

// Allowed expression operators
var AllowedOperators = map[string]string{
    "multiply": "*",
    "add":      "+",
    "subtract": "-",
    "divide":   "/",
}

// Allowed time intervals
var AllowedIntervals = map[string]string{
    "day":     "day",
    "week":    "week",
    "month":   "month",
    "quarter": "quarter",
    "year":    "year",
}
```

---

## Phase 2: Query Builder Changes (`builder.go`)

### Add Metric Query Builder

```go
// BuildQuery - modify to detect metric queries
func (qb *QueryBuilder) BuildQuery(ds *DataSource, params QueryParams, isPrimary bool) (string, map[string]interface{}, error) {
    // Check if this is a metric/chart query
    if len(ds.Metrics) > 0 {
        return qb.buildMetricQuery(ds, params, isPrimary)
    }

    // ... existing table query logic ...
}

// buildMetricQuery builds aggregation queries for charts
func (qb *QueryBuilder) buildMetricQuery(ds *DataSource, params QueryParams, isPrimary bool) (string, map[string]interface{}, error) {
    // 1. Validate all metrics
    // 2. Build SELECT with aggregates
    // 3. Build GROUP BY if present
    // 4. Apply joins from ForeignTables
    // 5. Apply filters and sorting
}

// buildMetricExpression safely builds a metric SQL expression
func (qb *QueryBuilder) buildMetricExpression(metric MetricConfig) (string, error) {
    // Validate function is in whitelist
    // Build expression or column reference
    // Return safe SQL fragment
}

// buildGroupByExpression safely builds GROUP BY clause
func (qb *QueryBuilder) buildGroupByExpression(groupBy *GroupByConfig) (string, string, error) {
    // Returns: (select expression, group by expression, error)
    // For time intervals: DATE_TRUNC('month', orders.created_date) AS month
}
```

### SQL Generation Examples

**KPI (no grouping):**
```sql
SELECT SUM(order_line_items.quantity * product_costs.selling_price) AS total_revenue
FROM sales.order_line_items
INNER JOIN products.products ON order_line_items.product_id = products.id
INNER JOIN products.product_costs ON products.id = product_costs.product_id
```

**Monthly time-series:**
```sql
SELECT
    DATE_TRUNC('month', orders.created_date) AS month,
    SUM(order_line_items.quantity * product_costs.selling_price) AS sales
FROM sales.order_line_items
INNER JOIN sales.orders ON order_line_items.order_id = orders.id
INNER JOIN products.products ON order_line_items.product_id = products.id
INNER JOIN products.product_costs ON products.id = product_costs.product_id
GROUP BY DATE_TRUNC('month', orders.created_date)
ORDER BY month ASC
```

**Count with categorical grouping:**
```sql
SELECT
    order_fulfillment_statuses.name AS status,
    COUNT(orders.id) AS count
FROM sales.orders
INNER JOIN sales.order_fulfillment_statuses ON orders.fulfillment_status_id = order_fulfillment_statuses.id
GROUP BY order_fulfillment_statuses.name
```

---

## Phase 3: Seed Data Updates (`seedmodels/charts.go`)

### Convert Existing Seeds

**Before:**
```go
var SeedKPITotalRevenue = &tablebuilder.Config{
    DataSource: []tablebuilder.DataSource{{
        Source: "order_line_items",
        Schema: "sales",
        Select: tablebuilder.SelectConfig{
            Columns: []tablebuilder.ColumnDefinition{
                {Name: "total_revenue", Alias: "total_revenue",
                 TableColumn: "SUM(order_line_items.quantity * product_costs.selling_price)"},
            },
            ForeignTables: []tablebuilder.ForeignTable{...},
        },
    }},
}
```

**After:**
```go
var SeedKPITotalRevenue = &tablebuilder.Config{
    Title:         "Total Revenue",
    WidgetType:    "chart",
    Visualization: "kpi",
    DataSource: []tablebuilder.DataSource{{
        Source: "order_line_items",
        Schema: "sales",
        Metrics: []tablebuilder.MetricConfig{
            {
                Name:     "total_revenue",
                Function: "sum",
                Expression: &tablebuilder.ExpressionConfig{
                    Operator: "multiply",
                    Columns:  []string{"order_line_items.quantity", "product_costs.selling_price"},
                },
            },
        },
        Select: tablebuilder.SelectConfig{
            ForeignTables: []tablebuilder.ForeignTable{
                // Same joins as before
            },
        },
    }},
    VisualSettings: tablebuilder.VisualSettings{...},
}
```

**Time-series example:**
```go
var SeedLineMonthlySales = &tablebuilder.Config{
    Title:         "Monthly Sales Trend",
    WidgetType:    "chart",
    Visualization: "line",
    DataSource: []tablebuilder.DataSource{{
        Source: "order_line_items",
        Schema: "sales",
        Metrics: []tablebuilder.MetricConfig{
            {
                Name:     "sales",
                Function: "sum",
                Expression: &tablebuilder.ExpressionConfig{
                    Operator: "multiply",
                    Columns:  []string{"order_line_items.quantity", "product_costs.selling_price"},
                },
            },
        },
        GroupBy: &tablebuilder.GroupByConfig{
            Column:   "orders.created_date",
            Interval: "month",
            Alias:    "month",
        },
        Select: tablebuilder.SelectConfig{
            ForeignTables: []tablebuilder.ForeignTable{
                // orders join, product_costs join
            },
        },
        Sort: []tablebuilder.Sort{
            {Column: "month", Direction: "asc"},
        },
    }},
}
```

### Charts to Convert

1. `SeedKPITotalRevenue` - KPI with SUM expression
2. `SeedKPIOrderCount` - KPI with COUNT
3. `SeedGaugeRevenueTarget` - Gauge with SUM expression
4. `SeedLineMonthlySales` - Line with monthly grouping
5. `SeedBarTopProducts` - Bar with product grouping
6. `SeedStackedBarRegionCategory` - Stacked bar with region grouping
7. `SeedStackedAreaCumulative` - Stacked area with monthly grouping
8. `SeedPieRevenueCategory` - Pie with category grouping
9. `SeedComboRevenueOrders` - Combo with multiple metrics + monthly grouping
10. `SeedWaterfallProfit` - Waterfall (static categories)
11. `SeedFunnelPipeline` - Funnel with status grouping
12. `SeedHeatmapSalesTime` - Heatmap with day/hour grouping
13. `SeedTreemapRevenue` - Treemap with product grouping
14. `SeedGanttProject` - Gantt (no aggregation, uses columns)

---

## Phase 4: Unit Tests (`tablebuilder_test.go`)

### Test Categories

#### 4.1 Model Serialization Tests

```go
func Test_MetricConfigSerialization(t *testing.T) {
    t.Run("simple metric with column", func(t *testing.T) {
        metric := tablebuilder.MetricConfig{
            Name:     "order_count",
            Function: "count",
            Column:   "orders.id",
        }
        // Test JSON round-trip
    })

    t.Run("metric with expression", func(t *testing.T) {
        metric := tablebuilder.MetricConfig{
            Name:     "total_revenue",
            Function: "sum",
            Expression: &tablebuilder.ExpressionConfig{
                Operator: "multiply",
                Columns:  []string{"quantity", "price"},
            },
        }
        // Test JSON round-trip
    })

    t.Run("group by with interval", func(t *testing.T) {
        groupBy := tablebuilder.GroupByConfig{
            Column:   "created_date",
            Interval: "month",
            Alias:    "month",
        }
        // Test JSON round-trip
    })
}
```

#### 4.2 Query Builder Unit Tests

```go
func Test_MetricQueryBuilder(t *testing.T) {
    t.Run("builds simple count query", func(t *testing.T) {
        ds := &tablebuilder.DataSource{
            Source: "orders",
            Schema: "sales",
            Metrics: []tablebuilder.MetricConfig{
                {Name: "count", Function: "count", Column: "orders.id"},
            },
        }
        builder := tablebuilder.NewQueryBuilder()
        sql, _, err := builder.BuildQuery(ds, tablebuilder.QueryParams{}, true)

        // Assert SQL contains "SELECT COUNT(orders.id) AS count"
        // Assert SQL contains "FROM sales.orders"
    })

    t.Run("builds sum with expression", func(t *testing.T) {
        ds := &tablebuilder.DataSource{
            Source: "order_line_items",
            Schema: "sales",
            Metrics: []tablebuilder.MetricConfig{
                {
                    Name:     "revenue",
                    Function: "sum",
                    Expression: &tablebuilder.ExpressionConfig{
                        Operator: "multiply",
                        Columns:  []string{"order_line_items.quantity", "product_costs.selling_price"},
                    },
                },
            },
            Select: tablebuilder.SelectConfig{
                ForeignTables: []tablebuilder.ForeignTable{
                    {Table: "products", Schema: "products", ...},
                },
            },
        }
        // Assert SQL contains "SUM(order_line_items.quantity * product_costs.selling_price)"
    })

    t.Run("builds query with group by interval", func(t *testing.T) {
        ds := &tablebuilder.DataSource{
            Source: "order_line_items",
            Schema: "sales",
            Metrics: []tablebuilder.MetricConfig{
                {Name: "sales", Function: "sum", Column: "amount"},
            },
            GroupBy: &tablebuilder.GroupByConfig{
                Column:   "orders.created_date",
                Interval: "month",
                Alias:    "month",
            },
        }
        // Assert SQL contains "DATE_TRUNC('month', orders.created_date) AS month"
        // Assert SQL contains "GROUP BY DATE_TRUNC('month', orders.created_date)"
    })

    t.Run("builds query with categorical group by", func(t *testing.T) {
        ds := &tablebuilder.DataSource{
            Source: "orders",
            Schema: "sales",
            Metrics: []tablebuilder.MetricConfig{
                {Name: "count", Function: "count", Column: "orders.id"},
            },
            GroupBy: &tablebuilder.GroupByConfig{
                Column: "order_fulfillment_statuses.name",
                Alias:  "status",
            },
        }
        // Assert SQL contains "GROUP BY order_fulfillment_statuses.name"
    })

    t.Run("rejects invalid function", func(t *testing.T) {
        ds := &tablebuilder.DataSource{
            Source: "orders",
            Schema: "sales",
            Metrics: []tablebuilder.MetricConfig{
                {Name: "bad", Function: "DROP TABLE", Column: "orders.id"},
            },
        }
        builder := tablebuilder.NewQueryBuilder()
        _, _, err := builder.BuildQuery(ds, tablebuilder.QueryParams{}, true)
        // Assert error is returned
    })

    t.Run("rejects invalid operator", func(t *testing.T) {
        ds := &tablebuilder.DataSource{
            Source: "orders",
            Schema: "sales",
            Metrics: []tablebuilder.MetricConfig{
                {
                    Name:     "bad",
                    Function: "sum",
                    Expression: &tablebuilder.ExpressionConfig{
                        Operator: "; DROP TABLE",
                        Columns:  []string{"a", "b"},
                    },
                },
            },
        }
        // Assert error is returned
    })

    t.Run("rejects invalid interval", func(t *testing.T) {
        ds := &tablebuilder.DataSource{
            Source: "orders",
            Schema: "sales",
            Metrics: []tablebuilder.MetricConfig{
                {Name: "count", Function: "count", Column: "id"},
            },
            GroupBy: &tablebuilder.GroupByConfig{
                Column:   "created_date",
                Interval: "DROP TABLE",
            },
        }
        // Assert error is returned
    })
}
```

#### 4.3 Integration Tests with Database

```go
func Test_MetricQueries(t *testing.T) {
    t.Parallel()

    db := dbtest.NewDatabase(t, "Test_MetricQueries")
    log := logger.New(io.Discard, logger.LevelInfo, "ADMIN", func(context.Context) string { return "00000000-0000-0000-0000-000000000000" })

    store := tablebuilder.NewStore(log, db.DB)

    // Seed test data including product_costs
    sd, err := insertMetricTestData(db.BusDomain)
    if err != nil {
        t.Fatalf("failed to insert seed data: %v", err)
    }

    t.Run("KPI total revenue", func(t *testing.T) {
        kpiTotalRevenueTest(context.Background(), t, store, sd)
    })

    t.Run("count orders", func(t *testing.T) {
        countOrdersTest(context.Background(), t, store, sd)
    })

    t.Run("monthly sales aggregation", func(t *testing.T) {
        monthlySalesTest(context.Background(), t, store, sd)
    })

    t.Run("revenue by category", func(t *testing.T) {
        revenueByCategoryTest(context.Background(), t, store, sd)
    })

    t.Run("multiple metrics", func(t *testing.T) {
        multipleMetricsTest(context.Background(), t, store, sd)
    })
}
```

#### 4.4 Test Helper: Seed Data with ProductCosts

```go
func insertMetricTestData(busDomain dbtest.BusDomain) (unitest.SeedData, error) {
    ctx := context.Background()

    // ... existing seed logic for users, products, orders, order_line_items ...

    // ADD: Seed product costs (required for revenue calculations)
    productCosts, err := productcostbus.TestSeedProductCosts(ctx, len(products), productIDs, busDomain.ProductCost)
    if err != nil {
        return unitest.SeedData{}, fmt.Errorf("seeding product costs: %w", err)
    }

    return unitest.SeedData{
        // ... existing fields ...
        ProductCosts: productCosts,
    }, nil
}
```

#### 4.5 Individual Test Functions

```go
func kpiTotalRevenueTest(ctx context.Context, t *testing.T, store *tablebuilder.Store, sd unitest.SeedData) {
    config := &tablebuilder.Config{
        Title:         "Total Revenue",
        WidgetType:    "chart",
        Visualization: "kpi",
        DataSource: []tablebuilder.DataSource{{
            Source: "order_line_items",
            Schema: "sales",
            Metrics: []tablebuilder.MetricConfig{
                {
                    Name:     "total_revenue",
                    Function: "sum",
                    Expression: &tablebuilder.ExpressionConfig{
                        Operator: "multiply",
                        Columns:  []string{"order_line_items.quantity", "product_costs.selling_price"},
                    },
                },
            },
            Select: tablebuilder.SelectConfig{
                ForeignTables: []tablebuilder.ForeignTable{
                    {
                        Table:            "products",
                        Schema:           "products",
                        RelationshipFrom: "order_line_items.product_id",
                        RelationshipTo:   "products.id",
                        JoinType:         "inner",
                        ForeignTables: []tablebuilder.ForeignTable{
                            {
                                Table:            "product_costs",
                                Schema:           "products",
                                RelationshipFrom: "products.id",
                                RelationshipTo:   "product_costs.product_id",
                                JoinType:         "inner",
                            },
                        },
                    },
                },
            },
        }},
    }

    params := tablebuilder.QueryParams{}
    result, err := store.FetchTableData(ctx, config, params)
    if err != nil {
        t.Fatalf("fetch failed: %v", err)
    }

    // Should return exactly 1 row with total_revenue
    if len(result.Data) != 1 {
        t.Errorf("expected 1 row, got %d", len(result.Data))
    }

    if _, ok := result.Data[0]["total_revenue"]; !ok {
        t.Error("expected total_revenue column in result")
    }

    // Value should be > 0 (we have seed data)
    revenue, ok := result.Data[0]["total_revenue"].(float64)
    if !ok {
        t.Error("total_revenue should be a number")
    }
    if revenue <= 0 {
        t.Errorf("expected positive revenue, got %f", revenue)
    }
}

func monthlySalesTest(ctx context.Context, t *testing.T, store *tablebuilder.Store, sd unitest.SeedData) {
    config := &tablebuilder.Config{
        Title:         "Monthly Sales",
        WidgetType:    "chart",
        Visualization: "line",
        DataSource: []tablebuilder.DataSource{{
            Source: "order_line_items",
            Schema: "sales",
            Metrics: []tablebuilder.MetricConfig{
                {
                    Name:     "sales",
                    Function: "sum",
                    Expression: &tablebuilder.ExpressionConfig{
                        Operator: "multiply",
                        Columns:  []string{"order_line_items.quantity", "product_costs.selling_price"},
                    },
                },
            },
            GroupBy: &tablebuilder.GroupByConfig{
                Column:   "orders.created_date",
                Interval: "month",
                Alias:    "month",
            },
            Select: tablebuilder.SelectConfig{
                ForeignTables: []tablebuilder.ForeignTable{
                    {
                        Table:            "orders",
                        Schema:           "sales",
                        RelationshipFrom: "order_line_items.order_id",
                        RelationshipTo:   "orders.id",
                        JoinType:         "inner",
                    },
                    {
                        Table:            "products",
                        Schema:           "products",
                        RelationshipFrom: "order_line_items.product_id",
                        RelationshipTo:   "products.id",
                        JoinType:         "inner",
                        ForeignTables: []tablebuilder.ForeignTable{
                            {
                                Table:            "product_costs",
                                Schema:           "products",
                                RelationshipFrom: "products.id",
                                RelationshipTo:   "product_costs.product_id",
                                JoinType:         "inner",
                            },
                        },
                    },
                },
            },
            Sort: []tablebuilder.Sort{
                {Column: "month", Direction: "asc"},
            },
        }},
    }

    params := tablebuilder.QueryParams{}
    result, err := store.FetchTableData(ctx, config, params)
    if err != nil {
        t.Fatalf("fetch failed: %v", err)
    }

    // Should return rows grouped by month
    if len(result.Data) == 0 {
        t.Error("expected at least 1 row")
    }

    // Each row should have month and sales columns
    for _, row := range result.Data {
        if _, ok := row["month"]; !ok {
            t.Error("expected month column")
        }
        if _, ok := row["sales"]; !ok {
            t.Error("expected sales column")
        }
    }
}

func countOrdersTest(ctx context.Context, t *testing.T, store *tablebuilder.Store, sd unitest.SeedData) {
    config := &tablebuilder.Config{
        Title:         "Order Count",
        WidgetType:    "chart",
        Visualization: "kpi",
        DataSource: []tablebuilder.DataSource{{
            Source: "orders",
            Schema: "sales",
            Metrics: []tablebuilder.MetricConfig{
                {Name: "order_count", Function: "count", Column: "orders.id"},
            },
        }},
    }

    params := tablebuilder.QueryParams{}
    result, err := store.FetchTableData(ctx, config, params)
    if err != nil {
        t.Fatalf("fetch failed: %v", err)
    }

    if len(result.Data) != 1 {
        t.Errorf("expected 1 row, got %d", len(result.Data))
    }

    count, ok := result.Data[0]["order_count"]
    if !ok {
        t.Error("expected order_count column")
    }

    // Should match seeded order count
    expectedCount := len(sd.Orders)
    actualCount := int(count.(int64))
    if actualCount != expectedCount {
        t.Errorf("expected %d orders, got %d", expectedCount, actualCount)
    }
}

func revenueByCategoryTest(ctx context.Context, t *testing.T, store *tablebuilder.Store, sd unitest.SeedData) {
    config := &tablebuilder.Config{
        Title:         "Revenue by Category",
        WidgetType:    "chart",
        Visualization: "pie",
        DataSource: []tablebuilder.DataSource{{
            Source: "order_line_items",
            Schema: "sales",
            Metrics: []tablebuilder.MetricConfig{
                {
                    Name:     "revenue",
                    Function: "sum",
                    Expression: &tablebuilder.ExpressionConfig{
                        Operator: "multiply",
                        Columns:  []string{"order_line_items.quantity", "product_costs.selling_price"},
                    },
                },
            },
            GroupBy: &tablebuilder.GroupByConfig{
                Column: "product_categories.name",
                Alias:  "category",
            },
            Select: tablebuilder.SelectConfig{
                ForeignTables: []tablebuilder.ForeignTable{
                    {
                        Table:            "products",
                        Schema:           "products",
                        RelationshipFrom: "order_line_items.product_id",
                        RelationshipTo:   "products.id",
                        JoinType:         "inner",
                        ForeignTables: []tablebuilder.ForeignTable{
                            {
                                Table:            "product_categories",
                                Schema:           "products",
                                RelationshipFrom: "products.category_id",
                                RelationshipTo:   "product_categories.id",
                                JoinType:         "inner",
                            },
                            {
                                Table:            "product_costs",
                                Schema:           "products",
                                RelationshipFrom: "products.id",
                                RelationshipTo:   "product_costs.product_id",
                                JoinType:         "inner",
                            },
                        },
                    },
                },
            },
        }},
    }

    params := tablebuilder.QueryParams{}
    result, err := store.FetchTableData(ctx, config, params)
    if err != nil {
        t.Fatalf("fetch failed: %v", err)
    }

    // Should have rows grouped by category
    for _, row := range result.Data {
        if _, ok := row["category"]; !ok {
            t.Error("expected category column")
        }
        if _, ok := row["revenue"]; !ok {
            t.Error("expected revenue column")
        }
    }
}

func multipleMetricsTest(ctx context.Context, t *testing.T, store *tablebuilder.Store, sd unitest.SeedData) {
    config := &tablebuilder.Config{
        Title:         "Revenue and Orders",
        WidgetType:    "chart",
        Visualization: "combo",
        DataSource: []tablebuilder.DataSource{{
            Source: "order_line_items",
            Schema: "sales",
            Metrics: []tablebuilder.MetricConfig{
                {
                    Name:     "revenue",
                    Function: "sum",
                    Expression: &tablebuilder.ExpressionConfig{
                        Operator: "multiply",
                        Columns:  []string{"order_line_items.quantity", "product_costs.selling_price"},
                    },
                },
                {
                    Name:     "order_count",
                    Function: "count_distinct",
                    Column:   "orders.id",
                },
            },
            GroupBy: &tablebuilder.GroupByConfig{
                Column:   "orders.created_date",
                Interval: "month",
                Alias:    "month",
            },
            Select: tablebuilder.SelectConfig{
                ForeignTables: []tablebuilder.ForeignTable{
                    {
                        Table:            "orders",
                        Schema:           "sales",
                        RelationshipFrom: "order_line_items.order_id",
                        RelationshipTo:   "orders.id",
                        JoinType:         "inner",
                    },
                    {
                        Table:            "products",
                        Schema:           "products",
                        RelationshipFrom: "order_line_items.product_id",
                        RelationshipTo:   "products.id",
                        JoinType:         "inner",
                        ForeignTables: []tablebuilder.ForeignTable{
                            {
                                Table:            "product_costs",
                                Schema:           "products",
                                RelationshipFrom: "products.id",
                                RelationshipTo:   "product_costs.product_id",
                                JoinType:         "inner",
                            },
                        },
                    },
                },
            },
            Sort: []tablebuilder.Sort{
                {Column: "month", Direction: "asc"},
            },
        }},
    }

    params := tablebuilder.QueryParams{}
    result, err := store.FetchTableData(ctx, config, params)
    if err != nil {
        t.Fatalf("fetch failed: %v", err)
    }

    // Each row should have all columns
    for _, row := range result.Data {
        if _, ok := row["month"]; !ok {
            t.Error("expected month column")
        }
        if _, ok := row["revenue"]; !ok {
            t.Error("expected revenue column")
        }
        if _, ok := row["order_count"]; !ok {
            t.Error("expected order_count column")
        }
    }
}
```

---

## Phase 5: Validation Functions (`builder.go` or new `validation.go`)

```go
// ValidateMetricConfig validates a metric configuration
func ValidateMetricConfig(metric MetricConfig) error {
    if metric.Name == "" {
        return fmt.Errorf("metric name is required")
    }

    if _, ok := AllowedAggregateFunctions[metric.Function]; !ok {
        return fmt.Errorf("invalid aggregate function: %s", metric.Function)
    }

    if metric.Column == "" && metric.Expression == nil {
        return fmt.Errorf("metric must have column or expression")
    }

    if metric.Expression != nil {
        if err := ValidateExpressionConfig(metric.Expression); err != nil {
            return fmt.Errorf("invalid expression: %w", err)
        }
    }

    return nil
}

// ValidateExpressionConfig validates an expression configuration
func ValidateExpressionConfig(expr *ExpressionConfig) error {
    if _, ok := AllowedOperators[expr.Operator]; !ok {
        return fmt.Errorf("invalid operator: %s", expr.Operator)
    }

    if len(expr.Columns) < 2 {
        return fmt.Errorf("expression requires at least 2 columns")
    }

    // Validate column names (no SQL injection)
    for _, col := range expr.Columns {
        if !isValidColumnReference(col) {
            return fmt.Errorf("invalid column reference: %s", col)
        }
    }

    return nil
}

// ValidateGroupByConfig validates a group by configuration
func ValidateGroupByConfig(groupBy *GroupByConfig) error {
    if groupBy.Column == "" {
        return fmt.Errorf("group by column is required")
    }

    if groupBy.Interval != "" {
        if _, ok := AllowedIntervals[groupBy.Interval]; !ok {
            return fmt.Errorf("invalid interval: %s", groupBy.Interval)
        }
    }

    if !isValidColumnReference(groupBy.Column) {
        return fmt.Errorf("invalid column reference: %s", groupBy.Column)
    }

    return nil
}

// isValidColumnReference checks if a column reference is safe
func isValidColumnReference(col string) bool {
    // Allow: table.column, schema.table.column, column
    // Disallow: SQL keywords, special characters, etc.
    validPattern := regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*(\.[a-zA-Z_][a-zA-Z0-9_]*)*$`)
    return validPattern.MatchString(col)
}
```

---

## File Changes Summary

| File | Changes |
|------|---------|
| `model.go` | Add `MetricConfig`, `ExpressionConfig`, `GroupByConfig` types; add to `DataSource`; add whitelist constants |
| `builder.go` | Add `buildMetricQuery()`, `buildMetricExpression()`, `buildGroupByExpression()`; modify `BuildQuery()` to detect metrics |
| `validation.go` (new) | Validation functions for metrics, expressions, group by |
| `seedmodels/charts.go` | Convert all 14 chart configs to use new Metrics/GroupBy structure |
| `tablebuilder_test.go` | Add model serialization tests, query builder unit tests, integration tests |

---

## Estimated Effort

| Phase | Description | Lines |
|-------|-------------|-------|
| 1 | Model changes | ~80 |
| 2 | Query builder | ~200 |
| 3 | Seed data conversion | ~400 |
| 4 | Unit tests | ~500 |
| 5 | Validation | ~80 |
| **Total** | | **~1,260** |

---

## Testing Strategy

1. **Unit tests** - Verify SQL generation without database
2. **Validation tests** - Verify whitelist enforcement
3. **Integration tests** - Full round-trip with test database
4. **Regression tests** - Ensure existing table queries still work

## Success Criteria

- [ ] All 14 chart seed configs converted to new format
- [ ] KPI queries return single aggregated value
- [ ] Time-series queries return grouped data with correct intervals
- [ ] Invalid functions/operators/intervals are rejected
- [ ] Existing table queries continue to work unchanged
- [ ] All tests pass
