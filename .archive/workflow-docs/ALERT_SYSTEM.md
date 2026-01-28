# Alert System Documentation

This document describes the workflow alert system architecture for future Claude sessions and developers working on frontend integration.

## Overview

The alert system allows workflow automation rules to create in-app alerts that are delivered to specific users or roles. Alerts are persistent, queryable, and support acknowledgment/dismissal workflows.

## Architecture

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                           WORKFLOW ENGINE                                    │
│                                                                             │
│  EventPublisher ──► RabbitMQ ──► QueueManager ──► Engine ──► ActionExecutor │
│                                                                             │
└────────────────────────────────────────┬────────────────────────────────────┘
                                         │
                                         ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                        CreateAlertHandler                                    │
│  (business/sdk/workflow/workflowactions/communication/alert.go)             │
│                                                                             │
│  • Parses action config (alert_type, severity, title, message, recipients)  │
│  • Resolves template variables: {{variable}} → execCtx.RawData[variable]    │
│  • Creates Alert via alertbus.Create()                                       │
│  • Creates AlertRecipients via alertbus.CreateRecipients()                  │
└────────────────────────────────────────┬────────────────────────────────────┘
                                         │
                                         ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                           alertbus (Business Layer)                          │
│  (business/domain/workflow/alertbus/alertbus.go)                            │
│                                                                             │
│  Methods:                                                                   │
│  • Create(ctx, alert) - Creates an alert                                    │
│  • CreateRecipients(ctx, recipients) - Batch insert recipients              │
│  • Query(ctx, filter, orderBy, page) - Admin query all alerts               │
│  • QueryMine(ctx, userID, roleIDs, filter, orderBy, page) - User's alerts   │
│  • QueryByID(ctx, id) - Single alert lookup                                 │
│  • Acknowledge(ctx, alertID, userID, roleIDs, notes, now) - Mark seen       │
│  • Dismiss(ctx, alertID, userID, roleIDs, now) - Hide alert                 │
│  • Count/CountMine - Pagination support                                     │
└────────────────────────────────────────┬────────────────────────────────────┘
                                         │
                                         ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                           alertdb (Store Layer)                              │
│  (business/domain/workflow/alertbus/stores/alertdb/alertdb.go)              │
│                                                                             │
│  • Pure CRUD operations against PostgreSQL                                  │
│  • Tables: workflow.alerts, workflow.alert_recipients,                      │
│            workflow.alert_acknowledgments                                   │
└─────────────────────────────────────────────────────────────────────────────┘
```

## Database Schema

### workflow.alerts
| Column | Type | Description |
|--------|------|-------------|
| id | UUID | Primary key |
| alert_type | TEXT | Category (e.g., "order_notification", "low_stock") |
| severity | TEXT | "low", "medium", "high", "critical" |
| title | TEXT | Short display title |
| message | TEXT | Full alert message |
| context | JSONB | Additional structured data |
| source_entity_name | TEXT | Entity that triggered alert (e.g., "orders") |
| source_entity_id | UUID | ID of triggering entity |
| source_rule_id | UUID | FK to workflow.automation_rules (nullable) |
| status | TEXT | "active", "acknowledged", "dismissed" |
| expires_date | TIMESTAMPTZ | Optional expiration |
| created_date | TIMESTAMPTZ | When alert was created |
| updated_date | TIMESTAMPTZ | Last modification |

### workflow.alert_recipients
| Column | Type | Description |
|--------|------|-------------|
| id | UUID | Primary key |
| alert_id | UUID | FK to workflow.alerts (CASCADE delete) |
| recipient_type | TEXT | "user" or "role" |
| recipient_id | UUID | User ID or Role ID |
| created_date | TIMESTAMPTZ | When recipient was added |

### workflow.alert_acknowledgments
| Column | Type | Description |
|--------|------|-------------|
| id | UUID | Primary key |
| alert_id | UUID | FK to workflow.alerts (CASCADE delete) |
| user_id | UUID | Who acknowledged |
| acknowledged_date | TIMESTAMPTZ | When acknowledged |
| notes | TEXT | Optional acknowledgment notes |

## API Endpoints

Base path: `/v1/workflow/alerts`

### User Endpoints (Authentication Required)

| Method | Path | Description |
|--------|------|-------------|
| GET | `/workflow/alerts/mine` | Get alerts for current user (filtered by user ID and role IDs) |
| GET | `/workflow/alerts/{id}` | Get single alert by ID |
| POST | `/workflow/alerts/{id}/acknowledge` | Mark alert as acknowledged |
| POST | `/workflow/alerts/{id}/dismiss` | Dismiss/hide alert |

### Admin Endpoints (Requires workflow.alerts read permission)

| Method | Path | Description |
|--------|------|-------------|
| GET | `/workflow/alerts` | Query all alerts (admin only) |

### Query Parameters (GET endpoints)

| Parameter | Description |
|-----------|-------------|
| page | Page number (default: 1) |
| rows | Rows per page (default: 20) |
| orderBy | Sort field (id, alertType, severity, status, createdDate, updatedDate) |
| id | Filter by alert ID |
| alertType | Filter by alert type |
| severity | Filter by severity |
| status | Filter by status |
| sourceEntityName | Filter by source entity |
| sourceEntityId | Filter by source entity ID |
| sourceRuleId | Filter by automation rule ID |

### Response Model

```json
{
  "id": "uuid",
  "alertType": "order_notification",
  "severity": "high",
  "title": "New Order: ORD-12345",
  "message": "A new order has been created: ORD-12345",
  "context": {},
  "sourceEntityName": "orders",
  "sourceEntityId": "uuid",
  "sourceRuleId": "uuid",
  "status": "active",
  "expiresDate": "2025-01-15T00:00:00Z",
  "createdDate": "2025-01-01T12:00:00Z",
  "updatedDate": "2025-01-01T12:00:00Z"
}
```

### Acknowledge Request Body

```json
{
  "notes": "I've reviewed this alert and taken action"
}
```

## Creating Alerts via Workflow Rules

Alerts are created by the `create_alert` action type in automation rules.

### Action Configuration Schema

```json
{
  "alert_type": "string (required)",
  "severity": "low|medium|high|critical (optional, defaults to medium)",
  "title": "string (required, supports {{template}} variables)",
  "message": "string (required, supports {{template}} variables)",
  "context": {},
  "recipients": {
    "users": ["uuid1", "uuid2"],
    "roles": ["uuid3", "uuid4"]
  }
}
```

### Template Variables

Title and message support `{{variable}}` syntax that gets replaced with values from the triggering entity's data:

```json
{
  "title": "Order {{number}} Status Changed",
  "message": "Order {{number}} for customer {{customer_name}} is now {{status}}"
}
```

When an order with `{"number": "ORD-001", "customer_name": "Acme Corp", "status": "shipped"}` triggers the rule, the message becomes:
"Order ORD-001 for customer Acme Corp is now shipped"

### Example Rule Action (via workflow API)

```json
{
  "automation_rule_id": "uuid",
  "name": "Send Order Alert",
  "action_config": {
    "alert_type": "order_notification",
    "severity": "high",
    "title": "New Order: {{number}}",
    "message": "A new order {{number}} has been created for {{customer_name}}",
    "recipients": {
      "users": ["5cf37266-3473-4006-984f-9325122678b7"],
      "roles": ["sales-team-role-uuid"]
    }
  },
  "execution_order": 1,
  "is_active": true
}
```

## Frontend Integration Guide

### Recommended Implementation

1. **Alert Notification Badge/Bell Icon**
   - Poll `GET /v1/workflow/alerts/mine?status=active` periodically (or use WebSocket when available)
   - Display count of active alerts
   - Filter by severity for visual indicators (red for critical/high, yellow for medium)

2. **Alert Dropdown/Panel**
   - Fetch recent alerts: `GET /v1/workflow/alerts/mine?rows=10&orderBy=-createdDate`
   - Show title, severity badge, time ago
   - Click to expand full message

3. **Alert Detail View**
   - Fetch single alert: `GET /v1/workflow/alerts/{id}`
   - Show full message, context data, source entity link
   - Provide acknowledge/dismiss buttons

4. **Acknowledge Flow**
   - POST to `/v1/workflow/alerts/{id}/acknowledge` with optional notes
   - Updates alert status to "acknowledged"
   - User can still see it but it's marked as reviewed

5. **Dismiss Flow**
   - POST to `/v1/workflow/alerts/{id}/dismiss`
   - Updates alert status to "dismissed"
   - Alert no longer appears in active queries

### Severity Levels for UI

| Severity | Suggested UI Treatment |
|----------|----------------------|
| critical | Red background, prominent placement, maybe sound/vibration |
| high | Red or orange badge, top of list |
| medium | Yellow/amber badge, normal placement |
| low | Gray or muted styling, lower priority |

### Status Flow

```
active ──► acknowledged ──► dismissed
   │                            ▲
   └────────────────────────────┘
```

Users can dismiss without acknowledging, or acknowledge first then dismiss later.

## Testing

### Unit Tests
Located in: `business/sdk/workflow/workflowactions/communication/alert_test.go`

Tests cover:
- Validation (missing fields, invalid severity, etc.)
- Execution (alert creation, recipients, template substitution)
- Error handling (invalid UUIDs)

### Integration Tests
Located in: `business/sdk/workflow/eventpublisher_integration_test.go`

`TestEventPublisher_CreateAlert` validates the full workflow:
1. Create automation rule with `create_alert` action
2. Fire event via EventPublisher
3. Verify alert persisted in database with correct properties
4. Verify template variables resolved
5. Verify recipients created

### Running Tests

```bash
# Unit tests only
go test -v ./business/sdk/workflow/workflowactions/communication/...

# Integration test
go test -v ./business/sdk/workflow/... -run TestEventPublisher_CreateAlert

# All workflow tests
go test -v ./business/sdk/workflow/...
```

## Key Files Reference

| File | Purpose |
|------|---------|
| `business/sdk/workflow/workflowactions/communication/alert.go` | CreateAlertHandler implementation |
| `business/sdk/workflow/workflowactions/communication/alert_test.go` | Unit tests |
| `business/domain/workflow/alertbus/alertbus.go` | Business layer |
| `business/domain/workflow/alertbus/model.go` | Domain models |
| `business/domain/workflow/alertbus/stores/alertdb/alertdb.go` | Database store |
| `api/domain/http/workflow/alertapi/alertapi.go` | HTTP handlers |
| `api/domain/http/workflow/alertapi/route.go` | Route registration |
| `api/domain/http/workflow/alertapi/model.go` | API models |
| `business/sdk/migrate/sql/migrate.sql` | Database migrations (1.76, 1.77, 1.78) |

## Common Patterns

### Checking if user is recipient
The business layer automatically filters alerts by recipient in `QueryMine`. It checks:
1. User is directly listed as a recipient (recipient_type='user', recipient_id=userID)
2. User has a role that is listed as recipient (recipient_type='role', recipient_id IN userRoleIDs)

### Authorization for acknowledge/dismiss
Users can only acknowledge/dismiss alerts they are recipients of. The business layer validates this before allowing the action.

### Batch recipient creation
Recipients are created in a single batch INSERT for performance. The `CreateRecipients` method handles this.

## Future Enhancements (Not Implemented)

- WebSocket/SSE for real-time alert delivery
- Email notification integration
- Push notification support
- Alert expiration auto-cleanup job
- Alert templates stored in database
- Alert categories/tags for filtering
- Bulk acknowledge/dismiss
