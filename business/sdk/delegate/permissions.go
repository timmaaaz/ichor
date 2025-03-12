package delegate

import (
	"encoding/json"
)

// For field visibility filtering
const (
	DomainPermission = "permission"
	ActionFilter     = "filter_fields"
)

// FieldFilterData represents data for field filtering across domains
type FieldFilterData struct {
	TableName string                 `json:"table_name"`
	Data      map[string]interface{} `json:"data"`
	Result    map[string]interface{} `json:"result,omitempty"`
}

// MarshalFieldFilterData converts FieldFilterData to raw params
func MarshalFieldFilterData(data FieldFilterData) ([]byte, error) {
	return json.Marshal(data)
}

// UnmarshalFieldFilterData extracts FieldFilterData from raw params
func UnmarshalFieldFilterData(raw []byte) (FieldFilterData, error) {
	var data FieldFilterData
	err := json.Unmarshal(raw, &data)
	return data, err
}
