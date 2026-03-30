package inspectionbus

import (
	"context"
	"math/rand"
	"sort"
	"time"

	"github.com/google/uuid"
)

func TestNewInspections(n int, productIDs, inspectorIDs, lotIDs uuid.UUIDs) []NewInspection {
	newInspections := make([]NewInspection, n)

	// floor_worker1 UUID — stable across all environments (from seed.sql)
	floorWorker1 := uuid.MustParse("c0000000-0000-4000-8000-000000000001")

	idx := rand.Intn(10000)

	for i := 0; i < n; i++ {
		idx++

		// First 5 inspections assigned to floor_worker1
		inspectorID := inspectorIDs[idx%len(inspectorIDs)]
		if i < 5 {
			inspectorID = floorWorker1
		}

		newInspections[i] = NewInspection{
			ProductID:          productIDs[idx%len(productIDs)],
			InspectorID:        inspectorID,
			LotID:              lotIDs[idx%len(lotIDs)],
			InspectionDate:     time.Now().AddDate(0, 0, -(i % 7)),
			Status:             "pending",
			NextInspectionDate: time.Now().AddDate(0, 0, i+7),
		}
	}

	return newInspections
}

func TestSeedInspections(ctx context.Context, n int, productIDs, inspectorIDs, lotIDs uuid.UUIDs, api *Business) ([]Inspection, error) {
	newInspections := TestNewInspections(n, productIDs, inspectorIDs, lotIDs)

	inspections := make([]Inspection, len(newInspections))
	for i, ni := range newInspections {
		inspection, err := api.Create(ctx, ni)
		if err != nil {
			return []Inspection{}, err
		}
		inspections[i] = inspection
	}

	sort.Slice(inspections, func(i, j int) bool {
		return inspections[i].InspectionID.String() < inspections[j].InspectionID.String()
	})

	return inspections, nil
}
