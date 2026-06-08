# jsonb-key-order-comparison

**Signal**: query-200 DIFF on a field backed by a Postgres **JSONB** column (`action_config`, `payload`, …) decoded into `json.RawMessage`; the JSON is semantically identical but key order differs, and/or whitespace differs; often intermittent (pagination decides whether the affected row is in-window)
**Root cause**: JSONB does not preserve key order (it reorders by length-then-bytewise) or whitespace (PG emits a space after `:`/`,`; Go `json.Marshal` over the wire compacts). Comparing the raw bytes of a JSONB-sourced `json.RawMessage` against an in-memory expectation fails on cosmetic differences even when the documents are equal.
**Fix**:
1. In the test `CmpFunc`, compare the `json.RawMessage` **semantically** via a `cmp.Transformer` that unmarshals to `map[string]any` (or re-marshals canonically) before diffing — never compare raw bytes.
2. Mirror the sibling bus test if one exists (e.g. `pageactionbus_test.go`'s map-unmarshal).
3. Do NOT "fix" it by re-querying the seed from the DB — that aligns key order but reintroduces the whitespace mismatch.
4. While there, copy the exp slice before any in-place sort (latent `cmpfunc-slice-mutation`).

**See also**: `docs/arch/testing.md`; related `cmpfunc-slice-mutation`
**Examples**:
- `pageactionapi_Test_PageAction_query-200-basic.md` — button `action_config` JSONB reordered keys + HTTP whitespace compaction; fixed with a `normalizeJSON` `cmp.Transformer` in `query_test.go` (re-query approach tried and reverted — exposed the whitespace mismatch).
