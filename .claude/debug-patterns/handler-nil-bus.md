# handler-nil-bus

**Signal**: Action handler test returns 404 or nil pointer panic; handler is registered via `RegisterCoreActions` but needs a real bus dependency (alertBus, notifBus, etc.) that was not wired in `all.go`
**Root cause**: A new action handler was added to `RegisterCoreActions` with `nil` for its bus dependency (acceptable for graceful degradation in tests), but the production wiring in `all.go` was not updated to pass the real bus. The handler silently no-ops or panics when called.
**Fix**:
1. Find the handler registration in `business/sdk/workflow/workflowactions/register.go`
2. Check how the existing `seek_approval` handler is wired in `api/cmd/services/ichor/build/all/all.go` — it is registered with nil first, then upgraded with a real bus
3. Add an equivalent upgrade call for the new handler in `all.go`, passing the real bus
4. Add a nil guard to the handler's `Execute()` method if not present (graceful degradation when bus is nil in tests)

**See also**: `docs/arch/workflow-engine.md`
**Examples**:
- `actionapi_Test_ActionAPI_execute-200-create-alert-user-with-alert-perm-executes-create-alert.md` — `create_alert` handler registered with `nil` alertBus; wired with real alertBus in `all.go:501` matching the `seek_approval` pattern
