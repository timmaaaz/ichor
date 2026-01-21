package productcostbus

import (
	"time"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/products/productcostbus/types"
)

// JSON tags are required for workflow event serialization. The workflow system
// (via EventPublisher) marshals business models to JSON for RawData in TriggerEvents.
// Without these tags, Go defaults to PascalCase keys, but workflow action handlers
// expect snake_case keys to match API conventions.

type ProductCost struct {
	ID                uuid.UUID          `json:"id"`
	ProductID         uuid.UUID          `json:"product_id"`
	PurchaseCost      types.Money        `json:"purchase_cost"`
	SellingPrice      types.Money        `json:"selling_price"`
	CurrencyID        uuid.UUID          `json:"currency_id"`
	MSRP              types.Money        `json:"msrp"`
	MarkupPercentage  types.RoundedFloat `json:"markup_percentage"`
	LandedCost        types.Money        `json:"landed_cost"`
	CarryingCost      types.Money        `json:"carrying_cost"`
	ABCClassification string             `json:"abc_classification"`
	DepreciationValue types.RoundedFloat `json:"depreciation_value"`
	InsuranceValue    types.Money        `json:"insurance_value"`
	EffectiveDate     time.Time          `json:"effective_date"`
	CreatedDate       time.Time          `json:"created_date"`
	UpdatedDate       time.Time          `json:"updated_date"`
}

type NewProductCost struct {
	ProductID         uuid.UUID          `json:"product_id"`
	PurchaseCost      types.Money        `json:"purchase_cost"`
	SellingPrice      types.Money        `json:"selling_price"`
	CurrencyID        uuid.UUID          `json:"currency_id"`
	MSRP              types.Money        `json:"msrp"`
	MarkupPercentage  types.RoundedFloat `json:"markup_percentage"`
	LandedCost        types.Money        `json:"landed_cost"`
	CarryingCost      types.Money        `json:"carrying_cost"`
	ABCClassification string             `json:"abc_classification"`
	DepreciationValue types.RoundedFloat `json:"depreciation_value"`
	InsuranceValue    types.Money        `json:"insurance_value"`
	EffectiveDate     time.Time          `json:"effective_date"`
}

type UpdateProductCost struct {
	ProductID         *uuid.UUID          `json:"product_id,omitempty"`
	PurchaseCost      *types.Money        `json:"purchase_cost,omitempty"`
	SellingPrice      *types.Money        `json:"selling_price,omitempty"`
	CurrencyID        *uuid.UUID          `json:"currency_id,omitempty"`
	MSRP              *types.Money        `json:"msrp,omitempty"`
	MarkupPercentage  *types.RoundedFloat `json:"markup_percentage,omitempty"`
	LandedCost        *types.Money        `json:"landed_cost,omitempty"`
	CarryingCost      *types.Money        `json:"carrying_cost,omitempty"`
	ABCClassification *string             `json:"abc_classification,omitempty"`
	DepreciationValue *types.RoundedFloat `json:"depreciation_value,omitempty"`
	InsuranceValue    *types.Money        `json:"insurance_value,omitempty"`
	EffectiveDate     *time.Time          `json:"effective_date,omitempty"`
}
