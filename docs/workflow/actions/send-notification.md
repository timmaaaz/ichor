# send_notification Action

Sends multi-channel notifications to specified recipients.

## Configuration Schema

```json
{
  "recipients": ["user-uuid-1", "user-uuid-2"],
  "channels": [
    {
      "type": "email|sms|push|in_app",
      "config": {}
    }
  ],
  "priority": "low|medium|high|critical"
}
```

**Source**: `business/sdk/workflow/workflowactions/communication/notification.go:32-39`

## Fields

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `recipients` | []string | **Yes** | - | Recipient user IDs |
| `channels` | []Channel | **Yes** | - | Delivery channels |
| `priority` | string | **Yes** | - | Priority level |

### Channel Types

| Type | Description |
|------|-------------|
| `email` | Send via email |
| `sms` | Send via SMS |
| `push` | Send push notification |
| `in_app` | In-app notification |

### Priority Levels

| Priority | Description |
|----------|-------------|
| `low` | Low priority |
| `medium` | Normal priority |
| `high` | High priority |
| `critical` | Urgent, immediate delivery |

## Validation Rules

1. `recipients` list is required and must not be empty
2. At least one channel is required
3. `priority` must be a valid level

**Source**: `business/sdk/workflow/workflowactions/communication/notification.go:32-61`

## Example Configurations

### Single Channel Notification

```json
{
  "recipients": ["5cf37266-3473-4006-984f-9325122678b7"],
  "channels": [
    {
      "type": "email",
      "config": {
        "subject": "Order Update",
        "body": "Your order has been updated."
      }
    }
  ],
  "priority": "medium"
}
```

### Multi-Channel Notification

```json
{
  "recipients": ["manager-uuid"],
  "channels": [
    {
      "type": "email",
      "config": {
        "subject": "Urgent: Approval Required",
        "body": "An order requires your approval."
      }
    },
    {
      "type": "push",
      "config": {
        "title": "Approval Required",
        "body": "Tap to review"
      }
    },
    {
      "type": "in_app",
      "config": {
        "message": "New approval request pending"
      }
    }
  ],
  "priority": "high"
}
```

### Critical Notification

```json
{
  "recipients": ["admin-uuid-1", "admin-uuid-2"],
  "channels": [
    {
      "type": "sms",
      "config": {
        "message": "CRITICAL: System alert requires immediate attention"
      }
    },
    {
      "type": "email",
      "config": {
        "subject": "CRITICAL: System Alert",
        "body": "Please check the system immediately."
      }
    }
  ],
  "priority": "critical"
}
```

## Channel Configuration

Each channel type has its own configuration schema:

### Email Channel

```json
{
  "type": "email",
  "config": {
    "subject": "string",
    "body": "string (templates supported)"
  }
}
```

### SMS Channel

```json
{
  "type": "sms",
  "config": {
    "message": "string (templates supported, max 160 chars recommended)"
  }
}
```

### Push Channel

```json
{
  "type": "push",
  "config": {
    "title": "string",
    "body": "string",
    "data": {}
  }
}
```

### In-App Channel

```json
{
  "type": "in_app",
  "config": {
    "message": "string"
  }
}
```

## Delivery Tracking

Notification deliveries are tracked in `workflow.notification_deliveries`:

| Column | Type | Description |
|--------|------|-------------|
| `id` | UUID | Primary key |
| `notification_id` | UUID | Parent notification |
| `channel` | TEXT | Channel type |
| `recipient_id` | UUID | User ID |
| `status` | TEXT | pending/sent/delivered/failed/bounced/retrying |
| `sent_at` | TIMESTAMPTZ | When sent |
| `delivered_at` | TIMESTAMPTZ | When delivered |
| `error_message` | TEXT | Error if failed |

### Delivery Status

| Status | Description |
|--------|-------------|
| `pending` | Awaiting delivery |
| `sent` | Sent to provider |
| `delivered` | Confirmed delivered |
| `failed` | Delivery failed |
| `bounced` | Bounced back |
| `retrying` | Retry in progress |

## Use Cases

1. **Multi-channel alerts** - Reach users on their preferred channel
2. **Escalation** - Start with in-app, escalate to SMS for critical
3. **Redundancy** - Send via multiple channels for important messages
4. **User preferences** - Different users receive via different channels

## Key Files

| File | Purpose |
|------|---------|
| `business/sdk/workflow/workflowactions/communication/notification.go` | Handler implementation |
