# test-seed-fake-user-fk

**Signal**: `creating rule: create: namedexeccontext: foreign key violation` (or similar FK violation) in test output; test seed creates a random `uuid.New()` and uses it as a `created_by`, `user_id`, or similar FK that references `core.users`
**Root cause**: Test seed uses `uuid.New()` to generate a random UUID for a FK field, but that UUID is never inserted into the referenced table (e.g., `core.users`). PostgreSQL rejects the insert with a FK violation.
**Fix**:
1. Find `uuid.New()` (or `uuid.MustParse(...)` with a hardcoded UUID) used as a FK to `core.users` in the test seed
2. Replace with `userbus.TestSeedUsersWithNoFKs(ctx, 1, userbus.Roles.Admin, db.BusDomain.User, ...)` to seed a real user
3. Use the returned user's ID for the FK field
4. If the call is in a helper function, add `userBus *userbus.Business` as a parameter and pass `db.BusDomain.User` from the caller

**See also**: `docs/arch/seeding.md`
**Examples**:
- `actionhandlers_TestCallWebhookAction.md` — `adminID := uuid.New()` used as `created_by` on `workflow.automation_rules`; fixed by seeding a real user
- `actionhandlers_TestSendEmailAction.md` — same pattern in shared `seedCommsActionRule` helper
