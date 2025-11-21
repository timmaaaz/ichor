package introspectionbus

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
