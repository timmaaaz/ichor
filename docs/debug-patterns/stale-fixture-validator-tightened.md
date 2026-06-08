# stale-fixture-validator-tightened

**Signal**: a previously-passing create-200 suddenly returns 400 with no test-code change; the rejected field value exceeds a recently-tightened `max=`/`min=` validator tag; seed rows (shorter) still pass, so create-409 / other subtests are unaffected
**Root cause**: A PR tightened a validator constraint in the app model (e.g. `Code` `max=32`→`max=12`), but static test fixtures still use violating values, so `errs.Check` rejects them as `InvalidArgument` → 400 before any business logic runs.
**Fix**:
1. Find the validator that changed: grep the app model for the field + `max=`/`min=`; check `git log -p` / the PR that tightened it.
2. If the tightening is intentional (usually — confirm the *why*), shorten/adjust the fixture to satisfy it.
3. Only if the constraint itself is wrong, relax it in the model — and state the reason.

**See also**: `docs/arch/errs.md`
**Examples**:
- `labelapi…create-200-basic` / `…with-entity-ref-and-payload` — `Code` `max=32`→`max=12` (PR #144, Zebra GK420t barcode render budget); fixtures `NEW-CREATE-00x` (14 chars) shortened to `NEW-CRT-00x` (11).
