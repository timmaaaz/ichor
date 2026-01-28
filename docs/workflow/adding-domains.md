# Adding Workflow Events to New Domains

This guide explains how to add workflow event firing to a new domain package.

## Overview

To enable workflow automation for a domain:

1. Create `event.go` with domain event definitions
2. Add `delegate.Call()` to CRUD methods
3. Register the domain in `all.go`

## Step 1: Create event.go

**Location**: `business/domain/{area}/{entity}bus/event.go`

### Template

Replace the placeholders:
- `{PACKAGE}` → Package name (e.g., `customersbus`)
- `{DOMAIN_NAME}` → Delegate routing key, singular (e.g., `"customer"`)
- `{ENTITY_NAME}` → Table name, NOT schema-qualified (e.g., `"customers"`)
- `{EntityType}` → Main struct type from model.go (e.g., `Customer`)
- `{entityVar}` → Variable name (e.g., `customer`)

```go
package {PACKAGE}

import (
	"encoding/json"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/sdk/delegate"
)

// DomainName represents the name of this domain for delegate events.
const DomainName = "{DOMAIN_NAME}"

// EntityName is the workflow entity name used for event matching.
// This should match the entity name in workflow.entities table.
const EntityName = "{ENTITY_NAME}"

// Delegate action constants.
const (
	ActionCreated = "created"
	ActionUpdated = "updated"
	ActionDeleted = "deleted"
)

// =============================================================================
// Created Event
// =============================================================================

// ActionCreatedParms represents the parameters for the created action.
type ActionCreatedParms struct {
	EntityID uuid.UUID   `json:"entityID"`
	UserID   uuid.UUID   `json:"userID"`
	Entity   {EntityType} `json:"entity"`
}

// Marshal returns the event parameters encoded as JSON.
func (p *ActionCreatedParms) Marshal() ([]byte, error) {
	return json.Marshal(p)
}

// ActionCreatedData constructs delegate data for {entityVar} creation events.
func ActionCreatedData({entityVar} {EntityType}) delegate.Data {
	params := ActionCreatedParms{
		EntityID: {entityVar}.ID,
		UserID:   {entityVar}.CreatedBy,
		Entity:   {entityVar},
	}

	rawParams, err := params.Marshal()
	if err != nil {
		panic(err)
	}

	return delegate.Data{
		Domain:    DomainName,
		Action:    ActionCreated,
		RawParams: rawParams,
	}
}

// =============================================================================
// Updated Event
// =============================================================================

// ActionUpdatedParms represents the parameters for the updated action.
type ActionUpdatedParms struct {
	EntityID uuid.UUID   `json:"entityID"`
	UserID   uuid.UUID   `json:"userID"`
	Entity   {EntityType} `json:"entity"`
}

// Marshal returns the event parameters encoded as JSON.
func (p *ActionUpdatedParms) Marshal() ([]byte, error) {
	return json.Marshal(p)
}

// ActionUpdatedData constructs delegate data for {entityVar} update events.
func ActionUpdatedData({entityVar} {EntityType}) delegate.Data {
	params := ActionUpdatedParms{
		EntityID: {entityVar}.ID,
		UserID:   {entityVar}.UpdatedBy,
		Entity:   {entityVar},
	}

	rawParams, err := params.Marshal()
	if err != nil {
		panic(err)
	}

	return delegate.Data{
		Domain:    DomainName,
		Action:    ActionUpdated,
		RawParams: rawParams,
	}
}

// =============================================================================
// Deleted Event
// =============================================================================

// ActionDeletedParms represents the parameters for the deleted action.
type ActionDeletedParms struct {
	EntityID uuid.UUID   `json:"entityID"`
	UserID   uuid.UUID   `json:"userID"`
	Entity   {EntityType} `json:"entity"`
}

// Marshal returns the event parameters encoded as JSON.
func (p *ActionDeletedParms) Marshal() ([]byte, error) {
	return json.Marshal(p)
}

// ActionDeletedData constructs delegate data for {entityVar} deletion events.
func ActionDeletedData({entityVar} {EntityType}) delegate.Data {
	params := ActionDeletedParms{
		EntityID: {entityVar}.ID,
		UserID:   {entityVar}.UpdatedBy,
		Entity:   {entityVar},
	}

	rawParams, err := params.Marshal()
	if err != nil {
		panic(err)
	}

	return delegate.Data{
		Domain:    DomainName,
		Action:    ActionDeleted,
		RawParams: rawParams,
	}
}
```

### Placeholder Reference

| Placeholder | Description | Example (customersbus) | Example (ordersbus) |
|-------------|-------------|------------------------|---------------------|
| `{PACKAGE}` | Go package name | `customersbus` | `ordersbus` |
| `{DOMAIN_NAME}` | Delegate routing key | `"customer"` | `"order"` |
| `{ENTITY_NAME}` | Table name (not schema) | `"customers"` | `"orders"` |
| `{EntityType}` | Main struct from model.go | `Customer` | `Order` |
| `{entityVar}` | Variable name | `customer` | `order` |

### Reference Tables (Without User Tracking)

For lookup tables without `CreatedBy`/`UpdatedBy` fields, use `uuid.Nil`:

```go
// ActionCreatedParms for reference tables.
// Note: UserID is uuid.Nil for system-level operations.
type ActionCreatedParms struct {
	EntityID uuid.UUID              `json:"entityID"`
	UserID   uuid.UUID              `json:"userID"`
	Entity   OrderFulfillmentStatus `json:"entity"`
}

func ActionCreatedData(status OrderFulfillmentStatus) delegate.Data {
	params := ActionCreatedParms{
		EntityID: status.ID,
		UserID:   uuid.Nil, // Reference table - no user tracking
		Entity:   status,
	}
	// ...
}
```

## Step 2: Add delegate.Call() to CRUD Methods

**File**: `business/domain/{area}/{entity}bus/{entity}bus.go`

### Create Method

Add after successful `storer.Create()`:

```go
func (b *Business) Create(ctx context.Context, nc NewCustomer) (Customer, error) {
	customer := Customer{
		ID:        b.delegate.GenerateUUID(),
		// ... field assignments
	}

	if err := b.storer.Create(ctx, customer); err != nil {
		return Customer{}, fmt.Errorf("create: %w", err)
	}

	// Fire delegate event for workflow automation
	if err := b.delegate.Call(ctx, ActionCreatedData(customer)); err != nil {
		b.log.Error(ctx, "customersbus: delegate call failed", "action", ActionCreated, "err", err)
	}

	return customer, nil
}
```

### Update Method

Add after successful `storer.Update()`:

```go
func (b *Business) Update(ctx context.Context, customer Customer, uc UpdateCustomer) (Customer, error) {
	// ... update logic

	if err := b.storer.Update(ctx, customer); err != nil {
		return Customer{}, fmt.Errorf("update: %w", err)
	}

	// Fire delegate event for workflow automation
	if err := b.delegate.Call(ctx, ActionUpdatedData(customer)); err != nil {
		b.log.Error(ctx, "customersbus: delegate call failed", "action", ActionUpdated, "err", err)
	}

	return customer, nil
}
```

### Delete Method

Add after successful `storer.Delete()`:

```go
func (b *Business) Delete(ctx context.Context, customer Customer) error {
	if err := b.storer.Delete(ctx, customer); err != nil {
		return fmt.Errorf("delete: %w", err)
	}

	// Fire delegate event for workflow automation
	if err := b.delegate.Call(ctx, ActionDeletedData(customer)); err != nil {
		b.log.Error(ctx, "customersbus: delegate call failed", "action", ActionDeleted, "err", err)
	}

	return nil
}
```

### Important Rules

1. Add delegate call **AFTER** `storer.Create/Update/Delete()` succeeds
2. Add delegate call **BEFORE** the final `return`
3. **Never** return an error from delegate.Call() - log and continue
4. Use the entity variable **after** modification (has correct IDs, timestamps)

## Step 3: Register in all.go

**File**: `api/cmd/services/ichor/build/all/all.go`

### Add Import

```go
import (
	// ... existing imports
	"github.com/timmaaaz/ichor/business/domain/sales/customersbus"
)
```

### Register Domain

Find the section with `delegateHandler.RegisterDomain()` calls and add:

```go
// Register customersbus domain -> workflow events
delegateHandler.RegisterDomain(delegate, customersbus.DomainName, customersbus.EntityName)
```

## Step 4: Add Entity to workflow.entities

Ensure the entity exists in the `workflow.entities` table. Either:

1. Add via migration:
```sql
INSERT INTO workflow.entities (name, entity_type_id, schema_name)
SELECT 'customers', id, 'sales' FROM workflow.entity_types WHERE name = 'table';
```

2. Or seed via `TestSeedFullWorkflow()` function

## Step 5: Verify

```bash
# Build to check compilation
go build ./...

# Run delegate handler tests
go test -v ./business/sdk/workflow/... -run TestDelegateHandler

# Run all tests
make test
```

## Complete Example: customersbus

### event.go

```go
package customersbus

import (
	"encoding/json"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/sdk/delegate"
)

const DomainName = "customer"
const EntityName = "customers"

const (
	ActionCreated = "created"
	ActionUpdated = "updated"
	ActionDeleted = "deleted"
)

type ActionCreatedParms struct {
	EntityID uuid.UUID `json:"entityID"`
	UserID   uuid.UUID `json:"userID"`
	Entity   Customer  `json:"entity"`
}

func (p *ActionCreatedParms) Marshal() ([]byte, error) {
	return json.Marshal(p)
}

func ActionCreatedData(customer Customer) delegate.Data {
	params := ActionCreatedParms{
		EntityID: customer.ID,
		UserID:   customer.CreatedBy,
		Entity:   customer,
	}
	rawParams, _ := params.Marshal()
	return delegate.Data{
		Domain:    DomainName,
		Action:    ActionCreated,
		RawParams: rawParams,
	}
}

// ... ActionUpdatedData, ActionDeletedData similarly
```

### customersbus.go (Create method)

```go
func (b *Business) Create(ctx context.Context, nc NewCustomer) (Customer, error) {
	ctx, span := otel.AddSpan(ctx, "business.customersbus.create")
	defer span.End()

	customer := Customer{
		ID:          b.delegate.GenerateUUID(),
		Name:        nc.Name,
		Email:       nc.Email,
		CreatedBy:   nc.CreatedBy,
		CreatedDate: b.delegate.Now(),
		UpdatedBy:   nc.CreatedBy,
		UpdatedDate: b.delegate.Now(),
	}

	if err := b.storer.Create(ctx, customer); err != nil {
		return Customer{}, fmt.Errorf("create: %w", err)
	}

	if err := b.delegate.Call(ctx, ActionCreatedData(customer)); err != nil {
		b.log.Error(ctx, "customersbus: delegate call failed", "action", ActionCreated, "err", err)
	}

	return customer, nil
}
```

### all.go

```go
import "github.com/timmaaaz/ichor/business/domain/sales/customersbus"

// In Add() function:
delegateHandler.RegisterDomain(delegate, customersbus.DomainName, customersbus.EntityName)
```

## Domain Implementation Status

Track which domains have workflow events implemented:

### Sales Domain

| Package | Entity Name | Status |
|---------|-------------|--------|
| ordersbus | orders | ✅ |
| orderlineitemsbus | order_line_items | ✅ |
| customersbus | customers | ✅ |
| orderfulfillmentstatusbus | order_fulfillment_statuses | ✅ |
| lineitemfulfillmentstatusbus | line_item_fulfillment_statuses | ✅ |

### Core Domain

| Package | Entity Name | Status |
|---------|-------------|--------|
| userbus | users | ✅ |
| rolebus | roles | ✅ |
| userrolebus | user_roles | ✅ |
| tableaccessbus | table_access | ✅ |
| permissionsbus | permissions | N/A (read-only) |

### Products Domain

| Package | Entity Name | Status |
|---------|-------------|--------|
| productbus | products | ✅ |
| productcategorybus | product_categories | ✅ |
| brandbus | brands | ✅ |

### Inventory Domain

| Package | Entity Name | Status |
|---------|-------------|--------|
| warehousebus | warehouses | ✅ |
| inventoryitembus | inventory_items | ✅ |
| inventorylocationbus | inventory_locations | ✅ |

## Related Documentation

- [Event Infrastructure](event-infrastructure.md) - EventPublisher and delegate pattern details
- [Architecture](architecture.md) - System overview and component details
- [Testing](testing.md) - Testing patterns for workflow events

## Troubleshooting

### Events Not Firing

1. Check entity name matches `workflow.entities` table
2. Verify `delegate.Call()` is after successful database operation
3. Check delegate handler is registered in `all.go`
4. Check logs for delegate call errors

### Build Errors

1. Verify import path is correct
2. Check entity type matches `model.go`
3. Ensure all three action functions are implemented

### Tests Failing

1. Verify entity is seeded in test data
2. Check workflow infrastructure is initialized
3. Review queue manager metrics for errors
