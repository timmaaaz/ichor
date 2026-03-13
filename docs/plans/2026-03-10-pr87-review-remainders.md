# PR #87 Code Review — Remaining Issues Fix Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Branch:** `feature/workflow-semantic-gaps`
**PR:** https://github.com/timmaaaz/ichor/pull/87
**Goal:** Fix the 4 open + 1 partial issue from the two code review comments on PR #87.

---

## Context

PR #87 has two code review comments with 10 total issues. 5 are addressed, 1 is partial, 4 are open. This plan fixes all remaining issues. Each task is independent and can be committed separately.

### Relevant arch files
- `docs/arch/sqldb.md` — for NamedQuery error handling patterns
- `docs/arch/errs.md` — for error code mapping conventions
- `docs/arch/domain-template.md` — for layer conventions

---

## Task 1: Fix `purchaseorderapp.Approve()` missing `ErrAlreadyRejected` mapping

**Review 1, Issue 4 (partial fix)**

The `Reject()` method correctly maps both sentinel errors, but `Approve()` only maps `ErrAlreadyApproved`. If a PO is rejected first, then a concurrent `Approve()` call returns HTTP 500 instead of 400.

### Files
- `app/domain/procurement/purchaseorderapp/purchaseorderapp.go` (~line 168)

### Steps

1. Open `app/domain/procurement/purchaseorderapp/purchaseorderapp.go`
2. Find the `Approve()` method's error handling block (around line 168):
   ```go
   if errors.Is(err, purchaseorderbus.ErrAlreadyApproved) {
       return PurchaseOrder{}, errs.New(errs.InvalidArgument, err)
   }
   ```
3. Change it to match the pattern in `Reject()`:
   ```go
   if errors.Is(err, purchaseorderbus.ErrAlreadyApproved) || errors.Is(err, purchaseorderbus.ErrAlreadyRejected) {
       return PurchaseOrder{}, errs.New(errs.InvalidArgument, err)
   }
   ```
4. Verify: `go build ./app/domain/procurement/purchaseorderapp/...`

### Commit
```
fix(purchaseorderapp): map ErrAlreadyRejected to HTTP 400 in Approve()
```

---

## Task 2: Fix `seek_approval` SupportsManual metadata contradiction

**Review 2, Issue 2**

`SeekApprovalHandler.SupportsManualExecution()` returns `true` but `actionschemas.go` has `SupportsManual: false`. The frontend discovery endpoint suppresses the "Run manually" button even though the backend supports it.

### Files
- `api/domain/http/workflow/referenceapi/actionschemas.go` (~line 129, the `"seek_approval"` entry)

### Steps

1. Open `api/domain/http/workflow/referenceapi/actionschemas.go`
2. Find the `"seek_approval"` entry in `actionTypeMetadata`
3. Change `SupportsManual: false` → `SupportsManual: true`
4. Verify: `go build ./api/domain/http/workflow/referenceapi/...`
5. Check if integration tests assert on this value:
   ```bash
   grep -rn 'seek_approval.*SupportsManual\|SupportsManual.*seek_approval\|supports_manual' api/cmd/services/ichor/tests/workflow/referenceapi/
   ```
   If any tests assert `SupportsManual: false` for seek_approval, update them to `true`.

### Commit
```
fix(referenceapi): set seek_approval SupportsManual to true (matches handler)
```

---

## Task 3: Fix `resolve_approval_request` IsAsync flag

**Review 2, Issue 3**

`ResolveApprovalHandler.IsAsync()` returns `true` and `actionschemas.go` has `IsAsync: true`, but the handler is **synchronous** — it calls `approvalBus.Resolve()` directly and returns the result. It is never registered in `AsyncRegistry`. Any workflow DAG that includes this action type will fail at runtime because `selectActivityFunc` routes it to `ExecuteAsyncActionActivity`, which looks it up in `AsyncRegistry` where it doesn't exist.

### Files
- `business/sdk/workflow/workflowactions/approval/resolve.go` (~line 37)
- `api/domain/http/workflow/referenceapi/actionschemas.go` (~line 227, the `"resolve_approval_request"` entry)

### Steps

1. Open `business/sdk/workflow/workflowactions/approval/resolve.go`
2. Change line 37:
   ```go
   func (h *ResolveApprovalHandler) IsAsync() bool { return true }
   ```
   to:
   ```go
   func (h *ResolveApprovalHandler) IsAsync() bool { return false }
   ```
3. Update the comment on line 36 to reflect the change:
   ```go
   // IsAsync returns false — resolution is a synchronous database operation.
   ```

4. Open `api/domain/http/workflow/referenceapi/actionschemas.go`
5. Find the `"resolve_approval_request"` entry and change `IsAsync: true` → `IsAsync: false`

6. Verify: `go build ./business/sdk/workflow/workflowactions/... && go build ./api/domain/http/workflow/referenceapi/...`
7. Check if integration tests assert on this value:
   ```bash
   grep -rn 'resolve_approval_request.*IsAsync\|resolve_approval_request.*is_async\|async.*resolve' api/cmd/services/ichor/tests/workflow/referenceapi/
   ```
   The `schema_alignment_test.go` likely has a test that asserts `IsAsync: true` for this type — update it to `false`.

### Commit
```
fix(workflow): set resolve_approval_request IsAsync to false (handler is synchronous)
```

---

## Task 4: Fix Resolve store conflating "not found" with "already resolved"

**Review 2, Issue 4**

The `Resolve` store method uses `UPDATE ... WHERE id = :id AND status = 'pending' RETURNING *`. When zero rows are returned, `NamedQueryStruct` returns `ErrDBNotFound`, which is unconditionally mapped to `ErrAlreadyResolved`. This means:
- Non-existent approval ID → `ErrAlreadyResolved` (should be `ErrNotFound`)
- Already-resolved approval → `ErrAlreadyResolved` (correct)

The fix: pre-check existence with `QueryByID` before the UPDATE, or use a two-step query pattern.

### Files
- `business/domain/workflow/approvalrequestbus/stores/approvalrequestdb/approvalrequestdb.go` (~line 97, `Resolve` method)

### Steps

1. Open `approvalrequestdb.go` and find the `Resolve` method
2. Replace the current error handling with a disambiguation approach. When `ErrDBNotFound` is returned from the UPDATE, do a follow-up `QueryByID` to check if the record exists:

   ```go
   if err := sqldb.NamedQueryStruct(ctx, s.log, s.db, q, data, &dbReq); err != nil {
       if errors.Is(err, sqldb.ErrDBNotFound) {
           // Disambiguate: does the record exist at all?
           if _, qErr := s.QueryByID(ctx, id); qErr != nil {
               return approvalrequestbus.ApprovalRequest{}, approvalrequestbus.ErrNotFound
           }
           return approvalrequestbus.ApprovalRequest{}, approvalrequestbus.ErrAlreadyResolved
       }
       return approvalrequestbus.ApprovalRequest{}, fmt.Errorf("resolve: %w", err)
   }
   ```

3. Verify `ErrNotFound` is defined in `approvalrequestbus` (it is — check `model.go` for `ErrNotFound = errors.New("approval request not found")`).

4. Update the `resolve` HTTP handler in `approvalapi.go` to handle the new `ErrNotFound` case. Currently the `resolve` method only has an `ErrAlreadyResolved` branch. Add:
   ```go
   if errors.Is(err, approvalrequestbus.ErrNotFound) {
       return errs.New(errs.NotFound, err)
   }
   ```
   This goes in the resolve method's error handling block, before the `ErrAlreadyResolved` check.

5. Verify: `go build ./business/domain/workflow/approvalrequestbus/... && go build ./api/domain/http/workflow/approvalapi/...`

6. Run existing unit tests to make sure nothing breaks:
   ```bash
   go test ./api/domain/http/workflow/approvalapi/... -run TestRetry -v
   ```

7. **Add a unit test** for the not-found case in `api/domain/http/workflow/approvalapi/resolve_test.go`. The existing `TestRetryTemporalCompletion_QueryByIDFails` uses a generic DB error. Add a new test where the storer's `Resolve` method returns `ErrNotFound`:

   Actually — `retryTemporalCompletion` is only called on the `ErrAlreadyResolved` path, so the not-found case is handled BEFORE the retry logic. The existing resolve_test.go tests `retryTemporalCompletion` only. The not-found case is best tested via the integration tests.

### Commit
```
fix(approvalrequestdb): disambiguate not-found from already-resolved in Resolve store
```

---

## Task 5: Restore deleted migration version 2.05

**Review 1, Issue 3**

Version 2.05 (`ALTER TABLE inventory.transfer_orders ALTER COLUMN approved_by DROP NOT NULL`) was deleted from `migrate.sql` and its effect was inlined into the `CREATE TABLE` at version 1.58. Any database that already ran version 2.05 will diverge from fresh installs.

The fix: restore version 2.05 as it exists on `master`, and add NEW versions for the additional columns added by this PR.

### Files
- `business/sdk/migrate/sql/migrate.sql`

### Steps

1. Check master for version 2.05's exact content:
   ```sql
   -- Version: 2.05
   -- Description: Make transfer_orders.approved_by nullable to support pending-approval workflow.
   ALTER TABLE inventory.transfer_orders
       ALTER COLUMN approved_by DROP NOT NULL;
   ```

2. The CREATE TABLE at version 1.57 (inventory_adjustments) and 1.58 (transfer_orders) currently include columns that didn't exist in the original master version:
   - `inventory_adjustments`: `approval_reason`, `rejected_by`, `rejection_reason` were added inline
   - `transfer_orders`: `approval_reason`, `rejected_by_id`, `rejection_reason` were added inline
   - `purchase_orders`: `approval_reason`, `rejected_by`, `rejected_date`, `rejection_reason` were added inline (at 1.87)

3. **Restore** the CREATE TABLE statements at 1.57 and 1.58 to match master (remove the inlined audit columns). Restore version 2.05 in its original position.

4. **Add new ALTER TABLE migrations** after version 2.08 (the last existing version) that add the new columns:

   ```sql
   -- Version: 2.09
   -- Description: Add rejection audit trail columns to inventory_adjustments.
   ALTER TABLE inventory.inventory_adjustments
       ADD COLUMN approval_reason TEXT NULL,
       ADD COLUMN rejected_by UUID NULL REFERENCES core.users(id),
       ADD COLUMN rejection_reason TEXT NULL;

   -- Version: 2.10
   -- Description: Add rejection audit trail columns to transfer_orders.
   ALTER TABLE inventory.transfer_orders
       ADD COLUMN approval_reason TEXT NULL,
       ADD COLUMN rejected_by_id UUID NULL REFERENCES core.users(id),
       ADD COLUMN rejection_reason TEXT NULL;

   -- Version: 2.11
   -- Description: Add rejection audit trail columns to purchase_orders.
   ALTER TABLE procurement.purchase_orders
       ADD COLUMN approval_reason TEXT NULL,
       ADD COLUMN rejected_by UUID NULL REFERENCES core.users(id) ON DELETE SET NULL,
       ADD COLUMN rejected_date TIMESTAMP NULL,
       ADD COLUMN rejection_reason TEXT NULL;
   ```

   > **Note:** Check which columns already exist on master for purchase_orders. On master, version 1.87 already has `rejected_by`, `rejected_date`, `rejection_reason`, and `approval_reason` in the CREATE TABLE. So the purchase_orders ALTER may not be needed — verify by diffing 1.87 between master and feature branch. If the columns were already there on master, only inventory_adjustments and transfer_orders need new ALTER versions.

5. Verify the migration file is syntactically valid:
   ```bash
   go build ./business/sdk/migrate/...
   ```

### Commit
```
fix(migrate): restore version 2.05 and add ALTER migrations for audit trail columns
```

---

## Verification

After all 5 tasks:

```bash
go build ./app/domain/procurement/purchaseorderapp/... && \
go build ./api/domain/http/workflow/referenceapi/... && \
go build ./business/sdk/workflow/workflowactions/... && \
go build ./business/domain/workflow/approvalrequestbus/... && \
go build ./api/domain/http/workflow/approvalapi/... && \
go build ./business/sdk/migrate/... && \
echo "all packages build OK"
```

Run targeted tests:
```bash
go test ./api/domain/http/workflow/approvalapi/... -run TestRetry -v
go test ./api/cmd/services/ichor/tests/workflow/referenceapi/... -v -timeout 120s
```

---

## Summary

| Task | Issue | Complexity | Files changed |
|------|-------|-----------|---------------|
| 1 | `purchaseorderapp.Approve()` error mapping | One-liner | 1 |
| 2 | `seek_approval` SupportsManual | One-liner + possible test update | 1-2 |
| 3 | `resolve_approval_request` IsAsync | Two files + possible test update | 2-3 |
| 4 | Resolve store not-found disambiguation | Store + handler + possible test | 2-3 |
| 5 | Restore migration version 2.05 | Migration file surgery | 1 |
