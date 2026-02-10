package executionapi

import (
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/sdk/workflow"
)

// QueryParams holds raw query parameter values.
type QueryParams struct {
	Page          string
	Rows          string
	OrderBy       string
	ID            string
	RuleID        string
	Status        string
	TriggerSource string
	DateFrom      string
	DateTo        string
}

// parseQueryParams extracts query parameters from the HTTP request.
func parseQueryParams(r *http.Request) QueryParams {
	return QueryParams{
		Page:          r.URL.Query().Get("page"),
		Rows:          r.URL.Query().Get("rows"),
		OrderBy:       r.URL.Query().Get("orderBy"),
		ID:            r.URL.Query().Get("id"),
		RuleID:        r.URL.Query().Get("rule_id"),
		Status:        r.URL.Query().Get("status"),
		TriggerSource: r.URL.Query().Get("trigger_source"),
		DateFrom:      r.URL.Query().Get("date_from"),
		DateTo:        r.URL.Query().Get("date_to"),
	}
}

// parseFilter converts query parameters to a workflow.ExecutionFilter.
func parseFilter(qp QueryParams) (workflow.ExecutionFilter, error) {
	var filter workflow.ExecutionFilter

	if qp.ID != "" {
		id, err := uuid.Parse(qp.ID)
		if err != nil {
			return filter, err
		}
		filter.ID = &id
	}

	if qp.RuleID != "" {
		ruleID, err := uuid.Parse(qp.RuleID)
		if err != nil {
			return filter, err
		}
		filter.RuleID = &ruleID
	}

	if qp.Status != "" {
		status := workflow.ExecutionStatus(qp.Status)
		filter.Status = &status
	}

	if qp.TriggerSource != "" {
		filter.TriggerSource = &qp.TriggerSource
	}

	if qp.DateFrom != "" {
		dateFrom, err := time.Parse(time.RFC3339, qp.DateFrom)
		if err != nil {
			// Try date-only format
			dateFrom, err = time.Parse("2006-01-02", qp.DateFrom)
			if err != nil {
				return filter, err
			}
		}
		filter.DateFrom = &dateFrom
	}

	if qp.DateTo != "" {
		dateTo, err := time.Parse(time.RFC3339, qp.DateTo)
		if err != nil {
			// Try date-only format
			dateTo, err = time.Parse("2006-01-02", qp.DateTo)
			if err != nil {
				return filter, err
			}
		}
		filter.DateTo = &dateTo
	}

	return filter, nil
}
