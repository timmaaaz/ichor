package citybus

import "github.com/google/uuid"

// City represents information about an individual city.
type City struct {
	ID       uuid.UUID
	RegionID uuid.UUID
	Name     string
}

type NewCity struct {
	RegionID uuid.UUID
	Name     string
}

type UpdateCity struct {
	RegionID *uuid.UUID
	Name     *string
}
