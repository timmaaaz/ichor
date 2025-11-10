package pageconfigdb

import (
	"database/sql"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/config/pageconfigbus"
	"github.com/timmaaaz/ichor/business/sdk/sqldb/nulltypes"
)

// dbPageConfig is the database representation of PageConfig with nullable user_id
type dbPageConfig struct {
	ID        uuid.UUID      `db:"id"`
	Name      string         `db:"name"`
	UserID    sql.NullString `db:"user_id"`
	IsDefault bool           `db:"is_default"`
}

// toDBPageConfig converts a PageConfig to its database representation
func toDBPageConfig(pc pageconfigbus.PageConfig) dbPageConfig {
	return dbPageConfig{
		ID:        pc.ID,
		Name:      pc.Name,
		UserID:    nulltypes.ToNullableUUID(pc.UserID),
		IsDefault: pc.IsDefault,
	}
}

// toBusPageConfig converts a database PageConfig to its business representation
func toBusPageConfig(db dbPageConfig) pageconfigbus.PageConfig {
	return pageconfigbus.PageConfig{
		ID:        db.ID,
		Name:      db.Name,
		UserID:    nulltypes.FromNullableUUID(db.UserID),
		IsDefault: db.IsDefault,
	}
}

// toBusPageConfigs converts multiple database PageConfigs to business representations
func toBusPageConfigs(dbs []dbPageConfig) []pageconfigbus.PageConfig {
	bus := make([]pageconfigbus.PageConfig, len(dbs))
	for i, db := range dbs {
		bus[i] = toBusPageConfig(db)
	}
	return bus
}
