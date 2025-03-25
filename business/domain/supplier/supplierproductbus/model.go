package supplierproductbus

import (
	"time"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/supplier/supplierproductbus/types"
)

type SupplierProduct struct {
	SupplierProductID  uuid.UUID
	SupplierID         uuid.UUID
	ProductID          uuid.UUID
	SupplierPartNumber string
	MinOrderQuantity   int
	MaxOrderQuantity   int
	LeadTimeDays       int
	UnitCost           types.Money
	IsPrimarySupplier  bool
	CreatedDate        time.Time
	UpdatedDate        time.Time
}

type NewSupplierProduct struct {
	SupplierID         uuid.UUID
	ProductID          uuid.UUID
	SupplierPartNumber string
	MinOrderQuantity   int
	MaxOrderQuantity   int
	LeadTimeDays       int
	UnitCost           types.Money
	IsPrimarySupplier  bool
}

type UpdateSupplierProduct struct {
	SupplierID         *uuid.UUID
	ProductID          *uuid.UUID
	SupplierPartNumber *string
	MinOrderQuantity   *int
	MaxOrderQuantity   *int
	LeadTimeDays       *int
	UnitCost           *types.Money
	IsPrimarySupplier  *bool
}
