# Page Content System - Frontend Implementation Guide

## Overview

The Page Content System is a flexible, user-customizable page builder that allows you to create pages with any combination of content types:

- **Tables** - Dynamic data tables
- **Forms** - User input forms
- **Tabs** - Tabbed containers with nested content
- **Charts** - Data visualizations (future)
- **Text** - Rich text content (future)
- **Containers** - Grid layouts and sections

**Key Features:**
- ✅ Mix different content types on the same page
- ✅ Nested content (tabs containing tables, forms, etc.)
- ✅ User-customizable layouts
- ✅ Responsive design with Tailwind CSS
- ✅ Backward compatible with existing `page_tab_configs`

---

## Architecture

### Database Structure

```
core.pages (System Pages)
    ↓ (name reference)
config.page_configs (Layout Configurations)
    ↓ (page_config_id FK)
config.page_content (Content Blocks)
    ↓ (parent_id for nesting)
config.page_content (Child Content - tabs, etc.)
```

**Key Tables:**

1. **`core.pages`** - Application pages (routes, access control)
2. **`config.page_configs`** - Page layout configurations (default + user-specific)
3. **`config.page_content`** - Flexible content blocks
4. **`config.forms`** - Dynamic form configurations
5. **`config.form_fields`** - Form field definitions
6. **`config.page_actions`** - Page action buttons and dropdowns

### Database Migrations

All page content system tables are defined in `business/sdk/migrate/sql/migrate.sql`:

| Table | Migration Version | Description |
|-------|------------------|-------------|
| `config.page_configs` | **1.61** | Page layout configurations |
| `config.forms` | **1.63** | Dynamic forms |
| `config.form_fields` | **1.64**, **1.69** | Form field definitions (1.69 added entity_schema/entity_table) |
| `config.page_actions` | **1.65** | Base page actions table |
| `config.page_action_buttons` | **1.66** | Button-specific data |
| `config.page_action_dropdowns` | **1.67** | Dropdown container data |
| `config.page_action_dropdown_items` | **1.68** | Dropdown menu items |
| `config.page_content` | **1.70** | Flexible content blocks |

**Migration File Location**: `/business/sdk/migrate/sql/migrate.sql`

**Apply Migrations**:
```bash
make migrate
```

### Content Types

| Type | Description | Has Children? | Example Use |
|------|-------------|---------------|-------------|
| `table` | Data table | No | User list, order history |
| `form` | Input form | No | Create user, edit product |
| `tabs` | Tab container | Yes | Multiple views on same page |
| `container` | Grid layout | Yes | Dashboard widgets |
| `text` | Rich text | No | Instructions, help text |
| `chart` | Data visualization | No | Sales graphs, metrics |

---

## Page Actions

Pages can have **action buttons** displayed in the header or toolbar area. Actions are configured per page_config and can be:

### Action Types

| Type | Description | Use Case |
|------|-------------|----------|
| `button` | Single clickable button | Create new item, Export data, Refresh |
| `dropdown` | Button with dropdown menu | Multiple related actions, Bulk operations |
| `separator` | Visual divider | Organize action groups |

### Database Schema

**Page Actions** are stored in four related tables:

```sql
-- Version 1.65: Base actions table
CREATE TABLE config.page_actions (
    id UUID PRIMARY KEY,
    page_config_id UUID NOT NULL,
    action_type TEXT CHECK (action_type IN ('button', 'dropdown', 'separator')),
    action_order INT DEFAULT 1,
    is_active BOOLEAN DEFAULT TRUE,
    FOREIGN KEY (page_config_id) REFERENCES config.page_configs(id)
);

-- Version 1.66: Button-specific data
CREATE TABLE config.page_action_buttons (
    action_id UUID PRIMARY KEY,
    label TEXT NOT NULL,
    icon TEXT,
    target_path TEXT NOT NULL,
    variant TEXT DEFAULT 'default' CHECK (variant IN
        ('default', 'secondary', 'outline', 'ghost', 'destructive')),
    alignment TEXT DEFAULT 'right' CHECK (alignment IN ('left', 'right')),
    confirmation_prompt TEXT,
    FOREIGN KEY (action_id) REFERENCES config.page_actions(id)
);

-- Version 1.67: Dropdown container
CREATE TABLE config.page_action_dropdowns (
    action_id UUID PRIMARY KEY,
    label TEXT NOT NULL,
    icon TEXT,
    FOREIGN KEY (action_id) REFERENCES config.page_actions(id)
);

-- Version 1.68: Dropdown items
CREATE TABLE config.page_action_dropdown_items (
    id UUID PRIMARY KEY,
    dropdown_action_id UUID NOT NULL,
    label TEXT NOT NULL,
    target_path TEXT NOT NULL,
    item_order INT NOT NULL,
    FOREIGN KEY (dropdown_action_id) REFERENCES config.page_action_dropdowns(action_id)
);
```

### Button Variants

| Variant | Appearance | Use Case |
|---------|------------|----------|
| `default` | Primary blue | Main actions (Create, Submit) |
| `secondary` | Muted gray | Secondary actions (Cancel, Back) |
| `outline` | Bordered | Neutral actions (View, Details) |
| `ghost` | Minimal | Tertiary actions (More options) |
| `destructive` | Red/warning | Delete, Remove, Disable |

### Example: Creating Page Actions

```sql
-- 1. Create a "Create New" button
INSERT INTO config.page_actions (id, page_config_id, action_type, action_order)
VALUES ('action-1', 'page-config-123', 'button', 1);

INSERT INTO config.page_action_buttons (
    action_id, label, icon, target_path, variant, alignment
) VALUES (
    'action-1',
    'Create New User',
    'plus',
    '/users/create',
    'default',
    'right'
);

-- 2. Create a separator
INSERT INTO config.page_actions (id, page_config_id, action_type, action_order)
VALUES ('action-2', 'page-config-123', 'separator', 2);

-- 3. Create a dropdown with multiple options
INSERT INTO config.page_actions (id, page_config_id, action_type, action_order)
VALUES ('action-3', 'page-config-123', 'dropdown', 3);

INSERT INTO config.page_action_dropdowns (action_id, label, icon)
VALUES ('action-3', 'Bulk Actions', 'menu');

INSERT INTO config.page_action_dropdown_items (
    id, dropdown_action_id, label, target_path, item_order
) VALUES
    ('item-1', 'action-3', 'Export CSV', '/users/export', 1),
    ('item-2', 'action-3', 'Import Users', '/users/import', 2),
    ('item-3', 'action-3', 'Archive All', '/users/archive', 3);
```

### API Response Format

```json
{
  "pageActions": [
    {
      "id": "action-1",
      "pageConfigId": "page-config-123",
      "actionType": "button",
      "actionOrder": 1,
      "isActive": true,
      "button": {
        "label": "Create New User",
        "icon": "plus",
        "targetPath": "/users/create",
        "variant": "default",
        "alignment": "right",
        "confirmationPrompt": null
      }
    },
    {
      "id": "action-2",
      "pageConfigId": "page-config-123",
      "actionType": "separator",
      "actionOrder": 2,
      "isActive": true
    },
    {
      "id": "action-3",
      "pageConfigId": "page-config-123",
      "actionType": "dropdown",
      "actionOrder": 3,
      "isActive": true,
      "dropdown": {
        "label": "Bulk Actions",
        "icon": "menu",
        "items": [
          {"id": "item-1", "label": "Export CSV", "targetPath": "/users/export", "itemOrder": 1},
          {"id": "item-2", "label": "Import Users", "targetPath": "/users/import", "itemOrder": 2},
          {"id": "item-3", "label": "Archive All", "targetPath": "/users/archive", "itemOrder": 3}
        ]
      }
    }
  ]
}
```

### Fetching Page Actions

```typescript
// Get all actions for a page
const response = await fetch(`/v1/config/page-configs/actions/${pageConfigId}`)
const actions = await response.json()

// Filter active actions and sort by order
const activeActions = actions
  .filter(a => a.isActive)
  .sort((a, b) => a.actionOrder - b.actionOrder)
```

---

## API Response Format

### Endpoint

**Primary Endpoint** (returns content WITH nested children):
```
GET /v1/config/page-configs/content/children/{page_config_id}
```

**Note**: To get page content, you first need the `page_config_id`. You can get this by querying:
```
GET /v1/config/page-configs/name/{page_name}
```

### Response Structure

```json
{
  "pageConfig": {
    "id": "uuid",
    "name": "user_management_example",
    "userId": null,
    "isDefault": true
  },
  "contents": [
    {
      "id": "content-1",
      "contentType": "form",
      "label": "Create New User",
      "formId": "form-uuid",
      "orderIndex": 1,
      "layout": {
        "colSpan": { "default": 12 }
      },
      "parentId": null,
      "isVisible": true,
      "isDefault": false
    },
    {
      "id": "content-2",
      "contentType": "tabs",
      "label": "User Lists",
      "orderIndex": 2,
      "layout": {
        "colSpan": { "default": 12 },
        "containerType": "tabs"
      },
      "parentId": null,
      "isVisible": true,
      "isDefault": false,
      "children": [
        {
          "id": "tab-1",
          "contentType": "table",
          "label": "Active Users",
          "tableConfigId": "table-config-uuid",
          "orderIndex": 1,
          "parentId": "content-2",
          "isVisible": true,
          "isDefault": true,
          "layout": {}
        },
        {
          "id": "tab-2",
          "contentType": "table",
          "label": "Roles",
          "tableConfigId": "roles-table-uuid",
          "orderIndex": 2,
          "parentId": "content-2",
          "isVisible": true,
          "isDefault": false,
          "layout": {}
        }
      ]
    }
  ]
}
```

---

## Vue Component Implementation

### Page Layout Component

```vue
<template>
  <div class="page-content-container">
    <!-- Render top-level content blocks -->
    <div
      v-for="content in topLevelContents"
      :key="content.id"
      :class="getContentClasses(content)"
    >
      <component
        :is="getContentComponent(content.contentType)"
        :content="content"
        @refresh="handleRefresh"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import FormContent from './content-types/FormContent.vue'
import TableContent from './content-types/TableContent.vue'
import TabsContent from './content-types/TabsContent.vue'
import ChartContent from './content-types/ChartContent.vue'

interface PageContent {
  id: string
  contentType: 'table' | 'form' | 'tabs' | 'chart' | 'container'
  label?: string
  tableConfigId?: string
  formId?: string
  orderIndex: number
  parentId?: string
  layout: LayoutConfig
  isVisible: boolean
  isDefault: boolean
  children?: PageContent[]
}

interface LayoutConfig {
  colSpan?: { default: number; sm?: number; md?: number; lg?: number }
  rowSpan?: number
  gap?: string
  className?: string
  containerType?: string
}

const props = defineProps<{
  contents: PageContent[]
}>()

// Get top-level content (no parent)
const topLevelContents = computed(() =>
  props.contents.filter((c) => !c.parentId)
)

// Map content types to components
const contentComponents = {
  form: FormContent,
  table: TableContent,
  tabs: TabsContent,
  chart: ChartContent,
  container: 'div', // Generic container
}

const getContentComponent = (type: string) => {
  return contentComponents[type as keyof typeof contentComponents] || 'div'
}

// Generate Tailwind classes from layout config
const getContentClasses = (content: PageContent): string => {
  const layout = content.layout
  const classes: string[] = []

  // Column span (responsive)
  if (layout.colSpan) {
    classes.push(`col-span-${layout.colSpan.default}`)
    if (layout.colSpan.sm) classes.push(`sm:col-span-${layout.colSpan.sm}`)
    if (layout.colSpan.md) classes.push(`md:col-span-${layout.colSpan.md}`)
    if (layout.colSpan.lg) classes.push(`lg:col-span-${layout.colSpan.lg}`)
  }

  // Row span
  if (layout.rowSpan) {
    classes.push(`row-span-${layout.rowSpan}`)
  }

  // Custom classes
  if (layout.className) {
    classes.push(layout.className)
  }

  return classes.join(' ')
}

const handleRefresh = () => {
  // Reload page content
  console.log('Refreshing page content')
}
</script>
```

### Tabs Content Component

```vue
<template>
  <div>
    <h3 v-if="content.label" class="mb-4 text-lg font-semibold">
      {{ content.label }}
    </h3>

    <Tabs :default-value="defaultTabId">
      <TabsList>
        <TabsTrigger
          v-for="child in content.children"
          :key="child.id"
          :value="child.id"
        >
          {{ child.label }}
        </TabsTrigger>
      </TabsList>

      <TabsContent
        v-for="child in content.children"
        :key="child.id"
        :value="child.id"
      >
        <!-- Recursively render child content -->
        <TableContent v-if="child.contentType === 'table'" :content="child" />
        <FormContent v-if="child.contentType === 'form'" :content="child" />
        <ChartContent v-if="child.contentType === 'chart'" :content="child" />
      </TabsContent>
    </Tabs>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { Tabs, TabsList, TabsTrigger, TabsContent } from '@/components/ui/tabs'
import TableContent from './TableContent.vue'
import FormContent from './FormContent.vue'
import ChartContent from './ChartContent.vue'

interface PageContent {
  id: string
  contentType: string
  label?: string
  tableConfigId?: string
  formId?: string
  isDefault: boolean
  children?: PageContent[]
}

const props = defineProps<{
  content: PageContent
}>()

// Get the default active tab
const defaultTabId = computed(() => {
  const defaultChild = props.content.children?.find((c) => c.isDefault)
  return defaultChild?.id || props.content.children?.[0]?.id || ''
})
</script>
```

### Table Content Component

```vue
<template>
  <div>
    <h4 v-if="content.label" class="mb-2 text-sm font-medium text-muted-foreground">
      {{ content.label }}
    </h4>

    <DynamicTable
      v-if="tableConfig"
      :config="tableConfig"
      :config-id="content.tableConfigId"
    />

    <div v-else class="text-muted-foreground">
      Loading table configuration...
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import DynamicTable from '@/components/DynamicTable.vue'

interface PageContent {
  id: string
  label?: string
  tableConfigId?: string
}

const props = defineProps<{
  content: PageContent
}>()

const tableConfig = ref(null)

onMounted(async () => {
  if (props.content.tableConfigId) {
    // Fetch table configuration
    const response = await fetch(
      `/v1/config/tables/${props.content.tableConfigId}`
    )
    tableConfig.value = await response.json()
  }
})
</script>
```

### Form Content Component

```vue
<template>
  <div>
    <h4 v-if="content.label" class="mb-4 text-lg font-semibold">
      {{ content.label }}
    </h4>

    <DynamicForm
      v-if="form"
      :form="form"
      :form-id="content.formId"
      @submit="handleSubmit"
    />

    <div v-else class="text-muted-foreground">
      Loading form...
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import DynamicForm from '@/components/DynamicForm.vue'

interface PageContent {
  id: string
  label?: string
  formId?: string
}

const props = defineProps<{
  content: PageContent
}>()

const form = ref(null)

onMounted(async () => {
  if (props.content.formId) {
    // Fetch form with fields
    const response = await fetch(`/v1/config/forms/${props.content.formId}/full`)
    form.value = await response.json()
  }
})

const handleSubmit = (data: any) => {
  console.log('Form submitted:', data)
  // Handle form submission
}
</script>
```

---

## Forms and Form Fields

Forms are dynamic, configuration-driven forms that render based on database configuration. This allows non-developers to create and modify forms without code changes.

### Database Schema

#### Forms Table (Version 1.63)

```sql
CREATE TABLE config.forms (
    id UUID PRIMARY KEY,
    name VARCHAR(255) UNIQUE NOT NULL,
    is_reference_data BOOLEAN DEFAULT false,
    allow_inline_create BOOLEAN DEFAULT true
);
```

**Fields:**
- `name` - Unique identifier for the form (e.g., `create_user_form`)
- `is_reference_data` - If true, this form manages reference/lookup data (admin-only)
- `allow_inline_create` - If true, this form can be embedded in other forms for inline creation

#### Form Fields Table (Versions 1.64, 1.69)

```sql
CREATE TABLE config.form_fields (
    id UUID PRIMARY KEY,
    form_id UUID NOT NULL,
    entity_id UUID NOT NULL,
    entity_schema TEXT NOT NULL,  -- Added in 1.69
    entity_table TEXT NOT NULL,   -- Added in 1.69
    name VARCHAR(255) NOT NULL,
    label VARCHAR(255) NOT NULL,
    field_type VARCHAR(50) NOT NULL,
    field_order INTEGER NOT NULL,
    required BOOLEAN DEFAULT false,
    config JSONB NOT NULL DEFAULT '{}'::jsonb,

    FOREIGN KEY (form_id) REFERENCES config.forms(id),
    FOREIGN KEY (entity_id) REFERENCES workflow.entities(id),
    UNIQUE(form_id, entity_id, name)
);
```

**Key Fields:**
- `entity_id` - Links to `workflow.entities` table (defines which entity this field belongs to)
- `entity_schema` - Database schema (e.g., `core`, `hr`, `assets`)
- `entity_table` - Database table name (e.g., `users`, `offices`, `assets`)
- `name` - Column name in the database
- `field_type` - UI component type (`text`, `select`, `checkbox`, `date`, etc.)
- `field_order` - Display order in the form (1, 2, 3...)
- `config` - JSONB containing field-specific configuration

### Field Types

| Field Type | Description | Example Use |
|------------|-------------|-------------|
| `text` | Single-line text input | Name, email, username |
| `textarea` | Multi-line text input | Description, notes, comments |
| `select` | Dropdown selection | Role, status, category |
| `checkbox` | Boolean toggle | Active/inactive, enabled/disabled |
| `date` | Date picker | Birth date, hire date |
| `datetime` | Date and time picker | Created at, updated at |
| `number` | Numeric input | Age, quantity, price |
| `email` | Email input with validation | Email address |
| `url` | URL input with validation | Website, social media link |

### Config JSON Structure

The `config` field contains field-specific settings. For **foreign key fields** (select dropdowns referencing other tables), use this structure:

```json
{
    "parent_entity_id": "uuid-of-parent-entity",
    "foreign_key_column": "role_id",
    "execution_order": 1,
    "display_fields": ["name", "description"]
}
```

**Foreign Key Config Fields:**
- `parent_entity_id` - UUID of the entity being referenced (from `workflow.entities`)
- `foreign_key_column` - Name of the FK column in the current table (e.g., `role_id`, `office_id`)
- `execution_order` - Order of creation in multi-entity forms (1 = created first, 2 = created second, etc.)
- `display_fields` - Array of fields to show in the dropdown (e.g., `["name"]`, `["first_name", "last_name"]`)

### Example: Creating a Form with Fields

```sql
-- 1. Create the form
INSERT INTO config.forms (id, name, is_reference_data, allow_inline_create)
VALUES ('form-uuid-1', 'create_user_form', false, false);

-- 2. Add text field (username)
INSERT INTO config.form_fields (
    id, form_id, entity_id, entity_schema, entity_table,
    name, label, field_type, field_order, required, config
) VALUES (
    'field-uuid-1',
    'form-uuid-1',
    'entity-uuid-users',
    'core',
    'users',
    'username',
    'Username',
    'text',
    1,
    true,
    '{}'::jsonb
);

-- 3. Add foreign key field (role selection)
INSERT INTO config.form_fields (
    id, form_id, entity_id, entity_schema, entity_table,
    name, label, field_type, field_order, required, config
) VALUES (
    'field-uuid-2',
    'form-uuid-1',
    'entity-uuid-users',
    'core',
    'users',
    'role_id',
    'Role',
    'select',
    2,
    true,
    '{
        "parent_entity_id": "entity-uuid-roles",
        "foreign_key_column": "role_id",
        "execution_order": 1,
        "display_fields": ["name", "description"]
    }'::jsonb
);

-- 4. Add email field
INSERT INTO config.form_fields (
    id, form_id, entity_id, entity_schema, entity_table,
    name, label, field_type, field_order, required, config
) VALUES (
    'field-uuid-3',
    'form-uuid-1',
    'entity-uuid-users',
    'core',
    'users',
    'email',
    'Email Address',
    'email',
    3,
    true,
    '{"maxLength": 255}'::jsonb
);
```

### API Response: Form with Fields

Fetching a form with the `/full` endpoint returns all fields in a single response:

```json
{
    "id": "form-uuid-1",
    "name": "create_user_form",
    "isReferenceData": false,
    "allowInlineCreate": false,
    "fields": [
        {
            "id": "field-uuid-1",
            "formId": "form-uuid-1",
            "entityId": "entity-uuid-users",
            "entitySchema": "core",
            "entityTable": "users",
            "name": "username",
            "label": "Username",
            "fieldType": "text",
            "fieldOrder": 1,
            "required": true,
            "config": {}
        },
        {
            "id": "field-uuid-2",
            "formId": "form-uuid-1",
            "entityId": "entity-uuid-users",
            "entitySchema": "core",
            "entityTable": "users",
            "name": "role_id",
            "label": "Role",
            "fieldType": "select",
            "fieldOrder": 2,
            "required": true,
            "config": {
                "parent_entity_id": "entity-uuid-roles",
                "foreign_key_column": "role_id",
                "execution_order": 1,
                "display_fields": ["name", "description"]
            }
        },
        {
            "id": "field-uuid-3",
            "formId": "form-uuid-1",
            "entityId": "entity-uuid-users",
            "entitySchema": "core",
            "entityTable": "users",
            "name": "email",
            "label": "Email Address",
            "fieldType": "email",
            "fieldOrder": 3,
            "required": true,
            "config": {"maxLength": 255}
        }
    ]
}
```

### Relationship with workflow.entities

The `workflow.entities` table defines all entities in the system and serves as the central registry for forms:

```sql
SELECT id, name, schema_name, table_name
FROM workflow.entities
WHERE schema_name = 'core' AND table_name = 'users';
```

**Why this matters:**
- Forms can span multiple entities (multi-entity creation)
- `entity_id` links each field to its target entity
- `entity_schema` and `entity_table` provide direct database mapping
- Foreign key relationships use `parent_entity_id` to reference other entities

---

## FormData Registry (Multi-Entity Transactions)

Forms and form fields support **multi-entity transactions** through the FormData system. This allows creating complex, related records in a single atomic transaction.

### Supported Entities

| Entity | Create Support | Update Support | Template Variables |
|--------|----------------|----------------|-------------------|
| `forms` | ✅ Yes | ✅ Yes | ✅ `{{forms.id}}` |
| `form_fields` | ✅ Yes | ✅ Yes | ✅ `{{form_fields.id}}` |
| `page_configs` | ❌ No | ❌ No | - |
| `page_content` | ❌ No | ❌ No | - |
| `page_actions` | ❌ No | ❌ No | - |

### Template Variables

Template variables allow referencing values from previously created entities within the same transaction:

**Syntax**: `{{entity_name.field_name}}`

**Examples:**
- `{{forms.id}}` - ID of a form created earlier in the transaction
- `{{users.id}}` - ID of a user created earlier
- `{{roles.name}}` - Name field from a created role

### Example: Create Form + Fields in One Transaction

```json
{
  "entities": [
    {
      "entity_name": "forms",
      "operation": "CREATE",
      "data": {
        "name": "create_employee_form",
        "isReferenceData": false,
        "allowInlineCreate": false
      }
    },
    {
      "entity_name": "form_fields",
      "operation": "CREATE",
      "data": {
        "formId": "{{forms.id}}",
        "entityId": "entity-uuid-for-employees",
        "entitySchema": "hr",
        "entityTable": "employees",
        "name": "first_name",
        "label": "First Name",
        "fieldType": "text",
        "fieldOrder": 1,
        "required": true,
        "config": {}
      }
    },
    {
      "entity_name": "form_fields",
      "operation": "CREATE",
      "data": {
        "formId": "{{forms.id}}",
        "entityId": "entity-uuid-for-employees",
        "entitySchema": "hr",
        "entityTable": "employees",
        "name": "last_name",
        "label": "Last Name",
        "fieldType": "text",
        "fieldOrder": 2,
        "required": true,
        "config": {}
      }
    },
    {
      "entity_name": "form_fields",
      "operation": "CREATE",
      "data": {
        "formId": "{{forms.id}}",
        "entityId": "entity-uuid-for-employees",
        "entitySchema": "hr",
        "entityTable": "employees",
        "name": "office_id",
        "label": "Office",
        "fieldType": "select",
        "fieldOrder": 3,
        "required": true,
        "config": {
          "parent_entity_id": "entity-uuid-for-offices",
          "foreign_key_column": "office_id",
          "execution_order": 1,
          "display_fields": ["name", "city"]
        }
      }
    }
  ]
}
```

### Benefits of Multi-Entity Transactions

✅ **Atomicity** - All operations succeed or fail together
✅ **Consistency** - No partial creates if validation fails
✅ **Template Variables** - Reference IDs automatically
✅ **Automatic Validation** - Each entity is validated before creation
✅ **Single API Call** - Reduces network overhead

### API Endpoint

```
POST /v1/data/formdata
```

**Request Body**: JSON with `entities` array (shown above)

**Response**: Array of created entities with their IDs

---

## Real-World Example: User Management Page

### Visual Layout

```
┌─────────────────────────────────────────────────────┐
│  [Create New User Form]                             │
│  ┌──────────────────────────────────────────────┐  │
│  │ Name: [_______________]                       │  │
│  │ Email: [______________]                       │  │
│  │ Role: [dropdown ▼]                            │  │
│  │ [Submit] [Cancel]                             │  │
│  └──────────────────────────────────────────────┘  │
├─────────────────────────────────────────────────────┤
│  ┌────────────────────────────────────────────┐    │
│  │ [Active Users] [Roles] [Permissions]       │    │
│  ├────────────────────────────────────────────┤    │
│  │                                            │    │
│  │  Active Users Table                        │    │
│  │  ┌──────┬─────────┬──────────────┐        │    │
│  │  │ ID   │ Name    │ Email        │        │    │
│  │  ├──────┼─────────┼──────────────┤        │    │
│  │  │ 1    │ John    │ john@...     │        │    │
│  │  │ 2    │ Jane    │ jane@...     │        │    │
│  │  └──────┴─────────┴──────────────┘        │    │
│  │                                            │    │
│  └────────────────────────────────────────────┘    │
└─────────────────────────────────────────────────────┘
```

### Database Records

```sql
-- 1. Page Config (default for all users)
INSERT INTO config.page_configs (id, name, user_id, is_default)
VALUES ('page-config-123', 'user_management_example', NULL, true);

-- 2. Content Block: Form
INSERT INTO config.page_content (
    id, page_config_id, content_type, label, form_id,
    order_index, layout, is_visible
) VALUES (
    'content-form-1',
    'page-config-123',
    'form',
    'Create New User',
    'user-form-456',
    1,
    '{"colSpan":{"default":12}}',
    true
);

-- 3. Content Block: Tabs Container
INSERT INTO config.page_content (
    id, page_config_id, content_type, label,
    order_index, layout, is_visible
) VALUES (
    'content-tabs-2',
    'page-config-123',
    'tabs',
    'User Lists',
    2,
    '{"colSpan":{"default":12},"containerType":"tabs"}',
    true
);

-- 4. Tab 1: Active Users (child of tabs container)
INSERT INTO config.page_content (
    id, page_config_id, content_type, label, table_config_id,
    order_index, parent_id, is_default, is_visible
) VALUES (
    'tab-1',
    'page-config-123',
    'table',
    'Active Users',
    'active-users-table-789',
    1,
    'content-tabs-2',  -- Parent is the tabs container
    true,  -- Default active tab
    true
);

-- 5. Tab 2: Roles (child of tabs container)
INSERT INTO config.page_content (
    id, page_config_id, content_type, label, table_config_id,
    order_index, parent_id, is_default, is_visible
) VALUES (
    'tab-2',
    'page-config-123',
    'table',
    'Roles',
    'roles-table-101',
    2,
    'content-tabs-2',  -- Same parent
    false,
    true
);
```

### API Response

```json
{
  "pageConfig": {
    "id": "page-config-123",
    "name": "user_management_example",
    "userId": null,
    "isDefault": true
  },
  "contents": [
    {
      "id": "content-form-1",
      "contentType": "form",
      "label": "Create New User",
      "formId": "user-form-456",
      "orderIndex": 1,
      "layout": { "colSpan": { "default": 12 } },
      "isVisible": true,
      "isDefault": false
    },
    {
      "id": "content-tabs-2",
      "contentType": "tabs",
      "label": "User Lists",
      "orderIndex": 2,
      "layout": {
        "colSpan": { "default": 12 },
        "containerType": "tabs"
      },
      "isVisible": true,
      "isDefault": false,
      "children": [
        {
          "id": "tab-1",
          "contentType": "table",
          "label": "Active Users",
          "tableConfigId": "active-users-table-789",
          "orderIndex": 1,
          "parentId": "content-tabs-2",
          "isDefault": true,
          "isVisible": true,
          "layout": {}
        },
        {
          "id": "tab-2",
          "contentType": "table",
          "label": "Roles",
          "tableConfigId": "roles-table-101",
          "orderIndex": 2,
          "parentId": "content-tabs-2",
          "isDefault": false,
          "isVisible": true,
          "layout": {}
        }
      ]
    }
  ]
}
```

---

## Complete Workflow: Building a Page from Scratch

This section shows the complete process of creating a dynamic page with forms, tables, and actions.

### Step 1: Create Page Config

First, query to get the page name or create a page config:

```typescript
// Get existing page config by name
const configResponse = await fetch(`/v1/config/page-configs/name/users`)
const pageConfig = await configResponse.json()
const pageConfigId = pageConfig.id
```

Or create a new page config:

```sql
INSERT INTO config.page_configs (id, name, is_default)
VALUES ('new-page-config-id', 'employee_management', true);
```

### Step 2: Create Forms (If Needed)

If your page needs forms, create them using the FormData multi-entity endpoint:

```json
POST /v1/data/formdata
{
  "entities": [
    {
      "entity_name": "forms",
      "operation": "CREATE",
      "data": {
        "name": "create_employee_form",
        "isReferenceData": false,
        "allowInlineCreate": false
      }
    },
    {
      "entity_name": "form_fields",
      "operation": "CREATE",
      "data": {
        "formId": "{{forms.id}}",
        "entityId": "entity-uuid-employees",
        "entitySchema": "hr",
        "entityTable": "employees",
        "name": "first_name",
        "label": "First Name",
        "fieldType": "text",
        "fieldOrder": 1,
        "required": true,
        "config": {}
      }
    }
  ]
}
```

Save the returned `forms.id` for later.

### Step 3: Create Page Content

Now create content blocks that use the forms/tables:

```sql
-- 1. Add form at top of page
INSERT INTO config.page_content (
    id, page_config_id, content_type, label, form_id,
    order_index, layout, is_visible
) VALUES (
    gen_random_uuid(),
    'page-config-id',
    'form',
    'Create New Employee',
    'form-id-from-step-2',
    1,
    '{"colSpan":{"default":12}}',
    true
);

-- 2. Create tabs container
INSERT INTO config.page_content (
    id, page_config_id, content_type, label,
    order_index, layout, is_visible
) VALUES (
    'tabs-container-id',
    'page-config-id',
    'tabs',
    'Employee Data',
    2,
    '{"colSpan":{"default":12},"containerType":"tabs"}',
    true
);

-- 3. Add table as first tab (child of container)
INSERT INTO config.page_content (
    id, page_config_id, content_type, label, table_config_id,
    parent_id, order_index, is_default, is_visible
) VALUES (
    gen_random_uuid(),
    'page-config-id',
    'table',
    'Active Employees',
    'existing-table-config-id',
    'tabs-container-id',  -- Parent is tabs container
    1,
    true,  -- Default active tab
    true
);
```

### Step 4: Add Page Actions (Optional)

Add action buttons to the page:

```sql
-- Create "Add Employee" button
INSERT INTO config.page_actions (id, page_config_id, action_type, action_order)
VALUES ('action-id-1', 'page-config-id', 'button', 1);

INSERT INTO config.page_action_buttons (
    action_id, label, icon, target_path, variant, alignment
) VALUES (
    'action-id-1',
    'Add Employee',
    'plus',
    '/employees/create',
    'default',
    'right'
);
```

### Step 5: Fetch Complete Page Data (Frontend)

On the frontend, fetch all data for the page:

```typescript
async function loadPageData(pageName: string) {
  // 1. Get page config
  const configResponse = await fetch(`/v1/config/page-configs/name/${pageName}`)
  const pageConfig = await configResponse.json()

  // 2. Get content with nested children
  const contentResponse = await fetch(
    `/v1/config/page-configs/content/children/${pageConfig.id}`
  )
  const pageData = await contentResponse.json()

  // 3. Get page actions
  const actionsResponse = await fetch(
    `/v1/config/page-configs/actions/${pageConfig.id}`
  )
  const actions = await actionsResponse.json()

  return {
    config: pageConfig,
    contents: pageData.contents,
    actions: actions.filter(a => a.isActive)
  }
}
```

### Step 6: Render the Page

```vue
<template>
  <div>
    <!-- Page Actions -->
    <div class="mb-4 flex justify-end gap-2">
      <PageActions :actions="pageData.actions" />
    </div>

    <!-- Page Content -->
    <PageContentRenderer :contents="pageData.contents" />
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'

const pageData = ref(null)

onMounted(async () => {
  pageData.value = await loadPageData('employee_management')
})
</script>
```

### Complete Flow Diagram

```
1. Query page_config by name → get page_config_id
                ↓
2. Create forms + fields (optional) → get form_id
                ↓
3. Create page_content blocks → link to forms/tables
                ↓
4. Create page_actions (optional) → buttons, dropdowns
                ↓
5. Frontend: Fetch content/children → nested structure
                ↓
6. Frontend: Fetch actions → active buttons
                ↓
7. Render: Components dynamically render based on content_type
```

### Key Points

✅ **Order matters**: Create forms before linking them in page_content
✅ **Parent-child relationships**: Tabs must have a parent container
✅ **Use /children endpoint**: This populates nested content automatically
✅ **Filter active**: Only show `isActive` actions and `isVisible` content
✅ **Use UUIDs**: Foreign keys reference forms, tables, and entities by ID

---

## Layout System (Tailwind CSS)

### Responsive Column Spans

The `layout.colSpan` object controls how wide content is at different breakpoints:

```json
{
  "colSpan": {
    "default": 12,  // Full width on mobile
    "md": 6,        // Half width on tablet
    "lg": 4         // Third width on desktop
  }
}
```

**Generates Tailwind classes:**
```html
<div class="col-span-12 md:col-span-6 lg:col-span-4">
```

### Common Layout Patterns

#### Full Width
```json
{ "colSpan": { "default": 12 } }
```

#### Sidebar + Main (Responsive)
```json
// Sidebar (left)
{ "colSpan": { "default": 12, "md": 4 } }

// Main content (right)
{ "colSpan": { "default": 12, "md": 8 } }
```

#### Three Column Dashboard
```json
{ "colSpan": { "default": 12, "sm": 6, "lg": 4 } }
```

#### Custom Classes
```json
{
  "colSpan": { "default": 12 },
  "className": "bg-card rounded-lg shadow-sm p-6"
}
```

---

## Migration Guide

### From Old `page_tab_configs` to New `page_content`

**Old System:**
- Only supported tabs with tables
- Fixed structure: Page → Tabs → Tables
- No mixing of content types

**New System:**
- Supports any content type anywhere
- Flexible nesting
- Responsive layouts

**Migration Steps:**

1. **Query existing tab configs:**
```sql
SELECT * FROM config.page_tab_configs WHERE page_config_id = 'your-page-id';
```

2. **For each tab, create equivalent content blocks:**

```go
// Old: page_tab_configs entry
oldTab := PageTabConfig{
    PageConfigID: pageID,
    Label:        "Users",
    ConfigID:     tableConfigID,
    TabOrder:     1,
    IsDefault:    true,
}

// New: Create tabs container first
tabsContainer := PageContent{
    PageConfigID: pageID,
    ContentType:  "tabs",
    Label:        "User Management Tabs",
    OrderIndex:   1,
    Layout:       json.RawMessage(`{"colSpan":{"default":12}}`),
}

// New: Create tab as child of container
tab := PageContent{
    PageConfigID:  pageID,
    ContentType:   "table",
    Label:         "Users",
    TableConfigID: tableConfigID,
    OrderIndex:    1,
    ParentID:      tabsContainer.ID,  // Link to parent
    IsDefault:     true,
}
```

3. **Update frontend to use new API endpoint:**

```typescript
// Old
const response = await fetch(`/v1/config/pages/${pageName}/tabs`)

// New - First get page config ID
const configResponse = await fetch(`/v1/config/page-configs/name/${pageName}`)
const pageConfig = await configResponse.json()

// Then get content with children
const contentResponse = await fetch(`/v1/config/page-configs/content/children/${pageConfig.id}`)
const pageData = await contentResponse.json()
```

---

## Advanced Use Cases

### Multiple Tab Groups on One Page

```json
{
  "contents": [
    {
      "contentType": "tabs",
      "label": "User Management",
      "orderIndex": 1,
      "children": [
        { "contentType": "table", "label": "Active" },
        { "contentType": "table", "label": "Inactive" }
      ]
    },
    {
      "contentType": "tabs",
      "label": "Role Management",
      "orderIndex": 2,
      "children": [
        { "contentType": "table", "label": "Roles" },
        { "contentType": "form", "label": "New Role" }
      ]
    }
  ]
}
```

### Dashboard with Mixed Content

```json
{
  "contents": [
    {
      "contentType": "chart",
      "label": "Sales Overview",
      "orderIndex": 1,
      "layout": { "colSpan": { "default": 12, "lg": 8 } }
    },
    {
      "contentType": "table",
      "label": "Recent Orders",
      "orderIndex": 2,
      "layout": { "colSpan": { "default": 12, "lg": 4 } }
    },
    {
      "contentType": "tabs",
      "label": "Details",
      "orderIndex": 3,
      "layout": { "colSpan": { "default": 12 } },
      "children": [...]
    }
  ]
}
```

### Form with Side-by-Side Tables

```json
{
  "contents": [
    {
      "contentType": "form",
      "orderIndex": 1,
      "layout": { "colSpan": { "default": 12 } }
    },
    {
      "contentType": "table",
      "label": "Available Items",
      "orderIndex": 2,
      "layout": { "colSpan": { "default": 12, "md": 6 } }
    },
    {
      "contentType": "table",
      "label": "Selected Items",
      "orderIndex": 3,
      "layout": { "colSpan": { "default": 12, "md": 6 } }
    }
  ]
}
```

---

## Troubleshooting

### Content Not Appearing

1. **Check if content is visible:**
   ```sql
   SELECT * FROM config.page_content WHERE is_visible = false;
   ```

2. **Verify parent-child relationships:**
   ```sql
   SELECT id, content_type, label, parent_id
   FROM config.page_content
   WHERE page_config_id = 'your-page-id'
   ORDER BY order_index;
   ```

3. **Check layout JSON validity:**
   ```sql
   SELECT id, label, layout::text
   FROM config.page_content
   WHERE layout IS NULL OR layout = '{}'::jsonb;
   ```

### Tabs Not Rendering Correctly

1. **Ensure tabs have a container:**
   - Tab items must have `parent_id` pointing to a tabs container
   - Tabs container must have `content_type = 'tabs'`

2. **Check default tab:**
   ```sql
   SELECT * FROM config.page_content
   WHERE parent_id = 'tabs-container-id' AND is_default = true;
   ```

3. **Verify children are populated:**
   - Use `QueryPageContentWithChildren()` method in Go
   - Frontend should check `content.children` array

### Layout Issues

1. **Tailwind classes not applying:**
   - Ensure classes are in your Tailwind config safelist
   - Check that `colSpan` values are 1-12

2. **Responsive not working:**
   - Verify breakpoint names: `sm`, `md`, `lg`, `xl`, `2xl`
   - Check mobile-first cascade: default → sm → md → lg → xl

---

## Best Practices

### 1. Content Organization

✅ **DO:**
- Group related tabs under one container
- Use consistent `orderIndex` (1, 2, 3...)
- Set `isDefault` on one tab per container

❌ **DON'T:**
- Create orphan tabs without a container
- Skip `orderIndex` numbers (use sequential)
- Set multiple tabs as `isDefault`

### 2. Layout Design

✅ **DO:**
- Start with mobile-first layouts (`default: 12`)
- Use Tailwind's standard breakpoints
- Test on multiple screen sizes

❌ **DON'T:**
- Hardcode pixel widths
- Use non-standard breakpoint names
- Forget to test mobile layouts

### 3. Performance

✅ **DO:**
- Lazy-load tab content
- Cache table/form configurations
- Use `/content/children/{page_config_id}` endpoint for nested content
- Use `QueryPageContentWithChildren()` in Go backend for nested queries

❌ **DON'T:**
- Load all tabs upfront
- Fetch content configs repeatedly
- Query `/content/{page_config_id}` endpoint when you need children (use `/content/children/{page_config_id}` instead)

### 4. Understanding the `children` Array

**Important**: The `children` array in `PageContent` is NOT stored in the database. It is populated by the backend during queries.

**How it works:**
1. Database stores flat records with `parent_id` foreign keys
2. Backend queries all content for a page_config_id
3. Backend groups records by `parent_id` to build hierarchy
4. Response includes nested `children` arrays

**Two query options:**

| Endpoint | Returns | Use When |
|----------|---------|----------|
| `GET /v1/config/page-configs/content/{page_config_id}` | Flat array of all content | You want to build hierarchy yourself |
| `GET /v1/config/page-configs/content/children/{page_config_id}` | Nested structure with `children` populated | You want tabs/containers with children (recommended) |

**Backend methods:**
```go
// Returns flat array
contents, err := pageContentBus.QueryByPageConfigID(ctx, pageConfigID)

// Returns nested structure (children populated)
contents, err := pageContentBus.QueryWithChildren(ctx, pageConfigID)
```

---

## Authorization Matrix

All endpoints require **authentication** (valid JWT token). The table below shows which operations require **admin privileges**.

### Operations by Auth Level

| Domain | Operation | Any Authenticated User | Admin Only |
|--------|-----------|----------------------|------------|
| **Page Configs** | Read (GET) | ✅ | ✅ |
| **Page Configs** | Create (POST) | ❌ | ✅ |
| **Page Configs** | Update (PUT) | ❌ | ✅ |
| **Page Configs** | Delete (DELETE) | ❌ | ✅ |
| **Page Content** | Read (GET) | ✅ | ✅ |
| **Page Content** | Create (POST) | ❌ | ✅ |
| **Page Content** | Update (PUT) | ❌ | ✅ |
| **Page Content** | Delete (DELETE) | ❌ | ✅ |
| **Forms** | Read (GET) | ✅ | ✅ |
| **Forms** | Create (POST) | ✅ | ✅ |
| **Forms** | Update (PUT) | ✅ | ✅ |
| **Forms** | Delete (DELETE) | ✅ | ✅ |
| **Form Fields** | Read (GET) | ✅ | ✅ |
| **Form Fields** | Create (POST) | ✅ | ✅ |
| **Form Fields** | Update (PUT) | ✅ | ✅ |
| **Form Fields** | Delete (DELETE) | ✅ | ✅ |
| **Page Actions** | Read (GET) | ✅ | ✅ |
| **Page Actions** | Create (POST) | ❌ | ✅ |
| **Page Actions** | Update (PUT) | ❌ | ✅ |
| **Page Actions** | Delete (DELETE) | ❌ | ✅ |
| **Page Actions** | Batch Create | ❌ | ✅ |

### Authentication Middleware

All endpoints use standard authentication middleware:

```go
// Authentication - Validates JWT token
authen := mid.Authenticate(cfg.AuthClient)

// Authorization - Checks permissions
mid.Authorize(cfg.AuthClient, cfg.PermissionsBus, RouteTable, Action, Rule)
```

**Rules:**
- `auth.RuleAdminOnly` - Requires admin role
- `auth.RuleAny` - Any authenticated user (non-admin OK)

### Permission Actions

Permissions are checked against specific table names and actions:

| Table Name | Supported Actions |
|------------|-------------------|
| `page_configs` | READ, CREATE, UPDATE, DELETE |
| `page_content` | READ, CREATE, UPDATE, DELETE |
| `forms` | READ, CREATE, UPDATE, DELETE |
| `form_fields` | READ, CREATE, UPDATE, DELETE |
| `page_actions` | READ, CREATE, UPDATE, DELETE |
| `page_action_buttons` | READ, CREATE, UPDATE, DELETE |
| `page_action_dropdowns` | READ, CREATE, UPDATE, DELETE |

**Note**: Forms and form fields are more permissive (any authenticated user) to allow dynamic form creation workflows. Page structure configuration (configs, content, actions) requires admin privileges.

---

## API Endpoints Summary

### Page Config API

| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| `GET` | `/v1/config/page-configs/id/{config_id}` | Any User | Get page config by ID |
| `GET` | `/v1/config/page-configs/name/{name}` | Any User | Get page config by name |
| `POST` | `/v1/config/page-configs` | **Admin Only** | Create new page config |
| `PUT` | `/v1/config/page-configs/id/{config_id}` | **Admin Only** | Update page config |
| `DELETE` | `/v1/config/page-configs/id/{config_id}` | **Admin Only** | Delete page config |

### Page Content API

| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| `GET` | `/v1/config/page-content/{content_id}` | Any User | Get single content block |
| `GET` | `/v1/config/page-configs/content/{page_config_id}` | Any User | Get all content for page config |
| `GET` | `/v1/config/page-configs/content/children/{page_config_id}` | Any User | **Get content WITH nested children** ⭐ |
| `POST` | `/v1/config/page-content` | **Admin Only** | Create new content block |
| `PUT` | `/v1/config/page-content/{content_id}` | **Admin Only** | Update content block |
| `DELETE` | `/v1/config/page-content/{content_id}` | **Admin Only** | Delete content block |

### Form API

| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| `GET` | `/v1/config/forms` | Any User | Query forms with pagination |
| `GET` | `/v1/config/forms/{form_id}` | Any User | Get form by ID |
| `GET` | `/v1/config/forms/{form_id}/full` | Any User | **Get form WITH all fields** ⭐ |
| `GET` | `/v1/config/forms/name/{form_name}/full` | Any User | **Get form by name WITH fields** ⭐ |
| `POST` | `/v1/config/forms` | Any User | Create new form |
| `PUT` | `/v1/config/forms/{form_id}` | Any User | Update form |
| `DELETE` | `/v1/config/forms/{form_id}` | Any User | Delete form |

### Form Field API

| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| `GET` | `/v1/config/form-fields` | Any User | Query form fields |
| `GET` | `/v1/config/form-fields/{field_id}` | Any User | Get field by ID |
| `POST` | `/v1/config/form-fields` | Any User | Create new field |
| `PUT` | `/v1/config/form-fields/{field_id}` | Any User | Update field |
| `DELETE` | `/v1/config/form-fields/{field_id}` | Any User | Delete field |

### Page Action API

| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| `GET` | `/v1/config/page-actions` | Any User | Query all actions |
| `GET` | `/v1/config/page-actions/{action_id}` | Any User | Get action by ID |
| `GET` | `/v1/config/page-configs/actions/{page_config_id}` | Any User | **Get all actions for page** ⭐ |
| `POST` | `/v1/config/page-actions/buttons` | **Admin Only** | Create button action |
| `PUT` | `/v1/config/page-actions/buttons/{action_id}` | **Admin Only** | Update button |
| `POST` | `/v1/config/page-actions/dropdowns` | **Admin Only** | Create dropdown action |
| `PUT` | `/v1/config/page-actions/dropdowns/{action_id}` | **Admin Only** | Update dropdown |
| `POST` | `/v1/config/page-actions/separators` | **Admin Only** | Create separator |
| `PUT` | `/v1/config/page-actions/separators/{action_id}` | **Admin Only** | Update separator |
| `DELETE` | `/v1/config/page-actions/{action_id}` | **Admin Only** | Delete any action type |
| `POST` | `/v1/config/page-configs/actions/batch/{page_config_id}` | **Admin Only** | **Batch create actions** ⭐ |

**Legend**: ⭐ = Special/recommended endpoint

---

## Advanced Topics

### Batch Page Action Creation

For creating multiple page actions at once, use the batch endpoint:

```typescript
POST /v1/config/page-configs/actions/batch/{page_config_id}

{
  "buttons": [
    {
      "label": "Create New",
      "icon": "plus",
      "targetPath": "/users/create",
      "variant": "default",
      "alignment": "right",
      "actionOrder": 1
    },
    {
      "label": "Export CSV",
      "icon": "download",
      "targetPath": "/users/export",
      "variant": "outline",
      "alignment": "left",
      "actionOrder": 3
    }
  ],
  "dropdowns": [
    {
      "label": "More Actions",
      "icon": "menu",
      "actionOrder": 4,
      "items": [
        {"label": "Import", "targetPath": "/users/import", "itemOrder": 1},
        {"label": "Archive", "targetPath": "/users/archive", "itemOrder": 2}
      ]
    }
  ],
  "separators": [
    {"actionOrder": 2}
  ]
}
```

**Benefits**:
- Single transaction for all actions
- Automatic rollback if any action fails
- Simpler frontend code

### When to Use `/children` vs Regular Endpoint

| Scenario | Recommended Endpoint | Reason |
|----------|---------------------|--------|
| Rendering tabs with nested content | `/content/children/{page_config_id}` | Children already populated |
| Building custom hierarchy | `/content/{page_config_id}` | You control grouping logic |
| Initial page load | `/content/children/{page_config_id}` | Complete structure in one call |
| Admin config UI (flat list) | `/content/{page_config_id}` | Easier to render flat lists |

### Optimizing Page Load Performance

**Option 1: Parallel Requests**
```typescript
async function loadPageOptimized(pageName: string) {
  // Fire all requests in parallel
  const [configRes, contentRes, actionsRes] = await Promise.all([
    fetch(`/v1/config/page-configs/name/${pageName}`),
    fetch(`/v1/config/page-configs/content/children/${pageConfigId}`),
    fetch(`/v1/config/page-configs/actions/${pageConfigId}`)
  ])

  return {
    config: await configRes.json(),
    contents: await contentRes.json(),
    actions: await actionsRes.json()
  }
}
```

**Option 2: Sequential with Early Exit**
```typescript
async function loadPageSafe(pageName: string) {
  // Get config first
  const configRes = await fetch(`/v1/config/page-configs/name/${pageName}`)
  if (!configRes.ok) return null  // Early exit

  const config = await configRes.json()

  // Then parallel for the rest
  const [contentRes, actionsRes] = await Promise.all([
    fetch(`/v1/config/page-configs/content/children/${config.id}`),
    fetch(`/v1/config/page-configs/actions/${config.id}`)
  ])

  return {
    config,
    contents: await contentRes.json(),
    actions: await actionsRes.json()
  }
}
```

### Dynamic Content Type Registration

If you need to add custom content types beyond the built-in ones (`table`, `form`, `tabs`, etc.), you'll need to:

1. **Update database constraint** in `config.page_content`:
   ```sql
   ALTER TABLE config.page_content DROP CONSTRAINT check_content_type;
   ALTER TABLE config.page_content ADD CONSTRAINT check_content_type
       CHECK (content_type IN ('table', 'form', 'tabs', 'container', 'text', 'chart', 'custom_type'));
   ```

2. **Register component in frontend**:
   ```typescript
   const contentComponents = {
     form: FormContent,
     table: TableContent,
     tabs: TabsContent,
     chart: ChartContent,
     custom_type: CustomTypeContent  // Your custom component
   }
   ```

3. **Handle in backend validation** (`business/domain/config/pagecontentbus/pagecontentbus.go`):
   ```go
   func validateContentType(contentType string) error {
       validTypes := []string{"table", "form", "tabs", "container", "text", "chart", "custom_type"}
       // ...
   }
   ```

### Caching Strategies

For production deployments, consider caching:

**Page Configs** (rarely change):
```typescript
const configCache = new Map<string, PageConfig>()

async function getPageConfig(name: string) {
  if (configCache.has(name)) {
    return configCache.get(name)
  }
  const config = await fetchPageConfig(name)
  configCache.set(name, config)
  return config
}
```

**Forms** (change infrequently):
```typescript
const formCache = new Map<string, FormFull>()

async function getForm(formId: string) {
  if (formCache.has(formId)) {
    return formCache.get(formId)
  }
  const form = await fetch(`/v1/config/forms/${formId}/full`).then(r => r.json())
  formCache.set(formId, form)
  return form
}
```

**Content and Actions** (DO NOT cache):
- These can change frequently as users customize their pages
- Always fetch fresh on page load

---

## Additional Resources

- **Tailwind CSS Grid Documentation:** https://tailwindcss.com/docs/grid-template-columns
- **shadcn-vue Tabs Component:** https://www.shadcn-vue.com/docs/components/tabs.html
- **Backend Implementation:** `business/sdk/tablebuilder/`
- **Seed Data Example:** `business/sdk/dbtest/seedFrontend.go` (line ~2941)

---

## Questions?

For questions or issues with the Page Content System:

1. Check the seed data example in `seedFrontend.go`
2. Review the Go models in `tablebuilder/model.go`
3. Test with the "User Management Example" page created in seed data
4. Reach out to the backend team with specific use cases

**Happy building! 🚀**
