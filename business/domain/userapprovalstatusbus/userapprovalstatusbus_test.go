package userapprovalstatusbus_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/timmaaaz/ichor/business/domain/userapprovalstatusbus"
	"github.com/timmaaaz/ichor/business/sdk/dbtest"
	"github.com/timmaaaz/ichor/business/sdk/order"
	"github.com/timmaaaz/ichor/business/sdk/page"
	"github.com/timmaaaz/ichor/business/sdk/unitest"
)

func Test_UserApprovalStatus(t *testing.T) {
	t.Parallel()

	db := dbtest.NewDatabase(t, "Test_UserApprovalStatus")

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

	as, err := userapprovalstatusbus.TestSeedUserApprovalStatus(ctx, 10, busDomain.UserApprovalStatus)
	if err != nil {
		return unitest.SeedData{}, fmt.Errorf("seeding approval statues : %w", err)
	}

	return unitest.SeedData{
		UserApprovalStatus: as,
	}, nil
}

func query(busdomain dbtest.BusDomain, sd unitest.SeedData) []unitest.Table {
	table := []unitest.Table{
		{
			Name: "query",
			ExpResp: []userapprovalstatusbus.UserApprovalStatus{
				{ID: sd.UserApprovalStatus[0].ID, Name: sd.UserApprovalStatus[0].Name, IconID: sd.UserApprovalStatus[0].IconID},
				{ID: sd.UserApprovalStatus[1].ID, Name: sd.UserApprovalStatus[1].Name, IconID: sd.UserApprovalStatus[1].IconID},
				{ID: sd.UserApprovalStatus[2].ID, Name: sd.UserApprovalStatus[2].Name, IconID: sd.UserApprovalStatus[2].IconID},
				{ID: sd.UserApprovalStatus[3].ID, Name: sd.UserApprovalStatus[3].Name, IconID: sd.UserApprovalStatus[3].IconID},
				{ID: sd.UserApprovalStatus[4].ID, Name: sd.UserApprovalStatus[4].Name, IconID: sd.UserApprovalStatus[4].IconID},
			},
			ExcFunc: func(ctx context.Context) any {
				aprvlStatuses, err := busdomain.UserApprovalStatus.Query(ctx, userapprovalstatusbus.QueryFilter{}, order.NewBy(userapprovalstatusbus.OrderByName, order.DESC), page.MustParse("1", "5"))
				if err != nil {
					return err
				}
				return aprvlStatuses
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
			ExpResp: userapprovalstatusbus.UserApprovalStatus{
				IconID: sd.UserApprovalStatus[0].IconID,
				Name:   "Test Approval Status",
			},
			ExcFunc: func(ctx context.Context) any {
				aprvlStatus, err := busDomain.UserApprovalStatus.Create(ctx, userapprovalstatusbus.NewUserApprovalStatus{
					Name:   "Test Approval Status",
					IconID: sd.UserApprovalStatus[0].IconID,
				})
				if err != nil {
					return err
				}
				return aprvlStatus
			},
			CmpFunc: func(got any, exp any) string {
				gotResp, exists := got.(userapprovalstatusbus.UserApprovalStatus)
				if !exists {
					return fmt.Sprintf("got is not an approval status %v", got)
				}

				expResp := exp.(userapprovalstatusbus.UserApprovalStatus)
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
			ExpResp: userapprovalstatusbus.UserApprovalStatus{
				ID:     sd.UserApprovalStatus[0].ID,
				IconID: sd.UserApprovalStatus[1].IconID,
				Name:   "Updated Approval Status",
			},
			ExcFunc: func(ctx context.Context) any {
				aprvlStatus, err := busDomain.UserApprovalStatus.Update(ctx, sd.UserApprovalStatus[0], userapprovalstatusbus.UpdateUserApprovalStatus{
					Name:   dbtest.StringPointer("Updated Approval Status"),
					IconID: &sd.UserApprovalStatus[1].IconID,
				})
				if err != nil {
					return err
				}
				return aprvlStatus
			},
			CmpFunc: func(got any, exp any) string {
				gotResp, exists := got.(userapprovalstatusbus.UserApprovalStatus)
				if !exists {
					return fmt.Sprintf("got is not an approval status %v", got)
				}
				expResp := exp.(userapprovalstatusbus.UserApprovalStatus)
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
				err := busDomain.UserApprovalStatus.Delete(ctx, sd.UserApprovalStatus[0])
				return err
			},
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}

	return table
}
