package contactinfosbus

import (
	"github.com/google/uuid"
)

type QueryFilter struct {
	ID                   *uuid.UUID
	FirstName            *string
	LastName             *string
	EmailAddress         *string
	PrimaryPhone         *string
	SecondaryPhone       *string
	StreetID             *uuid.UUID
	DeliveryAddressID    *uuid.UUID
	AvailableHoursStart  *string
	AvailableHoursEnd    *string
	TimezoneID           *uuid.UUID
	PreferredContactType *string
	Notes                *string
}
