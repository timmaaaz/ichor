package purchaseorderstatusbus_test

import (
	"context"
	"fmt"
	"sort"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/timmaaaz/ichor/business/domain/procurement/purchaseorderstatusbus"
	"github.com/timmaaaz/ichor/business/domain/procurement/purchaseorderstatusbus/stores/purchaseorderstatusdb"
	"github.com/timmaaaz/ichor/business/sdk/dbtest"
	"github.com/timmaaaz/ichor/business/sdk/delegate"
	"github.com/timmaaaz/ichor/business/sdk/order"
	"github.com/timmaaaz/ichor/business/sdk/page"
	"github.com/timmaaaz/ichor/business/sdk/unitest"
)

func Test_PurchaseOrderStatus(t *testing.T) {
	t.Parallel()

	db := dbtest.NewDatabase(t, "Test_PurchaseOrderStatus")

	// Create business layer for testing
	del := delegate.New(db.Log)
	bus := purchaseorderstatusbus.NewBusiness(db.Log, del, purchaseorderstatusdb.NewStore(db.Log, db.DB))

	sd, err := insertSeedData(bus)
	if err != nil {
		t.Fatalf("Seeding error: %s", err)
	}

	unitest.Run(t, query(bus, sd), "query")
	unitest.Run(t, create(bus), "create")
	unitest.Run(t, update(bus, sd), "update")
	unitest.Run(t, delete(bus, sd), "delete")
}

func insertSeedData(bus *purchaseorderstatusbus.Business) (unitest.SeedData, error) {
	ctx := context.Background()

	statuses, err := purchaseorderstatusbus.TestSeedPurchaseOrderStatuses(ctx, 3, bus)
	if err != nil {
		return unitest.SeedData{}, fmt.Errorf("seeding purchase order statuses: %w", err)
	}

	return unitest.SeedData{
		PurchaseOrderStatuses: statuses,
	}, nil
}

func query(bus *purchaseorderstatusbus.Business, sd unitest.SeedData) []unitest.Table {
	// make sorted copy for use
	exp := make([]purchaseorderstatusbus.PurchaseOrderStatus, len(sd.PurchaseOrderStatuses))
	copy(exp, sd.PurchaseOrderStatuses)
	sort.Slice(exp, func(i, j int) bool {
		return exp[i].SortOrder < exp[j].SortOrder
	})

	return []unitest.Table{
		{
			Name:    "Query",
			ExpResp: exp,
			ExcFunc: func(ctx context.Context) any {
				got, err := bus.Query(ctx, purchaseorderstatusbus.QueryFilter{}, order.NewBy(purchaseorderstatusbus.OrderBySortOrder, order.ASC), page.MustParse("1", "3"))
				if err != nil {
					return err
				}
				return got
			},
			CmpFunc: func(got, exp any) string {
				gotResp, exists := got.([]purchaseorderstatusbus.PurchaseOrderStatus)
				if !exists {
					return "error occurred"
				}

				expResp, exists := exp.([]purchaseorderstatusbus.PurchaseOrderStatus)
				if !exists {
					return "error occurred"
				}

				for i := range gotResp {
					expResp[i].ID = gotResp[i].ID
				}

				return cmp.Diff(exp, got)
			},
		},
	}
}

func create(bus *purchaseorderstatusbus.Business) []unitest.Table {
	return []unitest.Table{
		{
			Name: "Create",
			ExpResp: purchaseorderstatusbus.PurchaseOrderStatus{
				Name:        "TestStatus",
				Description: "TestStatus Description",
				SortOrder:   500,
			},
			ExcFunc: func(ctx context.Context) any {
				resp, err := bus.Create(ctx, purchaseorderstatusbus.NewPurchaseOrderStatus{
					Name:        "TestStatus",
					Description: "TestStatus Description",
					SortOrder:   500,
				})
				if err != nil {
					return err
				}
				return resp
			},
			CmpFunc: func(got, exp any) string {
				gotResp, exists := got.(purchaseorderstatusbus.PurchaseOrderStatus)
				if !exists {
					return "error occurred"
				}

				expResp, exists := exp.(purchaseorderstatusbus.PurchaseOrderStatus)
				if !exists {
					return "error occurred"
				}

				expResp.ID = gotResp.ID

				return cmp.Diff(expResp, gotResp)
			},
		},
	}
}

func update(bus *purchaseorderstatusbus.Business, sd unitest.SeedData) []unitest.Table {
	return []unitest.Table{
		{
			Name: "Update",
			ExpResp: purchaseorderstatusbus.PurchaseOrderStatus{
				ID:          sd.PurchaseOrderStatuses[0].ID,
				Name:        "UpdatedStatus",
				Description: "UpdatedStatus Description",
				SortOrder:   999,
			},
			ExcFunc: func(ctx context.Context) any {
				resp, err := bus.Update(ctx, sd.PurchaseOrderStatuses[0], purchaseorderstatusbus.UpdatePurchaseOrderStatus{
					Name:        dbtest.StringPointer("UpdatedStatus"),
					Description: dbtest.StringPointer("UpdatedStatus Description"),
					SortOrder:   dbtest.IntPointer(999),
				})
				if err != nil {
					return err
				}
				return resp
			},
			CmpFunc: func(got, exp any) string {
				gotResp, exists := got.(purchaseorderstatusbus.PurchaseOrderStatus)
				if !exists {
					return "error occurred"
				}

				expResp, exists := exp.(purchaseorderstatusbus.PurchaseOrderStatus)
				if !exists {
					return "error occurred"
				}

				return cmp.Diff(expResp, gotResp)
			},
		},
	}
}

func delete(bus *purchaseorderstatusbus.Business, sd unitest.SeedData) []unitest.Table {
	return []unitest.Table{
		{
			Name:    "Delete",
			ExpResp: nil,
			ExcFunc: func(ctx context.Context) any {
				return bus.Delete(ctx, sd.PurchaseOrderStatuses[0])
			},
			CmpFunc: func(got, exp any) string {
				return ""
			},
		},
	}
}