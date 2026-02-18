package approvalrequestdb

import (
	"bytes"
	"strings"

	"github.com/timmaaaz/ichor/business/domain/workflow/approvalrequestbus"
)

func applyFilter(filter approvalrequestbus.QueryFilter, data map[string]any, buf *bytes.Buffer) {
	var wc []string

	if filter.ID != nil {
		data["id"] = *filter.ID
		wc = append(wc, "approval_request_id = :id")
	}

	if filter.ExecutionID != nil {
		data["execution_id"] = *filter.ExecutionID
		wc = append(wc, "execution_id = :execution_id")
	}

	if filter.RuleID != nil {
		data["rule_id"] = *filter.RuleID
		wc = append(wc, "rule_id = :rule_id")
	}

	if filter.Status != nil {
		data["status"] = *filter.Status
		wc = append(wc, "status = :status")
	}

	if filter.ApproverID != nil {
		data["approver_id"] = filter.ApproverID.String()
		wc = append(wc, "approvers @> CAST(ARRAY[:approver_id] AS uuid[])")
	}

	if len(wc) > 0 {
		buf.WriteString(" WHERE ")
		buf.WriteString(strings.Join(wc, " AND "))
	}
}
