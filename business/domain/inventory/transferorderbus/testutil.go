package transferorderbus

import (
	"context"
	"math/rand"
	"sort"
	"time"

	"github.com/google/uuid"
)

func TestNewTransferOrders(n int, productIDs, fromLocationIDs, toLocationIDs, requestedBy, approvedBy []uuid.UUID) []NewTransferOrder {
	newTransferOrders := make([]NewTransferOrder, n)

	// Status distribution: ~40% pending, ~40% approved, ~20% completed
	transferStatuses := []string{StatusPending, StatusPending, StatusApproved, StatusApproved, StatusCompleted}

	idx := rand.Intn(10000)
	for i := range n {
		idx++
		newTransferOrders[i] = NewTransferOrder{
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

func TestSeedTransferOrders(ctx context.Context, n int, productIDs, fromLocationIDs, toLocationIDs, requestedBy, approvedBy []uuid.UUID, api *Business) ([]TransferOrder, error) {

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

	return transferOrders, nil
}
