package actionpermissionsdb

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/workflow/actionpermissionsbus"
)

// dbActionPermission represents the database structure for an action permission.
// Using "dbActionPermission" prefix to avoid conflicts with the page package.
type dbActionPermission struct {
	ID          uuid.UUID       `db:"id"`
	RoleID      uuid.UUID       `db:"role_id"`
	ActionType  string          `db:"action_type"`
	IsAllowed   bool            `db:"is_allowed"`
	Constraints json.RawMessage `db:"constraints"`
	CreatedAt   time.Time       `db:"created_date"`
	UpdatedAt   time.Time       `db:"updated_date"`
}

func toDBActionPermission(ap actionpermissionsbus.ActionPermission) dbActionPermission {
	return dbActionPermission{
		ID:          ap.ID,
		RoleID:      ap.RoleID,
		ActionType:  ap.ActionType,
		IsAllowed:   ap.IsAllowed,
		Constraints: ap.Constraints,
		CreatedAt:   ap.CreatedAt,
		UpdatedAt:   ap.UpdatedAt,
	}
}

func toBusActionPermission(db dbActionPermission) actionpermissionsbus.ActionPermission {
	return actionpermissionsbus.ActionPermission{
		ID:          db.ID,
		RoleID:      db.RoleID,
		ActionType:  db.ActionType,
		IsAllowed:   db.IsAllowed,
		Constraints: db.Constraints,
		CreatedAt:   db.CreatedAt,
		UpdatedAt:   db.UpdatedAt,
	}
}

func toBusActionPermissions(dbs []dbActionPermission) []actionpermissionsbus.ActionPermission {
	perms := make([]actionpermissionsbus.ActionPermission, len(dbs))
	for i, db := range dbs {
		perms[i] = toBusActionPermission(db)
	}
	return perms
}
