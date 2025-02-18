package contactinfobus

import (
	"github.com/google/uuid"
)

type ContactInfo struct {
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

type NewContactInfo struct {
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

type UpdateContactInfo struct {
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
