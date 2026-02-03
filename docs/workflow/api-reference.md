# API Reference

REST API endpoints for the workflow system.

## Alert API

Base path: `/v1/workflow/alerts`

### User Endpoints

These endpoints require authentication.

#### GET /workflow/alerts/mine

Get alerts for the current user (filtered by user ID and role IDs).

**Query Parameters:**

| Parameter | Type | Description |
|-----------|------|-------------|
| `page` | int | Page number (default: 1) |
| `rows` | int | Rows per page (default: 20) |
| `orderBy` | string | Sort field |
| `id` | string | Filter by alert ID |
| `alertType` | string | Filter by alert type |
| `severity` | string | Filter by severity |
| `status` | string | Filter by status |
| `sourceEntityName` | string | Filter by source entity |
| `sourceEntityId` | string | Filter by source entity ID |
| `sourceRuleId` | string | Filter by automation rule ID |

**Order By Values:**
- `id`
- `alertType`
- `severity`
- `status`
- `createdDate`
- `updatedDate`

Use `-` prefix for descending (e.g., `-createdDate`).

**Response:**

```json
{
  "items": [
    {
      "id": "uuid",
      "alertType": "order_notification",
      "severity": "high",
      "title": "New Order: ORD-12345",
      "message": "A new order has been created",
      "context": {},
      "sourceEntityName": "orders",
      "sourceEntityId": "uuid",
      "sourceRuleId": "uuid",
      "status": "active",
      "expiresDate": "2025-01-15T00:00:00Z",
      "createdDate": "2025-01-01T12:00:00Z",
      "updatedDate": "2025-01-01T12:00:00Z"
    }
  ],
  "total": 1,
  "page": 1,
  "rowsPerPage": 20
}
```

#### GET /workflow/alerts/{id}

Get a single alert by ID.

**Path Parameters:**

| Parameter | Type | Description |
|-----------|------|-------------|
| `id` | UUID | Alert ID |

**Response:**

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
  "expiresDate": null,
  "createdDate": "2025-01-01T12:00:00Z",
  "updatedDate": "2025-01-01T12:00:00Z"
}
```

#### POST /workflow/alerts/{id}/acknowledge

Mark an alert as acknowledged.

**Path Parameters:**

| Parameter | Type | Description |
|-----------|------|-------------|
| `id` | UUID | Alert ID |

**Request Body:**

```json
{
  "notes": "I've reviewed this alert and taken action"
}
```

**Response:**

```json
{
  "id": "uuid",
  "alertType": "order_notification",
  "severity": "high",
  "title": "New Order: ORD-12345",
  "message": "...",
  "status": "acknowledged",
  "createdDate": "2025-01-01T12:00:00Z",
  "updatedDate": "2025-01-01T12:30:00Z"
}
```

**Notes:**
- User must be a recipient of the alert
- Updates alert status to "acknowledged"
- Creates acknowledgment record with timestamp and notes

#### POST /workflow/alerts/{id}/dismiss

Dismiss/hide an alert.

**Path Parameters:**

| Parameter | Type | Description |
|-----------|------|-------------|
| `id` | UUID | Alert ID |

**Response:**

```json
{
  "id": "uuid",
  "alertType": "order_notification",
  "severity": "high",
  "title": "New Order: ORD-12345",
  "message": "...",
  "status": "dismissed",
  "createdDate": "2025-01-01T12:00:00Z",
  "updatedDate": "2025-01-01T12:30:00Z"
}
```

**Notes:**
- User must be a recipient of the alert
- Updates alert status to "dismissed"
- Alert no longer appears in active queries

### Bulk Action Endpoints

#### POST /workflow/alerts/acknowledge-selected

Acknowledge multiple alerts by ID.

**Request Body:**

```json
{
  "ids": ["uuid1", "uuid2", "uuid3"],
  "notes": "Reviewed all pending alerts (optional)"
}
```

**Response:**

```json
{
  "count": 3,
  "skipped": 0
}
```

**Notes:**
- Only acknowledges alerts where user is a recipient
- `skipped` indicates alerts where user was not a recipient

---

#### POST /workflow/alerts/acknowledge-all

Acknowledge all active alerts for the current user.

**Request Body:**

```json
{
  "notes": "Bulk acknowledgment (optional)"
}
```

**Response:**

```json
{
  "count": 15,
  "skipped": 0
}
```

---

#### POST /workflow/alerts/dismiss-selected

Dismiss multiple alerts by ID.

**Request Body:**

```json
{
  "ids": ["uuid1", "uuid2", "uuid3"],
  "notes": "Optional notes"
}
```

**Response:**

```json
{
  "count": 3,
  "skipped": 0
}
```

**Notes:**
- Only dismisses active alerts where user is a recipient
- `skipped` indicates alerts where user was not a recipient

---

#### POST /workflow/alerts/dismiss-all

Dismiss all active alerts for the current user.

**Request Body:**

```json
{
  "notes": "Optional notes"
}
```

**Response:**

```json
{
  "count": 10,
  "skipped": 0
}
```

### Debug/Test Endpoints

#### POST /workflow/alerts/test

Creates a test alert for the authenticated user (for E2E WebSocket testing).

**Request Body:** None required

**Response:**

```json
{
  "id": "uuid",
  "alertType": "test_alert",
  "severity": "medium",
  "title": "Test Alert",
  "message": "This is a test alert for E2E testing",
  "status": "active",
  "createdDate": "2025-01-01T12:00:00Z",
  "updatedDate": "2025-01-01T12:00:00Z"
}
```

**Notes:**
- Creates alert in database with current user as recipient
- Publishes to RabbitMQ for WebSocket delivery (if available)
- Useful for testing real-time alert functionality

### Admin Endpoints

These endpoints require `workflow.alerts` read permission.

#### GET /workflow/alerts

Query all alerts (admin only).

**Query Parameters:**

Same as `/workflow/alerts/mine`.

**Response:**

Same format as `/workflow/alerts/mine`.

## Response Models

### Alert

```json
{
  "id": "uuid",
  "alertType": "string",
  "severity": "low|medium|high|critical",
  "title": "string",
  "message": "string",
  "context": {},
  "sourceEntityName": "string",
  "sourceEntityId": "uuid",
  "sourceRuleId": "uuid",
  "status": "active|acknowledged|dismissed|resolved",
  "expiresDate": "timestamp or null",
  "createdDate": "timestamp",
  "updatedDate": "timestamp"
}
```

### Acknowledge Request

```json
{
  "notes": "string (optional)"
}
```

### Paginated Response

```json
{
  "items": [],
  "total": 0,
  "page": 1,
  "rowsPerPage": 20
}
```

## Error Responses

### 400 Bad Request

Invalid parameters or request body.

```json
{
  "error": "invalid_argument",
  "message": "Invalid alert ID format"
}
```

### 401 Unauthorized

Missing or invalid authentication.

```json
{
  "error": "unauthorized",
  "message": "Authentication required"
}
```

### 403 Forbidden

User is not a recipient of the alert.

```json
{
  "error": "forbidden",
  "message": "Not authorized to access this alert"
}
```

### 404 Not Found

Alert not found.

```json
{
  "error": "not_found",
  "message": "Alert not found"
}
```

## Authentication

All endpoints require JWT authentication via the `Authorization` header:

```
Authorization: Bearer <token>
```

## Authorization

- User endpoints: Any authenticated user (for their own alerts)
- Admin endpoints: Requires `workflow.alerts` table read permission

## Recipient Filtering

The `/workflow/alerts/mine` endpoint automatically filters alerts based on:

1. **Direct recipient**: User ID matches `recipient_type='user'` and `recipient_id`
2. **Role recipient**: User's role IDs match `recipient_type='role'` and `recipient_id`

This ensures users only see alerts they're supposed to receive.

## Rate Limiting

Consider implementing rate limiting for:
- Query endpoints: Prevent excessive polling
- Action endpoints: Prevent acknowledge/dismiss spam

## Pagination

All list endpoints support pagination:

```
GET /workflow/alerts/mine?page=2&rows=50
```

- Default page: 1
- Default rows: 20
- Maximum rows: 100 (recommended)

## Sorting

Use `orderBy` parameter with field name:

```
# Ascending (default)
GET /workflow/alerts/mine?orderBy=createdDate

# Descending
GET /workflow/alerts/mine?orderBy=-createdDate
```

## Filtering

Multiple filters are combined with AND logic:

```
GET /workflow/alerts/mine?status=active&severity=high
```

This returns alerts that are both active AND high severity.

## Edge API (Graph-Based Execution)

Base path: `/v1/workflow/rules/{ruleID}/edges`

The Edge API enables graph-based workflow execution with conditional branching. Edges define directed connections between actions, allowing for complex flows like if/else branching, diamond patterns, and parallel paths.

**Permission Table**: `workflow.automation_rules`

### Edge Types

| Type | Constant | Description |
|------|----------|-------------|
| `start` | `EdgeTypeStart` | Entry point into the graph (source is null) |
| `sequence` | `EdgeTypeSequence` | Linear progression, always followed |
| `always` | `EdgeTypeAlways` | Unconditional edge, always followed |
| `true_branch` | `EdgeTypeTrueBranch` | Only followed when condition evaluates to `true` |
| `false_branch` | `EdgeTypeFalseBranch` | Only followed when condition evaluates to `false` |

### GET /workflow/rules/{ruleID}/edges

List all edges for a rule.

**Path Parameters:**

| Parameter | Type | Description |
|-----------|------|-------------|
| `ruleID` | UUID | Rule ID |

**Response:**

```json
[
  {
    "id": "uuid",
    "rule_id": "uuid",
    "source_action_id": null,
    "target_action_id": "uuid",
    "edge_type": "start",
    "edge_order": 1,
    "created_date": "2025-01-01T12:00:00Z"
  },
  {
    "id": "uuid",
    "rule_id": "uuid",
    "source_action_id": "uuid",
    "target_action_id": "uuid",
    "edge_type": "sequence",
    "edge_order": 1,
    "created_date": "2025-01-01T12:00:00Z"
  }
]
```

---

### GET /workflow/rules/{ruleID}/edges/{edgeID}

Get a single edge by ID.

**Path Parameters:**

| Parameter | Type | Description |
|-----------|------|-------------|
| `ruleID` | UUID | Rule ID |
| `edgeID` | UUID | Edge ID |

**Response:**

```json
{
  "id": "uuid",
  "rule_id": "uuid",
  "source_action_id": "uuid",
  "target_action_id": "uuid",
  "edge_type": "true_branch",
  "edge_order": 1,
  "created_date": "2025-01-01T12:00:00Z"
}
```

---

### POST /workflow/rules/{ruleID}/edges

Create a new edge.

**Path Parameters:**

| Parameter | Type | Description |
|-----------|------|-------------|
| `ruleID` | UUID | Rule ID |

**Request Body:**

```json
{
  "source_action_id": "uuid-or-null",
  "target_action_id": "uuid",
  "edge_type": "sequence",
  "edge_order": 1
}
```

**Request Fields:**

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `source_action_id` | UUID/null | Conditional | Source action (null for `start` edges) |
| `target_action_id` | UUID | **Yes** | Target action |
| `edge_type` | string | **Yes** | One of: `start`, `sequence`, `true_branch`, `false_branch`, `always` |
| `edge_order` | int | No | Order for deterministic traversal (default: 0) |

**Validation Rules:**

- `start` edges MUST have `source_action_id` as null
- Non-start edges MUST have a `source_action_id`
- Target action must exist and belong to the specified rule
- Source action (if provided) must exist and belong to the specified rule

**Response:**

```json
{
  "id": "uuid",
  "rule_id": "uuid",
  "source_action_id": null,
  "target_action_id": "uuid",
  "edge_type": "start",
  "edge_order": 1,
  "created_date": "2025-01-01T12:00:00Z"
}
```

---

### DELETE /workflow/rules/{ruleID}/edges/{edgeID}

Delete a single edge.

**Path Parameters:**

| Parameter | Type | Description |
|-----------|------|-------------|
| `ruleID` | UUID | Rule ID |
| `edgeID` | UUID | Edge ID |

**Response:** `204 No Content`

---

### DELETE /workflow/rules/{ruleID}/edges-all

Delete all edges for a rule (reset to linear execution).

**Path Parameters:**

| Parameter | Type | Description |
|-----------|------|-------------|
| `ruleID` | UUID | Rule ID |

**Response:** `204 No Content`

**Notes:**
- After deleting all edges, the rule falls back to linear `execution_order` execution
- Useful for resetting a rule's action graph

---

## Cascade Visualization API

Endpoint for analyzing downstream workflow dependencies.

### GET /workflow/rules/{id}/cascade-map

Get all downstream workflows that could be triggered by a rule's actions.

**Path Parameters:**

| Parameter | Type | Description |
|-----------|------|-------------|
| `id` | UUID | Rule ID to analyze |

**Response:**

```json
{
  "rule_id": "uuid",
  "rule_name": "Order Status Update",
  "actions": [
    {
      "action_id": "uuid",
      "action_name": "Update Status",
      "action_type": "update_field",
      "modifies_entity": "sales.orders",
      "triggers_event": "on_update",
      "modified_fields": ["status"],
      "downstream_workflows": [
        {
          "rule_id": "uuid",
          "rule_name": "Order Shipped Notification",
          "trigger_conditions": {"status": {"operator": "equals", "value": "shipped"}},
          "will_trigger_if": "any sales.orders update (with conditions)"
        }
      ]
    },
    {
      "action_id": "uuid",
      "action_name": "Send Email",
      "action_type": "send_email",
      "downstream_workflows": []
    }
  ]
}
```

**Response Fields:**

| Field | Type | Description |
|-------|------|-------------|
| `rule_id` | string | ID of the analyzed rule |
| `rule_name` | string | Name of the analyzed rule |
| `actions` | array | Actions and their downstream effects |

**ActionCascadeInfo Fields:**

| Field | Type | Description |
|-------|------|-------------|
| `action_id` | string | Action ID |
| `action_name` | string | Action name |
| `action_type` | string | Action type (e.g., `update_field`) |
| `modifies_entity` | string | Entity being modified (if any) |
| `triggers_event` | string | Event type triggered (`on_create`, `on_update`, `on_delete`) |
| `modified_fields` | array | Fields being changed |
| `downstream_workflows` | array | Workflows that may be triggered |

**DownstreamWorkflowInfo Fields:**

| Field | Type | Description |
|-------|------|-------------|
| `rule_id` | string | Downstream rule ID |
| `rule_name` | string | Downstream rule name |
| `trigger_conditions` | object | Raw trigger conditions (nullable) |
| `will_trigger_if` | string | Human-readable trigger description |

**Notes:**
- Only active rules are included in downstream workflows
- Self-triggers are excluded (rule doesn't show itself)
- Only actions implementing `EntityModifier` show modifications (currently: `update_field`)

See [cascade-visualization.md](cascade-visualization.md) for detailed documentation.

---

## Rule Testing/Simulation API

Endpoints for testing rules without executing them.

### POST /workflow/rules/{id}/test

Test a rule with sample data (dry run simulation).

**Path Parameters:**

| Parameter | Type | Description |
|-----------|------|-------------|
| `id` | UUID | Rule ID to test |

**Request Body:**

```json
{
  "sample_data": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "status": "pending",
    "total": 1500.00,
    "customer_name": "Acme Corp"
  }
}
```

**Response:**

```json
{
  "rule_id": "uuid",
  "rule_name": "High Value Order Alert",
  "would_trigger": true,
  "matched_conditions": [
    {
      "field": "total",
      "operator": "greater_than",
      "expected": 1000,
      "actual": 1500,
      "matched": true
    }
  ],
  "actions_to_execute": [
    {
      "action_id": "uuid",
      "action_name": "Send Alert",
      "action_type": "create_alert",
      "order": 1,
      "preview": {
        "message": "High value order from Acme Corp"
      }
    }
  ],
  "template_preview": {
    "entity.id": "550e8400-e29b-41d4-a716-446655440000",
    "entity.status": "pending",
    "entity.total": "1500",
    "entity.customer_name": "Acme Corp"
  },
  "validation_errors": []
}
```

**Response Fields:**

| Field | Type | Description |
|-------|------|-------------|
| `rule_id` | UUID | Rule ID |
| `rule_name` | string | Rule name |
| `would_trigger` | bool | Whether the rule would trigger with this data |
| `matched_conditions` | array | Condition evaluation results |
| `actions_to_execute` | array | Actions that would run |
| `template_preview` | object | Template variable previews |
| `validation_errors` | array | Any validation issues |

**Condition Operators Supported:**

| Operator | Aliases | Description |
|----------|---------|-------------|
| `equals` | `eq`, `=`, `==` | Exact match |
| `not_equals` | `neq`, `!=`, `<>` | Not equal |
| `greater_than` | `gt`, `>` | Greater than |
| `greater_than_or_equals` | `gte`, `>=` | Greater than or equal |
| `less_than` | `lt`, `<` | Less than |
| `less_than_or_equals` | `lte`, `<=` | Less than or equal |
| `contains` | - | String contains |
| `not_contains` | - | String does not contain |
| `starts_with` | - | String starts with |
| `ends_with` | - | String ends with |
| `is_empty` | - | Value is empty/null |
| `is_not_empty` | - | Value is not empty |
| `in` | - | Value in list |
| `not_in` | - | Value not in list |

---

### GET /workflow/rules/{id}/executions

List execution history for a rule.

**Path Parameters:**

| Parameter | Type | Description |
|-----------|------|-------------|
| `id` | UUID | Rule ID |

**Query Parameters:**

| Parameter | Type | Description |
|-----------|------|-------------|
| `page` | int | Page number (default: 1) |
| `rows` | int | Rows per page (default: 20) |

**Response:**

```json
{
  "items": [
    {
      "id": "uuid",
      "rule_id": "uuid",
      "entity_id": "uuid",
      "entity_name": "orders",
      "trigger_type": "on_update",
      "status": "completed",
      "started_at": "2025-01-01T12:00:00Z",
      "completed_at": "2025-01-01T12:00:05Z",
      "error_message": null
    }
  ],
  "total": 50,
  "page": 1,
  "rowsPerPage": 20
}
```

---

## Error Responses

All APIs use consistent error responses.

### 400 Bad Request

Invalid parameters or request body.

```json
{
  "error": "invalid_argument",
  "message": "Invalid rule ID format"
}
```

### 401 Unauthorized

Missing or invalid authentication.

```json
{
  "error": "unauthenticated",
  "message": "Authentication required"
}
```

### 403 Forbidden

User lacks required permissions.

```json
{
  "error": "permission_denied",
  "message": "Not authorized to access this resource"
}
```

### 404 Not Found

Resource not found.

```json
{
  "error": "not_found",
  "message": "Rule not found"
}
```

### 412 Precondition Failed

Validation failed.

```json
{
  "error": "failed_precondition",
  "message": "start edges must not have a source_action_id"
}
```

---

## Key Files

| File | Purpose |
|------|---------|
| `api/domain/http/workflow/alertapi/alertapi.go` | Alert HTTP handlers |
| `api/domain/http/workflow/alertapi/route.go` | Alert route definitions |
| `api/domain/http/workflow/alertapi/model.go` | Alert API models |
| `api/domain/http/workflow/alertapi/filter.go` | Alert filter parsing |
| `api/domain/http/workflow/edgeapi/edgeapi.go` | Edge HTTP handlers |
| `api/domain/http/workflow/edgeapi/route.go` | Edge route definitions |
| `api/domain/http/workflow/edgeapi/model.go` | Edge API models |
| `api/domain/http/workflow/ruleapi/cascade.go` | Cascade visualization handler |
| `api/domain/http/workflow/ruleapi/simulate.go` | Rule simulation handler |
| `api/domain/http/workflow/ruleapi/route.go` | Rule route definitions |
