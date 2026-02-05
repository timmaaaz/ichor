# Workflow Save API - Frontend Integration Guide

This document provides frontend developers with everything needed to integrate with the Workflow Save API for creating and managing workflow automation rules.

---

## Table of Contents

1. [Overview](#overview)
2. [API Endpoints](#api-endpoints)
3. [Request/Response Models](#requestresponse-models)
4. [Action Types Reference](#action-types-reference)
5. [Edge Types & Graph Structure](#edge-types--graph-structure)
6. [Trigger System](#trigger-system)
7. [Temporary ID System](#temporary-id-system)
8. [Canvas Layout](#canvas-layout)
9. [Validation Rules](#validation-rules)
10. [Example Workflows](#example-workflows)
11. [Error Handling](#error-handling)

---

## Overview

The Workflow Save API allows atomic creation and updates of complete workflows including:
- Rule metadata (name, description, trigger configuration)
- Actions (the operations to perform)
- Edges (connections defining execution flow)
- Canvas layout (visual editor state)

All operations are transactional - either everything saves successfully or nothing changes.

---

## API Endpoints

### Create Workflow
```
POST /v1/workflow/rules/full
Authorization: Bearer <token>
Content-Type: application/json
```

### Update Workflow
```
PUT /v1/workflow/rules/{id}/full
Authorization: Bearer <token>
Content-Type: application/json
```

Both endpoints accept the same request body format.

---

## Request/Response Models

### SaveWorkflowRequest

```typescript
interface SaveWorkflowRequest {
  name: string;                    // Required, 1-255 chars
  description?: string;            // Optional, max 1000 chars
  is_active: boolean;              // Whether rule should fire
  entity_id: string;               // UUID - which entity triggers this
  trigger_type_id: string;         // UUID - when to trigger (on_create, on_update, etc.)
  trigger_conditions?: object;     // Optional - additional conditions
  actions: SaveActionRequest[];    // Required, min 1
  edges: SaveEdgeRequest[];        // Required for graph structure
  canvas_layout?: object;          // Optional - visual editor state
}
```

### SaveActionRequest

```typescript
interface SaveActionRequest {
  id?: string;                     // null for new, UUID for existing
  name: string;                    // Required, 1-255 chars
  description?: string;            // Optional, max 1000 chars
  action_type: ActionType;         // Required - see Action Types
  action_config: object;           // Required - type-specific config
  execution_order: number;         // Required, min 1
  is_active: boolean;              // Whether action should execute
}

type ActionType =
  | 'create_alert'
  | 'send_email'
  | 'send_notification'
  | 'update_field'
  | 'seek_approval'
  | 'allocate_inventory'
  | 'evaluate_condition';
```

### SaveEdgeRequest

```typescript
interface SaveEdgeRequest {
  source_action_id?: string;       // null for start edge, temp:N or UUID
  target_action_id: string;        // Required - temp:N or UUID
  edge_type: EdgeType;             // Required
  edge_order?: number;             // Optional, defaults to 0
}

type EdgeType =
  | 'start'           // Entry point (no source)
  | 'sequence'        // Linear flow
  | 'true_branch'     // Condition evaluated true
  | 'false_branch'    // Condition evaluated false
  | 'always';         // Always execute (parallel)
```

### SaveWorkflowResponse

```typescript
interface SaveWorkflowResponse {
  id: string;                      // UUID of saved rule
  name: string;
  description: string;
  is_active: boolean;
  entity_id: string;
  trigger_type_id: string;
  trigger_conditions: object | null;
  actions: SaveActionResponse[];
  edges: SaveEdgeResponse[];
  canvas_layout: object | null;
  created_date: string;            // ISO timestamp
  updated_date: string;            // ISO timestamp
}

interface SaveActionResponse {
  id: string;                      // Real UUID (temp IDs resolved)
  name: string;
  description: string;
  action_type: string;
  action_config: object;
  execution_order: number;
  is_active: boolean;
}

interface SaveEdgeResponse {
  id: string;                      // UUID
  source_action_id: string;        // Real UUID (temp IDs resolved)
  target_action_id: string;        // Real UUID (temp IDs resolved)
  edge_type: string;
  edge_order: number;
}
```

---

## Action Types Reference

### create_alert

Creates an alert record visible in the alerts dashboard.

```typescript
interface CreateAlertConfig {
  alert_type: string;              // Required - category of alert
  severity: 'info' | 'warning' | 'error' | 'critical';  // Required
  title: string;                   // Required
  message: string;                 // Required - supports template vars
}
```

**Example:**
```json
{
  "action_type": "create_alert",
  "action_config": {
    "alert_type": "order_issue",
    "severity": "warning",
    "title": "Order Requires Attention",
    "message": "Order #{{order_id}} has been pending for 24 hours"
  }
}
```

### send_email

Sends an email to specified recipients.

```typescript
interface SendEmailConfig {
  recipients: string[];            // Required - email addresses or template vars
  subject: string;                 // Required
  body: string;                    // Required - supports template vars
}
```

**Example:**
```json
{
  "action_type": "send_email",
  "action_config": {
    "recipients": ["{{customer_email}}", "sales@company.com"],
    "subject": "Your order has shipped!",
    "body": "Order #{{order_id}} is on its way. Track at: {{tracking_url}}"
  }
}
```

### send_notification

Sends in-app notifications.

```typescript
interface SendNotificationConfig {
  recipients: string[];            // Required - user IDs or template vars
  channels: string[];              // Required - notification channels
}
```

**Example:**
```json
{
  "action_type": "send_notification",
  "action_config": {
    "recipients": ["{{assigned_user_id}}"],
    "channels": ["in_app", "push"]
  }
}
```

### update_field

Updates a field on the target entity or related entity.

```typescript
interface UpdateFieldConfig {
  target_entity: string;           // Required - entity to update
  target_field: string;            // Required - field name
  value?: any;                     // New value (supports templates)
}
```

**Example:**
```json
{
  "action_type": "update_field",
  "action_config": {
    "target_entity": "orders",
    "target_field": "status",
    "value": "processed"
  }
}
```

### seek_approval

Creates an approval request.

```typescript
interface SeekApprovalConfig {
  approvers: string[];             // Required - user IDs
  approval_type: string;           // Required - type of approval
  timeout_hours?: number;          // Optional - auto-escalate after
}
```

**Example:**
```json
{
  "action_type": "seek_approval",
  "action_config": {
    "approvers": ["{{manager_id}}"],
    "approval_type": "expense_approval",
    "timeout_hours": 48
  }
}
```

### allocate_inventory

Allocates inventory items (async operation).

```typescript
interface AllocateInventoryConfig {
  inventory_items: InventoryItem[];  // Required
  allocation_mode: 'fifo' | 'lifo' | 'fefo';  // Required
}

interface InventoryItem {
  product_id: string;
  quantity: number;
}
```

**Example:**
```json
{
  "action_type": "allocate_inventory",
  "action_config": {
    "inventory_items": [
      { "product_id": "{{product_id}}", "quantity": "{{quantity}}" }
    ],
    "allocation_mode": "fifo"
  }
}
```

### evaluate_condition

Evaluates conditions and branches execution flow.

```typescript
interface EvaluateConditionConfig {
  conditions: Condition[];         // Required
}

interface Condition {
  field_name: string;
  operator: ConditionOperator;
  value: any;
}

type ConditionOperator =
  | 'equals'
  | 'not_equals'
  | 'greater_than'
  | 'less_than'
  | 'contains'
  | 'in';
```

**Example:**
```json
{
  "action_type": "evaluate_condition",
  "action_config": {
    "conditions": [
      { "field_name": "amount", "operator": "greater_than", "value": 1000 }
    ]
  }
}
```

**Important:** Condition actions MUST have both `true_branch` and `false_branch` edges connecting to subsequent actions.

---

## Edge Types & Graph Structure

### Edge Types

| Type | Source | Usage |
|------|--------|-------|
| `start` | null | Entry point - exactly ONE required |
| `sequence` | Action ID | Linear flow from one action to next |
| `true_branch` | Condition ID | Path when condition is true |
| `false_branch` | Condition ID | Path when condition is false |
| `always` | Action ID | Parallel execution path |

### Graph Rules

1. **Exactly one `start` edge** - Entry point to the workflow
2. **No cycles** - Cannot create loops (action A -> B -> A)
3. **All actions reachable** - Every action must have an incoming edge
4. **Conditions need both branches** - `evaluate_condition` must have `true_branch` AND `false_branch`

### Visual Examples

**Linear Sequence:**
```
[start] --> [Action 1] --sequence--> [Action 2] --sequence--> [Action 3]
```

**Branching:**
```
[start] --> [Condition] --true_branch--> [High Value Action]
                        --false_branch--> [Standard Action]
```

**Convergent Branches:**
```
[start] --> [Condition] --true_branch--> [Path A] --sequence--> [Final Action]
                        --false_branch--> [Path B] --sequence--> [Final Action]
```

---

## Trigger System

### Getting Trigger Types

First, fetch available trigger types:

```
GET /v1/workflow/trigger-types
```

Returns:
```json
[
  { "id": "uuid-1", "name": "on_create", "description": "Fired when entity is created" },
  { "id": "uuid-2", "name": "on_update", "description": "Fired when entity is updated" },
  { "id": "uuid-3", "name": "on_delete", "description": "Fired when entity is deleted" },
  { "id": "uuid-4", "name": "scheduled", "description": "Fired on a schedule" }
]
```

### Trigger Conditions

Define **when** within the trigger type the rule should fire.

```typescript
interface TriggerConditions {
  field_conditions: FieldCondition[];
}

interface FieldCondition {
  field_name: string;
  operator: TriggerOperator;
  value: any;
  previous_value?: any;            // For changed_from/changed_to
}

type TriggerOperator =
  | 'equals'
  | 'not_equals'
  | 'changed_to'
  | 'changed_from'
  | 'greater_than'
  | 'less_than'
  | 'contains'
  | 'in';
```

### Examples

**Fire on any create:**
```json
{
  "trigger_type_id": "uuid-of-on_create",
  "trigger_conditions": null
}
```

**Fire when status changes to "shipped":**
```json
{
  "trigger_type_id": "uuid-of-on_update",
  "trigger_conditions": {
    "field_conditions": [
      {
        "field_name": "status",
        "operator": "changed_to",
        "value": "shipped"
      }
    ]
  }
}
```

**Fire on high-value priority orders:**
```json
{
  "trigger_type_id": "uuid-of-on_update",
  "trigger_conditions": {
    "field_conditions": [
      {
        "field_name": "amount",
        "operator": "greater_than",
        "value": 1000
      },
      {
        "field_name": "is_priority",
        "operator": "equals",
        "value": true
      }
    ]
  }
}
```

**Note:** Multiple conditions use AND logic - all must match.

---

## Temporary ID System

When creating new actions, use temporary IDs (`temp:N`) to reference them in edges before they have real UUIDs.

### How It Works

1. Assign `id: null` to new actions
2. Reference by array index: `temp:0`, `temp:1`, etc.
3. Server resolves to real UUIDs in response

### Example

**Request:**
```json
{
  "actions": [
    { "id": null, "name": "First Action", ... },   // Index 0
    { "id": null, "name": "Second Action", ... }   // Index 1
  ],
  "edges": [
    { "source_action_id": null, "target_action_id": "temp:0", "edge_type": "start" },
    { "source_action_id": "temp:0", "target_action_id": "temp:1", "edge_type": "sequence" }
  ]
}
```

**Response:**
```json
{
  "actions": [
    { "id": "abc-123", "name": "First Action", ... },
    { "id": "def-456", "name": "Second Action", ... }
  ],
  "edges": [
    { "source_action_id": "", "target_action_id": "abc-123", "edge_type": "start" },
    { "source_action_id": "abc-123", "target_action_id": "def-456", "edge_type": "sequence" }
  ]
}
```

### Update Scenarios

| Action State | Request `id` | Result |
|--------------|--------------|--------|
| New action | `null` | Created, assigned UUID |
| Existing action | `"abc-123"` (UUID) | Updated |
| Removed from request | (not included) | Deleted |

---

## Canvas Layout

Store visual editor state in `canvas_layout`. This is passed through without validation - use any structure your frontend needs.

**Suggested structure:**
```typescript
interface CanvasLayout {
  viewport: {
    x: number;
    y: number;
    zoom: number;
  };
  node_positions: {
    [actionId: string]: {          // Use temp:N for new, UUID for existing
      x: number;
      y: number;
    };
  };
}
```

**Example:**
```json
{
  "canvas_layout": {
    "viewport": { "x": 0, "y": 0, "zoom": 1.0 },
    "node_positions": {
      "temp:0": { "x": 100, "y": 50 },
      "temp:1": { "x": 100, "y": 200 },
      "temp:2": { "x": 250, "y": 350 }
    }
  }
}
```

---

## Validation Rules

The API validates:

### Request Validation
- `name`: Required, 1-255 characters
- `description`: Max 1000 characters
- `entity_id`: Required, valid UUID
- `trigger_type_id`: Required, valid UUID
- `actions`: Required, minimum 1

### Action Validation
- `name`: Required, 1-255 characters
- `action_type`: Must be one of the valid types
- `action_config`: Required, must have type-specific required fields
- `execution_order`: Required, minimum 1

### Graph Validation
- No cycles detected
- All actions reachable from start
- Exactly one start edge
- Condition actions have both branch edges

### Action Config Validation

| Action Type | Required Fields |
|-------------|-----------------|
| `create_alert` | `alert_type`, `severity`, `title`, `message` |
| `send_email` | `recipients`, `subject`, `body` |
| `send_notification` | `recipients`, `channels` |
| `update_field` | `target_entity`, `target_field` |
| `seek_approval` | `approvers`, `approval_type` |
| `allocate_inventory` | `inventory_items`, `allocation_mode` |
| `evaluate_condition` | `conditions` |

---

## Example Workflows

### Simple Alert on Create

```json
{
  "name": "New Customer Alert",
  "description": "Alert sales team when new customer is created",
  "is_active": true,
  "entity_id": "customers-entity-uuid",
  "trigger_type_id": "on-create-trigger-uuid",
  "trigger_conditions": null,
  "actions": [
    {
      "id": null,
      "name": "Create Alert",
      "action_type": "create_alert",
      "execution_order": 1,
      "is_active": true,
      "action_config": {
        "alert_type": "new_customer",
        "severity": "info",
        "title": "New Customer Created",
        "message": "{{customer_name}} has been added to the system"
      }
    }
  ],
  "edges": [
    { "source_action_id": null, "target_action_id": "temp:0", "edge_type": "start" }
  ]
}
```

### Sequential Actions

```json
{
  "name": "Order Shipped Workflow",
  "is_active": true,
  "entity_id": "orders-entity-uuid",
  "trigger_type_id": "on-update-trigger-uuid",
  "trigger_conditions": {
    "field_conditions": [
      { "field_name": "status", "operator": "changed_to", "value": "shipped" }
    ]
  },
  "actions": [
    {
      "id": null,
      "name": "Update Shipped Date",
      "action_type": "update_field",
      "execution_order": 1,
      "is_active": true,
      "action_config": {
        "target_entity": "orders",
        "target_field": "shipped_date",
        "value": "{{current_timestamp}}"
      }
    },
    {
      "id": null,
      "name": "Email Customer",
      "action_type": "send_email",
      "execution_order": 2,
      "is_active": true,
      "action_config": {
        "recipients": ["{{customer_email}}"],
        "subject": "Your order has shipped!",
        "body": "Order #{{order_id}} is on its way."
      }
    },
    {
      "id": null,
      "name": "Create Shipping Alert",
      "action_type": "create_alert",
      "execution_order": 3,
      "is_active": true,
      "action_config": {
        "alert_type": "order_shipped",
        "severity": "info",
        "title": "Order Shipped",
        "message": "Order #{{order_id}} shipped to {{customer_name}}"
      }
    }
  ],
  "edges": [
    { "source_action_id": null, "target_action_id": "temp:0", "edge_type": "start" },
    { "source_action_id": "temp:0", "target_action_id": "temp:1", "edge_type": "sequence" },
    { "source_action_id": "temp:1", "target_action_id": "temp:2", "edge_type": "sequence" }
  ]
}
```

### Branching Workflow

```json
{
  "name": "High Value Order Processing",
  "is_active": true,
  "entity_id": "orders-entity-uuid",
  "trigger_type_id": "on-create-trigger-uuid",
  "actions": [
    {
      "id": null,
      "name": "Check Order Value",
      "action_type": "evaluate_condition",
      "execution_order": 1,
      "is_active": true,
      "action_config": {
        "conditions": [
          { "field_name": "total_amount", "operator": "greater_than", "value": 1000 }
        ]
      }
    },
    {
      "id": null,
      "name": "Request Approval",
      "action_type": "seek_approval",
      "execution_order": 2,
      "is_active": true,
      "action_config": {
        "approvers": ["{{sales_manager_id}}"],
        "approval_type": "high_value_order"
      }
    },
    {
      "id": null,
      "name": "Auto Approve",
      "action_type": "update_field",
      "execution_order": 2,
      "is_active": true,
      "action_config": {
        "target_entity": "orders",
        "target_field": "status",
        "value": "approved"
      }
    }
  ],
  "edges": [
    { "source_action_id": null, "target_action_id": "temp:0", "edge_type": "start" },
    { "source_action_id": "temp:0", "target_action_id": "temp:1", "edge_type": "true_branch" },
    { "source_action_id": "temp:0", "target_action_id": "temp:2", "edge_type": "false_branch" }
  ]
}
```

---

## Error Handling

### Error Response Format

```typescript
interface ErrorResponse {
  code: string;                    // Error code
  message: string;                 // Human-readable message
}
```

### Common Error Codes

| Code | HTTP Status | Meaning |
|------|-------------|---------|
| `InvalidArgument` | 400 | Validation failed |
| `NotFound` | 404 | Rule ID doesn't exist |
| `Unauthenticated` | 401 | Missing/invalid token |
| `PermissionDenied` | 403 | Insufficient permissions |
| `Internal` | 500 | Server error |

### Validation Error Examples

**Missing required field:**
```json
{
  "code": "InvalidArgument",
  "message": "validate: name is required"
}
```

**Invalid action type:**
```json
{
  "code": "InvalidArgument",
  "message": "validate: action_type must be one of [create_alert send_email send_notification update_field seek_approval allocate_inventory evaluate_condition]"
}
```

**Graph cycle detected:**
```json
{
  "code": "InvalidArgument",
  "message": "graph: cycle detected in workflow graph"
}
```

**Missing action config field:**
```json
{
  "code": "InvalidArgument",
  "message": "action config: create_alert requires severity"
}
```

**Invalid temp ID reference:**
```json
{
  "code": "InvalidArgument",
  "message": "edge references invalid action index: temp:99"
}
```

---

## TypeScript Client Example

```typescript
interface WorkflowClient {
  createWorkflow(request: SaveWorkflowRequest): Promise<SaveWorkflowResponse>;
  updateWorkflow(id: string, request: SaveWorkflowRequest): Promise<SaveWorkflowResponse>;
  getTriggerTypes(): Promise<TriggerType[]>;
  getEntityTypes(): Promise<EntityType[]>;
}

class WorkflowApiClient implements WorkflowClient {
  constructor(private baseUrl: string, private getToken: () => string) {}

  async createWorkflow(request: SaveWorkflowRequest): Promise<SaveWorkflowResponse> {
    const response = await fetch(`${this.baseUrl}/v1/workflow/rules/full`, {
      method: 'POST',
      headers: {
        'Authorization': `Bearer ${this.getToken()}`,
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(request),
    });

    if (!response.ok) {
      const error = await response.json();
      throw new WorkflowApiError(error.code, error.message);
    }

    return response.json();
  }

  async updateWorkflow(id: string, request: SaveWorkflowRequest): Promise<SaveWorkflowResponse> {
    const response = await fetch(`${this.baseUrl}/v1/workflow/rules/${id}/full`, {
      method: 'PUT',
      headers: {
        'Authorization': `Bearer ${this.getToken()}`,
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(request),
    });

    if (!response.ok) {
      const error = await response.json();
      throw new WorkflowApiError(error.code, error.message);
    }

    return response.json();
  }
}

class WorkflowApiError extends Error {
  constructor(public code: string, message: string) {
    super(message);
  }
}
```

---

## Summary

| Concept | Key Points |
|---------|------------|
| **Endpoints** | `POST /v1/workflow/rules/full` (create), `PUT /v1/workflow/rules/{id}/full` (update) |
| **Transactions** | All-or-nothing saves |
| **New Actions** | Use `id: null`, reference with `temp:N` |
| **Existing Actions** | Use real UUID |
| **Deleted Actions** | Simply omit from request |
| **Edges** | Define execution flow, exactly one `start` edge required |
| **Branching** | Use `evaluate_condition` with `true_branch`/`false_branch` edges |
| **Trigger Conditions** | Optional filtering on when rule fires |
| **Canvas Layout** | Passthrough JSON for visual editor state |
