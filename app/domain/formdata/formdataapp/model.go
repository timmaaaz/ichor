package formdataapp

import (
	"encoding/json"

	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/app/sdk/formdataregistry"
)

// FormDataRequest represents the incoming upsert request.
//
// The request structure uses explicit operations metadata separate from
// the actual entity data to avoid polluting business data with control fields.
//
// Example:
//
//	{
//	  "operations": {
//	    "users": {"operation": "create", "order": 1},
//	    "addresses": {"operation": "create", "order": 2}
//	  },
//	  "data": {
//	    "users": {"first_name": "John", "last_name": "Doe"},
//	    "addresses": {"user_id": "{{users.id}}", "street": "123 Main"}
//	  }
//	}
type FormDataRequest struct {
	Operations map[string]OperationMeta   `json:"operations" validate:"required,min=1"`
	Data       map[string]json.RawMessage `json:"data" validate:"required,min=1"`
}

// OperationMeta defines metadata for a single entity operation.
type OperationMeta struct {
	Operation formdataregistry.EntityOperation `json:"operation" validate:"required,oneof=create update"`
	Order     int                               `json:"order" validate:"required,min=1"`
}

// FormDataResponse is the successful response structure.
type FormDataResponse struct {
	Success bool                   `json:"success"`
	Results map[string]interface{} `json:"results"`
}

// ExecutionStep represents a single operation to execute in the plan.
type ExecutionStep struct {
	EntityName string
	Operation  formdataregistry.EntityOperation
	Order      int
	Registry   *formdataregistry.EntityRegistration
}

// Decode implements the decoder interface.
func (r *FormDataRequest) Decode(data []byte) error {
	return json.Unmarshal(data, r)
}

// Validate checks request validity.
func (r FormDataRequest) Validate() error {
	if err := errs.Check(r); err != nil {
		return errs.Newf(errs.InvalidArgument, "validate: %s", err)
	}

	// Ensure operations and data have matching keys
	for entityName := range r.Operations {
		if _, exists := r.Data[entityName]; !exists {
			return errs.Newf(errs.InvalidArgument, "entity %s in operations but missing from data", entityName)
		}
	}

	for entityName := range r.Data {
		if _, exists := r.Operations[entityName]; !exists {
			return errs.Newf(errs.InvalidArgument, "entity %s in data but missing from operations", entityName)
		}
	}

	// Validate operation types
	for entityName, meta := range r.Operations {
		if !meta.Operation.IsValid() {
			return errs.Newf(errs.InvalidArgument, "entity %s has invalid operation: %s", entityName, meta.Operation)
		}
	}

	return nil
}

// Encode implements the encoder interface.
func (r FormDataResponse) Encode() ([]byte, string, error) {
	data, err := json.Marshal(r)
	return data, "application/json", err
}
