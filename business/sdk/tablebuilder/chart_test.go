package tablebuilder_test

import (
	"math"
	"testing"

	"github.com/timmaaaz/ichor/business/sdk/tablebuilder"
)

// makeConfig returns a minimal Config that routes Transform to the given chart type.
func makeConfig(chartType, title string) *tablebuilder.Config {
	return &tablebuilder.Config{
		Title:         title,
		WidgetType:    "chart",
		Visualization: chartType,
		DataSource: []tablebuilder.DataSource{
			{Source: "orders", Schema: "sales"},
		},
	}
}

// makeTableData creates a TableData with the given rows.
func makeTableData(rows ...tablebuilder.TableRow) *tablebuilder.TableData {
	return &tablebuilder.TableData{Data: rows}
}

// makeTableDataWithMeta creates TableData with ColumnMetadata for deterministic column ordering.
// columns controls the field order that detectCategoricalColumns uses when metadata is present.
func makeTableDataWithMeta(columns []string, rows ...tablebuilder.TableRow) *tablebuilder.TableData {
	colMeta := make([]tablebuilder.ColumnMetadata, len(columns))
	for i, c := range columns {
		colMeta[i] = tablebuilder.ColumnMetadata{Field: c}
	}
	return &tablebuilder.TableData{
		Data: rows,
		Meta: tablebuilder.MetaData{Columns: colMeta},
	}
}

func TestChartTransformer_NilDataReturnsError(t *testing.T) {
	t.Parallel()
	ct := tablebuilder.NewChartTransformer()
	_, err := ct.Transform(nil, makeConfig("bar", "Test"))
	if err == nil {
		t.Fatal("expected error for nil data, got nil")
	}
}

func TestChartTransformer_UnknownChartTypeReturnsError(t *testing.T) {
	t.Parallel()
	ct := tablebuilder.NewChartTransformer()
	_, err := ct.Transform(makeTableData(), makeConfig("unknown_type", "Test"))
	if err == nil {
		t.Fatal("expected error for unknown chart type, got nil")
	}
}

func TestChartTransformer_KPI(t *testing.T) {
	t.Parallel()
	ct := tablebuilder.NewChartTransformer()

	t.Run("single row returns value", func(t *testing.T) {
		t.Parallel()
		data := makeTableData(tablebuilder.TableRow{"revenue": 125000.50})
		cfg := makeConfig(tablebuilder.ChartTypeKPI, "Total Revenue")
		resp, err := ct.Transform(data, cfg)
		if err != nil {
			t.Fatalf("Transform: %v", err)
		}
		if resp.KPI == nil {
			t.Fatal("expected KPI data, got nil")
		}
		if resp.KPI.Value != 125000.50 {
			t.Errorf("KPI.Value = %v, want 125000.50", resp.KPI.Value)
		}
	})

	t.Run("two rows calculates trend up", func(t *testing.T) {
		t.Parallel()
		data := makeTableData(
			tablebuilder.TableRow{"revenue": 125000.0},
			tablebuilder.TableRow{"revenue": 100000.0},
		)
		cfg := makeConfig(tablebuilder.ChartTypeKPI, "Revenue Trend")
		resp, err := ct.Transform(data, cfg)
		if err != nil {
			t.Fatalf("Transform: %v", err)
		}
		if resp.KPI.Trend != "up" {
			t.Errorf("Trend = %q, want %q", resp.KPI.Trend, "up")
		}
		// Change = ((125000 - 100000) / 100000) * 100 = 25%
		if math.Abs(resp.KPI.Change-25.0) > 0.01 {
			t.Errorf("Change = %v, want ~25.0", resp.KPI.Change)
		}
	})

	t.Run("two rows calculates trend down", func(t *testing.T) {
		t.Parallel()
		data := makeTableData(
			tablebuilder.TableRow{"revenue": 80000.0},
			tablebuilder.TableRow{"revenue": 100000.0},
		)
		cfg := makeConfig(tablebuilder.ChartTypeKPI, "Revenue Down")
		resp, err := ct.Transform(data, cfg)
		if err != nil {
			t.Fatalf("Transform: %v", err)
		}
		if resp.KPI.Trend != "down" {
			t.Errorf("Trend = %q, want %q", resp.KPI.Trend, "down")
		}
	})

	t.Run("empty data returns zero value KPI not error", func(t *testing.T) {
		t.Parallel()
		data := makeTableData()
		cfg := makeConfig(tablebuilder.ChartTypeKPI, "Empty")
		resp, err := ct.Transform(data, cfg)
		if err != nil {
			t.Fatalf("Transform on empty data: %v", err)
		}
		if resp.KPI == nil {
			t.Fatal("expected KPI, got nil")
		}
		if resp.KPI.Value != 0 {
			t.Errorf("KPI.Value = %v, want 0", resp.KPI.Value)
		}
	})

	t.Run("title set from config", func(t *testing.T) {
		t.Parallel()
		data := makeTableData(tablebuilder.TableRow{"revenue": 1.0})
		cfg := makeConfig(tablebuilder.ChartTypeKPI, "My KPI Title")
		resp, err := ct.Transform(data, cfg)
		if err != nil {
			t.Fatalf("Transform: %v", err)
		}
		if resp.Title != "My KPI Title" {
			t.Errorf("Title = %q, want %q", resp.Title, "My KPI Title")
		}
	})

	t.Run("meta.RowsProcessed matches input row count", func(t *testing.T) {
		t.Parallel()
		data := makeTableData(
			tablebuilder.TableRow{"revenue": 1.0},
			tablebuilder.TableRow{"revenue": 2.0},
		)
		cfg := makeConfig(tablebuilder.ChartTypeKPI, "KPI")
		resp, err := ct.Transform(data, cfg)
		if err != nil {
			t.Fatalf("Transform: %v", err)
		}
		if resp.Meta.RowsProcessed != 2 {
			t.Errorf("RowsProcessed = %d, want 2", resp.Meta.RowsProcessed)
		}
	})
}

func TestChartTransformer_Bar(t *testing.T) {
	t.Parallel()
	ct := tablebuilder.NewChartTransformer()

	t.Run("categories and values extracted in row order", func(t *testing.T) {
		t.Parallel()
		data := makeTableDataWithMeta(
			[]string{"month", "revenue"},
			tablebuilder.TableRow{"month": "Jan", "revenue": 1000.0},
			tablebuilder.TableRow{"month": "Feb", "revenue": 2000.0},
			tablebuilder.TableRow{"month": "Mar", "revenue": 1500.0},
		)
		cfg := makeConfig(tablebuilder.ChartTypeBar, "Monthly Revenue")
		resp, err := ct.Transform(data, cfg)
		if err != nil {
			t.Fatalf("Transform bar: %v", err)
		}
		if len(resp.Categories) != 3 {
			t.Fatalf("Categories len = %d, want 3", len(resp.Categories))
		}
		wantCats := []string{"Jan", "Feb", "Mar"}
		for i, cat := range wantCats {
			if resp.Categories[i] != cat {
				t.Errorf("Categories[%d] = %q, want %q", i, resp.Categories[i], cat)
			}
		}
		if len(resp.Series) == 0 {
			t.Fatal("expected at least one series")
		}
		wantVals := []float64{1000.0, 2000.0, 1500.0}
		for i, v := range wantVals {
			if resp.Series[0].Data[i] != v {
				t.Errorf("Series[0].Data[%d] = %v, want %v", i, resp.Series[0].Data[i], v)
			}
		}
	})

	t.Run("no category column returns error", func(t *testing.T) {
		t.Parallel()
		// All numeric columns → no category detected
		data := makeTableDataWithMeta(
			[]string{"a", "b"},
			tablebuilder.TableRow{"a": 1.0, "b": 2.0},
		)
		cfg := makeConfig(tablebuilder.ChartTypeBar, "Bad")
		_, err := ct.Transform(data, cfg)
		if err == nil {
			t.Fatal("expected error when no category column, got nil")
		}
	})
}

func TestChartTransformer_Line(t *testing.T) {
	t.Parallel()
	ct := tablebuilder.NewChartTransformer()

	t.Run("line chart same structure as bar", func(t *testing.T) {
		t.Parallel()
		data := makeTableDataWithMeta(
			[]string{"month", "units"},
			tablebuilder.TableRow{"month": "Q1", "units": 300.0},
			tablebuilder.TableRow{"month": "Q2", "units": 450.0},
		)
		cfg := makeConfig(tablebuilder.ChartTypeLine, "Units")
		resp, err := ct.Transform(data, cfg)
		if err != nil {
			t.Fatalf("Transform line: %v", err)
		}
		if resp.Type != tablebuilder.ChartTypeLine {
			t.Errorf("Type = %q, want %q", resp.Type, tablebuilder.ChartTypeLine)
		}
		if len(resp.Categories) != 2 {
			t.Errorf("Categories len = %d, want 2", len(resp.Categories))
		}
	})
}

func TestChartTransformer_StackedBar(t *testing.T) {
	t.Parallel()
	ct := tablebuilder.NewChartTransformer()

	t.Run("stacked bar series have Stack field set", func(t *testing.T) {
		t.Parallel()
		data := makeTableDataWithMeta(
			[]string{"region", "online", "offline"},
			tablebuilder.TableRow{"region": "North", "online": 500.0, "offline": 300.0},
			tablebuilder.TableRow{"region": "South", "online": 400.0, "offline": 200.0},
		)
		cfg := makeConfig(tablebuilder.ChartTypeStackedBar, "Sales by Channel")
		resp, err := ct.Transform(data, cfg)
		if err != nil {
			t.Fatalf("Transform stacked-bar: %v", err)
		}
		for i, s := range resp.Series {
			if s.Stack == "" {
				t.Errorf("Series[%d].Stack is empty, expected stacking to be set", i)
			}
		}
	})
}

func TestChartTransformer_Pie(t *testing.T) {
	t.Parallel()
	ct := tablebuilder.NewChartTransformer()

	t.Run("one series per data row", func(t *testing.T) {
		t.Parallel()
		data := makeTableData(
			tablebuilder.TableRow{"category": "Electronics", "value": 450.0},
			tablebuilder.TableRow{"category": "Clothing", "value": 300.0},
			tablebuilder.TableRow{"category": "Food", "value": 250.0},
		)
		// Inject explicit column mapping via _chart settings to avoid auto-detection.
		cfg := &tablebuilder.Config{
			Title:         "Sales Mix",
			WidgetType:    "chart",
			Visualization: tablebuilder.ChartTypePie,
			DataSource:    []tablebuilder.DataSource{{Source: "orders", Schema: "sales"}},
			VisualSettings: tablebuilder.VisualSettings{
				Columns: map[string]tablebuilder.ColumnConfig{
					"_chart": {
						CellTemplate: `{"chartType":"pie","categoryColumn":"category","valueColumns":["value"]}`,
					},
				},
			},
		}
		resp, err := ct.Transform(data, cfg)
		if err != nil {
			t.Fatalf("Transform pie: %v", err)
		}
		if len(resp.Series) != 3 {
			t.Fatalf("Series len = %d, want 3", len(resp.Series))
		}
		if resp.Series[0].Name != "Electronics" {
			t.Errorf("Series[0].Name = %q, want Electronics", resp.Series[0].Name)
		}
		if resp.Series[0].Data[0] != 450.0 {
			t.Errorf("Series[0].Data[0] = %v, want 450", resp.Series[0].Data[0])
		}
	})

	t.Run("missing category or value returns error", func(t *testing.T) {
		t.Parallel()
		// All numeric rows → no category detected by auto-detection fallback
		data := makeTableDataWithMeta(
			[]string{"a", "b"},
			tablebuilder.TableRow{"a": 1.0, "b": 2.0},
		)
		cfg := makeConfig(tablebuilder.ChartTypePie, "Bad Pie")
		_, err := ct.Transform(data, cfg)
		if err == nil {
			t.Fatal("expected error for missing category/value columns")
		}
	})
}

func TestChartTransformer_Heatmap(t *testing.T) {
	t.Parallel()
	ct := tablebuilder.NewChartTransformer()

	t.Run("x y value triple extracted correctly", func(t *testing.T) {
		t.Parallel()
		data := makeTableData(
			tablebuilder.TableRow{"day": "Mon", "hour": "9am", "count": 5.0},
			tablebuilder.TableRow{"day": "Mon", "hour": "10am", "count": 8.0},
			tablebuilder.TableRow{"day": "Tue", "hour": "9am", "count": 3.0},
		)
		// Inject heatmap settings via _chart key
		cfg := &tablebuilder.Config{
			Title:         "Order Heatmap",
			WidgetType:    "chart",
			Visualization: tablebuilder.ChartTypeHeatmap,
			DataSource:    []tablebuilder.DataSource{{Source: "orders", Schema: "sales"}},
			VisualSettings: tablebuilder.VisualSettings{
				Columns: map[string]tablebuilder.ColumnConfig{
					"_chart": {
						CellTemplate: `{"chartType":"heatmap","xCategoryColumn":"hour","yCategoryColumn":"day","valueColumns":["count"]}`,
					},
				},
			},
		}
		resp, err := ct.Transform(data, cfg)
		if err != nil {
			t.Fatalf("Transform heatmap: %v", err)
		}
		if resp.Heatmap == nil {
			t.Fatal("expected Heatmap data, got nil")
		}
		if len(resp.Heatmap.XCategories) == 0 {
			t.Error("XCategories is empty")
		}
		if len(resp.Heatmap.YCategories) == 0 {
			t.Error("YCategories is empty")
		}
	})

	t.Run("missing x or y category column returns error", func(t *testing.T) {
		t.Parallel()
		data := makeTableData(tablebuilder.TableRow{"x": "a", "y": "b", "v": 1.0})
		cfg := makeConfig(tablebuilder.ChartTypeHeatmap, "Bad Heatmap")
		// No xCategoryColumn or yCategoryColumn set in settings → error
		_, err := ct.Transform(data, cfg)
		if err == nil {
			t.Fatal("expected error for missing heatmap columns")
		}
	})
}

func TestChartTransformer_Gantt(t *testing.T) {
	t.Parallel()
	ct := tablebuilder.NewChartTransformer()

	t.Run("task start end mapped correctly", func(t *testing.T) {
		t.Parallel()
		data := makeTableData(
			tablebuilder.TableRow{
				"task":  "Design",
				"start": "2025-01-01",
				"end":   "2025-01-15",
			},
			tablebuilder.TableRow{
				"task":  "Build",
				"start": "2025-01-16",
				"end":   "2025-02-28",
			},
		)
		cfg := &tablebuilder.Config{
			Title:         "Project Timeline",
			WidgetType:    "chart",
			Visualization: tablebuilder.ChartTypeGantt,
			DataSource:    []tablebuilder.DataSource{{Source: "tasks", Schema: "project"}},
			VisualSettings: tablebuilder.VisualSettings{
				Columns: map[string]tablebuilder.ColumnConfig{
					"_chart": {
						CellTemplate: `{"chartType":"gantt","nameColumn":"task","startColumn":"start","endColumn":"end"}`,
					},
				},
			},
		}
		resp, err := ct.Transform(data, cfg)
		if err != nil {
			t.Fatalf("Transform gantt: %v", err)
		}
		if len(resp.Gantt) != 2 {
			t.Fatalf("Gantt len = %d, want 2", len(resp.Gantt))
		}
		if resp.Gantt[0].Name != "Design" {
			t.Errorf("Gantt[0].Name = %q, want Design", resp.Gantt[0].Name)
		}
		if resp.Gantt[0].StartDate != "2025-01-01" {
			t.Errorf("Gantt[0].StartDate = %q, want 2025-01-01", resp.Gantt[0].StartDate)
		}
		if resp.Gantt[1].EndDate != "2025-02-28" {
			t.Errorf("Gantt[1].EndDate = %q, want 2025-02-28", resp.Gantt[1].EndDate)
		}
	})

	t.Run("missing required columns returns error", func(t *testing.T) {
		t.Parallel()
		data := makeTableData(tablebuilder.TableRow{"task": "A"})
		cfg := makeConfig(tablebuilder.ChartTypeGantt, "Bad Gantt")
		_, err := ct.Transform(data, cfg)
		if err == nil {
			t.Fatal("expected error for missing gantt columns")
		}
	})
}

func TestChartTransformer_Treemap(t *testing.T) {
	t.Parallel()
	ct := tablebuilder.NewChartTransformer()

	t.Run("children extracted correctly", func(t *testing.T) {
		t.Parallel()
		data := makeTableData(
			tablebuilder.TableRow{"category": "Electronics", "revenue": 5000.0},
			tablebuilder.TableRow{"category": "Clothing", "revenue": 3000.0},
		)
		cfg := &tablebuilder.Config{
			Title:         "Revenue Treemap",
			WidgetType:    "chart",
			Visualization: tablebuilder.ChartTypeTreemap,
			DataSource:    []tablebuilder.DataSource{{Source: "sales", Schema: "sales"}},
			VisualSettings: tablebuilder.VisualSettings{
				Columns: map[string]tablebuilder.ColumnConfig{
					"_chart": {
						CellTemplate: `{"chartType":"treemap","categoryColumn":"category","valueColumns":["revenue"]}`,
					},
				},
			},
		}
		resp, err := ct.Transform(data, cfg)
		if err != nil {
			t.Fatalf("Transform treemap: %v", err)
		}
		if resp.Treemap == nil {
			t.Fatal("Treemap is nil")
		}
		if len(resp.Treemap.Children) != 2 {
			t.Fatalf("Children len = %d, want 2", len(resp.Treemap.Children))
		}
		if resp.Treemap.Children[0].Name != "Electronics" {
			t.Errorf("Children[0].Name = %q, want Electronics", resp.Treemap.Children[0].Name)
		}
	})
}

func TestChartTransformer_Waterfall(t *testing.T) {
	t.Parallel()
	ct := tablebuilder.NewChartTransformer()

	t.Run("step values extracted in order", func(t *testing.T) {
		t.Parallel()
		data := makeTableDataWithMeta(
			[]string{"label", "amount"},
			tablebuilder.TableRow{"label": "Start", "amount": 1000.0},
			tablebuilder.TableRow{"label": "Sales", "amount": 500.0},
			tablebuilder.TableRow{"label": "Costs", "amount": -200.0},
		)
		cfg := makeConfig(tablebuilder.ChartTypeWaterfall, "Cash Flow")
		resp, err := ct.Transform(data, cfg)
		if err != nil {
			t.Fatalf("Transform waterfall: %v", err)
		}
		if resp.Type != tablebuilder.ChartTypeWaterfall {
			t.Errorf("Type = %q, want waterfall", resp.Type)
		}
		if len(resp.Series) == 0 || len(resp.Series[0].Data) != 3 {
			t.Fatalf("Series[0].Data len = %d, want 3", len(resp.Series[0].Data))
		}
		if resp.Series[0].Data[2] != -200.0 {
			t.Errorf("Data[2] = %v, want -200", resp.Series[0].Data[2])
		}
	})
}

func TestChartTransformer_Funnel(t *testing.T) {
	t.Parallel()
	ct := tablebuilder.NewChartTransformer()

	t.Run("funnel stages extracted with explicit columns", func(t *testing.T) {
		t.Parallel()
		data := makeTableData(
			tablebuilder.TableRow{"stage": "Leads", "count": 1000.0},
			tablebuilder.TableRow{"stage": "Qualified", "count": 500.0},
			tablebuilder.TableRow{"stage": "Closed", "count": 100.0},
		)
		cfg := &tablebuilder.Config{
			Title:         "Sales Funnel",
			WidgetType:    "chart",
			Visualization: tablebuilder.ChartTypeFunnel,
			DataSource:    []tablebuilder.DataSource{{Source: "leads", Schema: "sales"}},
			VisualSettings: tablebuilder.VisualSettings{
				Columns: map[string]tablebuilder.ColumnConfig{
					"_chart": {
						CellTemplate: `{"chartType":"funnel","categoryColumn":"stage","valueColumns":["count"]}`,
					},
				},
			},
		}
		resp, err := ct.Transform(data, cfg)
		if err != nil {
			t.Fatalf("Transform funnel: %v", err)
		}
		if resp.Type != tablebuilder.ChartTypeFunnel {
			t.Errorf("Type = %q, want funnel", resp.Type)
		}
		if len(resp.Categories) != 3 {
			t.Fatalf("Categories len = %d, want 3", len(resp.Categories))
		}
		if resp.Categories[0] != "Leads" {
			t.Errorf("Categories[0] = %q, want Leads", resp.Categories[0])
		}
	})
}
