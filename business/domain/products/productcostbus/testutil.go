package productcostbus

import (
	"context"
	"fmt"
	"math/rand"
	"sort"
	"time"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/products/productcostbus/types"
)

const charString = "abcdefghijklmnopqrstuvwxyz"

// Common ISO 4217 currency codes for realistic test data
var currencyCodes = []string{"USD", "EUR", "GBP", "CAD", "AUD", "JPY", "CHF", "CNY", "INR", "MXN"}

func TestNewProductCosts(n int, productIDs uuid.UUIDs) []NewProductCost {
	newProductCosts := make([]NewProductCost, n)

	idx := rand.Intn(10000)
	for i := 0; i < n; i++ {

		idx++

		newProductCosts[i] = NewProductCost{
			ProductID:         productIDs[i%len(productIDs)],
			PurchaseCost:      types.MustParseMoney(fmt.Sprintf("%.2f", rand.Float64())),
			SellingPrice:      types.MustParseMoney(fmt.Sprintf("%.2f", rand.Float64()+float64(i))),
			Currency:          currencyCodes[0], // Default to USD for seed data; use i%len(currencyCodes) for variety
			MSRP:              types.MustParseMoney(fmt.Sprintf("%.2f", rand.Float64())),
			MarkupPercentage:  types.NewRoundedFloat(rand.Float64()),
			LandedCost:        types.MustParseMoney(fmt.Sprintf("%.2f", rand.Float64())),
			CarryingCost:      types.MustParseMoney(fmt.Sprintf("%.2f", rand.Float64())),
			ABCClassification: string(charString[idx%len(charString)]), // get a char from the char string
			DepreciationValue: types.NewRoundedFloat(rand.Float64()),
			InsuranceValue:    types.MustParseMoney(fmt.Sprintf("%.2f", rand.Float64())),
			EffectiveDate:     time.Now(),
		}
	}

	return newProductCosts
}

func TestSeedProductCosts(ctx context.Context, n int, productIDs uuid.UUIDs, api *Business) ([]ProductCost, error) {
	newProductCosts := TestNewProductCosts(n, productIDs)

	productCosts := make([]ProductCost, len(newProductCosts))

	for i, npc := range newProductCosts {
		pc, err := api.Create(ctx, npc)
		if err != nil {
			return []ProductCost{}, err
		}
		productCosts[i] = pc
	}

	sort.Slice(productCosts, func(i, j int) bool {
		return productCosts[i].ProductID.String() < productCosts[j].ProductID.String()
	})

	return productCosts, nil
}
