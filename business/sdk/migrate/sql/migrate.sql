-- Version: 1.01
-- Description: Create table asset_types
CREATE TABLE asset_types (
   asset_type_id UUID NOT NULL,
   name TEXT NOT NULL,
   description TEXT NULL,
   PRIMARY KEY (asset_type_id),
   UNIQUE (name)
);
-- Version: 1.02
-- Description: Create table asset_conditions
CREATE TABLE asset_conditions (
   asset_condition_id UUID NOT NULL,
   name TEXT NOT NULL,
   description TEXT NULL,
   PRIMARY KEY (asset_condition_id),
   UNIQUE (name)
);
-- Version: 1.03
-- Description: Create table country
CREATE TABLE countries (
   country_id UUID NOT NULL,
   number INT NOT NULL,
   name TEXT NOT NULL,
   alpha_2 VARCHAR(2) NOT NULL,
   alpha_3 VARCHAR(3) NOT NULL,
   PRIMARY KEY (country_id)
);
-- Version: 1.04
-- Description: Create table regions
CREATE TABLE regions (
   region_id UUID NOT NULL,
   country_id UUID NOT NULL,
   name TEXT NOT NULL,
   code TEXT NOT NULL,
   -- TODO: determine if these should be enforced unique.
   PRIMARY KEY (region_id),
   FOREIGN KEY (country_id) REFERENCES countries(country_id) ON DELETE CASCADE
);
-- Version: 1.05
-- Description: create table cities
CREATE TABLE cities (
   city_id UUID NOT NULL,
   region_id UUID NOT NULL,
   name TEXT NOT NULL,
   PRIMARY KEY (city_id),
   UNIQUE (region_id, name),
   FOREIGN KEY (region_id) REFERENCES regions(region_id) ON DELETE CASCADE
);
-- Version: 1.06
-- Description: create table streets
CREATE TABLE streets (
   street_id UUID NOT NULL,
   city_id UUID NOT NULL,
   line_1 TEXT NOT NULL,
   line_2 TEXT NULL,
   postal_code VARCHAR(20) NULL,
   PRIMARY KEY (street_id),
   FOREIGN KEY (city_id) REFERENCES cities(city_id) ON DELETE SET NULL-- Check this cascade relationship
);
-- Version: 1.07
-- Description: Create table users
CREATE TABLE users (
   user_id UUID NOT NULL,
   requested_by UUID NULL,
   approved_by UUID NULL,
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
   PRIMARY KEY (user_id)
);
-- Version: 1.08
-- Description: Create table products
CREATE TABLE products (
   product_id UUID NOT NULL,
   user_id UUID NOT NULL,
   NAME TEXT NOT NULL,
   COST NUMERIC(10, 2) NOT NULL,
   quantity INT NOT NULL,
   date_created TIMESTAMP NOT NULL,
   date_updated TIMESTAMP NOT NULL,
   PRIMARY KEY (product_id),
   FOREIGN KEY (user_id) REFERENCES users (user_id) ON DELETE CASCADE
);
-- Version: 1.09
-- Description: Add products view.
CREATE OR REPLACE VIEW view_products AS
SELECT p.product_id,
   p.user_id,
   p.name,
   p.cost,
   p.quantity,
   p.date_created,
   p.date_updated,
   u.username AS user_name
FROM products AS p
   JOIN users AS u ON u.user_id = p.user_id;
-- Version: 1.10
-- Description: Create table homes
CREATE TABLE homes (
   home_id UUID NOT NULL,
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
   PRIMARY KEY (home_id),
   FOREIGN KEY (user_id) REFERENCES users (user_id) ON DELETE CASCADE
);
-- Version: 1.09
-- Description: Add approval status 
CREATE TABLE approval_status (
   approval_status_id UUID NOT NULL, 
   icon_id UUID NOT NULL, 
   name TEXT NOT NULL,
   PRIMARY KEY (approval_status_id)
);
-- Version: 1.10
-- Description: Add fulfillment status
CREATE TABLE fulfillment_status (
   fulfillment_status_id UUID NOT NULL, 
   icon_id UUID NOT NULL, 
   name TEXT NOT NULL,
   PRIMARY KEY (fulfillment_status_id)
);
-- Version: 1.11
-- Description: add asset condition
CREATE TABLE asset_condition (
   asset_condition_id UUID NOT NULL, 
   name TEXT NOT NULL,
   PRIMARY KEY (asset_condition_id)
);