BEGIN;
pg_dump: warning: there are circular foreign-key constraints on this table:
pg_dump: detail: page_content
pg_dump: hint: You might not be able to restore the dump without using --disable-triggers or temporarily dropping the constraints.
pg_dump: hint: Consider using a full dump instead of a --data-only dump to avoid this problem.
--
-- PostgreSQL database dump
--


-- Dumped from database version 16.4 (Debian 16.4-1.pgdg120+2)
-- Dumped by pg_dump version 17.6

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;

--
-- Data for Name: page_configs; Type: TABLE DATA; Schema: config; Owner: -
--

INSERT INTO config.page_configs VALUES ('193aabb2-e3b0-474d-b86f-8409415d0b11', 'orders_page', NULL, true);
INSERT INTO config.page_configs VALUES ('2d918bfe-e839-4b6a-9c12-890f0ec37d0d', 'suppliers_page', NULL, true);
INSERT INTO config.page_configs VALUES ('ecf3d7c9-d0d1-40ec-b44b-5373c187b9ba', 'categories_page', NULL, true);
INSERT INTO config.page_configs VALUES ('6698349e-a4d6-4dd0-b28c-d3041223d2a6', 'order_line_items_page', NULL, true);
INSERT INTO config.page_configs VALUES ('b716ebfd-3c22-4294-81f3-4ee3e68f7b58', 'admin_users_page', NULL, true);
INSERT INTO config.page_configs VALUES ('592e7016-6e68-413b-a354-17ec6c38da83', 'admin_roles_page', NULL, true);
INSERT INTO config.page_configs VALUES ('9dc01ff4-5610-4392-993a-a9cc8c1ce5e7', 'admin_dashboard_page', NULL, true);
INSERT INTO config.page_configs VALUES ('339bc185-9045-4807-b463-de065529a152', 'assets_list_page', NULL, true);
INSERT INTO config.page_configs VALUES ('f0e94758-fb58-446b-9088-139c891d6e79', 'assets_requests_page', NULL, true);
INSERT INTO config.page_configs VALUES ('6850f0d9-ec28-434b-b2fe-16dfe96e2f0e', 'assets_dashboard_page', NULL, true);
INSERT INTO config.page_configs VALUES ('631d9418-04de-4c4f-8eae-3269f114c60b', 'hr_employees_page', NULL, true);
INSERT INTO config.page_configs VALUES ('570f2a2b-7898-4231-8931-5174f02f45b0', 'hr_offices_page', NULL, true);
INSERT INTO config.page_configs VALUES ('d25c649e-acf3-4629-9bf2-18c990c462d0', 'hr_dashboard_page', NULL, true);
INSERT INTO config.page_configs VALUES ('e5a036fb-5e93-437c-a41e-b184db8cbaee', 'inventory_warehouses_page', NULL, true);
INSERT INTO config.page_configs VALUES ('c0590f21-21f2-41b8-8a45-1293f0c194a4', 'inventory_items_page', NULL, true);
INSERT INTO config.page_configs VALUES ('aca655fd-1b0d-40b0-b02c-be56ebb7916e', 'inventory_adjustments_page', NULL, true);
INSERT INTO config.page_configs VALUES ('2f7b7513-8a92-4922-acfa-41cc94e8a134', 'inventory_transfers_page', NULL, true);
INSERT INTO config.page_configs VALUES ('72747ce9-f63a-43e3-b2d0-51a3530ac527', 'inventory_dashboard_page', NULL, true);
INSERT INTO config.page_configs VALUES ('e96c1df7-6c5c-463e-b604-da53cf8c9785', 'inventory_zones_page', NULL, true);
INSERT INTO config.page_configs VALUES ('8691a22f-7e0e-49d6-8143-4914d06ed6c4', 'inventory_locations_page', NULL, true);
INSERT INTO config.page_configs VALUES ('722d5db8-c3e8-4d35-9702-7d35e3dd7416', 'products_page', NULL, true);
INSERT INTO config.page_configs VALUES ('1b20927b-8682-4525-b18f-f0e7d1e1d1de', 'sales_customers_page', NULL, true);
INSERT INTO config.page_configs VALUES ('d25ec547-eb68-4847-8864-cf0d6532b913', 'sales_dashboard_page', NULL, true);
INSERT INTO config.page_configs VALUES ('47739fd5-ee02-4180-af5a-8db38bb41d5e', 'procurement_purchase_orders', NULL, true);
INSERT INTO config.page_configs VALUES ('e570d44d-6cc7-44d0-a146-9294598832cb', 'procurement_line_items', NULL, true);
INSERT INTO config.page_configs VALUES ('3d224c0d-b1d6-4d3b-bde1-8455e390c99e', 'procurement_approvals', NULL, true);
INSERT INTO config.page_configs VALUES ('4852ffae-1b5c-4ed3-a941-a88970f349de', 'procurement_dashboard', NULL, true);
INSERT INTO config.page_configs VALUES ('d1556062-a25d-428c-8841-e0c3509ab189', 'main_dashboard_page', NULL, true);
INSERT INTO config.page_configs VALUES ('d21c14ee-f005-4b13-88c3-b4e38be0bb37', 'user_management_example', NULL, true);
INSERT INTO config.page_configs VALUES ('e4d6667c-8749-4cfd-ad50-2e0d450eb379', 'sample_charts_dashboard', NULL, true);


--
-- Data for Name: page_content; Type: TABLE DATA; Schema: config; Owner: -
--

INSERT INTO config.page_content VALUES ('4f00a2aa-c022-4dfa-a006-8c4dfbdb04dd', '193aabb2-e3b0-474d-b86f-8409415d0b11', 'chart', 'Monthly Sales Trend', NULL, NULL, 1, NULL, '{"colSpan": {"lg": 8, "md": 8, "sm": 8, "xl": 8, "xs": 8}}', true, false, '2ea0c998-f4b4-465f-ae4d-65975baed934');
INSERT INTO config.page_content VALUES ('322c7a49-53c3-496c-a346-42fc97357c85', '193aabb2-e3b0-474d-b86f-8409415d0b11', 'chart', 'Sales Pipeline', NULL, NULL, 2, NULL, '{"colSpan": {"lg": 4, "md": 4, "sm": 4, "xl": 4, "xs": 4}}', true, false, 'd56ed628-87dd-43dc-8539-95197968336f');
INSERT INTO config.page_content VALUES ('fdb870a1-0465-4bea-8929-b7487ca50a6d', '193aabb2-e3b0-474d-b86f-8409415d0b11', 'table', '', 'b3696a06-7755-453f-948a-92792c233583', NULL, 3, NULL, '{}', true, true, NULL);
INSERT INTO config.page_content VALUES ('d1738b4e-54b5-40d3-bf45-3440c34f23b5', '2d918bfe-e839-4b6a-9c12-890f0ec37d0d', 'table', '', 'aa8b6e2b-6d58-4c57-ac91-59eb02cab636', NULL, 1, NULL, '{}', true, true, NULL);
INSERT INTO config.page_content VALUES ('ae9e8634-0170-46aa-9f7d-dfdc8b93a527', 'ecf3d7c9-d0d1-40ec-b44b-5373c187b9ba', 'chart', 'Revenue by Category', NULL, NULL, 1, NULL, '{"colSpan": {"md": 4, "xs": 12}}', true, false, '6f44562b-b955-42aa-bc1d-f4b30a05f45c');
INSERT INTO config.page_content VALUES ('0a7f5884-ccf7-4852-911a-992dfc05fef5', 'ecf3d7c9-d0d1-40ec-b44b-5373c187b9ba', 'chart', 'Top Products', NULL, NULL, 2, NULL, '{"colSpan": {"md": 4, "xs": 12}}', true, false, '3e069ce8-4014-4f9b-8ba4-cc816d8f8cda');
INSERT INTO config.page_content VALUES ('baa7f52c-2dfb-4b77-8ccf-bf8f133974e3', 'ecf3d7c9-d0d1-40ec-b44b-5373c187b9ba', 'chart', 'Revenue Breakdown', NULL, NULL, 3, NULL, '{"colSpan": {"md": 4, "xs": 12}}', true, false, '9b54798d-6485-4b1e-a4d3-8c271fcde95a');
INSERT INTO config.page_content VALUES ('7900042c-c67c-4d49-826b-70bb1631415f', 'ecf3d7c9-d0d1-40ec-b44b-5373c187b9ba', 'table', '', 'c94df673-e97d-4999-9bf8-8f916df69375', NULL, 4, NULL, '{}', true, true, NULL);
INSERT INTO config.page_content VALUES ('248edad9-e895-4b2b-91b4-9bf33afd2a5f', '6698349e-a4d6-4dd0-b28c-d3041223d2a6', 'table', '', '78ca37b1-3081-4eac-a35a-29c02c6eac05', NULL, 1, NULL, '{}', true, true, NULL);
INSERT INTO config.page_content VALUES ('39a1c206-1cce-4ae4-a217-c6c5829d0842', 'b716ebfd-3c22-4294-81f3-4ee3e68f7b58', 'table', '', '2b7deab8-b1f8-4b4c-8cf1-94a08f408bbb', NULL, 1, NULL, '{}', true, true, NULL);
INSERT INTO config.page_content VALUES ('e0e6f1d2-d18f-44cc-8c7c-14eb5c419827', '592e7016-6e68-413b-a354-17ec6c38da83', 'table', '', 'e739a9ce-df34-4ea7-b639-18edc9386153', NULL, 1, NULL, '{}', true, true, NULL);
INSERT INTO config.page_content VALUES ('cc7fd0bf-f52d-4324-829d-e87b2dd63b8b', '9dc01ff4-5610-4392-993a-a9cc8c1ce5e7', 'tabs', '', NULL, NULL, 1, NULL, '{"containerType": "tabs"}', true, false, NULL);
INSERT INTO config.page_content VALUES ('340ee0ee-0c2a-4e04-8c78-cc5b166c328e', '9dc01ff4-5610-4392-993a-a9cc8c1ce5e7', 'table', 'Users', '2b7deab8-b1f8-4b4c-8cf1-94a08f408bbb', NULL, 1, 'cc7fd0bf-f52d-4324-829d-e87b2dd63b8b', '{}', true, true, NULL);
INSERT INTO config.page_content VALUES ('1aa4bddb-7fdb-410c-bef1-3644728a1374', '9dc01ff4-5610-4392-993a-a9cc8c1ce5e7', 'table', 'Roles', 'e739a9ce-df34-4ea7-b639-18edc9386153', NULL, 2, 'cc7fd0bf-f52d-4324-829d-e87b2dd63b8b', '{}', true, false, NULL);
INSERT INTO config.page_content VALUES ('92ee2df2-bf70-4c87-ae77-acadd0c51aae', '9dc01ff4-5610-4392-993a-a9cc8c1ce5e7', 'table', 'Permissions', '5546207b-d00c-4382-8366-8069939a4734', NULL, 3, 'cc7fd0bf-f52d-4324-829d-e87b2dd63b8b', '{}', true, false, NULL);
INSERT INTO config.page_content VALUES ('70251213-67e0-4019-9327-3e48aa714d98', '9dc01ff4-5610-4392-993a-a9cc8c1ce5e7', 'table', 'Audit Logs', '4a550579-49e3-4665-aeb3-10a190a08b56', NULL, 4, 'cc7fd0bf-f52d-4324-829d-e87b2dd63b8b', '{}', true, false, NULL);
INSERT INTO config.page_content VALUES ('7be69cb7-c304-4b08-acd7-ae94fe262e89', '9dc01ff4-5610-4392-993a-a9cc8c1ce5e7', 'table', 'Configurations', '3a74548a-2188-4b21-b48f-69fe92e1ebbc', NULL, 5, 'cc7fd0bf-f52d-4324-829d-e87b2dd63b8b', '{}', true, false, NULL);
INSERT INTO config.page_content VALUES ('276d4bb0-9915-4200-9db0-a549213f58c4', '339bc185-9045-4807-b463-de065529a152', 'table', '', '3c376764-89c1-4eda-80c3-46e6b2f36388', NULL, 1, NULL, '{}', true, true, NULL);
INSERT INTO config.page_content VALUES ('caca36ec-c5c2-4f11-b333-a42ea540d447', 'f0e94758-fb58-446b-9088-139c891d6e79', 'table', '', 'bd948494-96e4-4891-8d45-e10f77dd00da', NULL, 1, NULL, '{}', true, true, NULL);
INSERT INTO config.page_content VALUES ('4eb5be75-95db-40f0-b528-7d5053169f7f', '6850f0d9-ec28-434b-b2fe-16dfe96e2f0e', 'tabs', '', NULL, NULL, 1, NULL, '{"containerType": "tabs"}', true, false, NULL);
INSERT INTO config.page_content VALUES ('29baf935-2105-4723-882e-a1e1f5ea6148', '6850f0d9-ec28-434b-b2fe-16dfe96e2f0e', 'table', 'Assets', '3c376764-89c1-4eda-80c3-46e6b2f36388', NULL, 1, '4eb5be75-95db-40f0-b528-7d5053169f7f', '{}', true, true, NULL);
INSERT INTO config.page_content VALUES ('bdd69c37-c636-4cdd-bd47-d222ce0af3a2', '6850f0d9-ec28-434b-b2fe-16dfe96e2f0e', 'table', 'Requests', 'bd948494-96e4-4891-8d45-e10f77dd00da', NULL, 2, '4eb5be75-95db-40f0-b528-7d5053169f7f', '{}', true, false, NULL);
INSERT INTO config.page_content VALUES ('c2f399e5-4412-46b5-97d2-e6cb2041a8b2', '631d9418-04de-4c4f-8eae-3269f114c60b', 'table', '', '67dd2fd0-13d2-4755-a650-152df92ec302', NULL, 1, NULL, '{}', true, true, NULL);
INSERT INTO config.page_content VALUES ('eaf15318-46bb-4e4b-a072-5c20a56bd903', '570f2a2b-7898-4231-8931-5174f02f45b0', 'table', '', '536128bd-caec-468a-853d-9ac5f6330f44', NULL, 1, NULL, '{}', true, true, NULL);
INSERT INTO config.page_content VALUES ('4efaf007-cfc2-4ef6-83a8-de18dc2839f0', 'd25c649e-acf3-4629-9bf2-18c990c462d0', 'tabs', '', NULL, NULL, 1, NULL, '{"containerType": "tabs"}', true, false, NULL);
INSERT INTO config.page_content VALUES ('329fb0e5-5031-470e-b5d6-5bcb69c163be', 'd25c649e-acf3-4629-9bf2-18c990c462d0', 'table', 'Employees', '67dd2fd0-13d2-4755-a650-152df92ec302', NULL, 1, '4efaf007-cfc2-4ef6-83a8-de18dc2839f0', '{}', true, true, NULL);
INSERT INTO config.page_content VALUES ('cd7f0009-5b16-4080-9678-f67b06781fe0', 'd25c649e-acf3-4629-9bf2-18c990c462d0', 'table', 'Offices', '536128bd-caec-468a-853d-9ac5f6330f44', NULL, 2, '4efaf007-cfc2-4ef6-83a8-de18dc2839f0', '{}', true, false, NULL);
INSERT INTO config.page_content VALUES ('5822115e-3e76-459b-ac7b-139b6b845b4e', 'e5a036fb-5e93-437c-a41e-b184db8cbaee', 'table', '', '1d80abe6-5640-43f1-a86e-ebb1f0b86c38', NULL, 1, NULL, '{}', true, true, NULL);
INSERT INTO config.page_content VALUES ('3386289c-2bbf-4789-8dc6-4392ad8a5841', 'c0590f21-21f2-41b8-8a45-1293f0c194a4', 'table', '', 'b40125e5-1ab7-4429-b417-0afc444c88b8', NULL, 1, NULL, '{}', true, true, NULL);
INSERT INTO config.page_content VALUES ('e9a76657-db0e-4e46-9a30-c438d8be2344', 'aca655fd-1b0d-40b0-b02c-be56ebb7916e', 'table', '', '3a2290ba-fb36-4f8e-b66f-80e37a659a4a', NULL, 1, NULL, '{}', true, true, NULL);
INSERT INTO config.page_content VALUES ('19b544e6-3862-46b2-9b2d-e198537dc5ba', '2f7b7513-8a92-4922-acfa-41cc94e8a134', 'table', '', 'd06aab8d-6530-4c6d-8209-0c68f61657f0', NULL, 1, NULL, '{}', true, true, NULL);
INSERT INTO config.page_content VALUES ('1af4a4f4-49ee-476f-abe8-c35b40b9020d', '72747ce9-f63a-43e3-b2d0-51a3530ac527', 'chart', 'Orders by Day and Hour', NULL, NULL, 1, NULL, '{"colSpan": {"xs": 12}}', true, false, 'ea2bfa2a-27b7-4352-9086-03448283b82d');
INSERT INTO config.page_content VALUES ('143dfcef-1c22-45f7-9783-e4c15b75fee3', '72747ce9-f63a-43e3-b2d0-51a3530ac527', 'tabs', '', NULL, NULL, 2, NULL, '{"containerType": "tabs"}', true, false, NULL);
INSERT INTO config.page_content VALUES ('ba406714-8ae5-4335-9a5b-70e4bc9db50e', '72747ce9-f63a-43e3-b2d0-51a3530ac527', 'table', 'Items', 'b40125e5-1ab7-4429-b417-0afc444c88b8', NULL, 1, '143dfcef-1c22-45f7-9783-e4c15b75fee3', '{}', true, true, NULL);
INSERT INTO config.page_content VALUES ('02a9edd6-0ac9-435b-afee-8f678135bf6d', '72747ce9-f63a-43e3-b2d0-51a3530ac527', 'table', 'Warehouses', '1d80abe6-5640-43f1-a86e-ebb1f0b86c38', NULL, 2, '143dfcef-1c22-45f7-9783-e4c15b75fee3', '{}', true, false, NULL);
INSERT INTO config.page_content VALUES ('70955489-b449-442d-b536-bb3b39b1864d', '72747ce9-f63a-43e3-b2d0-51a3530ac527', 'table', 'Adjustments', '3a2290ba-fb36-4f8e-b66f-80e37a659a4a', NULL, 3, '143dfcef-1c22-45f7-9783-e4c15b75fee3', '{}', true, false, NULL);
INSERT INTO config.page_content VALUES ('a9a1eec4-2e82-4127-94da-42ce6a0ca91d', '72747ce9-f63a-43e3-b2d0-51a3530ac527', 'table', 'Transfers', 'd06aab8d-6530-4c6d-8209-0c68f61657f0', NULL, 4, '143dfcef-1c22-45f7-9783-e4c15b75fee3', '{}', true, false, NULL);
INSERT INTO config.page_content VALUES ('8128cba4-db18-4413-9f6d-63660319d78d', 'e96c1df7-6c5c-463e-b604-da53cf8c9785', 'table', '', 'bf89212f-ef4e-4359-9db2-1a01eebb6a1e', NULL, 1, NULL, '{}', true, true, NULL);
INSERT INTO config.page_content VALUES ('1573581f-d61f-4000-8569-05354256563d', '8691a22f-7e0e-49d6-8143-4914d06ed6c4', 'table', '', '5c2adff6-21d9-4bff-9b52-cd19be221ad7', NULL, 1, NULL, '{}', true, true, NULL);
INSERT INTO config.page_content VALUES ('8913ac25-425d-40e7-a99e-ca4d72da9c92', '722d5db8-c3e8-4d35-9702-7d35e3dd7416', 'table', '', '25af3eaf-affd-465b-883d-ef78c7c6b2fe', NULL, 1, NULL, '{}', true, true, NULL);
INSERT INTO config.page_content VALUES ('bac4d0b9-f756-4de6-ad47-a8229ecd2c53', '1b20927b-8682-4525-b18f-f0e7d1e1d1de', 'table', '', 'e6a9cbba-9159-47a4-a390-1bd6b8afd672', NULL, 1, NULL, '{}', true, true, NULL);
INSERT INTO config.page_content VALUES ('5fd95981-2937-4f0b-b452-6cf9691e57e6', 'd25ec547-eb68-4847-8864-cf0d6532b913', 'chart', 'Total Revenue', NULL, NULL, 1, NULL, '{"colSpan": {"md": 4, "sm": 6, "xs": 12}}', true, false, '75fa5671-9cf4-4201-8e64-1ba67629226b');
INSERT INTO config.page_content VALUES ('5efac153-d1e1-4749-bb3a-57e9cd490e61', 'd25ec547-eb68-4847-8864-cf0d6532b913', 'chart', 'Total Orders', NULL, NULL, 2, NULL, '{"colSpan": {"md": 4, "sm": 6, "xs": 12}}', true, false, 'cc26a1a6-0b3c-42e2-b9a8-67345982f44a');
INSERT INTO config.page_content VALUES ('8c9e3767-02b2-4332-949b-aa3331b6cb28', 'd25ec547-eb68-4847-8864-cf0d6532b913', 'chart', 'Revenue Progress', NULL, NULL, 3, NULL, '{"colSpan": {"md": 4, "sm": 6, "xs": 12}}', true, false, '759cdca3-92dd-4f47-8835-9f4e30b2d974');
INSERT INTO config.page_content VALUES ('b8cf2f86-d487-4ae8-9114-dc896a3c2018', 'd25ec547-eb68-4847-8864-cf0d6532b913', 'tabs', '', NULL, NULL, 4, NULL, '{"containerType": "tabs"}', true, false, NULL);
INSERT INTO config.page_content VALUES ('346152f0-6443-4d5c-87d5-d7164759b620', 'd25ec547-eb68-4847-8864-cf0d6532b913', 'table', 'Orders', 'b3696a06-7755-453f-948a-92792c233583', NULL, 1, 'b8cf2f86-d487-4ae8-9114-dc896a3c2018', '{}', true, true, NULL);
INSERT INTO config.page_content VALUES ('b6c83cc1-3101-498f-99ee-143a86daaa6b', 'd25ec547-eb68-4847-8864-cf0d6532b913', 'table', 'Customers', 'e6a9cbba-9159-47a4-a390-1bd6b8afd672', NULL, 2, 'b8cf2f86-d487-4ae8-9114-dc896a3c2018', '{}', true, false, NULL);
INSERT INTO config.page_content VALUES ('eed4a3de-d522-4afd-84cf-4f45f4c5a43b', '47739fd5-ee02-4180-af5a-8db38bb41d5e', 'table', '', 'fc870b12-0cad-480e-8990-e9e720717d79', NULL, 1, NULL, '{}', true, true, NULL);
INSERT INTO config.page_content VALUES ('7cc36998-af48-4959-8198-3b205fd37fb0', 'e570d44d-6cc7-44d0-a146-9294598832cb', 'table', '', '65611cb2-8457-41cc-b882-c55b4b8c9db7', NULL, 1, NULL, '{}', true, true, NULL);
INSERT INTO config.page_content VALUES ('8330d8b3-5216-4178-a706-f1b0e9edbd06', '3d224c0d-b1d6-4d3b-bde1-8455e390c99e', 'tabs', '', NULL, NULL, 1, NULL, '{"containerType": "tabs"}', true, false, NULL);
INSERT INTO config.page_content VALUES ('1c625918-61c0-43ac-82a7-ebe494b66c3a', '3d224c0d-b1d6-4d3b-bde1-8455e390c99e', 'table', 'Open', 'ae10ff70-f723-49a9-9c65-3fbb7aa4b02e', NULL, 1, '8330d8b3-5216-4178-a706-f1b0e9edbd06', '{}', true, true, NULL);
INSERT INTO config.page_content VALUES ('f30f83f2-a249-492a-93f7-14894ca25b4d', '3d224c0d-b1d6-4d3b-bde1-8455e390c99e', 'table', 'Closed', 'd00b0a2f-9ca4-4a7b-a618-e2bd219d161f', NULL, 2, '8330d8b3-5216-4178-a706-f1b0e9edbd06', '{}', true, false, NULL);
INSERT INTO config.page_content VALUES ('5aec64d4-843e-4632-a729-1fd9c5b126e4', '4852ffae-1b5c-4ed3-a941-a88970f349de', 'chart', 'Purchase Order Timeline', NULL, NULL, 1, NULL, '{"colSpan": {"xs": 12}}', true, false, '8bd7cd3e-7132-494f-80b3-78fac980ec70');
INSERT INTO config.page_content VALUES ('7905c91e-d173-407d-a954-37c159c9d7db', '4852ffae-1b5c-4ed3-a941-a88970f349de', 'tabs', '', NULL, NULL, 2, NULL, '{"containerType": "tabs"}', true, false, NULL);
INSERT INTO config.page_content VALUES ('6dd3bcd8-543c-4e2e-8de5-f4e0d939792d', '4852ffae-1b5c-4ed3-a941-a88970f349de', 'table', 'Purchase Orders', 'fc870b12-0cad-480e-8990-e9e720717d79', NULL, 1, '7905c91e-d173-407d-a954-37c159c9d7db', '{}', true, true, NULL);
INSERT INTO config.page_content VALUES ('982575d2-5463-4f0c-b6de-e5498bcfdd10', '4852ffae-1b5c-4ed3-a941-a88970f349de', 'table', 'Line Items', '65611cb2-8457-41cc-b882-c55b4b8c9db7', NULL, 2, '7905c91e-d173-407d-a954-37c159c9d7db', '{}', true, false, NULL);
INSERT INTO config.page_content VALUES ('ecb3b764-5be0-47ea-8c6d-6ae1670d5ccb', '4852ffae-1b5c-4ed3-a941-a88970f349de', 'table', 'Suppliers', 'aa8b6e2b-6d58-4c57-ac91-59eb02cab636', NULL, 3, '7905c91e-d173-407d-a954-37c159c9d7db', '{}', true, false, NULL);
INSERT INTO config.page_content VALUES ('4029d574-42cf-46c4-b9fc-79843a12f5a4', '4852ffae-1b5c-4ed3-a941-a88970f349de', 'table', 'Approvals', 'ae10ff70-f723-49a9-9c65-3fbb7aa4b02e', NULL, 4, '7905c91e-d173-407d-a954-37c159c9d7db', '{}', true, false, NULL);
INSERT INTO config.page_content VALUES ('ccc54fb8-3203-4195-bf9f-f64277b87a10', 'd21c14ee-f005-4b13-88c3-b4e38be0bb37', 'form', 'Create New User', NULL, '3a1a985d-bd73-4525-bb61-21a88afb0f34', 1, NULL, '{"colSpan": {"xs": 12}}', true, false, NULL);
INSERT INTO config.page_content VALUES ('33363886-0c20-454b-a553-9774dad2a442', 'd21c14ee-f005-4b13-88c3-b4e38be0bb37', 'tabs', 'User Lists', NULL, NULL, 2, NULL, '{"colSpan": {"xs": 12}, "containerType": "tabs"}', true, false, NULL);
INSERT INTO config.page_content VALUES ('80b81b1b-b17f-47b5-9b2a-4e0fd040156c', 'd21c14ee-f005-4b13-88c3-b4e38be0bb37', 'table', 'Active Users', '2b7deab8-b1f8-4b4c-8cf1-94a08f408bbb', NULL, 1, '33363886-0c20-454b-a553-9774dad2a442', '{}', true, true, NULL);
INSERT INTO config.page_content VALUES ('6b5f6900-a7ee-4fe0-a483-ddaaeffa3b47', 'd21c14ee-f005-4b13-88c3-b4e38be0bb37', 'table', 'Roles', 'e739a9ce-df34-4ea7-b639-18edc9386153', NULL, 2, '33363886-0c20-454b-a553-9774dad2a442', '{}', true, false, NULL);
INSERT INTO config.page_content VALUES ('bf39dcb6-d100-448e-aa5e-b3681ea52264', 'e4d6667c-8749-4cfd-ad50-2e0d450eb379', 'chart', 'Sales by Region', NULL, NULL, 1, NULL, '{"colSpan": {"md": 6, "xs": 12}}', true, true, 'f9c8256c-0bca-4ca8-82ff-c9ff8274a0f1');
INSERT INTO config.page_content VALUES ('4dcae6c5-7cfc-44c6-92d8-db509088dd59', 'e4d6667c-8749-4cfd-ad50-2e0d450eb379', 'chart', 'Cumulative Revenue', NULL, NULL, 2, NULL, '{"colSpan": {"md": 6, "xs": 12}}', true, false, 'e09cfa6b-855c-4c2c-9860-e6ae60f9bd0b');
INSERT INTO config.page_content VALUES ('9ba7cc59-e72e-4f8c-bd4b-1abbe11c2b7b', 'e4d6667c-8749-4cfd-ad50-2e0d450eb379', 'chart', 'Revenue vs Orders', NULL, NULL, 3, NULL, '{"colSpan": {"md": 6, "xs": 12}}', true, false, 'd2dc1f2e-f95d-440a-82b6-3fffaacd0554');
INSERT INTO config.page_content VALUES ('5eafaa54-3f66-4969-a57f-904ae25b5752', 'e4d6667c-8749-4cfd-ad50-2e0d450eb379', 'chart', 'Profit Breakdown', NULL, NULL, 4, NULL, '{"colSpan": {"md": 6, "xs": 12}}', true, false, 'efe7e335-9aa6-4865-b2ba-1f6b684f7c6e');


--
-- Data for Name: pages; Type: TABLE DATA; Schema: core; Owner: -
--

INSERT INTO core.pages VALUES ('0a56b876-9fb2-48ef-91d8-9b208532cb43', '/dashboard', 'Main Dashboard', 'dashboard', 'material-symbols:space-dashboard', 1, true, true);
INSERT INTO core.pages VALUES ('3f81404c-0302-4b4a-bb1b-a2b9195935cf', '/dashboard/analytics', 'Analytics Dashboard', 'dashboard', 'material-symbols:analytics', 2, true, true);
INSERT INTO core.pages VALUES ('47c40487-95cb-4857-97ee-e23d613aab14', '/alerts', 'Notification Center', 'dashboard', 'material-symbols:notifications', 3, true, true);
INSERT INTO core.pages VALUES ('9a9c42ab-89ed-4cca-b513-f0cbfdc0ae17', '/sales', 'Sales Dashboard', 'sales', 'material-symbols:point-of-sale', 4, true, true);
INSERT INTO core.pages VALUES ('9562f714-d8d0-425d-b596-1ba8acb3fc1b', '/sales/orders', 'Order Management', 'sales', 'material-symbols:receipt-long', 5, true, true);
INSERT INTO core.pages VALUES ('6a4a94d1-e883-4e20-998c-9e11c8bdf27d', '/sales/customers', 'Customer Management', 'sales', 'material-symbols:group', 6, true, true);
INSERT INTO core.pages VALUES ('bbd287e1-b32d-48f3-9c95-6d561ab87fe5', '/sales/reports', 'Sales Reports', 'sales', 'material-symbols:assessment', 7, true, true);
INSERT INTO core.pages VALUES ('534d16a5-98c8-45e7-9f32-c1c7e644b7bd', '/sales/orders/:id/invoice', 'Invoice', 'sales', 'material-symbols:receipt-long', 8, true, false);
INSERT INTO core.pages VALUES ('ca26d749-5ef1-42b7-b696-fa3e3f8adbe3', '/sales/orders/new', 'New Order Form', 'sales', 'material-symbols:add-shopping-cart', 9, true, false);
INSERT INTO core.pages VALUES ('d656006a-7f95-4af3-9155-760f7bcf0f08', '/sales/customers/new', 'New Customer Form', 'sales', 'material-symbols:person-add', 10, true, false);
INSERT INTO core.pages VALUES ('5dcea00e-ea6d-4af7-b2b8-b8d3ed699f03', '/sales/customers/:id', 'Customer Details', 'sales', 'material-symbols:person', 11, true, false);
INSERT INTO core.pages VALUES ('59e762b9-a21b-4c2c-a957-b1713752d2d0', '/sales/orders/:id', 'Order Details', 'sales', 'material-symbols:receipt', 12, true, false);
INSERT INTO core.pages VALUES ('902fd18b-9334-43f7-805c-65dc625ecfcc', '/inventory', 'Inventory Dashboard', 'inventory', 'material-symbols:inventory-2', 13, true, true);
INSERT INTO core.pages VALUES ('b642ef29-633b-465e-9b83-e50463571e02', '/inventory/items', 'Item Management', 'inventory', 'material-symbols:category', 14, true, true);
INSERT INTO core.pages VALUES ('407eaa7d-d207-4337-93b6-4def37d8aba0', '/inventory/warehouses', 'Warehouse Management', 'inventory', 'material-symbols:warehouse', 15, true, true);
INSERT INTO core.pages VALUES ('ce239ae5-3b45-465e-bd04-d281ad83b9be', '/inventory/transfers', 'Transfer Orders', 'inventory', 'material-symbols:sync-alt', 16, true, true);
INSERT INTO core.pages VALUES ('1a9bc6da-2182-4bb0-8a83-cf6b30030050', '/inventory/adjustments', 'Stock Adjustments', 'inventory', 'material-symbols:tune', 17, true, true);
INSERT INTO core.pages VALUES ('53bfb19d-cbdc-4424-9b3a-14a92fea8471', '/inventory/reports', 'Inventory Reports', 'inventory', 'material-symbols:summarize', 18, true, true);
INSERT INTO core.pages VALUES ('e3191652-d8d2-40e1-bf46-898ca891d027', '/inventory/locations', 'Location Management', 'inventory', 'material-symbols:location-on', 19, true, true);
INSERT INTO core.pages VALUES ('f88f9093-ec79-4315-ae1c-ff68d7e9f02f', '/inventory/zones', 'Zone Management', 'inventory', 'material-symbols:grid-view', 20, true, true);
INSERT INTO core.pages VALUES ('2e1dd7c6-b7d7-481b-9f8f-678e84368870', '/inventory/items/new', 'New Inventory Item', 'inventory', 'material-symbols:add-box', 21, true, false);
INSERT INTO core.pages VALUES ('0a7fc67f-c261-4af9-9b38-5e8937864ec3', '/inventory/items/:id', 'Inventory Item Details', 'inventory', 'material-symbols:inventory', 22, true, false);
INSERT INTO core.pages VALUES ('4efaf488-6415-46f7-9b53-68759ebaa457', '/inventory/warehouses/new', 'New Warehouse', 'inventory', 'material-symbols:add-business', 23, true, false);
INSERT INTO core.pages VALUES ('ead475c7-1596-4271-a883-a8ea885a5a19', '/inventory/warehouses/:id', 'Warehouse Details', 'inventory', 'material-symbols:domain', 24, true, false);
INSERT INTO core.pages VALUES ('eaab4894-f42e-4391-a985-10faa0890740', '/inventory/transfers/new', 'New Transfer Order', 'inventory', 'material-symbols:add-circle', 25, true, false);
INSERT INTO core.pages VALUES ('eb70f06f-5616-4fd3-a9d5-d3d750a5410f', '/inventory/transfers/:id', 'Transfer Order Details', 'inventory', 'material-symbols:compare-arrows', 26, true, false);
INSERT INTO core.pages VALUES ('66894134-c357-4f79-925b-2e9dfd6a8361', '/inventory/adjustments/new', 'New Stock Adjustment', 'inventory', 'material-symbols:edit-note', 27, true, false);
INSERT INTO core.pages VALUES ('f714a23f-ccda-4c50-bd72-c283a02816a8', '/inventory/adjustments/:id', 'Stock Adjustment Details', 'inventory', 'material-symbols:analytics', 28, true, false);
INSERT INTO core.pages VALUES ('47be3fc4-de35-4f09-aeaf-145835f81d9c', '/inventory/locations/new', 'New Location', 'inventory', 'material-symbols:add-location', 29, true, false);
INSERT INTO core.pages VALUES ('ae7a67f1-fbaf-49d7-8911-e54b17341a09', '/inventory/locations/:id', 'Location Details', 'inventory', 'material-symbols:location-pin', 30, true, false);
INSERT INTO core.pages VALUES ('6cb0c7fa-e00a-4d6e-b988-67950fec631b', '/inventory/zones/new', 'New Zone', 'inventory', 'material-symbols:add-circle', 31, true, false);
INSERT INTO core.pages VALUES ('8c86d922-7202-416f-a9d8-1eb4ab3549bb', '/inventory/zones/:id', 'Zone Details', 'inventory', 'material-symbols:area-chart', 32, true, false);
INSERT INTO core.pages VALUES ('c482f872-96c2-477a-b5d2-3d89c4272a3b', '/procurement', 'Procurement Dashboard', 'procurement', 'material-symbols:shopping-cart', 33, true, true);
INSERT INTO core.pages VALUES ('2c456e1f-f582-4255-933b-6a55bdca20a0', '/procurement/suppliers', 'Supplier Management', 'procurement', 'material-symbols:local-shipping', 34, true, true);
INSERT INTO core.pages VALUES ('e05210d2-85c8-40f9-977e-4f25ef4f5557', '/procurement/orders', 'Purchase Orders', 'procurement', 'material-symbols:shopping-bag', 35, true, true);
INSERT INTO core.pages VALUES ('738b9570-67ce-4503-a629-2a03696fb751', '/procurement/approvals', 'Approval Queue', 'procurement', 'material-symbols:check-circle', 36, true, true);
INSERT INTO core.pages VALUES ('a460c1d8-3b97-42e6-88a1-d1690a04c883', '/procurement/suppliers/new', 'New Supplier', 'procurement', 'material-symbols:add-business', 37, true, false);
INSERT INTO core.pages VALUES ('6f24d056-5732-400c-9755-2e821eeefb63', '/procurement/suppliers/:id', 'Supplier Details', 'procurement', 'material-symbols:business-center', 38, true, false);
INSERT INTO core.pages VALUES ('592d2e7d-fff5-4679-a140-8f5c28105e99', '/procurement/orders/new', 'New Purchase Order', 'procurement', 'material-symbols:note-add', 39, true, false);
INSERT INTO core.pages VALUES ('21e8243b-d89d-41c2-9d24-666a195091f1', '/procurement/orders/:id', 'Purchase Order Details', 'procurement', 'material-symbols:description', 40, true, false);
INSERT INTO core.pages VALUES ('02978c56-cc65-4ff9-8c40-9b923dbf24eb', '/procurement/approvals/:id', 'Procurement Approval Detail', 'procurement', 'material-symbols:approval', 41, true, false);
INSERT INTO core.pages VALUES ('6259aa1f-dfcc-4732-82cb-04a0acbfb6d7', '/products', 'Products', 'products', 'material-symbols:inventory', 42, true, true);
INSERT INTO core.pages VALUES ('ce22b40c-fba8-41fa-aa0e-594c5d372528', '/products/new', 'New Product', 'products', 'material-symbols:add-box', 43, true, false);
INSERT INTO core.pages VALUES ('c8f3eeca-80fc-47d1-9714-541ac252a433', '/products/:id', 'Product Details', 'products', 'material-symbols:package-2', 44, true, false);
INSERT INTO core.pages VALUES ('e1a01fba-2543-483d-9cc9-0b18e68ce3d0', '/assets', 'Asset Dashboard', 'assets', 'material-symbols:apartment', 45, true, true);
INSERT INTO core.pages VALUES ('934cab7b-802d-449f-a7fe-fea5c3e81e55', '/assets/list', 'Asset List', 'assets', 'material-symbols:list', 46, true, true);
INSERT INTO core.pages VALUES ('2cec9547-5adb-49f8-8e21-2bfab4df1d48', '/assets/requests', 'Asset Requests', 'assets', 'material-symbols:request-quote', 47, true, true);
INSERT INTO core.pages VALUES ('563e6a0d-890a-417f-a648-10228aacda16', '/assets/maintenance', 'Maintenance Schedule', 'assets', 'material-symbols:build', 48, true, true);
INSERT INTO core.pages VALUES ('3851a545-2df5-41e5-bb54-5ba78802be8a', '/assets/list/new', 'New Asset', 'assets', 'material-symbols:add-circle', 49, true, false);
INSERT INTO core.pages VALUES ('4c18d6e0-abcc-4bd1-a293-80dd3d72aacd', '/assets/list/:id', 'Asset Details', 'assets', 'material-symbols:apartment', 50, true, false);
INSERT INTO core.pages VALUES ('bbeaa334-356f-4fb4-a2a5-7521df2948d5', '/hr', 'HR Dashboard', 'hr', 'material-symbols:badge', 51, true, true);
INSERT INTO core.pages VALUES ('91931f50-f456-4204-9f59-28e74aba4b16', '/hr/employees', 'Employee Directory', 'hr', 'material-symbols:groups', 52, true, true);
INSERT INTO core.pages VALUES ('866d2a71-b032-477b-832d-c214a424da8e', '/hr/onboarding', 'Onboarding', 'hr', 'material-symbols:how-to-reg', 53, true, true);
INSERT INTO core.pages VALUES ('7a335164-b9ac-448d-8b3a-789432220c4d', '/hr/offices', 'Office Management', 'hr', 'material-symbols:business', 54, true, true);
INSERT INTO core.pages VALUES ('ba20b037-1db8-4afb-a5ec-db38591f6a96', '/hr/reports', 'HR Reports', 'hr', 'material-symbols:bar-chart', 55, true, true);
INSERT INTO core.pages VALUES ('f673496e-4a28-4b37-bc0b-d995e918e842', '/hr/employees/new', 'New Employee', 'hr', 'material-symbols:person-add', 56, true, false);
INSERT INTO core.pages VALUES ('b047152b-d150-4099-bcf6-7e06fd68966e', '/hr/employees/:id', 'Employee Details', 'hr', 'material-symbols:badge', 57, true, false);
INSERT INTO core.pages VALUES ('f1ce8d09-bb46-4e41-b203-6637ddfcb52a', '/hr/offices/new', 'New Office', 'hr', 'material-symbols:add-business', 58, true, false);
INSERT INTO core.pages VALUES ('e83d3601-8bee-4cf2-bca2-ad7feb3fcc5a', '/hr/offices/:id', 'Office Details', 'hr', 'material-symbols:domain', 59, true, false);
INSERT INTO core.pages VALUES ('0ee4de80-e861-4981-8371-de0743714995', '/admin', 'Admin Dashboard', 'admin', 'material-symbols:admin-panel-settings', 60, true, true);
INSERT INTO core.pages VALUES ('c1484b63-ccff-4b80-877f-7994c1be12ec', '/admin/users', 'User Management', 'admin', 'material-symbols:manage-accounts', 61, true, true);
INSERT INTO core.pages VALUES ('e74586b2-1992-4a5a-a257-0172a2b8007a', '/admin/roles', 'Role Management', 'admin', 'material-symbols:security', 62, true, true);
INSERT INTO core.pages VALUES ('0156768c-8eee-42aa-94e8-728fd1421617', '/admin/config', 'System Configuration', 'admin', 'material-symbols:settings-applications', 63, true, true);
INSERT INTO core.pages VALUES ('b9c1ac35-ae92-4ae3-841d-d5aab5a8eedd', '/admin/audit', 'Audit Logs', 'admin', 'material-symbols:history', 64, true, true);
INSERT INTO core.pages VALUES ('9b603af3-96d2-4113-b8ec-1a49039052b4', '/admin/config/pages', 'Page Configs', 'admin', 'material-symbols:page-info', 65, true, true);
INSERT INTO core.pages VALUES ('1c9f6ea7-fdcb-4805-9aad-a5e1171c33de', '/admin/config/tables', 'Table Configs', 'admin', 'material-symbols:table', 66, true, true);
INSERT INTO core.pages VALUES ('68e2e884-7316-40fa-9a02-0a67634fcf21', '/admin/config/forms', 'Form Configs', 'admin', 'material-symbols:edit-document', 67, true, true);
INSERT INTO core.pages VALUES ('4f6fb988-620a-4d65-b66d-464b38695d35', '/admin/config/import-export', 'Import/Export', 'admin', 'material-symbols:import-export', 68, true, true);
INSERT INTO core.pages VALUES ('c0244de2-d1ef-4534-8190-1c15df8f8c60', '/admin/users/new', 'New User', 'admin', 'material-symbols:person-add', 69, true, false);
INSERT INTO core.pages VALUES ('9627387c-c188-4ed0-a427-a4fbf0b2b71e', '/admin/users/:id', 'User Details', 'admin', 'material-symbols:account-circle', 70, true, false);
INSERT INTO core.pages VALUES ('acd3b0b5-bcc6-41de-82b0-2cf665b094a0', '/admin/roles/new', 'New Role', 'admin', 'material-symbols:add-moderator', 71, true, false);
INSERT INTO core.pages VALUES ('9e361079-d1c0-4891-a048-4c303c5e082d', '/admin/roles/:id', 'Role Details', 'admin', 'material-symbols:admin-panel-settings', 72, true, false);
INSERT INTO core.pages VALUES ('412e09b1-f8b0-4d9f-80d4-1bda76a0d747', '/admin/config/page-configs/new', 'New Page Config', 'admin', 'material-symbols:add-circle', 73, true, false);
INSERT INTO core.pages VALUES ('96fb7765-abee-4051-af2c-049c0a04a139', '/admin/config/page-configs/:id', 'Edit Page Config', 'admin', 'material-symbols:edit', 74, true, false);
INSERT INTO core.pages VALUES ('6a60c808-936e-440f-94b5-282fd1794fa8', '/admin/config/page-content/:id', 'Page Content Editor', 'admin', 'material-symbols:edit-note', 75, true, false);
INSERT INTO core.pages VALUES ('815206a5-9c03-4962-a01c-6965b22f2699', '/admin/config/page-actions/:page_config_id', 'Page Actions Editor', 'admin', 'material-symbols:touch-app', 76, true, false);
INSERT INTO core.pages VALUES ('562af7fd-80ef-4715-83f3-1ce631c2d19f', '/admin/config/table-configs/new', 'New Table Config', 'admin', 'material-symbols:add-circle', 77, true, false);
INSERT INTO core.pages VALUES ('2239f427-14d5-4e62-911f-d9a45961799e', '/admin/config/table-configs/:id', 'Edit Table Config', 'admin', 'material-symbols:edit', 78, true, false);
INSERT INTO core.pages VALUES ('be97ead5-b671-4a1e-8663-1dc98912f1a3', '/admin/config/forms/new', 'New Form Config', 'admin', 'material-symbols:add-circle', 79, true, false);
INSERT INTO core.pages VALUES ('2e5e7d51-658d-4c06-b944-f56781ad4f95', '/admin/config/forms/:id', 'Edit Form Config', 'admin', 'material-symbols:edit', 80, true, false);
INSERT INTO core.pages VALUES ('591d352f-e777-4994-84c6-f99939780795', '/admin/config/:id', 'Edit Config', 'admin', 'material-symbols:edit', 81, true, false);
INSERT INTO core.pages VALUES ('9aae669f-94b5-48db-b599-fdf7126e6654', '/admin/config/simple-tables/:id', 'Simple Table', 'admin', 'material-symbols:edit', 82, true, false);
INSERT INTO core.pages VALUES ('a9d03946-dfe2-4fdc-95f6-26edf5ccce00', '/admin/config/simple-tables/', 'Simple Tables', 'admin', 'material-symbols:edit', 83, true, false);
INSERT INTO core.pages VALUES ('26f63361-0c41-481c-af2d-4bf42945ff9f', '/admin/config/simple-charts/:id', 'Simple Charts', 'admin', 'material-symbols:edit', 84, true, false);
INSERT INTO core.pages VALUES ('5db855bd-6132-4987-8c26-2f5b65b96af7', '/workflow', 'Automation Rules', 'workflow', 'material-symbols:account-tree', 85, true, true);
INSERT INTO core.pages VALUES ('45962d9a-2141-402f-9557-f26fb184930a', '/workflow/executions', 'Execution History', 'workflow', 'material-symbols:history', 86, true, true);
INSERT INTO core.pages VALUES ('7abf1232-f454-4292-8e5a-8b95f64db038', '/workflow/editor', 'Workflow Editor', 'workflow', 'material-symbols:history', 87, true, true);
INSERT INTO core.pages VALUES ('b52c625e-5daf-488e-a31a-3fcd73f767d7', '/workflow/editor/:id', 'Workflow Editor', 'workflow', 'material-symbols:schema', 88, true, false);


--
-- Data for Name: role_pages; Type: TABLE DATA; Schema: core; Owner: -
--

INSERT INTO core.role_pages VALUES ('bccfc451-4103-42f2-ad5c-0ef5f7fd24de', '54bb2165-71e1-41a6-af3e-7da4a0e1e2c1', '0a56b876-9fb2-48ef-91d8-9b208532cb43', true);
INSERT INTO core.role_pages VALUES ('ebced783-7604-4e16-a2f8-2be2f340c9f6', '54bb2165-71e1-41a6-af3e-7da4a0e1e2c1', '3f81404c-0302-4b4a-bb1b-a2b9195935cf', true);
INSERT INTO core.role_pages VALUES ('7864a6a6-4e9b-4373-998f-bf3c2142d619', '54bb2165-71e1-41a6-af3e-7da4a0e1e2c1', '47c40487-95cb-4857-97ee-e23d613aab14', true);
INSERT INTO core.role_pages VALUES ('7d03655b-3523-4a67-9f5a-4221eed00e2e', '54bb2165-71e1-41a6-af3e-7da4a0e1e2c1', '9a9c42ab-89ed-4cca-b513-f0cbfdc0ae17', true);
INSERT INTO core.role_pages VALUES ('dc507864-48aa-41a5-8e6d-797b0e248613', '54bb2165-71e1-41a6-af3e-7da4a0e1e2c1', '9562f714-d8d0-425d-b596-1ba8acb3fc1b', true);
INSERT INTO core.role_pages VALUES ('2eb2bcdf-24c4-4fc3-8280-da11ee6bc1ac', '54bb2165-71e1-41a6-af3e-7da4a0e1e2c1', '6a4a94d1-e883-4e20-998c-9e11c8bdf27d', true);
INSERT INTO core.role_pages VALUES ('abe5a1c3-4a47-4e66-938f-db87ab2750f6', '54bb2165-71e1-41a6-af3e-7da4a0e1e2c1', 'bbd287e1-b32d-48f3-9c95-6d561ab87fe5', true);
INSERT INTO core.role_pages VALUES ('6eb0bfa8-e08b-408a-8ec0-37f95cdc31b2', '54bb2165-71e1-41a6-af3e-7da4a0e1e2c1', '534d16a5-98c8-45e7-9f32-c1c7e644b7bd', true);
INSERT INTO core.role_pages VALUES ('e9f28432-c93c-4053-aa85-616eb149ab37', '54bb2165-71e1-41a6-af3e-7da4a0e1e2c1', 'ca26d749-5ef1-42b7-b696-fa3e3f8adbe3', true);
INSERT INTO core.role_pages VALUES ('7549da34-fb61-4d78-b428-a8c264a84126', '54bb2165-71e1-41a6-af3e-7da4a0e1e2c1', 'd656006a-7f95-4af3-9155-760f7bcf0f08', true);
INSERT INTO core.role_pages VALUES ('c42568cc-fd53-4236-8cde-a2cd68c131ff', '54bb2165-71e1-41a6-af3e-7da4a0e1e2c1', '5dcea00e-ea6d-4af7-b2b8-b8d3ed699f03', true);
INSERT INTO core.role_pages VALUES ('4bdba359-da85-4b63-a9b7-32a4a8328fd2', '54bb2165-71e1-41a6-af3e-7da4a0e1e2c1', '59e762b9-a21b-4c2c-a957-b1713752d2d0', true);
INSERT INTO core.role_pages VALUES ('11216851-9e2b-4577-9d32-6cfca7e6136b', '54bb2165-71e1-41a6-af3e-7da4a0e1e2c1', '902fd18b-9334-43f7-805c-65dc625ecfcc', true);
INSERT INTO core.role_pages VALUES ('740eaa3c-6d7b-4026-8e43-52ca2708a4e0', '54bb2165-71e1-41a6-af3e-7da4a0e1e2c1', 'b642ef29-633b-465e-9b83-e50463571e02', true);
INSERT INTO core.role_pages VALUES ('039b1593-1986-4104-94ba-f6f109608b2a', '54bb2165-71e1-41a6-af3e-7da4a0e1e2c1', '407eaa7d-d207-4337-93b6-4def37d8aba0', true);
INSERT INTO core.role_pages VALUES ('d655ac6c-a059-42be-8284-7b0efa811092', '54bb2165-71e1-41a6-af3e-7da4a0e1e2c1', 'ce239ae5-3b45-465e-bd04-d281ad83b9be', true);
INSERT INTO core.role_pages VALUES ('dc026692-09b3-40d1-8997-970047a313f1', '54bb2165-71e1-41a6-af3e-7da4a0e1e2c1', '1a9bc6da-2182-4bb0-8a83-cf6b30030050', true);
INSERT INTO core.role_pages VALUES ('dfa1853e-f8ae-46ee-b56e-327ac08df5b7', '54bb2165-71e1-41a6-af3e-7da4a0e1e2c1', '53bfb19d-cbdc-4424-9b3a-14a92fea8471', true);
INSERT INTO core.role_pages VALUES ('d3e1c0e5-72d1-4af5-8ed2-974fc076df30', '54bb2165-71e1-41a6-af3e-7da4a0e1e2c1', 'e3191652-d8d2-40e1-bf46-898ca891d027', true);
INSERT INTO core.role_pages VALUES ('c1724524-caaa-4243-a1c7-79680fa28f90', '54bb2165-71e1-41a6-af3e-7da4a0e1e2c1', 'f88f9093-ec79-4315-ae1c-ff68d7e9f02f', true);
INSERT INTO core.role_pages VALUES ('9bd18968-9090-4237-9979-5110b8dcf7c1', '54bb2165-71e1-41a6-af3e-7da4a0e1e2c1', '2e1dd7c6-b7d7-481b-9f8f-678e84368870', true);
INSERT INTO core.role_pages VALUES ('2ff1ba19-26ef-44c1-92b2-baa876d849df', '54bb2165-71e1-41a6-af3e-7da4a0e1e2c1', '0a7fc67f-c261-4af9-9b38-5e8937864ec3', true);
INSERT INTO core.role_pages VALUES ('e8e2fabc-15f2-4a9e-8f71-e897114f802f', '54bb2165-71e1-41a6-af3e-7da4a0e1e2c1', '4efaf488-6415-46f7-9b53-68759ebaa457', true);
INSERT INTO core.role_pages VALUES ('1fb5c3dc-9ccd-4e5d-a530-a67156d38c5e', '54bb2165-71e1-41a6-af3e-7da4a0e1e2c1', 'ead475c7-1596-4271-a883-a8ea885a5a19', true);
INSERT INTO core.role_pages VALUES ('a1deec8a-34bf-4eb3-8ed5-0017d59af4bf', '54bb2165-71e1-41a6-af3e-7da4a0e1e2c1', 'eaab4894-f42e-4391-a985-10faa0890740', true);
INSERT INTO core.role_pages VALUES ('b39c55ca-7d84-4525-a74c-debdc434aec4', '54bb2165-71e1-41a6-af3e-7da4a0e1e2c1', 'eb70f06f-5616-4fd3-a9d5-d3d750a5410f', true);
INSERT INTO core.role_pages VALUES ('8542f7e8-6fc7-4dbc-a13c-5a4c6eddc43b', '54bb2165-71e1-41a6-af3e-7da4a0e1e2c1', '66894134-c357-4f79-925b-2e9dfd6a8361', true);
INSERT INTO core.role_pages VALUES ('8298d019-8fb3-49a5-87e1-6493ad4899f6', '54bb2165-71e1-41a6-af3e-7da4a0e1e2c1', 'f714a23f-ccda-4c50-bd72-c283a02816a8', true);
INSERT INTO core.role_pages VALUES ('4746c95e-6b55-43a9-96de-96b969dfb3c3', '54bb2165-71e1-41a6-af3e-7da4a0e1e2c1', '47be3fc4-de35-4f09-aeaf-145835f81d9c', true);
INSERT INTO core.role_pages VALUES ('de0c61b4-10d9-4e73-b395-1368cd0901e2', '54bb2165-71e1-41a6-af3e-7da4a0e1e2c1', 'ae7a67f1-fbaf-49d7-8911-e54b17341a09', true);
INSERT INTO core.role_pages VALUES ('e4f51c3c-13bb-4ae0-92db-bb104b56377d', '54bb2165-71e1-41a6-af3e-7da4a0e1e2c1', '6cb0c7fa-e00a-4d6e-b988-67950fec631b', true);
INSERT INTO core.role_pages VALUES ('0c959493-f287-4419-b7a6-f28d795516b2', '54bb2165-71e1-41a6-af3e-7da4a0e1e2c1', '8c86d922-7202-416f-a9d8-1eb4ab3549bb', true);
INSERT INTO core.role_pages VALUES ('cfc30306-6113-4712-a5aa-ca5f8ef65e82', '54bb2165-71e1-41a6-af3e-7da4a0e1e2c1', 'c482f872-96c2-477a-b5d2-3d89c4272a3b', true);
INSERT INTO core.role_pages VALUES ('a812c44d-611c-4e5b-b8f0-05cc5afd2374', '54bb2165-71e1-41a6-af3e-7da4a0e1e2c1', '2c456e1f-f582-4255-933b-6a55bdca20a0', true);
INSERT INTO core.role_pages VALUES ('3b4f094d-0fc4-4546-9ea5-a1217fa95301', '54bb2165-71e1-41a6-af3e-7da4a0e1e2c1', 'e05210d2-85c8-40f9-977e-4f25ef4f5557', true);
INSERT INTO core.role_pages VALUES ('cab37a3e-cd38-4640-b566-698d1ba5072d', '54bb2165-71e1-41a6-af3e-7da4a0e1e2c1', '738b9570-67ce-4503-a629-2a03696fb751', true);
INSERT INTO core.role_pages VALUES ('a92013f1-38c4-4a7c-aab0-6a638b8e4164', '54bb2165-71e1-41a6-af3e-7da4a0e1e2c1', 'a460c1d8-3b97-42e6-88a1-d1690a04c883', true);
INSERT INTO core.role_pages VALUES ('012917f1-c362-4981-a2c0-42b1b3ba70cc', '54bb2165-71e1-41a6-af3e-7da4a0e1e2c1', '6f24d056-5732-400c-9755-2e821eeefb63', true);
INSERT INTO core.role_pages VALUES ('09bda91b-2fdd-438e-89cb-7f2d15c9f8ef', '54bb2165-71e1-41a6-af3e-7da4a0e1e2c1', '592d2e7d-fff5-4679-a140-8f5c28105e99', true);
INSERT INTO core.role_pages VALUES ('0d908e4f-92c6-497a-a228-805c515d5cfd', '54bb2165-71e1-41a6-af3e-7da4a0e1e2c1', '21e8243b-d89d-41c2-9d24-666a195091f1', true);
INSERT INTO core.role_pages VALUES ('eb1052fd-da41-40c3-9b9b-4876a195f445', '54bb2165-71e1-41a6-af3e-7da4a0e1e2c1', '02978c56-cc65-4ff9-8c40-9b923dbf24eb', true);
INSERT INTO core.role_pages VALUES ('02e0a4ab-c7b8-4479-95af-3b7e9bf26fc5', '54bb2165-71e1-41a6-af3e-7da4a0e1e2c1', '6259aa1f-dfcc-4732-82cb-04a0acbfb6d7', true);
INSERT INTO core.role_pages VALUES ('c10393d8-abbd-43ac-ab55-3c175bf6d7cc', '54bb2165-71e1-41a6-af3e-7da4a0e1e2c1', 'ce22b40c-fba8-41fa-aa0e-594c5d372528', true);
INSERT INTO core.role_pages VALUES ('aa7ae0fd-e29e-4ad6-a418-99f635a3dfb1', '54bb2165-71e1-41a6-af3e-7da4a0e1e2c1', 'c8f3eeca-80fc-47d1-9714-541ac252a433', true);
INSERT INTO core.role_pages VALUES ('9f9793f8-250e-40e8-a643-e8cab80d27e3', '54bb2165-71e1-41a6-af3e-7da4a0e1e2c1', 'e1a01fba-2543-483d-9cc9-0b18e68ce3d0', true);
INSERT INTO core.role_pages VALUES ('d2124f47-878c-4948-9af9-dab0bcc3841b', '54bb2165-71e1-41a6-af3e-7da4a0e1e2c1', '934cab7b-802d-449f-a7fe-fea5c3e81e55', true);
INSERT INTO core.role_pages VALUES ('0aae52e8-4f11-44ec-9bb6-680a45cdc7ac', '54bb2165-71e1-41a6-af3e-7da4a0e1e2c1', '2cec9547-5adb-49f8-8e21-2bfab4df1d48', true);
INSERT INTO core.role_pages VALUES ('bc6e8955-d020-4048-8b45-1263aa26a4de', '54bb2165-71e1-41a6-af3e-7da4a0e1e2c1', '563e6a0d-890a-417f-a648-10228aacda16', true);
INSERT INTO core.role_pages VALUES ('0e64f08f-4b19-4605-b4ca-babdf7f2d0bc', '54bb2165-71e1-41a6-af3e-7da4a0e1e2c1', '3851a545-2df5-41e5-bb54-5ba78802be8a', true);
INSERT INTO core.role_pages VALUES ('0bdeb1eb-e727-486b-a2d3-28cbcac0c071', '54bb2165-71e1-41a6-af3e-7da4a0e1e2c1', '4c18d6e0-abcc-4bd1-a293-80dd3d72aacd', true);
INSERT INTO core.role_pages VALUES ('452cc934-7e85-40df-99d7-ba6983017dd2', '54bb2165-71e1-41a6-af3e-7da4a0e1e2c1', 'bbeaa334-356f-4fb4-a2a5-7521df2948d5', true);
INSERT INTO core.role_pages VALUES ('27ccac78-8ff6-44d2-bc9a-1cbfeb49f94e', '54bb2165-71e1-41a6-af3e-7da4a0e1e2c1', '91931f50-f456-4204-9f59-28e74aba4b16', true);
INSERT INTO core.role_pages VALUES ('b49cbb51-d978-4f82-af00-4e1d0fedda09', '54bb2165-71e1-41a6-af3e-7da4a0e1e2c1', '866d2a71-b032-477b-832d-c214a424da8e', true);
INSERT INTO core.role_pages VALUES ('ddc674bb-f0a0-46f7-a498-7cac1fd9a6c8', '54bb2165-71e1-41a6-af3e-7da4a0e1e2c1', '7a335164-b9ac-448d-8b3a-789432220c4d', true);
INSERT INTO core.role_pages VALUES ('8df39306-2653-4635-b4c8-d93087e32b71', '54bb2165-71e1-41a6-af3e-7da4a0e1e2c1', 'ba20b037-1db8-4afb-a5ec-db38591f6a96', true);
INSERT INTO core.role_pages VALUES ('b4e63ba4-934a-4bf3-be58-63512f854e22', '54bb2165-71e1-41a6-af3e-7da4a0e1e2c1', 'f673496e-4a28-4b37-bc0b-d995e918e842', true);
INSERT INTO core.role_pages VALUES ('d73aa507-e600-44cd-b7a3-f437a8dcea1b', '54bb2165-71e1-41a6-af3e-7da4a0e1e2c1', 'b047152b-d150-4099-bcf6-7e06fd68966e', true);
INSERT INTO core.role_pages VALUES ('9c727f95-158b-45c8-9b1a-050d2b580675', '54bb2165-71e1-41a6-af3e-7da4a0e1e2c1', 'f1ce8d09-bb46-4e41-b203-6637ddfcb52a', true);
INSERT INTO core.role_pages VALUES ('1862f765-f2c3-49b4-b093-8154d96c52a3', '54bb2165-71e1-41a6-af3e-7da4a0e1e2c1', 'e83d3601-8bee-4cf2-bca2-ad7feb3fcc5a', true);
INSERT INTO core.role_pages VALUES ('1492cd7d-5822-406a-980c-f0236e596d46', '54bb2165-71e1-41a6-af3e-7da4a0e1e2c1', '0ee4de80-e861-4981-8371-de0743714995', true);
INSERT INTO core.role_pages VALUES ('7b30dff8-0b84-4c2a-bd4d-36bbd8ebb810', '54bb2165-71e1-41a6-af3e-7da4a0e1e2c1', 'c1484b63-ccff-4b80-877f-7994c1be12ec', true);
INSERT INTO core.role_pages VALUES ('af056f74-00c5-47c2-aa26-92cfc6915c70', '54bb2165-71e1-41a6-af3e-7da4a0e1e2c1', 'e74586b2-1992-4a5a-a257-0172a2b8007a', true);
INSERT INTO core.role_pages VALUES ('a5bbb42c-387b-4d33-852d-d2363ebfa1e5', '54bb2165-71e1-41a6-af3e-7da4a0e1e2c1', '0156768c-8eee-42aa-94e8-728fd1421617', true);
INSERT INTO core.role_pages VALUES ('05abdc0f-151b-4bd9-b78f-617d0a39daae', '54bb2165-71e1-41a6-af3e-7da4a0e1e2c1', 'b9c1ac35-ae92-4ae3-841d-d5aab5a8eedd', true);
INSERT INTO core.role_pages VALUES ('99ca3436-8d26-48ec-8e03-ed5b4abf7f5b', '54bb2165-71e1-41a6-af3e-7da4a0e1e2c1', '9b603af3-96d2-4113-b8ec-1a49039052b4', true);
INSERT INTO core.role_pages VALUES ('73046481-2296-464a-93e1-c8f4da09772f', '54bb2165-71e1-41a6-af3e-7da4a0e1e2c1', '1c9f6ea7-fdcb-4805-9aad-a5e1171c33de', true);
INSERT INTO core.role_pages VALUES ('2f1c6a2b-1125-47b3-95a2-79d0ce42cc91', '54bb2165-71e1-41a6-af3e-7da4a0e1e2c1', '68e2e884-7316-40fa-9a02-0a67634fcf21', true);
INSERT INTO core.role_pages VALUES ('086b3fee-8dff-4e80-8e3e-1f4b39651d31', '54bb2165-71e1-41a6-af3e-7da4a0e1e2c1', '4f6fb988-620a-4d65-b66d-464b38695d35', true);
INSERT INTO core.role_pages VALUES ('a4e8c278-8b99-4edb-bb43-6e02ae3aba70', '54bb2165-71e1-41a6-af3e-7da4a0e1e2c1', 'c0244de2-d1ef-4534-8190-1c15df8f8c60', true);
INSERT INTO core.role_pages VALUES ('574bca2c-1d43-4231-9115-602e6808eea6', '54bb2165-71e1-41a6-af3e-7da4a0e1e2c1', '9627387c-c188-4ed0-a427-a4fbf0b2b71e', true);
INSERT INTO core.role_pages VALUES ('89b03049-fb3b-47fd-9c02-ebff8cf6acd6', '54bb2165-71e1-41a6-af3e-7da4a0e1e2c1', 'acd3b0b5-bcc6-41de-82b0-2cf665b094a0', true);
INSERT INTO core.role_pages VALUES ('f71a8b6c-602d-4ecd-b3a0-583aa9e6495f', '54bb2165-71e1-41a6-af3e-7da4a0e1e2c1', '9e361079-d1c0-4891-a048-4c303c5e082d', true);
INSERT INTO core.role_pages VALUES ('8b7f8b7e-8c8c-4914-ad86-9f5e2f8203d1', '54bb2165-71e1-41a6-af3e-7da4a0e1e2c1', '412e09b1-f8b0-4d9f-80d4-1bda76a0d747', true);
INSERT INTO core.role_pages VALUES ('db625218-06e4-4df2-9864-1f9f2a15075b', '54bb2165-71e1-41a6-af3e-7da4a0e1e2c1', '96fb7765-abee-4051-af2c-049c0a04a139', true);
INSERT INTO core.role_pages VALUES ('127f7567-4809-476f-8f61-ea4afd4ec2ea', '54bb2165-71e1-41a6-af3e-7da4a0e1e2c1', '6a60c808-936e-440f-94b5-282fd1794fa8', true);
INSERT INTO core.role_pages VALUES ('95dc70a9-22b9-4ac0-90ce-665a8d3a301a', '54bb2165-71e1-41a6-af3e-7da4a0e1e2c1', '815206a5-9c03-4962-a01c-6965b22f2699', true);
INSERT INTO core.role_pages VALUES ('14102e43-3ddc-44ed-b68f-a05fd29857f3', '54bb2165-71e1-41a6-af3e-7da4a0e1e2c1', '562af7fd-80ef-4715-83f3-1ce631c2d19f', true);
INSERT INTO core.role_pages VALUES ('23eaf654-0b7a-4211-a5b2-872c2dcebb55', '54bb2165-71e1-41a6-af3e-7da4a0e1e2c1', '2239f427-14d5-4e62-911f-d9a45961799e', true);
INSERT INTO core.role_pages VALUES ('eb7c5d7a-58eb-4e9e-bbd8-c8d738ac9887', '54bb2165-71e1-41a6-af3e-7da4a0e1e2c1', 'be97ead5-b671-4a1e-8663-1dc98912f1a3', true);
INSERT INTO core.role_pages VALUES ('19adb80e-e7e3-4adf-b0f2-198f8c9075be', '54bb2165-71e1-41a6-af3e-7da4a0e1e2c1', '2e5e7d51-658d-4c06-b944-f56781ad4f95', true);
INSERT INTO core.role_pages VALUES ('de8091f5-6d6b-4862-8da4-3ca3aad2d529', '54bb2165-71e1-41a6-af3e-7da4a0e1e2c1', '591d352f-e777-4994-84c6-f99939780795', true);
INSERT INTO core.role_pages VALUES ('6d1366a6-5638-4a9a-a2cc-1f5e8fcdbe68', '54bb2165-71e1-41a6-af3e-7da4a0e1e2c1', '9aae669f-94b5-48db-b599-fdf7126e6654', true);
INSERT INTO core.role_pages VALUES ('28ac8dd5-cbb1-462f-bf6b-5539b38602b9', '54bb2165-71e1-41a6-af3e-7da4a0e1e2c1', 'a9d03946-dfe2-4fdc-95f6-26edf5ccce00', true);
INSERT INTO core.role_pages VALUES ('aa8b268a-96a3-4047-8756-bc59be421d86', '54bb2165-71e1-41a6-af3e-7da4a0e1e2c1', '26f63361-0c41-481c-af2d-4bf42945ff9f', true);
INSERT INTO core.role_pages VALUES ('40a68667-1fe1-42e2-a7fc-6c655259e42d', '54bb2165-71e1-41a6-af3e-7da4a0e1e2c1', '5db855bd-6132-4987-8c26-2f5b65b96af7', true);
INSERT INTO core.role_pages VALUES ('2ce95074-a2d5-4d86-8648-06ceff31099c', '54bb2165-71e1-41a6-af3e-7da4a0e1e2c1', '45962d9a-2141-402f-9557-f26fb184930a', true);
INSERT INTO core.role_pages VALUES ('6f8c9ec3-b1a4-4a63-9046-23101f390f9c', '54bb2165-71e1-41a6-af3e-7da4a0e1e2c1', '7abf1232-f454-4292-8e5a-8b95f64db038', true);
INSERT INTO core.role_pages VALUES ('08fffc1e-0235-4561-8eef-a5b3b2191309', '54bb2165-71e1-41a6-af3e-7da4a0e1e2c1', 'b52c625e-5daf-488e-a31a-3fcd73f767d7', true);


--
-- PostgreSQL database dump complete
--

\unrestrict cgMJO8iAfo4fcvwvxq9Rc0iBAWD3byfeQ2Kp1nUE9TDeQ65NAdb6AJ4k36mnVkw

COMMIT;
