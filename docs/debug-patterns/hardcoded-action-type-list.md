# hardcoded-action-type-list

**Signal**: DIFF in `list-actions-*` test; GOT contains an action type (e.g., `create_alert`, `send_notification`) not present in EXP; test slice is shorter than actual response
**Root cause**: Integration test hardcodes a slice of expected `AvailableAction` entries; when a new handler is added to `RegisterCoreActions` (or any action registration function) and it satisfies `SupportsManualExecution()=true` plus the test user's permission check, it appears in the API response but not in the test's `ExpResp`.
**Fix**:
1. Run the test and read the GOT output — it lists all currently registered and permitted actions
2. Add the missing action type(s) to the `ExpResp` slice in the failing test file (`tests/workflow/actionapi/list_test.go` or similar)
3. Match the `Type`, `Description`, and `IsAsync` fields from GOT exactly
4. Check `RegisterCoreActions` in `business/domain/workflow/workflowbus/register.go` (or `all.go`) to confirm the handler was intentionally registered

**See also**: `docs/arch/workflow-engine.md`
**Examples**:
- `actionapi_Test_ActionAPI_list-actions-200-user-with-permissions-user-sees-permitted-actions-only.md` — `create_alert` handler added to `RegisterCoreActions` with nil buses; both `send_notification` and `create_alert` satisfy permission check; test only had `send_notification` in expected slice
