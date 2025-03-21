package supplierbus

import (
	"context"
	"fmt"
	"math/rand"
	"sort"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/supplier/supplierbus/types"
)

func TestNewSupplier(n int, contactIDs uuid.UUIDs) []NewSupplier {
	newSuppliers := make([]NewSupplier, n)

	idx := rand.Intn(10000)
	for i := 0; i < n; i++ {
		idx++

		newSuppliers[i] = NewSupplier{
			ContactID:    contactIDs[idx%len(contactIDs)],
			Name:         fmt.Sprintf("Name%d", idx),
			PaymentTerms: fmt.Sprintf("PaymentTerms%d", idx),
			LeadTimeDays: idx,
			Rating:       types.MustParseRoundedFloat(fmt.Sprintf("%.2f", rand.Float64()*10)),
			IsActive:     idx%2 == 0,
		}
	}

	return newSuppliers
}

func TestSeedSuppliers(ctx context.Context, n int, contactIDs uuid.UUIDs, api *Business) ([]Supplier, error) {
	newSuppliers := TestNewSupplier(n, contactIDs)

	suppliers := make([]Supplier, len(newSuppliers))

	for i, ns := range newSuppliers {
		supplier, err := api.Create(ctx, ns)
		if err != nil {
			return []Supplier{}, err
		}
		suppliers[i] = supplier
	}

	sort.Slice(suppliers, func(i, j int) bool {
		return suppliers[i].Name < suppliers[j].Name
	})

	return suppliers, nil
}
