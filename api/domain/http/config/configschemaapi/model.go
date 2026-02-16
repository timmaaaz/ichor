package configschemaapi

import (
	"encoding/json"
)

// SchemaInfo describes a config schema and its purpose.
type SchemaInfo struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Schema      json.RawMessage `json:"schema"`
}

// Encode implements web.Encoder for a single SchemaInfo.
func (s SchemaInfo) Encode() ([]byte, string, error) {
	data, err := json.Marshal(s)
	return data, "application/json", err
}

// ContentTypeInfo describes a valid content type for page content blocks.
type ContentTypeInfo struct {
	Type               string `json:"type"`
	Name               string `json:"name"`
	Description        string `json:"description"`
	RequiresTableConfig bool  `json:"requires_table_config"`
	RequiresForm       bool   `json:"requires_form"`
	SupportsChildren   bool   `json:"supports_children"`
}

// ContentTypes is a slice of ContentTypeInfo for API responses.
type ContentTypes []ContentTypeInfo

// Encode implements web.Encoder.
func (ct ContentTypes) Encode() ([]byte, string, error) {
	data, err := json.Marshal(ct)
	return data, "application/json", err
}

// PageActionTypeField describes a field on a page action type.
type PageActionTypeField struct {
	Name        string   `json:"name"`
	Type        string   `json:"type"`
	Required    bool     `json:"required"`
	Description string   `json:"description"`
	Enum        []string `json:"enum,omitempty"`
}

// PageActionTypeInfo describes a valid page action type.
type PageActionTypeInfo struct {
	Type        string                `json:"type"`
	Name        string                `json:"name"`
	Description string                `json:"description"`
	Fields      []PageActionTypeField `json:"fields"`
}

// PageActionTypes is a slice of PageActionTypeInfo for API responses.
type PageActionTypes []PageActionTypeInfo

// Encode implements web.Encoder.
func (pat PageActionTypes) Encode() ([]byte, string, error) {
	data, err := json.Marshal(pat)
	return data, "application/json", err
}
