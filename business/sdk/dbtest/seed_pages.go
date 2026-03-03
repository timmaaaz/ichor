package dbtest

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/config/pageconfigbus"
	"github.com/timmaaaz/ichor/business/domain/config/pagecontentbus"
	"github.com/timmaaaz/ichor/business/domain/core/rolepagebus"
	"github.com/timmaaaz/ichor/business/domain/core/userrolebus"
	"github.com/timmaaaz/ichor/business/sdk/dbtest/seedmodels"
	"github.com/timmaaaz/ichor/business/sdk/page"
	"github.com/timmaaaz/ichor/foundation/logger"
)

func seedPages(ctx context.Context, log *logger.Logger, busDomain BusDomain) error {

	// =========================================================================
	// Create dedicated page configs for Orders, Suppliers, and Categories
	// =========================================================================

	// Get the stored config IDs for the new pages
	ordersTableStored, err := busDomain.ConfigStore.QueryByName(ctx, "orders_table")
	if err != nil {
		return fmt.Errorf("querying orders table config: %w", err)
	}

	suppliersTableStored, err := busDomain.ConfigStore.QueryByName(ctx, "suppliers_table")
	if err != nil {
		return fmt.Errorf("querying suppliers table config: %w", err)
	}

	categoriesTableStored, err := busDomain.ConfigStore.QueryByName(ctx, "categories_table")
	if err != nil {
		return fmt.Errorf("querying categories table config: %w", err)
	}

	orderLineItemsTableStored, err := busDomain.ConfigStore.QueryByName(ctx, "order_line_items_table")
	if err != nil {
		return fmt.Errorf("querying order line items table config: %w", err)
	}

	// Query Admin Module Configs
	adminUsersTableStored, err := busDomain.ConfigStore.QueryByName(ctx, "admin_users_table")
	if err != nil {
		return fmt.Errorf("querying admin users table config: %w", err)
	}

	adminRolesTableStored, err := busDomain.ConfigStore.QueryByName(ctx, "admin_roles_table")
	if err != nil {
		return fmt.Errorf("querying admin roles table config: %w", err)
	}

	adminTableAccessTableStored, err := busDomain.ConfigStore.QueryByName(ctx, "admin_table_access_table")
	if err != nil {
		return fmt.Errorf("querying admin table access table config: %w", err)
	}

	adminAuditTableStored, err := busDomain.ConfigStore.QueryByName(ctx, "admin_audit_table")
	if err != nil {
		return fmt.Errorf("querying admin audit table config: %w", err)
	}

	adminConfigTableStored, err := busDomain.ConfigStore.QueryByName(ctx, "admin_config_table")
	if err != nil {
		return fmt.Errorf("querying admin config table config: %w", err)
	}

	// Query Assets Module Configs
	assetsListTableStored, err := busDomain.ConfigStore.QueryByName(ctx, "assets_list_table")
	if err != nil {
		return fmt.Errorf("querying assets list table config: %w", err)
	}

	assetsRequestsTableStored, err := busDomain.ConfigStore.QueryByName(ctx, "assets_requests_table")
	if err != nil {
		return fmt.Errorf("querying assets requests table config: %w", err)
	}

	// Query HR Module Configs
	hrEmployeesTableStored, err := busDomain.ConfigStore.QueryByName(ctx, "hr_employees_table")
	if err != nil {
		return fmt.Errorf("querying hr employees table config: %w", err)
	}

	hrOfficesTableStored, err := busDomain.ConfigStore.QueryByName(ctx, "hr_offices_table")
	if err != nil {
		return fmt.Errorf("querying hr offices table config: %w", err)
	}

	// Query Inventory Module Configs
	inventoryWarehousesTableStored, err := busDomain.ConfigStore.QueryByName(ctx, "inventory_warehouses_table")
	if err != nil {
		return fmt.Errorf("querying inventory warehouses table config: %w", err)
	}

	inventoryItemsTableStored, err := busDomain.ConfigStore.QueryByName(ctx, "inventory_items_table")
	if err != nil {
		return fmt.Errorf("querying inventory items table config: %w", err)
	}

	inventoryAdjustmentsTableStored, err := busDomain.ConfigStore.QueryByName(ctx, "inventory_adjustments_table")
	if err != nil {
		return fmt.Errorf("querying inventory adjustments table config: %w", err)
	}

	inventoryTransfersTableStored, err := busDomain.ConfigStore.QueryByName(ctx, "inventory_transfers_table")
	if err != nil {
		return fmt.Errorf("querying inventory transfers table config: %w", err)
	}

	// Query Sales Module Configs
	salesCustomersTableStored, err := busDomain.ConfigStore.QueryByName(ctx, "sales_customers_table")
	if err != nil {
		return fmt.Errorf("querying sales customers table config: %w", err)
	}

	// Query Procurement Module Configs
	procurementPurchaseOrdersConfigStored, err := busDomain.ConfigStore.QueryByName(ctx, "procurement_purchase_orders_config")
	if err != nil {
		return fmt.Errorf("querying procurement purchase orders config: %w", err)
	}

	procurementLineItemsConfigStored, err := busDomain.ConfigStore.QueryByName(ctx, "procurement_line_items_config")
	if err != nil {
		return fmt.Errorf("querying procurement line items config: %w", err)
	}

	procurementApprovalsOpenConfigStored, err := busDomain.ConfigStore.QueryByName(ctx, "procurement_approvals_open_config")
	if err != nil {
		return fmt.Errorf("querying procurement approvals open config: %w", err)
	}

	procurementApprovalsClosedConfigStored, err := busDomain.ConfigStore.QueryByName(ctx, "procurement_approvals_closed_config")
	if err != nil {
		return fmt.Errorf("querying procurement approvals closed config: %w", err)
	}

	// Query Chart Configs for distribution across pages
	// These use _ for error since charts are optional - pages work without them
	kpiRevenueStored, _ := busDomain.ConfigStore.QueryByName(ctx, "seed_kpi_total_revenue")
	kpiOrdersStored, _ := busDomain.ConfigStore.QueryByName(ctx, "seed_kpi_order_count")
	gaugeRevenueStored, _ := busDomain.ConfigStore.QueryByName(ctx, "seed_gauge_revenue_target")
	lineMonthlySalesStored, _ := busDomain.ConfigStore.QueryByName(ctx, "seed_line_monthly_sales")
	barTopProductsStored, _ := busDomain.ConfigStore.QueryByName(ctx, "seed_bar_top_products")
	pieRevenueCategoryStored, _ := busDomain.ConfigStore.QueryByName(ctx, "seed_pie_revenue_category")
	funnelPipelineStored, _ := busDomain.ConfigStore.QueryByName(ctx, "seed_funnel_pipeline")
	heatmapSalesTimeStored, _ := busDomain.ConfigStore.QueryByName(ctx, "seed_heatmap_sales_time")
	treemapRevenueStored, _ := busDomain.ConfigStore.QueryByName(ctx, "seed_treemap_revenue")
	ganttProjectStored, _ := busDomain.ConfigStore.QueryByName(ctx, "seed_gantt_project")

	// Create Orders Page
	ordersPage, err := busDomain.PageConfig.Create(ctx, pageconfigbus.NewPageConfig{
		Name:      "orders_page",
		UserID:    uuid.Nil,
		IsDefault: true,
	})
	if err != nil {
		return fmt.Errorf("creating orders page: %w", err)
	}

	ordersPageOrderIndex := 1

	// Add charts to Orders Page
	if lineMonthlySalesStored != nil {
		_, err = busDomain.PageContent.Create(ctx, pagecontentbus.NewPageContent{
			PageConfigID:  ordersPage.ID,
			ContentType:   pagecontentbus.ContentTypeChart,
			Label:         "Monthly Sales Trend",
			ChartConfigID: lineMonthlySalesStored.ID,
			OrderIndex:    ordersPageOrderIndex,
			Layout:        json.RawMessage(`{"colSpan":{"xs":8,"sm":8,"md":8,"lg":8,"xl":8}}`),
			IsVisible:     true,
			IsDefault:     false,
		})
		if err != nil {
			return fmt.Errorf("creating orders page line chart: %w", err)
		}
		ordersPageOrderIndex++
	}

	if funnelPipelineStored != nil {
		_, err = busDomain.PageContent.Create(ctx, pagecontentbus.NewPageContent{
			PageConfigID:  ordersPage.ID,
			ContentType:   pagecontentbus.ContentTypeChart,
			Label:         "Sales Pipeline",
			ChartConfigID: funnelPipelineStored.ID,
			OrderIndex:    ordersPageOrderIndex,
			Layout:        json.RawMessage(`{"colSpan":{"xs":4,"sm":4,"md":4,"lg":4,"xl":4}}`),
			IsVisible:     true,
			IsDefault:     false,
		})
		if err != nil {
			return fmt.Errorf("creating orders page funnel chart: %w", err)
		}
		ordersPageOrderIndex++
	}

	_, err = busDomain.PageContent.Create(ctx, pagecontentbus.NewPageContent{
		PageConfigID:  ordersPage.ID,
		ParentID:      uuid.Nil,
		ContentType:   pagecontentbus.ContentTypeTable,
		Label:         "",
		TableConfigID: ordersTableStored.ID,
		OrderIndex:    ordersPageOrderIndex,
		Layout:        json.RawMessage(`{}`),
		IsVisible:     true,
		IsDefault:     true,
	})
	if err != nil {
		return fmt.Errorf("creating orders page content: %w", err)
	}

	// Create Suppliers Page
	suppliersPage, err := busDomain.PageConfig.Create(ctx, pageconfigbus.NewPageConfig{
		Name:      "suppliers_page",
		UserID:    uuid.Nil,
		IsDefault: true,
	})
	if err != nil {
		return fmt.Errorf("creating suppliers page: %w", err)
	}

	_, err = busDomain.PageContent.Create(ctx, pagecontentbus.NewPageContent{
		PageConfigID:  suppliersPage.ID,
		ParentID:      uuid.Nil,
		ContentType:   pagecontentbus.ContentTypeTable,
		Label:         "",
		TableConfigID: suppliersTableStored.ID,
		OrderIndex:    1,
		Layout:        json.RawMessage(`{}`),
		IsVisible:     true,
		IsDefault:     true,
	})
	if err != nil {
		return fmt.Errorf("creating suppliers page content: %w", err)
	}

	// Create Categories Page
	categoriesPage, err := busDomain.PageConfig.Create(ctx, pageconfigbus.NewPageConfig{
		Name:      "categories_page",
		UserID:    uuid.Nil,
		IsDefault: true,
	})
	if err != nil {
		return fmt.Errorf("creating categories page: %w", err)
	}

	categoriesPageOrderIndex := 1

	// Add charts to Categories Page
	if pieRevenueCategoryStored != nil {
		_, err = busDomain.PageContent.Create(ctx, pagecontentbus.NewPageContent{
			PageConfigID:  categoriesPage.ID,
			ContentType:   pagecontentbus.ContentTypeChart,
			Label:         "Revenue by Category",
			ChartConfigID: pieRevenueCategoryStored.ID,
			OrderIndex:    categoriesPageOrderIndex,
			Layout:        json.RawMessage(`{"colSpan":{"xs":12,"md":4}}`),
			IsVisible:     true,
			IsDefault:     false,
		})
		if err != nil {
			return fmt.Errorf("creating categories page pie chart: %w", err)
		}
		categoriesPageOrderIndex++
	}

	if barTopProductsStored != nil {
		_, err = busDomain.PageContent.Create(ctx, pagecontentbus.NewPageContent{
			PageConfigID:  categoriesPage.ID,
			ContentType:   pagecontentbus.ContentTypeChart,
			Label:         "Top Products",
			ChartConfigID: barTopProductsStored.ID,
			OrderIndex:    categoriesPageOrderIndex,
			Layout:        json.RawMessage(`{"colSpan":{"xs":12,"md":4}}`),
			IsVisible:     true,
			IsDefault:     false,
		})
		if err != nil {
			return fmt.Errorf("creating categories page bar chart: %w", err)
		}
		categoriesPageOrderIndex++
	}

	if treemapRevenueStored != nil {
		_, err = busDomain.PageContent.Create(ctx, pagecontentbus.NewPageContent{
			PageConfigID:  categoriesPage.ID,
			ContentType:   pagecontentbus.ContentTypeChart,
			Label:         "Revenue Breakdown",
			ChartConfigID: treemapRevenueStored.ID,
			OrderIndex:    categoriesPageOrderIndex,
			Layout:        json.RawMessage(`{"colSpan":{"xs":12,"md":4}}`),
			IsVisible:     true,
			IsDefault:     false,
		})
		if err != nil {
			return fmt.Errorf("creating categories page treemap chart: %w", err)
		}
		categoriesPageOrderIndex++
	}

	_, err = busDomain.PageContent.Create(ctx, pagecontentbus.NewPageContent{
		PageConfigID:  categoriesPage.ID,
		ParentID:      uuid.Nil,
		ContentType:   pagecontentbus.ContentTypeTable,
		Label:         "",
		TableConfigID: categoriesTableStored.ID,
		OrderIndex:    categoriesPageOrderIndex,
		Layout:        json.RawMessage(`{}`),
		IsVisible:     true,
		IsDefault:     true,
	})
	if err != nil {
		return fmt.Errorf("creating categories page content: %w", err)
	}

	// Create Order Line Items Page
	orderLineItemsPage, err := busDomain.PageConfig.Create(ctx, pageconfigbus.NewPageConfig{
		Name:      "order_line_items_page",
		UserID:    uuid.Nil,
		IsDefault: true,
	})
	if err != nil {
		return fmt.Errorf("creating order line items page: %w", err)
	}

	_, err = busDomain.PageContent.Create(ctx, pagecontentbus.NewPageContent{
		PageConfigID:  orderLineItemsPage.ID,
		ParentID:      uuid.Nil,
		ContentType:   pagecontentbus.ContentTypeTable,
		Label:         "",
		TableConfigID: orderLineItemsTableStored.ID,
		OrderIndex:    1,
		Layout:        json.RawMessage(`{}`),
		IsVisible:     true,
		IsDefault:     true,
	})
	if err != nil {
		return fmt.Errorf("creating order line items page content: %w", err)
	}

	// =========================================================================
	// Create Admin Module Pages
	// =========================================================================

	// Admin Users Page
	adminUsersPage, err := busDomain.PageConfig.Create(ctx, pageconfigbus.NewPageConfig{
		Name:      "admin_users_page",
		UserID:    uuid.Nil,
		IsDefault: true,
	})
	if err != nil {
		return fmt.Errorf("creating admin users page: %w", err)
	}

	_, err = busDomain.PageContent.Create(ctx, pagecontentbus.NewPageContent{
		PageConfigID:  adminUsersPage.ID,
		ParentID:      uuid.Nil,
		ContentType:   pagecontentbus.ContentTypeTable,
		Label:         "",
		TableConfigID: adminUsersTableStored.ID,
		OrderIndex:    1,
		Layout:        json.RawMessage(`{}`),
		IsVisible:     true,
		IsDefault:     true,
	})
	if err != nil {
		return fmt.Errorf("creating admin users page content: %w", err)
	}

	// Admin Roles Page
	adminRolesPage, err := busDomain.PageConfig.Create(ctx, pageconfigbus.NewPageConfig{
		Name:      "admin_roles_page",
		UserID:    uuid.Nil,
		IsDefault: true,
	})
	if err != nil {
		return fmt.Errorf("creating admin roles page: %w", err)
	}

	_, err = busDomain.PageContent.Create(ctx, pagecontentbus.NewPageContent{
		PageConfigID:  adminRolesPage.ID,
		ParentID:      uuid.Nil,
		ContentType:   pagecontentbus.ContentTypeTable,
		Label:         "",
		TableConfigID: adminRolesTableStored.ID,
		OrderIndex:    1,
		Layout:        json.RawMessage(`{}`),
		IsVisible:     true,
		IsDefault:     true,
	})
	if err != nil {
		return fmt.Errorf("creating admin roles page content: %w", err)
	}

	// Admin Dashboard Page (multi-tab: users, roles, table access)
	adminDashboardPage, err := busDomain.PageConfig.Create(ctx, pageconfigbus.NewPageConfig{
		Name:      "admin_dashboard_page",
		UserID:    uuid.Nil,
		IsDefault: true,
	})
	if err != nil {
		return fmt.Errorf("creating admin dashboard page: %w", err)
	}

	// Create tabs container (parent)
	adminDashboardTabsContainer, err := busDomain.PageContent.Create(ctx, pagecontentbus.NewPageContent{
		PageConfigID: adminDashboardPage.ID,
		ContentType:  pagecontentbus.ContentTypeTabs,
		OrderIndex:   1,
		Layout:       json.RawMessage(`{"containerType":"tabs"}`),
		IsVisible:    true,
		IsDefault:    false,
	})
	if err != nil {
		return fmt.Errorf("creating admin dashboard tabs container: %w", err)
	}

	// Tab 1: Users
	_, err = busDomain.PageContent.Create(ctx, pagecontentbus.NewPageContent{
		PageConfigID:  adminDashboardPage.ID,
		ParentID:      adminDashboardTabsContainer.ID,
		ContentType:   pagecontentbus.ContentTypeTable,
		Label:         "Users",
		TableConfigID: adminUsersTableStored.ID,
		OrderIndex:    1,
		Layout:        json.RawMessage(`{}`),
		IsVisible:     true,
		IsDefault:     true,
	})
	if err != nil {
		return fmt.Errorf("creating admin dashboard users tab: %w", err)
	}

	// Tab 2: Roles
	_, err = busDomain.PageContent.Create(ctx, pagecontentbus.NewPageContent{
		PageConfigID:  adminDashboardPage.ID,
		ParentID:      adminDashboardTabsContainer.ID,
		ContentType:   pagecontentbus.ContentTypeTable,
		Label:         "Roles",
		TableConfigID: adminRolesTableStored.ID,
		OrderIndex:    2,
		Layout:        json.RawMessage(`{}`),
		IsVisible:     true,
		IsDefault:     false,
	})
	if err != nil {
		return fmt.Errorf("creating admin dashboard roles tab: %w", err)
	}

	// Tab 3: Permissions
	_, err = busDomain.PageContent.Create(ctx, pagecontentbus.NewPageContent{
		PageConfigID:  adminDashboardPage.ID,
		ParentID:      adminDashboardTabsContainer.ID,
		ContentType:   pagecontentbus.ContentTypeTable,
		Label:         "Permissions",
		TableConfigID: adminTableAccessTableStored.ID,
		OrderIndex:    3,
		Layout:        json.RawMessage(`{}`),
		IsVisible:     true,
		IsDefault:     false,
	})
	if err != nil {
		return fmt.Errorf("creating admin dashboard permissions tab: %w", err)
	}

	// Tab 4: Audit Logs
	_, err = busDomain.PageContent.Create(ctx, pagecontentbus.NewPageContent{
		PageConfigID:  adminDashboardPage.ID,
		ParentID:      adminDashboardTabsContainer.ID,
		ContentType:   pagecontentbus.ContentTypeTable,
		Label:         "Audit Logs",
		TableConfigID: adminAuditTableStored.ID,
		OrderIndex:    4,
		Layout:        json.RawMessage(`{}`),
		IsVisible:     true,
		IsDefault:     false,
	})
	if err != nil {
		return fmt.Errorf("creating admin dashboard audit tab: %w", err)
	}

	// Tab 5: Configurations
	_, err = busDomain.PageContent.Create(ctx, pagecontentbus.NewPageContent{
		PageConfigID:  adminDashboardPage.ID,
		ParentID:      adminDashboardTabsContainer.ID,
		ContentType:   pagecontentbus.ContentTypeTable,
		Label:         "Configurations",
		TableConfigID: adminConfigTableStored.ID,
		OrderIndex:    5,
		Layout:        json.RawMessage(`{}`),
		IsVisible:     true,
		IsDefault:     false,
	})
	if err != nil {
		return fmt.Errorf("creating admin dashboard config tab: %w", err)
	}

	// =========================================================================
	// Create Assets Module Pages
	// =========================================================================

	// Assets List Page
	assetsPage, err := busDomain.PageConfig.Create(ctx, pageconfigbus.NewPageConfig{
		Name:      "assets_list_page",
		UserID:    uuid.Nil,
		IsDefault: true,
	})
	if err != nil {
		return fmt.Errorf("creating assets page: %w", err)
	}

	_, err = busDomain.PageContent.Create(ctx, pagecontentbus.NewPageContent{
		PageConfigID:  assetsPage.ID,
		ParentID:      uuid.Nil,
		ContentType:   pagecontentbus.ContentTypeTable,
		Label:         "",
		TableConfigID: assetsListTableStored.ID,
		OrderIndex:    1,
		Layout:        json.RawMessage(`{}`),
		IsVisible:     true,
		IsDefault:     true,
	})
	if err != nil {
		return fmt.Errorf("creating assets page content: %w", err)
	}

	// Asset Requests Page
	assetsRequestsPage, err := busDomain.PageConfig.Create(ctx, pageconfigbus.NewPageConfig{
		Name:      "assets_requests_page",
		UserID:    uuid.Nil,
		IsDefault: true,
	})
	if err != nil {
		return fmt.Errorf("creating assets requests page: %w", err)
	}

	_, err = busDomain.PageContent.Create(ctx, pagecontentbus.NewPageContent{
		PageConfigID:  assetsRequestsPage.ID,
		ParentID:      uuid.Nil,
		ContentType:   pagecontentbus.ContentTypeTable,
		Label:         "",
		TableConfigID: assetsRequestsTableStored.ID,
		OrderIndex:    1,
		Layout:        json.RawMessage(`{}`),
		IsVisible:     true,
		IsDefault:     true,
	})
	if err != nil {
		return fmt.Errorf("creating assets requests page content: %w", err)
	}

	// Assets Dashboard (multi-tab: assets, requests)
	assetsDashboardPage, err := busDomain.PageConfig.Create(ctx, pageconfigbus.NewPageConfig{
		Name:      "assets_dashboard_page",
		UserID:    uuid.Nil,
		IsDefault: true,
	})
	if err != nil {
		return fmt.Errorf("creating assets dashboard page: %w", err)
	}

	// Create tabs container (parent)
	assetsDashboardTabsContainer, err := busDomain.PageContent.Create(ctx, pagecontentbus.NewPageContent{
		PageConfigID: assetsDashboardPage.ID,
		ContentType:  pagecontentbus.ContentTypeTabs,
		OrderIndex:   1,
		Layout:       json.RawMessage(`{"containerType":"tabs"}`),
		IsVisible:    true,
		IsDefault:    false,
	})
	if err != nil {
		return fmt.Errorf("creating assets dashboard tabs container: %w", err)
	}

	// Tab 1: Assets
	_, err = busDomain.PageContent.Create(ctx, pagecontentbus.NewPageContent{
		PageConfigID:  assetsDashboardPage.ID,
		ParentID:      assetsDashboardTabsContainer.ID,
		ContentType:   pagecontentbus.ContentTypeTable,
		Label:         "Assets",
		TableConfigID: assetsListTableStored.ID,
		OrderIndex:    1,
		Layout:        json.RawMessage(`{}`),
		IsVisible:     true,
		IsDefault:     true,
	})
	if err != nil {
		return fmt.Errorf("creating assets dashboard assets tab: %w", err)
	}

	// Tab 2: Requests
	_, err = busDomain.PageContent.Create(ctx, pagecontentbus.NewPageContent{
		PageConfigID:  assetsDashboardPage.ID,
		ParentID:      assetsDashboardTabsContainer.ID,
		ContentType:   pagecontentbus.ContentTypeTable,
		Label:         "Requests",
		TableConfigID: assetsRequestsTableStored.ID,
		OrderIndex:    2,
		Layout:        json.RawMessage(`{}`),
		IsVisible:     true,
		IsDefault:     false,
	})
	if err != nil {
		return fmt.Errorf("creating assets dashboard requests tab: %w", err)
	}

	// =========================================================================
	// Create HR Module Pages
	// =========================================================================

	// HR Employees Page
	hrEmployeesPage, err := busDomain.PageConfig.Create(ctx, pageconfigbus.NewPageConfig{
		Name:      "hr_employees_page",
		UserID:    uuid.Nil,
		IsDefault: true,
	})
	if err != nil {
		return fmt.Errorf("creating hr employees page: %w", err)
	}

	_, err = busDomain.PageContent.Create(ctx, pagecontentbus.NewPageContent{
		PageConfigID:  hrEmployeesPage.ID,
		ParentID:      uuid.Nil,
		ContentType:   pagecontentbus.ContentTypeTable,
		Label:         "",
		TableConfigID: hrEmployeesTableStored.ID,
		OrderIndex:    1,
		Layout:        json.RawMessage(`{}`),
		IsVisible:     true,
		IsDefault:     true,
	})
	if err != nil {
		return fmt.Errorf("creating hr employees page content: %w", err)
	}

	// HR Offices Page
	hrOfficesPage, err := busDomain.PageConfig.Create(ctx, pageconfigbus.NewPageConfig{
		Name:      "hr_offices_page",
		UserID:    uuid.Nil,
		IsDefault: true,
	})
	if err != nil {
		return fmt.Errorf("creating hr offices page: %w", err)
	}

	_, err = busDomain.PageContent.Create(ctx, pagecontentbus.NewPageContent{
		PageConfigID:  hrOfficesPage.ID,
		ParentID:      uuid.Nil,
		ContentType:   pagecontentbus.ContentTypeTable,
		Label:         "",
		TableConfigID: hrOfficesTableStored.ID,
		OrderIndex:    1,
		Layout:        json.RawMessage(`{}`),
		IsVisible:     true,
		IsDefault:     true,
	})
	if err != nil {
		return fmt.Errorf("creating hr offices page content: %w", err)
	}

	// HR Dashboard (multi-tab: employees, offices)
	hrDashboardPage, err := busDomain.PageConfig.Create(ctx, pageconfigbus.NewPageConfig{
		Name:      "hr_dashboard_page",
		UserID:    uuid.Nil,
		IsDefault: true,
	})
	if err != nil {
		return fmt.Errorf("creating hr dashboard page: %w", err)
	}

	// Create tabs container (parent)
	hrDashboardTabsContainer, err := busDomain.PageContent.Create(ctx, pagecontentbus.NewPageContent{
		PageConfigID: hrDashboardPage.ID,
		ContentType:  pagecontentbus.ContentTypeTabs,
		OrderIndex:   1,
		Layout:       json.RawMessage(`{"containerType":"tabs"}`),
		IsVisible:    true,
		IsDefault:    false,
	})
	if err != nil {
		return fmt.Errorf("creating hr dashboard tabs container: %w", err)
	}

	// Tab 1: Employees
	_, err = busDomain.PageContent.Create(ctx, pagecontentbus.NewPageContent{
		PageConfigID:  hrDashboardPage.ID,
		ParentID:      hrDashboardTabsContainer.ID,
		ContentType:   pagecontentbus.ContentTypeTable,
		Label:         "Employees",
		TableConfigID: hrEmployeesTableStored.ID,
		OrderIndex:    1,
		Layout:        json.RawMessage(`{}`),
		IsVisible:     true,
		IsDefault:     true,
	})
	if err != nil {
		return fmt.Errorf("creating hr dashboard employees tab: %w", err)
	}

	// Tab 2: Offices
	_, err = busDomain.PageContent.Create(ctx, pagecontentbus.NewPageContent{
		PageConfigID:  hrDashboardPage.ID,
		ParentID:      hrDashboardTabsContainer.ID,
		ContentType:   pagecontentbus.ContentTypeTable,
		Label:         "Offices",
		TableConfigID: hrOfficesTableStored.ID,
		OrderIndex:    2,
		Layout:        json.RawMessage(`{}`),
		IsVisible:     true,
		IsDefault:     false,
	})
	if err != nil {
		return fmt.Errorf("creating hr dashboard offices tab: %w", err)
	}

	// =========================================================================
	// Create Inventory Module Pages
	// =========================================================================

	// Inventory Warehouses Page
	inventoryWarehousesPage, err := busDomain.PageConfig.Create(ctx, pageconfigbus.NewPageConfig{
		Name:      "inventory_warehouses_page",
		UserID:    uuid.Nil,
		IsDefault: true,
	})
	if err != nil {
		return fmt.Errorf("creating inventory warehouses page: %w", err)
	}

	_, err = busDomain.PageContent.Create(ctx, pagecontentbus.NewPageContent{
		PageConfigID:  inventoryWarehousesPage.ID,
		ParentID:      uuid.Nil,
		ContentType:   pagecontentbus.ContentTypeTable,
		Label:         "",
		TableConfigID: inventoryWarehousesTableStored.ID,
		OrderIndex:    1,
		Layout:        json.RawMessage(`{}`),
		IsVisible:     true,
		IsDefault:     true,
	})
	if err != nil {
		return fmt.Errorf("creating inventory warehouses page content: %w", err)
	}

	// Inventory Items Page
	inventoryItemsPage, err := busDomain.PageConfig.Create(ctx, pageconfigbus.NewPageConfig{
		Name:      "inventory_items_page",
		UserID:    uuid.Nil,
		IsDefault: true,
	})
	if err != nil {
		return fmt.Errorf("creating inventory items page: %w", err)
	}

	_, err = busDomain.PageContent.Create(ctx, pagecontentbus.NewPageContent{
		PageConfigID:  inventoryItemsPage.ID,
		ParentID:      uuid.Nil,
		ContentType:   pagecontentbus.ContentTypeTable,
		Label:         "",
		TableConfigID: inventoryItemsTableStored.ID,
		OrderIndex:    1,
		Layout:        json.RawMessage(`{}`),
		IsVisible:     true,
		IsDefault:     true,
	})
	if err != nil {
		return fmt.Errorf("creating inventory items page content: %w", err)
	}

	// Inventory Adjustments Page
	inventoryAdjustmentsPage, err := busDomain.PageConfig.Create(ctx, pageconfigbus.NewPageConfig{
		Name:      "inventory_adjustments_page",
		UserID:    uuid.Nil,
		IsDefault: true,
	})
	if err != nil {
		return fmt.Errorf("creating inventory adjustments page: %w", err)
	}

	_, err = busDomain.PageContent.Create(ctx, pagecontentbus.NewPageContent{
		PageConfigID:  inventoryAdjustmentsPage.ID,
		ParentID:      uuid.Nil,
		ContentType:   pagecontentbus.ContentTypeTable,
		Label:         "",
		TableConfigID: inventoryAdjustmentsTableStored.ID,
		OrderIndex:    1,
		Layout:        json.RawMessage(`{}`),
		IsVisible:     true,
		IsDefault:     true,
	})
	if err != nil {
		return fmt.Errorf("creating inventory adjustments page content: %w", err)
	}

	// Inventory Transfers Page
	inventoryTransfersPage, err := busDomain.PageConfig.Create(ctx, pageconfigbus.NewPageConfig{
		Name:      "inventory_transfers_page",
		UserID:    uuid.Nil,
		IsDefault: true,
	})
	if err != nil {
		return fmt.Errorf("creating inventory transfers page: %w", err)
	}

	_, err = busDomain.PageContent.Create(ctx, pagecontentbus.NewPageContent{
		PageConfigID:  inventoryTransfersPage.ID,
		ParentID:      uuid.Nil,
		ContentType:   pagecontentbus.ContentTypeTable,
		Label:         "",
		TableConfigID: inventoryTransfersTableStored.ID,
		OrderIndex:    1,
		Layout:        json.RawMessage(`{}`),
		IsVisible:     true,
		IsDefault:     true,
	})
	if err != nil {
		return fmt.Errorf("creating inventory transfers page content: %w", err)
	}

	// Inventory Dashboard (multi-tab: warehouses, items, adjustments, transfers)
	inventoryDashboardPage, err := busDomain.PageConfig.Create(ctx, pageconfigbus.NewPageConfig{
		Name:      "inventory_dashboard_page",
		UserID:    uuid.Nil,
		IsDefault: true,
	})
	if err != nil {
		return fmt.Errorf("creating inventory dashboard page: %w", err)
	}

	inventoryDashboardOrderIndex := 1

	// Add Heatmap chart to Inventory Dashboard (above tabs)
	if heatmapSalesTimeStored != nil {
		_, err = busDomain.PageContent.Create(ctx, pagecontentbus.NewPageContent{
			PageConfigID:  inventoryDashboardPage.ID,
			ContentType:   pagecontentbus.ContentTypeChart,
			Label:         "Orders by Day and Hour",
			ChartConfigID: heatmapSalesTimeStored.ID,
			OrderIndex:    inventoryDashboardOrderIndex,
			Layout:        json.RawMessage(`{"colSpan":{"xs":12}}`),
			IsVisible:     true,
			IsDefault:     false,
		})
		if err != nil {
			return fmt.Errorf("creating inventory dashboard heatmap chart: %w", err)
		}
		inventoryDashboardOrderIndex++
	}

	// Create tabs container (parent)
	inventoryDashboardTabsContainer, err := busDomain.PageContent.Create(ctx, pagecontentbus.NewPageContent{
		PageConfigID: inventoryDashboardPage.ID,
		ContentType:  pagecontentbus.ContentTypeTabs,
		OrderIndex:   inventoryDashboardOrderIndex,
		Layout:       json.RawMessage(`{"containerType":"tabs"}`),
		IsVisible:    true,
		IsDefault:    false,
	})
	if err != nil {
		return fmt.Errorf("creating inventory dashboard tabs container: %w", err)
	}

	// Tab 1: Items
	_, err = busDomain.PageContent.Create(ctx, pagecontentbus.NewPageContent{
		PageConfigID:  inventoryDashboardPage.ID,
		ParentID:      inventoryDashboardTabsContainer.ID,
		ContentType:   pagecontentbus.ContentTypeTable,
		Label:         "Items",
		TableConfigID: inventoryItemsTableStored.ID,
		OrderIndex:    1,
		Layout:        json.RawMessage(`{}`),
		IsVisible:     true,
		IsDefault:     true,
	})
	if err != nil {
		return fmt.Errorf("creating inventory dashboard items tab: %w", err)
	}

	// Tab 2: Warehouses
	_, err = busDomain.PageContent.Create(ctx, pagecontentbus.NewPageContent{
		PageConfigID:  inventoryDashboardPage.ID,
		ParentID:      inventoryDashboardTabsContainer.ID,
		ContentType:   pagecontentbus.ContentTypeTable,
		Label:         "Warehouses",
		TableConfigID: inventoryWarehousesTableStored.ID,
		OrderIndex:    2,
		Layout:        json.RawMessage(`{}`),
		IsVisible:     true,
		IsDefault:     false,
	})
	if err != nil {
		return fmt.Errorf("creating inventory dashboard warehouses tab: %w", err)
	}

	// Tab 3: Adjustments
	_, err = busDomain.PageContent.Create(ctx, pagecontentbus.NewPageContent{
		PageConfigID:  inventoryDashboardPage.ID,
		ParentID:      inventoryDashboardTabsContainer.ID,
		ContentType:   pagecontentbus.ContentTypeTable,
		Label:         "Adjustments",
		TableConfigID: inventoryAdjustmentsTableStored.ID,
		OrderIndex:    3,
		Layout:        json.RawMessage(`{}`),
		IsVisible:     true,
		IsDefault:     false,
	})
	if err != nil {
		return fmt.Errorf("creating inventory dashboard adjustments tab: %w", err)
	}

	// Tab 4: Transfers
	_, err = busDomain.PageContent.Create(ctx, pagecontentbus.NewPageContent{
		PageConfigID:  inventoryDashboardPage.ID,
		ParentID:      inventoryDashboardTabsContainer.ID,
		ContentType:   pagecontentbus.ContentTypeTable,
		Label:         "Transfers",
		TableConfigID: inventoryTransfersTableStored.ID,
		OrderIndex:    4,
		Layout:        json.RawMessage(`{}`),
		IsVisible:     true,
		IsDefault:     false,
	})
	if err != nil {
		return fmt.Errorf("creating inventory dashboard transfers tab: %w", err)
	}

	// =========================================================================
	// Create Sales Module Pages
	// =========================================================================

	// Sales Customers Page
	salesCustomersPage, err := busDomain.PageConfig.Create(ctx, pageconfigbus.NewPageConfig{
		Name:      "sales_customers_page",
		UserID:    uuid.Nil,
		IsDefault: true,
	})
	if err != nil {
		return fmt.Errorf("creating sales customers page: %w", err)
	}

	_, err = busDomain.PageContent.Create(ctx, pagecontentbus.NewPageContent{
		PageConfigID:  salesCustomersPage.ID,
		ParentID:      uuid.Nil,
		ContentType:   pagecontentbus.ContentTypeTable,
		Label:         "",
		TableConfigID: salesCustomersTableStored.ID,
		OrderIndex:    1,
		Layout:        json.RawMessage(`{}`),
		IsVisible:     true,
		IsDefault:     true,
	})
	if err != nil {
		return fmt.Errorf("creating sales customers page content: %w", err)
	}

	// Sales Dashboard (multi-tab: orders, customers)
	salesDashboardPage, err := busDomain.PageConfig.Create(ctx, pageconfigbus.NewPageConfig{
		Name:      "sales_dashboard_page",
		UserID:    uuid.Nil,
		IsDefault: true,
	})
	if err != nil {
		return fmt.Errorf("creating sales dashboard page: %w", err)
	}

	salesDashboardOrderIndex := 1

	// Add KPI charts row to Sales Dashboard (above tabs)
	if kpiRevenueStored != nil {
		_, err = busDomain.PageContent.Create(ctx, pagecontentbus.NewPageContent{
			PageConfigID:  salesDashboardPage.ID,
			ContentType:   pagecontentbus.ContentTypeChart,
			Label:         "Total Revenue",
			ChartConfigID: kpiRevenueStored.ID,
			OrderIndex:    salesDashboardOrderIndex,
			Layout:        json.RawMessage(`{"colSpan":{"xs":12,"sm":6,"md":4}}`),
			IsVisible:     true,
			IsDefault:     false,
		})
		if err != nil {
			return fmt.Errorf("creating sales dashboard KPI revenue chart: %w", err)
		}
		salesDashboardOrderIndex++
	}

	if kpiOrdersStored != nil {
		_, err = busDomain.PageContent.Create(ctx, pagecontentbus.NewPageContent{
			PageConfigID:  salesDashboardPage.ID,
			ContentType:   pagecontentbus.ContentTypeChart,
			Label:         "Total Orders",
			ChartConfigID: kpiOrdersStored.ID,
			OrderIndex:    salesDashboardOrderIndex,
			Layout:        json.RawMessage(`{"colSpan":{"xs":12,"sm":6,"md":4}}`),
			IsVisible:     true,
			IsDefault:     false,
		})
		if err != nil {
			return fmt.Errorf("creating sales dashboard KPI orders chart: %w", err)
		}
		salesDashboardOrderIndex++
	}

	if gaugeRevenueStored != nil {
		_, err = busDomain.PageContent.Create(ctx, pagecontentbus.NewPageContent{
			PageConfigID:  salesDashboardPage.ID,
			ContentType:   pagecontentbus.ContentTypeChart,
			Label:         "Revenue Progress",
			ChartConfigID: gaugeRevenueStored.ID,
			OrderIndex:    salesDashboardOrderIndex,
			Layout:        json.RawMessage(`{"colSpan":{"xs":12,"sm":6,"md":4}}`),
			IsVisible:     true,
			IsDefault:     false,
		})
		if err != nil {
			return fmt.Errorf("creating sales dashboard gauge chart: %w", err)
		}
		salesDashboardOrderIndex++
	}

	// Create tabs container (parent)
	salesDashboardTabsContainer, err := busDomain.PageContent.Create(ctx, pagecontentbus.NewPageContent{
		PageConfigID: salesDashboardPage.ID,
		ContentType:  pagecontentbus.ContentTypeTabs,
		OrderIndex:   salesDashboardOrderIndex,
		Layout:       json.RawMessage(`{"containerType":"tabs"}`),
		IsVisible:    true,
		IsDefault:    false,
	})
	if err != nil {
		return fmt.Errorf("creating sales dashboard tabs container: %w", err)
	}

	// Tab 1: Orders
	_, err = busDomain.PageContent.Create(ctx, pagecontentbus.NewPageContent{
		PageConfigID:  salesDashboardPage.ID,
		ParentID:      salesDashboardTabsContainer.ID,
		ContentType:   pagecontentbus.ContentTypeTable,
		Label:         "Orders",
		TableConfigID: ordersTableStored.ID,
		OrderIndex:    1,
		Layout:        json.RawMessage(`{}`),
		IsVisible:     true,
		IsDefault:     true,
	})
	if err != nil {
		return fmt.Errorf("creating sales dashboard orders tab: %w", err)
	}

	// Tab 2: Customers
	_, err = busDomain.PageContent.Create(ctx, pagecontentbus.NewPageContent{
		PageConfigID:  salesDashboardPage.ID,
		ParentID:      salesDashboardTabsContainer.ID,
		ContentType:   pagecontentbus.ContentTypeTable,
		Label:         "Customers",
		TableConfigID: salesCustomersTableStored.ID,
		OrderIndex:    2,
		Layout:        json.RawMessage(`{}`),
		IsVisible:     true,
		IsDefault:     false,
	})
	if err != nil {
		return fmt.Errorf("creating sales dashboard customers tab: %w", err)
	}

	// =========================================================================
	// Create Procurement Module Pages
	// =========================================================================

	// Purchase Orders Page
	procurementPurchaseOrdersPage, err := busDomain.PageConfig.Create(ctx, pageconfigbus.NewPageConfig{
		Name:      "procurement_purchase_orders",
		UserID:    uuid.Nil,
		IsDefault: true,
	})
	if err != nil {
		return fmt.Errorf("creating procurement purchase orders page: %w", err)
	}

	_, err = busDomain.PageContent.Create(ctx, pagecontentbus.NewPageContent{
		PageConfigID:  procurementPurchaseOrdersPage.ID,
		ParentID:      uuid.Nil,
		ContentType:   pagecontentbus.ContentTypeTable,
		Label:         "",
		TableConfigID: procurementPurchaseOrdersConfigStored.ID,
		OrderIndex:    1,
		Layout:        json.RawMessage(`{}`),
		IsVisible:     true,
		IsDefault:     true,
	})
	if err != nil {
		return fmt.Errorf("creating procurement purchase orders page content: %w", err)
	}

	// Purchase Order Line Items Page
	procurementLineItemsPage, err := busDomain.PageConfig.Create(ctx, pageconfigbus.NewPageConfig{
		Name:      "procurement_line_items",
		UserID:    uuid.Nil,
		IsDefault: true,
	})
	if err != nil {
		return fmt.Errorf("creating procurement line items page: %w", err)
	}

	_, err = busDomain.PageContent.Create(ctx, pagecontentbus.NewPageContent{
		PageConfigID:  procurementLineItemsPage.ID,
		ParentID:      uuid.Nil,
		ContentType:   pagecontentbus.ContentTypeTable,
		Label:         "",
		TableConfigID: procurementLineItemsConfigStored.ID,
		OrderIndex:    1,
		Layout:        json.RawMessage(`{}`),
		IsVisible:     true,
		IsDefault:     true,
	})
	if err != nil {
		return fmt.Errorf("creating procurement line items page content: %w", err)
	}

	// Procurement Approvals Page (multi-tab: open, closed)
	procurementApprovalsPage, err := busDomain.PageConfig.Create(ctx, pageconfigbus.NewPageConfig{
		Name:      "procurement_approvals",
		UserID:    uuid.Nil,
		IsDefault: true,
	})
	if err != nil {
		return fmt.Errorf("creating procurement approvals page: %w", err)
	}

	// Create tabs container (parent)
	procurementApprovalsTabsContainer, err := busDomain.PageContent.Create(ctx, pagecontentbus.NewPageContent{
		PageConfigID: procurementApprovalsPage.ID,
		ContentType:  pagecontentbus.ContentTypeTabs,
		OrderIndex:   1,
		Layout:       json.RawMessage(`{"containerType":"tabs"}`),
		IsVisible:    true,
		IsDefault:    false,
	})
	if err != nil {
		return fmt.Errorf("creating procurement approvals tabs container: %w", err)
	}

	// Tab 1: Open
	_, err = busDomain.PageContent.Create(ctx, pagecontentbus.NewPageContent{
		PageConfigID:  procurementApprovalsPage.ID,
		ParentID:      procurementApprovalsTabsContainer.ID,
		ContentType:   pagecontentbus.ContentTypeTable,
		Label:         "Open",
		TableConfigID: procurementApprovalsOpenConfigStored.ID,
		OrderIndex:    1,
		Layout:        json.RawMessage(`{}`),
		IsVisible:     true,
		IsDefault:     true,
	})
	if err != nil {
		return fmt.Errorf("creating procurement approvals open tab: %w", err)
	}

	// Tab 2: Closed
	_, err = busDomain.PageContent.Create(ctx, pagecontentbus.NewPageContent{
		PageConfigID:  procurementApprovalsPage.ID,
		ParentID:      procurementApprovalsTabsContainer.ID,
		ContentType:   pagecontentbus.ContentTypeTable,
		Label:         "Closed",
		TableConfigID: procurementApprovalsClosedConfigStored.ID,
		OrderIndex:    2,
		Layout:        json.RawMessage(`{}`),
		IsVisible:     true,
		IsDefault:     false,
	})
	if err != nil {
		return fmt.Errorf("creating procurement approvals closed tab: %w", err)
	}

	// Procurement Dashboard (multi-tab: purchase orders, line items, suppliers, approvals)
	procurementDashboardPage, err := busDomain.PageConfig.Create(ctx, pageconfigbus.NewPageConfig{
		Name:      "procurement_dashboard",
		UserID:    uuid.Nil,
		IsDefault: true,
	})
	if err != nil {
		return fmt.Errorf("creating procurement dashboard page: %w", err)
	}

	procurementDashboardOrderIndex := 1

	// Add Gantt chart to Procurement Dashboard (above tabs)
	if ganttProjectStored != nil {
		_, err = busDomain.PageContent.Create(ctx, pagecontentbus.NewPageContent{
			PageConfigID:  procurementDashboardPage.ID,
			ContentType:   pagecontentbus.ContentTypeChart,
			Label:         "Purchase Order Timeline",
			ChartConfigID: ganttProjectStored.ID,
			OrderIndex:    procurementDashboardOrderIndex,
			Layout:        json.RawMessage(`{"colSpan":{"xs":12}}`),
			IsVisible:     true,
			IsDefault:     false,
		})
		if err != nil {
			return fmt.Errorf("creating procurement dashboard gantt chart: %w", err)
		}
		procurementDashboardOrderIndex++
	}

	// Create tabs container (parent)
	procurementDashboardTabsContainer, err := busDomain.PageContent.Create(ctx, pagecontentbus.NewPageContent{
		PageConfigID: procurementDashboardPage.ID,
		ContentType:  pagecontentbus.ContentTypeTabs,
		OrderIndex:   procurementDashboardOrderIndex,
		Layout:       json.RawMessage(`{"containerType":"tabs"}`),
		IsVisible:    true,
		IsDefault:    false,
	})
	if err != nil {
		return fmt.Errorf("creating procurement dashboard tabs container: %w", err)
	}

	// Tab 1: Purchase Orders
	_, err = busDomain.PageContent.Create(ctx, pagecontentbus.NewPageContent{
		PageConfigID:  procurementDashboardPage.ID,
		ParentID:      procurementDashboardTabsContainer.ID,
		ContentType:   pagecontentbus.ContentTypeTable,
		Label:         "Purchase Orders",
		TableConfigID: procurementPurchaseOrdersConfigStored.ID,
		OrderIndex:    1,
		Layout:        json.RawMessage(`{}`),
		IsVisible:     true,
		IsDefault:     true,
	})
	if err != nil {
		return fmt.Errorf("creating procurement dashboard purchase orders tab: %w", err)
	}

	// Tab 2: Line Items
	_, err = busDomain.PageContent.Create(ctx, pagecontentbus.NewPageContent{
		PageConfigID:  procurementDashboardPage.ID,
		ParentID:      procurementDashboardTabsContainer.ID,
		ContentType:   pagecontentbus.ContentTypeTable,
		Label:         "Line Items",
		TableConfigID: procurementLineItemsConfigStored.ID,
		OrderIndex:    2,
		Layout:        json.RawMessage(`{}`),
		IsVisible:     true,
		IsDefault:     false,
	})
	if err != nil {
		return fmt.Errorf("creating procurement dashboard line items tab: %w", err)
	}

	// Tab 3: Suppliers
	_, err = busDomain.PageContent.Create(ctx, pagecontentbus.NewPageContent{
		PageConfigID:  procurementDashboardPage.ID,
		ParentID:      procurementDashboardTabsContainer.ID,
		ContentType:   pagecontentbus.ContentTypeTable,
		Label:         "Suppliers",
		TableConfigID: suppliersTableStored.ID,
		OrderIndex:    3,
		Layout:        json.RawMessage(`{}`),
		IsVisible:     true,
		IsDefault:     false,
	})
	if err != nil {
		return fmt.Errorf("creating procurement dashboard suppliers tab: %w", err)
	}

	// Tab 4: Approvals
	_, err = busDomain.PageContent.Create(ctx, pagecontentbus.NewPageContent{
		PageConfigID:  procurementDashboardPage.ID,
		ParentID:      procurementDashboardTabsContainer.ID,
		ContentType:   pagecontentbus.ContentTypeTable,
		Label:         "Approvals",
		TableConfigID: procurementApprovalsOpenConfigStored.ID,
		OrderIndex:    4,
		Layout:        json.RawMessage(`{}`),
		IsVisible:     true,
		IsDefault:     false,
	})
	if err != nil {
		return fmt.Errorf("creating procurement dashboard approvals tab: %w", err)
	}

	// =========================================================================
	// Main Dashboard (Testing Ground for UI Elements)
	// =========================================================================

	mainDashboardPage, err := busDomain.PageConfig.Create(ctx, pageconfigbus.NewPageConfig{
		Name:      "main_dashboard_page",
		UserID:    uuid.Nil,
		IsDefault: true,
	})
	if err != nil {
		return fmt.Errorf("creating main dashboard page: %w", err)
	}

	log.Info(ctx, "✅ Created Main Dashboard page (testing ground)",
		"page_config_id", mainDashboardPage.ID)

	// =========================================================================
	// Seed Page Action Buttons
	// =========================================================================

	pageConfigIDs := map[string]uuid.UUID{
		"admin_users_page":            adminUsersPage.ID,
		"admin_roles_page":            adminRolesPage.ID,
		"assets_list_page":            assetsPage.ID,
		"hr_employees_page":           hrEmployeesPage.ID,
		"hr_offices_page":             hrOfficesPage.ID,
		"inventory_items_page":        inventoryItemsPage.ID,
		"inventory_warehouses_page":   inventoryWarehousesPage.ID,
		"inventory_transfers_page":    inventoryTransfersPage.ID,
		"inventory_adjustments_page":  inventoryAdjustmentsPage.ID,
		"suppliers_page":              suppliersPage.ID,
		"procurement_purchase_orders": procurementPurchaseOrdersPage.ID,
		"sales_customers_page":        salesCustomersPage.ID,
		"orders_page":                 ordersPage.ID,
		"sales_dashboard_page":        salesDashboardPage.ID,
		"main_dashboard_page":         mainDashboardPage.ID,
	}

	if err := seedPageActionButtons(ctx, log, busDomain, pageConfigIDs); err != nil {
		return fmt.Errorf("seeding page action buttons: %w", err)
	}

	// PAGES
	var pageIDs uuid.UUIDs

	for _, p := range seedmodels.AllPages {
		created, err := busDomain.Page.Create(ctx, p)
		if err != nil {
			return fmt.Errorf("creating page %s : %w", p.Name, err)
		}
		pageIDs = append(pageIDs, created.ID)
	}

	// all user roles
	urs, err := busDomain.UserRole.Query(ctx, userrolebus.QueryFilter{}, userrolebus.DefaultOrderBy, page.MustParse("1", "100"))
	if err != nil {
		return fmt.Errorf("querying user roles : %w", err)
	}

	r, err := busDomain.Role.QueryByID(ctx, urs[0].RoleID)
	if err != nil {
		return fmt.Errorf("querying role : %w", err)
	}

	// Add all pages to role
	for i := range seedmodels.AllPages {
		_, err = busDomain.RolePage.Create(ctx, rolepagebus.NewRolePage{
			RoleID:    r.ID,
			PageID:    pageIDs[i],
			CanAccess: true,
		})
		if err != nil {
			return fmt.Errorf("creating role-page association : %w", err)
		}
	}

	return nil
}

// seedPageActionButtons creates button actions for pages.
// Supports multiple buttons per page with full customization.
func seedPageActionButtons(ctx context.Context, log *logger.Logger, busDomain BusDomain, pageConfigIDs map[string]uuid.UUID) error {
	// Get button definitions (now returns arrays)
	buttonDefs := seedmodels.GetNewButtonActionDefinitions()

	totalButtonsCreated := 0

	// Create button actions for each page config
	for configName, pageConfigID := range pageConfigIDs {
		buttonDefArray, exists := buttonDefs[configName]
		if !exists || len(buttonDefArray) == 0 {
			// Skip if no button definitions exist for this page config
			continue
		}

		// Create each button for this page
		for i, buttonDef := range buttonDefArray {
			// Validate required fields
			if buttonDef.Label == "" {
				return fmt.Errorf("button %d for page config %s: label is required", i, configName)
			}
			if buttonDef.TargetPath == "" {
				return fmt.Errorf("button %d for page config %s: target path is required", i, configName)
			}
			if buttonDef.ActionOrder <= 0 {
				return fmt.Errorf("button %d for page config %s: action order must be positive (got %d)", i, configName, buttonDef.ActionOrder)
			}

			// Create button action
			buttonAction := seedmodels.CreateNewButtonAction(pageConfigID, buttonDef)

			_, err := busDomain.PageAction.CreateButton(ctx, buttonAction)
			if err != nil {
				return fmt.Errorf("creating button action %d (%s) for %s: %w",
					i, buttonDef.Label, configName, err)
			}

			totalButtonsCreated++
		}
	}

	log.Info(ctx, "✅ Created page action buttons",
		"total", totalButtonsCreated)

	return nil
}
