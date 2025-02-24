package restrictedcolumnbus

import "github.com/google/uuid"

// RestrictedColumn represents information about an individual restricted column.
type RestrictedColumn struct {
	ID         uuid.UUID
	TableName  string
	ColumnName string
}

// NewRestrictedColumn contains information needed to create a new restricted
// column.
type NewRestrictedColumn struct {
	TableName  string
	ColumnName string
}
