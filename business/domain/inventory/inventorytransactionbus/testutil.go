package inventorytransactionbus

import (
	"context"
	"fmt"
	"math/rand"
	"sort"
	"time"

	"github.com/google/uuid"
)

func TestNewInventoryTransaction(n int, locationIDs, productIDs, userIDs uuid.UUIDs) []NewInventoryTransaction {
	newInventoryTransactions := make([]NewInventoryTransaction, n)

	floorWorker1 := uuid.MustParse("c0000000-0000-4000-8000-000000000001")

	transactionTypes := []string{"receive", "pick", "putaway", "transfer", "adjustment", "count"}

	idx := rand.Intn(10000)

	for i := 0; i < n; i++ {
		idx++

		userID := userIDs[idx%len(userIDs)]
		if i < 5 {
			userID = floorWorker1
		}

		newInventoryTransactions[i] = NewInventoryTransaction{
			LocationID:      locationIDs[idx%len(locationIDs)],
			ProductID:       productIDs[idx%len(productIDs)],
			UserID:          userID,
			TransactionType: transactionTypes[i%len(transactionTypes)],
			Quantity:        rand.Intn(100) + 1,
			ReferenceNumber: fmt.Sprintf("REF-%04d", i+1),
			TransactionDate: time.Now().AddDate(0, 0, -(i % 7)),
		}
	}

	return newInventoryTransactions
}

func TestSeedInventoryTransaction(ctx context.Context, n int, locationIDs, productIDs, userIDs uuid.UUIDs, api *Business) ([]InventoryTransaction, error) {
	newInventoryTransactions := TestNewInventoryTransaction(n, locationIDs, productIDs, userIDs)

	inventoryTransactions := make([]InventoryTransaction, len(newInventoryTransactions))
	for i, nit := range newInventoryTransactions {
		it, err := api.Create(ctx, nit)
		if err != nil {
			return []InventoryTransaction{}, err
		}
		inventoryTransactions[i] = it
	}

	sort.Slice(inventoryTransactions, func(i, j int) bool {
		return inventoryTransactions[i].InventoryTransactionID.String() < inventoryTransactions[j].InventoryTransactionID.String()
	})

	return inventoryTransactions, nil
}
