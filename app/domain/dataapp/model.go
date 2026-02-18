// api/tablebuilder/models.go

package dataapp

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/timmaaaz/ichor/app/domain/config/pageactionapp"
	"github.com/timmaaaz/ichor/app/domain/config/pageconfigapp"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/business/sdk/tablebuilder"
)

// QueryParams represents the set of possible query strings.
type QueryParams struct {
	Page             string
	Rows             string
	OrderBy          string
	ConfigID         string
	Name             string
	CreatedBy        string
	UpdatedBy        string
	StartCreatedDate string
	EndCreatedDate   string
}

// =============================================================================

// TableConfig represents a table configuration for the API.
type TableConfig struct {
	ID          string          `json:"id"`
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Config      json.RawMessage `json:"config"`
	CreatedBy   string          `json:"created_by"`
	UpdatedBy   string          `json:"updated_by"`
	CreatedDate string          `json:"created_date"`
	UpdatedDate string          `json:"updated_date"`
	IsSystem    bool            `json:"is_system"`
}

type TableConfigList struct {
	Items []TableConfig `json:"items"`
}

// Encode implements the encoder interface.
func (app TableConfig) Encode() ([]byte, string, error) {
	data, err := json.Marshal(app)
	return data, "application/json", err
}

// Encode implements the encoder interface.
func (app TableConfigList) Encode() ([]byte, string, error) {
	data, err := json.Marshal(app)
	return data, "application/json", err
}

func ToAppTableConfig(bus tablebuilder.StoredConfig) TableConfig {
	return TableConfig{
		ID:          bus.ID.String(),
		Name:        bus.Name,
		Description: bus.Description,
		Config:      bus.Config,
		CreatedBy:   bus.CreatedBy.String(),
		UpdatedBy:   bus.UpdatedBy.String(),
		CreatedDate: bus.CreatedDate.Format(time.RFC3339),
		UpdatedDate: bus.UpdatedDate.Format(time.RFC3339),
		IsSystem:    bus.IsSystem,
	}
}

func ToAppTableConfigs(configs []tablebuilder.StoredConfig) []TableConfig {
	app := make([]TableConfig, len(configs))
	for i, cfg := range configs {
		app[i] = ToAppTableConfig(cfg)
	}
	return app
}

func ToAppTableConfigList(bus []tablebuilder.StoredConfig) TableConfigList {
	items := make([]TableConfig, len(bus))
	for i, item := range bus {
		items[i] = ToAppTableConfig(item)
	}
	return TableConfigList{
		Items: items,
	}
}

// =============================================================================

// NewTableConfig defines the data needed to add a new table configuration.
type NewTableConfig struct {
	Name        string          `json:"name" validate:"required,min=3,max=100"`
	Description string          `json:"description" validate:"max=500"`
	Config      json.RawMessage `json:"config" validate:"required"`
}

// Decode implements the decoder interface.
func (app *NewTableConfig) Decode(data []byte) error {
	return json.Unmarshal(data, &app)
}

// Validate checks the data in the model is considered clean.
func (app NewTableConfig) Validate() error {
	if err := errs.Check(app); err != nil {
		return errs.Newf(errs.InvalidArgument, "validate: %s", err)
	}

	// Parse and validate the config structure
	var config tablebuilder.Config
	if err := json.Unmarshal(app.Config, &config); err != nil {
		return errs.Newf(errs.InvalidArgument, "invalid config JSON: %s", err)
	}

	if err := config.Validate(); err != nil {
		return errs.Newf(errs.InvalidArgument, "config validation: %s", err)
	}

	return nil
}

func toBusNewTableConfig(app NewTableConfig) (*tablebuilder.Config, error) {
	var config tablebuilder.Config
	if err := json.Unmarshal(app.Config, &config); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}

	return &config, nil
}

// =============================================================================

// UpdateTableConfig defines the data needed to update a table configuration.
type UpdateTableConfig struct {
	Name        *string          `json:"name" validate:"omitempty,min=3,max=100"`
	Description *string          `json:"description" validate:"omitempty,max=500"`
	Config      *json.RawMessage `json:"config"`
}

// Decode implements the decoder interface.
func (app *UpdateTableConfig) Decode(data []byte) error {
	return json.Unmarshal(data, &app)
}

// Validate checks the data in the model is considered clean.
func (app UpdateTableConfig) Validate() error {
	if err := errs.Check(app); err != nil {
		return errs.Newf(errs.InvalidArgument, "validate: %s", err)
	}

	// If config is being updated, validate it
	if app.Config != nil {
		var config tablebuilder.Config
		if err := json.Unmarshal(*app.Config, &config); err != nil {
			return errs.Newf(errs.InvalidArgument, "invalid config JSON: %s", err)
		}

		if err := config.Validate(); err != nil {
			return errs.Newf(errs.InvalidArgument, "config validation: %s", err)
		}
	}

	return nil
}

func toBusUpdateTableConfig(app UpdateTableConfig) (*tablebuilder.Config, error) {
	if app.Config == nil {
		return nil, nil
	}

	var config tablebuilder.Config
	if err := json.Unmarshal(*app.Config, &config); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}

	return &config, nil
}

// =============================================================================
// Export/Import Models

// ExportPackage represents a JSON export package for table configs.
type ExportPackage struct {
	Version    string        `json:"version"`
	Type       string        `json:"type"`
	ExportedAt string        `json:"exported_at"`
	Count      int           `json:"count"`
	Data       []TableConfig `json:"data"`
}

// Encode implements the encoder interface.
func (app ExportPackage) Encode() ([]byte, string, error) {
	data, err := json.Marshal(app)
	return data, "application/json", err
}

// ImportPackage represents a JSON import package for table configs.
type ImportPackage struct {
	Mode string        `json:"mode"` // "merge", "skip", "replace"
	Data []TableConfig `json:"data"`
}

// Decode implements the decoder interface.
func (app *ImportPackage) Decode(data []byte) error {
	return json.Unmarshal(data, &app)
}

// Validate checks the data in the model is considered clean.
func (app ImportPackage) Validate() error {
	if app.Mode != "merge" && app.Mode != "skip" && app.Mode != "replace" {
		return errs.Newf(errs.InvalidArgument, "mode must be 'merge', 'skip', or 'replace'")
	}

	if len(app.Data) == 0 {
		return errs.Newf(errs.InvalidArgument, "data cannot be empty")
	}

	// Validate each config
	for i, config := range app.Data {
		if config.Name == "" {
			return errs.Newf(errs.InvalidArgument, "config %d: name is required", i)
		}
		if len(config.Config) == 0 {
			return errs.Newf(errs.InvalidArgument, "config %d: config is required", i)
		}
	}

	return nil
}

// ImportResult represents the result of an import operation.
type ImportResult struct {
	ImportedCount int      `json:"imported_count"`
	SkippedCount  int      `json:"skipped_count"`
	UpdatedCount  int      `json:"updated_count"`
	Errors        []string `json:"errors,omitempty"`
}

// Encode implements the encoder interface.
func (app ImportResult) Encode() ([]byte, string, error) {
	data, err := json.Marshal(app)
	return data, "application/json", err
}

// =============================================================================

// TableQuery defines parameters for querying table data.
type TableQuery struct {
	Filters []FilterParam  `json:"filters" validate:"dive"`
	Sort    []SortParam    `json:"sort" validate:"dive"`
	Page    *int           `json:"page" validate:"omitempty,min=1"`
	Rows    *int           `json:"rows" validate:"omitempty,min=1,max=1000"`
	Dynamic map[string]any `json:"dynamic"`
}

// FilterParam represents a filter parameter.P
type FilterParam struct {
	Column   string `json:"column" validate:"required"`
	Operator string `json:"operator" validate:"required,oneof=eq neq gt gte lt lte in like ilike is_null is_not_null"`
	Value    any    `json:"value"`
	Dynamic  bool   `json:"dynamic,omitempty"`
}

// SortParam represents a sort parameter.
type SortParam struct {
	Column    string `json:"column" validate:"required"`
	Direction string `json:"direction" validate:"required,oneof=asc desc"`
}

// Decode implements the decoder interface.
func (app *TableQuery) Decode(data []byte) error {
	return json.Unmarshal(data, &app)
}

// Validate checks the data in the model is considered clean.
func (app TableQuery) Validate() error {
	if err := errs.Check(app); err != nil {
		return errs.Newf(errs.InvalidArgument, "validate: %s", err)
	}
	return nil
}

func toBusTableQuery(app TableQuery) tablebuilder.QueryParams {
	filters := make([]tablebuilder.Filter, len(app.Filters))
	for i, f := range app.Filters {
		filters[i] = tablebuilder.Filter{
			Column:   f.Column,
			Operator: f.Operator,
			Value:    f.Value,
			Dynamic:  f.Dynamic,
		}
	}

	sorts := make([]tablebuilder.Sort, len(app.Sort))
	for i, s := range app.Sort {
		sorts[i] = tablebuilder.Sort{
			Column:    s.Column,
			Direction: s.Direction,
		}
	}

	ret := tablebuilder.QueryParams{
		Filters: filters,
		Sort:    sorts,
		Dynamic: app.Dynamic,
	}

	if app.Page != nil {
		ret.Page = *app.Page
	}
	if app.Rows != nil {
		ret.Rows = *app.Rows
	}

	return ret
}

// =============================================================================

// TableData represents the table data response - MATCHES business layer exactly
type TableData struct {
	Data []map[string]any `json:"data"`
	Meta MetaData         `json:"meta"`
}

// MetaData contains metadata about the query result - MATCHES business layer exactly
type MetaData struct {
	Total         int                `json:"total"`
	Page          int                `json:"page,omitempty"`
	PageSize      int                `json:"page_size,omitempty"`
	TotalPages    int                `json:"total_pages,omitempty"`
	ExecutionTime int64              `json:"execution_time,omitempty"` // milliseconds
	Columns       []ColumnMetadata   `json:"columns,omitempty"`
	Relationships []RelationshipInfo `json:"relationships,omitempty"`
	Error         string             `json:"error,omitempty"`
}

// ColumnMetadata - MATCHES business layer exactly
type ColumnMetadata struct {
	// Core identification
	Field        string `json:"field"`
	DisplayName  string `json:"display_name"`
	DatabaseName string `json:"database_name"`

	// Type and source
	Type         string `json:"type"`
	SourceTable  string `json:"source_table,omitempty"`
	SourceColumn string `json:"source_column,omitempty"`
	SourceSchema string `json:"source_schema,omitempty"`
	Hidden       bool   `json:"hidden,omitempty"`

	// Flags
	IsPrimaryKey bool   `json:"is_primary_key,omitempty"`
	IsForeignKey bool   `json:"is_foreign_key,omitempty"`
	RelatedTable string `json:"related_table,omitempty"`

	// Visual settings (override DisplayName if present)
	Header     string          `json:"header,omitempty"`
	Width      int             `json:"width,omitempty"`
	Align      string          `json:"align,omitempty"`
	Sortable   bool            `json:"sortable,omitempty"`
	Filterable bool            `json:"filterable,omitempty"`
	Format     *FormatConfig   `json:"format,omitempty"`
	Editable   *EditableConfig `json:"editable,omitempty"`
	Link       *LinkConfig     `json:"link,omitempty"`

	// Conditional formatting rules for this column
	ConditionalFormatting []ConditionalFormat `json:"conditional_formatting,omitempty"`
}

// ConditionalFormat - MATCHES business layer exactly
type ConditionalFormat struct {
	Column     string `json:"column"`
	Condition  string `json:"condition"`
	Value      any    `json:"value"`
	Condition2 string `json:"condition2,omitempty"`
	Value2     any    `json:"value2,omitempty"`
	Color      string `json:"color,omitempty"`
	Background string `json:"background,omitempty"`
	Icon       string `json:"icon,omitempty"`
}

// FormatConfig - MATCHES business layer exactly
type FormatConfig struct {
	Type      string `json:"type"`
	Precision int    `json:"precision,omitempty"`
	Currency  string `json:"currency,omitempty"`
	Format    string `json:"format,omitempty"`
}

// EditableConfig - MATCHES business layer exactly
type EditableConfig struct {
	Type        string `json:"type"`
	Placeholder string `json:"placeholder,omitempty"`
}

// LinkConfig - MATCHES business layer exactly
type LinkConfig struct {
	URL         string `json:"url"`
	Label       string `json:"label,omitempty"`
	LabelColumn string `json:"label_column,omitempty"`
}

// RelationshipInfo - MATCHES business layer exactly
type RelationshipInfo struct {
	FromTable  string `json:"from_table"`
	FromColumn string `json:"from_column"`
	ToTable    string `json:"to_table"`
	ToColumn   string `json:"to_column"`
	Type       string `json:"type"` // "one-to-one", "one-to-many", "many-to-one"
}

type TableDataList struct {
	Items []TableData `json:"items"`
}

// Encode implements the encoder interface.
func (app TableData) Encode() ([]byte, string, error) {
	data, err := json.Marshal(app)
	return data, "application/json", err
}

type Count struct {
	Count int `json:"count"`
}

// Encode implements the encoder interface.
func (app Count) Encode() ([]byte, string, error) {
	data, err := json.Marshal(app)
	return data, "application/json", err
}

// =============================================================================
// Page and PageTab Configurations

// ActionsGroupedByType represents page actions grouped by type (buttons, dropdowns, separators).
// This is a type alias to pageactionapp.ActionsGroupedByType for convenience.
type ActionsGroupedByType = pageactionapp.ActionsGroupedByType

// PageConfig is a type alias to pageconfigapp.PageConfig for convenience.
type PageConfig = pageconfigapp.PageConfig

// NewPageConfig is a type alias to pageconfigapp.NewPageConfig for convenience.
type NewPageConfig = pageconfigapp.NewPageConfig

// UpdatePageConfig is a type alias to pageconfigapp.UpdatePageConfig for convenience.
type UpdatePageConfig = pageconfigapp.UpdatePageConfig

type FullPageConfig struct {
	PageConfig  PageConfig           `json:"page_config"`
	PageActions ActionsGroupedByType `json:"page_actions"`
}

// Encode implements the encoder interface.
func (app FullPageConfig) Encode() ([]byte, string, error) {
	data, err := json.Marshal(app)
	return data, "application/json", err
}

// toAppTableData - Now does a DIRECT pass-through with minimal conversion
func toAppTableData(bus *tablebuilder.TableData) TableData {
	// Convert data rows (simple map conversion)
	data := make([]map[string]any, len(bus.Data))
	for i, row := range bus.Data {
		// DEBUG: Log date-related fields before JSON serialization
		for key, val := range row {
			if strings.Contains(strings.ToLower(key), "date") || strings.Contains(strings.ToLower(key), "due") {
				fmt.Printf("[DEBUG-DATE] toAppTableData - Field: %s, Type: %T, Value: %v\n", key, val, val)
			}
		}
		data[i] = map[string]any(row)
	}

	// Convert metadata with direct field mapping
	return TableData{
		Data: data,
		Meta: MetaData{
			Total:         bus.Meta.Total,
			Page:          bus.Meta.Page,
			PageSize:      bus.Meta.PageSize,
			TotalPages:    bus.Meta.TotalPages,
			ExecutionTime: bus.Meta.ExecutionTime,
			Columns:       toAppColumnMetadata(bus.Meta.Columns),
			Relationships: toAppRelationships(bus.Meta.Relationships),
			Error:         bus.Meta.Error,
		},
	}
}

// toAppColumnMetadata - Direct 1:1 field mapping
func toAppColumnMetadata(busColumns []tablebuilder.ColumnMetadata) []ColumnMetadata {
	if busColumns == nil {
		return nil
	}

	appColumns := make([]ColumnMetadata, len(busColumns))
	for i, col := range busColumns {
		appColumns[i] = ColumnMetadata{
			Field:                 col.Field,
			DisplayName:           col.DisplayName,
			DatabaseName:          col.DatabaseName,
			Type:                  col.Type,
			SourceTable:           col.SourceTable,
			SourceColumn:          col.SourceColumn,
			SourceSchema:          col.SourceSchema,
			Hidden:                col.Hidden,
			IsPrimaryKey:          col.IsPrimaryKey,
			IsForeignKey:          col.IsForeignKey,
			RelatedTable:          col.RelatedTable,
			Header:                col.Header,
			Width:                 col.Width,
			Align:                 col.Align,
			Sortable:              col.Sortable,
			Filterable:            col.Filterable,
			Format:                toAppFormatConfig(col.Format),
			Editable:              toAppEditableConfig(col.Editable),
			Link:                  toAppLinkConfig(col.Link),
			ConditionalFormatting: toAppConditionalFormatting(col.ConditionalFormatting),
		}
	}
	return appColumns
}

// toAppFormatConfig - Direct 1:1 field mapping
func toAppFormatConfig(bus *tablebuilder.FormatConfig) *FormatConfig {
	if bus == nil {
		return nil
	}
	return &FormatConfig{
		Type:      bus.Type,
		Precision: bus.Precision,
		Currency:  bus.Currency,
		Format:    bus.Format,
	}
}

// toAppEditableConfig - Direct 1:1 field mapping
func toAppEditableConfig(bus *tablebuilder.EditableConfig) *EditableConfig {
	if bus == nil {
		return nil
	}
	return &EditableConfig{
		Type:        bus.Type,
		Placeholder: bus.Placeholder,
	}
}

// toAppLinkConfig - Direct 1:1 field mapping
func toAppLinkConfig(bus *tablebuilder.LinkConfig) *LinkConfig {
	if bus == nil {
		return nil
	}
	return &LinkConfig{
		URL:         bus.URL,
		Label:       bus.Label,
		LabelColumn: bus.LabelColumn,
	}
}

// toAppConditionalFormatting - Direct 1:1 field mapping
func toAppConditionalFormatting(busRules []tablebuilder.ConditionalFormat) []ConditionalFormat {
	if len(busRules) == 0 {
		return nil
	}

	appRules := make([]ConditionalFormat, len(busRules))
	for i, rule := range busRules {
		appRules[i] = ConditionalFormat{
			Column:     rule.Column,
			Condition:  rule.Condition,
			Value:      rule.Value,
			Condition2: rule.Condition2,
			Value2:     rule.Value2,
			Color:      rule.Color,
			Background: rule.Background,
			Icon:       rule.Icon,
		}
	}
	return appRules
}

// toAppRelationships - Direct 1:1 field mapping
func toAppRelationships(busRels []tablebuilder.RelationshipInfo) []RelationshipInfo {
	if busRels == nil {
		return nil
	}

	appRels := make([]RelationshipInfo, len(busRels))
	for i, rel := range busRels {
		appRels[i] = RelationshipInfo{
			FromTable:  rel.FromTable,
			FromColumn: rel.FromColumn,
			ToTable:    rel.ToTable,
			ToColumn:   rel.ToColumn,
			Type:       rel.Type,
		}
	}
	return appRels
}

// =============================================================================
// Chart Response Types
// =============================================================================

// ChartResponse represents the unified response for all chart types
type ChartResponse struct {
	Type       string       `json:"type"`
	Title      string       `json:"title,omitempty"`
	Subtitle   string       `json:"subtitle,omitempty"`
	Categories []string     `json:"categories,omitempty"`
	Series     []SeriesData `json:"series,omitempty"`
	KPI        *KPIData     `json:"kpi,omitempty"`
	Heatmap    *HeatmapData `json:"heatmap,omitempty"`
	Gantt      []GanttData  `json:"gantt,omitempty"`
	Treemap    *TreemapData `json:"treemap,omitempty"`
	Meta       ChartMeta    `json:"meta"`
}

// Encode implements the encoder interface.
func (app ChartResponse) Encode() ([]byte, string, error) {
	data, err := json.Marshal(app)
	return data, "application/json", err
}

// SeriesData represents a data series for categorical charts
type SeriesData struct {
	Name       string    `json:"name"`
	Type       string    `json:"type,omitempty"`
	YAxisIndex int       `json:"y_axis_index,omitempty"`
	Data       []float64 `json:"data"`
	Stack      string    `json:"stack,omitempty"`
}

// KPIData represents KPI-specific data
type KPIData struct {
	Value         float64 `json:"value"`
	PreviousValue float64 `json:"previous_value,omitempty"`
	Change        float64 `json:"change,omitempty"`
	Trend         string  `json:"trend,omitempty"`
	Label         string  `json:"label"`
	Format        string  `json:"format,omitempty"`
	Target        float64 `json:"target,omitempty"`
	Min           float64 `json:"min,omitempty"`
	Max           float64 `json:"max,omitempty"`
}

// ChartMeta contains metadata about the chart query
type ChartMeta struct {
	ExecutionTime int64  `json:"execution_time_ms"`
	RowsProcessed int    `json:"rows_processed"`
	Error         string `json:"error,omitempty"`
}

// HeatmapData for heatmap charts
type HeatmapData struct {
	XCategories []string    `json:"x_categories"`
	YCategories []string    `json:"y_categories"`
	Data        [][]float64 `json:"data"`
	Min         float64     `json:"min"`
	Max         float64     `json:"max"`
}

// GanttData for Gantt charts
type GanttData struct {
	Name      string `json:"name"`
	StartDate string `json:"start_date"`
	EndDate   string `json:"end_date"`
	Progress  int    `json:"progress,omitempty"`
	Category  string `json:"category,omitempty"`
}

// TreemapData for treemap charts
type TreemapData struct {
	Name     string        `json:"name"`
	Value    float64       `json:"value,omitempty"`
	Children []TreemapData `json:"children,omitempty"`
}

// toAppChartResponse converts business layer ChartResponse to app layer
func toAppChartResponse(bus *tablebuilder.ChartResponse) ChartResponse {
	resp := ChartResponse{
		Type:       bus.Type,
		Title:      bus.Title,
		Subtitle:   bus.Subtitle,
		Categories: bus.Categories,
		Meta: ChartMeta{
			ExecutionTime: bus.Meta.ExecutionTime,
			RowsProcessed: bus.Meta.RowsProcessed,
			Error:         bus.Meta.Error,
		},
	}

	// Convert series
	if len(bus.Series) > 0 {
		resp.Series = make([]SeriesData, len(bus.Series))
		for i, s := range bus.Series {
			resp.Series[i] = SeriesData{
				Name:       s.Name,
				Type:       s.Type,
				YAxisIndex: s.YAxisIndex,
				Data:       s.Data,
				Stack:      s.Stack,
			}
		}
	}

	// Convert KPI
	if bus.KPI != nil {
		resp.KPI = &KPIData{
			Value:         bus.KPI.Value,
			PreviousValue: bus.KPI.PreviousValue,
			Change:        bus.KPI.Change,
			Trend:         bus.KPI.Trend,
			Label:         bus.KPI.Label,
			Format:        bus.KPI.Format,
			Target:        bus.KPI.Target,
			Min:           bus.KPI.Min,
			Max:           bus.KPI.Max,
		}
	}

	// Convert Heatmap
	if bus.Heatmap != nil {
		resp.Heatmap = &HeatmapData{
			XCategories: bus.Heatmap.XCategories,
			YCategories: bus.Heatmap.YCategories,
			Data:        bus.Heatmap.Data,
			Min:         bus.Heatmap.Min,
			Max:         bus.Heatmap.Max,
		}
	}

	// Convert Gantt
	if len(bus.Gantt) > 0 {
		resp.Gantt = make([]GanttData, len(bus.Gantt))
		for i, g := range bus.Gantt {
			resp.Gantt[i] = GanttData{
				Name:      g.Name,
				StartDate: g.StartDate,
				EndDate:   g.EndDate,
				Progress:  g.Progress,
				Category:  g.Category,
			}
		}
	}

	// Convert Treemap
	if bus.Treemap != nil {
		resp.Treemap = toAppTreemapData(bus.Treemap)
	}

	return resp
}

// =============================================================================
// Chart Preview Request
// =============================================================================

// PreviewChartDataRequest represents a request to preview chart data with a config.
type PreviewChartDataRequest struct {
	Config json.RawMessage `json:"config" validate:"required"`
	Query  TableQuery      `json:"query"`
}

// Decode implements the decoder interface.
func (app *PreviewChartDataRequest) Decode(data []byte) error {
	return json.Unmarshal(data, &app)
}

// Validate checks the data in the model is considered clean.
func (app PreviewChartDataRequest) Validate() error {
	if err := errs.Check(app); err != nil {
		return errs.Newf(errs.InvalidArgument, "validate: %s", err)
	}
	if len(app.Config) == 0 {
		return errs.Newf(errs.InvalidArgument, "config is required")
	}
	return nil
}

// toAppTreemapData recursively converts treemap data
func toAppTreemapData(bus *tablebuilder.TreemapData) *TreemapData {
	if bus == nil {
		return nil
	}

	result := &TreemapData{
		Name:  bus.Name,
		Value: bus.Value,
	}

	if len(bus.Children) > 0 {
		result.Children = make([]TreemapData, len(bus.Children))
		for i, child := range bus.Children {
			childPtr := toAppTreemapData(&child)
			if childPtr != nil {
				result.Children[i] = *childPtr
			}
		}
	}

	return result
}
