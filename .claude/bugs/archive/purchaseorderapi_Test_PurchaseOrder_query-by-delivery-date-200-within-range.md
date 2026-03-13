# Test Failure: Test_PurchaseOrder/query-by-delivery-date-200-within-range

- **Package**: `github.com/timmaaaz/ichor/api/cmd/services/ichor/tests/procurement/purchaseorderapi`
- **Duration**: 0.01s

## Failure Output

```
:40:48 +0000 UTC","updated_date":"2026-03-13 11:40:48 +0000 UTC"},{"id":"b1e95201-a6cd-4daa-b4c1-76c1e8e02698","order_number":"PO-9","supplier_id":"61c39b44-2c58-4edb-a14b-ac8bdc4cf752","purchase_order_status_id":"6842b178-178f-4a6a-8b64-4daae303c3a1","delivery_warehouse_id":"2a3b57cb-a5f0-4a00-928b-5dbd316b4d18","delivery_location_id":"","delivery_street_id":"d4029822-1434-4a84-bc0b-b3dc0efcdd08","order_date":"2026-03-13 11:40:48 +0000 UTC","expected_delivery_date":"2026-03-27 11:40:48 +0000 UTC","actual_delivery_date":"2024-06-01 00:00:00 +0000 UTC","subtotal":"1800.00","tax_amount":"144.00","shipping_cost":"50.00","total_amount":"1994.00","currency_id":"c933c01d-c850-4562-bd4a-eb179163bd3f","requested_by":"ed887f5b-2db3-4e97-9e41-d197e61bd6d2","approved_by":"","approved_date":"","approval_reason":"","rejected_by":"","rejected_date":"","rejection_reason":"","notes":"Test purchase order 9","supplier_reference_number":"SUP-REF-9","created_by":"ed887f5b-2db3-4e97-9e41-d197e61bd6d2","updated_by":"ed887f5b-2db3-
4e97-9e41-d197e61bd6d2","created_date":"2026-03-13 11:40:48 +0000 UTC","updated_date":"2026-03-13 11:40:48 +0000 UTC"}],"total":2,"page":1,"rows_per_page":10}
    apitest.go:73: DIFF
    apitest.go:74:   &query.Result[github.com/timmaaaz/ichor/app/domain/procurement/purchaseorderapp.PurchaseOrder]{
          	Items: []purchaseorderapp.PurchaseOrder{
          			ID: strings.Join({
        - 				"83ffee6d-3faf-457d-9a87-f395e0322bd",
        + 				"2377df7e-f114-44ba-82b0-f0e74901c6a",
          				"f",
          			}, ""),
        - 			OrderNumber: "PO-10",
        + 			OrderNumber: "PO-8",
          			SupplierID: strings.Join({
        - 				"26757c4a-65fd-4963-aeaf-5d7a042b2528",
        + 				"ba5fe3de-1c43-4905-9bf9-b12660ec5ff4",
          			}, ""),
          			PurchaseOrderStatusID: strings.Join({
        - 				"40c80e11-6047-4154-9dfe-d7b6e622a200",
        + 				"f8ecc132-129e-45ee-985f-43d4c8fdc55d",
          			}, ""),
          			DeliveryWarehouseID: strings.Join({
        - 				"cc1a634d-9e4d-4be5-9b9c-76a1bcd86858",
        + 				"da3d6bd6-6b80-4fe7-a17b-8e26a7bbfef3",
          			}, ""),
          			DeliveryLocationID: "",
          			DeliveryStreetID: strings.Join({
        - 				"f0145b96-f2ec-48e2-a3bb-f43ea936e0b0",
        + 				"276421cb-3099-4093-8f01-2a4c832d9356",
          			}, ""),
          			OrderDate:            "2026-03-13 11:40:48 +0000 UTC",
          			ExpectedDeliveryDate: "2026-03-27 11:40:48 +0000 UTC",
        - 			ActualDeliveryDate:   "2024-06-01 00:00:00 +0000 UTC",
        + 			ActualDeliveryDate:   "",
        - 			Subtotal:             "1900.00",
        + 			Subtotal:             "1700.00",
        - 			TaxAmount:            "152.00",
        + 			TaxAmount:            "136.00",
          			ShippingCost:         "50.00",
        - 			TotalAmount:          "2102.00",
        + 			TotalAmount:          "1886.00",
          			CurrencyID: strings.Join({
        - 				"9462763d-cee2-43ff-abc7-bcb7eb219e97",
        + 				"1f2ee70b-47ba-42c0-9937-171ef1f301d0",
          			}, ""),
          			RequestedBy: "80b4c8e9-78c8-4616-bb78-ed39d5175530",
          			ApprovedBy:  "",
          			... // 3 identical fields
          			RejectedDate:            "",
          			RejectionReason:         "",
        - 			Notes:                   "Test purchase order 10",
        + 			Notes:                   "Test purchase order 8",
        - 			SupplierReferenceNumber: "SUP-REF-10",
        + 			SupplierReferenceNumber: "SUP-REF-8",
          			CreatedBy:               "80b4c8e9-78c8-4616-bb78-ed39d5175530",
          			UpdatedBy:               "80b4c8e9-78c8-4616-bb78-ed39d5175530",
          			... // 2 identical fields
          		},
          	},
          	Total:       2,
          	Page:        1,
          	RowsPerPage: 10,
          }
    apitest.go:75: GOT
    apitest.go:76: &query.Result[github.com/timmaaaz/ichor/app/domain/procurement/purchaseorderapp.PurchaseOrder]{Items:[]purchaseorderapp.PurchaseOrder{purchaseorderapp.PurchaseOrder{ID:"83ffee6d-3faf-457d-9a87-f395e0322bdf", OrderNumber:"PO-10", SupplierID:"26757c4a-65fd-4963-aeaf-5d7a042b2528", PurchaseOrderStatusID:"40c80e11-6047-4154-9dfe-d7b6e622a200", DeliveryWarehouseID:"cc1a634d-9e4d-4be5-9b9c-76a1bcd86858", DeliveryLocationID:"", DeliveryStreetID:"f0145b96-f2ec-48e2-a3bb-f43ea936e0b0", OrderDate:"2026-03-13 11:40:48 +0000 UTC", ExpectedDeliveryDate:"2026-03-27 11:40:48 +0000 UTC", ActualDeliveryDate:"2024-06-01 00:00:00 +0000 UTC", Subtotal:"1900.00", TaxAmount:"152.00", ShippingCost:"50.00", TotalAmount:"2102.00", CurrencyID:"9462763d-cee2-43ff-abc7-bcb7eb219e97", RequestedBy:"80b4c8e9-78c8-4616-bb78-ed39d5175530", ApprovedBy:"", ApprovedDate:"", ApprovalReason:"", RejectedBy:"", RejectedDate:"", RejectionReason:"", Notes:"Test purchase order 10", SupplierReferenceNumber:"SUP-REF-10", CreatedBy:"80
b4c8e9-78c8-4616-bb78-ed39d5175530", UpdatedBy:"80b4c8e9-78c8-4616-bb78-ed39d5175530", CreatedDate:"2026-03-13 11:40:48 +0000 UTC", UpdatedDate:"2026-03-13 11:40:48 +0000 UTC"}, purchaseorderapp.PurchaseOrder{ID:"b1e95201-a6cd-4daa-b4c1-76c1e8e02698", OrderNumber:"PO-9", SupplierID:"61c39b44-2c58-4edb-a14b-ac8bdc4cf752", PurchaseOrderStatusID:"6842b178-178f-4a6a-8b64-4daae303c3a1", DeliveryWarehouseID:"2a3b57cb-a5f0-4a00-928b-5dbd316b4d18", DeliveryLocationID:"", DeliveryStreetID:"d4029822-1434-4a84-bc0b-b3dc0efcdd08", OrderDate:"2026-03-13 11:40:48 +0000 UTC", ExpectedDeliveryDate:"2026-03-27 11:40:48 +0000 UTC", ActualDeliveryDate:"2024-06-01 00:00:00 +0000 UTC", Subtotal:"1800.00", TaxAmount:"144.00", ShippingCost:"50.00", TotalAmount:"1994.00", CurrencyID:"c933c01d-c850-4562-bd4a-eb179163bd3f", RequestedBy:"ed887f5b-2db3-4e97-9e41-d197e61bd6d2", ApprovedBy:"", ApprovedDate:"", ApprovalReason:"", RejectedBy:"", RejectedDate:"", RejectionReason:"", Notes:"Test purchase order 9", SupplierReferenceNumber:"SUP
-REF-9", CreatedBy:"ed887f5b-2db3-4e97-9e41-d197e61bd6d2", UpdatedBy:"ed887f5b-2db3-4e97-9e41-d197e61bd6d2", CreatedDate:"2026-03-13 11:40:48 +0000 UTC", UpdatedDate:"2026-03-13 11:40:48 +0000 UTC"}}, Total:2, Page:1, RowsPerPage:10}
    apitest.go:77: EXP
    apitest.go:78: &query.Result[github.com/timmaaaz/ichor/app/domain/procurement/purchaseorderapp.PurchaseOrder]{Items:[]purchaseorderapp.PurchaseOrder{purchaseorderapp.PurchaseOrder{ID:"2377df7e-f114-44ba-82b0-f0e74901c6af", OrderNumber:"PO-8", SupplierID:"ba5fe3de-1c43-4905-9bf9-b12660ec5ff4", PurchaseOrderStatusID:"f8ecc132-129e-45ee-985f-43d4c8fdc55d", DeliveryWarehouseID:"da3d6bd6-6b80-4fe7-a17b-8e26a7bbfef3", DeliveryLocationID:"", DeliveryStreetID:"276421cb-3099-4093-8f01-2a4c832d9356", OrderDate:"2026-03-13 11:40:48 +0000 UTC", ExpectedDeliveryDate:"2026-03-27 11:40:48 +0000 UTC", ActualDeliveryDate:"", Subtotal:"1700.00", TaxAmount:"136.00", ShippingCost:"50.00", TotalAmount:"1886.00", CurrencyID:"1f2ee70b-47ba-42c0-9937-171ef1f301d0", RequestedBy:"80b4c8e9-78c8-4616-bb78-ed39d5175530", ApprovedBy:"", ApprovedDate:"", ApprovalReason:"", RejectedBy:"", RejectedDate:"", RejectionReason:"", Notes:"Test purchase order 8", SupplierReferenceNumber:"SUP-REF-8", CreatedBy:"80b4c8e9-78c8-4616-bb78-ed39d51755
30", UpdatedBy:"80b4c8e9-78c8-4616-bb78-ed39d5175530", CreatedDate:"2026-03-13 11:40:48 +0000 UTC", UpdatedDate:"2026-03-13 11:40:48 +0000 UTC"}, purchaseorderapp.PurchaseOrder{ID:"b1e95201-a6cd-4daa-b4c1-76c1e8e02698", OrderNumber:"PO-9", SupplierID:"61c39b44-2c58-4edb-a14b-ac8bdc4cf752", PurchaseOrderStatusID:"6842b178-178f-4a6a-8b64-4daae303c3a1", DeliveryWarehouseID:"2a3b57cb-a5f0-4a00-928b-5dbd316b4d18", DeliveryLocationID:"", DeliveryStreetID:"d4029822-1434-4a84-bc0b-b3dc0efcdd08", OrderDate:"2026-03-13 11:40:48 +0000 UTC", ExpectedDeliveryDate:"2026-03-27 11:40:48 +0000 UTC", ActualDeliveryDate:"2024-06-01 00:00:00 +0000 UTC", Subtotal:"1800.00", TaxAmount:"144.00", ShippingCost:"50.00", TotalAmount:"1994.00", CurrencyID:"c933c01d-c850-4562-bd4a-eb179163bd3f", RequestedBy:"ed887f5b-2db3-4e97-9e41-d197e61bd6d2", ApprovedBy:"", ApprovedDate:"", ApprovalReason:"", RejectedBy:"", RejectedDate:"", RejectionReason:"", Notes:"Test purchase order 9", SupplierReferenceNumber:"SUP-REF-9", CreatedBy:"ed887f5b-2db
3-4e97-9e41-d197e61bd6d2", UpdatedBy:"ed887f5b-2db3-4e97-9e41-d197e61bd6d2", CreatedDate:"2026-03-13 11:40:48 +0000 UTC", UpdatedDate:"2026-03-13 11:40:48 +0000 UTC"}}, Total:2, Page:1, RowsPerPage:10}
    apitest.go:79: Should get the expected response
--- FAIL: Test_PurchaseOrder/query-by-delivery-date-200-within-range (0.01s)
```

## Fix
- **File**: `api/cmd/services/ichor/tests/procurement/purchaseorderapi/query_test.go:40-44`
- **Classification**: test bug — CmpFunc sort mutated shared slice backing array
- **Change**: In `query200`'s CmpFunc, copy `expResp.Items` via `append([]PurchaseOrder{}, expResp.Items...)` before sorting, preventing mutation of `sd.PurchaseOrders` which is shared across all subtests. The alphabetical sort of "PO-1".."PO-10" placed "PO-10" at index 1 (before "PO-2"), corrupting all subsequent index-based EXP slices (`sd.PurchaseOrders[8:]` became `{PO-8, PO-9}` instead of `{PO-9, PO-10}`).
- **Verified**: `go build ./api/cmd/services/ichor/tests/procurement/purchaseorderapi/...` ✓
