package contactinfobus

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
	Address              *string
	AvailableHoursStart  *string
	AvailableHoursEnd    *string
	Timezone             *string
	PreferredContactType *string
	Notes                *string
}
