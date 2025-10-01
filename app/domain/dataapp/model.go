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

func toAppTableConfig(bus tablebuilder.StoredConfig) TableConfig {
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

func toAppTableConfigs(configs []tablebuilder.StoredConfig) []TableConfig {
	app := make([]TableConfig, len(configs))
	for i, cfg := range configs {
		app[i] = toAppTableConfig(cfg)
	}
	return app
}

func toAppTableConfigList(bus []tablebuilder.StoredConfig) TableConfigList {
	items := make([]TableConfig, len(bus))
	for i, item := range bus {
		items[i] = toAppTableConfig(item)
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
	Limit   int            `json:"limit" validate:"min=1,max=1000"`
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
		Limit:   app.Limit,
		Dynamic: app.Dynamic,
	}
}

// =============================================================================

// TableData represents the table data response.
type TableData struct {
	Data []map[string]any `json:"data"`
	Meta TableMeta        `json:"meta"`
}

// TableMeta contains metadata about the query execution.
type TableMeta struct {
	Total         int               `json:"total"`
	Page          int               `json:"page,omitempty"`
	PageSize      int               `json:"page_size,omitempty"`
	TotalPages    int               `json:"total_pages,omitempty"`
	ExecutionTime int64             `json:"execution_time_ms"`
	AliasMap      map[string]string `json:"alias_map,omitempty"`
}

type TableDataList struct {
	Items []TableData `json:"items"`
}

// Encode implements the encoder interface.
func (app TableData) Encode() ([]byte, string, error) {
	data, err := json.Marshal(app)
	return data, "application/json", err
}

func toAppTableData(bus *tablebuilder.TableData) TableData {
	data := make([]map[string]any, len(bus.Data))
	for i, row := range bus.Data {
		data[i] = map[string]any(row)
	}

	return TableData{
		Data: data,
		Meta: TableMeta{
			Total:         bus.Meta.Total,
			Page:          bus.Meta.Page,
			PageSize:      bus.Meta.PageSize,
			TotalPages:    bus.Meta.TotalPages,
			ExecutionTime: bus.Meta.ExecutionTime,
			AliasMap:      bus.Meta.AliasMap,
		},
	}
}
