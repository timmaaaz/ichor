package transferorderbus

import (
	"context"
	"fmt"
	"math/rand"
	"sort"
	"time"

	"github.com/google/uuid"
)

func TestNewTransferOrders(n int, productIDs, fromLocationIDs, toLocationIDs, requestedBy, approvedBy []uuid.UUID) []NewTransferOrder {
	newTransferOrders := make([]NewTransferOrder, n)

	// Status distribution: ~40% pending, ~40% approved, ~20% completed
	transferStatuses := []string{StatusPending, StatusPending, StatusApproved, StatusApproved, StatusCompleted}

	today := time.Now().Format("060102")
	idx := rand.Intn(10000)
	for i := range n {
		idx++
		num := fmt.Sprintf("XFER-%s-%04d", today, i+1)
		newTransferOrders[i] = NewTransferOrder{
			TransferNumber: &num,
			ProductID:      productIDs[idx%len(productIDs)],
			FromLocationID: fromLocationIDs[idx%len(fromLocationIDs)],
			ToLocationID:   toLocationIDs[idx%len(toLocationIDs)],
			RequestedByID:  requestedBy[idx%len(requestedBy)],
			ApprovedByID:   func() *uuid.UUID { v := approvedBy[idx%len(approvedBy)]; return &v }(),
			Quantity:       idx,
			Status:         transferStatuses[i%len(transferStatuses)],
			TransferDate:   time.Now(),
		}
	}

	return newTransferOrders
}

// TestSeedTransferOrders creates n transfer orders for testing. If assigneeIDs
// is non-empty, each created order is claimed via Business.Update in a
// round-robin fashion over the provided IDs. Passing nil preserves the
// previous behavior of leaving orders unclaimed.
func TestSeedTransferOrders(ctx context.Context, n int, productIDs, fromLocationIDs, toLocationIDs, requestedBy, approvedBy, assigneeIDs []uuid.UUID, api *Business) ([]TransferOrder, error) {

	newTransferOrders := TestNewTransferOrders(n, productIDs, fromLocationIDs, toLocationIDs, requestedBy, approvedBy)
	transferOrders := make([]TransferOrder, len(newTransferOrders))
	for i, nto := range newTransferOrders {
		to, err := api.Create(ctx, nto)
		if err != nil {
			return []TransferOrder{}, err
		}
		transferOrders[i] = to
	}

	sort.Slice(transferOrders, func(i, j int) bool {
		return transferOrders[i].TransferID.String() < transferOrders[j].TransferID.String()
	})

	if len(assigneeIDs) > 0 {
		for i := range transferOrders {
			assignee := assigneeIDs[i%len(assigneeIDs)]
			updated, err := api.Update(ctx, transferOrders[i], UpdateTransferOrder{ClaimedByID: &assignee})
			if err != nil {
				return nil, fmt.Errorf("assign transfer order %d: %w", i, err)
			}
			transferOrders[i] = updated
		}
	}

	return transferOrders, nil
}
