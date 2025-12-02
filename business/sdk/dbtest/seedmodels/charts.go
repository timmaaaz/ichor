package seedmodels

import (
	"encoding/json"

	"github.com/timmaaaz/ichor/business/sdk/tablebuilder"
)

// =============================================================================
// CHART SEED CONFIGURATIONS
// =============================================================================
// These chart configurations demonstrate all 14 chart types supported by the
// Chart Integration system. They serve as examples and starting points for
// administrators to clone and customize.
//
// Chart types covered:
// - KPI Cards (kpi)
// - Gauges (gauge)
// - Line Charts (line)
// - Bar Charts (bar)
// - Stacked Bar Charts (stacked-bar)
// - Stacked Area Charts (stacked-area)
// - Pie Charts (pie)
// - Combo Charts (combo)
// - Waterfall Charts (waterfall)
// - Funnel Charts (funnel)
// - Heatmap Charts (heatmap)
// - Treemap Charts (treemap)
// - Gantt Charts (gantt)
// =============================================================================

// Helper function to create chart visual settings JSON
func createChartSettings(settings tablebuilder.ChartVisualSettings) string {
	bytes, _ := json.Marshal(settings)
	return string(bytes)
}

// =============================================================================
// KPI Cards
// =============================================================================

// SeedKPITotalRevenue - KPI card showing total revenue from orders
var SeedKPITotalRevenue = &tablebuilder.Config{
	Title:         "Total Revenue",
	WidgetType:    "chart",
	Visualization: "kpi",
	DataSource: []tablebuilder.DataSource{
		{
			Type:   "query",
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
		},
	},
	VisualSettings: tablebuilder.VisualSettings{
		Columns: map[string]tablebuilder.ColumnConfig{
			"_chart": {
				CellTemplate: createChartSettings(tablebuilder.ChartVisualSettings{
					ChartType:    "kpi",
					ValueColumns: []string{"total_revenue"},
					KPI: &tablebuilder.KPIConfig{
						Label:  "Total Revenue",
						Format: "currency",
					},
				}),
			},
		},
	},
}

// SeedKPIOrderCount - KPI card showing total order count
var SeedKPIOrderCount = &tablebuilder.Config{
	Title:         "Total Orders",
	WidgetType:    "chart",
	Visualization: "kpi",
	DataSource: []tablebuilder.DataSource{
		{
			Type:   "query",
			Source: "orders",
			Schema: "sales",
			Metrics: []tablebuilder.MetricConfig{
				{
					Name:     "order_count",
					Function: "count",
					Column:   "orders.id",
				},
			},
		},
	},
	VisualSettings: tablebuilder.VisualSettings{
		Columns: map[string]tablebuilder.ColumnConfig{
			"_chart": {
				CellTemplate: createChartSettings(tablebuilder.ChartVisualSettings{
					ChartType:    "kpi",
					ValueColumns: []string{"order_count"},
					KPI: &tablebuilder.KPIConfig{
						Label:  "Total Orders",
						Format: "number",
					},
				}),
			},
		},
	},
}

// =============================================================================
// Gauge Charts
// =============================================================================

// SeedGaugeRevenueTarget - Gauge showing revenue progress toward target
var SeedGaugeRevenueTarget = &tablebuilder.Config{
	Title:         "Revenue Progress",
	WidgetType:    "chart",
	Visualization: "gauge",
	DataSource: []tablebuilder.DataSource{
		{
			Type:   "query",
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
		},
	},
	VisualSettings: tablebuilder.VisualSettings{
		Columns: map[string]tablebuilder.ColumnConfig{
			"_chart": {
				CellTemplate: createChartSettings(tablebuilder.ChartVisualSettings{
					ChartType:    "gauge",
					ValueColumns: []string{"revenue"},
					KPI: &tablebuilder.KPIConfig{
						Label:             "Revenue vs Target",
						Format:            "currency",
						TargetValue:       1000000,
						ThresholdWarning:  500000,
						ThresholdCritical: 250000,
					},
				}),
			},
		},
	},
}

// =============================================================================
// Line Charts
// =============================================================================

// SeedLineMonthlySales - Line chart showing monthly sales trend
var SeedLineMonthlySales = &tablebuilder.Config{
	Title:         "Monthly Sales Trend",
	WidgetType:    "chart",
	Visualization: "line",
	DataSource: []tablebuilder.DataSource{
		{
			Type:   "query",
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
		},
	},
	VisualSettings: tablebuilder.VisualSettings{
		Columns: map[string]tablebuilder.ColumnConfig{
			"_chart": {
				CellTemplate: createChartSettings(tablebuilder.ChartVisualSettings{
					ChartType:      "line",
					CategoryColumn: "month",
					ValueColumns:   []string{"sales"},
					XAxis: &tablebuilder.AxisConfig{
						Title: "Month",
						Type:  "time",
					},
					YAxis: &tablebuilder.AxisConfig{
						Title:  "Sales ($)",
						Type:   "value",
						Format: "currency",
					},
				}),
			},
		},
	},
}

// =============================================================================
// Bar Charts
// =============================================================================

// SeedBarTopProducts - Bar chart showing top products by revenue
var SeedBarTopProducts = &tablebuilder.Config{
	Title:         "Top Products by Revenue",
	WidgetType:    "chart",
	Visualization: "bar",
	DataSource: []tablebuilder.DataSource{
		{
			Type:   "query",
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
				Column: "products.name",
				Alias:  "product_name",
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
			Sort: []tablebuilder.Sort{
				{Column: "revenue", Direction: "desc"},
			},
			Rows: 10,
		},
	},
	VisualSettings: tablebuilder.VisualSettings{
		Columns: map[string]tablebuilder.ColumnConfig{
			"_chart": {
				CellTemplate: createChartSettings(tablebuilder.ChartVisualSettings{
					ChartType:      "bar",
					CategoryColumn: "product_name",
					ValueColumns:   []string{"revenue"},
					XAxis: &tablebuilder.AxisConfig{
						Title: "Product",
						Type:  "category",
					},
					YAxis: &tablebuilder.AxisConfig{
						Title:  "Revenue ($)",
						Type:   "value",
						Format: "currency",
					},
				}),
			},
		},
	},
}

// =============================================================================
// Stacked Bar Charts
// =============================================================================

// SeedStackedBarRegionCategory - Stacked bar chart showing sales by region and category
var SeedStackedBarRegionCategory = &tablebuilder.Config{
	Title:         "Sales by Region",
	WidgetType:    "chart",
	Visualization: "stacked-bar",
	DataSource: []tablebuilder.DataSource{
		{
			Type:   "query",
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
				Column: "cities.name",
				Alias:  "region",
			},
			Select: tablebuilder.SelectConfig{
				ForeignTables: []tablebuilder.ForeignTable{
					{
						Table:            "orders",
						Schema:           "sales",
						RelationshipFrom: "order_line_items.order_id",
						RelationshipTo:   "orders.id",
						JoinType:         "inner",
						ForeignTables: []tablebuilder.ForeignTable{
							{
								Table:            "customers",
								Schema:           "sales",
								RelationshipFrom: "orders.customer_id",
								RelationshipTo:   "customers.id",
								JoinType:         "left",
								ForeignTables: []tablebuilder.ForeignTable{
									{
										Table:            "streets",
										Schema:           "geography",
										RelationshipFrom: "customers.delivery_address_id",
										RelationshipTo:   "streets.id",
										JoinType:         "left",
										ForeignTables: []tablebuilder.ForeignTable{
											{
												Table:            "cities",
												Schema:           "geography",
												RelationshipFrom: "streets.city_id",
												RelationshipTo:   "cities.id",
												JoinType:         "left",
											},
										},
									},
								},
							},
						},
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
		},
	},
	VisualSettings: tablebuilder.VisualSettings{
		Columns: map[string]tablebuilder.ColumnConfig{
			"_chart": {
				CellTemplate: createChartSettings(tablebuilder.ChartVisualSettings{
					ChartType:      "stacked-bar",
					CategoryColumn: "region",
					ValueColumns:   []string{"sales"},
				}),
			},
		},
	},
}

// =============================================================================
// Stacked Area Charts
// =============================================================================

// SeedStackedAreaCumulative - Stacked area chart showing cumulative revenue
var SeedStackedAreaCumulative = &tablebuilder.Config{
	Title:         "Cumulative Revenue Over Time",
	WidgetType:    "chart",
	Visualization: "stacked-area",
	DataSource: []tablebuilder.DataSource{
		{
			Type:   "query",
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
		},
	},
	VisualSettings: tablebuilder.VisualSettings{
		Columns: map[string]tablebuilder.ColumnConfig{
			"_chart": {
				CellTemplate: createChartSettings(tablebuilder.ChartVisualSettings{
					ChartType:      "stacked-area",
					CategoryColumn: "month",
					ValueColumns:   []string{"revenue"},
					XAxis: &tablebuilder.AxisConfig{
						Title: "Month",
						Type:  "time",
					},
					YAxis: &tablebuilder.AxisConfig{
						Title:  "Revenue ($)",
						Type:   "value",
						Format: "currency",
					},
				}),
			},
		},
	},
}

// =============================================================================
// Pie Charts
// =============================================================================

// SeedPieRevenueCategory - Pie chart showing revenue by product category
var SeedPieRevenueCategory = &tablebuilder.Config{
	Title:         "Revenue by Category",
	WidgetType:    "chart",
	Visualization: "pie",
	DataSource: []tablebuilder.DataSource{
		{
			Type:   "query",
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
		},
	},
	VisualSettings: tablebuilder.VisualSettings{
		Columns: map[string]tablebuilder.ColumnConfig{
			"_chart": {
				CellTemplate: createChartSettings(tablebuilder.ChartVisualSettings{
					ChartType:      "pie",
					CategoryColumn: "category",
					ValueColumns:   []string{"revenue"},
				}),
			},
		},
	},
}

// =============================================================================
// Combo Charts
// =============================================================================

// SeedComboRevenueOrders - Combo chart comparing revenue (bars) vs order count (line)
var SeedComboRevenueOrders = &tablebuilder.Config{
	Title:         "Revenue vs Order Count",
	WidgetType:    "chart",
	Visualization: "combo",
	DataSource: []tablebuilder.DataSource{
		{
			Type:   "query",
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
		},
	},
	VisualSettings: tablebuilder.VisualSettings{
		Columns: map[string]tablebuilder.ColumnConfig{
			"_chart": {
				CellTemplate: createChartSettings(tablebuilder.ChartVisualSettings{
					ChartType:      "combo",
					CategoryColumn: "month",
					ValueColumns:   []string{"revenue", "order_count"},
					SeriesConfig: []tablebuilder.SeriesConfig{
						{Column: "revenue", Type: "bar", YAxisIndex: 0, Label: "Revenue"},
						{Column: "order_count", Type: "line", YAxisIndex: 1, Label: "Orders"},
					},
					XAxis: &tablebuilder.AxisConfig{
						Title: "Month",
						Type:  "time",
					},
					YAxis: &tablebuilder.AxisConfig{
						Title:  "Revenue ($)",
						Type:   "value",
						Format: "currency",
					},
					Y2Axis: &tablebuilder.AxisConfig{
						Title: "Order Count",
						Type:  "value",
					},
				}),
			},
		},
	},
}

// =============================================================================
// Waterfall Charts
// =============================================================================

// SeedWaterfallProfit - Waterfall chart showing profit breakdown (uses static data)
// Note: Waterfall uses a simplified static category - in production would use multiple queries
var SeedWaterfallProfit = &tablebuilder.Config{
	Title:         "Profit Breakdown",
	WidgetType:    "chart",
	Visualization: "waterfall",
	DataSource: []tablebuilder.DataSource{
		{
			Type:   "query",
			Source: "order_line_items",
			Schema: "sales",
			Metrics: []tablebuilder.MetricConfig{
				{
					Name:     "amount",
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
		},
	},
	VisualSettings: tablebuilder.VisualSettings{
		Columns: map[string]tablebuilder.ColumnConfig{
			"_chart": {
				CellTemplate: createChartSettings(tablebuilder.ChartVisualSettings{
					ChartType:      "waterfall",
					CategoryColumn: "category",
					ValueColumns:   []string{"amount"},
				}),
			},
		},
	},
}

// =============================================================================
// Funnel Charts
// =============================================================================

// SeedFunnelPipeline - Funnel chart showing sales pipeline stages
var SeedFunnelPipeline = &tablebuilder.Config{
	Title:         "Sales Pipeline",
	WidgetType:    "chart",
	Visualization: "funnel",
	DataSource: []tablebuilder.DataSource{
		{
			Type:   "query",
			Source: "orders",
			Schema: "sales",
			Metrics: []tablebuilder.MetricConfig{
				{
					Name:     "count",
					Function: "count",
					Column:   "orders.id",
				},
			},
			GroupBy: &tablebuilder.GroupByConfig{
				Column: "order_fulfillment_statuses.name",
				Alias:  "status",
			},
			Select: tablebuilder.SelectConfig{
				ForeignTables: []tablebuilder.ForeignTable{
					{
						Table:            "order_fulfillment_statuses",
						Schema:           "sales",
						RelationshipFrom: "orders.order_fulfillment_status_id",
						RelationshipTo:   "order_fulfillment_statuses.id",
						JoinType:         "left",
					},
				},
			},
		},
	},
	VisualSettings: tablebuilder.VisualSettings{
		Columns: map[string]tablebuilder.ColumnConfig{
			"_chart": {
				CellTemplate: createChartSettings(tablebuilder.ChartVisualSettings{
					ChartType:      "funnel",
					CategoryColumn: "status",
					ValueColumns:   []string{"count"},
				}),
			},
		},
	},
}

// =============================================================================
// Heatmap Charts
// =============================================================================

// SeedHeatmapSalesTime - Heatmap showing order count by day of week and hour
// NOTE: Heatmaps require multi-dimensional grouping (day AND hour) which is beyond
// the single GroupBy config. This chart uses traditional Columns with EXTRACT functions
// and requires custom query handling. Future enhancement could add multi-group support.
var SeedHeatmapSalesTime = &tablebuilder.Config{
	Title:         "Orders by Day and Hour",
	WidgetType:    "chart",
	Visualization: "heatmap",
	DataSource: []tablebuilder.DataSource{
		{
			Type:   "query",
			Source: "orders",
			Schema: "sales",
			Select: tablebuilder.SelectConfig{
				Columns: []tablebuilder.ColumnDefinition{
					{Name: "day_of_week", Alias: "day", TableColumn: "EXTRACT(DOW FROM orders.created_date)"},
					{Name: "hour", Alias: "hour", TableColumn: "EXTRACT(HOUR FROM orders.created_date)"},
					{Name: "order_count", Alias: "count", TableColumn: "COUNT(orders.id)"},
				},
			},
		},
	},
	VisualSettings: tablebuilder.VisualSettings{
		Columns: map[string]tablebuilder.ColumnConfig{
			"_chart": {
				CellTemplate: createChartSettings(tablebuilder.ChartVisualSettings{
					ChartType:       "heatmap",
					XCategoryColumn: "hour",
					YCategoryColumn: "day",
					ValueColumns:    []string{"count"},
				}),
			},
		},
	},
}

// =============================================================================
// Treemap Charts
// =============================================================================

// SeedTreemapRevenue - Treemap showing revenue breakdown by product
var SeedTreemapRevenue = &tablebuilder.Config{
	Title:         "Revenue Breakdown by Product",
	WidgetType:    "chart",
	Visualization: "treemap",
	DataSource: []tablebuilder.DataSource{
		{
			Type:   "query",
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
				Column: "products.name",
				Alias:  "product",
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
			Rows: 20,
		},
	},
	VisualSettings: tablebuilder.VisualSettings{
		Columns: map[string]tablebuilder.ColumnConfig{
			"_chart": {
				CellTemplate: createChartSettings(tablebuilder.ChartVisualSettings{
					ChartType:      "treemap",
					CategoryColumn: "product",
					ValueColumns:   []string{"revenue"},
				}),
			},
		},
	},
}

// =============================================================================
// Gantt Charts
// =============================================================================

// SeedGanttProject - Gantt chart showing project timeline (demo with purchase orders)
// NOTE: Gantt charts don't use aggregation - they display row-level data directly.
// Uses standard Columns with proper column names from the base table.
var SeedGanttProject = &tablebuilder.Config{
	Title:         "Purchase Order Timeline",
	WidgetType:    "chart",
	Visualization: "gantt",
	DataSource: []tablebuilder.DataSource{
		{
			Type:   "query",
			Source: "purchase_orders",
			Schema: "procurement",
			Select: tablebuilder.SelectConfig{
				Columns: []tablebuilder.ColumnDefinition{
					{Name: "id", Alias: "id"},
					{Name: "order_number", Alias: "name"},
					{Name: "order_date", Alias: "start_date"},
					{Name: "expected_delivery_date", Alias: "end_date"},
				},
			},
			Rows: 20,
			Sort: []tablebuilder.Sort{
				{Column: "order_date", Direction: "asc"},
			},
		},
	},
	VisualSettings: tablebuilder.VisualSettings{
		Columns: map[string]tablebuilder.ColumnConfig{
			"_chart": {
				CellTemplate: createChartSettings(tablebuilder.ChartVisualSettings{
					ChartType:   "gantt",
					NameColumn:  "name",
					StartColumn: "start_date",
					EndColumn:   "end_date",
				}),
			},
		},
	},
}

// =============================================================================
// List of all chart configurations for seeding
// =============================================================================

// ChartConfigs contains all seed chart configurations with their names and descriptions
var ChartConfigs = []struct {
	Name        string
	Description string
	Config      *tablebuilder.Config
}{
	// KPI Cards
	{Name: "seed_kpi_total_revenue", Description: "KPI showing total revenue", Config: SeedKPITotalRevenue},
	{Name: "seed_kpi_order_count", Description: "KPI showing total order count", Config: SeedKPIOrderCount},

	// Gauges
	{Name: "seed_gauge_revenue_target", Description: "Gauge showing revenue progress toward target", Config: SeedGaugeRevenueTarget},

	// Line Charts
	{Name: "seed_line_monthly_sales", Description: "Line chart showing monthly sales trend", Config: SeedLineMonthlySales},

	// Bar Charts
	{Name: "seed_bar_top_products", Description: "Bar chart showing top products by revenue", Config: SeedBarTopProducts},

	// Stacked Bar Charts
	{Name: "seed_stacked_bar_region", Description: "Stacked bar chart showing sales by region", Config: SeedStackedBarRegionCategory},

	// Stacked Area Charts
	{Name: "seed_stacked_area_cumulative", Description: "Stacked area chart showing cumulative revenue", Config: SeedStackedAreaCumulative},

	// Pie Charts
	{Name: "seed_pie_revenue_category", Description: "Pie chart showing revenue by category", Config: SeedPieRevenueCategory},

	// Combo Charts
	{Name: "seed_combo_revenue_orders", Description: "Combo chart comparing revenue vs order count", Config: SeedComboRevenueOrders},

	// Waterfall Charts
	{Name: "seed_waterfall_profit", Description: "Waterfall chart showing profit breakdown", Config: SeedWaterfallProfit},

	// Funnel Charts
	{Name: "seed_funnel_pipeline", Description: "Funnel chart showing sales pipeline stages", Config: SeedFunnelPipeline},

	// Heatmap Charts
	{Name: "seed_heatmap_sales_time", Description: "Heatmap showing orders by day and hour", Config: SeedHeatmapSalesTime},

	// Treemap Charts
	{Name: "seed_treemap_revenue", Description: "Treemap showing revenue breakdown by product", Config: SeedTreemapRevenue},

	// Gantt Charts
	{Name: "seed_gantt_project", Description: "Gantt chart showing purchase order timeline", Config: SeedGanttProject},
}
