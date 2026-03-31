package inventoryadjustmentbus

import (
	"context"
	"fmt"
	"math/rand"
	"sort"
	"time"

	"github.com/google/uuid"
)

func TestNewInventoryAdjustment(n int, productIDs, locationIDs, adjustedByIDs uuid.UUIDs) []NewInventoryAdjustment {
	newInventoryAdjustments := make([]NewInventoryAdjustment, n)

	floorWorker1 := uuid.MustParse("c0000000-0000-4000-8000-000000000001")

	reasonCodes := []string{
		ReasonCodeDamaged,
		ReasonCodeTheft,
		ReasonCodeDataEntryError,
		ReasonCodeFoundStock,
		ReasonCodePickingError,
		ReasonCodeReceivingError,
		ReasonCodeCycleCount,
		ReasonCodeOther,
	}

	idx := rand.Intn(10000)

	for i := 0; i < n; i++ {
		approvedByID := adjustedByIDs[(idx+1)%len(adjustedByIDs)]

		adjustedBy := adjustedByIDs[idx%len(adjustedByIDs)]
		if i < 2 {
			adjustedBy = floorWorker1
		}

		newInventoryAdjustments[i] = NewInventoryAdjustment{
			ProductID:      productIDs[idx%len(productIDs)],
			LocationID:     locationIDs[idx%len(locationIDs)],
			AdjustedBy:     adjustedBy,
			ApprovedBy:     &approvedByID,
			QuantityChange: rand.Intn(100) - 50,
			ReasonCode:     reasonCodes[i%len(reasonCodes)],
			Notes:          fmt.Sprintf("Adjustment for %s", reasonCodes[i%len(reasonCodes)]),
			AdjustmentDate: time.Now().AddDate(0, 0, -(i % 30)),
		}
	}

	return newInventoryAdjustments
}

func TestSeedInventoryAdjustments(ctx context.Context, n int, productIDs, locationIDs, adjustedByIDs uuid.UUIDs, api *Business) ([]InventoryAdjustment, error) {
	newInventoryAdjustments := TestNewInventoryAdjustment(n, productIDs, locationIDs, adjustedByIDs)

	inventoryAdjustments := make([]InventoryAdjustment, len(newInventoryAdjustments))
	for i, nia := range newInventoryAdjustments {
		ia, err := api.Create(ctx, nia)
		if err != nil {
			return []InventoryAdjustment{}, err
		}
		inventoryAdjustments[i] = ia
	}

	sort.Slice(inventoryAdjustments, func(i, j int) bool {
		return inventoryAdjustments[i].InventoryAdjustmentID.String() < inventoryAdjustments[j].InventoryAdjustmentID.String()
	})

	return inventoryAdjustments, nil
}
