package countrybus

import (
	"github.com/google/uuid"
)

// NOTE: Countries are special. They are controlled solely in the database and
// therefore should have ONLY retrive actions available. No create, update, or
// delete actions are allowed. We want only the highest level admins to have any
// way to touch this.

// Country represents information about an individual country.
type Country struct {
	ID     uuid.UUID
	Number int
	Name   string
	Alpha2 string
	Alpha3 string
}
