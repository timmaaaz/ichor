package orderfulfillmentstatusbus

import "github.com/google/uuid"

type OrderFulfillmentStatus struct {
	ID             uuid.UUID
	Name           string
	Description    string
	PrimaryColor   string
	SecondaryColor string
	Icon           string
}

type NewOrderFulfillmentStatus struct {
	Name           string
	Description    string
	PrimaryColor   string
	SecondaryColor string
	Icon           string
}

type UpdateOrderFulfillmentStatus struct {
	Name           *string
	Description    *string
	PrimaryColor   *string
	SecondaryColor *string
	Icon           *string
}
