package approvalapi

import (
	"encoding/json"
	"time"

	"github.com/timmaaaz/ichor/business/domain/workflow/approvalrequestbus"
)

// Approval represents the API response model for an approval request.
type Approval struct {
	ID               string   `json:"id"`
	ExecutionID      string   `json:"executionId"`
	RuleID           string   `json:"ruleId"`
	ActionName       string   `json:"actionName"`
	Approvers        []string `json:"approvers"`
	ApprovalType     string   `json:"approvalType"`
	Status           string   `json:"status"`
	TimeoutHours     int      `json:"timeoutHours"`
	ApprovalMessage  string   `json:"approvalMessage,omitempty"`
	ResolvedBy       string   `json:"resolvedBy,omitempty"`
	ResolutionReason string   `json:"resolutionReason,omitempty"`
	CreatedDate      string   `json:"createdDate"`
	ResolvedDate     string   `json:"resolvedDate,omitempty"`
}

// Encode implements the web.Encoder interface.
func (app Approval) Encode() ([]byte, string, error) {
	data, err := json.Marshal(app)
	return data, "application/json", err
}

// toAppApproval converts a business approval request to an API approval.
func toAppApproval(bus approvalrequestbus.ApprovalRequest) Approval {
	approvers := make([]string, len(bus.Approvers))
	for i, id := range bus.Approvers {
		approvers[i] = id.String()
	}

	app := Approval{
		ID:           bus.ID.String(),
		ExecutionID:  bus.ExecutionID.String(),
		RuleID:       bus.RuleID.String(),
		ActionName:   bus.ActionName,
		Approvers:    approvers,
		ApprovalType: bus.ApprovalType,
		Status:       bus.Status,
		TimeoutHours: bus.TimeoutHours,
		CreatedDate:  bus.CreatedDate.Format(time.RFC3339),
	}

	if bus.ApprovalMessage != "" {
		app.ApprovalMessage = bus.ApprovalMessage
	}
	if bus.ResolvedBy != nil {
		app.ResolvedBy = bus.ResolvedBy.String()
	}
	if bus.ResolutionReason != "" {
		app.ResolutionReason = bus.ResolutionReason
	}
	if bus.ResolvedDate != nil {
		app.ResolvedDate = bus.ResolvedDate.Format(time.RFC3339)
	}

	return app
}

// toAppApprovals converts a slice of business approval requests to API models.
func toAppApprovals(bus []approvalrequestbus.ApprovalRequest) []Approval {
	app := make([]Approval, len(bus))
	for i, v := range bus {
		app[i] = toAppApproval(v)
	}
	return app
}

// ResolveRequest represents the request body for resolving an approval.
type ResolveRequest struct {
	Resolution string `json:"resolution"` // "approved" or "rejected"
	Reason     string `json:"reason"`
}

// Decode implements the web.Decoder interface.
func (r *ResolveRequest) Decode(data []byte) error {
	return json.Unmarshal(data, r)
}

// QueryParams holds query parameters for approval request queries.
type QueryParams struct {
	Page    string
	Rows    string
	OrderBy string
	Status  string
}
