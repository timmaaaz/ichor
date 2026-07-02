# scanner-without-valuer

**Signal**: query-200 DIFF on a JSONB / nullable-JSON column where GOT is a `{"Data":…,"Valid":true}` wrapper (or `{"Data":null,"Valid":false}`) but EXP is the bare payload. Tell-tale: a sibling plain `json.RawMessage` field in the *same* struct matches fine, so it is not a global JSON-comparison issue. GOT bytes are jsonb-reformatted (keys length-sorted, spaces after `:`/`,`), confirming the wrapper was *stored*, not produced on read.
**Root cause**: The nullable wrapper type (e.g. `nulltypes.NullRawMessage{Data, Valid}`) implements `sql.Scanner` (read) but NOT `driver.Valuer` (write). When bound as an INSERT/UPDATE param, pgx has no Valuer to call, so it JSON-marshals the whole `{Data, Valid}` Go struct into the JSONB column — persisting the wrapper object instead of the inner payload. The `!Valid` path marshals to `{"Data":null,"Valid":false}` instead of SQL NULL, also corrupting production NULL semantics.
**Fix**:
1. Grep the failing column's Go type in `business/sdk/sqldb/nulltypes/` and confirm it has a `Scan` method but no `Value` method.
2. Add `func (n T) Value() (driver.Value, error)` returning `[]byte(n.Data), nil` when valid and `nil, nil` (SQL NULL) when `!n.Valid || len(n.Data) == 0`. Mirror `sql.NullString`'s Scanner/Valuer symmetry.
3. **Use a value receiver, not a pointer receiver.** The db model binds the type by value, so only a value-receiver `Value()` lands in `T`'s method set and satisfies `driver.Valuer`; a `*T` receiver would be silently skipped and pgx would still marshal the wrapper.

**Distinguish from `jsonb-key-order-comparison`**: that is a *test bug* — the stored data is correct and only byte/key order differs, so you fix the comparison (`cmp.Transformer`). This is a *code bug* — the stored content is structurally wrong. Tell them apart by shape: key reorder / whitespace only → `jsonb-key-order-comparison`; extra `Data` / `Valid` keys wrapping the real value → this pattern.

**See also**: `docs/arch/sqldb.md`
**Examples**:
- `workflow_Test_Workflow_automationExecution-queryHistory.md` — `NullRawMessage` had a `Scan` but no `Value`; the `actions_executed` JSONB column persisted `{"Data":[…],"Valid":true}` instead of the bare action array while the sibling plain-`json.RawMessage` `TriggerData` matched. Fixed with a value-receiver `Value()` at `business/sdk/sqldb/nulltypes/nulltypes.go:45`.
