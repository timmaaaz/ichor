// Package workflowdomains is the single source of truth mapping domain buses to the
// workflow delegate subscriber. It exists so the in-process server (all.go) AND the
// standalone Temporal worker register the SAME set of (domain, entity) pairs, and so
// the generic write handlers (P4 M1) can resolve a schema-qualified target to the
// delegate domain + bare entity name to fire under (DESIGN §6 — one declared list,
// three consumers: the all.go RegisterDomain loop, the worker RegisterDomain loop, and
// the handler reverse map).
package workflowdomains

import (
	"github.com/timmaaaz/ichor/business/domain/assets/approvalstatusbus"
	"github.com/timmaaaz/ichor/business/domain/assets/assetbus"
	"github.com/timmaaaz/ichor/business/domain/assets/assetconditionbus"
	"github.com/timmaaaz/ichor/business/domain/assets/assettagbus"
	"github.com/timmaaaz/ichor/business/domain/assets/assettypebus"
	"github.com/timmaaaz/ichor/business/domain/assets/fulfillmentstatusbus"
	"github.com/timmaaaz/ichor/business/domain/assets/tagbus"
	"github.com/timmaaaz/ichor/business/domain/assets/userassetbus"
	"github.com/timmaaaz/ichor/business/domain/assets/validassetbus"
	"github.com/timmaaaz/ichor/business/domain/config/formbus"
	"github.com/timmaaaz/ichor/business/domain/config/formfieldbus"
	"github.com/timmaaaz/ichor/business/domain/config/pageactionbus"
	"github.com/timmaaaz/ichor/business/domain/config/pageconfigbus"
	"github.com/timmaaaz/ichor/business/domain/config/pagecontentbus"
	"github.com/timmaaaz/ichor/business/domain/core/contactinfosbus"
	"github.com/timmaaaz/ichor/business/domain/core/currencybus"
	"github.com/timmaaaz/ichor/business/domain/core/pagebus"
	"github.com/timmaaaz/ichor/business/domain/core/paymenttermbus"
	"github.com/timmaaaz/ichor/business/domain/core/rolebus"
	"github.com/timmaaaz/ichor/business/domain/core/rolepagebus"
	"github.com/timmaaaz/ichor/business/domain/core/tableaccessbus"
	"github.com/timmaaaz/ichor/business/domain/core/userbus"
	"github.com/timmaaaz/ichor/business/domain/core/userrolebus"
	"github.com/timmaaaz/ichor/business/domain/geography/citybus"
	"github.com/timmaaaz/ichor/business/domain/geography/streetbus"
	"github.com/timmaaaz/ichor/business/domain/geography/timezonebus"
	"github.com/timmaaaz/ichor/business/domain/hr/approvalbus"
	"github.com/timmaaaz/ichor/business/domain/hr/commentbus"
	"github.com/timmaaaz/ichor/business/domain/hr/homebus"
	"github.com/timmaaaz/ichor/business/domain/hr/officebus"
	"github.com/timmaaaz/ichor/business/domain/hr/reportstobus"
	"github.com/timmaaaz/ichor/business/domain/hr/titlebus"
	"github.com/timmaaaz/ichor/business/domain/inventory/cyclecountitembus"
	"github.com/timmaaaz/ichor/business/domain/inventory/cyclecountsessionbus"
	"github.com/timmaaaz/ichor/business/domain/inventory/inspectionbus"
	"github.com/timmaaaz/ichor/business/domain/inventory/inventoryadjustmentbus"
	"github.com/timmaaaz/ichor/business/domain/inventory/inventoryitembus"
	"github.com/timmaaaz/ichor/business/domain/inventory/inventorylocationbus"
	"github.com/timmaaaz/ichor/business/domain/inventory/inventorytransactionbus"
	"github.com/timmaaaz/ichor/business/domain/inventory/lotlocationbus"
	"github.com/timmaaaz/ichor/business/domain/inventory/lottrackingsbus"
	"github.com/timmaaaz/ichor/business/domain/inventory/picktaskbus"
	"github.com/timmaaaz/ichor/business/domain/inventory/putawaytaskbus"
	"github.com/timmaaaz/ichor/business/domain/inventory/serialnumberbus"
	"github.com/timmaaaz/ichor/business/domain/inventory/transferorderbus"
	"github.com/timmaaaz/ichor/business/domain/inventory/warehousebus"
	"github.com/timmaaaz/ichor/business/domain/inventory/zonebus"
	"github.com/timmaaaz/ichor/business/domain/labels/labelbus"
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
	"github.com/timmaaaz/ichor/business/domain/scenarios/scenariobus"
	"github.com/timmaaaz/ichor/business/sdk/workflow"
	"github.com/timmaaaz/ichor/business/sdk/workflow/workflowactions/data"
)

// EntityReg is one domain↔entity registration. Domain/Entity are the delegate domain
// and bare entity name (mirrors RegisterDomain(del, Domain, Entity)). Schema is the DB
// schema the entity's table lives in, used ONLY to build the reverse map key
// (schema.table) for generic-write synthesis; "" means the entity is registered for
// real-bus-write cascades but is not a generic-write synthesis target.
type EntityReg struct {
	Schema string
	Domain string
	Entity string
}

// Registrations is the authoritative list of delegate domains the workflow subscriber
// listens on. It reproduces, 1:1, the RegisterDomain block formerly hand-coded in
// all.go (so the worker registers the identical set). Domain/Entity reference the bus
// consts — a drifted name fails the build, not silently at runtime.
func Registrations() []EntityReg {
	return []EntityReg{
		// Sales domain
		{"sales", ordersbus.DomainName, ordersbus.EntityName},
		{"", ordersbus.BindingDomainName, ordersbus.BindingEntityName},
		{"sales", customersbus.DomainName, customersbus.EntityName},
		{"sales", orderlineitemsbus.DomainName, orderlineitemsbus.EntityName},
		{"sales", orderfulfillmentstatusbus.DomainName, orderfulfillmentstatusbus.EntityName},
		{"sales", lineitemfulfillmentstatusbus.DomainName, lineitemfulfillmentstatusbus.EntityName},

		// Assets domain
		{"assets", assetbus.DomainName, assetbus.EntityName},
		{"assets", validassetbus.DomainName, validassetbus.EntityName},
		{"assets", userassetbus.DomainName, userassetbus.EntityName},
		{"assets", assettypebus.DomainName, assettypebus.EntityName},
		{"assets", assetconditionbus.DomainName, assetconditionbus.EntityName},
		{"assets", assettagbus.DomainName, assettagbus.EntityName},
		{"assets", tagbus.DomainName, tagbus.EntityName},
		{"assets", approvalstatusbus.DomainName, approvalstatusbus.EntityName},
		{"assets", fulfillmentstatusbus.DomainName, fulfillmentstatusbus.EntityName},

		// Core domain
		{"core", userbus.DomainName, userbus.EntityName},
		{"core", rolebus.DomainName, rolebus.EntityName},
		{"core", userrolebus.DomainName, userrolebus.EntityName},
		{"core", tableaccessbus.DomainName, tableaccessbus.EntityName},
		{"core", pagebus.DomainName, pagebus.EntityName},
		{"core", paymenttermbus.DomainName, paymenttermbus.EntityName},
		{"core", currencybus.DomainName, currencybus.EntityName},
		{"core", rolepagebus.DomainName, rolepagebus.EntityName},
		{"core", contactinfosbus.DomainName, contactinfosbus.EntityName},

		// HR domain
		{"hr", approvalbus.DomainName, approvalbus.EntityName},
		{"hr", commentbus.DomainName, commentbus.EntityName},
		{"hr", homebus.DomainName, homebus.EntityName},
		{"hr", officebus.DomainName, officebus.EntityName},
		{"hr", reportstobus.DomainName, reportstobus.EntityName},
		{"hr", titlebus.DomainName, titlebus.EntityName},

		// Geography domain (countrybus/regionbus read-only, no events)
		{"geography", citybus.DomainName, citybus.EntityName},
		{"geography", streetbus.DomainName, streetbus.EntityName},
		{"geography", timezonebus.DomainName, timezonebus.EntityName},

		// Products domain
		{"products", productbus.DomainName, productbus.EntityName},
		{"products", productcategorybus.DomainName, productcategorybus.EntityName},
		{"products", brandbus.DomainName, brandbus.EntityName},
		{"products", productcostbus.DomainName, productcostbus.EntityName},
		{"products", costhistorybus.DomainName, costhistorybus.EntityName},
		{"products", physicalattributebus.DomainName, physicalattributebus.EntityName},
		{"products", metricsbus.DomainName, metricsbus.EntityName},

		// Procurement domain
		{"procurement", supplierbus.DomainName, supplierbus.EntityName},
		{"procurement", supplierproductbus.DomainName, supplierproductbus.EntityName},
		{"procurement", purchaseorderbus.DomainName, purchaseorderbus.EntityName},
		{"procurement", purchaseorderlineitembus.DomainName, purchaseorderlineitembus.EntityName},
		{"procurement", purchaseorderstatusbus.DomainName, purchaseorderstatusbus.EntityName},
		{"procurement", purchaseorderlineitemstatusbus.DomainName, purchaseorderlineitemstatusbus.EntityName},

		// Inventory domain
		{"inventory", warehousebus.DomainName, warehousebus.EntityName},
		{"inventory", zonebus.DomainName, zonebus.EntityName},
		{"inventory", inventorylocationbus.DomainName, inventorylocationbus.EntityName},
		{"inventory", inventoryitembus.DomainName, inventoryitembus.EntityName},
		{"inventory", inventorytransactionbus.DomainName, inventorytransactionbus.EntityName},
		{"inventory", inventoryadjustmentbus.DomainName, inventoryadjustmentbus.EntityName},
		{"inventory", putawaytaskbus.DomainName, putawaytaskbus.EntityName},
		{"inventory", picktaskbus.DomainName, picktaskbus.EntityName},
		{"inventory", cyclecountsessionbus.DomainName, cyclecountsessionbus.EntityName},
		{"inventory", cyclecountitembus.DomainName, cyclecountitembus.EntityName},
		{"inventory", transferorderbus.DomainName, transferorderbus.EntityName},
		{"inventory", inspectionbus.DomainName, inspectionbus.EntityName},
		{"inventory", lottrackingsbus.DomainName, lottrackingsbus.EntityName},
		{"inventory", serialnumberbus.DomainName, serialnumberbus.EntityName},
		{"inventory", lotlocationbus.DomainName, lotlocationbus.EntityName},

		// Config domain
		{"config", formbus.DomainName, formbus.EntityName},
		{"config", formfieldbus.DomainName, formfieldbus.EntityName},
		{"config", pageconfigbus.DomainName, pageconfigbus.EntityName},
		{"config", pagecontentbus.DomainName, pagecontentbus.EntityName},
		{"config", pageactionbus.DomainName, pageactionbus.EntityName},

		// Labels domain
		{"labels", labelbus.DomainName, labelbus.EntityName},

		// Scenarios domain
		{"scenarios", scenariobus.DomainName, scenariobus.EntityName},

		// Workflow domain — allocation_results (P4 M2). Fired by workflowbus, not a
		// domain bus; registered here so the subscriber dispatches its on_create event.
		// Not a generic-write synthesis target (Schema "" → absent from the reverse map).
		{"", workflow.AllocationResultDomainName, "allocation_results"},
	}
}

// ReverseMap builds the schema-qualified-table → EntityRef lookup the generic write
// handlers use to resolve which delegate domain + bare entity to fire under (P4 §E.2).
// It includes only entries whose schema.entity is an actual generically-writable table
// (data.IsValidTableName), so an entity whose bus EntityName does not equal its table
// (e.g. metricsbus "metrics" vs table "quality_metrics") or a non-writable target
// simply never enters the map — the handler then degrades safely (no event).
func ReverseMap() map[string]data.EntityRef {
	m := make(map[string]data.EntityRef)
	for _, r := range Registrations() {
		if r.Schema == "" {
			continue
		}
		qualified := r.Schema + "." + r.Entity
		if !data.IsValidTableName(qualified) {
			continue
		}
		m[qualified] = data.EntityRef{Domain: r.Domain, Entity: r.Entity}
	}
	return m
}
