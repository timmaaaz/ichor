# Page Actions API - Bundled Polymorphic System

## Overview

The Page Actions system is a **bundled polymorphic domain** that manages three types of UI actions that can appear on page configurations: buttons, dropdowns (with nested items), and separators. This synopsis explains the unique architectural decisions and patterns used in this implementation.

## Why a "Bundled" Approach?

Unlike typical domains where each entity has its own independent bus package, Page Actions uses a **single `pageactionbus` package** that handles all three action types plus dropdown items. This design choice reflects the **tightly coupled nature** of these entities:

1. **Shared base attributes**: All actions have common fields (page_config_id, action_order, is_active)
2. **Atomic operations**: Dropdowns must always be created/updated with their items in a single transaction
3. **Unified querying**: The UI needs to fetch all actions for a page together, grouped by type

## Database Schema

```sql
-- Base table for all action types
config.page_actions (
    id UUID PRIMARY KEY,
    page_config_id UUID NOT NULL,  -- FK to config.page_configs
    action_type TEXT NOT NULL,      -- 'button', 'dropdown', or 'separator'
    action_order INTEGER NOT NULL,  -- Unique per page_config_id
    is_active BOOLEAN DEFAULT true
)

-- Type-specific tables
config.page_action_buttons (
    action_id UUID PRIMARY KEY,     -- FK to page_actions.id
    label TEXT NOT NULL,
    icon TEXT,
    target_path TEXT NOT NULL,
    variant TEXT NOT NULL,          -- 'default', 'secondary', 'outline', 'ghost', 'destructive'
    alignment TEXT NOT NULL,        -- 'left', 'right'
    confirmation_prompt TEXT
)

config.page_action_dropdowns (
    action_id UUID PRIMARY KEY,     -- FK to page_actions.id
    label TEXT NOT NULL,
    icon TEXT
)

config.page_action_dropdown_items (
    id UUID PRIMARY KEY,
    dropdown_action_id UUID NOT NULL,  -- FK to page_actions.id
    label TEXT NOT NULL,
    target_path TEXT NOT NULL,
    item_order INTEGER NOT NULL    -- Order within the dropdown
)
```

## Action Types

### 1. Button Actions
Simple clickable buttons with optional confirmation prompts.
- **Variants**: default, secondary, outline, ghost, destructive
- **Alignment**: left, right
- **Target**: Navigation path

### 2. Dropdown Actions
Menus with 1+ child items, created/updated atomically.
- **Always bundled**: Items cannot exist without their parent dropdown
- **Minimum**: 1 item required
- **Cascading deletes**: Deleting dropdown removes all items

### 3. Separator Actions
Visual dividers with no additional data beyond base fields.

## API Endpoints

### Type-Specific CRUD
```
POST   /v1/config/page-actions/buttons       - Create button
PUT    /v1/config/page-actions/buttons/{id}  - Update button
POST   /v1/config/page-actions/dropdowns     - Create dropdown (with items)
PUT    /v1/config/page-actions/dropdowns/{id} - Update dropdown (replaces all items)
POST   /v1/config/page-actions/separators    - Create separator
PUT    /v1/config/page-actions/separators/{id} - Update separator
DELETE /v1/config/page-actions/{id}          - Delete any action type
```

### Query Endpoints
```
GET /v1/config/page-actions                    - Paginated query across all types
GET /v1/config/page-actions/{id}               - Get single action with full details
GET /v1/config/page-configs/{id}/actions       - Get all actions for a page, grouped by type
```

### Batch Operations
```
POST /v1/config/page-configs/{id}/actions/batch - Create multiple actions in one transaction
```

## Key Architectural Patterns

### 1. Polymorphic Response Model

```go
type PageAction struct {
    ID           string          `json:"id"`
    PageConfigID string          `json:"pageConfigId"`
    ActionType   string          `json:"actionType"`  // "button", "dropdown", "separator"
    ActionOrder  int             `json:"actionOrder"`
    IsActive     bool            `json:"isActive"`
    Button       *ButtonAction   `json:"button,omitempty"`    // Populated only for buttons
    Dropdown     *DropdownAction `json:"dropdown,omitempty"`  // Populated only for dropdowns
}
```

**Design Decision**: Use a single response type with optional nested objects rather than separate types. This simplifies querying and allows the frontend to handle all actions uniformly while still accessing type-specific data.

### 2. Type-Specific Create/Update Models

Rather than a generic polymorphic create model, we use **type-specific models**:

```go
type NewButtonAction struct {
    PageConfigID       string `json:"pageConfigId" validate:"required"`
    ActionOrder        int    `json:"actionOrder"`
    IsActive           bool   `json:"isActive"`
    Label              string `json:"label" validate:"required"`
    TargetPath         string `json:"targetPath" validate:"required"`
    Variant            string `json:"variant" validate:"required,oneof=default secondary outline ghost destructive"`
    Alignment          string `json:"alignment" validate:"required,oneof=left right"`
    ConfirmationPrompt string `json:"confirmationPrompt"`
}

type NewDropdownAction struct {
    PageConfigID string             `json:"pageConfigId" validate:"required"`
    ActionOrder  int                `json:"actionOrder"`
    IsActive     bool               `json:"isActive"`
    Label        string             `json:"label" validate:"required"`
    Icon         string             `json:"icon"`
    Items        []NewDropdownItem  `json:"items" validate:"required,min=1,dive"`  // Must have items
}
```

**Why separate models?**
- **Type safety**: Each action type has different required fields
- **Validation clarity**: Field requirements are explicit per type
- **Better API documentation**: Clear contracts for each endpoint
- **Simpler implementation**: No complex discriminator logic needed

### 3. Atomic Dropdown Operations

Dropdowns and their items are **always** handled together:

```go
// Business layer ensures atomicity
func (b *Business) CreateDropdown(ctx context.Context, nda NewDropdownAction) (PageAction, error) {
    // Single transaction creates:
    // 1. Base page_action record (type='dropdown')
    // 2. Dropdown-specific data in page_action_dropdowns
    // 3. All dropdown items in page_action_dropdown_items
    return b.storer.CreateDropdown(ctx, action)
}

// Updates replace ALL items
func (b *Business) UpdateDropdown(ctx context.Context, action PageAction, uda UpdateDropdownAction) (PageAction, error) {
    // Single transaction:
    // 1. Delete existing items
    // 2. Insert new items
    // 3. Update dropdown data
}
```

### 4. Deterministic Query Ordering

**Challenge**: `action_order` is only unique per page_config, so multiple pages can have actions with the same order value (e.g., multiple pages each having action_order=0, 1, 2).

**Solution**: Multi-level sorting to ensure consistent, logical results:

```sql
ORDER BY page_config_id ASC, action_order ASC, id ASC
```

This groups all actions for each page together, ordered by their action_order within that page, with ID as a final tiebreaker.

### 5. Full Details on Query

The paginated `Query()` method fetches full type-specific details for each action:

```go
func (b *Business) Query(ctx context.Context, filter QueryFilter, orderBy order.By, pg page.Page) ([]PageAction, error) {
    // Get paginated list of action IDs
    actions, err := b.storer.Query(ctx, filter, orderBy, pg)

    // Fetch full details for each (including button/dropdown data)
    fullActions := make([]PageAction, len(actions))
    for i, action := range actions {
        fullAction, err := b.storer.QueryByID(ctx, action.ID)  // Joins to type tables
        fullActions[i] = fullAction
    }
    return fullActions
}
```

**Trade-off**: This performs N+1 queries but ensures consistent response structure. For better performance with large result sets, consider implementing a batch detail fetch.

### 6. Batch Create with Transactions

The batch endpoint creates multiple actions of any type in a single transaction:

```go
func (a *App) BatchCreate(ctx context.Context, app BatchCreateRequest) (PageActions, error) {
    tx, err := a.db.BeginTxx(ctx, nil)
    defer tx.Rollback()

    pageactionBusTx, err := a.pageactionbus.NewWithTx(tx)

    var createdActions []PageAction
    for _, action := range app.Actions {
        switch action.ActionType {
        case "button":
            created, err := pageactionBusTx.CreateButton(ctx, *action.Button)
        case "dropdown":
            created, err := pageactionBusTx.CreateDropdown(ctx, *action.Dropdown)
        case "separator":
            created, err := pageactionBusTx.CreateSeparator(ctx, *action.Separator)
        }
        createdActions = append(createdActions, created)
    }

    tx.Commit()
    return createdActions, nil
}
```

## Authorization Model

Page actions use **table-level permissions**:

- `page_actions` - General access control
- `page_action_buttons` - Button-specific permissions
- `page_action_dropdowns` - Dropdown-specific permissions

**Note**: `page_action_dropdown_items` inherits access from `page_action_dropdowns` since they're always bundled.

All mutation endpoints require **admin-only** access (`auth.RuleAdminOnly`). Read operations require authentication but respect table-level read permissions.

## Testing Strategy

Integration tests cover all 26 scenarios:

- **Query tests** (4): Paginated query, query by ID, query by page config, unauthorized query
- **Button tests** (6): Create/update/delete with 200/400/401 scenarios
- **Dropdown tests** (6): Create/update/delete with 200/400/401 scenarios
- **Separator tests** (6): Create/update/delete with 200/400/401 scenarios
- **Batch tests** (3): Successful batch, validation errors, unauthorized
- **Delete test** (1): Delete any action type

## Common Patterns

### Creating a Button
```json
POST /v1/config/page-actions/buttons
{
  "pageConfigId": "uuid",
  "actionOrder": 1,
  "isActive": true,
  "label": "Save",
  "targetPath": "/save",
  "variant": "default",
  "alignment": "right",
  "confirmationPrompt": "Save changes?"
}
```

### Creating a Dropdown with Items
```json
POST /v1/config/page-actions/dropdowns
{
  "pageConfigId": "uuid",
  "actionOrder": 2,
  "isActive": true,
  "label": "Actions",
  "icon": "more-vertical",
  "items": [
    {"label": "Edit", "targetPath": "/edit", "itemOrder": 1},
    {"label": "Delete", "targetPath": "/delete", "itemOrder": 2}
  ]
}
```

### Batch Create Mixed Actions
```json
POST /v1/config/page-configs/{id}/actions/batch
{
  "actions": [
    {
      "actionType": "button",
      "button": { /* button fields */ }
    },
    {
      "actionType": "dropdown",
      "dropdown": { /* dropdown fields including items */ }
    },
    {
      "actionType": "separator",
      "separator": { /* separator fields */ }
    }
  ]
}
```

## Future Considerations

1. **Performance**: Consider implementing a single-query fetch for full details in `Query()` using LEFT JOINs
2. **Versioning**: Action history/audit trail if actions need to be reverted
3. **Validation**: Action-level validation rules (e.g., max actions per page)
4. **Caching**: Page action lists could be cached per page_config_id
5. **Events**: Consider emitting domain events for action creation/updates for audit purposes

## Related Documentation

- FormData System: For creating page configs with actions in a single request
- Table Configs: Actions are displayed based on table configuration
- Permissions System: Table-level access control for actions
