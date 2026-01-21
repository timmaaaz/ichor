package productcostbus

import (
	"time"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/products/productcostbus/types"
)

type QueryFilter struct {
	ID                *uuid.UUID
	ProductID         *uuid.UUID
	PurchaseCost      *types.Money
	SellingPrice      *types.Money
	CurrencyID        *uuid.UUID
	MSRP              *types.Money
	MarkupPercentage  *types.RoundedFloat
	LandedCost        *types.Money
	CarryingCost      *types.Money
	ABCClassification *string
	DepreciationValue *types.RoundedFloat
	InsuranceValue    *types.Money
	EffectiveDate     *time.Time
	CreatedDate       *time.Time
	UpdatedDate       *time.Time
}
