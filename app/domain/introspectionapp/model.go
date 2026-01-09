package introspectionapp

import (
	"encoding/json"

	"github.com/timmaaaz/ichor/business/domain/introspectionbus"
)

// Schema represents a database schema.
type Schema struct {
	Name string `json:"name"`
}

// Encode implements the encoder interface.
func (app Schema) Encode() ([]byte, string, error) {
	data, err := json.Marshal(app)
	return data, "application/json", err
}

// Schemas is a collection wrapper that implements the Encoder interface.
type Schemas []Schema

// Encode implements the encoder interface.
func (app Schemas) Encode() ([]byte, string, error) {
	data, err := json.Marshal(app)
	return data, "application/json", err
}

// Table represents a database table.
type Table struct {
	Schema           string `json:"schema"`
	Name             string `json:"name"`
	RowCountEstimate *int   `json:"rowCountEstimate"`
}

// Encode implements the encoder interface.
func (app Table) Encode() ([]byte, string, error) {
	data, err := json.Marshal(app)
	return data, "application/json", err
}

// Tables is a collection wrapper that implements the Encoder interface.
type Tables []Table

// Encode implements the encoder interface.
func (app Tables) Encode() ([]byte, string, error) {
	data, err := json.Marshal(app)
	return data, "application/json", err
}

// Column represents a table column.
type Column struct {
	Name         string `json:"name"`
	DataType     string `json:"dataType"`
	IsNullable   bool   `json:"isNullable"`
	IsPrimaryKey bool   `json:"isPrimaryKey"`
	DefaultValue string `json:"defaultValue"`
}

// Encode implements the encoder interface.
func (app Column) Encode() ([]byte, string, error) {
	data, err := json.Marshal(app)
	return data, "application/json", err
}

// Columns is a collection wrapper that implements the Encoder interface.
type Columns []Column

// Encode implements the encoder interface.
func (app Columns) Encode() ([]byte, string, error) {
	data, err := json.Marshal(app)
	return data, "application/json", err
}

// Relationship represents a foreign key relationship.
type Relationship struct {
	ForeignKeyName   string `json:"foreignKeyName"`
	ColumnName       string `json:"columnName"`
	ReferencedSchema string `json:"referencedSchema"`
	ReferencedTable  string `json:"referencedTable"`
	ReferencedColumn string `json:"referencedColumn"`
	RelationshipType string `json:"relationshipType"`
}

// Encode implements the encoder interface.
func (app Relationship) Encode() ([]byte, string, error) {
	data, err := json.Marshal(app)
	return data, "application/json", err
}

// Relationships is a collection wrapper that implements the Encoder interface.
type Relationships []Relationship

// Encode implements the encoder interface.
func (app Relationships) Encode() ([]byte, string, error) {
	data, err := json.Marshal(app)
	return data, "application/json", err
}

// ReferencingTable represents a table that has a foreign key pointing to another table.
type ReferencingTable struct {
	Schema           string `json:"schema"`
	Table            string `json:"table"`
	ForeignKeyColumn string `json:"foreignKeyColumn"`
	ConstraintName   string `json:"constraintName"`
}

// Encode implements the encoder interface.
func (app ReferencingTable) Encode() ([]byte, string, error) {
	data, err := json.Marshal(app)
	return data, "application/json", err
}

// ReferencingTables is a collection wrapper that implements the Encoder interface.
type ReferencingTables []ReferencingTable

// Encode implements the encoder interface.
func (app ReferencingTables) Encode() ([]byte, string, error) {
	data, err := json.Marshal(app)
	return data, "application/json", err
}

// =============================================================================
// Conversion functions

// ToAppSchema converts a business Schema to app Schema.
func ToAppSchema(bus introspectionbus.Schema) Schema {
	return Schema{
		Name: bus.Name,
	}
}

// ToAppSchemas converts a slice of business Schemas to app Schemas.
func ToAppSchemas(bus []introspectionbus.Schema) []Schema {
	schemas := make([]Schema, len(bus))
	for i, s := range bus {
		schemas[i] = ToAppSchema(s)
	}
	return schemas
}

// ToAppTable converts a business Table to app Table.
func ToAppTable(bus introspectionbus.Table) Table {
	return Table{
		Schema:           bus.Schema,
		Name:             bus.Name,
		RowCountEstimate: bus.RowCountEstimate,
	}
}

// ToAppTables converts a slice of business Tables to app Tables.
func ToAppTables(bus []introspectionbus.Table) []Table {
	tables := make([]Table, len(bus))
	for i, t := range bus {
		tables[i] = ToAppTable(t)
	}
	return tables
}

// ToAppColumn converts a business Column to app Column.
func ToAppColumn(bus introspectionbus.Column) Column {
	return Column{
		Name:         bus.Name,
		DataType:     bus.DataType,
		IsNullable:   bus.IsNullable,
		IsPrimaryKey: bus.IsPrimaryKey,
		DefaultValue: bus.DefaultValue,
	}
}

// ToAppColumns converts a slice of business Columns to app Columns.
func ToAppColumns(bus []introspectionbus.Column) []Column {
	columns := make([]Column, len(bus))
	for i, c := range bus {
		columns[i] = ToAppColumn(c)
	}
	return columns
}

// ToAppRelationship converts a business Relationship to app Relationship.
func ToAppRelationship(bus introspectionbus.Relationship) Relationship {
	return Relationship{
		ForeignKeyName:   bus.ForeignKeyName,
		ColumnName:       bus.ColumnName,
		ReferencedSchema: bus.ReferencedSchema,
		ReferencedTable:  bus.ReferencedTable,
		ReferencedColumn: bus.ReferencedColumn,
		RelationshipType: bus.RelationshipType,
	}
}

// ToAppRelationships converts a slice of business Relationships to app Relationships.
func ToAppRelationships(bus []introspectionbus.Relationship) []Relationship {
	relationships := make([]Relationship, len(bus))
	for i, r := range bus {
		relationships[i] = ToAppRelationship(r)
	}
	return relationships
}

// ToAppReferencingTable converts a business ReferencingTable to app ReferencingTable.
func ToAppReferencingTable(bus introspectionbus.ReferencingTable) ReferencingTable {
	return ReferencingTable{
		Schema:           bus.Schema,
		Table:            bus.Table,
		ForeignKeyColumn: bus.ForeignKeyColumn,
		ConstraintName:   bus.ConstraintName,
	}
}

// ToAppReferencingTables converts a slice of business ReferencingTables to app ReferencingTables.
func ToAppReferencingTables(bus []introspectionbus.ReferencingTable) []ReferencingTable {
	tables := make([]ReferencingTable, len(bus))
	for i, t := range bus {
		tables[i] = ToAppReferencingTable(t)
	}
	return tables
}
