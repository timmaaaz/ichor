---
name: handler-map-iteration-order
description: Handler iterates a Go map to build response slice — map iteration is non-deterministic, causing intermittent DIFF test failures
type: feedback
---

# handler-map-iteration-order

**Signal**: DIFF on a list/query test where GOT and EXP have the same items but in swapped order; failure is intermittent or varies between runs; the handler builds its response by ranging over a `map[string]...` or similar unordered structure (not a DB query)
**Root cause**: Go map iteration order is randomized per run by design. A handler that ranges over a map to build a `[]T` response will produce a different element order on each execution. Tests that hardcode the expected slice in a specific order will fail non-deterministically.
**Fix**:
1. In the failing test's `CmpFunc`, sort both GOT and EXP slices by a stable field before comparing:
   ```go
   CmpFunc: func(got, exp *actionapp.AvailableActions) string {
       gotItems := append([]actionapp.AvailableAction(nil), *got...)
       expItems := append([]actionapp.AvailableAction(nil), *exp...)
       sort.Slice(gotItems, func(i, j int) bool { return gotItems[i].Type < gotItems[j].Type })
       sort.Slice(expItems, func(i, j int) bool { return expItems[i].Type < expItems[j].Type })
       return cmp.Diff(expItems, gotItems)
   },
   ```
2. Copy each slice before sorting to avoid mutating shared test state (see `cmpfunc-slice-mutation`)
3. Do NOT add sorting to the handler itself — the handler's output is correct; only the test assertion needs to be order-agnostic

**See also**: `docs/arch/testing.md`, `cmpfunc-slice-mutation`
**Examples**:
- `actionapi_Test_ActionAPI_list-actions-200-user-with-permissions-user-sees-permitted-actions-only.md` — `listActionsHandler` ranged over a permissions map; `create_alert` and `send_notification` swapped order between runs; fixed by sorting both slices by `Type` in `CmpFunc`
