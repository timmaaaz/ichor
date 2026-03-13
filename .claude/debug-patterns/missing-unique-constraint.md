---
name: missing-unique-constraint
description: Test expects 409 on duplicate insert but DB has no UNIQUE constraint — conflict test always returns 200 instead
type: feedback
---

# missing-unique-constraint

**Signal**: `create-409-duplicate-X` or `update-409-duplicate-X` test returns 200 instead of 409; no constraint violation in DB logs; the field being tested for uniqueness has no `UNIQUE` constraint or index in `migrate.sql`
**Root cause**: The test was written assuming a uniqueness invariant (e.g., UPC code must be unique per product), but the migration never added a `UNIQUE` constraint or partial unique index for that column. Without the DB constraint, duplicates are silently accepted and the server returns 200.
**Fix**:
1. Verify the business intent: should the field actually be unique?
2. If yes, add a migration version in `business/sdk/migrate/sql/migrate.sql`:
   ```sql
   -- Version: X.YY
   -- Description: Add unique constraint on {table}.{column}
   ALTER TABLE {schema}.{table} ADD CONSTRAINT {table}_{column}_unique UNIQUE ({column});
   -- OR for a partial unique index:
   CREATE UNIQUE INDEX {table}_{column}_unique_idx ON {schema}.{table} ({column}) WHERE {condition};
   ```
3. Never edit an existing migration version — always add a new one
4. Re-run `make migrate` and re-seed to apply

**See also**: `docs/arch/sqldb.md`, `business/sdk/migrate/sql/migrate.sql`
**Examples**:
- `productapi_Test_InventoryProduct_create-409-duplicate-upc-code.md` — test expected 409 on duplicate `upc_code`; migration had no UNIQUE constraint; fixed by adding migration version 2.14 with `UNIQUE` on `products.upc_code`
- `productapi_Test_InventoryProduct_update-409-duplicate-upc-code.md` — same missing constraint; update path also accepted the duplicate; same migration fix resolved both create and update 409 tests
