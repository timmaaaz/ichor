package contactinfosbus

import (
	"github.com/google/uuid"
)

type ContactInfos struct {
	ID                   uuid.UUID
	FirstName            string
	LastName             string
	EmailAddress         string
	PrimaryPhone         string
	SecondaryPhone       string
	Address              string
	AvailableHoursStart  string
	AvailableHoursEnd    string
	Timezone             string
	PreferredContactType string
	Notes                string
}

type NewContactInfos struct {
	FirstName            string
	LastName             string
	EmailAddress         string
	PrimaryPhone         string
	SecondaryPhone       string
	Address              string
	AvailableHoursStart  string
	AvailableHoursEnd    string
	Timezone             string
	PreferredContactType string
	Notes                string
}

type UpdateContactInfos struct {
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
