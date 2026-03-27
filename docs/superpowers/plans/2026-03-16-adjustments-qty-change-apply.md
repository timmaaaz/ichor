# Adjustments Qty Change Apply Implementation Plan

> **For agentic workers:** REQUIRED: Use superpowers:subagent-driven-development (if subagents available) or superpowers:executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.
>
> **Worktree:** Create a worktree before executing: `create a worktree for blocker-003-adjustments and execute this plan`

**Goal:** Make adjustment approval actually mutate inventory by adding a 3-way atomic write (update adjustment status + create inventory transaction + upsert inventory item quantity).

**Architecture:** Extend the existing `inventoryadjustmentapp.App` struct with `invTransactionBus`, `invItemBus`, and `db *sqlx.DB`. The `Approve` method runs a 3-way atomic write inside a ReadCommitted transaction, following the exact `putawaytaskapp.complete()` pattern. Add an `AdjustQuantity` method to `inventoryitembus` to handle both positive and negative deltas.

**Tech Stack:** Go 1.23, PostgreSQL, Ardan Labs service architecture

---

## Step 1: Add `AdjustQuantity` to inventory item store

The existing `UpsertQuantity` rejects negative deltas (`quantityDelta <= 0`). The SQL itself already handles negatives correctly (`quantity + EXCLUDED.quantity`), so we add a parallel method that accepts any non-zero delta.

- [ ] **1a. Add `AdjustQuantity` to the Storer interface**

File: `business/domain/inventory/inventoryitembus/inventoryitembus.go` (line 25-37)

Add to the `Storer` interface:

```go
AdjustQuantity(ctx context.Context, newID, productID, locationID uuid.UUID, quantityDelta int) error
```

- [ ] **1b. Add `AdjustQuantity` to the DB store**

File: `business/domain/inventory/inventoryitembus/stores/inventoryitemdb/inventoryitemdb.go`

Add after `UpsertQuantity` (after line ~410):

```go
// AdjustQuantity atomically creates or updates the inventory item for the given
// (product_id, location_id) pair, adding quantityDelta (positive or negative)
// to the existing quantity. Used for inventory adjustments where shrinkage or
// damage may produce a negative delta.
func (s *Store) AdjustQuantity(ctx context.Context, newID, productID, locationID uuid.UUID, quantityDelta int) error {
	data := struct {
		ID         uuid.UUID `db:"id"`
		ProductID  uuid.UUID `db:"product_id"`
		LocationID uuid.UUID `db:"location_id"`
		Quantity   int       `db:"quantity"`
	}{
		ID:         newID,
		ProductID:  productID,
		LocationID: locationID,
		Quantity:   quantityDelta,
	}

	const q = `
	INSERT INTO inventory.inventory_items
		(id, product_id, location_id, quantity,
		 reserved_quantity, allocated_quantity,
		 minimum_stock, maximum_stock, reorder_point,
		 economic_order_quantity, safety_stock, avg_daily_usage,
		 created_date, updated_date)
	VALUES
		(:id, :product_id, :location_id, :quantity,
		 0, 0, 0, 0, 0, 0, 0, 0,
		 NOW(), NOW())
	ON CONFLICT (product_id, location_id)
	DO UPDATE SET
		quantity     = inventory.inventory_items.quantity + EXCLUDED.quantity,
		updated_date = NOW()
	`

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, data); err != nil {
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}
```

- [ ] **1c. Add `AdjustQuantity` business method**

File: `business/domain/inventory/inventoryitembus/inventoryitembus.go`

Add after `UpsertQuantity` (after line ~229):

```go
// AdjustQuantity creates or updates the inventory item for (productID, locationID),
// adding quantityDelta which may be positive or negative. Zero is rejected.
func (b *Business) AdjustQuantity(ctx context.Context, productID, locationID uuid.UUID, quantityDelta int) error {
	ctx, span := otel.AddSpan(ctx, "business.inventoryitembus.adjustquantity")
	defer span.End()

	if quantityDelta == 0 {
		return fmt.Errorf("adjust quantity: quantityDelta must be non-zero")
	}

	if err := b.storer.AdjustQuantity(ctx, uuid.New(), productID, locationID, quantityDelta); err != nil {
		return fmt.Errorf("adjust quantity: %w", err)
	}

	return nil
}
```

- [ ] **1d. Build and test**

```bash
go build ./business/domain/inventory/inventoryitembus/...
go build ./business/domain/inventory/inventoryitembus/stores/inventoryitemdb/...
```

**Commit:** `feat(inventory): add AdjustQuantity method to inventoryitembus for positive/negative deltas`

---

## Step 2: Expand `inventoryadjustmentapp.App` struct with transaction dependencies

- [ ] **2a. Update App struct and constructors**

File: `app/domain/inventory/inventoryadjustmentapp/inventoryadjustmentapp.go`

Replace the App struct (lines 18-21) and both constructors (lines 24-36):

```go
type App struct {
	inventoryadjustmentbus *inventoryadjustmentbus.Business
	invTransactionBus      *inventorytransactionbus.Business
	invItemBus             *inventoryitembus.Business
	db                     *sqlx.DB
	auth                   *auth.Auth
}

// NewApp constructs an inventory adjustment app API for use.
func NewApp(
	inventoryadjustmentbus *inventoryadjustmentbus.Business,
	invTransactionBus *inventorytransactionbus.Business,
	invItemBus *inventoryitembus.Business,
	db *sqlx.DB,
) *App {
	return &App{
		inventoryadjustmentbus: inventoryadjustmentbus,
		invTransactionBus:      invTransactionBus,
		invItemBus:             invItemBus,
		db:                     db,
	}
}

// NewAppWithAuth constructs an inventory adjustment app API for use with auth support.
func NewAppWithAuth(
	inventoryadjustmentbus *inventoryadjustmentbus.Business,
	invTransactionBus *inventorytransactionbus.Business,
	invItemBus *inventoryitembus.Business,
	db *sqlx.DB,
	ath *auth.Auth,
) *App {
	return &App{
		inventoryadjustmentbus: inventoryadjustmentbus,
		invTransactionBus:      invTransactionBus,
		invItemBus:             invItemBus,
		db:                     db,
		auth:                   ath,
	}
}
```

Add these imports to the file:

```go
"database/sql"
"time"

"github.com/jmoiron/sqlx"
"github.com/timmaaaz/ichor/business/domain/inventory/inventoryitembus"
"github.com/timmaaaz/ichor/business/domain/inventory/inventorytransactionbus"
```

- [ ] **2b. Build to confirm struct changes compile**

```bash
go build ./app/domain/inventory/inventoryadjustmentapp/...
```

This will fail because callers pass the old signature. That is expected — fixed in Step 3.

**Commit:** `refactor(inventory): expand inventoryadjustmentapp.App with transaction dependencies`

---

## Step 3: Update all callers of `NewApp` / `NewAppWithAuth`

There are exactly 3 call sites that must be updated.

- [ ] **3a. Update API routes constructor**

File: `api/domain/http/inventory/inventoryadjustmentapi/routes.go`

Update the `Config` struct (lines 16-21) to add the new dependencies:

```go
type Config struct {
	Log                    *logger.Logger
	InventoryAdjustmentBus *inventoryadjustmentbus.Business
	InvTransactionBus      *inventorytransactionbus.Business
	InvItemBus             *inventoryitembus.Business
	DB                     *sqlx.DB
	AuthClient             *authclient.Client
	PermissionsBus         *permissionsbus.Business
}
```

Add imports:

```go
"github.com/jmoiron/sqlx"
"github.com/timmaaaz/ichor/business/domain/inventory/inventoryitembus"
"github.com/timmaaaz/ichor/business/domain/inventory/inventorytransactionbus"
```

Update the `newAPI` call inside `Routes` (line 31):

```go
api := newAPI(inventoryadjustmentapp.NewApp(cfg.InventoryAdjustmentBus, cfg.InvTransactionBus, cfg.InvItemBus, cfg.DB))
```

- [ ] **3b. Update all.go API wiring**

File: `api/cmd/services/ichor/build/all/all.go` (lines 1047-1052)

Update the `inventoryadjustmentapi.Routes` call:

```go
inventoryadjustmentapi.Routes(app, inventoryadjustmentapi.Config{
	InventoryAdjustmentBus: inventoryAdjustmentBus,
	InvTransactionBus:      inventoryTransactionBus,
	InvItemBus:             inventoryItemBus,
	DB:                     cfg.DB,
	AuthClient:             cfg.AuthClient,
	Log:                    cfg.Log,
	PermissionsBus:         permissionsBus,
})
```

- [ ] **3c. Update all.go formdata registry call**

File: `api/cmd/services/ichor/build/all/all.go` (line 1423)

Update the `inventoryadjustmentapp.NewApp` call in the formdata registry:

```go
inventoryadjustmentapp.NewApp(inventoryAdjustmentBus, inventoryTransactionBus, inventoryItemBus, cfg.DB),
```

- [ ] **3d. Build to confirm all callers compile**

```bash
go build ./api/domain/http/inventory/inventoryadjustmentapi/...
go build ./api/cmd/services/ichor/...
```

Expected: clean build, no errors.

**Commit:** `refactor(inventory): update all NewApp callers for new adjustment app signature`

---

## Step 4: Implement the 3-way atomic write in `Approve`

- [ ] **4a. Replace the `Approve` method**

File: `app/domain/inventory/inventoryadjustmentapp/inventoryadjustmentapp.go`

Replace the existing `Approve` method (lines 134-157) with:

```go
// Approve marks an adjustment as approved and atomically applies the quantity
// change to inventory. This is a 3-way atomic write:
//  1. Update adjustment status to approved
//  2. Create an ADJUSTMENT inventory transaction (ledger entry)
//  3. Adjust inventory_item quantity at the adjustment's location
func (a *App) Approve(ctx context.Context, id uuid.UUID) (InventoryAdjustment, error) {
	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return InventoryAdjustment{}, errs.New(errs.Unauthenticated, err)
	}

	ia, err := a.inventoryadjustmentbus.QueryByID(ctx, id)
	if err != nil {
		if errors.Is(err, inventoryadjustmentbus.ErrNotFound) {
			return InventoryAdjustment{}, errs.New(errs.NotFound, err)
		}
		return InventoryAdjustment{}, fmt.Errorf("approve [querybyid]: %w", err)
	}

	if ia.ApprovalStatus != inventoryadjustmentbus.ApprovalStatusPending {
		return InventoryAdjustment{}, errs.New(errs.FailedPrecondition, inventoryadjustmentbus.ErrInvalidApprovalStatus)
	}

	// --- Begin atomic 3-way write ---

	tx, err := a.db.BeginTxx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return InventoryAdjustment{}, fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	// 1. Update adjustment status to approved.
	adjBusTx, err := a.inventoryadjustmentbus.NewWithTx(tx)
	if err != nil {
		return InventoryAdjustment{}, fmt.Errorf("new adjustment tx: %w", err)
	}

	updatedIA, err := adjBusTx.Approve(ctx, ia, userID, "")
	if err != nil {
		if errors.Is(err, inventoryadjustmentbus.ErrInvalidApprovalStatus) {
			return InventoryAdjustment{}, errs.New(errs.FailedPrecondition, err)
		}
		return InventoryAdjustment{}, fmt.Errorf("approve adjustment: %w", err)
	}

	// 2. Create ADJUSTMENT inventory transaction (ledger entry).
	txBusTx, err := a.invTransactionBus.NewWithTx(tx)
	if err != nil {
		return InventoryAdjustment{}, fmt.Errorf("new invtransaction tx: %w", err)
	}

	_, err = txBusTx.Create(ctx, inventorytransactionbus.NewInventoryTransaction{
		ProductID:       ia.ProductID,
		LocationID:      ia.LocationID,
		UserID:          userID,
		Quantity:        ia.QuantityChange,
		TransactionType: "ADJUSTMENT",
		ReferenceNumber: ia.InventoryAdjustmentID.String(),
		TransactionDate: time.Now(),
	})
	if err != nil {
		return InventoryAdjustment{}, fmt.Errorf("create inventory transaction: %w", err)
	}

	// 3. Adjust inventory_item quantity at the adjustment location.
	itemBusTx, err := a.invItemBus.NewWithTx(tx)
	if err != nil {
		return InventoryAdjustment{}, fmt.Errorf("new invitem tx: %w", err)
	}

	if err := itemBusTx.AdjustQuantity(ctx, ia.ProductID, ia.LocationID, ia.QuantityChange); err != nil {
		return InventoryAdjustment{}, fmt.Errorf("adjust inventory quantity: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return InventoryAdjustment{}, fmt.Errorf("commit transaction: %w", err)
	}

	return ToAppInventoryAdjustment(updatedIA), nil
}
```

- [ ] **4b. Build and verify**

```bash
go build ./app/domain/inventory/inventoryadjustmentapp/...
go build ./api/cmd/services/ichor/...
```

Expected: clean build.

**Commit:** `feat(inventory): implement 3-way atomic write on adjustment approval`

---

## Step 5: Run tests for all affected packages

- [ ] **5a. Run inventoryitembus tests**

```bash
go test ./business/domain/inventory/inventoryitembus/...
```

- [ ] **5b. Run inventoryadjustmentbus tests**

```bash
go test ./business/domain/inventory/inventoryadjustmentbus/...
```

- [ ] **5c. Run inventoryadjustmentapi integration tests (if they exist)**

```bash
go test ./api/cmd/services/ichor/tests/inventory/inventoryadjustmentapi/...
```

- [ ] **5d. Run inventorytransactionbus tests**

```bash
go test ./business/domain/inventory/inventorytransactionbus/...
```

- [ ] **5e. Fix any test failures**

If integration tests assert against the old `Approve` behavior (no inventory mutation), update them to verify the new behavior: after approval, an `ADJUSTMENT` transaction should exist and inventory quantity should reflect the delta.

**Commit (if needed):** `fix(inventory): update tests for adjustment approval atomic write`

---

## Summary of Files Changed

| File | Change |
|---|---|
| `business/domain/inventory/inventoryitembus/inventoryitembus.go` | Add `AdjustQuantity` method, add to `Storer` interface |
| `business/domain/inventory/inventoryitembus/stores/inventoryitemdb/inventoryitemdb.go` | Add `AdjustQuantity` store implementation |
| `app/domain/inventory/inventoryadjustmentapp/inventoryadjustmentapp.go` | Expand `App` struct, update constructors, rewrite `Approve` with 3-way atomic write |
| `api/domain/http/inventory/inventoryadjustmentapi/routes.go` | Add `InvTransactionBus`, `InvItemBus`, `DB` to `Config`, update `newAPI` call |
| `api/cmd/services/ichor/build/all/all.go` | Update both `NewApp` call sites (API routes + formdata registry) |

## Key Design Decisions

1. **New `AdjustQuantity` vs modifying `UpsertQuantity`:** Adding a separate method preserves the existing `UpsertQuantity` contract (positive-only) used by put-away tasks. `AdjustQuantity` accepts any non-zero delta. Both use the same SQL (the `ON CONFLICT ... quantity + EXCLUDED.quantity` handles negatives natively).

2. **Pre-check in app layer:** The `Approve` method checks `ApprovalStatus == pending` before starting the transaction, then the bus `Approve` rechecks inside the transaction. This avoids opening a transaction for already-approved adjustments (fast fail) while maintaining correctness under concurrency.

3. **Transaction type `"ADJUSTMENT"`:** Distinct from `"PUT_AWAY"` to clearly identify the source of inventory mutations in the ledger.

4. **Reference number:** Uses `ia.InventoryAdjustmentID.String()` to link the transaction back to the adjustment record.
