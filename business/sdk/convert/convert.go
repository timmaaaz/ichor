// Package Convert allows us to convert one type to another
package convert

import (
	"fmt"
	"reflect"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/foundation/timeutil"
)

// PopulateSameTypes populates a struct from another struct, in the business layer
// This function should be called when the structs have the same field type and names
// Dest must be a pointer
func PopulateSameTypes(src, dest interface{}) error {
	if reflect.TypeOf(src).Kind() != reflect.Struct {
		return fmt.Errorf("src must be a struct")
	}
	if reflect.TypeOf(dest).Kind() != reflect.Ptr || reflect.TypeOf(dest).Elem().Kind() != reflect.Struct {
		return fmt.Errorf("dest must be a pointer to a struct")
	}

	val := reflect.ValueOf(src)
	typ := reflect.TypeOf(src)

	dVal := reflect.ValueOf(dest)
	if dVal.Kind() != reflect.Ptr || dVal.IsNil() {
		panic("dest must be a non-nil pointer")
	}

	dVal = dVal.Elem()

	for i := 0; i < typ.NumField(); i++ {

		// for each field we want to get the type name and value from the source
		field := typ.Field(i)
		value := val.Field(i)

		// we check if the source value exists
		if value.IsValid() && !value.IsNil() {

			// we find the field in the destination
			dField := dVal.FieldByName(field.Name)

			// we check if the value is a pointer, if it is we take the literal
			// the value should always be a string in this method
			if value.Kind() == reflect.Ptr {
				value = value.Elem()
			}

			// if the destination field can be set then we set the value
			if dField.IsValid() && dField.CanSet() {
				if dField.Type() == value.Type() {
					dField.Set(value)
				} else if value.Type().ConvertibleTo(dField.Type()) {
					dField.Set(value.Convert(dField.Type()))
				}
			}
		}
	}
	return nil

}

// PopulateTypesFromStrings populates the typed destination struct from a
// src struct that contains strings
// Dest must be a pointer
func PopulateTypesFromStrings(src, dest interface{}) error {
	if reflect.TypeOf(src).Kind() != reflect.Struct {
		return fmt.Errorf("src must be a struct")
	}
	if reflect.TypeOf(dest).Kind() != reflect.Ptr || reflect.TypeOf(dest).Elem().Kind() != reflect.Struct {
		return fmt.Errorf("dest must be a pointer to a struct")
	}

	val := reflect.ValueOf(src)
	typ := reflect.TypeOf(src)

	dVal := reflect.ValueOf(dest)
	if dVal.Kind() != reflect.Ptr || dVal.IsNil() {
		panic("dest must be a non-nil pointer")
	}

	dVal = dVal.Elem()

	for i := 0; i < typ.NumField(); i++ {

		// we get the field name and value
		field := typ.Field(i)
		value := val.Field(i)

		// we check if the value is a pointer, if it is we take the literal
		// the value should always be a string in this method
		if value.Kind() == reflect.Ptr {
			value = value.Elem()
		}

		// we make sure the value is valid and not empty
		if value.IsValid() {
			// Check if we should skip this field based on its zero value
			shouldProcess := false

			switch value.Kind() {
			case reflect.String:
				shouldProcess = value.String() != ""
			case reflect.Bool:
				// Always process booleans (both true and false are valid)
				shouldProcess = true
			default:
				// For other types, check if it's not the zero value
				shouldProcess = !value.IsZero()
			}

			if shouldProcess {
				// get the destination field
				dField := dVal.FieldByName(field.Name)

				if dField.Kind() != reflect.Ptr {
					err := populateStructFieldLiteral(&dField, value)
					if err != nil {
						return err
					}
				} else {
					err := populateStructFieldPointer(&dField, value)
					if err != nil {
						return err
					}
				}
			}
		}
	}
	return nil
}

func populateStructFieldPointer(dField *reflect.Value, value reflect.Value) error {
	// Ensure the pointer is valid and can be set
	if dField.IsValid() && dField.Kind() == reflect.Ptr {
		// If the pointer is nil, initialize it
		if dField.IsNil() {
			newValue := reflect.New(dField.Type().Elem()) // Create a new instance of the underlying type
			dField.Set(newValue)                          // Set the pointer to the new instance
		}
		elem := dField.Elem()

		// Make sure the underlying value can be set
		if elem.IsValid() && elem.CanSet() {
			switch elem.Type().Kind() {
			case reflect.String:
				elem.SetString(value.String())
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				i, err := strconv.ParseInt(value.String(), 10, 64)
				if err == nil {
					elem.SetInt(i)
				}
			case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
				u, err := strconv.ParseUint(value.String(), 10, 64)
				if err == nil {
					elem.SetUint(u)
				}
			case reflect.Float32, reflect.Float64:
				f, err := strconv.ParseFloat(value.String(), 64)
				if err == nil {
					elem.SetFloat(f)
				}
			case reflect.Bool:
				// Check if source value is already a bool
				if value.Kind() == reflect.Bool {
					elem.SetBool(value.Bool())
				} else {
					b, err := strconv.ParseBool(value.String())
					if err == nil {
						elem.SetBool(b)
					}
				}
			default:
				if elem.Type() == reflect.TypeOf(uuid.UUID{}) {
					uuidStr := value.String()
					uuidVal, err := uuid.Parse(uuidStr)
					if err != nil {
						return fmt.Errorf("failed to parse UUID: %w", err)
					}
					elem.Set(reflect.ValueOf(uuidVal))
				} else if elem.Type() == reflect.TypeOf(time.Time{}) {
					t, err := time.Parse(timeutil.FORMAT, value.String())
					if err != nil {
						return fmt.Errorf("failed to parse time: %w", err)
					}
					elem.Set(reflect.ValueOf(t))
				}
			}
		} else {
			return fmt.Errorf("field cannot be set")
		}
	} else {
		return fmt.Errorf("field is not a valid, non-nil pointer")
	}

	return nil
}

func populateStructFieldLiteral(dField *reflect.Value, value reflect.Value) error {
	// make sure that the destination field can be populated
	if dField.IsValid() && dField.CanSet() {
		switch dField.Type().Kind() {
		case reflect.String:
			dField.SetString(value.String())
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			i, err := strconv.ParseInt(value.String(), 10, 64)
			if err == nil {
				dField.SetInt(i)
			}
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			u, err := strconv.ParseUint(value.String(), 10, 64)
			if err == nil {
				dField.SetUint(u)
			}
		case reflect.Float32, reflect.Float64:
			f, err := strconv.ParseFloat(value.String(), 64)
			if err == nil {
				dField.SetFloat(f)
			}
		case reflect.Bool:
			// Check if source value is already a bool
			if value.Kind() == reflect.Bool {
				dField.SetBool(value.Bool())
			} else {
				b, err := strconv.ParseBool(value.String())
				if err == nil {
					dField.SetBool(b)
				}
			}
		default:
			// this area is for custom types (not defined in the reflect package)
			// we can set custom options per type
			if dField.Type() == reflect.TypeOf(uuid.UUID{}) {
				uuidStr := value.String()
				uuidVal, err := uuid.Parse(uuidStr)
				if err != nil {
					return fmt.Errorf("failed to parse UUID: %w", err)
				}
				dField.Set(reflect.ValueOf(uuidVal))
			} else if dField.Type() == reflect.TypeOf(time.Time{}) {
				t, err := time.Parse(timeutil.FORMAT, value.String())
				if err != nil {
					return fmt.Errorf("failed to parse time: %w", err)
				}
				dField.Set(reflect.ValueOf(t))
			}
		}
	}

	return nil
}
