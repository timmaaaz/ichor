package contactinfosbus

import (
	"context"
	"fmt"
	"math/rand"
	"sort"

	"github.com/google/uuid"
)

func TestNewContactInfos(n int, streetIDs, timezoneIDs uuid.UUIDs) []NewContactInfos {
	newContactInfos := make([]NewContactInfos, n)

	idx := rand.Intn(10000)
	for i := 0; i < n; i++ {
		idx++

		nc := NewContactInfos{
			FirstName:            fmt.Sprintf("FirstName%d", idx),
			LastName:             fmt.Sprintf("LastName%d", idx),
			EmailAddress:         fmt.Sprintf("EmailAddress%d", idx),
			PrimaryPhone:         fmt.Sprintf("PrimaryPhone%d", idx),
			SecondaryPhone:       fmt.Sprintf("SecondaryPhone%d", idx),
			StreetID:             streetIDs[i%len(streetIDs)],
			DeliveryAddressID:    streetIDs[(i+1)%len(streetIDs)],
			AvailableHoursStart:  "8:00:00",
			AvailableHoursEnd:    "5:00:00",
			TimezoneID:           timezoneIDs[i%len(timezoneIDs)],
			PreferredContactType: "phone",
			Notes:                fmt.Sprintf("Notes%d", idx),
		}
		newContactInfos[i] = nc
	}

	return newContactInfos
}

func TestSeedContactInfos(ctx context.Context, n int, streetIDs, timezoneIDs uuid.UUIDs, api *Business) ([]ContactInfos, error) {
	newContactInfoss := TestNewContactInfos(n, streetIDs, timezoneIDs)

	contactInfoss := make([]ContactInfos, len(newContactInfoss))

	for i, nci := range newContactInfoss {
		contactInfos, err := api.Create(ctx, nci)
		if err != nil {
			return nil, fmt.Errorf("seeding contact info: idx: %d : %w", i, err)
		}

		contactInfoss[i] = contactInfos
	}

	sort.Slice(contactInfoss, func(i, j int) bool {
		return contactInfoss[i].FirstName <= contactInfoss[j].FirstName
	})

	return contactInfoss, nil
}
