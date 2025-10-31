package purchaseorderlineitemstatusbus_test

import (
	"context"
	"fmt"
	"sort"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/timmaaaz/ichor/business/domain/procurement/purchaseorderlineitemstatusbus"
	"github.com/timmaaaz/ichor/business/domain/procurement/purchaseorderlineitemstatusbus/stores/purchaseorderlineitemstatusdb"
	"github.com/timmaaaz/ichor/business/sdk/dbtest"
	"github.com/timmaaaz/ichor/business/sdk/delegate"
	"github.com/timmaaaz/ichor/business/sdk/order"
	"github.com/timmaaaz/ichor/business/sdk/page"
	"github.com/timmaaaz/ichor/business/sdk/unitest"
)

func Test_PurchaseOrderLineItemStatus(t *testing.T) {
	t.Parallel()

	db := dbtest.NewDatabase(t, "Test_PurchaseOrderLineItemStatus")

	// Create business layer for testing
	del := delegate.New(db.Log)
	bus := purchaseorderlineitemstatusbus.NewBusiness(db.Log, del, purchaseorderlineitemstatusdb.NewStore(db.Log, db.DB))

	sd, err := insertSeedData(bus)
	if err != nil {
		t.Fatalf("Seeding error: %s", err)
	}

	unitest.Run(t, query(bus, sd), "query")
	unitest.Run(t, create(bus), "create")
	unitest.Run(t, update(bus, sd), "update")
	unitest.Run(t, delete(bus, sd), "delete")
}

func insertSeedData(bus *purchaseorderlineitemstatusbus.Business) (unitest.SeedData, error) {
	ctx := context.Background()

	statuses, err := purchaseorderlineitemstatusbus.TestSeedPurchaseOrderLineItemStatuses(ctx, 3, bus)
	if err != nil {
		return unitest.SeedData{}, fmt.Errorf("seeding purchase order line item statuses: %w", err)
	}

	return unitest.SeedData{
		PurchaseOrderLineItemStatuses: statuses,
	}, nil
}

func query(bus *purchaseorderlineitemstatusbus.Business, sd unitest.SeedData) []unitest.Table {
	// make sorted copy for use
	exp := make([]purchaseorderlineitemstatusbus.PurchaseOrderLineItemStatus, len(sd.PurchaseOrderLineItemStatuses))
	copy(exp, sd.PurchaseOrderLineItemStatuses)
	sort.Slice(exp, func(i, j int) bool {
		return exp[i].SortOrder < exp[j].SortOrder
	})

	return []unitest.Table{
		{
			Name:    "Query",
			ExpResp: exp,
			ExcFunc: func(ctx context.Context) any {
				got, err := bus.Query(ctx, purchaseorderlineitemstatusbus.QueryFilter{}, order.NewBy(purchaseorderlineitemstatusbus.OrderBySortOrder, order.ASC), page.MustParse("1", "3"))
				if err != nil {
					return err
				}
				return got
			},
			CmpFunc: func(got, exp any) string {
				gotResp, exists := got.([]purchaseorderlineitemstatusbus.PurchaseOrderLineItemStatus)
				if !exists {
					return "error occurred"
				}

				expResp, exists := exp.([]purchaseorderlineitemstatusbus.PurchaseOrderLineItemStatus)
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

func create(bus *purchaseorderlineitemstatusbus.Business) []unitest.Table {
	return []unitest.Table{
		{
			Name: "Create",
			ExpResp: purchaseorderlineitemstatusbus.PurchaseOrderLineItemStatus{
				Name:        "TestStatus",
				Description: "TestStatus Description",
				SortOrder:   500,
			},
			ExcFunc: func(ctx context.Context) any {
				resp, err := bus.Create(ctx, purchaseorderlineitemstatusbus.NewPurchaseOrderLineItemStatus{
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
				gotResp, exists := got.(purchaseorderlineitemstatusbus.PurchaseOrderLineItemStatus)
				if !exists {
					return "error occurred"
				}

				expResp, exists := exp.(purchaseorderlineitemstatusbus.PurchaseOrderLineItemStatus)
				if !exists {
					return "error occurred"
				}

				expResp.ID = gotResp.ID

				return cmp.Diff(expResp, gotResp)
			},
		},
	}
}

func update(bus *purchaseorderlineitemstatusbus.Business, sd unitest.SeedData) []unitest.Table {
	return []unitest.Table{
		{
			Name: "Update",
			ExpResp: purchaseorderlineitemstatusbus.PurchaseOrderLineItemStatus{
				ID:          sd.PurchaseOrderLineItemStatuses[0].ID,
				Name:        "UpdatedStatus",
				Description: "UpdatedStatus Description",
				SortOrder:   999,
			},
			ExcFunc: func(ctx context.Context) any {
				resp, err := bus.Update(ctx, sd.PurchaseOrderLineItemStatuses[0], purchaseorderlineitemstatusbus.UpdatePurchaseOrderLineItemStatus{
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
				gotResp, exists := got.(purchaseorderlineitemstatusbus.PurchaseOrderLineItemStatus)
				if !exists {
					return "error occurred"
				}

				expResp, exists := exp.(purchaseorderlineitemstatusbus.PurchaseOrderLineItemStatus)
				if !exists {
					return "error occurred"
				}

				return cmp.Diff(expResp, gotResp)
			},
		},
	}
}

func delete(bus *purchaseorderlineitemstatusbus.Business, sd unitest.SeedData) []unitest.Table {
	return []unitest.Table{
		{
			Name:    "Delete",
			ExpResp: nil,
			ExcFunc: func(ctx context.Context) any {
				return bus.Delete(ctx, sd.PurchaseOrderLineItemStatuses[0])
			},
			CmpFunc: func(got, exp any) string {
				return ""
			},
		},
	}
}
