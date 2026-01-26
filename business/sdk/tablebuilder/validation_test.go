package tablebuilder_test

import (
	"errors"
	"testing"

	"github.com/timmaaaz/ichor/business/sdk/tablebuilder"
)

// =============================================================================
// Required Fields Tests
// =============================================================================

func TestValidateConfig_RequiredFields(t *testing.T) {
	tests := []struct {
		name        string
		config      tablebuilder.Config
		expectError bool
		errorField  string
	}{
		{
			name:        "empty title",
			config:      tablebuilder.Config{},
			expectError: true,
			errorField:  "title",
		},
		{
			name: "missing data source",
			config: tablebuilder.Config{
				Title: "Test",
			},
			expectError: true,
			errorField:  "data_source",
		},
		{
			name: "data source missing source",
			config: tablebuilder.Config{
				Title: "Test",
				DataSource: []tablebuilder.DataSource{
					{Type: "query"},
				},
			},
			expectError: true,
			errorField:  "data_source[0].source",
		},
		{
			name: "valid minimal config",
			config: tablebuilder.Config{
				Title: "Test",
				DataSource: []tablebuilder.DataSource{
					{Source: "test_table"},
				},
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.config.ValidateConfig()

			if tt.expectError && !result.HasErrors() {
				t.Errorf("expected error for field %s, got none", tt.errorField)
			}

			if !tt.expectError && result.HasErrors() {
				t.Errorf("unexpected errors: %v", result.Errors)
			}

			if tt.expectError && result.HasErrors() {
				found := false
				for _, err := range result.Errors {
					if err.Field == tt.errorField {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("expected error for field %s, got errors: %v", tt.errorField, result.Errors)
				}
			}
		})
	}
}

// =============================================================================
// Whitelist Validation Tests
// =============================================================================

func TestValidateConfig_WidgetType(t *testing.T) {
	tests := []struct {
		name        string
		widgetType  string
		expectError bool
	}{
		{"valid table", "table", false},
		{"valid chart", "chart", false},
		{"empty is valid", "", false},
		{"invalid type", "invalid", true},
		{"invalid graph", "graph", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := tablebuilder.Config{
				Title:      "Test",
				WidgetType: tt.widgetType,
				DataSource: []tablebuilder.DataSource{{Source: "test"}},
			}
			result := config.ValidateConfig()

			hasWidgetError := false
			for _, err := range result.Errors {
				if err.Field == "widget_type" {
					hasWidgetError = true
					break
				}
			}

			if tt.expectError && !hasWidgetError {
				t.Errorf("expected widget_type error for %q", tt.widgetType)
			}
			if !tt.expectError && hasWidgetError {
				t.Errorf("unexpected widget_type error for %q", tt.widgetType)
			}
		})
	}
}

func TestValidateConfig_DataSourceType(t *testing.T) {
	tests := []struct {
		name        string
		dsType      string
		expectError bool
	}{
		{"valid query", "query", false},
		{"valid view", "view", false},
		{"valid viewcount", "viewcount", false},
		{"valid rpc", "rpc", false},
		{"empty is valid", "", false},
		{"invalid type", "invalid", true},
		{"invalid sql", "sql", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := tablebuilder.Config{
				Title: "Test",
				DataSource: []tablebuilder.DataSource{
					{Source: "test", Type: tt.dsType},
				},
			}
			result := config.ValidateConfig()

			hasTypeError := false
			for _, err := range result.Errors {
				if err.Field == "data_source[0].type" {
					hasTypeError = true
					break
				}
			}

			if tt.expectError && !hasTypeError {
				t.Errorf("expected data_source[0].type error for %q", tt.dsType)
			}
			if !tt.expectError && hasTypeError {
				t.Errorf("unexpected data_source[0].type error for %q", tt.dsType)
			}
		})
	}
}

func TestValidateConfig_FilterOperators(t *testing.T) {
	validOperators := []string{"eq", "neq", "gt", "gte", "lt", "lte", "in", "like", "ilike"}
	invalidOperators := []string{"equals", "!=", ">", ">=", "contains", "between"}

	for _, op := range validOperators {
		t.Run("valid_"+op, func(t *testing.T) {
			config := tablebuilder.Config{
				Title: "Test",
				DataSource: []tablebuilder.DataSource{
					{
						Source: "test",
						Filters: []tablebuilder.Filter{
							{Column: "col", Operator: op, Value: "test"},
						},
					},
				},
			}
			result := config.ValidateConfig()

			for _, err := range result.Errors {
				if err.Field == "data_source[0].filters[0].operator" {
					t.Errorf("unexpected error for valid operator %q", op)
				}
			}
		})
	}

	for _, op := range invalidOperators {
		t.Run("invalid_"+op, func(t *testing.T) {
			config := tablebuilder.Config{
				Title: "Test",
				DataSource: []tablebuilder.DataSource{
					{
						Source: "test",
						Filters: []tablebuilder.Filter{
							{Column: "col", Operator: op, Value: "test"},
						},
					},
				},
			}
			result := config.ValidateConfig()

			found := false
			for _, err := range result.Errors {
				if err.Field == "data_source[0].filters[0].operator" {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("expected error for invalid operator %q", op)
			}
		})
	}
}

func TestValidateConfig_SortDirections(t *testing.T) {
	tests := []struct {
		name        string
		direction   string
		expectError bool
	}{
		{"valid asc", "asc", false},
		{"valid desc", "desc", false},
		{"invalid ASC", "ASC", true},
		{"invalid ascending", "ascending", true},
		{"invalid empty", "", true}, // direction is required
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := tablebuilder.Config{
				Title: "Test",
				DataSource: []tablebuilder.DataSource{
					{
						Source: "test",
						Sort: []tablebuilder.Sort{
							{Column: "col", Direction: tt.direction},
						},
					},
				},
			}
			result := config.ValidateConfig()

			hasDirectionError := false
			for _, err := range result.Errors {
				if err.Field == "data_source[0].sort[0].direction" {
					hasDirectionError = true
					break
				}
			}

			if tt.expectError && !hasDirectionError {
				t.Errorf("expected direction error for %q", tt.direction)
			}
			if !tt.expectError && hasDirectionError {
				t.Errorf("unexpected direction error for %q", tt.direction)
			}
		})
	}
}

func TestValidateConfig_JoinTypes(t *testing.T) {
	tests := []struct {
		name        string
		joinType    string
		expectError bool
	}{
		{"valid inner", "inner", false},
		{"valid left", "left", false},
		{"valid right", "right", false},
		{"valid full", "full", false},
		{"invalid INNER", "INNER", true},
		{"invalid outer", "outer", true},
		{"invalid cross", "cross", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := tablebuilder.Config{
				Title: "Test",
				DataSource: []tablebuilder.DataSource{
					{
						Source: "test",
						Joins: []tablebuilder.Join{
							{Table: "other", Type: tt.joinType, On: "test.id = other.id"},
						},
					},
				},
			}
			result := config.ValidateConfig()

			hasJoinError := false
			for _, err := range result.Errors {
				if err.Field == "data_source[0].joins[0].type" {
					hasJoinError = true
					break
				}
			}

			if tt.expectError && !hasJoinError {
				t.Errorf("expected join type error for %q", tt.joinType)
			}
			if !tt.expectError && hasJoinError {
				t.Errorf("unexpected join type error for %q", tt.joinType)
			}
		})
	}
}

func TestValidateConfig_Alignments(t *testing.T) {
	tests := []struct {
		name        string
		align       string
		expectError bool
	}{
		{"valid left", "left", false},
		{"valid center", "center", false},
		{"valid right", "right", false},
		{"empty is valid", "", false},
		{"invalid justify", "justify", true},
		{"invalid LEFT", "LEFT", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := tablebuilder.Config{
				Title: "Test",
				DataSource: []tablebuilder.DataSource{
					{Source: "test"},
				},
				VisualSettings: tablebuilder.VisualSettings{
					Columns: map[string]tablebuilder.ColumnConfig{
						"col1": {Name: "col1", Type: "string", Align: tt.align},
					},
				},
			}
			result := config.ValidateConfig()

			hasAlignError := false
			for _, err := range result.Errors {
				if err.Field == "visual_settings.columns[col1].align" {
					hasAlignError = true
					break
				}
			}

			if tt.expectError && !hasAlignError {
				t.Errorf("expected align error for %q", tt.align)
			}
			if !tt.expectError && hasAlignError {
				t.Errorf("unexpected align error for %q", tt.align)
			}
		})
	}
}

func TestValidateConfig_ColumnTypes(t *testing.T) {
	validTypes := []string{"string", "number", "datetime", "boolean", "uuid", "status", "computed", "lookup"}
	invalidTypes := []string{"text", "int", "integer", "date", "bool"}

	for _, colType := range validTypes {
		t.Run("valid_"+colType, func(t *testing.T) {
			config := tablebuilder.Config{
				Title: "Test",
				DataSource: []tablebuilder.DataSource{
					{Source: "test"},
				},
				VisualSettings: tablebuilder.VisualSettings{
					Columns: map[string]tablebuilder.ColumnConfig{
						"col1": {Name: "col1", Type: colType},
					},
				},
			}
			result := config.ValidateConfig()

			for _, err := range result.Errors {
				if err.Field == "visual_settings.columns[col1].type" {
					t.Errorf("unexpected error for valid type %q", colType)
				}
			}
		})
	}

	for _, colType := range invalidTypes {
		t.Run("invalid_"+colType, func(t *testing.T) {
			config := tablebuilder.Config{
				Title: "Test",
				DataSource: []tablebuilder.DataSource{
					{Source: "test"},
				},
				VisualSettings: tablebuilder.VisualSettings{
					Columns: map[string]tablebuilder.ColumnConfig{
						"col1": {Name: "col1", Type: colType},
					},
				},
			}
			result := config.ValidateConfig()

			found := false
			for _, err := range result.Errors {
				if err.Field == "visual_settings.columns[col1].type" {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("expected error for invalid type %q", colType)
			}
		})
	}
}

func TestValidateConfig_FormatTypes(t *testing.T) {
	tests := []struct {
		name        string
		formatType  string
		expectError bool
	}{
		{"valid number", "number", false},
		{"valid currency", "currency", false},
		{"valid date", "date", false},
		{"valid datetime", "datetime", false},
		{"valid boolean", "boolean", false},
		{"valid percent", "percent", false},
		{"empty is valid", "", false},
		{"invalid money", "money", true},
		{"invalid decimal", "decimal", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := tablebuilder.Config{
				Title: "Test",
				DataSource: []tablebuilder.DataSource{
					{Source: "test"},
				},
				VisualSettings: tablebuilder.VisualSettings{
					Columns: map[string]tablebuilder.ColumnConfig{
						"col1": {
							Name: "col1",
							Type: "number",
							Format: &tablebuilder.FormatConfig{
								Type: tt.formatType,
							},
						},
					},
				},
			}
			result := config.ValidateConfig()

			hasFormatError := false
			for _, err := range result.Errors {
				if err.Field == "visual_settings.columns[col1].format.type" {
					hasFormatError = true
					break
				}
			}

			if tt.expectError && !hasFormatError {
				t.Errorf("expected format type error for %q", tt.formatType)
			}
			if !tt.expectError && hasFormatError {
				t.Errorf("unexpected format type error for %q", tt.formatType)
			}
		})
	}
}

func TestValidateConfig_EditableTypes(t *testing.T) {
	tests := []struct {
		name         string
		editableType string
		expectError  bool
	}{
		{"valid text", "text", false},
		{"valid number", "number", false},
		{"valid checkbox", "checkbox", false},
		{"valid boolean", "boolean", false}, // alias for checkbox
		{"valid select", "select", false},
		{"valid date", "date", false},
		{"valid textarea", "textarea", false},
		{"invalid input", "input", true},
		{"invalid string", "string", true},
		{"empty is error", "", true}, // type is required for editable
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := tablebuilder.Config{
				Title: "Test",
				DataSource: []tablebuilder.DataSource{
					{Source: "test"},
				},
				VisualSettings: tablebuilder.VisualSettings{
					Columns: map[string]tablebuilder.ColumnConfig{
						"col1": {
							Name: "col1",
							Type: "string",
							Editable: &tablebuilder.EditableConfig{
								Type: tt.editableType,
							},
						},
					},
				},
			}
			result := config.ValidateConfig()

			hasEditableError := false
			for _, err := range result.Errors {
				if err.Field == "visual_settings.columns[col1].editable.type" {
					hasEditableError = true
					break
				}
			}

			if tt.expectError && !hasEditableError {
				t.Errorf("expected editable type error for %q", tt.editableType)
			}
			if !tt.expectError && hasEditableError {
				t.Errorf("unexpected editable type error for %q", tt.editableType)
			}
		})
	}
}

func TestValidateConfig_ActionTypes(t *testing.T) {
	tests := []struct {
		name        string
		actionType  string
		expectError bool
	}{
		{"valid link", "link", false},
		{"valid modal", "modal", false},
		{"valid export", "export", false},
		{"valid print", "print", false},
		{"valid custom", "custom", false},
		{"invalid button", "button", true},
		{"invalid click", "click", true},
		{"empty is error", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := tablebuilder.Config{
				Title: "Test",
				DataSource: []tablebuilder.DataSource{
					{Source: "test"},
				},
				VisualSettings: tablebuilder.VisualSettings{
					RowActions: map[string]tablebuilder.Action{
						"test_action": {
							Name:       "test",
							Label:      "Test",
							ActionType: tt.actionType,
							URL:        "/test", // needed for link type
						},
					},
				},
			}
			result := config.ValidateConfig()

			hasActionError := false
			for _, err := range result.Errors {
				if err.Field == "visual_settings.row_actions[test_action].action_type" {
					hasActionError = true
					break
				}
			}

			if tt.expectError && !hasActionError {
				t.Errorf("expected action type error for %q", tt.actionType)
			}
			if !tt.expectError && hasActionError {
				t.Errorf("unexpected action type error for %q", tt.actionType)
			}
		})
	}
}

// =============================================================================
// Column Reference Validation Tests
// =============================================================================

func TestValidateConfig_ColumnReferences(t *testing.T) {
	tests := []struct {
		name        string
		column      string
		expectError bool
	}{
		{"valid simple", "column", false},
		{"valid with table", "table.column", false},
		{"valid with schema", "schema.table.column", false},
		{"valid underscore", "my_column", false},
		{"valid with number", "col1", false},
		{"invalid starts with number", "1column", true},
		{"invalid special chars", "col-name", true},
		{"invalid spaces", "col name", true},
		{"invalid sql injection", "col; DROP TABLE", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := tablebuilder.Config{
				Title: "Test",
				DataSource: []tablebuilder.DataSource{
					{
						Source: "test",
						Filters: []tablebuilder.Filter{
							{Column: tt.column, Operator: "eq", Value: "test"},
						},
					},
				},
			}
			result := config.ValidateConfig()

			hasColumnError := false
			for _, err := range result.Errors {
				if err.Field == "data_source[0].filters[0].column" && err.Code == "INVALID_FORMAT" {
					hasColumnError = true
					break
				}
			}

			if tt.expectError && !hasColumnError {
				t.Errorf("expected column reference error for %q", tt.column)
			}
			if !tt.expectError && hasColumnError {
				t.Errorf("unexpected column reference error for %q", tt.column)
			}
		})
	}
}

// =============================================================================
// Foreign Table Validation Tests
// =============================================================================

func TestValidateConfig_ForeignTable(t *testing.T) {
	tests := []struct {
		name        string
		ft          tablebuilder.ForeignTable
		expectError bool
		errorField  string
	}{
		{
			name: "valid foreign table",
			ft: tablebuilder.ForeignTable{
				Table:            "related",
				RelationshipFrom: "main.id",
				RelationshipTo:   "related.main_id",
			},
			expectError: false,
		},
		{
			name: "missing table",
			ft: tablebuilder.ForeignTable{
				RelationshipFrom: "main.id",
				RelationshipTo:   "related.main_id",
			},
			expectError: true,
			errorField:  "data_source[0].select.foreign_tables[0].table",
		},
		{
			name: "missing relationship_from",
			ft: tablebuilder.ForeignTable{
				Table:          "related",
				RelationshipTo: "related.main_id",
			},
			expectError: true,
			errorField:  "data_source[0].select.foreign_tables[0].relationship_from",
		},
		{
			name: "invalid join type",
			ft: tablebuilder.ForeignTable{
				Table:            "related",
				RelationshipFrom: "main.id",
				RelationshipTo:   "related.main_id",
				JoinType:         "invalid",
			},
			expectError: true,
			errorField:  "data_source[0].select.foreign_tables[0].join_type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := tablebuilder.Config{
				Title: "Test",
				DataSource: []tablebuilder.DataSource{
					{
						Source: "main",
						Select: tablebuilder.SelectConfig{
							ForeignTables: []tablebuilder.ForeignTable{tt.ft},
						},
					},
				},
			}
			result := config.ValidateConfig()

			if tt.expectError {
				found := false
				for _, err := range result.Errors {
					if err.Field == tt.errorField {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("expected error for field %s, got errors: %v", tt.errorField, result.Errors)
				}
			} else {
				// Check no foreign table related errors
				for _, err := range result.Errors {
					if err.Field == tt.errorField || (len(tt.errorField) > 0 && err.Field[:len("data_source[0].select.foreign_tables")] == "data_source[0].select.foreign_tables") {
						t.Errorf("unexpected error: %v", err)
					}
				}
			}
		})
	}
}

// =============================================================================
// Lookup Config Validation Tests
// =============================================================================

func TestValidateConfig_LookupConfig(t *testing.T) {
	t.Run("lookup type requires lookup config", func(t *testing.T) {
		config := tablebuilder.Config{
			Title: "Test",
			DataSource: []tablebuilder.DataSource{
				{Source: "test"},
			},
			VisualSettings: tablebuilder.VisualSettings{
				Columns: map[string]tablebuilder.ColumnConfig{
					"col1": {Name: "col1", Type: "lookup"}, // missing Lookup config
				},
			},
		}
		result := config.ValidateConfig()

		found := false
		for _, err := range result.Errors {
			if err.Field == "visual_settings.columns[col1].lookup" {
				found = true
				break
			}
		}
		if !found {
			t.Error("expected error for missing lookup config when type is 'lookup'")
		}
	})

	t.Run("lookup config requires entity", func(t *testing.T) {
		config := tablebuilder.Config{
			Title: "Test",
			DataSource: []tablebuilder.DataSource{
				{Source: "test"},
			},
			VisualSettings: tablebuilder.VisualSettings{
				Columns: map[string]tablebuilder.ColumnConfig{
					"col1": {
						Name: "col1",
						Type: "lookup",
						Lookup: &tablebuilder.LookupConfig{
							LabelColumn: "name",
							ValueColumn: "id",
						},
					},
				},
			},
		}
		result := config.ValidateConfig()

		found := false
		for _, err := range result.Errors {
			if err.Field == "visual_settings.columns[col1].lookup.entity" {
				found = true
				break
			}
		}
		if !found {
			t.Error("expected error for missing entity in lookup config")
		}
	})
}

// =============================================================================
// Pagination Config Validation Tests
// =============================================================================

func TestValidateConfig_PaginationConfig(t *testing.T) {
	t.Run("page sizes must be positive", func(t *testing.T) {
		config := tablebuilder.Config{
			Title: "Test",
			DataSource: []tablebuilder.DataSource{
				{Source: "test"},
			},
			VisualSettings: tablebuilder.VisualSettings{
				Pagination: &tablebuilder.PaginationConfig{
					Enabled:   true,
					PageSizes: []int{10, 0, 25}, // 0 is invalid
				},
			},
		}
		result := config.ValidateConfig()

		found := false
		for _, err := range result.Errors {
			if err.Field == "visual_settings.pagination.page_sizes[1]" {
				found = true
				break
			}
		}
		if !found {
			t.Error("expected error for zero page size")
		}
	})

	t.Run("default page size warning when not in page sizes", func(t *testing.T) {
		config := tablebuilder.Config{
			Title: "Test",
			DataSource: []tablebuilder.DataSource{
				{Source: "test"},
			},
			VisualSettings: tablebuilder.VisualSettings{
				Pagination: &tablebuilder.PaginationConfig{
					Enabled:         true,
					PageSizes:       []int{10, 25, 50},
					DefaultPageSize: 20, // not in page_sizes
				},
			},
		}
		result := config.ValidateConfig()

		found := false
		for _, warn := range result.Warnings {
			if warn.Field == "visual_settings.pagination.default_page_size" {
				found = true
				break
			}
		}
		if !found {
			t.Error("expected warning for default_page_size not in page_sizes")
		}
	})
}

// =============================================================================
// Numeric Validation Tests
// =============================================================================

func TestValidateConfig_NumericFields(t *testing.T) {
	t.Run("negative values validation", func(t *testing.T) {
		config := tablebuilder.Config{
			Title:           "Test",
			RefreshInterval: -1,
			PositionX:       -1,
			PositionY:       -1,
			Width:           -1,
			Height:          -1,
			DataSource: []tablebuilder.DataSource{
				{Source: "test", Rows: -1},
			},
			VisualSettings: tablebuilder.VisualSettings{
				Columns: map[string]tablebuilder.ColumnConfig{
					"col1": {
						Name:  "col1",
						Type:  "number",
						Width: -1,
						Format: &tablebuilder.FormatConfig{
							Precision: -1,
						},
					},
				},
			},
		}
		result := config.ValidateConfig()

		expectedFields := []string{
			"refresh_interval",
			"position_x",
			"position_y",
			"width",
			"height",
			"data_source[0].rows",
			"visual_settings.columns[col1].width",
			"visual_settings.columns[col1].format.precision",
		}

		for _, field := range expectedFields {
			found := false
			for _, err := range result.Errors {
				if err.Field == field {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("expected error for negative %s", field)
			}
		}
	})
}

// =============================================================================
// Link Action Validation Tests
// =============================================================================

func TestValidateConfig_LinkActionRequiresURL(t *testing.T) {
	config := tablebuilder.Config{
		Title: "Test",
		DataSource: []tablebuilder.DataSource{
			{Source: "test"},
		},
		VisualSettings: tablebuilder.VisualSettings{
			RowActions: map[string]tablebuilder.Action{
				"view": {
					Name:       "view",
					Label:      "View",
					ActionType: "link",
					// Missing URL
				},
			},
		},
	}
	result := config.ValidateConfig()

	found := false
	for _, err := range result.Errors {
		if err.Field == "visual_settings.row_actions[view].url" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected error for missing URL when action_type is 'link'")
	}
}

// =============================================================================
// Permission Actions Validation Tests
// =============================================================================

func TestValidateConfig_PermissionActions(t *testing.T) {
	tests := []struct {
		name        string
		actions     []string
		expectError bool
	}{
		{"valid view", []string{"view"}, false},
		{"valid edit", []string{"edit"}, false},
		{"valid delete", []string{"delete"}, false},
		{"valid export", []string{"export"}, false},
		{"valid create", []string{"create"}, false},
		{"valid adjust", []string{"adjust"}, false},
		{"valid approve", []string{"approve"}, false},
		{"valid reject", []string{"reject"}, false},
		{"valid multiple", []string{"view", "edit", "delete", "export", "create"}, false},
		{"invalid read", []string{"read"}, true},
		{"invalid update", []string{"update"}, true},
		{"mixed valid/invalid", []string{"view", "read"}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := tablebuilder.Config{
				Title: "Test",
				DataSource: []tablebuilder.DataSource{
					{Source: "test"},
				},
				Permissions: tablebuilder.Permissions{
					Actions: tt.actions,
				},
			}
			result := config.ValidateConfig()

			hasPermError := false
			for _, err := range result.Errors {
				if len(err.Field) > 19 && err.Field[:19] == "permissions.actions" {
					hasPermError = true
					break
				}
			}

			if tt.expectError && !hasPermError {
				t.Errorf("expected permission action error for %v", tt.actions)
			}
			if !tt.expectError && hasPermError {
				t.Errorf("unexpected permission action error for %v", tt.actions)
			}
		})
	}
}

// =============================================================================
// Error Interface Tests
// =============================================================================

func TestValidationResult_Error(t *testing.T) {
	t.Run("empty result returns empty string", func(t *testing.T) {
		result := &tablebuilder.ValidationResult{}
		if result.Error() != "" {
			t.Error("expected empty string for no errors")
		}
	})

	t.Run("errors are formatted correctly", func(t *testing.T) {
		result := &tablebuilder.ValidationResult{}
		result.AddError("field1", "message1", "CODE1")
		result.AddError("field2", "message2", "CODE2")

		errStr := result.Error()
		if errStr != "field1: message1; field2: message2" {
			t.Errorf("unexpected error format: %s", errStr)
		}
	})
}

// =============================================================================
// Computed Column Validation Tests
// =============================================================================

func TestValidateConfig_ComputedColumn(t *testing.T) {
	t.Run("computed column requires name", func(t *testing.T) {
		config := tablebuilder.Config{
			Title: "Test",
			DataSource: []tablebuilder.DataSource{
				{
					Source: "test",
					Select: tablebuilder.SelectConfig{
						ClientComputedColumns: []tablebuilder.ComputedColumn{
							{Expression: "a + b"}, // Missing name
						},
					},
				},
			},
		}
		result := config.ValidateConfig()

		found := false
		for _, err := range result.Errors {
			if err.Field == "data_source[0].select.client_computed_columns[0].name" {
				found = true
				break
			}
		}
		if !found {
			t.Error("expected error for missing computed column name")
		}
	})

	t.Run("computed column requires expression", func(t *testing.T) {
		config := tablebuilder.Config{
			Title: "Test",
			DataSource: []tablebuilder.DataSource{
				{
					Source: "test",
					Select: tablebuilder.SelectConfig{
						ClientComputedColumns: []tablebuilder.ComputedColumn{
							{Name: "computed"}, // Missing expression
						},
					},
				},
			},
		}
		result := config.ValidateConfig()

		found := false
		for _, err := range result.Errors {
			if err.Field == "data_source[0].select.client_computed_columns[0].expression" {
				found = true
				break
			}
		}
		if !found {
			t.Error("expected error for missing computed column expression")
		}
	})
}

// =============================================================================
// Conditional Format Validation Tests
// =============================================================================

func TestValidateConfig_ConditionalFormat(t *testing.T) {
	t.Run("requires column", func(t *testing.T) {
		config := tablebuilder.Config{
			Title: "Test",
			DataSource: []tablebuilder.DataSource{
				{Source: "test"},
			},
			VisualSettings: tablebuilder.VisualSettings{
				ConditionalFormatting: []tablebuilder.ConditionalFormat{
					{Condition: "eq", Value: "test"},
				},
			},
		}
		result := config.ValidateConfig()

		found := false
		for _, err := range result.Errors {
			if err.Field == "visual_settings.conditional_formatting[0].column" {
				found = true
				break
			}
		}
		if !found {
			t.Error("expected error for missing column in conditional format")
		}
	})

	t.Run("validates condition operator", func(t *testing.T) {
		config := tablebuilder.Config{
			Title: "Test",
			DataSource: []tablebuilder.DataSource{
				{Source: "test"},
			},
			VisualSettings: tablebuilder.VisualSettings{
				ConditionalFormatting: []tablebuilder.ConditionalFormat{
					{Column: "status", Condition: "invalid", Value: "test"},
				},
			},
		}
		result := config.ValidateConfig()

		found := false
		for _, err := range result.Errors {
			if err.Field == "visual_settings.conditional_formatting[0].condition" {
				found = true
				break
			}
		}
		if !found {
			t.Error("expected error for invalid condition in conditional format")
		}
	})
}

// =============================================================================
// Chart Widget Skip Validation Tests
// =============================================================================

func TestValidateConfig_ChartWidgetSkipsColumnTypeValidation(t *testing.T) {
	config := tablebuilder.Config{
		Title:      "Test Chart",
		WidgetType: "chart",
		DataSource: []tablebuilder.DataSource{
			{Source: "test"},
		},
		// No VisualSettings.Columns - should not error for chart widgets
	}
	result := config.ValidateConfig()

	// Should not have any column type errors
	for _, err := range result.Errors {
		if err.Code == "REQUIRED" && err.Field == "visual_settings.columns" {
			t.Error("chart widgets should not require visual_settings.columns")
		}
	}
}

// =============================================================================
// Date Format Validation Tests
// =============================================================================

func TestValidateDateFormatString(t *testing.T) {
	validFormats := []string{
		"yyyy-MM-dd",
		"MM-dd-yyyy",
		"yyyy-MM-dd HH:mm",
		"yyyy-MM-dd HH:mm:ss",
		"dd/MM/yyyy",
		"MMM dd, yyyy",
		"EEEE, MMMM do, yyyy",
		"HH:mm:ss",
		"hh:mm a",
		"yyyy-MM-dd'T'HH:mm:ss",
		"", // Empty is valid
	}

	for _, format := range validFormats {
		t.Run("valid_"+format, func(t *testing.T) {
			err := tablebuilder.ValidateDateFormatString(format)
			if err != nil {
				t.Errorf("ValidateDateFormatString(%q) unexpected error: %v", format, err)
			}
		})
	}
}

func TestValidateDateFormatString_RejectsGoFormats(t *testing.T) {
	goFormats := []struct {
		name   string
		format string
	}{
		{"basic date", "2006-01-02"},
		{"US date", "01-02-2006"},
		{"datetime", "2006-01-02 15:04"},
		{"full datetime", "2006-01-02 15:04:05"},
		{"time only", "15:04:05"},
		{"ISO datetime", "2006-01-02T15:04:05"},
	}

	for _, tt := range goFormats {
		t.Run(tt.name, func(t *testing.T) {
			err := tablebuilder.ValidateDateFormatString(tt.format)
			if err == nil {
				t.Errorf("ValidateDateFormatString(%q) expected error for Go format, got nil", tt.format)
			}
			if err != nil && !errors.Is(err, tablebuilder.ErrGoDateFormatDetected) {
				t.Errorf("ValidateDateFormatString(%q) expected ErrGoDateFormatDetected, got: %v", tt.format, err)
			}
		})
	}
}

func TestValidateDateFormatString_RejectsInvalidTokens(t *testing.T) {
	invalidFormats := []struct {
		name   string
		format string
	}{
		{"invalid token YYYY", "YYYY-mm-dd"}, // YYYY is not valid, should be yyyy
		{"invalid quarter", "QQQ"},           // Quarter tokens not supported
		{"invalid week", "ww"},               // Week tokens not supported
	}

	for _, tt := range invalidFormats {
		t.Run(tt.name, func(t *testing.T) {
			err := tablebuilder.ValidateDateFormatString(tt.format)
			if err == nil {
				t.Errorf("ValidateDateFormatString(%q) expected error for invalid token, got nil", tt.format)
			}
		})
	}
}

func TestValidateConfig_DateFormatInFormatConfig(t *testing.T) {
	tests := []struct {
		name        string
		format      string
		expectError bool
	}{
		{"valid date-fns format", "yyyy-MM-dd", false},
		{"valid datetime format", "yyyy-MM-dd HH:mm:ss", false},
		{"Go date format rejected", "2006-01-02", true},
		{"Go datetime format rejected", "2006-01-02 15:04:05", true},
		{"empty format valid", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := tablebuilder.Config{
				Title: "Test",
				DataSource: []tablebuilder.DataSource{
					{Source: "test"},
				},
				VisualSettings: tablebuilder.VisualSettings{
					Columns: map[string]tablebuilder.ColumnConfig{
						"date_col": {
							Name: "date_col",
							Type: "datetime",
							Format: &tablebuilder.FormatConfig{
								Type:   "date",
								Format: tt.format,
							},
						},
					},
				},
			}
			result := config.ValidateConfig()

			hasFormatError := false
			for _, err := range result.Errors {
				if err.Field == "visual_settings.columns[date_col].format.format" {
					hasFormatError = true
					break
				}
			}

			if tt.expectError && !hasFormatError {
				t.Errorf("expected format error for %q, got none", tt.format)
			}
			if !tt.expectError && hasFormatError {
				t.Errorf("unexpected format error for %q", tt.format)
			}
		})
	}
}
