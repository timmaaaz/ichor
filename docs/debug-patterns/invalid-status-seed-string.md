# invalid-status-seed-string

**Signal**: `must be pending, got status4`, `must be approved, got status3`, `not in a state that allows`; testutil.go generates status via `fmt.Sprintf("status%d", idx%N)` or similar computed pattern; state machine rejects the fabricated status
**Root cause**: Shared test seed utility (`testutil.go`) generates placeholder status strings using index-based formatting instead of using domain status constants. When the bus layer enforces a state machine (e.g., must be "pending" before approval), the fabricated string fails validation.
**Fix**:
1. Open the domain's `testutil.go` (e.g., `business/domain/inventory/transferorderbus/testutil.go`)
2. Find the status field generation (typically `fmt.Sprintf("status%d", ...)`)
3. Replace with the appropriate domain constant (e.g., `StatusPending`, `StatusApproved`)
4. If multiple tests depend on different statuses, seed entities in the correct initial state and use bus methods to transition them

**See also**: `docs/arch/domain-template.md`
**Examples**:
- `transferorderbus_Test_TransferOrders_approve-approve-pending-succeeds.md` — testutil used `fmt.Sprintf("status%d", idx%5)` producing "status4"; fixed by using `StatusPending` constant
- `transferorderbus_Test_TransferOrders_claim-claim-approved-succeeds.md` — same testutil fix; claim requires "approved" status
- `transferorderbus_Test_TransferOrders_execute-execute-in-transit-succeeds.md` — same testutil fix; execute requires "in_transit" status
