package commands

import (
	"fmt"

	"github.com/timmaaaz/ichor/business/sdk/dbtest"
	"github.com/timmaaaz/ichor/business/sdk/sqldb"
	"github.com/timmaaaz/ichor/foundation/logger"
)

// SeedPlatform seeds only platform configuration (pages, forms, table configs,
// workflows, alerts) without any demo data.
func SeedPlatform(log *logger.Logger, cfg sqldb.Config) error {
	if err := dbtest.InsertPlatformConfig(log, cfg); err != nil {
		return fmt.Errorf("inserting platform config: %w", err)
	}

	fmt.Println("platform config seeded")
	return nil
}
