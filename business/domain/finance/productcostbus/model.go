package productcostbus

import (
	"time"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/finance/productcostbus/types"
)

type ProductCost struct {
	ID                uuid.UUID
	ProductID         uuid.UUID
	PurchaseCost      types.Money
	SellingPrice      types.Money
	Currency          string
	MSRP              types.Money
	MarkupPercentage  types.RoundedFloat
	LandedCost        types.Money
	CarryingCost      types.Money
	ABCClassification string
	DepreciationValue types.RoundedFloat
	InsuranceValue    types.Money
	EffectiveDate     time.Time
	CreatedDate       time.Time
	UpdatedDate       time.Time
}

type NewProductCost struct {
	ProductID         uuid.UUID
	PurchaseCost      types.Money
	SellingPrice      types.Money
	Currency          string
	MSRP              types.Money
	MarkupPercentage  types.RoundedFloat
	LandedCost        types.Money
	CarryingCost      types.Money
	ABCClassification string
	DepreciationValue types.RoundedFloat
	InsuranceValue    types.Money
	EffectiveDate     time.Time
}

type UpdateProductCost struct {
	ProductID         *uuid.UUID
	PurchaseCost      *types.Money
	SellingPrice      *types.Money
	Currency          *string
	MSRP              *types.Money
	MarkupPercentage  *types.RoundedFloat
	LandedCost        *types.Money
	CarryingCost      *types.Money
	ABCClassification *string
	DepreciationValue *types.RoundedFloat
	InsuranceValue    *types.Money
	EffectiveDate     *time.Time
}
