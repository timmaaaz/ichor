package alertapi

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/workflow/alertbus"
)

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
		filter.Severity = &qp.Severity
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

	return filter, nil
}
