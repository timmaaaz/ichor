package customersbus

import (
	"time"

	"github.com/google/uuid"
)

type Customers struct {
	ID                uuid.UUID
	Name              string
	ContactID         uuid.UUID
	DeliveryAddressID uuid.UUID
	Notes             string
	CreatedBy         uuid.UUID
	UpdatedBy         uuid.UUID
	CreatedDate       time.Time
	UpdatedDate       time.Time
}

type NewCustomers struct {
	Name              string
	ContactID         uuid.UUID
	DeliveryAddressID uuid.UUID
	Notes             string
	CreatedBy         uuid.UUID
}

type UpdateCustomers struct {
	Name              *string
	ContactID         *uuid.UUID
	DeliveryAddressID *uuid.UUID
	Notes             *string
	UpdatedBy         *uuid.UUID
}
