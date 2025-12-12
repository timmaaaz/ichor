package approvalstatusdb

import (
	"database/sql"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/assets/approvalstatusbus"
)

type approvalStatus struct {
	ID             uuid.UUID      `db:"id"`
	Name           string         `db:"name"`
	IconID         uuid.UUID      `db:"icon_id"`
	PrimaryColor   sql.NullString `db:"primary_color"`
	SecondaryColor sql.NullString `db:"secondary_color"`
	Icon           sql.NullString `db:"icon"`
}

func toDBApprovalStatus(as approvalstatusbus.ApprovalStatus) approvalStatus {
	db := approvalStatus{
		ID:     as.ID,
		Name:   as.Name,
		IconID: as.IconID,
	}

	if as.PrimaryColor != "" {
		db.PrimaryColor = sql.NullString{String: as.PrimaryColor, Valid: true}
	}

	if as.SecondaryColor != "" {
		db.SecondaryColor = sql.NullString{String: as.SecondaryColor, Valid: true}
	}

	if as.Icon != "" {
		db.Icon = sql.NullString{String: as.Icon, Valid: true}
	}

	return db
}

func toBusApprovalStatus(dbAS approvalStatus) approvalstatusbus.ApprovalStatus {
	bus := approvalstatusbus.ApprovalStatus{
		ID:     dbAS.ID,
		Name:   dbAS.Name,
		IconID: dbAS.IconID,
	}

	if dbAS.PrimaryColor.Valid {
		bus.PrimaryColor = dbAS.PrimaryColor.String
	}

	if dbAS.SecondaryColor.Valid {
		bus.SecondaryColor = dbAS.SecondaryColor.String
	}

	if dbAS.Icon.Valid {
		bus.Icon = dbAS.Icon.String
	}

	return bus
}

func toBusApprovalStatuses(dbAS []approvalStatus) []approvalstatusbus.ApprovalStatus {
	aprvlStatuses := make([]approvalstatusbus.ApprovalStatus, len(dbAS))
	for i, as := range dbAS {
		aprvlStatuses[i] = toBusApprovalStatus(as)
	}

	return aprvlStatuses
}
