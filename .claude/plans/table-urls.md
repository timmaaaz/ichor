# Plan: Fix Link Config URL and Add Label Templating Support

## Problem Summary

In the orders table configuration, the `customer_id_fk` link config has two issues:

1. **URL Template Issue**: Uses `{customers.id}` but the column is aliased to `customer_id_fk` in the result set
2. **Label Templating**: The `Label` field doesn't support template variables - `{customers.name}` is treated as a literal string

## Root Cause Analysis

### URL Issue
The columns from joined tables are aliased (line 477-478 of tables.go):
```go
{Name: "id", Alias: "customer_id_fk", TableColumn: "customers.id"},
{Name: "name", Alias: "customer_name", TableColumn: "customers.name"},
```

So the template must use the **alias** (`customer_id_fk`), not the original table.column (`customers.id`).

### Label Issue
`LinkConfig.Label` is just a plain string field with no template processing. The validation only checks it's non-empty.

---

## Implementation Plan

### Part 1: Fix the URL Template (Quick Fix)

**File**: [tables.go](business/sdk/dbtest/seedmodels/tables.go#L615-L624)

Change from:
```go
"customer_id_fk": {
    Name:   "customer_id_fk",
    Header: "Customer",
    Width:  100,
    Type:   "uuid",
    Link: &tablebuilder.LinkConfig{
        URL:   "/sales/customers/{customers.id}",
        Label: "{customers.name}",
    },
},
```

To:
```go
"customer_id_fk": {
    Name:   "customer_id_fk",
    Header: "Customer",
    Width:  100,
    Type:   "uuid",
    Link: &tablebuilder.LinkConfig{
        URL:   "/sales/customers/{customer_id_fk}",
        LabelColumn: "customer_name",  // Will be used as column reference after Part 2
    },
},
```

### Part 2: Add Label Templating Support

The label templating needs to be handled by the **frontend** since it has access to row data at render time. The backend just needs to indicate that the label should be treated as a column reference.

#### Approach: Add `LabelColumn` Field

Add a new field to `LinkConfig` to indicate the label should come from a column:

**File**: [model.go](business/sdk/tablebuilder/model.go#L216-L220)

```go
// LinkConfig defines link configuration
type LinkConfig struct {
    URL         string `json:"url"`
    Label       string `json:"label"`                   // Static label text
    LabelColumn string `json:"label_column,omitempty"`  // Column name to use as dynamic label
}
```

**Behavior**:
- If `LabelColumn` is set, frontend uses that column's value from the row as the link text
- If only `Label` is set, use the static text
- `LabelColumn` takes precedence over `Label` if both are set

#### Files to Modify

1. **[model.go](business/sdk/tablebuilder/model.go#L216-L220)** - Add `LabelColumn` field to `LinkConfig`
2. **[validation.go](business/sdk/tablebuilder/validation.go#L514-L523)** - Update validation to allow either `Label` or `LabelColumn`
3. **[tables.go](business/sdk/dbtest/seedmodels/tables.go#L615-L624)** - Update test seed data to use new field

### Part 3: Update Validation

**File**: [validation.go](business/sdk/tablebuilder/validation.go#L514-L523)

Change from:
```go
func (c *Config) validateLinkConfig(result *ValidationResult, l *LinkConfig, prefix string) {
    if l.URL == "" {
        result.AddError(prefix+".url", "url is required", "REQUIRED")
    }

    if l.Label == "" {
        result.AddError(prefix+".label", "label is required", "REQUIRED")
    }
}
```

To:
```go
func (c *Config) validateLinkConfig(result *ValidationResult, l *LinkConfig, prefix string) {
    if l.URL == "" {
        result.AddError(prefix+".url", "url is required", "REQUIRED")
    }

    // Either Label or LabelColumn must be provided
    if l.Label == "" && l.LabelColumn == "" {
        result.AddError(prefix+".label", "either label or label_column is required", "REQUIRED")
    }
}
```

---

## Final Configuration Example

After implementation, the config would look like:
```go
"customer_id_fk": {
    Name:   "customer_id_fk",
    Header: "Customer",
    Width:  100,
    Type:   "uuid",
    Link: &tablebuilder.LinkConfig{
        URL:         "/sales/customers/{customer_id_fk}",
        LabelColumn: "customer_name",  // Dynamic - uses row value
    },
},
```

Or with a static fallback:
```go
Link: &tablebuilder.LinkConfig{
    URL:         "/sales/customers/{customer_id_fk}",
    Label:       "View Customer",  // Static fallback
    LabelColumn: "customer_name",  // Dynamic - takes precedence
},
```

---

## Verification

1. Run `make test` to ensure all existing tests pass
2. Check that the orders table config is valid: `go run ./api/cmd/tooling/admin table-config validate`
3. Verify JSON output includes the new `label_column` field when set

---

## Frontend Note

The frontend will need to be updated to:
1. Check if `link.label_column` exists in the column metadata
2. If so, use `row[link.label_column]` as the link text
3. Otherwise, fall back to `link.label` as the static text

This is outside the scope of this backend change but should be communicated to the frontend team.
