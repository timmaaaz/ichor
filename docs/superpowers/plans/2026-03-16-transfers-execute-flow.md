# Transfers Execute Flow Implementation Plan

> **For agentic workers:** REQUIRED: Use superpowers:subagent-driven-development (if subagents available) or superpowers:executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.
>
> **Worktree:** Create a worktree before executing: `create a worktree for blocker-004-transfers and execute this plan`

**Goal:** Add Claim and Execute endpoints to transfer orders so floor workers can pick up approved transfers and atomically move stock between locations.

**Architecture:** Add `in_transit` and `completed` statuses. `Claim` is a simple status transition. `Execute` uses the 3-way atomic write pattern (update transfer status + create two inventory transactions + decrement source + increment destination). Extend `transferorderapp.App` with inventory buses and `db *sqlx.DB`, following the `putawaytaskapp` pattern.

**Tech Stack:** Go 1.23, PostgreSQL, Ardan Labs service architecture

---

## Phase 1: Migration — Add new columns

- [ ] **Step 1.1** Add migration v2.17 to `business/sdk/migrate/sql/migrate.sql`

Append after line 2259 (the end of v2.14):

```sql
-- Version: 2.17
-- Description: Add claim/complete tracking columns to transfer_orders.
ALTER TABLE inventory.transfer_orders ADD COLUMN claimed_by UUID NULL REFERENCES core.users(id);
ALTER TABLE inventory.transfer_orders ADD COLUMN claimed_at TIMESTAMP NULL;
ALTER TABLE inventory.transfer_orders ADD COLUMN completed_by UUID NULL REFERENCES core.users(id);
ALTER TABLE inventory.transfer_orders ADD COLUMN completed_at TIMESTAMP NULL;
```

---

## Phase 2: Business layer — New statuses, model fields, Claim/Execute methods

- [ ] **Step 2.1** Add new status constants in `business/domain/inventory/transferorderbus/transferorderbus.go`

At line 31 (after `StatusRejected`), add:

```go
	StatusInTransit = "in_transit"
	StatusCompleted = "completed"
```

- [ ] **Step 2.2** Add new fields to `TransferOrder` model in `business/domain/inventory/transferorderbus/model.go`

Add these fields to the `TransferOrder` struct (after `RejectionReason string` on line 23):

```go
	ClaimedByID *uuid.UUID `json:"claimed_by_id"`
	ClaimedAt   *time.Time `json:"claimed_at"`
	CompletedByID *uuid.UUID `json:"completed_by_id"`
	CompletedAt   *time.Time `json:"completed_at"`
```

Add these fields to the `UpdateTransferOrder` struct (after `RejectionReason *string` on line 51):

```go
	ClaimedByID   *uuid.UUID `json:"claimed_by_id,omitempty"`
	ClaimedAt     *time.Time `json:"claimed_at,omitempty"`
	CompletedByID *uuid.UUID `json:"completed_by_id,omitempty"`
	CompletedAt   *time.Time `json:"completed_at,omitempty"`
```

- [ ] **Step 2.3** Update the `Update` method in `transferorderbus.go` to handle the new fields

Add these blocks inside the `Update` method (after the `ut.Status` block around line 137):

```go
	if ut.ClaimedByID != nil {
		to.ClaimedByID = ut.ClaimedByID
	}
	if ut.ClaimedAt != nil {
		to.ClaimedAt = ut.ClaimedAt
	}
	if ut.CompletedByID != nil {
		to.CompletedByID = ut.CompletedByID
	}
	if ut.CompletedAt != nil {
		to.CompletedAt = ut.CompletedAt
	}
```

- [ ] **Step 2.4** Add `Claim` method to `transferorderbus.go`

Add after the `Reject` method (after line 262):

```go
// Claim marks an approved transfer order as in_transit, recording who claimed it.
func (b *Business) Claim(ctx context.Context, to TransferOrder, claimedBy uuid.UUID) (TransferOrder, error) {
	ctx, span := otel.AddSpan(ctx, "business.transferorderbus.claim")
	defer span.End()

	if to.Status != StatusApproved {
		return TransferOrder{}, fmt.Errorf("claim: %w: must be approved, got %s", ErrInvalidTransferStatus, to.Status)
	}

	before := to

	now := time.Now()
	to.ClaimedByID = &claimedBy
	to.ClaimedAt = &now
	to.Status = StatusInTransit
	to.UpdatedDate = now

	if err := b.storer.Update(ctx, to); err != nil {
		return TransferOrder{}, fmt.Errorf("claim: %w", err)
	}

	if err := b.delegate.Call(ctx, ActionUpdatedData(before, to)); err != nil {
		b.log.Error(ctx, "transferorderbus: delegate call failed", "action", ActionUpdated, "err", err)
	}

	return to, nil
}
```

- [ ] **Step 2.5** Add `Execute` method to `transferorderbus.go`

Add after the `Claim` method:

```go
// Execute marks an in_transit transfer order as completed, recording who completed it.
// This is a simple status transition — the atomic stock move happens at the app layer.
func (b *Business) Execute(ctx context.Context, to TransferOrder, completedBy uuid.UUID) (TransferOrder, error) {
	ctx, span := otel.AddSpan(ctx, "business.transferorderbus.execute")
	defer span.End()

	if to.Status != StatusInTransit {
		return TransferOrder{}, fmt.Errorf("execute: %w: must be in_transit, got %s", ErrInvalidTransferStatus, to.Status)
	}

	before := to

	now := time.Now()
	to.CompletedByID = &completedBy
	to.CompletedAt = &now
	to.Status = StatusCompleted
	to.UpdatedDate = now

	if err := b.storer.Update(ctx, to); err != nil {
		return TransferOrder{}, fmt.Errorf("execute: %w", err)
	}

	if err := b.delegate.Call(ctx, ActionUpdatedData(before, to)); err != nil {
		b.log.Error(ctx, "transferorderbus: delegate call failed", "action", ActionUpdated, "err", err)
	}

	return to, nil
}
```

- [ ] **Step 2.6** Verify build

```bash
go build ./business/domain/inventory/transferorderbus/...
```

---

## Phase 3: DB layer — Add new columns to store

- [ ] **Step 3.1** Add new fields to DB model in `business/domain/inventory/transferorderbus/stores/transferorderdb/model.go`

Add to the `transferOrder` struct (after `RejectionReason string` on line 19):

```go
	ClaimedByID   uuid.NullUUID `db:"claimed_by"`
	ClaimedAt     sql.NullTime  `db:"claimed_at"`
	CompletedByID uuid.NullUUID `db:"completed_by"`
	CompletedAt   sql.NullTime  `db:"completed_at"`
```

Add `"database/sql"` to the imports.

- [ ] **Step 3.2** Update `toBusTransferOrder` in `model.go`

Add these blocks after the `RejectedByID` handling (after line 49):

```go
	if db.ClaimedByID.Valid {
		to.ClaimedByID = &db.ClaimedByID.UUID
	}
	if db.ClaimedAt.Valid {
		t := db.ClaimedAt.Time
		to.ClaimedAt = &t
	}
	if db.CompletedByID.Valid {
		to.CompletedByID = &db.CompletedByID.UUID
	}
	if db.CompletedAt.Valid {
		t := db.CompletedAt.Time
		to.CompletedAt = &t
	}
```

- [ ] **Step 3.3** Update `toDBTransferOrder` in `model.go`

Add these blocks after the `RejectedByID` handling (after line 84):

```go
	if bus.ClaimedByID != nil {
		db.ClaimedByID = uuid.NullUUID{UUID: *bus.ClaimedByID, Valid: true}
	}
	if bus.ClaimedAt != nil {
		db.ClaimedAt = sql.NullTime{Time: *bus.ClaimedAt, Valid: true}
	}
	if bus.CompletedByID != nil {
		db.CompletedByID = uuid.NullUUID{UUID: *bus.CompletedByID, Valid: true}
	}
	if bus.CompletedAt != nil {
		db.CompletedAt = sql.NullTime{Time: *bus.CompletedAt, Valid: true}
	}
```

- [ ] **Step 3.4** Update SQL in `transferorderdb.go`

**Create query** (line 44): Add `claimed_by, claimed_at, completed_by, completed_at` to both the column list and the VALUES list:

```sql
INSERT INTO inventory.transfer_orders (
    id, product_id, from_location_id, to_location_id, requested_by,
    approved_by, rejected_by_id, approval_reason, rejection_reason,
    claimed_by, claimed_at, completed_by, completed_at,
    quantity, status, transfer_date, created_date, updated_date
) VALUES (
    :id, :product_id, :from_location_id, :to_location_id, :requested_by,
    :approved_by, :rejected_by_id, :approval_reason, :rejection_reason,
    :claimed_by, :claimed_at, :completed_by, :completed_at,
    :quantity, :status, :transfer_date, :created_date, :updated_date
)
```

**Update query** (line 70): Add to the SET clause:

```sql
claimed_by = :claimed_by,
claimed_at = :claimed_at,
completed_by = :completed_by,
completed_at = :completed_at,
```

**Query/QueryByID SELECT lists** (lines 123 and 179): Add the four new columns to the SELECT:

```sql
id, product_id, from_location_id, to_location_id, requested_by, approved_by,
rejected_by_id, approval_reason, rejection_reason,
claimed_by, claimed_at, completed_by, completed_at,
quantity, status, transfer_date, created_date, updated_date
```

- [ ] **Step 3.5** Verify build

```bash
go build ./business/domain/inventory/transferorderbus/...
```

---

## Phase 4: Inventory item bus — Add DecrementQuantity

The existing `UpsertQuantity` rejects non-positive deltas. We need a `DecrementQuantity` method for the source location.

- [ ] **Step 4.1** Add `DecrementQuantity` to the `Storer` interface in `business/domain/inventory/inventoryitembus/inventoryitembus.go`

Add after `UpsertQuantity` (line 36):

```go
	DecrementQuantity(ctx context.Context, productID, locationID uuid.UUID, quantity int) error
```

- [ ] **Step 4.2** Add `DecrementQuantity` business method in `inventoryitembus.go`

Add after `UpsertQuantity` (after line 229):

```go
// DecrementQuantity subtracts the given quantity from the inventory item at
// (productID, locationID). Returns an error if the item does not exist or if
// the resulting quantity would go negative.
func (b *Business) DecrementQuantity(ctx context.Context, productID, locationID uuid.UUID, quantity int) error {
	ctx, span := otel.AddSpan(ctx, "business.inventoryitembus.decrementquantity")
	defer span.End()

	if quantity <= 0 {
		return fmt.Errorf("decrement quantity: quantity must be positive, got %d", quantity)
	}

	if err := b.storer.DecrementQuantity(ctx, productID, locationID, quantity); err != nil {
		return fmt.Errorf("decrement quantity: %w", err)
	}

	return nil
}
```

- [ ] **Step 4.3** Add `DecrementQuantity` store method in `business/domain/inventory/inventoryitembus/stores/inventoryitemdb/inventoryitemdb.go`

Add after the `UpsertQuantity` method (after line 410):

```go
// DecrementQuantity atomically subtracts quantity from the inventory item at
// (product_id, location_id). The UPDATE uses a WHERE guard to prevent the
// quantity from going negative. If no row is updated (item missing or
// insufficient stock), an ErrDBNotFound is returned.
func (s *Store) DecrementQuantity(ctx context.Context, productID, locationID uuid.UUID, quantity int) error {
	data := struct {
		ProductID  uuid.UUID `db:"product_id"`
		LocationID uuid.UUID `db:"location_id"`
		Quantity   int       `db:"quantity"`
	}{
		ProductID:  productID,
		LocationID: locationID,
		Quantity:   quantity,
	}

	const q = `
	UPDATE inventory.inventory_items
	SET
		quantity     = quantity - :quantity,
		updated_date = NOW()
	WHERE
		product_id  = :product_id
		AND location_id = :location_id
		AND quantity >= :quantity
	`

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, data); err != nil {
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}
```

**Note:** `NamedExecContext` does NOT check rows-affected by default. If zero rows are affected (insufficient stock or missing item), the operation silently succeeds. To enforce the guard, we need to check rows affected. Look at how `sqldb.NamedExecContext` works — if it does NOT return rows-affected, we may need to use raw `sqlx.NamedExecContext` and check `result.RowsAffected()`. Verify this during implementation and add a `ErrInsufficientStock` error to `inventoryitembus` if needed.

**Alternative approach** (simpler, recommended): Instead of adding `DecrementQuantity` to the Storer interface, use a raw SQL approach in the app-layer `execute()` method via the transaction directly. However, this violates layer boundaries. The cleaner path is the Storer method above.

- [ ] **Step 4.4** Verify build

```bash
go build ./business/domain/inventory/inventoryitembus/...
```

---

## Phase 5: App layer — Extend App struct, add Claim/Execute methods

- [ ] **Step 5.1** Update `App` struct and constructors in `app/domain/inventory/transferorderapp/transferorderapp.go`

Replace the struct and constructors:

```go
type App struct {
	transferorderbus *transferorderbus.Business
	invTransactionBus *inventorytransactionbus.Business
	invItemBus        *inventoryitembus.Business
	db                *sqlx.DB
	auth              *auth.Auth
}

func NewApp(
	transferorderbus *transferorderbus.Business,
	invTransactionBus *inventorytransactionbus.Business,
	invItemBus *inventoryitembus.Business,
	db *sqlx.DB,
) *App {
	return &App{
		transferorderbus:  transferorderbus,
		invTransactionBus: invTransactionBus,
		invItemBus:        invItemBus,
		db:                db,
	}
}

func NewAppWithAuth(
	transferorderbus *transferorderbus.Business,
	invTransactionBus *inventorytransactionbus.Business,
	invItemBus *inventoryitembus.Business,
	db *sqlx.DB,
	auth *auth.Auth,
) *App {
	return &App{
		transferorderbus:  transferorderbus,
		invTransactionBus: invTransactionBus,
		invItemBus:        invItemBus,
		db:                db,
		auth:              auth,
	}
}
```

Add imports:
```go
"database/sql"
"time"

"github.com/jmoiron/sqlx"
"github.com/timmaaaz/ichor/business/domain/inventory/inventoryitembus"
"github.com/timmaaaz/ichor/business/domain/inventory/inventorytransactionbus"
```

- [ ] **Step 5.2** Add `Claim` method to `transferorderapp.go`

Add after the `Reject` method:

```go
func (a *App) Claim(ctx context.Context, id uuid.UUID) (TransferOrder, error) {
	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return TransferOrder{}, errs.New(errs.Unauthenticated, err)
	}

	to, err := a.transferorderbus.QueryByID(ctx, id)
	if err != nil {
		if errors.Is(err, transferorderbus.ErrNotFound) {
			return TransferOrder{}, errs.New(errs.NotFound, err)
		}
		return TransferOrder{}, fmt.Errorf("claim [querybyid]: %w", err)
	}

	claimed, err := a.transferorderbus.Claim(ctx, to, userID)
	if err != nil {
		if errors.Is(err, transferorderbus.ErrInvalidTransferStatus) {
			return TransferOrder{}, errs.New(errs.FailedPrecondition, err)
		}
		return TransferOrder{}, fmt.Errorf("claim: %w", err)
	}

	return ToAppTransferOrder(claimed), nil
}
```

- [ ] **Step 5.3** Add `Execute` method to `transferorderapp.go`

This is the atomic 4-way write following the `putawaytaskapp.complete()` pattern:

```go
// Execute atomically completes a transfer: marks the order completed, creates
// two inventory transactions (TRANSFER_OUT and TRANSFER_IN), decrements source
// location stock, and increments destination location stock.
func (a *App) Execute(ctx context.Context, id uuid.UUID) (TransferOrder, error) {
	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return TransferOrder{}, errs.New(errs.Unauthenticated, err)
	}

	to, err := a.transferorderbus.QueryByID(ctx, id)
	if err != nil {
		if errors.Is(err, transferorderbus.ErrNotFound) {
			return TransferOrder{}, errs.New(errs.NotFound, err)
		}
		return TransferOrder{}, fmt.Errorf("execute [querybyid]: %w", err)
	}

	if to.Status != transferorderbus.StatusInTransit {
		return TransferOrder{}, errs.Newf(errs.FailedPrecondition,
			"transfer must be in_transit, got %s", to.Status)
	}

	tx, err := a.db.BeginTxx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return TransferOrder{}, fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	// 1. Mark transfer order as completed.
	toBusTx, err := a.transferorderbus.NewWithTx(tx)
	if err != nil {
		return TransferOrder{}, fmt.Errorf("new transferorder tx: %w", err)
	}

	completed, err := toBusTx.Execute(ctx, to, userID)
	if err != nil {
		if errors.Is(err, transferorderbus.ErrInvalidTransferStatus) {
			return TransferOrder{}, errs.New(errs.FailedPrecondition, err)
		}
		return TransferOrder{}, fmt.Errorf("execute transfer: %w", err)
	}

	// 2. Create TRANSFER_OUT inventory transaction (source location).
	txBusTx, err := a.invTransactionBus.NewWithTx(tx)
	if err != nil {
		return TransferOrder{}, fmt.Errorf("new invtransaction tx: %w", err)
	}

	refNum := to.TransferID.String()
	now := time.Now()

	_, err = txBusTx.Create(ctx, inventorytransactionbus.NewInventoryTransaction{
		ProductID:       to.ProductID,
		LocationID:      to.FromLocationID,
		UserID:          userID,
		Quantity:        to.Quantity,
		TransactionType: "TRANSFER_OUT",
		ReferenceNumber: refNum,
		TransactionDate: now,
	})
	if err != nil {
		return TransferOrder{}, fmt.Errorf("create transfer_out transaction: %w", err)
	}

	// 3. Create TRANSFER_IN inventory transaction (destination location).
	_, err = txBusTx.Create(ctx, inventorytransactionbus.NewInventoryTransaction{
		ProductID:       to.ProductID,
		LocationID:      to.ToLocationID,
		UserID:          userID,
		Quantity:        to.Quantity,
		TransactionType: "TRANSFER_IN",
		ReferenceNumber: refNum,
		TransactionDate: now,
	})
	if err != nil {
		return TransferOrder{}, fmt.Errorf("create transfer_in transaction: %w", err)
	}

	// 4. Decrement source location stock.
	itemBusTx, err := a.invItemBus.NewWithTx(tx)
	if err != nil {
		return TransferOrder{}, fmt.Errorf("new invitem tx: %w", err)
	}

	if err := itemBusTx.DecrementQuantity(ctx, to.ProductID, to.FromLocationID, to.Quantity); err != nil {
		return TransferOrder{}, errs.Newf(errs.FailedPrecondition, "insufficient stock at source location: %s", err)
	}

	// 5. Increment destination location stock.
	if err := itemBusTx.UpsertQuantity(ctx, to.ProductID, to.ToLocationID, to.Quantity); err != nil {
		return TransferOrder{}, fmt.Errorf("upsert destination inventory: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return TransferOrder{}, fmt.Errorf("commit transaction: %w", err)
	}

	return ToAppTransferOrder(completed), nil
}
```

- [ ] **Step 5.4** Update app model `TransferOrder` in `app/domain/inventory/transferorderapp/model.go`

Add fields to the app `TransferOrder` struct (after `RejectionReason` on line 41):

```go
	ClaimedByID   string `json:"claimed_by_id"`
	ClaimedAt     string `json:"claimed_at"`
	CompletedByID string `json:"completed_by_id"`
	CompletedAt   string `json:"completed_at"`
```

- [ ] **Step 5.5** Update `ToAppTransferOrder` in `model.go`

Add handling for the new fields (after the `rejectedByID` handling, before the return):

```go
	claimedByID := ""
	if bus.ClaimedByID != nil {
		claimedByID = bus.ClaimedByID.String()
	}

	claimedAt := ""
	if bus.ClaimedAt != nil {
		claimedAt = bus.ClaimedAt.Format(timeutil.FORMAT)
	}

	completedByID := ""
	if bus.CompletedByID != nil {
		completedByID = bus.CompletedByID.String()
	}

	completedAt := ""
	if bus.CompletedAt != nil {
		completedAt = bus.CompletedAt.Format(timeutil.FORMAT)
	}
```

And add these to the returned struct literal:

```go
	ClaimedByID:   claimedByID,
	ClaimedAt:     claimedAt,
	CompletedByID: completedByID,
	CompletedAt:   completedAt,
```

- [ ] **Step 5.6** Verify build

```bash
go build ./app/domain/inventory/transferorderapp/...
```

---

## Phase 6: API layer — Add claim/execute routes and handlers

- [ ] **Step 6.1** Update `Config` in `api/domain/http/inventory/transferorderapi/routes.go`

Add the new dependencies to the Config struct:

```go
type Config struct {
	Log               *logger.Logger
	TransferOrderBus  *transferorderbus.Business
	InvTransactionBus *inventorytransactionbus.Business
	InvItemBus        *inventoryitembus.Business
	DB                *sqlx.DB
	AuthClient        *authclient.Client
	PermissionsBus    *permissionsbus.Business
}
```

Add imports:

```go
"github.com/jmoiron/sqlx"
"github.com/timmaaaz/ichor/business/domain/inventory/inventoryitembus"
"github.com/timmaaaz/ichor/business/domain/inventory/inventorytransactionbus"
```

- [ ] **Step 6.2** Update `Routes` function to pass new deps and add new routes

Update the `api` construction line:

```go
api := newAPI(transferorderapp.NewApp(
    cfg.TransferOrderBus,
    cfg.InvTransactionBus,
    cfg.InvItemBus,
    cfg.DB,
))
```

Add two new routes (after the reject route):

```go
app.HandlerFunc(http.MethodPost, version, "/inventory/transfer-orders/{transfer_id}/claim", api.claim, authen,
    mid.Authorize(cfg.AuthClient, cfg.PermissionsBus, RouteTable, permissionsbus.Actions.Update, auth.RuleAny))

app.HandlerFunc(http.MethodPost, version, "/inventory/transfer-orders/{transfer_id}/execute", api.execute, authen,
    mid.Authorize(cfg.AuthClient, cfg.PermissionsBus, RouteTable, permissionsbus.Actions.Update, auth.RuleAny))
```

- [ ] **Step 6.3** Add `claim` and `execute` handler methods in `transferorderapi.go`

Add after the `reject` method (after line 128):

```go
func (api *api) claim(ctx context.Context, r *http.Request) web.Encoder {
	toID := web.Param(r, "transfer_id")
	parsed, err := uuid.Parse(toID)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	to, err := api.transferorderapp.Claim(ctx, parsed)
	if err != nil {
		return errs.NewError(err)
	}

	return to
}

func (api *api) execute(ctx context.Context, r *http.Request) web.Encoder {
	toID := web.Param(r, "transfer_id")
	parsed, err := uuid.Parse(toID)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	to, err := api.transferorderapp.Execute(ctx, parsed)
	if err != nil {
		return errs.NewError(err)
	}

	return to
}
```

- [ ] **Step 6.4** Verify build

```bash
go build ./api/domain/http/inventory/transferorderapi/...
```

---

## Phase 7: Wiring — Update all.go

- [ ] **Step 7.1** Update `transferorderapi.Routes` call in `api/cmd/services/ichor/build/all/all.go`

Replace lines 1064-1069:

```go
transferorderapi.Routes(app, transferorderapi.Config{
    Log:               cfg.Log,
    TransferOrderBus:  transferOrderBus,
    InvTransactionBus: inventoryTransactionBus,
    InvItemBus:        inventoryItemBus,
    DB:                cfg.DB,
    AuthClient:        cfg.AuthClient,
    PermissionsBus:    permissionsBus,
})
```

Add imports if not already present:

```go
"github.com/timmaaaz/ichor/business/domain/inventory/inventoryitembus"
"github.com/timmaaaz/ichor/business/domain/inventory/inventorytransactionbus"
```

(These imports likely already exist from putawaytaskapi wiring.)

- [ ] **Step 7.2** Verify full build

```bash
go build ./api/cmd/services/ichor/...
```

---

## Phase 8: Tests

- [ ] **Step 8.1** Add unit tests for `Claim` and `Execute` status guards in `business/domain/inventory/transferorderbus/transferorderbus_test.go`

Test cases:
- `Claim` on a `pending` order returns `ErrInvalidTransferStatus`
- `Claim` on a `rejected` order returns `ErrInvalidTransferStatus`
- `Claim` on an `approved` order succeeds, status becomes `in_transit`
- `Execute` on an `approved` order returns `ErrInvalidTransferStatus`
- `Execute` on an `in_transit` order succeeds, status becomes `completed`

- [ ] **Step 8.2** Run business layer tests

```bash
go test ./business/domain/inventory/transferorderbus/...
```

- [ ] **Step 8.3** Run inventory item bus tests (for DecrementQuantity)

```bash
go test ./business/domain/inventory/inventoryitembus/...
```

- [ ] **Step 8.4** Check for hardcoded route/tool counts in test files

Search for any test assertions that count routes or tools for transferorderapi and update them to account for the 2 new endpoints.

```bash
grep -rn "transfer.order" api/cmd/services/ichor/tests/ --include="*.go" | head -20
```

---

## File Change Summary

| File | Change |
|------|--------|
| `business/sdk/migrate/sql/migrate.sql` | Add v2.17: 4 new columns |
| `business/domain/inventory/transferorderbus/transferorderbus.go` | Add statuses, `Claim()`, `Execute()` |
| `business/domain/inventory/transferorderbus/model.go` | Add 4 fields to `TransferOrder` + `UpdateTransferOrder` |
| `business/domain/inventory/transferorderbus/stores/transferorderdb/model.go` | Add 4 DB fields, update converters |
| `business/domain/inventory/transferorderbus/stores/transferorderdb/transferorderdb.go` | Update all SQL queries |
| `business/domain/inventory/inventoryitembus/inventoryitembus.go` | Add `DecrementQuantity` to Storer + Business |
| `business/domain/inventory/inventoryitembus/stores/inventoryitemdb/inventoryitemdb.go` | Add `DecrementQuantity` store method |
| `app/domain/inventory/transferorderapp/transferorderapp.go` | Extend `App`, add `Claim()`, `Execute()` (atomic) |
| `app/domain/inventory/transferorderapp/model.go` | Add 4 app fields, update `ToAppTransferOrder` |
| `api/domain/http/inventory/transferorderapi/routes.go` | Extend `Config`, add 2 routes |
| `api/domain/http/inventory/transferorderapi/transferorderapi.go` | Add `claim()`, `execute()` handlers |
| `api/cmd/services/ichor/build/all/all.go` | Update wiring with new deps |
| `business/domain/inventory/transferorderbus/transferorderbus_test.go` | Add Claim/Execute guard tests |

---

## Status Machine

```
pending ──approve──→ approved ──claim──→ in_transit ──execute──→ completed
   │
   └──reject──→ rejected (terminal)
```

- `Claim`: `approved` → `in_transit` (simple status update + claimed_by/claimed_at)
- `Execute`: `in_transit` → `completed` (atomic: status + 2 transactions + source decrement + dest increment)
- `completed` and `rejected` are terminal states
