package chatapi

import (
	"encoding/json"

	"github.com/timmaaaz/ichor/app/sdk/errs"
)

// ChatRequest is the POST body for /v1/agent/chat.
type ChatRequest struct {
	Message        string          `json:"message" validate:"required"`
	ContextType    string          `json:"context_type" validate:"required"`
	Context        json.RawMessage `json:"context,omitempty"`
	ConversationID string          `json:"conversation_id,omitempty"`
}

// Decode implements the Decoder interface.
func (r *ChatRequest) Decode(data []byte) error {
	return json.Unmarshal(data, r)
}

// Validate checks required fields.
func (r ChatRequest) Validate() error {
	if err := errs.Check(r); err != nil {
		return errs.Newf(errs.InvalidArgument, "validate: %s", err)
	}
	return nil
}
