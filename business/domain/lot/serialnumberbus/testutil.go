package serialnumberbus

import (
	"context"
	"fmt"
	"math/rand"
	"sort"

	"github.com/google/uuid"
)

func TestNewSerialNumbers(n int, lotIDs, productIDs, locationIDs []uuid.UUID) []NewSerialNumber {

	newSerialNumbers := make([]NewSerialNumber, n)

	idx := rand.Intn(10000)
	for i := 0; i < n; i++ {
		idx++
		newSerialNumbers[i] = NewSerialNumber{
			LotID:        lotIDs[idx%len(lotIDs)],
			ProductID:    productIDs[idx%len(productIDs)],
			LocationID:   locationIDs[idx%len(locationIDs)],
			SerialNumber: fmt.Sprintf("SN-%d", idx),
			Status:       fmt.Sprintf("Status-%d", idx%2),
		}
	}

	return newSerialNumbers
}

func TestSeedSerialNumbers(ctx context.Context, n int, lotIDs, productIDs, locationIDs []uuid.UUID, api *Business) ([]SerialNumber, error) {
	newSerialNumbers := TestNewSerialNumbers(n, lotIDs, productIDs, locationIDs)

	serialNumbers := make([]SerialNumber, len(newSerialNumbers))

	for i, ns := range newSerialNumbers {
		sn, err := api.Create(ctx, ns)
		if err != nil {
			return []SerialNumber{}, err
		}
		serialNumbers[i] = sn
	}

	sort.Slice(serialNumbers, func(i, j int) bool {
		return serialNumbers[i].SerialID.String() < serialNumbers[j].SerialID.String()
	})

	return serialNumbers, nil
}
