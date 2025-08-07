package supplierbus

import (
	"time"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/supplier/supplierbus/types"
)

type QueryFilter struct {
	SupplierID    *uuid.UUID
	ContactInfoID *uuid.UUID
	Name          *string
	PaymentTerms  *string
	LeadTimeDays  *int
	Rating        *types.RoundedFloat
	IsActive      *bool
	CreatedDate   *time.Time
	UpdatedDate   *time.Time
}
