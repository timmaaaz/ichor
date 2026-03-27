# invalid-reason-code-seed

**Signal**: `create: invalid reason code`, `invalid reason`, business-layer validation rejection on reason/cause field; seed or test helper uses a free-text string instead of a valid enum value
**Root cause**: Business layer validates reason codes (e.g., adjustment reasons) against a fixed set of valid values. Test seed functions use arbitrary strings like `"correction"` or `"damage"` that aren't in the valid set. Unlike DB CHECK constraints (which fail at 500), these fail at the business layer with a descriptive validation error.
**Fix**:
1. Find the valid reason codes — grep for the validation logic in the bus layer (e.g., `validReasonCodes`, `isValidReason`, or the `Parse` function for the reason type)
2. Replace the invalid string in the test seed function with a valid constant
3. If the seed function is shared across multiple tests, verify all consumers still pass

**See also**: `docs/arch/domain-template.md`
**Examples**:
- `inventory_Test_ApproveInventoryAdjustment.md` — seed used `"correction"` and `"damage"` as reason codes; valid values are `"data_entry_error"` and `"damaged"`; fixed in shared seed function
- `inventory_Test_RejectInventoryAdjustment.md` — same shared seed function fix as above
