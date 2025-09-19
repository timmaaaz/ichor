package commands

import (
	"fmt"

	"github.com/timmaaaz/ichor/business/sdk/dbtest"
	"github.com/timmaaaz/ichor/business/sdk/sqldb"
	"github.com/timmaaaz/ichor/foundation/logger"
)

// SeedFrontend loads test data into the database.
func SeedFrontend(log *logger.Logger, cfg sqldb.Config) error {
	db, err := sqldb.Open(cfg)
	if err != nil {
		return fmt.Errorf("connect database: %w", err)
	}
	defer db.Close()

	if err := dbtest.InsertSeedData(log, cfg); err != nil {
		return fmt.Errorf("inserting seed data: %w", err)
	}

	fmt.Println("seed data complete")
	return nil
}
