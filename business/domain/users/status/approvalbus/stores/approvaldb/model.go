package approvaldb

import (
	"database/sql"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/users/status/approvalbus"
	"github.com/timmaaaz/ichor/business/sdk/sqldb/nulltypes"
)

type userApprovalStatus struct {
	ID     uuid.UUID      `db:"id"`
	Name   string         `db:"name"`
	IconID sql.NullString `db:"icon_id"`
}

func toDBUserApprovalStatus(as approvalbus.UserApprovalStatus) userApprovalStatus {
	return userApprovalStatus{
		ID:     as.ID,
		Name:   as.Name,
		IconID: nulltypes.ToNullableUUID(as.IconID),
	}
}

func toBusUserApprovalStatus(dbAS userApprovalStatus) approvalbus.UserApprovalStatus {
	return approvalbus.UserApprovalStatus{
		ID:     dbAS.ID,
		Name:   dbAS.Name,
		IconID: nulltypes.FromNullableUUID(dbAS.IconID),
	}
}

func toBusUserApprovalStatuses(dbAS []userApprovalStatus) []approvalbus.UserApprovalStatus {
	aprvlStatuses := make([]approvalbus.UserApprovalStatus, len(dbAS))
	for i, as := range dbAS {
		aprvlStatuses[i] = toBusUserApprovalStatus(as)
	}

	return aprvlStatuses
}
