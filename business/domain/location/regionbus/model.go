package regionbus

import "github.com/google/uuid"

// NOTE: Regions are special. They are controlled solely in the database and
// therefore should have ONLY retrive actions available. No create, update, or
// delete actions are allowed. We want only the highest level admins to have
// any way to touch this because it denotes areas we support.
type Region struct {
	ID        uuid.UUID
	CountryID uuid.UUID
	Name      string
	Code      string
}
