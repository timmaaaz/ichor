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

## Key Files

| File | Purpose |
|------|---------|
| `api/domain/http/workflow/alertapi/alertapi.go` | HTTP handlers |
| `api/domain/http/workflow/alertapi/route.go` | Route definitions |
| `api/domain/http/workflow/alertapi/model.go` | API models |
| `api/domain/http/workflow/alertapi/filter.go` | Filter parsing |
