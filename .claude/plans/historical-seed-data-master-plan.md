# Historical Seed Data Implementation - Master Plan

## Overview

This plan implements optional historical date support for seed data across all domains in the Ichor ERP system. This enables realistic time-distributed data for charts, reports, and analytics while maintaining Go idioms and Ardan Labs architectural patterns.

## Architecture Principles

### Core Design Decisions

1. **Optional Field Pattern**: Use pointer fields (`*time.Time`) in `New*` structs to distinguish between:
   - `nil` = Use business layer default (`time.Now()`) - Production behavior
   - `&time.Time` = Use explicit date - Seeding/testing behavior

2. **Business Layer Control**: The business layer (`*bus.Create()`) retains full control:
   - Always defaults to `time.Now().UTC()` when not provided
   - Accepts explicit dates only when intentionally provided
   - Maintains audit trail integrity

3. **Separation of Concerns**:
   - **Production API**: Uses `nil` → automatic timestamps
   - **Existing Unit Tests**: Unchanged - continue using existing `TestSeed*()` with `nil` → automatic timestamps
   - **New Seed Data**: New `TestSeed*Historical()` functions with explicit timestamps

4. **Zero Breaking Changes**:
   - ✅ All existing code works without modification
   - ✅ All existing tests pass unchanged
   - ✅ Only `seedFrontend.go` needs updates to use new historical functions
   - ✅ Production API behavior unchanged (always passes `nil`)

## Domain Organization

### Phase 1: Sales Domain (High Priority - Chart Dependency)
**Why First**: Required for "Last 30 Days Sales" chart and monthly revenue trends

#### 1.1 ordersbus
**Files to Modify**:
- `business/domain/sales/ordersbus/model.go`
- `business/domain/sales/ordersbus/ordersbus.go`
- `business/domain/sales/ordersbus/testutil.go`
- `app/domain/sales/ordersapp/model.go`

**Current State**:
```go
// model.go
type Order struct {
    ID                  uuid.UUID
    Number              string
    CustomerID          uuid.UUID
    DueDate             time.Time
    FulfillmentStatusID uuid.UUID
    CreatedBy           uuid.UUID
    UpdatedBy           uuid.UUID
    CreatedDate         time.Time  // Set by business layer
    UpdatedDate         time.Time  // Set by business layer
}

type NewOrder struct {
    Number              string
    CustomerID          uuid.UUID
    DueDate             time.Time
    FulfillmentStatusID uuid.UUID
    CreatedBy           uuid.UUID
    // NO CreatedDate field
}
```

**Changes Required**:

**Step 1**: Update `model.go`
```go
type NewOrder struct {
    Number              string
    CustomerID          uuid.UUID
    DueDate             time.Time
    FulfillmentStatusID uuid.UUID
    CreatedBy           uuid.UUID
    CreatedDate         *time.Time  // NEW: Optional for seeding
}
```

**Step 2**: Update `ordersbus.go` Create method
```go
func (b *Business) Create(ctx context.Context, no NewOrder) (Order, error) {
    ctx, span := otel.AddSpan(ctx, "business.ordersbus.create")
    defer span.End()

    now := time.Now().UTC()
    if no.CreatedDate != nil {
        now = *no.CreatedDate  // Use provided date for seeding
    }

    order := Order{
        ID:                  uuid.New(),
        Number:              no.Number,
        CustomerID:          no.CustomerID,
        DueDate:             no.DueDate,
        FulfillmentStatusID: no.FulfillmentStatusID,
        CreatedBy:           no.CreatedBy,
        UpdatedBy:           no.CreatedBy,
        CreatedDate:         now,
        UpdatedDate:         now,
    }

    if err := b.storer.Create(ctx, order); err != nil {
        return Order{}, err
    }
    return order, nil
}
```

**Step 3**: Add to `testutil.go` (APPEND - don't modify existing functions)
```go
// TestNewOrdersHistorical creates orders distributed across a time range for seeding.
// daysBack specifies how many days of history to generate (e.g., 30, 90, 365).
// Orders are evenly distributed across the time range.
func TestNewOrdersHistorical(n int, daysBack int, userIDs uuid.UUIDs, customerIDs uuid.UUIDs, ofIDs uuid.UUIDs) []NewOrder {
    orders := make([]NewOrder, 0, n)
    now := time.Now()

    for i := 0; i < n; i++ {
        // Distribute evenly across the time range
        daysAgo := (i * daysBack) / n
        createdDate := now.AddDate(0, 0, -daysAgo)

        orders = append(orders, NewOrder{
            Number:              fmt.Sprintf("SEED-%d", i+1),
            CustomerID:          customerIDs[i%len(customerIDs)],
            DueDate:             createdDate.AddDate(0, 0, 7), // Due 7 days after creation
            FulfillmentStatusID: ofIDs[i%len(ofIDs)],
            CreatedBy:           userIDs[i%len(userIDs)],
            CreatedDate:         &createdDate,  // Explicit historical date
        })
    }
    return orders
}

// TestSeedOrdersHistorical seeds orders with historical date distribution.
func TestSeedOrdersHistorical(ctx context.Context, n int, daysBack int, userIDs uuid.UUIDs, customerIDs uuid.UUIDs, ofIDs uuid.UUIDs, api *Business) ([]Order, error) {
    newOrders := TestNewOrdersHistorical(n, daysBack, userIDs, customerIDs, ofIDs)
    orders := make([]Order, len(newOrders))
    for i, no := range newOrders {
        order, err := api.Create(ctx, no)
        if err != nil {
            return []Order{}, err
        }
        orders[i] = order
    }
    return orders, nil
}
```

**IMPORTANT**: Keep existing `TestNewOrders()` and `TestSeedOrders()` unchanged - all existing tests continue to work!

**Step 4**: Update `app/domain/sales/ordersapp/model.go`
```go
type NewOrder struct {
    Number              string     `json:"number" validate:"required"`
    CustomerID          string     `json:"customerId" validate:"required,uuid"`
    DueDate             string     `json:"dueDate" validate:"required"`
    FulfillmentStatusID string     `json:"fulfillmentStatusId" validate:"required,uuid"`
    CreatedBy           string     `json:"createdBy" validate:"required,uuid"`
    CreatedDate         *string    `json:"createdDate"` // NEW: Optional for seeding/import
}

// Update toBusNewOrder function
func toBusNewOrder(app NewOrder) (ordersbus.NewOrder, error) {
    customerID, err := uuid.Parse(app.CustomerID)
    if err != nil {
        return ordersbus.NewOrder{}, fmt.Errorf("parse customerid: %w", err)
    }

    fulfillmentStatusID, err := uuid.Parse(app.FulfillmentStatusID)
    if err != nil {
        return ordersbus.NewOrder{}, fmt.Errorf("parse fulfillmentstatusid: %w", err)
    }

    createdBy, err := uuid.Parse(app.CreatedBy)
    if err != nil {
        return ordersbus.NewOrder{}, fmt.Errorf("parse createdby: %w", err)
    }

    dueDate, err := time.Parse(time.RFC3339, app.DueDate)
    if err != nil {
        return ordersbus.NewOrder{}, fmt.Errorf("parse duedate: %w", err)
    }

    bus := ordersbus.NewOrder{
        Number:              app.Number,
        CustomerID:          customerID,
        DueDate:             dueDate,
        FulfillmentStatusID: fulfillmentStatusID,
        CreatedBy:           createdBy,
        // CreatedDate: nil by default - API always uses server time
    }

    // NEW: Handle optional CreatedDate (for imports/admin tools only)
    if app.CreatedDate != nil && *app.CreatedDate != "" {
        createdDate, err := time.Parse(time.RFC3339, *app.CreatedDate)
        if err != nil {
            return ordersbus.NewOrder{}, fmt.Errorf("parse createddate: %w", err)
        }
        bus.CreatedDate = &createdDate
    }

    return bus, nil
}
```

**Step 5**: Update `business/sdk/dbtest/seedFrontend.go`
```go
// Change from:
orders, err := ordersbus.TestSeedOrders(ctx, 100, userIDs, customerIDs, ofIDs, busDomain.Orders)

// To:
orders, err := ordersbus.TestSeedOrdersHistorical(ctx, 100, 90, userIDs, customerIDs, ofIDs, busDomain.Orders)
// 100 orders distributed across last 90 days
```

**What Tests Run Unchanged**:
- ✅ `api/cmd/services/ichor/tests/sales/ordersapi/*_test.go` - All continue using `TestSeedOrders()` with automatic timestamps
- ✅ `business/domain/sales/ordersbus/ordersbus_test.go` - All unit tests unchanged

#### 1.2 orderlineitemsbus
**Files to Modify**:
- `business/domain/sales/orderlineitemsbus/model.go`
- `business/domain/sales/orderlineitemsbus/orderlineitemsbus.go`
- `business/domain/sales/orderlineitemsbus/testutil.go`
- `app/domain/sales/orderlineitemsapp/model.go`

**Pattern**: Same as ordersbus above

**Key Consideration**: Line items should inherit the order's created date or be slightly after:
```go
func TestNewOrderLineItemsHistorical(n int, orderDates map[uuid.UUID]time.Time, orderIDs uuid.UUIDs, productIDs uuid.UUIDs, fulfillmentStatusIDs uuid.UUIDs, userIDs uuid.UUIDs) []NewOrderLineItem {
    items := make([]NewOrderLineItem, 0, n)

    for i := 0; i < n; i++ {
        orderID := orderIDs[i%len(orderIDs)]
        orderDate := orderDates[orderID]
        // Line item created 0-2 hours after order
        lineItemDate := orderDate.Add(time.Duration(rand.Intn(120)) * time.Minute)

        items = append(items, NewOrderLineItem{
            OrderID:                     orderID,
            ProductID:                   productIDs[i%len(productIDs)],
            Quantity:                    i%10 + 1,
            UnitPrice:                   float64(50 + i*5),
            LineItemFulfillmentStatusID: fulfillmentStatusIDs[i%len(fulfillmentStatusIDs)],
            CreatedBy:                   userIDs[i%len(userIDs)],
            CreatedDate:                 &lineItemDate,
        })
    }
    return items
}

// Usage in seedFrontend.go:
// Create map of order IDs to their created dates
orderDates := make(map[uuid.UUID]time.Time)
for _, order := range orders {
    orderDates[order.ID] = order.CreatedDate
}

lineItems, err := orderlineitemsbus.TestSeedOrderLineItemsHistorical(ctx, 250, orderDates, orderIDs, productIDs, lineItemStatusIDs, userIDs, busDomain.OrderLineItems)
```

#### 1.3 customersbus
**Files to Modify**:
- `business/domain/sales/customersbus/model.go`
- `business/domain/sales/customersbus/customersbus.go`
- `business/domain/sales/customersbus/testutil.go`
- `app/domain/sales/customersapp/model.go`

**Pattern**: Same pattern, but distribute customer creation further back (180-365 days)

```go
func TestNewCustomersHistorical(n int, daysBack int, streetIDs uuid.UUIDs, userIDs uuid.UUIDs) []NewCustomer {
    // Distribute across 180-365 days (customers exist longer than orders)
}
```

---

### Phase 2: Procurement Domain
**Why Second**: Purchase orders need historical data for supply chain analytics

#### 2.1 purchaseorderbus
**Files to Modify**:
- `business/domain/procurement/purchaseorderbus/model.go`
- `business/domain/procurement/purchaseorderbus/purchaseorderbus.go`
- `business/domain/procurement/purchaseorderbus/testutil.go`
- `app/domain/procurement/purchaseorderapp/model.go` (if exists)

**Special Field**: Has both `OrderDate` and `CreatedDate`
```go
type NewPurchaseOrder struct {
    // ... existing fields
    OrderDate            time.Time
    ExpectedDeliveryDate time.Time
    CreatedBy            uuid.UUID
    CreatedDate          *time.Time  // NEW: Optional for seeding (typically same as OrderDate)
}
```

**Historical Pattern**:
```go
func TestNewPurchaseOrdersHistorical(n int, daysBack int, supplierIDs uuid.UUIDs, statusIDs uuid.UUIDs, warehouseIDs uuid.UUIDs, streetIDs uuid.UUIDs, userIDs uuid.UUIDs) []NewPurchaseOrder {
    orders := make([]NewPurchaseOrder, 0, n)
    now := time.Now()

    for i := 0; i < n; i++ {
        daysAgo := (i * daysBack) / n
        orderDate := now.AddDate(0, 0, -daysAgo)
        expectedDelivery := orderDate.AddDate(0, 0, 14) // 2 weeks out

        subtotal := 1000.00 + float64(i*100)
        tax := subtotal * 0.08
        shipping := 50.00
        total := subtotal + tax + shipping

        orders = append(orders, NewPurchaseOrder{
            OrderNumber:              fmt.Sprintf("PO-HIST-%d", i+1),
            SupplierID:               supplierIDs[i%len(supplierIDs)],
            PurchaseOrderStatusID:    statusIDs[i%len(statusIDs)],
            DeliveryWarehouseID:      warehouseIDs[i%len(warehouseIDs)],
            DeliveryLocationID:       uuid.Nil,
            DeliveryStreetID:         streetIDs[i%len(streetIDs)],
            OrderDate:                orderDate,  // Still regular field
            ExpectedDeliveryDate:     expectedDelivery,
            Subtotal:                 subtotal,
            TaxAmount:                tax,
            ShippingCost:             shipping,
            TotalAmount:              total,
            Currency:                 "USD",
            RequestedBy:              userIDs[i%len(userIDs)],
            Notes:                    fmt.Sprintf("Historical PO %d", i+1),
            SupplierReferenceNumber:  fmt.Sprintf("SUP-HIST-%d", i+1),
            CreatedBy:                userIDs[i%len(userIDs)],
            CreatedDate:              &orderDate,  // Use same as order date
        })
    }
    return orders
}

func TestSeedPurchaseOrdersHistorical(ctx context.Context, n int, daysBack int, supplierIDs uuid.UUIDs, statusIDs uuid.UUIDs, warehouseIDs uuid.UUIDs, streetIDs uuid.UUIDs, userIDs uuid.UUIDs, api *Business) ([]PurchaseOrder, error) {
    newOrders := TestNewPurchaseOrdersHistorical(n, daysBack, supplierIDs, statusIDs, warehouseIDs, streetIDs, userIDs)
    orders := make([]PurchaseOrder, len(newOrders))
    for i, no := range newOrders {
        order, err := api.Create(ctx, no)
        if err != nil {
            return []PurchaseOrder{}, fmt.Errorf("creating purchase order: %w", err)
        }
        orders[i] = order
    }
    return orders, nil
}
```

#### 2.2 purchaseorderlineitembus
**Files to Modify**: Same pattern as above

#### 2.3 supplierbus
**Files to Modify**: Same pattern, distribute 1-2 years back

#### 2.4 supplierproductbus
**Files to Modify**: Same pattern

---

### Phase 3: Inventory Domain
**Why Third**: Inventory tracking and movements need timestamps

#### 3.1 warehousebus
#### 3.2 zonebus
#### 3.3 inventorylocationbus
#### 3.4 inventoryitembus
#### 3.5 lottrackingsbus
#### 3.6 serialnumberbus
#### 3.7 inspectionbus
#### 3.8 inventoryadjustmentbus
#### 3.9 inventorytransactionbus
#### 3.10 transferorderbus

**Files to Modify for Each**:
- `model.go`
- `{entity}bus.go`
- `testutil.go`
- `app/domain/inventory/{entity}app/model.go` (if exists)

**Special Considerations**:
- Warehouses: Created 1+ years ago
- Inventory items: Distributed across 90-180 days
- Transactions: High frequency, last 60 days
- Adjustments: Last 90 days
- Transfer orders: Last 120 days

---

### Phase 4: Products Domain
**Why Fourth**: Product catalog with historical cost changes

#### 4.1 productbus
#### 4.2 brandbus
#### 4.3 productcategorybus
#### 4.4 productcostbus
**Special**: Cost history needs tiered historical dates
```go
// Oldest costs 1-2 years back, newer costs recent
```

#### 4.5 costhistorybus
**Special**: Multiple cost changes per product over time

#### 4.6 physicalattributebus
#### 4.7 metricsbus

**Files to Modify for Each**: Same pattern as above

---

### Phase 5: HR Domain
**Why Fifth**: Employee data and comments

#### 5.1 commentbus
#### 5.2 homebus

**Files to Modify**: Same pattern

**Considerations**:
- Comments: Distributed across last 30-180 days
- Homes: 1-5 years back (stable data)

---

### Phase 6: Assets Domain
**Why Last**: Asset tracking less critical for initial analytics

#### 6.1 validassetbus

**Files to Modify**: Same pattern

**Considerations**:
- Assets: Distributed 1-3 years back

---

### Phase 7: Core Domain (User)

#### 7.1 userbus
**Files to Modify**:
- `business/domain/core/userbus/model.go`
- `business/domain/core/userbus/userbus.go`
- `business/domain/core/userbus/testutil.go`

**Special Consideration**: Users are foundational and should be created far in the past
```go
// Admin users: 2+ years ago
// Regular users: Distributed 6-24 months ago
```

---

## Implementation Template

For each domain, follow this exact pattern:

### 1. Update `model.go`
```go
// Add to New{Entity} struct
type New{Entity} struct {
    // ... existing fields
    CreatedDate *time.Time  // Optional: if nil, uses time.Now(), otherwise explicit date for seeding
}
```

### 2. Update `{entity}bus.go`
```go
func (b *Business) Create(ctx context.Context, n New{Entity}) ({Entity}, error) {
    ctx, span := otel.AddSpan(ctx, "business.{entity}bus.create")
    defer span.End()

    now := time.Now().UTC()
    if n.CreatedDate != nil {
        now = *n.CreatedDate  // Use provided date for seeding
    }

    entity := {Entity}{
        // ... fields
        CreatedDate: now,
        UpdatedDate: now,
    }

    if err := b.storer.Create(ctx, entity); err != nil {
        return {Entity}{}, err
    }
    return entity, nil
}
```

### 3. Add to `testutil.go` (APPEND - keep existing functions)
```go
// TestNew{Entity}Historical creates entities distributed across a time range.
func TestNew{Entity}Historical(n int, daysBack int, /* other params */) []New{Entity} {
    entities := make([]New{Entity}, 0, n)
    now := time.Now()

    for i := 0; i < n; i++ {
        // Even distribution across time range
        daysAgo := (i * daysBack) / n
        createdDate := now.AddDate(0, 0, -daysAgo)

        entities = append(entities, New{Entity}{
            // ... fields
            CreatedDate: &createdDate,
        })
    }
    return entities
}

// TestSeed{Entity}Historical seeds entities with historical dates.
func TestSeed{Entity}Historical(ctx context.Context, n int, daysBack int, /* other params */, api *Business) ([]{Entity}, error) {
    newEntities := TestNew{Entity}Historical(n, daysBack, /* params */)
    entities := make([]{Entity}, len(newEntities))
    for i, ne := range newEntities {
        entity, err := api.Create(ctx, ne)
        if err != nil {
            return []{Entity}{}, err
        }
        entities[i] = entity
    }
    return entities, nil
}
```

**CRITICAL**: Do NOT modify existing `TestNew{Entity}()` or `TestSeed{Entity}()` functions!

### 4. Update `app/domain/{area}/{entity}app/model.go` (if exists)
```go
type New{Entity} struct {
    // ... existing fields
    CreatedDate *string `json:"createdDate"` // Optional: ISO8601 format
}

// Update conversion function
func toBusNew{Entity}(app New{Entity}) ({entity}bus.New{Entity}, error) {
    bus := {entity}bus.New{Entity}{
        // ... field conversions
        // CreatedDate: nil  - default behavior for API
    }

    // Handle optional CreatedDate (for imports/admin only)
    if app.CreatedDate != nil && *app.CreatedDate != "" {
        createdDate, err := time.Parse(time.RFC3339, *app.CreatedDate)
        if err != nil {
            return {entity}bus.New{Entity}{}, fmt.Errorf("parse createddate: %w", err)
        }
        bus.CreatedDate = &createdDate
    }

    return bus, nil
}
```

### 5. Update `business/sdk/dbtest/seedFrontend.go`
```go
// Change from:
entities, err := {entity}bus.TestSeed{Entity}(ctx, count, /* params */, busDomain.{Entity})

// To:
entities, err := {entity}bus.TestSeed{Entity}Historical(ctx, count, daysBack, /* params */, busDomain.{Entity})
```

---

## Testing Strategy

### Existing Tests (UNCHANGED)
✅ **No modifications needed** - all existing tests continue to use:
- `TestNew{Entity}()` functions
- `TestSeed{Entity}()` functions
- These pass `nil` for `CreatedDate` → automatic `time.Now()`

Example - existing test continues to work:
```go
// api/cmd/services/ichor/tests/sales/ordersapi/create_test.go
func create200(sd apitest.SeedData) []apitest.Table {
    // Uses TestSeedOrders() internally - still works!
    // CreatedDate is nil, so time.Now() is used automatically
}
```

### New Seed Data Tests
Only `make seed-frontend` uses historical functions - no test changes needed!

### Verification (Manual)
After implementing a phase, verify in dev environment:
```bash
# Reseed with historical data
make dev-database-recreate
make migrate
make seed-frontend

# Verify data distribution (manual inspection)
make pgcli
# Then in psql:
SELECT
    DATE(created_date) as day,
    COUNT(*) as count
FROM sales.orders
WHERE created_date > NOW() - INTERVAL '90 days'
GROUP BY DATE(created_date)
ORDER BY day;
```

---

## Chart Configuration Updates

After implementing Phase 1 (Sales), update chart configurations:

### New Chart: Last 30 Days Daily Sales
```go
// business/sdk/dbtest/seedmodels/charts.go

var SeedLineLast30DaysSales = &tablebuilder.Config{
    Title:         "Last 30 Days Sales",
    WidgetType:    "chart",
    Visualization: "line",
    DataSource: []tablebuilder.DataSource{
        {
            Type:   "query",
            Source: "order_line_items",
            Schema: "sales",
            Metrics: []tablebuilder.MetricConfig{
                {
                    Name:     "daily_sales",
                    Function: "sum",
                    Expression: &tablebuilder.ExpressionConfig{
                        Operator: "multiply",
                        Columns:  []string{"order_line_items.quantity", "product_costs.selling_price"},
                    },
                },
            },
            GroupBy: &tablebuilder.GroupByConfig{
                Column:   "orders.created_date",
                Interval: "day",
                Alias:    "day",
            },
            Select: tablebuilder.SelectConfig{
                ForeignTables: []tablebuilder.ForeignTable{
                    {
                        Table:            "orders",
                        Schema:           "sales",
                        RelationshipFrom: "order_line_items.order_id",
                        RelationshipTo:   "orders.id",
                        JoinType:         "inner",
                    },
                    {
                        Table:            "products",
                        Schema:           "products",
                        RelationshipFrom: "order_line_items.product_id",
                        RelationshipTo:   "products.id",
                        JoinType:         "inner",
                        ForeignTables: []tablebuilder.ForeignTable{
                            {
                                Table:            "product_costs",
                                Schema:           "products",
                                RelationshipFrom: "products.id",
                                RelationshipTo:   "product_costs.product_id",
                                JoinType:         "inner",
                            },
                        },
                    },
                },
            },
            Sort: []tablebuilder.Sort{
                {Column: "day", Direction: "asc"},
            },
        },
    },
    VisualSettings: tablebuilder.VisualSettings{
        Columns: map[string]tablebuilder.ColumnConfig{
            "_chart": {
                CellTemplate: createChartSettings(tablebuilder.ChartVisualSettings{
                    ChartType:      "line",
                    CategoryColumn: "day",
                    ValueColumns:   []string{"daily_sales"},
                    XAxis: &tablebuilder.AxisConfig{
                        Title: "Day",
                        Type:  "time",
                    },
                    YAxis: &tablebuilder.AxisConfig{
                        Title:  "Sales ($)",
                        Type:   "value",
                        Format: "currency",
                    },
                }),
            },
        },
    },
}

// Add to ChartConfigs slice (around line 982)
{Name: "seed_line_last_30_days", Description: "Line chart showing daily sales for last 30 days", Config: SeedLineLast30DaysSales},
```

---

## Rollout Plan

### Week 1: Phase 1 (Sales)
- Day 1-2: ordersbus + orderlineitemsbus
- Day 3: customersbus
- Day 4: Update charts, verify seed data
- Day 5: Integration verification (no test changes!)

### Week 2: Phase 2 (Procurement)
- Day 1-2: purchaseorderbus + lineitem
- Day 3: supplierbus + supplierproductbus
- Day 4-5: Verification

### Week 3: Phase 3 (Inventory - Part 1)
- Day 1: warehousebus, zonebus, locationbus
- Day 2: inventoryitembus, lottrackingsbus
- Day 3: serialnumberbus, inspectionbus
- Day 4: adjustmentbus, transactionbus
- Day 5: transferorderbus, verification

### Week 4: Phases 4-6
- Day 1-2: Products domain
- Day 3: HR domain
- Day 4: Assets domain
- Day 5: Core/userbus, final verification

---

## Success Criteria

✅ All seed data distributed across historical time ranges
✅ **Zero test modifications** - all existing tests pass unchanged
✅ Charts display realistic time-series data
✅ Optional field pattern consistently applied
✅ Production API still uses automatic timestamps (always `nil`)
✅ `make seed-frontend` generates historical data
✅ `make test` passes without any changes

---

## Rollback Strategy

If issues arise, each phase is independent:
1. Revert the business layer changes (restore `time.Now()` only)
2. Revert testutil additions (remove new *Historical functions)
3. Revert seedFrontend.go calls
4. No test changes to rollback - tests never changed!
5. No data migration needed - just reseed

---

## Files Changed Per Domain

For each domain entity (e.g., `ordersbus`):

### Modified:
1. `business/domain/{area}/{entity}bus/model.go` - Add `CreatedDate *time.Time` to `New{Entity}`
2. `business/domain/{area}/{entity}bus/{entity}bus.go` - Update `Create()` method
3. `business/domain/{area}/{entity}bus/testutil.go` - **ADD** new functions (don't modify existing)
4. `app/domain/{area}/{entity}app/model.go` - Add `CreatedDate *string` and update conversion

### Unchanged:
- ❌ No test files modified
- ❌ No API files modified
- ❌ No database stores modified
- ✅ Only `seedFrontend.go` calls new historical functions

---

## Notes

- **Thread Safety**: No concerns - `time.Now()` vs explicit times are both safe
- **Performance**: No impact - same code path, just conditional timestamp
- **Data Quality**: Improved - realistic time distribution for analytics
- **Maintainability**: High - consistent pattern across all domains
- **Go Idioms**: Follows pointer-for-optional pattern
- **Ardan Labs**: Respects business layer ownership of timestamp generation
- **Backward Compatibility**: 100% - all existing code works unchanged
