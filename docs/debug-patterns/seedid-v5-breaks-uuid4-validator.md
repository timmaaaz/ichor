# seedid-v5-breaks-uuid4-validator

**Signal**: create-200 returns 400, or a create-400 test gets an *extra* spurious error; message `must be a valid version 4 UUID`; the field is a foreign-key UUID (`product_id`, `substitute_product_id`, …) referencing a seeded entity; also appears on the FormData/order path as `decode: validate: [...version 4 UUID]`
**Root cause**: `seedid.Stable` (= `uuid.NewSHA1`, PR #137) mints deterministic **v5** UUIDs. A field tagged `validate:"...,uuid4"` enforces RFC-4122 *version 4* only, so a valid seeded FK reference is rejected. The DB column is version-agnostic `uuid` — the validator over-constrains the data contract.
**Fix**:
1. Grep the app model for the offending tag: `grep -rn "uuid4" app/domain/`
2. On any field that holds a *reference* to a deterministically-seeded entity, change `validate:"...,uuid4"` → `validate:"...,uuid"` (any RFC-4122 version).
3. Leave `uuid4` only on fields that are genuinely client-generated v4 — FK references to seed data must be `uuid`.
4. If a create-400 test now sees an extra `product_id ... version 4 UUID` alongside the real expected error, the fix is this model change, NOT a test edit.

**See also**: `docs/arch/seeding.md` (⚠ "uuid4 validators reject deterministic seed IDs")
**Examples**:
- `orderlineitemsapp/model.go` ProductID `uuid4`→`uuid` cleared 5 bugs (formdataapi ×3, orderlineitemsapi ×2 incl. a create-400 spurious-extra-error); latent twin `pickingapp/model.go:41` substitute_product_id fixed in the same follow-on.
