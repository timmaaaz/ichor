package streetbus

import "github.com/google/uuid"

// Street represents information about an individual street.
type Street struct {
	ID         uuid.UUID
	CityID     uuid.UUID
	Line1      string
	Line2      string
	PostalCode string
}

type NewStreet struct {
	CityID     uuid.UUID
	Line1      string
	Line2      string
	PostalCode string
}

type UpdateStreet struct {
	CityID     *uuid.UUID
	Line1      *string
	Line2      *string
	PostalCode *string
}
