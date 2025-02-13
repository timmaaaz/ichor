package fulfillmentstatusbus_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/timmaaaz/ichor/business/domain/assets/fulfillmentstatusbus"
	"github.com/timmaaaz/ichor/business/sdk/dbtest"
	"github.com/timmaaaz/ichor/business/sdk/order"
	"github.com/timmaaaz/ichor/business/sdk/page"
	"github.com/timmaaaz/ichor/business/sdk/unitest"
)

func Test_FulfillmentStatus(t *testing.T) {
	t.Parallel()

	db := dbtest.NewDatabase(t, "Test_FulfillmentStatus")

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

// =============================================================================

func insertSeedData(busDomain dbtest.BusDomain) (unitest.SeedData, error) {
	ctx := context.Background()

	as, err := fulfillmentstatusbus.TestSeedFulfillmentStatus(ctx, 10, busDomain.FulfillmentStatus)
	if err != nil {
		return unitest.SeedData{}, fmt.Errorf("seeding fulfillment statues : %w", err)
	}

	return unitest.SeedData{
		FulfillmentStatus: as,
	}, nil
}

func query(busdomain dbtest.BusDomain, sd unitest.SeedData) []unitest.Table {
	table := []unitest.Table{
		{
			Name: "query",
			ExpResp: []fulfillmentstatusbus.FulfillmentStatus{
				{ID: sd.FulfillmentStatus[0].ID, Name: sd.FulfillmentStatus[0].Name, IconID: sd.FulfillmentStatus[0].IconID},
				{ID: sd.FulfillmentStatus[1].ID, Name: sd.FulfillmentStatus[1].Name, IconID: sd.FulfillmentStatus[1].IconID},
				{ID: sd.FulfillmentStatus[2].ID, Name: sd.FulfillmentStatus[2].Name, IconID: sd.FulfillmentStatus[2].IconID},
				{ID: sd.FulfillmentStatus[3].ID, Name: sd.FulfillmentStatus[3].Name, IconID: sd.FulfillmentStatus[3].IconID},
				{ID: sd.FulfillmentStatus[4].ID, Name: sd.FulfillmentStatus[4].Name, IconID: sd.FulfillmentStatus[4].IconID},
			},
			ExcFunc: func(ctx context.Context) any {
				fulfillmentstatuses, err := busdomain.FulfillmentStatus.Query(ctx, fulfillmentstatusbus.QueryFilter{}, order.NewBy(fulfillmentstatusbus.OrderByName, order.ASC), page.MustParse("1", "5"))
				if err != nil {
					return err
				}
				return fulfillmentstatuses
			},
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}

	return table
}

func create(busDomain dbtest.BusDomain, sd unitest.SeedData) []unitest.Table {
	table := []unitest.Table{
		{
			Name: "create",
			ExpResp: fulfillmentstatusbus.FulfillmentStatus{
				IconID: sd.FulfillmentStatus[0].IconID,
				Name:   "Test Approval Status",
			},
			ExcFunc: func(ctx context.Context) any {
				aprvlStatus, err := busDomain.FulfillmentStatus.Create(ctx, fulfillmentstatusbus.NewFulfillmentStatus{
					Name:   "Test Approval Status",
					IconID: sd.FulfillmentStatus[0].IconID,
				})
				if err != nil {
					return err
				}
				return aprvlStatus
			},
			CmpFunc: func(got any, exp any) string {
				gotResp, exists := got.(fulfillmentstatusbus.FulfillmentStatus)
				if !exists {
					return fmt.Sprintf("got is not an approval status %v", got)
				}

				expResp := exp.(fulfillmentstatusbus.FulfillmentStatus)
				expResp.ID = gotResp.ID

				return cmp.Diff(gotResp, expResp)
			},
		},
	}

	return table
}

func update(busDomain dbtest.BusDomain, sd unitest.SeedData) []unitest.Table {
	table := []unitest.Table{
		{
			Name: "update",
			ExpResp: fulfillmentstatusbus.FulfillmentStatus{
				ID:     sd.FulfillmentStatus[0].ID,
				IconID: sd.FulfillmentStatus[1].IconID,
				Name:   "Updated Approval Status",
			},
			ExcFunc: func(ctx context.Context) any {
				aprvlStatus, err := busDomain.FulfillmentStatus.Update(ctx, sd.FulfillmentStatus[0], fulfillmentstatusbus.UpdateFulfillmentStatus{
					Name:   dbtest.StringPointer("Updated Approval Status"),
					IconID: &sd.FulfillmentStatus[1].IconID,
				})
				if err != nil {
					return err
				}
				return aprvlStatus
			},
			CmpFunc: func(got any, exp any) string {
				gotResp, exists := got.(fulfillmentstatusbus.FulfillmentStatus)
				if !exists {
					return fmt.Sprintf("got is not an approval status %v", got)
				}
				expResp := exp.(fulfillmentstatusbus.FulfillmentStatus)
				return cmp.Diff(gotResp, expResp)
			},
		},
	}

	return table
}

func delete(busDomain dbtest.BusDomain, sd unitest.SeedData) []unitest.Table {
	table := []unitest.Table{
		{
			Name:    "delete",
			ExpResp: nil,
			ExcFunc: func(ctx context.Context) any {
				err := busDomain.FulfillmentStatus.Delete(ctx, sd.FulfillmentStatus[0])
				return err
			},
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}

	return table
}
