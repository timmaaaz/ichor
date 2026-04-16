package labelbus

import (
	"context"
	"fmt"
	"sort"
)

// TestNewLabels generates n NewLabelCatalog values with deterministic codes.
func TestNewLabels(n int) []NewLabelCatalog {
	labels := make([]NewLabelCatalog, n)
	for i := range n {
		labels[i] = NewLabelCatalog{
			Code:      fmt.Sprintf("TEST-%04d", i+1),
			Type:      TypeLocation,
			EntityRef: "",
		}
	}
	return labels
}

// TestSeedLabels creates n labels in the database for testing and returns
// them sorted by ID for stable comparison.
func TestSeedLabels(ctx context.Context, n int, api *Business) ([]LabelCatalog, error) {
	newLabels := TestNewLabels(n)

	labels := make([]LabelCatalog, len(newLabels))
	for i, nl := range newLabels {
		lc, err := api.Create(ctx, nl)
		if err != nil {
			return nil, fmt.Errorf("seeding label %d: %w", i, err)
		}
		labels[i] = lc
	}

	sort.Slice(labels, func(i, j int) bool {
		return labels[i].ID.String() < labels[j].ID.String()
	})

	return labels, nil
}
