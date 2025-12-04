package lineitemfulfillmentstatusbus

import "github.com/google/uuid"

type LineItemFulfillmentStatus struct {
	ID             uuid.UUID
	Name           string
	Description    string
	PrimaryColor   string
	SecondaryColor string
	Icon           string
}

type NewLineItemFulfillmentStatus struct {
	Name           string
	Description    string
	PrimaryColor   string
	SecondaryColor string
	Icon           string
}

type UpdateLineItemFulfillmentStatus struct {
	Name           *string
	Description    *string
	PrimaryColor   *string
	SecondaryColor *string
	Icon           *string
}
