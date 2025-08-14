package tablebuilder

import (
	"fmt"
	"reflect"
	"strconv"
	"time"
)

// Converter handles type conversions for dynamic data
type Converter struct{}

// NewConverter creates a new type converter
func NewConverter() *Converter {
	return &Converter{}
}

// Convert converts a value to the specified type
func (c *Converter) Convert(value interface{}, targetType string) (interface{}, error) {
	if value == nil {
		return nil, nil
	}

	switch targetType {
	case "string":
		return c.ToString(value), nil
	case "int", "integer":
		return c.ToInt(value)
	case "float", "float64", "number":
		return c.ToFloat64(value)
	case "bool", "boolean":
		return c.ToBool(value)
	case "uuid":
		return c.ToUUID(value)
	case "time", "timestamp", "datetime":
		return c.ToTime(value)
	case "date":
		return c.ToDate(value)
	case "json":
		return value, nil // Already in the right format
	default:
		return value, nil // Return as-is for unknown types
	}
}

// ToString converts a value to string
func (c *Converter) ToString(value interface{}) string {
	if value == nil {
		return ""
	}

	switch v := value.(type) {
	case string:
		return v
	case []byte:
		return string(v)
	case fmt.Stringer:
		return v.String()
	default:
		return fmt.Sprintf("%v", value)
	}
}

// ToInt converts a value to int
func (c *Converter) ToInt(value interface{}) (int, error) {
	switch v := value.(type) {
	case int:
		return v, nil
	case int32:
		return int(v), nil
	case int64:
		return int(v), nil
	case float32:
		return int(v), nil
	case float64:
		return int(v), nil
	case string:
		return strconv.Atoi(v)
	case []byte:
		return strconv.Atoi(string(v))
	default:
		return 0, fmt.Errorf("cannot convert %T to int", value)
	}
}

// ToFloat64 converts a value to float64
func (c *Converter) ToFloat64(value interface{}) (float64, error) {
	switch v := value.(type) {
	case float64:
		return v, nil
	case float32:
		return float64(v), nil
	case int:
		return float64(v), nil
	case int32:
		return float64(v), nil
	case int64:
		return float64(v), nil
	case string:
		return strconv.ParseFloat(v, 64)
	case []byte:
		return strconv.ParseFloat(string(v), 64)
	default:
		return 0, fmt.Errorf("cannot convert %T to float64", value)
	}
}

// ToBool converts a value to bool
func (c *Converter) ToBool(value interface{}) (bool, error) {
	switch v := value.(type) {
	case bool:
		return v, nil
	case int:
		return v != 0, nil
	case string:
		switch v {
		case "true", "TRUE", "True", "1", "yes", "YES", "Yes", "on", "ON":
			return true, nil
		case "false", "FALSE", "False", "0", "no", "NO", "No", "off", "OFF":
			return false, nil
		default:
			return false, fmt.Errorf("cannot convert string %q to bool", v)
		}
	case []byte:
		return c.ToBool(string(v))
	default:
		return false, fmt.Errorf("cannot convert %T to bool", value)
	}
}

// ToUUID converts a value to UUID string
func (c *Converter) ToUUID(value interface{}) (string, error) {
	switch v := value.(type) {
	case string:
		// Validate UUID format
		if len(v) == 36 { // Basic UUID length check
			return v, nil
		}
		return "", fmt.Errorf("invalid UUID format: %s", v)
	case []byte:
		return c.ToUUID(string(v))
	default:
		return "", fmt.Errorf("cannot convert %T to UUID", value)
	}
}

// ToTime converts a value to time.Time
func (c *Converter) ToTime(value interface{}) (time.Time, error) {
	switch v := value.(type) {
	case time.Time:
		return v, nil
	case string:
		// Try multiple formats
		formats := []string{
			time.RFC3339,
			time.RFC3339Nano,
			"2006-01-02 15:04:05",
			"2006-01-02T15:04:05Z",
			"2006-01-02",
		}

		for _, format := range formats {
			if t, err := time.Parse(format, v); err == nil {
				return t, nil
			}
		}
		return time.Time{}, fmt.Errorf("cannot parse time string: %s", v)
	case []byte:
		return c.ToTime(string(v))
	default:
		return time.Time{}, fmt.Errorf("cannot convert %T to time.Time", value)
	}
}

// ToDate converts a value to date (time with zero hour)
func (c *Converter) ToDate(value interface{}) (time.Time, error) {
	t, err := c.ToTime(value)
	if err != nil {
		return time.Time{}, err
	}

	// Zero out the time portion
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location()), nil
}

// ConvertMap converts all values in a map according to a type map
func (c *Converter) ConvertMap(data map[string]interface{}, typeMap map[string]string) map[string]interface{} {
	result := make(map[string]interface{})

	for key, value := range data {
		if targetType, ok := typeMap[key]; ok {
			converted, err := c.Convert(value, targetType)
			if err != nil {
				// Log error but keep original value
				result[key] = value
			} else {
				result[key] = converted
			}
		} else {
			result[key] = value
		}
	}

	return result
}

// ConvertSlice converts a slice of values to the target type
func (c *Converter) ConvertSlice(values []interface{}, targetType string) ([]interface{}, error) {
	result := make([]interface{}, len(values))

	for i, value := range values {
		converted, err := c.Convert(value, targetType)
		if err != nil {
			return nil, fmt.Errorf("convert slice element %d: %w", i, err)
		}
		result[i] = converted
	}

	return result, nil
}

// InferType attempts to infer the type of a value
func (c *Converter) InferType(value interface{}) string {
	if value == nil {
		return "null"
	}

	v := reflect.ValueOf(value)
	switch v.Kind() {
	case reflect.String:
		// Check if it's a special string type
		str := v.String()
		if _, err := time.Parse(time.RFC3339, str); err == nil {
			return "datetime"
		}
		if _, err := time.Parse("2006-01-02", str); err == nil {
			return "date"
		}
		if len(str) == 36 && containsUUIDChars(str) {
			return "uuid"
		}
		return "string"
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return "int"
	case reflect.Float32, reflect.Float64:
		return "float"
	case reflect.Bool:
		return "bool"
	case reflect.Slice, reflect.Array:
		return "array"
	case reflect.Map, reflect.Struct:
		return "object"
	default:
		return "unknown"
	}
}

// Helper function to check if string contains UUID-like characters
func containsUUIDChars(s string) bool {
	// Simple check for UUID format (8-4-4-4-12)
	if len(s) != 36 {
		return false
	}
	for i, ch := range s {
		if i == 8 || i == 13 || i == 18 || i == 23 {
			if ch != '-' {
				return false
			}
		} else if !((ch >= '0' && ch <= '9') || (ch >= 'a' && ch <= 'f') || (ch >= 'A' && ch <= 'F')) {
			return false
		}
	}
	return true
}
