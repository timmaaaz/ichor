package tablebuilder

import (
	"errors"
	"testing"
)

func Test_IsValidColumnType(t *testing.T) {
	validTypes := []string{"string", "number", "datetime", "boolean", "uuid", "status", "computed", "lookup"}
	invalidTypes := []string{"", "invalid", "text", "integer", "int", "date", "VARCHAR"}

	for _, typ := range validTypes {
		t.Run("valid_"+typ, func(t *testing.T) {
			if !IsValidColumnType(typ) {
				t.Errorf("IsValidColumnType(%q) = false, want true", typ)
			}
		})
	}

	for _, typ := range invalidTypes {
		t.Run("invalid_"+typ, func(t *testing.T) {
			if IsValidColumnType(typ) {
				t.Errorf("IsValidColumnType(%q) = true, want false", typ)
			}
		})
	}
}

func Test_ConfigValidateColumnTypes(t *testing.T) {
	tests := []struct {
		name        string
		config      Config
		wantErr     bool
		errContains error
	}{
		{
			name: "valid config with all types",
			config: Config{
				Title: "Test Config",
				DataSource: []DataSource{
					{
						Source: "test_table",
						Schema: "test",
						Select: SelectConfig{
							Columns: []ColumnDefinition{
								{Name: "id"},
								{Name: "name"},
							},
						},
					},
				},
				VisualSettings: VisualSettings{
					Columns: map[string]ColumnConfig{
						"id":   {Type: "uuid"},
						"name": {Type: "string"},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "missing type for column",
			config: Config{
				Title: "Test Config",
				DataSource: []DataSource{
					{
						Source: "test_table",
						Schema: "test",
						Select: SelectConfig{
							Columns: []ColumnDefinition{
								{Name: "id"},
								{Name: "name"},
							},
						},
					},
				},
				VisualSettings: VisualSettings{
					Columns: map[string]ColumnConfig{
						"id": {Type: "uuid"},
						// "name" is missing
					},
				},
			},
			wantErr:     true,
			errContains: ErrMissingColumnType,
		},
		{
			name: "empty type for column",
			config: Config{
				Title: "Test Config",
				DataSource: []DataSource{
					{
						Source: "test_table",
						Schema: "test",
						Select: SelectConfig{
							Columns: []ColumnDefinition{
								{Name: "id"},
								{Name: "name"},
							},
						},
					},
				},
				VisualSettings: VisualSettings{
					Columns: map[string]ColumnConfig{
						"id":   {Type: "uuid"},
						"name": {Type: ""},
					},
				},
			},
			wantErr:     true,
			errContains: ErrMissingColumnType,
		},
		{
			name: "invalid type for column",
			config: Config{
				Title: "Test Config",
				DataSource: []DataSource{
					{
						Source: "test_table",
						Schema: "test",
						Select: SelectConfig{
							Columns: []ColumnDefinition{
								{Name: "id"},
								{Name: "name"},
							},
						},
					},
				},
				VisualSettings: VisualSettings{
					Columns: map[string]ColumnConfig{
						"id":   {Type: "uuid"},
						"name": {Type: "invalid_type"},
					},
				},
			},
			wantErr:     true,
			errContains: ErrInvalidColumn,
		},
		{
			name: "missing type for computed column",
			config: Config{
				Title: "Test Config",
				DataSource: []DataSource{
					{
						Source: "test_table",
						Schema: "test",
						Select: SelectConfig{
							Columns: []ColumnDefinition{
								{Name: "id"},
							},
							ClientComputedColumns: []ComputedColumn{
								{Name: "computed_field", Expression: "id + 1"},
							},
						},
					},
				},
				VisualSettings: VisualSettings{
					Columns: map[string]ColumnConfig{
						"id": {Type: "uuid"},
						// "computed_field" is missing
					},
				},
			},
			wantErr:     true,
			errContains: ErrMissingColumnType,
		},
		{
			name: "missing type for foreign table column",
			config: Config{
				Title: "Test Config",
				DataSource: []DataSource{
					{
						Source: "test_table",
						Schema: "test",
						Select: SelectConfig{
							Columns: []ColumnDefinition{
								{Name: "id"},
							},
							ForeignTables: []ForeignTable{
								{
									Table:  "foreign_table",
									Schema: "test",
									Columns: []ColumnDefinition{
										{Name: "foreign_col"},
									},
								},
							},
						},
					},
				},
				VisualSettings: VisualSettings{
					Columns: map[string]ColumnConfig{
						"id": {Type: "uuid"},
						// "foreign_col" is missing
					},
				},
			},
			wantErr:     true,
			errContains: ErrMissingColumnType,
		},
		{
			name: "chart widget skips column type validation",
			config: Config{
				Title:         "Test Chart",
				WidgetType:    "chart",
				Visualization: "gantt",
				DataSource: []DataSource{
					{
						Source: "test_table",
						Schema: "test",
						Select: SelectConfig{
							Columns: []ColumnDefinition{
								{Name: "id"},
								{Name: "name"},
								{Name: "start_date"},
								{Name: "end_date"},
							},
						},
					},
				},
				VisualSettings: VisualSettings{
					Columns: map[string]ColumnConfig{
						"_chart": {CellTemplate: `{"chartType":"gantt"}`},
						// No column types defined - should be OK for charts
					},
				},
			},
			wantErr: false, // Chart widgets skip column type validation
		},
		{
			name: "datetime column without format config",
			config: Config{
				Title: "Test Config",
				DataSource: []DataSource{
					{
						Source: "test_table",
						Schema: "test",
						Select: SelectConfig{
							Columns: []ColumnDefinition{
								{Name: "id"},
								{Name: "created_date"},
							},
						},
					},
				},
				VisualSettings: VisualSettings{
					Columns: map[string]ColumnConfig{
						"id":           {Type: "uuid"},
						"created_date": {Type: "datetime"}, // Missing Format config
					},
				},
			},
			wantErr:     true,
			errContains: ErrMissingDatetimeFormat,
		},
		{
			name: "datetime column with format config",
			config: Config{
				Title: "Test Config",
				DataSource: []DataSource{
					{
						Source: "test_table",
						Schema: "test",
						Select: SelectConfig{
							Columns: []ColumnDefinition{
								{Name: "id"},
								{Name: "created_date"},
							},
						},
					},
				},
				VisualSettings: VisualSettings{
					Columns: map[string]ColumnConfig{
						"id": {Type: "uuid"},
						"created_date": {
							Type: "datetime",
							Format: &FormatConfig{
								Type:   "date",
								Format: "yyyy-MM-dd",
							},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "datetime foreign table column without format config",
			config: Config{
				Title: "Test Config",
				DataSource: []DataSource{
					{
						Source: "test_table",
						Schema: "test",
						Select: SelectConfig{
							Columns: []ColumnDefinition{
								{Name: "id"},
							},
							ForeignTables: []ForeignTable{
								{
									Table:            "foreign_table",
									Schema:           "test",
									RelationshipFrom: "id",
									RelationshipTo:   "foreign_id",
									Columns: []ColumnDefinition{
										{Name: "foreign_date"},
									},
								},
							},
						},
					},
				},
				VisualSettings: VisualSettings{
					Columns: map[string]ColumnConfig{
						"id":           {Type: "uuid"},
						"foreign_date": {Type: "datetime"}, // Missing Format config
					},
				},
			},
			wantErr:     true,
			errContains: ErrMissingDatetimeFormat,
		},
		{
			name: "datetime computed column without format config",
			config: Config{
				Title: "Test Config",
				DataSource: []DataSource{
					{
						Source: "test_table",
						Schema: "test",
						Select: SelectConfig{
							Columns: []ColumnDefinition{
								{Name: "id"},
							},
							ClientComputedColumns: []ComputedColumn{
								{Name: "computed_date", Expression: "NOW()"},
							},
						},
					},
				},
				VisualSettings: VisualSettings{
					Columns: map[string]ColumnConfig{
						"id":            {Type: "uuid"},
						"computed_date": {Type: "datetime"}, // Missing Format config
					},
				},
			},
			wantErr:     true,
			errContains: ErrMissingDatetimeFormat,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantErr {
				if err == nil {
					t.Errorf("Validate() expected error, got nil")
					return
				}
				if tt.errContains != nil && !errors.Is(err, tt.errContains) {
					t.Errorf("Validate() error = %v, want error containing %v", err, tt.errContains)
				}
			} else {
				if err != nil {
					t.Errorf("Validate() unexpected error: %v", err)
				}
			}
		})
	}
}

func Test_MapPostgreSQLType(t *testing.T) {
	tests := []struct {
		name     string
		pgType   string
		expected string
	}{
		// UUID
		{"uuid", "uuid", "uuid"},
		{"UUID uppercase", "UUID", "uuid"},

		// Date/time
		{"timestamp with time zone", "timestamp with time zone", "datetime"},
		{"timestamp without time zone", "timestamp without time zone", "datetime"},
		{"timestamp plain", "timestamp", "datetime"},
		{"date", "date", "datetime"},
		{"time", "time", "datetime"},
		{"time without time zone", "time without time zone", "datetime"},
		{"time with time zone", "time with time zone", "datetime"},
		{"interval", "interval", "datetime"},

		// Numeric
		{"integer", "integer", "number"},
		{"bigint", "bigint", "number"},
		{"smallint", "smallint", "number"},
		{"numeric", "numeric", "number"},
		{"numeric with precision", "numeric(10,2)", "number"},
		{"decimal", "decimal", "number"},
		{"decimal with precision", "decimal(18,4)", "number"},
		{"real", "real", "number"},
		{"double precision", "double precision", "number"},
		{"serial", "serial", "number"},
		{"bigserial", "bigserial", "number"},
		{"smallserial", "smallserial", "number"},
		{"money", "money", "number"},

		// Boolean
		{"boolean", "boolean", "boolean"},
		{"boolean uppercase", "BOOLEAN", "boolean"},

		// Text/character
		{"character varying", "character varying", "string"},
		{"varchar with length", "character varying(255)", "string"},
		{"character", "character", "string"},
		{"char with length", "character(10)", "string"},
		{"text", "text", "string"},
		{"citext", "citext", "string"},

		// JSON
		{"json", "json", "string"},
		{"jsonb", "jsonb", "string"},

		// Array types (default to string)
		{"integer array", "integer[]", "string"},
		{"text array", "text[]", "string"},
		{"text ARRAY", "text ARRAY", "string"},
		{"varchar array", "character varying[]", "string"},

		// Unknown types (default to string)
		{"unknown geometry", "geometry", "string"},
		{"unknown point", "point", "string"},
		{"unknown bytea", "bytea", "string"},
		{"unknown inet", "inet", "string"},
		{"unknown macaddr", "macaddr", "string"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := mapPostgreSQLType(tt.pgType)
			if result != tt.expected {
				t.Errorf("mapPostgreSQLType(%q) = %q, want %q",
					tt.pgType, result, tt.expected)
			}
		})
	}
}
