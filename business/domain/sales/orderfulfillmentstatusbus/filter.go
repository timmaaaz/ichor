package orderfulfillmentstatusbus

import "github.com/google/uuid"

type QueryFilter struct {
	ID          *uuid.UUID
	Name        *string
	Description *string
}
