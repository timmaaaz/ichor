package pageactionbus

import "github.com/google/uuid"

// QueryFilter holds the available fields a query can be filtered on.
type QueryFilter struct {
	ID           *uuid.UUID
	PageConfigID *uuid.UUID
	ActionType   *ActionType
	IsActive     *bool
}
