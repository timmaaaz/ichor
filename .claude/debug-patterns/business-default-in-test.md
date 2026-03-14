# business-default-in-test

**Signal**: DIFF on a create-200 test: field is non-empty in GOT but empty string / zero in EXP; field name is a status, state, or classification field (e.g., `ApprovalStatus`, `Status`, `State`)
**Root cause**: The business layer sets a default value on Create (e.g., `ApprovalStatus = "pending"`), but the test's `ExpResp` was written before that default was added or without knowledge of it.
**Fix**:
1. Identify the field in GOT that is populated but missing from EXP
2. Trace to the business layer `Create` method to confirm the default is intentional
3. Add the expected default value to the test's `ExpResp` struct

**See also**: `docs/arch/domain-template.md`
**Examples**:
- `inventoryadjustmentapi_Test_InventoryLocations_create-200-basic.md` — `ApprovalStatus: "pending"` was being set by business layer on create; test `ExpResp` was written with empty string
- `inventoryadjustmentbus_Test_InventoryAdjustment_create-create.md` + `update-update.md` — bus-layer tests also missing `ApprovalStatus: ApprovalStatusPending` in `ExpResp`; default is set in `inventoryadjustmentbus.go Create()` at line 90
- `orderlineitemsapi_Test_OrderLineItem_create-200-basic.md` — `PickedQuantity: "0"` and `BackorderedQuantity: "0"` absent from `ExpResp`; fields are `int` in bus model so Go zero-value is 0; `strconv.Itoa(0)` → `"0"` in response
- `orderlineitemsapi_Test_OrderLineItem_update-200-basic.md` — same `PickedQuantity: "0"` and `BackorderedQuantity: "0"` missing from update test `ExpResp`; pattern applies to update-200 tests too, not only create-200
- `ordersapi_Test_Order_update-200-basic.md` — `AssignedTo` field missing from update `ExpResp`; business layer populates it, test was written without it
- `ordersapi_Test_Order.md` — same `AssignedTo` fix; composite bug file covering the update-200 subtest
- `productapi_Test_InventoryProduct_create-200-basic.md` — `TrackingType: "none"` missing from create `ExpResp`; business layer sets default `"none"` on product create
- `productapi_Test_InventoryProduct_update-200-basic.md` — same `TrackingType: "none"` missing from update `ExpResp`
- `workflowsaveapi_Test_WorkflowSaveAPI_exec-no-matching-rules.md` — `Timestamp: time.Now()` missing from `TriggerEvent` construction in test; bus-layer validation requires it; pattern extends to required input fields, not just response fields
