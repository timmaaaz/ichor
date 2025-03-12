package testing

import (
	"reflect"
	"strings"
	"time"
)

// MapToStruct converts a map[string]interface{} to a struct of type T
// It matches map keys to struct field names (case-insensitive)
func MapToStruct[T any](data map[string]interface{}) (T, error) {
	var result T
	resultValue := reflect.ValueOf(&result).Elem()
	resultType := resultValue.Type()

	// Create a map of lowercase field names to their indices
	fieldMap := make(map[string]int)
	for i := 0; i < resultType.NumField(); i++ {
		field := resultType.Field(i)

		// Check for json tag first
		tagName := field.Tag.Get("json")
		if tagName != "" {
			// Handle commas in json tag (e.g. "name,omitempty")
			tagName = strings.Split(tagName, ",")[0]
			fieldMap[strings.ToLower(tagName)] = i
		} else {
			// If no json tag, use field name
			fieldMap[strings.ToLower(field.Name)] = i
		}

		// Also map standard field name
		fieldMap[strings.ToLower(field.Name)] = i
	}

	// Populate struct fields from map
	for key, value := range data {
		if value == nil {
			continue
		}

		fieldIndex, ok := fieldMap[strings.ToLower(key)]
		if !ok {
			continue // Skip if field not found in struct
		}

		field := resultValue.Field(fieldIndex)
		if !field.CanSet() {
			continue // Skip if field cannot be set (unexported)
		}

		// Handle different field types
		switch field.Kind() {
		case reflect.String:
			if strValue, ok := value.(string); ok {
				field.SetString(strValue)
			}
		case reflect.Bool:
			if boolValue, ok := value.(bool); ok {
				field.SetBool(boolValue)
			}
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			switch v := value.(type) {
			case int:
				field.SetInt(int64(v))
			case int64:
				field.SetInt(v)
			case int32:
				field.SetInt(int64(v))
			case float64: // JSON often unmarshals integers as float64
				field.SetInt(int64(v))
			}
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			switch v := value.(type) {
			case uint:
				field.SetUint(uint64(v))
			case uint64:
				field.SetUint(v)
			case uint32:
				field.SetUint(uint64(v))
			case int:
				field.SetUint(uint64(v))
			case float64: // JSON often unmarshals integers as float64
				field.SetUint(uint64(v))
			}
		case reflect.Float32, reflect.Float64:
			if floatValue, ok := value.(float64); ok {
				field.SetFloat(floatValue)
			}
		case reflect.Struct:
			// Special handling for time.Time
			if field.Type() == reflect.TypeOf(time.Time{}) {
				if timeValue, ok := value.(time.Time); ok {
					field.Set(reflect.ValueOf(timeValue))
				}
			}
		case reflect.Interface:
			// For interface{} fields, just set the value directly
			field.Set(reflect.ValueOf(value))
		case reflect.Ptr:
			// Handle pointer types
			if value != nil {
				// Create new pointer of the right type
				ptrValue := reflect.New(field.Type().Elem())

				// If the pointer is to a primitive like string
				if field.Type().Elem().Kind() == reflect.String {
					if strValue, ok := value.(string); ok {
						ptrValue.Elem().SetString(strValue)
					}
				} else if field.Type().Elem().Kind() == reflect.Bool {
					if boolValue, ok := value.(bool); ok {
						ptrValue.Elem().SetBool(boolValue)
					}
				} else if field.Type().Elem().Kind() == reflect.Int {
					if intValue, ok := value.(int); ok {
						ptrValue.Elem().SetInt(int64(intValue))
					} else if floatValue, ok := value.(float64); ok {
						ptrValue.Elem().SetInt(int64(floatValue))
					}
				}

				field.Set(ptrValue)
			}
		}
	}

	return result, nil
}
