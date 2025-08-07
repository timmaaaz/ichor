package productcostapp

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/business/domain/finance/productcostbus"
	"github.com/timmaaaz/ichor/business/domain/finance/productcostbus/types"
	"github.com/timmaaaz/ichor/foundation/timeutil"
)

type QueryParams struct {
	Page    string
	Rows    string
	OrderBy string

	ID                string
	ProductID         string
	PurchaseCost      string
	SellingPrice      string
	Currency          string
	MSRP              string
	MarkupPercentage  string
	LandedCost        string
	CarryingCost      string
	ABCClassification string
	DepreciationValue string
	InsuranceValue    string
	EffectiveDate     string
	CreatedDate       string
	UpdatedDate       string
}

type ProductCost struct {
	ID                string `json:"cost_id"` // TODO: Take a look at what the proper id of this should be
	ProductID         string `json:"product_id"`
	PurchaseCost      string `json:"purchase_cost"`
	SellingPrice      string `json:"selling_price"`
	Currency          string `json:"currency"`
	MSRP              string `json:"msrp"`
	MarkupPercentage  string `json:"markup_percentage"`
	LandedCost        string `json:"landed_cost"`
	CarryingCost      string `json:"carrying_cost"`
	ABCClassification string `json:"abc_classification"`
	DepreciationValue string `json:"depreciation_value"`
	InsuranceValue    string `json:"insurance_value"`
	EffectiveDate     string `json:"effective_date"`
	CreatedDate       string `json:"created_date"`
	UpdatedDate       string `json:"updated_date"`
}

func (app ProductCost) Encode() ([]byte, string, error) {
	data, err := json.Marshal(app)
	return data, "application/json", err
}

func ToAppProductCost(bus productcostbus.ProductCost) ProductCost {
	return ProductCost{
		ID:                bus.ID.String(),
		ProductID:         bus.ProductID.String(),
		PurchaseCost:      bus.PurchaseCost.Value(),
		SellingPrice:      bus.SellingPrice.Value(),
		Currency:          bus.Currency,
		MSRP:              bus.MSRP.Value(),
		MarkupPercentage:  bus.MarkupPercentage.String(),
		LandedCost:        bus.LandedCost.Value(),
		CarryingCost:      bus.CarryingCost.Value(),
		ABCClassification: bus.ABCClassification,
		DepreciationValue: bus.DepreciationValue.String(),
		InsuranceValue:    bus.InsuranceValue.Value(),
		EffectiveDate:     bus.EffectiveDate.Format(timeutil.FORMAT),
		CreatedDate:       bus.CreatedDate.Format(timeutil.FORMAT),
		UpdatedDate:       bus.UpdatedDate.Format(timeutil.FORMAT),
	}
}

func ToAppProductCosts(bus []productcostbus.ProductCost) []ProductCost {
	app := make([]ProductCost, len(bus))
	for i, v := range bus {
		app[i] = ToAppProductCost(v)
	}
	return app
}

// =========================================================================

type NewProductCost struct {
	ProductID         string `json:"product_id" validate:"required,min=36,max=36"`
	PurchaseCost      string `json:"purchase_cost" validate:"required"`
	SellingPrice      string `json:"selling_price" validate:"required"`
	Currency          string `json:"currency" validate:"required"`
	MSRP              string `json:"msrp" validate:"required"`
	MarkupPercentage  string `json:"markup_percentage" validate:"required"`
	LandedCost        string `json:"landed_cost" validate:"required"`
	CarryingCost      string `json:"carrying_cost" validate:"required"`
	ABCClassification string `json:"abc_classification" validate:"required"`
	DepreciationValue string `json:"depreciation_value" validate:"required"`
	InsuranceValue    string `json:"insurance_value" validate:"required"`
	EffectiveDate     string `json:"effective_date" validate:"required"`
}

func (app *NewProductCost) Decode(data []byte) error {
	return json.Unmarshal(data, &app)
}

func (app NewProductCost) Validate() error {
	if err := errs.Check(app); err != nil {
		return errs.Newf(errs.InvalidArgument, "validate: %s", err)
	}

	return nil
}

func toBusNewProductCost(app NewProductCost) (productcostbus.NewProductCost, error) {

	productID, err := uuid.Parse(app.ProductID)
	if err != nil {
		return productcostbus.NewProductCost{}, errs.NewFieldsError("productID", err)
	}

	purchaseCost, err := types.ParseMoney(app.PurchaseCost)
	if err != nil {
		return productcostbus.NewProductCost{}, errs.NewFieldsError("purchaseCost", err)
	}

	sellingPrice, err := types.ParseMoney(app.SellingPrice)
	if err != nil {
		return productcostbus.NewProductCost{}, errs.NewFieldsError("sellingPrice", err)
	}

	msrp, err := types.ParseMoney(app.MSRP)
	if err != nil {
		return productcostbus.NewProductCost{}, errs.NewFieldsError("msrp", err)
	}

	markupPercentage, err := types.ParseRoundedFloat(app.MarkupPercentage)
	if err != nil {
		return productcostbus.NewProductCost{}, errs.NewFieldsError("markupPercentage", err)
	}

	landedCost, err := types.ParseMoney(app.LandedCost)
	if err != nil {
		return productcostbus.NewProductCost{}, errs.NewFieldsError("landedCost", err)
	}

	carryingCost, err := types.ParseMoney(app.CarryingCost)
	if err != nil {
		return productcostbus.NewProductCost{}, errs.NewFieldsError("carryingCost", err)
	}

	depreciationValue, err := types.ParseRoundedFloat(app.DepreciationValue)
	if err != nil {
		return productcostbus.NewProductCost{}, errs.NewFieldsError("depreciationValue", err)
	}

	insuranceValue, err := types.ParseMoney(app.InsuranceValue)
	if err != nil {
		return productcostbus.NewProductCost{}, errs.NewFieldsError("insuranceValue", err)
	}

	effectiveDate, err := time.Parse(timeutil.FORMAT, app.EffectiveDate)
	if err != nil {
		return productcostbus.NewProductCost{}, errs.NewFieldsError("effectiveDate", err)
	}

	return productcostbus.NewProductCost{
		ProductID:         productID,
		PurchaseCost:      purchaseCost,
		SellingPrice:      sellingPrice,
		Currency:          app.Currency,
		MSRP:              msrp,
		MarkupPercentage:  markupPercentage,
		LandedCost:        landedCost,
		CarryingCost:      carryingCost,
		ABCClassification: app.ABCClassification,
		DepreciationValue: depreciationValue,
		InsuranceValue:    insuranceValue,
		EffectiveDate:     effectiveDate,
	}, nil
}

// =========================================================================

type UpdateProductCost struct {
	ProductID         *string `json:"product_id" validate:"omitempty,min=36,max=36"`
	PurchaseCost      *string `json:"purchase_cost"`
	SellingPrice      *string `json:"selling_price"`
	Currency          *string `json:"currency"`
	MSRP              *string `json:"msrp"`
	MarkupPercentage  *string `json:"markup_percentage"`
	LandedCost        *string `json:"landed_cost"`
	CarryingCost      *string `json:"carrying_cost"`
	ABCClassification *string `json:"abc_classification"`
	DepreciationValue *string `json:"depreciation_value"`
	InsuranceValue    *string `json:"insurance_value"`
	EffectiveDate     *string `json:"effective_date"`
}

func (app *UpdateProductCost) Decode(data []byte) error {
	return json.Unmarshal(data, &app)
}

func (app UpdateProductCost) Validate() error {
	if err := errs.Check(app); err != nil {
		return errs.Newf(errs.InvalidArgument, "validate: %s", err)
	}

	return nil
}

func toBusUpdateProductCost(app UpdateProductCost) (productcostbus.UpdateProductCost, error) {

	upc := productcostbus.UpdateProductCost{}

	if app.ProductID != nil {
		productID, err := uuid.Parse(*app.ProductID)
		if err != nil {
			return productcostbus.UpdateProductCost{}, errs.NewFieldsError("productID", err)
		}
		upc.ProductID = &productID
	}

	if app.PurchaseCost != nil {
		purchaseCost, err := types.ParseMoney(*app.PurchaseCost)
		if err != nil {
			return productcostbus.UpdateProductCost{}, errs.NewFieldsError("purchaseCost", err)
		}
		upc.PurchaseCost = &purchaseCost
	}

	if app.SellingPrice != nil {
		sellingPrice, err := types.ParseMoney(*app.SellingPrice)
		if err != nil {
			return productcostbus.UpdateProductCost{}, errs.NewFieldsError("sellingPrice", err)
		}
		upc.SellingPrice = &sellingPrice
	}

	if app.MSRP != nil {
		msrp, err := types.ParseMoney(*app.MSRP)
		if err != nil {
			return productcostbus.UpdateProductCost{}, errs.NewFieldsError("msrp", err)
		}
		upc.MSRP = &msrp
	}

	if app.MarkupPercentage != nil {
		markupPercentage, err := types.ParseRoundedFloat(*app.MarkupPercentage)
		if err != nil {
			return productcostbus.UpdateProductCost{}, errs.NewFieldsError("markupPercentage", err)
		}
		upc.MarkupPercentage = &markupPercentage
	}

	if app.LandedCost != nil {
		landedCost, err := types.ParseMoney(*app.LandedCost)
		if err != nil {
			return productcostbus.UpdateProductCost{}, errs.NewFieldsError("landedCost", err)
		}

		upc.LandedCost = &landedCost
	}

	if app.CarryingCost != nil {
		carryingCost, err := types.ParseMoney(*app.CarryingCost)
		if err != nil {
			return productcostbus.UpdateProductCost{}, errs.NewFieldsError("carryingCost", err)
		}
		upc.CarryingCost = &carryingCost
	}

	if app.DepreciationValue != nil {
		depreciationValue, err := types.ParseRoundedFloat(*app.DepreciationValue)
		if err != nil {
			return productcostbus.UpdateProductCost{}, errs.NewFieldsError("depreciationValue", err)
		}
		upc.DepreciationValue = &depreciationValue
	}

	if app.InsuranceValue != nil {
		insuranceValue, err := types.ParseMoney(*app.InsuranceValue)
		if err != nil {
			return productcostbus.UpdateProductCost{}, errs.NewFieldsError("insuranceValue", err)
		}

		upc.InsuranceValue = &insuranceValue
	}

	if app.EffectiveDate != nil {
		effectiveDate, err := time.Parse(timeutil.FORMAT, *app.EffectiveDate)
		if err != nil {
			return productcostbus.UpdateProductCost{}, errs.NewFieldsError("effectiveDate", err)
		}

		upc.EffectiveDate = &effectiveDate
	}

	if app.Currency != nil {
		upc.Currency = app.Currency
	}

	if app.ABCClassification != nil {
		upc.ABCClassification = app.ABCClassification
	}

	return upc, nil
}
