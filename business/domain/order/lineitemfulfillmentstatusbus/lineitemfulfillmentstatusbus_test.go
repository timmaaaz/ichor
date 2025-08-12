package lineitemfulfillmentstatusbus_test

import (
	"context"
	"fmt"
	"sort"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/timmaaaz/ichor/business/domain/order/lineitemfulfillmentstatusbus"
	"github.com/timmaaaz/ichor/business/domain/users/userbus"
	"github.com/timmaaaz/ichor/business/sdk/dbtest"
	"github.com/timmaaaz/ichor/business/sdk/page"
	"github.com/timmaaaz/ichor/business/sdk/unitest"
)

func Test_LineItemFulfillmentStatus(t *testing.T) {
	t.Parallel()

	db := dbtest.NewDatabase(t, "Test_LineItemFulfillmentStatus")

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

	ofls, err := lineitemfulfillmentstatusbus.TestSeedLineItemFulfillmentStatuses(ctx, busDomain.LineItemFulfillmentStatus)
	if err != nil {
		return unitest.SeedData{}, fmt.Errorf("seeding order fulfillment statuses: %w", err)
	}

	return unitest.SeedData{
		Admins:                      []unitest.User{{User: admins[0]}},
		LineItemFulfillmentStatuses: ofls,
	}, nil
}

func query(busDomain dbtest.BusDomain, sd unitest.SeedData) []unitest.Table {
	return []unitest.Table{
		{
			Name: "Query",
			ExpResp: []lineitemfulfillmentstatusbus.LineItemFulfillmentStatus{
				sd.LineItemFulfillmentStatuses[0],
				sd.LineItemFulfillmentStatuses[1],
				sd.LineItemFulfillmentStatuses[2],
				sd.LineItemFulfillmentStatuses[3],
				sd.LineItemFulfillmentStatuses[4],
				sd.LineItemFulfillmentStatuses[5],
			},
			ExcFunc: func(ctx context.Context) any {
				got, err := busDomain.LineItemFulfillmentStatus.Query(ctx, lineitemfulfillmentstatusbus.QueryFilter{}, lineitemfulfillmentstatusbus.DefaultOrderBy, page.MustParse("1", "6"))
				if err != nil {
					return err
				}
				return got
			},
			CmpFunc: func(got, exp any) string {
				gotResp, exists := got.([]lineitemfulfillmentstatusbus.LineItemFulfillmentStatus)
				sort.Slice(gotResp, func(i, j int) bool {
					return gotResp[i].Name < gotResp[j].Name
				})

				if !exists {
					return fmt.Sprintf("expected []lineitemfulfillmentstatusbus.LineItemFulfillmentStatus, got %T", got)
				}

				expResp, exists := exp.([]lineitemfulfillmentstatusbus.LineItemFulfillmentStatus)
				if !exists {
					return fmt.Sprintf("expected []lineitemfulfillmentstatusbus.LineItemFulfillmentStatus, got %T", exp)
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
			ExpResp: lineitemfulfillmentstatusbus.LineItemFulfillmentStatus{
				Name:        "TEST STATUS",
				Description: "Description for TEST STATUS",
			},
			ExcFunc: func(ctx context.Context) any {
				ofs, err := busDomain.LineItemFulfillmentStatus.Create(ctx, lineitemfulfillmentstatusbus.NewLineItemFulfillmentStatus{
					Name:        "TEST STATUS",
					Description: "Description for TEST STATUS",
				})
				if err != nil {
					return err
				}
				return ofs
			},
			CmpFunc: func(got, exp any) string {
				gotResp, exists := got.(lineitemfulfillmentstatusbus.LineItemFulfillmentStatus)
				if !exists {
					return fmt.Sprintf("expected lineitemfulfillmentstatusbus.LineItemFulfillmentStatus, got %T", got)
				}

				expResp, exists := exp.(lineitemfulfillmentstatusbus.LineItemFulfillmentStatus)
				if !exists {
					return fmt.Sprintf("expected lineitemfulfillmentstatusbus.LineItemFulfillmentStatus, got %T", exp)
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
			ExpResp: lineitemfulfillmentstatusbus.LineItemFulfillmentStatus{
				ID:          sd.LineItemFulfillmentStatuses[0].ID,
				Name:        "UPDATED STATUS",
				Description: "UPDATED STATUS DESCRIPTION",
			},
			ExcFunc: func(ctx context.Context) any {
				got, err := busDomain.LineItemFulfillmentStatus.Update(ctx, sd.LineItemFulfillmentStatuses[0], lineitemfulfillmentstatusbus.UpdateLineItemFulfillmentStatus{
					Name:        dbtest.StringPointer("UPDATED STATUS"),
					Description: dbtest.StringPointer("UPDATED STATUS DESCRIPTION"),
				})
				if err != nil {
					return err
				}
				return got
			},
			CmpFunc: func(got, exp any) string {
				gotResp, exists := got.(lineitemfulfillmentstatusbus.LineItemFulfillmentStatus)
				if !exists {
					return fmt.Sprintf("expected lineitemfulfillmentstatusbus.LineItemFulfillmentStatus, got %T", got)
				}

				expResp, exists := exp.(lineitemfulfillmentstatusbus.LineItemFulfillmentStatus)
				if !exists {
					return fmt.Sprintf("expected lineitemfulfillmentstatusbus.LineItemFulfillmentStatus, got %T", exp)
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
				err := busDomain.LineItemFulfillmentStatus.Delete(ctx, sd.LineItemFulfillmentStatuses[0])
				return err
			},
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}

	return table
}
