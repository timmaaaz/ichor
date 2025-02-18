package contactinfobus_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/timmaaaz/ichor/business/domain/core/contactinfobus"
	"github.com/timmaaaz/ichor/business/domain/users/userbus"
	"github.com/timmaaaz/ichor/business/sdk/dbtest"
	"github.com/timmaaaz/ichor/business/sdk/order"
	"github.com/timmaaaz/ichor/business/sdk/page"
	"github.com/timmaaaz/ichor/business/sdk/unitest"
)

func Test_ContactInfo(t *testing.T) {
	t.Parallel()

	db := dbtest.NewDatabase(t, "Test_ContactInfo")

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

func insertSeedData(busdomain dbtest.BusDomain) (unitest.SeedData, error) {
	ctx := context.Background()

	admins, err := userbus.TestSeedUsersWithNoFKs(ctx, 1, userbus.Roles.Admin, busdomain.User)
	if err != nil {
		return unitest.SeedData{}, fmt.Errorf("seeding user : %w", err)
	}

	contactInfo, err := contactinfobus.TestSeedContactInfo(ctx, 15, busdomain.ContactInfo)
	if err != nil {
		return unitest.SeedData{}, fmt.Errorf("seeding contact info : %w", err)
	}

	return unitest.SeedData{
		Admins:      []unitest.User{{User: admins[0]}},
		ContactInfo: contactInfo,
	}, nil
}

func query(busDomain dbtest.BusDomain, sd unitest.SeedData) []unitest.Table {
	return []unitest.Table{
		{
			Name: "Query",
			ExpResp: []contactinfobus.ContactInfo{
				sd.ContactInfo[0],
				sd.ContactInfo[1],
				sd.ContactInfo[2],
				sd.ContactInfo[3],
				sd.ContactInfo[4],
			},
			ExcFunc: func(ctx context.Context) any {
				got, err := busDomain.ContactInfo.Query(ctx, contactinfobus.QueryFilter{}, order.NewBy(contactinfobus.OrderByFirstName, order.ASC), page.MustParse("1", "5"))
				if err != nil {
					return err
				}
				return got
			},
			CmpFunc: func(got, exp any) string {
				gotResp, exists := got.([]contactinfobus.ContactInfo)
				if !exists {
					return fmt.Sprintf("got is not a slice of contact info: %v", got)
				}

				expResp := exp.([]contactinfobus.ContactInfo)

				return cmp.Diff(gotResp, expResp)

			},
		},
	}
}

func create(busDomain dbtest.BusDomain, sd unitest.SeedData) []unitest.Table {
	return []unitest.Table{
		{
			Name: "Create",
			ExpResp: contactinfobus.ContactInfo{
				FirstName:            "John",
				LastName:             "Doe",
				EmailAddress:         "johndoe@email.com",
				PrimaryPhone:         "222-222-2222",
				SecondaryPhone:       "333-333-3333",
				Address:              "123 Main St",
				AvailableHoursStart:  "8:00:00",
				AvailableHoursEnd:    "5:00:00",
				Timezone:             "EST",
				PreferredContactType: "phone",
			},
			ExcFunc: func(ctx context.Context) any {
				newContactInfo := contactinfobus.NewContactInfo{
					FirstName:            "John",
					LastName:             "Doe",
					EmailAddress:         "johndoe@email.com",
					PrimaryPhone:         "222-222-2222",
					SecondaryPhone:       "333-333-3333",
					Address:              "123 Main St",
					AvailableHoursStart:  "8:00:00",
					AvailableHoursEnd:    "5:00:00",
					Timezone:             "EST",
					PreferredContactType: "phone",
				}

				ci, err := busDomain.ContactInfo.Create(ctx, newContactInfo)
				if err != nil {
					return err
				}

				return ci
			},
			CmpFunc: func(got, exp any) string {
				gotResp, exists := got.(contactinfobus.ContactInfo)
				if !exists {
					return fmt.Sprintf("got is not a contact info: %v", got)
				}

				expResp := exp.(contactinfobus.ContactInfo)

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
			ExpResp: contactinfobus.ContactInfo{
				ID:                   sd.ContactInfo[0].ID,
				FirstName:            "Jane",
				LastName:             "Doe",
				EmailAddress:         "janedoe@email.com",
				PrimaryPhone:         "444-444-4444",
				SecondaryPhone:       "555-555-5555",
				Address:              "456 Elm St",
				AvailableHoursStart:  "9:00:00",
				AvailableHoursEnd:    "6:00:00",
				Timezone:             "PST",
				PreferredContactType: "email",
				Notes:                sd.ContactInfo[0].Notes,
			},
			ExcFunc: func(ctx context.Context) any {
				uc := contactinfobus.UpdateContactInfo{
					FirstName:            dbtest.StringPointer("Jane"),
					LastName:             dbtest.StringPointer("Doe"),
					EmailAddress:         dbtest.StringPointer("janedoe@email.com"),
					PrimaryPhone:         dbtest.StringPointer("444-444-4444"),
					SecondaryPhone:       dbtest.StringPointer("555-555-5555"),
					Address:              dbtest.StringPointer("456 Elm St"),
					AvailableHoursStart:  dbtest.StringPointer("9:00:00"),
					AvailableHoursEnd:    dbtest.StringPointer("6:00:00"),
					Timezone:             dbtest.StringPointer("PST"),
					PreferredContactType: dbtest.StringPointer("email"),
				}

				got, err := busDomain.ContactInfo.Update(ctx, sd.ContactInfo[0], uc)
				if err != nil {
					return err
				}

				return got
			},
			CmpFunc: func(got, exp any) string {
				gotResp, exists := got.(contactinfobus.ContactInfo)
				if !exists {
					return fmt.Sprintf("got is not a contact info: %v", got)
				}

				expResp := exp.(contactinfobus.ContactInfo)

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
				err := busDomain.ContactInfo.Delete(ctx, sd.ContactInfo[0])
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
