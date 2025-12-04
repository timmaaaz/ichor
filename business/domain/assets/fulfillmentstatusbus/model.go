package fulfillmentstatusbus

import "github.com/google/uuid"

type FulfillmentStatus struct {
	ID             uuid.UUID
	Name           string
	IconID         uuid.UUID
	PrimaryColor   string
	SecondaryColor string
	Icon           string
}

type NewFulfillmentStatus struct {
	Name           string
	IconID         uuid.UUID
	PrimaryColor   string
	SecondaryColor string
	Icon           string
}

type UpdateFulfillmentStatus struct {
	Name           *string
	IconID         *uuid.UUID
	PrimaryColor   *string
	SecondaryColor *string
	Icon           *string
}
