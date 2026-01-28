# allocate_inventory Action

Reserves or allocates inventory items from warehouses. Supports multiple allocation strategies and asynchronous processing.

## Configuration Schema

```json
{
  "inventory_items": [
    {
      "product_id": "uuid",
      "quantity": 10,
      "warehouse_id": "uuid (optional)",
      "location_id": "uuid (optional)"
    }
  ],
  "source_from_line_item": false,
  "allocation_mode": "reserve|allocate",
  "allocation_strategy": "fifo|lifo|nearest_expiry|lowest_cost",
  "allow_partial": false,
  "reservation_duration_hours": 24,
  "priority": "low|medium|high|critical",
  "timeout_ms": 30000,
  "reference_id": "string",
  "reference_type": "string"
}
```

**Source**: `business/sdk/workflow/workflowactions/inventory/allocate.go:30-41`

## Fields

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `inventory_items` | []AllocationItem | Conditional | - | Items to allocate |
| `source_from_line_item` | bool | No | `false` | Extract from trigger event |
| `allocation_mode` | string | **Yes** | - | Reserve or allocate |
| `allocation_strategy` | string | **Yes** | - | How to select inventory |
| `allow_partial` | bool | No | `false` | Accept partial fulfillment |
| `reservation_duration_hours` | int | Conditional | `24` | TTL for reservations |
| `priority` | string | **Yes** | - | Queue priority |
| `timeout_ms` | int | No | `30000` | Operation timeout |
| `reference_id` | string | No | - | Order/transfer ID |
| `reference_type` | string | No | - | Type of reference |

### Allocation Item

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `product_id` | UUID | **Yes** | Product to allocate |
| `quantity` | int | **Yes** | Quantity (must be > 0) |
| `warehouse_id` | UUID | No | Specific warehouse |
| `location_id` | UUID | No | Specific location |

## Allocation Modes

| Mode | Description | Use Case |
|------|-------------|----------|
| `reserve` | Temporarily holds inventory | Shopping cart, pending orders |
| `allocate` | Permanently commits inventory | Confirmed orders |

### Reserve Mode

- Creates a temporary hold on inventory
- Has an expiration time (`reservation_duration_hours`)
- Inventory is "reserved_quantity" - not available to others
- Can be converted to allocation or released

### Allocate Mode

- Permanently commits inventory
- No expiration
- Inventory is "allocated_quantity"
- Creates audit trail in inventory_transactions

## Allocation Strategies

| Strategy | Description | Implementation |
|----------|-------------|----------------|
| `fifo` | First In, First Out | Oldest inventory first |
| `lifo` | Last In, First Out | Newest inventory first |
| `nearest_expiry` | Closest to expiration | Uses FIFO (simplified) |
| `lowest_cost` | Cheapest warehouse | Uses FIFO (simplified) |

**Note**: `nearest_expiry` and `lowest_cost` are simplified implementations that currently use FIFO logic.

## Priority Levels

| Priority | Queue Priority Value |
|----------|---------------------|
| `low` | 1 |
| `medium` | 5 |
| `high` | 10 |
| `critical` | 20 |

Higher priority items are processed first from the queue.

## Validation Rules

1. `inventory_items` required unless `source_from_line_item`
2. `allocation_strategy` must be valid
3. `allocation_mode` must be valid
4. `priority` must be valid
5. Each item must have `product_id` and `quantity > 0`

**Source**: `business/sdk/workflow/workflowactions/inventory/allocate.go:179-222`

## Example Configurations

### Basic Reservation

```json
{
  "inventory_items": [
    {
      "product_id": "product-uuid",
      "quantity": 10
    }
  ],
  "allocation_mode": "reserve",
  "allocation_strategy": "fifo",
  "reservation_duration_hours": 24,
  "priority": "medium"
}
```

### Order Allocation

```json
{
  "inventory_items": [
    {
      "product_id": "product-1-uuid",
      "quantity": 5,
      "warehouse_id": "warehouse-uuid"
    },
    {
      "product_id": "product-2-uuid",
      "quantity": 3
    }
  ],
  "allocation_mode": "allocate",
  "allocation_strategy": "fifo",
  "allow_partial": false,
  "priority": "high",
  "reference_id": "order-uuid",
  "reference_type": "order"
}
```

### From Line Item (Automated)

```json
{
  "source_from_line_item": true,
  "allocation_mode": "allocate",
  "allocation_strategy": "fifo",
  "allow_partial": true,
  "priority": "high"
}
```

This extracts product and quantity from the triggering order line item.

## Asynchronous Processing

The allocation system uses a two-phase approach:

```
┌─────────────┐     ┌──────────────┐     ┌─────────────────┐
│   Execute() │────▶│   RabbitMQ   │────▶│ProcessAllocation│
└─────────────┘     └──────────────┘     └─────────────────┘
       │                    │                    │
   Returns                Queued            Background
   Tracking ID          Processing          Processing
```

### Phase 1: Execute()

- Validates configuration
- Checks idempotency
- Queues request to RabbitMQ
- Returns immediately with tracking ID

### Phase 2: ProcessAllocation()

- Background worker processes asynchronously
- Performs actual inventory allocation
- Updates database with results

## Idempotency

Each request has an idempotency key: `{ExecutionID}_{RuleID}_{ActionType}`

- Prevents duplicate processing if same request sent multiple times
- Critical for preventing double-allocation
- Results stored in `allocation_results` table

## Database Tables

### inventory_items

| Column | Description |
|--------|-------------|
| `quantity` | Total quantity |
| `reserved_quantity` | Temporarily held |
| `allocated_quantity` | Permanently committed |

**Available** = `quantity - reserved_quantity - allocated_quantity`

### allocation_results

Tracks allocation request results for idempotency.

### inventory_transactions

Audit trail of all inventory movements.

## Response Format

### Immediate Response (Queued)

```json
{
  "allocation_id": "uuid",
  "status": "queued",
  "idempotency_key": "exec-id_rule-id_allocate_inventory",
  "priority": 10,
  "message": "Allocation request queued for processing"
}
```

### Final Result (After Processing)

```json
{
  "allocation_id": "uuid",
  "status": "success",
  "allocated_items": [
    {
      "product_id": "uuid",
      "quantity_allocated": 10,
      "location_id": "uuid"
    }
  ],
  "failed_items": [],
  "total_requested": 10,
  "total_allocated": 10,
  "execution_time_ms": 45
}
```

### Failure Response

```json
{
  "product_id": "uuid",
  "requested_quantity": 100,
  "available_quantity": 50,
  "reason": "insufficient_inventory",
  "error_message": "Only 50 available, 100 requested"
}
```

## Error Handling

### Failure Reasons

| Reason | Description |
|--------|-------------|
| `insufficient_inventory` | Not enough stock |
| `transaction_setup_failed` | Database transaction issue |
| `query_failed` | Unable to query inventory |
| `update_failed` | Failed to update records |
| `no_allocation` | Unable to allocate any inventory |

### Retry Logic

- Default max retries: 3
- Exponential backoff
- Dead letter queue for permanent failures

## Safety Features

| Feature | Description |
|---------|-------------|
| **Row-level locking** | `FOR UPDATE` prevents race conditions |
| **Read Committed isolation** | Balances consistency/performance |
| **Atomic operations** | All-or-nothing within transactions |
| **Idempotency keys** | Prevents duplicate processing |
| **Audit trail** | Complete transaction history |

## Current Limitations

1. **Query Limit**: Fetches max 10 inventory items at a time
2. **Simplified Strategies**: nearest_expiry and lowest_cost use FIFO
3. **No Auto-Cleanup**: Expired reservations need separate cleanup process

## Use Cases

1. **Order fulfillment** - Allocate inventory when order is confirmed
2. **Shopping cart** - Reserve inventory while customer checks out
3. **Transfer orders** - Allocate from source warehouse
4. **Pre-orders** - Reserve inventory for upcoming products

## Related Documentation

- [Database Schema](../database-schema.md) - allocation_results table
- [Architecture](../architecture.md) - Queue processing and async handlers
- [Rules](../configuration/rules.md) - How to create automation rules

## Key Files

| File | Purpose |
|------|---------|
| `business/sdk/workflow/workflowactions/inventory/allocate.go` | Handler implementation |
| `business/domain/inventory/inventoryitembus/` | Inventory business layer |
