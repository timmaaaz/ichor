package types

import (
	"database/sql"
	"fmt"
	"regexp"
	"strings"

	"github.com/timmaaaz/ichor/app/sdk/errs"
)

// Interval represents a time interval.
type Interval struct {
	value string
}

// ParseInterval parses a string into a Interval.
func ParseInterval(value string) (Interval, error) {
	if err := ValidateIntervalFormat(value); err != nil {
		return Interval{}, fmt.Errorf("invalid interval format: %q", err)
	}

	// we lowercase the interval string because postgres doesn't
	// 'Days' but it does recognize 'days'
	value = strings.ToLower(value)

	return Interval{value: value}, nil
}

func MustParseInterval(value string) Interval {
	i, err := ParseInterval(value)
	if err != nil {
		panic(fmt.Sprintf("mustparseinterval failed %q", err))
	}
	return i
}

// ParseIntervalPtr parses a string into a Interval pointer.
func ParseIntervalPtr(value string) (*Interval, error) {
	if value == "" {
		return nil, nil
	}
	i, err := ParseInterval(value)
	if err != nil {
		return nil, fmt.Errorf("invalid interval format: %q", err)
	}
	return &i, nil
}

// Value returns the value of the Interval.
func (i Interval) Value() string {
	return i.value
}

// DBValue returns the value of the Interval for database storage.
func (i Interval) DBValue() sql.NullString {
	if i.value == "" {
		return sql.NullString{Valid: false}
	}
	return sql.NullString{String: i.value, Valid: true}
}

// ValuePtr returns a pointer to the value of the Interval.
func (i Interval) ValuePtr() *string {
	return &i.value
}

// UnmarshalText unmarshals a Interval from a string implementing the
// unmarshal interface for JSON conversions.
func (i *Interval) UnmarshalText(data []byte) error {
	interval, err := ParseInterval(string(data))
	if err != nil {
		return errs.NewFieldsError("interval", err)
	}
	i.value = interval.Value()
	return nil
}

// MarshalText marshals a Interval into a string implementing the marshal
// implementing the marshal interface for JSON conversions.
func (i Interval) MarshalText() ([]byte, error) {
	return []byte(i.value), nil
}

// Equal returns true if the two intervals are equal.
func (i Interval) Equal(other Interval) bool {
	return i.value == other.value
}

func (i Interval) Ptr() *Interval {
	return &i
}

// ValidateIntervalFormat validates the format of an interval.
func ValidateIntervalFormat(value string) error {
	r, _ := regexp.Compile(`(?i)^\s*(?:(\d+)\s*(?:year|years))?\s*(?:(\d+)\s*(?:month|months))?\s*(?:(\d+)\s*(?:day|days))?\s*$`)
	if !r.MatchString(value) {
		return fmt.Errorf("invalid interval format for %s, must be A year(s) B month(s) C day(s)", value)
	}
	return nil
}
