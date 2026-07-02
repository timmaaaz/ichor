// Package positivefixture is a static fixture for the apiconvention guard.
//
// It lives under testdata/, so `go build`/`go test`/`go vet` never compile it —
// but the guard's detector parses it as source to prove it still flags a known
// json:"items"-without-"total" encoder (see Test_Detector_FlagsKnownViolation).
//
// Do NOT "fix" KnownViolation to a documented shape: its convention violation is
// intentional and is the thing the detector's positive-path test asserts on.
package positivefixture

// KnownViolation exposes json:"items" WITHOUT json:"total" and implements Encode,
// so it is exactly the shape the guard must detect. If the detector stops
// flagging this type, its detection logic has regressed.
type KnownViolation struct {
	Items []int `json:"items"`
}

// Encode gives KnownViolation the encoder signature the detector keys on. The
// body is irrelevant — the guard only parses the source, it never runs this.
func (KnownViolation) Encode() ([]byte, string, error) {
	return nil, "application/json", nil
}
