package formdataapp

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"github.com/shopspring/decimal"
)

// Set of error variables for type conversion operations.
var (
	ErrParseInt     = errors.New("parse int failed")
	ErrParseDecimal = errors.New("parse decimal failed")
	ErrParseString  = errors.New("parse string failed")
	ErrMissingField = errors.New("required field missing")
	ErrInvalidType  = errors.New("invalid type for conversion")
)

// parseIntFromAny safely extracts an integer from interface{} values.
// Returns error if the value cannot be parsed.
//
// Handles the following types:
//   - nil: returns 0, nil (nil is a valid zero value)
//   - int, int64: direct conversion
//   - float64: truncates to int
//   - json.Number: parses as int64
//   - string: parses as int
func parseIntFromAny(val any) (int, error) {
	if val == nil {
		return 0, nil
	}
	switch v := val.(type) {
	case int:
		return v, nil
	case int64:
		return int(v), nil
	case float64:
		return int(v), nil
	case json.Number:
		i, err := v.Int64()
		if err != nil {
			return 0, fmt.Errorf("%w: json.Number %q: %v", ErrParseInt, v.String(), err)
		}
		return int(i), nil
	case string:
		i, err := strconv.Atoi(v)
		if err != nil {
			return 0, fmt.Errorf("%w: string %q: %v", ErrParseInt, v, err)
		}
		return i, nil
	default:
		return 0, fmt.Errorf("%w: unsupported type %T for int conversion", ErrInvalidType, val)
	}
}

// parseDecimalFromAny safely extracts a decimal from interface{} values.
// Returns error if the value cannot be parsed.
//
// Handles the following types:
//   - nil: returns decimal.Zero, nil (nil is a valid zero value)
//   - string: parses as decimal (e.g., "10.50")
//   - float64: converts to decimal
//   - int, int64: converts to decimal
//   - json.Number: parses string representation
func parseDecimalFromAny(val any) (decimal.Decimal, error) {
	if val == nil {
		return decimal.Zero, nil
	}
	switch v := val.(type) {
	case string:
		if v == "" {
			return decimal.Zero, fmt.Errorf("%w: empty string", ErrParseDecimal)
		}
		d, err := decimal.NewFromString(v)
		if err != nil {
			return decimal.Zero, fmt.Errorf("%w: string %q: %v", ErrParseDecimal, v, err)
		}
		return d, nil
	case float64:
		return decimal.NewFromFloat(v), nil
	case int:
		return decimal.NewFromInt(int64(v)), nil
	case int64:
		return decimal.NewFromInt(v), nil
	case json.Number:
		d, err := decimal.NewFromString(v.String())
		if err != nil {
			return decimal.Zero, fmt.Errorf("%w: json.Number %q: %v", ErrParseDecimal, v.String(), err)
		}
		return d, nil
	default:
		return decimal.Zero, fmt.Errorf("%w: unsupported type %T for decimal conversion", ErrInvalidType, val)
	}
}

// parseStringFromAny safely extracts a string from interface{} values.
// Returns error if the value is not a string type.
//
// Handles the following types:
//   - nil: returns "", nil (nil is a valid empty value)
//   - string: returns the string value
func parseStringFromAny(val any) (string, error) {
	if val == nil {
		return "", nil
	}
	switch v := val.(type) {
	case string:
		return v, nil
	default:
		return "", fmt.Errorf("%w: unsupported type %T for string conversion", ErrInvalidType, val)
	}
}
