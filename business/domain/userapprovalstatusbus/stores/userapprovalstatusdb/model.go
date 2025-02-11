package userapprovalstatusdb

import (
	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/userapprovalstatusbus"
)

type userApprovalStatus struct {
	ID     uuid.UUID `db:"user_approval_status_id"`
	Name   string    `db:"name"`
	IconID uuid.UUID `db:"icon_id"`
}

func toDBUserApprovalStatus(as userapprovalstatusbus.UserApprovalStatus) userApprovalStatus {
	return userApprovalStatus{
		ID:     as.ID,
		Name:   as.Name,
		IconID: as.IconID,
	}
}

func toBusUserApprovalStatus(dbAS userApprovalStatus) userapprovalstatusbus.UserApprovalStatus {
	return userapprovalstatusbus.UserApprovalStatus{
		ID:     dbAS.ID,
		Name:   dbAS.Name,
		IconID: dbAS.IconID,
	}
}

func toBusUserApprovalStatuses(dbAS []userApprovalStatus) []userapprovalstatusbus.UserApprovalStatus {
	aprvlStatuses := make([]userapprovalstatusbus.UserApprovalStatus, len(dbAS))
	for i, as := range dbAS {
		aprvlStatuses[i] = toBusUserApprovalStatus(as)
	}

	return aprvlStatuses
}
