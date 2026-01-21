# Plan: Fully Qualified Column Names in Table Builder

## Overview

Change the table builder to return data row keys using fully qualified `table.column` format instead of simple column names. This ensures consistency between config references (filters, sorts, LabelColumn) and actual data keys.

## Repository Root

All paths in this plan are relative to:
```
/Users/jaketimmer/src/work/superior/ichor/ichor/
```

## Current Behavior

```go
// Config
{Name: "name", TableColumn: "products.name"}

// Returns
{"name": "Widget", "selling_price": 99.99}

// But filters/sorts use:
{Column: "products.name", Operator: "eq", Value: "Widget"}
```

## Target Behavior

```go
// Config (unchanged)
{Name: "name", TableColumn: "products.name"}

// Returns (NEW: qualified keys)
{"products.name": "Widget", "product_costs.selling_price": 99.99}

// Filters/sorts (unchanged - now consistent)
{Column: "products.name", Operator: "eq", Value: "Widget"}
```

## Design Rules

1. **No Alias + Has TableColumn** → Use `TableColumn` as key
2. **Has Alias** → Use `Alias` as key (explicit override)
3. **No Alias + No TableColumn** → Use `Name` as key (fallback for simple cases)

---

## Validation Patterns (Run After Each Phase)

### Pattern 1: LabelColumn must be `table.column` format
```bash
cd /Users/jaketimmer/src/work/superior/ichor/ichor
# Find LabelColumn values that are NOT fully qualified (missing dot)
grep -rn 'LabelColumn:' --include="*.go" business/sdk/dbtest/ | grep -vE 'LabelColumn:\s*"[a-z_]+\.[a-z_]+"'
# Expected: Should return 0 matches after fix
```

### Pattern 2: ValueColumn must be `table.column` format
```bash
grep -rn 'ValueColumn:' --include="*.go" business/sdk/dbtest/ | grep -vE 'ValueColumn:\s*"[a-z_]+\.[a-z_]+"'
# Expected: Should return 0 matches after fix
```

### Pattern 3: SourceColumn (AutoPopulate) must be `table.column` format
```bash
grep -rn 'SourceColumn:' --include="*.go" business/sdk/dbtest/ | grep -vE 'SourceColumn:\s*"[a-z_]+\.[a-z_]+"'
# Expected: Should return 0 matches after fix
```

### Pattern 4: Count violations (quick check)
```bash
# Before fix - count how many need updating:
grep -rn 'LabelColumn:\|ValueColumn:\|SourceColumn:' --include="*.go" business/sdk/dbtest/ | grep -vE '\w+\.\w+' | wc -l
```

### Pattern 5: VisualSettings keys without dots (potential violations)
```bash
# Find simple keys in VisualSettings Columns maps that might need updating
# Note: Some simple keys are valid (aliases), so cross-reference with alias list
grep -rn 'Columns: map\[string\]tablebuilder.ColumnConfig{' -A 100 --include="*.go" business/sdk/dbtest/ | grep -E '^\s*"[a-z_]+":\s*\{' | grep -v '\.' | head -50
```

### Pattern 6: Positive validation - all LabelColumn/ValueColumn have table.column format
```bash
# These should match ALL occurrences after fix (count should equal total)
grep -rn 'LabelColumn:\s*"[a-z_]\+\.[a-z_]\+"' --include="*.go" business/sdk/dbtest/ | wc -l
grep -rn 'ValueColumn:\s*"[a-z_]\+\.[a-z_]\+"' --include="*.go" business/sdk/dbtest/ | wc -l
grep -rn 'SourceColumn:\s*"[a-z_]\+\.[a-z_]\+"' --include="*.go" business/sdk/dbtest/ | wc -l

# Compare against total count (should be equal after fix)
grep -rn 'LabelColumn:' --include="*.go" business/sdk/dbtest/ | wc -l
grep -rn 'ValueColumn:' --include="*.go" business/sdk/dbtest/ | wc -l
grep -rn 'SourceColumn:' --include="*.go" business/sdk/dbtest/ | wc -l
```

---

## Phase 1: Core Query Builder Changes

### File: `business/sdk/tablebuilder/builder.go`

**Function: `buildSelectColumns` (~line 130)**

Find:
```go
for _, col := range config.Columns {
    colExpr := goqu.T(baseTable).Col(col.Name)

    if col.Alias != "" {
        cols = append(cols, colExpr.As(col.Alias))
    } else {
        cols = append(cols, colExpr)
    }
}
```

Replace with:
```go
for _, col := range config.Columns {
    colExpr := goqu.T(baseTable).Col(col.Name)

    if col.Alias != "" {
        cols = append(cols, colExpr.As(col.Alias))
    } else if col.TableColumn != "" {
        cols = append(cols, colExpr.As(col.TableColumn))
    } else {
        cols = append(cols, colExpr)
    }
}
```

**Function: `buildForeignColumns` (~line 150)**

Find:
```go
for _, col := range ft.Columns {
    colExpr := goqu.T(tableRef).Col(col.Name)

    if col.Alias != "" {
        cols = append(cols, colExpr.As(col.Alias))
    } else {
        cols = append(cols, colExpr)
    }
}
```

Replace with:
```go
for _, col := range ft.Columns {
    colExpr := goqu.T(tableRef).Col(col.Name)

    if col.Alias != "" {
        cols = append(cols, colExpr.As(col.Alias))
    } else if col.TableColumn != "" {
        cols = append(cols, colExpr.As(col.TableColumn))
    } else {
        cols = append(cols, colExpr)
    }
}
```

---

### File: `business/sdk/tablebuilder/store.go`

**Function: `getFieldName` (~line 534)**

Find:
```go
func getFieldName(col ColumnDefinition) string {
    if col.Alias != "" {
        return col.Alias
    }
    return col.Name
}
```

Replace with:
```go
func getFieldName(col ColumnDefinition) string {
    if col.Alias != "" {
        return col.Alias
    }
    if col.TableColumn != "" {
        return col.TableColumn
    }
    return col.Name
}
```

---

## Phase 2: Update VisualSettings Keys

### Transformation Rule

For each `VisualSettings.Columns` map:
- If column has **Alias**: key = Alias (no change needed)
- If column has **no Alias but has TableColumn**: key must change from `"name"` to `"table.column"`

### Example Transformation

**Before:**
```go
VisualSettings: tablebuilder.VisualSettings{
    Columns: map[string]tablebuilder.ColumnConfig{
        "id": {
            Name:   "id",
            Header: "ID",
        },
        "name": {
            Name:   "name",
            Header: "Product Name",
        },
    },
},
```

**After:**
```go
VisualSettings: tablebuilder.VisualSettings{
    Columns: map[string]tablebuilder.ColumnConfig{
        "products.id": {
            Name:   "products.id",
            Header: "ID",
        },
        "products.name": {
            Name:   "products.name",
            Header: "Product Name",
        },
    },
},
```

### File: `business/sdk/dbtest/seedmodels/tables.go`

#### ProductsTableConfig (line ~36)
| Old Key | New Key | TableColumn Reference |
|---------|---------|----------------------|
| `"id"` | `"products.id"` | line 27 |
| `"name"` | `"products.name"` | line 28 |
| `"sku"` | `"products.sku"` | line 29 |
| `"is_active"` | `"products.is_active"` | line 30 |

#### InventoryItemsTableConfig (line ~127)
Columns with aliases keep their alias as key. Columns without alias:
| Old Key | New Key |
|---------|---------|
| `"inventory_items.id"` | Keep (has TableColumn) |
| `"inventory_items.reorder_point"` | Keep |
| `"inventory_items.maximum_stock"` | Keep |

#### OrdersTableConfig (line ~461)
| Old Key | New Key | Notes |
|---------|---------|-------|
| `"orders.id"` | Keep | |
| `"order_number"` | Keep | Has alias |
| `"orders.due_date"` | Keep | |
| `"orders.customer_id"` | Keep | |
| `"customer_name"` | Keep | Has alias |
| `"status_name"` | Keep | Has alias |

#### SuppliersTableConfig (line ~684)
| Old Key | New Key |
|---------|---------|
| `"id"` | `"suppliers.id"` |
| `"name"` | `"suppliers.name"` |
| `"payment_terms"` | `"suppliers.payment_terms"` |
| `"lead_time_days"` | `"suppliers.lead_time_days"` |
| `"rating"` | `"suppliers.rating"` |
| `"is_active"` | `"suppliers.is_active"` |
| `"contact_infos_id"` | `"suppliers.contact_infos_id"` |
| `"created_date"` | `"suppliers.created_date"` |
| `"updated_date"` | `"suppliers.updated_date"` |
| `"primary_phone_number"` | Keep (has alias) |
| `"secondary_phone_number"` | Keep (has alias) |
| `"email_address"` | Keep (has alias) |

#### OrderLineItemsTableConfig (line ~952)
| Old Key | New Key |
|---------|---------|
| `"id"` | `"order_line_items.id"` |
| `"order_id"` | `"order_line_items.order_id"` |
| `"product_id"` | `"order_line_items.product_id"` |
| `"quantity"` | `"order_line_items.quantity"` |
| `"discount"` | `"order_line_items.discount"` |
| `"line_item_fulfillment_statuses_id"` | `"order_line_items.line_item_fulfillment_statuses_id"` |
| `"created_by"` | `"order_line_items.created_by"` |
| `"created_date"` | `"order_line_items.created_date"` |
| `"updated_by"` | `"order_line_items.updated_by"` |
| `"updated_date"` | `"order_line_items.updated_date"` |
| `"order_number"` | Keep (has alias) |
| `"product_name"` | Keep (has alias) |
| `"fulfillment_status_name"` | Keep (has alias) |

#### ProductCategoriesTableConfig (line ~1354)
| Old Key | New Key |
|---------|---------|
| `"id"` | `"product_categories.id"` |
| `"name"` | `"product_categories.name"` |
| `"description"` | `"product_categories.description"` |
| `"created_date"` | `"product_categories.created_date"` |
| `"updated_date"` | `"product_categories.updated_date"` |

#### AssetsListTableConfig (line ~1468)
| Old Key | New Key |
|---------|---------|
| `"id"` | `"assets.id"` |
| `"serial_number"` | `"assets.serial_number"` |
| `"last_maintenance_time"` | `"assets.last_maintenance_time"` |
| `"asset_name"` | Keep (has alias) |
| `"asset_type_name"` | Keep (has alias) |
| `"condition_name"` | Keep (has alias) |

#### AssetsRequestsTableConfig (line ~1631)
| Old Key | New Key |
|---------|---------|
| `"id"` | `"user_assets.id"` |
| `"date_received"` | `"user_assets.date_received"` |
| `"last_maintenance"` | `"user_assets.last_maintenance"` |
| All others | Keep (have aliases) |

#### HrEmployeesTableConfig (line ~1883)
| Old Key | New Key |
|---------|---------|
| `"id"` | `"users.id"` |
| `"first_name"` | `"users.first_name"` |
| `"last_name"` | `"users.last_name"` |
| `"email"` | `"users.email"` |
| `"enabled"` | `"users.enabled"` |
| `"date_hired"` | `"users.date_hired"` |
| `"title_name"` | Keep (has alias) |
| `"office_name"` | Keep (has alias) |

#### HrOfficesTableConfig (line ~2100)
| Old Key | New Key |
|---------|---------|
| `"id"` | `"offices.id"` |
| `"name"` | `"offices.name"` |
| `"street_line_1"` | Keep (has alias) |
| `"city_name"` | Keep (has alias) |
| `"region_name"` | Keep (has alias) |
| `"country_name"` | Keep (has alias) |

#### InventoryWarehousesTableConfig (line ~2279)
| Old Key | New Key |
|---------|---------|
| `"id"` | `"warehouses.id"` |
| `"name"` | `"warehouses.name"` |
| `"is_active"` | `"warehouses.is_active"` |
| `"created_date"` | `"warehouses.created_date"` |
| `"updated_date"` | `"warehouses.updated_date"` |
| `"street_line_1"` | Keep (has alias) |
| `"postal_code"` | `"streets.postal_code"` |
| `"city_name"` | Keep (has alias) |
| `"region_code"` | Keep (has alias) |
| `"created_by_username"` | Keep (has alias) |

#### InventoryAdjustmentsTableConfig (line ~2477)
| Old Key | New Key |
|---------|---------|
| `"id"` | `"inventory_adjustments.id"` |
| `"quantity_change"` | `"inventory_adjustments.quantity_change"` |
| `"reason_code"` | `"inventory_adjustments.reason_code"` |
| `"notes"` | `"inventory_adjustments.notes"` |
| `"adjustment_date"` | `"inventory_adjustments.adjustment_date"` |
| `"created_date"` | `"inventory_adjustments.created_date"` |
| All others | Keep (have aliases) |

#### InventoryTransfersTableConfig (line ~2731)
| Old Key | New Key |
|---------|---------|
| `"id"` | `"transfer_orders.id"` |
| `"quantity"` | `"transfer_orders.quantity"` |
| `"status"` | `"transfer_orders.status"` |
| `"transfer_date"` | `"transfer_orders.transfer_date"` |
| `"created_date"` | `"transfer_orders.created_date"` |
| All others | Keep (have aliases) |

#### SalesCustomersTableConfig (line ~3032)
| Old Key | New Key |
|---------|---------|
| `"id"` | `"customers.id"` |
| `"name"` | `"customers.name"` |
| `"notes"` | `"customers.notes"` |
| `"created_date"` | `"customers.created_date"` |
| `"updated_date"` | `"customers.updated_date"` |
| All others | Keep (have aliases) |

#### PurchaseOrderTableConfig (line ~3259)
| Old Key | New Key |
|---------|---------|
| `"id"` | `"purchase_orders.id"` |
| `"order_number"` | `"purchase_orders.order_number"` |
| `"supplier_id"` | `"purchase_orders.supplier_id"` |
| `"purchase_order_status_id"` | `"purchase_orders.purchase_order_status_id"` |
| `"delivery_warehouse_id"` | `"purchase_orders.delivery_warehouse_id"` |
| `"order_date"` | `"purchase_orders.order_date"` |
| `"expected_delivery_date"` | `"purchase_orders.expected_delivery_date"` |
| `"actual_delivery_date"` | `"purchase_orders.actual_delivery_date"` |
| `"subtotal"` | `"purchase_orders.subtotal"` |
| `"tax_amount"` | `"purchase_orders.tax_amount"` |
| `"shipping_cost"` | `"purchase_orders.shipping_cost"` |
| `"total_amount"` | `"purchase_orders.total_amount"` |
| `"requested_by"` | `"purchase_orders.requested_by"` |
| `"approved_by"` | `"purchase_orders.approved_by"` |
| `"approved_date"` | `"purchase_orders.approved_date"` |
| `"notes"` | `"purchase_orders.notes"` |
| `"created_date"` | `"purchase_orders.created_date"` |
| `"updated_date"` | `"purchase_orders.updated_date"` |
| All others | Keep (have aliases) |

#### PurchaseOrderLineItemTableConfig (line ~3671)
| Old Key | New Key |
|---------|---------|
| `"id"` | `"purchase_order_line_items.id"` |
| `"purchase_order_id"` | `"purchase_order_line_items.purchase_order_id"` |
| `"supplier_product_id"` | `"purchase_order_line_items.supplier_product_id"` |
| `"quantity_ordered"` | `"purchase_order_line_items.quantity_ordered"` |
| `"quantity_received"` | `"purchase_order_line_items.quantity_received"` |
| `"quantity_cancelled"` | `"purchase_order_line_items.quantity_cancelled"` |
| `"unit_cost"` | `"purchase_order_line_items.unit_cost"` |
| `"discount"` | `"purchase_order_line_items.discount"` |
| `"line_total"` | `"purchase_order_line_items.line_total"` |
| `"line_item_status_id"` | `"purchase_order_line_items.line_item_status_id"` |
| `"expected_delivery_date"` | `"purchase_order_line_items.expected_delivery_date"` |
| `"actual_delivery_date"` | `"purchase_order_line_items.actual_delivery_date"` |
| `"notes"` | `"purchase_order_line_items.notes"` |
| `"created_by"` | `"purchase_order_line_items.created_by"` |
| `"created_date"` | `"purchase_order_line_items.created_date"` |
| `"updated_by"` | `"purchase_order_line_items.updated_by"` |
| `"updated_date"` | `"purchase_order_line_items.updated_date"` |
| All others | Keep (have aliases) |

#### ProcurementOpenApprovalsTableConfig (line ~4056)
Same pattern as PurchaseOrderTableConfig - update `"purchase_orders.*"` keys

#### ProcurementClosedApprovalsTableConfig (line ~4407)
Same pattern as PurchaseOrderTableConfig - update `"purchase_orders.*"` keys

#### ProductsWithPricesLookup (line ~4836)
| Old Key | New Key |
|---------|---------|
| `"id"` | `"products.id"` |
| `"name"` | `"products.name"` |
| `"description"` | `"products.description"` |
| `"sku"` | `"products.sku"` |
| `"is_active"` | `"products.is_active"` |
| `"selling_price"` | `"product_costs.selling_price"` |
| `"purchase_cost"` | `"product_costs.purchase_cost"` |
| `"currency_code"` | Keep (has alias) |

---

### File: `business/sdk/dbtest/model.go`

#### adminUsersTableConfig (line ~783)
| Old Key | New Key |
|---------|---------|
| `"full_name"` | Keep (computed/alias) |
| `"id"` | `"users.id"` |
| `"username"` | `"users.username"` |
| `"first_name"` | `"users.first_name"` |
| `"last_name"` | `"users.last_name"` |
| `"email"` | `"users.email"` |
| `"enabled"` | `"users.enabled"` |
| `"date_hired"` | `"users.date_hired"` |
| `"created_date"` | `"users.created_date"` |
| `"title_name"` | Keep (has alias) |
| `"office_name"` | Keep (has alias) |
| `"approval_status_name"` | Keep (has alias) |

#### adminRolesTableConfig (line ~954)
| Old Key | New Key |
|---------|---------|
| `"name"` | `"roles.name"` |
| `"id"` | `"roles.id"` |
| `"description"` | `"roles.description"` |

#### adminTableAccessTableConfig (line ~1041)
| Old Key | New Key |
|---------|---------|
| `"id"` | `"table_access.id"` |
| `"table_name"` | `"table_access.table_name"` |
| `"can_create"` | `"table_access.can_create"` |
| `"can_read"` | `"table_access.can_read"` |
| `"can_update"` | `"table_access.can_update"` |
| `"can_delete"` | `"table_access.can_delete"` |
| `"role_name"` | Keep (has alias) |

#### adminAuditTableConfig (line ~1199)
| Old Key | New Key |
|---------|---------|
| `"executed_at"` | `"automation_executions.executed_at"` |
| `"id"` | `"automation_executions.id"` |
| `"entity_type"` | `"automation_executions.entity_type"` |
| `"status"` | `"automation_executions.status"` |
| `"error_message"` | `"automation_executions.error_message"` |
| `"execution_time_ms"` | `"automation_executions.execution_time_ms"` |
| `"rule_name"` | Keep (has alias) |

#### adminConfigTableConfig (line ~1362)
| Old Key | New Key |
|---------|---------|
| `"id"` | `"table_configs.id"` |
| `"name"` | `"table_configs.name"` |
| `"description"` | `"table_configs.description"` |
| `"created_date"` | `"table_configs.created_date"` |
| `"updated_date"` | `"table_configs.updated_date"` |
| `"created_by_username"` | Keep (has alias) |
| `"updated_by_username"` | Keep (has alias) |

---

## Phase 3: Update LookupConfig (LabelColumn, ValueColumn)

### File: `business/sdk/dbtest/seedmodels/tables.go`

Each `LookupConfig` has an `Entity` field showing the target table. Use that to derive the qualified column names.

| Line | Entity | Old LabelColumn | New LabelColumn | Old ValueColumn | New ValueColumn |
|------|--------|-----------------|-----------------|-----------------|-----------------|
| 201 | `products.products` | `"name"` | `"products.name"` | `"id"` | `"products.id"` |
| 356 | `sales.order_fulfillment_statuses` | `"name"` | `"order_fulfillment_statuses.name"` | `"id"` | `"order_fulfillment_statuses.id"` |
| 368 | `sales.customers` | `"notes"` | `"customers.notes"` | `"id"` | `"customers.id"` |
| 380 | `sales.customers` | `"name"` | `"customers.name"` | `"id"` | `"customers.id"` |
| 392 | `core.contact_infos` | `"email_address"` | `"contact_infos.email_address"` | `"id"` | `"contact_infos.id"` |
| 404 | `geography.addresses` | `"street"` | `"addresses.street"` | `"id"` | `"addresses.id"` |
| 585 | `sales.customers` | `"name"` | `"customers.name"` | `"id"` | `"customers.id"` |
| 597 | `sales.order_fulfillment_statuses` | `"name"` | `"order_fulfillment_statuses.name"` | `"id"` | `"order_fulfillment_statuses.id"` |
| 838 | `core.contact_infos` | `"email_address"` | `"contact_infos.email_address"` | `"id"` | `"contact_infos.id"` |
| 1197 | `sales.orders` | `"number"` | `"orders.number"` | `"id"` | `"orders.id"` |
| 1209 | `products.products` | `"name"` | `"products.name"` | `"id"` | `"products.id"` |
| 1221 | `sales.line_item_fulfillment_statuses` | `"name"` | `"line_item_fulfillment_statuses.name"` | `"id"` | `"line_item_fulfillment_statuses.id"` |
| 1251 | `sales.customers` | `"name"` | `"customers.name"` | `"id"` | `"customers.id"` |
| 3467 | `procurement.suppliers` | `"name"` | `"suppliers.name"` | `"id"` | `"suppliers.id"` |
| 3479 | `procurement.purchase_order_statuses` | `"name"` | `"purchase_order_statuses.name"` | `"id"` | `"purchase_order_statuses.id"` |
| 3491 | `inventory.warehouses` | `"name"` | `"warehouses.name"` | `"id"` | `"warehouses.id"` |
| 3911 | `procurement.purchase_orders` | `"order_number"` | `"purchase_orders.order_number"` | `"id"` | `"purchase_orders.id"` |
| 3923 | `procurement.supplier_products` | `"supplier_part_number"` | `"supplier_products.supplier_part_number"` | `"id"` | `"supplier_products.id"` |
| 3947 | `procurement.purchase_order_line_item_statuses` | `"name"` | `"purchase_order_line_item_statuses.name"` | `"id"` | `"purchase_order_line_item_statuses.id"` |
| 3995 | `procurement.suppliers` | `"name"` | `"suppliers.name"` | `"id"` | `"suppliers.id"` |
| 4013 | `products.products` | `"name"` | `"products.name"` | `"id"` | `"products.id"` |
| 4250 | `procurement.suppliers` | `"name"` | `"suppliers.name"` | `"id"` | `"suppliers.id"` |
| 4262 | `procurement.purchase_order_statuses` | `"name"` | `"purchase_order_statuses.name"` | `"id"` | `"purchase_order_statuses.id"` |
| 4274 | `inventory.warehouses` | `"name"` | `"warehouses.name"` | `"id"` | `"warehouses.id"` |
| 4640 | `procurement.suppliers` | `"name"` | `"suppliers.name"` | `"id"` | `"suppliers.id"` |
| 4652 | `procurement.purchase_order_statuses` | `"name"` | `"purchase_order_statuses.name"` | `"id"` | `"purchase_order_statuses.id"` |
| 4664 | `inventory.warehouses` | `"name"` | `"warehouses.name"` | `"id"` | `"warehouses.id"` |

---

### File: `business/sdk/dbtest/seedmodels/forms.go`

#### Line 266-270: product_id dropdown
```go
// BEFORE
DropdownConfig: &formfieldbus.DropdownConfig{
    TableConfigName: "products_with_prices_lookup",
    LabelColumn:     "products.name",  // Already correct!
    ValueColumn:     "id",
    AutoPopulate: []formfieldbus.AutoPopulateMapping{
        {SourceColumn: "selling_price", TargetField: "unit_price"},
        {SourceColumn: "description", TargetField: "description"},
    },
},

// AFTER
DropdownConfig: &formfieldbus.DropdownConfig{
    TableConfigName: "products_with_prices_lookup",
    LabelColumn:     "products.name",
    ValueColumn:     "products.id",
    AutoPopulate: []formfieldbus.AutoPopulateMapping{
        {SourceColumn: "product_costs.selling_price", TargetField: "unit_price"},
        {SourceColumn: "products.description", TargetField: "description"},
    },
},
```

#### Line 345-346: fulfillment_status dropdown
```go
// BEFORE
DropdownConfig: &formfieldbus.DropdownConfig{
    Entity:      "sales.line_item_fulfillment_statuses",
    LabelColumn: "name",
    ValueColumn: "id",
},

// AFTER
DropdownConfig: &formfieldbus.DropdownConfig{
    Entity:      "sales.line_item_fulfillment_statuses",
    LabelColumn: "line_item_fulfillment_statuses.name",
    ValueColumn: "line_item_fulfillment_statuses.id",
},
```

---

## Phase 4: Chart Configs (NO CHANGES NEEDED)

Chart configs in `business/sdk/dbtest/seedmodels/charts.go` use **output aliases** from metrics and group-by clauses, not raw table column names.

Example:
```go
Metrics: []tablebuilder.MetricConfig{
    {Name: "sales", Function: "sum", ...},  // Output alias: "sales"
},
GroupBy: []tablebuilder.GroupByConfig{
    {Column: "orders.created_date", Interval: "month", Alias: "month"},  // Output alias: "month"
},
// ChartVisualSettings references these aliases:
CategoryColumn: "month",      // References GroupBy alias
ValueColumns:   []string{"sales"},  // References Metric name
```

These aliases are independent of the underlying table columns. **No changes required.**

---

## Phase 5: Frontend Updates

### File: `vue/ichor/src/services/form/tableOptionsService.ts`

The proposed `getColumnValue` helper from the dropdown exploration plan is **no longer needed**.

The existing direct access will work because data keys now match:
```typescript
const label = row[labelColumn]  // labelColumn = "products.name", row["products.name"] = "Widget"
```

**Action:** Do NOT implement the fallback helper. Delete the exploration plan file if desired:
```
vue/ichor/.claude/plans/dropdown-column-names-exploration.md
```

### Frontend Hardcoded Column References

Search for any hardcoded column access:
```bash
cd /Users/jaketimmer/src/work/superior/ichor/vue/ichor
grep -rn 'row\["[a-z_]*"\]' --include="*.ts" --include="*.vue" src/
```

Any matches need review - they may need to change to qualified names or use metadata lookup.

---

## Phase 6: Update Tests

### Backend Tests

#### `business/sdk/tablebuilder/builder_test.go`
Update expected SQL to include `AS "table.column"` clauses.

#### `business/sdk/tablebuilder/store_test.go`
Update expected `Field` values in metadata assertions.

### Frontend Tests
```bash
cd /Users/jaketimmer/src/work/superior/ichor/vue/ichor
grep -rn '"name"\|"id"' --include="*.test.ts" --include="*.spec.ts" src/
```

---

## Execution Checklist

### Phase 1: Core Changes
- [ ] Edit `business/sdk/tablebuilder/builder.go` - `buildSelectColumns`
- [ ] Edit `business/sdk/tablebuilder/builder.go` - `buildForeignColumns`
- [ ] Edit `business/sdk/tablebuilder/store.go` - `getFieldName`
- [ ] Run `go build ./...`
- [ ] Run `go test ./business/sdk/tablebuilder/...` (expect failures)

### Phase 2: VisualSettings Keys
- [ ] Update `business/sdk/dbtest/seedmodels/tables.go` - all configs listed above
- [ ] Update `business/sdk/dbtest/model.go` - admin configs
- [ ] Run `go build ./...`

### Phase 3: Lookup/Dropdown Configs
- [ ] Update `business/sdk/dbtest/seedmodels/tables.go` - all LookupConfigs (28 locations)
- [ ] Update `business/sdk/dbtest/seedmodels/forms.go` - 2 DropdownConfigs
- [ ] Run validation patterns - should return 0 violations
- [ ] Run `go build ./...`

### Phase 4: Charts
- [ ] Verify no changes needed (already confirmed)

### Phase 5: Frontend
- [ ] Delete `vue/ichor/.claude/plans/dropdown-column-names-exploration.md`
- [ ] Search for hardcoded column references
- [ ] Run `npm run type-check`
- [ ] Run `npm run test:unit`

### Phase 6: Tests
- [ ] Update backend test assertions
- [ ] Run `go test ./...`

### Final Verification
- [ ] Start backend: `go run cmd/main.go`
- [ ] Seed database
- [ ] Start frontend: `npm run dev`
- [ ] Test `/sales/orders` table
- [ ] Test `/sales/orders/new` form with dropdown
- [ ] Verify product dropdown shows names
- [ ] Verify auto-populate works

---

## Summary Statistics

| Category | Count |
|----------|-------|
| Core code changes | 3 functions |
| VisualSettings configs | ~25 |
| LookupConfig updates | 28 |
| Form DropdownConfig updates | 2 |
| Chart changes | 0 |
| Files to modify | 4 Go files |
| Estimated lines changed | ~600-800 |
