// Package tablebuilder provides a dynamic table query builder for PostgreSQL
// that supports complex configurations including joins, filters, computed columns,
// and nested data structures.
package tablebuilder

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
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
	Limit        int            `json:"limit,omitempty"`
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
	Align        string          `json:"align,omitempty"` // left, center, right
	Type         string          `json:"type,omitempty"`  // linktotal, etc.
	Sortable     bool            `json:"sortable,omitempty"`
	Filterable   bool            `json:"filterable,omitempty"`
	CellTemplate string          `json:"cell_template,omitempty"`
	Format       *FormatConfig   `json:"format,omitempty"`
	Editable     *EditableConfig `json:"editable,omitempty"`
	Link         *LinkConfig     `json:"link,omitempty"`
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
	URL   string `json:"url"`
	Label string `json:"label"`
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

// TableData represents the result of a table query
type TableData struct {
	Data []TableRow `json:"data"`
	Meta MetaData   `json:"meta"`
}

// TableRow represents a single row of data
type TableRow map[string]any

// MetaData contains metadata about the query result
type MetaData struct {
	Total         int               `json:"total"`
	Config        *Config           `json:"config,omitempty"`
	AliasMap      map[string]string `json:"alias_map,omitempty"`
	Error         string            `json:"error,omitempty"`
	ExecutionTime int64             `json:"execution_time,omitempty"` // milliseconds
	Page          int               `json:"page,omitempty"`
	PageSize      int               `json:"page_size,omitempty"`
	TotalPages    int               `json:"total_pages,omitempty"`
}

// =============================================================================
// Query Parameters
// =============================================================================

// QueryParams represents runtime query parameters
type QueryParams struct {
	Filters []Filter       `json:"filters,omitempty"`
	Sort    []Sort         `json:"sort,omitempty"`
	Page    int            `json:"page,omitempty"`
	Limit   int            `json:"limit,omitempty"`
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
