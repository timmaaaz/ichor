# Plan: Fix JSON Tag Naming Inconsistency (camelCase to snake_case)

## How to Continue This Work

**To start or resume work on this plan, use this prompt:**
```
Continue fixing camelCase JSON tags per .claude/plans/json-tag-snake-case-fix.md
- Check the Progress Tracking table to see which domains are complete
- Work on the next incomplete domain
- After each domain, update the progress table and verify with grep
- After all domains, run make test
```

**To work on a specific domain:**
```
Fix the [Domain Name] domain JSON tags from .claude/plans/json-tag-snake-case-fix.md
```

---

## Problem Summary
Form validation fails when backend models use camelCase JSON tags (e.g., `json:"regionID"`) but form fields use snake_case names (e.g., `region_id`). The validation system extracts required field names from JSON struct tags, so they must match.

## Scope
- **Total Files**: 17 model.go files in app/domain/
- **Total Tags to Fix**: 234 camelCase JSON tags
- **No test files require changes** (verified via grep)

## Transformation Rule
Change JSON tags from camelCase to snake_case. Do NOT change Go field names.

```go
// Before
RegionID string `json:"regionID" validate:"required"`

// After
RegionID string `json:"region_id" validate:"required"`
```

---

## Progress Tracking

**Status**: COMPLETE

| Domain | Files | Tags | Status |
|--------|-------|------|--------|
| Geography | 2 | 11 | ✅ Complete |
| Core | 3 | 17 | ✅ Complete |
| HR | 1 | 6 | ✅ Complete |
| Procurement | 4 | 73 | ✅ Complete |
| Config | 4 | 97 | ✅ Complete |
| Introspection | 1 | 15 | ✅ Complete |
| Data/Check | 2 | 15 | ✅ Complete |

**Legend**: ⬜ Not Started | 🔄 In Progress | ✅ Complete

### Per-Domain Completion Checklist
After completing each domain:
1. Update the status in the table above (change ⬜ to ✅)
2. Mark individual file checkboxes as [x]
3. Run verification: `grep -rn 'json:"[a-z]\+[A-Z]' app/domain/{domain}/` to confirm no camelCase remains in that domain

### Final Verification (after all domains complete)
```bash
# Verify no camelCase JSON tags remain anywhere
grep -rn 'json:"[a-z]\+[A-Z]' app/domain/

# Run full test suite
make test
```

---

## Complete Checklist by File

### Geography Domain

#### [ ] app/domain/geography/cityapp/model.go (2 tags)
| Line | Current | Change To |
|------|---------|-----------|
| 53 | `json:"regionID"` | `json:"region_id"` |
| 89 | `json:"regionID"` | `json:"region_id"` |

#### [ ] app/domain/geography/timezoneapp/model.go (9 tags)
| Line | Current | Change To |
|------|---------|-----------|
| 26 | `json:"displayName"` | `json:"display_name"` |
| 27 | `json:"utcOffset"` | `json:"utc_offset"` |
| 28 | `json:"isActive"` | `json:"is_active"` |
| 62 | `json:"displayName"` | `json:"display_name"` |
| 63 | `json:"utcOffset"` | `json:"utc_offset"` |
| 64 | `json:"isActive"` | `json:"is_active"` |
| 95 | `json:"displayName"` | `json:"display_name"` |
| 96 | `json:"utcOffset"` | `json:"utc_offset"` |
| 97 | `json:"isActive"` | `json:"is_active"` |

---

### Core Domain

#### [ ] app/domain/core/pageapp/model.go (9 tags)
| Line | Current | Change To |
|------|---------|-----------|
| 28 | `json:"sortOrder"` | `json:"sort_order"` |
| 29 | `json:"isActive"` | `json:"is_active"` |
| 30 | `json:"showInMenu"` | `json:"show_in_menu"` |
| 75 | `json:"sortOrder"` | `json:"sort_order"` |
| 76 | `json:"isActive"` | `json:"is_active"` |
| 77 | `json:"showInMenu"` | `json:"show_in_menu"` |
| 110 | `json:"sortOrder"` | `json:"sort_order"` |
| 111 | `json:"isActive"` | `json:"is_active"` |
| 112 | `json:"showInMenu"` | `json:"show_in_menu"` |

#### [ ] app/domain/core/rolepageapp/model.go (7 tags)
| Line | Current | Change To |
|------|---------|-----------|
| 23 | `json:"roleId"` | `json:"role_id"` |
| 24 | `json:"pageId"` | `json:"page_id"` |
| 25 | `json:"canAccess"` | `json:"can_access"` |
| 53 | `json:"roleId"` | `json:"role_id"` |
| 54 | `json:"pageId"` | `json:"page_id"` |
| 55 | `json:"canAccess"` | `json:"can_access"` |
| 91 | `json:"canAccess"` | `json:"can_access"` |

#### [ ] app/domain/core/userapp/model.go (1 tag)
| Line | Current | Change To |
|------|---------|-----------|
| 257 | `json:"approvedBy"` | `json:"approved_by"` |

---

### HR Domain

#### [ ] app/domain/hr/homeapp/model.go (6 tags)
| Line | Current | Change To |
|------|---------|-----------|
| 32 | `json:"zipCode"` | `json:"zip_code"` |
| 41 | `json:"userID"` | `json:"user_id"` |
| 44 | `json:"dateCreated"` | `json:"date_created"` |
| 45 | `json:"dateUpdated"` | `json:"date_updated"` |
| 87 | `json:"zipCode"` | `json:"zip_code"` |
| 157 | `json:"zipCode"` | `json:"zip_code"` |

---

### Procurement Domain

#### [ ] app/domain/procurement/purchaseorderstatusapp/model.go (3 tags)
| Line | Current | Change To |
|------|---------|-----------|
| 25 | `json:"sortOrder"` | `json:"sort_order"` |
| 53 | `json:"sortOrder"` | `json:"sort_order"` |
| 78 | `json:"sortOrder"` | `json:"sort_order"` |

#### [ ] app/domain/procurement/purchaseorderlineitemstatusapp/model.go (3 tags)
| Line | Current | Change To |
|------|---------|-----------|
| 25 | `json:"sortOrder"` | `json:"sort_order"` |
| 53 | `json:"sortOrder"` | `json:"sort_order"` |
| 78 | `json:"sortOrder"` | `json:"sort_order"` |

#### [ ] app/domain/procurement/purchaseorderapp/model.go (40 tags)
| Line | Current | Change To |
|------|---------|-----------|
| 35 | `json:"orderNumber"` | `json:"order_number"` |
| 36 | `json:"supplierId"` | `json:"supplier_id"` |
| 37 | `json:"purchaseOrderStatusId"` | `json:"purchase_order_status_id"` |
| 38 | `json:"deliveryWarehouseId"` | `json:"delivery_warehouse_id"` |
| 39 | `json:"deliveryLocationId"` | `json:"delivery_location_id"` |
| 40 | `json:"deliveryStreetId"` | `json:"delivery_street_id"` |
| 41 | `json:"orderDate"` | `json:"order_date"` |
| 42 | `json:"expectedDeliveryDate"` | `json:"expected_delivery_date"` |
| 43 | `json:"actualDeliveryDate"` | `json:"actual_delivery_date"` |
| 45 | `json:"taxAmount"` | `json:"tax_amount"` |
| 46 | `json:"shippingCost"` | `json:"shipping_cost"` |
| 47 | `json:"totalAmount"` | `json:"total_amount"` |
| 49 | `json:"requestedBy"` | `json:"requested_by"` |
| 50 | `json:"approvedBy"` | `json:"approved_by"` |
| 51 | `json:"approvedDate"` | `json:"approved_date"` |
| 53 | `json:"supplierReferenceNumber"` | `json:"supplier_reference_number"` |
| 54 | `json:"createdBy"` | `json:"created_by"` |
| 55 | `json:"updatedBy"` | `json:"updated_by"` |
| 56 | `json:"createdDate"` | `json:"created_date"` |
| 57 | `json:"updatedDate"` | `json:"updated_date"` |
| 122 | `json:"orderNumber"` | `json:"order_number"` |
| 123 | `json:"supplierId"` | `json:"supplier_id"` |
| 124 | `json:"purchaseOrderStatusId"` | `json:"purchase_order_status_id"` |
| 125 | `json:"deliveryWarehouseId"` | `json:"delivery_warehouse_id"` |
| 126 | `json:"deliveryLocationId"` | `json:"delivery_location_id"` |
| 127 | `json:"deliveryStreetId"` | `json:"delivery_street_id"` |
| 128 | `json:"orderDate"` | `json:"order_date"` |
| 129 | `json:"expectedDeliveryDate"` | `json:"expected_delivery_date"` |
| 131 | `json:"taxAmount"` | `json:"tax_amount"` |
| 132 | `json:"shippingCost"` | `json:"shipping_cost"` |
| 133 | `json:"totalAmount"` | `json:"total_amount"` |
| 135 | `json:"requestedBy"` | `json:"requested_by"` |
| 137 | `json:"supplierReferenceNumber"` | `json:"supplier_reference_number"` |
| 138 | `json:"createdBy"` | `json:"created_by"` |
| 139 | `json:"createdDate"` | `json:"created_date"` |
| 263 | `json:"orderNumber"` | `json:"order_number"` |
| 264 | `json:"supplierId"` | `json:"supplier_id"` |
| 265 | `json:"purchaseOrderStatusId"` | `json:"purchase_order_status_id"` |
| 266 | `json:"deliveryWarehouseId"` | `json:"delivery_warehouse_id"` |
| 267 | `json:"deliveryLocationId"` | `json:"delivery_location_id"` |
| 268 | `json:"deliveryStreetId"` | `json:"delivery_street_id"` |
| 269 | `json:"orderDate"` | `json:"order_date"` |
| 270 | `json:"expectedDeliveryDate"` | `json:"expected_delivery_date"` |
| 271 | `json:"actualDeliveryDate"` | `json:"actual_delivery_date"` |
| 273 | `json:"taxAmount"` | `json:"tax_amount"` |
| 274 | `json:"shippingCost"` | `json:"shipping_cost"` |
| 275 | `json:"totalAmount"` | `json:"total_amount"` |
| 277 | `json:"approvedBy"` | `json:"approved_by"` |
| 278 | `json:"approvedDate"` | `json:"approved_date"` |
| 280 | `json:"supplierReferenceNumber"` | `json:"supplier_reference_number"` |
| 281 | `json:"updatedBy"` | `json:"updated_by"` |
| 469 | `json:"approvedBy"` | `json:"approved_by"` |

#### [ ] app/domain/procurement/purchaseorderlineitemapp/model.go (27 tags)
| Line | Current | Change To |
|------|---------|-----------|
| 39 | `json:"purchaseOrderId"` | `json:"purchase_order_id"` |
| 40 | `json:"supplierProductId"` | `json:"supplier_product_id"` |
| 41 | `json:"quantityOrdered"` | `json:"quantity_ordered"` |
| 42 | `json:"quantityReceived"` | `json:"quantity_received"` |
| 43 | `json:"quantityCancelled"` | `json:"quantity_cancelled"` |
| 44 | `json:"unitCost"` | `json:"unit_cost"` |
| 46 | `json:"lineTotal"` | `json:"line_total"` |
| 47 | `json:"lineItemStatusId"` | `json:"line_item_status_id"` |
| 48 | `json:"expectedDeliveryDate"` | `json:"expected_delivery_date"` |
| 49 | `json:"actualDeliveryDate"` | `json:"actual_delivery_date"` |
| 51 | `json:"createdBy"` | `json:"created_by"` |
| 52 | `json:"updatedBy"` | `json:"updated_by"` |
| 53 | `json:"createdDate"` | `json:"created_date"` |
| 54 | `json:"updatedDate"` | `json:"updated_date"` |
| 102 | `json:"purchaseOrderId"` | `json:"purchase_order_id"` |
| 103 | `json:"supplierProductId"` | `json:"supplier_product_id"` |
| 104 | `json:"quantityOrdered"` | `json:"quantity_ordered"` |
| 105 | `json:"unitCost"` | `json:"unit_cost"` |
| 107 | `json:"lineTotal"` | `json:"line_total"` |
| 108 | `json:"lineItemStatusId"` | `json:"line_item_status_id"` |
| 109 | `json:"expectedDeliveryDate"` | `json:"expected_delivery_date"` |
| 111 | `json:"createdBy"` | `json:"created_by"` |
| 192 | `json:"supplierProductId"` | `json:"supplier_product_id"` |
| 193 | `json:"quantityOrdered"` | `json:"quantity_ordered"` |
| 194 | `json:"quantityReceived"` | `json:"quantity_received"` |
| 195 | `json:"quantityCancelled"` | `json:"quantity_cancelled"` |
| 196 | `json:"unitCost"` | `json:"unit_cost"` |
| 198 | `json:"lineTotal"` | `json:"line_total"` |
| 199 | `json:"lineItemStatusId"` | `json:"line_item_status_id"` |
| 200 | `json:"expectedDeliveryDate"` | `json:"expected_delivery_date"` |
| 201 | `json:"actualDeliveryDate"` | `json:"actual_delivery_date"` |
| 203 | `json:"updatedBy"` | `json:"updated_by"` |
| 230 | `json:"receivedBy"` | `json:"received_by"` |

---

### Config Domain

#### [ ] app/domain/config/formapp/model.go (12 tags)
| Line | Current | Change To |
|------|---------|-----------|
| 26 | `json:"isReferenceData"` | `json:"is_reference_data"` |
| 27 | `json:"allowInlineCreate"` | `json:"allow_inline_create"` |
| 65 | `json:"isReferenceData"` | `json:"is_reference_data"` |
| 66 | `json:"allowInlineCreate"` | `json:"allow_inline_create"` |
| 91 | `json:"isReferenceData"` | `json:"is_reference_data"` |
| 92 | `json:"allowInlineCreate"` | `json:"allow_inline_create"` |
| 118 | `json:"isReferenceData"` | `json:"is_reference_data"` |
| 119 | `json:"allowInlineCreate"` | `json:"allow_inline_create"` |
| 143 | `json:"exportedAt"` | `json:"exported_at"` |
| 190 | `json:"importedCount"` | `json:"imported_count"` |
| 191 | `json:"skippedCount"` | `json:"skipped_count"` |
| 192 | `json:"updatedCount"` | `json:"updated_count"` |

#### [ ] app/domain/config/pageconfigapp/model.go (28 tags)
| Line | Current | Change To |
|------|---------|-----------|
| 27 | `json:"userId,omitempty"` | `json:"user_id,omitempty"` |
| 28 | `json:"isDefault"` | `json:"is_default"` |
| 49 | `json:"userId"` | `json:"user_id"` |
| 50 | `json:"isDefault"` | `json:"is_default"` |
| 69 | `json:"userId"` | `json:"user_id"` |
| 70 | `json:"isDefault"` | `json:"is_default"` |
| 158 | `json:"exportedAt"` | `json:"exported_at"` |
| 171 | `json:"pageConfig"` | `json:"page_config"` |
| 179 | `json:"pageConfigId"` | `json:"page_config_id"` |
| 180 | `json:"contentType"` | `json:"content_type"` |
| 182 | `json:"tableConfigId,omitempty"` | `json:"table_config_id,omitempty"` |
| 183 | `json:"formId,omitempty"` | `json:"form_id,omitempty"` |
| 184 | `json:"orderIndex"` | `json:"order_index"` |
| 185 | `json:"parentId,omitempty"` | `json:"parent_id,omitempty"` |
| 187 | `json:"isVisible"` | `json:"is_visible"` |
| 188 | `json:"isDefault"` | `json:"is_default"` |
| 201 | `json:"pageConfigId"` | `json:"page_config_id"` |
| 202 | `json:"actionType"` | `json:"action_type"` |
| 203 | `json:"actionOrder"` | `json:"action_order"` |
| 204 | `json:"isActive"` | `json:"is_active"` |
| 213 | `json:"targetPath"` | `json:"target_path"` |
| 216 | `json:"confirmationPrompt,omitempty"` | `json:"confirmation_prompt,omitempty"` |
| 230 | `json:"targetPath"` | `json:"target_path"` |
| 231 | `json:"itemOrder"` | `json:"item_order"` |
| 266 | `json:"importedCount"` | `json:"imported_count"` |
| 267 | `json:"skippedCount"` | `json:"skipped_count"` |
| 268 | `json:"updatedCount"` | `json:"updated_count"` |

#### [ ] app/domain/config/pagecontentapp/model.go (22 tags)
| Line | Current | Change To |
|------|---------|-----------|
| 27 | `json:"pageConfigId"` | `json:"page_config_id"` |
| 28 | `json:"contentType"` | `json:"content_type"` |
| 30 | `json:"tableConfigId,omitempty"` | `json:"table_config_id,omitempty"` |
| 31 | `json:"formId,omitempty"` | `json:"form_id,omitempty"` |
| 32 | `json:"chartConfigId,omitempty"` | `json:"chart_config_id,omitempty"` |
| 33 | `json:"orderIndex"` | `json:"order_index"` |
| 34 | `json:"parentId,omitempty"` | `json:"parent_id,omitempty"` |
| 36 | `json:"isVisible"` | `json:"is_visible"` |
| 37 | `json:"isDefault"` | `json:"is_default"` |
| 58 | `json:"pageConfigId"` | `json:"page_config_id"` |
| 59 | `json:"contentType"` | `json:"content_type"` |
| 61 | `json:"tableConfigId"` | `json:"table_config_id"` |
| 62 | `json:"formId"` | `json:"form_id"` |
| 63 | `json:"chartConfigId"` | `json:"chart_config_id"` |
| 64 | `json:"orderIndex"` | `json:"order_index"` |
| 65 | `json:"parentId"` | `json:"parent_id"` |
| 67 | `json:"isVisible"` | `json:"is_visible"` |
| 68 | `json:"isDefault"` | `json:"is_default"` |
| 99 | `json:"orderIndex"` | `json:"order_index"` |
| 101 | `json:"isVisible"` | `json:"is_visible"` |
| 102 | `json:"isDefault"` | `json:"is_default"` |

#### [ ] app/domain/config/pageactionapp/model.go (35 tags)
| Line | Current | Change To |
|------|---------|-----------|
| 29 | `json:"pageConfigId"` | `json:"page_config_id"` |
| 30 | `json:"actionType"` | `json:"action_type"` |
| 31 | `json:"actionOrder"` | `json:"action_order"` |
| 32 | `json:"isActive"` | `json:"is_active"` |
| 46 | `json:"actionUrl"` | `json:"action_url"` |
| 49 | `json:"confirmationPrompt,omitempty"` | `json:"confirmation_prompt,omitempty"` |
| 63 | `json:"targetPath"` | `json:"target_path"` |
| 64 | `json:"itemOrder"` | `json:"item_order"` |
| 152 | `json:"pageConfigId"` | `json:"page_config_id"` |
| 153 | `json:"actionOrder"` | `json:"action_order"` |
| 154 | `json:"isActive"` | `json:"is_active"` |
| 157 | `json:"actionUrl"` | `json:"action_url"` |
| 160 | `json:"confirmationPrompt"` | `json:"confirmation_prompt"` |
| 196 | `json:"targetPath"` | `json:"target_path"` |
| 197 | `json:"itemOrder"` | `json:"item_order"` |
| 202 | `json:"pageConfigId"` | `json:"page_config_id"` |
| 203 | `json:"actionOrder"` | `json:"action_order"` |
| 204 | `json:"isActive"` | `json:"is_active"` |
| 248 | `json:"pageConfigId"` | `json:"page_config_id"` |
| 249 | `json:"actionOrder"` | `json:"action_order"` |
| 250 | `json:"isActive"` | `json:"is_active"` |
| 283 | `json:"pageConfigId"` | `json:"page_config_id"` |
| 284 | `json:"actionOrder"` | `json:"action_order"` |
| 285 | `json:"isActive"` | `json:"is_active"` |
| 288 | `json:"actionUrl"` | `json:"action_url"` |
| 291 | `json:"confirmationPrompt"` | `json:"confirmation_prompt"` |
| 330 | `json:"pageConfigId"` | `json:"page_config_id"` |
| 331 | `json:"actionOrder"` | `json:"action_order"` |
| 332 | `json:"isActive"` | `json:"is_active"` |
| 382 | `json:"pageConfigId"` | `json:"page_config_id"` |
| 383 | `json:"actionOrder"` | `json:"action_order"` |
| 384 | `json:"isActive"` | `json:"is_active"` |
| 421 | `json:"actionType"` | `json:"action_type"` |

---

### Introspection Domain

#### [ ] app/domain/introspectionapp/model.go (14 tags)
| Line | Current | Change To |
|------|---------|-----------|
| 33 | `json:"rowCountEstimate"` | `json:"row_count_estimate"` |
| 54 | `json:"dataType"` | `json:"data_type"` |
| 55 | `json:"isNullable"` | `json:"is_nullable"` |
| 56 | `json:"isPrimaryKey"` | `json:"is_primary_key"` |
| 57 | `json:"defaultValue"` | `json:"default_value"` |
| 77 | `json:"foreignKeyName"` | `json:"foreign_key_name"` |
| 78 | `json:"columnName"` | `json:"column_name"` |
| 79 | `json:"referencedSchema"` | `json:"referenced_schema"` |
| 80 | `json:"referencedTable"` | `json:"referenced_table"` |
| 81 | `json:"referencedColumn"` | `json:"referenced_column"` |
| 82 | `json:"relationshipType"` | `json:"relationship_type"` |
| 104 | `json:"foreignKeyColumn"` | `json:"foreign_key_column"` |
| 105 | `json:"constraintName"` | `json:"constraint_name"` |
| 248 | `json:"sortOrder"` | `json:"sort_order"` |
| 305 | `json:"sortOrder"` | `json:"sort_order"` |

---

### Data/Check Domains

#### [ ] app/domain/checkapp/model.go (1 tag)
| Line | Current | Change To |
|------|---------|-----------|
| 11 | `json:"podIP,omitempty"` | `json:"pod_ip,omitempty"` |

#### [ ] app/domain/dataapp/model.go (14 tags)
| Line | Current | Change To |
|------|---------|-----------|
| 187 | `json:"exportedAt"` | `json:"exported_at"` |
| 234 | `json:"importedCount"` | `json:"imported_count"` |
| 235 | `json:"skippedCount"` | `json:"skipped_count"` |
| 236 | `json:"updatedCount"` | `json:"updated_count"` |
| 435 | `json:"pageConfig"` | `json:"page_config"` |
| 436 | `json:"pageActions"` | `json:"page_actions"` |
| 584 | `json:"yAxisIndex,omitempty"` | `json:"y_axis_index,omitempty"` |
| 592 | `json:"previousValue,omitempty"` | `json:"previous_value,omitempty"` |
| 604 | `json:"executionTimeMs"` | `json:"execution_time_ms"` |
| 605 | `json:"rowsProcessed"` | `json:"rows_processed"` |
| 611 | `json:"xCategories"` | `json:"x_categories"` |
| 612 | `json:"yCategories"` | `json:"y_categories"` |
| 621 | `json:"startDate"` | `json:"start_date"` |
| 622 | `json:"endDate"` | `json:"end_date"` |

---

## Implementation Strategy

### Recommended Order (by domain complexity)
1. **Geography** (2 files, 11 tags) - Simple, good for initial validation
2. **Core** (3 files, 17 tags) - Important for page/role functionality
3. **HR** (1 file, 6 tags) - Simple domain
4. **Procurement** (4 files, 73 tags) - Most affected domain, critical for forms
5. **Config** (4 files, 97 tags) - Complex, affects page configurations
6. **Introspection** (1 file, 14 tags) - Database introspection APIs
7. **Data/Check** (2 files, 15 tags) - Utility endpoints

### For Each File
1. Open the file
2. Use find-and-replace for each camelCase tag listed
3. Verify the line numbers match (they may shift if file was modified)
4. Save the file

### Verification
After all changes:
```bash
# 1. Verify no camelCase JSON tags remain
grep -rn 'json:"[a-z]\+[A-Z]' app/domain/

# 2. Run tests (includes compilation check, linting, and vulnerability checks)
make test
```

---

## Files NOT to Modify
- `business/domain/**/model.go` - Business layer uses Go naming, not JSON
- `business/domain/**/stores/**/*.go` - Database layer uses `db` tags

---

## IMPORTANT: Test File Updates Required

**When changing JSON tags, test files with validation error expectations MUST also be updated.**

The Go validation library (`validate` tags) uses JSON field names in error messages. When you change:
```go
PageConfigID string `json:"page_config_id" validate:"required"`  // was "pageConfigId"
```

Test files expecting validation errors must also be updated:
```go
// BEFORE (broken after JSON tag change):
ExpResp: errs.Newf(errs.InvalidArgument, "validate: [{\"field\":\"pageConfigId\",\"error\":\"...\"

// AFTER (matches new JSON tag):
ExpResp: errs.Newf(errs.InvalidArgument, "validate: [{\"field\":\"page_config_id\",\"error\":\"...\"
```

### How to Find Test Files Needing Updates

After changing JSON tags in a domain, search for test files with old field names:
```bash
# For a specific field
grep -rn '"pageConfigId"' api/cmd/services/ichor/tests/

# For all camelCase patterns in test files
grep -rn 'field.*"[a-z]\+[A-Z]' api/cmd/services/ichor/tests/
```

### Test Files Updated in This Plan

The following test files were updated with new field names:
- `api/cmd/services/ichor/tests/geography/cityapi/create_test.go` - `regionID` → `region_id`
- `api/cmd/services/ichor/tests/config/pagecontentapi/create_test.go` - `page_config_id`, `content_type`
- `api/cmd/services/ichor/tests/config/pageactionapi/create_test.go` - `page_config_id`, `action_url`, `label`, `variant`, `alignment`
- `api/cmd/services/ichor/tests/procurement/purchaseorderlineitemapi/create_test.go` - various snake_case fields

### Files Also Changed (SDK level)

- `app/sdk/query/query.go` - `rowsPerPage` → `rows_per_page`

---

## Summary Statistics
| Domain | Files | Tags |
|--------|-------|------|
| Geography | 2 | 11 |
| Core | 3 | 17 |
| HR | 1 | 6 |
| Procurement | 4 | 73 |
| Config | 4 | 97 |
| Introspection | 1 | 15 |
| Data/Check | 2 | 15 |
| **TOTAL** | **17** | **234** |

Note: Some tags appear in multiple structs within the same file (e.g., Entity, NewEntity, UpdateEntity), which is why the per-file counts may seem high.

---

## Post-Implementation Failing Tests

**Test Run Date**: 2026-01-19
**Total Failing Packages**: 38

### Summary of Failure Types

The failures fall into several categories:

1. **Validation Error Field Name Mismatches** - Test files expect camelCase field names in validation error messages but code now returns snake_case
2. **JSON Unmarshalling Failures** - Test code unmarshalling response into structs with camelCase JSON tags fails because API returns snake_case
3. **Unrelated Test Failures** - Some tests fail due to issues unrelated to JSON tag changes (e.g., missing required fields, changed defaults)

### Failing Test Packages

#### Config Domain (2 packages)
- [ ] `config/pageactionapi` - Validation error field names
- [ ] `config/pageconfigapi` - Export tests - JSON unmarshalling fails (camelCase struct tags vs snake_case response)

#### Core Domain (5 packages)
- [ ] `core/pageapi` - Validation error field names
- [ ] `core/roleapi` - Validation error field names
- [ ] `core/tableaccessapi` - Validation error field names
- [ ] `core/userapi` - Validation error field names
- [ ] `core/userroleapi` - Validation error field names

#### Geography Domain (4 packages)
- [ ] `geography/cityapi` - Validation error field names
- [ ] `geography/countryapi` - Validation error field names
- [ ] `geography/regionapi` - Validation error field names
- [ ] `geography/streetapi` - Validation error field names

#### HR Domain (5 packages)
- [ ] `hr/officeapi` - Validation error field names
- [ ] `hr/reportstoapi` - Validation error field names
- [ ] `hr/titleapi` - Validation error field names
- [ ] `hr/userapprovalcommentapi` - Validation error field names
- [ ] `hr/userapprovalstatusapi` - Validation error field names

#### Inventory Domain (10 packages)
- [ ] `inventory/inspectionapi` - Validation error field names
- [ ] `inventory/inventoryadjustmentapi` - Validation error field names
- [ ] `inventory/inventoryitemapi` - Validation error field names
- [ ] `inventory/inventorylocationapi` - Validation error field names
- [ ] `inventory/inventorytransactionapi` - Validation error field names
- [ ] `inventory/lottrackingsapi` - Validation error field names
- [ ] `inventory/serialnumberapi` - Validation error field names
- [ ] `inventory/transferorderapi` - Validation error field names
- [ ] `inventory/warehouseapi` - Validation error field names
- [ ] `inventory/zoneapi` - Validation error field names

#### Procurement Domain (5 packages)
- [ ] `procurement/purchaseorderapi` - Validation error - expects `orderNumber` but got `order_number`
- [ ] `procurement/purchaseorderlineitemapi` - Validation error field names
- [ ] `procurement/purchaseorderlineitemstatusapi` - Validation error field names
- [ ] `procurement/purchaseorderstatusapi` - Validation error field names
- [ ] `procurement/supplierapi` - Validation error field names

#### Sales Domain (2 packages)
- [ ] `sales/orderlineitemsapi` - Mixed - some validation errors, some unrelated (default value changes)
- [ ] `sales/ordersapi` - Multiple failures - validation errors + possible unrelated issues (`order_date` required)

#### Other (3 packages)
- [ ] `data` - Likely JSON unmarshalling issues
- [ ] `formdata/formdataapi` - Validation errors + unrelated (`fulfillment_status_id` missing)
- [ ] `introspectionapi` - Validation error field names

### Specific Failure Examples

#### 1. Validation Error Field Name Mismatch (Most Common)
```
ExpResp: errs.Newf(errs.InvalidArgument, "validate: [{\"field\":\"orderNumber\",...
Got:     errs.Newf(errs.InvalidArgument, "validate: [{\"field\":\"order_number\",...
```
**Fix**: Update test files to use snake_case field names in expected validation errors.

#### 2. PageConfigPackage JSON Unmarshalling (Export Tests)
```
Response: {"pageConfig":{"id":"...","name":"...","userId":"00000000-...","isDefault":true},...}
Test struct uses: `json:"user_id"` and `json:"is_default"`
Result: Empty struct because JSON keys don't match
```
**Fix**: The `PageConfigPackage` struct in `pageconfigapp/model.go` needs snake_case JSON tags for `userId` → `user_id` and `isDefault` → `is_default`.

### Recommended Fix Approach

1. **Search for test files with camelCase validation expectations**:
   ```bash
   grep -rn '"field":"[a-z]\+[A-Z]' api/cmd/services/ichor/tests/
   ```

2. **Update each test file** to use snake_case field names in expected error messages

3. **Verify PageConfigPackage struct** in `app/domain/config/pageconfigapp/model.go` - ensure all nested struct JSON tags are snake_case

4. **Re-run tests after fixes**:
   ```bash
   go clean -testcache && make test
   ```

### Note on Unrelated Failures

Some tests in `sales/ordersapi` and `formdata/formdataapi` show errors related to:
- `order_date` being a required field (may be a new validation added)
- `fulfillment_status_id` missing from form validation

These may be unrelated to the JSON tag changes and should be investigated separately.
