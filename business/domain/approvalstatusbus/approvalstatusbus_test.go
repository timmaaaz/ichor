package approvalstatusbus_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/timmaaaz/ichor/business/domain/approvalstatusbus"
	"github.com/timmaaaz/ichor/business/sdk/dbtest"
	"github.com/timmaaaz/ichor/business/sdk/order"
	"github.com/timmaaaz/ichor/business/sdk/page"
	"github.com/timmaaaz/ichor/business/sdk/unitest"
)

func Test_ApprovalStatus(t *testing.T) {
	t.Parallel()

	db := dbtest.NewDatabase(t, "Test_ApprovalStatus")

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

	as, err := approvalstatusbus.TestSeedApprovalStatus(ctx, 10, busDomain.ApprovalStatus)
	if err != nil {
		return unitest.SeedData{}, fmt.Errorf("seeding approval statues : %w", err)
	}

	return unitest.SeedData{
		ApprovalStatus: as,
	}, nil
}

func query(busdomain dbtest.BusDomain, sd unitest.SeedData) []unitest.Table {
	table := []unitest.Table{
		{
			Name: "query",
			ExpResp: []approvalstatusbus.ApprovalStatus{
				{ID: sd.ApprovalStatus[0].ID, Name: sd.ApprovalStatus[0].Name, IconID: sd.ApprovalStatus[0].IconID},
				{ID: sd.ApprovalStatus[1].ID, Name: sd.ApprovalStatus[1].Name, IconID: sd.ApprovalStatus[1].IconID},
				{ID: sd.ApprovalStatus[2].ID, Name: sd.ApprovalStatus[2].Name, IconID: sd.ApprovalStatus[2].IconID},
				{ID: sd.ApprovalStatus[3].ID, Name: sd.ApprovalStatus[3].Name, IconID: sd.ApprovalStatus[3].IconID},
				{ID: sd.ApprovalStatus[4].ID, Name: sd.ApprovalStatus[4].Name, IconID: sd.ApprovalStatus[4].IconID},
			},
			ExcFunc: func(ctx context.Context) any {
				aprvlStatuses, err := busdomain.ApprovalStatus.Query(ctx, approvalstatusbus.QueryFilter{}, order.NewBy(approvalstatusbus.OrderByName, order.ASC), page.MustParse("1", "5"))
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
			ExpResp: approvalstatusbus.ApprovalStatus{
				IconID: sd.ApprovalStatus[0].IconID,
				Name:   "Test Approval Status",
			},
			ExcFunc: func(ctx context.Context) any {
				aprvlStatus, err := busDomain.ApprovalStatus.Create(ctx, approvalstatusbus.NewApprovalStatus{
					Name:   "Test Approval Status",
					IconID: sd.ApprovalStatus[0].IconID,
				})
				if err != nil {
					return err
				}
				return aprvlStatus
			},
			CmpFunc: func(got any, exp any) string {
				gotResp, exists := got.(approvalstatusbus.ApprovalStatus)
				if !exists {
					return fmt.Sprintf("got is not an approval status %v", got)
				}

				expResp := exp.(approvalstatusbus.ApprovalStatus)
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
			ExpResp: approvalstatusbus.ApprovalStatus{
				ID:     sd.ApprovalStatus[0].ID,
				IconID: sd.ApprovalStatus[1].IconID,
				Name:   "Updated Approval Status",
			},
			ExcFunc: func(ctx context.Context) any {
				aprvlStatus, err := busDomain.ApprovalStatus.Update(ctx, sd.ApprovalStatus[0], approvalstatusbus.UpdateApprovalStatus{
					Name:   dbtest.StringPointer("Updated Approval Status"),
					IconID: &sd.ApprovalStatus[1].IconID,
				})
				if err != nil {
					return err
				}
				return aprvlStatus
			},
			CmpFunc: func(got any, exp any) string {
				gotResp, exists := got.(approvalstatusbus.ApprovalStatus)
				if !exists {
					return fmt.Sprintf("got is not an approval status %v", got)
				}
				expResp := exp.(approvalstatusbus.ApprovalStatus)
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
				err := busDomain.ApprovalStatus.Delete(ctx, sd.ApprovalStatus[0])
				return err
			},
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}

	return table
}
