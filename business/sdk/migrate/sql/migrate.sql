-- UPDATED SCHEMA - Primary keys changed to 'id', foreign keys remain descriptive

-- Version: 1.01
-- Description: Create table asset_types
CREATE TABLE asset_types (
   id UUID NOT NULL,
   name TEXT NOT NULL,
   description TEXT NULL,
   PRIMARY KEY (id),
   UNIQUE (name)
);

-- Version: 1.02
-- Description: Create table asset_conditions
CREATE TABLE asset_conditions (
   id UUID NOT NULL,
   name TEXT NOT NULL,
   description TEXT NULL,
   PRIMARY KEY (id),
   UNIQUE (name)
);

-- Version: 1.03
-- Description: Create table countries
CREATE TABLE countries (
   id UUID NOT NULL,
   number INT NOT NULL,
   name TEXT NOT NULL,
   alpha_2 VARCHAR(2) NOT NULL,
   alpha_3 VARCHAR(3) NOT NULL,
   PRIMARY KEY (id)
);

-- Version: 1.04
-- Description: Create table regions
CREATE TABLE regions (
   id UUID NOT NULL,
   country_id UUID NOT NULL,
   name TEXT NOT NULL,
   code TEXT NOT NULL,
   -- TODO: determine if these should be enforced unique.
   PRIMARY KEY (id),
   FOREIGN KEY (country_id) REFERENCES countries(id) ON DELETE CASCADE
);

-- Version: 1.05
-- Description: create table cities
CREATE TABLE cities (
   id UUID NOT NULL,
   region_id UUID NOT NULL,
   name TEXT NOT NULL,
   PRIMARY KEY (id),
   UNIQUE (region_id, name),
   FOREIGN KEY (region_id) REFERENCES regions(id) ON DELETE CASCADE
);

-- Version: 1.06
-- Description: create table streets
CREATE TABLE streets (
   id UUID NOT NULL,
   city_id UUID NOT NULL,
   line_1 TEXT NOT NULL,
   line_2 TEXT NULL,
   postal_code VARCHAR(20) NULL,
   PRIMARY KEY (id),
   FOREIGN KEY (city_id) REFERENCES cities(id) ON DELETE SET NULL-- Check this cascade relationship
);

-- Version: 1.07
-- Description: Create table user_approval_status
CREATE TABLE user_approval_status (
   id UUID NOT NULL, 
   icon_id UUID NOT NULL, 
   name TEXT NOT NULL,
   PRIMARY KEY (id)
);

-- Version: 1.08
-- Description: Create table titles
CREATE TABLE titles (
   id UUID NOT NULL, 
   name TEXT NOT NULL,
   description TEXT NULL,
   PRIMARY KEY (id)
);

-- Version: 1.09
-- Description: Create table offices
CREATE TABLE offices (
   id UUID NOT NULL, 
   name TEXT NOT NULL,
   street_id UUID NOT NULL,
   PRIMARY KEY (id),
   FOREIGN KEY (street_id) REFERENCES streets(id) ON DELETE CASCADE
);

-- Version: 1.10
-- Description: Create table users
CREATE TABLE users (
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
   date_created TIMESTAMP NOT NULL,
   date_updated TIMESTAMP NOT NULL,
   PRIMARY KEY (id),
   FOREIGN KEY (requested_by) REFERENCES users(id) ON DELETE SET NULL, -- we don't want to delete someone if their boss is deleted
   FOREIGN KEY (approved_by) REFERENCES users(id) ON DELETE SET NULL, -- we don't want to delete someone if their boss is deleted
   FOREIGN KEY (title_id) REFERENCES titles(id) ON DELETE CASCADE,
   FOREIGN KEY (office_id) REFERENCES offices(id) ON DELETE CASCADE,
   FOREIGN KEY (user_approval_status_id) REFERENCES user_approval_status(id) ON DELETE CASCADE
);

-- Version: 1.11
-- Description: Create table valid_assets
CREATE TABLE valid_assets (
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
   date_created TIMESTAMP NOT NULL,
   date_updated TIMESTAMP NOT NULL,
   created_by UUID NOT NULL,
   updated_by UUID NOT NULL,
   PRIMARY KEY (id),
   
   -- UNIQUE named constraint
   CONSTRAINT unique_asset_name UNIQUE (name),

   -- named foreign keys
   CONSTRAINT fk_assets_type_id FOREIGN KEY (type_id) REFERENCES asset_types(id) ON DELETE CASCADE,
   CONSTRAINT fk_assets_created_by FOREIGN KEY (created_by) REFERENCES users(id) ON DELETE CASCADE,
   CONSTRAINT fk_assets_updated_by FOREIGN KEY (updated_by) REFERENCES users(id) ON DELETE CASCADE
);

-- Version: 1.12
-- Description: Create table homes
CREATE TABLE homes (
   id UUID NOT NULL,
   TYPE TEXT NOT NULL,
   user_id UUID NOT NULL,
   address_1 TEXT NOT NULL,
   address_2 TEXT NULL,
   zip_code TEXT NOT NULL,
   city TEXT NOT NULL,
   state TEXT NOT NULL,
   country TEXT NOT NULL,
   date_created TIMESTAMP NOT NULL,
   date_updated TIMESTAMP NOT NULL,
   PRIMARY KEY (id),
   FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- Version: 1.13
-- Description: Add approval status 
CREATE TABLE approval_status (
   id UUID NOT NULL, 
   icon_id UUID NOT NULL, 
   name TEXT NOT NULL,
   PRIMARY KEY (id)
);

-- Version: 1.14
-- Description: Add fulfillment status
CREATE TABLE fulfillment_status (
   id UUID NOT NULL, 
   icon_id UUID NOT NULL, 
   name TEXT NOT NULL,
   PRIMARY KEY (id)
);

-- Version: 1.15
-- Description: Add Tags
CREATE TABLE tags (
   id UUID NOT NULL, 
   name TEXT NOT NULL,
   description TEXT NULL,
   PRIMARY KEY (id)
);

-- Version: 1.16
-- Description: Add asset_tags
CREATE TABLE asset_tags (
   id UUID NOT NULL,
   valid_asset_id UUID NOT NULL,
   tag_id UUID NOT NULL,
   PRIMARY KEY (id),
   FOREIGN KEY (valid_asset_id) REFERENCES valid_assets(id) ON DELETE CASCADE,
   FOREIGN KEY (tag_id) REFERENCES tags(id) ON DELETE CASCADE
);

-- Version: 1.17
-- Description: Creates reports to table
CREATE TABLE reports_to(
   id UUID NOT NULL,
   reporter_id UUID NOT NULL,
   boss_id UUID NOT NULL,
   PRIMARY KEY (id),
   FOREIGN KEY (reporter_id) REFERENCES users(id) ON DELETE CASCADE,
   FOREIGN KEY (boss_id) REFERENCES users(id) ON DELETE CASCADE
);

-- Version: 1.18
-- Description: Add assets
CREATE TABLE assets (
   id UUID NOT NULL,
   valid_asset_id UUID NOT NULL,
   last_maintenance_time TIMESTAMP NOT NULL,
   serial_number TEXT NOT NULL,
   asset_condition_id UUID NOT NULL,
   PRIMARY KEY (id),
   FOREIGN KEY (valid_asset_id) REFERENCES valid_assets(id) ON DELETE CASCADE,
   FOREIGN KEY (asset_condition_id) REFERENCES asset_conditions(id) ON DELETE CASCADE
);

-- Version: 1.19
-- Description: Add user_assets
CREATE TABLE user_assets (
   id UUID NOT NULL,
   user_id UUID NOT NULL,
   asset_id UUID NOT NULL,
   approval_status_id UUID NOT NULL,
   last_maintenance TIMESTAMP NOT NULL,
   date_received TIMESTAMP NOT NULL,
   approved_by UUID NOT NULL,
   fulfillment_status_id UUID NOT NULL,
   PRIMARY KEY (id),
   FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
   FOREIGN KEY (approved_by) REFERENCES users(id) ON DELETE CASCADE,
   FOREIGN KEY (asset_id) REFERENCES assets(id) ON DELETE CASCADE,
   FOREIGN KEY (approval_status_id) REFERENCES approval_status(id) ON DELETE CASCADE,
   FOREIGN KEY (fulfillment_status_id) REFERENCES fulfillment_status(id) ON DELETE CASCADE
);

-- Version: 1.20
-- Description: Add user_approval_comments
CREATE TABLE user_approval_comments (
   id UUID NOT NULL,
   comment VARCHAR(255) NOT NULL,
   commenter_id UUID NOT NULL,
   user_id UUID NOT NULL,
   created_date TIMESTAMP NOT NULL,
   PRIMARY KEY (id),
   FOREIGN KEY (commenter_id) REFERENCES users(id) ON DELETE SET NULL,
   FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL
);

CREATE TYPE contact_type as ENUM ('phone', 'email', 'mail', 'fax');

-- Version: 1.21
-- Description: Add contact_infos
CREATE TABLE contact_infos (
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
CREATE TABLE brands (
   id UUID NOT NULL,
   name TEXT NOT NULL,
   contact_infos_id UUID NOT NULL,
   created_date TIMESTAMP NOT NULL,
   updated_date TIMESTAMP NOT NULL,
   PRIMARY KEY (id),
   FOREIGN KEY (contact_infos_id) REFERENCES contact_infos(id)
);

-- Version: 1.23
-- Description: add product_categoriesp
CREATE TABLE product_categories (
   id UUID NOT NULL,
   name TEXT NOT NULL,
   description text NOT NULL,
   created_date TIMESTAMP NOT NULL,
   updated_date TIMESTAMP NOT NULL,
   PRIMARY KEY (id)
);

-- Version: 1.24
-- Description: Create table warehouses
CREATE TABLE warehouses (
   id UUID NOT NULL,
   name TEXT NOT NULL,
   street_id UUID NOT NULL,
   is_active BOOLEAN NOT NULL,
   date_created TIMESTAMP NOT NULL,
   date_updated TIMESTAMP NOT NULL,
   created_by UUID NOT NULL,
   updated_by UUID NOT NULL,
   PRIMARY KEY (id),
   FOREIGN KEY (street_id) REFERENCES streets(id) ON DELETE CASCADE
);

-- =============================================================================
-- Core Permissions
-- =============================================================================

-- Version: 1.25
-- Description: Create table roles
CREATE TABLE roles (
    id UUID PRIMARY KEY,
    name VARCHAR(50) NOT NULL UNIQUE,
    description TEXT
);

-- Version: 1.26
-- Description: Create table user_roles
CREATE TABLE user_roles (
      id UUID NOT NULL,
      user_id UUID NOT NULL,
      role_id UUID NOT NULL,
      PRIMARY KEY (id),
      UNIQUE (user_id, role_id),
      FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
      FOREIGN KEY (role_id) REFERENCES roles(id) ON DELETE CASCADE
);

-- Version: 1.27
-- Description: Create table table_access
CREATE TABLE table_access (
    id UUID PRIMARY KEY,
    role_id UUID REFERENCES roles(id) ON DELETE CASCADE,
    table_name VARCHAR(50) NOT NULL,
    can_create BOOLEAN DEFAULT FALSE,
    can_read BOOLEAN DEFAULT FALSE,
    can_update BOOLEAN DEFAULT FALSE,
    can_delete BOOLEAN DEFAULT FALSE,
    UNIQUE(role_id, table_name)
);

-- Version: 1.28
-- Description: add products
CREATE TABLE products (
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
   FOREIGN KEY (brand_id) REFERENCES brands(id),
   FOREIGN KEY (category_id) REFERENCES product_categories(id)
);

-- Version: 1.29
-- Description: add physical_attributes
CREATE TABLE physical_attributes (
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
   FOREIGN KEY (product_id) REFERENCES products(id)
);

-- Version: 1.30
-- Description: add product_costs
CREATE TABLE product_costs (
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
   FOREIGN KEY (product_id) REFERENCES products(id)
);

-- Version: 1.31
-- Description: add suppliers
CREATE TABLE suppliers (
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
   FOREIGN KEY (contact_infos_id) REFERENCES contact_infos(id)
);

-- Version: 1.32
-- Description: add cost_history
CREATE TABLE cost_history (
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
   FOREIGN KEY (product_id) REFERENCES products(id)
);

-- Version: 1.33
-- Description: add supplier_products
CREATE TABLE supplier_products (
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
   FOREIGN KEY (supplier_id) REFERENCES suppliers(id),
   FOREIGN KEY (product_id) REFERENCES products(id)
);

-- Version: 1.34
-- Description: add quality_metrics
CREATE TABLE quality_metrics (
   id UUID NOT NULL,
   product_id UUID NOT NULL,
   return_rate NUMERIC(10, 4) NOT NULL,
   defect_rate NUMERIC(10, 4) NOT NULL,
   measurement_period INTERVAL NOT NULL,
   created_date TIMESTAMP NOT NULL,
   updated_date TIMESTAMP NOT NULL,
   PRIMARY KEY (id),
   FOREIGN KEY (product_id) REFERENCES products(id)
);

-- Version: 1.35
-- Description: add lot tracking
CREATE TABLE lot_trackings (
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
   FOREIGN KEY (supplier_product_id) REFERENCES supplier_products(id)
);

-- Version: 1.36
-- Description: add zones
CREATE TABLE zones (
   id UUID NOT NULL,
   warehouse_id UUID NOT NULL,
   name VARCHAR(50) NOT NULL,
   description TEXT NULL,
   created_date TIMESTAMP NOT NULL,
   updated_date TIMESTAMP NOT NULL,
   PRIMARY KEY (id),
   FOREIGN KEY (warehouse_id) REFERENCES warehouses(id)
);

-- Version: 1.37
-- Description: add inventory_locations
CREATE TABLE inventory_locations (
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
   FOREIGN KEY (zone_id) REFERENCES zones(id),
   FOREIGN KEY (warehouse_id) REFERENCES warehouses(id)
);

-- Version: 1.38
-- Description: add inventory_items
CREATE TABLE inventory_items (
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
   FOREIGN KEY (product_id) REFERENCES products(id),
   FOREIGN KEY (location_id) REFERENCES inventory_locations(id)
);

-- Version: 1.39
-- Description: add serial_numbers
CREATE TABLE serial_numbers (
   id UUID NOT NULL,
   product_id UUID NOT NULL,  
   location_id UUID NOT NULL,
   lot_id UUID NOT NULL,
   serial_number VARCHAR(100) NOT NULL,
   status VARCHAR(20) NOT NULL,
   created_date TIMESTAMP NOT NULL,
   updated_date TIMESTAMP NOT NULL,
   PRIMARY KEY (id),
   FOREIGN KEY (product_id) REFERENCES products(id),
   FOREIGN KEY (location_id) REFERENCES inventory_locations(id),
   FOREIGN KEY (lot_id) REFERENCES lot_trackings(id)
);

-- Version: 1.40
-- Description: add quality_inspections
CREATE TABLE quality_inspections (
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
   FOREIGN KEY (product_id) REFERENCES products(id),
   FOREIGN KEY (inspector_id) REFERENCES users(id),
   FOREIGN KEY (lot_id) REFERENCES lot_trackings(id)
);

-- Version: 1.41 
-- Description: add inventory_transactions
CREATE TABLE inventory_transactions (
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
   FOREIGN KEY (product_id) REFERENCES products(id),
   FOREIGN KEY (location_id) REFERENCES inventory_locations(id),
   FOREIGN KEY (user_id) REFERENCES users(id)
);

-- Version: 1.42
-- Description: add inventory_adjustments
CREATE TABLE inventory_adjustments (
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
   FOREIGN KEY (product_id) REFERENCES products(id),
   FOREIGN KEY (location_id) REFERENCES inventory_locations(id),
   FOREIGN KEY (adjusted_by) REFERENCES users(id),
   FOREIGN KEY (approved_by) REFERENCES users(id)
);

-- Version: 1.43
-- Description: transfer_orders
CREATE TABLE transfer_orders (
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
   FOREIGN KEY (product_id) REFERENCES products(id),
   FOREIGN KEY (from_location_id) REFERENCES inventory_locations(id),
   FOREIGN KEY (to_location_id) REFERENCES inventory_locations(id),
   FOREIGN KEY (requested_by) REFERENCES users(id),
   FOREIGN KEY (approved_by) REFERENCES users(id)
);


-- =============================================================================
-- ORDERS
-- =============================================================================
CREATE TABLE order_fulfillment_statuses (
   id UUID NOT NULL,
   name VARCHAR(50) NOT NULL,
   description TEXT NULL,
   PRIMARY KEY (id),
   UNIQUE (name)
);

CREATE TABLE line_item_fulfillment_statuses (
   id UUID NOT NULL,
   name VARCHAR(50) NOT NULL,
   description TEXT NULL,
   PRIMARY KEY (id),
   UNIQUE (name)
);


CREATE TABLE customers (
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
   FOREIGN KEY (contact_id) REFERENCES contact_infos(id) ON DELETE CASCADE,
   FOREIGN KEY (delivery_address_id) REFERENCES streets(id) ON DELETE CASCADE
);

CREATE TABLE orders (
   id UUID NOT NULL,
   order_number VARCHAR(50) NOT NULL,
   customer_id UUID NOT NULL,
   order_date TIMESTAMP NOT NULL,
   due_date TIMESTAMP NOT NULL,
   order_fulfillment_statuses_id UUID NOT NULL,
   created_date TIMESTAMP NOT NULL,
   updated_date TIMESTAMP NOT NULL,
   PRIMARY KEY (id),
   FOREIGN KEY (customer_id) REFERENCES customers(id),
   FOREIGN KEY (order_fulfillment_statuses_id) REFERENCES order_fulfillment_statuses(id) ON DELETE SET NULL
);

CREATE TABLE order_line_items (
   id UUID NOT NULL,
   order_id UUID NOT NULL,
   product_id UUID NOT NULL,
   quantity INT NOT NULL, 
   discount NUMERIC(10,2) NULL, -- TODO: Refactor this to be either percent or flat amount
   line_item_fulfillment_statuses_id UUID NOT NULL,
   created_date TIMESTAMP NOT NULL,
   updated_date TIMESTAMP NOT NULL,
   PRIMARY KEY (id),
   FOREIGN KEY (order_id) REFERENCES orders(id) ON DELETE CASCADE,
   FOREIGN KEY (product_id) REFERENCES products(id) ON DELETE CASCADE,
   FOREIGN KEY (line_item_fulfillment_statuses_id) REFERENCES line_item_fulfillment_statuses(id) ON DELETE SET NULL
);