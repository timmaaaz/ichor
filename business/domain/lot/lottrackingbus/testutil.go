package lottrackingbus

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

func TestNewLotTracking(n int, supplierProductIDs uuid.UUIDs) []NewLotTracking {
	newLotTrackings := make([]NewLotTracking, n)

	idx := rand.Intn(10000)
	for i := 0; i < n; i++ {
		idx++
		newLotTrackings[i] = NewLotTracking{
			SupplierProductID: supplierProductIDs[i%len(supplierProductIDs)],
			LotNumber:         fmt.Sprintf("LotNumber%d", idx),
			ManufactureDate:   RandomDate(),
			ExpirationDate:    RandomDate(),
			RecievedDate:      RandomDate(),
			QualityStatus:     fmt.Sprintf("QualityStatus%d", idx),
			Quantity:          rand.Intn(1000),
		}
	}

	return newLotTrackings

}

func TestSeedLotTracking(ctx context.Context, n int, supplierProductIDs uuid.UUIDs, api *Business) ([]LotTracking, error) {
	newLotTracking := TestNewLotTracking(n, supplierProductIDs)

	lotTrackings := make([]LotTracking, len(newLotTracking))

	for i, nl := range newLotTracking {
		lt, err := api.Create(ctx, nl)
		if err != nil {
			return nil, fmt.Errorf("seeding error: %v", err)
		}
		lotTrackings[i] = lt
	}

	sort.Slice(lotTrackings, func(i, j int) bool {
		return lotTrackings[i].LotID.String() < lotTrackings[j].LotID.String()
	})

	return lotTrackings, nil
}
