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

// validQualityStatuses are the values allowed by the lot_trackings_quality_status_check constraint.
var validQualityStatuses = []string{"good", "on_hold", "quarantined", "released", "expired"}

func TestNewLotTrackings(n int, supplierProductIDs uuid.UUIDs) []NewLotTrackings {
	newLotTrackingss := make([]NewLotTrackings, n)

	now := time.Now()

	idx := rand.Intn(10000)
	for i := 0; i < n; i++ {
		idx++

		// Expiration dates distributed for dashboard buckets:
		// idx 0: urgent (within 7 days), idx 1: warning (within 30 days),
		// idx 2: monitor (within 90 days), rest: 30-180 days out
		var expirationDate time.Time
		switch i {
		case 0:
			expirationDate = now.AddDate(0, 0, 5) // urgent
		case 1:
			expirationDate = now.AddDate(0, 0, 20) // warning
		case 2:
			expirationDate = now.AddDate(0, 0, 60) // monitor
		default:
			expirationDate = now.AddDate(0, 0, 30+rand.Intn(150))
		}

		// First 3 are "good" (for dashboard visibility), then cycle to ensure quarantined
		qualityStatuses := []string{"good", "good", "good", "quarantined", "on_hold", "released", "expired"}
		qualityStatus := qualityStatuses[i%len(qualityStatuses)]

		newLotTrackingss[i] = NewLotTrackings{
			SupplierProductID: supplierProductIDs[i%len(supplierProductIDs)],
			LotNumber:         fmt.Sprintf("LOT-2026-%03d", i+1),
			ManufactureDate:   now.AddDate(0, -3, 0),
			ExpirationDate:    expirationDate,
			ReceivedDate:      now.AddDate(0, 0, -rand.Intn(30)),
			QualityStatus:     qualityStatus,
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
		// Re-fetch via QueryByID to pick up JOIN-enriched fields (ProductID, ProductName, ProductSKU).
		enriched, err := api.QueryByID(ctx, lt.LotID)
		if err != nil {
			return nil, fmt.Errorf("seeding re-fetch error: %v", err)
		}
		lotTrackingss[i] = enriched
	}

	sort.Slice(lotTrackingss, func(i, j int) bool {
		return lotTrackingss[i].LotID.String() < lotTrackingss[j].LotID.String()
	})

	return lotTrackingss, nil
}
