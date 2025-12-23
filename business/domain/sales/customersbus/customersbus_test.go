package customersbus_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/core/contactinfosbus"
	"github.com/timmaaaz/ichor/business/domain/core/userbus"
	"github.com/timmaaaz/ichor/business/domain/geography/citybus"
	"github.com/timmaaaz/ichor/business/domain/geography/regionbus"
	"github.com/timmaaaz/ichor/business/domain/geography/streetbus"
	"github.com/timmaaaz/ichor/business/domain/sales/customersbus"
	"github.com/timmaaaz/ichor/business/sdk/dbtest"
	"github.com/timmaaaz/ichor/business/sdk/order"
	"github.com/timmaaaz/ichor/business/sdk/page"
	"github.com/timmaaaz/ichor/business/sdk/unitest"
)

func Test_Customers(t *testing.T) {
	t.Parallel()

	db := dbtest.NewDatabase(t, "Test_Customers")

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

	customersCount := 5

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

	ctys, err := citybus.TestSeedCities(ctx, customersCount, ids, busDomain.City)
	if err != nil {
		return unitest.SeedData{}, fmt.Errorf("seeding cities : %w", err)
	}

	ctyIDs := make([]uuid.UUID, 0, len(ctys))
	for _, c := range ctys {
		ctyIDs = append(ctyIDs, c.ID)
	}

	strs, err := streetbus.TestSeedStreets(ctx, customersCount, ctyIDs, busDomain.Street)
	if err != nil {
		return unitest.SeedData{}, fmt.Errorf("seeding streets : %w", err)
	}
	strIDs := make([]uuid.UUID, 0, len(strs))
	for _, s := range strs {
		strIDs = append(strIDs, s.ID)
	}

	// Query timezones from seed data
	tzs, err := busDomain.Timezone.QueryAll(ctx)
	if err != nil {
		return unitest.SeedData{}, fmt.Errorf("querying timezones : %w", err)
	}
	tzIDs := make([]uuid.UUID, 0, len(tzs))
	for _, tz := range tzs {
		tzIDs = append(tzIDs, tz.ID)
	}

	contactInfos, err := contactinfosbus.TestSeedContactInfos(ctx, customersCount, strIDs, tzIDs, busDomain.ContactInfos)
	if err != nil {
		return unitest.SeedData{}, fmt.Errorf("seeding contact info : %w", err)
	}
	contactInfoIDs := make([]uuid.UUID, 0, len(contactInfos))
	for _, ci := range contactInfos {
		contactInfoIDs = append(contactInfoIDs, ci.ID)
	}

	customers, err := customersbus.TestSeedCustomers(ctx, customersCount, strIDs, contactInfoIDs, uuid.UUIDs{admins[0].ID}, busDomain.Customers)
	if err != nil {
		return unitest.SeedData{}, fmt.Errorf("seeding customers : %w", err)
	}

	return unitest.SeedData{
		Admins:       []unitest.User{{User: admins[0]}},
		ContactInfos: contactInfos,
		Streets:      strs,
		Customers:    customers,
	}, nil
}

func query(busDomain dbtest.BusDomain, sd unitest.SeedData) []unitest.Table {
	return []unitest.Table{
		{
			Name: "Query",
			ExpResp: []customersbus.Customers{
				sd.Customers[0],
				sd.Customers[1],
				sd.Customers[2],
				sd.Customers[3],
				sd.Customers[4],
			},
			ExcFunc: func(ctx context.Context) any {
				got, err := busDomain.Customers.Query(ctx, customersbus.QueryFilter{}, order.NewBy(customersbus.OrderByName, order.ASC), page.MustParse("1", "5"))
				if err != nil {
					return err
				}
				return got
			},
			CmpFunc: func(got, exp any) string {
				gotResp, exists := got.([]customersbus.Customers)
				if !exists {
					return fmt.Sprintf("got is not a slice of contact info: %v", got)
				}

				expResp := exp.([]customersbus.Customers)

				for i := range gotResp {

					if gotResp[i].CreatedDate.Format(time.RFC3339) == expResp[i].CreatedDate.Format(time.RFC3339) {
						expResp[i].CreatedDate = gotResp[i].CreatedDate
					}
					if gotResp[i].UpdatedDate.Format(time.RFC3339) == expResp[i].UpdatedDate.Format(time.RFC3339) {
						expResp[i].UpdatedDate = gotResp[i].UpdatedDate
					}
				}

				return cmp.Diff(gotResp, expResp)

			},
		},
	}
}

func create(busDomain dbtest.BusDomain, sd unitest.SeedData) []unitest.Table {
	return []unitest.Table{
		{
			Name: "Create",
			ExpResp: customersbus.Customers{
				Name:              "New Customer",
				ContactID:         sd.ContactInfos[0].ID,
				DeliveryAddressID: sd.Streets[0].ID,
				Notes:             "New customer notes",
				CreatedDate:       time.Now().UTC(),
				UpdatedDate:       time.Now().UTC(),
				CreatedBy:         sd.Admins[0].ID,
				UpdatedBy:         sd.Admins[0].ID,
			},
			ExcFunc: func(ctx context.Context) any {
				newCustomers := customersbus.NewCustomers{
					Name:              "New Customer",
					ContactID:         sd.ContactInfos[0].ID,
					DeliveryAddressID: sd.Streets[0].ID,
					Notes:             "New customer notes",
					CreatedBy:         sd.Admins[0].ID,
				}

				ci, err := busDomain.Customers.Create(ctx, newCustomers)
				if err != nil {
					return err
				}

				return ci
			},
			CmpFunc: func(got, exp any) string {
				gotResp, exists := got.(customersbus.Customers)
				if !exists {
					return fmt.Sprintf("got is not a customer: %v", got)
				}

				expResp := exp.(customersbus.Customers)

				expResp.ID = gotResp.ID
				expResp.CreatedDate = gotResp.CreatedDate
				expResp.UpdatedDate = gotResp.UpdatedDate

				return cmp.Diff(gotResp, expResp)
			},
		},
	}
}

func update(busDomain dbtest.BusDomain, sd unitest.SeedData) []unitest.Table {
	return []unitest.Table{
		{
			Name: "Update",
			ExpResp: customersbus.Customers{
				ID:                sd.Customers[0].ID,
				Name:              "Updated Test Customer",
				ContactID:         sd.Customers[0].ContactID,
				DeliveryAddressID: sd.Customers[0].DeliveryAddressID,
				Notes:             sd.Customers[0].Notes,
				CreatedBy:         sd.Customers[0].CreatedBy,
				UpdatedBy:         sd.Admins[0].ID,
				CreatedDate:       sd.Customers[0].CreatedDate,
			},
			ExcFunc: func(ctx context.Context) any {
				uc := customersbus.UpdateCustomers{
					Name: dbtest.StringPointer("Updated Test Customer"),
				}

				got, err := busDomain.Customers.Update(ctx, sd.Customers[0], uc)
				if err != nil {
					return err
				}

				return got
			},
			CmpFunc: func(got, exp any) string {
				gotResp, exists := got.(customersbus.Customers)
				if !exists {
					return fmt.Sprintf("got is not a contact info: %v", got)
				}

				expResp := exp.(customersbus.Customers)

				expResp.ID = gotResp.ID
				expResp.UpdatedDate = gotResp.UpdatedDate

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
				err := busDomain.Customers.Delete(ctx, sd.Customers[0])
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
