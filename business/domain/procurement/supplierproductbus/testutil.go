package supplierproductbus

import (
	"context"
	"fmt"
	"math/rand"
	"sort"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/procurement/supplierproductbus/types"
)

func TestNewSupplierProducts(n int, productIDs, supplierIDs uuid.UUIDs) []NewSupplierProduct {
	newSupplierProducts := make([]NewSupplierProduct, n)

	idx := rand.Intn(10000)
	for i := 0; i < n; i++ {
		idx++

		newSupplierProducts[i] = NewSupplierProduct{
			SupplierID:         supplierIDs[idx%len(supplierIDs)],
			ProductID:          productIDs[idx%len(productIDs)],
			SupplierPartNumber: fmt.Sprintf("SupplierPartNumber%d", idx),
			MinOrderQuantity:   idx - 10,
			MaxOrderQuantity:   idx + 10,
			LeadTimeDays:       idx,
			UnitCost:           types.MustParseMoney(fmt.Sprintf("%.2f", rand.Float64()*10)),
			IsPrimarySupplier:  idx%2 == 0,
		}
	}

	return newSupplierProducts
}

func TestSeedSupplierProducts(ctx context.Context, n int, productIDs, supplierIDs uuid.UUIDs, api *Business) ([]SupplierProduct, error) {
	newSupplierProducts := TestNewSupplierProducts(n, productIDs, supplierIDs)

	supplierProducts := make([]SupplierProduct, len(newSupplierProducts))

	for i, nsp := range newSupplierProducts {
		sp, err := api.Create(ctx, nsp)
		if err != nil {
			return []SupplierProduct{}, err
		}
		supplierProducts[i] = sp
	}

	sort.Slice(supplierProducts, func(i, j int) bool {
		return supplierProducts[i].ProductID.String() < supplierProducts[j].ProductID.String()
	})

	return supplierProducts, nil
}
