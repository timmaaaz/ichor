package physicalattributebus

import (
	"context"
	"fmt"
	"math/rand"
	"sort"

	"github.com/google/uuid"
)

func TestNewPhysicalAttribute(n int, productIDs []uuid.UUID) []NewPhysicalAttribute {
	newPhysicalAttributes := make([]NewPhysicalAttribute, n)

	idx := rand.Intn(10000)
	for i := 0; i < n; i++ {
		idx++
		newPhysicalAttributes[i] = NewPhysicalAttribute{
			ProductID:           productIDs[i%len(productIDs)],
			Length:              rand.Float32(),
			Width:               rand.Float32(),
			Height:              rand.Float32(),
			Weight:              rand.Float32(),
			WeightUnit:          fmt.Sprintf("wu%d", idx),
			Color:               fmt.Sprintf("color%d", idx),
			Size:                fmt.Sprintf("size%d", idx),
			Material:            fmt.Sprintf("material%d", idx),
			StorageRequirements: fmt.Sprintf("storageRequirements%d", idx),
			HazmatClass:         fmt.Sprintf("hazmatClass%d", idx),
			ShelfLifeDays:       idx,
		}
	}

	return newPhysicalAttributes
}

func TestSeedPhysicalAttributes(ctx context.Context, n int, productIDs []uuid.UUID, api *Business) ([]PhysicalAttribute, error) {
	NewPhysicalAttributes := TestNewPhysicalAttribute(n, productIDs)

	physicalAttributes := make([]PhysicalAttribute, len(NewPhysicalAttributes))

	for i, np := range NewPhysicalAttributes {
		physicalAttribute, err := api.Create(ctx, np)
		if err != nil {
			return nil, err
		}
		physicalAttributes[i] = physicalAttribute
	}

	sort.Slice(physicalAttributes, func(i, j int) bool {
		return physicalAttributes[i].ProductID.String() < physicalAttributes[j].ProductID.String()
	})

	return physicalAttributes, nil
}
