// api/tablebuilder/models.go

package dataapp

import (
	"encoding/json"
	"fmt"
	"time"

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

// TableQuery defines parameters for querying table data.
type TableQuery struct {
	Filters []FilterParam  `json:"filters" validate:"dive"`
	Sort    []SortParam    `json:"sort" validate:"dive"`
	Page    int            `json:"page" validate:"min=0"`
	Rows    int            `json:"rows" validate:"min=1,max=1000"`
	Dynamic map[string]any `json:"dynamic"`
}

// FilterParam represents a filter parameter.
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

	return tablebuilder.QueryParams{
		Filters: filters,
		Sort:    sorts,
		Page:    app.Page,
		Rows:    app.Rows,
		Dynamic: app.Dynamic,
	}
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
	URL   string `json:"url"`
	Label string `json:"label"`
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

// toAppTableData - Now does a DIRECT pass-through with minimal conversion
func toAppTableData(bus *tablebuilder.TableData) TableData {
	// Convert data rows (simple map conversion)
	data := make([]map[string]any, len(bus.Data))
	for i, row := range bus.Data {
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
			Field:        col.Field,
			DisplayName:  col.DisplayName,
			DatabaseName: col.DatabaseName,
			Type:         col.Type,
			SourceTable:  col.SourceTable,
			SourceColumn: col.SourceColumn,
			Hidden:       col.Hidden,
			IsPrimaryKey: col.IsPrimaryKey,
			IsForeignKey: col.IsForeignKey,
			RelatedTable: col.RelatedTable,
			Header:       col.Header,
			Width:        col.Width,
			Align:        col.Align,
			Sortable:     col.Sortable,
			Filterable:   col.Filterable,
			Format:       toAppFormatConfig(col.Format),
			Editable:     toAppEditableConfig(col.Editable),
			Link:         toAppLinkConfig(col.Link),
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
		URL:   bus.URL,
		Label: bus.Label,
	}
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
