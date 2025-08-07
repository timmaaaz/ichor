package userassetdb

import (
	"fmt"

	"github.com/timmaaaz/ichor/business/domain/assets/userassetbus"
	"github.com/timmaaaz/ichor/business/sdk/order"
)

var orderByFields = map[string]string{
	userassetbus.OrderByID:                  "id",
	userassetbus.OrderByUserID:              "user_id",
	userassetbus.OrderByAssetID:             "asset_id",
	userassetbus.OrderByApprovedBy:          "approved_by",
	userassetbus.OrderByApprovalStatusID:    "approval_status_id",
	userassetbus.OrderByFulfillmentStatusID: "fulfillment_status_id",
	userassetbus.OrderByDateReceived:        "date_received",
	userassetbus.OrderByLastMaintenance:     "last_maintenance",
}

func orderByClause(orderBy order.By) (string, error) {
	by, exists := orderByFields[orderBy.Field]
	if !exists {
		return "", fmt.Errorf("field %q does not exist", orderBy.Field)
	}

	return " ORDER BY " + by + " " + orderBy.Direction, nil
}
