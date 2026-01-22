package commands

import (
	"fmt"

	"github.com/timmaaaz/ichor/business/sdk/dbtest/seedmodels"
	"github.com/timmaaaz/ichor/business/sdk/tablebuilder"
)

// configEntry pairs a config with its name for validation reporting
type configEntry struct {
	name   string
	config *tablebuilder.Config
}

// ValidateConfigs validates all seed table and chart configurations.
// Returns nil if all configs are valid, otherwise returns validation errors.
// This command does not require a database connection.
func ValidateConfigs() error {
	// Collect all configs to validate
	configs := []configEntry{
		// Table configs from seedmodels/tables.go
		{"TableConfig", seedmodels.TableConfig},
		{"ComplexConfig", seedmodels.ComplexConfig},
		{"OrdersConfig", seedmodels.OrdersConfig},
		{"OrdersTableConfig", seedmodels.OrdersTableConfig},
		{"SuppliersTableConfig", seedmodels.SuppliersTableConfig},
		{"OrderLineItemsTableConfig", seedmodels.OrderLineItemsTableConfig},
		{"CategoriesTableConfig", seedmodels.CategoriesTableConfig},
		{"AssetsListTableConfig", seedmodels.AssetsListTableConfig},
		{"AssetsRequestsTableConfig", seedmodels.AssetsRequestsTableConfig},
		{"HrEmployeesTableConfig", seedmodels.HrEmployeesTableConfig},
		{"HrOfficesTableConfig", seedmodels.HrOfficesTableConfig},
		{"InventoryWarehousesTableConfig", seedmodels.InventoryWarehousesTableConfig},
		{"InventoryAdjustmentsTableConfig", seedmodels.InventoryAdjustmentsTableConfig},
		{"InventoryTransfersTableConfig", seedmodels.InventoryTransfersTableConfig},
		{"SalesCustomersTableConfig", seedmodels.SalesCustomersTableConfig},
		{"PurchaseOrderTableConfig", seedmodels.PurchaseOrderTableConfig},
		{"PurchaseOrderLineItemTableConfig", seedmodels.PurchaseOrderLineItemTableConfig},
		{"ProcurementOpenApprovalsTableConfig", seedmodels.ProcurementOpenApprovalsTableConfig},
		{"ProcurementClosedApprovalsTableConfig", seedmodels.ProcurementClosedApprovalsTableConfig},
		{"ProductsWithPricesLookup", seedmodels.ProductsWithPricesLookup},

		// Chart configs from seedmodels/charts.go
		{"SeedKPITotalRevenue", seedmodels.SeedKPITotalRevenue},
		{"SeedKPIOrderCount", seedmodels.SeedKPIOrderCount},
		{"SeedGaugeRevenueTarget", seedmodels.SeedGaugeRevenueTarget},
		{"SeedLineMonthlySales", seedmodels.SeedLineMonthlySales},
		{"SeedBarTopProducts", seedmodels.SeedBarTopProducts},
		{"SeedStackedBarRegionCategory", seedmodels.SeedStackedBarRegionCategory},
		{"SeedStackedAreaCumulative", seedmodels.SeedStackedAreaCumulative},
		{"SeedPieRevenueCategory", seedmodels.SeedPieRevenueCategory},
		{"SeedComboRevenueOrders", seedmodels.SeedComboRevenueOrders},
		{"SeedWaterfallProfit", seedmodels.SeedWaterfallProfit},
		{"SeedFunnelPipeline", seedmodels.SeedFunnelPipeline},
		{"SeedHeatmapSalesTime", seedmodels.SeedHeatmapSalesTime},
		{"SeedTreemapRevenue", seedmodels.SeedTreemapRevenue},
		{"SeedGanttProject", seedmodels.SeedGanttProject},
	}

	var (
		hasErrors    bool
		validCount   int
		invalidCount int
		warnCount    int
	)

	fmt.Println("Validating seed configurations...")
	fmt.Println()

	for _, entry := range configs {
		result := entry.config.ValidateConfig()

		if result.HasErrors() {
			hasErrors = true
			invalidCount++
			fmt.Printf("❌ %s:\n", entry.name)
			for _, err := range result.Errors {
				fmt.Printf("   • %s: %s\n", err.Field, err.Message)
			}
		} else {
			validCount++
			fmt.Printf("✓ %s\n", entry.name)
		}

		// Show warnings regardless of error status
		for _, warn := range result.Warnings {
			warnCount++
			fmt.Printf("   ⚠ %s: %s\n", warn.Field, warn.Message)
		}
	}

	fmt.Println()
	fmt.Printf("Summary: %d valid, %d invalid, %d warnings\n", validCount, invalidCount, warnCount)

	if hasErrors {
		return fmt.Errorf("validation failed: %d config(s) have errors", invalidCount)
	}

	fmt.Println("\nAll configurations valid!")
	return nil
}
