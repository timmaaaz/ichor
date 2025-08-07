package supplierbus

import (
	"time"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/supplier/supplierbus/types"
)

type Supplier struct {
	SupplierID    uuid.UUID
	ContactInfoID uuid.UUID
	Name          string
	PaymentTerms  string
	LeadTimeDays  int
	Rating        types.RoundedFloat
	IsActive      bool
	CreatedDate   time.Time
	UpdatedDate   time.Time
}

type NewSupplier struct {
	ContactInfoID uuid.UUID
	Name          string
	PaymentTerms  string
	LeadTimeDays  int
	Rating        types.RoundedFloat
	IsActive      bool
}

type UpdateSupplier struct {
	ContactInfoID *uuid.UUID
	Name          *string
	PaymentTerms  *string
	LeadTimeDays  *int
	Rating        *types.RoundedFloat
	IsActive      *bool
}
