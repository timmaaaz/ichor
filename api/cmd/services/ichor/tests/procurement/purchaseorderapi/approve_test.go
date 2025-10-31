package purchaseorderapi_test

import (
	"github.com/google/go-cmp/cmp"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/app/domain/procurement/purchaseorderapp"
	"github.com/timmaaaz/ichor/app/sdk/errs"
)

func approve200(sd apitest.SeedData) []apitest.Table {
	table := []apitest.Table{
		{
			Name:       "basic",
			URL:        "/v1/procurement/purchase-orders/" + sd.PurchaseOrders[0].ID + "/approve",
			Token:      sd.Admins[0].Token,
			StatusCode: 200,
			Method:     "POST",
			Input: &purchaseorderapp.ApproveRequest{
				ApprovedBy: sd.Admins[0].ID.String(),
			},
			GotResp: &purchaseorderapp.PurchaseOrder{},
			ExpResp: &purchaseorderapp.PurchaseOrder{
				ID:                      sd.PurchaseOrders[0].ID,
				OrderNumber:             sd.PurchaseOrders[0].OrderNumber,
				SupplierID:              sd.PurchaseOrders[0].SupplierID,
				PurchaseOrderStatusID:   sd.PurchaseOrders[0].PurchaseOrderStatusID,
				DeliveryWarehouseID:     sd.PurchaseOrders[0].DeliveryWarehouseID,
				DeliveryLocationID:      sd.PurchaseOrders[0].DeliveryLocationID,
				DeliveryStreetID:        sd.PurchaseOrders[0].DeliveryStreetID,
				OrderDate:               sd.PurchaseOrders[0].OrderDate,
				ExpectedDeliveryDate:    sd.PurchaseOrders[0].ExpectedDeliveryDate,
				ActualDeliveryDate:      sd.PurchaseOrders[0].ActualDeliveryDate,
				Subtotal:                sd.PurchaseOrders[0].Subtotal,
				TaxAmount:               sd.PurchaseOrders[0].TaxAmount,
				ShippingCost:            sd.PurchaseOrders[0].ShippingCost,
				TotalAmount:             sd.PurchaseOrders[0].TotalAmount,
				Currency:                sd.PurchaseOrders[0].Currency,
				RequestedBy:             sd.PurchaseOrders[0].RequestedBy,
				ApprovedBy:              sd.Admins[0].ID.String(),
				Notes:                   sd.PurchaseOrders[0].Notes,
				SupplierReferenceNumber: sd.PurchaseOrders[0].SupplierReferenceNumber,
				CreatedBy:               sd.PurchaseOrders[0].CreatedBy,
				UpdatedBy:               sd.Admins[0].ID.String(),
				CreatedDate:             sd.PurchaseOrders[0].CreatedDate,
			},
			CmpFunc: func(got, exp any) string {
				gotResp := got.(*purchaseorderapp.PurchaseOrder)
				expResp := exp.(*purchaseorderapp.PurchaseOrder)

				expResp.ApprovedDate = gotResp.ApprovedDate
				expResp.UpdatedDate = gotResp.UpdatedDate

				return cmp.Diff(got, exp)
			},
		},
	}
	return table
}

func approve401(sd apitest.SeedData) []apitest.Table {
	table := []apitest.Table{
		{
			Name:       "unauthorized",
			URL:        "/v1/procurement/purchase-orders/" + sd.PurchaseOrders[1].ID + "/approve",
			Token:      sd.Users[0].Token,
			StatusCode: 401,
			Method:     "POST",
			Input: &purchaseorderapp.ApproveRequest{
				ApprovedBy: sd.Users[0].ID.String(),
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.Unauthenticated, "authorize: you are not authorized for that action, claims[[USER]] rule[rule_admin_only]: rego evaluation failed : bindings results[[{[true] map[x:false]}]] ok[true]"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
	return table
}

func approve404(sd apitest.SeedData) []apitest.Table {
	table := []apitest.Table{
		{
			Name:       "not-found",
			URL:        "/v1/procurement/purchase-orders/00000000-0000-0000-0000-000000000000/approve",
			Token:      sd.Admins[0].Token,
			StatusCode: 404,
			Method:     "POST",
			Input: &purchaseorderapp.ApproveRequest{
				ApprovedBy: sd.Admins[0].ID.String(),
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
