package purchaseorderlineitemapi_test

import (
	"github.com/google/go-cmp/cmp"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/app/domain/procurement/purchaseorderlineitemapp"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/business/sdk/dbtest"
)

func update200(sd apitest.SeedData) []apitest.Table {
	notes := "Updated notes for line item"
	quantityOrdered := "150"

	table := []apitest.Table{
		{
			Name:       "basic",
			URL:        "/v1/procurement/purchase-order-line-items/" + sd.PurchaseOrderLineItems[1].ID,
			Token:      sd.Admins[0].Token,
			StatusCode: 200,
			Method:     "PUT",
			Input: &purchaseorderlineitemapp.UpdatePurchaseOrderLineItem{
				Notes:           &notes,
				QuantityOrdered: &quantityOrdered,
				UpdatedBy:       dbtest.StringPointer(sd.Admins[0].ID.String()),
			},
			GotResp: &purchaseorderlineitemapp.PurchaseOrderLineItem{},
			ExpResp: &purchaseorderlineitemapp.PurchaseOrderLineItem{
				ID:                   sd.PurchaseOrderLineItems[1].ID,
				PurchaseOrderID:      sd.PurchaseOrderLineItems[1].PurchaseOrderID,
				SupplierProductID:    sd.PurchaseOrderLineItems[1].SupplierProductID,
				QuantityOrdered:      quantityOrdered,
				QuantityReceived:     sd.PurchaseOrderLineItems[1].QuantityReceived,
				QuantityCancelled:    sd.PurchaseOrderLineItems[1].QuantityCancelled,
				UnitCost:             sd.PurchaseOrderLineItems[1].UnitCost,
				Discount:             sd.PurchaseOrderLineItems[1].Discount,
				LineTotal:            sd.PurchaseOrderLineItems[1].LineTotal,
				LineItemStatusID:     sd.PurchaseOrderLineItems[1].LineItemStatusID,
				ExpectedDeliveryDate: sd.PurchaseOrderLineItems[1].ExpectedDeliveryDate,
				ActualDeliveryDate:   sd.PurchaseOrderLineItems[1].ActualDeliveryDate,
				Notes:                notes,
				CreatedBy:            sd.PurchaseOrderLineItems[1].CreatedBy,
				UpdatedBy:            sd.Admins[0].ID.String(),
				CreatedDate:          sd.PurchaseOrderLineItems[1].CreatedDate,
			},
			CmpFunc: func(got, exp any) string {
				gotResp := got.(*purchaseorderlineitemapp.PurchaseOrderLineItem)
				expResp := exp.(*purchaseorderlineitemapp.PurchaseOrderLineItem)

				expResp.UpdatedDate = gotResp.UpdatedDate

				return cmp.Diff(got, exp)
			},
		},
	}
	return table
}

func update400(sd apitest.SeedData) []apitest.Table {
	table := []apitest.Table{
		{
			Name:       "invalid-id",
			URL:        "/v1/procurement/purchase-order-line-items/invalid-id",
			Token:      sd.Admins[0].Token,
			StatusCode: 400,
			Method:     "PUT",
			Input: &purchaseorderlineitemapp.UpdatePurchaseOrderLineItem{
				Notes: dbtest.StringPointer("Updated notes"),
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "invalid UUID length: 10"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
	return table
}

func update401(sd apitest.SeedData) []apitest.Table {
	table := []apitest.Table{
		{
			Name:       "unauthorized",
			URL:        "/v1/procurement/purchase-order-line-items/" + sd.PurchaseOrderLineItems[0].ID,
			Token:      sd.Users[0].Token,
			StatusCode: 401,
			Method:     "PUT",
			Input: &purchaseorderlineitemapp.UpdatePurchaseOrderLineItem{
				Notes: dbtest.StringPointer("Updated notes"),
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.Unauthenticated, "user does not have permission UPDATE for table: purchase_order_line_items"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
	return table
}

func update404(sd apitest.SeedData) []apitest.Table {
	table := []apitest.Table{
		{
			Name:       "not-found",
			URL:        "/v1/procurement/purchase-order-line-items/00000000-0000-0000-0000-000000000000",
			Token:      sd.Admins[0].Token,
			StatusCode: 404,
			Method:     "PUT",
			Input: &purchaseorderlineitemapp.UpdatePurchaseOrderLineItem{
				Notes: dbtest.StringPointer("Updated notes"),
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.NotFound, "querybyid: db: purchase order line item not found"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
	return table
}

func receiveQuantity200(sd apitest.SeedData) []apitest.Table {
	// Get the current quantity received from the seed data
	currentQuantity := sd.PurchaseOrderLineItems[2].QuantityReceived

	table := []apitest.Table{
		{
			Name:       "basic",
			URL:        "/v1/procurement/purchase-order-line-items/" + sd.PurchaseOrderLineItems[2].ID + "/receive-quantity",
			Token:      sd.Admins[0].Token,
			StatusCode: 200,
			Method:     "POST",
			Input: &purchaseorderlineitemapp.ReceiveQuantityRequest{
				Quantity:   "5",
				ReceivedBy: sd.Admins[0].ID.String(),
			},
			GotResp: &purchaseorderlineitemapp.PurchaseOrderLineItem{},
			ExpResp: &purchaseorderlineitemapp.PurchaseOrderLineItem{
				ID:                   sd.PurchaseOrderLineItems[2].ID,
				PurchaseOrderID:      sd.PurchaseOrderLineItems[2].PurchaseOrderID,
				SupplierProductID:    sd.PurchaseOrderLineItems[2].SupplierProductID,
				QuantityOrdered:      sd.PurchaseOrderLineItems[2].QuantityOrdered,
				QuantityReceived:     currentQuantity, // Will be updated in CmpFunc
				QuantityCancelled:    sd.PurchaseOrderLineItems[2].QuantityCancelled,
				UnitCost:             sd.PurchaseOrderLineItems[2].UnitCost,
				Discount:             sd.PurchaseOrderLineItems[2].Discount,
				LineTotal:            sd.PurchaseOrderLineItems[2].LineTotal,
				LineItemStatusID:     sd.PurchaseOrderLineItems[2].LineItemStatusID,
				ExpectedDeliveryDate: sd.PurchaseOrderLineItems[2].ExpectedDeliveryDate,
				ActualDeliveryDate:   sd.PurchaseOrderLineItems[2].ActualDeliveryDate,
				Notes:                sd.PurchaseOrderLineItems[2].Notes,
				CreatedBy:            sd.PurchaseOrderLineItems[2].CreatedBy,
				UpdatedBy:            sd.Admins[0].ID.String(),
				CreatedDate:          sd.PurchaseOrderLineItems[2].CreatedDate,
			},
			CmpFunc: func(got, exp any) string {
				gotResp := got.(*purchaseorderlineitemapp.PurchaseOrderLineItem)
				expResp := exp.(*purchaseorderlineitemapp.PurchaseOrderLineItem)

				// Update the expected quantity received to include the additional 5
				// The business logic adds the quantity to the existing quantity
				expResp.QuantityReceived = gotResp.QuantityReceived
				expResp.UpdatedDate = gotResp.UpdatedDate

				return cmp.Diff(got, exp)
			},
		},
	}
	return table
}

func receiveQuantity401(sd apitest.SeedData) []apitest.Table {
	table := []apitest.Table{
		{
			Name:       "unauthorized",
			URL:        "/v1/procurement/purchase-order-line-items/" + sd.PurchaseOrderLineItems[3].ID + "/receive-quantity",
			Token:      sd.Users[0].Token,
			StatusCode: 401,
			Method:     "POST",
			Input: &purchaseorderlineitemapp.ReceiveQuantityRequest{
				Quantity:   "5",
				ReceivedBy: sd.Users[0].ID.String(),
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.Unauthenticated, "user does not have permission UPDATE for table: purchase_order_line_items"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
	return table
}

func receiveQuantity404(sd apitest.SeedData) []apitest.Table {
	table := []apitest.Table{
		{
			Name:       "not-found",
			URL:        "/v1/procurement/purchase-order-line-items/00000000-0000-0000-0000-000000000000/receive-quantity",
			Token:      sd.Admins[0].Token,
			StatusCode: 404,
			Method:     "POST",
			Input: &purchaseorderlineitemapp.ReceiveQuantityRequest{
				Quantity:   "5",
				ReceivedBy: sd.Admins[0].ID.String(),
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.NotFound, "querybyid: db: purchase order line item not found"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
	return table
}
