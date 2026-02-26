package alertapi

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/workflow/alertbus"
)

// validSeverities contains the allowed severity values.
var validSeverities = map[string]bool{
	alertbus.SeverityLow:      true,
	alertbus.SeverityMedium:   true,
	alertbus.SeverityHigh:     true,
	alertbus.SeverityCritical: true,
}

func parseQueryParams(r *http.Request) QueryParams {
	values := r.URL.Query()

	return QueryParams{
		Page:             values.Get("page"),
		Rows:             values.Get("rows"),
		OrderBy:          values.Get("orderBy"),
		ID:               values.Get("id"),
		AlertType:        values.Get("alertType"),
		Severity:         values.Get("severity"),
		Status:           values.Get("status"),
		SourceEntityName: values.Get("sourceEntityName"),
		SourceEntityID:   values.Get("sourceEntityId"),
		SourceRuleID:     values.Get("sourceRuleId"),
		CreatedAfter:     values.Get("createdAfter"),
		CreatedBefore:    values.Get("createdBefore"),
	}
}

func parseFilter(qp QueryParams) (alertbus.QueryFilter, error) {
	var filter alertbus.QueryFilter

	if qp.ID != "" {
		id, err := uuid.Parse(qp.ID)
		if err != nil {
			return alertbus.QueryFilter{}, err
		}
		filter.ID = &id
	}

	if qp.AlertType != "" {
		filter.AlertType = &qp.AlertType
	}

	if qp.Severity != "" {
		severities := strings.Split(qp.Severity, ",")
		for i := range severities {
			severities[i] = strings.TrimSpace(severities[i])
			if !validSeverities[severities[i]] {
				return alertbus.QueryFilter{}, fmt.Errorf("invalid severity value: %s", severities[i])
			}
		}
		filter.Severities = severities
	}

	if qp.Status != "" {
		filter.Status = &qp.Status
	}

	if qp.SourceEntityName != "" {
		filter.SourceEntityName = &qp.SourceEntityName
	}

	if qp.SourceEntityID != "" {
		id, err := uuid.Parse(qp.SourceEntityID)
		if err != nil {
			return alertbus.QueryFilter{}, err
		}
		filter.SourceEntityID = &id
	}

	if qp.SourceRuleID != "" {
		id, err := uuid.Parse(qp.SourceRuleID)
		if err != nil {
			return alertbus.QueryFilter{}, err
		}
		filter.SourceRuleID = &id
	}

	if qp.CreatedAfter != "" {
		t, err := time.Parse(time.RFC3339, qp.CreatedAfter)
		if err != nil {
			return alertbus.QueryFilter{}, fmt.Errorf("invalid createdAfter: %w", err)
		}
		filter.CreatedAfter = &t
	}

	if qp.CreatedBefore != "" {
		t, err := time.Parse(time.RFC3339, qp.CreatedBefore)
		if err != nil {
			return alertbus.QueryFilter{}, fmt.Errorf("invalid createdBefore: %w", err)
		}
		filter.CreatedBefore = &t
	}

	return filter, nil
}
