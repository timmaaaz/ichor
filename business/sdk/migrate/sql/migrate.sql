-- UPDATED SCHEMA - Primary keys changed to 'id', foreign keys remain descriptive

-- Version: 1.01
-- Core/System schema for authentication and base configuration
CREATE SCHEMA IF NOT EXISTS  core;
-- users, roles, user_roles, table_access, contact_infos

-- Version: 1.02
-- Human Resources schema
CREATE SCHEMA IF NOT EXISTS  hr;
-- titles, offices, reports_to, user_approval_status, 
-- user_approval_comments, homes

-- Version: 1.03
-- Location/Geography schema (shared reference data)
CREATE SCHEMA IF NOT EXISTS  geography;
-- countries, regions, cities, streets

-- Version: 1.04
-- Asset Management schema
CREATE SCHEMA IF NOT EXISTS  assets;
-- asset_types, asset_conditions, valid_assets, assets, 
-- user_assets, asset_tags, tags, approval_status, fulfillment_status

-- Version: 1.05
-- Inventory/Warehouse Management schema
CREATE SCHEMA IF NOT EXISTS  inventory;
-- warehouses, zones, inventory_locations, inventory_items, 
-- inventory_transactions, inventory_adjustments, transfer_orders,
-- serial_numbers, lot_trackings, quality_inspections

-- Version: 1.06
-- Product Information Management schema
CREATE SCHEMA IF NOT EXISTS  products;
-- products, product_categories, product_costs, physical_attributes,
-- brands, quality_metrics, cost_history

-- Version: 1.07
-- Supply Chain/Procurement schema
CREATE SCHEMA IF NOT EXISTS  procurement;
-- suppliers, supplier_products

-- Version: 1.08
-- Sales/Order Management schema
CREATE SCHEMA IF NOT EXISTS  sales;
-- customers, orders, order_line_items, order_fulfillment_statuses,
-- line_item_fulfillment_statuses

-- Version: 1.09
-- Workflow/Automation schema
CREATE SCHEMA IF NOT EXISTS  workflow;
-- automation_rules, automation_executions, action_templates, 
-- rule_actions, rule_dependencies, trigger_types, entity_types,
-- entities, notification_deliveries, allocation_results

-- Version: 1.10
-- Configuration schema
CREATE SCHEMA IF NOT EXISTS  config;
-- table_configs

-- Version: 1.11
-- Description: Create table asset_types
CREATE TABLE assets.asset_types (
   id UUID NOT NULL,
   name TEXT NOT NULL,
   description TEXT NULL,
   PRIMARY KEY (id),
   UNIQUE (name)
);

-- Version: 1.12
-- Description: Create table asset_conditions
CREATE TABLE assets.asset_conditions (
   id UUID NOT NULL,
   name TEXT NOT NULL,
   description TEXT NULL,
   PRIMARY KEY (id),
   UNIQUE (name)
);

-- Version: 1.13
-- Description: Create payment_terms lookup table
CREATE TABLE core.payment_terms (
   id UUID PRIMARY KEY,
   name VARCHAR(100) UNIQUE NOT NULL,
   description TEXT
);

-- Version: 1.14
-- Description: Create table countries
CREATE TABLE geography.countries (
   id UUID NOT NULL,
   number INT NOT NULL,
   name TEXT NOT NULL,
   alpha_2 VARCHAR(2) NOT NULL,
   alpha_3 VARCHAR(3) NOT NULL,
   PRIMARY KEY (id)
);

-- Version: 1.15
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

-- Version: 1.16
-- Description: create table cities
CREATE TABLE geography.cities (
   id UUID NOT NULL,
   region_id UUID NOT NULL,
   name TEXT NOT NULL,
   PRIMARY KEY (id),
   UNIQUE (region_id, name),
   FOREIGN KEY (region_id) REFERENCES geography.regions(id) ON DELETE CASCADE
);

-- Version: 1.17
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

-- Version: 1.18
-- Description: Create timezones table
CREATE TABLE geography.timezones (
   id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
   name TEXT UNIQUE NOT NULL,
   display_name TEXT NOT NULL,
   utc_offset TEXT NOT NULL,
   is_active BOOLEAN DEFAULT TRUE
);

-- Version: 1.19
-- Description: Create table user_approval_status
CREATE TABLE hr.user_approval_status (
   id UUID NOT NULL,
   icon_id UUID NULL,
   name TEXT NOT NULL,
   primary_color VARCHAR(50) NULL,
   secondary_color VARCHAR(50) NULL,
   icon VARCHAR(100) NULL,
   PRIMARY KEY (id)
);

-- Version: 1.20
-- Description: Create table titles
CREATE TABLE hr.titles (
   id UUID NOT NULL, 
   name TEXT NOT NULL,
   description TEXT NULL,
   PRIMARY KEY (id)
);

-- Version: 1.21
-- Description: Create table offices
CREATE TABLE hr.offices (
   id UUID NOT NULL, 
   name TEXT NOT NULL,
   street_id UUID NOT NULL,
   PRIMARY KEY (id),
   FOREIGN KEY (street_id) REFERENCES geography.streets(id) ON DELETE CASCADE
);

-- Version: 1.22
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

-- Version: 1.23
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

-- Version: 1.24
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

-- Version: 1.25
-- Description: Add approval status
CREATE TABLE assets.approval_status (
   id UUID NOT NULL,
   icon_id UUID NOT NULL,
   name TEXT NOT NULL,
   primary_color VARCHAR(50) NULL,
   secondary_color VARCHAR(50) NULL,
   icon VARCHAR(100) NULL,
   PRIMARY KEY (id)
);

-- Version: 1.26
-- Description: Add fulfillment status
CREATE TABLE assets.fulfillment_status (
   id UUID NOT NULL,
   icon_id UUID NOT NULL,
   name TEXT NOT NULL,
   primary_color VARCHAR(50) NULL,
   secondary_color VARCHAR(50) NULL,
   icon VARCHAR(100) NULL,
   PRIMARY KEY (id)
);

-- Version: 1.27
-- Description: Add Tags
CREATE TABLE assets.tags (
   id UUID NOT NULL, 
   name TEXT NOT NULL,
   description TEXT NULL,
   PRIMARY KEY (id)
);

-- Version: 1.28
-- Description: Add asset_tags
CREATE TABLE assets.asset_tags (
   id UUID NOT NULL,
   valid_asset_id UUID NOT NULL,
   tag_id UUID NOT NULL,
   PRIMARY KEY (id),
   FOREIGN KEY (valid_asset_id) REFERENCES assets.valid_assets(id) ON DELETE CASCADE,
   FOREIGN KEY (tag_id) REFERENCES assets.tags(id) ON DELETE CASCADE
);

-- Version: 1.29
-- Description: Creates reports to table
CREATE TABLE hr.reports_to (
   id UUID NOT NULL,
   reporter_id UUID NOT NULL,
   boss_id UUID NOT NULL,
   PRIMARY KEY (id),
   FOREIGN KEY (reporter_id) REFERENCES core.users(id) ON DELETE CASCADE,
   FOREIGN KEY (boss_id) REFERENCES core.users(id) ON DELETE CASCADE
);

-- Version: 1.30
-- Description: Add assets
CREATE TABLE assets.assets (
   id UUID NOT NULL,
   valid_asset_id UUID NOT NULL,
   last_maintenance_time TIMESTAMP,
   serial_number TEXT NOT NULL,
   asset_condition_id UUID NOT NULL,
   PRIMARY KEY (id),
   FOREIGN KEY (valid_asset_id) REFERENCES assets.valid_assets(id) ON DELETE CASCADE,
   FOREIGN KEY (asset_condition_id) REFERENCES assets.asset_conditions(id) ON DELETE CASCADE
);

-- Version: 1.31
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

-- Version: 1.32
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

-- ENUM types for dropdown fields (schema-scoped for organization)
CREATE TYPE sales.discount_type AS ENUM ('flat', 'percent');
CREATE TYPE config.content_type AS ENUM ('table', 'form', 'tabs', 'text', 'chart', 'container');
CREATE TYPE config.action_type AS ENUM ('button', 'dropdown', 'separator');
CREATE TYPE config.button_variant AS ENUM ('default', 'secondary', 'outline', 'ghost', 'destructive');
CREATE TYPE config.button_alignment AS ENUM ('left', 'right');
CREATE TYPE workflow.alert_severity AS ENUM ('low', 'medium', 'high', 'critical');
CREATE TYPE workflow.alert_status AS ENUM ('active', 'acknowledged', 'dismissed', 'resolved');
CREATE TYPE workflow.recipient_type AS ENUM ('user', 'role');

-- Version: 1.33
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
   timezone_id UUID NOT NULL REFERENCES geography.timezones(id),
   preferred_contact_type contact_type NOT NULL,
   notes TEXT NULL,
   PRIMARY KEY (id)
);

-- Version: 1.34
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

-- Version: 1.35
-- Description: add product_categoriesp
CREATE TABLE products.product_categories (
   id UUID NOT NULL,
   name TEXT NOT NULL,
   description text NOT NULL,
   created_date TIMESTAMP NOT NULL,
   updated_date TIMESTAMP NOT NULL,
   PRIMARY KEY (id)
);

-- Version: 1.36
-- Description: Create table warehouses
CREATE TABLE inventory.warehouses (
   id UUID NOT NULL,
   code TEXT NOT NULL,
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

-- Version: 1.37
-- Description: Create table roles
CREATE TABLE core.roles (
    id UUID PRIMARY KEY,
    name VARCHAR(50) NOT NULL UNIQUE,
    description TEXT
);

-- Version: 1.38
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

-- Version: 1.39
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

-- Version: 1.40
-- Description: Create table pages
CREATE TABLE core.pages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    path TEXT UNIQUE NOT NULL,
    name TEXT NOT NULL,
    module TEXT NOT NULL,
    icon TEXT,
    sort_order INTEGER DEFAULT 1000,
    is_active BOOLEAN DEFAULT TRUE,
    show_in_menu BOOLEAN DEFAULT TRUE
);

-- Version: 1.41
-- Description: Create table role_pages
CREATE TABLE core.role_pages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    role_id UUID REFERENCES core.roles(id) ON DELETE CASCADE,
    page_id UUID REFERENCES core.pages(id) ON DELETE CASCADE,
    can_access BOOLEAN DEFAULT TRUE,
    UNIQUE(role_id, page_id)
);

-- Version: 1.42
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
   tracking_type VARCHAR(20) NOT NULL DEFAULT 'none' CHECK (tracking_type IN ('none', 'lot', 'serial')),
   created_date TIMESTAMP NOT NULL,
   updated_date TIMESTAMP NOT NULL,
   PRIMARY KEY (id),
   FOREIGN KEY (brand_id) REFERENCES products.brands(id),
   FOREIGN KEY (category_id) REFERENCES products.product_categories(id)
);

-- Version: 1.43
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

-- Version: 1.44
-- Description: Create currencies reference table
CREATE TABLE core.currencies (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code VARCHAR(3) NOT NULL UNIQUE,
    name VARCHAR(100) NOT NULL,
    symbol VARCHAR(10) NOT NULL,
    locale VARCHAR(10) NOT NULL,
    decimal_places INT NOT NULL DEFAULT 2,
    is_active BOOLEAN NOT NULL DEFAULT true,
    sort_order INT NOT NULL DEFAULT 0,
    created_by UUID REFERENCES core.users(id),
    created_date TIMESTAMPTZ DEFAULT NOW(),
    updated_by UUID REFERENCES core.users(id),
    updated_date TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_currencies_code ON core.currencies(code);
CREATE INDEX idx_currencies_is_active ON core.currencies(is_active);
CREATE INDEX idx_currencies_sort_order ON core.currencies(sort_order);

COMMENT ON TABLE core.currencies IS 'Reference table for supported currencies';
COMMENT ON COLUMN core.currencies.code IS 'ISO 4217 currency code (e.g., USD, EUR)';
COMMENT ON COLUMN core.currencies.symbol IS 'Currency symbol for display (e.g., $, €)';
COMMENT ON COLUMN core.currencies.locale IS 'Locale identifier for formatting (e.g., en-US)';
COMMENT ON COLUMN core.currencies.decimal_places IS 'Number of decimal places for this currency (e.g., 2 for USD, 0 for JPY)';

-- Version: 1.45
-- Description: add product_costs
CREATE TABLE products.product_costs (
   id UUID NOT NULL,
   product_id UUID NOT NULL,
   purchase_cost NUMERIC(10,2) NOT NULL,
   selling_price NUMERIC(10,2) NOT NULL,
   currency_id UUID NOT NULL,
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
   FOREIGN KEY (product_id) REFERENCES products.products(id),
   FOREIGN KEY (currency_id) REFERENCES core.currencies(id)
);

-- Version: 1.46
-- Description: add suppliers
CREATE TABLE procurement.suppliers (
   id UUID NOT NULL,
   contact_infos_id UUID NOT NULL,
   name VARCHAR(100) NOT NULL,
   payment_term_id UUID NULL,
   lead_time_days INTEGER NOT NULL,
   rating NUMERIC(10, 2) NOT NULL,
   is_active BOOLEAN NOT NULL,
   created_date TIMESTAMP NOT NULL,
   updated_date TIMESTAMP NOT NULL,
   PRIMARY KEY (id),
   FOREIGN KEY (contact_infos_id) REFERENCES core.contact_infos(id),
   FOREIGN KEY (payment_term_id) REFERENCES core.payment_terms(id) ON DELETE SET NULL
);

-- Version: 1.47
-- Description: add cost_history
CREATE TABLE products.cost_history (
   id UUID NOT NULL,
   product_id UUID NOT NULL,
   cost_type VARCHAR(50) NOT NULL,
   amount NUMERIC(10,2) NOT NULL,
   currency_id UUID NOT NULL,
   effective_date TIMESTAMP NOT NULL,
   end_date TIMESTAMP NOT NULL,
   created_date TIMESTAMP NOT NULL,
   updated_date TIMESTAMP NOT NULL,
   PRIMARY KEY (id),
   FOREIGN KEY (product_id) REFERENCES products.products(id),
   FOREIGN KEY (currency_id) REFERENCES core.currencies(id)
);

-- Version: 1.48
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

-- Version: 1.49
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

-- Version: 1.50
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

-- Version: 1.51
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

-- Version: 1.52
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

-- Version: 1.53
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

-- Version: 1.54
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

-- Version: 1.55
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

-- Version: 1.56
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
   lot_id UUID NULL,
   created_date TIMESTAMP NOT NULL,
   updated_date TIMESTAMP NOT NULL,
   PRIMARY KEY (id),
   FOREIGN KEY (product_id) REFERENCES products.products(id),
   FOREIGN KEY (location_id) REFERENCES inventory.inventory_locations(id),
   FOREIGN KEY (user_id) REFERENCES core.users(id),
   FOREIGN KEY (lot_id) REFERENCES inventory.lot_trackings(id)
);

-- Version: 1.57
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

-- Version: 1.58
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

-- Version: 1.59
-- Description: Create table order_fulfillment_statuses
CREATE TABLE sales.order_fulfillment_statuses (
   id UUID NOT NULL,
   name VARCHAR(50) NOT NULL,
   description TEXT NULL,
   primary_color VARCHAR(50) NULL,
   secondary_color VARCHAR(50) NULL,
   icon VARCHAR(100) NULL,
   PRIMARY KEY (id),
   UNIQUE (name)
);

-- Version: 1.60
-- Description: Create table line_item_fulfillment_statuses
CREATE TABLE sales.line_item_fulfillment_statuses (
   id UUID NOT NULL,
   name VARCHAR(50) NOT NULL,
   description TEXT NULL,
   primary_color VARCHAR(50) NULL,
   secondary_color VARCHAR(50) NULL,
   icon VARCHAR(100) NULL,
   PRIMARY KEY (id),
   UNIQUE (name)
);

-- Version: 1.61
-- Description: Create table customers
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

-- Version: 1.62
-- Description: Create table orders
CREATE TABLE sales.orders (
   id UUID NOT NULL,
   number VARCHAR(100) NOT NULL UNIQUE,
   customer_id UUID NOT NULL,
   due_date TIMESTAMP NOT NULL,
   order_fulfillment_status_id UUID NOT NULL,
   -- Address references (same pattern as customers.delivery_address_id)
   billing_address_id UUID NULL,
   shipping_address_id UUID NULL,
   -- Order details
   order_date DATE NULL,
   payment_term_id UUID NULL,
   notes TEXT NULL,
   -- Financial fields
   subtotal DECIMAL(12,2) DEFAULT 0,
   tax_rate DECIMAL(5,2) DEFAULT 0,    -- Whole percentage (e.g., 8.25 for 8.25%)
   tax_amount DECIMAL(12,2) DEFAULT 0,
   shipping_cost DECIMAL(12,2) DEFAULT 0,
   total_amount DECIMAL(12,2) DEFAULT 0,
   currency_id UUID NOT NULL,
   -- Audit columns
   created_by UUID NOT NULL,
   updated_by UUID NOT NULL,
   created_date TIMESTAMP NOT NULL,
   updated_date TIMESTAMP NOT NULL,
   PRIMARY KEY (id),
   FOREIGN KEY (customer_id) REFERENCES sales.customers(id),
   FOREIGN KEY (order_fulfillment_status_id) REFERENCES sales.order_fulfillment_statuses(id) ON DELETE SET NULL,
   FOREIGN KEY (billing_address_id) REFERENCES geography.streets(id) ON DELETE SET NULL,
   FOREIGN KEY (shipping_address_id) REFERENCES geography.streets(id) ON DELETE SET NULL,
   FOREIGN KEY (payment_term_id) REFERENCES core.payment_terms(id) ON DELETE SET NULL,
   FOREIGN KEY (currency_id) REFERENCES core.currencies(id),
   FOREIGN KEY (created_by) REFERENCES core.users(id),
   FOREIGN KEY (updated_by) REFERENCES core.users(id)
);

-- Version: 1.63
-- Description: Create table order_line_items
CREATE TABLE sales.order_line_items (
   id UUID NOT NULL,
   order_id UUID NOT NULL,
   product_id UUID NOT NULL,
   description TEXT NULL,                          -- Line item description
   quantity INT NOT NULL DEFAULT 1,
   unit_price DECIMAL(12,2) DEFAULT 0,             -- Price per unit
   discount NUMERIC(10,2) DEFAULT 0,               -- Discount amount or percent value
   discount_type sales.discount_type DEFAULT 'flat',
   line_total DECIMAL(12,2) DEFAULT 0,             -- Calculated total
   line_item_fulfillment_statuses_id UUID NOT NULL,
   picked_quantity      INTEGER      NOT NULL DEFAULT 0,
   backordered_quantity INTEGER      NOT NULL DEFAULT 0,
   short_pick_reason    VARCHAR(100) NULL,
   -- Audit columns
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

-- Version: 1.64
-- Description: Create table trigger_types
CREATE TABLE workflow.trigger_types (
   id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
   name VARCHAR(50) NOT NULL UNIQUE,
   description TEXT NULL,
   is_active BOOLEAN NOT NULL DEFAULT TRUE,
   deactivated_by UUID NULL,
   FOREIGN KEY (deactivated_by) REFERENCES core.users(id)
);

-- Version: 1.65
-- Description: Create table entity_types
CREATE TABLE workflow.entity_types (
   id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
   name VARCHAR(50) NOT NULL UNIQUE,
   description TEXT NULL,
   is_active BOOLEAN NOT NULL DEFAULT TRUE,
   deactivated_by UUID NULL,
   FOREIGN KEY (deactivated_by) REFERENCES core.users(id)
);

-- Version: 1.66
-- Description: Create table automation_rules
CREATE TABLE workflow.automation_rules (
   id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
   name VARCHAR(100) NOT NULL,
   description TEXT,
   entity_id UUID NOT NULL, -- table or view name, maybe others in the future
   entity_type_id UUID NOT NULL REFERENCES workflow.entity_types(id),
   
   -- Trigger conditions
   trigger_type_id UUID NOT NULL REFERENCES workflow.trigger_types(id),

   trigger_conditions JSONB NULL, -- When to trigger

   -- Visual editor state
   canvas_layout JSONB DEFAULT '{}',

   -- Control
   is_active BOOLEAN NOT NULL DEFAULT TRUE,
   
   created_date TIMESTAMP NOT NULL DEFAULT NOW(),
   updated_date TIMESTAMP NOT NULL DEFAULT NOW(),
   created_by UUID NOT NULL REFERENCES core.users(id),
   updated_by UUID NOT NULL REFERENCES core.users(id),

   deactivated_by UUID NULL REFERENCES core.users(id)
);

-- Version: 1.67
-- Description: Create table automation_executions (supports both automation and manual triggers)
CREATE TABLE workflow.automation_executions (
   id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
   automation_rules_id UUID REFERENCES workflow.automation_rules(id),  -- NULLABLE for manual executions
   entity_type VARCHAR(50) NOT NULL,
   trigger_data JSONB,
   actions_executed JSONB,
   status VARCHAR(20) NOT NULL, -- 'success', 'failed', 'partial', 'queued', 'processing'
   error_message TEXT,
   execution_time_ms INTEGER,
   executed_at TIMESTAMP NOT NULL DEFAULT NOW(),
   trigger_source VARCHAR(20) NOT NULL DEFAULT 'automation',  -- 'automation' or 'manual'
   executed_by UUID REFERENCES core.users(id),                -- User who triggered (required for manual)
   action_type VARCHAR(100)                                   -- Action type for manual executions
);

CREATE INDEX idx_automation_executions_trigger_source ON workflow.automation_executions(trigger_source);
CREATE INDEX idx_automation_executions_executed_by ON workflow.automation_executions(executed_by);
CREATE INDEX idx_automation_executions_status ON workflow.automation_executions(status);

-- Version: 1.68
-- Description: Create table action_templates
CREATE TABLE workflow.action_templates (
   id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
   name VARCHAR(100) NOT NULL UNIQUE,
   description TEXT,
   action_type VARCHAR(50) NOT NULL,
   icon VARCHAR(100) NULL,
   default_config JSONB NOT NULL,
   created_date TIMESTAMP NOT NULL DEFAULT NOW(),
   created_by UUID NOT NULL REFERENCES core.users(id),
   is_active BOOLEAN NOT NULL DEFAULT TRUE,
   deactivated_by UUID NULL REFERENCES core.users(id)
);

-- Version: 1.69
-- Description: Create table rule_actions
CREATE TABLE workflow.rule_actions (
   id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
   automation_rules_id UUID NOT NULL REFERENCES workflow.automation_rules(id),
   name VARCHAR(100) NOT NULL,
   description TEXT,
   action_config JSONB NOT NULL,
   is_active BOOLEAN DEFAULT TRUE,
   template_id UUID NULL REFERENCES workflow.action_templates(id),
   deactivated_by UUID NULL REFERENCES core.users(id)
);

-- Version: 1.70
-- Description: Create table rule_dependencies
CREATE TABLE workflow.rule_dependencies (
   id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
   parent_rule_id UUID REFERENCES workflow.automation_rules(id),
   child_rule_id UUID REFERENCES workflow.automation_rules(id)
);

-- Version: 1.71
-- Description: Create table entities
CREATE TABLE workflow.entities (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL UNIQUE,
    entity_type_id UUID NOT NULL REFERENCES workflow.entity_types(id),
    schema_name VARCHAR(50) DEFAULT 'public',
    is_active BOOLEAN DEFAULT TRUE,
    created_date TIMESTAMP DEFAULT NOW(),
    deactivated_by UUID NULL REFERENCES core.users(id)
);

-- Version: 1.72
-- Description: Create table notification_deliveries
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

-- Version: 1.73
-- Description: Create table allocation_results
CREATE TABLE workflow.allocation_results (
    id UUID PRIMARY KEY,
    idempotency_key VARCHAR(255) UNIQUE NOT NULL,
    allocation_data JSONB NOT NULL,
    created_date TIMESTAMP NOT NULL
);
CREATE INDEX idx_allocation_idempotency ON workflow.allocation_results(idempotency_key);

-- Migration: Create table_configs table for storing table configurations
-- Version: 1.74
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
    is_system BOOLEAN NOT NULL DEFAULT FALSE,

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
COMMENT ON COLUMN config.table_configs.is_system IS 'When true, this configuration is system-defined and cannot be deleted via the API';


-- Migration: Create page_configs for storing page configurations
-- Version: 1.75
-- Description: Create table for storing page configurations
CREATE TABLE IF NOT EXISTS config.page_configs (
   id UUID PRIMARY KEY,
   name TEXT NOT NULL,
   user_id UUID NULL,
   is_default BOOLEAN NOT NULL DEFAULT FALSE,

   CONSTRAINT fk_page_configs_user FOREIGN KEY (user_id) REFERENCES core.users(id) ON DELETE CASCADE,
   CONSTRAINT unique_page_config_name_user UNIQUE (name, user_id),
   CONSTRAINT check_default_no_user CHECK (
      (is_default = true AND user_id IS NULL) OR
      (is_default = false AND user_id IS NOT NULL)
   )
);

-- Create unique partial index to ensure only one default per page name
CREATE UNIQUE INDEX IF NOT EXISTS unique_default_page_config
   ON config.page_configs (name)
   WHERE is_default = true;

COMMENT ON INDEX config.unique_default_page_config IS 'Ensures only one default page configuration exists per page name';
COMMENT ON CONSTRAINT check_default_no_user ON config.page_configs IS 'Ensures default configs have no user_id and non-default configs have a user_id';

-- Migration: Create page_tab_configs for storing tab configurations within pages
-- Version: 1.76
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

-- Version: 1.77
-- Description: Create forms table for form configurations
CREATE TABLE IF NOT EXISTS config.forms (
   id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
   name VARCHAR(255) NOT NULL,
   is_reference_data BOOLEAN DEFAULT false,
   allow_inline_create BOOLEAN DEFAULT true,
   UNIQUE(name)
);

-- Create indexes for forms
CREATE INDEX IF NOT EXISTS idx_forms_name ON config.forms(name);
CREATE INDEX idx_forms_reference_data ON config.forms(is_reference_data);
CREATE INDEX idx_forms_inline_create ON config.forms(allow_inline_create);

-- Comments
COMMENT ON TABLE config.forms IS 'Stores form configuration definitions';
COMMENT ON COLUMN config.forms.id IS 'Unique identifier for the form';
COMMENT ON COLUMN config.forms.name IS 'Unique name for the form configuration';
COMMENT ON COLUMN config.forms.is_reference_data IS
    'If true, this form represents stable reference data managed by admins only (no inline creation allowed)';
COMMENT ON COLUMN config.forms.allow_inline_create IS
    'If true, this form can be embedded for inline entity creation within other forms';

-- Version: 1.78
-- Description: Create form_fields table for form field configurations
CREATE TABLE IF NOT EXISTS config.form_fields (
   id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
   form_id UUID NOT NULL,
   entity_id UUID NOT NULL,
   name VARCHAR(255) NOT NULL,
   label VARCHAR(255) NOT NULL,
   field_type VARCHAR(50) NOT NULL,
   field_order INTEGER NOT NULL,
   required BOOLEAN DEFAULT false,
   config JSONB NOT NULL DEFAULT '{}'::jsonb,

   CONSTRAINT fk_form_fields_form FOREIGN KEY (form_id) REFERENCES config.forms(id) ON DELETE CASCADE,
   CONSTRAINT fk_form_fields_entity FOREIGN KEY (entity_id) REFERENCES workflow.entities(id) ON DELETE CASCADE,
   UNIQUE(form_id, entity_id, name)
);

-- Create indexes for form_fields
CREATE INDEX IF NOT EXISTS idx_form_fields_form_id ON config.form_fields(form_id);
CREATE INDEX IF NOT EXISTS idx_form_fields_entity_id ON config.form_fields(entity_id);
CREATE INDEX IF NOT EXISTS idx_form_fields_field_order ON config.form_fields(form_id, field_order);
CREATE INDEX IF NOT EXISTS idx_form_fields_config ON config.form_fields USING GIN (config);

-- Comments
COMMENT ON TABLE config.form_fields IS 'Stores individual field configurations for forms';
COMMENT ON COLUMN config.form_fields.id IS 'Unique identifier for the form field';
COMMENT ON COLUMN config.form_fields.form_id IS 'Foreign key reference to the parent form';
COMMENT ON COLUMN config.form_fields.entity_id IS 'Foreign key reference to the entity (table) this field belongs to';
COMMENT ON COLUMN config.form_fields.name IS 'Field name (column name in the entity table)';
COMMENT ON COLUMN config.form_fields.label IS 'Display label for the field';
COMMENT ON COLUMN config.form_fields.field_type IS 'Type of field (text, select, checkbox, etc.)';
COMMENT ON COLUMN config.form_fields.field_order IS 'Display order of the field within the form';
COMMENT ON COLUMN config.form_fields.required IS 'Whether the field is required';
COMMENT ON COLUMN config.form_fields.config IS 'JSONB configuration for field-specific settings (parent_entity_id, foreign_key_column, execution_order, etc.)';

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
    tc.is_system,
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

-- Version: 1.79
-- Description: Create table page_actions (base table for all action types)
CREATE TABLE config.page_actions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    page_config_id UUID NOT NULL,
    action_type config.action_type NOT NULL,
    action_order INT NOT NULL DEFAULT 1,
    is_active BOOLEAN DEFAULT TRUE,

    CONSTRAINT fk_page_actions_page FOREIGN KEY (page_config_id)
        REFERENCES config.page_configs(id) ON DELETE CASCADE
);

-- Version: 1.80
-- Description: Create table page_action_buttons
CREATE TABLE config.page_action_buttons (
    action_id UUID PRIMARY KEY,
    label TEXT NOT NULL,
    icon TEXT,
    target_path TEXT NOT NULL,
    variant config.button_variant DEFAULT 'default',
    alignment config.button_alignment DEFAULT 'right',
    confirmation_prompt TEXT,

    CONSTRAINT fk_button_action FOREIGN KEY (action_id)
        REFERENCES config.page_actions(id) ON DELETE CASCADE
);

-- Version: 1.81
-- Description: Create table page_action_dropdowns
CREATE TABLE config.page_action_dropdowns (
    action_id UUID PRIMARY KEY,
    label TEXT NOT NULL,
    icon TEXT,

    CONSTRAINT fk_dropdown_action FOREIGN KEY (action_id)
        REFERENCES config.page_actions(id) ON DELETE CASCADE
);

-- Version: 1.82
-- Description: Create table page_action_dropdown_items
CREATE TABLE config.page_action_dropdown_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    dropdown_action_id UUID NOT NULL,
    label TEXT NOT NULL,
    target_path TEXT NOT NULL,
    item_order INT NOT NULL,

    CONSTRAINT fk_dropdown_items FOREIGN KEY (dropdown_action_id)
        REFERENCES config.page_action_dropdowns(action_id) ON DELETE CASCADE
);

-- Version: 1.83
-- Description: Add entity_schema and entity_table columns to form_fields
ALTER TABLE config.form_fields
    ADD COLUMN entity_schema TEXT NOT NULL DEFAULT 'core',
    ADD COLUMN entity_table TEXT NOT NULL DEFAULT 'users';

-- Remove defaults after adding columns (defaults were only for existing rows)
ALTER TABLE config.form_fields
    ALTER COLUMN entity_schema DROP DEFAULT,
    ALTER COLUMN entity_table DROP DEFAULT;

-- Create indexes for filtering and ordering
CREATE INDEX IF NOT EXISTS idx_form_fields_entity_schema ON config.form_fields(entity_schema);
CREATE INDEX IF NOT EXISTS idx_form_fields_entity_table ON config.form_fields(entity_table);
CREATE INDEX IF NOT EXISTS idx_form_fields_schema_table ON config.form_fields(entity_schema, entity_table);

-- Add column comments
COMMENT ON COLUMN config.form_fields.entity_schema IS 'Database schema name for the entity this field belongs to';
COMMENT ON COLUMN config.form_fields.entity_table IS 'Database table name for the entity this field belongs to';

-- Version: 1.84
-- Description: Create flexible page content blocks system for user-customizable layouts
CREATE TABLE IF NOT EXISTS config.page_content (
   id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
   page_config_id UUID NOT NULL,

   -- Content identification
   content_type config.content_type NOT NULL,
   label TEXT,

   -- Content references (only one should be set based on content_type)
   table_config_id UUID NULL,
   form_id UUID NULL,

   -- Simple ordering for stacked layouts
   order_index INT DEFAULT 0,

   -- Container/nesting (for tabs, accordions, sections, etc.)
   parent_id UUID NULL,

   -- ALL layout configuration in JSON (Tailwind-compatible)
   -- Stores: responsive column spans, grid config, gap, custom classes, etc.
   -- Example: {"colSpan":{"default":12,"md":6},"gap":"gap-4"}
   layout JSONB DEFAULT '{}'::jsonb,

   -- Display flags
   is_visible BOOLEAN DEFAULT TRUE,
   is_default BOOLEAN DEFAULT FALSE,  -- For tabs: which tab is active by default

   -- Foreign keys
   CONSTRAINT fk_page_content_page_config
      FOREIGN KEY (page_config_id) REFERENCES config.page_configs(id) ON DELETE CASCADE,
   CONSTRAINT fk_page_content_table
      FOREIGN KEY (table_config_id) REFERENCES config.table_configs(id) ON DELETE CASCADE,
   CONSTRAINT fk_page_content_form
      FOREIGN KEY (form_id) REFERENCES config.forms(id) ON DELETE CASCADE,
   CONSTRAINT fk_page_content_parent
      FOREIGN KEY (parent_id) REFERENCES config.page_content(id) ON DELETE CASCADE,

   -- Business rule: content must reference appropriate entity
   CONSTRAINT check_content_reference CHECK (
      (content_type = 'table' AND table_config_id IS NOT NULL) OR
      (content_type = 'form' AND form_id IS NOT NULL) OR
      (content_type IN ('tabs', 'container', 'text', 'chart'))
   )
);

-- Indexes for performance
CREATE INDEX idx_page_content_config ON config.page_content(page_config_id);
CREATE INDEX idx_page_content_table ON config.page_content(table_config_id);
CREATE INDEX idx_page_content_form ON config.page_content(form_id);
CREATE INDEX idx_page_content_parent ON config.page_content(parent_id);
CREATE INDEX idx_page_content_order ON config.page_content(page_config_id, order_index);
CREATE INDEX idx_page_content_layout ON config.page_content USING GIN (layout);

-- Comments
COMMENT ON TABLE config.page_content IS 'User-customizable content blocks (tables, forms, charts, etc.) with flexible JSON layout';
COMMENT ON COLUMN config.page_content.page_config_id IS 'References page_configs for user-specific or default layouts';
COMMENT ON COLUMN config.page_content.content_type IS 'Type of content: table, form, tabs, container, text, chart';
COMMENT ON COLUMN config.page_content.parent_id IS 'Parent content block ID for nested content (e.g., tabs within a tabs container)';
COMMENT ON COLUMN config.page_content.layout IS 'JSONB configuration for responsive grid layout (Tailwind-compatible), spacing, and styling';
COMMENT ON COLUMN config.page_content.order_index IS 'Display order for simple stacked layouts (0-based)';
COMMENT ON COLUMN config.page_content.is_default IS 'For tabs: indicates which tab is active by default';

CREATE OR REPLACE VIEW sales.orders_base AS
SELECT
   o.id AS orders_id,
   o.number AS orders_number,
   o.order_date AS orders_order_date,
   o.due_date AS orders_due_date,
   o.created_date AS orders_created_date,
   o.updated_date AS orders_updated_date,
   o.order_fulfillment_status_id AS orders_fulfillment_status_id,
   o.customer_id AS orders_customer_id,
   -- New financial and address fields
   o.billing_address_id AS orders_billing_address_id,
   o.shipping_address_id AS orders_shipping_address_id,
   o.payment_term_id AS orders_payment_term_id,
   o.notes AS orders_notes,
   o.subtotal AS orders_subtotal,
   o.tax_rate AS orders_tax_rate,
   o.tax_amount AS orders_tax_amount,
   o.shipping_cost AS orders_shipping_cost,
   o.total_amount AS orders_total_amount,
   o.currency_id AS orders_currency_id,

   c.id AS customers_id,
   c.name AS customers_name,
   c.contact_id AS customers_contact_infos_id,
   c.delivery_address_id AS customers_delivery_address_id,
   c.notes AS customers_notes,
   c.created_date AS customers_created_date,
   c.updated_date AS customers_updated_date,

   ofs.name AS order_fulfillment_statuses_name,
   ofs.description AS order_fulfillment_statuses_description,

   pt.id AS payment_term_id,
   pt.name AS payment_term_name,
   pt.description AS payment_term_description
FROM sales.orders o
   INNER JOIN sales.customers c ON o.customer_id = c.id
   LEFT JOIN sales.order_fulfillment_statuses ofs ON o.order_fulfillment_status_id = ofs.id
   LEFT JOIN core.payment_terms pt ON o.payment_term_id = pt.id;

CREATE OR REPLACE VIEW sales.order_line_items_base AS
SELECT
   oli.id AS order_line_item_id,
   oli.order_id AS order_line_item_order_id,
   oli.product_id AS order_line_item_product_id,
   oli.description AS order_line_item_description,
   oli.quantity AS order_line_item_quantity,
   oli.unit_price AS order_line_item_unit_price,
   oli.discount AS order_line_item_discount,
   oli.discount_type AS order_line_item_discount_type,
   oli.line_total AS order_line_item_line_total,
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
    ar.canvas_layout,
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
      ra.is_active,
      ra.template_id,
      at.name as template_name,
      at.action_type as template_action_type,
      at.icon as template_icon,
      at.default_config as template_default_config
   FROM workflow.rule_actions ra
   LEFT JOIN workflow.action_templates at ON ra.template_id = at.id;

-- Version: 1.85
-- Description: Create purchase order status table
CREATE TABLE procurement.purchase_order_statuses (
   id UUID PRIMARY KEY,
   name VARCHAR(50) NOT NULL UNIQUE,
   description TEXT,
   sort_order INTEGER DEFAULT 1000
);

-- Version: 1.86
-- Description: Create purchase order line item status table
CREATE TABLE procurement.purchase_order_line_item_statuses (
   id UUID PRIMARY KEY,
   name VARCHAR(50) NOT NULL UNIQUE,
   description TEXT,
   sort_order INTEGER DEFAULT 1000
);

-- Version: 1.87
-- Description: Create purchase orders table
CREATE TABLE procurement.purchase_orders (
   id UUID PRIMARY KEY,
   order_number VARCHAR(100) NOT NULL UNIQUE,
   supplier_id UUID NOT NULL,
   purchase_order_status_id UUID NOT NULL,

   -- Delivery information
   delivery_warehouse_id UUID NOT NULL,
   delivery_location_id UUID NULL,
   delivery_street_id UUID NULL,

   -- Dates
   order_date TIMESTAMP NOT NULL,
   expected_delivery_date TIMESTAMP NOT NULL,
   actual_delivery_date TIMESTAMP NULL,

   -- Financial
   subtotal NUMERIC(10,2) NOT NULL DEFAULT 0.00,
   tax_amount NUMERIC(10,2) NOT NULL DEFAULT 0.00,
   shipping_cost NUMERIC(10,2) NOT NULL DEFAULT 0.00,
   total_amount NUMERIC(10,2) NOT NULL DEFAULT 0.00,
   currency_id UUID NOT NULL,

   -- Workflow
   requested_by UUID NOT NULL,
   approved_by UUID NULL,
   approved_date TIMESTAMP NULL,

   -- Notes and reference
   notes TEXT NULL,
   supplier_reference_number VARCHAR(100) NULL,

   -- Standard audit fields
   created_by UUID NOT NULL,
   updated_by UUID NOT NULL,
   created_date TIMESTAMP NOT NULL,
   updated_date TIMESTAMP NOT NULL,

   FOREIGN KEY (supplier_id) REFERENCES procurement.suppliers(id) ON DELETE RESTRICT,
   FOREIGN KEY (purchase_order_status_id) REFERENCES procurement.purchase_order_statuses(id) ON DELETE RESTRICT,
   FOREIGN KEY (delivery_warehouse_id) REFERENCES inventory.warehouses(id) ON DELETE RESTRICT,
   FOREIGN KEY (delivery_location_id) REFERENCES inventory.inventory_locations(id) ON DELETE SET NULL,
   FOREIGN KEY (delivery_street_id) REFERENCES geography.streets(id) ON DELETE SET NULL,
   FOREIGN KEY (currency_id) REFERENCES core.currencies(id) ON DELETE RESTRICT,
   FOREIGN KEY (requested_by) REFERENCES core.users(id) ON DELETE RESTRICT,
   FOREIGN KEY (approved_by) REFERENCES core.users(id) ON DELETE SET NULL,
   FOREIGN KEY (created_by) REFERENCES core.users(id) ON DELETE RESTRICT,
   FOREIGN KEY (updated_by) REFERENCES core.users(id) ON DELETE RESTRICT,

   -- Constraint: Must have either warehouse OR street address for delivery
   CONSTRAINT check_delivery_location CHECK (
      (delivery_warehouse_id IS NOT NULL) OR (delivery_street_id IS NOT NULL)
   )
);

-- Indexes for common queries
CREATE INDEX idx_purchase_orders_supplier ON procurement.purchase_orders(supplier_id);
CREATE INDEX idx_purchase_orders_status ON procurement.purchase_orders(purchase_order_status_id);
CREATE INDEX idx_purchase_orders_order_date ON procurement.purchase_orders(order_date DESC);
CREATE INDEX idx_purchase_orders_expected_delivery ON procurement.purchase_orders(expected_delivery_date);
CREATE INDEX idx_purchase_orders_requested_by ON procurement.purchase_orders(requested_by);

-- Version: 1.88
-- Description: Create purchase order line items table
CREATE TABLE procurement.purchase_order_line_items (
   id UUID PRIMARY KEY,
   purchase_order_id UUID NOT NULL,
   supplier_product_id UUID NOT NULL,

   -- Quantities
   quantity_ordered INT NOT NULL,
   quantity_received INT NOT NULL DEFAULT 0,
   quantity_cancelled INT NOT NULL DEFAULT 0,

   -- Pricing (captured at time of order)
   unit_cost NUMERIC(10,2) NOT NULL,
   discount NUMERIC(10,2) NULL DEFAULT 0.00,
   line_total NUMERIC(10,2) NOT NULL,

   -- Status and dates
   line_item_status_id UUID NOT NULL,
   expected_delivery_date TIMESTAMP NULL,
   actual_delivery_date TIMESTAMP NULL,

   -- Notes
   notes TEXT NULL,

   -- Standard audit fields
   created_by UUID NOT NULL,
   updated_by UUID NOT NULL,
   created_date TIMESTAMP NOT NULL,
   updated_date TIMESTAMP NOT NULL,

   FOREIGN KEY (purchase_order_id) REFERENCES procurement.purchase_orders(id) ON DELETE CASCADE,
   FOREIGN KEY (supplier_product_id) REFERENCES procurement.supplier_products(id) ON DELETE RESTRICT,
   FOREIGN KEY (line_item_status_id) REFERENCES procurement.purchase_order_line_item_statuses(id) ON DELETE RESTRICT,
   FOREIGN KEY (created_by) REFERENCES core.users(id) ON DELETE RESTRICT,
   FOREIGN KEY (updated_by) REFERENCES core.users(id) ON DELETE RESTRICT,

   -- Constraint: Received + Cancelled should not exceed Ordered
   CONSTRAINT check_quantities CHECK (
      quantity_received + quantity_cancelled <= quantity_ordered
   )
);

-- Indexes
CREATE INDEX idx_po_line_items_po ON procurement.purchase_order_line_items(purchase_order_id);
CREATE INDEX idx_po_line_items_supplier_product ON procurement.purchase_order_line_items(supplier_product_id);
CREATE INDEX idx_po_line_items_status ON procurement.purchase_order_line_items(line_item_status_id);

-- Version: 1.89
-- Description: Drop legacy page_tab_configs table (replaced by page_content system)
DROP TABLE IF EXISTS config.page_tab_configs;

-- Version: 1.90
-- Description: Increase table_name column size to accommodate schema prefixes
ALTER TABLE core.table_access ALTER COLUMN table_name TYPE VARCHAR(100);

-- Version: 1.91
-- Description: Add schema prefixes to existing table names
-- Assets schema
UPDATE core.table_access SET table_name = 'assets.asset_types' WHERE table_name = 'asset_types';
UPDATE core.table_access SET table_name = 'assets.asset_conditions' WHERE table_name = 'asset_conditions';
UPDATE core.table_access SET table_name = 'assets.tags' WHERE table_name = 'tags';
UPDATE core.table_access SET table_name = 'assets.asset_tags' WHERE table_name = 'asset_tags';
UPDATE core.table_access SET table_name = 'assets.assets' WHERE table_name = 'assets';
UPDATE core.table_access SET table_name = 'assets.user_assets' WHERE table_name = 'user_assets';

-- Geography schema
UPDATE core.table_access SET table_name = 'geography.countries' WHERE table_name = 'countries';
UPDATE core.table_access SET table_name = 'geography.regions' WHERE table_name = 'regions';
UPDATE core.table_access SET table_name = 'geography.cities' WHERE table_name = 'cities';
UPDATE core.table_access SET table_name = 'geography.streets' WHERE table_name = 'streets';

-- HR schema
UPDATE core.table_access SET table_name = 'hr.user_approval_status' WHERE table_name = 'user_approval_status';
UPDATE core.table_access SET table_name = 'hr.titles' WHERE table_name = 'titles';
UPDATE core.table_access SET table_name = 'hr.offices' WHERE table_name = 'offices';
UPDATE core.table_access SET table_name = 'hr.homes' WHERE table_name = 'homes';
UPDATE core.table_access SET table_name = 'assets.approval_status' WHERE table_name = 'approval_status';
UPDATE core.table_access SET table_name = 'hr.reports_to' WHERE table_name = 'reports_to';
UPDATE core.table_access SET table_name = 'hr.user_approval_comments' WHERE table_name = 'user_approval_comments';

-- Core schema
UPDATE core.table_access SET table_name = 'core.users' WHERE table_name = 'users';
UPDATE core.table_access SET table_name = 'core.contact_infos' WHERE table_name = 'contact_infos';
UPDATE core.table_access SET table_name = 'core.roles' WHERE table_name = 'roles';
UPDATE core.table_access SET table_name = 'core.user_roles' WHERE table_name = 'user_roles';
UPDATE core.table_access SET table_name = 'core.pages' WHERE table_name = 'pages';
UPDATE core.table_access SET table_name = 'core.role_pages' WHERE table_name = 'role_pages';
UPDATE core.table_access SET table_name = 'core.table_access' WHERE table_name = 'table_access';

-- Sales schema
UPDATE core.table_access SET table_name = 'assets.fulfillment_status' WHERE table_name = 'fulfillment_status';
UPDATE core.table_access SET table_name = 'sales.customers' WHERE table_name = 'customers';
UPDATE core.table_access SET table_name = 'sales.orders' WHERE table_name = 'orders';
UPDATE core.table_access SET table_name = 'sales.order_line_items' WHERE table_name = 'order_line_items';
UPDATE core.table_access SET table_name = 'sales.order_fulfillment_statuses' WHERE table_name = 'order_fulfillment_statuses';
UPDATE core.table_access SET table_name = 'sales.line_item_fulfillment_statuses' WHERE table_name = 'line_item_fulfillment_statuses';

-- Products schema
UPDATE core.table_access SET table_name = 'products.brands' WHERE table_name = 'brands';
UPDATE core.table_access SET table_name = 'products.product_categories' WHERE table_name = 'product_categories';
UPDATE core.table_access SET table_name = 'products.products' WHERE table_name = 'products';
UPDATE core.table_access SET table_name = 'products.physical_attributes' WHERE table_name = 'physical_attributes';
UPDATE core.table_access SET table_name = 'products.product_costs' WHERE table_name = 'product_costs';
UPDATE core.table_access SET table_name = 'products.cost_history' WHERE table_name = 'cost_history';
UPDATE core.table_access SET table_name = 'products.quality_metrics' WHERE table_name = 'quality_metrics';

-- Inventory schema
UPDATE core.table_access SET table_name = 'inventory.warehouses' WHERE table_name = 'warehouses';
UPDATE core.table_access SET table_name = 'inventory.zones' WHERE table_name = 'zones';
UPDATE core.table_access SET table_name = 'inventory.inventory_locations' WHERE table_name = 'inventory_locations';
UPDATE core.table_access SET table_name = 'inventory.inventory_items' WHERE table_name = 'inventory_items';
UPDATE core.table_access SET table_name = 'inventory.lot_trackings' WHERE table_name = 'lot_trackings';
UPDATE core.table_access SET table_name = 'inventory.quality_inspections' WHERE table_name = 'quality_inspections';
UPDATE core.table_access SET table_name = 'inventory.serial_numbers' WHERE table_name = 'serial_numbers';
UPDATE core.table_access SET table_name = 'inventory.inventory_transactions' WHERE table_name = 'inventory_transactions';
UPDATE core.table_access SET table_name = 'inventory.inventory_adjustments' WHERE table_name = 'inventory_adjustments';
UPDATE core.table_access SET table_name = 'inventory.transfer_orders' WHERE table_name = 'transfer_orders';

-- Procurement schema
UPDATE core.table_access SET table_name = 'procurement.suppliers' WHERE table_name = 'suppliers';
UPDATE core.table_access SET table_name = 'procurement.supplier_products' WHERE table_name = 'supplier_products';
UPDATE core.table_access SET table_name = 'procurement.purchase_order_statuses' WHERE table_name = 'purchase_order_statuses';
UPDATE core.table_access SET table_name = 'procurement.purchase_order_line_item_statuses' WHERE table_name = 'purchase_order_line_item_statuses';
UPDATE core.table_access SET table_name = 'procurement.purchase_orders' WHERE table_name = 'purchase_orders';
UPDATE core.table_access SET table_name = 'procurement.purchase_order_line_items' WHERE table_name = 'purchase_order_line_items';

-- Config schema
UPDATE core.table_access SET table_name = 'config.table_configs' WHERE table_name = 'table_configs';
UPDATE core.table_access SET table_name = 'config.forms' WHERE table_name = 'forms';
UPDATE core.table_access SET table_name = 'config.form_fields' WHERE table_name = 'form_fields';
UPDATE core.table_access SET table_name = 'config.page_configs' WHERE table_name = 'page_configs';
UPDATE core.table_access SET table_name = 'config.page_content' WHERE table_name = 'page_content';
UPDATE core.table_access SET table_name = 'config.page_actions' WHERE table_name = 'page_actions';
UPDATE core.table_access SET table_name = 'config.page_action_buttons' WHERE table_name = 'page_action_buttons';
UPDATE core.table_access SET table_name = 'config.page_action_dropdowns' WHERE table_name = 'page_action_dropdowns';
UPDATE core.table_access SET table_name = 'config.page_action_dropdown_items' WHERE table_name = 'page_action_dropdown_items';

-- Version: 1.92
-- Description: Add chart_config_id column to page_content table for chart content type
ALTER TABLE config.page_content
ADD COLUMN chart_config_id UUID NULL;

ALTER TABLE config.page_content
ADD CONSTRAINT fk_page_content_chart
   FOREIGN KEY (chart_config_id) REFERENCES config.table_configs(id) ON DELETE CASCADE;

CREATE INDEX idx_page_content_chart ON config.page_content(chart_config_id);

-- Update CHECK constraint to require chart_config_id for chart content type
ALTER TABLE config.page_content
DROP CONSTRAINT check_content_reference;

ALTER TABLE config.page_content
ADD CONSTRAINT check_content_reference CHECK (
   (content_type = 'table' AND table_config_id IS NOT NULL) OR
   (content_type = 'form' AND form_id IS NOT NULL) OR
   (content_type = 'chart' AND chart_config_id IS NOT NULL) OR
   (content_type IN ('tabs', 'container', 'text'))
);

COMMENT ON COLUMN config.page_content.chart_config_id IS 'References table_configs for chart widget configurations';

-- Version: 1.93
-- Description: Convert GroupBy from object to array in table_configs for multi-dimensional grouping support
UPDATE config.table_configs
SET config = jsonb_set(
    config,
    '{data_source,0,group_by}',
    CASE
        WHEN config->'data_source'->0->'group_by' IS NOT NULL
         AND jsonb_typeof(config->'data_source'->0->'group_by') = 'object'
        THEN jsonb_build_array(config->'data_source'->0->'group_by')
        ELSE config->'data_source'->0->'group_by'
    END
)
WHERE config->'data_source'->0->'group_by' IS NOT NULL
  AND jsonb_typeof(config->'data_source'->0->'group_by') = 'object';

-- Version: 1.94
-- Description: Create workflow alert tables
CREATE TABLE workflow.alerts (
   id UUID NOT NULL,
   alert_type VARCHAR(100) NOT NULL,
   severity workflow.alert_severity NOT NULL,
   title TEXT NOT NULL,
   message TEXT NOT NULL,
   context JSONB DEFAULT '{}',
   source_entity_name VARCHAR(100) NULL,
   source_entity_id UUID NULL,
   source_rule_id UUID NULL,
   status workflow.alert_status NOT NULL DEFAULT 'active',
   expires_date TIMESTAMP NULL,
   created_date TIMESTAMP NOT NULL,
   updated_date TIMESTAMP NOT NULL,
   PRIMARY KEY (id),
   FOREIGN KEY (source_rule_id) REFERENCES workflow.automation_rules(id) ON DELETE SET NULL
);

CREATE INDEX idx_alerts_status ON workflow.alerts(status);
CREATE INDEX idx_alerts_severity ON workflow.alerts(severity);
CREATE INDEX idx_alerts_created_date ON workflow.alerts(created_date DESC);
CREATE INDEX idx_alerts_source_rule ON workflow.alerts(source_rule_id);
CREATE INDEX idx_alerts_expires_date ON workflow.alerts(expires_date) WHERE expires_date IS NOT NULL;
CREATE INDEX idx_alerts_source_entity_type_status ON workflow.alerts(source_entity_id, alert_type, status) WHERE source_entity_id IS NOT NULL;

-- Version: 1.95
-- Description: Create alert recipients table
CREATE TABLE workflow.alert_recipients (
   id UUID NOT NULL,
   alert_id UUID NOT NULL,
   recipient_type workflow.recipient_type NOT NULL,
   recipient_id UUID NOT NULL,
   created_date TIMESTAMP NOT NULL,
   PRIMARY KEY (id),
   FOREIGN KEY (alert_id) REFERENCES workflow.alerts(id) ON DELETE CASCADE,
   UNIQUE (alert_id, recipient_type, recipient_id)
);

CREATE INDEX idx_alert_recipients_alert ON workflow.alert_recipients(alert_id);
CREATE INDEX idx_alert_recipients_recipient ON workflow.alert_recipients(recipient_type, recipient_id);
CREATE INDEX idx_alert_recipients_lookup ON workflow.alert_recipients(recipient_id, recipient_type, alert_id);

-- Version: 1.96
-- Description: Create alert acknowledgments table
CREATE TABLE workflow.alert_acknowledgments (
   id UUID NOT NULL,
   alert_id UUID NOT NULL,
   acknowledged_by UUID NOT NULL,
   acknowledged_date TIMESTAMP NOT NULL,
   notes TEXT NULL,
   PRIMARY KEY (id),
   FOREIGN KEY (alert_id) REFERENCES workflow.alerts(id) ON DELETE CASCADE,
   FOREIGN KEY (acknowledged_by) REFERENCES core.users(id) ON DELETE CASCADE,
   UNIQUE (alert_id, acknowledged_by)
);

CREATE INDEX idx_alert_ack_alert ON workflow.alert_acknowledgments(alert_id);
CREATE INDEX idx_alert_ack_user ON workflow.alert_acknowledgments(acknowledged_by);

-- Version: 1.97
-- Description: Create enum_labels table for human-friendly display labels
CREATE TABLE config.enum_labels (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    enum_name VARCHAR(100) NOT NULL,      -- e.g., 'sales.discount_type' (schema.enum_name format)
    value VARCHAR(100) NOT NULL,          -- e.g., 'flat' (must match enum value)
    label VARCHAR(255) NOT NULL,          -- e.g., 'Flat Amount'
    sort_order INTEGER DEFAULT 1000,      -- Override pg_enum sort if needed
    created_date TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE(enum_name, value)
);

CREATE INDEX idx_enum_labels_name ON config.enum_labels(enum_name);

COMMENT ON TABLE config.enum_labels IS 'Human-friendly labels for PostgreSQL ENUM values';
COMMENT ON COLUMN config.enum_labels.enum_name IS 'Full enum name in schema.enum_name format (e.g., sales.discount_type)';
COMMENT ON COLUMN config.enum_labels.value IS 'Raw enum value that must match the PostgreSQL enum value';
COMMENT ON COLUMN config.enum_labels.label IS 'Human-friendly display label for the enum value';
COMMENT ON COLUMN config.enum_labels.sort_order IS 'Custom sort order (overrides pg_enum enumsortorder if needed)';

-- Version: 1.98
-- Description: Add itemized discount type
ALTER TYPE sales.discount_type ADD VALUE 'itemized';

-- Version: 1.990
-- Description: Create workflow action_permissions table for manual action authorization
CREATE TABLE workflow.action_permissions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    role_id UUID NOT NULL REFERENCES core.roles(id) ON DELETE CASCADE,
    action_type VARCHAR(100) NOT NULL,
    is_allowed BOOLEAN NOT NULL DEFAULT TRUE,
    constraints JSONB DEFAULT '{}',  -- Stubbed for future constraint implementation
    created_date TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_date TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE(role_id, action_type)
);

CREATE INDEX idx_action_permissions_role ON workflow.action_permissions(role_id);
CREATE INDEX idx_action_permissions_type ON workflow.action_permissions(action_type);

COMMENT ON TABLE workflow.action_permissions IS 'Controls which roles can manually execute workflow actions';
COMMENT ON COLUMN workflow.action_permissions.role_id IS 'Role that is granted this permission';
COMMENT ON COLUMN workflow.action_permissions.action_type IS 'Action type identifier (e.g., allocate_inventory, send_email)';
COMMENT ON COLUMN workflow.action_permissions.is_allowed IS 'Whether the role is allowed to execute this action';
COMMENT ON COLUMN workflow.action_permissions.constraints IS 'Future: JSONB constraints for fine-grained permission control';

-- Version: 1.991
-- Description: Seed default action permissions for admin role
-- Note: update_field is intentionally excluded from manual execution
INSERT INTO workflow.action_permissions (role_id, action_type, is_allowed)
SELECT r.id, action_type, true
FROM core.roles r
CROSS JOIN (VALUES
    ('allocate_inventory'),
    ('create_alert'),
    ('send_email'),
    ('send_notification'),
    ('seek_approval')
) AS actions(action_type)
WHERE r.name = 'admin';

-- Version: 1.992
-- Description: Add action edges for workflow branching (condition nodes)
CREATE TABLE workflow.action_edges (
   id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
   rule_id UUID NOT NULL REFERENCES workflow.automation_rules(id) ON DELETE CASCADE,
   source_action_id UUID REFERENCES workflow.rule_actions(id) ON DELETE CASCADE,
   target_action_id UUID NOT NULL REFERENCES workflow.rule_actions(id) ON DELETE CASCADE,
   edge_type VARCHAR(20) NOT NULL CHECK (edge_type IN ('start', 'sequence', 'true_branch', 'false_branch', 'always')),
   edge_order INTEGER DEFAULT 0,
   created_date TIMESTAMP NOT NULL DEFAULT NOW(),
   CONSTRAINT unique_edge UNIQUE(source_action_id, target_action_id, edge_type)
);

CREATE INDEX idx_action_edges_source ON workflow.action_edges(source_action_id);
CREATE INDEX idx_action_edges_target ON workflow.action_edges(target_action_id);
CREATE INDEX idx_action_edges_rule ON workflow.action_edges(rule_id);

COMMENT ON TABLE workflow.action_edges IS 'Defines edges between actions in a workflow graph for branching support';
COMMENT ON COLUMN workflow.action_edges.rule_id IS 'The automation rule this edge belongs to';
COMMENT ON COLUMN workflow.action_edges.source_action_id IS 'Source action (NULL for start edges)';
COMMENT ON COLUMN workflow.action_edges.target_action_id IS 'Target action to execute';
COMMENT ON COLUMN workflow.action_edges.edge_type IS 'Type: start, sequence, true_branch, false_branch, always';
COMMENT ON COLUMN workflow.action_edges.edge_order IS 'Order when multiple edges have same source (for deterministic traversal)';

-- Version: 1.993
-- Description: Create partitioned audit log table for workflow audit trail entries
CREATE TABLE workflow.audit_log (
    id              UUID DEFAULT gen_random_uuid(),
    entity_name     TEXT NOT NULL,
    entity_id       UUID NOT NULL,
    action          TEXT NOT NULL,
    message         TEXT NOT NULL,
    metadata        JSONB,
    rule_id         UUID,
    execution_id    UUID,
    user_id         UUID,
    created_date    TIMESTAMP NOT NULL DEFAULT NOW(),
    PRIMARY KEY (id, created_date)
) PARTITION BY RANGE (created_date);

CREATE TABLE workflow.audit_log_2026_02 PARTITION OF workflow.audit_log
    FOR VALUES FROM ('2026-02-01') TO ('2026-03-01');
CREATE TABLE workflow.audit_log_2026_03 PARTITION OF workflow.audit_log
    FOR VALUES FROM ('2026-03-01') TO ('2026-04-01');
CREATE TABLE workflow.audit_log_2026_04 PARTITION OF workflow.audit_log
    FOR VALUES FROM ('2026-04-01') TO ('2026-05-01');
CREATE TABLE workflow.audit_log_2026_05 PARTITION OF workflow.audit_log
    FOR VALUES FROM ('2026-05-01') TO ('2026-06-01');

CREATE TABLE workflow.audit_log_default PARTITION OF workflow.audit_log DEFAULT;

CREATE INDEX idx_audit_log_created ON workflow.audit_log USING BRIN (created_date);
CREATE INDEX idx_audit_log_entity ON workflow.audit_log(entity_name, entity_id);
CREATE INDEX idx_audit_log_rule ON workflow.audit_log(rule_id);

-- Version: 1.994
-- Description: Add action permissions for new workflow action types
INSERT INTO workflow.action_permissions (role_id, action_type, is_allowed)
SELECT r.id, action_type, true
FROM core.roles r
CROSS JOIN (VALUES
    ('create_entity'),
    ('lookup_entity'),
    ('transition_status'),
    ('log_audit_entry')
) AS actions(action_type)
WHERE r.name = 'admin';

-- Version: 1.995
-- Description: Add source_output column to action_edges for output port routing
ALTER TABLE workflow.action_edges ADD COLUMN source_output VARCHAR(50);

COMMENT ON COLUMN workflow.action_edges.source_output IS
  'Output port name this edge connects from. NULL for start and always edges.';

-- Backfill existing edges:
UPDATE workflow.action_edges SET source_output = 'true' WHERE edge_type = 'true_branch';
UPDATE workflow.action_edges SET source_output = 'false' WHERE edge_type = 'false_branch';
UPDATE workflow.action_edges SET source_output = 'success' WHERE edge_type = 'sequence';
-- start and always edges: source_output stays NULL

-- Normalize edge_type: true_branch/false_branch -> sequence (output port handles routing now)
UPDATE workflow.action_edges SET edge_type = 'sequence' WHERE edge_type IN ('true_branch', 'false_branch');

-- Update unique constraint to include source_output instead of edge_type
ALTER TABLE workflow.action_edges DROP CONSTRAINT unique_edge;
ALTER TABLE workflow.action_edges ADD CONSTRAINT unique_edge
  UNIQUE(source_action_id, target_action_id, source_output);

-- Version: 1.996
-- Description: Add is_default flag to automation_rules; update automation_rules_view
ALTER TABLE workflow.automation_rules ADD COLUMN is_default BOOLEAN NOT NULL DEFAULT FALSE;

COMMENT ON COLUMN workflow.automation_rules.is_default IS
  'True for system-seeded default workflows. Immutable after creation — cannot be modified via UpdateRule.';

CREATE OR REPLACE VIEW workflow.automation_rules_view AS
SELECT
    ar.id,
    ar.name,
    ar.description,
    ar.trigger_conditions,
    ar.canvas_layout,
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
    du.username as deactivated_by_username,

    -- Added in 1.996
    ar.is_default

FROM workflow.automation_rules ar
LEFT JOIN workflow.trigger_types tt ON ar.trigger_type_id = tt.id
LEFT JOIN workflow.entity_types et ON ar.entity_type_id = et.id
LEFT JOIN workflow.entities e ON ar.entity_id = e.id
LEFT JOIN core.users cu ON ar.created_by = cu.id
LEFT JOIN core.users uu ON ar.updated_by = uu.id
LEFT JOIN core.users du ON ar.deactivated_by = du.id
WHERE ar.is_active = true;

-- Version: 1.997
-- Description: Add workflow approval requests table for seek_approval async activity
CREATE TABLE workflow.approval_requests (
    approval_request_id UUID        NOT NULL,
    execution_id        UUID        NOT NULL REFERENCES workflow.automation_executions(id),
    rule_id             UUID        NOT NULL REFERENCES workflow.automation_rules(id),
    action_name         VARCHAR(100) NOT NULL,
    approvers           UUID[]      NOT NULL,
    approval_type       VARCHAR(20) NOT NULL DEFAULT 'any' CHECK (approval_type IN ('any', 'all', 'majority')),
    status              VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'approved', 'rejected', 'timed_out', 'expired')),
    timeout_hours       INT         NOT NULL DEFAULT 72,
    task_token          TEXT        NOT NULL,
    approval_message    TEXT,
    resolved_by         UUID        REFERENCES core.users(id),
    resolution_reason   TEXT,
    created_date        TIMESTAMP   NOT NULL DEFAULT NOW(),
    resolved_date       TIMESTAMP,

    PRIMARY KEY (approval_request_id)
);

CREATE INDEX idx_approval_requests_execution ON workflow.approval_requests(execution_id);
CREATE INDEX idx_approval_requests_status ON workflow.approval_requests(status) WHERE status = 'pending';
CREATE INDEX idx_approval_requests_approvers ON workflow.approval_requests USING GIN (approvers);

-- Version: 1.998
-- Description: Add unique constraint on inventory_items (product_id, location_id).
-- One stock position per product per location. Future lot tracking belongs in a
-- separate inventory_lots table, not in inventory_items.
ALTER TABLE inventory.inventory_items
    ADD CONSTRAINT unique_product_location UNIQUE (product_id, location_id);

-- Version: 1.999
-- Description: Add inventory put_away_tasks table. A put-away task is a single
-- directed work instruction for a floor worker: take N units of a product to a
-- specific warehouse location. Created from PO receipts or manually by supervisors.
CREATE TABLE inventory.put_away_tasks (
    id               UUID          NOT NULL,
    product_id       UUID          NOT NULL REFERENCES products.products(id),
    location_id      UUID          NOT NULL REFERENCES inventory.inventory_locations(id),
    quantity         INT           NOT NULL CHECK (quantity > 0),
    reference_number VARCHAR(100)  NOT NULL DEFAULT '',
    status           VARCHAR(20)   NOT NULL DEFAULT 'pending'
                         CHECK (status IN ('pending', 'in_progress', 'completed', 'cancelled')),
    assigned_to      UUID          REFERENCES core.users(id),
    assigned_at      TIMESTAMP,
    completed_by     UUID          REFERENCES core.users(id),
    completed_at     TIMESTAMP,
    created_by       UUID          NOT NULL REFERENCES core.users(id),
    created_date     TIMESTAMP     NOT NULL,
    updated_date     TIMESTAMP     NOT NULL,
    PRIMARY KEY (id)
);

CREATE INDEX idx_put_away_tasks_status   ON inventory.put_away_tasks(status);
CREATE INDEX idx_put_away_tasks_product  ON inventory.put_away_tasks(product_id);
CREATE INDEX idx_put_away_tasks_location ON inventory.put_away_tasks(location_id);
CREATE INDEX idx_put_away_tasks_assigned ON inventory.put_away_tasks(assigned_to)
    WHERE assigned_to IS NOT NULL;
