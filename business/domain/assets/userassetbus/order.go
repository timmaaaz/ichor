package userassetbus

import "github.com/timmaaaz/ichor/business/sdk/order"

var DefaultOrderBy = order.NewBy(OrderByID, order.ASC)

const (
	OrderByID                  = "id"
	OrderByUserID              = "user_id"
	OrderByAssetID             = "asset_id"
	OrderByApprovedBy          = "approved_by"
	OrderByApprovalStatusID    = "approval_status_id"
	OrderByFulfillmentStatusID = "fulfillment_status_id"
	OrderByDateReceived        = "date_received"
	OrderByLastMaintenance     = "last_maintenance"
)
