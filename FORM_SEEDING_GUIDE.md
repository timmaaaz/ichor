# Backend Prompt: Seed Database with Form Configurations

## Objective

Create a seeding script that generates `config.forms` and `config.form_fields` entries for all user-facing tables in the database. These forms will enable:

1. CRUD operations via dynamic forms
2. Inline entity creation within other forms (where appropriate)
3. Automatic field rendering based on database schema

## Prerequisites

**Required Backend Changes:**

First, add these columns to `config.forms`:

```sql
-- Version: 2.11
-- Description: Add inline creation metadata to forms
ALTER TABLE config.forms
    ADD COLUMN is_reference_data BOOLEAN DEFAULT false,
    ADD COLUMN allow_inline_create BOOLEAN DEFAULT true;

CREATE INDEX idx_forms_reference_data ON config.forms(is_reference_data);
CREATE INDEX idx_forms_inline_create ON config.forms(allow_inline_create);

COMMENT ON COLUMN config.forms.is_reference_data IS
    'If true, this form represents stable reference data managed by admins only (no inline creation allowed)';
COMMENT ON COLUMN config.forms.allow_inline_create IS
    'If true, this form can be embedded for inline entity creation within other forms';
```

;

These columns need to be added at each level of this domain including seeding and tests.

## Form Generation Logic

### Step 1: Identify Tables to Generate Forms For

Query `information_schema.tables` for all tables EXCEPT:

**Exclude System/Internal Tables:**

- All tables in `config` schema (these define the form system itself)
- `workflow.automation_executions` (system logs)
- `workflow.notification_deliveries` (system logs)
- `workflow.allocation_results` (internal data)
- `core.table_access` (admin-only, manage via separate UI)
- `core.role_pages` (admin-only, manage via separate UI)
- `core.pages` (admin-only, manage via separate UI)
- `workflow.trigger_types`
- `workflow.entity_types`
- `workflow.action_templates`

**Exclude Junction Tables** (for now, handle separately):

- `assets.asset_tags` (many-to-many)
- `core.user_roles` (many-to-many)
- `hr.reports_to` (many-to-many, self-referential)
- `workflow.rule_dependencies` (system config)

**Result:** ~50-60 tables will get forms

---

### Step 2: Categorize Tables by Type

Use this classification to set `is_reference_data` and `allow_inline_create`:

#### **Reference/Lookup Data** (Stable, Admin-Managed)

Set: `is_reference_data = true`, `allow_inline_create = false`

**Geography (External Data Sources):**

- `geography.countries` - ISO country list
- `geography.regions` - State/province data

**Workflow Statuses:**

- `hr.user_approval_status`
- `assets.approval_status`
- `assets.fulfillment_status`
- `sales.order_fulfillment_statuses`
- `sales.line_item_fulfillment_statuses`
- `procurement.purchase_order_statuses`
- `procurement.purchase_order_line_item_statuses`

**Controlled Vocabularies:**

- `assets.asset_types`
- `assets.asset_conditions`
- `products.product_categories` - Business taxonomy

#### **User-Created Transactional Data** (Allow Inline Creation)

Set: `is_reference_data = false`, `allow_inline_create = true`

**Core/HR:**

- `core.contact_infos` - Created with customers, suppliers, brands
- `hr.titles` - Can be added on-the-fly
- `hr.offices` - Depends on street, can be created inline
- `hr.homes` - User-specific addresses
- `core.users` - Employee onboarding (but mark special, see notes below)

**Geography (User-Created):**

- `geography.cities` - Can create new cities as needed
- `geography.streets` - Frequently created with addresses

**Sales:**

- `sales.customers` - Primary entity
- `sales.orders` - Transactional
- `sales.order_line_items` - Nested within orders

**Products:**

- `products.brands` - Can create new brands
- `products.products` - Product catalog
- `products.physical_attributes` - Associated with products
- `products.product_costs` - Pricing history
- `products.quality_metrics` - Product QA data
- `products.cost_history` - Financial tracking

**Procurement:**

- `procurement.suppliers` - Supplier management
- `procurement.supplier_products` - Supplier catalog
- `procurement.purchase_orders` - Transactional
- `procurement.purchase_order_line_items` - Nested within POs

**Inventory:**

- `inventory.warehouses` - Warehouse setup
- `inventory.zones` - Warehouse organization
- `inventory.inventory_locations` - Bin locations
- `inventory.inventory_items` - Stock levels
- `inventory.inventory_transactions` - Stock movements
- `inventory.inventory_adjustments` - Corrections
- `inventory.transfer_orders` - Transfers
- `inventory.serial_numbers` - Serialized tracking
- `inventory.lot_trackings` - Lot/batch tracking
- `inventory.quality_inspections` - QA records

**Assets:**

- `assets.valid_assets` - Asset catalog
- `assets.assets` - Asset instances
- `assets.user_assets` - Asset assignments
- `assets.tags` - Tagging system
- `hr.user_approval_comments` - Comments/notes

**Workflow:**

- `workflow.automation_rules` - User-defined rules
- `workflow.rule_actions` - Rule configuration
- `workflow.entities` - Entity registry

---

### Step 3: Generate Form Records

For each table, create ONE formbus.NewForm entry in business/sdk/dbtest/seedmodels/tableforms.go
as well as one array of formfieldbus.NewFormField that contains one entry for
each field in the corresponding table. Fields should be as follows

formbus.NewForm

- Name: same name as the entity/table to which it corresponds

FormID

- FormID: uuid returned from the previous form
- EntityID: entity id for the workflow.entities item for that table
- EntitySchema: name of table schema
- EntityTable: name of the table
- Name: name of the field
- Label: name of the user interface label
- FieldType: type of field from the following
  - text - Standard text input
  - email - Email input
  - tel / phone - Telephone input
  - password - Password input
  - url - URL input
  - number - Number input
  - textarea - Multi-line text area
  - date - Date picker
  - time - Time input (currently using text input with HH:MM placeholder)
  - boolean - Checkbox
  - select - Dropdown (either dynamic from database table or static with options array)
  - multiselect - Multi-select (not yet fully supported, falls back to text input)
- FieldOrder: treat it like an iota starting at 1 in a given form
- Required: match up with Nullable / Not Nullable in the database
- Config: just set to json.RawMessage(`{}`) for the time being

These should have functions like you see in business/sdk/dbtest/seedmodels/forms.go

I know there may be duplicates in forms.go, my goal is to have everything that
is specifically a table form in the tableforms.go file.

Below is a fieldmap that is an example of how the data should be structured for
selecting a FieldType, above though are the ACTUAL field types from the
frontend. The exceptions for \_id and others are valid.

---

### Step 4: Generate Form Fields

For each table, query `information_schema.columns` and create form fields:

Follow the exclusion rules below for fields that should ALWAYS not be on the
form, these fields are generated by the backend.

#### **Field Exclusion Rules**

**ALWAYS Exclude:**

- `id` - Auto-generated UUID primary key
- `created_date`, `updated_date` - Auto-populated by system
- `created_by`, `updated_by` - Auto-populated from auth context
- `password_hash` - Security sensitive (handle separately via password change UI)
- `roles`, `system_roles` - Complex array fields (handle via separate role assignment UI)

**Conditionally Exclude:**

- `deactivated_by` - Only show in "deactivate" action, not create/edit forms
- `enabled` (on users table) - Admin-only field, show in separate UI
- Self-referential FKs like `requested_by`, `approved_by` - Often auto-populated by workflow

#### **Field Type Mapping**

Map PostgreSQL types to form field types:

```go
typeMap := map[string]string{
    "character varying": "text",
    "varchar":           "text",
    "text":             "text",
    "char":             "text",

    "integer":          "number",
    "int":              "number",
    "smallint":         "number",
    "bigint":           "number",

    "numeric":          "number",  // Or "money" if column name contains "cost", "price", "amount"
    "decimal":          "number",

    "boolean":          "checkbox",
    "bool":             "checkbox",

    "date":             "datepicker",
    "timestamp":        "datetime",
    "timestamp without time zone": "datetime",

    "interval":         "text",  // Or create custom duration field

    "USER-DEFINED":     "text",  // ENUMs - detect and convert to dropdown
    "ARRAY":            "text",  // Handle specially if needed
}
```

**Special Cases:**

- Columns ending in `_id` (UUIDs) → Likely foreign keys, set to `"dropdown-from-table"` or `"combobox-from-table"`
- Columns named `email`, `email_address` → `"email"`
- Columns named `phone`, `*_phone_*` → `"tel"`
- Columns named `url`, `website` → `"url"`
- Columns named `notes`, `description`, `comment*` → `"textarea"`
- Columns with `numeric` type containing "price", "cost", "amount" → `"money"`

#### **Detect Foreign Keys**

Query `information_schema.table_constraints` and `information_schema.key_column_usage` to find foreign key relationships:

```sql
SELECT
    kcu.column_name,
    ccu.table_schema AS foreign_table_schema,
    ccu.table_name AS foreign_table_name,
    ccu.column_name AS foreign_column_name
FROM information_schema.table_constraints AS tc
JOIN information_schema.key_column_usage AS kcu
    ON tc.constraint_name = kcu.constraint_name
JOIN information_schema.constraint_column_usage AS ccu
    ON ccu.constraint_name = tc.constraint_name
WHERE tc.constraint_type = 'FOREIGN KEY'
    AND tc.table_schema = '{schema}'
    AND tc.table_name = '{table}';
```

For FK fields:

- Set `field_type = 'combobox-from-table'`
- Add to `config` JSON:
- note that all foreign keys should be <table_name>\_id, but there may be
  exceptions such as in cases where a table may be referenced by foreign key
  more than once

```json
{
  "table_option": {
    "schema": "{foreign_table_schema}",
    "table": "{foreign_table_name}",
    "valueColumn": "id",
    "labelColumn": "{best_guess_label_column}",
    "searchColumns": ["{best_guess_label_column}"],
    "pageSize": 50
  },
  "inline_form_name": "{foreign_table_name}",
  "allow_inline_create": {determined_from_forms_table}
}
```

**Label Column Detection** (for dropdowns):
Priority order to guess display label for FK dropdown:

1. Column named `name` → Use this
2. Column named `number` → Use this (for order numbers, etc.)
3. Column named `title` → Use this
4. First VARCHAR/TEXT column that isn't `id` or `description`
5. Fallback: Concatenate multiple columns (e.g., `first_name || ' ' || last_name`)

#### **Detect Required Fields**

Query `information_schema.columns.is_nullable`:

- `is_nullable = 'NO'` AND column is not auto-generated → `required = true`
- `is_nullable = 'YES'` → `required = false`

#### **Field Order**

Order fields logically:

1. Natural/business key fields first (name, number, title)
2. Primary business fields next
3. Foreign keys in middle
4. Optional fields (notes, description) last

You can use `field_order = ROW_NUMBER()` based on ordinal position from `information_schema.columns`.

#### **Generate form_fields Records**

FormFields should be in the model functions defined above

---

### Step 5: Handle Special Cases

#### **ENUM Types**

NOTE: These are only examples, I do not want direct database operations seeding
this stuff if at all avoidable.

Detect PostgreSQL ENUMs (like `contact_type`):

```sql
SELECT
    t.typname,
    e.enumlabel
FROM pg_type t
JOIN pg_enum e ON t.oid = e.enumtypid
WHERE t.typname = 'contact_type';
```

For ENUM columns:

- Set `field_type = 'dropdown'` (not combobox-from-table)
- Add to config JSON:

```json
{
  "options": [
    { "value": "phone", "label": "Phone" },
    { "value": "email", "label": "Email" },
    { "value": "mail", "label": "Mail" },
    { "value": "fax", "label": "Fax" }
  ]
}
```

#### **Self-Referential Foreign Keys**

Tables like `core.users` with `requested_by` and `approved_by` pointing to `core.users(id)`:

- Allow `inline_form_name = null` (don't allow recursive inline creation of users within user form)
- Set `allow_inline_create = false` in field config

#### **Multi-Column Labels**

I DON'T KNOW THAT WE HAVE THE INFRASTRUCTURE CREATED TO SUPPORT THE FOLLOWING YET

For tables like `core.contact_infos` or `core.users`, dropdown labels should combine multiple columns:

- `contact_infos` → `first_name || ' ' || last_name || ' (' || email_address || ')'`
- `users` → `first_name || ' ' || last_name || ' (' || username || ')'`

Store in `table_option.labelExpression` (backend will need to support this):

```json
{
  "table_option": {
    "schema": "core",
    "table": "users",
    "valueColumn": "id",
    "labelExpression": "first_name || ' ' || last_name",
    "searchColumns": ["first_name", "last_name", "email", "username"],
    "pageSize": 50
  }
}
```

#### **Array Fields**

Tables like `core.users` have `roles TEXT[]` and `system_roles TEXT[]`:

- Skip these in form generation (handle via separate role assignment UI)
- Or set `field_type = 'multi-select'` if you want to support them later

#### **Audit Fields in Existing Records**

For tables that already have data, ensure the seeding script:

- Uses a system user UUID for `created_by` / `updated_by` on config records
- Sets appropriate timestamps

---

## Expected Output

The script should generate:

- **~55 form records** in `config.forms`
- **~400-500 field records** in `config.form_fields` (average 7-9 fields per form)

---

## Implementation Approach

All of this should be implemented in the dbtest/_ packages where the actual data
is defined in seedmodels/_ and the structures are initialized in seedFrontend.go

## Notes for Complex Tables

### `core.users`

- Exclude: `password_hash`, `roles`, `system_roles`, `enabled`
- Self-referential FKs (`requested_by`, `approved_by`) should NOT allow inline creation
- Consider creating TWO forms:
  - `users` (full employee profile)
  - `users_quick` (minimal fields for quick add)

### `sales.customers`

- Has 3 FK dependencies: `contact_id`, `delivery_address_id`, both should allow inline creation
- Your example form config shows this is already correctly configured

### `inventory.*` tables

- Complex interdependencies (warehouse → zone → location)
- All should allow inline creation for warehouse setup workflow

### `workflow.automation_rules`

- Has JSONB fields (`trigger_conditions`, `action_config`)
- Field type should be `textarea` or create custom `json-editor` type

---

## Post-Seeding Tasks

After seeding:

1. **Manually review** forms for tables with complex business logic:

   - `core.users`
   - `sales.orders` + `sales.order_line_items` (parent-child)
   - `procurement.purchase_orders` + line items

2. **Create custom forms** for special workflows:

   - User registration (subset of users table)
   - Quick customer add (subset of customers table)

3. **Update existing form configs** if they conflict with seeded ones

4. **Add to workflow.entities** any tables missing from that registry

---

## Questions?

If you encounter edge cases or need clarification on specific tables, refer back to this categorization. The key principles:

1. **Reference data** = Admin-managed, stable, no inline creation
2. **Transactional data** = User-created, allow inline forms
3. **System data** = No user forms at all
4. **Audit fields** = Always exclude from forms
