package introspectionbus

import "github.com/timmaaaz/ichor/business/sdk/sqldb/dbarray"

// Schema represents a PostgreSQL schema.
type Schema struct {
	Name string `db:"name"`
}

// Table represents a table within a schema.
type Table struct {
	Schema           string `db:"schema"`
	Name             string `db:"name"`
	RowCountEstimate *int   `db:"row_count_estimate"`
}

// Column represents a column within a table.
type Column struct {
	Name         string `db:"name"`
	DataType     string `db:"data_type"`
	IsNullable   bool   `db:"is_nullable"`
	IsPrimaryKey bool   `db:"is_primary_key"`
	DefaultValue string `db:"default_value"`
	// Foreign key metadata (NULL if not a FK)
	IsForeignKey     bool    `db:"is_foreign_key"`
	ReferencedSchema *string `db:"referenced_schema"`
	ReferencedTable  *string `db:"referenced_table"`
	ReferencedColumn *string `db:"referenced_column"`
}

// Relationship represents a foreign key relationship.
type Relationship struct {
	ForeignKeyName   string `db:"foreign_key_name"`
	ColumnName       string `db:"column_name"`
	ReferencedSchema string `db:"referenced_schema"`
	ReferencedTable  string `db:"referenced_table"`
	ReferencedColumn string `db:"referenced_column"`
	RelationshipType string `db:"relationship_type"` // "many-to-one", "one-to-many", etc.
}

// ReferencingTable represents a table that has a foreign key pointing to another table.
type ReferencingTable struct {
	Schema           string `db:"schema"`
	Table            string `db:"table"`
	ForeignKeyColumn string `db:"fk_column"`
	ConstraintName   string `db:"constraint_name"`
}

// EnumType represents a PostgreSQL ENUM type with its values.
type EnumType struct {
	Name   string        `db:"enum_name"`
	Schema string        `db:"schema_name"`
	Values dbarray.String `db:"values"` // Aggregated from pg_enum using array_agg
}

// EnumValue represents a single value within a PostgreSQL ENUM type.
type EnumValue struct {
	Value     string `db:"value"`
	SortOrder int    `db:"sort_order"`
}

// EnumLabel represents a human-friendly label for an enum value.
type EnumLabel struct {
	EnumName  string `db:"enum_name"`
	Value     string `db:"value"`
	Label     string `db:"label"`
	SortOrder int    `db:"sort_order"`
}

// EnumOption represents a ready-to-use dropdown option (merged enum value + label).
type EnumOption struct {
	Value     string `json:"value"`
	Label     string `json:"label"`
	SortOrder int    `json:"sortOrder"`
}
