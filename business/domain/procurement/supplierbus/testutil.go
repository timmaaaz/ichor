package supplierbus

import (
	"context"
	"fmt"
	"math/rand"
	"sort"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/procurement/supplierbus/types"
)

func TestNewSupplier(n int, ContactInfosIDs uuid.UUIDs) []NewSupplier {
	newSuppliers := make([]NewSupplier, n)

	idx := rand.Intn(10000)
	for i := 0; i < n; i++ {
		idx++

		newSuppliers[i] = NewSupplier{
			ContactInfosID: ContactInfosIDs[idx%len(ContactInfosIDs)],
			Name:           fmt.Sprintf("Name%d", idx),
			LeadTimeDays:   idx,
			Rating:         types.MustParseRoundedFloat(fmt.Sprintf("%.2f", rand.Float64()*10)),
			IsActive:       idx%2 == 0,
		}
	}

	return newSuppliers
}

func TestSeedSuppliers(ctx context.Context, n int, ContactInfosIDs uuid.UUIDs, api *Business) ([]Supplier, error) {
	newSuppliers := TestNewSupplier(n, ContactInfosIDs)

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
