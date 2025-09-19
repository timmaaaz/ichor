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

	idx := rand.Intn(10000)

	for i := 0; i < n; i++ {
		idx++
		newInventoryTransactions[i] = NewInventoryTransaction{
			LocationID:      locationIDs[idx%len(locationIDs)],
			ProductID:       productIDs[idx%len(productIDs)],
			UserID:          userIDs[idx%len(userIDs)],
			TransactionType: "Movement",
			Quantity:        rand.Intn(100),
			ReferenceNumber: fmt.Sprintf("ref_%d", idx),
			TransactionDate: time.Now(),
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
