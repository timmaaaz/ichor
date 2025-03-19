package supplierbus

import (
	"time"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/supplier/supplierbus/types"
)

type Supplier struct {
	SupplierID   uuid.UUID
	ContactID    uuid.UUID
	Name         string
	PaymentTerms string
	LeadTimeDays int
	Rating       types.RoundedFloat
	IsActive     bool
	CreatedDate  time.Time
	UpdatedDate  time.Time
}

type NewSupplier struct {
	ContactID    uuid.UUID
	Name         string
	PaymentTerms string
	LeadTimeDays int
	Rating       types.RoundedFloat
	IsActive     bool
}

type UpdateSupplier struct {
	ContactID    *uuid.UUID
	Name         *string
	PaymentTerms *string
	LeadTimeDays *int
	Rating       *types.RoundedFloat
	IsActive     *bool
}
