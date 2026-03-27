# Debug Pattern Index

Read this file fully when checking for known patterns. Match by: test name, package name, error keywords.

| Pattern ID | Signal Keywords | Fix Summary |
|---|---|---|
| `table-access-count` | `len(gotResp.TableAccess) != N`, `permissionsbus_test.go`, count mismatch | Update hardcoded count in `permissionsbus_test.go` |
| `missing-table-access-seed` | 401 on ALL endpoints for one entity, newly added domain, other entities work | Add table to `seed.sql` table_access INSERT **and** `tableaccessbus/testutil.go` TestSeedTableAccess |
| `pg-time-leading-zeros` | DIFF: `"8:00:00"` vs `"08:00:00"`, `AvailableHoursStart`, `AvailableHoursEnd`, time column | Change expected to zero-padded: `"08:00:00"` |
| `business-default-in-test` | DIFF on create-200: status/state field empty in EXP but populated in GOT, `ApprovalStatus`, `Status` | Add business-layer default value to test `ExpResp` |
| `test-seed-fake-user-fk` | `foreign key violation`, `uuid.New()` or hardcoded UUID as `created_by`/`user_id` in seed | Replace `uuid.New()` with `userbus.TestSeedUsersWithNoFKs()` |
| `handler-nil-bus` | Action handler returns 404 or panics, `RegisterCoreActions`, nil bus dependency | Wire real bus in `all.go` (match `seek_approval` pattern); add nil guard in handler |
| `formconfig-value-column` | `DROPDOWN_COLUMN_NOT_FOUND`, `value column "X" not found in table Y`, `TestFormConfigsAgainstSchema` | Set `value_column` to `"id"` (target table PK, not the FK column name) |
| `omitempty-to-required` | create-400 missing-X returns 200, field should be required but isn't, `NewEntity` struct | Change `validate:"omitempty"` → `validate:"required"` in `app/domain/{area}/{entity}app/model.go` |
| `return-type-changed-to-map` | `interface conversion`, panic, typed struct assertion on action handler result | Replace `result.(pkg.Struct)` with `result.(map[string]any)` and `.Field` with `["field_name"]` |
| `seed-product-index-exhausted` | `409`, `unique_violation`, inventory item create, `Products[0]` | Use `Products[idx]` where `idx >= ceil(seed_count/location_count)` to skip saturated products |
| `formconfig-column-name-mismatch` | `COLUMN_NOT_FOUND`, `TestFormConfigsAgainstSchema`, form field Name wrong | Change `Name` in `tableforms.go` from FK column name to actual column name in target table |
| `nil-uuid-field-validation-400` | `update-200 returns 400`, `&sd.Entity[n].Field`, nullable `*uuid.UUID` seeded as nil, `min=36` | Fix testutil: change `FieldName: nil` → `FieldName: &validUUID`; validator v10 `omitempty` skips nil ptrs only |
| `create-missing-querybyid-join-fields` | `create-200 returns 200 but DIFF`, `00000000-0000-0000-0000-000000000000`, JOIN fields empty in response | Add `QueryByID` after `bus.Create` in `App.Create` to populate JOIN-enriched fields |
| `invalid-enum-check-constraint` | `create-200 returns 500`, string literal status/quality field, DB CHECK constraint | Check `migrate.sql` for valid CHECK values; replace invalid string literal in test |
| `seed-unique-pair-exhausted` | `Seeding error`, `entry is not unique`, junction table, n > unique pairs | Reduce n ≤ len(aIDs)*len(bIDs); update Total in query tests; check create-200 for collision with seeded pair |
| `cmpfunc-slice-mutation` | Intermittent DIFF on query-200, items at wrong index, `sort.Slice` in CmpFunc, sibling subtests | Copy slice before sorting: `append([]T(nil), expResp.Items...)` to prevent shared backing array mutation |
| `missing-order-by` | Non-deterministic DIFF on `QueryByIDs`/`QueryByFilter`, row order varies, no ORDER BY in SQL | Add `ORDER BY {id or natural key} ASC` to the SQL query in `{entity}db.go` |
| `missing-unique-constraint` | `create-409-duplicate-X` returns 200, no constraint violation in DB, UNIQUE missing from migration | Add `ALTER TABLE ... ADD CONSTRAINT ... UNIQUE (column)` as a new migration version in `migrate.sql` |
| `hardcoded-action-type-list` | DIFF in `list-actions-*` test, GOT has more action types than EXP, `create_alert`/`send_notification` missing from expected slice | Add missing action type to `ExpResp` in test file; match Type/Description/IsAsync from GOT |
| `bus-test-seed-contamination` | Bus query test DIFF, GOT has rows from different ProductID/entity group than EXP, `sd.Items[0:N]` wrong, items interleaved across products | Add entity-specific filter (e.g., `ProductID`) to QueryFilter; compute expected slice dynamically |
| `wrong-seed-index-after-deletion` | 200→404 after delete step, `sd.Entities[N]` hardcoded index, seed slice shifted | Update hardcoded seed index to account for deleted entity |
| `wrong-error-code-for-table-permission` | 403→401, table permission denial, `authorize.go`, `errs.Unauthenticated` | Change `errs.Unauthenticated` to `errs.PermissionDenied` in authorize.go |
| `missing-test-table-fields` | `json: Unmarshal(nil)`, test table entry missing GotResp/ExpResp/CmpFunc | Add missing struct fields to test table entry |
| `invalid-reason-code-seed` | `invalid reason code`, business-layer validation, seed uses free-text reason | Replace with valid reason constant in seed function |
| `invalid-status-seed-string` | `must be pending, got status4`, `fmt.Sprintf("status%d")`, state machine rejection | Replace computed status string with domain constant in testutil.go |
| `reference-category-count-mismatch` | `expected N types, got N+M`, category consistency test, action type registry grew | Add new action type to expected list and update count |
