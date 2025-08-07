package approvalstatusdb

import (
	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/assets/approvalstatusbus"
)

type approvalStatus struct {
	ID     uuid.UUID `db:"id"`
	Name   string    `db:"name"`
	IconID uuid.UUID `db:"icon_id"`
}

func toDBApprovalStatus(as approvalstatusbus.ApprovalStatus) approvalStatus {
	return approvalStatus{
		ID:     as.ID,
		Name:   as.Name,
		IconID: as.IconID,
	}
}

func toBusApprovalStatus(dbAS approvalStatus) approvalstatusbus.ApprovalStatus {
	return approvalstatusbus.ApprovalStatus{
		ID:     dbAS.ID,
		Name:   dbAS.Name,
		IconID: dbAS.IconID,
	}
}

func toBusApprovalStatuses(dbAS []approvalStatus) []approvalstatusbus.ApprovalStatus {
	aprvlStatuses := make([]approvalstatusbus.ApprovalStatus, len(dbAS))
	for i, as := range dbAS {
		aprvlStatuses[i] = toBusApprovalStatus(as)
	}

	return aprvlStatuses
}
