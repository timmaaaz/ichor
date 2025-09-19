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

	idx := rand.Intn(10000)

	for i := 0; i < n; i++ {
		idx++
		newInspections[i] = NewInspection{
			ProductID:          productIDs[idx%len(productIDs)],
			InspectorID:        inspectorIDs[idx%len(inspectorIDs)],
			LotID:              lotIDs[idx%len(lotIDs)],
			InspectionDate:     time.Now(),
			Status:             "pending",
			NextInspectionDate: time.Now().AddDate(0, 0, 14),
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
