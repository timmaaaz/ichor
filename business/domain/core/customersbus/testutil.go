package customersbus

import (
	"context"
	"fmt"
	"math/rand"
	"sort"

	"github.com/google/uuid"
)

func TestNewCustomers(n int, streetIDs uuid.UUIDs, contactIDs uuid.UUIDs, userIDs uuid.UUIDs) []NewCustomers {
	newCustomers := make([]NewCustomers, n)

	idx := rand.Intn(10000)
	for i := 0; i < n; i++ {
		idx++

		nc := NewCustomers{
			Name:              fmt.Sprintf("Customer%d", idx),
			ContactID:         contactIDs[i%len(contactIDs)],
			DeliveryAddressID: streetIDs[(i+1)%len(streetIDs)],
			Notes:             fmt.Sprintf("Notes%d", idx),
			CreatedBy:         userIDs[i%len(userIDs)],
		}
		newCustomers[i] = nc
	}

	return newCustomers
}

func TestSeedCustomers(ctx context.Context, n int, streetIDs uuid.UUIDs, contactIDs uuid.UUIDs, userIDs uuid.UUIDs, api *Business) ([]Customers, error) {
	newCustomerss := TestNewCustomers(n, streetIDs, contactIDs, userIDs)

	customerss := make([]Customers, len(newCustomerss))

	for i, nci := range newCustomerss {
		customers, err := api.Create(ctx, nci)
		if err != nil {
			return nil, fmt.Errorf("seeding contact info: idx: %d : %w", i, err)
		}

		customerss[i] = customers
	}

	// Match up with the queryfilter
	sort.Slice(customerss, func(i, j int) bool {
		return customerss[i].Name <= customerss[j].Name
	})

	return customerss, nil
}
