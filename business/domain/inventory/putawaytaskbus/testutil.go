package putawaytaskbus

import (
	"context"
	"fmt"
	"sort"

	"github.com/google/uuid"
)

// TestNewPutAwayTasks generates n new put-away tasks for testing.
func TestNewPutAwayTasks(n int, productIDs, locationIDs, createdByIDs []uuid.UUID) []NewPutAwayTask {
	tasks := make([]NewPutAwayTask, n)

	// floor_worker1 UUID — stable across all environments (from seed.sql)
	floorWorker1 := uuid.MustParse("c0000000-0000-4000-8000-000000000001")

	for i := range n {
		createdBy := createdByIDs[i%len(createdByIDs)]
		if i == 0 {
			createdBy = floorWorker1
		}

		tasks[i] = NewPutAwayTask{
			ProductID:       productIDs[i%len(productIDs)],
			LocationID:      locationIDs[i%len(locationIDs)],
			Quantity:        (i + 1) * 10,
			ReferenceNumber: fmt.Sprintf("PO-HIST-%d", i+1),
			CreatedBy:       createdBy,
		}
	}

	return tasks
}

// TestSeedPutAwayTasks creates n put-away tasks in the database for testing.
func TestSeedPutAwayTasks(ctx context.Context, n int, productIDs, locationIDs, createdByIDs []uuid.UUID, api *Business) ([]PutAwayTask, error) {
	newTasks := TestNewPutAwayTasks(n, productIDs, locationIDs, createdByIDs)

	tasks := make([]PutAwayTask, len(newTasks))
	for i, npt := range newTasks {
		task, err := api.Create(ctx, npt)
		if err != nil {
			return nil, err
		}
		tasks[i] = task
	}

	sort.Slice(tasks, func(i, j int) bool {
		return tasks[i].ID.String() < tasks[j].ID.String()
	})

	return tasks, nil
}
