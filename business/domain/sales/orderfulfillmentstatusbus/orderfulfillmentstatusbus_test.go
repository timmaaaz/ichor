package orderfulfillmentstatusbus_test

import (
	"context"
	"fmt"
	"sort"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/timmaaaz/ichor/business/domain/core/userbus"
	"github.com/timmaaaz/ichor/business/domain/sales/orderfulfillmentstatusbus"
	"github.com/timmaaaz/ichor/business/sdk/dbtest"
	"github.com/timmaaaz/ichor/business/sdk/page"
	"github.com/timmaaaz/ichor/business/sdk/unitest"
)

func Test_OrderFulfillmentStatus(t *testing.T) {
	t.Parallel()

	db := dbtest.NewDatabase(t, "Test_OrderFulfillmentStatus")

	sd, err := insertSeedData(db.BusDomain)
	if err != nil {
		t.Fatalf("Seeding error: %s", err)
	}

	unitest.Run(t, query(db.BusDomain, sd), "query")
	unitest.Run(t, create(db.BusDomain), "create")
	unitest.Run(t, update(db.BusDomain, sd), "update")
	unitest.Run(t, delete(db.BusDomain, sd), "delete")
}

func insertSeedData(busDomain dbtest.BusDomain) (unitest.SeedData, error) {
	ctx := context.Background()

	admins, err := userbus.TestSeedUsersWithNoFKs(ctx, 1, userbus.Roles.Admin, busDomain.User)
	if err != nil {
		return unitest.SeedData{}, fmt.Errorf("seeding user : %w", err)
	}

	ofls, err := orderfulfillmentstatusbus.TestSeedOrderFulfillmentStatuses(ctx, busDomain.OrderFulfillmentStatus)
	if err != nil {
		return unitest.SeedData{}, fmt.Errorf("seeding order fulfillment statuses: %w", err)
	}

	return unitest.SeedData{
		Admins:                   []unitest.User{{User: admins[0]}},
		OrderFulfillmentStatuses: ofls,
	}, nil
}

func query(busDomain dbtest.BusDomain, sd unitest.SeedData) []unitest.Table {
	return []unitest.Table{
		{
			Name: "Query",
			ExpResp: []orderfulfillmentstatusbus.OrderFulfillmentStatus{
				sd.OrderFulfillmentStatuses[0],
				sd.OrderFulfillmentStatuses[1],
				sd.OrderFulfillmentStatuses[2],
				sd.OrderFulfillmentStatuses[3],
				sd.OrderFulfillmentStatuses[4],
			},
			ExcFunc: func(ctx context.Context) any {
				got, err := busDomain.OrderFulfillmentStatus.Query(ctx, orderfulfillmentstatusbus.QueryFilter{}, orderfulfillmentstatusbus.DefaultOrderBy, page.MustParse("1", "5"))
				if err != nil {
					return err
				}
				return got
			},
			CmpFunc: func(got, exp any) string {
				gotResp, exists := got.([]orderfulfillmentstatusbus.OrderFulfillmentStatus)
				sort.Slice(gotResp, func(i, j int) bool {
					return gotResp[i].Name < gotResp[j].Name
				})

				if !exists {
					return fmt.Sprintf("expected []orderfulfillmentstatusbus.OrderFulfillmentStatus, got %T", got)
				}

				expResp, exists := exp.([]orderfulfillmentstatusbus.OrderFulfillmentStatus)
				if !exists {
					return fmt.Sprintf("expected []orderfulfillmentstatusbus.OrderFulfillmentStatus, got %T", exp)
				}
				sort.Slice(expResp, func(i, j int) bool {
					return expResp[i].Name < expResp[j].Name
				})

				return cmp.Diff(gotResp, expResp)
			},
		},
	}
}

func create(busDomain dbtest.BusDomain) []unitest.Table {
	return []unitest.Table{
		{
			Name: "Create",
			ExpResp: orderfulfillmentstatusbus.OrderFulfillmentStatus{
				Name:        "TEST STATUS",
				Description: "Description for TEST STATUS",
			},
			ExcFunc: func(ctx context.Context) any {
				ofs, err := busDomain.OrderFulfillmentStatus.Create(ctx, orderfulfillmentstatusbus.NewOrderFulfillmentStatus{
					Name:        "TEST STATUS",
					Description: "Description for TEST STATUS",
				})
				if err != nil {
					return err
				}
				return ofs
			},
			CmpFunc: func(got, exp any) string {
				gotResp, exists := got.(orderfulfillmentstatusbus.OrderFulfillmentStatus)
				if !exists {
					return fmt.Sprintf("expected orderfulfillmentstatusbus.OrderFulfillmentStatus, got %T", got)
				}

				expResp, exists := exp.(orderfulfillmentstatusbus.OrderFulfillmentStatus)
				if !exists {
					return fmt.Sprintf("expected orderfulfillmentstatusbus.OrderFulfillmentStatus, got %T", exp)
				}

				expResp.ID = gotResp.ID // Ignore ID for comparison

				return cmp.Diff(gotResp, expResp)
			},
		},
	}
}

func update(busDomain dbtest.BusDomain, sd unitest.SeedData) []unitest.Table {
	return []unitest.Table{
		{
			Name: "Update",
			ExpResp: orderfulfillmentstatusbus.OrderFulfillmentStatus{
				ID:          sd.OrderFulfillmentStatuses[0].ID,
				Name:        "UPDATED STATUS",
				Description: "UPDATED STATUS DESCRIPTION",
			},
			ExcFunc: func(ctx context.Context) any {
				got, err := busDomain.OrderFulfillmentStatus.Update(ctx, sd.OrderFulfillmentStatuses[0], orderfulfillmentstatusbus.UpdateOrderFulfillmentStatus{
					Name:        dbtest.StringPointer("UPDATED STATUS"),
					Description: dbtest.StringPointer("UPDATED STATUS DESCRIPTION"),
				})
				if err != nil {
					return err
				}
				return got
			},
			CmpFunc: func(got, exp any) string {
				gotResp, exists := got.(orderfulfillmentstatusbus.OrderFulfillmentStatus)
				if !exists {
					return fmt.Sprintf("expected orderfulfillmentstatusbus.OrderFulfillmentStatus, got %T", got)
				}

				expResp, exists := exp.(orderfulfillmentstatusbus.OrderFulfillmentStatus)
				if !exists {
					return fmt.Sprintf("expected orderfulfillmentstatusbus.OrderFulfillmentStatus, got %T", exp)
				}

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
				err := busDomain.OrderFulfillmentStatus.Delete(ctx, sd.OrderFulfillmentStatuses[0])
				return err
			},
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}

	return table
}
