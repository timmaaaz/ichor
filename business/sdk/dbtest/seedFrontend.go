package dbtest

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/sdk/sqldb"
	"github.com/timmaaaz/ichor/foundation/logger"
)

// InsertSeedData seeds the database for local development and make seed-frontend.
// Each domain's seed data lives in its own seed_*.go file in this package.
// Static config data (table configs, forms, charts, page actions) lives in seedmodels/.
// Architecture reference: docs/arch/seeding.md
func InsertSeedData(log *logger.Logger, cfg sqldb.Config) error {
	db, err := sqldb.Open(cfg)
	if err != nil {
		return fmt.Errorf("connect database: %w", err)
	}
	defer db.Close()
	busDomain := newBusDomains(log, db)

	ctx := context.Background()

	foundation, err := seedFoundation(ctx, busDomain)
	if err != nil {
		return fmt.Errorf("seeding foundation: %w", err)
	}
	adminID := foundation.Admins[0].ID

	// Seed user preferences for all foundation users.
	var allUserIDs uuid.UUIDs
	for _, u := range foundation.Admins {
		allUserIDs = append(allUserIDs, u.ID)
	}
	for _, u := range foundation.Reporters {
		allUserIDs = append(allUserIDs, u.ID)
	}
	for _, u := range foundation.Bosses {
		allUserIDs = append(allUserIDs, u.ID)
	}

	if err := seedUserPreferences(ctx, busDomain, allUserIDs); err != nil {
		return fmt.Errorf("seeding user preferences: %w", err)
	}

	geoHR, err := seedGeographyHR(ctx, busDomain)
	if err != nil {
		return fmt.Errorf("seeding geography and hr: %w", err)
	}

	if err := seedAssets(ctx, busDomain, foundation); err != nil {
		return fmt.Errorf("seeding assets: %w", err)
	}

	if err := seedLabels(ctx, busDomain.Label); err != nil {
		return fmt.Errorf("seed labels: %w", err)
	}

	products, err := seedProducts(ctx, busDomain, geoHR, foundation)
	if err != nil {
		return fmt.Errorf("seeding products: %w", err)
	}

	inventory, err := seedInventory(ctx, busDomain, foundation, geoHR, products)
	if err != nil {
		return fmt.Errorf("seeding inventory: %w", err)
	}

	sales, err := seedSales(ctx, busDomain, foundation, geoHR, products)
	if err != nil {
		return fmt.Errorf("seeding sales: %w", err)
	}

	if err := seedProcurement(ctx, busDomain, foundation, geoHR, products, inventory); err != nil {
		return fmt.Errorf("seeding procurement: %w", err)
	}

	if _, err := seedTasks(ctx, busDomain, foundation, products, inventory, sales); err != nil {
		return fmt.Errorf("seeding tasks: %w", err)
	}

	if err := seedTableBuilder(ctx, busDomain, adminID); err != nil {
		return fmt.Errorf("seeding table builder: %w", err)
	}

	if err := seedPages(ctx, log, busDomain); err != nil {
		return fmt.Errorf("seeding pages: %w", err)
	}

	if err := seedForms(ctx, log, busDomain, db); err != nil {
		return fmt.Errorf("seeding forms: %w", err)
	}

	if err := seedWorkflow(ctx, log, busDomain, adminID); err != nil {
		return fmt.Errorf("seeding workflow: %w", err)
	}

	if err := seedAlerts(ctx, log, busDomain, adminID); err != nil {
		return fmt.Errorf("seeding alerts: %w", err)
	}

	if err := seedCycleCounts(ctx, busDomain, foundation, products, inventory); err != nil {
		return fmt.Errorf("seeding cycle counts: %w", err)
	}

	if err := seedApprovals(ctx, busDomain, foundation); err != nil {
		return fmt.Errorf("seeding approvals: %w", err)
	}

	return nil
}

// InsertPlatformConfig seeds only platform configuration — pages, forms,
// table configs, workflows, and alerts. No demo users, products, or orders.
// Requires migrate + seed.sql to have run first (provides admin_gopher).
func InsertPlatformConfig(log *logger.Logger, cfg sqldb.Config) error {
	db, err := sqldb.Open(cfg)
	if err != nil {
		return fmt.Errorf("connect database: %w", err)
	}
	defer db.Close()
	busDomain := newBusDomains(log, db)

	ctx := context.Background()

	// admin_gopher UUID from seed.sql — hardcoded, stable across all environments.
	adminID := uuid.MustParse("5cf37266-3473-4006-984f-9325122678b7")

	if err := seedTableBuilder(ctx, busDomain, adminID); err != nil {
		return fmt.Errorf("seeding table builder: %w", err)
	}

	if err := seedPages(ctx, log, busDomain); err != nil {
		return fmt.Errorf("seeding pages: %w", err)
	}

	if err := seedForms(ctx, log, busDomain, db); err != nil {
		return fmt.Errorf("seeding forms: %w", err)
	}

	if err := seedWorkflow(ctx, log, busDomain, adminID); err != nil {
		return fmt.Errorf("seeding workflow: %w", err)
	}

	if err := seedAlerts(ctx, log, busDomain, adminID); err != nil {
		return fmt.Errorf("seeding alerts: %w", err)
	}

	return nil
}
