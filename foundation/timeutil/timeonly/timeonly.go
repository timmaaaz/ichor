package timeonly

import (
	"fmt"
	"time"
)

const timeFmt = "15:04:05"
const timeFmtShort = "15:04"

// TimeOnly represents a time with no date value.
type TimeOnly struct {
	value string
}

// ValidateTimeOnlyFmt validates the format of a time string.
// Accepts both "HH:MM:SS" and "HH:MM" formats.
func ValidateTimeOnlyFmt(val string) bool {
	_, err := time.Parse(timeFmt, val)
	if err != nil {
		_, err = time.Parse(timeFmtShort, val)
	}
	return err == nil
}

// NormalizeTimeOnly validates and normalizes a time string to "HH:MM:SS" format.
// Accepts both "HH:MM:SS" and "HH:MM" formats, always returns "HH:MM:SS".
func NormalizeTimeOnly(val string) (string, error) {
	if t, err := time.Parse(timeFmt, val); err == nil {
		return t.Format(timeFmt), nil
	}
	if t, err := time.Parse(timeFmtShort, val); err == nil {
		return t.Format(timeFmt), nil
	}
	return "", fmt.Errorf("invalid time format: %q, expected HH:MM:SS or HH:MM", val)
}

// GetTimeFmt returns the time format.
func GetTimeFmt() string {
	return timeFmt
}

// ParseTimeOnly parses a time.Time into a TimeOnly.
func ParseTimeOnly(value time.Time) (TimeOnly, error) {
	return TimeOnly{value: value.Format(timeFmt)}, nil
}

// ParseTimeOnlyFromString parses a string into a TimeOnly.
func ParseTimeOnlyFromString(value string) (TimeOnly, error) {
	if !ValidateTimeOnlyFmt(value) {
		return TimeOnly{}, fmt.Errorf("invalid time format: %q", value)
	}
	return TimeOnly{value: value}, nil
}

// ParseTimeOnlyPtr parses a time.Time into a TimeOnly pointer.
func ParseTimeOnlyPtr(value time.Time) (*TimeOnly, error) {
	var zero time.Time
	if value == zero {
		return nil, nil
	}
	m, err := ParseTimeOnly(value)
	if err != nil {
		return nil, fmt.Errorf("invalid time format: %q", err)
	}
	return &m, nil
}

// Value returns the value of the TimeOnly.
func (to TimeOnly) Value() string {
	return to.value
}

// ValuePtr returns a pointer to the value of the TimeOnly.
func (to TimeOnly) ValuePtr() *string {
	return &to.value
}
