package productuombus

import (
	"context"
	"fmt"
	"math/rand"
	"sort"

	"github.com/google/uuid"
)

func TestNewProductUOMs(n int, productIDs []uuid.UUID) []NewProductUOM {
	newUOMs := make([]NewProductUOM, n)

	idx := rand.Intn(10000)
	for i := 0; i < n; i++ {
		idx++
		newUOMs[i] = NewProductUOM{
			ProductID:        productIDs[rand.Intn(len(productIDs))],
			Name:             fmt.Sprintf("UOM%d", idx),
			Abbreviation:     fmt.Sprintf("U%d", idx),
			ConversionFactor: float64(idx) * 0.1,
			IsBase:           i == 0,
			IsApproximate:    false,
			Notes:            fmt.Sprintf("notes%d", idx),
		}
	}

	return newUOMs
}

func TestSeedProductUOMs(ctx context.Context, n int, productIDs []uuid.UUID, api *Business) ([]ProductUOM, error) {
	newUOMs := TestNewProductUOMs(n, productIDs)

	uoms := make([]ProductUOM, len(newUOMs))

	for i, nu := range newUOMs {
		uom, err := api.Create(ctx, nu)
		if err != nil {
			return nil, err
		}
		uoms[i] = uom
	}

	sort.Slice(uoms, func(i, j int) bool {
		return uoms[i].Name < uoms[j].Name
	})

	return uoms, nil
}
