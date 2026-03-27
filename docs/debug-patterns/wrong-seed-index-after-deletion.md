# wrong-seed-index-after-deletion

**Signal**: API test expects 200 but gets 404; test deletes a seed entity in an earlier step then references a later seed entity by hardcoded index; `sd.Entities[N]` where N is now invalid after deletion
**Root cause**: Seed data slice indices shift when a preceding test step deletes an entity. The hardcoded index in a later test step still points to the old position, which may now reference the deleted entity (404) or a different entity entirely (wrong data).
**Fix**:
1. Identify which seed entity was deleted in a prior test step
2. Determine which index in the seed slice now points to the intended entity (accounting for the gap)
3. Update the hardcoded index in the failing test step

**See also**: `docs/arch/testing.md`, `docs/arch/seeding.md`
**Examples**:
- `inspectionapi_Test_Inspections_fail-200-no-quarantine-fail-without-quarantine.md` — test deleted sd.Inspections[0], then referenced sd.Inspections[1] which was now the deleted item's former neighbor; fixed by changing to sd.Inspections[3]
