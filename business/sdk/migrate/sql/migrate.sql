-- UPDATED SCHEMA - Primary keys changed to 'id', foreign keys remain descriptive

-- Version: 0.01
-- Core/System schema for authentication and base configuration
CREATE SCHEMA IF NOT EXISTS  core;
-- users, roles, user_roles, table_access, contact_infos

-- Version: 0.02
-- Human Resources schema
CREATE SCHEMA IF NOT EXISTS  hr;
-- titles, offices, reports_to, user_approval_status, 
-- user_approval_comments, homes

-- Version: 0.03
-- Location/Geography schema (shared reference data)
CREATE SCHEMA IF NOT EXISTS  geography;
-- countries, regions, cities, streets

-- Version: 0.04
-- Asset Management schema
CREATE SCHEMA IF NOT EXISTS  assets;
-- asset_types, asset_conditions, valid_assets, assets, 
-- user_assets, asset_tags, tags, approval_status, fulfillment_status

-- Version: 0.05
-- Inventory/Warehouse Management schema
CREATE SCHEMA IF NOT EXISTS  inventory;
-- warehouses, zones, inventory_locations, inventory_items, 
-- inventory_transactions, inventory_adjustments, transfer_orders,
-- serial_numbers, lot_trackings, quality_inspections

-- Version: 0.06
-- Product Information Management schema
CREATE SCHEMA IF NOT EXISTS  products;
-- products, product_categories, product_costs, physical_attributes,
-- brands, quality_metrics, cost_history

-- Version: 0.07
-- Supply Chain/Procurement schema
CREATE SCHEMA IF NOT EXISTS  procurement;
-- suppliers, supplier_products

-- Version: 0.08
-- Sales/Order Management schema
CREATE SCHEMA IF NOT EXISTS  sales;
-- customers, orders, order_line_items, order_fulfillment_statuses,
-- line_item_fulfillment_statuses

-- Version: 0.09
-- Workflow/Automation schema
CREATE SCHEMA IF NOT EXISTS  workflow;
-- automation_rules, automation_executions, action_templates, 
-- rule_actions, rule_dependencies, trigger_types, entity_types,
-- entities, notification_deliveries, allocation_results

-- Version: 0.10
-- Configuration schema
CREATE SCHEMA IF NOT EXISTS  config;
-- table_configs

-- Version: 1.01
-- Description: Create table asset_types
CREATE TABLE assets.asset_types (
   id UUID NOT NULL,
   name TEXT NOT NULL,
   description TEXT NULL,
   PRIMARY KEY (id),
   UNIQUE (name)
);

-- Version: 1.02
-- Description: Create table asset_conditions
CREATE TABLE assets.asset_conditions (
   id UUID NOT NULL,
   name TEXT NOT NULL,
   description TEXT NULL,
   PRIMARY KEY (id),
   UNIQUE (name)
);

-- Version: 1.03
-- Description: Create table countries
CREATE TABLE geography.countries (
   id UUID NOT NULL,
   number INT NOT NULL,
   name TEXT NOT NULL,
   alpha_2 VARCHAR(2) NOT NULL,
   alpha_3 VARCHAR(3) NOT NULL,
   PRIMARY KEY (id)
);

-- Version: 1.04
-- Description: Create table regions
CREATE TABLE geography.regions (
   id UUID NOT NULL,
   country_id UUID NOT NULL,
   name TEXT NOT NULL,
   code TEXT NOT NULL,
   -- TODO: determine if these should be enforced unique.
   PRIMARY KEY (id),
   FOREIGN KEY (country_id) REFERENCES geography.countries(id) ON DELETE CASCADE
);

-- Version: 1.05
-- Description: create table cities
CREATE TABLE geography.cities (
   id UUID NOT NULL,
   region_id UUID NOT NULL,
   name TEXT NOT NULL,
   PRIMARY KEY (id),
   UNIQUE (region_id, name),
   FOREIGN KEY (region_id) REFERENCES geography.regions(id) ON DELETE CASCADE
);

-- Version: 1.06
-- Description: create table streets
CREATE TABLE geography.streets (
   id UUID NOT NULL,
   city_id UUID NOT NULL,
   line_1 TEXT NOT NULL,
   line_2 TEXT NULL,
   postal_code VARCHAR(20) NULL,
   PRIMARY KEY (id),
   FOREIGN KEY (city_id) REFERENCES geography.cities(id) ON DELETE SET NULL-- Check this cascade relationship
);

-- Version: 1.07
-- Description: Create table user_approval_status
CREATE TABLE hr.user_approval_status (
   id UUID NOT NULL, 
   icon_id UUID NULL, 
   name TEXT NOT NULL,
   PRIMARY KEY (id)
);

-- Version: 1.08
-- Description: Create table titles
CREATE TABLE hr.titles (
   id UUID NOT NULL, 
   name TEXT NOT NULL,
   description TEXT NULL,
   PRIMARY KEY (id)
);

-- Version: 1.09
-- Description: Create table offices
CREATE TABLE hr.offices (
   id UUID NOT NULL, 
   name TEXT NOT NULL,
   street_id UUID NOT NULL,
   PRIMARY KEY (id),
   FOREIGN KEY (street_id) REFERENCES geography.streets(id) ON DELETE CASCADE
);

-- Version: 1.10
-- Description: Create table users
CREATE TABLE core.users (
   id UUID NOT NULL,
   requested_by UUID NULL,
   approved_by UUID NULL,
   user_approval_status_id UUID NOT NULL,
   title_id UUID NULL,
   office_id UUID NULL,
   work_phone_id UUID NULL,
   cell_phone_id UUID NULL,
   username TEXT UNIQUE NOT NULL,
   first_name TEXT NOT NULL,
   last_name TEXT NOT NULL,
   email TEXT UNIQUE NOT NULL,
   birthday DATE NULL,
   roles TEXT [] NOT NULL,
   system_roles TEXT [] NOT NULL,
   password_hash TEXT NOT NULL,
   enabled BOOLEAN NOT NULL,
   date_hired TIMESTAMP NULL,
   date_requested TIMESTAMP NULL,
   date_approved TIMESTAMP NULL,
   created_date TIMESTAMP NOT NULL,
   updated_date TIMESTAMP NOT NULL,
   PRIMARY KEY (id),
   FOREIGN KEY (requested_by) REFERENCES core.users(id) ON DELETE SET NULL, -- we don't want to delete someone if their boss is deleted
   FOREIGN KEY (approved_by) REFERENCES core.users(id) ON DELETE SET NULL, -- we don't want to delete someone if their boss is deleted
   FOREIGN KEY (title_id) REFERENCES hr.titles(id) ON DELETE CASCADE,
   FOREIGN KEY (office_id) REFERENCES hr.offices(id) ON DELETE CASCADE,
   FOREIGN KEY (user_approval_status_id) REFERENCES hr.user_approval_status(id) ON DELETE CASCADE
);

-- Version: 1.11
-- Description: Create table valid_assets
CREATE TABLE assets.valid_assets (
   id UUID NOT NULL,
   type_id UUID NOT NULL,
   name TEXT NOT NULL,
   est_price NUMERIC(10,2) NULL,
   price NUMERIC(10,2) NULL,
   maintenance_interval INTERVAL NULL,
   life_expectancy INTERVAL NULL,
   serial_number TEXT NULL,
   model_number TEXT NULL,
   is_enabled BOOLEAN NOT NULL,
   created_date TIMESTAMP NOT NULL,
   updated_date TIMESTAMP NOT NULL,
   created_by UUID NOT NULL,
   updated_by UUID NOT NULL,
   PRIMARY KEY (id),
   
   -- UNIQUE named constraint
   CONSTRAINT unique_asset_name UNIQUE (name),

   -- named foreign keys
   CONSTRAINT fk_assets_type_id FOREIGN KEY (type_id) REFERENCES assets.asset_types(id) ON DELETE CASCADE,
   CONSTRAINT fk_assets_created_by FOREIGN KEY (created_by) REFERENCES core.users(id) ON DELETE CASCADE,
   CONSTRAINT fk_assets_updated_by FOREIGN KEY (updated_by) REFERENCES core.users(id) ON DELETE CASCADE
);

-- Version: 1.12
-- Description: Create table homes
CREATE TABLE hr.homes (
   id UUID NOT NULL,
   TYPE TEXT NOT NULL,
   user_id UUID NOT NULL,
   address_1 TEXT NOT NULL,
   address_2 TEXT NULL,
   zip_code TEXT NOT NULL,
   city TEXT NOT NULL,
   state TEXT NOT NULL,
   country TEXT NOT NULL,
   created_date TIMESTAMP NOT NULL,
   updated_date TIMESTAMP NOT NULL,
   PRIMARY KEY (id),
   FOREIGN KEY (user_id) REFERENCES core.users(id) ON DELETE CASCADE
);

-- Version: 1.13
-- Description: Add approval status 
CREATE TABLE assets.approval_status (
   id UUID NOT NULL, 
   icon_id UUID NOT NULL, 
   name TEXT NOT NULL,
   PRIMARY KEY (id)
);

-- Version: 1.14
-- Description: Add fulfillment status
CREATE TABLE assets.fulfillment_status (
   id UUID NOT NULL, 
   icon_id UUID NOT NULL, 
   name TEXT NOT NULL,
   PRIMARY KEY (id)
);

-- Version: 1.15
-- Description: Add Tags
CREATE TABLE assets.tags (
   id UUID NOT NULL, 
   name TEXT NOT NULL,
   description TEXT NULL,
   PRIMARY KEY (id)
);

-- Version: 1.16
-- Description: Add asset_tags
CREATE TABLE assets.asset_tags (
   id UUID NOT NULL,
   valid_asset_id UUID NOT NULL,
   tag_id UUID NOT NULL,
   PRIMARY KEY (id),
   FOREIGN KEY (valid_asset_id) REFERENCES assets.valid_assets(id) ON DELETE CASCADE,
   FOREIGN KEY (tag_id) REFERENCES assets.tags(id) ON DELETE CASCADE
);

-- Version: 1.17
-- Description: Creates reports to table
CREATE TABLE hr.reports_to (
   id UUID NOT NULL,
   reporter_id UUID NOT NULL,
   boss_id UUID NOT NULL,
   PRIMARY KEY (id),
   FOREIGN KEY (reporter_id) REFERENCES core.users(id) ON DELETE CASCADE,
   FOREIGN KEY (boss_id) REFERENCES core.users(id) ON DELETE CASCADE
);

-- Version: 1.18
-- Description: Add assets
CREATE TABLE assets.assets (
   id UUID NOT NULL,
   valid_asset_id UUID NOT NULL,
   last_maintenance_time TIMESTAMP NOT NULL,
   serial_number TEXT NOT NULL,
   asset_condition_id UUID NOT NULL,
   PRIMARY KEY (id),
   FOREIGN KEY (valid_asset_id) REFERENCES assets.valid_assets(id) ON DELETE CASCADE,
   FOREIGN KEY (asset_condition_id) REFERENCES assets.asset_conditions(id) ON DELETE CASCADE
);

-- Version: 1.19
-- Description: Add user_assets
CREATE TABLE assets.user_assets (
   id UUID NOT NULL,
   user_id UUID NOT NULL,
   asset_id UUID NOT NULL,
   approval_status_id UUID NOT NULL,
   last_maintenance TIMESTAMP NOT NULL,
   date_received TIMESTAMP NOT NULL,
   approved_by UUID NOT NULL,
   fulfillment_status_id UUID NOT NULL,
   PRIMARY KEY (id),
   FOREIGN KEY (user_id) REFERENCES core.users(id) ON DELETE CASCADE,
   FOREIGN KEY (approved_by) REFERENCES core.users(id) ON DELETE CASCADE,
   FOREIGN KEY (asset_id) REFERENCES assets.assets(id) ON DELETE CASCADE,
   FOREIGN KEY (approval_status_id) REFERENCES assets.approval_status(id) ON DELETE CASCADE,
   FOREIGN KEY (fulfillment_status_id) REFERENCES assets.fulfillment_status(id) ON DELETE CASCADE
);

-- Version: 1.20
-- Description: Add user_approval_comments
CREATE TABLE hr.user_approval_comments (
   id UUID NOT NULL,
   comment VARCHAR(255) NOT NULL,
   commenter_id UUID NOT NULL,
   user_id UUID NOT NULL,
   created_date TIMESTAMP NOT NULL,
   PRIMARY KEY (id),
   FOREIGN KEY (commenter_id) REFERENCES core.users(id) ON DELETE SET NULL,
   FOREIGN KEY (user_id) REFERENCES core.users(id) ON DELETE SET NULL
);

CREATE TYPE contact_type as ENUM ('phone', 'email', 'mail', 'fax');

-- Version: 1.21
-- Description: Add contact_infos
CREATE TABLE core.contact_infos (
   id UUID NOT NULL,
   first_name VARCHAR(50) NOT NULL,
   last_name VARCHAR(50) NOT NULL,
   primary_phone_number VARCHAR(50) NOT NULL,
   secondary_phone_number VARCHAR(50) NULL,
   email_address VARCHAR(50) NOT NULL,
   street_id UUID NOT NULL,
   delivery_address_id UUID NULL,
   available_hours_start VARCHAR(50) NOT NULL,
   available_hours_end VARCHAR(50) NOT NULL,
   timezone VARCHAR(50) NOT NULL,
   preferred_contact_type contact_type NOT NULL,
   notes TEXT NULL,
   PRIMARY KEY (id)
);

-- Version: 1.22
-- Description: add brands
CREATE TABLE products.brands (
   id UUID NOT NULL,
   name TEXT NOT NULL,
   contact_infos_id UUID NOT NULL,
   created_date TIMESTAMP NOT NULL,
   updated_date TIMESTAMP NOT NULL,
   PRIMARY KEY (id),
   FOREIGN KEY (contact_infos_id) REFERENCES core.contact_infos(id)
);

-- Version: 1.23
-- Description: add product_categoriesp
CREATE TABLE products.product_categories (
   id UUID NOT NULL,
   name TEXT NOT NULL,
   description text NOT NULL,
   created_date TIMESTAMP NOT NULL,
   updated_date TIMESTAMP NOT NULL,
   PRIMARY KEY (id)
);

-- Version: 1.24
-- Description: Create table warehouses
CREATE TABLE inventory.warehouses (
   id UUID NOT NULL,
   name TEXT NOT NULL,
   street_id UUID NOT NULL,
   is_active BOOLEAN NOT NULL,
   created_date TIMESTAMP NOT NULL,
   updated_date TIMESTAMP NOT NULL,
   created_by UUID NOT NULL,
   updated_by UUID NOT NULL,
   PRIMARY KEY (id),
   FOREIGN KEY (street_id) REFERENCES geography.streets(id) ON DELETE CASCADE
);

-- =============================================================================
-- Core Permissions
-- =============================================================================

-- Version: 1.25
-- Description: Create table roles
CREATE TABLE core.roles (
    id UUID PRIMARY KEY,
    name VARCHAR(50) NOT NULL UNIQUE,
    description TEXT
);

-- Version: 1.26
-- Description: Create table user_roles
CREATE TABLE core.user_roles (
      id UUID NOT NULL,
      user_id UUID NOT NULL,
      role_id UUID NOT NULL,
      PRIMARY KEY (id),
      UNIQUE (user_id, role_id),
      FOREIGN KEY (user_id) REFERENCES core.users(id) ON DELETE CASCADE,
      FOREIGN KEY (role_id) REFERENCES core.roles(id) ON DELETE CASCADE
);

-- Version: 1.27
-- Description: Create table table_access
CREATE TABLE core.table_access (
    id UUID PRIMARY KEY,
    role_id UUID REFERENCES core.roles(id) ON DELETE CASCADE,
    table_name VARCHAR(50) NOT NULL,
    can_create BOOLEAN DEFAULT FALSE,
    can_read BOOLEAN DEFAULT FALSE,
    can_update BOOLEAN DEFAULT FALSE,
    can_delete BOOLEAN DEFAULT FALSE,
    UNIQUE(role_id, table_name)
);

-- Version: 1.28
-- Description: add products
CREATE TABLE products.products (
   id UUID NOT NULL,
   sku VARCHAR(50) NOT NULL,
   brand_id UUID NOT NULL,
   category_id UUID NOT NULL,
   name VARCHAR(255) NOT NULL,
   description TEXT NOT NULL,
   model_number VARCHAR(100),
   upc_code VARCHAR(50) NOT NULL,
   status VARCHAR(20) NOT NULL,
   is_active BOOLEAN NOT NULL,
   is_perishable BOOLEAN NOT NULL,
   handling_instructions TEXT NULL,
   units_per_case INT NOT NULL,
   created_date TIMESTAMP NOT NULL,
   updated_date TIMESTAMP NOT NULL,
   PRIMARY KEY (id),
   FOREIGN KEY (brand_id) REFERENCES products.brands(id),
   FOREIGN KEY (category_id) REFERENCES products.product_categories(id)
);

-- Version: 1.29
-- Description: add physical_attributes
CREATE TABLE products.physical_attributes (
   id UUID NOT NULL,
   product_id UUID NOT NULL,
   length NUMERIC(10, 4) NOT NULL,
   width NUMERIC(10, 4) NOT NULL,
   height NUMERIC(10, 4) NOT NULL,
   weight NUMERIC(10, 4) NOT NULL,
   weight_unit VARCHAR(10) NOT NULL,
   color VARCHAR(50) NULL,
   size VARCHAR(50) NULL,
   material VARCHAR(100) NULL,
   storage_requirements text NOT NULL,
   hazmat_class VARCHAR(50) NOT NULL,
   shelf_life_days INTEGER NOT NULL,
   created_date TIMESTAMP NOT NULL,
   updated_date TIMESTAMP NOT NULL,
   PRIMARY KEY (id),
   FOREIGN KEY (product_id) REFERENCES products.products(id)
);

-- Version: 1.30
-- Description: add product_costs
CREATE TABLE products.product_costs (
   id UUID NOT NULL,
   product_id UUID NOT NULL,
   purchase_cost NUMERIC(10,2) NOT NULL,
   selling_price NUMERIC(10,2) NOT NULL,
   currency VARCHAR(50) NOT NULL,
   msrp NUMERIC(10,2) NOT NULL,
   markup_percentage NUMERIC(10,4) NOT NULL,
   landed_cost NUMERIC(10,2) NOT NULL,
   carrying_cost NUMERIC(10,2) NOT NULL,
   abc_classification char(1) NOT NULL,
   depreciation_value NUMERIC(10,4) NOT NULL,
   insurance_value NUMERIC(10,2) NOT NULL,
   effective_date TIMESTAMP NOT NULL,
   created_date TIMESTAMP NOT NULL,
   updated_date TIMESTAMP NOT NULL,
   PRIMARY KEY (id),
   FOREIGN KEY (product_id) REFERENCES products.products(id)
);

-- Version: 1.31
-- Description: add suppliers
CREATE TABLE procurement.suppliers (
   id UUID NOT NULL,
   contact_infos_id UUID NOT NULL,
   name VARCHAR(100) NOT NULL,
   payment_terms TEXT NOT NULL,
   lead_time_days INTEGER NOT NULL,
   rating NUMERIC(10, 2) NOT NULL,
   is_active BOOLEAN NOT NULL,
   created_date TIMESTAMP NOT NULL,
   updated_date TIMESTAMP NOT NULL,
   PRIMARY KEY (id),
   FOREIGN KEY (contact_infos_id) REFERENCES core.contact_infos(id)
);

-- Version: 1.32
-- Description: add cost_history
CREATE TABLE products.cost_history (
   id UUID NOT NULL,
   product_id UUID NOT NULL,
   cost_type VARCHAR(50) NOT NULL,
   amount NUMERIC(10,2) NOT NULL,
   currency VARCHAR(50) NOT NULL,
   effective_date TIMESTAMP NOT NULL,
   end_date TIMESTAMP NOT NULL,
   created_date TIMESTAMP NOT NULL,
   updated_date TIMESTAMP NOT NULL,
   PRIMARY KEY (id),
   FOREIGN KEY (product_id) REFERENCES products.products(id)
);

-- Version: 1.33
-- Description: add supplier_products
CREATE TABLE procurement.supplier_products (
   id UUID NOT NULL,
   supplier_id UUID NOT NULL,
   product_id UUID NOT NULL,
   supplier_part_number varchar(100) NOT NULL,
   min_order_quantity INT NOT NULL,
   max_order_quantity INT NOT NULL,
   lead_time_days INT NOT NULL,
   unit_cost NUMERIC(10,2) NOT NULL,
   is_primary_supplier BOOLEAN NOT NULL,
   created_date TIMESTAMP NOT NULL,
   updated_date TIMESTAMP NOT NULL,
   PRIMARY KEY (id),
   FOREIGN KEY (supplier_id) REFERENCES procurement.suppliers(id),
   FOREIGN KEY (product_id) REFERENCES products.products(id)
);

-- Version: 1.34
-- Description: add quality_metrics
CREATE TABLE products.quality_metrics (
   id UUID NOT NULL,
   product_id UUID NOT NULL,
   return_rate NUMERIC(10, 4) NOT NULL,
   defect_rate NUMERIC(10, 4) NOT NULL,
   measurement_period INTERVAL NOT NULL,
   created_date TIMESTAMP NOT NULL,
   updated_date TIMESTAMP NOT NULL,
   PRIMARY KEY (id),
   FOREIGN KEY (product_id) REFERENCES products.products(id)
);

-- Version: 1.35
-- Description: add lot tracking
CREATE TABLE inventory.lot_trackings (
   id UUID NOT NULL,
   supplier_product_id UUID NOT NULL,
   lot_number VARCHAR(100) NOT NULL,
   manufacture_date TIMESTAMP NOT NULL,
   expiration_date TIMESTAMP NOT NULL,
   received_date TIMESTAMP NOT NULL,
   quantity INT NOT NULL,
   quality_status varchar(20) NOT NULL,
   created_date TIMESTAMP NOT NULL,
   updated_date TIMESTAMP NOT NULL,
   PRIMARY KEY (id),
   FOREIGN KEY (supplier_product_id) REFERENCES procurement.supplier_products(id)
);

-- Version: 1.36
-- Description: add zones
CREATE TABLE inventory.zones (
   id UUID NOT NULL,
   warehouse_id UUID NOT NULL,
   name VARCHAR(50) NOT NULL,
   description TEXT NULL,
   created_date TIMESTAMP NOT NULL,
   updated_date TIMESTAMP NOT NULL,
   PRIMARY KEY (id),
   FOREIGN KEY (warehouse_id) REFERENCES inventory.warehouses(id)
);

-- Version: 1.37
-- Description: add inventory_locations
CREATE TABLE inventory.inventory_locations (
   id UUID NOT NULL,
   zone_id UUID NOT NULL,
   warehouse_id UUID NOT NULL,
   aisle varchar(20) NOT NULL,
   rack varchar(20) NOT NULL,
   shelf varchar(20) NOT NULL,
   bin varchar(20) NOT NULL,
   is_pick_location boolean NOT NULL,
   is_reserve_location boolean NOT NULL,
   max_capacity integer NOT NULL,
   current_utilization numeric(10,4) NOT NULL,
   created_date TIMESTAMP NOT NULL,
   updated_date TIMESTAMP NOT NULL,
   PRIMARY KEY (id),
   FOREIGN KEY (zone_id) REFERENCES inventory.zones(id),
   FOREIGN KEY (warehouse_id) REFERENCES inventory.warehouses(id)
);

-- Version: 1.38
-- Description: add inventory_items
CREATE TABLE inventory.inventory_items (
   id UUID NOT NULL,
   product_id UUID NOT NULL, 
   location_id UUID NOT NULL,
   quantity INT NOT NULL,
   reserved_quantity INT NOT NULL,
   allocated_quantity INT NOT NULL,
   minimum_stock INT NOT NULL,
   maximum_stock INT NOT NULL,
   reorder_point INT NOT NULL,
   economic_order_quantity INT NOT NULL,
   safety_stock INT NOT NULL,
   avg_daily_usage INT NOT NULL,
   created_date TIMESTAMP NOT NULL,
   updated_date TIMESTAMP NOT NULL,
   PRIMARY KEY (id),
   FOREIGN KEY (product_id) REFERENCES products.products(id),
   FOREIGN KEY (location_id) REFERENCES inventory.inventory_locations(id)
);

-- Version: 1.39
-- Description: add serial_numbers
CREATE TABLE inventory.serial_numbers (
   id UUID NOT NULL,
   product_id UUID NOT NULL,  
   location_id UUID NOT NULL,
   lot_id UUID NOT NULL,
   serial_number VARCHAR(100) NOT NULL,
   status VARCHAR(20) NOT NULL,
   created_date TIMESTAMP NOT NULL,
   updated_date TIMESTAMP NOT NULL,
   PRIMARY KEY (id),
   FOREIGN KEY (product_id) REFERENCES products.products(id),
   FOREIGN KEY (location_id) REFERENCES inventory.inventory_locations(id),
   FOREIGN KEY (lot_id) REFERENCES inventory.lot_trackings(id)
);

-- Version: 1.40
-- Description: add quality_inspections
CREATE TABLE inventory.quality_inspections (
   id UUID NOT NULL,
   product_id UUID NOT NULL,
   inspector_id UUID NOT NULL,
   lot_id UUID NOT NULL,
   inspection_date TIMESTAMP NOT NULL,
   next_inspection_date TIMESTAMP NOT NULL,
   status VARCHAR(20) NOT NULL,
   notes TEXT NULL,
   created_date TIMESTAMP NOT NULL,
   updated_date TIMESTAMP NOT NULL,
   PRIMARY KEY (id),
   FOREIGN KEY (product_id) REFERENCES products.products(id),
   FOREIGN KEY (inspector_id) REFERENCES core.users(id),
   FOREIGN KEY (lot_id) REFERENCES inventory.lot_trackings(id)
);

-- Version: 1.41 
-- Description: add inventory_transactions
CREATE TABLE inventory.inventory_transactions (
   id UUID NOT NULL,
   product_id UUID NOT NULL,
   location_id UUID NOT NULL,
   user_id UUID NOT NULL,
   transaction_type varchar(50) NOT NULL,
   quantity INT NOT NULL,
   reference_number varchar(100) NOT NULL,
   transaction_date TIMESTAMP NOT NULL,
   created_date TIMESTAMP NOT NULL,
   updated_date TIMESTAMP NOT NULL,
   PRIMARY KEY (id),
   FOREIGN KEY (product_id) REFERENCES products.products(id),
   FOREIGN KEY (location_id) REFERENCES inventory.inventory_locations(id),
   FOREIGN KEY (user_id) REFERENCES core.users(id)
);

-- Version: 1.42
-- Description: add inventory_adjustments
CREATE TABLE inventory.inventory_adjustments (
   id UUID NOT NULL,
   product_id UUID NOT NULL,
   location_id UUID NOT NULL,
   adjusted_by UUID NOT NULL,
   approved_by UUID NOT NULL,
   quantity_change INT NOT NULL,
   reason_code varchar(50) NOT NULL,
   notes TEXT NOT NULL,
   adjustment_date TIMESTAMP NOT NULL,
   created_date TIMESTAMP NOT NULL,
   updated_date TIMESTAMP NOT NULL,
   PRIMARY KEY (id),
   FOREIGN KEY (product_id) REFERENCES products.products(id),
   FOREIGN KEY (location_id) REFERENCES inventory.inventory_locations(id),
   FOREIGN KEY (adjusted_by) REFERENCES core.users(id),
   FOREIGN KEY (approved_by) REFERENCES core.users(id)
);

-- Version: 1.43
-- Description: transfer_orders
CREATE TABLE inventory.transfer_orders (
   id UUID NOT NULL,
   product_id UUID NOT NULL,
   from_location_id UUID NOT NULL,
   to_location_id UUID NOT NULL,
   requested_by UUID NOT NULL,
   approved_by UUID NOT NULL,
   quantity int NOT NULL,
   status varchar(20) NOT NULL,
   transfer_date TIMESTAMP NOT NULL,
   created_date TIMESTAMP NOT NULL,
   updated_date TIMESTAMP NOT NULL,
   PRIMARY KEY (id),
   FOREIGN KEY (product_id) REFERENCES products.products(id),
   FOREIGN KEY (from_location_id) REFERENCES inventory.inventory_locations(id),
   FOREIGN KEY (to_location_id) REFERENCES inventory.inventory_locations(id),
   FOREIGN KEY (requested_by) REFERENCES core.users(id),
   FOREIGN KEY (approved_by) REFERENCES core.users(id)
);


-- =============================================================================
-- ORDERS
-- =============================================================================
CREATE TABLE sales.order_fulfillment_statuses (
   id UUID NOT NULL,
   name VARCHAR(50) NOT NULL,
   description TEXT NULL,
   PRIMARY KEY (id),
   UNIQUE (name)
);

CREATE TABLE sales.line_item_fulfillment_statuses (
   id UUID NOT NULL,
   name VARCHAR(50) NOT NULL,
   description TEXT NULL,
   PRIMARY KEY (id),
   UNIQUE (name)
);


CREATE TABLE sales.customers (
   id UUID NOT NULL,
   name VARCHAR(100) NOT NULL,
   contact_id UUID NOT NULL,
   delivery_address_id UUID NOT NULL,
   notes TEXT NULL,
   created_by UUID NOT NULL,
   updated_by UUID NOT NULL,
   created_date TIMESTAMP NOT NULL,
   updated_date TIMESTAMP NOT NULL,
   PRIMARY KEY (id),
   FOREIGN KEY (contact_id) REFERENCES core.contact_infos(id) ON DELETE CASCADE,
   FOREIGN KEY (delivery_address_id) REFERENCES geography.streets(id) ON DELETE CASCADE,
   FOREIGN KEY (created_by) REFERENCES core.users(id),
   FOREIGN KEY (updated_by) REFERENCES core.users(id)
);

CREATE TABLE sales.orders (
   id UUID NOT NULL,
   number VARCHAR(100) NOT NULL,
   customer_id UUID NOT NULL,
   due_date TIMESTAMP NOT NULL,
   order_fulfillment_status_id UUID NOT NULL,
   created_by UUID NOT NULL,
   updated_by UUID NOT NULL,
   created_date TIMESTAMP NOT NULL,
   updated_date TIMESTAMP NOT NULL,
   PRIMARY KEY (id),
   FOREIGN KEY (customer_id) REFERENCES sales.customers(id),
   FOREIGN KEY (order_fulfillment_status_id) REFERENCES sales.order_fulfillment_statuses(id) ON DELETE SET NULL,
   FOREIGN KEY (created_by) REFERENCES core.users(id),
   FOREIGN KEY (updated_by) REFERENCES core.users(id)
);

CREATE TABLE sales.order_line_items (
   id UUID NOT NULL,
   order_id UUID NOT NULL,
   product_id UUID NOT NULL,
   quantity INT NOT NULL, 
   discount NUMERIC(10,2) NULL, -- TODO: Refactor this to be either percent or flat amount
   line_item_fulfillment_statuses_id UUID NOT NULL,
   created_by UUID NOT NULL,
   created_date TIMESTAMP NOT NULL,
   updated_by UUID NOT NULL,
   updated_date TIMESTAMP NOT NULL,
   PRIMARY KEY (id),
   FOREIGN KEY (order_id) REFERENCES sales.orders(id) ON DELETE CASCADE,
   FOREIGN KEY (product_id) REFERENCES products.products(id) ON DELETE CASCADE,
   FOREIGN KEY (line_item_fulfillment_statuses_id) REFERENCES sales.line_item_fulfillment_statuses(id) ON DELETE SET NULL,
   FOREIGN KEY (created_by) REFERENCES core.users(id),
   FOREIGN KEY (updated_by) REFERENCES core.users(id)
);

-- =============================================================================
-- WORKFLOW TABLES
-- =============================================================================
CREATE TABLE workflow.trigger_types (
   id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
   name VARCHAR(50) NOT NULL UNIQUE,
   description TEXT NULL,
   is_active BOOLEAN NOT NULL DEFAULT TRUE,
   deactivated_by UUID NULL,
   FOREIGN KEY (deactivated_by) REFERENCES core.users(id)
);

CREATE TABLE workflow.entity_types (
   id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
   name VARCHAR(50) NOT NULL UNIQUE,
   description TEXT NULL,
   is_active BOOLEAN NOT NULL DEFAULT TRUE,
   deactivated_by UUID NULL,
   FOREIGN KEY (deactivated_by) REFERENCES core.users(id)
);

-- Define automation rules
CREATE TABLE workflow.automation_rules (
   id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
   name VARCHAR(100) NOT NULL,
   description TEXT,
   entity_id UUID NOT NULL, -- table or view name, maybe others in the future
   entity_type_id UUID NOT NULL REFERENCES workflow.entity_types(id),
   
   -- Trigger conditions
   trigger_type_id UUID NOT NULL REFERENCES workflow.trigger_types(id),

   trigger_conditions JSONB NULL, -- When to trigger
      
   -- Control
   is_active BOOLEAN NOT NULL DEFAULT TRUE,
   
   created_date TIMESTAMP NOT NULL DEFAULT NOW(),
   updated_date TIMESTAMP NOT NULL DEFAULT NOW(),
   created_by UUID NOT NULL REFERENCES core.users(id),
   updated_by UUID NOT NULL REFERENCES core.users(id),

   deactivated_by UUID NULL REFERENCES core.users(id)
);

-- Track rule executions
CREATE TABLE workflow.automation_executions (
   id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
   automation_rules_id UUID NOT NULL REFERENCES workflow.automation_rules(id),
   entity_type VARCHAR(50) NOT NULL,
   trigger_data JSONB,
   actions_executed JSONB,
   status VARCHAR(20) NOT NULL, -- 'success', 'failed', 'partial'
   error_message TEXT,
   execution_time_ms INTEGER,
   executed_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE workflow.action_templates (
   id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
   name VARCHAR(100) NOT NULL UNIQUE,
   description TEXT,
   action_type VARCHAR(50) NOT NULL,
   default_config JSONB NOT NULL,
   created_date TIMESTAMP NOT NULL DEFAULT NOW(),
   created_by UUID NOT NULL REFERENCES core.users(id),
   is_active BOOLEAN NOT NULL DEFAULT TRUE,
   deactivated_by UUID NULL REFERENCES core.users(id)
);

CREATE TABLE workflow.rule_actions (
   id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
   automation_rules_id UUID NOT NULL REFERENCES workflow.automation_rules(id),
   name VARCHAR(100) NOT NULL,
   description TEXT,
   action_config JSONB NOT NULL,
   execution_order INTEGER NOT NULL DEFAULT 1,
   is_active BOOLEAN DEFAULT TRUE,
   template_id UUID NULL REFERENCES workflow.action_templates(id),
   deactivated_by UUID NULL REFERENCES core.users(id)
);

CREATE TABLE workflow.rule_dependencies (
   id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
   parent_rule_id UUID REFERENCES workflow.automation_rules(id),
   child_rule_id UUID REFERENCES workflow.automation_rules(id)
);

-- Create entities table
CREATE TABLE workflow.entities (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL UNIQUE,
    entity_type_id UUID NOT NULL REFERENCES workflow.entity_types(id),
    schema_name VARCHAR(50) DEFAULT 'public',
    is_active BOOLEAN DEFAULT TRUE,
    created_date TIMESTAMP DEFAULT NOW(),
    deactivated_by UUID NULL REFERENCES core.users(id)
);

-- Track notification deliveries from workflow actions
CREATE TABLE workflow.notification_deliveries (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    notification_id UUID NOT NULL, -- References the data structure NotificationPayload in the notification queue, probably this will be referencing logs
    automation_execution_id UUID NULL REFERENCES workflow.automation_executions(id), -- Links to workflow execution
    rule_id UUID NULL REFERENCES workflow.automation_rules(id),
    action_id UUID NULL REFERENCES workflow.rule_actions(id),
    recipient_id UUID NOT NULL, -- User ID or email
    channel VARCHAR(50) NOT NULL, -- email, sms, push, in_app
    status VARCHAR(20) NOT NULL, -- pending, sent, delivered, failed, bounced, retrying
    attempts INTEGER DEFAULT 1,
    sent_at TIMESTAMP,
    delivered_at TIMESTAMP,
    failed_at TIMESTAMP,
    error_message TEXT,
    provider_response JSONB, -- Store provider-specific response data
    created_date TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_date TIMESTAMP NOT NULL DEFAULT NOW()
    
    -- Indexes for common queries
   --  INDEX idx_notification_deliveries_notification_id (notification_id),
   --  INDEX idx_notification_deliveries_workflow_id (workflow_id),
   --  INDEX idx_notification_deliveries_recipient_id (recipient_id),
   --  INDEX idx_notification_deliveries_status (status),
   --  INDEX idx_notification_deliveries_created_at (created_at)
);

-- Add to your migration file
CREATE TABLE workflow.allocation_results (
    id UUID PRIMARY KEY,
    idempotency_key VARCHAR(255) UNIQUE NOT NULL,
    allocation_data JSONB NOT NULL,
    created_date TIMESTAMP NOT NULL
);
CREATE INDEX idx_allocation_idempotency ON workflow.allocation_results(idempotency_key);

-- Migration: Create table_configs table for storing table configurations
-- Version: 2.01
-- Description: Create table for storing dynamic table configurations
CREATE TABLE IF NOT EXISTS config.table_configs (
    id UUID PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    description TEXT,
    config JSONB NOT NULL,
    created_by UUID NOT NULL,
    updated_by UUID NOT NULL,
    created_date TIMESTAMP NOT NULL,
    updated_date TIMESTAMP NOT NULL,
    
    -- Foreign keys
    CONSTRAINT fk_table_configs_created_by FOREIGN KEY (created_by) REFERENCES core.users(id) ON DELETE CASCADE,
    CONSTRAINT fk_table_configs_updated_by FOREIGN KEY (updated_by) REFERENCES core.users(id) ON DELETE CASCADE
);

-- Create indexes for better query performance
CREATE INDEX idx_table_configs_name ON config.table_configs(name);
CREATE INDEX idx_table_configs_created_by ON config.table_configs(created_by);
CREATE INDEX idx_table_configs_updated_date ON config.table_configs(updated_date DESC);

-- GIN index for JSONB queries
CREATE INDEX idx_table_configs_config ON config.table_configs USING GIN (config);

-- Comments
COMMENT ON TABLE config.table_configs IS 'Stores user-defined table configurations for dynamic table generation';
COMMENT ON COLUMN config.table_configs.id IS 'Unique identifier for the configuration';
COMMENT ON COLUMN config.table_configs.name IS 'Unique name for the configuration';
COMMENT ON COLUMN config.table_configs.description IS 'Optional description of what this configuration displays';
COMMENT ON COLUMN config.table_configs.config IS 'JSON configuration matching the TableConfig structure';
COMMENT ON COLUMN config.table_configs.created_by IS 'User who created this configuration';
COMMENT ON COLUMN config.table_configs.updated_by IS 'User who last updated this configuration';


-- Migration: Create page_configs for storing page configurations
-- Version: 2.02
-- Description: Create table for storing page configurations
CREATE TABLE IF NOT EXISTS config.page_configs (
   id UUID PRIMARY KEY,
   name TEXT NOT NULL,
   user_id UUID NOT NULL,
   is_default BOOLEAN NOT NULL DEFAULT FALSE,

   CONSTRAINT fk_page_configs_user FOREIGN KEY (user_id) REFERENCES core.users(id) ON DELETE CASCADE,
   CONSTRAINT unique_page_config_name_user UNIQUE (name, user_id)
);

-- Migration: Create page_tab_configs for storing tab configurations within pages
-- Version: 2.03
-- Description: Create table for storing tab configurations within pages
CREATE TABLE IF NOT EXISTS config.page_tab_configs (
   id UUID PRIMARY KEY,
   page_config_id UUID NOT NULL,
   label TEXT NOT NULL,
   config_id UUID NOT NULL,
   is_default BOOLEAN NOT NULL DEFAULT FALSE,
   tab_order INT NOT NULL,

   CONSTRAINT fk_page_tab_configs_page FOREIGN KEY (page_config_id) REFERENCES config.page_configs(id) ON DELETE CASCADE,
   CONSTRAINT fk_page_tab_configs_config FOREIGN KEY (config_id) REFERENCES config.table_configs(id) ON DELETE CASCADE
);

-- Optional: Create a view for commonly accessed configurations
CREATE OR REPLACE VIEW config.active_table_configs AS
SELECT 
    tc.id,
    tc.name,
    tc.description,
    tc.config,
    tc.created_by,
    tc.updated_by,
    tc.created_date,
    tc.updated_date,
    u1.username as created_by_username,
    u2.username as updated_by_username
FROM config.table_configs tc
JOIN core.users u1 ON tc.created_by = u1.id
JOIN core.users u2 ON tc.updated_by = u2.id
WHERE u1.enabled = true
ORDER BY tc.updated_date DESC;

-- Grant appropriate permissions (adjust based on your needs)
-- GRANT SELECT ON config.table_configs TO authenticated;
-- GRANT INSERT, UPDATE, DELETE ON config.table_configs TO authenticated;
-- GRANT SELECT ON config.active_table_configs TO authenticated;

CREATE OR REPLACE VIEW sales.orders_base AS
SELECT
   o.id AS orders_id,
   o.number AS orders_number,
   o.created_date AS orders_order_date,
   o.due_date AS orders_due_date,
   o.created_date AS orders_created_date,
   o.updated_date AS orders_updated_date,
   o.order_fulfillment_status_id AS orders_fulfillment_status_id,
   o.customer_id AS orders_customer_id,

   c.id AS customers_id,
   c.name AS customers_name,
   c.contact_id AS customers_contact_infos_id,
   c.delivery_address_id AS customers_delivery_address_id,
   c.notes AS customers_notes,
   c.created_date AS customers_created_date,
   c.updated_date AS customers_updated_date,

   ofs.name AS order_fulfillment_statuses_name,
   ofs.description AS order_fulfillment_statuses_description
FROM sales.orders o
   INNER JOIN sales.customers c ON o.customer_id = c.id
   LEFT JOIN sales.order_fulfillment_statuses ofs ON o.order_fulfillment_status_id = ofs.id;

CREATE OR REPLACE VIEW sales.order_line_items_base AS
SELECT
   oli.id AS order_line_item_id,
   oli.order_id AS order_line_item_order_id,
   oli.product_id AS order_line_item_product_id,
   oli.quantity AS order_line_item_quantity,
   oli.discount AS order_line_item_discount, 
   oli.line_item_fulfillment_statuses_id AS line_item_fulfillment_statuses_id,
   oli.created_date AS order_line_item_created_date,
   oli.updated_date AS order_line_item_updated_date,

   p.name AS product_name,
   p.sku AS product_sku,
   p.brand_id AS product_brand_id,
   p.category_id AS product_category_id,
   p.description AS product_description,
   p.model_number AS product_model_number,
   p.upc_code AS product_upc_code,
   p.status AS product_status,
   p.is_active AS product_is_active,
   p.is_perishable AS product_is_perishable,
   p.handling_instructions AS product_handling_instructions,
   p.units_per_case AS product_units_per_case,
   p.created_date AS product_created_date,
   p.updated_date AS product_updated_date,

   b.name AS product_brand_name,
   b.contact_infos_id AS product_brand_contact_infos_id,
   b.created_date AS product_brand_created_date,
   b.updated_date AS product_brand_updated_date,

   c.name AS product_category_name,
   c.created_date AS product_category_created_date,
   c.updated_date AS product_category_updated_date

FROM sales.order_line_items oli
   INNER JOIN products.products p ON oli.product_id = p.id
   LEFT JOIN products.brands b ON p.brand_id = b.id
   LEFT JOIN products.product_categories c ON p.category_id = c.id
   LEFT JOIN sales.line_item_fulfillment_statuses ofs ON oli.id = ofs.id;

CREATE OR REPLACE VIEW workflow.automation_rules_view AS 
SELECT 
    ar.id,
    ar.name,
    ar.description,
    ar.trigger_conditions,
    ar.is_active,
    ar.created_date,
    ar.updated_date,
    ar.created_by,
    ar.updated_by,
    ar.deactivated_by,
    
    -- Trigger type fields
    tt.id as trigger_type_id,
    tt.name as trigger_type_name,
    tt.description as trigger_type_description,
    
    -- Entity type fields (from automation_rules.entity_type_id)
    et.id as entity_type_id,
    et.name as entity_type_name,
    et.description as entity_type_description,
    
    -- Entity fields (from automation_rules.entity_id)
    e.id as entity_id,
    e.name as entity_name,
    e.schema_name as entity_schema_name,
    
    -- User fields for better context
    cu.username as created_by_username,
    uu.username as updated_by_username,
    du.username as deactivated_by_username
    
FROM workflow.automation_rules ar
LEFT JOIN workflow.trigger_types tt ON ar.trigger_type_id = tt.id
LEFT JOIN workflow.entity_types et ON ar.entity_type_id = et.id
LEFT JOIN workflow.entities e ON ar.entity_id = e.id
LEFT JOIN core.users cu ON ar.created_by = cu.id
LEFT JOIN core.users uu ON ar.updated_by = uu.id
LEFT JOIN core.users du ON ar.deactivated_by = du.id
WHERE ar.is_active = true;


CREATE OR REPLACE VIEW workflow.rule_actions_view AS 
   SELECT 
      ra.id,
      ra.automation_rules_id,
      ra.name,
      ra.description,
      ra.action_config,
      ra.execution_order,
      ra.is_active,
      ra.template_id,
      at.name as template_name,
      at.action_type as template_action_type,
      at.default_config as template_default_config
   FROM workflow.rule_actions ra
   LEFT JOIN workflow.action_templates at ON ra.template_id = at.id;
