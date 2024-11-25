package fulfillmentstatusbus

import "github.com/google/uuid"

type FulfillmentStatus struct {
	ID     uuid.UUID
	Name   string
	IconID uuid.UUID
}

type NewFulfillmentStatus struct {
	Name   string
	IconID uuid.UUID
}

type UpdateFulfillmentStatus struct {
	Name   *string
	IconID *uuid.UUID
}
