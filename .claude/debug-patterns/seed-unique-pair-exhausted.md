# seed-unique-pair-exhausted

**Signal**: `Seeding error: seeding {entities}`, `namedexeccontext: {entity} entry is not unique`, seed fails before any test runs; entity is a junction/association table with a UNIQUE constraint on two FK columns
**Root cause**: `TestSeed*` called with `n` greater than `len(fkA_ids) * len(fkB_ids)` — the number of unique (A, B) pairs is exhausted. Random index cycling eventually repeats a pair → unique constraint violation.
**Fix**:
1. Count the IDs passed for each FK dimension in the seed call
2. Ensure `n ≤ product of distinct FK counts`: `n ≤ len(aIDs) * len(bIDs)`
3. Reduce `n` in the seed call, and update any hardcoded `Total:` count in query tests to match

**Also check**: if the test's `create-200` subtest uses `Entity[0]` from seed data for both FKs, it may collide with an already-seeded pair — use a different index (e.g., `Entity[1]`) for the create test input.

**See also**: `docs/arch/seeding.md`
**Examples**:
- `lotlocationapi_Test_LotLocation.md` — `TestSeedLotLocations(n=15, lotIDs, locationIDs)` with only 10 unique pairs available → unique violation; fixed by `n=10`; also required fixing `create-200` to use `InventoryLocations[1]` to avoid colliding with a seeded pair
