# FormData App Package

## Overview

The `formdataapp` package provides the application layer for handling dynamic multi-entity form data operations. It coordinates transactions across multiple domain entities while preserving business logic and validation in their respective domain packages.

## Table of Contents

- [Architecture](#architecture)
- [Key Features](#key-features)
- [Request Format](#request-format)
- [Foreign Key Resolution](#foreign-key-resolution)
- [Transaction Handling](#transaction-handling)
- [Error Handling](#error-handling)
- [Usage Examples](#usage-examples)
- [Best Practices](#best-practices)

## Architecture

### Layer Responsibilities

```
┌────────────────────────────────────────────────────┐
│              Frontend/API Layer                    │
│  - Receives form submissions                       │
│  - Routes to /formdata/{form_id}/upsert           │
└──────────────────┬─────────────────────────────────┘
                   │
┌──────────────────▼─────────────────────────────────┐
│             FormData App Layer                     │
│  - Loads form configuration                        │
│  - Validates request structure                     │
│  - Plans execution order                           │
│  - Manages transactions                            │
│  - Resolves template variables                     │
└──────────────────┬─────────────────────────────────┘
                   │
┌──────────────────▼─────────────────────────────────┐
│               Registry Layer                       │
│  - Maps entities to operations                     │
│  - Provides decode/validate functions              │
│  - Provides CRUD functions                         │
└──────────────────┬─────────────────────────────────┘
                   │
┌──────────────────▼─────────────────────────────────┐
│            Domain App Layers                       │
│  (userapp, assetapp, productapp, etc.)            │
│  - Validates business rules                        │
│  - Converts between app/bus models                 │
└──────────────────┬─────────────────────────────────┘
                   │
┌──────────────────▼─────────────────────────────────┐
│           Domain Business Layers                   │
│  (userbus, assetbus, productbus, etc.)            │
│  - Enforces domain logic                           │
│  - Persists to database                            │
└────────────────────────────────────────────────────┘
```

### Data Flow

```
1. Frontend submits form with operations + data
       ↓
2. FormDataApp.UpsertFormData() receives request
       ↓
3. Load form config from config.forms + config.form_fields
       ↓
4. Validate request against form definition
       ↓
5. Build execution plan (sort by order)
       ↓
6. Begin transaction
       ↓
7. For each operation:
   a. Process template variables ({{users.id}})
   b. Look up entity in registry
   c. Decode & validate using app layer model
   d. Execute create/update via registry function
   e. Store result for template context
       ↓
8. Commit transaction
       ↓
9. Return all results to frontend
```

## Key Features

### 1. Multi-Entity Transactions

Execute operations across multiple entities atomically:

```json
{
  "operations": {
    "users": {"operation": "create", "order": 1},
    "addresses": {"operation": "create", "order": 2},
    "phones": {"operation": "create", "order": 3}
  },
  "data": {
    "users": {...},
    "addresses": {...},
    "phones": {...}
  }
}
```

All operations succeed or all fail (rollback).

### 2. Foreign Key Resolution

Automatically resolve foreign keys using template variables:

```json
{
  "data": {
    "users": {
      "first_name": "John",
      "last_name": "Doe"
    },
    "addresses": {
      "user_id": "{{users.id}}",  // Resolved after user creation
      "street": "123 Main St"
    }
  }
}
```

### 3. Execution Ordering

Control execution order explicitly to handle dependencies:

```json
{
  "operations": {
    "users": {"operation": "create", "order": 1},
    "addresses": {"operation": "create", "order": 2}
  }
}
```

Order 1 executes first, then order 2 can reference its results.

### 4. Domain Validation Preserved

All validation remains in domain app layers:

```go
// formapp/model.go - Existing validation is used
func (app NewUser) Validate() error {
    // Email format check
    // Password strength check
    // Business rules
}

// formdata app calls this via registry
model, err := reg.DecodeNew(data)  // Calls Validate() internally
```

## Request Format

### Complete Request Example

```json
{
  "operations": {
    "users": {
      "operation": "create",
      "order": 1
    },
    "addresses": {
      "operation": "create",
      "order": 2
    }
  },
  "data": {
    "users": {
      "first_name": "John",
      "last_name": "Doe",
      "email": "john@example.com",
      "password": "SecurePass123!"
    },
    "addresses": {
      "user_id": "{{users.id}}",
      "street": "123 Main St",
      "city": "Portland",
      "state": "OR",
      "postal_code": "97201"
    }
  }
}
```

### Operations Structure

Each key in `operations` is an entity name (matching `workflow.entities.name`):

```typescript
{
  "entity_name": {
    "operation": "create" | "update",  // Required
    "order": number                    // Required, >= 1
  }
}
```

**Fields:**
- `operation`: Must be "create" or "update"
- `order`: Execution order (1-based), determines sequence

**Rules:**
- Entity names must match registered entities in registry
- Order values should be sequential but gaps are allowed
- Lower order values execute first

### Data Structure

Each key in `data` matches an entity name from `operations`:

```typescript
{
  "entity_name": {
    // Entity-specific fields
    // For create: fields from NewEntity model
    // For update: fields from UpdateEntity model + id field
  }
}
```

**For CREATE operations:**
```json
{
  "users": {
    "first_name": "John",
    "last_name": "Doe",
    "email": "john@example.com"
  }
}
```

**For UPDATE operations:**
```json
{
  "users": {
    "id": "existing-uuid-here",
    "first_name": "Jane"
  }
}
```

The `id` field is required for updates and is used to identify the record.

## Foreign Key Resolution

### Template Syntax

Use `{{entity_name.field}}` syntax to reference results from previous operations:

```json
{
  "operations": {
    "users": {"operation": "create", "order": 1},
    "addresses": {"operation": "create", "order": 2}
  },
  "data": {
    "users": {
      "first_name": "John",
      "email": "john@example.com"
    },
    "addresses": {
      "user_id": "{{users.id}}",         // References created user's ID
      "street": "{{users.street}}",       // Can reference any field
      "description": "Home of {{users.first_name}}"  // Even non-UUID fields
    }
  }
}
```

### How It Works

1. **Execute first operation** (users create)
   ```
   Result: {id: "uuid-123", first_name: "John", ...}
   ```

2. **Build template context**
   ```go
   context["users"] = {id: "uuid-123", first_name: "John", ...}
   ```

3. **Process second operation's data**
   ```
   Input:  {"user_id": "{{users.id}}"}
   Output: {"user_id": "uuid-123"}
   ```

4. **Execute second operation** with resolved data

### Nested Field Access

Access nested fields using dot notation:

```json
{
  "product_id": "{{products.id}}",
  "brand_name": "{{products.brand.name}}",
  "category": "{{products.category.display_name}}"
}
```

### Supported Template Features

Powered by `business/sdk/workflow/template.go`:

**Filters:**
```json
{
  "full_name": "{{users.first_name | uppercase}}",
  "formatted_date": "{{users.created_date | formatDate:short}}"
}
```

**Default values:**
```json
{
  "middle_name": "{{users.middle_name | default:N/A}}"
}
```

See `business/sdk/workflow/template.go` README for full template syntax.

## Transaction Handling

### All-or-Nothing Semantics

Transactions are atomic - all operations succeed or all fail:

```
BEGIN TRANSACTION
  1. CREATE users → Success ✓
  2. CREATE addresses → Success ✓
  3. CREATE phones → FAIL ✗
ROLLBACK

Result: Nothing persisted, users and addresses rolled back
```

### Transaction Isolation

Uses `READ COMMITTED` isolation level:

```go
tx, err := db.BeginTxx(ctx, &sql.TxOptions{
    Isolation: sql.LevelReadCommitted,
})
```

This prevents dirty reads while allowing good concurrency.

### When Transactions Fail

Common failure scenarios:

**Validation Error:**
```
Operation: CREATE users
Error: "email format invalid"
Action: Immediate rollback, no database writes
```

**Foreign Key Violation:**
```
Operation: CREATE addresses
Error: "foreign key constraint violation on user_id"
Action: Rollback entire transaction
```

**Unique Constraint:**
```
Operation: CREATE users
Error: "duplicate key value violates unique constraint: email"
Action: Rollback entire transaction
```

### Transaction Ordering

Always execute parent entities before children:

```json
{
  "operations": {
    "users": {"operation": "create", "order": 1},      // Parent
    "addresses": {"operation": "create", "order": 2},   // Child
    "phones": {"operation": "create", "order": 3}       // Child
  }
}
```

**BAD ordering** (will fail):
```json
{
  "operations": {
    "addresses": {"operation": "create", "order": 1},  // ✗ Child first
    "users": {"operation": "create", "order": 2}       // ✗ Parent second
  }
}
```

This fails because `addresses` needs `users.id` which doesn't exist yet.

## Error Handling

### Error Response Format

All errors follow the standard app error format:

```json
{
  "error": {
    "code": "INVALID_ARGUMENT",
    "message": "execute users: decode: email format invalid",
    "details": {
      "field": "email",
      "value": "not-an-email"
    }
  }
}
```

### Error Types

**InvalidArgument:**
```json
{
  "code": "INVALID_ARGUMENT",
  "message": "entity users in operations but missing from data"
}
```

Causes:
- Operations/data mismatch
- Invalid operation type
- Missing required fields
- Validation failures

**NotFound:**
```json
{
  "code": "NOT_FOUND",
  "message": "form not found"
}
```

Causes:
- Invalid form_id
- Form deleted
- No permission to access form

**Internal:**
```json
{
  "code": "INTERNAL",
  "message": "begin transaction: connection refused"
}
```

Causes:
- Database connection issues
- Transaction failures
- System errors

### Error Context

Errors include the operation context:

```
"execute users: decode: email format invalid"
 ^^^^^^ ^^^^^ ^^^^^^ ^^^^^^^^^^^^^^^^^^^^^^
   |     |      |            |
   |     |      |            └─ Root cause
   |     |      └─ Which function failed
   |     └─ Which entity
   └─ Which phase
```

This helps debugging multi-step operations.

### Validation Errors

Validation uses existing app layer validation:

```go
// In userapp/model.go
func (app NewUser) Validate() error {
    if !validEmail(app.Email) {
        return errs.New(errs.InvalidArgument, "invalid email format")
    }
    // ... more validation
}
```

These errors are passed through unchanged to the response.

## Usage Examples

### Example 1: Single Entity Create

**Request:**
```json
POST /v1/formdata/{form_id}/upsert

{
  "operations": {
    "users": {
      "operation": "create",
      "order": 1
    }
  },
  "data": {
    "users": {
      "first_name": "John",
      "last_name": "Doe",
      "email": "john@example.com",
      "password": "SecurePass123!"
    }
  }
}
```

**Response:**
```json
{
  "success": true,
  "results": {
    "users": {
      "id": "uuid-123",
      "first_name": "John",
      "last_name": "Doe",
      "email": "john@example.com",
      "created_date": "2025-01-15T10:30:00Z"
    }
  }
}
```

### Example 2: Multi-Entity with Foreign Keys

**Request:**
```json
{
  "operations": {
    "users": {"operation": "create", "order": 1},
    "addresses": {"operation": "create", "order": 2}
  },
  "data": {
    "users": {
      "first_name": "Jane",
      "last_name": "Smith",
      "email": "jane@example.com"
    },
    "addresses": {
      "user_id": "{{users.id}}",
      "street": "456 Oak Ave",
      "city": "Seattle",
      "state": "WA",
      "postal_code": "98101"
    }
  }
}
```

**Response:**
```json
{
  "success": true,
  "results": {
    "users": {
      "id": "uuid-456",
      "first_name": "Jane",
      "last_name": "Smith",
      "email": "jane@example.com"
    },
    "addresses": {
      "id": "uuid-789",
      "user_id": "uuid-456",
      "street": "456 Oak Ave",
      "city": "Seattle"
    }
  }
}
```

### Example 3: Update Existing Record

**Request:**
```json
{
  "operations": {
    "users": {"operation": "update", "order": 1}
  },
  "data": {
    "users": {
      "id": "uuid-456",
      "first_name": "Janet",
      "email": "janet@example.com"
    }
  }
}
```

**Response:**
```json
{
  "success": true,
  "results": {
    "users": {
      "id": "uuid-456",
      "first_name": "Janet",
      "last_name": "Smith",
      "email": "janet@example.com",
      "updated_date": "2025-01-15T11:45:00Z"
    }
  }
}
```

### Example 4: Mixed Create and Update

**Request:**
```json
{
  "operations": {
    "users": {"operation": "update", "order": 1},
    "addresses": {"operation": "create", "order": 2}
  },
  "data": {
    "users": {
      "id": "uuid-456",
      "phone": "+1-555-0123"
    },
    "addresses": {
      "user_id": "{{users.id}}",
      "street": "789 Pine St",
      "city": "Portland"
    }
  }
}
```

## Best Practices

### 1. Order Dependencies Explicitly

Always set `order` to reflect dependencies:

```json
{
  "operations": {
    "categories": {"operation": "create", "order": 1},
    "products": {"operation": "create", "order": 2},
    "inventory": {"operation": "create", "order": 3}
  }
}
```

Even if operations *could* run in parallel, explicit ordering makes intent clear.

### 2. Use Template Variables for All FKs

Don't hardcode IDs, use templates:

```json
// ✅ GOOD: Template variable
{
  "addresses": {
    "user_id": "{{users.id}}"
  }
}

// ❌ BAD: Hardcoded (fragile, error-prone)
{
  "addresses": {
    "user_id": "uuid-from-previous-request"
  }
}
```

### 3. Validate on Frontend First

Catch simple errors before API call:

```typescript
// Frontend validation
if (!isValidEmail(email)) {
  showError("Invalid email format");
  return;
}

// Then call API
await fetch('/v1/formdata/{id}/upsert', ...);
```

This gives better UX than waiting for server validation.

### 4. Handle Partial Failures Gracefully

Since transactions rollback entirely, inform users clearly:

```typescript
try {
  const result = await upsertFormData(data);
} catch (error) {
  if (error.code === 'INVALID_ARGUMENT') {
    showFieldError(error.details.field, error.message);
  } else {
    showGenericError("Failed to save. Please try again.");
  }
}
```

### 5. Use Appropriate Form Configurations

Define forms that match your business processes:

```
user-creation-form:
  - Entity: users
  - Fields: first_name, last_name, email, password

user-with-address-form:
  - Entity: users (order 1)
  - Entity: addresses (order 2)
  - Fields: user fields + address fields

order-entry-form:
  - Entity: orders (order 1)
  - Entity: order_line_items (order 2)
  - Fields: order info + line items
```

### 6. Test Transaction Rollback

Always test failure scenarios:

```go
func TestUpsertFormData_RollbackOnFailure(t *testing.T) {
    // Setup: Create invalid data for second entity
    req := FormDataRequest{
        Operations: map[string]OperationMeta{
            "users": {Operation: "create", Order: 1},
            "addresses": {Operation: "create", Order: 2},
        },
        Data: map[string]json.RawMessage{
            "users": validUserData,
            "addresses": invalidAddressData,  // Will fail
        },
    }

    // Execute
    _, err := app.UpsertFormData(ctx, formID, req)
    assert.Error(t, err)

    // Verify: User was not persisted (rollback worked)
    _, err = userBus.QueryByEmail(ctx, "test@example.com")
    assert.Error(t, err)  // Should not exist
}
```

### 7. Log Execution Plans

Log the execution plan for debugging:

```go
log.Info("Executing form data plan",
    "form_id", formID,
    "entities", len(plan),
    "operations", formatOperations(plan))
```

This helps trace issues in production.

## Integration with Form Configuration

The service uses form configurations from `config.forms` and `config.form_fields`:

### Form Field Structure

```sql
SELECT
    ff.id,
    ff.form_id,
    ff.entity_id,      -- References workflow.entities
    ff.name,           -- Column name
    ff.label,          -- Display label
    ff.field_type,     -- UI field type
    ff.config          -- JSONB with execution_order, parent_entity_id, etc.
FROM config.form_fields ff
WHERE ff.form_id = $1
ORDER BY ff.field_order;
```

### Form Field Config JSONB

```json
{
  "parent_entity_id": "uuid-of-parent",
  "foreign_key_column": "user_id",
  "execution_order": 2,
  "validation_rules": {},
  "ui_config": {}
}
```

This configuration drives the frontend form generation and can inform
validation/ordering on the backend (future enhancement).

## Performance Considerations

### Transaction Duration

Keep transactions short:
- ✅ Good: 2-3 entity operations
- ⚠️  Caution: 5-10 entity operations
- ❌ Avoid: >10 entity operations (consider splitting)

Long transactions hold locks and increase contention.

### Template Processing Overhead

Template processing adds minimal overhead (~1ms per entity), but:

```json
// ✅ Efficient: Simple references
{
  "user_id": "{{users.id}}"
}

// ⚠️ Less efficient: Complex filters
{
  "description": "{{users.addresses | join:, | uppercase}}"
}
```

Use complex templates sparingly.

### Database Round Trips

Each operation = 1 DB round trip:

```
Operation 1 (users):      DB call
Operation 2 (addresses):  DB call
Operation 3 (phones):     DB call
Total: 3 DB calls in 1 transaction
```

This is acceptable for typical forms (<5 entities).

## Troubleshooting

### "entity X not registered"

**Cause:** Entity missing from registry

**Solution:** Check `build.go` for registration:
```go
registry.Register("users", ...)
registry.Register("addresses", ...)
```

### "template processing errors"

**Cause:** Invalid template syntax or missing variable

**Solution:** Check template references:
```json
{
  "user_id": "{{users.id}}"  // Correct
  "user_id": "{{user.id}}"   // Wrong (typo)
}
```

### "id required for update operations"

**Cause:** Missing `id` field in update data

**Solution:** Include ID in update payload:
```json
{
  "users": {
    "id": "uuid-here",  // Required for updates
    "first_name": "New Name"
  }
}
```

### Transaction deadlocks

**Cause:** Concurrent requests updating same records

**Solution:**
1. Use optimistic locking (version fields)
2. Retry with exponential backoff
3. Queue writes for hot records
