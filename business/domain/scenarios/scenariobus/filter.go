package scenariobus

import "github.com/google/uuid"

// QueryFilter holds the available fields for scenario queries.
type QueryFilter struct {
	ID   *uuid.UUID
	Name *string // prefix search for the admin UI
}
