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
	Timezone             *string
	PreferredContactType *string
	Notes                *string
}
