package purchaseorderlineitemapi_test

import (
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/app/domain/procurement/purchaseorderlineitemapp"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/foundation/timeutil"
)

func create200(sd apitest.SeedData) []apitest.Table {
	table := []apitest.Table{
		{
			Name:       "basic",
			URL:        "/v1/procurement/purchase-order-line-items",
			Token:      sd.Admins[0].Token,
			StatusCode: 200,
			Method:     "POST",
			Input: &purchaseorderlineitemapp.NewPurchaseOrderLineItem{
				PurchaseOrderID:      sd.PurchaseOrders[0].ID,
				SupplierProductID:    sd.SupplierProducts[0].SupplierProductID,
				QuantityOrdered:      "100",
				UnitCost:             "25.50",
				Discount:             "2.50",
				LineTotal:            "2300.00",
				LineItemStatusID:     sd.PurchaseOrderLineItemStatuses[0].ID,
				ExpectedDeliveryDate: time.Now().UTC().Add(time.Hour * 24 * 14).Format(timeutil.FORMAT),
				Notes:                "Test line item creation",
				CreatedBy:            sd.Admins[0].ID.String(),
			},
			GotResp: &purchaseorderlineitemapp.PurchaseOrderLineItem{},
			ExpResp: &purchaseorderlineitemapp.PurchaseOrderLineItem{
				PurchaseOrderID:   sd.PurchaseOrders[0].ID,
				SupplierProductID: sd.SupplierProducts[0].SupplierProductID,
				QuantityOrdered:   "100",
				QuantityReceived:  "0",
				QuantityCancelled: "0",
				UnitCost:          "25.50",
				Discount:          "2.50",
				LineTotal:         "2300.00",
				LineItemStatusID:  sd.PurchaseOrderLineItemStatuses[0].ID,
				Notes:             "Test line item creation",
				CreatedBy:         sd.Admins[0].ID.String(),
			},
			CmpFunc: func(got, exp any) string {
				gotResp := got.(*purchaseorderlineitemapp.PurchaseOrderLineItem)
				expResp := exp.(*purchaseorderlineitemapp.PurchaseOrderLineItem)

				expResp.ID = gotResp.ID
				expResp.UpdatedBy = gotResp.UpdatedBy
				expResp.CreatedDate = gotResp.CreatedDate
				expResp.UpdatedDate = gotResp.UpdatedDate
				expResp.ExpectedDeliveryDate = gotResp.ExpectedDeliveryDate
				expResp.ActualDeliveryDate = gotResp.ActualDeliveryDate

				return cmp.Diff(got, exp)
			},
		},
	}
	return table
}

func create400(sd apitest.SeedData) []apitest.Table {
	table := []apitest.Table{
		{
			Name:       "missing-purchase-order-id",
			URL:        "/v1/procurement/purchase-order-line-items",
			Token:      sd.Admins[0].Token,
			StatusCode: 400,
			Method:     "POST",
			Input: &purchaseorderlineitemapp.NewPurchaseOrderLineItem{
				SupplierProductID:    sd.SupplierProducts[0].SupplierProductID,
				QuantityOrdered:      "100",
				UnitCost:             "25.50",
				Discount:             "2.50",
				LineTotal:            "2300.00",
				LineItemStatusID:     sd.PurchaseOrderLineItemStatuses[0].ID,
				ExpectedDeliveryDate: time.Now().UTC().Add(time.Hour * 24 * 14).Format(timeutil.FORMAT),
				CreatedBy:            sd.Admins[0].ID.String(),
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "validate: [{\"field\":\"purchaseOrderId\",\"error\":\"purchaseOrderId is a required field\"}]"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
	return table
}

func create401(sd apitest.SeedData) []apitest.Table {
	table := []apitest.Table{
		{
			Name:       "unauthorized",
			URL:        "/v1/procurement/purchase-order-line-items",
			Token:      sd.Users[0].Token,
			StatusCode: 401,
			Method:     "POST",
			Input: &purchaseorderlineitemapp.NewPurchaseOrderLineItem{
				PurchaseOrderID:      sd.PurchaseOrders[0].ID,
				SupplierProductID:    sd.SupplierProducts[0].SupplierProductID,
				QuantityOrdered:      "100",
				UnitCost:             "25.50",
				Discount:             "2.50",
				LineTotal:            "2300.00",
				LineItemStatusID:     sd.PurchaseOrderLineItemStatuses[0].ID,
				ExpectedDeliveryDate: time.Now().UTC().Add(time.Hour * 24 * 14).Format(timeutil.FORMAT),
				CreatedBy:            sd.Users[0].ID.String(),
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.Unauthenticated, "user does not have permission CREATE for table: procurement.purchase_order_line_items"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
	return table
}
