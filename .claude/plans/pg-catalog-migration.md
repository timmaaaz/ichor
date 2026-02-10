# Migrate All `information_schema` Queries to `pg_catalog`

## Context

The FK enrichment plan (`introspection-fk-enrichment.md`) already converts `QueryColumns` to use `pg_catalog`. This plan covers **all remaining** `information_schema` usage across the codebase for consistency, since Ichor is PostgreSQL-only and `pg_catalog` is the idiomatic, faster approach (no view abstraction layer overhead).

**Note**: `QueryColumns` is NOT included here -- it's covered by the FK enrichment plan.

---

## Files to Modify

| # | File | What changes |
|---|------|-------------|
| 1 | `business/domain/introspectionbus/introspectionbus.go` | 4 queries: QuerySchemas, QueryTables, QueryRelationships, QueryReferencingTables |
| 2 | `business/domain/config/formfieldbus/stores/formfielddb/formfielddb.go` | 2 table-existence checks (Create, Update) |
| 3 | `business/domain/core/tableaccessbus/stores/tableaccessdb/tableaccessdb.go` | 1 table-existence check (Create) |
| 4 | `business/sdk/migrate/sql/seed.sql` | 2 refs in workflow entity seeding CTE |
| 5 | `business/sdk/tablebuilder/typemapper.go` | Update 2 comments (no logic changes) |

---

## 1. introspectionbus.go -- QuerySchemas (line 32)

**Current**: `information_schema.schemata`

**New**:
```sql
SELECT
    nspname AS name
FROM
    pg_catalog.pg_namespace
WHERE
    nspname NOT IN ('pg_catalog', 'information_schema', 'pg_toast')
    AND nspname NOT LIKE 'pg_temp_%'
    AND nspname NOT LIKE 'pg_toast_temp_%'
ORDER BY
    nspname
```

**Why extra filters**: `pg_namespace` includes temp schemas that `information_schema.schemata` hides.

---

## 2. introspectionbus.go -- QueryTables (line 55)

**Current**: `information_schema.tables` joined with `pg_class`/`pg_namespace` (redundant mix)

**New** (pure `pg_catalog`, eliminates the redundant join):
```sql
SELECT
    n.nspname AS schema,
    c.relname AS name,
    CAST(c.reltuples AS bigint) AS row_count_estimate
FROM
    pg_catalog.pg_class c
JOIN
    pg_catalog.pg_namespace n ON c.relnamespace = n.oid
WHERE
    n.nspname = :schema
    AND c.relkind = 'r'
ORDER BY
    c.relname
```

**Note**: `relkind = 'r'` is ordinary tables (equivalent to `table_type = 'BASE TABLE'`). This is cleaner since the current query was already joining to pg_class for reltuples.

---

## 3. introspectionbus.go -- QueryRelationships (line 137)

**Current**: 3-way join on `information_schema.table_constraints`, `key_column_usage`, `constraint_column_usage`

**New**:
```sql
SELECT
    con.conname AS foreign_key_name,
    att.attname AS column_name,
    ref_ns.nspname AS referenced_schema,
    ref_cl.relname AS referenced_table,
    ref_att.attname AS referenced_column,
    'many-to-one' AS relationship_type
FROM
    pg_catalog.pg_constraint con
JOIN pg_catalog.pg_class cl ON con.conrelid = cl.oid
JOIN pg_catalog.pg_namespace ns ON cl.relnamespace = ns.oid
JOIN pg_catalog.pg_class ref_cl ON con.confrelid = ref_cl.oid
JOIN pg_catalog.pg_namespace ref_ns ON ref_cl.relnamespace = ref_ns.oid
CROSS JOIN LATERAL unnest(con.conkey, con.confkey)
    WITH ORDINALITY AS cols(src_attnum, ref_attnum, ord)
JOIN pg_catalog.pg_attribute att
    ON att.attrelid = con.conrelid AND att.attnum = cols.src_attnum
JOIN pg_catalog.pg_attribute ref_att
    ON ref_att.attrelid = con.confrelid AND ref_att.attnum = cols.ref_attnum
WHERE
    con.contype = 'f'
    AND ns.nspname = :schema
    AND cl.relname = :table
ORDER BY
    con.conname, cols.ord
```

**Why LATERAL unnest**: Correctly handles composite foreign keys by pairing source/referenced columns by ordinal position.

---

## 4. introspectionbus.go -- QueryReferencingTables (line 182)

**Current**: Same 3-way `information_schema` join, but filtering on the *referenced* side

**New**:
```sql
SELECT
    ns.nspname AS schema,
    cl.relname AS table,
    att.attname AS fk_column,
    con.conname AS constraint_name
FROM
    pg_catalog.pg_constraint con
JOIN pg_catalog.pg_class cl ON con.conrelid = cl.oid
JOIN pg_catalog.pg_namespace ns ON cl.relnamespace = ns.oid
JOIN pg_catalog.pg_class ref_cl ON con.confrelid = ref_cl.oid
JOIN pg_catalog.pg_namespace ref_ns ON ref_cl.relnamespace = ref_ns.oid
CROSS JOIN LATERAL unnest(con.conkey)
    WITH ORDINALITY AS cols(attnum, ord)
JOIN pg_catalog.pg_attribute att
    ON att.attrelid = con.conrelid AND att.attnum = cols.attnum
WHERE
    con.contype = 'f'
    AND ref_ns.nspname = :schema
    AND ref_cl.relname = :table
ORDER BY
    ns.nspname, cl.relname
```

---

## 5. formfielddb.go -- Table existence checks (lines 51-57, 99-105)

Both `Create` and `Update` have identical checks. Replace both with:

```sql
SELECT EXISTS (
    SELECT 1
    FROM pg_catalog.pg_class c
    JOIN pg_catalog.pg_namespace n ON c.relnamespace = n.oid
    WHERE n.nspname = :schema
    AND c.relname = :table
    AND c.relkind IN ('r', 'v')
)
```

**Note**: `relkind IN ('r', 'v')` matches both tables and views, same as `information_schema.tables` behavior.

---

## 6. tableaccessdb.go -- Table existence check (lines 68-74)

Same pattern as formfielddb:

```sql
SELECT EXISTS (
    SELECT 1
    FROM pg_catalog.pg_class c
    JOIN pg_catalog.pg_namespace n ON c.relnamespace = n.oid
    WHERE n.nspname = :schema_name
    AND c.relname = :table_name
    AND c.relkind IN ('r', 'v')
)
```

---

## 7. seed.sql -- Workflow entity seeding CTE (lines 573-611)

Replace the `entity_data` CTE:

```sql
WITH entity_data AS (
    -- Get all tables from your schemas
    SELECT DISTINCT
        c.relname as name,
        n.nspname as schema_name,
        'table' as entity_type_name,
        CASE
            WHEN n.nspname = 'core' THEN 'Core system tables for authentication and base configuration'
            WHEN n.nspname = 'hr' THEN 'Human Resources tables'
            WHEN n.nspname = 'geography' THEN 'Location and geography reference tables'
            WHEN n.nspname = 'assets' THEN 'Asset management tables'
            WHEN n.nspname = 'inventory' THEN 'Inventory and warehouse management tables'
            WHEN n.nspname = 'products' THEN 'Product information management tables'
            WHEN n.nspname = 'procurement' THEN 'Supply chain and procurement tables'
            WHEN n.nspname = 'sales' THEN 'Sales and order management tables'
            WHEN n.nspname = 'workflow' THEN 'Workflow and automation tables'
            WHEN n.nspname = 'config' THEN 'Configuration tables'
            ELSE 'Database table'
        END as description
    FROM pg_catalog.pg_class c
    JOIN pg_catalog.pg_namespace n ON c.relnamespace = n.oid
    WHERE n.nspname IN ('core', 'hr', 'geography', 'assets', 'inventory', 'products', 'procurement', 'sales', 'workflow', 'config')
      AND c.relkind = 'r'

    UNION ALL

    -- Get all views from your schemas
    SELECT DISTINCT
        c.relname as name,
        n.nspname as schema_name,
        'view' as entity_type_name,
        CASE
            WHEN n.nspname = 'sales' THEN 'Sales management view'
            WHEN n.nspname = 'workflow' THEN 'Workflow management view'
            WHEN n.nspname = 'config' THEN 'Configuration view'
            ELSE 'Database view'
        END as description
    FROM pg_catalog.pg_class c
    JOIN pg_catalog.pg_namespace n ON c.relnamespace = n.oid
    WHERE n.nspname IN ('core', 'hr', 'geography', 'assets', 'inventory', 'products', 'procurement', 'sales', 'workflow', 'config')
      AND c.relkind = 'v'
)
```

---

## 8. typemapper.go -- Comment updates only

- Line 23: `"from information_schema"` → `"from pg_catalog"`
- Line 77: same change

No logic changes.

---

## Implementation Order

1. `introspectionbus.go` - all 4 queries (biggest change, most value)
2. `formfielddb.go` - both existence checks
3. `tableaccessdb.go` - existence check
4. `seed.sql` - CTE rewrite
5. `typemapper.go` - comments

---

## Verification

```bash
# Compile check
go build ./...

# Run introspection tests
go test -v ./api/cmd/services/ichor/tests/introspectionapi/...

# Run formfield tests
go test -v ./business/domain/config/formfieldbus/...
go test -v ./api/cmd/services/ichor/tests/config/formfieldapi/...

# Run tableaccess tests
go test -v ./business/domain/core/tableaccessbus/...
go test -v ./api/cmd/services/ichor/tests/core/tableaccessapi/...

# Run all tests to catch any regressions
make test

# Verify seed still works
make seed
```
