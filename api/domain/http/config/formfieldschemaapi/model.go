package formfieldschemaapi

import (
	"encoding/json"
)

// FieldTypeInfo describes a form field type and its config JSON schema.
type FieldTypeInfo struct {
	Type         string          `json:"type"`
	Name         string          `json:"name"`
	Description  string          `json:"description"`
	ConfigSchema json.RawMessage `json:"config_schema"`
}

// Encode implements web.Encoder for a single FieldTypeInfo.
func (f FieldTypeInfo) Encode() ([]byte, string, error) {
	data, err := json.Marshal(f)
	return data, "application/json", err
}

// FieldTypes is a slice of FieldTypeInfo for API responses.
type FieldTypes []FieldTypeInfo

// Encode implements web.Encoder.
func (ft FieldTypes) Encode() ([]byte, string, error) {
	data, err := json.Marshal(ft)
	return data, "application/json", err
}
