package contactinfosbus_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/core/contactinfosbus"
	"github.com/timmaaaz/ichor/business/domain/location/citybus"
	"github.com/timmaaaz/ichor/business/domain/location/regionbus"
	"github.com/timmaaaz/ichor/business/domain/location/streetbus"
	"github.com/timmaaaz/ichor/business/domain/users/userbus"
	"github.com/timmaaaz/ichor/business/sdk/dbtest"
	"github.com/timmaaaz/ichor/business/sdk/order"
	"github.com/timmaaaz/ichor/business/sdk/page"
	"github.com/timmaaaz/ichor/business/sdk/unitest"
)

func Test_ContactInfos(t *testing.T) {
	t.Parallel()

	db := dbtest.NewDatabase(t, "Test_ContactInfos")

	sd, err := insertSeedData(db.BusDomain)
	if err != nil {
		t.Fatalf("Seeding error: %s", err)
	}

	// -------------------------------------------------------------------------
	unitest.Run(t, query(db.BusDomain, sd), "query")
	unitest.Run(t, create(db.BusDomain, sd), "create")
	unitest.Run(t, update(db.BusDomain, sd), "update")
	unitest.Run(t, delete(db.BusDomain, sd), "delete")
}

func insertSeedData(busDomain dbtest.BusDomain) (unitest.SeedData, error) {
	ctx := context.Background()

	contactInfosCount := 5

	admins, err := userbus.TestSeedUsersWithNoFKs(ctx, 1, userbus.Roles.Admin, busDomain.User)
	if err != nil {
		return unitest.SeedData{}, fmt.Errorf("seeding user : %w", err)
	}

	// ADDRESSES
	regions, err := busDomain.Region.Query(ctx, regionbus.QueryFilter{}, regionbus.DefaultOrderBy, page.MustParse("1", "5"))
	if err != nil {
		return unitest.SeedData{}, fmt.Errorf("querying regions : %w", err)
	}

	ids := make([]uuid.UUID, 0, len(regions))
	for _, r := range regions {
		ids = append(ids, r.ID)
	}

	ctys, err := citybus.TestSeedCities(ctx, contactInfosCount, ids, busDomain.City)
	if err != nil {
		return unitest.SeedData{}, fmt.Errorf("seeding cities : %w", err)
	}

	ctyIDs := make([]uuid.UUID, 0, len(ctys))
	for _, c := range ctys {
		ctyIDs = append(ctyIDs, c.ID)
	}

	strs, err := streetbus.TestSeedStreets(ctx, contactInfosCount, ctyIDs, busDomain.Street)
	if err != nil {
		return unitest.SeedData{}, fmt.Errorf("seeding streets : %w", err)
	}
	strIDs := make([]uuid.UUID, 0, len(strs))
	for _, s := range strs {
		strIDs = append(strIDs, s.ID)
	}

	contactInfos, err := contactinfosbus.TestSeedContactInfos(ctx, contactInfosCount, strIDs, busDomain.ContactInfos)
	if err != nil {
		return unitest.SeedData{}, fmt.Errorf("seeding contact info : %w", err)
	}

	return unitest.SeedData{
		Admins:       []unitest.User{{User: admins[0]}},
		ContactInfos: contactInfos,
		Streets:      strs,
	}, nil
}

func query(busDomain dbtest.BusDomain, sd unitest.SeedData) []unitest.Table {
	return []unitest.Table{
		{
			Name: "Query",
			ExpResp: []contactinfosbus.ContactInfos{
				sd.ContactInfos[0],
				sd.ContactInfos[1],
				sd.ContactInfos[2],
				sd.ContactInfos[3],
				sd.ContactInfos[4],
			},
			ExcFunc: func(ctx context.Context) any {
				got, err := busDomain.ContactInfos.Query(ctx, contactinfosbus.QueryFilter{}, order.NewBy(contactinfosbus.OrderByFirstName, order.ASC), page.MustParse("1", "5"))
				if err != nil {
					return err
				}
				return got
			},
			CmpFunc: func(got, exp any) string {
				gotResp, exists := got.([]contactinfosbus.ContactInfos)
				if !exists {
					return fmt.Sprintf("got is not a slice of contact info: %v", got)
				}

				expResp := exp.([]contactinfosbus.ContactInfos)

				return cmp.Diff(gotResp, expResp)

			},
		},
	}
}

func create(busDomain dbtest.BusDomain, sd unitest.SeedData) []unitest.Table {
	return []unitest.Table{
		{
			Name: "Create",
			ExpResp: contactinfosbus.ContactInfos{
				FirstName:            "John",
				LastName:             "Doe",
				EmailAddress:         "johndoe@email.com",
				PrimaryPhone:         "222-222-2222",
				SecondaryPhone:       "333-333-3333",
				StreetID:             sd.Streets[0].ID,
				AvailableHoursStart:  "8:00:00",
				AvailableHoursEnd:    "5:00:00",
				Timezone:             "EST",
				PreferredContactType: "phone",
			},
			ExcFunc: func(ctx context.Context) any {
				newContactInfos := contactinfosbus.NewContactInfos{
					FirstName:            "John",
					LastName:             "Doe",
					EmailAddress:         "johndoe@email.com",
					PrimaryPhone:         "222-222-2222",
					SecondaryPhone:       "333-333-3333",
					StreetID:             sd.Streets[0].ID,
					AvailableHoursStart:  "8:00:00",
					AvailableHoursEnd:    "5:00:00",
					Timezone:             "EST",
					PreferredContactType: "phone",
				}

				ci, err := busDomain.ContactInfos.Create(ctx, newContactInfos)
				if err != nil {
					return err
				}

				return ci
			},
			CmpFunc: func(got, exp any) string {
				gotResp, exists := got.(contactinfosbus.ContactInfos)
				if !exists {
					return fmt.Sprintf("got is not a contact info: %v", got)
				}

				expResp := exp.(contactinfosbus.ContactInfos)

				expResp.ID = gotResp.ID

				return cmp.Diff(gotResp, expResp)
			},
		},
	}
}

func update(busDomain dbtest.BusDomain, sd unitest.SeedData) []unitest.Table {
	return []unitest.Table{
		{
			Name: "Update",
			ExpResp: contactinfosbus.ContactInfos{
				ID:                   sd.ContactInfos[0].ID,
				FirstName:            "Jane",
				LastName:             "Doe",
				EmailAddress:         "janedoe@email.com",
				PrimaryPhone:         "444-444-4444",
				SecondaryPhone:       "555-555-5555",
				StreetID:             sd.Streets[1].ID,
				AvailableHoursStart:  "9:00:00",
				AvailableHoursEnd:    "6:00:00",
				Timezone:             "PST",
				PreferredContactType: "email",
				Notes:                sd.ContactInfos[0].Notes,
			},
			ExcFunc: func(ctx context.Context) any {
				uc := contactinfosbus.UpdateContactInfos{
					FirstName:            dbtest.StringPointer("Jane"),
					LastName:             dbtest.StringPointer("Doe"),
					EmailAddress:         dbtest.StringPointer("janedoe@email.com"),
					PrimaryPhone:         dbtest.StringPointer("444-444-4444"),
					SecondaryPhone:       dbtest.StringPointer("555-555-5555"),
					StreetID:             &sd.Streets[1].ID,
					AvailableHoursStart:  dbtest.StringPointer("9:00:00"),
					AvailableHoursEnd:    dbtest.StringPointer("6:00:00"),
					Timezone:             dbtest.StringPointer("PST"),
					PreferredContactType: dbtest.StringPointer("email"),
				}

				got, err := busDomain.ContactInfos.Update(ctx, sd.ContactInfos[0], uc)
				if err != nil {
					return err
				}

				return got
			},
			CmpFunc: func(got, exp any) string {
				gotResp, exists := got.(contactinfosbus.ContactInfos)
				if !exists {
					return fmt.Sprintf("got is not a contact info: %v", got)
				}

				expResp := exp.(contactinfosbus.ContactInfos)

				expResp.ID = gotResp.ID

				return cmp.Diff(gotResp, expResp)
			},
		},
	}
}

func delete(busDomain dbtest.BusDomain, sd unitest.SeedData) []unitest.Table {
	return []unitest.Table{
		{
			Name:    "Delete",
			ExpResp: nil,
			ExcFunc: func(ctx context.Context) any {
				err := busDomain.ContactInfos.Delete(ctx, sd.ContactInfos[0])
				if err != nil {
					return err
				}
				return nil
			},
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
}
