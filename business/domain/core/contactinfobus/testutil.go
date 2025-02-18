package contactinfobus

import (
	"context"
	"fmt"
	"math/rand"
	"sort"
)

func TestNewContactInfo(n int) []NewContactInfo {
	newContactInfo := make([]NewContactInfo, n)

	idx := rand.Intn(10000)
	for i := 0; i < n; i++ {
		idx++

		nc := NewContactInfo{
			FirstName:            fmt.Sprintf("FirstName%d", idx),
			LastName:             fmt.Sprintf("LastName%d", idx),
			EmailAddress:         fmt.Sprintf("EmailAddress%d", idx),
			PrimaryPhone:         fmt.Sprintf("PrimaryPhone%d", idx),
			SecondaryPhone:       fmt.Sprintf("SecondaryPhone%d", idx),
			Address:              fmt.Sprintf("Address%d", idx),
			AvailableHoursStart:  "8:00:00",
			AvailableHoursEnd:    "5:00:00",
			Timezone:             "EST",
			PreferredContactType: "phone",
			Notes:                fmt.Sprintf("Notes%d", idx),
		}
		newContactInfo[i] = nc
	}

	return newContactInfo
}

func TestSeedContactInfo(ctx context.Context, n int, api *Business) ([]ContactInfo, error) {
	newContactInfos := TestNewContactInfo(n)

	contactInfos := make([]ContactInfo, len(newContactInfos))

	for i, nci := range newContactInfos {
		contactInfo, err := api.Create(ctx, nci)
		if err != nil {
			return nil, fmt.Errorf("seeding contact info: idx: %d : %w", i, err)
		}

		contactInfos[i] = contactInfo
	}

	sort.Slice(contactInfos, func(i, j int) bool {
		return contactInfos[i].FirstName <= contactInfos[j].FirstName
	})

	return contactInfos, nil
}
