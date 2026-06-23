package executionapp

import (
	"encoding/json"

	"github.com/google/uuid"
)

// RerunResponse is returned when an execution is re-run.
type RerunResponse struct {
	OriginalExecutionID uuid.UUID `json:"original_execution_id"`
	NewExecutionID      uuid.UUID `json:"new_execution_id"`
}

// Encode implements web.Encoder.
func (r RerunResponse) Encode() ([]byte, string, error) {
	data, err := json.Marshal(r)
	return data, "application/json", err
}
