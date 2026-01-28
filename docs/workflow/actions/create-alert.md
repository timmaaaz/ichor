# create_alert Action

Creates workflow alerts for users or roles. Alerts are persistent, queryable, and support acknowledgment/dismissal workflows.

## Configuration Schema

```json
{
  "alert_type": "string",
  "severity": "low|medium|high|critical",
  "title": "string (supports templates)",
  "message": "string (required, supports templates)",
  "context": {},
  "recipients": {
    "users": ["uuid1", "uuid2"],
    "roles": ["uuid3", "uuid4"]
  },
  "resolve_prior": false
}
```

**Source**: `business/sdk/workflow/workflowactions/communication/alert.go:24-35`

## Fields

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `alert_type` | string | No | - | Alert category (e.g., "order_notification") |
| `severity` | string | No | `medium` | Priority level |
| `title` | string | No | - | Alert title (template supported) |
| `message` | string | **Yes** | - | Alert message (template supported) |
| `context` | object | No | `{}` | Additional structured data |
| `recipients.users` | []string | Conditional | - | User UUIDs |
| `recipients.roles` | []string | Conditional | - | Role UUIDs |
| `resolve_prior` | bool | No | `false` | Auto-resolve prior alerts |

### Severity Levels

| Value | Description |
|-------|-------------|
| `low` | Low priority, informational |
| `medium` | Normal priority (default) |
| `high` | Important, requires attention |
| `critical` | Urgent, immediate action needed |

### Recipient Requirements

At least one recipient (user or role) is required. Alerts can be sent to:
- Specific users by UUID
- All users with a specific role

## Validation Rules

1. `message` is required and must be non-empty
2. At least one recipient (user or role) is required
3. All UUIDs must be valid format
4. `severity` must be valid if provided

**Source**: `business/sdk/workflow/workflowactions/communication/alert.go:59-81`

## Example Configurations

### Basic Alert

```json
{
  "alert_type": "order_notification",
  "severity": "medium",
  "title": "New Order",
  "message": "A new order has been created.",
  "recipients": {
    "users": ["5cf37266-3473-4006-984f-9325122678b7"]
  }
}
```

### Alert with Templates

```json
{
  "alert_type": "order_notification",
  "severity": "high",
  "title": "New Order: {{number}}",
  "message": "Order {{number}} for {{customer_name}} totaling {{total | currency:USD}} needs review.",
  "recipients": {
    "roles": ["sales-manager-role-uuid"]
  }
}
```

### Alert to Multiple Recipients

```json
{
  "alert_type": "approval_required",
  "severity": "high",
  "title": "Approval Required: {{number}}",
  "message": "Order {{number}} exceeds threshold and requires approval.",
  "recipients": {
    "users": ["manager-1-uuid", "manager-2-uuid"],
    "roles": ["approvers-role-uuid"]
  }
}
```

### Alert with Context Data

```json
{
  "alert_type": "inventory_warning",
  "severity": "critical",
  "title": "Low Stock: {{product_name}}",
  "message": "Product {{sku}} has only {{quantity}} units remaining.",
  "context": {
    "product_id": "{{product_id}}",
    "warehouse": "{{warehouse_name}}",
    "reorder_point": 100
  },
  "recipients": {
    "roles": ["inventory-managers"]
  }
}
```

## Execution Flow

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                        CreateAlertHandler                                    │
│                                                                             │
│  1. Parse action config                                                     │
│  2. Process template variables (title, message)                             │
│  3. Create Alert via alertbus.Create()                                      │
│  4. Create AlertRecipients via alertbus.CreateRecipients()                  │
└─────────────────────────────────────────────────────────────────────────────┘
```

## Database Tables

### workflow.alerts

| Column | Type | Description |
|--------|------|-------------|
| `id` | UUID | Primary key |
| `alert_type` | TEXT | Category |
| `severity` | TEXT | low/medium/high/critical |
| `title` | TEXT | Alert title |
| `message` | TEXT | Alert message |
| `context` | JSONB | Additional data |
| `source_entity_name` | TEXT | Triggering entity |
| `source_entity_id` | UUID | Triggering entity ID |
| `source_rule_id` | UUID | Rule that created alert |
| `status` | TEXT | active/acknowledged/dismissed/resolved |
| `expires_date` | TIMESTAMPTZ | Optional expiration |
| `created_date` | TIMESTAMPTZ | When created |
| `updated_date` | TIMESTAMPTZ | Last update |

### workflow.alert_recipients

| Column | Type | Description |
|--------|------|-------------|
| `id` | UUID | Primary key |
| `alert_id` | UUID | FK to alerts |
| `recipient_type` | TEXT | "user" or "role" |
| `recipient_id` | UUID | User/Role UUID |
| `created_date` | TIMESTAMPTZ | When created |

### workflow.alert_acknowledgments

| Column | Type | Description |
|--------|------|-------------|
| `id` | UUID | Primary key |
| `alert_id` | UUID | FK to alerts |
| `user_id` | UUID | Who acknowledged |
| `acknowledged_date` | TIMESTAMPTZ | When acknowledged |
| `notes` | TEXT | Optional notes |

## Alert Status Flow

```
active ──► acknowledged ──► dismissed
   │                            ▲
   └────────────────────────────┘
```

- `active` - New, actionable
- `acknowledged` - User has seen it
- `dismissed` - User hid it
- `resolved` - Auto-resolved by system

## API Integration

Alerts can be queried and managed via the Alert API:

| Endpoint | Description |
|----------|-------------|
| `GET /v1/workflow/alerts/mine` | Get current user's alerts |
| `GET /v1/workflow/alerts/{id}` | Get single alert |
| `POST /v1/workflow/alerts/{id}/acknowledge` | Acknowledge alert |
| `POST /v1/workflow/alerts/{id}/dismiss` | Dismiss alert |
| `GET /v1/workflow/alerts` | Admin: query all alerts |

See [API Reference](../api-reference.md) for full details.

## Frontend Integration

### Recommended UI

1. **Alert Badge** - Show count of active alerts
2. **Alert Dropdown** - List recent alerts with severity indicators
3. **Alert Detail** - Full message, context, acknowledge/dismiss buttons

### Severity Styling

| Severity | Suggested Treatment |
|----------|---------------------|
| `critical` | Red, prominent, possibly sound |
| `high` | Orange/red badge, top of list |
| `medium` | Yellow/amber, normal placement |
| `low` | Gray/muted, lower priority |

## Testing

Unit tests: `business/sdk/workflow/workflowactions/communication/alert_test.go`

Integration tests: `business/sdk/workflow/eventpublisher_integration_test.go`

```bash
# Run unit tests
go test -v ./business/sdk/workflow/workflowactions/communication/... -run Alert

# Run integration tests
go test -v ./business/sdk/workflow/... -run TestEventPublisher_CreateAlert
```

## Related Documentation

- [API Reference](../api-reference.md) - REST endpoints for alert management
- [Database Schema](../database-schema.md) - Alert table structure
- [Templates](../configuration/templates.md) - Template variables and filters
- [Rules](../configuration/rules.md) - How to create automation rules

## Key Files

| File | Purpose |
|------|---------|
| `business/sdk/workflow/workflowactions/communication/alert.go` | Handler implementation |
| `business/domain/workflow/alertbus/alertbus.go` | Business layer |
| `business/domain/workflow/alertbus/model.go` | Domain models |
| `api/domain/http/workflow/alertapi/` | HTTP API |
