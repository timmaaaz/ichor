# Missing Tables and Future Work

This document tracks database tables, features, and pages that are referenced but not yet implemented.

## Missing Database Tables

### Procurement Module
- **`procurement.purchase_orders`** - Referenced by `/procurement/orders/index.vue`
  - Would need: id, order_number, supplier_id, status, order_date, expected_delivery, etc.
  - Related: Should integrate with `procurement.suppliers` and `products.products`

### General Requests
- **Requests table** - Referenced by `/requests/index.vue`
  - Could be a unified view across multiple request types (asset requests, purchase requests, etc.)
  - May need a polymorphic design or union view

## Future Features (Not Implemented Yet)

### Dashboard Widgets
- Currently using multi-tabbed tables for dashboards
- Future: Widget system with draggable/resizable panels
- Future: Chart widgets, KPI cards, activity feeds

### Reporting System
- Pages exist but configs skipped for now:
  - `/reports/index.vue`
  - `/hr/reports/index.vue`
  - `/inventory/reports/index.vue`
  - `/sales/reports/index.vue`
- Future: Aggregated views, custom report builder, scheduled reports

### Profile & Settings Pages
- `/profile/index.vue` - User profile management
- `/settings/index.vue` - User preferences/settings
- These may not need table configs (form-based rather than table-based)

## Pages Needing Special Handling

### Procurement Approvals (`/procurement/approvals/index.vue`)
- Depends on workflow integration
- May need a custom view joining:
  - `workflow.automation_executions`
  - Multiple entity types (purchase orders, inventory adjustments, etc.)
  - Approval status tracking

### General Orders (`/orders/index.vue`)
- Unclear if this is different from `/sales/orders/index.vue`
- May be a unified view of multiple order types
- Could be removed/consolidated

## Notes
- Some pages may be placeholders for future functionality
- Dashboard pages will evolve as widget system is implemented
- Report pages need dedicated reporting architecture
