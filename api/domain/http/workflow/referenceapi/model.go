package referenceapi

import (
	"encoding/json"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/sdk/workflow"
)

// =============================================================================
// Response Types

// TriggerType represents a trigger type in API responses.
type TriggerType struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	IsActive    bool      `json:"is_active"`
}

// EntityType represents an entity type in API responses.
type EntityType struct {
	ID            uuid.UUID `json:"id"`
	Name          string    `json:"name"`
	Description   string    `json:"description"`
	IsActive      bool      `json:"is_active"`
	FrontendRoute string    `json:"frontendRoute,omitempty"`
}

// Entity represents an entity in API responses.
type Entity struct {
	ID           uuid.UUID `json:"id"`
	Name         string    `json:"name"`
	EntityTypeID uuid.UUID `json:"entity_type_id"`
	SchemaName   string    `json:"schema_name"`
	IsActive     bool      `json:"is_active"`
}

// =============================================================================
// Slice Types (implement web.Encoder)

// TriggerTypes is a slice of TriggerType for API responses.
type TriggerTypes []TriggerType

// Encode implements web.Encoder.
func (tt TriggerTypes) Encode() ([]byte, string, error) {
	data, err := json.Marshal(tt)
	return data, "application/json", err
}

// EntityTypes is a slice of EntityType for API responses.
type EntityTypes []EntityType

// Encode implements web.Encoder.
func (et EntityTypes) Encode() ([]byte, string, error) {
	data, err := json.Marshal(et)
	return data, "application/json", err
}

// Entities is a slice of Entity for API responses.
type Entities []Entity

// Encode implements web.Encoder.
func (e Entities) Encode() ([]byte, string, error) {
	data, err := json.Marshal(e)
	return data, "application/json", err
}

// ActionTypes is a slice of ActionTypeInfo for API responses.
type ActionTypes []ActionTypeInfo

// Encode implements web.Encoder.
func (at ActionTypes) Encode() ([]byte, string, error) {
	data, err := json.Marshal(at)
	return data, "application/json", err
}

// ActionTemplate represents an action template in API responses.
type ActionTemplate struct {
	ID            uuid.UUID       `json:"id"`
	Name          string          `json:"name"`
	Description   string          `json:"description"`
	ActionType    string          `json:"actionType"`
	Icon          string          `json:"icon"`
	DefaultConfig json.RawMessage `json:"defaultConfig"`
	CreatedDate   string          `json:"createdDate"`
	CreatedBy     uuid.UUID       `json:"createdBy"`
	IsActive      bool            `json:"isActive"`
	DeactivatedBy uuid.UUID       `json:"deactivatedBy"`
}

// ActionTemplates is a slice of ActionTemplate for API responses.
type ActionTemplates []ActionTemplate

// Encode implements web.Encoder.
func (at ActionTemplates) Encode() ([]byte, string, error) {
	data, err := json.Marshal(at)
	return data, "application/json", err
}

// =============================================================================
// Converter Functions

// toTriggerType converts a business trigger type to an API response.
func toTriggerType(tt workflow.TriggerType) TriggerType {
	return TriggerType{
		ID:          tt.ID,
		Name:        tt.Name,
		Description: tt.Description,
		IsActive:    tt.IsActive,
	}
}

// toTriggerTypes converts a slice of business trigger types to API responses.
func toTriggerTypes(types []workflow.TriggerType) TriggerTypes {
	resp := make(TriggerTypes, len(types))
	for i, tt := range types {
		resp[i] = toTriggerType(tt)
	}
	return resp
}

// toEntityType converts a business entity type to an API response.
func toEntityType(et workflow.EntityType) EntityType {
	return EntityType{
		ID:            et.ID,
		Name:          et.Name,
		Description:   et.Description,
		IsActive:      et.IsActive,
		FrontendRoute: et.FrontendRoute,
	}
}

// toEntityTypes converts a slice of business entity types to API responses.
func toEntityTypes(types []workflow.EntityType) EntityTypes {
	resp := make(EntityTypes, len(types))
	for i, et := range types {
		resp[i] = toEntityType(et)
	}
	return resp
}

// toEntity converts a business entity to an API response.
func toEntity(e workflow.Entity) Entity {
	return Entity{
		ID:           e.ID,
		Name:         e.Name,
		EntityTypeID: e.EntityTypeID,
		SchemaName:   e.SchemaName,
		IsActive:     e.IsActive,
	}
}

// toEntities converts a slice of business entities to API responses.
func toEntities(entities []workflow.Entity) Entities {
	resp := make(Entities, len(entities))
	for i, e := range entities {
		resp[i] = toEntity(e)
	}
	return resp
}

// toActionTemplate converts a business action template to an API response.
func toActionTemplate(at workflow.ActionTemplate) ActionTemplate {
	return ActionTemplate{
		ID:            at.ID,
		Name:          at.Name,
		Description:   at.Description,
		ActionType:    at.ActionType,
		Icon:          at.Icon,
		DefaultConfig: at.DefaultConfig,
		CreatedDate:   at.CreatedDate.Format("2006-01-02T15:04:05Z"),
		CreatedBy:     at.CreatedBy,
		IsActive:      at.IsActive,
		DeactivatedBy: at.DeactivatedBy,
	}
}

// toActionTemplates converts a slice of business action templates to API responses.
func toActionTemplates(templates []workflow.ActionTemplate) ActionTemplates {
	resp := make(ActionTemplates, len(templates))
	for i, at := range templates {
		resp[i] = toActionTemplate(at)
	}
	return resp
}
