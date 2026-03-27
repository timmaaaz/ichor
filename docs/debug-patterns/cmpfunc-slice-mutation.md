---
name: cmpfunc-slice-mutation
description: CmpFunc sorts expResp.Items in-place, mutating the shared backing array — causes DIFF failures in other subtests that observe the mutated order
type: feedback
---

# cmpfunc-slice-mutation

**Signal**: Intermittent or parallel-subtest DIFF failure on a query-200 test; specific items appear at wrong index in GOT vs EXP; passes in isolation but fails when run with sibling subtests; `sort.Slice` present in `CmpFunc`
**Root cause**: `CmpFunc` calls `sort.Slice(expResp.Items, ...)` directly on the slice value inside `expResp`. Because `expResp` was built from a loop or shared seed-derived variable, multiple subtests hold slices backed by the same underlying array. Sorting in one subtest reorders the backing array, so sibling subtests see a different order than they expect.
**Fix**:
1. In the `CmpFunc`, before sorting, copy the slice to a new backing array:
   ```go
   expItems := append([]ExpectedType(nil), expResp.Items...)
   sort.Slice(expItems, func(i, j int) bool { ... })
   gotItems := append([]ExpectedType(nil), gotResp.Items...)
   sort.Slice(gotItems, func(i, j int) bool { ... })
   ```
2. Compare `expItems` vs `gotItems` instead of the originals
3. Apply the same copy to `gotResp.Items` if it is also sourced from a shared structure

**See also**: `docs/arch/testing.md`
**Examples**:
- `purchaseorderapi_Test_PurchaseOrder_query-by-delivery-date-200-within-range.md` — `CmpFunc` sorted `expResp.Items` in-place; PO-10 appeared at wrong index when run alongside other subtests; fixed by copying slice before sort
- `purchaseorderapi_Test_PurchaseOrder_query-is-undelivered-200-is-undelivered-true.md` — same root cause; `expResp.Items` sorted in-place corrupted the backing array shared across `is-undelivered` subtests
