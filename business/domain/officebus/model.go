package officebus

import "github.com/google/uuid"

type Office struct {
	ID       uuid.UUID
	Name     string
	StreetID uuid.UUID
}

type NewOffice struct {
	Name     string
	StreetID uuid.UUID
}

type UpdateOffice struct {
	Name     *string
	StreetID *uuid.UUID
}
