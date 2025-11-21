package formfieldbus

import "github.com/google/uuid"

// QueryFilter holds the available fields a query can be filtered on.
type QueryFilter struct {
	ID           *uuid.UUID
	FormID       *uuid.UUID
	EntitySchema *string
	EntityTable  *string
	Name         *string
	FieldType    *string
	Required     *bool
}
