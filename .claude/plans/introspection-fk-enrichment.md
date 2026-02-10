# Introspection API: FK Column Enrichment

## Overview

Enhance the introspection `QueryColumns` endpoint to include foreign key metadata directly in the column response, eliminating the need for a separate relationships call.

**Purpose**: Enable frontend workflow condition builder to detect FK columns and render SmartCombobox for value selection instead of raw UUID input.

---

## Current State

**QueryColumns SQL** (`introspectionbus.go` lines 91-114):
```sql
SELECT
    c.column_name AS name,
    c.data_type AS data_type,
    c.is_nullable = 'YES' AS is_nullable,
    COALESCE(c.column_default, '') AS default_value,
    EXISTS(...) AS is_primary_key
FROM information_schema.columns c
WHERE c.table_schema = :schema AND c.table_name = :table
```

**Current Response**:
```json
{
  "name": "customer_id",
  "data_type": "uuid",
  "is_nullable": true,
  "is_primary_key": false,
  "default_value": ""
}
```

**Problem**: No FK information - frontend can't know this column references another table.

---

## Proposed Changes

### New Response Format

```json
{
  "name": "customer_id",
  "data_type": "uuid",
  "is_nullable": true,
  "is_primary_key": false,
  "default_value": "",
  "is_foreign_key": true,
  "referenced_schema": "sales",
  "referenced_table": "customers",
  "referenced_column": "id"
}
```

Non-FK columns will have:
- `is_foreign_key: false`
- `referenced_schema`, `referenced_table`, `referenced_column` omitted (null/omitempty)

---

## Files to Modify

### 1. Business Layer Model

**File**: `business/domain/introspectionbus/model.go`

**Current**:
```go
type Column struct {
    Name         string `db:"name"`
    DataType     string `db:"data_type"`
    IsNullable   bool   `db:"is_nullable"`
    IsPrimaryKey bool   `db:"is_primary_key"`
    DefaultValue string `db:"default_value"`
}
```

**Proposed**:
```go
type Column struct {
    Name         string `db:"name"`
    DataType     string `db:"data_type"`
    IsNullable   bool   `db:"is_nullable"`
    IsPrimaryKey bool   `db:"is_primary_key"`
    DefaultValue string `db:"default_value"`
    // Foreign key metadata (NULL if not a FK)
    IsForeignKey     bool    `db:"is_foreign_key"`
    ReferencedSchema *string `db:"referenced_schema"`
    ReferencedTable  *string `db:"referenced_table"`
    ReferencedColumn *string `db:"referenced_column"`
}
```

**Notes**:
- Use `*string` (pointer) for nullable fields - Go convention for optional DB values
- `db` tags must match SQL column aliases exactly

---

### 2. Business Layer Query

**File**: `business/domain/introspectionbus/introspectionbus.go`

**Current**: `QueryColumns` function uses `information_schema.columns`

**Proposed**: Rewrite using `pg_catalog` (idiomatic PostgreSQL, faster than `information_schema` views)

```go
// QueryColumns returns all columns for a given table, enriched with FK metadata.
func (b *Business) QueryColumns(ctx context.Context, schema, table string) ([]Column, error) {
    ctx, span := otel.AddSpan(ctx, "business.introspectionbus.querycolumns")
    defer span.End()

    const q = `
    SELECT
        a.attname AS name,
        pg_catalog.format_type(a.atttypid, a.atttypmod) AS data_type,
        NOT a.attnotnull AS is_nullable,
        COALESCE(pg_get_expr(d.adbin, d.adrelid), '') AS default_value,
        COALESCE(pk.is_pk, FALSE) AS is_primary_key,
        fk.conname IS NOT NULL AS is_foreign_key,
        fk_ns.nspname AS referenced_schema,
        fk_cl.relname AS referenced_table,
        fk_att.attname AS referenced_column
    FROM
        pg_catalog.pg_attribute a
    JOIN pg_catalog.pg_class c ON a.attrelid = c.oid
    JOIN pg_catalog.pg_namespace n ON c.relnamespace = n.oid
    -- Default values
    LEFT JOIN pg_catalog.pg_attrdef d ON a.attrelid = d.adrelid AND a.attnum = d.adnum
    -- Primary key detection
    LEFT JOIN (
        SELECT
            conrelid,
            unnest(conkey) AS attnum,
            TRUE AS is_pk
        FROM pg_catalog.pg_constraint
        WHERE contype = 'p'
    ) pk ON a.attrelid = pk.conrelid AND a.attnum = pk.attnum
    -- Foreign key detection with referenced table info
    LEFT JOIN LATERAL (
        SELECT
            con.conname,
            con.confrelid,
            col_idx.ord,
            (con.confkey)[col_idx.ord] AS ref_attnum
        FROM pg_catalog.pg_constraint con
        CROSS JOIN LATERAL unnest(con.conkey) WITH ORDINALITY AS col_idx(attnum, ord)
        WHERE con.contype = 'f'
          AND con.conrelid = a.attrelid
          AND col_idx.attnum = a.attnum
    ) fk ON TRUE
    LEFT JOIN pg_catalog.pg_class fk_cl ON fk.confrelid = fk_cl.oid
    LEFT JOIN pg_catalog.pg_namespace fk_ns ON fk_cl.relnamespace = fk_ns.oid
    LEFT JOIN pg_catalog.pg_attribute fk_att ON fk.confrelid = fk_att.attrelid AND fk.ref_attnum = fk_att.attnum
    WHERE
        n.nspname = :schema
        AND c.relname = :table
        AND a.attnum > 0           -- Exclude system columns
        AND NOT a.attisdropped     -- Exclude dropped columns
    ORDER BY
        a.attnum`

    data := struct {
        Schema string `db:"schema"`
        Table  string `db:"table"`
    }{
        Schema: schema,
        Table:  table,
    }

    var columns []Column
    if err := sqldb.NamedQuerySlice(ctx, b.log, b.db, q, data, &columns); err != nil {
        return nil, fmt.Errorf("query columns: %w", err)
    }

    return columns, nil
}
```

**Why `pg_catalog` instead of `information_schema`**:
- `information_schema` is a SQL standard abstraction layer built on top of `pg_catalog`
- `pg_catalog` is PostgreSQL's native system catalog - direct access, no view overhead
- Ichor is PostgreSQL-only, so portability to other databases is not a concern
- This is the idiomatic PostgreSQL approach

**SQL Explanation**:
- `pg_attribute` contains column metadata (name, type, nullability)
- `pg_constraint` with `contype = 'p'` finds primary keys, `contype = 'f'` finds foreign keys
- `LATERAL` join with `unnest` handles composite FKs correctly (maps source column to referenced column by position)
- `format_type()` gives human-readable type names (e.g., "character varying(255)" not just "varchar")
- `attnum > 0` excludes system columns (ctid, xmin, etc.)
- Non-FK columns get NULL for `referenced_*` fields, `is_foreign_key` becomes FALSE

---

### 3. Application Layer Model

**File**: `app/domain/introspectionapp/model.go`

**Current**:
```go
type Column struct {
    Name         string `json:"name"`
    DataType     string `json:"data_type"`
    IsNullable   bool   `json:"is_nullable"`
    IsPrimaryKey bool   `json:"is_primary_key"`
    DefaultValue string `json:"default_value"`
}
```

**Proposed**:
```go
type Column struct {
    Name         string  `json:"name"`
    DataType     string  `json:"data_type"`
    IsNullable   bool    `json:"is_nullable"`
    IsPrimaryKey bool    `json:"is_primary_key"`
    DefaultValue string  `json:"default_value"`
    // Foreign key metadata
    IsForeignKey     bool    `json:"is_foreign_key"`
    ReferencedSchema *string `json:"referenced_schema,omitempty"`
    ReferencedTable  *string `json:"referenced_table,omitempty"`
    ReferencedColumn *string `json:"referenced_column,omitempty"`
}
```

**Notes**:
- `omitempty` excludes null pointer fields from JSON (cleaner response for non-FK columns)
- Matches business layer types exactly

---

### 4. Application Layer Conversion

**File**: `app/domain/introspectionapp/model.go`

**Current** `ToAppColumn` function:
```go
func ToAppColumn(bus introspectionbus.Column) Column {
    return Column{
        Name:         bus.Name,
        DataType:     bus.DataType,
        IsNullable:   bus.IsNullable,
        IsPrimaryKey: bus.IsPrimaryKey,
        DefaultValue: bus.DefaultValue,
    }
}
```

**Proposed**:
```go
func ToAppColumn(bus introspectionbus.Column) Column {
    return Column{
        Name:             bus.Name,
        DataType:         bus.DataType,
        IsNullable:       bus.IsNullable,
        IsPrimaryKey:     bus.IsPrimaryKey,
        DefaultValue:     bus.DefaultValue,
        IsForeignKey:     bus.IsForeignKey,
        ReferencedSchema: bus.ReferencedSchema,
        ReferencedTable:  bus.ReferencedTable,
        ReferencedColumn: bus.ReferencedColumn,
    }
}
```

---

### 5. Tests

**File**: `api/cmd/services/ichor/tests/introspectionapi/introspection_test.go`

The existing `queryColumns200` test uses subset comparison and won't break. However, we should add explicit FK verification.

**Add new test case** for FK columns:

```go
func queryColumnsWithFK200(sd apitest.SeedData) []apitest.Table {
    // Test table with known FK: core.user_roles has user_id -> core.users.id
    return []apitest.Table{
        {
            Name:       "fk-columns",
            URL:        "/v1/introspection/tables/core/user_roles/columns",
            Method:     http.MethodGet,
            Token:      sd.Admins[0].Token,
            StatusCode: http.StatusOK,
            GotResp:    &introspectionapp.Columns{},
            ExpResp:    &introspectionapp.Columns{},
            CmpFunc: func(got, exp any) string {
                gotCols := got.(*introspectionapp.Columns)

                // Find user_id column and verify FK metadata
                for _, c := range *gotCols {
                    if c.Name == "user_id" {
                        if !c.IsForeignKey {
                            return "user_id should be marked as foreign key"
                        }
                        if c.ReferencedSchema == nil || *c.ReferencedSchema != "core" {
                            return "user_id should reference core schema"
                        }
                        if c.ReferencedTable == nil || *c.ReferencedTable != "users" {
                            return "user_id should reference users table"
                        }
                        if c.ReferencedColumn == nil || *c.ReferencedColumn != "id" {
                            return "user_id should reference id column"
                        }
                        return "" // Success
                    }
                }
                return "user_id column not found"
            },
        },
    }
}
```

**Update test runner** - add after the existing `queryColumns200` test run:
```go
test.Run(t, queryColumnsWithFK200(sd), "query-columns-fk-200")
```

---

## Files NOT Modified

These files use introspection but don't need changes:

| File | Reason |
|------|--------|
| `introspectionapi/introspectionapi.go` | Handler just passes through - no struct knowledge |
| `introspectionapi/routes.go` | Only defines routes, no data handling |
| `formfieldbus/validation_deep.go` | Uses columns for name lookup, not FK metadata |
| `tablebuilder/store.go` | Has separate FK detection logic, doesn't use Column struct |

---

## Backwards Compatibility

**API Contract**: Additive only
- New fields added to response
- Existing fields unchanged
- Non-FK columns: new fields are `false`/null (omitted with `omitempty`)

**Existing Clients**:
- Frontend `useEntityDiscovery.ts` already expects `is_foreign_key` and `foreign_table` fields
- Unknown fields in JSON are safely ignored by most clients

**Database**:
- Query uses `pg_catalog` system tables (always present in PostgreSQL)
- No new tables or migrations required

---

## Testing Strategy

### Manual Testing

```bash
# After deployment, verify with curl:

# 1. Test table without FKs (core.users primary table)
curl -H "Authorization: Bearer $TOKEN" \
  http://localhost:8080/v1/introspection/tables/core/users/columns | jq '.[] | select(.name == "id")'
# Expected: is_foreign_key: false, no referenced_* fields

# 2. Test table with FKs (core.user_roles junction table)
curl -H "Authorization: Bearer $TOKEN" \
  http://localhost:8080/v1/introspection/tables/core/user_roles/columns | jq '.[] | select(.name == "user_id")'
# Expected: is_foreign_key: true, referenced_schema: "core", referenced_table: "users"

# 3. Test sales table (orders -> customers)
curl -H "Authorization: Bearer $TOKEN" \
  http://localhost:8080/v1/introspection/tables/sales/orders/columns | jq '.[] | select(.is_foreign_key == true)'
# Expected: customer_id with FK metadata
```

### Automated Testing

```bash
# Run introspection tests
go test -v ./api/cmd/services/ichor/tests/introspectionapi/...
```

---

## Implementation Order

1. **Business layer model** (`business/domain/introspectionbus/model.go`) - Add 4 new fields to `Column` struct
2. **SQL query** (`business/domain/introspectionbus/introspectionbus.go`) - Replace `QueryColumns` with `pg_catalog` version
3. **App layer model** (`app/domain/introspectionapp/model.go`) - Add 4 new fields to `Column` struct
4. **Conversion function** (`app/domain/introspectionapp/model.go`) - Update `ToAppColumn` to map new fields
5. **Tests** (`api/cmd/services/ichor/tests/introspectionapi/introspection_test.go`) - Add FK verification test
6. **Run tests** - `go test -v ./api/cmd/services/ichor/tests/introspectionapi/...`
7. **Manual verification** - Test with curl against running server

---

## Rollback Plan

If issues arise:
1. Revert model changes (remove 4 fields from each Column struct)
2. Revert SQL query to original simple SELECT
3. Revert conversion function
4. Tests will still pass (subset comparison)

The API remains backwards compatible throughout - old responses are a subset of new responses.

---

## Performance Considerations

**Query Impact**:
- `pg_catalog` tables are PostgreSQL's native system catalogs with excellent indexing
- Faster than `information_schema` which is a view layer on top of `pg_catalog`
- LATERAL join for FK lookup is efficient - only executed for matching rows
- Acceptable: Columns are queried once per entity selection, cached by frontend

**Why `pg_catalog` over `information_schema`**:
- Direct access to system catalogs vs. abstraction layer overhead
- `information_schema` is for SQL standard portability - irrelevant for PostgreSQL-only project
- This is how PostgreSQL's own tools (psql `\d`, pg_dump) query metadata

**Alternative Considered**:
- Two separate queries (columns + relationships) merged in Go
- Rejected: More code, same DB work, extra serialization overhead

---

## Related Frontend Changes

After this backend change is deployed, the frontend can be updated:

1. `useEntityDiscovery.ts` - Map `referenced_schema.referenced_table` to `foreign_table`
2. `TypedValueInput.vue` - Render SmartCombobox when `fkColumn.is_foreign_key` is true
3. `ConditionRow.vue` - Pass selected column to TypedValueInput

See: `vue/ichor/.claude/plans/fk-value-lookup-workflow-conditions.md`
