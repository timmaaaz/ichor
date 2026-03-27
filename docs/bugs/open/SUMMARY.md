# Test Failures Summary
Generated: 2026-03-27T15:29:51Z
Total: 12 failures across 6 packages

## Failures

- inspectionapi_Test_Inspections_fail-200-no-quarantine-fail-without-quarantine.md — apitest.go:57: fail-without-quarantine: Should receive a status code of 200 for the response : 404
- inspectionapi_Test_Inspections_fail-403-fail-forbidden.md — apitest.go:57: fail-forbidden: Should receive a status code of 403 for the response : 401
- inspectionapi_Test_Inspections_fail-404-fail-not-found.md — apitest.go:68: Should be able to unmarshal the response : json: Unmarshal(nil)
- referenceapi_Test_ActionTypeSchemas_CategoryConsistency.md — schema_alignment_test.go:171: Category "inventory": expected 11 types, got 12 (found: [allocate_inve
- permissionsbus_Test_Permissions_query-Query.md — unittest.go:17: DIFF
- inspectionbus_Test_Inspections_create-Create.md — unittest.go:17: DIFF
- inspectionbus_Test_Inspections_update-Update.md — unittest.go:17: DIFF
- transferorderbus_Test_TransferOrders_approve-approve-pending-succeeds.md — unittest.go:17: DIFF
- transferorderbus_Test_TransferOrders_claim-claim-approved-succeeds.md — unittest.go:17: DIFF
- transferorderbus_Test_TransferOrders_execute-execute-in-transit-succeeds.md — unittest.go:17: DIFF
- inventory_Test_ApproveInventoryAdjustment.md — approve_adjustment_test.go:38: seeding: creating pre-approve adjustment: create: invalid reason code
- inventory_Test_RejectInventoryAdjustment.md — reject_adjustment_test.go:23: seeding: creating pre-approve adjustment: create: invalid reason code
