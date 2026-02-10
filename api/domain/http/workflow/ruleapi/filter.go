package ruleapi

import (
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/sdk/workflow"
)

// QueryParams holds parsed query string parameters.
type QueryParams struct {
	Page          string
	Rows          string
	OrderBy       string
	ID            string
	Name          string
	IsActive      string
	EntityID      string
	EntityTypeID  string
	TriggerTypeID string
	CreatedBy     string
}

// parseQueryParams extracts query parameters from the request.
func parseQueryParams(r *http.Request) QueryParams {
	return QueryParams{
		Page:          r.URL.Query().Get("page"),
		Rows:          r.URL.Query().Get("rows"),
		OrderBy:       r.URL.Query().Get("orderBy"),
		ID:            r.URL.Query().Get("id"),
		Name:          r.URL.Query().Get("name"),
		IsActive:      r.URL.Query().Get("is_active"),
		EntityID:      r.URL.Query().Get("entity_id"),
		EntityTypeID:  r.URL.Query().Get("entity_type_id"),
		TriggerTypeID: r.URL.Query().Get("trigger_type_id"),
		CreatedBy:     r.URL.Query().Get("created_by"),
	}
}

// parseFilter converts QueryParams to a workflow.AutomationRuleFilter.
func parseFilter(qp QueryParams) (workflow.AutomationRuleFilter, error) {
	var filter workflow.AutomationRuleFilter

	if qp.ID != "" {
		id, err := uuid.Parse(qp.ID)
		if err != nil {
			return filter, err
		}
		filter.ID = &id
	}

	if qp.Name != "" {
		filter.Name = &qp.Name
	}

	if qp.IsActive != "" {
		active, err := strconv.ParseBool(qp.IsActive)
		if err != nil {
			return filter, err
		}
		filter.IsActive = &active
	}

	if qp.EntityID != "" {
		id, err := uuid.Parse(qp.EntityID)
		if err != nil {
			return filter, err
		}
		filter.EntityID = &id
	}

	if qp.EntityTypeID != "" {
		id, err := uuid.Parse(qp.EntityTypeID)
		if err != nil {
			return filter, err
		}
		filter.EntityTypeID = &id
	}

	if qp.TriggerTypeID != "" {
		id, err := uuid.Parse(qp.TriggerTypeID)
		if err != nil {
			return filter, err
		}
		filter.TriggerTypeID = &id
	}

	if qp.CreatedBy != "" {
		id, err := uuid.Parse(qp.CreatedBy)
		if err != nil {
			return filter, err
		}
		filter.CreatedBy = &id
	}

	return filter, nil
}
