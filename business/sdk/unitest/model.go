package unitest

import (
	"context"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/assets/approvalstatusbus"
	"github.com/timmaaaz/ichor/business/domain/assets/assetbus"
	"github.com/timmaaaz/ichor/business/domain/assets/assetconditionbus"
	"github.com/timmaaaz/ichor/business/domain/assets/assettagbus"
	"github.com/timmaaaz/ichor/business/domain/assets/assettypebus"
	"github.com/timmaaaz/ichor/business/domain/assets/fulfillmentstatusbus"
	"github.com/timmaaaz/ichor/business/domain/assets/tagbus"
	"github.com/timmaaaz/ichor/business/domain/assets/userassetbus"
	"github.com/timmaaaz/ichor/business/domain/assets/validassetbus"
	"github.com/timmaaaz/ichor/business/domain/core/contactinfosbus"
	"github.com/timmaaaz/ichor/business/domain/core/permissionsbus"
	"github.com/timmaaaz/ichor/business/domain/core/rolebus"
	"github.com/timmaaaz/ichor/business/domain/core/tableaccessbus"
	"github.com/timmaaaz/ichor/business/domain/core/userbus"
	"github.com/timmaaaz/ichor/business/domain/core/userrolebus"
	"github.com/timmaaaz/ichor/business/domain/geography/citybus"
	"github.com/timmaaaz/ichor/business/domain/geography/countrybus"
	"github.com/timmaaaz/ichor/business/domain/geography/regionbus"
	"github.com/timmaaaz/ichor/business/domain/geography/streetbus"
	"github.com/timmaaaz/ichor/business/domain/geography/timezonebus"
	"github.com/timmaaaz/ichor/business/domain/hr/approvalbus"
	"github.com/timmaaaz/ichor/business/domain/hr/commentbus"
	"github.com/timmaaaz/ichor/business/domain/hr/homebus"
	"github.com/timmaaaz/ichor/business/domain/hr/officebus"
	"github.com/timmaaaz/ichor/business/domain/hr/reportstobus"
	"github.com/timmaaaz/ichor/business/domain/hr/titlebus"
	"github.com/timmaaaz/ichor/business/domain/inventory/inspectionbus"
	"github.com/timmaaaz/ichor/business/domain/inventory/inventoryadjustmentbus"
	"github.com/timmaaaz/ichor/business/domain/inventory/inventoryitembus"
	"github.com/timmaaaz/ichor/business/domain/inventory/inventorylocationbus"
	"github.com/timmaaaz/ichor/business/domain/inventory/inventorytransactionbus"
	"github.com/timmaaaz/ichor/business/domain/inventory/lottrackingsbus"
	"github.com/timmaaaz/ichor/business/domain/inventory/serialnumberbus"
	"github.com/timmaaaz/ichor/business/domain/inventory/transferorderbus"
	"github.com/timmaaaz/ichor/business/domain/inventory/warehousebus"
	"github.com/timmaaaz/ichor/business/domain/inventory/zonebus"
	"github.com/timmaaaz/ichor/business/domain/procurement/purchaseorderbus"
	"github.com/timmaaaz/ichor/business/domain/procurement/purchaseorderlineitembus"
	"github.com/timmaaaz/ichor/business/domain/procurement/purchaseorderlineitemstatusbus"
	"github.com/timmaaaz/ichor/business/domain/procurement/purchaseorderstatusbus"
	"github.com/timmaaaz/ichor/business/domain/procurement/supplierbus"
	"github.com/timmaaaz/ichor/business/domain/procurement/supplierproductbus"
	"github.com/timmaaaz/ichor/business/domain/products/brandbus"
	"github.com/timmaaaz/ichor/business/domain/products/costhistorybus"
	"github.com/timmaaaz/ichor/business/domain/products/metricsbus"
	"github.com/timmaaaz/ichor/business/domain/products/physicalattributebus"
	"github.com/timmaaaz/ichor/business/domain/products/productbus"
	"github.com/timmaaaz/ichor/business/domain/products/productcategorybus"
	"github.com/timmaaaz/ichor/business/domain/products/productcostbus"
	"github.com/timmaaaz/ichor/business/domain/sales/customersbus"
	"github.com/timmaaaz/ichor/business/domain/sales/lineitemfulfillmentstatusbus"
	"github.com/timmaaaz/ichor/business/domain/sales/orderfulfillmentstatusbus"
	"github.com/timmaaaz/ichor/business/domain/sales/orderlineitemsbus"
	"github.com/timmaaaz/ichor/business/domain/sales/ordersbus"
	"github.com/timmaaaz/ichor/business/sdk/tablebuilder"

	"github.com/timmaaaz/ichor/business/domain/config/formbus"
	"github.com/timmaaaz/ichor/business/domain/config/formfieldbus"
	"github.com/timmaaaz/ichor/business/domain/config/pageactionbus"
	"github.com/timmaaaz/ichor/business/domain/config/pageconfigbus"
	"github.com/timmaaaz/ichor/business/domain/config/pagecontentbus"
)

// User represents an app user specified for the test.
type User struct {
	userbus.User
	Homes []homebus.Home
}

// SeedData represents data that was seeded for the test.
type SeedData struct {
	Users                         []User
	Admins                        []User
	AssetConditions               []assetconditionbus.AssetCondition
	ValidAssets                   []validassetbus.ValidAsset
	Countries                     []countrybus.Country
	Regions                       []regionbus.Region
	Cities                        []citybus.City
	Streets                       []streetbus.Street
	Timezones                     []timezonebus.Timezone
	ApprovalStatus                []approvalstatusbus.ApprovalStatus
	UserApprovalStatus            []approvalbus.UserApprovalStatus
	UserApprovalComment           []commentbus.UserApprovalComment
	FulfillmentStatus             []fulfillmentstatusbus.FulfillmentStatus
	AssetCondition                []assetconditionbus.AssetCondition
	AssetTypes                    []assettypebus.AssetType
	Tags                          []tagbus.Tag
	AssetTags                     []assettagbus.AssetTag
	Title                         []titlebus.Title
	ReportsTo                     []reportstobus.ReportsTo
	Offices                       []officebus.Office
	UserAssets                    []userassetbus.UserAsset
	Assets                        []assetbus.Asset
	ContactInfos                  []contactinfosbus.ContactInfos
	Brands                        []brandbus.Brand
	ProductCategories             []productcategorybus.ProductCategory
	Warehouses                    []warehousebus.Warehouse
	Roles                         []rolebus.Role
	UserRoles                     []userrolebus.UserRole
	TableAccesses                 []tableaccessbus.TableAccess
	UserPermissions               []permissionsbus.UserPermissions
	Products                      []productbus.Product
	PhysicalAttributes            []physicalattributebus.PhysicalAttribute
	ProductCosts                  []productcostbus.ProductCost
	Suppliers                     []supplierbus.Supplier
	CostHistory                   []costhistorybus.CostHistory
	SupplierProducts              []supplierproductbus.SupplierProduct
	PurchaseOrderStatuses         []purchaseorderstatusbus.PurchaseOrderStatus
	PurchaseOrderLineItemStatuses []purchaseorderlineitemstatusbus.PurchaseOrderLineItemStatus
	PurchaseOrders                []purchaseorderbus.PurchaseOrder
	PurchaseOrderLineItems        []purchaseorderlineitembus.PurchaseOrderLineItem
	Metrics                       []metricsbus.Metric
	LotTrackings                  []lottrackingsbus.LotTrackings
	Zones                         []zonebus.Zone
	InventoryLocations            []inventorylocationbus.InventoryLocation
	InventoryItems                []inventoryitembus.InventoryItem
	Inspections                   []inspectionbus.Inspection
	SerialNumbers                 []serialnumberbus.SerialNumber
	InventoryTransactions         []inventorytransactionbus.InventoryTransaction
	InventoryAdjustments          []inventoryadjustmentbus.InventoryAdjustment
	TransferOrders                []transferorderbus.TransferOrder
	OrderFulfillmentStatuses      []orderfulfillmentstatusbus.OrderFulfillmentStatus
	LineItemFulfillmentStatuses   []lineitemfulfillmentstatusbus.LineItemFulfillmentStatus
	Customers                     []customersbus.Customers
	Orders                        []ordersbus.Order
	OrderLineItems                []orderlineitemsbus.OrderLineItem
	TableBuilderConfigs           []tablebuilder.StoredConfig
	Forms                         []formbus.Form
	FormFields                    []formfieldbus.FormField
	PageActions                   []pageactionbus.PageAction
	PageConfigs                   []pageconfigbus.PageConfig
	PageContents                  []pagecontentbus.PageContent
	PageConfigIDs                 []uuid.UUID
}

type Table struct {
	Name    string
	ExpResp any
	ExcFunc func(ctx context.Context) any
	CmpFunc func(got any, exp any) string
}
