package data

// validTables is a map-based whitelist of table names that workflow actions
// can interact with. Map lookup is O(1) vs O(n) slice iteration.
var validTables = map[string]bool{
	// sales schema
	"sales.customers":                      true,
	"sales.orders":                         true,
	"sales.order_line_items":               true,
	"sales.order_fulfillment_statuses":     true,
	"sales.line_item_fulfillment_statuses": true,
	// products schema
	"products.products":            true,
	"products.brands":              true,
	"products.product_categories":  true,
	"products.physical_attributes": true,
	"products.product_costs":       true,
	"products.cost_history":        true,
	"products.quality_metrics":     true,
	// inventory schema
	"inventory.inventory_items":        true,
	"inventory.inventory_locations":    true,
	"inventory.inventory_transactions": true,
	"inventory.warehouses":             true,
	"inventory.zones":                  true,
	"inventory.lot_trackings":          true,
	"inventory.serial_numbers":         true,
	"inventory.inspections":            true,
	"inventory.inventory_adjustments":  true,
	"inventory.transfer_orders":        true,
	// procurement schema
	"procurement.suppliers":                          true,
	"procurement.supplier_products":                  true,
	"procurement.purchase_orders":                    true,
	"procurement.purchase_order_line_items":          true,
	"procurement.purchase_order_statuses":            true,
	"procurement.purchase_order_line_item_statuses":  true,
	// core schema
	"core.users":        true,
	"core.roles":        true,
	"core.user_roles":   true,
	"core.contact_infos": true,
	"core.table_access": true,
	// hr schema
	"hr.offices":                  true,
	"hr.titles":                   true,
	"hr.reports_to":               true,
	"hr.homes":                    true,
	"hr.user_approval_status":     true,
	"hr.user_approval_comments":   true,
	// geography schema
	"geography.countries": true,
	"geography.regions":   true,
	"geography.cities":    true,
	"geography.streets":   true,
	// assets schema
	"assets.assets":            true,
	"assets.valid_assets":      true,
	"assets.user_assets":       true,
	"assets.asset_conditions":  true,
	"assets.asset_types":       true,
	"assets.asset_tags":        true,
	// config schema
	"config.table_configs": true,
	// workflow schema
	"workflow.automation_rules":        true,
	"workflow.rule_actions":            true,
	"workflow.action_templates":        true,
	"workflow.rule_dependencies":       true,
	"workflow.trigger_types":           true,
	"workflow.entity_types":            true,
	"workflow.entities":                true,
	"workflow.automation_executions":   true,
	"workflow.notification_deliveries": true,
}

// IsValidTableName validates table names against the shared whitelist.
func IsValidTableName(tableName string) bool {
	return validTables[tableName]
}

// validOperators is a map-based set of valid condition operators.
var validOperators = map[string]bool{
	"equals":      true,
	"not_equals":  true,
	"greater_than": true,
	"less_than":   true,
	"contains":    true,
	"is_null":     true,
	"is_not_null": true,
	"in":          true,
	"not_in":      true,
}

// IsValidOperator validates condition operators against the shared set.
func IsValidOperator(operator string) bool {
	return validOperators[operator]
}
