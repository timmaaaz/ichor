# Task: Create Integration Test for Order-Grouped Inventory Allocation

## Background

We implemented a workflow system where inventory allocation is triggered when order line items are created. The key requirement is **order preservation**: all line items from Order A should be allocated before any line items from Order B, even if they're created nearly simultaneously.

This prevents the scenario where:
- User A orders items 1, 2, 3
- User B orders items 1, 2 a second later
- Without ordering: User A gets item 1, User B takes item 2, User A can't complete their order

## What Was Implemented

### 1. Modified `business/sdk/workflow/workflowactions/inventory/allocate.go`

Added `SourceFromLineItem` config option that extracts allocation data from the triggering line item's `RawData`:

```go
type AllocateInventoryConfig struct {
    InventoryItems     []AllocationItem `json:"inventory_items"`
    SourceFromLineItem bool             `json:"source_from_line_item"` // NEW
    AllocationMode     string           `json:"allocation_mode"`
    AllocationStrategy string           `json:"allocation_strategy"`
    AllowPartial       bool             `json:"allow_partial"`
    Priority           string           `json:"priority"`
    ReferenceType      string           `json:"reference_type"`
    ReferenceID        string           `json:"reference_id"`
    ReservationHours   int              `json:"reservation_hours"`
    IdempotencyKey     string           `json:"idempotency_key"`
}
```

**In `Validate()` (~line 183-185):**
```go
// Allow empty inventory_items when sourcing from line item (extracted at execute time)
if len(cfg.InventoryItems) == 0 && !cfg.SourceFromLineItem {
    return errors.New("inventory_items list is required and must not be empty")
}
```

**In `Execute()` (~lines 240-276):**
```go
// If sourcing from line item, extract product_id and quantity from raw data
if cfg.SourceFromLineItem {
    productIDStr, _ := execContext.RawData["product_id"].(string)
    productID, err := uuid.Parse(productIDStr)
    if err != nil {
        return QueuedAllocationResponse{}, fmt.Errorf("invalid product_id in line item: %w", err)
    }

    quantity, ok := execContext.RawData["quantity"].(float64)
    if !ok || quantity <= 0 {
        if qInt, ok := execContext.RawData["quantity"].(int); ok {
            quantity = float64(qInt)
        }
    }
    if quantity <= 0 {
        return QueuedAllocationResponse{}, errors.New("quantity must be greater than 0")
    }

    orderIDStr, _ := execContext.RawData["order_id"].(string)

    cfg.InventoryItems = []AllocationItem{{
        ProductID: productID,
        Quantity:  int(quantity),
    }}
    cfg.ReferenceID = orderIDStr
    cfg.ReferenceType = "order"
}
```

### 2. Modified `business/sdk/dbtest/seedFrontend.go` (~lines 3674-3765)

Changed the workflow rule to trigger on `order_line_items.on_create` instead of `orders.on_create`:

```go
// Query order_line_items entity
orderLineItemsEntity, err := busDomain.Workflow.QueryEntityByName(ctx, "order_line_items")

allocateConfig := map[string]interface{}{
    "source_from_line_item": true,
    "allocation_mode":       "reserve",
    "allocation_strategy":   "fifo",
    "allow_partial":         false,
    "priority":              "high",
    "reference_type":        "order",
}

orderAllocateRule, err := busDomain.Workflow.CreateRule(ctx, workflow.NewAutomationRule{
    Name:        "Line Item Created - Allocate Inventory",
    Description: "When an order line item is created, attempt to reserve inventory for that product",
    EntityID:    orderLineItemsEntity.ID,
    EventType:   "on_create",
    // ...
})
```

## Test Requirements

Create a new test function in `business/sdk/workflow/workflowactions/inventory/allocate_test.go` that verifies:

### Scenario 1: Single Order with Multiple Line Items
1. Create Order A with 3 line items (different products)
2. Each line item triggers allocation via `SourceFromLineItem`
3. Verify all 3 allocations succeed
4. Verify all allocations have the same `ReferenceID` (Order A's ID)
5. Verify `ReferenceType` is "order"

### Scenario 2: Competing Orders (Order Preservation)
1. Create Order A with 3 line items for products 1, 2, 3
2. Create Order B with 2 line items for products 1, 2 (same products, competing)
3. Inventory: Product 1 has qty 10, Product 2 has qty 5
4. Order A requests: Product 1 qty 8, Product 2 qty 4, Product 3 qty 2
5. Order B requests: Product 1 qty 5, Product 2 qty 3

**Expected behavior:**
- Process in order: A1, A2, A3, B1, B2
- Order A gets: Product 1 (8), Product 2 (4), Product 3 (2) ✓
- Order B gets: Product 1 (2 remaining), Product 2 (1 remaining) - partial or fail depending on `allow_partial`

### Test Implementation Pattern

Follow the existing test pattern in `allocate_test.go`:

```go
func testSourceFromLineItem(busDomain dbtest.BusDomain, db *sqlx.DB, sd allocateSeedData) unitest.Table {
    return unitest.Table{
        Name:    "test_source_from_line_item",
        ExpResp: "success",
        ExcFunc: func(ctx context.Context) any {
            orderID := uuid.New()

            // Create config with SourceFromLineItem enabled
            config := inventory.AllocateInventoryConfig{
                SourceFromLineItem: true,
                AllocationMode:     "reserve",
                AllocationStrategy: "fifo",
                AllowPartial:       false,
                Priority:           "high",
            }
            configJSON, _ := json.Marshal(config)

            // Create execution context with RawData simulating line item
            execContext := workflow.ActionExecutionContext{
                EntityID:    uuid.New(), // line item ID
                EntityName:  "order_line_items",
                EventType:   "on_create",
                UserID:      sd.Admins[0].ID,
                RuleID:      uuid.New(),
                RuleName:    "Test Line Item Allocation",
                ExecutionID: uuid.New(),
                Timestamp:   time.Now().UTC(),
                RawData: map[string]interface{}{
                    "product_id": sd.Products[0].ProductID.String(),
                    "quantity":   float64(5),
                    "order_id":   orderID.String(),
                },
            }

            // Execute - should extract from RawData
            result, err := sd.Handler.Execute(ctx, configJSON, execContext)
            if err != nil {
                return err
            }

            // Verify queued response
            queuedResp, ok := result.(inventory.QueuedAllocationResponse)
            if !ok {
                return fmt.Errorf("expected QueuedAllocationResponse, got %T", result)
            }

            return queuedResp.Status
        },
        CmpFunc: func(got any, exp any) string {
            if got != exp {
                return fmt.Sprintf("got %v, want %v", got, exp)
            }
            return ""
        },
    }
}

func testOrderGroupedAllocation(busDomain dbtest.BusDomain, db *sqlx.DB, sd allocateSeedData) unitest.Table {
    return unitest.Table{
        Name:    "test_order_grouped_allocation",
        ExpResp: true,
        ExcFunc: func(ctx context.Context) any {
            orderA := uuid.New()
            orderB := uuid.New()

            config := inventory.AllocateInventoryConfig{
                SourceFromLineItem: true,
                AllocationMode:     "reserve",
                AllocationStrategy: "fifo",
                AllowPartial:       false,
                Priority:           "high",
            }
            configJSON, _ := json.Marshal(config)

            // Track allocation order
            var allocationOrder []string

            // Simulate Order A's line items (3 items)
            for i := 0; i < 3 && i < len(sd.Products); i++ {
                execCtx := workflow.ActionExecutionContext{
                    EntityID:    uuid.New(),
                    EntityName:  "order_line_items",
                    EventType:   "on_create",
                    UserID:      sd.Admins[0].ID,
                    RuleID:      uuid.New(),
                    RuleName:    "Test Grouped Allocation",
                    ExecutionID: uuid.New(),
                    Timestamp:   time.Now().UTC(),
                    RawData: map[string]interface{}{
                        "product_id": sd.Products[i].ProductID.String(),
                        "quantity":   float64(5),
                        "order_id":   orderA.String(),
                    },
                }

                _, err := sd.Handler.Execute(ctx, configJSON, execCtx)
                if err != nil {
                    return fmt.Errorf("order A item %d failed: %w", i, err)
                }
                allocationOrder = append(allocationOrder, fmt.Sprintf("A%d", i+1))
            }

            // Simulate Order B's line items (2 items)
            for i := 0; i < 2 && i < len(sd.Products); i++ {
                execCtx := workflow.ActionExecutionContext{
                    EntityID:    uuid.New(),
                    EntityName:  "order_line_items",
                    EventType:   "on_create",
                    UserID:      sd.Admins[0].ID,
                    RuleID:      uuid.New(),
                    RuleName:    "Test Grouped Allocation",
                    ExecutionID: uuid.New(),
                    Timestamp:   time.Now().UTC(),
                    RawData: map[string]interface{}{
                        "product_id": sd.Products[i].ProductID.String(),
                        "quantity":   float64(3),
                        "order_id":   orderB.String(),
                    },
                }

                _, err := sd.Handler.Execute(ctx, configJSON, execCtx)
                if err != nil {
                    return fmt.Errorf("order B item %d failed: %w", i, err)
                }
                allocationOrder = append(allocationOrder, fmt.Sprintf("B%d", i+1))
            }

            // Verify order: A1, A2, A3, B1, B2
            expectedOrder := []string{"A1", "A2", "A3", "B1", "B2"}
            for i, exp := range expectedOrder {
                if i >= len(allocationOrder) || allocationOrder[i] != exp {
                    return false
                }
            }

            return true
        },
        CmpFunc: func(got any, exp any) string {
            if got != exp {
                return fmt.Sprintf("order grouping failed: got %v, want %v", got, exp)
            }
            return ""
        },
    }
}
```

### Add Tests to Test Suite

In `allocateInventoryTests()` function, add the new tests:

```go
func allocateInventoryTests(busDomain dbtest.BusDomain, db *sqlx.DB, sd allocateSeedData) []unitest.Table {
    return []unitest.Table{
        validateAllocationConfig(sd),
        executeBasicAllocation(busDomain, db, sd),
        executePartialAllocation(busDomain, db, sd),
        testIdempotency(busDomain, db, sd),
        testReservationMode(busDomain, db, sd),
        testFIFOStrategy(busDomain, db, sd),
        testSourceFromLineItem(busDomain, db, sd),           // NEW
        testOrderGroupedAllocation(busDomain, db, sd),       // NEW
    }
}
```

## Key Files to Read

1. `business/sdk/workflow/workflowactions/inventory/allocate.go` - Main handler implementation
2. `business/sdk/workflow/workflowactions/inventory/allocate_test.go` - Existing tests to follow pattern
3. `business/sdk/workflow/workflow.go` - `ActionExecutionContext` struct definition
4. `business/sdk/unitest/unitest.go` - Test table pattern

## Validation Criteria

1. ✅ `SourceFromLineItem: true` correctly extracts `product_id`, `quantity`, `order_id` from `RawData`
2. ✅ Empty `inventory_items` is allowed when `SourceFromLineItem: true`
3. ✅ `ReferenceID` is set to the extracted `order_id`
4. ✅ `ReferenceType` is set to "order"
5. ✅ Sequential processing maintains order (A1, A2, A3 before B1, B2)
6. ✅ Invalid/missing `product_id` or `quantity` returns appropriate error

## Edge Cases to Test

1. Missing `product_id` in RawData → should error
2. Invalid UUID for `product_id` → should error
3. Zero or negative `quantity` → should error
4. Missing `order_id` → should work (ReferenceID will be empty string)
5. Integer quantity (not float64) → should work via type assertion fallback

## Run Tests

```bash
go test -v ./business/sdk/workflow/workflowactions/inventory/... -run "Test_AllocateInventory"
```
