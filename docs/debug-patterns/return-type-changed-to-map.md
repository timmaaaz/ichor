# return-type-changed-to-map

**Signal**: `interface conversion`, `panic`, `.(inventory.QueuedAllocationResponse)`, `.(SomeTypedStruct)`, type assertion failure on action handler result
**Root cause**: Action handler `Execute()` return type changed from domain-specific struct to `map[string]any` (Temporal activity serialization); test type assertions not updated.
**Fix**:
1. Find the failing type assertion in the test (e.g., `result.(pkg.SomeStruct)`)
2. Replace with `result.(map[string]any)`
3. Replace field access `.FieldName` with map key `["field_name"]` (snake_case JSON key)
4. Cast numeric values: map values are `float64` after JSON round-trip (use type assertion or `.(float64)`)
5. Nested structs become `map[string]any` as well -- chain map lookups

**See also**: `docs/arch/workflow-engine.md`
**Examples**:
- `inventory_Test_AllocateInventory_allocateInventory-test_order_grouped_allocation.md` -- Execute() returned map[string]any, test asserted QueuedAllocationResponse
- `inventory_Test_AllocateInventory_allocateInventory-test_source_from_line_item.md` -- same root cause, same fix pattern
