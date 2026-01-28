# seek_approval Action

Initiates approval workflows by requesting approval from designated approvers.

## Configuration Schema

```json
{
  "approvers": ["user-uuid-1", "user-uuid-2"],
  "approval_type": "any|all|majority"
}
```

**Source**: `business/sdk/workflow/workflowactions/approval/seek.go:32-36`

## Fields

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `approvers` | []string | **Yes** | - | Approver user UUIDs |
| `approval_type` | string | **Yes** | - | Approval mode |

### Approval Types

| Type | Description |
|------|-------------|
| `any` | Any one approver can approve |
| `all` | All approvers must approve |
| `majority` | Majority of approvers must approve |

## Validation Rules

1. `approvers` list is required and must not be empty
2. `approval_type` must be: `any`, `all`, or `majority`

**Source**: `business/sdk/workflow/workflowactions/approval/seek.go:32-52`

## Example Configurations

### Single Approver

```json
{
  "approvers": ["manager-uuid"],
  "approval_type": "any"
}
```

### Require All Approvers

```json
{
  "approvers": [
    "finance-manager-uuid",
    "operations-manager-uuid",
    "ceo-uuid"
  ],
  "approval_type": "all"
}
```

### Majority Approval

```json
{
  "approvers": [
    "board-member-1-uuid",
    "board-member-2-uuid",
    "board-member-3-uuid",
    "board-member-4-uuid",
    "board-member-5-uuid"
  ],
  "approval_type": "majority"
}
```

## Approval Flow

### Any Approval

```
Request ──► Approver 1
        ──► Approver 2   ──► Any approves ──► Approved
        ──► Approver 3
```

First approval completes the workflow.

### All Approval

```
Request ──► Approver 1 ──► Approved
        ──► Approver 2 ──► Approved  ──► All approved ──► Approved
        ──► Approver 3 ──► Approved
```

All must approve. One rejection = rejected.

### Majority Approval

```
Request ──► Approver 1 ──► Approved
        ──► Approver 2 ──► Approved   ──► 3/5 approved ──► Approved
        ──► Approver 3 ──► Approved
        ──► Approver 4 ──► Pending
        ──► Approver 5 ──► Rejected
```

More than half must approve.

## Use Cases

1. **Purchase approval** - Orders over threshold need manager approval
2. **Discount approval** - Large discounts need sales director approval
3. **User provisioning** - New user creation needs HR approval
4. **Document approval** - Documents need stakeholder sign-off

## Combining with Other Actions

Typically used with `create_alert` to notify approvers:

```json
// Rule with two actions:

// Action 1: Create alert
{
  "action_type": "create_alert",
  "config": {
    "alert_type": "approval_required",
    "title": "Approval Needed: Order {{number}}",
    "message": "Order {{number}} totaling {{total | currency:USD}} requires your approval.",
    "recipients": {
      "users": ["manager-uuid"]
    }
  },
  "execution_order": 1
}

// Action 2: Seek approval
{
  "action_type": "seek_approval",
  "config": {
    "approvers": ["manager-uuid"],
    "approval_type": "any"
  },
  "execution_order": 1
}
```

## Related Documentation

- [create_alert](create-alert.md) - Often used together to notify approvers
- [Rules](../configuration/rules.md) - How to create automation rules
- [Architecture](../architecture.md) - How action handlers are executed

## Key Files

| File | Purpose |
|------|---------|
| `business/sdk/workflow/workflowactions/approval/seek.go` | Handler implementation |
