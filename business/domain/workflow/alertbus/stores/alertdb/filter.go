package alertdb

import (
	"bytes"
	"strings"

	"github.com/timmaaaz/ichor/business/domain/workflow/alertbus"
)

// hasSeverities returns true if the filter has multi-value severities that require IN clause.
func hasSeverities(filter alertbus.QueryFilter) bool {
	return len(filter.Severities) > 0
}

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

	if len(filter.Severities) > 0 {
		data["severities"] = filter.Severities
		wc = append(wc, "severity IN (:severities)")
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

	if filter.CreatedAfter != nil {
		data["created_after"] = *filter.CreatedAfter
		wc = append(wc, "created_date >= :created_after")
	}

	if filter.CreatedBefore != nil {
		data["created_before"] = *filter.CreatedBefore
		wc = append(wc, "created_date <= :created_before")
	}

	if filter.Status != nil && *filter.Status == alertbus.StatusActive {
		wc = append(wc, "(expires_date IS NULL OR expires_date > NOW())")
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

	if len(filter.Severities) > 0 {
		data["severities"] = filter.Severities
		wc = append(wc, "a.severity IN (:severities)")
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

	if filter.CreatedAfter != nil {
		data["created_after"] = *filter.CreatedAfter
		wc = append(wc, "a.created_date >= :created_after")
	}

	if filter.CreatedBefore != nil {
		data["created_before"] = *filter.CreatedBefore
		wc = append(wc, "a.created_date <= :created_before")
	}

	if filter.Status != nil && *filter.Status == alertbus.StatusActive {
		wc = append(wc, "(a.expires_date IS NULL OR a.expires_date > NOW())")
	}

	if len(wc) > 0 {
		buf.WriteString(" AND ")
		buf.WriteString(strings.Join(wc, " AND "))
	}
}
