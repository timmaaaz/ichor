package lottrackingsbus

import (
	"context"
	"fmt"
	"math/rand"
	"sort"
	"time"

	"github.com/google/uuid"
)

func RandomDate() time.Time {
	start := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2030, 12, 31, 23, 59, 59, 0, time.UTC)
	diff := end.Sub(start)
	randomDays := rand.Intn(int(diff.Hours() / 24))
	return start.Add(time.Duration(randomDays) * 24 * time.Hour)
}

func TestNewLotTrackings(n int, supplierProductIDs uuid.UUIDs) []NewLotTrackings {
	newLotTrackingss := make([]NewLotTrackings, n)

	idx := rand.Intn(10000)
	for i := 0; i < n; i++ {
		idx++
		newLotTrackingss[i] = NewLotTrackings{
			SupplierProductID: supplierProductIDs[i%len(supplierProductIDs)],
			LotNumber:         fmt.Sprintf("LotNumber%d", idx),
			ManufactureDate:   RandomDate(),
			ExpirationDate:    RandomDate(),
			RecievedDate:      RandomDate(),
			QualityStatus:     fmt.Sprintf("QualityStatus%d", idx),
			Quantity:          rand.Intn(1000),
		}
	}

	return newLotTrackingss

}

func TestSeedLotTrackings(ctx context.Context, n int, supplierProductIDs uuid.UUIDs, api *Business) ([]LotTrackings, error) {
	newLotTrackings := TestNewLotTrackings(n, supplierProductIDs)

	lotTrackingss := make([]LotTrackings, len(newLotTrackings))

	for i, nl := range newLotTrackings {
		lt, err := api.Create(ctx, nl)
		if err != nil {
			return nil, fmt.Errorf("seeding error: %v", err)
		}
		lotTrackingss[i] = lt
	}

	sort.Slice(lotTrackingss, func(i, j int) bool {
		return lotTrackingss[i].LotID.String() < lotTrackingss[j].LotID.String()
	})

	return lotTrackingss, nil
}
