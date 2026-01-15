// Package types provides value objects for the orders domain.
//
// Percentage represents a tax rate or other percentage value in whole number form.
// For example, an 8.25% tax rate is stored as "8.25" (not "0.0825").
// This matches user expectations and simplifies form validation.
//
// For arithmetic operations, convert to decimal.Decimal:
//
//	taxRate, _ := types.ParsePercentage("8.25")
//	taxDec, _ := decimal.NewFromString(taxRate.Value())
//	// The calculations package handles division by 100 internally
//
// See also: Money for monetary values, calculations package for arithmetic.
package types

import (
	"database/sql"
	"fmt"
	"regexp"
	"strconv"

	"github.com/timmaaaz/ichor/app/sdk/errs"
)

// percentageFormatRegex validates percentage format:
// - No leading zeros (except "0" itself or "0.xx")
// - Optional decimal with 1-2 decimal places
// Valid: "0", "8", "8.5", "8.25", "100", "0.01", "0.1"
// Invalid: "08", "8.", "8.125", " 8", "8 "
// Compiled once at package init for performance.
var percentageFormatRegex = regexp.MustCompile(`^(0|[1-9]\d*)(\.\d{1,2})?$`)

// Percentage represents a percentage value (0-100) stored as a validated string.
// Used for tax rates and other percentage-based values.
//
// Two decimal places (e.g., 8.25%) provides sufficient precision for tax rates
// while matching the database schema DECIMAL(5,2).
type Percentage struct {
	value string
}

// ParsePercentage parses a string into a Percentage.
// Valid range is 0-100 with up to 2 decimal places.
func ParsePercentage(value string) (Percentage, error) {
	if value == "" {
		return Percentage{}, nil
	}
	if err := ValidatePercentageFormat(value); err != nil {
		return Percentage{}, fmt.Errorf("invalid percentage format: %w", err)
	}
	return Percentage{value: value}, nil
}

// ParsePercentagePtr parses a string into a Percentage pointer.
func ParsePercentagePtr(value string) (*Percentage, error) {
	if value == "" {
		return nil, nil
	}
	p, err := ParsePercentage(value)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

// MustParsePercentage parses a string into a Percentage or panics.
// For use in test code and seed data only.
func MustParsePercentage(value string) Percentage {
	p, err := ParsePercentage(value)
	if err != nil {
		panic(fmt.Sprintf("invalid percentage value: %s", value))
	}
	return p
}

// Value returns the string value of the Percentage.
func (p Percentage) Value() string {
	return p.value
}

// DBValue returns the value for database storage as sql.NullString.
func (p Percentage) DBValue() sql.NullString {
	if p.value == "" {
		return sql.NullString{Valid: false}
	}
	return sql.NullString{String: p.value, Valid: true}
}

// ValuePtr returns a pointer to the string value.
func (p Percentage) ValuePtr() *string {
	return &p.value
}

// UnmarshalText implements encoding.TextUnmarshaler for JSON.
func (p *Percentage) UnmarshalText(data []byte) error {
	pct, err := ParsePercentage(string(data))
	if err != nil {
		return errs.NewFieldsError("percentage", err)
	}
	p.value = pct.Value()
	return nil
}

// MarshalText implements encoding.TextMarshaler for JSON.
func (p Percentage) MarshalText() ([]byte, error) {
	return []byte(p.value), nil
}

// Equal returns true if two Percentage values are equal.
func (p Percentage) Equal(p2 Percentage) bool {
	return p.value == p2.value
}

// ValidatePercentageFormat validates percentage string format.
// Accepts: "0", "8", "8.25", "100", up to 2 decimal places.
// Must be in range 0-100.
//
// Two decimal places (e.g., 8.25%) provides sufficient precision for tax rates
// while matching the database schema DECIMAL(5,2).
func ValidatePercentageFormat(val string) error {
	// Use pre-compiled regex for performance
	if !percentageFormatRegex.MatchString(val) {
		return fmt.Errorf("invalid percentage format for %s", val)
	}

	// Parse and validate range
	f, err := strconv.ParseFloat(val, 64)
	if err != nil {
		return fmt.Errorf("invalid percentage value: %s", val)
	}
	if f < 0 || f > 100 {
		return fmt.Errorf("percentage must be between 0 and 100, got %s", val)
	}

	return nil
}
