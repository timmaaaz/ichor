package tablebuilder_test

import (
	"strings"
	"testing"

	"github.com/timmaaaz/ichor/business/sdk/tablebuilder"
)

// TestMultiGroupBy tests the multi-dimensional grouping functionality
func TestMultiGroupBy(t *testing.T) {
	t.Parallel()

	t.Run("single GroupBy in slice - backward compatibility", func(t *testing.T) {
		t.Parallel()
		config := &tablebuilder.Config{
			Title:         "Monthly Revenue",
			WidgetType:    "chart",
			Visualization: "bar",
			DataSource: []tablebuilder.DataSource{
				{
					Type:   "query",
					Source: "orders",
					Schema: "sales",
					GroupBy: []tablebuilder.GroupByConfig{
						{
							Column:   "orders.created_date",
							Interval: "month",
							Alias:    "month",
						},
					},
					Metrics: []tablebuilder.MetricConfig{
						{
							Name:     "revenue",
							Function: "sum",
							Column:   "orders.amount",
						},
					},
				},
			},
		}

		// Should work exactly like old *GroupByConfig
		if len(config.DataSource[0].GroupBy) != 1 {
			t.Errorf("expected 1 GroupBy, got %d", len(config.DataSource[0].GroupBy))
		}
		if config.DataSource[0].GroupBy[0].Column != "orders.created_date" {
			t.Errorf("GroupBy column mismatch: got %s", config.DataSource[0].GroupBy[0].Column)
		}
	})

	t.Run("empty GroupBy slice - no grouping", func(t *testing.T) {
		t.Parallel()
		config := &tablebuilder.Config{
			Title:         "Total Revenue",
			WidgetType:    "kpi",
			Visualization: "kpi",
			DataSource: []tablebuilder.DataSource{
				{
					Type:   "query",
					Source: "orders",
					Schema: "sales",
					GroupBy: []tablebuilder.GroupByConfig{}, // Empty slice
					Metrics: []tablebuilder.MetricConfig{
						{
							Name:     "total",
							Function: "sum",
							Column:   "orders.amount",
						},
					},
				},
			},
		}

		// Empty slice should behave like no grouping
		if len(config.DataSource[0].GroupBy) != 0 {
			t.Errorf("expected 0 GroupBy, got %d", len(config.DataSource[0].GroupBy))
		}
	})

	t.Run("multiple GroupBy with simple columns", func(t *testing.T) {
		t.Parallel()
		config := &tablebuilder.Config{
			Title:         "Revenue by Region and Product",
			WidgetType:    "chart",
			Visualization: "bar",
			DataSource: []tablebuilder.DataSource{
				{
					Type:   "query",
					Source: "order_line_items",
					Schema: "sales",
					GroupBy: []tablebuilder.GroupByConfig{
						{
							Column: "cities.name",
							Alias:  "region",
						},
						{
							Column: "products.name",
							Alias:  "product",
						},
					},
					Metrics: []tablebuilder.MetricConfig{
						{
							Name:     "revenue",
							Function: "sum",
							Column:   "order_line_items.total_price",
						},
					},
				},
			},
		}

		if len(config.DataSource[0].GroupBy) != 2 {
			t.Errorf("expected 2 GroupBy, got %d", len(config.DataSource[0].GroupBy))
		}
		if config.DataSource[0].GroupBy[0].Column != "cities.name" {
			t.Errorf("first GroupBy column mismatch: got %s", config.DataSource[0].GroupBy[0].Column)
		}
		if config.DataSource[0].GroupBy[1].Column != "products.name" {
			t.Errorf("second GroupBy column mismatch: got %s", config.DataSource[0].GroupBy[1].Column)
		}
	})

	t.Run("multiple GroupBy with SQL expressions", func(t *testing.T) {
		t.Parallel()
		config := &tablebuilder.Config{
			Title:         "Orders by Day and Hour",
			WidgetType:    "chart",
			Visualization: "heatmap",
			DataSource: []tablebuilder.DataSource{
				{
					Type:   "query",
					Source: "orders",
					Schema: "sales",
					GroupBy: []tablebuilder.GroupByConfig{
						{
							Column:     "EXTRACT(DOW FROM orders.created_date)",
							Alias:      "day",
							Expression: true,
						},
						{
							Column:     "EXTRACT(HOUR FROM orders.created_date)",
							Alias:      "hour",
							Expression: true,
						},
					},
					Metrics: []tablebuilder.MetricConfig{
						{
							Name:     "count",
							Function: "count",
							Column:   "orders.id",
						},
					},
				},
			},
		}

		if len(config.DataSource[0].GroupBy) != 2 {
			t.Errorf("expected 2 GroupBy, got %d", len(config.DataSource[0].GroupBy))
		}

		// Verify first GroupBy (day)
		dayGroup := config.DataSource[0].GroupBy[0]
		if dayGroup.Column != "EXTRACT(DOW FROM orders.created_date)" {
			t.Errorf("day GroupBy column mismatch: got %s", dayGroup.Column)
		}
		if dayGroup.Alias != "day" {
			t.Errorf("day GroupBy alias mismatch: got %s", dayGroup.Alias)
		}
		if !dayGroup.Expression {
			t.Error("day GroupBy should be marked as Expression")
		}

		// Verify second GroupBy (hour)
		hourGroup := config.DataSource[0].GroupBy[1]
		if hourGroup.Column != "EXTRACT(HOUR FROM orders.created_date)" {
			t.Errorf("hour GroupBy column mismatch: got %s", hourGroup.Column)
		}
		if hourGroup.Alias != "hour" {
			t.Errorf("hour GroupBy alias mismatch: got %s", hourGroup.Alias)
		}
		if !hourGroup.Expression {
			t.Error("hour GroupBy should be marked as Expression")
		}
	})

	t.Run("mixed simple and expression GroupBy", func(t *testing.T) {
		t.Parallel()
		config := &tablebuilder.Config{
			Title:         "Revenue by Month and Product Category",
			WidgetType:    "chart",
			Visualization: "bar",
			DataSource: []tablebuilder.DataSource{
				{
					Type:   "query",
					Source: "order_line_items",
					Schema: "sales",
					GroupBy: []tablebuilder.GroupByConfig{
						{
							Column:   "orders.created_date",
							Interval: "month",
							Alias:    "month",
						},
						{
							Column:     "COALESCE(product_categories.name, 'Uncategorized')",
							Alias:      "category",
							Expression: true,
						},
					},
					Metrics: []tablebuilder.MetricConfig{
						{
							Name:     "revenue",
							Function: "sum",
							Column:   "order_line_items.total_price",
						},
					},
				},
			},
		}

		if len(config.DataSource[0].GroupBy) != 2 {
			t.Errorf("expected 2 GroupBy, got %d", len(config.DataSource[0].GroupBy))
		}

		// First GroupBy is date interval (simple)
		monthGroup := config.DataSource[0].GroupBy[0]
		if monthGroup.Interval != "month" {
			t.Errorf("month GroupBy interval mismatch: got %s", monthGroup.Interval)
		}
		if monthGroup.Expression {
			t.Error("month GroupBy should not be marked as Expression")
		}

		// Second GroupBy is SQL expression
		categoryGroup := config.DataSource[0].GroupBy[1]
		if !categoryGroup.Expression {
			t.Error("category GroupBy should be marked as Expression")
		}
		if categoryGroup.Alias != "category" {
			t.Errorf("category GroupBy alias mismatch: got %s", categoryGroup.Alias)
		}
	})

	t.Run("Expression GroupBy requires alias", func(t *testing.T) {
		t.Parallel()

		// Expression GroupBy with alias — should be valid.
		groupBy := &tablebuilder.GroupByConfig{
			Column:     "EXTRACT(DOW FROM orders.created_date)",
			Expression: true,
			Alias:      "day_of_week",
		}
		if err := tablebuilder.ValidateGroupByConfig(groupBy); err != nil {
			t.Errorf("expression GroupBy with alias should be valid: %v", err)
		}

		// Without alias — confirm validate behavior (builder handles alias at query time).
		groupByNoAlias := &tablebuilder.GroupByConfig{
			Column:     "EXTRACT(DOW FROM orders.created_date)",
			Expression: true,
		}
		// ValidateGroupByConfig permits expression without alias; alias enforcement happens at BuildQuery time.
		if err := tablebuilder.ValidateGroupByConfig(groupByNoAlias); err != nil {
			t.Errorf("ValidateGroupByConfig should allow expression without alias (enforced at query build time): %v", err)
		}
	})

	t.Run("three-dimensional grouping", func(t *testing.T) {
		t.Parallel()
		config := &tablebuilder.Config{
			Title:         "Sales by Region, Product, and Month",
			WidgetType:    "chart",
			Visualization: "bar",
			DataSource: []tablebuilder.DataSource{
				{
					Type:   "query",
					Source: "order_line_items",
					Schema: "sales",
					GroupBy: []tablebuilder.GroupByConfig{
						{
							Column: "cities.name",
							Alias:  "region",
						},
						{
							Column: "products.name",
							Alias:  "product",
						},
						{
							Column:   "orders.created_date",
							Interval: "month",
							Alias:    "month",
						},
					},
					Metrics: []tablebuilder.MetricConfig{
						{
							Name:     "revenue",
							Function: "sum",
							Column:   "order_line_items.total_price",
						},
					},
				},
			},
		}

		if len(config.DataSource[0].GroupBy) != 3 {
			t.Errorf("expected 3 GroupBy, got %d", len(config.DataSource[0].GroupBy))
		}

		// Verify all three dimensions
		aliases := []string{"region", "product", "month"}
		for i, gb := range config.DataSource[0].GroupBy {
			if gb.Alias != aliases[i] {
				t.Errorf("GroupBy %d alias mismatch: got %s, want %s", i, gb.Alias, aliases[i])
			}
		}

		// Validate each GroupBy config is individually valid.
		for i, gb := range config.DataSource[0].GroupBy {
			gbCopy := gb
			if err := tablebuilder.ValidateGroupByConfig(&gbCopy); err != nil {
				t.Errorf("GroupBy[%d] (%q) failed validation: %v", i, gb.Alias, err)
			}
		}
	})

	t.Run("complex SQL expressions in GroupBy", func(t *testing.T) {
		t.Parallel()
		config := &tablebuilder.Config{
			Title:         "Complex Grouping",
			WidgetType:    "chart",
			Visualization: "bar",
			DataSource: []tablebuilder.DataSource{
				{
					Type:   "query",
					Source: "orders",
					Schema: "sales",
					GroupBy: []tablebuilder.GroupByConfig{
						{
							Column:     "CASE WHEN orders.amount > 1000 THEN 'High' WHEN orders.amount > 500 THEN 'Medium' ELSE 'Low' END",
							Alias:      "value_tier",
							Expression: true,
						},
						{
							Column:     "DATE_PART('quarter', orders.created_date)",
							Alias:      "quarter",
							Expression: true,
						},
					},
					Metrics: []tablebuilder.MetricConfig{
						{
							Name:     "count",
							Function: "count",
							Column:   "orders.id",
						},
					},
				},
			},
		}

		if len(config.DataSource[0].GroupBy) != 2 {
			t.Errorf("expected 2 GroupBy, got %d", len(config.DataSource[0].GroupBy))
		}

		// Both should be expressions
		for i, gb := range config.DataSource[0].GroupBy {
			if !gb.Expression {
				t.Errorf("GroupBy %d should be marked as Expression", i)
			}
			if gb.Alias == "" {
				t.Errorf("GroupBy %d is missing required alias", i)
			}
		}
	})
}

// TestGroupByValidation tests the validation logic for GroupBy configs
func TestGroupByValidation(t *testing.T) {
	t.Parallel()

	t.Run("validates interval values", func(t *testing.T) {
		t.Parallel()
		validIntervals := []string{"day", "week", "month", "quarter", "year"}

		for _, interval := range validIntervals {
			groupBy := &tablebuilder.GroupByConfig{
				Column:   "orders.created_date",
				Interval: interval,
				Alias:    "period",
			}

			err := tablebuilder.ValidateGroupByConfig(groupBy)
			if err != nil {
				t.Errorf("valid interval %q should not produce error: %v", interval, err)
			}
		}
	})

	t.Run("rejects invalid interval", func(t *testing.T) {
		t.Parallel()
		groupBy := &tablebuilder.GroupByConfig{
			Column:   "orders.created_date",
			Interval: "invalid_interval",
			Alias:    "period",
		}

		err := tablebuilder.ValidateGroupByConfig(groupBy)
		if err == nil {
			t.Error("invalid interval should produce error")
		}
	})

	t.Run("validates expression requires alias", func(t *testing.T) {
		t.Parallel()
		groupBy := &tablebuilder.GroupByConfig{
			Column:     "EXTRACT(DOW FROM orders.created_date)",
			Expression: true,
			// Missing alias
		}

		// ValidateGroupByConfig is permissive on alias; alias enforcement happens at BuildQuery time.
		if err := tablebuilder.ValidateGroupByConfig(groupBy); err != nil {
			t.Errorf("ValidateGroupByConfig should allow expression without alias (enforced at query build time): %v", err)
		}
	})

	t.Run("allows simple column without alias", func(t *testing.T) {
		t.Parallel()
		groupBy := &tablebuilder.GroupByConfig{
			Column: "products.name",
			// No alias - should use column name
		}

		err := tablebuilder.ValidateGroupByConfig(groupBy)
		if err != nil {
			t.Errorf("simple column without alias should be valid: %v", err)
		}
	})
}

// TestGroupByJSONSerialization tests that the new GroupBy format serializes correctly
func TestGroupByJSONSerialization(t *testing.T) {
	t.Parallel()

	t.Run("serializes single GroupBy", func(t *testing.T) {
		ds := tablebuilder.DataSource{
			Type:   "query",
			Source: "orders",
			Schema: "sales",
			GroupBy: []tablebuilder.GroupByConfig{
				{
					Column:   "orders.created_date",
					Interval: "month",
					Alias:    "month",
				},
			},
		}

		// Should serialize as an array
		if len(ds.GroupBy) != 1 {
			t.Errorf("expected 1 GroupBy, got %d", len(ds.GroupBy))
		}
	})

	t.Run("serializes multiple GroupBy", func(t *testing.T) {
		ds := tablebuilder.DataSource{
			Type:   "query",
			Source: "orders",
			Schema: "sales",
			GroupBy: []tablebuilder.GroupByConfig{
				{
					Column:     "EXTRACT(DOW FROM orders.created_date)",
					Alias:      "day",
					Expression: true,
				},
				{
					Column:     "EXTRACT(HOUR FROM orders.created_date)",
					Alias:      "hour",
					Expression: true,
				},
			},
		}

		if len(ds.GroupBy) != 2 {
			t.Errorf("expected 2 GroupBy, got %d", len(ds.GroupBy))
		}

		// Verify Expression flags are preserved
		if !ds.GroupBy[0].Expression || !ds.GroupBy[1].Expression {
			t.Error("Expression flags should be preserved")
		}
	})

	t.Run("serializes empty GroupBy slice", func(t *testing.T) {
		ds := tablebuilder.DataSource{
			Type:    "query",
			Source:  "orders",
			Schema:  "sales",
			GroupBy: []tablebuilder.GroupByConfig{},
		}

		if ds.GroupBy == nil {
			t.Error("GroupBy should be empty slice, not nil")
		}
		if len(ds.GroupBy) != 0 {
			t.Errorf("expected 0 GroupBy, got %d", len(ds.GroupBy))
		}
	})
}

// TestMultiGroupBy_SQLGeneration verifies that the query builder emits correct SQL
// for various GroupBy configurations. No database required — pure SQL generation.
func TestMultiGroupBy_SQLGeneration(t *testing.T) {
	t.Parallel()

	t.Run("single_groupby_date_trunc", func(t *testing.T) {
		t.Parallel()
		qb := tablebuilder.NewQueryBuilder()
		ds := tablebuilder.DataSource{
			Type:   "query",
			Source: "orders",
			Schema: "sales",
			GroupBy: []tablebuilder.GroupByConfig{
				{Column: "orders.created_date", Interval: "month", Alias: "month"},
			},
			Metrics: []tablebuilder.MetricConfig{
				{Name: "revenue", Function: "sum", Column: "orders.amount"},
			},
		}
		sql, _, err := qb.BuildQuery(&ds, tablebuilder.QueryParams{}, true)
		if err != nil {
			t.Fatalf("BuildQuery: %v", err)
		}
		if !strings.Contains(sql, "DATE_TRUNC('month'") {
			t.Errorf("SQL missing DATE_TRUNC('month')\nFull SQL: %s", sql)
		}
		if !strings.Contains(sql, "GROUP BY") {
			t.Errorf("SQL missing GROUP BY\nFull SQL: %s", sql)
		}
		if !strings.Contains(sql, `SUM(orders.amount)`) {
			t.Errorf("SQL missing SUM aggregate\nFull SQL: %s", sql)
		}
	})

	t.Run("multiple_groupby_categorical", func(t *testing.T) {
		t.Parallel()
		qb := tablebuilder.NewQueryBuilder()
		ds := tablebuilder.DataSource{
			Type:   "query",
			Source: "orders",
			Schema: "sales",
			GroupBy: []tablebuilder.GroupByConfig{
				{Column: "orders.region", Alias: "region"},
				{Column: "orders.status", Alias: "status"},
			},
			Metrics: []tablebuilder.MetricConfig{
				{Name: "count", Function: "count", Column: "orders.id"},
			},
		}
		sql, _, err := qb.BuildQuery(&ds, tablebuilder.QueryParams{}, true)
		if err != nil {
			t.Fatalf("BuildQuery: %v", err)
		}
		if !strings.Contains(sql, "orders.region") {
			t.Errorf("SQL missing orders.region\nFull SQL: %s", sql)
		}
		if !strings.Contains(sql, "orders.status") {
			t.Errorf("SQL missing orders.status\nFull SQL: %s", sql)
		}
		if !strings.Contains(sql, "GROUP BY") {
			t.Errorf("SQL missing GROUP BY\nFull SQL: %s", sql)
		}
	})

	t.Run("expression_groupby_raw_sql_in_select_and_group_by", func(t *testing.T) {
		t.Parallel()
		qb := tablebuilder.NewQueryBuilder()
		ds := tablebuilder.DataSource{
			Type:   "query",
			Source: "orders",
			Schema: "sales",
			GroupBy: []tablebuilder.GroupByConfig{
				{Column: "EXTRACT(DOW FROM orders.created_date)", Expression: true, Alias: "day_of_week"},
			},
			Metrics: []tablebuilder.MetricConfig{
				{Name: "count", Function: "count", Column: "orders.id"},
			},
		}
		sql, _, err := qb.BuildQuery(&ds, tablebuilder.QueryParams{}, true)
		if err != nil {
			t.Fatalf("BuildQuery: %v", err)
		}
		if !strings.Contains(sql, "EXTRACT(DOW FROM orders.created_date)") {
			t.Errorf("SQL missing raw expression\nFull SQL: %s", sql)
		}
		if !strings.Contains(sql, "GROUP BY") {
			t.Errorf("SQL missing GROUP BY\nFull SQL: %s", sql)
		}
	})

	t.Run("mixed_interval_and_expression", func(t *testing.T) {
		t.Parallel()
		qb := tablebuilder.NewQueryBuilder()
		ds := tablebuilder.DataSource{
			Type:   "query",
			Source: "orders",
			Schema: "sales",
			GroupBy: []tablebuilder.GroupByConfig{
				{Column: "orders.created_date", Interval: "quarter", Alias: "quarter"},
				{Column: "orders.region", Alias: "region"},
			},
			Metrics: []tablebuilder.MetricConfig{
				{Name: "total", Function: "sum", Column: "orders.amount"},
			},
		}
		sql, _, err := qb.BuildQuery(&ds, tablebuilder.QueryParams{}, true)
		if err != nil {
			t.Fatalf("BuildQuery: %v", err)
		}
		if !strings.Contains(sql, "DATE_TRUNC('quarter'") {
			t.Errorf("SQL missing DATE_TRUNC('quarter')\nFull SQL: %s", sql)
		}
		if !strings.Contains(sql, "orders.region") {
			t.Errorf("SQL missing orders.region\nFull SQL: %s", sql)
		}
		if !strings.Contains(sql, "GROUP BY") {
			t.Errorf("SQL missing GROUP BY\nFull SQL: %s", sql)
		}
	})

	t.Run("no_groupby_produces_no_group_clause", func(t *testing.T) {
		t.Parallel()
		qb := tablebuilder.NewQueryBuilder()
		ds := tablebuilder.DataSource{
			Type:   "query",
			Source: "orders",
			Schema: "sales",
			Metrics: []tablebuilder.MetricConfig{
				{Name: "total", Function: "sum", Column: "orders.amount"},
			},
		}
		sql, _, err := qb.BuildQuery(&ds, tablebuilder.QueryParams{}, true)
		if err != nil {
			t.Fatalf("BuildQuery: %v", err)
		}
		if strings.Contains(sql, "GROUP BY") {
			t.Errorf("SQL should not contain GROUP BY when no GroupBy configured\nFull SQL: %s", sql)
		}
	})
}

// TestMultiGroupBy_ErrorCases verifies that invalid GroupBy configurations return errors.
func TestMultiGroupBy_ErrorCases(t *testing.T) {
	t.Parallel()

	t.Run("invalid_interval_returns_error", func(t *testing.T) {
		t.Parallel()
		qb := tablebuilder.NewQueryBuilder()
		ds := tablebuilder.DataSource{
			Type:   "query",
			Source: "orders",
			Schema: "sales",
			GroupBy: []tablebuilder.GroupByConfig{
				{Column: "orders.created_date", Interval: "fortnight", Alias: "period"},
			},
			Metrics: []tablebuilder.MetricConfig{
				{Name: "count", Function: "count", Column: "orders.id"},
			},
		}
		_, _, err := qb.BuildQuery(&ds, tablebuilder.QueryParams{}, true)
		if err == nil {
			t.Error("expected error for invalid interval 'fortnight', got nil")
		}
	})

	t.Run("expression_missing_alias_returns_error", func(t *testing.T) {
		t.Parallel()
		qb := tablebuilder.NewQueryBuilder()
		ds := tablebuilder.DataSource{
			Type:   "query",
			Source: "orders",
			Schema: "sales",
			GroupBy: []tablebuilder.GroupByConfig{
				{Column: "EXTRACT(DOW FROM orders.created_date)", Expression: true},
				// no Alias — should be rejected
			},
			Metrics: []tablebuilder.MetricConfig{
				{Name: "count", Function: "count", Column: "orders.id"},
			},
		}
		_, _, err := qb.BuildQuery(&ds, tablebuilder.QueryParams{}, true)
		if err == nil {
			t.Error("expected error for expression GroupBy missing alias, got nil")
		}
	})
}
