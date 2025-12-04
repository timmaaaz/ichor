package approvaldb

import (
	"database/sql"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/hr/approvalbus"
	"github.com/timmaaaz/ichor/business/sdk/sqldb/nulltypes"
)

type userApprovalStatus struct {
	ID             uuid.UUID      `db:"id"`
	Name           string         `db:"name"`
	IconID         sql.NullString `db:"icon_id"`
	PrimaryColor   sql.NullString `db:"primary_color"`
	SecondaryColor sql.NullString `db:"secondary_color"`
	Icon           sql.NullString `db:"icon"`
}

func toDBUserApprovalStatus(as approvalbus.UserApprovalStatus) userApprovalStatus {
	db := userApprovalStatus{
		ID:     as.ID,
		Name:   as.Name,
		IconID: nulltypes.ToNullableUUID(as.IconID),
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

func toBusUserApprovalStatus(dbAS userApprovalStatus) approvalbus.UserApprovalStatus {
	bus := approvalbus.UserApprovalStatus{
		ID:     dbAS.ID,
		Name:   dbAS.Name,
		IconID: nulltypes.FromNullableUUID(dbAS.IconID),
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

func toBusUserApprovalStatuses(dbAS []userApprovalStatus) []approvalbus.UserApprovalStatus {
	aprvlStatuses := make([]approvalbus.UserApprovalStatus, len(dbAS))
	for i, as := range dbAS {
		aprvlStatuses[i] = toBusUserApprovalStatus(as)
	}

	return aprvlStatuses
}
