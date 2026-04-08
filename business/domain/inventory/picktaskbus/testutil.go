package picktaskbus

import (
	"context"
	"fmt"
	"sort"

	"github.com/google/uuid"
)

// TestNewPickTasks generates n new pick tasks for testing.
func TestNewPickTasks(n int, salesOrderIDs, salesOrderLineItemIDs, productIDs, locationIDs, createdByIDs []uuid.UUID) []NewPickTask {
	tasks := make([]NewPickTask, n)

	for i := range n {
		tasks[i] = NewPickTask{
			SalesOrderID:         salesOrderIDs[i%len(salesOrderIDs)],
			SalesOrderLineItemID: salesOrderLineItemIDs[i%len(salesOrderLineItemIDs)],
			ProductID:            productIDs[i%len(productIDs)],
			LocationID:           locationIDs[i%len(locationIDs)],
			QuantityToPick:       (i + 1) * 5,
			CreatedBy:            createdByIDs[i%len(createdByIDs)],
		}
	}

	return tasks
}

// TestSeedPickTasks creates n pick tasks in the database for testing.
// If assigneeIDs is non-nil and non-empty, each task is round-robin assigned
// to a user via Business.Update after creation. Passing nil preserves the
// existing unassigned behavior.
func TestSeedPickTasks(ctx context.Context, n int, salesOrderIDs, salesOrderLineItemIDs, productIDs, locationIDs, createdByIDs, assigneeIDs []uuid.UUID, api *Business) ([]PickTask, error) {
	newTasks := TestNewPickTasks(n, salesOrderIDs, salesOrderLineItemIDs, productIDs, locationIDs, createdByIDs)

	tasks := make([]PickTask, len(newTasks))
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

	if len(assigneeIDs) > 0 {
		for i := range tasks {
			assignee := assigneeIDs[i%len(assigneeIDs)]
			updated, err := api.Update(ctx, tasks[i], UpdatePickTask{AssignedTo: &assignee})
			if err != nil {
				return nil, fmt.Errorf("assign pick task %d: %w", i, err)
			}
			tasks[i] = updated
		}
	}

	return tasks, nil
}
