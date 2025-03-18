package types

import (
	"database/sql"
	"fmt"
	"regexp"

	"github.com/timmaaaz/ichor/app/sdk/errs"
)

// Money represents a monetary value.
type Money struct {
	value string
}

func MustParseMoney(value string) Money {
	m, err := ParseMoney(value)
	if err != nil {
		panic(err)
	}
	return m
}

// ParseMoney parses a string into a Money.
func ParseMoney(value string) (Money, error) {
	if value == "" {
		return Money{}, nil
	}
	if err := ValidateCurrencyFormat(value); err != nil {
		return Money{}, fmt.Errorf("invalid currency format: %q", err)
	}
	return Money{value: value}, nil
}

// ParseMoneyPtr parses a string into a Money pointer.
func ParseMoneyPtr(value string) (*Money, error) {
	if value == "" {
		return nil, nil
	}
	m, err := ParseMoney(value)
	if err != nil {
		return nil, fmt.Errorf("invalid currency format: %q", err)
	}
	return &m, nil
}

// Value returns the value of the Money.
func (m Money) Value() string {
	return m.value
}

// DBValue returns the value of the Interval for database storage.
func (m Money) DBValue() sql.NullString {
	if m.value == "" {
		return sql.NullString{Valid: false}
	}
	return sql.NullString{String: m.value, Valid: true}
}

// ValuePtr returns a pointer to the value of the Money.
func (m Money) ValuePtr() *string {
	return &m.value
}

// UnmarshalText unmarshals a Money from a string implementing the
// unmarshal interface for JSON conversions.
func (m *Money) UnmarshalText(data []byte) error {
	money, err := ParseMoney(string(data))
	if err != nil {
		return errs.NewFieldsError("money", err)
	}
	m.value = money.Value()
	return nil
}

// MarshalText marshals a Money into a string implementing the marshal
// implementing the marshal interface for JSON conversions.
func (m Money) MarshalText() ([]byte, error) {
	return []byte(m.value), nil
}

// Equal returns true if the Money is equal to the given Money.
func (m Money) Equal(m2 Money) bool {
	return m.value == m2.value
}

// ValidateCurrencyFormat validates the format of a currency string.
func ValidateCurrencyFormat(val string) error {
	r, _ := regexp.Compile(`^\-?\d+\.?\d{0,4}$`)
	if ok := r.MatchString(val); !ok {
		return fmt.Errorf("invalid currency format for %s", val)
	}
	return nil
}
