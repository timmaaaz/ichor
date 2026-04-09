package directedworkapi

import (
	"encoding/json"

	"github.com/timmaaaz/ichor/app/domain/floor/directedworkapp"
)

// Response is the JSON envelope for the endpoint. WorkItem is an explicit
// pointer so the JSON body reads { "work_item": null } when nothing is
// directed — see judgment-call #1 in the spec (we do not return HTTP 204).
type Response struct {
	WorkItem *directedworkapp.WorkItem `json:"work_item"`
}

// Encode implements web.Encoder for the standard JSON path.
func (r Response) Encode() ([]byte, string, error) {
	data, err := json.Marshal(r)
	return data, "application/json", err
}
