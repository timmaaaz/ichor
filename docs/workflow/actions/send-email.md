# send_email Action

Sends email notifications to specified recipients.

## Configuration Schema

```json
{
  "recipients": ["email@example.com"],
  "subject": "string (required)",
  "body": "string (supports templates)",
  "simulate_failure": false,
  "failure_message": "string"
}
```

**Source**: `business/sdk/workflow/workflowactions/communication/email.go:32-38`

## Fields

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `recipients` | []string | **Yes** | - | Email addresses |
| `subject` | string | **Yes** | - | Email subject line |
| `body` | string | No | - | Email body (template supported) |
| `simulate_failure` | bool | No | `false` | Testing flag to simulate failure |
| `failure_message` | string | No | - | Error message for simulated failure |

## Validation Rules

1. `recipients` is required and must not be empty (unless `simulate_failure`)
2. `subject` is required (unless `simulate_failure`)
3. All recipient addresses should be valid email format

**Source**: `business/sdk/workflow/workflowactions/communication/email.go:32-58`

## Example Configurations

### Basic Email

```json
{
  "recipients": ["customer@example.com"],
  "subject": "Your Order Confirmation",
  "body": "Thank you for your order!"
}
```

### Email with Templates

```json
{
  "recipients": ["{{customer_email}}"],
  "subject": "Order {{number}} - Confirmation",
  "body": "Dear {{customer_name}},\n\nYour order {{number}} has been received.\n\nTotal: {{total | currency:USD}}\n\nThank you for your business!"
}
```

### Multi-Recipient Email

```json
{
  "recipients": [
    "sales@company.com",
    "manager@company.com",
    "{{customer_email}}"
  ],
  "subject": "New High-Value Order: {{number}}",
  "body": "A new order exceeding $10,000 has been placed.\n\nOrder: {{number}}\nCustomer: {{customer_name}}\nTotal: {{total | currency:USD}}"
}
```

### Test Configuration (Simulated Failure)

```json
{
  "recipients": ["test@example.com"],
  "subject": "Test",
  "body": "Test",
  "simulate_failure": true,
  "failure_message": "SMTP connection failed"
}
```

## Template Support

Both `subject` and `body` support template variables:

```json
{
  "subject": "{{entity_name | capitalize}} Update: {{number}}",
  "body": "Status changed from {{old_status}} to {{new_status}}\n\nUpdated by: {{user_id}}\nTimestamp: {{timestamp | formatDate:datetime}}"
}
```

See [Templates](../configuration/templates.md) for all available variables and filters.

## Execution

1. Parse and validate configuration
2. Process templates in subject and body
3. Send email via configured SMTP provider
4. Return success/failure result

## Email Provider

The email handler integrates with the configured email service. In testing, emails may be logged rather than sent.

## Error Handling

- Invalid email addresses cause validation errors
- SMTP failures are recorded in action result
- Timeouts are configurable
- Retries may be configured at the queue level

## Use Cases

1. **Order confirmations** - Send to customer on order creation
2. **Status updates** - Notify when order ships
3. **Internal notifications** - Alert staff to important events
4. **Approval requests** - Request approval via email

## Related Documentation

- [Templates](../configuration/templates.md) - Template variables and filters
- [Rules](../configuration/rules.md) - How to create automation rules
- [send_notification](send-notification.md) - Multi-channel notifications

## Key Files

| File | Purpose |
|------|---------|
| `business/sdk/workflow/workflowactions/communication/email.go` | Handler implementation |
