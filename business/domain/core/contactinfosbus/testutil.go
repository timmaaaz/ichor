package contactinfosbus

import (
	"context"
	"fmt"
	"math/rand"
	"sort"

	"github.com/google/uuid"
)

// generatePhoneNumber creates a random phone number in (XXX) XXX-XXXX format.
func generatePhoneNumber() string {
	areaCode := rand.Intn(900) + 100    // 100-999
	exchange := rand.Intn(900) + 100    // 100-999
	subscriber := rand.Intn(10000)      // 0000-9999
	return fmt.Sprintf("(%03d) %03d-%04d", areaCode, exchange, subscriber)
}

// generateEmail creates an email using first and last name.
func generateEmail(firstName, lastName string) string {
	domains := []string{"gmail.com", "yahoo.com", "outlook.com", "email.com"}
	domain := domains[rand.Intn(len(domains))]
	return fmt.Sprintf("%s.%s@%s", firstName, lastName, domain)
}

func TestNewContactInfos(n int, streetIDs, timezoneIDs uuid.UUIDs) []NewContactInfos {
	newContactInfos := make([]NewContactInfos, n)

	idx := rand.Intn(10000)
	for i := 0; i < n; i++ {
		idx++

		firstName := fmt.Sprintf("First%d", idx)
		lastName := fmt.Sprintf("Last%d", idx)

		nc := NewContactInfos{
			FirstName:            firstName,
			LastName:             lastName,
			EmailAddress:         generateEmail(firstName, lastName),
			PrimaryPhone:         generatePhoneNumber(),
			SecondaryPhone:       generatePhoneNumber(),
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
