package workflowdb

import (
	"bytes"
	"strings"

	"github.com/timmaaaz/ichor/business/sdk/workflow"
)

func applyAutomationRuleFilter(filter workflow.AutomationRuleFilter, data map[string]any, buf *bytes.Buffer) {
	var wc []string

	if filter.ID != nil {
		data["id"] = filter.ID.String()
		wc = append(wc, "ar.id = :id")
	}

	if filter.Name != nil {
		data["name"] = "%" + *filter.Name + "%"
		wc = append(wc, "ar.name ILIKE :name")
	}

	if filter.IsActive != nil {
		data["is_active"] = *filter.IsActive
		wc = append(wc, "ar.is_active = :is_active")
	}

	if filter.EntityID != nil {
		data["entity_id"] = filter.EntityID.String()
		wc = append(wc, "ar.entity_id = :entity_id")
	}

	if filter.EntityTypeID != nil {
		data["entity_type_id"] = filter.EntityTypeID.String()
		wc = append(wc, "ar.entity_type_id = :entity_type_id")
	}

	if filter.TriggerTypeID != nil {
		data["trigger_type_id"] = filter.TriggerTypeID.String()
		wc = append(wc, "ar.trigger_type_id = :trigger_type_id")
	}

	if filter.CreatedBy != nil {
		data["created_by"] = filter.CreatedBy.String()
		wc = append(wc, "ar.created_by = :created_by")
	}

	if len(wc) > 0 {
		buf.WriteString(" WHERE ")
		buf.WriteString(strings.Join(wc, " AND "))
	}
}
