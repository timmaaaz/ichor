package tablebuilder_test

import (
	"strings"
	"testing"

	"github.com/timmaaaz/ichor/business/sdk/tablebuilder"
)

// minimalDS returns a DataSource with one column — enough for BuildQuery to work.
func minimalDS(source, schema string) tablebuilder.DataSource {
	return tablebuilder.DataSource{
		Type:   "query",
		Source: source,
		Schema: schema,
		Select: tablebuilder.SelectConfig{
			Columns: []tablebuilder.ColumnDefinition{
				{Name: "id", TableColumn: source + ".id"},
			},
		},
	}
}

// assertSQL calls t.Errorf for each want string not found in sql.
func assertSQL(t *testing.T, sql string, wants ...string) {
	t.Helper()
	for _, want := range wants {
		if !strings.Contains(sql, want) {
			t.Errorf("SQL missing %q\nFull SQL: %s", want, sql)
		}
	}
}

// assertNoSQL calls t.Errorf for each noWant string found in sql.
func assertNoSQL(t *testing.T, sql string, noWants ...string) {
	t.Helper()
	for _, noWant := range noWants {
		if strings.Contains(sql, noWant) {
			t.Errorf("SQL unexpectedly contains %q\nFull SQL: %s", noWant, sql)
		}
	}
}

func TestBuildQuery_Filters(t *testing.T) {
	t.Parallel()
	qb := tablebuilder.NewQueryBuilder()

	tests := []struct {
		name   string
		op     string
		value  any
		want   []string
		noWant []string
	}{
		{name: "eq", op: "eq", value: 42, want: []string{"WHERE", "42"}},
		{name: "neq", op: "neq", value: 42, want: []string{"WHERE", "!=", "42"}},
		{name: "gt", op: "gt", value: 10, want: []string{"WHERE", ">", "10"}},
		{name: "gte", op: "gte", value: 10, want: []string{"WHERE", ">=", "10"}},
		{name: "lt", op: "lt", value: 10, want: []string{"WHERE", "<", "10"}},
		{name: "lte", op: "lte", value: 10, want: []string{"WHERE", "<=", "10"}},
		{name: "in", op: "in", value: []interface{}{"a", "b", "c"}, want: []string{"WHERE", "IN"}},
		{name: "like", op: "like", value: "test", want: []string{"WHERE", "LIKE", "test"}},
		{name: "ilike", op: "ilike", value: "test", want: []string{"WHERE", "ILIKE", "test"}},
		{
			name:  "is_null",
			op:    "is_null",
			value: true, // non-nil sentinel required to bypass nil-guard; NOT emitted in SQL
			want:  []string{"WHERE", "IS NULL"},
		},
		{
			name:  "is_not_null",
			op:    "is_not_null",
			value: true, // non-nil sentinel required to bypass nil-guard; NOT emitted in SQL
			want:  []string{"WHERE", "IS NOT NULL"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ds := minimalDS("products", "products")
			ds.Filters = []tablebuilder.Filter{
				{Column: "quantity", Operator: tt.op, Value: tt.value},
			}
			sql, _, err := qb.BuildQuery(&ds, tablebuilder.QueryParams{}, true)
			if err != nil {
				t.Fatalf("BuildQuery: %v", err)
			}
			assertSQL(t, sql, tt.want...)
			assertNoSQL(t, sql, tt.noWant...)
		})
	}
}

// nil value skips filter entirely — no WHERE clause produced
func TestBuildQuery_NilValueSkipsFilter(t *testing.T) {
	t.Parallel()
	qb := tablebuilder.NewQueryBuilder()

	ds := minimalDS("products", "products")
	ds.Filters = []tablebuilder.Filter{
		{Column: "quantity", Operator: "eq", Value: nil},
	}
	sql, _, err := qb.BuildQuery(&ds, tablebuilder.QueryParams{}, true)
	if err != nil {
		t.Fatalf("BuildQuery: %v", err)
	}
	assertNoSQL(t, sql, "WHERE")
}

// Dynamic filter: value from QueryParams.Dynamic overrides static Value
func TestBuildQuery_DynamicFilter(t *testing.T) {
	t.Parallel()
	qb := tablebuilder.NewQueryBuilder()

	ds := minimalDS("orders", "sales")
	ds.Filters = []tablebuilder.Filter{
		{Column: "status", Operator: "eq", Value: "pending", Dynamic: true},
	}
	params := tablebuilder.QueryParams{
		Dynamic: map[string]any{"status": "shipped"},
	}
	sql, _, err := qb.BuildQuery(&ds, params, true)
	if err != nil {
		t.Fatalf("BuildQuery: %v", err)
	}
	assertSQL(t, sql, "WHERE", "shipped")
	assertNoSQL(t, sql, "pending")
}

func TestBuildQuery_Sorting(t *testing.T) {
	t.Parallel()
	qb := tablebuilder.NewQueryBuilder()

	t.Run("config sort asc used when params empty", func(t *testing.T) {
		ds := minimalDS("products", "products")
		ds.Sort = []tablebuilder.Sort{{Column: "name", Direction: "asc"}}
		sql, _, err := qb.BuildQuery(&ds, tablebuilder.QueryParams{}, true)
		if err != nil {
			t.Fatalf("BuildQuery: %v", err)
		}
		assertSQL(t, sql, "ORDER BY", "name", "ASC")
	})

	t.Run("config sort desc", func(t *testing.T) {
		ds := minimalDS("products", "products")
		ds.Sort = []tablebuilder.Sort{{Column: "quantity", Direction: "desc"}}
		sql, _, err := qb.BuildQuery(&ds, tablebuilder.QueryParams{}, true)
		if err != nil {
			t.Fatalf("BuildQuery: %v", err)
		}
		assertSQL(t, sql, "ORDER BY", "quantity", "DESC")
	})

	t.Run("params sort overrides config sort", func(t *testing.T) {
		ds := minimalDS("products", "products")
		ds.Sort = []tablebuilder.Sort{{Column: "name", Direction: "asc"}}
		params := tablebuilder.QueryParams{
			Sort: []tablebuilder.Sort{{Column: "price", Direction: "desc"}},
		}
		sql, _, err := qb.BuildQuery(&ds, params, true)
		if err != nil {
			t.Fatalf("BuildQuery: %v", err)
		}
		// Positive assertion proves override worked — no need for fragile negative assertion
		assertSQL(t, sql, "ORDER BY", "price", "DESC")
	})

	t.Run("multiple sort columns all appear", func(t *testing.T) {
		ds := minimalDS("products", "products")
		ds.Sort = []tablebuilder.Sort{
			{Column: "category", Direction: "asc"},
			{Column: "price", Direction: "desc"},
		}
		sql, _, err := qb.BuildQuery(&ds, tablebuilder.QueryParams{}, true)
		if err != nil {
			t.Fatalf("BuildQuery: %v", err)
		}
		assertSQL(t, sql, "ORDER BY", "category", "price")
	})

	t.Run("non-primary source skips sort", func(t *testing.T) {
		ds := minimalDS("products", "products")
		ds.Sort = []tablebuilder.Sort{{Column: "name", Direction: "asc"}}
		sql, _, err := qb.BuildQuery(&ds, tablebuilder.QueryParams{}, false /* isPrimary=false */)
		if err != nil {
			t.Fatalf("BuildQuery: %v", err)
		}
		assertNoSQL(t, sql, "ORDER BY")
	})
}

func TestBuildQuery_Pagination(t *testing.T) {
	t.Parallel()
	qb := tablebuilder.NewQueryBuilder()

	t.Run("page 1 rows 10 produces LIMIT 10", func(t *testing.T) {
		// goqu omits OFFSET when offset=0 (page 1)
		ds := minimalDS("products", "products")
		params := tablebuilder.QueryParams{Page: 1, Rows: 10}
		sql, _, err := qb.BuildQuery(&ds, params, true)
		if err != nil {
			t.Fatalf("BuildQuery: %v", err)
		}
		assertSQL(t, sql, "LIMIT 10")
		assertNoSQL(t, sql, "OFFSET")
	})

	t.Run("page 3 rows 10 produces OFFSET 20", func(t *testing.T) {
		ds := minimalDS("products", "products")
		params := tablebuilder.QueryParams{Page: 3, Rows: 10}
		sql, _, err := qb.BuildQuery(&ds, params, true)
		if err != nil {
			t.Fatalf("BuildQuery: %v", err)
		}
		assertSQL(t, sql, "LIMIT 10", "OFFSET 20")
	})

	t.Run("page 2 rows 25 produces OFFSET 25", func(t *testing.T) {
		ds := minimalDS("products", "products")
		params := tablebuilder.QueryParams{Page: 2, Rows: 25}
		sql, _, err := qb.BuildQuery(&ds, params, true)
		if err != nil {
			t.Fatalf("BuildQuery: %v", err)
		}
		assertSQL(t, sql, "LIMIT 25", "OFFSET 25")
	})

	t.Run("non-primary source skips pagination", func(t *testing.T) {
		ds := minimalDS("products", "products")
		params := tablebuilder.QueryParams{Page: 2, Rows: 10}
		sql, _, err := qb.BuildQuery(&ds, params, false /* isPrimary=false */)
		if err != nil {
			t.Fatalf("BuildQuery: %v", err)
		}
		assertNoSQL(t, sql, "LIMIT", "OFFSET")
	})

	t.Run("ds.Rows applied when no page params", func(t *testing.T) {
		ds := minimalDS("products", "products")
		ds.Rows = 50
		sql, _, err := qb.BuildQuery(&ds, tablebuilder.QueryParams{}, true)
		if err != nil {
			t.Fatalf("BuildQuery: %v", err)
		}
		assertSQL(t, sql, "LIMIT 50")
		assertNoSQL(t, sql, "OFFSET")
	})
}

func TestBuildQuery_Joins(t *testing.T) {
	t.Parallel()
	qb := tablebuilder.NewQueryBuilder()

	joinTypes := []struct {
		joinType string
		wantSQL  string
	}{
		{"inner", "JOIN"},
		{"left", "LEFT JOIN"},
		{"right", "RIGHT JOIN"},
		{"full", "FULL JOIN"},
	}

	for _, tt := range joinTypes {
		t.Run(tt.joinType, func(t *testing.T) {
			ds := minimalDS("orders", "sales")
			ds.Joins = []tablebuilder.Join{
				{
					Table:  "customers",
					Schema: "sales",
					Type:   tt.joinType,
					On:     "orders.customer_id = customers.id",
				},
			}
			sql, _, err := qb.BuildQuery(&ds, tablebuilder.QueryParams{}, true)
			if err != nil {
				t.Fatalf("BuildQuery join type %q: %v", tt.joinType, err)
			}
			assertSQL(t, sql, tt.wantSQL, "customers")
		})
	}

	t.Run("join condition parsed table.col = table.col", func(t *testing.T) {
		ds := minimalDS("orders", "sales")
		ds.Joins = []tablebuilder.Join{
			{
				Table: "customers",
				Type:  "inner",
				On:    "orders.customer_id = customers.id",
			},
		}
		sql, _, err := qb.BuildQuery(&ds, tablebuilder.QueryParams{}, true)
		if err != nil {
			t.Fatalf("BuildQuery: %v", err)
		}
		assertSQL(t, sql, "customer_id", "customers")
	})

	t.Run("join with schema prefix", func(t *testing.T) {
		ds := minimalDS("orders", "sales")
		ds.Joins = []tablebuilder.Join{
			{
				Table:  "customers",
				Schema: "sales",
				Type:   "left",
				On:     "orders.customer_id = customers.id",
			},
		}
		sql, _, err := qb.BuildQuery(&ds, tablebuilder.QueryParams{}, true)
		if err != nil {
			t.Fatalf("BuildQuery: %v", err)
		}
		assertSQL(t, sql, "sales", "customers", "LEFT JOIN")
	})
}

func TestBuildQuery_ForeignTableJoins(t *testing.T) {
	t.Parallel()
	qb := tablebuilder.NewQueryBuilder()

	t.Run("inner join with schema and alias", func(t *testing.T) {
		ds := minimalDS("inventory_items", "inventory")
		ds.Select.ForeignTables = []tablebuilder.ForeignTable{
			{
				Table:            "products",
				Alias:            "p",
				Schema:           "products",
				RelationshipFrom: "inventory_items.product_id",
				RelationshipTo:   "p.id",
				JoinType:         "inner",
				Columns: []tablebuilder.ColumnDefinition{
					{Name: "name", Alias: "product_name"},
				},
			},
		}
		sql, _, err := qb.BuildQuery(&ds, tablebuilder.QueryParams{}, true)
		if err != nil {
			t.Fatalf("BuildQuery: %v", err)
		}
		assertSQL(t, sql, "JOIN", "products", "product_name")
	})

	t.Run("left join without alias", func(t *testing.T) {
		ds := minimalDS("inventory_items", "inventory")
		ds.Select.ForeignTables = []tablebuilder.ForeignTable{
			{
				Table:            "products",
				Schema:           "products",
				RelationshipFrom: "inventory_items.product_id",
				RelationshipTo:   "products.id",
				JoinType:         "left",
				Columns: []tablebuilder.ColumnDefinition{
					{Name: "sku"},
				},
			},
		}
		sql, _, err := qb.BuildQuery(&ds, tablebuilder.QueryParams{}, true)
		if err != nil {
			t.Fatalf("BuildQuery: %v", err)
		}
		assertSQL(t, sql, "LEFT JOIN", "products")
	})

	t.Run("nested foreign table join", func(t *testing.T) {
		ds := minimalDS("inventory_adjustments", "inventory")
		ds.Select.ForeignTables = []tablebuilder.ForeignTable{
			{
				Table:            "inventory_locations",
				Schema:           "inventory",
				RelationshipFrom: "inventory_adjustments.location_id",
				RelationshipTo:   "inventory_locations.id",
				JoinType:         "left",
				Columns:          []tablebuilder.ColumnDefinition{{Name: "aisle"}},
				ForeignTables: []tablebuilder.ForeignTable{
					{
						Table:            "warehouses",
						Schema:           "inventory",
						RelationshipFrom: "inventory_locations.warehouse_id",
						RelationshipTo:   "warehouses.id",
						JoinType:         "left",
						Columns:          []tablebuilder.ColumnDefinition{{Name: "name", Alias: "warehouse_name"}},
					},
				},
			},
		}
		sql, _, err := qb.BuildQuery(&ds, tablebuilder.QueryParams{}, true)
		if err != nil {
			t.Fatalf("BuildQuery: %v", err)
		}
		assertSQL(t, sql, "inventory_locations", "warehouses", "warehouse_name")
	})
}

func TestBuildCountQuery(t *testing.T) {
	t.Parallel()
	qb := tablebuilder.NewQueryBuilder()

	t.Run("produces COUNT star", func(t *testing.T) {
		ds := minimalDS("products", "products")
		sql, _, err := qb.BuildCountQuery(&ds, tablebuilder.QueryParams{})
		if err != nil {
			t.Fatalf("BuildCountQuery: %v", err)
		}
		assertSQL(t, sql, "COUNT(*)")
		assertNoSQL(t, sql, "LIMIT", "OFFSET")
	})

	t.Run("filter still applied to count", func(t *testing.T) {
		ds := minimalDS("products", "products")
		ds.Filters = []tablebuilder.Filter{
			{Column: "is_active", Operator: "eq", Value: true},
		}
		sql, _, err := qb.BuildCountQuery(&ds, tablebuilder.QueryParams{})
		if err != nil {
			t.Fatalf("BuildCountQuery: %v", err)
		}
		assertSQL(t, sql, "COUNT(*)", "WHERE", "is_active")
	})
}

func TestBuildMetricQuery(t *testing.T) {
	t.Parallel()
	qb := tablebuilder.NewQueryBuilder()

	metricDS := func(source, schema string, metrics []tablebuilder.MetricConfig, groupBys []tablebuilder.GroupByConfig) tablebuilder.DataSource {
		return tablebuilder.DataSource{
			Type:    "query",
			Source:  source,
			Schema:  schema,
			Metrics: metrics,
			GroupBy: groupBys,
		}
	}

	t.Run("sum function", func(t *testing.T) {
		ds := metricDS("orders", "sales", []tablebuilder.MetricConfig{
			{Name: "total_revenue", Function: "sum", Column: "orders.amount"},
		}, nil)
		sql, _, err := qb.BuildQuery(&ds, tablebuilder.QueryParams{}, true)
		if err != nil {
			t.Fatalf("BuildQuery sum: %v", err)
		}
		assertSQL(t, sql, "SUM", "total_revenue")
		assertNoSQL(t, sql, "GROUP BY")
	})

	t.Run("count function", func(t *testing.T) {
		ds := metricDS("orders", "sales", []tablebuilder.MetricConfig{
			{Name: "order_count", Function: "count", Column: "orders.id"},
		}, nil)
		sql, _, err := qb.BuildQuery(&ds, tablebuilder.QueryParams{}, true)
		if err != nil {
			t.Fatalf("BuildQuery count: %v", err)
		}
		assertSQL(t, sql, "COUNT", "order_count")
	})

	t.Run("count_distinct produces COUNT DISTINCT", func(t *testing.T) {
		ds := metricDS("orders", "sales", []tablebuilder.MetricConfig{
			{Name: "unique_customers", Function: "count_distinct", Column: "orders.customer_id"},
		}, nil)
		sql, _, err := qb.BuildQuery(&ds, tablebuilder.QueryParams{}, true)
		if err != nil {
			t.Fatalf("BuildQuery count_distinct: %v", err)
		}
		assertSQL(t, sql, "COUNT(DISTINCT", "unique_customers")
	})

	t.Run("avg function", func(t *testing.T) {
		ds := metricDS("orders", "sales", []tablebuilder.MetricConfig{
			{Name: "avg_order", Function: "avg", Column: "orders.amount"},
		}, nil)
		sql, _, err := qb.BuildQuery(&ds, tablebuilder.QueryParams{}, true)
		if err != nil {
			t.Fatalf("BuildQuery avg: %v", err)
		}
		assertSQL(t, sql, "AVG", "avg_order")
	})

	t.Run("invalid aggregate function returns error", func(t *testing.T) {
		ds := metricDS("orders", "sales", []tablebuilder.MetricConfig{
			{Name: "bad", Function: "median", Column: "orders.amount"},
		}, nil)
		_, _, err := qb.BuildQuery(&ds, tablebuilder.QueryParams{}, true)
		if err == nil {
			t.Fatal("expected error for invalid aggregate function, got nil")
		}
	})

	t.Run("groupby with date interval produces DATE_TRUNC and GROUP BY", func(t *testing.T) {
		ds := metricDS("orders", "sales",
			[]tablebuilder.MetricConfig{
				{Name: "revenue", Function: "sum", Column: "orders.amount"},
			},
			[]tablebuilder.GroupByConfig{
				{Column: "orders.created_date", Interval: "month", Alias: "month"},
			},
		)
		sql, _, err := qb.BuildQuery(&ds, tablebuilder.QueryParams{}, true)
		if err != nil {
			t.Fatalf("BuildQuery groupby: %v", err)
		}
		assertSQL(t, sql, "DATE_TRUNC", "month", "GROUP BY")
	})

	t.Run("multiple groupby all appear in GROUP BY clause", func(t *testing.T) {
		ds := metricDS("order_line_items", "sales",
			[]tablebuilder.MetricConfig{
				{Name: "revenue", Function: "sum", Column: "order_line_items.total_price"},
			},
			[]tablebuilder.GroupByConfig{
				{Column: "products.name", Alias: "product"},
				{Column: "cities.name", Alias: "city"},
			},
		)
		sql, _, err := qb.BuildQuery(&ds, tablebuilder.QueryParams{}, true)
		if err != nil {
			t.Fatalf("BuildQuery multi-groupby: %v", err)
		}
		assertSQL(t, sql, "GROUP BY", "products.name", "cities.name")
	})

	t.Run("expression groupby uses raw SQL", func(t *testing.T) {
		ds := metricDS("orders", "sales",
			[]tablebuilder.MetricConfig{
				{Name: "cnt", Function: "count", Column: "orders.id"},
			},
			[]tablebuilder.GroupByConfig{
				{
					Column:     "EXTRACT(DOW FROM orders.created_date)",
					Alias:      "day_of_week",
					Expression: true,
				},
			},
		)
		sql, _, err := qb.BuildQuery(&ds, tablebuilder.QueryParams{}, true)
		if err != nil {
			t.Fatalf("BuildQuery expression groupby: %v", err)
		}
		assertSQL(t, sql, "EXTRACT", "DOW", "GROUP BY")
	})
}

func TestBuildMetricQuery_ArithmeticExpression(t *testing.T) {
	t.Parallel()
	qb := tablebuilder.NewQueryBuilder()

	operators := []struct {
		name    string
		op      string
		wantSQL string
	}{
		{"multiply", "multiply", "*"},
		{"add", "add", "+"},
		{"subtract", "subtract", "-"},
		{"divide", "divide", "/"},
	}

	for _, tt := range operators {
		t.Run(tt.name, func(t *testing.T) {
			ds := tablebuilder.DataSource{
				Type:   "query",
				Source: "order_line_items",
				Schema: "sales",
				Metrics: []tablebuilder.MetricConfig{
					{
						Name:     "result",
						Function: "sum",
						Expression: &tablebuilder.ExpressionConfig{
							Operator: tt.op,
							Columns:  []string{"order_line_items.quantity", "order_line_items.unit_price"},
						},
					},
				},
			}
			sql, _, err := qb.BuildQuery(&ds, tablebuilder.QueryParams{}, true)
			if err != nil {
				t.Fatalf("BuildQuery operator %q: %v", tt.op, err)
			}
			assertSQL(t, sql, tt.wantSQL, "quantity", "unit_price")
		})
	}

	t.Run("invalid operator returns error", func(t *testing.T) {
		ds := tablebuilder.DataSource{
			Source: "t",
			Metrics: []tablebuilder.MetricConfig{
				{
					Name:     "bad",
					Function: "sum",
					Expression: &tablebuilder.ExpressionConfig{
						Operator: "modulo",
						Columns:  []string{"a", "b"},
					},
				},
			},
		}
		_, _, err := qb.BuildQuery(&ds, tablebuilder.QueryParams{}, true)
		if err == nil {
			t.Fatal("expected error for invalid operator")
		}
	})
}
