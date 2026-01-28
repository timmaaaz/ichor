# UpdateFieldHandler Overview

## Purpose

A workflow action handler that updates database fields dynamically based on configuration. Part of a larger automation/workflow system that can modify entity fields in response to events.

## Core Functionality

### What It Does

- Updates specific fields in database tables based on configurable rules
- Supports conditional updates (WHERE clauses)
- Handles foreign key resolution and creation
- Processes template variables for dynamic values
- Validates against a whitelist of allowed tables

### Key Components

#### Configuration Structure (`UpdateFieldConfig`)

```json
{
  "target_entity": "orders",           // Table to update
  "target_field": "status",            // Column to update
  "new_value": "{{status_value}}",     // New value (supports templates)
  "field_type": "foreign_key",         // Optional: specify special handling
  "conditions": [...],                 // WHERE conditions
  "foreign_key_config": {...}          // FK resolution settings
}
```

#### Foreign Key Resolution

- Automatically looks up foreign key IDs from display values
- Can create missing referenced records if configured
- Validates existing UUIDs against reference tables

#### Template Processing

- Replaces variables like `{{entity_id}}`, `{{user_id}}`, `{{old_fieldname}}`, `{{new_fieldname}}`
- Pulls values from execution context (event data, field changes, metadata)

## Execution Flow

1. **Validation Phase**

   - Validates configuration structure
   - Checks table names against whitelist
   - Validates operators and conditions

2. **Template Processing**

   - Processes template variables in the new value
   - Builds context from execution metadata

3. **Foreign Key Resolution** (if applicable)

   - Looks up ID from display value
   - Creates new record if allowed and not found

4. **Database Update**

   - Builds SQL UPDATE query with conditions
   - Executes update and returns affected row count

5. **Result Generation**
   - Returns execution metadata including:
     - Status (success/failed)
     - Records affected
     - Execution time
     - Any warnings or errors

## Security Features

- **Table Whitelist**: Only allows updates to predefined tables
- **Operator Validation**: Restricts SQL operators to safe set
- **Parameterized Queries**: Uses named parameters to prevent SQL injection
- **Transaction Support**: Can participate in larger database transactions

## Common Use Cases

- Update order status when conditions are met
- Set foreign key relationships using lookup values
- Bulk update records matching specific criteria
- Trigger cascading field updates in workflow automation

## Error Handling

- Validates all inputs before execution
- Returns detailed error messages in result
- Logs operations for debugging
- Handles missing foreign keys gracefully (create or fail)
