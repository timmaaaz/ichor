package inventory

// quantityFromRawData extracts a line item's requested quantity from a trigger
// event's raw entity data.
//
// Trigger RawData is reconstructed by a JSON round-trip (see
// temporal/enrichment.go: marshal entity -> unmarshal into map[string]any), so
// numeric quantities always arrive as float64 on the real line-item path;
// direct unit-test injection may use a plain int. There is deliberately no
// string case: no production path produces a string quantity, and accepting one
// here would diverge from the identical parse used by reserve_inventory and
// allocate. Returns false when no usable quantity is present.
func quantityFromRawData(raw map[string]any) (int, bool) {
	switch q := raw["quantity"].(type) {
	case float64:
		return int(q), true
	case int:
		return q, true
	}
	return 0, false
}
