package costhistorybus

import (
	"context"
	"fmt"
	"math/rand"
	"sort"
	"time"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/products/costhistorybus/types"
)

func TestNewCostHistory(n int, productIDs uuid.UUIDs) []NewCostHistory {
	newCostHistories := make([]NewCostHistory, n)

	idx := rand.Intn(10000)
	for i := 0; i < n; i++ {
		idx++

		newCostHistories[i] = NewCostHistory{
			ProductID:     productIDs[i%len(productIDs)],
			CostType:      fmt.Sprintf("CostType%d", idx),
			Amount:        types.MustParseMoney(fmt.Sprintf("%.2f", rand.Float64()*10)),
			Currency:      fmt.Sprintf("Currency%d", idx),
			EffectiveDate: time.Now(),
			EndDate:       time.Now().AddDate(0, 3, 0),
		}
	}

	return newCostHistories
}

func TestSeedCostHistories(ctx context.Context, n int, productIDs uuid.UUIDs, api *Business) ([]CostHistory, error) {
	newCostHistories := TestNewCostHistory(n, productIDs)

	costHistories := make([]CostHistory, len(newCostHistories))

	for i, nch := range newCostHistories {
		ch, err := api.Create(ctx, nch)
		if err != nil {
			return []CostHistory{}, err
		}
		costHistories[i] = ch
	}

	sort.Slice(costHistories, func(i, j int) bool {
		return costHistories[i].Amount.Value() < costHistories[j].Amount.Value()
	})

	return costHistories, nil
}

// TestNewCostHistoryHistorical creates cost histories distributed across a time range for seeding.
// daysBack specifies how many days of history to generate (180-365 days recommended).
// Cost histories are evenly distributed across the time range.
func TestNewCostHistoryHistorical(n int, daysBack int, productIDs uuid.UUIDs) []NewCostHistory {
	newCostHistories := make([]NewCostHistory, n)
	now := time.Now()

	idx := rand.Intn(10000)
	for i := 0; i < n; i++ {
		idx++

		// Distribute evenly across the time range
		daysAgo := (i * daysBack) / n
		createdDate := now.AddDate(0, 0, -daysAgo)
		effectiveDate := createdDate
		endDate := effectiveDate.AddDate(0, 3, 0) // 3 months validity

		newCostHistories[i] = NewCostHistory{
			ProductID:     productIDs[i%len(productIDs)],
			CostType:      fmt.Sprintf("CostType%d", idx),
			Amount:        types.MustParseMoney(fmt.Sprintf("%.2f", rand.Float64()*10)),
			Currency:      fmt.Sprintf("Currency%d", idx),
			EffectiveDate: effectiveDate,
			EndDate:       endDate,
			CreatedDate:   &createdDate,
		}
	}

	return newCostHistories
}

// TestSeedCostHistoriesHistorical seeds cost histories with historical date distribution.
func TestSeedCostHistoriesHistorical(ctx context.Context, n int, daysBack int, productIDs uuid.UUIDs, api *Business) ([]CostHistory, error) {
	newCostHistories := TestNewCostHistoryHistorical(n, daysBack, productIDs)

	costHistories := make([]CostHistory, len(newCostHistories))

	for i, nch := range newCostHistories {
		ch, err := api.Create(ctx, nch)
		if err != nil {
			return []CostHistory{}, err
		}
		costHistories[i] = ch
	}

	sort.Slice(costHistories, func(i, j int) bool {
		return costHistories[i].Amount.Value() < costHistories[j].Amount.Value()
	})

	return costHistories, nil
}
