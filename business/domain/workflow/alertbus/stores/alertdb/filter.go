package alertdb

import (
	"bytes"
	"strings"

	"github.com/timmaaaz/ichor/business/domain/workflow/alertbus"
)

func applyFilter(filter alertbus.QueryFilter, data map[string]any, buf *bytes.Buffer) {
	var wc []string

	if filter.ID != nil {
		data["id"] = *filter.ID
		wc = append(wc, "id = :id")
	}

	if filter.AlertType != nil {
		data["alert_type"] = *filter.AlertType
		wc = append(wc, "alert_type = :alert_type")
	}

	if filter.Severity != nil {
		data["severity"] = *filter.Severity
		wc = append(wc, "severity = :severity")
	}

	if filter.Status != nil {
		data["status"] = *filter.Status
		wc = append(wc, "status = :status")
	}

	if filter.SourceEntityName != nil {
		data["source_entity_name"] = *filter.SourceEntityName
		wc = append(wc, "source_entity_name = :source_entity_name")
	}

	if filter.SourceEntityID != nil {
		data["source_entity_id"] = filter.SourceEntityID.String()
		wc = append(wc, "source_entity_id = :source_entity_id")
	}

	if filter.SourceRuleID != nil {
		data["source_rule_id"] = filter.SourceRuleID.String()
		wc = append(wc, "source_rule_id = :source_rule_id")
	}

	if len(wc) > 0 {
		buf.WriteString(" WHERE ")
		buf.WriteString(strings.Join(wc, " AND "))
	}
}

func applyFilterWithJoin(filter alertbus.QueryFilter, data map[string]any, buf *bytes.Buffer) {
	var wc []string

	if filter.ID != nil {
		data["id"] = *filter.ID
		wc = append(wc, "a.id = :id")
	}

	if filter.AlertType != nil {
		data["alert_type"] = *filter.AlertType
		wc = append(wc, "a.alert_type = :alert_type")
	}

	if filter.Severity != nil {
		data["severity"] = *filter.Severity
		wc = append(wc, "a.severity = :severity")
	}

	if filter.Status != nil {
		data["status"] = *filter.Status
		wc = append(wc, "a.status = :status")
	}

	if filter.SourceEntityName != nil {
		data["source_entity_name"] = *filter.SourceEntityName
		wc = append(wc, "a.source_entity_name = :source_entity_name")
	}

	if filter.SourceEntityID != nil {
		data["source_entity_id"] = filter.SourceEntityID.String()
		wc = append(wc, "a.source_entity_id = :source_entity_id")
	}

	if filter.SourceRuleID != nil {
		data["source_rule_id"] = filter.SourceRuleID.String()
		wc = append(wc, "a.source_rule_id = :source_rule_id")
	}

	if len(wc) > 0 {
		buf.WriteString(" AND ")
		buf.WriteString(strings.Join(wc, " AND "))
	}
}
