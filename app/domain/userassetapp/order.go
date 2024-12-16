package userassetapp

import (
	"github.com/timmaaaz/ichor/business/domain/userassetbus"
	"github.com/timmaaaz/ichor/business/sdk/order"
)

var defaultOrderBy = order.NewBy("user_asset_id", order.ASC)

var orderByFields = map[string]string{
	"user_asset_id":         userassetbus.OrderByID,
	"user_id":               userassetbus.OrderByUserID,
	"asset_id":              userassetbus.OrderByAssetID,
	"approved_by":           userassetbus.OrderByApprovedBy,
	"condition_id":          userassetbus.OrderByConditionID,
	"approval_status_id":    userassetbus.OrderByApprovalStatusID,
	"fulfillment_status_id": userassetbus.OrderByFulfillmentStatusID,
	"date_received":         userassetbus.OrderByDateReceived,
	"last_maintenance":      userassetbus.OrderByLastMaintenance,
}
