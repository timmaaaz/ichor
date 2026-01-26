// Package tablebuilder provides chart data transformation logic
package tablebuilder

import (
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"strconv"
	"time"
)

// ChartTransformer transforms table data into chart-ready format
type ChartTransformer struct{}

// NewChartTransformer creates a new chart transformer
func NewChartTransformer() *ChartTransformer {
	return &ChartTransformer{}
}

// Transform converts TableData to ChartResponse based on config
func (ct *ChartTransformer) Transform(data *TableData, config *Config) (*ChartResponse, error) {
	if data == nil {
		return nil, fmt.Errorf("no data to transform")
	}

	// Get chart settings from visual_settings
	chartSettings, err := ct.extractChartSettings(config)
	if err != nil {
		return nil, fmt.Errorf("extract chart settings: %w", err)
	}
	if chartSettings == nil {
		return nil, fmt.Errorf("no chart settings in config")
	}

	startTime := time.Now()

	var response *ChartResponse

	switch chartSettings.ChartType {
	case ChartTypeKPI, ChartTypeGauge:
		response, err = ct.transformKPI(data, chartSettings, config)
	case ChartTypeLine, ChartTypeBar, ChartTypeStackedBar, ChartTypeStackedArea:
		response, err = ct.transformCategorical(data, chartSettings, config)
	case ChartTypeCombo:
		response, err = ct.transformCombo(data, chartSettings, config)
	case ChartTypePie:
		response, err = ct.transformPie(data, chartSettings)
	case ChartTypeFunnel:
		response, err = ct.transformFunnel(data, chartSettings)
	case ChartTypeWaterfall:
		response, err = ct.transformWaterfall(data, chartSettings)
	case ChartTypeHeatmap:
		response, err = ct.transformHeatmap(data, chartSettings)
	case ChartTypeTreemap:
		response, err = ct.transformTreemap(data, chartSettings)
	case ChartTypeGantt:
		response, err = ct.transformGantt(data, chartSettings)
	default:
		return nil, fmt.Errorf("unsupported chart type: %s", chartSettings.ChartType)
	}

	if err != nil {
		return nil, err
	}

	// Add title from config
	if response.Title == "" {
		response.Title = config.Title
	}

	// Add metadata
	response.Meta = ChartMeta{
		ExecutionTime: time.Since(startTime).Milliseconds(),
		RowsProcessed: len(data.Data),
	}

	return response, nil
}

// extractChartSettings extracts chart configuration from VisualSettings
func (ct *ChartTransformer) extractChartSettings(config *Config) (*ChartVisualSettings, error) {
	// Chart settings can be stored in visual_settings as a "chart" key
	// or inferred from widget_type and visualization fields

	// First, try to get from visual_settings columns (as raw JSON)
	// We'll look for a special "chart" field in the visual settings

	// For now, we'll construct from config fields and visual_settings
	settings := &ChartVisualSettings{
		ChartType: config.Visualization,
	}

	// If visualization is empty, try widget_type
	if settings.ChartType == "" {
		settings.ChartType = config.WidgetType
	}

	// If still empty, this isn't a chart config
	if settings.ChartType == "" || settings.ChartType == "table" {
		return nil, nil
	}

	// Look for chart-specific settings in visual_settings columns
	// This is a convention: column configs can contain chart settings
	for key, colConfig := range config.VisualSettings.Columns {
		switch key {
		case "_chart":
			// Special key for chart settings stored as JSON in column config
			if colConfig.CellTemplate != "" {
				var chartSettings ChartVisualSettings
				if err := json.Unmarshal([]byte(colConfig.CellTemplate), &chartSettings); err == nil {
					// Preserve chart type from config
					chartType := settings.ChartType
					settings = &chartSettings
					if settings.ChartType == "" {
						settings.ChartType = chartType
					}
				}
			}
		}
	}

	return settings, nil
}

// transformKPI transforms data for KPI cards and gauges
func (ct *ChartTransformer) transformKPI(data *TableData, settings *ChartVisualSettings, config *Config) (*ChartResponse, error) {
	if len(data.Data) == 0 {
		return &ChartResponse{
			Type: settings.ChartType,
			KPI: &KPIData{
				Value: 0,
				Label: ct.getKPILabel(settings, config),
			},
		}, nil
	}

	// Get the first value column or first numeric column
	var valueCol string
	if len(settings.ValueColumns) > 0 {
		valueCol = settings.ValueColumns[0]
	} else {
		// Find first numeric-looking column from data
		for key := range data.Data[0] {
			if ct.isNumericValue(data.Data[0][key]) {
				valueCol = key
				break
			}
		}
	}

	if valueCol == "" {
		return nil, fmt.Errorf("KPI requires at least one value column")
	}

	currentValue := ct.extractFloat(data.Data[0], valueCol)

	kpi := &KPIData{
		Value:  currentValue,
		Label:  ct.getKPILabel(settings, config),
		Format: ct.getKPIFormat(settings),
	}

	// Calculate trend if compare column specified or multiple rows exist
	compareCol := valueCol
	if settings.KPI != nil && settings.KPI.CompareColumn != "" {
		compareCol = settings.KPI.CompareColumn
	}

	// If we have previous value data (either in separate column or second row)
	if len(data.Data) > 1 {
		previousValue := ct.extractFloat(data.Data[1], compareCol)
		kpi.PreviousValue = previousValue

		if previousValue != 0 {
			kpi.Change = ((currentValue - previousValue) / math.Abs(previousValue)) * 100
		}

		if kpi.Change > 0.01 {
			kpi.Trend = "up"
		} else if kpi.Change < -0.01 {
			kpi.Trend = "down"
		} else {
			kpi.Trend = "flat"
		}
	}

	// Set target for gauge charts
	if settings.ChartType == ChartTypeGauge {
		if settings.KPI != nil && settings.KPI.TargetValue > 0 {
			kpi.Target = settings.KPI.TargetValue
			kpi.Min = 0
			kpi.Max = settings.KPI.TargetValue * 1.2 // 120% of target as max
		} else {
			// Default gauge range
			kpi.Min = 0
			kpi.Max = currentValue * 1.5
			kpi.Target = currentValue
		}
	}

	return &ChartResponse{
		Type: settings.ChartType,
		KPI:  kpi,
	}, nil
}

// transformCategorical transforms data for line/bar/stacked charts
func (ct *ChartTransformer) transformCategorical(data *TableData, settings *ChartVisualSettings, config *Config) (*ChartResponse, error) {
	categoryCol := settings.CategoryColumn
	valueColumns := settings.ValueColumns

	// Auto-detect columns if not specified
	if categoryCol == "" || len(valueColumns) == 0 {
		categoryCol, valueColumns = ct.detectCategoricalColumns(data)
	}

	if categoryCol == "" {
		return nil, fmt.Errorf("categorical charts require a category column")
	}
	if len(valueColumns) == 0 {
		return nil, fmt.Errorf("categorical charts require at least one value column")
	}

	categories := make([]string, 0, len(data.Data))
	seriesMap := make(map[string][]float64)

	// Initialize series
	for _, col := range valueColumns {
		seriesMap[col] = make([]float64, 0, len(data.Data))
	}

	// Extract data
	for _, row := range data.Data {
		// Get category value
		catValue := ct.extractString(row, categoryCol)
		categories = append(categories, catValue)

		// Get values for each series
		for _, col := range valueColumns {
			value := ct.extractFloat(row, col)
			seriesMap[col] = append(seriesMap[col], value)
		}
	}

	// Build series array
	series := make([]SeriesData, 0, len(valueColumns))
	for _, col := range valueColumns {
		s := SeriesData{
			Name: ct.getSeriesLabel(col, settings),
			Data: seriesMap[col],
		}

		// Add stack for stacked charts
		if settings.ChartType == ChartTypeStackedBar || settings.ChartType == ChartTypeStackedArea {
			s.Stack = "total"
		}

		series = append(series, s)
	}

	return &ChartResponse{
		Type:       settings.ChartType,
		Categories: categories,
		Series:     series,
	}, nil
}

// transformCombo transforms data for dual-axis combo charts
func (ct *ChartTransformer) transformCombo(data *TableData, settings *ChartVisualSettings, config *Config) (*ChartResponse, error) {
	// Start with categorical transformation
	resp, err := ct.transformCategorical(data, settings, config)
	if err != nil {
		return nil, err
	}

	resp.Type = ChartTypeCombo

	// Apply series config if available
	if len(settings.SeriesConfig) > 0 {
		for i, sc := range settings.SeriesConfig {
			if i < len(resp.Series) {
				if sc.Type != "" {
					resp.Series[i].Type = sc.Type
				}
				resp.Series[i].YAxisIndex = sc.YAxisIndex
				if sc.Stack != "" {
					resp.Series[i].Stack = sc.Stack
				}
				if sc.Label != "" {
					resp.Series[i].Name = sc.Label
				}
			}
		}
	} else {
		// Default: first series = bar, rest = line on secondary axis
		for i := range resp.Series {
			if i == 0 {
				resp.Series[i].Type = "bar"
				resp.Series[i].YAxisIndex = 0
			} else {
				resp.Series[i].Type = "line"
				resp.Series[i].YAxisIndex = 1
			}
		}
	}

	return resp, nil
}

// transformPie transforms data for pie/donut charts
func (ct *ChartTransformer) transformPie(data *TableData, settings *ChartVisualSettings) (*ChartResponse, error) {
	categoryCol := settings.CategoryColumn
	var valueCol string
	if len(settings.ValueColumns) > 0 {
		valueCol = settings.ValueColumns[0]
	}

	// Auto-detect if not specified
	if categoryCol == "" || valueCol == "" {
		categoryCol, valueCols := ct.detectCategoricalColumns(data)
		if len(valueCols) > 0 {
			valueCol = valueCols[0]
		}
		if categoryCol == "" || valueCol == "" {
			return nil, fmt.Errorf("pie charts require category and value columns")
		}
	}

	// Create one series per data point (segment)
	series := make([]SeriesData, 0, len(data.Data))

	for _, row := range data.Data {
		name := ct.extractString(row, categoryCol)
		value := ct.extractFloat(row, valueCol)

		series = append(series, SeriesData{
			Name: name,
			Data: []float64{value},
		})
	}

	return &ChartResponse{
		Type:   ChartTypePie,
		Series: series,
	}, nil
}

// transformFunnel transforms data for funnel charts
func (ct *ChartTransformer) transformFunnel(data *TableData, settings *ChartVisualSettings) (*ChartResponse, error) {
	// Extract category and value columns
	categoryCol := settings.CategoryColumn
	var valueCol string
	if len(settings.ValueColumns) > 0 {
		valueCol = settings.ValueColumns[0]
	}

	if categoryCol == "" || valueCol == "" {
		return nil, fmt.Errorf("transformFunnel: categoryColumn and valueColumns[0] required")
	}

	// Build categories and data arrays
	categories := make([]string, 0, len(data.Data))
	values := make([]float64, 0, len(data.Data))

	for _, row := range data.Data {
		name := ct.extractString(row, categoryCol)
		value := ct.extractFloat(row, valueCol)

		categories = append(categories, name)
		values = append(values, value)
	}

	// Create single series with all values
	series := []SeriesData{
		{
			Name: ct.getSeriesLabel(valueCol, settings),
			Data: values,
		},
	}

	return &ChartResponse{
		Type:       ChartTypeFunnel,
		Categories: categories,
		Series:     series,
	}, nil
}

// transformWaterfall transforms data for waterfall charts
func (ct *ChartTransformer) transformWaterfall(data *TableData, settings *ChartVisualSettings) (*ChartResponse, error) {
	categoryCol := settings.CategoryColumn
	var valueCol string
	if len(settings.ValueColumns) > 0 {
		valueCol = settings.ValueColumns[0]
	}

	if categoryCol == "" || valueCol == "" {
		detectedCat, valueCols := ct.detectCategoricalColumns(data)
		if categoryCol == "" {
			categoryCol = detectedCat
		}
		if valueCol == "" && len(valueCols) > 0 {
			valueCol = valueCols[0]
		}
	}

	if categoryCol == "" || valueCol == "" {
		return nil, fmt.Errorf("waterfall charts require category and value columns")
	}

	categories := make([]string, 0, len(data.Data))
	values := make([]float64, 0, len(data.Data))

	for _, row := range data.Data {
		cat := ct.extractString(row, categoryCol)
		val := ct.extractFloat(row, valueCol)
		categories = append(categories, cat)
		values = append(values, val)
	}

	return &ChartResponse{
		Type:       ChartTypeWaterfall,
		Categories: categories,
		Series: []SeriesData{
			{Name: valueCol, Data: values},
		},
	}, nil
}

// transformHeatmap transforms data for heatmap charts
func (ct *ChartTransformer) transformHeatmap(data *TableData, settings *ChartVisualSettings) (*ChartResponse, error) {
	if settings.XCategoryColumn == "" || settings.YCategoryColumn == "" {
		return nil, fmt.Errorf("heatmap requires xCategoryColumn and yCategoryColumn")
	}

	var valueCol string
	if len(settings.ValueColumns) > 0 {
		valueCol = settings.ValueColumns[0]
	}

	// Build unique categories
	xCatSet := make(map[string]int)
	yCatSet := make(map[string]int)
	var xCategories, yCategories []string

	for _, row := range data.Data {
		xCat := ct.extractString(row, settings.XCategoryColumn)
		yCat := ct.extractString(row, settings.YCategoryColumn)

		if _, exists := xCatSet[xCat]; !exists {
			xCatSet[xCat] = len(xCategories)
			xCategories = append(xCategories, xCat)
		}
		if _, exists := yCatSet[yCat]; !exists {
			yCatSet[yCat] = len(yCategories)
			yCategories = append(yCategories, yCat)
		}
	}

	// Initialize data matrix
	heatData := make([][]float64, len(yCategories))
	for i := range heatData {
		heatData[i] = make([]float64, len(xCategories))
	}

	// Fill data matrix
	var minVal, maxVal float64 = math.MaxFloat64, -math.MaxFloat64
	for _, row := range data.Data {
		xCat := ct.extractString(row, settings.XCategoryColumn)
		yCat := ct.extractString(row, settings.YCategoryColumn)
		val := ct.extractFloat(row, valueCol)

		xIdx := xCatSet[xCat]
		yIdx := yCatSet[yCat]
		heatData[yIdx][xIdx] = val

		if val < minVal {
			minVal = val
		}
		if val > maxVal {
			maxVal = val
		}
	}

	return &ChartResponse{
		Type: ChartTypeHeatmap,
		Heatmap: &HeatmapData{
			XCategories: xCategories,
			YCategories: yCategories,
			Data:        heatData,
			Min:         minVal,
			Max:         maxVal,
		},
	}, nil
}

// transformTreemap transforms data for treemap charts
func (ct *ChartTransformer) transformTreemap(data *TableData, settings *ChartVisualSettings) (*ChartResponse, error) {
	categoryCol := settings.CategoryColumn
	var valueCol string
	if len(settings.ValueColumns) > 0 {
		valueCol = settings.ValueColumns[0]
	}

	if categoryCol == "" || valueCol == "" {
		return nil, fmt.Errorf("treemap requires category and value columns")
	}

	// Build flat treemap (no hierarchy for now)
	children := make([]TreemapData, 0, len(data.Data))
	for _, row := range data.Data {
		name := ct.extractString(row, categoryCol)
		value := ct.extractFloat(row, valueCol)
		children = append(children, TreemapData{
			Name:  name,
			Value: value,
		})
	}

	return &ChartResponse{
		Type: ChartTypeTreemap,
		Treemap: &TreemapData{
			Name:     "root",
			Children: children,
		},
	}, nil
}

// transformGantt transforms data for Gantt charts
func (ct *ChartTransformer) transformGantt(data *TableData, settings *ChartVisualSettings) (*ChartResponse, error) {
	nameCol := settings.NameColumn
	startCol := settings.StartColumn
	endCol := settings.EndColumn

	if nameCol == "" || startCol == "" || endCol == "" {
		return nil, fmt.Errorf("gantt requires nameColumn, startColumn, and endColumn")
	}

	ganttData := make([]GanttData, 0, len(data.Data))
	for _, row := range data.Data {
		g := GanttData{
			Name:      ct.extractString(row, nameCol),
			StartDate: ct.extractDateString(row, startCol),
			EndDate:   ct.extractDateString(row, endCol),
		}

		if settings.ProgressColumn != "" {
			g.Progress = int(ct.extractFloat(row, settings.ProgressColumn))
		}
		if settings.CategoryColumn != "" {
			g.Category = ct.extractString(row, settings.CategoryColumn)
		}

		ganttData = append(ganttData, g)
	}

	return &ChartResponse{
		Type:  ChartTypeGantt,
		Gantt: ganttData,
	}, nil
}

// =============================================================================
// Helper Functions
// =============================================================================

func (ct *ChartTransformer) extractFloat(row TableRow, column string) float64 {
	if val, ok := row[column]; ok {
		switch v := val.(type) {
		case float64:
			return v
		case float32:
			return float64(v)
		case int:
			return float64(v)
		case int64:
			return float64(v)
		case int32:
			return float64(v)
		case string:
			if f, err := strconv.ParseFloat(v, 64); err == nil {
				return f
			}
		case json.Number:
			if f, err := v.Float64(); err == nil {
				return f
			}
		}
	}
	return 0
}

func (ct *ChartTransformer) extractString(row TableRow, column string) string {
	if val, ok := row[column]; ok {
		switch v := val.(type) {
		case string:
			return v
		case time.Time:
			return v.Format("2006-01-02")
		default:
			return fmt.Sprintf("%v", v)
		}
	}
	return ""
}

func (ct *ChartTransformer) extractDateString(row TableRow, column string) string {
	if val, ok := row[column]; ok {
		switch v := val.(type) {
		case time.Time:
			return v.Format(time.RFC3339)
		case string:
			return v
		default:
			return fmt.Sprintf("%v", v)
		}
	}
	return ""
}

func (ct *ChartTransformer) isNumericValue(val any) bool {
	switch val.(type) {
	case float64, float32, int, int64, int32:
		return true
	case string:
		_, err := strconv.ParseFloat(val.(string), 64)
		return err == nil
	}
	return false
}

func (ct *ChartTransformer) detectCategoricalColumns(data *TableData) (string, []string) {
	if len(data.Data) == 0 {
		return "", nil
	}

	var categoryCol string
	var valueColumns []string

	row := data.Data[0]

	// Use metadata columns for deterministic ordering if available
	if len(data.Meta.Columns) > 0 {
		for _, col := range data.Meta.Columns {
			key := col.Field
			val, exists := row[key]
			if !exists {
				continue
			}

			if ct.isNumericValue(val) {
				valueColumns = append(valueColumns, key)
			} else if categoryCol == "" {
				categoryCol = key
			}
		}
	} else {
		// Fallback to deterministic map iteration using sorted keys
		// This handles edge cases where metadata isn't populated
		keys := make([]string, 0, len(row))
		for key := range row {
			keys = append(keys, key)
		}
		sort.Strings(keys)

		for _, key := range keys {
			val := row[key]
			if ct.isNumericValue(val) {
				valueColumns = append(valueColumns, key)
			} else if categoryCol == "" {
				categoryCol = key
			}
		}
	}

	return categoryCol, valueColumns
}

func (ct *ChartTransformer) getKPILabel(settings *ChartVisualSettings, config *Config) string {
	if settings.KPI != nil && settings.KPI.Label != "" {
		return settings.KPI.Label
	}
	if config != nil && config.Title != "" {
		return config.Title
	}
	return "Value"
}

func (ct *ChartTransformer) getKPIFormat(settings *ChartVisualSettings) string {
	if settings.KPI != nil && settings.KPI.Format != "" {
		return settings.KPI.Format
	}
	return "number"
}

func (ct *ChartTransformer) getSeriesLabel(column string, settings *ChartVisualSettings) string {
	// Check series config for label override
	for _, sc := range settings.SeriesConfig {
		if sc.Column == column && sc.Label != "" {
			return sc.Label
		}
	}
	return column
}
