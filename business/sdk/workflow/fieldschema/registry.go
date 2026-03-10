// Package fieldschema provides a static registry of known enum field schemas
// for workflow entity fields. This powers the field discovery endpoint used
// by the workflow builder to present valid values for trigger field_conditions.
package fieldschema

import "encoding/json"

// FieldSchema describes a single field on an entity that has known discrete values.
type FieldSchema struct {
	Name        string   `json:"name"`
	Type        string   `json:"type"`             // "enum", "string", "uuid", etc.
	Values      []string `json:"values,omitempty"` // only for type="enum"
	Description string   `json:"description,omitempty"`
}

// EntitySchema groups the field schemas for one entity.
type EntitySchema struct {
	Entity string        `json:"entity"`
	Fields []FieldSchema `json:"fields"`
}

// Encode implements the web.Encoder interface so EntitySchema can be returned
// directly from HTTP handlers.
func (e EntitySchema) Encode() ([]byte, string, error) {
	data, err := json.Marshal(e)
	return data, "application/json", err
}

// KnownEnumFields maps DB entity names (schema.table format) to their registered
// enum field schemas. This is the authoritative list of discoverable enum values
// for workflow trigger field_conditions.
//
// NOTE: When adding new status/enum fields to any bus model, also add an entry here.
var KnownEnumFields = map[string][]FieldSchema{
	"inventory.put_away_tasks": {
		{Name: "status", Type: "enum", Values: []string{"pending", "in_progress", "completed", "cancelled"}, Description: "Lifecycle state of the putaway task"},
	},
	"inventory.inventory_adjustments": {
		{Name: "approval_status", Type: "enum", Values: []string{"pending", "approved", "rejected"}, Description: "Approval state of the inventory adjustment"},
	},
	"inventory.lot_trackings": {
		{Name: "quality_status", Type: "enum", Values: []string{"good", "on_hold", "quarantined", "released", "expired"}, Description: "Quality control state of the lot"},
	},
	"workflow.alerts": {
		{Name: "status", Type: "enum", Values: []string{"active", "acknowledged", "dismissed", "resolved"}, Description: "Alert lifecycle state"},
		{Name: "severity", Type: "enum", Values: []string{"low", "medium", "high", "critical"}, Description: "Alert severity level"},
	},
	"workflow.approval_requests": {
		{Name: "status", Type: "enum", Values: []string{"pending", "approved", "rejected", "timed_out", "expired"}, Description: "Approval request resolution state"},
		{Name: "approval_type", Type: "enum", Values: []string{"any", "all", "majority"}, Description: "Required approval quorum type"},
	},
}

// GetEntitySchema returns the field schemas for the given entity.
// Returns (fields, true) if the entity has registered enum fields.
// Returns (nil, false) if the entity is not in the registry.
func GetEntitySchema(entity string) ([]FieldSchema, bool) {
	fields, ok := KnownEnumFields[entity]
	return fields, ok
}
