package ruleapi

import "github.com/timmaaaz/ichor/business/sdk/workflow"

// orderByFields maps API field names to database field names for ordering.
var orderByFields = map[string]string{
	"id":          workflow.OrderByID,
	"name":        workflow.OrderByName,
	"createdDate": workflow.OrderByCreatedDate,
	"updatedDate": workflow.OrderByUpdatedDate,
	"isActive":    workflow.OrderByIsActive,
}
