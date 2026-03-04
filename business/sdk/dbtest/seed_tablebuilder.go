package dbtest

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/sdk/dbtest/seedmodels"
)

func seedTableBuilder(ctx context.Context, busDomain BusDomain, adminID uuid.UUID) error {
	_, err := busDomain.ConfigStore.Create(ctx, "orders_dashboard", "orders_base", seedmodels.OrdersConfig, adminID)
	if err != nil {
		return fmt.Errorf("creating stored config: %w", err)
	}

	_, err = busDomain.ConfigStore.Create(ctx, "products_dashboard", "products", seedmodels.TableConfig, adminID)
	if err != nil {
		return fmt.Errorf("creating stored config: %w", err)
	}

	_, err = busDomain.ConfigStore.Create(ctx, "inventory_dashboard", "inventory_items", seedmodels.ComplexConfig, adminID)
	if err != nil {
		return fmt.Errorf("creating stored config: %w", err)
	}

	// Create dedicated page configs for orders, suppliers, categories, and order line items
	_, err = busDomain.ConfigStore.Create(ctx, "orders_table", "orders", seedmodels.OrdersTableConfig, adminID)
	if err != nil {
		return fmt.Errorf("creating orders table config: %w", err)
	}

	_, err = busDomain.ConfigStore.Create(ctx, "suppliers_table", "suppliers", seedmodels.SuppliersTableConfig, adminID)
	if err != nil {
		return fmt.Errorf("creating suppliers table config: %w", err)
	}

	_, err = busDomain.ConfigStore.Create(ctx, "categories_table", "product_categories", seedmodels.CategoriesTableConfig, adminID)
	if err != nil {
		return fmt.Errorf("creating categories table config: %w", err)
	}

	_, err = busDomain.ConfigStore.Create(ctx, "order_line_items_table", "order_line_items", seedmodels.OrderLineItemsTableConfig, adminID)
	if err != nil {
		return fmt.Errorf("creating order line items table config: %w", err)
	}

	// Admin Module Configs
	_, err = busDomain.ConfigStore.Create(ctx, "admin_users_table", "users", adminUsersTableConfig, adminID)
	if err != nil {
		return fmt.Errorf("creating admin users table config: %w", err)
	}

	_, err = busDomain.ConfigStore.Create(ctx, "admin_roles_table", "roles", adminRolesTableConfig, adminID)
	if err != nil {
		return fmt.Errorf("creating admin roles table config: %w", err)
	}

	_, err = busDomain.ConfigStore.Create(ctx, "admin_table_access_table", "table_access", adminTableAccessTableConfig, adminID)
	if err != nil {
		return fmt.Errorf("creating admin table access table config: %w", err)
	}

	_, err = busDomain.ConfigStore.Create(ctx, "admin_audit_table", "automation_executions", adminAuditTableConfig, adminID)
	if err != nil {
		return fmt.Errorf("creating admin audit table config: %w", err)
	}

	_, err = busDomain.ConfigStore.Create(ctx, "admin_config_table", "table_configs", adminConfigTableConfig, adminID)
	if err != nil {
		return fmt.Errorf("creating admin config table config: %w", err)
	}

	// Assets Module Configs
	_, err = busDomain.ConfigStore.Create(ctx, "assets_list_table", "assets", seedmodels.AssetsListTableConfig, adminID)
	if err != nil {
		return fmt.Errorf("creating assets list table config: %w", err)
	}

	_, err = busDomain.ConfigStore.Create(ctx, "assets_requests_table", "user_assets", seedmodels.AssetsRequestsTableConfig, adminID)
	if err != nil {
		return fmt.Errorf("creating assets requests table config: %w", err)
	}

	// HR Module Configs
	_, err = busDomain.ConfigStore.Create(ctx, "hr_employees_table", "users", seedmodels.HrEmployeesTableConfig, adminID)
	if err != nil {
		return fmt.Errorf("creating hr employees table config: %w", err)
	}

	_, err = busDomain.ConfigStore.Create(ctx, "hr_offices_table", "offices", seedmodels.HrOfficesTableConfig, adminID)
	if err != nil {
		return fmt.Errorf("creating hr offices table config: %w", err)
	}

	// Inventory Module Configs
	_, err = busDomain.ConfigStore.Create(ctx, "inventory_warehouses_table", "warehouses", seedmodels.InventoryWarehousesTableConfig, adminID)
	if err != nil {
		return fmt.Errorf("creating inventory warehouses table config: %w", err)
	}

	_, err = busDomain.ConfigStore.Create(ctx, "inventory_items_table", "inventory_items", seedmodels.InventoryItemsTableConfig, adminID)
	if err != nil {
		return fmt.Errorf("creating inventory items table config: %w", err)
	}

	_, err = busDomain.ConfigStore.Create(ctx, "inventory_adjustments_table", "inventory_adjustments", seedmodels.InventoryAdjustmentsTableConfig, adminID)
	if err != nil {
		return fmt.Errorf("creating inventory adjustments table config: %w", err)
	}

	_, err = busDomain.ConfigStore.Create(ctx, "inventory_transfers_table", "transfer_orders", seedmodels.InventoryTransfersTableConfig, adminID)
	if err != nil {
		return fmt.Errorf("creating inventory transfers table config: %w", err)
	}

	_, err = busDomain.ConfigStore.Create(ctx, "inventory_zones_table", "zones", seedmodels.InventoryZonesTableConfig, adminID)
	if err != nil {
		return fmt.Errorf("creating inventory zones table config: %w", err)
	}

	_, err = busDomain.ConfigStore.Create(ctx, "inventory_locations_table", "inventory_locations", seedmodels.InventoryLocationsTableConfig, adminID)
	if err != nil {
		return fmt.Errorf("creating inventory locations table config: %w", err)
	}

	_, err = busDomain.ConfigStore.Create(ctx, "products_table", "products", seedmodels.ProductsListTableConfig, adminID)
	if err != nil {
		return fmt.Errorf("creating products table config: %w", err)
	}

	// Sales Module Configs
	_, err = busDomain.ConfigStore.Create(ctx, "sales_customers_table", "customers", seedmodels.SalesCustomersTableConfig, adminID)
	if err != nil {
		return fmt.Errorf("creating sales customers table config: %w", err)
	}

	_, err = busDomain.ConfigStore.Create(ctx, "products_with_prices_lookup", "products", seedmodels.ProductsWithPricesLookup, adminID)
	if err != nil {
		return fmt.Errorf("creating products with prices lookup config: %w", err)
	}

	// Procurement Module Configs
	_, err = busDomain.ConfigStore.Create(ctx, "procurement_purchase_orders_config", "purchase_orders", seedmodels.PurchaseOrderTableConfig, adminID)
	if err != nil {
		return fmt.Errorf("creating procurement purchase orders config: %w", err)
	}

	_, err = busDomain.ConfigStore.Create(ctx, "procurement_line_items_config", "purchase_order_line_items", seedmodels.PurchaseOrderLineItemTableConfig, adminID)
	if err != nil {
		return fmt.Errorf("creating procurement line items config: %w", err)
	}

	_, err = busDomain.ConfigStore.Create(ctx, "procurement_approvals_open_config", "purchase_orders", seedmodels.ProcurementOpenApprovalsTableConfig, adminID)
	if err != nil {
		return fmt.Errorf("creating procurement approvals open config: %w", err)
	}

	_, err = busDomain.ConfigStore.Create(ctx, "procurement_approvals_closed_config", "purchase_orders", seedmodels.ProcurementClosedApprovalsTableConfig, adminID)
	if err != nil {
		return fmt.Errorf("creating procurement approvals closed config: %w", err)
	}

	// =========================================================================
	// Chart Configurations - 14 seed charts covering all chart types
	// =========================================================================
	for _, chartConfig := range seedmodels.ChartConfigs {
		_, err = busDomain.ConfigStore.Create(ctx, chartConfig.Name, chartConfig.Description, chartConfig.Config, adminID)
		if err != nil {
			return fmt.Errorf("creating chart config %s: %w", chartConfig.Name, err)
		}
	}

	return nil
}
