package metricsbus

import (
	"context"
	"fmt"
	"math/rand"
	"sort"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/quality/metricsbus/types"
)

func TestNewMetrics(n int, productIDs uuid.UUIDs) []NewMetric {

	newMetrics := make([]NewMetric, n)
	idx := rand.Intn(10000)

	for i := 0; i < n; i++ {
		nm := NewMetric{
			ProductID:         productIDs[idx%len(productIDs)],
			ReturnRate:        types.NewRoundedFloat(float64(idx / (i + 3))),
			DefectRate:        types.NewRoundedFloat(float64(idx / n)),
			MeasurementPeriod: types.MustParseInterval("3 days"),
		}

		newMetrics[i] = nm

	}

	return newMetrics
}

func TestSeedMetrics(ctx context.Context, n int, productIDs uuid.UUIDs, api *Business) ([]Metric, error) {
	newMetrics := TestNewMetrics(n, productIDs)

	metrics := make([]Metric, len(newMetrics))

	for i, nm := range newMetrics {
		m, err := api.Create(ctx, nm)
		if err != nil {
			return nil, fmt.Errorf("seeding error: %v", err)
		}
		metrics[i] = m
	}

	sort.Slice(metrics, func(i, j int) bool {
		return metrics[i].MetricID.String() < metrics[j].MetricID.String()
	})

	return metrics, nil
}
