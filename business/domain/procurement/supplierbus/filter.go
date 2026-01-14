package supplierbus

import (
	"time"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/procurement/supplierbus/types"
)

type QueryFilter struct {
	SupplierID     *uuid.UUID
	ContactInfosID *uuid.UUID
	Name           *string
	PaymentTermID  *uuid.UUID
	LeadTimeDays   *int
	Rating         *types.RoundedFloat
	IsActive       *bool
	CreatedDate    *time.Time
	UpdatedDate    *time.Time
}
