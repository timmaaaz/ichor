BEGIN;
--
-- PostgreSQL database dump
--

\restrict tg2r7S5elF0h8E73AeZATmLFxKFMvO7cjaPfGueWWAkxph8IwOT7bR4s1yZOOVD

-- Dumped from database version 16.4 (Debian 16.4-1.pgdg120+2)
-- Dumped by pg_dump version 17.6

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET transaction_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;

--
-- Data for Name: table_access; Type: TABLE DATA; Schema: core; Owner: -
--

INSERT INTO core.table_access VALUES ('35753130-0c1d-45ea-9beb-0967c840d5d3', '54bb2165-71e1-41a6-af3e-7da4a0e1e2c1', 'assets.approval_status', true, true, true, true);
INSERT INTO core.table_access VALUES ('3df16a42-b875-4dc8-87b7-e2b7a6c7a140', '54bb2165-71e1-41a6-af3e-7da4a0e1e2c1', 'assets.asset_conditions', true, true, true, true);
INSERT INTO core.table_access VALUES ('f4adda3b-e704-4594-873a-ef2f18a43012', '54bb2165-71e1-41a6-af3e-7da4a0e1e2c1', 'assets.asset_tags', true, true, true, true);
INSERT INTO core.table_access VALUES ('1f732b90-4b29-43cb-aee6-ebdf6c65eda1', '54bb2165-71e1-41a6-af3e-7da4a0e1e2c1', 'assets.asset_types', true, true, true, true);
INSERT INTO core.table_access VALUES ('002c7fcd-2cd5-42b6-b29b-00a5ae42b372', '54bb2165-71e1-41a6-af3e-7da4a0e1e2c1', 'assets.assets', true, true, true, true);
INSERT INTO core.table_access VALUES ('d790318c-d758-4c12-a3a4-147ff3eec38d', '54bb2165-71e1-41a6-af3e-7da4a0e1e2c1', 'assets.fulfillment_status', true, true, true, true);
INSERT INTO core.table_access VALUES ('21f7f7e4-ba9c-41b6-86e9-0ec347f46651', '54bb2165-71e1-41a6-af3e-7da4a0e1e2c1', 'assets.tags', true, true, true, true);
INSERT INTO core.table_access VALUES ('e37aa67a-eebe-473e-91c6-acb1c431e6c3', '54bb2165-71e1-41a6-af3e-7da4a0e1e2c1', 'assets.user_assets', true, true, true, true);
INSERT INTO core.table_access VALUES ('f483556c-42b1-4079-9133-cb4c644b5e6f', '54bb2165-71e1-41a6-af3e-7da4a0e1e2c1', 'assets.valid_assets', true, true, true, true);
INSERT INTO core.table_access VALUES ('b57a5d99-2254-4c29-a642-216a44b6c3f9', '54bb2165-71e1-41a6-af3e-7da4a0e1e2c1', 'config.form_fields', true, true, true, true);
INSERT INTO core.table_access VALUES ('243cc834-eb58-424d-8b3a-94ca4de2d984', '54bb2165-71e1-41a6-af3e-7da4a0e1e2c1', 'formdata', true, true, true, true);
INSERT INTO core.table_access VALUES ('293375c2-0ad9-40bc-9161-c91ff5b48f5b', '54bb2165-71e1-41a6-af3e-7da4a0e1e2c1', 'config.forms', true, true, true, true);
INSERT INTO core.table_access VALUES ('bbd13510-57d2-4445-b84e-ba5e943a6f3d', '54bb2165-71e1-41a6-af3e-7da4a0e1e2c1', 'config.page_action_buttons', true, true, true, true);
INSERT INTO core.table_access VALUES ('2c3df62a-ba04-4d9a-86ae-03d7dcb149e4', '54bb2165-71e1-41a6-af3e-7da4a0e1e2c1', 'config.page_action_dropdown_items', true, true, true, true);
INSERT INTO core.table_access VALUES ('e5f3b60a-80ec-444f-aa50-5009125da29b', '54bb2165-71e1-41a6-af3e-7da4a0e1e2c1', 'config.page_action_dropdowns', true, true, true, true);
INSERT INTO core.table_access VALUES ('f566d310-666f-4677-bbb0-d67abb95bc24', '54bb2165-71e1-41a6-af3e-7da4a0e1e2c1', 'config.page_actions', true, true, true, true);
INSERT INTO core.table_access VALUES ('1700b5f9-a6f4-4e26-ab07-c20487044c73', '54bb2165-71e1-41a6-af3e-7da4a0e1e2c1', 'config.page_configs', true, true, true, true);
INSERT INTO core.table_access VALUES ('0d91c394-f34c-4f2d-aabc-1d803abe6ecc', '54bb2165-71e1-41a6-af3e-7da4a0e1e2c1', 'config.page_content', true, true, true, true);
INSERT INTO core.table_access VALUES ('0faaa628-a6ad-47ce-a7a7-71d9ada9cf3c', '54bb2165-71e1-41a6-af3e-7da4a0e1e2c1', 'config.table_configs', true, true, true, true);
INSERT INTO core.table_access VALUES ('3aab6db6-51f0-44b4-a6f2-3eb22823fb13', '54bb2165-71e1-41a6-af3e-7da4a0e1e2c1', 'core.contact_infos', true, true, true, true);
INSERT INTO core.table_access VALUES ('509c4501-163c-463d-bf01-e0223a32980c', '54bb2165-71e1-41a6-af3e-7da4a0e1e2c1', 'core.currencies', true, true, true, true);
INSERT INTO core.table_access VALUES ('b1db7f71-d4c7-471d-b8ad-1e98faef0fdf', '54bb2165-71e1-41a6-af3e-7da4a0e1e2c1', 'core.pages', true, true, true, true);
INSERT INTO core.table_access VALUES ('77488ab4-17db-48f3-b7c4-684aa5ec9741', '54bb2165-71e1-41a6-af3e-7da4a0e1e2c1', 'core.payment_terms', true, true, true, true);
INSERT INTO core.table_access VALUES ('52d634d9-f781-49f3-bc39-c54b3db03999', '54bb2165-71e1-41a6-af3e-7da4a0e1e2c1', 'core.role_pages', true, true, true, true);
INSERT INTO core.table_access VALUES ('17bf9d56-43a3-4693-970f-9414ef86b1de', '54bb2165-71e1-41a6-af3e-7da4a0e1e2c1', 'core.roles', true, true, true, true);
INSERT INTO core.table_access VALUES ('c4413527-bb3d-4092-b094-c2a114edd03a', '54bb2165-71e1-41a6-af3e-7da4a0e1e2c1', 'core.table_access', true, true, true, true);
INSERT INTO core.table_access VALUES ('5d830b2c-0d88-4a62-8e58-1c27eff38898', '54bb2165-71e1-41a6-af3e-7da4a0e1e2c1', 'core.user_roles', true, true, true, true);
INSERT INTO core.table_access VALUES ('675dab8a-f9c7-4058-be87-4dd39d923ff4', '54bb2165-71e1-41a6-af3e-7da4a0e1e2c1', 'core.users', true, true, true, true);
INSERT INTO core.table_access VALUES ('57404c10-b283-4bbd-9298-8b360862b9d3', '54bb2165-71e1-41a6-af3e-7da4a0e1e2c1', 'geography.cities', true, true, true, true);
INSERT INTO core.table_access VALUES ('6e30a879-d447-4ad0-9ccd-be4cce95b948', '54bb2165-71e1-41a6-af3e-7da4a0e1e2c1', 'geography.countries', true, true, true, true);
INSERT INTO core.table_access VALUES ('b9f304ef-d199-4dd0-aaa9-47e56c6854e6', '54bb2165-71e1-41a6-af3e-7da4a0e1e2c1', 'geography.regions', true, true, true, true);
INSERT INTO core.table_access VALUES ('4e3b5339-ee3c-4faa-9bf1-7e0a2fa99380', '54bb2165-71e1-41a6-af3e-7da4a0e1e2c1', 'geography.streets', true, true, true, true);
INSERT INTO core.table_access VALUES ('8a66e096-71db-4a1c-af39-2a343cc2b85f', '54bb2165-71e1-41a6-af3e-7da4a0e1e2c1', 'geography.timezones', true, true, true, true);
INSERT INTO core.table_access VALUES ('85c863ef-d34e-475f-b8f8-13f06b99694b', '54bb2165-71e1-41a6-af3e-7da4a0e1e2c1', 'hr.homes', true, true, true, true);
INSERT INTO core.table_access VALUES ('a8a26c8a-1dec-4ffb-b9be-b6bd9d36bac6', '54bb2165-71e1-41a6-af3e-7da4a0e1e2c1', 'hr.offices', true, true, true, true);
INSERT INTO core.table_access VALUES ('38a64ab8-06f1-4f53-bcf3-16511ce014b1', '54bb2165-71e1-41a6-af3e-7da4a0e1e2c1', 'hr.reports_to', true, true, true, true);
INSERT INTO core.table_access VALUES ('574027d5-b9b8-4445-89bc-8cc1b458302e', '54bb2165-71e1-41a6-af3e-7da4a0e1e2c1', 'hr.titles', true, true, true, true);
INSERT INTO core.table_access VALUES ('6c10ae1a-4eb3-4b4c-b57b-f9c0a5dc8589', '54bb2165-71e1-41a6-af3e-7da4a0e1e2c1', 'hr.user_approval_comments', true, true, true, true);
INSERT INTO core.table_access VALUES ('741356dc-65ca-4e43-a125-10a746574d6d', '54bb2165-71e1-41a6-af3e-7da4a0e1e2c1', 'hr.user_approval_status', true, true, true, true);
INSERT INTO core.table_access VALUES ('3608c85f-5403-4daf-b3d1-27fd42d7638c', '54bb2165-71e1-41a6-af3e-7da4a0e1e2c1', 'introspection', true, true, true, true);
INSERT INTO core.table_access VALUES ('645569b4-5d0a-4c82-9082-e623d8c9ff8c', '54bb2165-71e1-41a6-af3e-7da4a0e1e2c1', 'inventory.inventory_adjustments', true, true, true, true);
INSERT INTO core.table_access VALUES ('236879a2-4ab4-4715-9438-571a7c5797cc', '54bb2165-71e1-41a6-af3e-7da4a0e1e2c1', 'inventory.inventory_items', true, true, true, true);
INSERT INTO core.table_access VALUES ('e00cd0a7-d8bb-4179-8b00-5626cec61fd3', '54bb2165-71e1-41a6-af3e-7da4a0e1e2c1', 'inventory.inventory_locations', true, true, true, true);
INSERT INTO core.table_access VALUES ('bceb4caf-fead-4798-969e-556de2aa2b48', '54bb2165-71e1-41a6-af3e-7da4a0e1e2c1', 'inventory.inventory_transactions', true, true, true, true);
INSERT INTO core.table_access VALUES ('fc002780-d71f-4444-83f0-30b8ae3d22af', '54bb2165-71e1-41a6-af3e-7da4a0e1e2c1', 'inventory.lot_trackings', true, true, true, true);
INSERT INTO core.table_access VALUES ('51aac7a3-0d2a-4426-92e7-81c2862246c2', '54bb2165-71e1-41a6-af3e-7da4a0e1e2c1', 'inventory.lot_locations', true, true, true, true);
INSERT INTO core.table_access VALUES ('a2a87c34-cdff-451b-a10e-6ba921b3bc02', '54bb2165-71e1-41a6-af3e-7da4a0e1e2c1', 'inventory.quality_inspections', true, true, true, true);
INSERT INTO core.table_access VALUES ('286fb78f-88b2-43ae-9395-fb69e0b7583e', '54bb2165-71e1-41a6-af3e-7da4a0e1e2c1', 'inventory.put_away_tasks', true, true, true, true);
INSERT INTO core.table_access VALUES ('4a4f44ca-d165-429c-bdfe-d7569f0083c2', '54bb2165-71e1-41a6-af3e-7da4a0e1e2c1', 'inventory.serial_numbers', true, true, true, true);
INSERT INTO core.table_access VALUES ('3a97932c-9c9d-490c-b212-923eecb9c097', '54bb2165-71e1-41a6-af3e-7da4a0e1e2c1', 'inventory.transfer_orders', true, true, true, true);
INSERT INTO core.table_access VALUES ('f035354b-3793-4456-ad94-74131166c0c5', '54bb2165-71e1-41a6-af3e-7da4a0e1e2c1', 'inventory.warehouses', true, true, true, true);
INSERT INTO core.table_access VALUES ('c88aa11f-b5ad-4f61-8acf-e7a6d60fe802', '54bb2165-71e1-41a6-af3e-7da4a0e1e2c1', 'inventory.zones', true, true, true, true);
INSERT INTO core.table_access VALUES ('0899eb2e-6179-4790-baf4-3d1cba052d18', '54bb2165-71e1-41a6-af3e-7da4a0e1e2c1', 'procurement.purchase_order_line_item_statuses', true, true, true, true);
INSERT INTO core.table_access VALUES ('c2592f47-1fa4-44ba-a59a-f8a94ac3cfc3', '54bb2165-71e1-41a6-af3e-7da4a0e1e2c1', 'procurement.purchase_order_line_items', true, true, true, true);
INSERT INTO core.table_access VALUES ('2f1aa63a-16a7-4c26-be40-30a05449a6b0', '54bb2165-71e1-41a6-af3e-7da4a0e1e2c1', 'procurement.purchase_order_statuses', true, true, true, true);
INSERT INTO core.table_access VALUES ('def00305-b209-43b1-bf68-708a4aa43f49', '54bb2165-71e1-41a6-af3e-7da4a0e1e2c1', 'procurement.purchase_orders', true, true, true, true);
INSERT INTO core.table_access VALUES ('c2530a98-de3b-4c3e-bce2-5feeef613a2e', '54bb2165-71e1-41a6-af3e-7da4a0e1e2c1', 'procurement.supplier_products', true, true, true, true);
INSERT INTO core.table_access VALUES ('701df4b0-a021-41b0-b52a-f3c6101d6ee0', '54bb2165-71e1-41a6-af3e-7da4a0e1e2c1', 'procurement.suppliers', true, true, true, true);
INSERT INTO core.table_access VALUES ('8b433ff9-99b3-47cb-bd83-b49b732341d5', '54bb2165-71e1-41a6-af3e-7da4a0e1e2c1', 'products.brands', true, true, true, true);
INSERT INTO core.table_access VALUES ('55ed6e28-0497-43a4-a9a5-af2102c24d29', '54bb2165-71e1-41a6-af3e-7da4a0e1e2c1', 'products.cost_history', true, true, true, true);
INSERT INTO core.table_access VALUES ('492610b2-ff91-4e15-8868-2f7115240b49', '54bb2165-71e1-41a6-af3e-7da4a0e1e2c1', 'products.physical_attributes', true, true, true, true);
INSERT INTO core.table_access VALUES ('a4e479cc-ac61-41b7-894d-491a1ff2fc36', '54bb2165-71e1-41a6-af3e-7da4a0e1e2c1', 'products.product_categories', true, true, true, true);
INSERT INTO core.table_access VALUES ('9ec0a1c0-1fbe-4c24-89f7-db88bd288057', '54bb2165-71e1-41a6-af3e-7da4a0e1e2c1', 'products.product_costs', true, true, true, true);
INSERT INTO core.table_access VALUES ('fba8be77-db26-483e-a363-e8e56da391c3', '54bb2165-71e1-41a6-af3e-7da4a0e1e2c1', 'products.products', true, true, true, true);
INSERT INTO core.table_access VALUES ('59aeb976-24b1-429e-9254-c7b7badeae69', '54bb2165-71e1-41a6-af3e-7da4a0e1e2c1', 'products.quality_metrics', true, true, true, true);
INSERT INTO core.table_access VALUES ('408b8bf8-1531-495b-bcdb-8d34c5994614', '54bb2165-71e1-41a6-af3e-7da4a0e1e2c1', 'sales.customers', true, true, true, true);
INSERT INTO core.table_access VALUES ('d077bd45-e37b-414f-9932-3b4b8a63f7da', '54bb2165-71e1-41a6-af3e-7da4a0e1e2c1', 'sales.line_item_fulfillment_statuses', true, true, true, true);
INSERT INTO core.table_access VALUES ('696e3f26-46a1-4691-845e-ab9708f720ac', '54bb2165-71e1-41a6-af3e-7da4a0e1e2c1', 'sales.order_fulfillment_statuses', true, true, true, true);
INSERT INTO core.table_access VALUES ('1a31834e-34c0-4f1a-a062-88060c439972', '54bb2165-71e1-41a6-af3e-7da4a0e1e2c1', 'sales.order_line_items', true, true, true, true);
INSERT INTO core.table_access VALUES ('66219ba9-e722-4b71-836e-1756b20916d7', '54bb2165-71e1-41a6-af3e-7da4a0e1e2c1', 'sales.orders', true, true, true, true);
INSERT INTO core.table_access VALUES ('e50d80d7-c764-4ce3-9e78-57ef2a124d44', '54bb2165-71e1-41a6-af3e-7da4a0e1e2c1', 'workflow.action_edges', true, true, true, true);
INSERT INTO core.table_access VALUES ('baeca2bd-ce2c-46d2-a1b3-384cd213ba8d', '54bb2165-71e1-41a6-af3e-7da4a0e1e2c1', 'workflow.action_permissions', true, true, true, true);
INSERT INTO core.table_access VALUES ('638a1519-e098-4e38-a685-1cb8220e3412', '54bb2165-71e1-41a6-af3e-7da4a0e1e2c1', 'workflow.action_templates', true, true, true, true);
INSERT INTO core.table_access VALUES ('16451224-05ef-4bbd-a8fb-c0febac47bc4', '54bb2165-71e1-41a6-af3e-7da4a0e1e2c1', 'workflow.alert_acknowledgments', true, true, true, true);
INSERT INTO core.table_access VALUES ('ffe7e1c2-f1f0-4c61-b356-d70f1574797d', '54bb2165-71e1-41a6-af3e-7da4a0e1e2c1', 'workflow.alert_recipients', true, true, true, true);
INSERT INTO core.table_access VALUES ('c002ea02-661e-40a5-b387-6fe7b1a38014', '54bb2165-71e1-41a6-af3e-7da4a0e1e2c1', 'workflow.alerts', true, true, true, true);
INSERT INTO core.table_access VALUES ('0e15d4c1-5f13-4bad-89c0-7b5407aee7c9', '54bb2165-71e1-41a6-af3e-7da4a0e1e2c1', 'workflow.allocation_results', true, true, true, true);
INSERT INTO core.table_access VALUES ('d1f4a26d-ee64-47c5-b9c4-8cfd6766b133', '54bb2165-71e1-41a6-af3e-7da4a0e1e2c1', 'workflow.automation_executions', true, true, true, true);
INSERT INTO core.table_access VALUES ('eaddc6aa-51b7-4f4e-939d-8c0b61e7344b', '54bb2165-71e1-41a6-af3e-7da4a0e1e2c1', 'workflow.automation_rules', true, true, true, true);
INSERT INTO core.table_access VALUES ('6d3a0834-e0b5-48f4-9f9e-c82589517b44', '54bb2165-71e1-41a6-af3e-7da4a0e1e2c1', 'workflow.entities', true, true, true, true);
INSERT INTO core.table_access VALUES ('d3cf22c0-6a6f-49c8-b5af-7c4aef2acea0', '54bb2165-71e1-41a6-af3e-7da4a0e1e2c1', 'workflow.entity_types', true, true, true, true);
INSERT INTO core.table_access VALUES ('686ed1e1-713c-4066-8e80-3a2739197b6d', '54bb2165-71e1-41a6-af3e-7da4a0e1e2c1', 'workflow.notification_deliveries', true, true, true, true);
INSERT INTO core.table_access VALUES ('2688051b-911e-4820-b0be-004250a70004', '54bb2165-71e1-41a6-af3e-7da4a0e1e2c1', 'workflow.approval_requests', true, true, true, true);
INSERT INTO core.table_access VALUES ('55ed3ded-abf6-40b0-b491-ff5dc44d6291', '54bb2165-71e1-41a6-af3e-7da4a0e1e2c1', 'workflow.rule_actions', true, true, true, true);
INSERT INTO core.table_access VALUES ('fc5a1fcc-3e05-422c-b90f-56711b6a202d', '54bb2165-71e1-41a6-af3e-7da4a0e1e2c1', 'workflow.rule_dependencies', true, true, true, true);
INSERT INTO core.table_access VALUES ('e1aa7615-3e14-4fa3-989d-36488e39f1a7', '54bb2165-71e1-41a6-af3e-7da4a0e1e2c1', 'workflow.trigger_types', true, true, true, true);
INSERT INTO core.table_access VALUES ('239acee1-5c84-490c-a443-7360ef00e910', '54bb2165-71e1-41a6-af3e-7da4a0e1e2c1', 'workflow.user_approval_comments', true, true, true, true);
INSERT INTO core.table_access VALUES ('98b4b082-c19e-47ac-9639-fd16920a800a', '54bb2165-71e1-41a6-af3e-7da4a0e1e2c1', 'workflow.user_approval_status', true, true, true, true);
INSERT INTO core.table_access VALUES ('c88866ff-e210-4815-ae2e-44beee4d86a4', '54bb2165-71e1-41a6-af3e-7da4a0e1e2c1', 'workflow.audit_log', true, true, true, true);


--
-- PostgreSQL database dump complete
--

\unrestrict tg2r7S5elF0h8E73AeZATmLFxKFMvO7cjaPfGueWWAkxph8IwOT7bR4s1yZOOVD

COMMIT;
