package supplierbus

import (
	"time"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/procurement/supplierbus/types"
)

type Supplier struct {
	SupplierID     uuid.UUID
	ContactInfosID uuid.UUID
	Name           string
	PaymentTerms   string
	LeadTimeDays   int
	Rating         types.RoundedFloat
	IsActive       bool
	CreatedDate    time.Time
	UpdatedDate    time.Time
}

type NewSupplier struct {
	ContactInfosID uuid.UUID
	Name           string
	PaymentTerms   string
	LeadTimeDays   int
	Rating         types.RoundedFloat
	IsActive       bool
}

type UpdateSupplier struct {
	ContactInfosID *uuid.UUID
	Name           *string
	PaymentTerms   *string
	LeadTimeDays   *int
	Rating         *types.RoundedFloat
	IsActive       *bool
}
