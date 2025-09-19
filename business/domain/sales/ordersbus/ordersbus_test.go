package ordersbus_test

import (
	"context"
	"fmt"
	"sort"
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
	"github.com/timmaaaz/ichor/business/domain/sales/orderfulfillmentstatusbus"
	"github.com/timmaaaz/ichor/business/domain/sales/ordersbus"
	"github.com/timmaaaz/ichor/business/sdk/dbtest"
	"github.com/timmaaaz/ichor/business/sdk/page"
	"github.com/timmaaaz/ichor/business/sdk/unitest"
)

func Test_Order(t *testing.T) {
	t.Parallel()

	db := dbtest.NewDatabase(t, "Test_Order")

	sd, err := insertSeedData(db.BusDomain)
	if err != nil {
		t.Fatalf("Seeding error: %s", err)
	}

	unitest.Run(t, query(db.BusDomain, sd), "query")
	unitest.Run(t, create(db.BusDomain, sd), "create")
	unitest.Run(t, update(db.BusDomain, sd), "update")
	unitest.Run(t, delete(db.BusDomain, sd), "delete")
}

func insertSeedData(busDomain dbtest.BusDomain) (unitest.SeedData, error) {
	ctx := context.Background()

	admins, err := userbus.TestSeedUsersWithNoFKs(ctx, 1, userbus.Roles.Admin, busDomain.User)
	if err != nil {
		return unitest.SeedData{}, fmt.Errorf("seeding user : %w", err)
	}

	count := 5

	// ADDRESSES
	regions, err := busDomain.Region.Query(ctx, regionbus.QueryFilter{}, regionbus.DefaultOrderBy, page.MustParse("1", "5"))
	if err != nil {
		return unitest.SeedData{}, fmt.Errorf("querying regions : %w", err)
	}

	ids := make([]uuid.UUID, 0, len(regions))
	for _, r := range regions {
		ids = append(ids, r.ID)
	}

	ctys, err := citybus.TestSeedCities(ctx, count, ids, busDomain.City)
	if err != nil {
		return unitest.SeedData{}, fmt.Errorf("seeding cities : %w", err)
	}

	ctyIDs := make([]uuid.UUID, 0, len(ctys))
	for _, c := range ctys {
		ctyIDs = append(ctyIDs, c.ID)
	}

	strs, err := streetbus.TestSeedStreets(ctx, count, ctyIDs, busDomain.Street)
	if err != nil {
		return unitest.SeedData{}, fmt.Errorf("seeding streets : %w", err)
	}
	strIDs := make([]uuid.UUID, 0, len(strs))
	for _, s := range strs {
		strIDs = append(strIDs, s.ID)
	}

	contactInfos, err := contactinfosbus.TestSeedContactInfos(ctx, count, strIDs, busDomain.ContactInfos)
	if err != nil {
		return unitest.SeedData{}, fmt.Errorf("seeding contact info : %w", err)
	}
	contactInfoIDs := make([]uuid.UUID, 0, len(contactInfos))
	for _, ci := range contactInfos {
		contactInfoIDs = append(contactInfoIDs, ci.ID)
	}

	customers, err := customersbus.TestSeedCustomers(ctx, count, strIDs, contactInfoIDs, uuid.UUIDs{admins[0].ID}, busDomain.Customers)
	if err != nil {
		return unitest.SeedData{}, fmt.Errorf("seeding customers : %w", err)
	}
	customerIDs := make([]uuid.UUID, 0, len(customers))
	for _, c := range customers {
		customerIDs = append(customerIDs, c.ID)
	}

	ofls, err := orderfulfillmentstatusbus.TestSeedOrderFulfillmentStatuses(ctx, busDomain.OrderFulfillmentStatus)
	if err != nil {
		return unitest.SeedData{}, fmt.Errorf("seeding order fulfillment statuses: %w", err)
	}
	oflIDs := make([]uuid.UUID, 0, len(ofls))
	for _, ofl := range ofls {
		oflIDs = append(oflIDs, ofl.ID)
	}

	orders, err := ordersbus.TestSeedOrders(ctx, count, uuid.UUIDs{admins[0].ID}, customerIDs, oflIDs, busDomain.Order)
	if err != nil {
		return unitest.SeedData{}, fmt.Errorf("seeding Orders: %w", err)
	}

	return unitest.SeedData{
		Admins:                   []unitest.User{{User: admins[0]}},
		Orders:                   orders,
		Customers:                customers,
		OrderFulfillmentStatuses: ofls,
	}, nil
}

func query(busDomain dbtest.BusDomain, sd unitest.SeedData) []unitest.Table {
	return []unitest.Table{
		{
			Name: "Query",
			ExpResp: []ordersbus.Order{
				sd.Orders[0],
				sd.Orders[1],
				sd.Orders[2],
				sd.Orders[3],
				sd.Orders[4],
			},
			ExcFunc: func(ctx context.Context) any {
				got, err := busDomain.Order.Query(ctx, ordersbus.QueryFilter{}, ordersbus.DefaultOrderBy, page.MustParse("1", "5"))
				if err != nil {
					return err
				}
				return got
			},
			CmpFunc: func(got, exp any) string {
				gotResp, exists := got.([]ordersbus.Order)
				sort.Slice(gotResp, func(i, j int) bool {
					return gotResp[i].Number < gotResp[j].Number
				})

				if !exists {
					return fmt.Sprintf("expected []ordersbus.Order, got %T", got)
				}

				expResp, exists := exp.([]ordersbus.Order)
				if !exists {
					return fmt.Sprintf("expected []ordersbus.Order, got %T", exp)
				}
				sort.Slice(expResp, func(i, j int) bool {
					return expResp[i].Number < expResp[j].Number
				})

				return cmp.Diff(gotResp, expResp)
			},
		},
	}
}

func create(busDomain dbtest.BusDomain, sd unitest.SeedData) []unitest.Table {

	dueDate := time.Now().AddDate(0, 0, 10)
	return []unitest.Table{
		{
			Name: "Create",
			ExpResp: ordersbus.Order{
				Number:              "ORD-123",
				DueDate:             dueDate,
				CustomerID:          sd.Customers[0].ID,
				FulfillmentStatusID: sd.OrderFulfillmentStatuses[0].ID,
				CreatedBy:           sd.Admins[0].ID,
				UpdatedBy:           sd.Admins[0].ID,
			},
			ExcFunc: func(ctx context.Context) any {
				ofs, err := busDomain.Order.Create(ctx, ordersbus.NewOrder{
					Number:              "ORD-123",
					CustomerID:          sd.Customers[0].ID,
					DueDate:             dueDate,
					FulfillmentStatusID: sd.OrderFulfillmentStatuses[0].ID,
					CreatedBy:           sd.Admins[0].ID,
				})
				if err != nil {
					return err
				}
				return ofs
			},
			CmpFunc: func(got, exp any) string {
				gotResp, exists := got.(ordersbus.Order)
				if !exists {
					return fmt.Sprintf("expected ordersbus.Order, got %T", got)
				}

				expResp, exists := exp.(ordersbus.Order)
				if !exists {
					return fmt.Sprintf("expected ordersbus.Order, got %T", exp)
				}

				expResp.ID = gotResp.ID // Ignore ID for comparison
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
			ExpResp: ordersbus.Order{
				ID:                  sd.Orders[0].ID,
				Number:              "ORD-123-UPDATED",
				CustomerID:          sd.Customers[0].ID,
				FulfillmentStatusID: sd.OrderFulfillmentStatuses[0].ID,
				CreatedBy:           sd.Admins[0].ID,
				UpdatedBy:           sd.Admins[0].ID,
			},
			ExcFunc: func(ctx context.Context) any {
				got, err := busDomain.Order.Update(ctx, sd.Orders[0], ordersbus.UpdateOrder{
					Number: dbtest.StringPointer("ORD-123-UPDATED"),
				})
				if err != nil {
					return err
				}
				return got
			},
			CmpFunc: func(got, exp any) string {
				gotResp, exists := got.(ordersbus.Order)
				if !exists {
					return fmt.Sprintf("expected ordersbus.Order, got %T", got)
				}

				expResp, exists := exp.(ordersbus.Order)
				if !exists {
					return fmt.Sprintf("expected ordersbus.Order, got %T", exp)
				}

				expResp.DueDate = gotResp.DueDate
				expResp.CreatedDate = gotResp.CreatedDate
				expResp.UpdatedDate = gotResp.UpdatedDate

				return cmp.Diff(gotResp, expResp)
			},
		},
	}
}

func delete(busDomain dbtest.BusDomain, sd unitest.SeedData) []unitest.Table {
	table := []unitest.Table{
		{
			Name:    "delete",
			ExpResp: nil,
			ExcFunc: func(ctx context.Context) any {
				err := busDomain.Order.Delete(ctx, sd.Orders[0])
				return err
			},
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}

	return table
}
