package reportstobus

import (
	"context"
	"fmt"
	"math/rand"
	"sort"

	"github.com/google/uuid"
)

func TestNewReportsTo(n int, reporterID, bossID []uuid.UUID) []NewReportsTo {
	newReportsTo := make([]NewReportsTo, n)

	for i := 0; i < n; i++ {
		nrt := NewReportsTo{
			ReporterID: reporterID[rand.Intn(len(reporterID))],
			BossID:     bossID[rand.Intn(len(bossID))],
		}
		newReportsTo[i] = nrt
	}

	return newReportsTo
}

func TestSeedReportsTo(ctx context.Context, n int, reporterID, bossID []uuid.UUID, api *Business) ([]ReportsTo, error) {

	newReportsTos := TestNewReportsTo(20, reporterID, bossID)

	reportsTos := make([]ReportsTo, len(newReportsTos))

	for i, nrt := range newReportsTos {
		reportsTo, err := api.Create(ctx, nrt)
		if err != nil {
			return nil, fmt.Errorf("seeding reportstos: idx: %d : %w", i, err)
		}
		reportsTos[i] = reportsTo
	}

	sort.Slice(reportsTos, func(i, j int) bool {
		return reportsTos[i].ID.String() < reportsTos[j].ID.String()
	})

	return reportsTos, nil
}
