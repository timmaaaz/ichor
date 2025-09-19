package productcostdb

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/products/productcostbus"
	"github.com/timmaaaz/ichor/business/domain/products/productcostbus/types"
)

type productCost struct {
	ID                uuid.UUID      `db:"id"`
	ProductID         uuid.UUID      `db:"product_id"`
	PurchaseCost      sql.NullString `db:"purchase_cost"`
	SellingPrice      sql.NullString `db:"selling_price"`
	Currency          string         `db:"currency"`
	MSRP              sql.NullString `db:"msrp"`
	MarkupPercentage  sql.NullString `db:"markup_percentage"`
	LandedCost        sql.NullString `db:"landed_cost"`
	CarryingCost      sql.NullString `db:"carrying_cost"`
	ABCClassification string         `db:"abc_classification"`
	DepreciationValue sql.NullString `db:"depreciation_value"`
	InsuranceValue    sql.NullString `db:"insurance_value"`
	EffectiveDate     time.Time      `db:"effective_date"`
	CreatedDate       time.Time      `db:"created_date"`
	UpdatedDate       time.Time      `db:"updated_date"`
}

func toDBProductCost(bus productcostbus.ProductCost) productCost {
	return productCost{
		ID:                bus.ID,
		ProductID:         bus.ProductID,
		PurchaseCost:      bus.PurchaseCost.DBValue(),
		SellingPrice:      bus.SellingPrice.DBValue(),
		Currency:          bus.Currency,
		MSRP:              bus.MSRP.DBValue(),
		MarkupPercentage:  bus.MarkupPercentage.DBValue(),
		LandedCost:        bus.LandedCost.DBValue(),
		CarryingCost:      bus.CarryingCost.DBValue(),
		ABCClassification: bus.ABCClassification,
		DepreciationValue: bus.DepreciationValue.DBValue(),
		InsuranceValue:    bus.InsuranceValue.DBValue(),
		EffectiveDate:     bus.EffectiveDate.UTC(),
		CreatedDate:       bus.CreatedDate.UTC(),
		UpdatedDate:       bus.UpdatedDate.UTC(),
	}
}

func toBusProductCost(db productCost) (productcostbus.ProductCost, error) {

	purchaseCost, err := types.ParseMoney(db.PurchaseCost.String)
	if err != nil {
		return productcostbus.ProductCost{}, fmt.Errorf("tobusproudctcost: %v", err)
	}

	sellingPrice, err := types.ParseMoney(db.SellingPrice.String)
	if err != nil {
		return productcostbus.ProductCost{}, fmt.Errorf("tobusproudctcost: %v", err)
	}

	MSRP, err := types.ParseMoney(db.MSRP.String)
	if err != nil {
		return productcostbus.ProductCost{}, fmt.Errorf("tobusproudctcost: %v", err)
	}

	landedCost, err := types.ParseMoney(db.LandedCost.String)
	if err != nil {
		return productcostbus.ProductCost{}, fmt.Errorf("tobusproudctcost: %v", err)
	}

	carryingCost, err := types.ParseMoney(db.CarryingCost.String)
	if err != nil {
		return productcostbus.ProductCost{}, fmt.Errorf("tobusproudctcost: %v", err)
	}

	insuranceValue, err := types.ParseMoney(db.InsuranceValue.String)
	if err != nil {
		return productcostbus.ProductCost{}, fmt.Errorf("tobusproudctcost: %v", err)
	}

	markupPercentage, err := types.ParseRoundedFloat(db.MarkupPercentage.String)
	if err != nil {
		return productcostbus.ProductCost{}, fmt.Errorf("tobusproudctcost: %v", err)
	}

	depreciationValue, err := types.ParseRoundedFloat(db.DepreciationValue.String)
	if err != nil {
		return productcostbus.ProductCost{}, fmt.Errorf("tobusproudctcost: %v", err)
	}

	return productcostbus.ProductCost{
		ID:                db.ID,
		ProductID:         db.ProductID,
		PurchaseCost:      purchaseCost,
		SellingPrice:      sellingPrice,
		Currency:          db.Currency,
		MSRP:              MSRP,
		MarkupPercentage:  markupPercentage,
		LandedCost:        landedCost,
		CarryingCost:      carryingCost,
		ABCClassification: db.ABCClassification,
		DepreciationValue: depreciationValue,
		InsuranceValue:    insuranceValue,
		EffectiveDate:     db.EffectiveDate.Local(),
		CreatedDate:       db.CreatedDate.Local(),
		UpdatedDate:       db.UpdatedDate.Local(),
	}, nil
}

func toBusProductCosts(dbProductCosts []productCost) ([]productcostbus.ProductCost, error) {
	busProductCosts := make([]productcostbus.ProductCost, len(dbProductCosts))

	for i, dbProductCost := range dbProductCosts {
		busProductCost, err := toBusProductCost(dbProductCost)
		if err != nil {
			return nil, fmt.Errorf("tobusproductcosts: %v", err)
		}
		busProductCosts[i] = busProductCost
	}

	return busProductCosts, nil
}
