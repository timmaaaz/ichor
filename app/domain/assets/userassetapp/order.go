package userassetapp

import (
	"github.com/timmaaaz/ichor/business/domain/assets/userassetbus"
	"github.com/timmaaaz/ichor/business/sdk/order"
)

var defaultOrderBy = order.NewBy("id", order.ASC)

var orderByFields = map[string]string{
	"id":                    userassetbus.OrderByID,
	"user_id":               userassetbus.OrderByUserID,
	"asset_id":              userassetbus.OrderByAssetID,
	"approved_by":           userassetbus.OrderByApprovedBy,
	"approval_status_id":    userassetbus.OrderByApprovalStatusID,
	"fulfillment_status_id": userassetbus.OrderByFulfillmentStatusID,
	"date_received":         userassetbus.OrderByDateReceived,
	"last_maintenance":      userassetbus.OrderByLastMaintenance,
}
