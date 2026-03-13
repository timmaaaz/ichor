# Phase 4: Entity Field Schema Discovery API

## Objective

Eliminate the trigger configuration semantic leakiness by exposing a field schema endpoint.
When users configure `field_conditions` on `on_update` triggers, they currently must know the
exact internal string values for status enums. This phase adds an API that makes those values
discoverable.

## The Problem (Concrete Example)

User wants: "When a putaway task is completed, notify the warehouse manager."

What they must currently figure out:
1. Entity: `inventory.put_away_tasks` ✓ (discoverable via GET /v1/workflow/entities)
2. Trigger type: `on_update` ✓ (discoverable via GET /v1/workflow/trigger-types)
3. Field condition: `status` field, operator `changed_to`, value `"completed"` ✗ (not discoverable — must know the string)

After this phase:
```
GET /v1/workflow/entities/inventory.put_away_tasks/fields

{
  "entity": "inventory.put_away_tasks",
  "fields": [
    {
      "name": "status",
      "type": "enum",
      "values": ["pending", "in_progress", "completed", "cancelled"],
      "description": "Current lifecycle state of the putaway task"
    },
    ...
  ]
}
```

## Hardcoded Enum Registry

These 7 enums must be registered. The source of truth is Go constants already defined
in each bus package:

| Entity (API name) | Field | Source Constants | Values |
|-------------------|-------|-----------------|--------|
| `inventory.put_away_tasks` | `status` | `putawaytaskbus.Statuses` | pending, in_progress, completed, cancelled |
| `inventory.inventory_adjustments` | `approval_status` | inline in bus | pending, approved, rejected |
| `inventory.lot_trackings` | `quality_status` | inline in bus | good, on_hold, quarantined, released, expired |
| `workflow.alerts` | `status` | `alertbus.Status*` constants | active, acknowledged, dismissed, resolved |
| `workflow.alerts` | `severity` | `alertbus.Severity*` constants | low, medium, high, critical |
| `workflow.approval_requests` | `status` | `approvalrequestbus.Status*` constants | pending, approved, rejected, timed_out, expired |
| `workflow.approval_requests` | `approval_type` | `approvalrequestbus.ApprovalType*` constants | any, all, majority |

## Implementation Design

### Option A: Static Registry (Recommended)

Define a Go map in the workflow SDK or rule API package that maps entity names to their known
enum fields. This avoids runtime DB introspection and is always correct.

```go
// business/sdk/workflow/fieldschema/registry.go
package fieldschema

type FieldSchema struct {
    Name        string   `json:"name"`
    Type        string   `json:"type"`   // "enum", "string", "int", "bool", "uuid", "timestamp"
    Values      []string `json:"values,omitempty"`
    Description string   `json:"description,omitempty"`
}

type EntitySchema struct {
    Entity string        `json:"entity"`
    Fields []FieldSchema `json:"fields"`
}

// KnownEnumFields maps DB entity names to their known enum field schemas.
// Add new entries here as new status fields are added to the codebase.
var KnownEnumFields = map[string][]FieldSchema{
    "inventory.put_away_tasks": {
        {Name: "status", Type: "enum", Values: []string{"pending", "in_progress", "completed", "cancelled"}, Description: "Lifecycle state of the putaway task"},
    },
    "inventory.inventory_adjustments": {
        {Name: "approval_status", Type: "enum", Values: []string{"pending", "approved", "rejected"}, Description: "Approval state of the inventory adjustment"},
    },
    "inventory.lot_trackings": {
        {Name: "quality_status", Type: "enum", Values: []string{"good", "on_hold", "quarantined", "released", "expired"}, Description: "Quality control state of the lot"},
    },
    "workflow.alerts": {
        {Name: "status", Type: "enum", Values: []string{"active", "acknowledged", "dismissed", "resolved"}, Description: "Alert lifecycle state"},
        {Name: "severity", Type: "enum", Values: []string{"low", "medium", "high", "critical"}, Description: "Alert severity level"},
    },
    "workflow.approval_requests": {
        {Name: "status", Type: "enum", Values: []string{"pending", "approved", "rejected", "timed_out", "expired"}, Description: "Approval request resolution state"},
        {Name: "approval_type", Type: "enum", Values: []string{"any", "all", "majority"}, Description: "Required approval quorum type"},
    },
}
```

### Option B: DB Introspection

Query `information_schema.check_constraints` and `pg_enum` to dynamically extract enum values.
This auto-discovers new enums but requires complex SQL parsing of CHECK constraints.

**Recommendation**: Use Option A (static registry). It's explicit, testable, and immune to DB
schema quirks. The set of business-significant enums is small and stable. Add a `// NOTE: update
fieldschema/registry.go when adding new status enums` comment to each status.go file.

## API Endpoint

### Route

```
GET /v1/workflow/entities/{entity_name}/fields
```

Where `entity_name` is URL-encoded (e.g., `inventory.put_away_tasks` → `inventory.put_away_tasks`
or using `%2E` for dot — check existing entity name encoding in GET /v1/workflow/entities).

### Handler

**File**: `api/domain/http/workflow/ruleapi/fields.go` (or add to existing route file)

```go
func handleGetEntityFields(app *web.App, cfg Config) {
    app.HandleFunc(http.MethodGet, version, "/workflow/entities/{entity_name}/fields", ...)
}

// Response:
// 200 — entity found in registry, returns field schemas
// 200 with empty fields — entity exists in DB catalog but has no registered enum fields (not 404, entities without enums are valid)
// 404 — entity not found in DB catalog at all
```

### Response Shape

```json
{
  "entity": "inventory.put_away_tasks",
  "fields": [
    {
      "name": "status",
      "type": "enum",
      "values": ["pending", "in_progress", "completed", "cancelled"],
      "description": "Lifecycle state of the putaway task"
    }
  ]
}
```

## Route Registration

**File**: `api/domain/http/workflow/ruleapi/route.go`

Add the new route to the existing route registration alongside:
- `GET /workflow/trigger-types`
- `GET /workflow/entities`
- `GET /workflow/action-types`

## Tests

**File**: `api/cmd/services/ichor/tests/workflow/ruleapi/fields_test.go` (or existing test file)

Test cases:
1. GET `inventory.put_away_tasks/fields` → 200, returns status enum with 4 values
2. GET `workflow.approval_requests/fields` → 200, returns status + approval_type
3. GET `inventory.inventory_items/fields` → 200, empty fields array (entity exists, no registered enums)
4. GET `nonexistent.table/fields` → 404

## Documentation / Developer Guidance

Add a comment in each `status.go` (or equivalent) file that defines a hardcoded enum:

```go
// NOTE: These status values are registered in business/sdk/workflow/fieldschema/registry.go
// for workflow trigger discovery. Update both places when adding new status values.
```

This creates a maintenance link without coupling the bus package to the workflow SDK.

## Verification

```bash
go build ./business/sdk/workflow/fieldschema/...
go build ./api/domain/http/workflow/ruleapi/...
go build ./api/cmd/services/ichor/...
go test ./api/cmd/services/ichor/tests/workflow/ruleapi/...
```

Manually verify:
```bash
make token
curl -H "Authorization: Bearer $TOKEN" http://localhost:3000/v1/workflow/entities/inventory.put_away_tasks/fields
```

## Definition of Done

- [ ] `business/sdk/workflow/fieldschema/registry.go` created with 7 enums
- [ ] `GET /v1/workflow/entities/{entity}/fields` endpoint implemented
- [ ] Route registered alongside existing workflow discovery routes
- [ ] Integration tests for the endpoint (known enum entity, no-enum entity, 404)
- [ ] `// NOTE: update fieldschema/registry.go` comment added to each registered enum's source file
- [ ] `go build` passes on all affected packages
