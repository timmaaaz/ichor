package purchaseorderapi_test

import (
	"github.com/google/go-cmp/cmp"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/app/domain/procurement/purchaseorderapp"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/business/sdk/dbtest"
)

func update200(sd apitest.SeedData) []apitest.Table {
	notes := "Updated notes for purchase order"

	table := []apitest.Table{
		{
			Name:       "basic",
			URL:        "/v1/procurement/purchase-orders/" + sd.PurchaseOrders[1].ID,
			Token:      sd.Admins[0].Token,
			StatusCode: 200,
			Method:     "PUT",
			Input: &purchaseorderapp.UpdatePurchaseOrder{
				Notes:     &notes,
				UpdatedBy: dbtest.StringPointer(sd.Admins[0].ID.String()),
			},
			GotResp: &purchaseorderapp.PurchaseOrder{},
			ExpResp: &purchaseorderapp.PurchaseOrder{
				ID:                      sd.PurchaseOrders[1].ID,
				OrderNumber:             sd.PurchaseOrders[1].OrderNumber,
				SupplierID:              sd.PurchaseOrders[1].SupplierID,
				PurchaseOrderStatusID:   sd.PurchaseOrders[1].PurchaseOrderStatusID,
				DeliveryWarehouseID:     sd.PurchaseOrders[1].DeliveryWarehouseID,
				DeliveryLocationID:      sd.PurchaseOrders[1].DeliveryLocationID,
				DeliveryStreetID:        sd.PurchaseOrders[1].DeliveryStreetID,
				OrderDate:               sd.PurchaseOrders[1].OrderDate,
				ExpectedDeliveryDate:    sd.PurchaseOrders[1].ExpectedDeliveryDate,
				ActualDeliveryDate:      sd.PurchaseOrders[1].ActualDeliveryDate,
				Subtotal:                sd.PurchaseOrders[1].Subtotal,
				TaxAmount:               sd.PurchaseOrders[1].TaxAmount,
				ShippingCost:            sd.PurchaseOrders[1].ShippingCost,
				TotalAmount:             sd.PurchaseOrders[1].TotalAmount,
				Currency:                sd.PurchaseOrders[1].Currency,
				RequestedBy:             sd.PurchaseOrders[1].RequestedBy,
				ApprovedBy:              sd.PurchaseOrders[1].ApprovedBy,
				ApprovedDate:            sd.PurchaseOrders[1].ApprovedDate,
				Notes:                   notes,
				SupplierReferenceNumber: sd.PurchaseOrders[1].SupplierReferenceNumber,
				CreatedBy:               sd.PurchaseOrders[1].CreatedBy,
				UpdatedBy:               sd.Admins[0].ID.String(),
				CreatedDate:             sd.PurchaseOrders[1].CreatedDate,
			},
			CmpFunc: func(got, exp any) string {
				gotResp := got.(*purchaseorderapp.PurchaseOrder)
				expResp := exp.(*purchaseorderapp.PurchaseOrder)

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
			URL:        "/v1/procurement/purchase-orders/invalid-id",
			Token:      sd.Admins[0].Token,
			StatusCode: 400,
			Method:     "PUT",
			Input: &purchaseorderapp.UpdatePurchaseOrder{
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
			URL:        "/v1/procurement/purchase-orders/" + sd.PurchaseOrders[0].ID,
			Token:      sd.Users[0].Token,
			StatusCode: 401,
			Method:     "PUT",
			Input: &purchaseorderapp.UpdatePurchaseOrder{
				Notes: dbtest.StringPointer("Updated notes"),
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.Unauthenticated, "user does not have permission UPDATE for table: purchase_orders"),
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
			URL:        "/v1/procurement/purchase-orders/00000000-0000-0000-0000-000000000000",
			Token:      sd.Admins[0].Token,
			StatusCode: 404,
			Method:     "PUT",
			Input: &purchaseorderapp.UpdatePurchaseOrder{
				Notes: dbtest.StringPointer("Updated notes"),
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.NotFound, "querybyid: db: purchase order not found"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
	return table
}
