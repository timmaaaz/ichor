package putawaytaskbus

import "fmt"

type statusSet struct {
	Pending    Status
	InProgress Status
	Completed  Status
	Cancelled  Status
}

// Statuses represents the set of valid put-away task statuses.
var Statuses = statusSet{
	Pending:    newStatus("pending"),
	InProgress: newStatus("in_progress"),
	Completed:  newStatus("completed"),
	Cancelled:  newStatus("cancelled"),
}

// =============================================================================

// Set of known statuses.
var statuses = make(map[string]Status)

// Status represents a put-away task status in the system.
type Status struct {
	name string
}

func newStatus(s string) Status {
	st := Status{s}
	statuses[s] = st
	return st
}

// String returns the name of the status.
func (s Status) String() string {
	return s.name
}

// Equal provides support for the go-cmp package and testing.
func (s Status) Equal(s2 Status) bool {
	return s.name == s2.name
}

// MarshalText implements encoding.TextMarshaler.
func (s Status) MarshalText() ([]byte, error) {
	return []byte(s.name), nil
}

// UnmarshalText implements encoding.TextUnmarshaler.
func (s *Status) UnmarshalText(data []byte) error {
	st, err := ParseStatus(string(data))
	if err != nil {
		return err
	}
	*s = st
	return nil
}

// =============================================================================

// ParseStatus parses the string value and returns a status if one exists.
func ParseStatus(value string) (Status, error) {
	st, exists := statuses[value]
	if !exists {
		return Status{}, fmt.Errorf("invalid status %q", value)
	}
	return st, nil
}

// MustParseStatus parses the string value and returns a status if one exists.
// Panics if the status is invalid.
func MustParseStatus(value string) Status {
	st, err := ParseStatus(value)
	if err != nil {
		panic(err)
	}
	return st
}
