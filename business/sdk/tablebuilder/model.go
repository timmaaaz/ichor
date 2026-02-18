// Package tablebuilder provides a dynamic table query builder for PostgreSQL
// that supports complex configurations including joins, filters, computed columns,
// and nested data structures.
package tablebuilder

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/sdk/sqldb/nulltypes"
)

// =============================================================================
// Core Configuration Types
// =============================================================================

// Config represents the main table configuration
type Config struct {
	Title           string         `json:"title" db:"title"`
	WidgetType      string         `json:"widget_type" db:"widget_type"`
	Visualization   string         `json:"visualization" db:"visualization"`
	PositionX       int            `json:"position_x" db:"position_x"`
	PositionY       int            `json:"position_y" db:"position_y"`
	Width           int            `json:"width" db:"width"`
	Height          int            `json:"height" db:"height"`
	DataSource      []DataSource   `json:"data_source" db:"data_source"`
	RefreshInterval int            `json:"refresh_interval" db:"refresh_interval"`
	RefreshMode     string         `json:"refresh_mode" db:"refresh_mode"`
	VisualSettings  VisualSettings `json:"visual_settings" db:"visual_settings"`
	Permissions     Permissions    `json:"permissions" db:"permissions"`
}

// DataSource represents a data source configuration
type DataSource struct {
	Type         string         `json:"type"`   // "query", "view", "viewcount", "rpc"
	Source       string         `json:"source"` // Table/view/function name
	Schema       string         `json:"schema,omitempty"`
	Select       SelectConfig   `json:"select"`
	Args         map[string]any `json:"args,omitempty"`
	SelectBy     string         `json:"select_by,omitempty"`
	ParentSource string         `json:"parent_source,omitempty"`
	Joins        []Join         `json:"joins,omitempty"`
	Filters      []Filter       `json:"filters,omitempty"`
	Sort         []Sort         `json:"sort,omitempty"`
	Rows         int            `json:"rows,omitempty"`

	// Chart aggregation (safe, structured)
	Metrics []MetricConfig  `json:"metrics,omitempty"`
	GroupBy []GroupByConfig `json:"group_by,omitempty"`
}

// =============================================================================
// Chart Metric Configuration Types
// =============================================================================

// MetricConfig defines an aggregated value for charts
type MetricConfig struct {
	Name       string            `json:"name"`                 // Output alias: "total_revenue"
	Function   string            `json:"function"`             // sum, count, avg, min, max, count_distinct
	Column     string            `json:"column,omitempty"`     // Simple column: "quantity"
	Expression *ExpressionConfig `json:"expression,omitempty"` // For multi-column math
}

// ExpressionConfig for multi-column arithmetic
type ExpressionConfig struct {
	Operator string   `json:"operator"` // multiply, add, subtract, divide
	Columns  []string `json:"columns"`  // ["order_line_items.quantity", "product_costs.selling_price"]
}

// GroupByConfig for time-series and categorical grouping
type GroupByConfig struct {
	Column     string `json:"column"`               // "orders.created_date" or "categories.name" or SQL expression
	Interval   string `json:"interval,omitempty"`   // day, week, month, quarter, year (dates only)
	Alias      string `json:"alias,omitempty"`      // Output name: "month"
	Expression bool   `json:"expression,omitempty"` // If true, Column is treated as raw SQL
}

// =============================================================================
// Metric Validation Whitelists
// =============================================================================

// AllowedAggregateFunctions contains the allowed aggregate functions (whitelist)
var AllowedAggregateFunctions = map[string]string{
	"sum":            "SUM",
	"count":          "COUNT",
	"count_distinct": "COUNT_DISTINCT", // Special handling in builder
	"avg":            "AVG",
	"min":            "MIN",
	"max":            "MAX",
}

// AllowedOperators contains the allowed expression operators
var AllowedOperators = map[string]string{
	"multiply": "*",
	"add":      "+",
	"subtract": "-",
	"divide":   "/",
}

// AllowedIntervals contains the allowed time intervals for GROUP BY
var AllowedIntervals = map[string]string{
	"day":     "day",
	"week":    "week",
	"month":   "month",
	"quarter": "quarter",
	"year":    "year",
}

// SelectConfig defines what columns to select
type SelectConfig struct {
	Columns               []ColumnDefinition `json:"columns"`
	ForeignTables         []ForeignTable     `json:"foreign_tables,omitempty"`
	ClientComputedColumns []ComputedColumn   `json:"client_computed_columns,omitempty"`
}

// ColumnDefinition represents a column selection
type ColumnDefinition struct {
	Name        string `json:"name"`                   // Column name in database
	Alias       string `json:"alias,omitempty"`        // Optional alias
	TableColumn string `json:"table_column,omitempty"` // Format: "table.column"
}

// ForeignTable represents a related table configuration
type ForeignTable struct {
	Table                 string             `json:"table"`
	Alias                 string             `json:"alias,omitempty"`  // Optional alias for the table (required for multiple joins to same table)
	Schema                string             `json:"schema,omitempty"` // Optional, defaults to public
	RelationshipFrom      string             `json:"relationship_from"`
	RelationshipTo        string             `json:"relationship_to"`
	JoinType              string             `json:"join_type,omitempty"` // inner, left, right, full
	Columns               []ColumnDefinition `json:"columns"`
	ForeignTables         []ForeignTable     `json:"foreign_tables,omitempty"`
	RelationshipDirection string             `json:"relationship_direction,omitempty"` // parent_to_child, child_to_parent
}

// ComputedColumn represents a client-side computed column
type ComputedColumn struct {
	Name       string `json:"name"`
	Expression string `json:"expression"`
}

// Join represents a table join
type Join struct {
	Table  string `json:"table"`
	Schema string `json:"schema,omitempty"` // Optional, defaults to public
	Type   string `json:"type"`             // inner, left, right, full
	On     string `json:"on"`
}

// Filter represents a query filter
type Filter struct {
	Column      string `json:"column"`
	Operator    string `json:"operator"` // eq, neq, gt, gte, lt, lte, in, like, ilike
	Value       any    `json:"value"`
	Dynamic     bool   `json:"dynamic,omitempty"`
	Label       string `json:"label,omitempty"`
	ControlType string `json:"control_type,omitempty"`
}

// Sort represents a sort configuration
type Sort struct {
	Column      string   `json:"column"`
	Direction   string   `json:"direction"` // asc, desc
	Priority    int      `json:"priority,omitempty"`
	CustomOrder []string `json:"custom_order,omitempty"`
}

// =============================================================================
// Visual Settings Types
// =============================================================================

// VisualSettings contains all visual configuration
type VisualSettings struct {
	Columns               map[string]ColumnConfig `json:"columns"`
	ConditionalFormatting []ConditionalFormat     `json:"conditional_formatting,omitempty"`
	RowActions            map[string]Action       `json:"row_actions,omitempty"`
	TableActions          map[string]Action       `json:"table_actions,omitempty"`
	Pagination            *PaginationConfig       `json:"pagination,omitempty"`
	Theme                 string                  `json:"theme,omitempty"`
}

// ColumnConfig represents visual configuration for a column
type ColumnConfig struct {
	Name         string          `json:"name"`
	Header       string          `json:"header"`
	Width        int             `json:"width,omitempty"`
	Align        string          `json:"align,omitempty"`  // left, center, right
	Type         string          `json:"type,omitempty"`   // linktotal, status, lookup, etc.
	Hidden       bool            `json:"hidden,omitempty"` // Column selected but not displayed (e.g., for lookup labels)
	Order        int             `json:"order,omitempty"`  // Display order (lower = earlier, 0 = implicit order)
	Sortable     bool            `json:"sortable,omitempty"`
	Filterable   bool            `json:"filterable,omitempty"`
	CellTemplate string          `json:"cell_template,omitempty"`
	Format       *FormatConfig   `json:"format,omitempty"`
	Editable     *EditableConfig `json:"editable,omitempty"`
	Link         *LinkConfig     `json:"link,omitempty"`
	Lookup       *LookupConfig   `json:"lookup,omitempty"` // Lookup dropdown config for filters
}

// FormatConfig defines how to format a value
type FormatConfig struct {
	Type      string `json:"type"` // number, currency, date, datetime, boolean, percent
	Precision int    `json:"precision,omitempty"`
	Currency  string `json:"currency,omitempty"`
	Format    string `json:"format,omitempty"` // For dates
}

// EditableConfig defines if/how a field is editable
type EditableConfig struct {
	Type        string `json:"type"` // text, number, checkbox, select, date, textarea
	Placeholder string `json:"placeholder,omitempty"`
}

// LinkConfig defines link configuration
type LinkConfig struct {
	URL         string `json:"url"`
	Label       string `json:"label,omitempty"`        // Static label text
	LabelColumn string `json:"label_column,omitempty"` // Column name to use as dynamic label (takes precedence over Label)
}

// LookupConfig defines configuration for lookup dropdown filters.
// Used when Type="lookup" to specify how to populate the filter dropdown.
// Mirrors formfieldbus.DropdownConfig for consistency across the application.
type LookupConfig struct {
	Entity         string   `json:"entity"`                    // Format: "schema.table" (e.g., "core.users")
	LabelColumn    string   `json:"label_column"`              // Column to display as option label
	ValueColumn    string   `json:"value_column"`              // Column to use as option value (usually 'id')
	DisplayColumns []string `json:"display_columns,omitempty"` // Additional columns to show in dropdown
}

// ConditionalFormat represents conditional formatting rules
type ConditionalFormat struct {
	Column     string `json:"column"`
	Condition  string `json:"condition"` // eq, neq, gt, lt, etc.
	Value      any    `json:"value"`
	Condition2 string `json:"condition2,omitempty"`
	Value2     any    `json:"value2,omitempty"`
	Color      string `json:"color,omitempty"`
	Background string `json:"background,omitempty"`
	Icon       string `json:"icon,omitempty"`
}

// Action represents an action configuration
type Action struct {
	Name       string         `json:"name"`
	Label      string         `json:"label"`
	Icon       string         `json:"icon,omitempty"`
	ActionType string         `json:"action_type"` // link, modal, export, print, custom
	URL        string         `json:"url,omitempty"`
	Component  string         `json:"component,omitempty"`
	Params     map[string]any `json:"params,omitempty"`
	Format     string         `json:"format,omitempty"`
}

// PaginationConfig represents pagination settings
type PaginationConfig struct {
	Enabled         bool  `json:"enabled"`
	PageSizes       []int `json:"page_sizes"`
	DefaultPageSize int   `json:"default_page_size"`
}

// Permissions represents access permissions
type Permissions struct {
	Roles   []string `json:"roles"`
	Actions []string `json:"actions"` // view, edit, delete, export
}

// =============================================================================
// Result Types
// =============================================================================
type TableData struct {
	Data []TableRow `json:"data"`
	Meta MetaData   `json:"meta"`
}

// TableRow represents a single row of data
type TableRow map[string]any

// MetaData contains metadata about the query result
type MetaData struct {
	Total         int   `json:"total"`
	Page          int   `json:"page,omitempty"`
	PageSize      int   `json:"page_size,omitempty"`
	TotalPages    int   `json:"total_pages,omitempty"`
	ExecutionTime int64 `json:"execution_time,omitempty"` // milliseconds

	Columns       []ColumnMetadata   `json:"columns,omitempty"`
	Relationships []RelationshipInfo `json:"relationships,omitempty"`

	Error string `json:"error,omitempty"`
}

type ColumnMetadata struct {
	// Core identification
	Field        string `json:"field"`         // Key in data row (uses alias if present)
	DisplayName  string `json:"display_name"`  // What user sees (header or alias or name)
	DatabaseName string `json:"database_name"` // Original column name in DB

	// Type and source
	Type         string `json:"type"`
	SourceTable  string `json:"source_table,omitempty"`
	SourceColumn string `json:"source_column,omitempty"`
	SourceSchema string `json:"source_schema,omitempty"`
	Hidden       bool   `json:"hidden,omitempty"`
	Order        int    `json:"order,omitempty"` // Display order for frontend (lower = earlier)

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

	// Conditional formatting rules filtered for this column
	ConditionalFormatting []ConditionalFormat `json:"conditional_formatting,omitempty"`
}

type RelationshipInfo struct {
	FromTable  string `json:"from_table"`
	FromColumn string `json:"from_column"`
	ToTable    string `json:"to_table"`
	ToColumn   string `json:"to_column"`
	Type       string `json:"type"` // "one-to-one", "one-to-many", "many-to-one"
}

// =============================================================================
// Query Parameters
// =============================================================================

// QueryParams represents runtime query parameters
type QueryParams struct {
	Filters []Filter       `json:"filters,omitempty"`
	Sort    []Sort         `json:"sort,omitempty"`
	Page    int            `json:"page,omitempty"`
	Rows    int            `json:"rows,omitempty"`
	Dynamic map[string]any `json:"dynamic,omitempty"` // Dynamic filter values
}

// =============================================================================
// Configuration Storage Types
// =============================================================================

// StoredConfig represents a saved configuration in the database
type StoredConfig struct {
	ID          uuid.UUID       `db:"id" json:"id"`
	Name        string          `db:"name" json:"name"`
	Description string          `db:"description" json:"description,omitempty"`
	Config      json.RawMessage `db:"config" json:"config"`
	CreatedBy   uuid.UUID       `db:"created_by" json:"created_by"`
	UpdatedBy   uuid.UUID       `db:"updated_by" json:"updated_by"`
	CreatedDate time.Time       `db:"created_date" json:"created_date"`
	UpdatedDate time.Time       `db:"updated_date" json:"updated_date"`
	IsSystem    bool            `db:"is_system" json:"is_system"`
}

// =============================================================================
// Helper Methods
// =============================================================================

// Validate performs basic validation on the configuration
func (c *Config) Validate() error {
	if c.Title == "" {
		return ErrInvalidConfig
	}
	if len(c.DataSource) == 0 {
		return ErrNoDataSource
	}
	for _, ds := range c.DataSource {
		if ds.Source == "" {
			return ErrInvalidDataSource
		}
	}

	// Validate that all columns have explicit types in VisualSettings
	if err := c.validateColumnTypes(); err != nil {
		return err
	}

	return nil
}

// validateColumnTypes ensures all columns have valid types defined in VisualSettings.
// This is required for type-aware filtering, charting, and other type-dependent features.
// Chart widgets are skipped because they use ChartVisualSettings via the "_chart" key instead.
// Exempt columns (LabelColumn references, hidden columns) don't require a Type field.
func (c *Config) validateColumnTypes() error {
	if len(c.DataSource) == 0 {
		return nil
	}

	// Skip validation for chart widgets - they use ChartVisualSettings via "_chart" key
	// and don't require column-level Type fields
	if c.WidgetType == "chart" {
		return nil
	}

	ds := c.DataSource[0]

	// Collect columns exempt from Type validation
	exemptColumns := c.collectExemptColumnsForValidation()

	// Check regular columns
	for _, col := range ds.Select.Columns {
		fieldName := col.Name
		if col.Alias != "" {
			fieldName = col.Alias
		} else if col.TableColumn != "" {
			fieldName = col.TableColumn
		}

		// Skip exempt columns
		if exemptColumns[fieldName] {
			continue
		}

		vs, ok := c.VisualSettings.Columns[fieldName]
		if !ok || vs.Type == "" {
			return fmt.Errorf("%w: %s", ErrMissingColumnType, fieldName)
		}
		if !IsValidColumnType(vs.Type) {
			return fmt.Errorf("%w: %s has invalid type %q", ErrInvalidColumn, fieldName, vs.Type)
		}
		// Datetime columns must have a Format config
		if vs.Type == "datetime" && vs.Format == nil {
			return fmt.Errorf("%w: %s", ErrMissingDatetimeFormat, fieldName)
		}
	}

	// Check foreign table columns recursively
	if err := c.validateForeignTableColumnTypes(ds.Select.ForeignTables, exemptColumns); err != nil {
		return err
	}

	// Check computed columns (they should have type "computed" or a specific type)
	for _, cc := range ds.Select.ClientComputedColumns {
		// Skip exempt columns
		if exemptColumns[cc.Name] {
			continue
		}

		vs, ok := c.VisualSettings.Columns[cc.Name]
		if !ok || vs.Type == "" {
			return fmt.Errorf("%w: %s (computed)", ErrMissingColumnType, cc.Name)
		}
		if !IsValidColumnType(vs.Type) {
			return fmt.Errorf("%w: %s has invalid type %q", ErrInvalidColumn, cc.Name, vs.Type)
		}
		// Datetime columns must have a Format config
		if vs.Type == "datetime" && vs.Format == nil {
			return fmt.Errorf("%w: %s (computed)", ErrMissingDatetimeFormat, cc.Name)
		}
	}

	return nil
}

// collectExemptColumnsForValidation returns a set of column names that are exempt from Type validation.
// This includes:
// 1. Columns used as LabelColumn in LinkConfig (display purposes in links)
// 2. Columns marked as Hidden (selected for data but not displayed)
func (c *Config) collectExemptColumnsForValidation() map[string]bool {
	exempt := make(map[string]bool)
	for name, colConfig := range c.VisualSettings.Columns {
		// Exempt LabelColumn references
		if colConfig.Link != nil && colConfig.Link.LabelColumn != "" {
			exempt[colConfig.Link.LabelColumn] = true
		}
		// Exempt hidden columns
		if colConfig.Hidden {
			exempt[name] = true
		}
	}
	return exempt
}

// validateForeignTableColumnTypes recursively validates column types for foreign tables.
func (c *Config) validateForeignTableColumnTypes(foreignTables []ForeignTable, exemptColumns map[string]bool) error {
	for _, ft := range foreignTables {
		for _, col := range ft.Columns {
			fieldName := col.Name
			if col.Alias != "" {
				fieldName = col.Alias
			} else if col.TableColumn != "" {
				fieldName = col.TableColumn
			}

			// Skip exempt columns
			if exemptColumns[fieldName] {
				continue
			}

			vs, ok := c.VisualSettings.Columns[fieldName]
			if !ok || vs.Type == "" {
				return fmt.Errorf("%w: %s (from %s.%s)", ErrMissingColumnType, fieldName, ft.Schema, ft.Table)
			}
			if !IsValidColumnType(vs.Type) {
				return fmt.Errorf("%w: %s has invalid type %q", ErrInvalidColumn, fieldName, vs.Type)
			}
			// Datetime columns must have a Format config
			if vs.Type == "datetime" && vs.Format == nil {
				return fmt.Errorf("%w: %s (from %s.%s)", ErrMissingDatetimeFormat, fieldName, ft.Schema, ft.Table)
			}
		}

		// Recursively check nested foreign tables
		if err := c.validateForeignTableColumnTypes(ft.ForeignTables, exemptColumns); err != nil {
			return err
		}
	}

	return nil
}

// GetColumnByName finds a column definition by its name
func (s *SelectConfig) GetColumnByName(name string) *ColumnDefinition {
	for i := range s.Columns {
		if s.Columns[i].Name == name {
			return &s.Columns[i]
		}
	}
	return nil
}

type PageConfig struct {
	ID        uuid.UUID `db:"id" json:"id"`
	Name      string    `db:"name" json:"name"`
	UserID    uuid.UUID `db:"user_id" json:"user_id"`
	IsDefault bool      `db:"is_default" json:"is_default"`
}

type UpdatePageConfig struct {
	Name      *string
	UserID    *uuid.UUID
	IsDefault *bool
}

// =============================================================================
// Database Models and Conversion Functions
// =============================================================================

// dbPageConfig is the database representation of PageConfig with nullable user_id
type dbPageConfig struct {
	ID        uuid.UUID      `db:"id"`
	Name      string         `db:"name"`
	UserID    sql.NullString `db:"user_id"`
	IsDefault bool           `db:"is_default"`
}

// toDBPageConfig converts a PageConfig to its database representation
func toDBPageConfig(pc PageConfig) dbPageConfig {
	return dbPageConfig{
		ID:        pc.ID,
		Name:      pc.Name,
		UserID:    nulltypes.ToNullableUUID(pc.UserID),
		IsDefault: pc.IsDefault,
	}
}

// toBusPageConfig converts a database PageConfig to its business representation
func toBusPageConfig(db dbPageConfig) PageConfig {
	return PageConfig{
		ID:        db.ID,
		Name:      db.Name,
		UserID:    nulltypes.FromNullableUUID(db.UserID),
		IsDefault: db.IsDefault,
	}
}

// toBusPageConfigs converts multiple database PageConfigs to business representations
func toBusPageConfigs(dbs []dbPageConfig) []PageConfig {
	bus := make([]PageConfig, len(dbs))
	for i, db := range dbs {
		bus[i] = toBusPageConfig(db)
	}
	return bus
}

// =============================================================================
// Page Content System - Flexible content blocks with user-customizable layouts
// =============================================================================

// Content type constants
const (
	ContentTypeTable     = "table"
	ContentTypeForm      = "form"
	ContentTypeTabs      = "tabs"
	ContentTypeContainer = "container"
	ContentTypeText      = "text"
	ContentTypeChart     = "chart"
)

// Container type constants
const (
	ContainerTypeTab       = "tab"
	ContainerTypeAccordion = "accordion"
	ContainerTypeSection   = "section"
	ContainerTypeGrid      = "grid"
)

// PageContent represents a flexible content block on a page
type PageContent struct {
	ID            uuid.UUID       `db:"id" json:"id"`
	PageConfigID  uuid.UUID       `db:"page_config_id" json:"pageConfigId"`
	ContentType   string          `db:"content_type" json:"contentType"`
	Label         string          `db:"label" json:"label,omitempty"`
	TableConfigID uuid.UUID       `db:"table_config_id" json:"tableConfigId,omitempty"`
	FormID        uuid.UUID       `db:"form_id" json:"formId,omitempty"`
	OrderIndex    int             `db:"order_index" json:"orderIndex"`
	ParentID      uuid.UUID       `db:"parent_id" json:"parentId,omitempty"`
	Layout        json.RawMessage `db:"layout" json:"layout"`
	IsVisible     bool            `db:"is_visible" json:"isVisible"`
	IsDefault     bool            `db:"is_default" json:"isDefault"`
	Children      []PageContent   `json:"children,omitempty"` // Populated by queries, not stored in DB
}

// LayoutConfig holds all layout/styling configuration (stored as JSONB)
type LayoutConfig struct {
	// Responsive column spans (Tailwind col-span-*)
	ColSpan *ResponsiveValue `json:"colSpan,omitempty"`

	// Row span
	RowSpan int `json:"rowSpan,omitempty"`

	// Explicit grid positioning
	ColStart *int `json:"colStart,omitempty"`
	RowStart *int `json:"rowStart,omitempty"`

	// Container-specific (if this content is a container)
	GridCols *ResponsiveValue `json:"gridCols,omitempty"`
	Gap      string           `json:"gap,omitempty"` // Tailwind gap class: "gap-4", "gap-x-4 gap-y-6"

	// Additional Tailwind classes
	ClassName string `json:"className,omitempty"`

	// Container behavior
	ContainerType string `json:"containerType,omitempty"` // "tab", "accordion", "section", "grid"

	// Display options
	Collapsible bool `json:"collapsible,omitempty"`
}

// ResponsiveValue holds mobile-first responsive values (Tailwind breakpoints)
type ResponsiveValue struct {
	Default int  `json:"default"`
	Sm      *int `json:"sm,omitempty"`
	Md      *int `json:"md,omitempty"`
	Lg      *int `json:"lg,omitempty"`
	Xl      *int `json:"xl,omitempty"`
	Xl2     *int `json:"2xl,omitempty"`
}

// GetTailwindClasses generates Tailwind classes from layout config
func (lc *LayoutConfig) GetTailwindClasses() string {
	classes := []string{}

	// Column span
	if lc.ColSpan != nil {
		classes = append(classes, lc.ColSpan.ToColSpanClasses()...)
	}

	// Row span
	if lc.RowSpan > 0 {
		classes = append(classes, fmt.Sprintf("row-span-%d", lc.RowSpan))
	}

	// Positioning
	if lc.ColStart != nil {
		classes = append(classes, fmt.Sprintf("col-start-%d", *lc.ColStart))
	}
	if lc.RowStart != nil {
		classes = append(classes, fmt.Sprintf("row-start-%d", *lc.RowStart))
	}

	// Custom classes
	if lc.ClassName != "" {
		classes = append(classes, lc.ClassName)
	}

	return strings.Join(classes, " ")
}

// GetContainerClasses generates container Tailwind classes
func (lc *LayoutConfig) GetContainerClasses() string {
	if lc.GridCols == nil {
		return ""
	}

	classes := []string{"grid"}
	classes = append(classes, lc.GridCols.ToGridColsClasses()...)

	if lc.Gap != "" {
		classes = append(classes, lc.Gap)
	}

	return strings.Join(classes, " ")
}

// ToColSpanClasses converts responsive values to col-span-* classes
func (rv *ResponsiveValue) ToColSpanClasses() []string {
	classes := []string{fmt.Sprintf("col-span-%d", rv.Default)}

	if rv.Sm != nil {
		classes = append(classes, fmt.Sprintf("sm:col-span-%d", *rv.Sm))
	}
	if rv.Md != nil {
		classes = append(classes, fmt.Sprintf("md:col-span-%d", *rv.Md))
	}
	if rv.Lg != nil {
		classes = append(classes, fmt.Sprintf("lg:col-span-%d", *rv.Lg))
	}
	if rv.Xl != nil {
		classes = append(classes, fmt.Sprintf("xl:col-span-%d", *rv.Xl))
	}
	if rv.Xl2 != nil {
		classes = append(classes, fmt.Sprintf("2xl:col-span-%d", *rv.Xl2))
	}

	return classes
}

// ToGridColsClasses converts responsive values to grid-cols-* classes
func (rv *ResponsiveValue) ToGridColsClasses() []string {
	classes := []string{fmt.Sprintf("grid-cols-%d", rv.Default)}

	if rv.Sm != nil {
		classes = append(classes, fmt.Sprintf("sm:grid-cols-%d", *rv.Sm))
	}
	if rv.Md != nil {
		classes = append(classes, fmt.Sprintf("md:grid-cols-%d", *rv.Md))
	}
	if rv.Lg != nil {
		classes = append(classes, fmt.Sprintf("lg:grid-cols-%d", *rv.Lg))
	}
	if rv.Xl != nil {
		classes = append(classes, fmt.Sprintf("xl:grid-cols-%d", *rv.Xl))
	}
	if rv.Xl2 != nil {
		classes = append(classes, fmt.Sprintf("2xl:grid-cols-%d", *rv.Xl2))
	}

	return classes
}

// =============================================================================
// Chart Response Types
// =============================================================================

// Chart type constants
const (
	ChartTypeLine       = "line"
	ChartTypeBar        = "bar"
	ChartTypeStackedBar = "stacked-bar"
	ChartTypeStackedArea = "stacked-area"
	ChartTypeCombo      = "combo"
	ChartTypeKPI        = "kpi"
	ChartTypeGauge      = "gauge"
	ChartTypePie        = "pie"
	ChartTypeWaterfall  = "waterfall"
	ChartTypeFunnel     = "funnel"
	ChartTypeHeatmap    = "heatmap"
	ChartTypeTreemap    = "treemap"
	ChartTypeGantt      = "gantt"
)

// ChartResponse is the unified response for all chart types
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

// SeriesData represents a data series for categorical charts
type SeriesData struct {
	Name       string    `json:"name"`
	Type       string    `json:"type,omitempty"`       // For combo charts: "bar", "line"
	YAxisIndex int       `json:"yAxisIndex,omitempty"` // For dual-axis charts
	Data       []float64 `json:"data"`
	Stack      string    `json:"stack,omitempty"` // For stacked charts
}

// KPIData represents KPI-specific data
type KPIData struct {
	Value         float64 `json:"value"`
	PreviousValue float64 `json:"previousValue,omitempty"`
	Change        float64 `json:"change,omitempty"` // Percentage change
	Trend         string  `json:"trend,omitempty"`  // "up", "down", "flat"
	Label         string  `json:"label"`
	Format        string  `json:"format,omitempty"` // "currency", "percent", "number", "compact"
	Target        float64 `json:"target,omitempty"` // For gauge charts
	Min           float64 `json:"min,omitempty"`    // For gauge charts
	Max           float64 `json:"max,omitempty"`    // For gauge charts
}

// ChartMeta contains metadata about the chart query
type ChartMeta struct {
	ExecutionTime int64  `json:"executionTimeMs"`
	RowsProcessed int    `json:"rowsProcessed"`
	Error         string `json:"error,omitempty"`
}

// HeatmapData for heatmap charts
type HeatmapData struct {
	XCategories []string      `json:"xCategories"`
	YCategories []string      `json:"yCategories"`
	Data        [][]float64   `json:"data"` // [y][x] = value
	Min         float64       `json:"min"`
	Max         float64       `json:"max"`
}

// GanttData for Gantt charts
type GanttData struct {
	Name      string `json:"name"`
	StartDate string `json:"startDate"` // ISO format
	EndDate   string `json:"endDate"`
	Progress  int    `json:"progress,omitempty"` // 0-100
	Category  string `json:"category,omitempty"`
}

// TreemapData for treemap charts
type TreemapData struct {
	Name     string        `json:"name"`
	Value    float64       `json:"value,omitempty"`
	Children []TreemapData `json:"children,omitempty"`
}

// =============================================================================
// Chart Visual Settings
// =============================================================================

// ChartVisualSettings extends VisualSettings for chart-specific configuration
type ChartVisualSettings struct {
	ChartType string `json:"chartType"`

	// Data mapping
	CategoryColumn string   `json:"categoryColumn,omitempty"`
	ValueColumns   []string `json:"valueColumns,omitempty"`

	// Aggregation
	AggregateFunction string `json:"aggregateFunction,omitempty"` // sum, avg, count, min, max
	GroupBy           string `json:"groupBy,omitempty"`

	// Axes
	XAxis  *AxisConfig `json:"xAxis,omitempty"`
	YAxis  *AxisConfig `json:"yAxis,omitempty"`
	Y2Axis *AxisConfig `json:"y2Axis,omitempty"` // For dual-axis

	// Appearance
	Colors  []string       `json:"colors,omitempty"`
	Legend  *LegendConfig  `json:"legend,omitempty"`
	Tooltip *TooltipConfig `json:"tooltip,omitempty"`

	// KPI-specific
	KPI *KPIConfig `json:"kpi,omitempty"`

	// Heatmap-specific
	XCategoryColumn string `json:"xCategoryColumn,omitempty"`
	YCategoryColumn string `json:"yCategoryColumn,omitempty"`

	// Gantt-specific
	NameColumn     string `json:"nameColumn,omitempty"`
	StartColumn    string `json:"startColumn,omitempty"`
	EndColumn      string `json:"endColumn,omitempty"`
	ProgressColumn string `json:"progressColumn,omitempty"`

	// Series configuration for combo charts
	SeriesConfig []SeriesConfig `json:"seriesConfig,omitempty"`
}

// SeriesConfig defines how each value column should be rendered
type SeriesConfig struct {
	Column     string `json:"column"`
	Type       string `json:"type,omitempty"`       // "bar", "line", "area"
	YAxisIndex int    `json:"yAxisIndex,omitempty"` // 0 or 1 for dual-axis
	Stack      string `json:"stack,omitempty"`      // Stack group name
	Label      string `json:"label,omitempty"`      // Display name
}

// AxisConfig for chart axes
type AxisConfig struct {
	Title  string   `json:"title,omitempty"`
	Type   string   `json:"type,omitempty"`   // "value", "category", "time"
	Min    *float64 `json:"min,omitempty"`
	Max    *float64 `json:"max,omitempty"`
	Format string   `json:"format,omitempty"` // "currency", "percent", "number"
}

// LegendConfig for chart legend
type LegendConfig struct {
	Show     bool   `json:"show"`
	Position string `json:"position,omitempty"` // "top", "bottom", "left", "right"
}

// TooltipConfig for chart tooltips
type TooltipConfig struct {
	Show   bool   `json:"show"`
	Format string `json:"format,omitempty"`
}

// KPIConfig for KPI cards
type KPIConfig struct {
	Label             string  `json:"label"`
	Format            string  `json:"format,omitempty"`            // "currency", "percent", "number", "compact"
	CompareColumn     string  `json:"compareColumn,omitempty"`     // For trend calculation
	TargetValue       float64 `json:"targetValue,omitempty"`       // For gauge
	ThresholdWarning  float64 `json:"thresholdWarning,omitempty"`
	ThresholdCritical float64 `json:"thresholdCritical,omitempty"`
}
