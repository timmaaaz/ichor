package purchaseorderapi_test

import (
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/app/domain/procurement/purchaseorderapp"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/foundation/timeutil"
)

func create200(sd apitest.SeedData) []apitest.Table {
	table := []apitest.Table{
		{
			Name:       "basic",
			URL:        "/v1/procurement/purchase-orders",
			Token:      sd.Admins[0].Token,
			StatusCode: 200,
			Method:     "POST",
			Input: &purchaseorderapp.NewPurchaseOrder{
				OrderNumber:             "PO-TEST-001",
				SupplierID:              sd.Suppliers[0].SupplierID,
				PurchaseOrderStatusID:   sd.PurchaseOrderStatuses[0].ID,
				DeliveryWarehouseID:     sd.Warehouses[0].ID,
				DeliveryLocationID:      sd.InventoryLocations[0].LocationID,
				DeliveryStreetID:        sd.ContactInfos[0].StreetID,
				OrderDate:               time.Now().UTC().Format(timeutil.FORMAT),
				ExpectedDeliveryDate:    time.Now().UTC().Add(time.Hour * 24 * 14).Format(timeutil.FORMAT),
				Subtotal:                "1000.00",
				TaxAmount:               "80.00",
				ShippingCost:            "50.00",
				TotalAmount:             "1130.00",
				Currency:                "USD",
				RequestedBy:             sd.Admins[0].ID.String(),
				Notes:                   "Test purchase order",
				SupplierReferenceNumber: "SUP-REF-TEST",
				CreatedBy:               sd.Admins[0].ID.String(),
			},
			GotResp: &purchaseorderapp.PurchaseOrder{},
			ExpResp: &purchaseorderapp.PurchaseOrder{
				OrderNumber:             "PO-TEST-001",
				SupplierID:              sd.Suppliers[0].SupplierID,
				PurchaseOrderStatusID:   sd.PurchaseOrderStatuses[0].ID,
				DeliveryWarehouseID:     sd.Warehouses[0].ID,
				DeliveryLocationID:      sd.InventoryLocations[0].LocationID,
				DeliveryStreetID:        sd.ContactInfos[0].StreetID,
				Subtotal:                "1000.00",
				TaxAmount:               "80.00",
				ShippingCost:            "50.00",
				TotalAmount:             "1130.00",
				Currency:                "USD",
				RequestedBy:             sd.Admins[0].ID.String(),
				Notes:                   "Test purchase order",
				SupplierReferenceNumber: "SUP-REF-TEST",
				CreatedBy:               sd.Admins[0].ID.String(),
			},
			CmpFunc: func(got, exp any) string {
				gotResp := got.(*purchaseorderapp.PurchaseOrder)
				expResp := exp.(*purchaseorderapp.PurchaseOrder)

				expResp.ID = gotResp.ID
				expResp.UpdatedBy = gotResp.UpdatedBy
				expResp.CreatedDate = gotResp.CreatedDate
				expResp.UpdatedDate = gotResp.UpdatedDate
				expResp.OrderDate = gotResp.OrderDate
				expResp.ExpectedDeliveryDate = gotResp.ExpectedDeliveryDate

				return cmp.Diff(got, exp)
			},
		},
	}
	return table
}

func create400(sd apitest.SeedData) []apitest.Table {
	table := []apitest.Table{
		{
			Name:       "missing-order-number",
			URL:        "/v1/procurement/purchase-orders",
			Token:      sd.Admins[0].Token,
			StatusCode: 400,
			Method:     "POST",
			Input: &purchaseorderapp.NewPurchaseOrder{
				SupplierID:            sd.Suppliers[0].SupplierID,
				PurchaseOrderStatusID: sd.PurchaseOrderStatuses[0].ID,
				DeliveryWarehouseID:   sd.Warehouses[0].ID,
				DeliveryLocationID:    sd.InventoryLocations[0].LocationID,
				DeliveryStreetID:      sd.ContactInfos[0].StreetID,
				OrderDate:             time.Now().UTC().Format(timeutil.FORMAT),
				ExpectedDeliveryDate:  time.Now().UTC().Add(time.Hour * 24 * 14).Format(timeutil.FORMAT),
				Subtotal:              "1000.00",
				TaxAmount:             "80.00",
				ShippingCost:          "50.00",
				TotalAmount:           "1130.00",
				Currency:              "USD",
				RequestedBy:           sd.Admins[0].ID.String(),
				CreatedBy:             sd.Admins[0].ID.String(),
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "validate: [{\"field\":\"orderNumber\",\"error\":\"orderNumber is a required field\"}]"),
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
			URL:        "/v1/procurement/purchase-orders",
			Token:      sd.Users[0].Token,
			StatusCode: 401,
			Method:     "POST",
			Input: &purchaseorderapp.NewPurchaseOrder{
				OrderNumber:           "PO-TEST-002",
				SupplierID:            sd.Suppliers[0].SupplierID,
				PurchaseOrderStatusID: sd.PurchaseOrderStatuses[0].ID,
				DeliveryWarehouseID:   sd.Warehouses[0].ID,
				DeliveryLocationID:    sd.InventoryLocations[0].LocationID,
				DeliveryStreetID:      sd.ContactInfos[0].StreetID,
				OrderDate:             time.Now().UTC().Format(timeutil.FORMAT),
				ExpectedDeliveryDate:  time.Now().UTC().Add(time.Hour * 24 * 14).Format(timeutil.FORMAT),
				Subtotal:              "1000.00",
				TaxAmount:             "80.00",
				ShippingCost:          "50.00",
				TotalAmount:           "1130.00",
				Currency:              "USD",
				RequestedBy:           sd.Users[0].ID.String(),
				CreatedBy:             sd.Users[0].ID.String(),
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.Unauthenticated, "user does not have permission CREATE for table: purchase_orders"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
	return table
}
